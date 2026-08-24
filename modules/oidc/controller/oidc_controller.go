package controller

import (
	"encoding/json"
	"log"
	"net/http"
	"net/url"
	"strings"

	"nid-backend/modules/oidc/dto"
	"nid-backend/modules/oidc/service"
	"nid-backend/pkg/helpers"
)

type OIDCController struct {
	service *service.OIDCService
}

func NewOIDCController(
	service *service.OIDCService,
) *OIDCController {

	return &OIDCController{
		service: service,
	}
}

// ============================================================
// Client Registration
// ============================================================
//
// POST /oauth/clients
//
// This endpoint MUST be protected by your admin/auth middleware.
// Do NOT expose it publicly.
//

func (c *OIDCController) RegisterClientHandler(
	w http.ResponseWriter,
	r *http.Request,
) {

	if r.Method != http.MethodPost {
		http.Error(
			w,
			"method not allowed",
			http.StatusMethodNotAllowed,
		)
		return
	}

	var req dto.RegisterClientRequest

	if err := json.NewDecoder(
		r.Body,
	).Decode(&req); err != nil {

		http.Error(
			w,
			"invalid request body",
			http.StatusBadRequest,
		)
		return
	}

	result, err :=
		c.service.RegisterClient(req)

	if err != nil {
		http.Error(
			w,
			err.Error(),
			http.StatusBadRequest,
		)
		return
	}

	w.Header().Set(
		"Content-Type",
		"application/json",
	)

	w.WriteHeader(http.StatusCreated)

	json.NewEncoder(w).Encode(result)
}

// ============================================================
// Authorization
// ============================================================

func (c *OIDCController) AuthorizeHandler(
	w http.ResponseWriter,
	r *http.Request,
) {
	q := r.URL.Query()

	clientID := q.Get("client_id")
	redirectURI := q.Get("redirect_uri")
	responseType := q.Get("response_type")
	scope := q.Get("scope")
	state := q.Get("state")
	nonce := q.Get("nonce")
	codeChallenge := q.Get("code_challenge")
	codeChallengeMethod := q.Get("code_challenge_method")

	log.Printf(
		"OAuth authorize: client_id=%q redirect_uri=%q response_type=%q scope=%q",
		clientID,
		redirectURI,
		responseType,
		scope,
	)

	// ------------------------------------------------------------
	// Validate required OAuth parameters
	// ------------------------------------------------------------

	if clientID == "" {
		http.Error(w, "missing client_id", http.StatusBadRequest)
		return
	}

	if redirectURI == "" {
		http.Error(w, "missing redirect_uri", http.StatusBadRequest)
		return
	}

	if responseType != "code" {
		http.Error(
			w,
			"response_type must be code",
			http.StatusBadRequest,
		)
		return
	}

	if scope == "" {
		scope = "openid"
	}

	// ------------------------------------------------------------
	// Validate scope
	// ------------------------------------------------------------

	hasOpenID := false

	for _, s := range strings.Fields(scope) {
		if s == "openid" {
			hasOpenID = true
			break
		}
	}

	if !hasOpenID {
		http.Error(
			w,
			"openid scope required",
			http.StatusBadRequest,
		)
		return
	}

	// ------------------------------------------------------------
	// Validate PKCE
	// ------------------------------------------------------------

	if codeChallenge == "" {
		http.Error(
			w,
			"code_challenge is required",
			http.StatusBadRequest,
		)
		return
	}

	if codeChallengeMethod != "S256" {
		http.Error(
			w,
			"code_challenge_method must be S256",
			http.StatusBadRequest,
		)
		return
	}

	// ------------------------------------------------------------
	// Validate OAuth client + redirect URI
	// ------------------------------------------------------------

	err := c.service.ValidateAuthorizationRequest(
		clientID,
		redirectURI,
	)

	if err != nil {
		log.Printf(
			"OAuth client validation failed: %v",
			err,
		)

		http.Error(
			w,
			"invalid client or redirect URI",
			http.StatusBadRequest,
		)
		return
	}

	// ------------------------------------------------------------
	// Check NID authentication
	// ------------------------------------------------------------

	userID, err := helpers.GetUserIDFromRequest(r)

	if err != nil {
		/*
			User is NOT authenticated.

			Send them to the React OAuth authorization page.

			IMPORTANT:
			We preserve the complete OAuth request.
		*/

		oauthURL := "http://localhost:5173/oauth/authorize?" +
			r.URL.RawQuery

		http.Redirect(
			w,
			r,
			oauthURL,
			http.StatusFound,
		)

		return
	}

	// ------------------------------------------------------------
	// User is authenticated
	// ------------------------------------------------------------

	log.Printf(
		"OAuth authenticated user: user_id=%s client_id=%s",
		userID,
		clientID,
	)

	// ------------------------------------------------------------
	// Generate authorization code
	// ------------------------------------------------------------

	code, err := c.service.GenerateAuthCode(
		clientID,
		userID,
		redirectURI,
		scope,
		nonce,
		codeChallenge,
	)

	if err != nil {
		log.Printf(
			"failed generating authorization code: %v",
			err,
		)

		http.Error(
			w,
			"failed to generate authorization code",
			http.StatusInternalServerError,
		)

		return
	}

	// ------------------------------------------------------------
	// Redirect to client callback
	// ------------------------------------------------------------

	callback, err := url.Parse(redirectURI)

	if err != nil {
		http.Error(
			w,
			"invalid redirect URI",
			http.StatusBadRequest,
		)

		return
	}

	params := callback.Query()

	params.Set("code", code)

	if state != "" {
		params.Set("state", state)
	}

	callback.RawQuery = params.Encode()

	http.Redirect(
		w,
		r,
		callback.String(),
		http.StatusFound,
	)
}

// ============================================================
// Token
// ============================================================

func (c *OIDCController) TokenHandler(
	w http.ResponseWriter,
	r *http.Request,
) {

	if r.Method != http.MethodPost {

		http.Error(
			w,
			"method not allowed",
			http.StatusMethodNotAllowed,
		)

		return
	}

	if err := r.ParseForm(); err != nil {

		http.Error(
			w,
			"invalid form",
			http.StatusBadRequest,
		)

		return
	}

	req := dto.TokenRequest{
		GrantType:    r.FormValue("grant_type"),
		Code:         r.FormValue("code"),
		ClientID:     r.FormValue("client_id"),
		ClientSecret: r.FormValue("client_secret"),
		RedirectURI:  r.FormValue("redirect_uri"),
		CodeVerifier: r.FormValue("code_verifier"),
	}

	if req.GrantType != "authorization_code" {

		http.Error(
			w,
			"unsupported grant type",
			http.StatusBadRequest,
		)

		return
	}

	result, err :=
		c.service.ExchangeCodeForToken(req)

	if err != nil {

		http.Error(
			w,
			err.Error(),
			http.StatusBadRequest,
		)

		return
	}

	w.Header().Set(
		"Content-Type",
		"application/json",
	)

	json.NewEncoder(w).Encode(result)
}

// ============================================================
// UserInfo
// ============================================================

func (c *OIDCController) UserInfoHandler(
	w http.ResponseWriter,
	r *http.Request,
) {

	authHeader := r.Header.Get("Authorization")

	if !strings.HasPrefix(
		authHeader,
		"Bearer ",
	) {

		http.Error(
			w,
			"missing bearer token",
			http.StatusUnauthorized,
		)

		return
	}

	accessToken := strings.TrimPrefix(
		authHeader,
		"Bearer ",
	)

	userInfo, err :=
		c.service.UserInfo(accessToken)

	if err != nil {

		http.Error(
			w,
			"invalid access token",
			http.StatusUnauthorized,
		)

		return
	}

	w.Header().Set(
		"Content-Type",
		"application/json",
	)

	json.NewEncoder(w).Encode(userInfo)
}

// ============================================================
// Discovery
// ============================================================

func (c *OIDCController) DiscoveryHandler(
	w http.ResponseWriter,
	r *http.Request,
) {

	baseURL := "https://nid.xyz"

	config := map[string]interface{}{
		"issuer": baseURL,

		"authorization_endpoint":
			baseURL + "/oauth/authorize",

		"token_endpoint":
			baseURL + "/oauth/token",

		"userinfo_endpoint":
			baseURL + "/oauth/userinfo",

		"jwks_uri":
			baseURL + "/.well-known/jwks.json",

		"response_types_supported": []string{
			"code",
		},

		"subject_types_supported": []string{
			"public",
		},

		"id_token_signing_alg_values_supported": []string{
			"RS256",
		},

		"grant_types_supported": []string{
			"authorization_code",
		},

		"scopes_supported": []string{
			"openid",
			"profile",
			"wallet",
		},

		"code_challenge_methods_supported": []string{
			"S256",
		},
	}

	w.Header().Set(
		"Content-Type",
		"application/json",
	)

	json.NewEncoder(w).Encode(config)
}

// ============================================================
// JWKS
// ============================================================

func (c *OIDCController) JWKSHandler(
	w http.ResponseWriter,
	r *http.Request,
) {

	jwks, err := c.service.JWKS()

	if err != nil {

		http.Error(
			w,
			"failed to generate JWKS",
			http.StatusInternalServerError,
		)

		return
	}

	w.Header().Set(
		"Content-Type",
		"application/json",
	)

	json.NewEncoder(w).Encode(jwks)
}
