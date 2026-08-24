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
	query := `SELECT name FROM handles WHERE user_id = $1 AND status = 'active'`
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
