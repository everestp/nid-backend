package controller

import (
    "encoding/json"
    "net/http"
    "nid-backend/modules/auth/dto"
    "nid-backend/modules/auth/service"
)

type AuthController struct {
    service *service.AuthService
}

func NewAuthController(service *service.AuthService) *AuthController {
    return &AuthController{service: service}
}

func (c *AuthController) setTokenCookie(w http.ResponseWriter, token string) {
    cookie := http.Cookie{
        Name:     "nid_token",
        Value:    token,
        Path:     "/",
        HttpOnly: true,
        Secure:   true,
        SameSite: http.SameSiteLaxMode,
        MaxAge:   86400 * 7,
    }
    http.SetCookie(w, &cookie)
}

func (c *AuthController) LoginHandler(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }

    var req dto.LoginRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "Invalid request body", http.StatusBadRequest)
        return
    }

    userID, token, err := c.service.AuthenticateWithHandle(req.Handle, req.Address, req.Signature, req.Message, req.Chain)
    if err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    c.setTokenCookie(w, token)

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(dto.LoginResponse{
        Token:  token,
        UserID: userID,
        Handle: req.Handle,
    })
}
