// pkg/helpers/oidc_key.go

package helpers

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"os"
)

func LoadOIDCPrivateKey() (*rsa.PrivateKey, error) {

	// ============================================================
	// Load private key file path from environment
	// ============================================================

	keyPath := os.Getenv("NID_OIDC_PRIVATE_KEY_FILE")

	if keyPath == "" {
		return nil, errors.New(
			"NID_OIDC_PRIVATE_KEY_FILE is not configured",
		)
	}

	// ============================================================
	// Read PEM file
	// ============================================================

	privateKeyPEM, err := os.ReadFile(keyPath)
	if err != nil {
		return nil, errors.New(
			"failed to read OIDC private key: " + err.Error(),
		)
	}

	// ============================================================
	// Decode PEM
	// ============================================================

	block, _ := pem.Decode(privateKeyPEM)

	if block == nil {
		return nil, errors.New(
			"invalid OIDC private key PEM",
		)
	}

	// ============================================================
	// PKCS#8
	//
	// -----BEGIN PRIVATE KEY-----
	// ============================================================

	if block.Type == "PRIVATE KEY" {

		key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
		if err != nil {
			return nil, errors.New(
				"invalid PKCS#8 OIDC private key",
			)
		}

		rsaKey, ok := key.(*rsa.PrivateKey)
		if !ok {
			return nil, errors.New(
				"OIDC private key is not RSA",
			)
		}

		return rsaKey, nil
	}

	// ============================================================
	// PKCS#1
	//
	// -----BEGIN RSA PRIVATE KEY-----
	// ============================================================

	if block.Type == "RSA PRIVATE KEY" {

		key, err := x509.ParsePKCS1PrivateKey(block.Bytes)
		if err != nil {
			return nil, errors.New(
				"invalid PKCS#1 OIDC private key",
			)
		}

		return key, nil
	}

	// ============================================================
	// Unsupported PEM type
	// ============================================================

	return nil, errors.New(
		"unsupported OIDC private key type: " + block.Type,
	)
}
