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
// GetCurrentLoggedInUserHandler returns the current authenticated user's profile
func (c *UserController) GetCurrentLoggedInUserHandler(w http.ResponseWriter, r *http.Request) {
if r.Method != http.MethodGet {
http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
return
}

userID, ok := r.Context().Value(middleware.UserIDKey).(string)
if !ok || userID == "" {
	http.Error(w, "Unauthorized", http.StatusUnauthorized)
	return
}

user, err := c.service.GetCurrentLoggedInUser(userID)
if err != nil {
	http.Error(w, err.Error(), http.StatusNotFound)
	return
}

w.Header().Set("Content-Type", "application/json")
w.WriteHeader(http.StatusOK)

if err := json.NewEncoder(w).Encode(user); err != nil {
	http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	return
}

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

// GetDashboardHandler returns the at-a-glance dashboard data for the authenticated user
func (c *UserController) GetDashboardHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID, ok := r.Context().Value(middleware.UserIDKey).(string)
	if !ok || userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	dashboard, err := c.service.GetUserDashboard(userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	fmt.Println("Tis is the dashboara at controller",dashboard,err)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(dashboard)
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
