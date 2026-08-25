package dto

// ============================================================
// OAuth / OIDC Client Registration
// ============================================================


type RegisterClientRequest struct {
	Name        string `json:"name"`
	RedirectURI string `json:"redirect_uri"`
	ClientType  string `json:"client_type"`
	ClientLogo  string `json:"client_logo"`
	ClientURI   string `json:"client_uri"`
	PolicyURI   string `json:"policy_uri"`
}

type RegisterClientResponse struct {
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret,omitempty"`
	Name         string `json:"name"`
	RedirectURI  string `json:"redirect_uri"`
	ClientType   string `json:"client_type"`
	ClientLogo   string `json:"client_logo"`
	ClientURI    string `json:"client_uri"`
	PolicyURI    string `json:"policy_uri"`
}

// ============================================================
// Authorization Request
// ============================================================
//
// GET /oauth/authorize
//
// Example:
//
// /oauth/authorize?
//   client_id=abc
//   &redirect_uri=https://client.xyz/callback
//   &response_type=code
//   &scope=openid%20profile
//   &state=xyz
//   &nonce=abc
//   &code_challenge=xxx
//   &code_challenge_method=S256
//

type AuthorizationRequest struct {
	ClientID            string `json:"client_id"`
	RedirectURI         string `json:"redirect_uri"`
	ResponseType        string `json:"response_type"`
	Scope               string `json:"scope"`
	State               string `json:"state"`
	Nonce               string `json:"nonce"`
	CodeChallenge       string `json:"code_challenge"`
	CodeChallengeMethod string `json:"code_challenge_method"`
}

// ============================================================
// Authorization Response
// ============================================================
//
// Returned through browser redirect:
//
// https://client.xyz/callback?
//     code=xxx
//     &state=xyz
//

type AuthorizationResponse struct {
	Code  string `json:"code"`
	State string `json:"state,omitempty"`
}

// ============================================================
// Token Request
// ============================================================
//
// POST /oauth/token
//
// Content-Type:
// application/x-www-form-urlencoded
//
// grant_type=authorization_code
// code=xxx
// client_id=xxx
// client_secret=xxx
// redirect_uri=https://client.xyz/callback
// code_verifier=xxx
//

type TokenRequest struct {
	GrantType    string `json:"grant_type"`
	Code         string `json:"code"`
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret,omitempty"`
	RedirectURI  string `json:"redirect_uri"`
	CodeVerifier string `json:"code_verifier"`
}

// ============================================================
// Token Response
// ============================================================

type TokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
	IDToken     string `json:"id_token,omitempty"`
	Scope       string `json:"scope,omitempty"`
}

// ============================================================
// UserInfo
// ============================================================
//
// GET /oauth/userinfo
//
// Authorization:
// Bearer <access_token>
//

type UserInfoResponse struct {
	Subject           string `json:"sub"`
	Name              string `json:"name,omitempty"`
	PreferredUsername string `json:"preferred_username,omitempty"`
}

// ============================================================
// OIDC Discovery
// ============================================================
//
// GET /.well-known/openid-configuration
//

type DiscoveryResponse struct {
	Issuer        string `json:"issuer"`
	AuthorizationEndpoint string `json:"authorization_endpoint"`
	TokenEndpoint string `json:"token_endpoint"`
	UserInfoEndpoint string `json:"userinfo_endpoint"`
	JWKSURI       string `json:"jwks_uri"`

	ResponseTypesSupported []string `json:"response_types_supported"`
	SubjectTypesSupported  []string `json:"subject_types_supported"`

	IDTokenSigningAlgValuesSupported []string `json:"id_token_signing_alg_values_supported"`

	GrantTypesSupported []string `json:"grant_types_supported"`
	ScopesSupported     []string `json:"scopes_supported"`

	CodeChallengeMethodsSupported []string `json:"code_challenge_methods_supported"`
}

// ============================================================
// OIDC Error Response
// ============================================================
//
// Standard OAuth/OIDC error format.
//

type ErrorResponse struct {
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description,omitempty"`
	ErrorURI         string `json:"error_uri,omitempty"`
	State            string `json:"state,omitempty"`
}


type ClientInfoResponse struct {
	ClientName string `json:"client_name"`
	ClientLogo string `json:"client_logo"`
	ClientURI  string `json:"client_uri"`
	PolicyURI  string `json:"policy_uri"`
}
