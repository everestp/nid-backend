package service

import (
	"errors"
	"strings"

	"nid-backend/modules/wallet_list/dto"
	"nid-backend/modules/wallet_list/repository"
)

type WalletService struct {
	repository *repository.WalletRepository
}

func NewwalletListService(
	repository *repository.WalletRepository,
) *WalletService {

	return &WalletService{
		repository: repository,
	}
}

// ============================================================
// CREATE
// ============================================================

func (s *WalletService) Create(
	userID string,
	req dto.CreateWalletRequest,
) (*dto.WalletResponse, error) {

	if userID == "" {
		return nil, errors.New(
			"user id is required",
		)
	}

	req.Chain = strings.TrimSpace(
		strings.ToLower(req.Chain),
	)

	req.Network = strings.TrimSpace(
		strings.ToLower(req.Network),
	)

	req.Address = strings.TrimSpace(
		req.Address,
	)

	req.Status = strings.TrimSpace(
		strings.ToLower(req.Status),
	)

	if req.Chain == "" {
		return nil, errors.New(
			"chain is required",
		)
	}

	if req.Network == "" {
		return nil, errors.New(
			"network is required",
		)
	}

	if req.Address == "" {
		return nil, errors.New(
			"address is required",
		)
	}

	// Default status
	if req.Status == "" {
		req.Status = "verified"
	}

	if req.Status != "verified" &&
		req.Status != "pending" {

		return nil, errors.New(
			"invalid wallet status",
		)
	}

	return s.repository.Create(
		userID,
		req,
	)
}

// ============================================================
// GET ALL
// ============================================================

func (s *WalletService) GetAll(
	userID string,
) ([]dto.WalletResponse, error) {

	if userID == "" {
		return nil, errors.New(
			"user id is required",
		)
	}

	return s.repository.GetAll(
		userID,
	)
}

// ============================================================
// GET BY ID
// ============================================================

func (s *WalletService) GetByID(
	userID string,
	id string,
) (*dto.WalletResponse, error) {

	if userID == "" {
		return nil, errors.New(
			"user id is required",
		)
	}

	if id == "" {
		return nil, errors.New(
			"wallet id is required",
		)
	}

	return s.repository.GetByID(
		userID,
		id,
	)
}

// ============================================================
// UPDATE
// ============================================================

func (s *WalletService) Update(
	userID string,
	id string,
	req dto.UpdateWalletRequest,
) (*dto.WalletResponse, error) {

	if userID == "" {
		return nil, errors.New(
			"user id is required",
		)
	}

	if id == "" {
		return nil, errors.New(
			"wallet id is required",
		)
	}

	req.Chain = strings.TrimSpace(
		strings.ToLower(req.Chain),
	)

	req.Network = strings.TrimSpace(
		strings.ToLower(req.Network),
	)

	req.Address = strings.TrimSpace(
		req.Address,
	)

	req.Status = strings.TrimSpace(
		strings.ToLower(req.Status),
	)

	if req.Chain == "" {
		return nil, errors.New(
			"chain is required",
		)
	}

	if req.Network == "" {
		return nil, errors.New(
			"network is required",
		)
	}

	if req.Address == "" {
		return nil, errors.New(
			"address is required",
		)
	}

	if req.Status != "verified" &&
		req.Status != "pending" {

		return nil, errors.New(
			"invalid wallet status",
		)
	}

	return s.repository.Update(
		userID,
		id,
		req,
	)
}

// ============================================================
// DELETE
// ============================================================

func (s *WalletService) Delete(
	userID string,
	id string,
) error {

	if userID == "" {
		return errors.New(
			"user id is required",
		)
	}

	if id == "" {
		return errors.New(
			"wallet id is required",
		)
	}

	return s.repository.Delete(
		userID,
		id,
	)
}
