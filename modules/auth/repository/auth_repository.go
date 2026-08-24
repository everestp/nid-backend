// modules/auth/repository/auth_repository.go
package repository

import (
	"database/sql"
)

type AuthRepository struct {
	db *sql.DB
}

func NewAuthRepository(db *sql.DB) *AuthRepository {
	return &AuthRepository{db: db}
}

func (r *AuthRepository) FindOrCreateUserByWallet(address, chain string) (string, error) {
	var userID string
	query := `SELECT user_id FROM wallets WHERE address = $1`
	err := r.db.QueryRow(query, address).Scan(&userID)
	if err == nil {
		return userID, nil
	}

	if err != sql.ErrNoRows {
		return "", err
	}

	tx, err := r.db.Begin()
	if err != nil {
		return "", err
	}
	defer tx.Rollback()

	err = tx.QueryRow(`INSERT INTO users DEFAULT VALUES RETURNING id`).Scan(&userID)
	if err != nil {
		return "", err
	}

	_, err = tx.Exec(`INSERT INTO wallets (user_id, chain, network, address, status) VALUES ($1, $2, 'mainnet', $3, 'verified')`, userID, chain, address)
	if err != nil {
		return "", err
	}

	return userID, tx.Commit()
}
