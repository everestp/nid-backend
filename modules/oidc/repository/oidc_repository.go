package repository

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"nid-backend/models"
	"nid-backend/modules/oidc/dto"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// ============================================================
// OIDC Repository
// ============================================================

type OIDCRepository struct {
	db *sql.DB
}

// ============================================================
// Constructor
// ============================================================

func NewOIDCRepository(db *sql.DB) *OIDCRepository {
	return &OIDCRepository{
		db: db,
	}
}

// ============================================================
// Models
// ============================================================

type OAuthClient struct {
	ClientID         string
	ClientSecretHash string
	Name             string
	RedirectURI      string
	ClientType       string
}

type AuthorizationCode struct {
	UserID              string
	ClientID            string
	RedirectURI         string
	Scope               string
	Nonce               string
	CodeChallenge       string
	CodeChallengeMethod string
	ExpiresAt           time.Time
}

type AccessToken struct {
	UserID    string
	ClientID  string
	Scope     string
	ExpiresAt time.Time
}

// ============================================================
// Client Registration
// ============================================================

func (r *OIDCRepository) CreateClient(
    clientID string,
	userID   string,
    clientSecretHash string,
    clientName string,
    redirectURI string,
    clientType string,
    clientLogo string,
    clientURI string,
    policyURI string,
) error {

    _, err := r.db.Exec(`
        INSERT INTO oauth_clients (
            client_id,
            client_secret_hash,
            name,
            redirect_uri,
            client_type,
            client_logo,
            client_uri,
            policy_uri,
			user_id
        )
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8 , $9)
    `,
        clientID,
        clientSecretHash,
        clientName,
        redirectURI,
        clientType,
        clientLogo,
        clientURI,
        policyURI,
		userID,
    )

    return err
}

// ============================================================
// Get Client
// ============================================================

func (r *OIDCRepository) GetClient(
	clientID string,
) (*OAuthClient, error) {

	var client OAuthClient

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
		&client.ClientID,
		&client.ClientSecretHash,
		&client.Name,
		&client.RedirectURI,
		&client.ClientType,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("client not found")
		}

		return nil, err
	}

	return &client, nil
}

// ============================================================
// Validate Client
//
// Used by /oauth/token.
//
// Confidential client:
//   - client_secret required
//
// Public client:
//   - no client_secret required
// ============================================================

func (r *OIDCRepository) ValidateClient(
	clientID string,
	clientSecret string,
	redirectURI string,
) error {

	client, err := r.GetClient(clientID)
	if err != nil {
		return errors.New("invalid client")
	}

	// Exact redirect URI validation.
	if client.RedirectURI != redirectURI {
		return errors.New("invalid redirect uri")
	}

	// Public clients do not have a secret.
	if client.ClientType == "public" {
		return nil
	}

	// Confidential client must have secret hash.
	if client.ClientSecretHash == "" {
		return errors.New("client authentication required")
	}

	// Validate secret.
	if err := bcrypt.CompareHashAndPassword(
		[]byte(client.ClientSecretHash),
		[]byte(clientSecret),
	); err != nil {
		return errors.New("invalid client credentials")
	}

	return nil
}

// ============================================================
// Validate Redirect URI
//
// IMPORTANT:
//
// Authorization endpoint must validate redirect_uri
// before doing any redirect.
//
// Never redirect to an unregistered URI.
// ============================================================

func (r *OIDCRepository) ValidateRedirectURI(
	clientID string,
	redirectURI string,
) error {

	var registeredURI string

	err := r.db.QueryRow(`
		SELECT redirect_uri
		FROM oauth_clients
		WHERE client_id = $1
	`,
		clientID,
	).Scan(&registeredURI)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errors.New("client not found")
		}

		return err
	}

	if registeredURI != redirectURI {
		return errors.New("invalid redirect uri")
	}

	return nil
}

// ============================================================
// Authorization Code
//
// Raw code is NEVER stored.
//
// DB stores:
//
// SHA256(code)
//
// Code expires after 60 seconds.
// ============================================================

func (r *OIDCRepository) SaveAuthorizationCode(
	code string,
	clientID string,
	userID string,
	redirectURI string,
	scope string,
	nonce string,
	codeChallenge string,
	codeChallengeMethod string,
) error {

	hash := sha256.Sum256([]byte(code))
	codeHash := hex.EncodeToString(hash[:])

	expiresAt := time.Now().Add(60 * time.Second)

	if codeChallengeMethod == "" {
		codeChallengeMethod = "S256"
	}

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
		VALUES (
			$1,
			$2,
			$3,
			$4,
			$5,
			$6,
			$7,
			$8,
			$9
		)
	`,
		codeHash,
		clientID,
		userID,
		redirectURI,
		scope,
		nonce,
		codeChallenge,
		codeChallengeMethod,
		expiresAt,
	)

	return err
}

// ============================================================
// Consume Authorization Code
//
// Steps:
//
// 1. Hash raw authorization code
// 2. Start transaction
// 3. SELECT ... FOR UPDATE
// 4. Check client
// 5. Check redirect URI
// 6. Check unused
// 7. Check expiration
// 8. Mark as used
// 9. Commit
// 10. Return authorization code data
//
// This prevents authorization-code replay.
// ============================================================

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
			COALESCE(code_challenge_method, 'S256'),
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
		&result.CodeChallengeMethod,
		&result.ExpiresAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New(
				"invalid authorization code",
			)
		}

		return nil, err
	}

	// --------------------------------------------------------
	// Expiration
	// --------------------------------------------------------

	if time.Now().After(result.ExpiresAt) {
		return nil, errors.New(
			"authorization code expired",
		)
	}

	// --------------------------------------------------------
	// Mark authorization code as used
	// --------------------------------------------------------

	res, err := tx.Exec(`
		UPDATE oauth_codes
		SET used_at = CURRENT_TIMESTAMP
		WHERE code_hash = $1
			AND used_at IS NULL
	`,
		codeHash,
	)

	if err != nil {
		return nil, err
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return nil, err
	}

	if rowsAffected != 1 {
		return nil, errors.New(
			"authorization code already used",
		)
	}

	// --------------------------------------------------------
	// Commit
	// --------------------------------------------------------

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
    sessionID string,
    clientID string,
    userID string,
    scope string,
    expiresAt time.Time,
) error {

    _, err := r.db.ExecContext(
        context.Background(),
        `
        INSERT INTO oauth_access_tokens (
            token_hash,
            session_id,
            client_id,
            user_id,
            scope,
            expires_at
        )
        VALUES ($1, $2, $3, $4, $5, $6)
        `,
        tokenHash,
        sessionID,
        clientID,
        userID,
        scope,
        expiresAt,
    )

    return err
}

// ============================================================
// Get User By Access Token
//
// Used by:
//
// GET /oauth/userinfo
//
// Only active and non-expired tokens are accepted.
// ============================================================

func (r *OIDCRepository) GetUserByAccessToken(
    tokenHash string,
) (string, error) {

    var userID string

    err := r.db.QueryRowContext(
        context.Background(),
        `
        SELECT t.user_id
        FROM oauth_access_tokens t
        INNER JOIN oauth_sessions s
            ON s.id = t.session_id
        WHERE t.token_hash = $1
          AND t.revoked_at IS NULL
          AND t.expires_at > CURRENT_TIMESTAMP
          AND s.status = 'active'
          AND (
              s.expires_at IS NULL
              OR s.expires_at > CURRENT_TIMESTAMP
          )
        `,
        tokenHash,
    ).Scan(&userID)

    if err != nil {
        return "", err
    }

    return userID, nil
}

// ============================================================
// Get Access Token
//
// Useful when you need:
//
// - user_id
// - client_id
// - scope
// - expiry
// ============================================================

func (r *OIDCRepository) GetAccessToken(
	tokenHash string,
) (*AccessToken, error) {

	var token AccessToken

	err := r.db.QueryRow(`
		SELECT
			user_id,
			client_id,
			scope,
			expires_at
		FROM oauth_access_tokens
		WHERE token_hash = $1
			AND revoked_at IS NULL
			AND expires_at > CURRENT_TIMESTAMP
	`,
		tokenHash,
	).Scan(
		&token.UserID,
		&token.ClientID,
		&token.Scope,
		&token.ExpiresAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New(
				"invalid access token",
			)
		}

		return nil, err
	}

	return &token, nil
}

// ============================================================
// Revoke Access Token
// ============================================================

func (r *OIDCRepository) RevokeAccessToken(
	tokenHash string,
) error {

	_, err := r.db.Exec(`
		UPDATE oauth_access_tokens
		SET revoked_at = CURRENT_TIMESTAMP
		WHERE token_hash = $1
			AND revoked_at IS NULL
	`,
		tokenHash,
	)

	return err
}

// ============================================================
// User Handle
//
// Returns primary .nid handle.
//
// Fallback:
// oldest handle.
// ============================================================

func (r *OIDCRepository) GetPrimaryHandleByUserID(
	userID string,
) (string, error) {

	var handle string

	// --------------------------------------------------------
	// First: primary handle
	// --------------------------------------------------------

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

	// --------------------------------------------------------
	// Fallback: oldest handle
	// --------------------------------------------------------

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
		if errors.Is(err, sql.ErrNoRows) {
			return "", errors.New(
				"no handle found for user",
			)
		}

		return "", err
	}

	return handle, nil
}


func (r *OIDCRepository) GetClientInfo(clientID string) (*dto.ClientInfoResponse, error) {
    var client dto.ClientInfoResponse

    err := r.db.QueryRow(`
        SELECT client_name, client_logo, client_uri, policy_uri
        FROM oauth_clients
        WHERE client_id = $1
    `, clientID).Scan(
        &client.ClientName,
        &client.ClientLogo,
        &client.ClientURI,
        &client.PolicyURI,
    )

    if err != nil {
        return nil, err
    }

    return &client, nil
}
func (r *OIDCRepository) CreateOAuthSession(
    userID string,
    clientID string,
    expiresAt time.Time,
) (string, error) {

    var sessionID string

    err := r.db.QueryRowContext(
        context.Background(),
        `
        INSERT INTO oauth_sessions (
            user_id,
            client_id,
            status,
            expires_at
        )
        VALUES ($1, $2, 'active', $3)
        RETURNING id
        `,
        userID,
        clientID,
        expiresAt,
    ).Scan(&sessionID)

    if err != nil {
        return "", err
    }

    return sessionID, nil
}


func (r *OIDCRepository) TouchSession(
    tokenHash string,
) error {

    _, err := r.db.ExecContext(
        context.Background(),
        `
        UPDATE oauth_sessions
        SET last_used_at = CURRENT_TIMESTAMP
        WHERE id = (
            SELECT session_id
            FROM oauth_access_tokens
            WHERE token_hash = $1
        )
        `,
        tokenHash,
    )

    return err
}



func (r *OIDCRepository) ListAllByUser(
	userID string,
) ([]dto.OAuthClientInfo, error) {

	rows, err := r.db.QueryContext(
		context.Background(),
		`
		SELECT
			id,
			client_id,
			user_id,
			name,
			redirect_uri,
			client_type,
			client_logo,
			client_uri,
			policy_uri,
			created_at,
			updated_at
		FROM oauth_clients
		WHERE user_id = $1
		ORDER BY created_at DESC
		`,
		userID,
	)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	clients := make([]dto.OAuthClientInfo, 0)

	for rows.Next() {
		var client dto.OAuthClientInfo

		err := rows.Scan(
			&client.ID,
			&client.ClientID,
			&client.UserID,
			&client.ClientName,
			&client.RedirectURI,
			&client.ClientType,
			&client.ClientLogo,
			&client.ClientURI,
			&client.PolicyURI,
			&client.CreatedAt,
			&client.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		clients = append(clients, client)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return clients, nil
}

func (r *OIDCRepository) GetUserSessions(
    userID string,
) ([]models.OAuthSession, error) {

    rows, err := r.db.QueryContext(
        context.Background(),
        `
        SELECT
            s.id,
            s.user_id,
            s.client_id,
            s.status,
            s.created_at,
            s.expires_at,
            s.last_used_at,
            s.revoked_at
        FROM oauth_sessions s
        WHERE s.user_id = $1
        ORDER BY s.created_at DESC
        `,
        userID,
    )
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var sessions []models.OAuthSession

    for rows.Next() {
        var session models.OAuthSession

        err := rows.Scan(
            &session.ID,
            &session.UserID,
            &session.ClientID,
            &session.Status,
            &session.CreatedAt,
            &session.ExpiresAt,
            &session.LastUsedAt,
            &session.RevokedAt,
        )
        if err != nil {
            return nil, err
        }

        sessions = append(sessions, session)
    }

    return sessions, rows.Err()
}
// ============================================================
// Delete Client By Internal ID
// ============================================================

func (r *OIDCRepository) DeleteByID(id string) error {
    result, err := r.db.ExecContext(
        context.Background(),
        `DELETE FROM oauth_clients WHERE id = $1`,
        id,
    )
    if err != nil {
        return err
    }

    rowsAffected, err := result.RowsAffected()
    if err != nil {
        return err
    }

    if rowsAffected == 0 {
        return errors.New("client not found")
    }

    return nil
}

// ============================================================
// Rotate Client Secret By Internal ID
// ============================================================

func (r *OIDCRepository) RotateSecretByID(
    id string,
    newSecretHash string,
) error {

    result, err := r.db.ExecContext(
        context.Background(),
        `
        UPDATE oauth_clients
        SET client_secret_hash = $1,
            updated_at = CURRENT_TIMESTAMP
        WHERE id = $2
        `,
        newSecretHash,
        id,
    )
    if err != nil {
        return err
    }

    rowsAffected, err := result.RowsAffected()
    if err != nil {
        return err
    }

    if rowsAffected == 0 {
        return errors.New("client not found")
    }

    return nil
}
