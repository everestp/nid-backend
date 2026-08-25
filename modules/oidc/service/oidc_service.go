package service

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"nid-backend/modules/oidc/dto"
	"nid-backend/modules/oidc/repository"
)

// ============================================================
// OIDC Service
// ============================================================

type OIDCService struct {
	repo       *repository.OIDCRepository
	privateKey *rsa.PrivateKey
	issuer     string
	keyID      string
}

// ============================================================
// Constructor
// ============================================================

func NewOIDCService(
	repo *repository.OIDCRepository,
	privateKey *rsa.PrivateKey,
	issuer string,
	keyID string,
) *OIDCService {
	return &OIDCService{
		repo:       repo,
		privateKey: privateKey,
		issuer:     strings.TrimRight(issuer, "/"),
		keyID:      keyID,
	}
}

// ============================================================
// Random String
// ============================================================

func generateRandomString(size int) (string, error) {
	if size <= 0 {
		return "", errors.New("invalid random string size")
	}

	b := make([]byte, size)

	if _, err := rand.Read(b); err != nil {
		return "", err
	}

	return base64.RawURLEncoding.EncodeToString(b), nil
}

// ============================================================
// SHA-256 Hash
// ============================================================

func hashSHA256(value string) string {
	hash := sha256.Sum256([]byte(value))
	return hex.EncodeToString(hash[:])
}

// ============================================================
// PKCE Verification
// ============================================================
//
// code_challenge =
// BASE64URL(SHA256(code_verifier))
//
// Only S256 is supported.
//

func verifyPKCE(
	codeVerifier string,
	codeChallenge string,
	codeChallengeMethod string,
) bool {

	if codeVerifier == "" {
		return false
	}

	if codeChallenge == "" {
		return false
	}

	if codeChallengeMethod != "S256" {
		return false
	}

	hash := sha256.Sum256([]byte(codeVerifier))

	computed := base64.RawURLEncoding.EncodeToString(
		hash[:],
	)

	return subtle.ConstantTimeCompare(
		[]byte(computed),
		[]byte(codeChallenge),
	) == 1
}

// ============================================================
// Register OAuth Client
// ============================================================

func (s *OIDCService) RegisterClient(
	req dto.RegisterClientRequest,
) (*dto.RegisterClientResponse, error) {

	// --------------------------------------------------------
	// Normalize
	// --------------------------------------------------------

	req.Name = strings.TrimSpace(req.Name)

	req.RedirectURI = strings.TrimSpace(
		req.RedirectURI,
	)

	req.ClientType = strings.ToLower(
		strings.TrimSpace(req.ClientType),
	)

	// --------------------------------------------------------
	// Validate name
	// --------------------------------------------------------

	if req.Name == "" {
		return nil, errors.New(
			"client name is required",
		)
	}

	// --------------------------------------------------------
	// Validate redirect URI
	// --------------------------------------------------------

	if req.RedirectURI == "" {
		return nil, errors.New(
			"redirect_uri is required",
		)
	}

	// --------------------------------------------------------
	// Default client type
	// --------------------------------------------------------

	if req.ClientType == "" {
		req.ClientType = "confidential"
	}

	// --------------------------------------------------------
	// Validate client type
	// --------------------------------------------------------

	if req.ClientType != "confidential" &&
		req.ClientType != "public" {

		return nil, errors.New(
			"invalid client_type",
		)
	}

	// --------------------------------------------------------
	// Generate client ID
	// --------------------------------------------------------

	clientID, err := generateRandomString(24)
	if err != nil {
		return nil, err
	}

	// --------------------------------------------------------
	// Client secret
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

	// --------------------------------------------------------
	// Response
	// --------------------------------------------------------

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
	responseType string,
	scope string,
	codeChallenge string,
	codeChallengeMethod string,
	nonce string,
) error {

	// --------------------------------------------------------
	// client_id
	// --------------------------------------------------------

	clientID = strings.TrimSpace(clientID)

	if clientID == "" {
		return errors.New(
			"client_id is required",
		)
	}

	// --------------------------------------------------------
	// redirect_uri
	// --------------------------------------------------------

	redirectURI = strings.TrimSpace(redirectURI)

	if redirectURI == "" {
		return errors.New(
			"redirect_uri is required",
		)
	}

	// --------------------------------------------------------
	// response_type
	// --------------------------------------------------------

	responseType = strings.TrimSpace(responseType)

	if responseType != "code" {
		return errors.New(
			"only response_type=code is supported",
		)
	}

	// --------------------------------------------------------
	// scope
	// --------------------------------------------------------

	scope = strings.TrimSpace(scope)

	if scope == "" {
		scope = "openid"
	}

	if !containsScope(scope, "openid") {
		return errors.New(
			"openid scope is required",
		)
	}

	// --------------------------------------------------------
	// PKCE code_challenge
	// --------------------------------------------------------

	codeChallenge = strings.TrimSpace(
		codeChallenge,
	)

	if codeChallenge == "" {
		return errors.New(
			"code_challenge is required",
		)
	}

	// --------------------------------------------------------
	// PKCE code_challenge_method
	// --------------------------------------------------------

	codeChallengeMethod = strings.TrimSpace(
		codeChallengeMethod,
	)

	if codeChallengeMethod == "" {
		codeChallengeMethod = "S256"
	}

	if codeChallengeMethod != "S256" {
		return errors.New(
			"only S256 PKCE is supported",
		)
	}

	// --------------------------------------------------------
	// nonce
	// --------------------------------------------------------

	nonce = strings.TrimSpace(nonce)

	if nonce == "" {
		return errors.New(
			"nonce is required",
		)
	}

	// --------------------------------------------------------
	// Get registered client
	// --------------------------------------------------------

	client, err := s.repo.GetClient(clientID)
	if err != nil {
		return errors.New(
			"invalid client",
		)
	}

	// --------------------------------------------------------
	// Exact redirect URI matching
	// --------------------------------------------------------

	if client.RedirectURI != redirectURI {
		return errors.New(
			"invalid redirect_uri",
		)
	}

	return nil
}

// ============================================================
// Scope Helper
// ============================================================

func containsScope(
	scope string,
	required string,
) bool {

	for _, item := range strings.Fields(scope) {
		if item == required {
			return true
		}
	}

	return false
}

// ============================================================
// Generate Authorization Code
// ============================================================
//
// Stores:
//
// - client_id
// - user_id
// - redirect_uri
// - scope
// - nonce
// - code_challenge
// - code_challenge_method
//
// The raw authorization code is NEVER stored.
//

func (s *OIDCService) GenerateAuthCode(
	clientID string,
	userID string,
	redirectURI string,
	scope string,
	nonce string,
	codeChallenge string,
	codeChallengeMethod string,
) (string, error) {

	// --------------------------------------------------------
	// client_id
	// --------------------------------------------------------

	clientID = strings.TrimSpace(clientID)

	if clientID == "" {
		return "", errors.New(
			"client_id is required",
		)
	}

	// --------------------------------------------------------
	// user_id
	// --------------------------------------------------------

	userID = strings.TrimSpace(userID)

	if userID == "" {
		return "", errors.New(
			"user_id is required",
		)
	}

	// --------------------------------------------------------
	// redirect_uri
	// --------------------------------------------------------

	redirectURI = strings.TrimSpace(
		redirectURI,
	)

	if redirectURI == "" {
		return "", errors.New(
			"redirect_uri is required",
		)
	}

	// --------------------------------------------------------
	// scope
	// --------------------------------------------------------

	scope = strings.TrimSpace(scope)

	if scope == "" {
		scope = "openid"
	}

	if !containsScope(scope, "openid") {
		return "", errors.New(
			"openid scope is required",
		)
	}

	// --------------------------------------------------------
	// nonce
	// --------------------------------------------------------

	nonce = strings.TrimSpace(nonce)

	if nonce == "" {
		return "", errors.New(
			"nonce is required",
		)
	}

	// --------------------------------------------------------
	// code_challenge
	// --------------------------------------------------------

	codeChallenge = strings.TrimSpace(
		codeChallenge,
	)

	if codeChallenge == "" {
		return "", errors.New(
			"code_challenge is required",
		)
	}

	// --------------------------------------------------------
	// code_challenge_method
	// --------------------------------------------------------

	codeChallengeMethod = strings.TrimSpace(
		codeChallengeMethod,
	)

	if codeChallengeMethod == "" {
		codeChallengeMethod = "S256"
	}

	if codeChallengeMethod != "S256" {
		return "", errors.New(
			"only S256 PKCE is supported",
		)
	}

	// --------------------------------------------------------
	// Validate client + redirect URI again
	// --------------------------------------------------------

	client, err := s.repo.GetClient(clientID)
	if err != nil {
		return "", errors.New(
			"invalid client",
		)
	}

	if client.RedirectURI != redirectURI {
		return "", errors.New(
			"invalid redirect_uri",
		)
	}

	// --------------------------------------------------------
	// Generate authorization code
	// --------------------------------------------------------

	code, err := generateRandomString(48)
	if err != nil {
		return "", err
	}

	// --------------------------------------------------------
	// Store authorization code
	// --------------------------------------------------------

	err = s.repo.SaveAuthorizationCode(
		code,
		clientID,
		userID,
		redirectURI,
		scope,
		nonce,
		codeChallenge,
		codeChallengeMethod,
	)
	if err != nil {
		return "", err
	}

	return code, nil
}

// ============================================================
// Exchange Authorization Code
// ============================================================

func (s *OIDCService) ExchangeCodeForToken(
	req dto.TokenRequest,
) (*dto.TokenResponse, error) {

	// --------------------------------------------------------
	// grant_type
	// --------------------------------------------------------

	req.GrantType = strings.TrimSpace(
		req.GrantType,
	)

	if req.GrantType != "authorization_code" {
		return nil, errors.New(
			"unsupported grant_type",
		)
	}

	// --------------------------------------------------------
	// client_id
	// --------------------------------------------------------

	req.ClientID = strings.TrimSpace(
		req.ClientID,
	)

	if req.ClientID == "" {
		return nil, errors.New(
			"client_id is required",
		)
	}

	// --------------------------------------------------------
	// code
	// --------------------------------------------------------

	req.Code = strings.TrimSpace(
		req.Code,
	)

	if req.Code == "" {
		return nil, errors.New(
			"code is required",
		)
	}

	// --------------------------------------------------------
	// redirect_uri
	// --------------------------------------------------------

	req.RedirectURI = strings.TrimSpace(
		req.RedirectURI,
	)

	if req.RedirectURI == "" {
		return nil, errors.New(
			"redirect_uri is required",
		)
	}

	// --------------------------------------------------------
	// code_verifier
	// --------------------------------------------------------

	req.CodeVerifier = strings.TrimSpace(
		req.CodeVerifier,
	)

	if req.CodeVerifier == "" {
		return nil, errors.New(
			"code_verifier is required",
		)
	}

	// --------------------------------------------------------
	// Get client
	// --------------------------------------------------------

	client, err := s.repo.GetClient(
		req.ClientID,
	)
	if err != nil {
		return nil, errors.New(
			"invalid client",
		)
	}

	// --------------------------------------------------------
	// Validate redirect URI
	// --------------------------------------------------------

	if client.RedirectURI != req.RedirectURI {
		return nil, errors.New(
			"redirect_uri mismatch",
		)
	}

	// --------------------------------------------------------
	// Confidential client authentication
	// --------------------------------------------------------

	if client.ClientType == "confidential" {

		req.ClientSecret = strings.TrimSpace(
			req.ClientSecret,
		)

		if req.ClientSecret == "" {
			return nil, errors.New(
				"client_secret is required",
			)
		}

		if client.ClientSecretHash == "" {
			return nil, errors.New(
				"client secret not configured",
			)
		}

		if err := bcrypt.CompareHashAndPassword(
			[]byte(client.ClientSecretHash),
			[]byte(req.ClientSecret),
		); err != nil {
			return nil, errors.New(
				"invalid client credentials",
			)
		}
	}

	// --------------------------------------------------------
	// Consume authorization code
	// --------------------------------------------------------

	authCode, err :=
		s.repo.ConsumeAuthorizationCode(
			req.Code,
			req.ClientID,
			req.RedirectURI,
		)

	if err != nil {
		return nil, err
	}

	// --------------------------------------------------------
	// Validate PKCE method
	// --------------------------------------------------------

	if authCode.CodeChallengeMethod != "S256" {
		return nil, errors.New(
			"unsupported code_challenge_method",
		)
	}

	// --------------------------------------------------------
	// Verify PKCE
	// --------------------------------------------------------

	if !verifyPKCE(
		req.CodeVerifier,
		authCode.CodeChallenge,
		authCode.CodeChallengeMethod,
	) {
		return nil, errors.New(
			"invalid code_verifier",
		)
	}

	// --------------------------------------------------------
	// Generate access token
	// --------------------------------------------------------

	accessToken, err :=
		generateRandomString(48)

	if err != nil {
		return nil, err
	}

	// --------------------------------------------------------
	// Hash access token
	// --------------------------------------------------------

	tokenHash := hashSHA256(
		accessToken,
	)

	accessTokenExpires :=
		time.Now().Add(time.Hour)

	// --------------------------------------------------------
	// Store access token
	// --------------------------------------------------------

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
	// Get user's NID handle
	// --------------------------------------------------------

	handle, err :=
		s.repo.GetPrimaryHandleByUserID(
			authCode.UserID,
		)

	if err != nil {
		return nil, err
	}

	// --------------------------------------------------------
	// Generate OIDC ID token
	// --------------------------------------------------------

	idToken, err :=
		s.GenerateIDToken(
			authCode.UserID,
			handle,
			req.ClientID,
			authCode.Nonce,
		)

	if err != nil {
		return nil, err
	}

	// --------------------------------------------------------
	// Token response
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

	// --------------------------------------------------------
	// Validate signing key
	// --------------------------------------------------------

	if s.privateKey == nil {
		return "", errors.New(
			"oidc signing key is not configured",
		)
	}

	// --------------------------------------------------------
	// Validate user
	// --------------------------------------------------------

	userID = strings.TrimSpace(userID)

	if userID == "" {
		return "", errors.New(
			"user_id is required",
		)
	}

	// --------------------------------------------------------
	// Validate client
	// --------------------------------------------------------

	clientID = strings.TrimSpace(clientID)

	if clientID == "" {
		return "", errors.New(
			"client_id is required",
		)
	}

	// --------------------------------------------------------
	// Validate nonce
	// --------------------------------------------------------

	nonce = strings.TrimSpace(nonce)

	if nonce == "" {
		return "", errors.New(
			"nonce is required",
		)
	}

	// --------------------------------------------------------
	// Token timestamps
	// --------------------------------------------------------

	now := time.Now()

	expiresAt :=
		now.Add(time.Hour)

	// --------------------------------------------------------
	// OIDC claims
	// --------------------------------------------------------

	claims := jwt.MapClaims{
		"iss": s.issuer,

		"sub": userID,

		"aud": clientID,

		"iat": now.Unix(),

		"exp": expiresAt.Unix(),

		"preferred_username": handle,

		"nonce": nonce,
	}

	// --------------------------------------------------------
	// Create JWT
	// --------------------------------------------------------

	token := jwt.NewWithClaims(
		jwt.SigningMethodRS256,
		claims,
	)

	// --------------------------------------------------------
	// Key ID
	// --------------------------------------------------------

	token.Header["kid"] = s.keyID

	// --------------------------------------------------------
	// Sign JWT
	// --------------------------------------------------------

	signedToken, err :=
		token.SignedString(
			s.privateKey,
		)

	if err != nil {
		return "", err
	}

	return signedToken, nil
}

// ============================================================
// UserInfo
// ============================================================

func (s *OIDCService) UserInfo(
	accessToken string,
) (*dto.UserInfoResponse, error) {

	// --------------------------------------------------------
	// Validate token
	// --------------------------------------------------------

	accessToken = strings.TrimSpace(
		accessToken,
	)

	if accessToken == "" {
		return nil, errors.New(
			"access_token is required",
		)
	}

	// --------------------------------------------------------
	// Hash token
	// --------------------------------------------------------

	tokenHash := hashSHA256(
		accessToken,
	)

	// --------------------------------------------------------
	// Lookup token
	// --------------------------------------------------------

	userID, err :=
		s.repo.GetUserByAccessToken(
			tokenHash,
		)

	if err != nil {
		return nil, errors.New(
			"invalid access token",
		)
	}

	// --------------------------------------------------------
	// Get handle
	// --------------------------------------------------------

	handle, err :=
		s.repo.GetPrimaryHandleByUserID(
			userID,
		)

	if err != nil {
		return nil, err
	}

	// --------------------------------------------------------
	// UserInfo response
	// --------------------------------------------------------

	return &dto.UserInfoResponse{
		Subject:           userID,
		Name:              handle,
		PreferredUsername: handle,
	}, nil
}

// ============================================================
// JWKS
// ============================================================

func (s *OIDCService) JWKS() (
	map[string]interface{},
	error,
) {

	// --------------------------------------------------------
	// Validate signing key
	// --------------------------------------------------------

	if s.privateKey == nil {
		return nil, errors.New(
			"oidc signing key is not configured",
		)
	}

	// --------------------------------------------------------
	// Public RSA key
	// --------------------------------------------------------

	publicKey := &s.privateKey.PublicKey

	// --------------------------------------------------------
	// RSA modulus
	// --------------------------------------------------------

	n := base64.RawURLEncoding.EncodeToString(
		publicKey.N.Bytes(),
	)

	// --------------------------------------------------------
	// RSA exponent
	// --------------------------------------------------------

	e := uint64(publicKey.E)

	if e == 0 {
		return nil, errors.New(
			"invalid RSA public exponent",
		)
	}

	var exponentBytes []byte

	for e > 0 {

		exponentBytes = append(
			[]byte{
				byte(e & 0xff),
			},
			exponentBytes...,
		)

		e >>= 8
	}

	// --------------------------------------------------------
	// JWKS
	// --------------------------------------------------------

	return map[string]interface{}{
		"keys": []map[string]interface{}{
			{
				"kty": "RSA",
				"use": "sig",
				"alg": "RS256",
				"kid": s.keyID,
				"n":   n,
				"e": base64.RawURLEncoding.EncodeToString(
					exponentBytes,
				),
			},
		},
	}, nil
}
