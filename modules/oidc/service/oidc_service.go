package service

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"strings"
	"time"

	"nid-backend/modules/oidc/dto"
	"nid-backend/modules/oidc/repository"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type OIDCService struct {
	repo       *repository.OIDCRepository
	privateKey *rsa.PrivateKey
	issuer     string
	keyID      string
}

func NewOIDCService(
	repo *repository.OIDCRepository,
	privateKey *rsa.PrivateKey,
	issuer string,
	keyID string,
) *OIDCService {
	return &OIDCService{
		repo:       repo,
		privateKey: privateKey,
		issuer:     issuer,
		keyID:       keyID,
	}
}

// ============================================================
// Generate Random String
// ============================================================

func generateRandomString(size int) (string, error) {
	b := make([]byte, size)

	if _, err := rand.Read(b); err != nil {
		return "", err
	}

	return base64.RawURLEncoding.EncodeToString(b), nil
}

// ============================================================
// Hash Secret / Code / Token
// ============================================================

func hashSHA256(value string) string {
	hash := sha256.Sum256([]byte(value))
	return hex.EncodeToString(hash[:])
}

// ============================================================
// Register OAuth Client
// ============================================================

func (s *OIDCService) RegisterClient(
	req dto.RegisterClientRequest,
) (*dto.RegisterClientResponse, error) {

	req.Name = strings.TrimSpace(req.Name)
	req.RedirectURI = strings.TrimSpace(req.RedirectURI)
	req.ClientType = strings.ToLower(strings.TrimSpace(req.ClientType))

	if req.Name == "" {
		return nil, errors.New("client name is required")
	}

	if req.RedirectURI == "" {
		return nil, errors.New("redirect uri is required")
	}

	if req.ClientType == "" {
		req.ClientType = "confidential"
	}

	if req.ClientType != "confidential" &&
		req.ClientType != "public" {

		return nil, errors.New("invalid client type")
	}

	// --------------------------------------------------------
	// Generate client ID
	// --------------------------------------------------------

	clientID, err := generateRandomString(24)
	if err != nil {
		return nil, err
	}

	// --------------------------------------------------------
	// Generate client secret
	// --------------------------------------------------------

	clientSecret := ""
	secretHash := ""

	if req.ClientType == "confidential" {

		clientSecret, err = generateRandomString(48)
		if err != nil {
			return nil, err
		}

		hash, err := bcrypt.GenerateFromPassword(
			[]byte(clientSecret),
			bcrypt.DefaultCost,
		)
		if err != nil {
			return nil, err
		}

		secretHash = string(hash)
	}

	// --------------------------------------------------------
	// Store client
	// --------------------------------------------------------

	err = s.repo.CreateClient(
		clientID,
		secretHash,
		req.Name,
		req.RedirectURI,
		req.ClientType,
	)
	if err != nil {
		return nil, err
	}

	return &dto.RegisterClientResponse{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Name:         req.Name,
		RedirectURI:  req.RedirectURI,
		ClientType:   req.ClientType,
	}, nil
}

// ============================================================
// Validate Authorization Request
// ============================================================

func (s *OIDCService) ValidateAuthorizationRequest(
	clientID string,
	redirectURI string,
) error {

	_, _, _, registeredRedirectURI, _, err :=
		s.repo.GetClient(clientID)

	if err != nil {
		return errors.New("invalid client")
	}

	if registeredRedirectURI != redirectURI {
		return errors.New("invalid redirect uri")
	}

	return nil
}

// ============================================================
// Generate Authorization Code
// ============================================================

func (s *OIDCService) GenerateAuthCode(
	clientID string,
	userID string,
	redirectURI string,
	scope string,
	nonce string,
	codeChallenge string,
) (string, error) {

	if clientID == "" {
		return "", errors.New("client id is required")
	}

	if userID == "" {
		return "", errors.New("user id is required")
	}

	if redirectURI == "" {
		return "", errors.New("redirect uri is required")
	}

	if codeChallenge == "" {
		return "", errors.New("code challenge is required")
	}

	if scope == "" {
		scope = "openid"
	}

	// Authorization code returned to browser/client.
	// Only its SHA-256 hash should be stored in DB.
	code, err := generateRandomString(48)
	if err != nil {
		return "", err
	}

	err = s.repo.SaveAuthorizationCode(
		code,
		clientID,
		userID,
		redirectURI,
		scope,
		nonce,
		codeChallenge,
	)
	if err != nil {
		return "", err
	}

	return code, nil
}

// ============================================================
// PKCE Verification
// ============================================================

func verifyPKCE(
	codeVerifier string,
	codeChallenge string,
) bool {

	if codeVerifier == "" || codeChallenge == "" {
		return false
	}

	hash := sha256.Sum256(
		[]byte(codeVerifier),
	)

	computed := base64.RawURLEncoding.EncodeToString(
		hash[:],
	)

	return computed == codeChallenge
}

// ============================================================
// Exchange Authorization Code
// ============================================================

func (s *OIDCService) ExchangeCodeForToken(
	req dto.TokenRequest,
) (*dto.TokenResponse, error) {

	// --------------------------------------------------------
	// 1. Validate required fields
	// --------------------------------------------------------

	if req.ClientID == "" {
		return nil, errors.New("client id is required")
	}

	if req.Code == "" {
		return nil, errors.New("authorization code is required")
	}

	if req.RedirectURI == "" {
		return nil, errors.New("redirect uri is required")
	}

	if req.CodeVerifier == "" {
		return nil, errors.New("code verifier is required")
	}

	// --------------------------------------------------------
	// 2. Get client
	// --------------------------------------------------------

	_, secretHash, _, registeredRedirectURI, clientType, err :=
		s.repo.GetClient(req.ClientID)

	if err != nil {
		return nil, errors.New("invalid client")
	}

	// --------------------------------------------------------
	// 3. Validate redirect URI
	// --------------------------------------------------------

	if registeredRedirectURI != req.RedirectURI {
		return nil, errors.New("redirect uri mismatch")
	}

	// --------------------------------------------------------
	// 4. Validate client secret
	// --------------------------------------------------------

	if clientType == "confidential" {

		if secretHash == "" {
			return nil, errors.New("client secret not configured")
		}

		if req.ClientSecret == "" {
			return nil, errors.New("client secret is required")
		}

		err := bcrypt.CompareHashAndPassword(
			[]byte(secretHash),
			[]byte(req.ClientSecret),
		)

		if err != nil {
			return nil, errors.New("invalid client credentials")
		}
	}

	// --------------------------------------------------------
	// 5. Consume authorization code
	// --------------------------------------------------------
	//
	// Repository should:
	// - hash req.Code
	// - find matching code_hash
	// - verify client_id
	// - verify redirect_uri
	// - verify expiration
	// - verify used_at IS NULL
	// - mark used_at
	//
	// This makes the code single-use.
	//

	authCode, err := s.repo.ConsumeAuthorizationCode(
		req.Code,
		req.ClientID,
		req.RedirectURI,
	)
	if err != nil {
		return nil, err
	}

	// --------------------------------------------------------
	// 6. Verify PKCE
	// --------------------------------------------------------

	if !verifyPKCE(
		req.CodeVerifier,
		authCode.CodeChallenge,
	) {
		return nil, errors.New("invalid code verifier")
	}

	// --------------------------------------------------------
	// 7. Generate access token
	// --------------------------------------------------------

	accessToken, err := generateRandomString(48)
	if err != nil {
		return nil, err
	}

	tokenHash := hashSHA256(accessToken)

	accessTokenExpires := time.Now().Add(time.Hour)

	err = s.repo.SaveAccessToken(
		tokenHash,
		req.ClientID,
		authCode.UserID,
		authCode.Scope,
		accessTokenExpires,
	)
	if err != nil {
		return nil, err
	}

	// --------------------------------------------------------
	// 8. Get user's primary handle
	// --------------------------------------------------------

	userHandle, err :=
		s.repo.GetPrimaryHandleByUserID(
			authCode.UserID,
		)

	if err != nil {
		return nil, err
	}

	// --------------------------------------------------------
	// 9. Generate OIDC ID token
	// --------------------------------------------------------

	idToken, err := s.GenerateIDToken(
		authCode.UserID,
		userHandle,
		req.ClientID,
		authCode.Nonce,
	)
	if err != nil {
		return nil, err
	}

	// --------------------------------------------------------
	// 10. Return OAuth/OIDC tokens
	// --------------------------------------------------------

	return &dto.TokenResponse{
		AccessToken: accessToken,
		TokenType:   "Bearer",
		ExpiresIn:   3600,
		IDToken:     idToken,
		Scope:       authCode.Scope,
	}, nil
}

// ============================================================
// Generate OIDC ID Token
// ============================================================

func (s *OIDCService) GenerateIDToken(
	userID string,
	handle string,
	clientID string,
	nonce string,
) (string, error) {

	if s.privateKey == nil {
		return "", errors.New("oidc signing key is not configured")
	}

	now := time.Now()

	claims := jwt.MapClaims{
		"iss":                s.issuer,
		"sub":                userID,
		"aud":                clientID,
		"iat":                now.Unix(),
		"exp":                now.Add(time.Hour).Unix(),
		"preferred_username": handle,
	}

	// OIDC requires nonce to be returned if it was
	// included in the authorization request.
	if nonce != "" {
		claims["nonce"] = nonce
	}

	token := jwt.NewWithClaims(
		jwt.SigningMethodRS256,
		claims,
	)

	token.Header["kid"] = s.keyID

	return token.SignedString(
		s.privateKey,
	)
}

// ============================================================
// UserInfo
// ============================================================

func (s *OIDCService) UserInfo(
	accessToken string,
) (*dto.UserInfoResponse, error) {

	if accessToken == "" {
		return nil, errors.New("access token is required")
	}

	tokenHash := hashSHA256(accessToken)

	userID, err :=
		s.repo.GetUserByAccessToken(tokenHash)

	if err != nil {
		return nil, errors.New("invalid access token")
	}

	handle, err :=
		s.repo.GetPrimaryHandleByUserID(userID)

	if err != nil {
		return nil, err
	}

	return &dto.UserInfoResponse{
		Subject:           userID,
		Name:              handle,
		PreferredUsername: handle,
	}, nil
}

// ============================================================
// JWKS
// ============================================================

func (s *OIDCService) JWKS() (map[string]interface{}, error) {

	if s.privateKey == nil {
		return nil, errors.New("oidc signing key is not configured")
	}

	publicKey := &s.privateKey.PublicKey

	// --------------------------------------------------------
	// RSA modulus
	// --------------------------------------------------------

	n := base64.RawURLEncoding.EncodeToString(
		publicKey.N.Bytes(),
	)

	// --------------------------------------------------------
	// RSA exponent
	//
	// Normally this is 65537.
	// Encode it as an unsigned big-endian integer.
	// --------------------------------------------------------

	e := uint64(publicKey.E)

	var exponentBytes []byte

	for e > 0 {
		exponentBytes = append(
			[]byte{byte(e & 0xff)},
			exponentBytes...,
		)

		e >>= 8
	}

	// --------------------------------------------------------
	// JWKS response
	// --------------------------------------------------------

	return map[string]interface{}{
		"keys": []map[string]interface{}{
			{
				"kty": "RSA",
				"use": "sig",
				"alg": "RS256",
				"kid": s.keyID,
				"n":   n,
				"e":   base64.RawURLEncoding.EncodeToString(exponentBytes),
			},
		},
	}, nil
}
