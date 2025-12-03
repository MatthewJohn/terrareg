package terraform

import (
	"net/http"

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
		terrareg.RespondError(w, http.StatusInternalServerError, "Failed to generate OpenID configuration")
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

	jwksData, err := h.idp.GetJWKS()
	if err != nil {
		terrareg.RespondError(w, http.StatusInternalServerError, "Failed to generate JWKS")
		return
	}

	terrareg.RespondJSON(w, http.StatusOK, jwksData)
}

// HandleAuth handles GET /oauth2/auth - OIDC authorization endpoint
func (h *TerraformIDPHandler) HandleAuth(w http.ResponseWriter, r *http.Request) {
	if h.idp == nil || !h.idp.IsEnabled() {
		terrareg.RespondError(w, http.StatusNotFound, "Terraform IDP is not enabled")
		return
	}

	// TODO: Implement full OIDC authorization flow when JWT library is integrated
	response := map[string]interface{}{
		"message": "OIDC authorization endpoint - implementation pending",
		"status":  "stub",
	}
	terrareg.RespondJSON(w, http.StatusOK, response)
}

// HandleToken handles POST /oauth2/token - OIDC token endpoint
func (h *TerraformIDPHandler) HandleToken(w http.ResponseWriter, r *http.Request) {
	if h.idp == nil || !h.idp.IsEnabled() {
		terrareg.RespondError(w, http.StatusNotFound, "Terraform IDP is not enabled")
		return
	}

	// TODO: Implement full OIDC token exchange when JWT library is integrated
	response, err := h.idp.HandleTokenRequest(r.Context(), map[string]interface{}{})
	if err != nil {
		terrareg.RespondError(w, http.StatusInternalServerError, "Token exchange failed")
		return
	}

	terrareg.RespondJSON(w, http.StatusOK, response)
}

// HandleUserInfo handles GET /userinfo - OIDC userinfo endpoint
func (h *TerraformIDPHandler) HandleUserInfo(w http.ResponseWriter, r *http.Request) {
	if h.idp == nil || !h.idp.IsEnabled() {
		terrareg.RespondError(w, http.StatusNotFound, "Terraform IDP is not enabled")
		return
	}

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

	userInfo, err := h.idp.HandleUserInfoRequest(r.Context(), token)
	if err != nil {
		terrareg.RespondError(w, http.StatusUnauthorized, "Invalid token")
		return
	}

	terrareg.RespondJSON(w, http.StatusOK, userInfo)
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
		"valid":        true,
		"identity_id":  "terraform-static-user",
		"subject":     "Terraform Static Token",
		"permissions":  []string{"read:modules", "read:providers"},
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