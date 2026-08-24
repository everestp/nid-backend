package service

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"nid-backend/modules/oidc/repository"
	"nid-backend/pkg/helpers"

)

type OIDCService struct {
	repo *repository.OIDCRepository
}

func NewOIDCService(repo *repository.OIDCRepository) *OIDCService {
	return &OIDCService{repo: repo}
}

func (s *OIDCService) GenerateAuthCode(clientID string, userID string) (string, error) {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	code := hex.EncodeToString(bytes)

	err := s.repo.SaveAuthorizationCode(code, clientID, userID)
	if err != nil {
		return "", err
	}
	return code, nil
}

func (s *OIDCService) ExchangeCodeForToken(clientID, clientSecret, code string) (string, string, error) {
	// 1. Verify Client Credentials
	valid, err := s.repo.ValidateClient(clientID, clientSecret)
	if err != nil || !valid {
		return "", "", errors.New("invalid client credentials")
	}

	// 2. Consume Authorization Code & get UserID
	userID, err := s.repo.ConsumeAuthorizationCode(code, clientID)
	if err != nil {
		return "", "", err
	}

	// 3. Get User's Handle
	userHandle, err := s.repo.GetPrimaryHandleByUserID(userID)
	if err != nil {
		return "", "", err
	}

	// 4. Generate Standard OIDC ID Token (JWT) — FIXED with error handling
	idToken, err := helpers.GenerateIDToken(userID, userHandle, clientID)
	if err != nil {
		return "", "", err
	}

	return userID, idToken, nil
}
