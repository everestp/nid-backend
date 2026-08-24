// modules/auth/service/auth_service.go
package service

import (
	"errors"
	"nid-backend/modules/auth/repository"
	"nid-backend/pkg/helpers"
)

type AuthService struct {
	repo *repository.AuthRepository
}

func NewAuthService(repo *repository.AuthRepository) *AuthService {
	return &AuthService{repo: repo}
}

func (s *AuthService) AuthenticateWallet(address, signature, message, chain string) (string, string, error) {
	if address == "" || chain == "" {
		return "", "", errors.New("address and chain are required")
	}

	userID, err := s.repo.FindOrCreateUserByWallet(address, chain)
	if err != nil {
		return "", "", err
	}

	token := helpers.GenerateToken(userID)
	return userID, token, nil
}
