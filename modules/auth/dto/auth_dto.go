// modules/auth/dto/auth_dto.go
package dto

type LoginRequest struct {
	Address   string `json:"address"`
	Signature string `json:"signature"`
	Message   string `json:"message"`
	Chain     string `json:"chain"`
}

type LoginResponse struct {
	Token  string `json:"token"`
	UserID string `json:"user_id"`
}
