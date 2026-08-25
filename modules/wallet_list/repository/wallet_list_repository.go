package repository

import (
	"database/sql"
	"errors"

	"nid-backend/modules/wallet_list/dto"
)

type WalletRepository struct {
	db *sql.DB
}

func NewwalletListRepository(db *sql.DB) *WalletRepository {
	return &WalletRepository{
		db: db,
	}
}

// ============================================================
// CREATE
// ============================================================

func (r *WalletRepository) Create(
	userID string,
	req dto.CreateWalletRequest,
) (*dto.WalletResponse, error) {

	var wallet dto.WalletResponse

	query := `
		INSERT INTO wallet_list (
			user_id,
			chain,
			network,
			address,
			status
		)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING
			id,
			chain,
			network,
			address,
			status,
			linked_at
	`

	err := r.db.QueryRow(
		query,
		userID,
		req.Chain,
		req.Network,
		req.Address,
		req.Status,
	).Scan(
		&wallet.ID,
		&wallet.Chain,
		&wallet.Network,
		&wallet.Address,
		&wallet.Status,
		&wallet.LinkedAt,
	)

	if err != nil {
		return nil, err
	}

	return &wallet, nil
}

// ============================================================
// GET ALL USER WALLETS
// ============================================================

func (r *WalletRepository) GetAll(
	userID string,
) ([]dto.WalletResponse, error) {

	query := `
		SELECT
			id,
			chain,
			network,
			address,
			status,
			linked_at
		FROM wallet_list
		WHERE user_id = $1
		ORDER BY linked_at DESC
	`

	rows, err := r.db.Query(query, userID)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	wallets := make(
		[]dto.WalletResponse,
		0,
	)

	for rows.Next() {

		var wallet dto.WalletResponse

		err := rows.Scan(
			&wallet.ID,
			&wallet.Chain,
			&wallet.Network,
			&wallet.Address,
			&wallet.Status,
			&wallet.LinkedAt,
		)

		if err != nil {
			return nil, err
		}

		wallets = append(
			wallets,
			wallet,
		)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return wallets, nil
}

// ============================================================
// GET BY ID
// ============================================================

func (r *WalletRepository) GetByID(
	userID string,
	id string,
) (*dto.WalletResponse, error) {

	var wallet dto.WalletResponse

	query := `
		SELECT
			id,
			chain,
			network,
			address,
			status,
			linked_at
		FROM wallet_list
		WHERE id = $1
		  AND user_id = $2
	`

	err := r.db.QueryRow(
		query,
		id,
		userID,
	).Scan(
		&wallet.ID,
		&wallet.Chain,
		&wallet.Network,
		&wallet.Address,
		&wallet.Status,
		&wallet.LinkedAt,
	)

	if err != nil {

		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New(
				"wallet not found",
			)
		}

		return nil, err
	}

	return &wallet, nil
}

// ============================================================
// UPDATE
// ============================================================

func (r *WalletRepository) Update(
	userID string,
	id string,
	req dto.UpdateWalletRequest,
) (*dto.WalletResponse, error) {

	var wallet dto.WalletResponse

	query := `
		UPDATE wallet_list
		SET
			chain = $1,
			network = $2,
			address = $3,
			status = $4
		WHERE id = $5
		  AND user_id = $6
		RETURNING
			id,
			chain,
			network,
			address,
			status,
			linked_at
	`

	err := r.db.QueryRow(
		query,
		req.Chain,
		req.Network,
		req.Address,
		req.Status,
		id,
		userID,
	).Scan(
		&wallet.ID,
		&wallet.Chain,
		&wallet.Network,
		&wallet.Address,
		&wallet.Status,
		&wallet.LinkedAt,
	)

	if err != nil {

		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New(
				"wallet not found",
			)
		}

		return nil, err
	}

	return &wallet, nil
}

// ============================================================
// DELETE
// ============================================================

func (r *WalletRepository) Delete(
	userID string,
	id string,
) error {

	query := `
		DELETE FROM wallet_list
		WHERE id = $1
		  AND user_id = $2
	`

	result, err := r.db.Exec(
		query,
		id,
		userID,
	)

	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()

	if err != nil {
		return err
	}

	if rows == 0 {
		return errors.New(
			"wallet not found",
		)
	}

	return nil
}
