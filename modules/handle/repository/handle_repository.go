package repository

import (
	"database/sql"
	"errors"
	"time"
)

type HandleModel struct {
	ID        string
	UserID    string
	Name      string
	Status    string
	Primary   bool
	CreatedAt time.Time
}

type HandleRepository struct {
	db *sql.DB
}

func NewHandleRepository(db *sql.DB) *HandleRepository {
	return &HandleRepository{db: db}
}

// ClaimHandleForWallet handles the public homepage registration:
// Creates user, links wallet, and claims the handle atomically.
func (r *HandleRepository) ClaimHandleForWallet(name, address, chain string) (*HandleModel, error) {
	// Check if handle is already taken
	var exists bool
	err := r.db.QueryRow(`SELECT EXISTS(SELECT 1 FROM handles WHERE handle = $1)`, name).Scan(&exists)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.New("this handle is already taken")
	}

	// Check if wallet is already registered
	var userID string
	err = r.db.QueryRow(`SELECT user_id FROM wallets WHERE address = $1`, address).Scan(&userID)

	tx, errTx := r.db.Begin()
	if errTx != nil {
		return nil, errTx
	}
	defer tx.Rollback()

	if err == sql.ErrNoRows {
		// New user registration
		err = tx.QueryRow(`INSERT INTO users DEFAULT VALUES RETURNING id`).Scan(&userID)
		if err != nil {
			return nil, err
		}

		_, err = tx.Exec(`INSERT INTO wallets (user_id, chain, network, address, status) VALUES ($1, $2, 'mainnet', $3, 'verified')`, userID, chain, address)
		if err != nil {
			return nil, err
		}
	} else if err != nil {
		return nil, err
	}

	// Insert Handle
	var handle HandleModel
	handle.UserID = userID
	handle.Name = name
	handle.Status = "active"
	handle.Primary = true

	err = tx.QueryRow(
		`INSERT INTO handles (user_id, handle, status, is_primary) VALUES ($1, $2, $3, $4) RETURNING id, created_at`,
		handle.UserID, handle.Name, handle.Status, handle.Primary,
	).Scan(&handle.ID, &handle.CreatedAt)
	if err != nil {
		return nil, errors.New("failed to insert handle, it may already be registered")
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return &handle, nil
}
// GetAllByUserID returns all handles owned by a user.
func (r *HandleRepository) GetAllByUserID(userID string) ([]*HandleModel, error) {
	rows, err := r.db.Query(`
		SELECT id, user_id, handle, status, is_primary, created_at
		FROM handles
		WHERE user_id = $1
		ORDER BY is_primary DESC, created_at ASC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var handles []*HandleModel

	for rows.Next() {
		var handle HandleModel

		err := rows.Scan(
			&handle.ID,
			&handle.UserID,
			&handle.Name,
			&handle.Status,
			&handle.Primary,
			&handle.CreatedAt,
		)
		if err != nil {
			return nil, err
		}

		handles = append(handles, &handle)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return handles, nil
}
