// modules/wallet/dto/wallet_dto.go
package dto

type LinkWalletRequest struct {
	Chain   string `json:"chain"`
	Network string `json:"network"`
	Address string `json:"address"`
}

type WalletResponse struct {
	ID       string `json:"id"`
	UserID   string `json:"user_id"`
	Chain    string `json:"chain"`
	Network  string `json:"network"`
	Address  string `json:"address"`
	Status   string `json:"status"`
	LinkedAt string `json:"linked_at"`
}
