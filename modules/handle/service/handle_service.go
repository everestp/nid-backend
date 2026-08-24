package service

import (
	"errors"
	"fmt"
	"nid-backend/modules/handle/repository"
	"nid-backend/pkg/helpers"
	"strings"
)

type HandleService struct {
	repo *repository.HandleRepository
}

func NewHandleService(repo *repository.HandleRepository) *HandleService {
	return &HandleService{repo: repo}
}

func (s *HandleService) ClaimHandle(name, address, chain, signature string) (*repository.HandleModel, error) {
	name = strings.TrimSpace(strings.ToLower(name))
	if name == "" || address == "" || chain == "" || signature == "" {
		return nil, errors.New("name, address, chain, and signature are required")
	}

	// Reconstruct the exact message your frontend signed
	expectedMessage := fmt.Sprintf("Claim %s.nid", name)

	// --- CRYPTOGRAPHIC VERIFICATION ---
	err := helpers.VerifySignature(chain, address, expectedMessage, signature)
	if err != nil {
		return nil, fmt.Errorf("invalid signature: %w", err)
	}

	// If signature is valid, proceed to save in the database
	return s.repo.ClaimHandleForWallet(name, address, chain)
}
