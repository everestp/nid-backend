package service

import (
	"errors"
	"strings"
	"nid-backend/modules/handle/repository"
)

type HandleService struct {
	repo *repository.HandleRepository
}

func NewHandleService(repo *repository.HandleRepository) *HandleService {
	return &HandleService{repo: repo}
}

func (s *HandleService) ClaimHandle(name, address, chain, signature string) (*repository.HandleModel, error) {
	name = strings.TrimSpace(strings.ToLower(name))
	if name == "" || address == "" || chain == "" {
		return nil, errors.New("name, address, and chain are required")
	}

	// TODO: Verify cryptographic signature here if needed

	return s.repo.ClaimHandleForWallet(name, address, chain)
}
