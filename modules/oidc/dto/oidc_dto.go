package dto

// ============================================================
// Client Registration
// ============================================================

type RegisterClientRequest struct {
	Name        string `json:"name"`
	RedirectURI string `json:"redirect_uri"`
	ClientType  string `json:"client_type"`
}

type RegisterClientResponse struct {
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret,omitempty"`
	Name         string `json:"name"`
	RedirectURI  string `json:"redirect_uri"`
	ClientType   string `json:"client_type"`
}

// ============================================================
// Authorization
// ============================================================

type AuthorizationRequest struct {
	ClientID            string
	RedirectURI         string
	ResponseType        string
	Scope               string
	State               string
	Nonce               string
	CodeChallenge       string
	CodeChallengeMethod string
}

// ============================================================
// Token
// ============================================================

type TokenRequest struct {
	GrantType    string `json:"grant_type"`
	Code         string `json:"code"`
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`

	RedirectURI  string `json:"redirect_uri"`
	CodeVerifier string `json:"code_verifier"`
}

type TokenResponse struct {
	AccessToken string `json:"access_token"`

	TokenType string `json:"token_type"`

	ExpiresIn int `json:"expires_in"`

	IDToken string `json:"id_token,omitempty"`

	Scope string `json:"scope,omitempty"`
}

// ============================================================
// UserInfo
// ============================================================

type UserInfoResponse struct {
	Subject           string `json:"sub"`
	Name              string `json:"name,omitempty"`
	PreferredUsername string `json:"preferred_username,omitempty"`
}
