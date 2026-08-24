// modules/auth/service/auth_service.go
package service

import (
	"errors"
	"fmt"
	"nid-backend/modules/auth/repository"
	"nid-backend/pkg/helpers"
	"strings"
)

type AuthService struct {
	repo *repository.AuthRepository
}

func NewAuthService(repo *repository.AuthRepository) *AuthService {
	return &AuthService{repo: repo}
}

func (s *AuthService) AuthenticateWithHandle(handle, address, signature, message, chain string) (string, string, error) {
	handle = strings.TrimSpace(strings.ToLower(handle))
	if handle == "" || address == "" || chain == "" || signature == "" || message == "" {
		return "", "", errors.New("handle, address, chain, signature, and message are required")
	}

	// --- CRYPTOGRAPHIC SIGNATURE VERIFICATION ---
	// Ensures the person logging in actually owns the private key of the provided address
	err := helpers.VerifySignature(chain, address, message, signature)
	if err != nil {
		return "", "", fmt.Errorf("invalid authentication signature: %w", err)
	}

	// Updated: Strictly looks up existing users. Fails if the handle is not claimed yet.
	userID, err := s.repo.FindUserByHandleAndWallet(handle, address)
	if err != nil {
		return "", "", fmt.Errorf("account not found or wallet does not match handle: %w", err)
	}

	// Generate secure session token
	token := helpers.GenerateToken(userID)
	return userID, token, nil
}
