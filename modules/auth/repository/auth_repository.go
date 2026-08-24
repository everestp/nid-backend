package repository

import (
	"database/sql"
	"errors"
)

type AuthRepository struct {
	db *sql.DB
}

func NewAuthRepository(db *sql.DB) *AuthRepository {
	return &AuthRepository{db: db}
}

// FindUserByHandleAndWallet strictly handles RETURNING users logging in:
// 1. Checks if the handle exists. If not, returns an error (NO CREATION).
// 2. Verifies if the incoming wallet address is authorized/linked to this handle.
func (r *AuthRepository) FindUserByHandleAndWallet(handle, address string) (string, error) {
	var userID string

	// 1. Check if the handle exists in the system
	err := r.db.QueryRow(`SELECT user_id FROM handles WHERE handle = $1`, handle).Scan(&userID)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", errors.New("handle not found, please claim it first on the homepage")
		}
		return "", err
	}

	// 2. Check if the incoming wallet address belongs to this user
	var walletMatch bool
	err = r.db.QueryRow(`SELECT EXISTS(SELECT 1 FROM wallets WHERE user_id = $1 AND address = $2)`, userID, address).Scan(&walletMatch)
	if err != nil {
		return "", err
	}

	if !walletMatch {
		return "", errors.New("this wallet address is not authorized for this handle")
	}

	// Successfully validated! Return the user ID so a session token can be issued.
	return userID, nil
}
