// modules/resolution/repository/resolution_repository.go
package repository

import (
	"database/sql"
	"fmt"
)

type ResolutionRepository struct {
	db *sql.DB
}

func NewResolutionRepository(db *sql.DB) *ResolutionRepository {
	return &ResolutionRepository{db: db}
}

func (r *ResolutionRepository) ResolveAddress(handleName, chain string) (string, error) {
	query := `
		SELECT w.address
		FROM handles h
		JOIN wallets w ON h.user_id = w.user_id
		WHERE h.handle = $1 AND w.chain = $2 AND h.status = 'active'
	`
	var address string
	err := r.db.QueryRow(query, handleName, chain).Scan(&address)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", fmt.Errorf("no address found for handle %s on chain %s", handleName, chain)
		}
		return "", err
	}
	return address, nil
}
