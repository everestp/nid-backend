package controller

import (
	"encoding/json"
	"net/http"
	"nid-backend/modules/oidc/service"
	"nid-backend/pkg/helpers"
)

type OIDCController struct {
	service *service.OIDCService
}

func NewOIDCController(service *service.OIDCService) *OIDCController {
	return &OIDCController{service: service}
}

// 1. Authorize Handler: /oauth/authorize
func (c *OIDCController) AuthorizeHandler(w http.ResponseWriter, r *http.Request) {
	clientID := r.URL.Query().Get("client_id")
	redirectURI := r.URL.Query().Get("redirect_uri")
	responseType := r.URL.Query().Get("response_type")

	if clientID == "" || redirectURI == "" || responseType != "code" {
		http.Error(w, "Invalid OAuth parameters", http.StatusBadRequest)
		return
	}

	// Extract logged-in NID user from active session/token header
	userID, err := helpers.GetUserIDFromRequest(r)
	if err != nil {
		// Not logged in to NID? Redirect to frontend login with return callback URL
		loginURL := "/login?client_id=" + clientID + "&redirect_uri=" + redirectURI
		http.Redirect(w, r, loginURL, http.StatusFound)
		return
	}

	// Generate authorization code for this user session
	code, err := c.service.GenerateAuthCode(clientID, userID)
	if err != nil {
		http.Error(w, "Failed to generate authorization code", http.StatusInternalServerError)
		return
	}

	// Redirect back to third-party app with auth code
	http.Redirect(w, r, redirectURI+"?code="+code, http.StatusFound)
}

// 2. Token Handler: /oauth/token
func (c *OIDCController) TokenHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Supports both form-urlencoded and JSON payloads
	_ = r.ParseForm()
	code := r.FormValue("code")
	clientID := r.FormValue("client_id")
	clientSecret := r.FormValue("client_secret")
	grantType := r.FormValue("grant_type")

	if grantType != "authorization_code" {
		http.Error(w, "Unsupported grant type", http.StatusBadRequest)
		return
	}

	_, idToken, err := c.service.ExchangeCodeForToken(clientID, clientSecret, code)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"access_token": idToken,
		"token_type":   "Bearer",
		"id_token":     idToken,
		"expires_in":   3600,
	})
}

// 3. OIDC Discovery Config Endpoint: /.well-known/openid-configuration
func (c *OIDCController) DiscoveryHandler(w http.ResponseWriter, r *http.Request) {
	baseURL := "https://nid.xyz" // Replace with your production domain/env variable

	config := map[string]interface{}{
		"issuer":                 baseURL,
		"authorization_endpoint": baseURL + "/oauth/authorize",
		"token_endpoint":         baseURL + "/oauth/token",
		"userinfo_endpoint":      baseURL + "/oauth/userinfo",
		"response_types_supported": []string{"code"},
		"subject_types_supported":  []string{"public"},
		"id_token_signing_alg_values_supported": []string{"HS256"},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(config)
}
