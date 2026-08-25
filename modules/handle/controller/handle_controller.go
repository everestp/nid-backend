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

// ClaimHandler handles POST /handles/claim
func (c *HandleController) ClaimHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req dto.ClaimHandleRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	handle, err := c.service.ClaimHandle(
		req.Name,
		req.Address,
		req.Chain,
		req.Signature,
	)
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

// GetAllByUserIDHandler handles GET /handles
func (c *HandleController) GetAllByUserIDHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

   // --------------------------------------------------------
    // Authenticate user
    // --------------------------------------------------------

    userID, ok := r.Context().Value(
        middleware.UserIDKey,
    ).(string)

    if !ok || userID == "" {
        http.Error(
            w,
            "Unauthorized",
            http.StatusUnauthorized,
        )
        return
    }

	handles, err := c.service.GetAllByUserID(userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	response := make([]dto.HandleResponse, 0, len(handles))

	for _, handle := range handles {
		response = append(response, dto.HandleResponse{
			ID:        handle.ID,
			UserID:    handle.UserID,
			Name:      handle.Name,
			Status:    handle.Status,
			Primary:   handle.Primary,
			CreatedAt: handle.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		})
	}

	w.Header().Set("Content-Type", "application/json")

	json.NewEncoder(w).Encode(response)
}
