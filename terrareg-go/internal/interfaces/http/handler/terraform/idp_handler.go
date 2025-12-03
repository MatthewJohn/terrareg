package terraform

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/go-chi/chi/v5"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/auth/terraform"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/interfaces/http/handler/terrareg"
)

// TerraformIDPHandler handles Terraform OIDC Identity Provider endpoints
type TerraformIDPHandler struct {
	idp *terraform.TerraformIDP
}

// NewTerraformIDPHandler creates a new Terraform IDP handler
func NewTerraformIDPHandler(idp *terraform.TerraformIDP) *TerraformIDPHandler {
	return &TerraformIDPHandler{
		idp: idp,
	}
}

// HandleOpenIDConfiguration handles GET /.well-known/openid-configuration
func (h *TerraformIDPHandler) HandleOpenIDConfiguration(w http.ResponseWriter, r *http.Request) {
	if h.idp == nil || !h.idp.IsEnabled() {
		terrareg.RespondError(w, http.StatusNotFound, "Terraform IDP is not enabled")
		return
	}

	config := h.idp.GetOpenIDConfiguration()
	if config == nil {
		terrareg.RespondError(w, http.StatusInternalServerError, "Failed to generate OIDC configuration")
		return
	}

	terrareg.RespondJSON(w, http.StatusOK, config)
}

// HandleJWKS handles GET /.well-known/jwks.json
func (h *TerraformIDPHandler) HandleJWKS(w http.ResponseWriter, r *http.Request) {
	if h.idp == nil || !h.idp.IsEnabled() {
		terrareg.RespondError(w, http.StatusNotFound, "Terraform IDP is not enabled")
		return
	}

	jwks, err := h.idp.GetJWKS()
	if err != nil {
		terrareg.RespondError(w, http.StatusInternalServerError, "Failed to generate JWKS")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(jwks))
}

// HandleAuth handles GET /oauth2/auth - OIDC authorization endpoint
func (h *TerraformIDPHandler) HandleAuth(w http.ResponseWriter, r *http.Request) {
	if h.idp == nil || !h.idp.IsEnabled() {
		terrareg.RespondError(w, http.StatusNotFound, "Terraform IDP is not enabled")
		return
	}

	// Extract authorization parameters
	responseType := r.URL.Query().Get("response_type")
	clientID := r.URL.Query().Get("client_id")
	redirectURI := r.URL.Query().Get("redirect_uri")
	scope := r.URL.Query().Get("scope")
	state := r.URL.Query().Get("state")

	// Validate required parameters
	if responseType == "" || clientID == "" || redirectURI == "" {
		terrareg.RespondError(w, http.StatusBadRequest, "Missing required OAuth parameters")
		return
	}

	// Create authorization request
	authRequest := terraform.OIDCAuthRequest{
		ResponseType: responseType,
		ClientID:     clientID,
		RedirectURI:  redirectURI,
		Scope:        scope,
		State:        state,
	}

	// For now, we'll redirect to a simple success page with the authorization code
	// In a real implementation, this would involve user authentication and consent
	authCode := "sample-auth-code" // This should be properly generated

	redirectURL, err := url.Parse(redirectURI)
	if err != nil {
		terrareg.RespondError(w, http.StatusBadRequest, "Invalid redirect URI")
		return
	}

	params := redirectURL.Query()
	params.Set("code", authCode)
	if state != "" {
		params.Set("state", state)
	}
	redirectURL.RawQuery = params.Encode()
	redirectURL.Fragment = ""

	http.Redirect(w, r, redirectURL.String(), http.StatusFound)
}

// HandleToken handles POST /oauth2/token - OIDC token endpoint
func (h *TerraformIDPHandler) HandleToken(w http.ResponseWriter, r *http.Request) {
	if h.idp == nil || !h.idp.IsEnabled() {
		terrareg.RespondError(w, http.StatusNotFound, "Terraform IDP is not enabled")
		return
	}

	// Parse form data
	if err := r.ParseForm(); err != nil {
		terrareg.RespondError(w, http.StatusBadRequest, "Failed to parse form data")
		return
	}

	grantType := r.FormValue("grant_type")
	clientID := r.FormValue("client_id")
	clientSecret := r.FormValue("client_secret")
	code := r.FormValue("code")
	redirectURI := r.FormValue("redirect_uri")
	scope := r.FormValue("scope")

	// Create token request
	tokenRequest := terraform.OIDCTokenRequest{
		GrantType:    grantType,
		Code:         code,
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Scope:        scope,
	}

	// Handle token request
	tokenResponse, err := h.idp.HandleTokenRequest(r.Context(), tokenRequest)
	if err != nil {
		terrareg.RespondError(w, http.StatusBadRequest, fmt.Sprintf("Token request failed: %s", err.Error()))
		return
	}

	terrareg.RespondJSON(w, http.StatusOK, tokenResponse)
}

// HandleUserInfo handles GET /oauth2/userinfo - OIDC userinfo endpoint
func (h *TerraformIDPHandler) HandleUserInfo(w http.ResponseWriter, r *http.Request) {
	if h.idp == nil || !h.idp.IsEnabled() {
		terrareg.RespondError(w, http.StatusNotFound, "Terraform IDP is not enabled")
		return
	}

	// Extract access token from Authorization header
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		terrareg.RespondError(w, http.StatusUnauthorized, "Missing Authorization header")
		return
	}

	token := terraform.ExtractTokenFromHeader(authHeader)
	if token == "" {
		terrareg.RespondError(w, http.StatusUnauthorized, "Invalid Authorization header")
		return
	}

	// Handle userinfo request
	userInfo, err := h.idp.HandleUserInfoRequest(r.Context(), token)
	if err != nil {
		terrareg.RespondError(w, http.StatusUnauthorized, fmt.Sprintf("Userinfo request failed: %s", err.Error()))
		return
	}

	terrareg.RespondJSON(w, http.StatusOK, userInfo)
}

// TerraformStaticTokenHandler handles Terraform static token authentication
type TerraformStaticTokenHandler struct {
	auth *terraform.TerraformAuthenticator
}

// NewTerraformStaticTokenHandler creates a new Terraform static token handler
func NewTerraformStaticTokenHandler(auth *terraform.TerraformAuthenticator) *TerraformStaticTokenHandler {
	return &TerraformStaticTokenHandler{
		auth: auth,
	}
}

// HandleValidateToken handles token validation for Terraform operations
func (h *TerraformStaticTokenHandler) HandleValidateToken(w http.ResponseWriter, r *http.Request) {
	if h.auth == nil {
		terrareg.RespondError(w, http.StatusInternalServerError, "Terraform authenticator not configured")
		return
	}

	// Extract token from request
	authHeader := r.Header.Get("Authorization")
	token := terraform.ExtractTokenFromHeader(authHeader)

	if token == "" {
		terrareg.RespondError(w, http.StatusUnauthorized, "Missing authentication token")
		return
	}

	// Get required permissions from query parameters
	requiredPermissions := r.URL.Query()["permission"]
	if len(requiredPermissions) == 0 {
		// Default to read permission if none specified
		requiredPermissions = []string{"read"}
	}

	// Validate token with permissions
	if staticToken := h.auth.(*terraform.TerraformStaticTokenAuth); staticToken != nil {
		err := staticToken.ValidateToken(r.Context(), token, requiredPermissions)
		if err != nil {
			terrareg.RespondError(w, http.StatusUnauthorized, fmt.Sprintf("Token validation failed: %s", err.Error()))
			return
		}

		// Get token type information
		tokenType, err := staticToken.GetTokenType(r.Context(), token)
		if err != nil {
			terrareg.RespondError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to determine token type: %s", err.Error()))
			return
		}

		// Return token information
		response := map[string]interface{}{
			"valid": true,
			"type":  tokenType,
		}

		terrareg.RespondJSON(w, http.StatusOK, response)
	} else {
		terrareg.RespondError(w, http.StatusInternalServerError, "Static token authentication not available")
	}
}

// HandleAuthStatus handles authentication status check
func (h *TerraformStaticTokenHandler) HandleAuthStatus(w http.ResponseWriter, r *http.Request) {
	if h.auth == nil {
		terrareg.RespondError(w, http.StatusInternalServerError, "Terraform authenticator not configured")
		return
	}

	// Check which authentication methods are enabled
	enabledMethods := h.auth.GetEnabledMethods()
	methodNames := make([]string, len(enabledMethods))
	for i, method := range enabledMethods {
		methodNames[i] = fmt.Sprintf("%d", method)
	}

	response := map[string]interface{}{
		"enabled_methods": methodNames,
		"idp_enabled":     h.auth.IsMethodEnabled(terraform.TerraformOIDCMethod),
		"static_token_enabled": h.auth.IsMethodEnabled(terraform.TerraformStaticTokenMethod),
	}

	terrareg.RespondJSON(w, http.StatusOK, response)
}