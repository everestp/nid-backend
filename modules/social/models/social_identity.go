package models

import (
	"time"

	"github.com/google/uuid"
)

type SocialIdentity struct {
	ID                 uuid.UUID              `json:"id"`
	UserID             uuid.UUID              `json:"user_id"`
	Platform           string                 `json:"platform"`
	Handle             string                 `json:"handle"`
	NormalizedHandle   string                 `json:"normalized_handle"`
	Verified           bool                   `json:"verified"`
	PubliclyVisible    bool                   `json:"publicly_visible"`
	Metadata           map[string]interface{} `json:"metadata"`
	CreatedAt          time.Time              `json:"created_at"`
	UpdatedAt          time.Time              `json:"updated_at"`
}
