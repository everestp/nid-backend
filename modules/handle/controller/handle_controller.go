// modules/handle/controller/handle_controller.go
package controller

import (
	"encoding/json"
	"net/http"
	"nid-backend/modules/handle/dto"
	"nid-backend/modules/handle/service"
	"nid-backend/pkg/middleware"
)

type HandleController struct {
	service *service.HandleService
}

func NewHandleController(service *service.HandleService) *HandleController {
	return &HandleController{service: service}
}

func (c *HandleController) ClaimHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID, ok := r.Context().Value(middleware.UserIDKey).(string)
	if !ok || userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req dto.ClaimHandleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	handle, err := c.service.ClaimHandle(userID, req.Name)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(dto.HandleResponse{
		ID:        handle.ID,
		UserID:    handle.UserID,
		Name:      handle.Name,
		Status:    handle.Status,
		Primary:   handle.Primary,
		CreatedAt: handle.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	})
}
