package controller

import (
	"encoding/json"
	"log"
	"net/http"
	"net/url"
	"strings"

	"nid-backend/config"
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
// POST /oauth/register
//
// Example:
//
// {
//   "name": "Demo App",
//   "redirect_uri": "http://localhost:5174/oauth/callback",
//   "client_type": "public"
// }
//
// Public clients:
// - No client secret
// - PKCE required
//
// Confidential clients:
// - Client secret
// - PKCE required
//
// ============================================================

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

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(
			w,
			"invalid request body",
			http.StatusBadRequest,
		)
		return
	}

	result, err := c.service.RegisterClient(req)
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

	if err := json.NewEncoder(w).Encode(result); err != nil {
		log.Printf(
			"failed encoding registration response: %v",
			err,
		)
	}
}

// ============================================================
// Authorization Endpoint
// ============================================================
//
// GET /oauth/authorize
//
// Example:
//
// /oauth/authorize?
//   client_id=xxx
//   &redirect_uri=http://localhost:5174/oauth/callback
//   &response_type=code
//   &scope=openid%20profile
//   &state=xxx
//   &nonce=xxx
//   &code_challenge=xxx
//   &code_challenge_method=S256
//
// ============================================================

func (c *OIDCController) AuthorizeHandler(
	w http.ResponseWriter,
	r *http.Request,
) {
	// ========================================================
	// Method
	// ========================================================

	if r.Method != http.MethodGet {
		http.Error(
			w,
			"method not allowed",
			http.StatusMethodNotAllowed,
		)
		return
	}

	// ========================================================
	// Parse query parameters
	// ========================================================

	query := r.URL.Query()

	clientID := strings.TrimSpace(
		query.Get("client_id"),
	)

	redirectURI := strings.TrimSpace(
		query.Get("redirect_uri"),
	)

	responseType := strings.TrimSpace(
		query.Get("response_type"),
	)

	scope := strings.TrimSpace(
		query.Get("scope"),
	)

	codeChallenge := strings.TrimSpace(
		query.Get("code_challenge"),
	)

	codeChallengeMethod := strings.TrimSpace(
		query.Get("code_challenge_method"),
	)

	nonce := strings.TrimSpace(
		query.Get("nonce"),
	)

	state := query.Get("state")

	// ========================================================
	// Validate client_id
	// ========================================================

	if clientID == "" {
		http.Error(
			w,
			"client_id is required",
			http.StatusBadRequest,
		)
		return
	}

	// ========================================================
	// Validate redirect_uri
	// ========================================================

	if redirectURI == "" {
		http.Error(
			w,
			"redirect_uri is required",
			http.StatusBadRequest,
		)
		return
	}

	// ========================================================
	// Validate response_type
	// ========================================================

	if responseType == "" {
		responseType = "code"
	}

	if responseType != "code" {
		http.Error(
			w,
			"only response_type=code is supported",
			http.StatusBadRequest,
		)
		return
	}

	// ========================================================
	// Default scope
	// ========================================================

	if scope == "" {
		scope = "openid"
	}

	// ========================================================
	// Validate openid scope
	// ========================================================

	hasOpenID := false

	for _, item := range strings.Fields(scope) {
		if item == "openid" {
			hasOpenID = true
			break
		}
	}

	if !hasOpenID {
		http.Error(
			w,
			"openid scope is required",
			http.StatusBadRequest,
		)
		return
	}

	// ========================================================
	// Validate PKCE code_challenge
	// ========================================================

	if codeChallenge == "" {
		http.Error(
			w,
			"code_challenge is required",
			http.StatusBadRequest,
		)
		return
	}

	// ========================================================
	// Validate PKCE method
	// ========================================================

	if codeChallengeMethod == "" {
		codeChallengeMethod = "S256"
	}

	if codeChallengeMethod != "S256" {
		http.Error(
			w,
			"only S256 PKCE is supported",
			http.StatusBadRequest,
		)
		return
	}

	// ========================================================
	// Validate nonce
	// ========================================================

	if nonce == "" {
		http.Error(
			w,
			"nonce is required",
			http.StatusBadRequest,
		)
		return
	}

	// ========================================================
	// IMPORTANT
	//
	// Validate redirect URI BEFORE showing login/consent.
	//
	// Never redirect to an unregistered URI.
	// ========================================================

	if err := c.service.ValidateAuthorizationRequest(
		clientID,
		redirectURI,
		responseType,
		scope,
		codeChallenge,
		codeChallengeMethod,
		nonce,
	); err != nil {
		log.Printf(
			"invalid OAuth authorization request: %v",
			err,
		)

		http.Error(
			w,
			"invalid authorization request",
			http.StatusBadRequest,
		)
		return
	}

	// ========================================================
	// Build authorization/consent URL
	// ========================================================
	//
	// The user is not approved yet.
	//
	// We send the authorization request to the NID
	// authentication/consent UI.
	//
	// Change this path to whatever your frontend uses.
	// ========================================================

	authorizationURL, err := url.Parse(
		config.LoadConfig().FrontendURL+ "/oauth/authorize",
	)
	if err != nil {
		log.Printf(
			"failed parsing authorization URL: %v",
			err,
		)

		http.Error(
			w,
			"internal server error",
			http.StatusInternalServerError,
		)
		return
	}

	// ========================================================
	// Forward OAuth parameters
	// ========================================================

	params := authorizationURL.Query()

	params.Set(
		"client_id",
		clientID,
	)

	params.Set(
		"redirect_uri",
		redirectURI,
	)

	params.Set(
		"response_type",
		responseType,
	)

	params.Set(
		"scope",
		scope,
	)

	params.Set(
		"code_challenge",
		codeChallenge,
	)

	params.Set(
		"code_challenge_method",
		codeChallengeMethod,
	)

	params.Set(
		"nonce",
		nonce,
	)

	if state != "" {
		params.Set(
			"state",
			state,
		)
	}

	authorizationURL.RawQuery = params.Encode()

	// ========================================================
	// Redirect user to NID authorization UI
	// ========================================================

	log.Printf(
		"OAuth authorization request: client_id=%s redirect_uri=%s",
		clientID,
		redirectURI,
	)

	http.Redirect(
		w,
		r,
		authorizationURL.String(),
		http.StatusFound,
	)
}

// ============================================================
// Authorization Approval
// ============================================================
//
// POST /oauth/authorize/approve
//
// This endpoint is called by NID's authorization UI after
// the user clicks:
//
//     "Allow"
//
// Expected JSON:
//
// {
//   "client_id": "...",
//   "redirect_uri": "...",
//   "response_type": "code",
//   "scope": "openid profile",
//   "state": "...",
//   "nonce": "...",
//   "code_challenge": "...",
//   "code_challenge_method": "S256"
// }
//
// ============================================================

func (c *OIDCController) ApproveAuthorizationHandler(
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

    // ========================================================
    // Verify NID user
    // ========================================================

    userID, err := helpers.GetUserIDFromRequest(r)
    if err != nil {
        http.Error(
            w,
            "nid authentication required",
            http.StatusUnauthorized,
        )
        return
    }

    // ========================================================
    // Parse authorization request
    // ========================================================

    var req dto.AuthorizationRequest

    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(
            w,
            "invalid authorization request",
            http.StatusBadRequest,
        )
        return
    }

    // ========================================================
    // Normalize
    // ========================================================

    req.ClientID = strings.TrimSpace(req.ClientID)
    req.RedirectURI = strings.TrimSpace(req.RedirectURI)
    req.ResponseType = strings.TrimSpace(req.ResponseType)
    req.Scope = strings.TrimSpace(req.Scope)
    req.CodeChallenge = strings.TrimSpace(req.CodeChallenge)
    req.CodeChallengeMethod = strings.TrimSpace(
        req.CodeChallengeMethod,
    )
    req.Nonce = strings.TrimSpace(req.Nonce)

    // ========================================================
    // Validate client_id
    // ========================================================

    if req.ClientID == "" {
        http.Error(
            w,
            "missing client_id",
            http.StatusBadRequest,
        )
        return
    }

    // ========================================================
    // Validate redirect_uri
    // ========================================================

    if req.RedirectURI == "" {
        http.Error(
            w,
            "missing redirect_uri",
            http.StatusBadRequest,
        )
        return
    }

    // ========================================================
    // Validate response_type
    // ========================================================

    if req.ResponseType == "" {
        req.ResponseType = "code"
    }

    if req.ResponseType != "code" {
        http.Error(
            w,
            "response_type must be code",
            http.StatusBadRequest,
        )
        return
    }

    // ========================================================
    // Default scope
    // ========================================================

    if req.Scope == "" {
        req.Scope = "openid"
    }

    // ========================================================
    // Validate openid scope
    // ========================================================

    hasOpenID := false

    for _, scope := range strings.Fields(req.Scope) {
        if scope == "openid" {
            hasOpenID = true
            break
        }
    }

    if !hasOpenID {
        http.Error(
            w,
            "openid scope is required",
            http.StatusBadRequest,
        )
        return
    }

    // ========================================================
    // Validate PKCE code_challenge
    // ========================================================

    if req.CodeChallenge == "" {
        http.Error(
            w,
            "code_challenge is required",
            http.StatusBadRequest,
        )
        return
    }

    // ========================================================
    // Validate PKCE method
    // ========================================================

    if req.CodeChallengeMethod == "" {
        req.CodeChallengeMethod = "S256"
    }

    if req.CodeChallengeMethod != "S256" {
        http.Error(
            w,
            "code_challenge_method must be S256",
            http.StatusBadRequest,
        )
        return
    }

    // ========================================================
    // Validate nonce
    // ========================================================

    if req.Nonce == "" {
        http.Error(
            w,
            "nonce is required",
            http.StatusBadRequest,
        )
        return
    }

    // ========================================================
    // Validate client + redirect + OAuth parameters
    // ========================================================

    if err := c.service.ValidateAuthorizationRequest(
        req.ClientID,
        req.RedirectURI,
        req.ResponseType,
        req.Scope,
        req.CodeChallenge,
        req.CodeChallengeMethod,
        req.Nonce,
    ); err != nil {
        http.Error(
            w,
            "invalid authorization request",
            http.StatusBadRequest,
        )
        return
    }

    // ========================================================
    // Generate authorization code
    // ========================================================

    code, err := c.service.GenerateAuthCode(
        req.ClientID,
        userID,
        req.RedirectURI,
        req.Scope,
        req.Nonce,
        req.CodeChallenge,
        req.CodeChallengeMethod,
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

    // ========================================================
    // Build callback URL
    // ========================================================

    callback, err := url.Parse(req.RedirectURI)
    if err != nil {
        http.Error(
            w,
            "invalid redirect URI",
            http.StatusBadRequest,
        )
        return
    }

    params := callback.Query()

    // ========================================================
    // Authorization code
    // ========================================================

    params.Set(
        "code",
        code,
    )

    // ========================================================
    // OAuth state
    // ========================================================

    if req.State != "" {
        params.Set(
            "state",
            req.State,
        )
    }

    callback.RawQuery = params.Encode()

    // ========================================================
    // Log approval
    // ========================================================

    log.Printf(
        "OAuth authorization approved: client_id=%s user_id=%s redirect=%s",
        req.ClientID,
        userID,
        req.RedirectURI,
    )

    // ========================================================
    // Response Handling (JSON vs Redirect)
    // ========================================================

    redirectURL := callback.String()

    if strings.Contains(r.Header.Get("Accept"), "application/json") {
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusOK)
        _ = json.NewEncoder(w).Encode(map[string]string{
            "redirect_uri": redirectURL,
        })
        return
    }

    // Fallback standard redirect
    http.Redirect(
        w,
        r,
        redirectURL,
        http.StatusFound,
    )
}
// ============================================================
// Authorization Denied
// ============================================================
//
// POST /oauth/authorize/deny
//
// Called when user clicks "Deny".
//
// ============================================================

func (c *OIDCController) DenyAuthorizationHandler(
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

	var req struct {
		RedirectURI string `json:"redirect_uri"`
		State       string `json:"state"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(
			w,
			"invalid request body",
			http.StatusBadRequest,
		)
		return
	}

	if req.RedirectURI == "" {
		http.Error(
			w,
			"missing redirect_uri",
			http.StatusBadRequest,
		)
		return
	}

	// ========================================================
	// Validate redirect URI before redirecting
	// ========================================================

	// NOTE:
	// In production, validate this redirect URI against the
	// client stored in DB before redirecting.
	//
	// Never blindly redirect to arbitrary URLs.

	callback, err := url.Parse(
		req.RedirectURI,
	)

	if err != nil {
		http.Error(
			w,
			"invalid redirect URI",
			http.StatusBadRequest,
		)
		return
	}

	params := callback.Query()

	params.Set(
		"error",
		"access_denied",
	)

	params.Set(
		"error_description",
		"user denied authorization",
	)

	if req.State != "" {
		params.Set(
			"state",
			req.State,
		)
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
// Token Endpoint
// ============================================================
//
// POST /oauth/token
//
// Exchanges:
//
// authorization_code
// +
// code_verifier
//
// for:
//
// access_token
// id_token
//
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

	// ========================================================
	// Parse form
	// ========================================================

	if err := r.ParseForm(); err != nil {
		http.Error(
			w,
			"invalid form",
			http.StatusBadRequest,
		)
		return
	}

	req := dto.TokenRequest{
		GrantType: r.FormValue(
			"grant_type",
		),

		Code: r.FormValue(
			"code",
		),

		ClientID: r.FormValue(
			"client_id",
		),

		ClientSecret: r.FormValue(
			"client_secret",
		),

		RedirectURI: r.FormValue(
			"redirect_uri",
		),

		CodeVerifier: r.FormValue(
			"code_verifier",
		),
	}

	// ========================================================
	// Validate grant type
	// ========================================================

	if req.GrantType != "authorization_code" {
		http.Error(
			w,
			"unsupported grant type",
			http.StatusBadRequest,
		)
		return
	}

	// ========================================================
	// Exchange code
	// ========================================================

	result, err := c.service.ExchangeCodeForToken(
		req,
	)

	if err != nil {
		log.Printf(
			"OAuth token exchange failed: %v",
			err,
		)

		http.Error(
			w,
			err.Error(),
			http.StatusBadRequest,
		)
		return
	}

	// ========================================================
	// Response
	// ========================================================

	w.Header().Set(
		"Content-Type",
		"application/json",
	)

	w.Header().Set(
		"Cache-Control",
		"no-store",
	)

	w.Header().Set(
		"Pragma",
		"no-cache",
	)

	if err := json.NewEncoder(w).Encode(
		result,
	); err != nil {
		log.Printf(
			"failed encoding token response: %v",
			err,
		)
	}
}

// ============================================================
// UserInfo Endpoint
// ============================================================
//
// GET /oauth/userinfo
//
// Authorization:
//
// Bearer <access_token>
//
// ============================================================

func (c *OIDCController) UserInfoHandler(
	w http.ResponseWriter,
	r *http.Request,
) {
	if r.Method != http.MethodGet {
		http.Error(
			w,
			"method not allowed",
			http.StatusMethodNotAllowed,
		)
		return
	}

	authHeader := strings.TrimSpace(
		r.Header.Get("Authorization"),
	)

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

	accessToken := strings.TrimSpace(
		strings.TrimPrefix(
			authHeader,
			"Bearer ",
		),
	)

	if accessToken == "" {
		http.Error(
			w,
			"missing access token",
			http.StatusUnauthorized,
		)
		return
	}

	userInfo, err := c.service.UserInfo(
		accessToken,
	)

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

	w.Header().Set(
		"Cache-Control",
		"no-store",
	)

	if err := json.NewEncoder(w).Encode(
		userInfo,
	); err != nil {
		log.Printf(
			"failed encoding userinfo response: %v",
			err,
		)
	}
}

// ============================================================
// OpenID Connect Discovery
// ============================================================
//
// GET /.well-known/openid-configuration
//
// ============================================================

func (c *OIDCController) DiscoveryHandler(
	w http.ResponseWriter,
	r *http.Request,
) {
	if r.Method != http.MethodGet {
		http.Error(
			w,
			"method not allowed",
			http.StatusMethodNotAllowed,
		)
		return
	}

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

		"token_endpoint_auth_methods_supported": []string{
			"none",
			"client_secret_post",
		},
	}

	w.Header().Set(
		"Content-Type",
		"application/json",
	)

	w.Header().Set(
		"Cache-Control",
		"public, max-age=3600",
	)

	if err := json.NewEncoder(w).Encode(
		config,
	); err != nil {
		log.Printf(
			"failed encoding discovery response: %v",
			err,
		)
	}
}

// ============================================================
// JWKS
// ============================================================
//
// GET /.well-known/jwks.json
//
// ============================================================

func (c *OIDCController) JWKSHandler(
	w http.ResponseWriter,
	r *http.Request,
) {
	if r.Method != http.MethodGet {
		http.Error(
			w,
			"method not allowed",
			http.StatusMethodNotAllowed,
		)
		return
	}

	jwks, err := c.service.JWKS()

	if err != nil {
		log.Printf(
			"failed to generate JWKS: %v",
			err,
		)

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

	w.Header().Set(
		"Cache-Control",
		"public, max-age=3600",
	)

	if err := json.NewEncoder(w).Encode(
		jwks,
	); err != nil {
		log.Printf(
			"failed encoding JWKS response: %v",
			err,
		)
	}
}


func (c *OIDCController) GetClientInfoHandler(
	w http.ResponseWriter,
	r *http.Request,
) {
	if r.Method != http.MethodGet {
		http.Error(
			w,
			"method not allowed",
			http.StatusMethodNotAllowed,
		)
		return
	}

	clientID := r.URL.Query().Get("client_id")
	if clientID == "" {
		http.Error(
			w,
			"missing client_id parameter",
			http.StatusBadRequest,
		)
		return
	}

	clientInfo, err := c.service.GetClientInfo(clientID)
	if err != nil {
		http.Error(
			w,
			err.Error(),
			http.StatusNotFound,
		)
		return
	}

	w.Header().Set(
		"Content-Type",
		"application/json",
	)
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(clientInfo); err != nil {
		log.Printf(
			"failed encoding client info response: %v",
			err,
		)
	}
}
