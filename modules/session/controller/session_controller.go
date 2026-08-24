// modules/session/controller/session_controller.go
package controller

import (
	"encoding/json"
	"net/http"
	"nid-backend/modules/session/service"
)

type SessionController struct {
	service *service.SessionService
}

func NewSessionController(service *service.SessionService) *SessionController {
	return &SessionController{service: service}
}

func (c *SessionController) RevokeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	sessionID := r.URL.Query().Get("id")
	if sessionID == "" {
		http.Error(w, "Missing session ID", http.StatusBadRequest)
		return
	}

	if err := c.service.Revoke(sessionID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status": "success",
		"message": "session revoked",
	})
}
