// modules/session/repository/session_repository.go
package repository

import (
	"database/sql"
)

type SessionRepository struct {
	db *sql.DB
}

func NewSessionRepository(db *sql.DB) *SessionRepository {
	return &SessionRepository{db: db}
}

func (r *SessionRepository) RevokeSession(sessionID string) error {
	query := `UPDATE oauth_sessions SET status = 'expired' WHERE id = $1`
	_, err := r.db.Exec(query, sessionID)
	return err
}
