// modules/user/repository/user_repository.go
package repository

import (
	"database/sql"
	"encoding/json"

	"nid-backend/models"
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}


// GetUserDashboard retrieves all dashboard data for a given userID in a single query.
func (r *UserRepository) GetUserDashboard(userID string) (*models.UserDashboard, error) {
	query := `
		SELECT
			u.id,
			u.created_at,

			-- Fetch all handles
			COALESCE(
				(SELECT json_agg(json_build_object(
					'id', h.id,
					'handle', h.handle,
					'is_primary', h.is_primary,
					'status', h.status
				) ORDER BY h.is_primary DESC, h.created_at ASC)
				FROM handles h
				WHERE h.user_id = u.id),
				'[]'::json
			) AS handles,

			-- Fetch social identities
			COALESCE(
				(SELECT json_agg(json_build_object(
					'id', s.id,
					'platform', s.platform,
					'handle', s.handle,
					'verified', s.verified,
					'publicly_visible', s.publicly_visible
				) ORDER BY s.verified DESC, s.created_at DESC)
				FROM social_identities s
				WHERE s.user_id = u.id),
				'[]'::json
			) AS socials,

			-- Fetch wallets
			COALESCE(
				(SELECT json_agg(json_build_object(
					'id', w.id,
					'chain', w.chain,
					'network', w.network,
					'address', w.address,
					'status', w.status
				) ORDER BY w.linked_at DESC)
				FROM wallet_list w
				WHERE w.user_id = u.id),
				'[]'::json
			) AS wallets,

			-- Fetch active sessions
			COALESCE(
				(SELECT json_agg(json_build_object(
					'id', os.id,
					'client_id', os.client_id,
					'client_name', os.client_name,
					'last_used_at', os.last_used_at,
					'created_at', os.created_at
				) ORDER BY COALESCE(os.last_used_at, os.created_at) DESC)
				FROM oauth_sessions os
				WHERE os.user_id = u.id AND os.status = 'active'),
				'[]'::json
			) AS active_sessions

		FROM users u
		WHERE u.id = $1;
	`

	dash := &models.UserDashboard{}
	var handlesJSON, socialsJSON, walletsJSON, sessionsJSON []byte

	err := r.db.QueryRow(query, userID).Scan(
		&dash.UserID,
		&dash.CreatedAt,
		&handlesJSON,
		&socialsJSON,
		&walletsJSON,
		&sessionsJSON,
	)
	if err != nil {
		return nil, err
	}

	// Unmarshal JSON aggregations into Go structs
	if err := json.Unmarshal(handlesJSON, &dash.Handles); err != nil {
		return nil, err
	}
	if err := json.Unmarshal(socialsJSON, &dash.Socials); err != nil {
		return nil, err
	}
	if err := json.Unmarshal(walletsJSON, &dash.Wallets); err != nil {
		return nil, err
	}
	if err := json.Unmarshal(sessionsJSON, &dash.ActiveSessions); err != nil {
		return nil, err
	}

	return dash, nil
}
// FindUserByID returns a user by ID.
func (r *UserRepository) FindUserByID(userID string) (*models.User, error) {
	query := `
		SELECT id, created_at
		FROM users
		WHERE id = $1
	`

	u := &models.User{}

	err := r.db.QueryRow(query, userID).Scan(
		&u.ID,
		&u.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	return u, nil
}
func (r *UserRepository) GetCurrentLoggedInUser(userID string) (*models.User, error) {
	query := `
		SELECT id, created_at
		FROM users
		WHERE id = $1
	`

	u := &models.User{}

	err := r.db.QueryRow(query, userID).Scan(
		&u.ID,
		&u.CreatedAt,
	)

	if err != nil {
		return nil, err
	}

	return u, nil
}
// FindHandlesByUserID returns all active handles for a user.
func (r *UserRepository) FindHandlesByUserID(userID string) ([]string, error) {
	query := `
		SELECT handle
		FROM handles
		WHERE user_id = $1
		  AND status = 'active'
	`

	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var handles []string

	for rows.Next() {
		var name string

		if err := rows.Scan(&name); err != nil {
			return nil, err
		}

		handles = append(handles, name)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return handles, nil
}

// FindWalletsByUserID returns all wallets for a user.
func (r *UserRepository) FindWalletsByUserID(userID string) ([]models.Wallet, error) {
	query := `
		SELECT
			id,
			user_id,
			chain,
			network,
			address,
			status,
			linked_at
		FROM wallets
		WHERE user_id = $1
	`

	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var wallets []models.Wallet

	for rows.Next() {
		var w models.Wallet

		if err := rows.Scan(
			&w.ID,
			&w.UserID,
			&w.Chain,
			&w.Network,
			&w.Address,
			&w.Status,
			&w.LinkedAt,
		); err != nil {
			return nil, err
		}

		wallets = append(wallets, w)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return wallets, nil
}

// GetPublicProfileByHandle returns the public profile associated
// with an active handle.
func (r *UserRepository) GetPublicProfileByHandle(handle string) (*models.PublicProfile, error) {

	// ------------------------------------------------------------------
	// 1. Find user by handle
	// ------------------------------------------------------------------

	var userID string

	err := r.db.QueryRow(`
		SELECT user_id
		FROM handles
		WHERE handle = $1
		  AND status = 'active'
	`, handle).Scan(&userID)

	if err != nil {
		return nil, err
	}

	// ------------------------------------------------------------------
	// 2. Get user
	// ------------------------------------------------------------------

	user := &models.User{}

	err = r.db.QueryRow(`
		SELECT id, created_at
		FROM users
		WHERE id = $1
	`, userID).Scan(
		&user.ID,
		&user.CreatedAt,
	)

	if err != nil {
		return nil, err
	}

	// ------------------------------------------------------------------
	// 3. Get all active handles
	// ------------------------------------------------------------------

	rows, err := r.db.Query(`
		SELECT handle, is_primary
		FROM handles
		WHERE user_id = $1
		  AND status = 'active'
		ORDER BY is_primary DESC, created_at ASC
	`, userID)

	if err != nil {
		return nil, err
	}

	var handles []models.HandleInfo

	for rows.Next() {
		var h models.HandleInfo

		if err := rows.Scan(
			&h.Handle,
			&h.IsPrimary,
		); err != nil {
			rows.Close()
			return nil, err
		}

		handles = append(handles, h)
	}

	if err := rows.Err(); err != nil {
		rows.Close()
		return nil, err
	}

	rows.Close()

	// ------------------------------------------------------------------
	// 4. Get publicly visible social identities
	// ------------------------------------------------------------------

	rows, err = r.db.Query(`
		SELECT
			platform,
			handle,
			verified,
			metadata
		FROM social_identities
		WHERE user_id = $1
		  AND publicly_visible = TRUE
		ORDER BY platform, handle
	`, userID)

	if err != nil {
		return nil, err
	}

	var identities []models.SocialIdentity

	for rows.Next() {
		var s models.SocialIdentity

		// PostgreSQL JSONB is returned by database/sql as []byte.
		var metadata []byte

		if err := rows.Scan(
			&s.Platform,
			&s.Handle,
			&s.Verified,
			&metadata,
		); err != nil {
			rows.Close()
			return nil, err
		}

		// Convert JSONB []byte into map[string]interface{}.
		if err := json.Unmarshal(metadata, &s.Metadata); err != nil {
			rows.Close()
			return nil, err
		}

		identities = append(identities, s)
	}

	if err := rows.Err(); err != nil {
		rows.Close()
		return nil, err
	}

	rows.Close()

	// ------------------------------------------------------------------
	// 5. Get verified wallets
	// ------------------------------------------------------------------

	rows, err = r.db.Query(`
		SELECT
			chain,
			network,
			address,
			linked_at
		FROM wallet_list
		WHERE user_id = $1
		  AND status = 'verified'
		ORDER BY linked_at ASC
	`, userID)

	if err != nil {
		return nil, err
	}

	var wallets []models.WalletList

	for rows.Next() {
		var w models.WalletList

		if err := rows.Scan(
			&w.Chain,
			&w.Network,
			&w.Address,
			&w.LinkedAt,
		); err != nil {
			rows.Close()
			return nil, err
		}

		wallets = append(wallets, w)
	}

	if err := rows.Err(); err != nil {
		rows.Close()
		return nil, err
	}

	rows.Close()

	// ------------------------------------------------------------------
	// 6. Build public profile
	// ------------------------------------------------------------------

	return &models.PublicProfile{
		User:       user,
		Handles:    handles,
		Identities: identities,
		Wallets:    wallets,
	}, nil
}
