package models

import "time"

type UserDashboard struct {
	UserID         string           `json:"user_id"`
	CreatedAt      time.Time        `json:"created_at"`
	Handles        []DashboardHandle `json:"handles"`
	Socials        []DashboardSocial `json:"socials"`
	Wallets        []DashboardWallet `json:"wallets"`
	ActiveSessions []DashboardSession `json:"active_sessions"`
}

type DashboardHandle struct {
	ID        string `json:"id"`
	Handle    string `json:"handle"`
	IsPrimary bool   `json:"is_primary"`
	Status    string `json:"status"`
}

type DashboardSocial struct {
	ID               string `json:"id"`
	Platform         string `json:"platform"`
	Handle           string `json:"handle"`
	Verified         bool   `json:"verified"`
	PubliclyVisible  bool   `json:"publicly_visible"`
}

type DashboardWallet struct {
	ID       string `json:"id"`
	Chain    string `json:"chain"`
	Network  string `json:"network"`
	Address  string `json:"address"`
	Status   string `json:"status"`
}

type DashboardSession struct {
	ID         string     `json:"id"`
	ClientID   *string    `json:"client_id"`
	ClientName *string    `json:"client_name"`
	LastUsedAt *time.Time `json:"last_used_at"`
	CreatedAt  time.Time  `json:"created_at"`
}
