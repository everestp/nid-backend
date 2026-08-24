// modules/auth/service/auth_service.go
package service

import (
	"errors"
	"strings"
	"nid-backend/modules/auth/repository"
	"nid-backend/pkg/helpers"
)

type AuthService struct {
	repo *repository.AuthRepository
}

func NewAuthService(repo *repository.AuthRepository) *AuthService {
	return &AuthService{repo: repo}
}

func (s *AuthService) AuthenticateWithHandle(handle, address, signature, message, chain string) (string, string, error) {
	handle = strings.TrimSpace(strings.ToLower(handle))
	if handle == "" || address == "" || chain == "" {
		return "", "", errors.New("handle, address, and chain are required")
	}

	// TODO: Add cryptographic signature verification here
	// (e.g., verifying that `signature` matches `message` signed by `address`)

	// Updated: Strictly looks up existing users. Fails if the handle is not claimed yet.
	userID, err := s.repo.FindUserByHandleAndWallet(handle, address)
	if err != nil {
		return "", "", err
	}

	// Generate secure session token
	token := helpers.GenerateToken(userID)
	return userID, token, nil
}
