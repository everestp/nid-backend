package helpers

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const defaultSecret = "nid-secret-key"

// ============================================================
// INTERNAL NID SESSION AUTH
// ============================================================

func getInternalAuthSecret() []byte {
	secret := strings.TrimSpace(os.Getenv("NID_INTERNAL_AUTH_SECRET"))

	if secret == "" {
		secret = defaultSecret
	}

	return []byte(secret)
}

// ------------------------------------------------------------
// Generate Internal Session Token
// ------------------------------------------------------------
//
// This token is ONLY for NID's own frontend/backend.
//
// Example:
//
// nid_token = userID.signature
//
// It is NOT an OIDC token.
// It is NOT sent to external applications.
//

func GenerateInternalSessionToken(userID string) string {
	userID = strings.TrimSpace(userID)

	mac := hmac.New(
		sha256.New,
		getInternalAuthSecret(),
	)

	_, _ = mac.Write([]byte(userID))

	signature := hex.EncodeToString(mac.Sum(nil))

	return userID + "." + signature
}

// ------------------------------------------------------------
// Validate Internal Session Token
// ------------------------------------------------------------

func ValidateInternalSessionToken(
	token string,
) (string, bool) {

	token = strings.TrimSpace(token)

	parts := strings.Split(token, ".")

	if len(parts) != 2 {
		return "", false
	}

	userID := strings.TrimSpace(parts[0])

	if userID == "" {
		return "", false
	}

	expected := GenerateInternalSessionToken(userID)

	if !hmac.Equal(
		[]byte(token),
		[]byte(expected),
	) {
		return "", false
	}

	return userID, true
}

// ------------------------------------------------------------
// Get NID Internal User From Cookie
// ------------------------------------------------------------

func GetInternalUserIDFromRequest(
	r *http.Request,
) (string, error) {

	cookie, err := r.Cookie("nid_token")

	if err != nil {
		return "", errors.New(
			"nid session cookie missing",
		)
	}

	token := strings.TrimSpace(cookie.Value)

	if token == "" {
		return "", errors.New(
			"nid session token is empty",
		)
	}

	userID, valid := ValidateInternalSessionToken(token)

	if !valid {
		return "", errors.New(
			"invalid nid session",
		)
	}

	return userID, nil
}

// ============================================================
// OIDC AUTH
// ============================================================

// IMPORTANT:
//
// These functions are for EXTERNAL applications.
//
// Do NOT use the internal NID session token here.
//
// OIDC tokens should be signed with NID's OIDC signing key.
//
// Example:
//
// NID
//   |
//   ├── Authorization Code
//   |
//   ├── Access Token
//   |
//   └── ID Token
//
// External app consumes these tokens.
//

// ------------------------------------------------------------
// OIDC ID Token Claims
// ------------------------------------------------------------

type OIDCClaims struct {
	Handle string `json:"handle,omitempty"`

	jwt.RegisteredClaims
}

// ------------------------------------------------------------
// Generate OIDC ID Token
// ------------------------------------------------------------
//
// NOTE:
// This version uses HS256.
//
// For a real OIDC provider, I recommend RS256/ES256
// with your OIDC private key and JWKS endpoint.
//

func GenerateOIDCIDToken(
	userID string,
	handle string,
	clientID string,
) (string, error) {

	userID = strings.TrimSpace(userID)
	handle = strings.TrimSpace(handle)
	clientID = strings.TrimSpace(clientID)

	if userID == "" {
		return "", errors.New(
			"user id is required",
		)
	}

	if clientID == "" {
		return "", errors.New(
			"client id is required",
		)
	}

	now := time.Now()

	claims := OIDCClaims{
		Handle: handle,

		RegisteredClaims: jwt.RegisteredClaims{
			Issuer: "https://nid.xyz",

			Subject: userID,

			Audience: jwt.ClaimStrings{
				clientID,
			},

			IssuedAt: jwt.NewNumericDate(now),

			ExpiresAt: jwt.NewNumericDate(
				now.Add(1 * time.Hour),
			),
		},
	}

	token := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		claims,
	)

	return token.SignedString(
		getInternalAuthSecret(),
	)
}

// ------------------------------------------------------------
// Validate OIDC ID Token
// ------------------------------------------------------------
//
// This is for OIDC token validation.
// External applications should normally validate
// NID's ID token using NID's JWKS endpoint.
//

func ValidateOIDCIDToken(
	tokenString string,
	clientID string,
) (string, bool) {

	tokenString = strings.TrimSpace(tokenString)

	if tokenString == "" {
		return "", false
	}

	clientID = strings.TrimSpace(clientID)

	if clientID == "" {
		return "", false
	}

	token, err := jwt.ParseWithClaims(
		tokenString,
		&OIDCClaims{},
		func(token *jwt.Token) (interface{}, error) {

			if token.Method != jwt.SigningMethodHS256 {
				return nil, errors.New(
					"unexpected oidc signing method",
				)
			}

			return getInternalAuthSecret(), nil
		},
		jwt.WithIssuer("https://nid.xyz"),
		jwt.WithAudience(clientID),
	)

	if err != nil || token == nil || !token.Valid {
		return "", false
	}

	claims, ok := token.Claims.(*OIDCClaims)

	if !ok {
		return "", false
	}

	sub := strings.TrimSpace(claims.Subject)

	if sub == "" {
		return "", false
	}

	return sub, true
}
