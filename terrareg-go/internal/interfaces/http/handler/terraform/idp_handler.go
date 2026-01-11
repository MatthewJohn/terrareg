package terraform

import (
	"net/http"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/auth/service"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/config"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/interfaces/http/handler/terrareg"
)

// TerraformIDPHandler handles Terraform OIDC Identity Provider endpoints
type TerraformIDPHandler struct {
	idpService  *service.TerraformIdpService
	infraConfig *config.InfrastructureConfig
}

// NewTerraformIDPHandler creates a new Terraform IDP handler
func NewTerraformIDPHandler(idpService *service.TerraformIdpService, infraConfig *config.InfrastructureConfig) *TerraformIDPHandler {
	return &TerraformIDPHandler{
		idpService:  idpService,
		infraConfig: infraConfig,
	}
}

// HandleOpenIDConfiguration handles GET /.well-known/openid-configuration
func (h *TerraformIDPHandler) HandleOpenIDConfiguration(w http.ResponseWriter, r *http.Request) {
	// Build base URL from PublicURL config (matching Python pattern)
	baseURL := h.infraConfig.PublicURL

	// Return OpenID Connect discovery document
	config := map[string]interface{}{
		"issuer":                                baseURL,
		"authorization_endpoint":                baseURL + "/terraform/v1/idp/authorize",
		"token_endpoint":                        baseURL + "/terraform/v1/idp/token",
		"userinfo_endpoint":                     baseURL + "/terraform/v1/idp/userinfo",
		"jwks_uri":                              baseURL + "/.well-known/jwks.json",
		"response_types_supported":              []string{"code"},
		"grant_types_supported":                 []string{"authorization_code"},
		"subject_types_supported":               []string{"public"},
		"id_token_signing_alg_values_supported": []string{"RS256"},
		"scopes_supported":                      []string{"openid", "profile", "email", "terraform"},
		"terraform_version":                     "1",
		"terraform_supported_audiences":         []string{"terraform.workspaces"},
	}

	terrareg.RespondJSON(w, http.StatusOK, config)
}

// HandleJWKS handles GET /.well-known/jwks.json
func (h *TerraformIDPHandler) HandleJWKS(w http.ResponseWriter, r *http.Request) {
	// Return mock JWKS - in production, this would serve actual signing keys
	jwks := map[string]interface{}{
		"keys": []map[string]interface{}{
			{
				"kty": "RSA",
				"kid": "terraform-idp-key-1",
				"use": "sig",
				"alg": "RS256",
				"n":   "mock-modulus-for-development-purposes",
				"e":   "AQAB",
			},
		},
	}

	terrareg.RespondJSON(w, http.StatusOK, jwks)
}

// HandleAuth handles GET /terraform/v1/idp/authorize - OIDC authorization endpoint
func (h *TerraformIDPHandler) HandleAuth(w http.ResponseWriter, r *http.Request) {
	clientID := r.URL.Query().Get("client_id")
	redirectURI := r.URL.Query().Get("redirect_uri")
	scope := r.URL.Query().Get("scope")
	state := r.URL.Query().Get("state")
	responseType := r.URL.Query().Get("response_type")

	if responseType != "code" {
		terrareg.RespondError(w, http.StatusBadRequest, "unsupported_response_type")
		return
	}

	if clientID == "" || redirectURI == "" {
		terrareg.RespondError(w, http.StatusBadRequest, "client_id and redirect_uri are required")
		return
	}

	// Create authorization request
	req := service.AuthorizationCodeRequest{
		ClientID:     clientID,
		RedirectURI:  redirectURI,
		Scope:        scope,
		State:        state,
		ResponseType: responseType,
	}

	// Generate authorization code
	resp, err := h.idpService.CreateAuthorizationCode(r.Context(), req)
	if err != nil {
		terrareg.RespondError(w, http.StatusInternalServerError, "Failed to create authorization code")
		return
	}

	// For development, redirect directly with the code
	redirectURL := redirectURI + "?code=" + resp.Code
	if state != "" {
		redirectURL += "&state=" + state
	}

	http.Redirect(w, r, redirectURL, http.StatusFound)
}

// HandleToken handles POST /terraform/v1/idp/token - OIDC token endpoint
func (h *TerraformIDPHandler) HandleToken(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		terrareg.RespondError(w, http.StatusBadRequest, "Failed to parse form data")
		return
	}

	req := service.AccessTokenRequest{
		GrantType:   r.FormValue("grant_type"),
		Code:        r.FormValue("code"),
		RedirectURI: r.FormValue("redirect_uri"),
		ClientID:    r.FormValue("client_id"),
	}

	if req.GrantType != "authorization_code" {
		terrareg.RespondError(w, http.StatusBadRequest, "unsupported_grant_type")
		return
	}

	if req.Code == "" || req.ClientID == "" || req.RedirectURI == "" {
		terrareg.RespondError(w, http.StatusBadRequest, "code, client_id, and redirect_uri are required")
		return
	}

	// Exchange code for token
	resp, err := h.idpService.ExchangeCodeForToken(r.Context(), req)
	if err != nil {
		terrareg.RespondError(w, http.StatusBadRequest, "Invalid or expired authorization code")
		return
	}

	terrareg.RespondJSON(w, http.StatusOK, resp)
}

// HandleUserInfo handles GET /terraform/v1/idp/userinfo - OIDC userinfo endpoint
func (h *TerraformIDPHandler) HandleUserInfo(w http.ResponseWriter, r *http.Request) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		terrareg.RespondError(w, http.StatusUnauthorized, "Missing Authorization header")
		return
	}

	// Extract token from Authorization header
	token := ""
	if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
		token = authHeader[7:]
	}

	if token == "" {
		terrareg.RespondError(w, http.StatusUnauthorized, "Invalid authorization header format")
		return
	}

	// Validate token and get user info
	userInfo, err := h.idpService.ValidateToken(r.Context(), token)
	if err != nil {
		terrareg.RespondError(w, http.StatusUnauthorized, "Invalid or expired access token")
		return
	}

	terrareg.RespondJSON(w, http.StatusOK, userInfo)
}

// HandleRevoke handles POST /terraform/v1/idp/revoke - OAuth token revocation endpoint
func (h *TerraformIDPHandler) HandleRevoke(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		terrareg.RespondError(w, http.StatusBadRequest, "Failed to parse form data")
		return
	}

	token := r.FormValue("token")
	if token == "" {
		terrareg.RespondError(w, http.StatusBadRequest, "token parameter is required")
		return
	}

	// Revoke token
	err := h.idpService.RevokeToken(r.Context(), token)
	if err != nil {
		terrareg.RespondError(w, http.StatusBadRequest, "Failed to revoke token")
		return
	}

	w.WriteHeader(http.StatusOK)
}

// HandleCleanup handles POST /terraform/v1/idp/cleanup - cleanup expired tokens (admin endpoint)
func (h *TerraformIDPHandler) HandleCleanup(w http.ResponseWriter, r *http.Request) {
	err := h.idpService.CleanupExpired(r.Context())
	if err != nil {
		terrareg.RespondError(w, http.StatusInternalServerError, "Failed to cleanup expired tokens")
		return
	}

	response := map[string]interface{}{
		"message": "Expired tokens cleaned up successfully",
	}
	terrareg.RespondJSON(w, http.StatusOK, response)
}

// HandleHealthCheck handles GET /terraform/v1/idp/health - health check endpoint
func (h *TerraformIDPHandler) HandleHealthCheck(w http.ResponseWriter, r *http.Request) {
	health := map[string]interface{}{
		"status":    "healthy",
		"timestamp": "2024-01-01T00:00:00Z",
		"version":   "1.0.0",
	}
	terrareg.RespondJSON(w, http.StatusOK, health)
}

// TerraformStaticTokenHandler handles Terraform static token authentication
type TerraformStaticTokenHandler struct {
	// TODO: Implement when static token authentication is integrated with new auth domain
}

// NewTerraformStaticTokenHandler creates a new Terraform static token handler
func NewTerraformStaticTokenHandler() *TerraformStaticTokenHandler {
	return &TerraformStaticTokenHandler{}
}

// HandleValidateToken handles token validation for Terraform operations using the integrated auth system
func (h *TerraformStaticTokenHandler) HandleValidateToken(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement when integrated with new auth domain
	// For now, return a mock successful response
	mockResponse := map[string]interface{}{
		"valid":         true,
		"identity_id":   "terraform-static-user",
		"subject":       "Terraform Static Token",
		"permissions":   []string{"read:modules", "read:providers"},
		"identity_type": "TERRAFORM_STATIC_TOKEN",
	}
	terrareg.RespondJSON(w, http.StatusOK, mockResponse)
}

// HandleAuthStatus handles authentication status check
func (h *TerraformStaticTokenHandler) HandleAuthStatus(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement when integrated with new auth domain
	response := map[string]interface{}{
		"message": "Static token authentication integration pending",
		"status":  "stub",
	}
	terrareg.RespondJSON(w, http.StatusOK, response)
}
