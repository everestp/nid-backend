// modules/resolution/service/resolution_service.go
package service

import (
	"nid-backend/modules/resolution/repository"
)

type ResolutionService struct {
	repo *repository.ResolutionRepository
}

func NewResolutionService(repo *repository.ResolutionRepository) *ResolutionService {
	return &ResolutionService{repo: repo}
}

func (s *ResolutionService) Resolve(handleName, chain string) (string, error) {
	return s.repo.ResolveAddress(handleName, chain)
}
