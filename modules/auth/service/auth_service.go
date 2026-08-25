package service

import (
	"errors"
	"fmt"
	"strings"

	"nid-backend/modules/auth/repository"
	"nid-backend/pkg/helpers"
)

type AuthService struct {
	repo *repository.AuthRepository
}

func NewAuthService(
	repo *repository.AuthRepository,
) *AuthService {
	return &AuthService{
		repo: repo,
	}
}

// ============================================================================
// IN-HOUSE AUTHENTICATION
// ============================================================================
//
// Used ONLY by NID itself.
//
// Flow:
//
//   wallet signature
//        ↓
//   verify wallet ownership
//        ↓
//   find NID account
//        ↓
//   generate internal session token
//        ↓
//   app_session
//
// This function has NOTHING to do with OAuth/OIDC.
//
// ============================================================================

func (s *AuthService) AuthenticateInHouse(
	handle string,
	address string,
	signature string,
	message string,
	chain string,
) (string, string, error) {

	handle = strings.TrimSpace(
		strings.ToLower(handle),
	)

	address = strings.TrimSpace(address)
	signature = strings.TrimSpace(signature)
	message = strings.TrimSpace(message)
	chain = strings.TrimSpace(
		strings.ToLower(chain),
	)

	// ------------------------------------------------------------------------
	// Validate input
	// ------------------------------------------------------------------------

	if handle == "" ||
		address == "" ||
		signature == "" ||
		message == "" ||
		chain == "" {

		return "", "", errors.New(
			"handle, address, chain, signature, and message are required",
		)
	}

	// ------------------------------------------------------------------------
	// Verify wallet signature
	// ------------------------------------------------------------------------

	if err := helpers.VerifySignature(
		chain,
		address,
		message,
		signature,
	); err != nil {

		return "", "", fmt.Errorf(
			"invalid authentication signature: %w",
			err,
		)
	}

	// ------------------------------------------------------------------------
	// Find existing NID account
	// ------------------------------------------------------------------------

	userID, err := s.repo.FindUserByHandleAndWallet(
		handle,
		address,
	)

	if err != nil {
		return "", "", fmt.Errorf(
			"account not found or wallet does not match handle: %w",
			err,
		)
	}

	userID = strings.TrimSpace(userID)

	if userID == "" {
		return "", "", errors.New(
			"authenticated user id is empty",
		)
	}

	// ------------------------------------------------------------------------
	// Generate INTERNAL session token
	// ------------------------------------------------------------------------

	sessionToken := helpers.GenerateInternalSessionToken(userID)

	if strings.TrimSpace(sessionToken) == "" {
		return "", "", errors.New(
			"failed to generate session token",
		)
	}

	return userID, sessionToken, nil
}

// ============================================================================
// OIDC AUTHENTICATION
// ============================================================================
//
// Used ONLY by the NID Identity Provider.
//
// This function should be called by your OIDC authorization flow.
//
// It does NOT generate app_session.
//
// It generates an OIDC ID token for the external application.
//
// ============================================================================

func (s *AuthService) AuthenticateOIDC(
	handle string,
	address string,
	signature string,
	message string,
	chain string,
	clientID string,
) (string, string, error) {

	handle = strings.TrimSpace(
		strings.ToLower(handle),
	)

	address = strings.TrimSpace(address)
	signature = strings.TrimSpace(signature)
	message = strings.TrimSpace(message)
	chain = strings.TrimSpace(
		strings.ToLower(chain),
	)
	clientID = strings.TrimSpace(clientID)

	// ------------------------------------------------------------------------
	// Validate input
	// ------------------------------------------------------------------------

	if handle == "" ||
		address == "" ||
		signature == "" ||
		message == "" ||
		chain == "" ||
		clientID == "" {

		return "", "", errors.New(
			"handle, address, chain, signature, message, and client_id are required",
		)
	}

	// ------------------------------------------------------------------------
	// Verify wallet signature
	// ------------------------------------------------------------------------

	if err := helpers.VerifySignature(
		chain,
		address,
		message,
		signature,
	); err != nil {

		return "", "", fmt.Errorf(
			"invalid OIDC authentication signature: %w",
			err,
		)
	}

	// ------------------------------------------------------------------------
	// Find NID account
	// ------------------------------------------------------------------------

	userID, err := s.repo.FindUserByHandleAndWallet(
		handle,
		address,
	)

	if err != nil {
		return "", "", fmt.Errorf(
			"account not found or wallet does not match handle: %w",
			err,
		)
	}

	userID = strings.TrimSpace(userID)

	if userID == "" {
		return "", "", errors.New(
			"authenticated user id is empty",
		)
	}

	// ------------------------------------------------------------------------
	// Generate OIDC ID Token
	// ------------------------------------------------------------------------

	idToken, err := helpers.GenerateOIDCIDToken(
		userID,
		handle,
		clientID,
	)

	if err != nil {
		return "", "", fmt.Errorf(
			"failed to generate OIDC ID token: %w",
			err,
		)
	}

	return userID, idToken, nil
}
