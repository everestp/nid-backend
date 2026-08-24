// pkg/helpers/auth.go
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

// getJWTSecret retrieves your secret key from environment variables with a fallback
func getJWTSecret() []byte {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = defaultSecret
	}
	return []byte(secret)
}

// 1. GenerateToken creates a lightweight HMAC session token for your internal frontend
func GenerateToken(userID string) string {
	mac := hmac.New(sha256.New, []byte(defaultSecret))
	mac.Write([]byte(userID))
	return userID + "." + hex.EncodeToString(mac.Sum(nil))
}

// 2. ValidateToken verifies your internal HMAC session token
func ValidateToken(token string) (string, bool) {
	parts := strings.Split(token, ".")
	if len(parts) != 2 {
		return "", false
	}
	userID := parts[0]
	expected := GenerateToken(userID)
	return userID, token == expected
}

// 3. GenerateIDToken generates an OIDC-compliant JWT ID Token for external apps
// containing standard OIDC claims (issuer, subject, audience, handle, expiration).
func GenerateIDToken(userID, handle, clientID string) (string, error) {
	claims := jwt.MapClaims{
		"iss":    "https://nid.xyz",                            // Your Identity Provider domain name
		"sub":    userID,                                       // Unique user identifier
		"aud":    clientID,                                     // The third-party app client ID requesting login
		"handle": handle,                                       // The user's primary .nid handle
		"iat":    time.Now().Unix(),
		"exp":    time.Now().Add(1 * time.Hour).Unix(),         // OIDC ID tokens are typically short-lived
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(getJWTSecret())
}

// 4. GetUserIDFromRequest extracts and validates the Bearer token from incoming HTTP requests
// (Used by OIDC Authorize/Token controllers to recognize logged-in users safely)
func GetUserIDFromRequest(r *http.Request) (string, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return "", errors.New("authorization header missing")
	}

	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		return "", errors.New("invalid authorization format, expected 'Bearer <token>'")
	}

	token := parts[1]
	userID, valid := ValidateToken(token)
	if !valid {
		return "", errors.New("invalid or expired session token")
	}

	return userID, nil
}
