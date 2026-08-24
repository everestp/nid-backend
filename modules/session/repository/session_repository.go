// modules/session/repository/session_repository.go
package repository

import (
	"database/sql"
	"errors"
)

type SessionRepository struct {
	db *sql.DB
}

func NewSessionRepository(db *sql.DB) *SessionRepository {
	return &SessionRepository{db: db}
}

func (r *SessionRepository) RevokeSession(sessionID, userID string) error {
	query := `UPDATE oauth_sessions SET status = 'expired' WHERE id = $1 AND user_id = $2`
	res, err := r.db.Exec(query, sessionID, userID)
	if err != nil {
		return err
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return errors.New("session not found or unauthorized")
	}
	return nil
}
