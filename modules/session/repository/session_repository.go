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
// CREATE / RESTORE SESSION
// ============================================================
//
// One session per:
//
//	user_id + client_id
//
// If the same user logs into the same client again,
// the existing session is reused and activated.
//
// Examples:
//
//	user1 + client1 -> INSERT
//	user1 + client2 -> INSERT
//	user1 + client2 -> UPDATE existing
//	user2 + client1 -> INSERT
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
			status,
			last_used_at
		)
		VALUES ($1, $2, 'active', CURRENT_TIMESTAMP)

		ON CONFLICT ON CONSTRAINT oauth_sessions_user_client_unique
		DO UPDATE SET
			status = 'active',
			last_used_at = CURRENT_TIMESTAMP,
			revoked_at = NULL

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

		if err := rows.Scan(
			&session.ID,
			&session.ClientID,
			&session.AppName,
			&session.Status,
			&session.CreatedAt,
		); err != nil {
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
// sessionID + userID are both checked.
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
// Revokes only one active session belonging to the user.
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
// REVOKE SESSION + ITS ACCESS TOKENS
// ============================================================
//
// Both operations happen in one transaction.
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
	// Revoke access tokens belonging to session
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
	// Revoke all user sessions
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
	// Revoke all user access tokens
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
// Active sessions older than 30 days become expired.
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
// Tokens are marked revoked once their expiration time passes.
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
