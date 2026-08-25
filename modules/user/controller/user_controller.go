// modules/user/controller/user_controller.go
package controller

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"nid-backend/modules/user/service"
	"nid-backend/pkg/middleware"
)

type UserController struct {
	service *service.UserService
}

func NewUserController(service *service.UserService) *UserController {
	return &UserController{service: service}
}

// GetProfileHandler returns the private profile of the authenticated user
func (c *UserController) GetProfileHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID, ok := r.Context().Value(middleware.UserIDKey).(string)
	if !ok || userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	profile, err := c.service.GetProfile(userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(profile)
}

// GetPublicProfileByHandleHandler returns the public profile for any handle
// Route: GET /api/v1/profile/{handle}
func (c *UserController) GetPublicProfileByHandleHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	handle := strings.TrimSpace(r.PathValue("handle"))
	handle = strings.TrimPrefix(handle, "@")
	fmt.Println("Tjios is clesn handfle",handle)

	if handle == "" {
		http.Error(w, "handle is required", http.StatusBadRequest)
		return
	}

	profile, err := c.service.GetPublicProfileByHandle(handle)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(profile)
}
