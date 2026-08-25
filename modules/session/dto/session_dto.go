// modules/session/dto/session_dto.go

package dto

import "time"

// ============================================================
// SESSION
// ============================================================

type Session struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	ClientID  string    `json:"client_id"`
	AppName   string    `json:"app_name"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

// ============================================================
// SESSION RESPONSE
// ============================================================

type SessionResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// ============================================================
// SESSION LIST RESPONSE
// ============================================================

type SessionListResponse struct {
	Success  bool      `json:"success"`
	Sessions []Session `json:"sessions"`
	Count    int       `json:"count"`
}
