// modules/resolution/service/resolution_service.go
package service

import (
	"errors"
	"nid-backend/modules/resolution/repository"
)

type ResolutionService struct {
	repo *repository.ResolutionRepository
}

func NewResolutionService(repo *repository.ResolutionRepository) *ResolutionService {
	return &ResolutionService{repo: repo}
}

func (s *ResolutionService) Resolve(handleName, chain string) (string, error) {
	if handleName == "" || chain == "" {
		return "", errors.New("handle and chain are required")
	}
	return s.repo.ResolveAddress(handleName, chain)
}
