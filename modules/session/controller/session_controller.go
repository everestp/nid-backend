// modules/session/controller/session_controller.go
package controller

import (
	"encoding/json"
	"fmt"
	"net/http"
	"nid-backend/modules/session/dto"
	"nid-backend/modules/session/service"
	"nid-backend/pkg/middleware"
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

	userID, ok := r.Context().Value(middleware.UserIDKey).(string)
	if !ok || userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	sessionID := r.URL.Query().Get("id")
	if sessionID == "" {
		http.Error(w, "Missing session ID", http.StatusBadRequest)
		return
	}

	if err := c.service.Revoke(sessionID, userID); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(dto.SessionResponse{
		Success:  true,
		Message: "session revoked successfully",
	})
}
func (c *SessionController) ListHandler(
    w http.ResponseWriter,
    r *http.Request,
) {

    userID, ok :=
        r.Context().Value(
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

    sessions, err :=
        c.service.List(userID)
        fmt.Println("Tjis is the session data and  err",sessions,err)

    if err != nil {
        http.Error(
            w,
            err.Error(),
            http.StatusInternalServerError,
        )
        return
    }

    w.Header().Set(
        "Content-Type",
        "application/json",
    )

    json.NewEncoder(w).Encode(
        map[string]interface{}{
            "success":  true,
            "sessions": sessions,
            "count":    len(sessions),
        },
    )
}
