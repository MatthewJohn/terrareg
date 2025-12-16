package terrareg

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net/http"

	"github.com/rs/zerolog/log"

	authCmd "github.com/matthewjohn/terrareg/terrareg-go/internal/application/command/auth"
	authQuery "github.com/matthewjohn/terrareg/terrareg-go/internal/application/query/auth"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/auth/service"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/config"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/interfaces/http/dto"
)

// AuthHandler handles authentication-related requests
type AuthHandler struct {
	adminLoginCmd        *authCmd.AdminLoginCommand
	checkSessionQuery    *authQuery.CheckSessionQuery
	isAuthenticatedQuery *authQuery.IsAuthenticatedQuery
	oidcLoginCmd         *authCmd.OidcLoginCommand
	oidcCallbackCmd      *authCmd.OidcCallbackCommand
	samlLoginCmd         *authCmd.SamlLoginCommand
	samlMetadataCmd      *authCmd.SamlMetadataCommand
	githubOAuthCmd       *authCmd.GithubOAuthCommand
	authService          *service.AuthenticationService
	stateStorageService  *service.StateStorageService
	infraConfig          *config.InfrastructureConfig
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(
	adminLoginCmd *authCmd.AdminLoginCommand,
	checkSessionQuery *authQuery.CheckSessionQuery,
	isAuthenticatedQuery *authQuery.IsAuthenticatedQuery,
	oidcLoginCmd *authCmd.OidcLoginCommand,
	oidcCallbackCmd *authCmd.OidcCallbackCommand,
	samlLoginCmd *authCmd.SamlLoginCommand,
	samlMetadataCmd *authCmd.SamlMetadataCommand,
	githubOAuthCmd *authCmd.GithubOAuthCommand,
	authService *service.AuthenticationService,
	stateStorageService *service.StateStorageService,
	infraConfig *config.InfrastructureConfig,
) *AuthHandler {
	return &AuthHandler{
		adminLoginCmd:        adminLoginCmd,
		checkSessionQuery:    checkSessionQuery,
		isAuthenticatedQuery: isAuthenticatedQuery,
		oidcLoginCmd:         oidcLoginCmd,
		oidcCallbackCmd:      oidcCallbackCmd,
		samlLoginCmd:         samlLoginCmd,
		samlMetadataCmd:      samlMetadataCmd,
		githubOAuthCmd:       githubOAuthCmd,
		authService:          authService,
		stateStorageService:  stateStorageService,
		infraConfig:          infraConfig,
	}
}

// HandleAdminLogin handles POST /v1/terrareg/auth/admin/login
// Now follows DDD principles and matches Python API contract
func (h *AuthHandler) HandleAdminLogin(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Only allow POST requests
	if r.Method != http.MethodPost {
		RespondJSON(w, http.StatusMethodNotAllowed, map[string]interface{}{
			"message": "Method not allowed",
		})
		return
	}

	// Execute admin login command using the request (extracts X-Terrareg-ApiKey header)
	response, err := h.adminLoginCmd.ExecuteWithRequest(ctx, r)
	if err != nil {
		// Handle different error scenarios appropriately
		if !h.adminLoginCmd.IsConfigured() {
			RespondJSON(w, http.StatusForbidden, map[string]interface{}{
				"message": "Admin authentication is not enabled",
			})
		} else {
			RespondJSON(w, http.StatusUnauthorized, map[string]interface{}{
				"message": "Invalid API key",
			})
		}
		return
	}

	// If authentication failed, return 401
	if !response.Authenticated {
		RespondJSON(w, http.StatusUnauthorized, map[string]interface{}{
			"message": "Authentication failed",
		})
		return
	}

	// Create admin session using the authentication service (DDD-compliant approach)
	// The authentication service orchestrates session and cookie operations
	log.Info().
		Str("session_id", response.SessionID).
		Msg("Creating admin session")

	if err := h.authService.CreateAdminSession(ctx, w, response.SessionID); err != nil {
		// If session creation fails, return error
		log.Error().
			Err(err).
			Str("session_id", response.SessionID).
			Msg("Failed to create admin session")

		RespondJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"message": "Failed to create session",
		})
		return
	}

	log.Info().
		Str("session_id", response.SessionID).
		Msg("Admin session created successfully")

	// Respond with Python-compatible format
	RespondJSON(w, http.StatusOK, map[string]interface{}{
		"authenticated": response.Authenticated,
	})
}

// HandleIsAuthenticated handles GET /v1/terrareg/auth/admin/is_authenticated
func (h *AuthHandler) HandleIsAuthenticated(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get authentication status using the query
	response, err := h.isAuthenticatedQuery.Execute(ctx)
	if err != nil {
		// If there's an error, return unauthenticated status
		response = &dto.IsAuthenticatedResponse{
			Authenticated:        false,
			ReadAccess:           false,
			SiteAdmin:            false,
			NamespacePermissions: make(map[string]string),
		}
	}

	RespondJSON(w, http.StatusOK, response)
}

// HandleOIDCLogin handles GET /v1/terrareg/auth/oidc/login
func (h *AuthHandler) HandleOIDCLogin(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get redirect URL from query parameter
	redirectURL := r.URL.Query().Get("redirect_url")
	if redirectURL == "" {
		redirectURL = "/"
	}

	// Generate state parameter for CSRF protection
	state := generateRandomState()

	// Execute OIDC login command
	response, err := h.oidcLoginCmd.Execute(ctx, &authCmd.OidcLoginRequest{
		RedirectURL: redirectURL,
		State:       state,
	})
	if err != nil {
		RespondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Store state using secure state storage
	if h.stateStorageService != nil {
		// Generate and store secure state for CSRF protection
		redirectURL := getRedirectURL(r)
		stateParam, err := h.stateStorageService.GenerateAndStoreState(ctx, redirectURL, "oidc")
		if err != nil {
			RespondError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to generate secure state: %v", err))
			return
		}

		// Use the generated state instead of the one from response
		response.State = stateParam
	}

	// Redirect to OIDC provider
	http.Redirect(w, r, response.AuthURL, http.StatusFound)
}

// HandleOIDCCallback handles GET /v1/terrareg/auth/oidc/callback
func (h *AuthHandler) HandleOIDCCallback(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get authorization code and state from query parameters
	code := r.URL.Query().Get("code")
	state := r.URL.Query().Get("state")
	errorParam := r.URL.Query().Get("error")
	errorDescription := r.URL.Query().Get("error_description")

	// Execute OIDC callback command
	response, err := h.oidcCallbackCmd.Execute(ctx, &authCmd.OidcCallbackRequest{
		Code:             code,
		State:            state,
		Error:            errorParam,
		ErrorDescription: errorDescription,
	})
	if err != nil {
		RespondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Check if authentication failed
	if !response.Authenticated {
		RespondError(w, http.StatusUnauthorized, response.ErrorMessage)
		return
	}

	// Create session using auth service
	if err := h.authService.CreateSession(ctx, w, response.SessionID); err != nil {
		RespondError(w, http.StatusInternalServerError, "Failed to create session")
		return
	}

	// Redirect to application
	http.Redirect(w, r, "/", http.StatusFound)
}

// HandleSAMLLogin handles GET /v1/terrareg/auth/saml/login
func (h *AuthHandler) HandleSAMLLogin(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get redirect URL from query parameter
	redirectURL := r.URL.Query().Get("redirect_url")
	if redirectURL == "" {
		redirectURL = "/"
	}

	// Execute SAML login command
	response, err := h.samlLoginCmd.Execute(ctx, &authCmd.SamlLoginRequest{
		RelayState: redirectURL,
	})
	if err != nil {
		RespondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Redirect to SAML IDP
	http.Redirect(w, r, response.AuthURL, http.StatusFound)
}

// HandleSAMLMetadata handles GET /v1/terrareg/auth/saml/metadata
func (h *AuthHandler) HandleSAMLMetadata(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Execute SAML metadata command
	response, err := h.samlMetadataCmd.Execute(ctx, &authCmd.SamlMetadataRequest{})
	if err != nil {
		RespondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Return SAML metadata XML
	w.Header().Set("Content-Type", h.samlMetadataCmd.GetMetadataContentType())
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(response.Metadata))
}

// HandleGitHubOAuth handles GET /v1/terrareg/auth/github/oauth
func (h *AuthHandler) HandleGitHubOAuth(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get redirect URL from query parameter
	redirectURL := r.URL.Query().Get("redirect_url")
	if redirectURL == "" {
		redirectURL = "/"
	}

	// Generate state parameter for CSRF protection
	state := generateRandomState()

	// Execute GitHub OAuth login command
	response, err := h.githubOAuthCmd.ExecuteLogin(ctx, &authCmd.GithubOAuthLoginRequest{
		RedirectURL: redirectURL,
		State:       state,
	})
	if err != nil {
		RespondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Store state using secure state storage
	if h.stateStorageService != nil {
		// Generate and store secure state for CSRF protection
		redirectURL := getRedirectURL(r)
		stateParam, err := h.stateStorageService.GenerateAndStoreState(ctx, redirectURL, "github")
		if err != nil {
			RespondError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to generate secure state: %v", err))
			return
		}

		// Use the generated state instead of the one from response
		response.State = stateParam
	}

	// Redirect to GitHub OAuth provider
	http.Redirect(w, r, response.AuthURL, http.StatusFound)
}

// HandleLogout handles POST /v1/terrareg/auth/logout
func (h *AuthHandler) HandleLogout(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Clear session
	if err := h.authService.ClearSession(ctx, w, r); err != nil {
		RespondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Return success response
	RespondJSON(w, http.StatusOK, map[string]interface{}{
		"message": "Successfully logged out",
	})
}

// generateRandomState generates a random state parameter for CSRF protection
func generateRandomState() string {
	b := make([]byte, 32)
	rand.Read(b)
	return base64.StdEncoding.EncodeToString(b)
}

// getRedirectURL extracts the redirect URL from the request
func getRedirectURL(r *http.Request) string {
	redirectURL := r.URL.Query().Get("redirect_url")
	if redirectURL == "" {
		redirectURL = "/"
	}
	return redirectURL
}
