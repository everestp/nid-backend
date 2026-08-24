package repository

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"time"

	"nid-backend/modules/oidc/dto"

	"golang.org/x/crypto/bcrypt"
)

type OIDCRepository struct {
	db *sql.DB
}

func NewOIDCRepository(db *sql.DB) *OIDCRepository {
	return &OIDCRepository{
		db: db,
	}
}

// ============================================================
// Client Registration
// ============================================================

func (r *OIDCRepository) CreateClient(
	clientID string,
	clientSecretHash string,
	name string,
	redirectURI string,
	clientType string,
) error {

	_, err := r.db.Exec(`
		INSERT INTO oauth_clients (
			client_id,
			client_secret_hash,
			name,
			redirect_uri,
			client_type
		)
		VALUES ($1, $2, $3, $4, $5)
	`,
		clientID,
		clientSecretHash,
		name,
		redirectURI,
		clientType,
	)

	return err
}

func (r *OIDCRepository) GetClient(
	clientID string,
) (
	string,
	string,
	string,
	string,
	string,
	error,
) {

	var (
		dbClientID      string
		clientSecretHash string
		name            string
		redirectURI     string
		clientType      string
	)

	err := r.db.QueryRow(`
		SELECT
			client_id,
			client_secret_hash,
			name,
			redirect_uri,
			client_type
		FROM oauth_clients
		WHERE client_id = $1
	`,
		clientID,
	).Scan(
		&dbClientID,
		&clientSecretHash,
		&name,
		&redirectURI,
		&clientType,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return "", "", "", "", "", errors.New("client not found")
		}

		return "", "", "", "", "", err
	}

	return dbClientID,
		clientSecretHash,
		name,
		redirectURI,
		clientType,
		nil
}

// ============================================================
// Validate Client
// ============================================================

func (r *OIDCRepository) ValidateClient(
	clientID string,
	clientSecret string,
	redirectURI string,
) error {

	_, secretHash, _, registeredRedirectURI, _, err :=
		r.GetClient(clientID)

	if err != nil {
		return errors.New("invalid client")
	}

	if registeredRedirectURI != redirectURI {
		return errors.New("invalid redirect uri")
	}

	if secretHash == "" {
		return errors.New("client authentication required")
	}

	if err := bcrypt.CompareHashAndPassword(
		[]byte(secretHash),
		[]byte(clientSecret),
	); err != nil {
		return errors.New("invalid client credentials")
	}

	return nil
}

// ============================================================
// Authorization Code
// ============================================================

func (r *OIDCRepository) SaveAuthorizationCode(
	code string,
	clientID string,
	userID string,
	redirectURI string,
	scope string,
	nonce string,
	codeChallenge string,
) error {

	hash := sha256.Sum256([]byte(code))

	codeHash := hex.EncodeToString(hash[:])

	expiresAt := time.Now().Add(60 * time.Second)

	_, err := r.db.Exec(`
		INSERT INTO oauth_codes (
			code_hash,
			client_id,
			user_id,
			redirect_uri,
			scope,
			nonce,
			code_challenge,
			code_challenge_method,
			expires_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, 'S256', $8)
	`,
		codeHash,
		clientID,
		userID,
		redirectURI,
		scope,
		nonce,
		codeChallenge,
		expiresAt,
	)

	return err
}

// ============================================================
// Consume Authorization Code
// ============================================================

type AuthorizationCode struct {
	UserID          string
	ClientID        string
	RedirectURI     string
	Scope           string
	Nonce           string
	CodeChallenge   string
	ExpiresAt       time.Time
}

func (r *OIDCRepository) ConsumeAuthorizationCode(
	code string,
	clientID string,
	redirectURI string,
) (*AuthorizationCode, error) {

	hash := sha256.Sum256([]byte(code))

	codeHash := hex.EncodeToString(hash[:])

	tx, err := r.db.Begin()
	if err != nil {
		return nil, err
	}

	defer tx.Rollback()

	var result AuthorizationCode

	err = tx.QueryRow(`
		SELECT
			user_id,
			client_id,
			redirect_uri,
			scope,
			COALESCE(nonce, ''),
			code_challenge,
			expires_at
		FROM oauth_codes
		WHERE code_hash = $1
		  AND client_id = $2
		  AND redirect_uri = $3
		  AND used_at IS NULL
		FOR UPDATE
	`,
		codeHash,
		clientID,
		redirectURI,
	).Scan(
		&result.UserID,
		&result.ClientID,
		&result.RedirectURI,
		&result.Scope,
		&result.Nonce,
		&result.CodeChallenge,
		&result.ExpiresAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New(
				"invalid authorization code",
			)
		}

		return nil, err
	}

	if time.Now().After(result.ExpiresAt) {
		return nil, errors.New(
			"authorization code expired",
		)
	}

	_, err = tx.Exec(`
		UPDATE oauth_codes
		SET used_at = CURRENT_TIMESTAMP
		WHERE code_hash = $1
	`,
		codeHash,
	)

	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return &result, nil
}

// ============================================================
// Access Token
// ============================================================

func (r *OIDCRepository) SaveAccessToken(
	tokenHash string,
	clientID string,
	userID string,
	scope string,
	expiresAt time.Time,
) error {

	_, err := r.db.Exec(`
		INSERT INTO oauth_access_tokens (
			token_hash,
			client_id,
			user_id,
			scope,
			expires_at
		)
		VALUES ($1, $2, $3, $4, $5)
	`,
		tokenHash,
		clientID,
		userID,
		scope,
		expiresAt,
	)

	return err
}

// ============================================================
// Access Token Validation
// ============================================================

func (r *OIDCRepository) GetUserByAccessToken(
	tokenHash string,
) (string, error) {

	var userID string

	err := r.db.QueryRow(`
		SELECT user_id
		FROM oauth_access_tokens
		WHERE token_hash = $1
		  AND revoked_at IS NULL
		  AND expires_at > CURRENT_TIMESTAMP
	`,
		tokenHash,
	).Scan(&userID)

	if err != nil {
		if err == sql.ErrNoRows {
			return "", errors.New("invalid access token")
		}

		return "", err
	}

	return userID, nil
}

// ============================================================
// User Handle
// ============================================================

func (r *OIDCRepository) GetPrimaryHandleByUserID(
	userID string,
) (string, error) {

	var handle string

	err := r.db.QueryRow(`
		SELECT handle
		FROM handles
		WHERE user_id = $1
		  AND is_primary = true
		LIMIT 1
	`,
		userID,
	).Scan(&handle)

	if err == nil {
		return handle, nil
	}

	err = r.db.QueryRow(`
		SELECT handle
		FROM handles
		WHERE user_id = $1
		ORDER BY created_at ASC
		LIMIT 1
	`,
		userID,
	).Scan(&handle)

	if err != nil {
		return "", errors.New(
			"no handle found for user",
		)
	}

	return handle, nil
}

// Keep DTO imported if you later expand repository methods.
var _ = dto.TokenResponse{}
