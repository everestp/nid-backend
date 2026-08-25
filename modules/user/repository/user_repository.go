// modules/user/repository/user_repository.go
package repository

import (
	"database/sql"
	"nid-backend/models"
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) FindUserByID(userID string) (*models.User, error) {
	query := `SELECT id, created_at FROM users WHERE id = $1`
	u := &models.User{}
	err := r.db.QueryRow(query, userID).Scan(&u.ID, &u.CreatedAt)
	if err != nil {
		return nil, err
	}
	return u, nil
}

func (r *UserRepository) FindHandlesByUserID(userID string) ([]string, error) {
	query := `SELECT handle FROM handles WHERE user_id = $1 AND status = 'active'`
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
	return handles, nil
}

func (r *UserRepository) FindWalletsByUserID(userID string) ([]models.Wallet, error) {
	query := `SELECT id, user_id, chain, network, address, status, linked_at FROM wallets WHERE user_id = $1`
	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var wallets []models.Wallet
	for rows.Next() {
		var w models.Wallet
		if err := rows.Scan(&w.ID, &w.UserID, &w.Chain, &w.Network, &w.Address, &w.Status, &w.LinkedAt); err != nil {
			return nil, err
		}
		wallets = append(wallets, w)
	}
	return wallets, nil
}
func (r *UserRepository) GetPublicProfileByHandle(handle string) (*models.PublicProfile, error) {
	var userID string
	var isPrimary bool

	err := r.db.QueryRow(`
		SELECT user_id, is_primary
		FROM handles
		WHERE handle = $1 AND status = 'active'
	`, handle).Scan(&userID, &isPrimary)
	if err != nil {
		return nil, err
	}

	user := &models.User{}
	err = r.db.QueryRow(`
		SELECT id, created_at
		FROM users
		WHERE id = $1
	`, userID).Scan(&user.ID, &user.CreatedAt)
	if err != nil {
		return nil, err
	}

	// All active handles
	rows, err := r.db.Query(`
		SELECT handle, is_primary
		FROM handles
		WHERE user_id = $1 AND status = 'active'
		ORDER BY is_primary DESC, created_at ASC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var handles []models.HandleInfo
	for rows.Next() {
		var h models.HandleInfo
		if err := rows.Scan(&h.Handle, &h.IsPrimary); err != nil {
			return nil, err
		}
		handles = append(handles, h)
	}

	// Public social identities
	rows, err = r.db.Query(`
		SELECT platform, handle, verified, metadata
		FROM social_identities
		WHERE user_id = $1 AND publicly_visible = TRUE
		ORDER BY platform, handle
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var identities []models.SocialIdentity
	for rows.Next() {
		var s models.SocialIdentity
		if err := rows.Scan(&s.Platform, &s.Handle, &s.Verified, &s.Metadata); err != nil {
			return nil, err
		}
		identities = append(identities, s)
	}

	// Verified wallets
	rows, err = r.db.Query(`
		SELECT chain, network, address, linked_at
		FROM wallet_list
		WHERE user_id = $1 AND status = 'verified'
		ORDER BY linked_at ASC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var wallets []models.WalletList
	for rows.Next() {
		var w models.WalletList
		if err := rows.Scan(&w.Chain, &w.Network, &w.Address, &w.LinkedAt); err != nil {
			return nil, err
		}
		wallets = append(wallets, w)
	}

	return &models.PublicProfile{
		User:       user,
		Handles:    handles,
		Identities: identities,
		Wallets:    wallets,
	}, nil
}
