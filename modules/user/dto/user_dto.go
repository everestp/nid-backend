// modules/user/dto/user_dto.go
package dto

type UserProfileResponse struct {
	ID        string       `json:"id"`
	CreatedAt string       `json:"created_at"`
	Handles   []string     `json:"handles"`
	Wallets   []WalletInfo `json:"wallets"`
}

type WalletInfo struct {
	Chain   string `json:"chain"`
	Address string `json:"address"`
	Network string `json:"network"`
}



type SocialIdentityInfo struct {
	Platform string                 `json:"platform"`
	Handle   string                 `json:"handle"`
	Verified bool                   `json:"verified"`
	Metadata map[string]interface{} `json:"metadata"`
}

type PublicProfileResponse struct {
	ID         string               `json:"id"`
	CreatedAt  string               `json:"created_at"`
	Handles    []HandleInfo         `json:"handles"`
	Identities []SocialIdentityInfo `json:"identities"`
	Wallets    []WalletInfo         `json:"wallets"`
}
type UserDashboardResponse struct {
	UserID         string             `json:"user_id"`
	CreatedAt      string             `json:"created_at"`
	Handles        []HandleInfo       `json:"handles"`
	Socials        []SocialInfo       `json:"socials"`
	Wallets        []WalletDashboard  `json:"wallets"`
	ActiveSessions []SessionInfo      `json:"active_sessions"`
}

type HandleInfo struct {
	ID        string `json:"id"`
	Handle    string `json:"handle"`
	IsPrimary bool   `json:"is_primary"`
	Status    string `json:"status"`
}

type SocialInfo struct {
	ID              string `json:"id"`
	Platform        string `json:"platform"`
	Handle          string `json:"handle"`
	Verified        bool   `json:"verified"`
	PubliclyVisible bool   `json:"publicly_visible"`
}

type WalletDashboard struct {
	ID      string `json:"id"`
	Chain   string `json:"chain"`
	Network string `json:"network"`
	Address string `json:"address"`
	Status  string `json:"status"`
}

type SessionInfo struct {
	ID         string  `json:"id"`
	ClientID   *string `json:"client_id"`
	ClientName *string `json:"client_name"`
	LastUsedAt *string `json:"last_used_at,omitempty"`
	CreatedAt  string  `json:"created_at"`
}
