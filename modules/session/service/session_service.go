// modules/session/service/session_service.go
package service

import (
	"nid-backend/modules/session/repository"
)

type SessionService struct {
	repo *repository.SessionRepository
}

func NewSessionService(repo *repository.SessionRepository) *SessionService {
	return &SessionService{repo: repo}
}

func (s *SessionService) Revoke(sessionID string) error {
	return s.repo.RevokeSession(sessionID)
}
