// modules/session/service/session_service.go

package service

import (
	"errors"
	"strings"

	"nid-backend/modules/session/dto"
	"nid-backend/modules/session/repository"
)

type SessionService struct {
	repo *repository.SessionRepository
}

// ============================================================
// Constructor
// ============================================================

func NewSessionService(
	repo *repository.SessionRepository,
) *SessionService {
	return &SessionService{
		repo: repo,
	}
}

// ============================================================
// List User Sessions
// ============================================================

func (s *SessionService) List(
	userID string,
) ([]dto.Session, error) {

	if userID == "" {
		return nil, errors.New("user ID is required")
	}

	return s.repo.GetUserSessions(userID)
}

// ============================================================
// Revoke Session
// ============================================================

func (s *SessionService) Revoke(
	sessionID string,
	userID string,
) error {

	sessionID = strings.TrimSpace(sessionID)
	userID = strings.TrimSpace(userID)

	if sessionID == "" {
		return errors.New("session ID is required")
	}

	if userID == "" {
		return errors.New("user ID is required")
	}

	// Repository must verify that this session belongs
	// to the authenticated user before revoking it.
	err := s.repo.RevokeSession(
		sessionID,
		userID,
	)

	if err != nil {
		return err
	}

	return nil
}
