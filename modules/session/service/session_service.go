// modules/session/service/session_service.go
package service

import (
	"errors"
	"nid-backend/modules/session/repository"
)

type SessionService struct {
	repo *repository.SessionRepository
}

func NewSessionService(repo *repository.SessionRepository) *SessionService {
	return &SessionService{repo: repo}
}

func (s *SessionService) Revoke(sessionID, userID string) error {
	if sessionID == "" {
		return errors.New("session ID is required")
	}
	return s.repo.RevokeSession(sessionID, userID)
}
