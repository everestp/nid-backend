// pkg/helpers/auth.go
package helpers

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"strings"
)

func GenerateToken(userID string) string {
	secret := "nid-secret-key"
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(userID))
	return userID + "." + hex.EncodeToString(mac.Sum(nil))
}

func ValidateToken(token string) (string, bool) {
	parts := strings.Split(token, ".")
	if len(parts) != 2 {
		return "", false
	}
	userID := parts[0]
	expected := GenerateToken(userID)
	return userID, token == expected
}
