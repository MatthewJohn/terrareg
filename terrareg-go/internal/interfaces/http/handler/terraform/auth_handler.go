package terraform

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/application/terraform"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/interfaces/http/handler/terrareg"
)

// TerraformAuthHandler handles Terraform authentication endpoints
type TerraformAuthHandler struct {
	authenticateOIDCCommand *terraform.AuthenticateOIDCTokenCommand
	validateTokenCommand    *terraform.ValidateTokenCommand
	getUserCommand          *terraform.GetUserCommand
}

// NewTerraformAuthHandler creates a new Terraform auth handler
func NewTerraformAuthHandler(
	authenticateOIDCCommand *terraform.AuthenticateOIDCTokenCommand,
	validateTokenCommand *terraform.ValidateTokenCommand,
	getUserCommand *terraform.GetUserCommand,
) *TerraformAuthHandler {
	return &TerraformAuthHandler{
		authenticateOIDCCommand: authenticateOIDCCommand,
		validateTokenCommand:    validateTokenCommand,
		getUserCommand:          getUserCommand,
	}
}

// HandleAuthenticateOIDCToken handles POST /terraform/v1/authenticate/oidc
func (h *TerraformAuthHandler) HandleAuthenticateOIDCToken(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		terrareg.RespondError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req terraform.AuthenticateOIDCTokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		terrareg.RespondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	resp, err := h.authenticateOIDCCommand.Execute(r.Context(), req)
	if err != nil {
		terrareg.RespondError(w, http.StatusUnauthorized, fmt.Sprintf("Authentication failed: %s", err.Error()))
		return
	}

	terrareg.RespondJSON(w, http.StatusOK, resp)
}

// HandleValidateToken handles POST /terraform/v1/validate/token
func (h *TerraformAuthHandler) HandleValidateToken(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		terrareg.RespondError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req terraform.ValidateTokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		terrareg.RespondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// If no authorization header provided, check it
	if req.AuthorizationHeader == "" {
		authHeader := r.Header.Get("Authorization")
		if authHeader != "" {
			req.AuthorizationHeader = authHeader
		}
	}

	// Get required permissions from query params if not in request
	if len(req.RequiredPermissions) == 0 {
		permissions := r.URL.Query()["permission"]
		if len(permissions) > 0 {
			req.RequiredPermissions = permissions
		} else {
			// Default to read permissions for Terraform registry access
			req.RequiredPermissions = []string{"read:modules", "read:providers"}
		}
	}

	resp, err := h.validateTokenCommand.Execute(r.Context(), req)
	if err != nil {
		terrareg.RespondError(w, http.StatusInternalServerError, fmt.Sprintf("Token validation failed: %s", err.Error()))
		return
	}

	terrareg.RespondJSON(w, http.StatusOK, resp)
}

// HandleGetUser handles GET /terraform/v1/users/{userID}
func (h *TerraformAuthHandler) HandleGetUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		terrareg.RespondError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Extract user ID from URL path
	path := strings.TrimPrefix(r.URL.Path, "/terraform/v1/users/")
	if path == "" || path == r.URL.Path {
		terrareg.RespondError(w, http.StatusBadRequest, "User ID is required")
		return
	}

	// Handle URL-encoded user ID
	userID, err := url.PathUnescape(path)
	if err != nil {
		terrareg.RespondError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	resp, err := h.getUserCommand.Execute(r.Context(), userID)
	if err != nil {
		terrareg.RespondError(w, http.StatusNotFound, fmt.Sprintf("User not found: %s", err.Error()))
		return
	}

	terrareg.RespondJSON(w, http.StatusOK, resp)
}

// HandleTerraformLogin handles GET /terraform/v1/login - Simplified Terraform login
func (h *TerraformAuthHandler) HandleTerraformLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		terrareg.RespondError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// For Terraform CLI authentication, we typically redirect to an external auth provider
	// For this implementation, we'll provide a simple login response
	response := map[string]interface{}{
		"login_url": "/terraform/v1/auth/oidc",
		"methods": []string{
			"oidc",
			"static_token",
		},
		"status": "configured",
	}

	terrareg.RespondJSON(w, http.StatusOK, response)
}

// HandleTerraformOIDCAuth handles GET /terraform/v1/auth/oidc - OIDC authentication start
func (h *TerraformAuthHandler) HandleTerraformOIDCAuth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		terrareg.RespondError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Extract OIDC parameters
	responseType := r.URL.Query().Get("response_type")
	clientID := r.URL.Query().Get("client_id")
	redirectURI := r.URL.Query().Get("redirect_uri")
	state := r.URL.Query().Get("state")

	// Validate required parameters
	if responseType == "" || clientID == "" || redirectURI == "" {
		terrareg.RespondError(w, http.StatusBadRequest, "Missing required OIDC parameters")
		return
	}

	// For a real implementation, this would redirect to an actual OIDC provider
	// For now, we'll simulate the authentication flow with a mock authorization code
	authCode := "mock-terraform-oidc-code"

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

	// Redirect back to Terraform with authorization code
	http.Redirect(w, r, redirectURL.String(), http.StatusFound)
}

// HandleTerraformToken handles POST /terraform/v1/auth/token - OIDC token exchange
func (h *TerraformAuthHandler) HandleTerraformToken(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		terrareg.RespondError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Parse form data
	if err := r.ParseForm(); err != nil {
		terrareg.RespondError(w, http.StatusBadRequest, "Failed to parse form data")
		return
	}

	grantType := r.FormValue("grant_type")
	code := r.FormValue("code")

	// Validate required parameters
	if grantType == "" || code == "" {
		terrareg.RespondError(w, http.StatusBadRequest, "Missing required token parameters")
		return
	}

	// For a real implementation, this would validate the authorization code
	// For now, we'll generate a mock access token
	accessToken := fmt.Sprintf("terraform-oidc-token-%s", code)

	// Authenticate the OIDC token using the integrated system
	authReq := terraform.AuthenticateOIDCTokenRequest{
		AuthorizationHeader: accessToken,
	}

	resp, err := h.authenticateOIDCCommand.Execute(r.Context(), authReq)
	if err != nil {
		terrareg.RespondError(w, http.StatusUnauthorized, fmt.Sprintf("Token authentication failed: %s", err.Error()))
		return
	}

	// Return OIDC token response
	tokenResponse := map[string]interface{}{
		"access_token": accessToken,
		"token_type":   "Bearer",
		"expires_in":   3600,
		"scope":        "openid profile email",
		"identity":     resp,
	}

	terrareg.RespondJSON(w, http.StatusOK, tokenResponse)
}
