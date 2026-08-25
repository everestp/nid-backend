package repository

import (
	"database/sql"
	"encoding/json"
	"errors"


	"nid-backend/modules/social/models"
)

var (
	ErrSocialNotFound = errors.New("social identity not found")
)

type SocialRepository struct {
	db *sql.DB
}

func NewSocialRepository(db *sql.DB) *SocialRepository {
	return &SocialRepository{db: db}
}

// ============================================================================
// CREATE
// ============================================================================

func (r *SocialRepository) Create(
	userID string,
	platform string,
	handle string,
	normalizedHandle string,
	publiclyVisible bool,
	metadata map[string]interface{},
) (*models.SocialIdentity, error) {

	metadataJSON, err := json.Marshal(metadata)
	if err != nil {
		return nil, err
	}

	var social models.SocialIdentity
	var metadataBytes []byte

	err = r.db.QueryRow(`
		INSERT INTO social_identities (
			user_id,
			platform,
			handle,
			normalized_handle,
			publicly_visible,
			metadata
		)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING
			id,
			user_id,
			platform,
			handle,
			normalized_handle,
			verified,
			publicly_visible,
			metadata,
			created_at,
			updated_at
	`,
		userID,
		platform,
		handle,
		normalizedHandle,
		publiclyVisible,
		metadataJSON,
	).Scan(
		&social.ID,
		&social.UserID,
		&social.Platform,
		&social.Handle,
		&social.NormalizedHandle,
		&social.Verified,
		&social.PubliclyVisible,
		&metadataBytes,
		&social.CreatedAt,
		&social.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(metadataBytes, &social.Metadata); err != nil {
		return nil, err
	}

	return &social, nil
}

// ============================================================================
// FIND BY ID
// ============================================================================

func (r *SocialRepository) FindByID(
	id string,
) (*models.SocialIdentity, error) {

	var social models.SocialIdentity
	var metadataBytes []byte

	err := r.db.QueryRow(`
		SELECT
			id,
			user_id,
			platform,
			handle,
			normalized_handle,
			verified,
			publicly_visible,
			metadata,
			created_at,
			updated_at
		FROM social_identities
		WHERE id = $1
	`,
		id,
	).Scan(
		&social.ID,
		&social.UserID,
		&social.Platform,
		&social.Handle,
		&social.NormalizedHandle,
		&social.Verified,
		&social.PubliclyVisible,
		&metadataBytes,
		&social.CreatedAt,
		&social.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrSocialNotFound
		}

		return nil, err
	}

	if err := json.Unmarshal(metadataBytes, &social.Metadata); err != nil {
		return nil, err
	}

	return &social, nil
}

// ============================================================================
// FIND BY ID AND USER
// ============================================================================
//
// Used for authenticated operations.
// This guarantees the social identity belongs to the authenticated user.
//

func (r *SocialRepository) FindByIDAndUser(
	id string,
	userID string,
) (*models.SocialIdentity, error) {

	var social models.SocialIdentity
	var metadataBytes []byte

	err := r.db.QueryRow(`
		SELECT
			id,
			user_id,
			platform,
			handle,
			normalized_handle,
			verified,
			publicly_visible,
			metadata,
			created_at,
			updated_at
		FROM social_identities
		WHERE id = $1
		  AND user_id = $2
	`,
		id,
		userID,
	).Scan(
		&social.ID,
		&social.UserID,
		&social.Platform,
		&social.Handle,
		&social.NormalizedHandle,
		&social.Verified,
		&social.PubliclyVisible,
		&metadataBytes,
		&social.CreatedAt,
		&social.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrSocialNotFound
		}

		return nil, err
	}

	if err := json.Unmarshal(metadataBytes, &social.Metadata); err != nil {
		return nil, err
	}

	return &social, nil
}

// ============================================================================
// FIND ALL USER SOCIALS
// ============================================================================

func (r *SocialRepository) FindByUser(
	userID string,
) ([]models.SocialIdentity, error) {

	rows, err := r.db.Query(`
		SELECT
			id,
			user_id,
			platform,
			handle,
			normalized_handle,
			verified,
			publicly_visible,
			metadata,
			created_at,
			updated_at
		FROM social_identities
		WHERE user_id = $1
		ORDER BY created_at ASC
	`,
		userID,
	)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	socials := make([]models.SocialIdentity, 0)

	for rows.Next() {

		var social models.SocialIdentity
		var metadataBytes []byte

		err := rows.Scan(
			&social.ID,
			&social.UserID,
			&social.Platform,
			&social.Handle,
			&social.NormalizedHandle,
			&social.Verified,
			&social.PubliclyVisible,
			&metadataBytes,
			&social.CreatedAt,
			&social.UpdatedAt,
		)

		if err != nil {
			return nil, err
		}

		if err := json.Unmarshal(metadataBytes, &social.Metadata); err != nil {
			return nil, err
		}

		socials = append(socials, social)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return socials, nil
}

// ============================================================================
// FIND PUBLIC SOCIALS
// ============================================================================
//
// Used by:
// GET /users/:handle/socials
//
// Only publicly visible identities are returned.
//

func (r *SocialRepository) FindPublicByUser(
	userID string,
) ([]models.SocialIdentity, error) {

	rows, err := r.db.Query(`
		SELECT
			id,
			user_id,
			platform,
			handle,
			normalized_handle,
			verified,
			publicly_visible,
			metadata,
			created_at,
			updated_at
		FROM social_identities
		WHERE user_id = $1
		  AND publicly_visible = TRUE
		ORDER BY created_at ASC
	`,
		userID,
	)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	socials := make([]models.SocialIdentity, 0)

	for rows.Next() {

		var social models.SocialIdentity
		var metadataBytes []byte

		err := rows.Scan(
			&social.ID,
			&social.UserID,
			&social.Platform,
			&social.Handle,
			&social.NormalizedHandle,
			&social.Verified,
			&social.PubliclyVisible,
			&metadataBytes,
			&social.CreatedAt,
			&social.UpdatedAt,
		)

		if err != nil {
			return nil, err
		}

		if err := json.Unmarshal(metadataBytes, &social.Metadata); err != nil {
			return nil, err
		}

		socials = append(socials, social)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return socials, nil
}

// ============================================================================
// UPDATE
// ============================================================================
//
// Platform is intentionally NOT updated.
// A social identity belongs to a platform permanently.
//

func (r *SocialRepository) Update(
	id string,
	userID string,
	handle string,
	normalizedHandle string,
	publiclyVisible bool,
	metadata map[string]interface{},
) (*models.SocialIdentity, error) {

	metadataJSON, err := json.Marshal(metadata)
	if err != nil {
		return nil, err
	}

	var social models.SocialIdentity
	var metadataBytes []byte

	err = r.db.QueryRow(`
		UPDATE social_identities
		SET
			handle = $1,
			normalized_handle = $2,
			publicly_visible = $3,
			metadata = $4,
			updated_at = CURRENT_TIMESTAMP
		WHERE id = $5
		  AND user_id = $6
		RETURNING
			id,
			user_id,
			platform,
			handle,
			normalized_handle,
			verified,
			publicly_visible,
			metadata,
			created_at,
			updated_at
	`,
		handle,
		normalizedHandle,
		publiclyVisible,
		metadataJSON,
		id,
		userID,
	).Scan(
		&social.ID,
		&social.UserID,
		&social.Platform,
		&social.Handle,
		&social.NormalizedHandle,
		&social.Verified,
		&social.PubliclyVisible,
		&metadataBytes,
		&social.CreatedAt,
		&social.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrSocialNotFound
		}

		return nil, err
	}

	if err := json.Unmarshal(metadataBytes, &social.Metadata); err != nil {
		return nil, err
	}

	return &social, nil
}

// ============================================================================
// TOGGLE VISIBILITY
// ============================================================================

func (r *SocialRepository) ToggleVisibility(
	id string,
	userID string,
	publiclyVisible bool,
) (*models.SocialIdentity, error) {

	var social models.SocialIdentity
	var metadataBytes []byte

	err := r.db.QueryRow(`
		UPDATE social_identities
		SET
			publicly_visible = $1,
			updated_at = CURRENT_TIMESTAMP
		WHERE id = $2
		  AND user_id = $3
		RETURNING
			id,
			user_id,
			platform,
			handle,
			normalized_handle,
			verified,
			publicly_visible,
			metadata,
			created_at,
			updated_at
	`,
		publiclyVisible,
		id,
		userID,
	).Scan(
		&social.ID,
		&social.UserID,
		&social.Platform,
		&social.Handle,
		&social.NormalizedHandle,
		&social.Verified,
		&social.PubliclyVisible,
		&metadataBytes,
		&social.CreatedAt,
		&social.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrSocialNotFound
		}

		return nil, err
	}

	if err := json.Unmarshal(metadataBytes, &social.Metadata); err != nil {
		return nil, err
	}

	return &social, nil
}

// ============================================================================
// DELETE
// ============================================================================

func (r *SocialRepository) Delete(
	id string,
	userID string,
) error {

	result, err := r.db.Exec(`
		DELETE FROM social_identities
		WHERE id = $1
		  AND user_id = $2
	`,
		id,
		userID,
	)

	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrSocialNotFound
	}

	return nil
}

// ============================================================================
// CHECK DUPLICATE SOCIAL IDENTITY
// ============================================================================
//
// Useful before creating a social identity.
// Example:
//
// user_id = xxx
// platform = github
// normalized_handle = everest
//
// prevents the same account from being added twice.
//

func (r *SocialRepository) Exists(
	userID string,
	platform string,
	normalizedHandle string,
) (bool, error) {

	var exists bool

	err := r.db.QueryRow(`
		SELECT EXISTS (
			SELECT 1
			FROM social_identities
			WHERE user_id = $1
			  AND platform = $2
			  AND normalized_handle = $3
		)
	`,
		userID,
		platform,
		normalizedHandle,
	).Scan(&exists)

	if err != nil {
		return false, err
	}

	return exists, nil
}
