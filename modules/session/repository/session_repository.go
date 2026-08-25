// modules/session/repository/session_repository.go

package repository

import (
	"database/sql"
	"errors"

	"nid-backend/modules/session/dto"
)

type SessionRepository struct {
	db *sql.DB
}

// ============================================================
// CONSTRUCTOR
// ============================================================

func NewSessionRepository(db *sql.DB) *SessionRepository {
	return &SessionRepository{
		db: db,
	}
}

// ============================================================
// CREATE SESSION
// ============================================================
//
// Creates a login session for a user + OAuth client.
//
// Example:
// user A logs into example.com through NID
//
// oauth_sessions:
// user_id   = user A
// client_id = example.com
// status    = active
//
// ============================================================

func (r *SessionRepository) CreateSession(
	userID string,
	clientID string,
) (string, error) {

	var sessionID string

	err := r.db.QueryRow(
		`
		INSERT INTO oauth_sessions (
			user_id,
			client_id,
			status
		)
		VALUES ($1, $2, 'active')
		RETURNING id
		`,
		userID,
		clientID,
	).Scan(&sessionID)

	if err != nil {
		return "", err
	}

	return sessionID, nil
}

// ============================================================
// GET USER SESSIONS
// ============================================================
//
// Returns all sessions belonging to the authenticated user.
//
// ============================================================

func (r *SessionRepository) GetUserSessions(
	userID string,
) ([]dto.Session, error) {

	rows, err := r.db.Query(
		`
		SELECT
			s.id,
			COALESCE(s.client_id, ''),
			COALESCE(c.name, 'Unknown App'),
			s.status,
			s.created_at
		FROM oauth_sessions s
		LEFT JOIN oauth_clients c
			ON c.client_id = s.client_id
		WHERE s.user_id = $1
		ORDER BY s.created_at DESC
		`,
		userID,
	)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	sessions := make([]dto.Session, 0)

	for rows.Next() {

		var session dto.Session

		err := rows.Scan(
			&session.ID,
			&session.ClientID,
			&session.AppName,
			&session.Status,
			&session.CreatedAt,
		)

		if err != nil {
			return nil, err
		}

		sessions = append(sessions, session)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return sessions, nil
}

// ============================================================
// GET SINGLE SESSION
// ============================================================
//
// Important:
//
// sessionID + userID are both checked.
//
// A user cannot access another user's session.
//
// ============================================================

func (r *SessionRepository) GetSession(
	sessionID string,
	userID string,
) (*dto.Session, error) {

	var session dto.Session

	err := r.db.QueryRow(
		`
		SELECT
			s.id,
			COALESCE(s.client_id, ''),
			COALESCE(c.name, 'Unknown App'),
			s.status,
			s.created_at
		FROM oauth_sessions s
		LEFT JOIN oauth_clients c
			ON c.client_id = s.client_id
		WHERE s.id = $1
		  AND s.user_id = $2
		`,
		sessionID,
		userID,
	).Scan(
		&session.ID,
		&session.ClientID,
		&session.AppName,
		&session.Status,
		&session.CreatedAt,
	)

	if err != nil {

		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("session not found")
		}

		return nil, err
	}

	return &session, nil
}

// ============================================================
// REVOKE SESSION
// ============================================================
//
// Revokes only one session.
//
// IMPORTANT:
//
// user_id is included in WHERE clause so a user cannot revoke
// somebody else's session.
//
// ============================================================

func (r *SessionRepository) RevokeSession(
	sessionID string,
	userID string,
) error {

	result, err := r.db.Exec(
		`
		UPDATE oauth_sessions
		SET status = 'revoked'
		WHERE id = $1
		  AND user_id = $2
		  AND status = 'active'
		`,
		sessionID,
		userID,
	)

	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()

	if err != nil {
		return err
	}

	if rows == 0 {
		return errors.New(
			"session not found or already revoked",
		)
	}

	return nil
}

// ============================================================
// REVOKE ALL SESSIONS
// ============================================================
//
// Revokes every active session belonging to a user.
//
// ============================================================

func (r *SessionRepository) RevokeAllSessions(
	userID string,
) error {

	_, err := r.db.Exec(
		`
		UPDATE oauth_sessions
		SET status = 'revoked'
		WHERE user_id = $1
		  AND status = 'active'
		`,
		userID,
	)

	return err
}

// ============================================================
// EXPIRE OLD SESSIONS
// ============================================================
//
// Sessions older than 30 days become expired.
//
// This can be called from a background cleanup job.
//
// ============================================================

func (r *SessionRepository) ExpireSessions() error {

	_, err := r.db.Exec(
		`
		UPDATE oauth_sessions
		SET status = 'expired'
		WHERE status = 'active'
		  AND created_at < NOW() - INTERVAL '30 days'
		`,
	)

	return err
}

// ============================================================
// REVOKE SESSION + ITS ACCESS TOKENS
// ============================================================
//
// When user revokes:
//
//     Session
//
// we also revoke every access token belonging to that session.
//
// Both operations happen inside one transaction.
//
// ============================================================

func (r *SessionRepository) RevokeSessionWithTokens(
	sessionID string,
	userID string,
) error {

	tx, err := r.db.Begin()

	if err != nil {
		return err
	}

	defer tx.Rollback()

	// --------------------------------------------------------
	// Revoke session
	// --------------------------------------------------------

	result, err := tx.Exec(
		`
		UPDATE oauth_sessions
		SET status = 'revoked'
		WHERE id = $1
		  AND user_id = $2
		  AND status = 'active'
		`,
		sessionID,
		userID,
	)

	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()

	if err != nil {
		return err
	}

	if rows == 0 {
		return errors.New(
			"session not found or already revoked",
		)
	}

	// --------------------------------------------------------
	// Revoke access tokens belonging to this session
	// --------------------------------------------------------

	_, err = tx.Exec(
		`
		UPDATE oauth_access_tokens
		SET revoked_at = NOW()
		WHERE session_id = $1
		  AND revoked_at IS NULL
		`,
		sessionID,
	)

	if err != nil {
		return err
	}

	// --------------------------------------------------------
	// Commit
	// --------------------------------------------------------

	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}

// ============================================================
// REVOKE ALL SESSIONS + ALL TOKENS
// ============================================================
//
// "Log out everywhere"
//
// This revokes:
//
//     1. All NID sessions
//     2. All OAuth access tokens
//
// ============================================================

func (r *SessionRepository) RevokeAllSessionsWithTokens(
	userID string,
) error {

	tx, err := r.db.Begin()

	if err != nil {
		return err
	}

	defer tx.Rollback()

	// --------------------------------------------------------
	// Revoke all sessions
	// --------------------------------------------------------

	_, err = tx.Exec(
		`
		UPDATE oauth_sessions
		SET status = 'revoked'
		WHERE user_id = $1
		  AND status = 'active'
		`,
		userID,
	)

	if err != nil {
		return err
	}

	// --------------------------------------------------------
	// Revoke all access tokens
	// --------------------------------------------------------

	_, err = tx.Exec(
		`
		UPDATE oauth_access_tokens
		SET revoked_at = NOW()
		WHERE user_id = $1
		  AND revoked_at IS NULL
		`,
		userID,
	)

	if err != nil {
		return err
	}

	// --------------------------------------------------------
	// Commit
	// --------------------------------------------------------

	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}

// ============================================================
// CLEANUP EXPIRED SESSIONS
// ============================================================
//
// Marks old active sessions as expired.
//
// ============================================================

func (r *SessionRepository) CleanupExpiredSessions() error {

	_, err := r.db.Exec(
		`
		UPDATE oauth_sessions
		SET status = 'expired'
		WHERE status = 'active'
		  AND created_at < NOW() - INTERVAL '30 days'
		`,
	)

	return err
}

// ============================================================
// CLEANUP EXPIRED TOKENS
// ============================================================
//
// Access tokens are revoked after expiration.
//
// ============================================================

func (r *SessionRepository) CleanupExpiredTokens() error {

	_, err := r.db.Exec(
		`
		UPDATE oauth_access_tokens
		SET revoked_at = NOW()
		WHERE revoked_at IS NULL
		  AND expires_at <= NOW()
		`,
	)

	return err
}
