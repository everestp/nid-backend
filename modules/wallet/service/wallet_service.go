// modules/wallet/service/wallet_service.go
package service

import (
	"errors"
	"nid-backend/models"
	"nid-backend/modules/wallet/repository"
)

type WalletService struct {
	repo *repository.WalletRepository
}

func NewWalletService(repo *repository.WalletRepository) *WalletService {
	return &WalletService{repo: repo}
}

func (s *WalletService) LinkWallet(userID, chain, network, address string) (*models.Wallet, error) {
	if address == "" || chain == "" {
		return nil, errors.New("invalid chain or address")
	}
	existing, _ := s.repo.FindByAddress(address)
	if existing != nil {
		return nil, errors.New("wallet already linked")
	}
	return s.repo.Create(userID, chain, network, address)
}
