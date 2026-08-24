package helpers

import (
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"

	"github.com/gagliardetto/solana-go"
	"github.com/ethereum/go-ethereum/common/hexutil"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"
)

// Helper function to handle EVM and Solana signature checks
func VerifySignature(chain, expectedAddress, message, signatureStr string) error {

	switch strings.ToLower(chain) {

	case "evm", "ethereum":
		return verifyEVMSignature(
			expectedAddress,
			message,
			signatureStr,
		)

	case "solana":
		return verifySolanaSignature(
			expectedAddress,
			message,
			signatureStr,
		)

	default:
		return fmt.Errorf(
			"unsupported chain: %s",
			chain,
		)
	}
}

// 1. EVM (MetaMask / personal_sign) Verification
func verifyEVMSignature(
	expectedAddress,
	message,
	signatureStr string,
) error {

	// Format Ethereum personal_sign message prefix
	hash := ethcrypto.Keccak256Hash([]byte(message))

	// Decode hex signature
	sigBytes, err := hexutil.Decode(signatureStr)
	if err != nil {
		return errors.New("failed to decode hex signature")
	}

	if len(sigBytes) != 65 {
		return errors.New("invalid signature length")
	}

	// Adjust recovery ID for go-ethereum (27/28 -> 0/1)
	if sigBytes[64] >= 27 {
		sigBytes[64] -= 27
	}

	// Recover public key from signature
	pubKey, err := ethcrypto.SigToPub(
		hash.Bytes(),
		sigBytes,
	)
	if err != nil {
		return errors.New(
			"failed to recover public key from signature",
		)
	}

	// Derive address from public key
	recoveredAddress := ethcrypto.PubkeyToAddress(
		*pubKey,
	).Hex()

	// Compare with expected address
	if !strings.EqualFold(
		recoveredAddress,
		expectedAddress,
	) {
		return fmt.Errorf(
			"address mismatch: expected %s, got %s",
			expectedAddress,
			recoveredAddress,
		)
	}

	return nil
}

// 2. Solana (Phantom / signMessage) Verification
func verifySolanaSignature(
	expectedAddress,
	message,
	signatureStr string,
) error {

	// --------------------------------------------------
	// Decode Solana public key
	//
	// expectedAddress is the wallet public key.
	// Example:
	//
	// 8KFYXCqqv8BkHhZ6aGkydmC57gYi8tmNDNVAfQ9yQRZf
	// --------------------------------------------------

	pubKey, err := solana.PublicKeyFromBase58(
		expectedAddress,
	)
	if err != nil {
		return errors.New(
			"invalid solana public key address",
		)
	}

	// --------------------------------------------------
	// Decode signature
	//
	// Your frontend does:
	//
	// btoa(String.fromCharCode(...sigBytes))
	//
	// Therefore signatureStr is Base64.
	// --------------------------------------------------

	var sigBytes []byte

	if strings.HasPrefix(signatureStr, "0x") {

		// Keep support for hex signatures too.
		sigBytes, err = hex.DecodeString(
			signatureStr[2:],
		)

	} else {

		// Phantom signature from your frontend is Base64.
		sigBytes, err = base64.StdEncoding.DecodeString(
			signatureStr,
		)
	}

	if err != nil {
		return errors.New(
			"failed to decode solana signature",
		)
	}

	// Ed25519 signature must be exactly 64 bytes.
	if len(sigBytes) != 64 {
		return errors.New(
			"invalid solana signature length",
		)
	}

	// Convert []byte to solana.Signature.
	var sig solana.Signature

	copy(
		sig[:],
		sigBytes,
	)

	// --------------------------------------------------
	// Verify:
	//
	// Public Key
	// +
	// Exact Message
	// +
	// Signature
	//
	// --------------------------------------------------

	messageBytes := []byte(message)

	if !pubKey.Verify(
		messageBytes,
		sig,
	) {
		return errors.New(
			"solana signature verification failed",
		)
	}

	return nil
}
