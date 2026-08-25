// models/entity.go
package models

import "time"

type User struct {
	ID        string    `json:"id"`
	CreatedAt time.Time `json:"created_at"`
}

type Handle struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Name      string    `json:"name"`
	Status    string    `json:"status"`
	Primary   bool      `json:"primary"`
	CreatedAt time.Time `json:"created_at"`
}

type Wallet struct {
	ID       string    `json:"id"`
	UserID   string    `json:"user_id"`
	Chain    string    `json:"chain"`
	Network  string    `json:"network"`
	Address  string    `json:"address"`
	Status   string    `json:"status"`
	LinkedAt time.Time `json:"linked_at"`
}
type OAuthSession struct {
    ID           string    `json:"id"`
    UserID       string    `json:"user_id"`
    ClientID     string    `json:"client_id"`
    Status       string    `json:"status"`
    CreatedAt    time.Time `json:"created_at"`
    ExpiresAt    time.Time `json:"expires_at"`
    LastUsedAt   *time.Time `json:"last_used_at,omitempty"`
    RevokedAt    *time.Time `json:"revoked_at,omitempty"`
}
