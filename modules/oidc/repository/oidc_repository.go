package repository

import (
	"database/sql"
	"errors"
	"time"
	
)

type OIDCRepository struct {
	db *sql.DB
}

func NewOIDCRepository(db *sql.DB) *OIDCRepository {
	return &OIDCRepository{db: db}
}

// ValidateClient checks if client_id and client_secret are registered
func (r *OIDCRepository) ValidateClient(clientID, clientSecret string) (bool, error) {
	var exists bool
	err := r.db.QueryRow(
		`SELECT EXISTS(SELECT 1 FROM oauth_clients WHERE client_id = $1 AND client_secret = $2)`,
		clientID, clientSecret,
	).Scan(&exists)
	return exists, err
}

// SaveAuthorizationCode stores a temporary 60-second auth code
func (r *OIDCRepository) SaveAuthorizationCode(code, clientID string, userID string) error {
	expiresAt := time.Now().Add(60 * time.Second)
	_, err := r.db.Exec(
		`INSERT INTO oauth_codes (code, client_id, user_id, expires_at) VALUES ($1, $2, $3, $4)`,
		code, clientID, userID, expiresAt,
	)
	return err
}

// ConsumeAuthorizationCode checks, returns the userID, and deletes the code (single-use)
func (r *OIDCRepository) ConsumeAuthorizationCode(code, clientID string) (string, error) {
	var userID string
	var expiresAt time.Time

	err := r.db.QueryRow(
		`SELECT user_id, expires_at FROM oauth_codes WHERE code = $1 AND client_id = $2`,
		code, clientID,
	).Scan(&userID, &expiresAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return "", errors.New("invalid or expired authorization code")
		}
		return "", err
	}

	// Delete code immediately (One-time use)
	_, _ = r.db.Exec(`DELETE FROM oauth_codes WHERE code = $1`, code)

	if time.Now().After(expiresAt) {
		return "", errors.New("authorization code has expired")
	}

	return userID, nil
}

// GetPrimaryHandleByUserID fetches the user's primary .nid handle
func (r *OIDCRepository) GetPrimaryHandleByUserID(userID string) (string, error) {
	var handle string
	err := r.db.QueryRow(
		`SELECT handle FROM handles WHERE user_id = $1 AND is_primary = true`,
		userID,
	).Scan(&handle)
	if err != nil {
		// Fallback to any handle if primary isn't explicitly flagged
		err = r.db.QueryRow(`SELECT handle FROM handles WHERE user_id = $1 LIMIT 1`, userID).Scan(&handle)
		if err != nil {
			return "", errors.New("no handle found for user")
		}
	}
	return handle, nil
}
