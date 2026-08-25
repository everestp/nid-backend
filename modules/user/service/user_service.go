// modules/user/service/user_service.go
package service

import (
	"errors"
	"fmt"
	"nid-backend/modules/user/dto"
	"nid-backend/modules/user/repository"
	"time"
)

type UserService struct {
	repo *repository.UserRepository
}

func NewUserService(repo *repository.UserRepository) *UserService {
	return &UserService{repo: repo}
}

// GetUserDashboard retrieves and formats the complete single-glance dashboard data for a user.
func (s *UserService) GetUserDashboard(userID string) (*dto.UserDashboardResponse, error) {
	if userID == "" {
		return nil, errors.New("invalid user id")
	}

	dash, err := s.repo.GetUserDashboard(userID)
	if err != nil {
		return nil, errors.New("failed to fetch user dashboard data")
	}


	// Map repository model handles to DTO
	handles := make([]dto.HandleInfo, len(dash.Handles))
	for i, h := range dash.Handles {
		handles[i] = dto.HandleInfo{
			ID:        h.ID,
			Handle:    h.Handle,
			IsPrimary: h.IsPrimary,
			Status:    h.Status,
		}
	}

	// Map repository model socials to DTO
	socials := make([]dto.SocialInfo, len(dash.Socials))
	for i, soc := range dash.Socials {
		socials[i] = dto.SocialInfo{
			ID:              soc.ID,
			Platform:        soc.Platform,
			Handle:          soc.Handle,
			Verified:        soc.Verified,
			PubliclyVisible: soc.PubliclyVisible,
		}
	}

	// Map repository model wallets to DTO
	wallets := make([]dto.WalletDashboard, len(dash.Wallets))
	for i, w := range dash.Wallets {
		wallets[i] = dto.WalletDashboard{
			ID:      w.ID,
			Chain:   w.Chain,
			Network: w.Network,
			Address: w.Address,
			Status:  w.Status,
		}
	}

	// Map repository model active sessions to DTO with ISO 8601 formatting
	sessions := make([]dto.SessionInfo, len(dash.ActiveSessions))
	for i, sess := range dash.ActiveSessions {
		var lastUsedAtFormatted *string
		if sess.LastUsedAt != nil {
			formatted := sess.LastUsedAt.Format(time.RFC3339)
			lastUsedAtFormatted = &formatted
		}

		sessions[i] = dto.SessionInfo{
			ID:         sess.ID,
			ClientID:   sess.ClientID,
			ClientName: sess.ClientName,
			LastUsedAt: lastUsedAtFormatted,
			CreatedAt:  sess.CreatedAt.Format(time.RFC3339),
		}
	}

	return &dto.UserDashboardResponse{
		UserID:         dash.UserID,
		CreatedAt:      dash.CreatedAt.Format(time.RFC3339),
		Handles:        handles,
		Socials:        socials,
		Wallets:        wallets,
		ActiveSessions: sessions,
	}, nil
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
	fmt.Println("Thi is the profile",profile , err)
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
func (s *UserService) GetCurrentLoggedInUser(userID string) (*dto.UserProfileResponse, error) {
	user, err := s.repo.GetCurrentLoggedInUser(userID)
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
