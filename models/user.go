package models

import (
	"time"
)



type HandleInfo struct {
	Handle    string `json:"handle"`
	IsPrimary bool   `json:"is_primary"`
}

type SocialIdentity struct {
	Platform string                 `json:"platform"`
	Handle   string                 `json:"handle"`
	Verified bool                   `json:"verified"`
	Metadata map[string]interface{} `json:"metadata"`
}

type WalletList struct {
	Chain    string    `json:"chain"`
	Network  string    `json:"network"`
	Address  string    `json:"address"`
	LinkedAt time.Time `json:"linked_at"`
}

type PublicProfile struct {
	User       *User             `json:"user"`
	Handles    []HandleInfo      `json:"handles"`
	Identities []SocialIdentity  `json:"identities"`
	Wallets    []WalletList        `json:"wallets"`
}
