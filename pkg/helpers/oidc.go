// pkg/helpers/oidc_key.go

package helpers

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"log"
	"os"
	"strings"
)

func LoadOIDCPrivateKey() (*rsa.PrivateKey, error) {

	privateKeyPEM := strings.TrimSpace(
		os.Getenv("NID_OIDC_PRIVATE_KEY"),
	)

	// ============================================================
	// Development fallback
	// ============================================================

	if privateKeyPEM == "" {

		log.Println(
			"WARNING: NID_OIDC_PRIVATE_KEY is not configured",
		)

		log.Println(
			"Generating temporary RSA key for development",
		)

		return rsa.GenerateKey(
			rand.Reader,
			2048,
		)
	}

	// ============================================================
	// Decode PEM
	// ============================================================

	block, _ := pem.Decode(
		[]byte(privateKeyPEM),
	)

	if block == nil {
		return nil, errors.New(
			"invalid OIDC RSA private key PEM",
		)
	}

	// ============================================================
	// PKCS#1
	// ============================================================

	if key, err := x509.ParsePKCS1PrivateKey(
		block.Bytes,
	); err == nil {

		return key, nil
	}

	// ============================================================
	// PKCS#8
	// ============================================================

	key, err := x509.ParsePKCS8PrivateKey(
		block.Bytes,
	)

	if err != nil {
		return nil, errors.New(
			"invalid PKCS#8 OIDC private key",
		)
	}

	// ============================================================
	// Ensure RSA
	// ============================================================

	rsaKey, ok := key.(*rsa.PrivateKey)

	if !ok {
		return nil, errors.New(
			"OIDC private key is not RSA",
		)
	}

	return rsaKey, nil
}
