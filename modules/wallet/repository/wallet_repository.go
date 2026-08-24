// modules/wallet/repository/wallet_repository.go
package repository

import (
	"database/sql"
	"nid-backend/models"
)

type WalletRepository struct {
	db *sql.DB
}

func NewWalletRepository(db *sql.DB) *WalletRepository {
	return &WalletRepository{db: db}
}

func (r *WalletRepository) Create(userID, chain, network, address string) (*models.Wallet, error) {
	query := `INSERT INTO wallets (user_id, chain, network, address, status) VALUES ($1, $2, $3, $4, 'verified') RETURNING id, user_id, chain, network, address, status, linked_at`
	w := &models.Wallet{}
	err := r.db.QueryRow(query, userID, chain, network, address).Scan(&w.ID, &w.UserID, &w.Chain, &w.Network, &w.Address, &w.Status, &w.LinkedAt)
	if err != nil {
		return nil, err
	}
	return w, nil
}

func (r *WalletRepository) FindByAddress(address string) (*models.Wallet, error) {
	query := `SELECT id, user_id, chain, network, address, status, linked_at FROM wallets WHERE address = $1`
	w := &models.Wallet{}
	err := r.db.QueryRow(query, address).Scan(&w.ID, &w.UserID, &w.Chain, &w.Network, &w.Address, &w.Status, &w.LinkedAt)
	if err != nil {
		return nil, err
	}
	return w, nil
}
