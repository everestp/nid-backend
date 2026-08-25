package controller

import (
	"encoding/json"
	"net/http"
	"strings"

	"nid-backend/modules/auth/dto"
	"nid-backend/modules/auth/service"
)

type AuthController struct {
	service *service.AuthService
}

func NewAuthController(
	service *service.AuthService,
) *AuthController {
	return &AuthController{
		service: service,
	}
}

// ============================================================
// POST /api/v1/auth/login
//
// IN-HOUSE LOGIN
//
// Wallet signature -> verify ownership -> internal session
// token -> nid_token cookie.
//
// This flow is completely independent from OIDC.
// ============================================================

func (c *AuthController) LoginHandler(
	w http.ResponseWriter,
	r *http.Request,
) {
	if r.Method != http.MethodPost {
		writeError(
			w,
			http.StatusMethodNotAllowed,
			"method not allowed",
		)
		return
	}

	var req dto.LoginRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(
			w,
			http.StatusBadRequest,
			"invalid request body",
		)
		return
	}

	req.Handle = strings.TrimSpace(req.Handle)
	req.Address = strings.TrimSpace(req.Address)
	req.Signature = strings.TrimSpace(req.Signature)
	req.Message = strings.TrimSpace(req.Message)
	req.Chain = strings.TrimSpace(req.Chain)

	// ------------------------------------------------------------
	// Validate request
	// ------------------------------------------------------------

	if req.Handle == "" {
		writeError(
			w,
			http.StatusBadRequest,
			"handle is required",
		)
		return
	}

	if req.Address == "" {
		writeError(
			w,
			http.StatusBadRequest,
			"address is required",
		)
		return
	}

	if req.Signature == "" {
		writeError(
			w,
			http.StatusBadRequest,
			"signature is required",
		)
		return
	}

	if req.Message == "" {
		writeError(
			w,
			http.StatusBadRequest,
			"message is required",
		)
		return
	}

	if req.Chain == "" {
		writeError(
			w,
			http.StatusBadRequest,
			"chain is required",
		)
		return
	}

	// ------------------------------------------------------------
	// Authenticate using wallet signature
	//
	// IMPORTANT:
	// This generates an INTERNAL NID session token.
	// It is NOT an OIDC token.
	// ------------------------------------------------------------

	userID, token, err := c.service.AuthenticateInHouse(
		req.Handle,
		req.Address,
		req.Signature,
		req.Message,
		req.Chain,
	)

	if err != nil {
		writeError(
			w,
			http.StatusUnauthorized,
			err.Error(),
		)
		return
	}

	// ------------------------------------------------------------
	// Set internal session cookie
	// ------------------------------------------------------------

	c.setTokenCookie(w, token)

	// ------------------------------------------------------------
	// Response
	// ------------------------------------------------------------

	writeJSON(
		w,
		http.StatusOK,
		map[string]interface{}{
			"token":token,
			"success": true,
			"user_id": userID,
			"handle":  req.Handle,
		},
	)
}

// ============================================================
// INTERNAL SESSION COOKIE
// ============================================================

func (c *AuthController) setTokenCookie(
	w http.ResponseWriter,
	token string,
) {
	http.SetCookie(
		w,
		&http.Cookie{
			Name:     "nid_token",
			Value:    token,
			Path:     "/",
			HttpOnly: true,

			// For local HTTP development this must be false.
			// In production behind HTTPS set this to true.
			Secure: false,

			SameSite: http.SameSiteLaxMode,

			MaxAge: 86400 * 7,
		},
	)
}

// ============================================================
// POST /api/v1/auth/logout
//
// Clears the internal NID session.
// ============================================================

func (c *AuthController) LogoutHandler(
	w http.ResponseWriter,
	r *http.Request,
) {
	if r.Method != http.MethodPost {
		writeError(
			w,
			http.StatusMethodNotAllowed,
			"method not allowed",
		)
		return
	}

	http.SetCookie(
		w,
		&http.Cookie{
			Name:     "nid_token",
			Value:    "",
			Path:     "/",
			HttpOnly: true,
			Secure:   false,
			SameSite: http.SameSiteLaxMode,
			MaxAge:   -1,
		},
	)

	writeJSON(
		w,
		http.StatusOK,
		map[string]interface{}{
			"success": true,
			"message": "logged out successfully",
		},
	)
}

// ============================================================
// GET /api/v1/auth/me
//
// Returns the currently authenticated NID user.
//
// AuthMiddleware should protect this route.
// ============================================================

func (c *AuthController) MeHandler(
	w http.ResponseWriter,
	r *http.Request,
) {
	userID, ok := r.Context().Value("user_id").(string)

	if !ok || strings.TrimSpace(userID) == "" {
		writeError(
			w,
			http.StatusUnauthorized,
			"unauthorized",
		)
		return
	}

	writeJSON(
		w,
		http.StatusOK,
		map[string]interface{}{
			"success": true,
			"user_id": userID,
		},
	)
}

// ============================================================
// HELPERS
// ============================================================

func writeJSON(
	w http.ResponseWriter,
	status int,
	data interface{},
) {
	w.Header().Set(
		"Content-Type",
		"application/json",
	)

	w.WriteHeader(status)

	_ = json.NewEncoder(w).Encode(data)
}

func writeError(
	w http.ResponseWriter,
	status int,
	message string,
) {
	writeJSON(
		w,
		status,
		map[string]interface{}{
			"success": false,
			"message": message,
		},
	)
}
