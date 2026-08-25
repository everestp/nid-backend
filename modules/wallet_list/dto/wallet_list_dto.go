package dto

import "time"

// ============================================================
// CREATE WALLET
// ============================================================

type CreateWalletRequest struct {
	Chain   string `json:"chain"`
	Network string `json:"network"`
	Address string `json:"address"`
	Status  string `json:"status"`
}

// ============================================================
// UPDATE WALLET
// ============================================================

type UpdateWalletRequest struct {
	Chain   string `json:"chain"`
	Network string `json:"network"`
	Address string `json:"address"`
	Status  string `json:"status"`
}

// ============================================================
// WALLET RESPONSE
// ============================================================

type WalletResponse struct {
	ID       string    `json:"id"`
	Chain    string    `json:"chain"`
	Network  string    `json:"network"`
	Address  string    `json:"address"`
	Status   string    `json:"status"`
	LinkedAt time.Time `json:"linkedAt"`
}
