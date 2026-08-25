package service

import (
	"errors"
	"strings"

	"nid-backend/modules/social/dto"
	"nid-backend/modules/social/models"
	"nid-backend/modules/social/repository"
)

// ============================================================================
// ERRORS
// ============================================================================

var (
	ErrInvalidPlatform = errors.New("invalid social platform")
	ErrInvalidHandle   = errors.New("invalid social handle")
	ErrSocialExists    = errors.New("social identity already exists")
	ErrSocialNotFound  = repository.ErrSocialNotFound
)

// ============================================================================
// SERVICE
// ============================================================================

type SocialService struct {
	repository *repository.SocialRepository
}

func NewSocialService(
	repository *repository.SocialRepository,
) *SocialService {
	return &SocialService{
		repository: repository,
	}
}

// ============================================================================
// SUPPORTED PLATFORMS
// ============================================================================

var allowedPlatforms = map[string]bool{

	// Social
	"twitter":   true,
	"instagram": true,
	"facebook":  true,
	"tiktok":    true,
	"linkedin":  true,
	"threads":   true,
	"bluesky":   true,
	"mastodon":  true,
	"snapchat":  true,
	"pinterest": true,
	"reddit":    true,

	// Web3
	"farcaster": true,
	"lens":      true,
	"warpcast":  true,
	"mirror":    true,
	"zora":      true,
	"opensea":   true,
	"ens":       true,

	// Developer
	"github":        true,
	"gitlab":        true,
	"bitbucket":     true,
	"stackoverflow": true,
	"codepen":       true,

	// Messaging / Community
	"discord":  true,
	"telegram": true,
	"whatsapp": true,
	"signal":   true,
	"twitch":   true,

	// Content
	"youtube":  true,
	"medium":   true,
	"substack": true,

	// Contact
	"email":   true,
	"phone":   true,
	"website": true,
}

// ============================================================================
// CREATE
// ============================================================================

func (s *SocialService) Create(
	userID string,
	req dto.CreateSocialRequest,
) (*models.SocialIdentity, error) {

	userID = strings.TrimSpace(userID)

	if userID == "" {
		return nil, errors.New("user id is required")
	}

	platform := strings.ToLower(
		strings.TrimSpace(req.Platform),
	)

	handle := strings.TrimSpace(req.Handle)

	// ------------------------------------------------------------
	// Validate platform
	// ------------------------------------------------------------

	if !allowedPlatforms[platform] {
		return nil, ErrInvalidPlatform
	}

	// ------------------------------------------------------------
	// Validate handle
	// ------------------------------------------------------------

	if handle == "" {
		return nil, ErrInvalidHandle
	}

	// ------------------------------------------------------------
	// Normalize handle
	// ------------------------------------------------------------

	normalizedHandle := normalizeHandle(handle)

	if normalizedHandle == "" {
		return nil, ErrInvalidHandle
	}

	// ------------------------------------------------------------
	// Metadata
	// ------------------------------------------------------------

	if req.Metadata == nil {
		req.Metadata = map[string]interface{}{}
	}

	// ------------------------------------------------------------
	// Check duplicate
	// ------------------------------------------------------------

	exists, err := s.repository.Exists(
		userID,
		platform,
		normalizedHandle,
	)
	if err != nil {
		return nil, err
	}

	if exists {
		return nil, ErrSocialExists
	}

	// ------------------------------------------------------------
	// Create
	// ------------------------------------------------------------

	return s.repository.Create(
		userID,
		platform,
		handle,
		normalizedHandle,
		req.PubliclyVisible,
		req.Metadata,
	)
}

// ============================================================================
// GET MY SOCIAL IDENTITIES
// ============================================================================

func (s *SocialService) GetMySocials(
	userID string,
) ([]models.SocialIdentity, error) {

	userID = strings.TrimSpace(userID)

	if userID == "" {
		return nil, errors.New("user id is required")
	}

	return s.repository.FindByUser(userID)
}

// ============================================================================
// GET PUBLIC SOCIAL IDENTITIES
// ============================================================================

func (s *SocialService) GetPublicSocials(
	userID string,
) ([]models.SocialIdentity, error) {

	userID = strings.TrimSpace(userID)

	if userID == "" {
		return nil, errors.New("user id is required")
	}

	return s.repository.FindPublicByUser(userID)
}

// ============================================================================
// GET SINGLE SOCIAL IDENTITY
// ============================================================================

func (s *SocialService) GetByID(
	userID string,
	socialID string,
) (*models.SocialIdentity, error) {

	userID = strings.TrimSpace(userID)
	socialID = strings.TrimSpace(socialID)

	if userID == "" {
		return nil, errors.New("user id is required")
	}

	if socialID == "" {
		return nil, ErrSocialNotFound
	}

	return s.repository.FindByIDAndUser(
		socialID,
		userID,
	)
}

// ============================================================================
// UPDATE
// ============================================================================
//
// Platform cannot be changed.
//
// User can update:
//   - handle
//   - publicly_visible
//   - metadata
//
// User CANNOT update:
//   - verified
//   - platform
//
// ============================================================================

func (s *SocialService) Update(
	userID string,
	socialID string,
	req dto.UpdateSocialRequest,
) (*models.SocialIdentity, error) {

	userID = strings.TrimSpace(userID)
	socialID = strings.TrimSpace(socialID)

	if userID == "" {
		return nil, errors.New("user id is required")
	}

	if socialID == "" {
		return nil, ErrSocialNotFound
	}

	// ------------------------------------------------------------
	// Verify ownership
	// ------------------------------------------------------------

	existing, err := s.repository.FindByIDAndUser(
		socialID,
		userID,
	)
	if err != nil {
		return nil, err
	}

	// ------------------------------------------------------------
	// Existing values
	// ------------------------------------------------------------

	handle := existing.Handle
	normalizedHandle := existing.NormalizedHandle
	publiclyVisible := existing.PubliclyVisible
	metadata := existing.Metadata

	// ------------------------------------------------------------
	// Update handle
	// ------------------------------------------------------------

	if req.Handle != nil {

		handle = strings.TrimSpace(
			*req.Handle,
		)

		if handle == "" {
			return nil, ErrInvalidHandle
		}

		normalizedHandle = normalizeHandle(handle)

		if normalizedHandle == "" {
			return nil, ErrInvalidHandle
		}
	}

	// ------------------------------------------------------------
	// Update visibility
	// ------------------------------------------------------------

	if req.PubliclyVisible != nil {
		publiclyVisible = *req.PubliclyVisible
	}

	// ------------------------------------------------------------
	// Update metadata
	// ------------------------------------------------------------

	if req.Metadata != nil {
		metadata = req.Metadata
	}

	// ------------------------------------------------------------
	// Check duplicate only when handle changed
	// ------------------------------------------------------------

	if normalizedHandle != existing.NormalizedHandle {

		exists, err := s.repository.Exists(
			userID,
			existing.Platform,
			normalizedHandle,
		)

		if err != nil {
			return nil, err
		}

		if exists {
			return nil, ErrSocialExists
		}
	}

	// ------------------------------------------------------------
	// Update repository
	// ------------------------------------------------------------

	return s.repository.Update(
		socialID,
		userID,
		handle,
		normalizedHandle,
		publiclyVisible,
		metadata,
	)
}

// ============================================================================
// TOGGLE VISIBILITY
// ============================================================================
//
// Explicitly receives the new visibility value.
//
// Example:
//
// publiclyVisible = true
//
// or
//
// publiclyVisible = false
//
// ============================================================================

func (s *SocialService) ToggleVisibility(
	userID string,
	socialID string,
	publiclyVisible bool,
) (*models.SocialIdentity, error) {

	userID = strings.TrimSpace(userID)
	socialID = strings.TrimSpace(socialID)

	if userID == "" {
		return nil, errors.New("user id is required")
	}

	if socialID == "" {
		return nil, ErrSocialNotFound
	}

	return s.repository.ToggleVisibility(
		socialID,
		userID,
		publiclyVisible,
	)
}

// ============================================================================
// DELETE
// ============================================================================

func (s *SocialService) Delete(
	userID string,
	socialID string,
) error {

	userID = strings.TrimSpace(userID)
	socialID = strings.TrimSpace(socialID)

	if userID == "" {
		return errors.New("user id is required")
	}

	if socialID == "" {
		return ErrSocialNotFound
	}

	return s.repository.Delete(
		socialID,
		userID,
	)
}

// ============================================================================
// NORMALIZE HANDLE
// ============================================================================

func normalizeHandle(
	handle string,
) string {

	handle = strings.TrimSpace(handle)

	// Remove @ prefix.
	//
	// @everest
	//      ↓
	// everest

	handle = strings.TrimPrefix(
		handle,
		"@",
	)

	return strings.ToLower(handle)
}
