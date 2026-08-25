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

func (s *UserService) GetPublicProfileByHandle(handle string) (*dto.PublicProfileResponse, error) {
	profile, err := s.repo.GetPublicProfileByHandle(handle)
	if err != nil {
		return nil, errors.New("profile not found")
	}

	var handles []dto.HandleInfo
	for _, h := range profile.Handles {
		handles = append(handles, dto.HandleInfo{
			Handle:    h.Handle,
			IsPrimary: h.IsPrimary,
		})
	}

	var identities []dto.SocialIdentityInfo
	for _, i := range profile.Identities {
		identities = append(identities, dto.SocialIdentityInfo{
			Platform: i.Platform,
			Handle:   i.Handle,
			Verified: i.Verified,
			Metadata: i.Metadata,
		})
	}

	var wallets []dto.WalletInfo
	for _, w := range profile.Wallets {
		wallets = append(wallets, dto.WalletInfo{
			Chain:   w.Chain,
			Network: w.Network,
			Address: w.Address,
		})
	}

	return &dto.PublicProfileResponse{
		ID:         profile.User.ID,
		CreatedAt:  profile.User.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		Handles:    handles,
		Identities: identities,
		Wallets:    wallets,
	}, nil
}
