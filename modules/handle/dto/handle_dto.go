package dto

type ClaimHandleRequest struct {
	Name      string `json:"name"`
	Address   string `json:"address"`
	Chain     string `json:"chain"`
	Signature string `json:"signature"`
	Message   string `json:"message"`
}

type HandleResponse struct {
	ID        string `json:"id"`
	UserID    string `json:"user_id"`
	Name      string `json:"name"`
	Status    string `json:"status"`
	Primary   bool   `json:"is_primary"`
	CreatedAt string `json:"created_at"`
}
