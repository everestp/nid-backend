// modules/handle/service/handle_service.go
package service

import (
	"errors"
	"nid-backend/models"
	"nid-backend/modules/handle/repository"
)

type HandleService struct {
	repo *repository.HandleRepository
}

func NewHandleService(repo *repository.HandleRepository) *HandleService {
	return &HandleService{repo: repo}
}

func (s *HandleService) ClaimHandle(userID, name string) (*models.Handle, error) {
	if len(name) < 3 {
		return nil, errors.New("handle name too short")
	}
	existing, _ := s.repo.FindByName(name)
	if existing != nil {
		return nil, errors.New("handle already taken")
	}
	return s.repo.Create(userID, name)
}

func (s *HandleService) ResolveHandle(name string) (*models.Handle, error) {
	return s.repo.FindByName(name)
}
