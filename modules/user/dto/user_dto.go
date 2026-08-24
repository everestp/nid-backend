// modules/user/dto/user_dto.go
package dto

type UserProfileResponse struct {
	ID        string    `json:"id"`
	CreatedAt string    `json:"created_at"`
	Handles   []string  `json:"handles"`
	Wallets   []WalletInfo `json:"wallets"`
}

type WalletInfo struct {
	Chain   string `json:"chain"`
	Address string `json:"address"`
	Network string `json:"network"`
}
