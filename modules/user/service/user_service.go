// modules/user/service/user_service.go
package service

import (
	"errors"
	"nid-backend/modules/user/repository"
	"nid-backend/modules/user/dto"
)

type UserService struct {
	repo *repository.UserRepository
}

func NewUserService(repo *repository.UserRepository) *UserService {
	return &UserService{repo: repo}
}

func (s *UserService) GetProfile(userID string) (*dto.UserProfileResponse, error) {
	user, err := s.repo.FindUserByID(userID)
	if err != nil {
		return nil, errors.New("user not found")
	}

	handles, err := s.repo.FindHandlesByUserID(userID)
	if err != nil {
		return nil, err
	}

	wallets, err := s.repo.FindWalletsByUserID(userID)
	if err != nil {
		return nil, err
	}

	var walletInfos []dto.WalletInfo
	for _, w := range wallets {
		walletInfos = append(walletInfos, dto.WalletInfo{
			Chain:   w.Chain,
			Address: w.Address,
			Network: w.Network,
		})
	}

	return &dto.UserProfileResponse{
		ID:        user.ID,
		CreatedAt: user.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		Handles:   handles,
		Wallets:   walletInfos,
	}, nil
}
