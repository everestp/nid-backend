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

type HandleInfo struct {
	Handle    string `json:"handle"`
	IsPrimary bool   `json:"is_primary"`
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
