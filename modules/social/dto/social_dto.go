package dto

// ============================================================================
// CREATE SOCIAL IDENTITY
// ============================================================================
//
// Used when adding a new social identity.
//
// Example:
//
// {
//     "platform": "github",
//     "handle": "everestpaudel",
//     "publicly_visible": true,
//     "metadata": {
//         "url": "https://github.com/everestpaudel"
//     }
// }
//
// ============================================================================

type CreateSocialRequest struct {
	Platform        string                 `json:"platform"`
	Handle          string                 `json:"handle"`
	PubliclyVisible bool                   `json:"publicly_visible"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
}

// ============================================================================
// UPDATE SOCIAL IDENTITY
// ============================================================================
//
// All fields are optional.
//
// Only supplied fields are updated.
//
// Platform cannot be changed.
// Verified cannot be changed through normal CRUD.
//
// ============================================================================

type UpdateSocialRequest struct {
	Handle          *string                `json:"handle,omitempty"`
	PubliclyVisible *bool                  `json:"publicly_visible,omitempty"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
}

// ============================================================================
// TOGGLE PUBLIC VISIBILITY
// ============================================================================
//
// Used specifically by:
//
// PATCH /api/v1/social/{id}/visibility
//
// Example:
//
// {
//     "publicly_visible": true
// }
//
// ============================================================================

type ToggleVisibilityRequest struct {
	PubliclyVisible bool `json:"publicly_visible"`
}

// ============================================================================
// SOCIAL RESPONSE
// ============================================================================
//
// Returned to the authenticated owner.
//
// Verified is response-only.
// The client cannot modify it through Create/Update.
//
// ============================================================================

type SocialResponse struct {
	ID              string                 `json:"id"`
	Platform        string                 `json:"platform"`
	Handle          string                 `json:"handle"`
	Verified        bool                   `json:"verified"`
	PubliclyVisible bool                   `json:"publicly_visible"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt       string                 `json:"created_at"`
	UpdatedAt       string                 `json:"updated_at"`
}

// ============================================================================
// SOCIAL LIST RESPONSE
// ============================================================================
//
// Example:
//
// {
//     "socials": [
//         {
//             "id": "uuid",
//             "platform": "github",
//             "handle": "everestpaudel",
//             "verified": true,
//             "publicly_visible": true,
//             "metadata": {},
//             "created_at": "2026-08-25T10:00:00Z",
//             "updated_at": "2026-08-25T10:00:00Z"
//         }
//     ],
//     "count": 1
// }
//
// ============================================================================

type SocialListResponse struct {
	Socials []SocialResponse `json:"socials"`
	Count   int              `json:"count"`
}
