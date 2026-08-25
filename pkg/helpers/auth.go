package helpers

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const defaultSecret = "nid-secret-key"

// ============================================================
// JWT SECRET
// ============================================================

func getJWTSecret() []byte {
	secret := strings.TrimSpace(os.Getenv("JWT_SECRET"))

	if secret == "" {
		secret = defaultSecret
	}

	return []byte(secret)
}

// ============================================================
// 1. EXISTING HMAC TOKEN
// ============================================================

func GenerateToken(userID string) string {
	userID = strings.TrimSpace(userID)

	mac := hmac.New(
		sha256.New,
		getJWTSecret(),
	)

	_, _ = mac.Write([]byte(userID))

	return userID + "." + hex.EncodeToString(mac.Sum(nil))
}

// ============================================================
// 2. EXISTING HMAC TOKEN VALIDATION
// ============================================================

func ValidateToken(token string) (string, bool) {
	token = strings.TrimSpace(token)

	parts := strings.Split(token, ".")

	if len(parts) != 2 {
		return "", false
	}

	userID := strings.TrimSpace(parts[0])

	if userID == "" {
		return "", false
	}

	expected := GenerateToken(userID)

	if !hmac.Equal(
		[]byte(token),
		[]byte(expected),
	) {
		return "", false
	}

	return userID, true
}

// ============================================================
// 3. OIDC ID TOKEN
// ============================================================

func GenerateIDToken(
	userID string,
	handle string,
	clientID string,
) (string, error) {

	userID = strings.TrimSpace(userID)
	handle = strings.TrimSpace(handle)
	clientID = strings.TrimSpace(clientID)

	if userID == "" {
		return "", errors.New("user id is required")
	}

	if clientID == "" {
		return "", errors.New("client id is required")
	}

	now := time.Now()

	claims := jwt.MapClaims{
		"iss":    "https://nid.xyz",
		"sub":    userID,
		"aud":    clientID,
		"handle": handle,
		"iat":    now.Unix(),
		"exp":    now.Add(1 * time.Hour).Unix(),
	}

	token := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		claims,
	)

	return token.SignedString(getJWTSecret())
}

// ============================================================
// 4. GET USER ID FROM REQUEST
// ============================================================

func GetUserIDFromRequest(
	r *http.Request,
) (string, error) {

	cookie, err := r.Cookie("nid_token")

	if err != nil {
		return "", errors.New(
			"authentication token cookie missing",
		)
	}

	token := strings.TrimSpace(cookie.Value)

	if token == "" {
		return "", errors.New(
			"authentication token is empty",
		)
	}

	userID, valid := ValidateToken(token)

	if !valid {
		return "", errors.New(
			"invalid or expired session token",
		)
	}

	return userID, nil
}

// ============================================================
// 5. VALIDATE JWT TOKEN
// ============================================================
//
// This is ONLY for JWT tokens.
//
// Example:
//
// app_session=<JWT>
//
// It validates:
//   - JWT format
//   - HS256 algorithm
//   - signature
//   - expiration
//   - subject
//
// ============================================================

func ValidateJWTToken(
	tokenString string,
) (string, bool) {

	tokenString = strings.TrimSpace(tokenString)

	if tokenString == "" {
		return "", false
	}

	// ------------------------------------------------------------
	// Parse JWT
	// ------------------------------------------------------------

	token, err := jwt.Parse(
		tokenString,
		func(token *jwt.Token) (interface{}, error) {

			// Only allow HMAC algorithms.
			// Specifically HS256 for your current implementation.

			if token.Method != jwt.SigningMethodHS256 {
				return nil, fmt.Errorf(
					"unexpected signing method: %s",
					token.Method.Alg(),
				)
			}

			return getJWTSecret(), nil
		},
	)

	if err != nil {
		return "", false
	}

	if token == nil || !token.Valid {
		return "", false
	}

	// ------------------------------------------------------------
	// Claims
	// ------------------------------------------------------------

	claims, ok := token.Claims.(jwt.MapClaims)

	if !ok {
		return "", false
	}

	// ------------------------------------------------------------
	// Subject
	// ------------------------------------------------------------

	sub, ok := claims["sub"].(string)

	if !ok {
		return "", false
	}

	sub = strings.TrimSpace(sub)

	if sub == "" {
		return "", false
	}

	return sub, true
}
