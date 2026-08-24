// modules/handle/dto/handle_dto.go
package dto

type ClaimHandleRequest struct {
	Name string `json:"name"`
}

type HandleResponse struct {
	ID        string `json:"id"`
	UserID    string `json:"user_id"`
	Name      string `json:"name"`
	Status    string `json:"status"`
	Primary   bool   `json:"primary"`
	CreatedAt string `json:"created_at"`
}
