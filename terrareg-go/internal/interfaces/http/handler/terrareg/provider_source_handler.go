package terrareg

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog/log"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/auth"
	provider_source_service "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/provider_source/service"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/persistence/sqldb"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/interfaces/http/middleware"
)

// ProviderSourceFactoryInterface defines the interface for provider source factory
// This allows for mocking in tests
type ProviderSourceFactoryInterface interface {
	GetProviderSourceByApiName(ctx context.Context, apiName string) (provider_source_service.ProviderSourceInstance, error)
}

// SessionManagementServiceInterface defines the interface for session management
// This allows for mocking in tests
type SessionManagementServiceInterface interface {
	CreateSessionAndCookie(ctx context.Context, w http.ResponseWriter, authMethod auth.AuthMethodType, username string, isAdmin bool, userGroups []string, permissions map[string]string, providerData map[string]interface{}, ttl *time.Duration) error
}

// ProviderSourceHandler handles provider source OAuth flow
// Python reference: server/api/github/github_login_callback.py
type ProviderSourceHandler struct {
	providerSourceFactory    ProviderSourceFactoryInterface
	sessionManagementService SessionManagementServiceInterface
}

// NewProviderSourceHandler creates a new provider source handler
func NewProviderSourceHandler(
	providerSourceFactory ProviderSourceFactoryInterface,
	sessionManagementService SessionManagementServiceInterface,
) *ProviderSourceHandler {
	return &ProviderSourceHandler{
		providerSourceFactory:    providerSourceFactory,
		sessionManagementService: sessionManagementService,
	}
}

// HandleLogin handles GET /{provider_source}/login
// Redirects to the provider's OAuth authorization URL
// Python reference: server/api/github/github_login.py
func (h *ProviderSourceHandler) HandleLogin(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Extract provider_source from URL
	providerSourceName := chi.URLParam(r, "provider_source")
	if providerSourceName == "" {
		h.writeError(w, http.StatusBadRequest, "provider_source is required")
		return
	}

	// Get provider source from factory
	providerSource, err := h.providerSourceFactory.GetProviderSourceByApiName(ctx, providerSourceName)
	if err != nil || providerSource == nil {
		log.Error().Err(err).Str("provider_source", providerSourceName).Msg("Failed to get provider source")
		h.writeError(w, http.StatusNotFound, "Provider source not found")
		return
	}

	// Get login redirect URL
	loginURL, err := providerSource.GetLoginRedirectURL(ctx)
	if err != nil {
		log.Error().Err(err).Str("provider_source", providerSourceName).Msg("Failed to get login redirect URL")
		h.writeError(w, http.StatusInternalServerError, "Failed to generate login URL")
		return
	}

	// Redirect to provider's OAuth page
	http.Redirect(w, r, loginURL, http.StatusFound)
}

// HandleCallback handles GET /{provider_source}/callback
// Exchanges OAuth code for access token, fetches user info, creates session
// Python reference: server/api/github/github_login_callback.py
func (h *ProviderSourceHandler) HandleCallback(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Extract provider_source from URL
	providerSourceName := chi.URLParam(r, "provider_source")
	if providerSourceName == "" {
		h.writeError(w, http.StatusBadRequest, "provider_source is required")
		return
	}

	// Extract authorization code from query params
	code := r.URL.Query().Get("code")
	if code == "" {
		h.writeError(w, http.StatusBadRequest, "authorization code is required")
		return
	}

	// Get provider source from factory
	providerSource, err := h.providerSourceFactory.GetProviderSourceByApiName(ctx, providerSourceName)
	if err != nil || providerSource == nil {
		log.Error().Err(err).Str("provider_source", providerSourceName).Msg("Failed to get provider source")
		h.writeError(w, http.StatusNotFound, "Provider source not found")
		return
	}

	// Exchange code for access token
	accessToken, err := providerSource.GetUserAccessToken(ctx, code)
	if err != nil || accessToken == "" {
		log.Error().Err(err).Str("provider_source", providerSourceName).Msg("Failed to get access token")
		h.writeError(w, http.StatusUnauthorized, "Invalid authorization code")
		return
	}

	// Get username
	username, err := providerSource.GetUsername(ctx, accessToken)
	if err != nil || username == "" {
		log.Error().Err(err).Str("provider_source", providerSourceName).Msg("Failed to get username")
		h.writeError(w, http.StatusUnauthorized, "Failed to get user information")
		return
	}

	// Get organizations (admin role only)
	organizations := providerSource.GetUserOrganizations(ctx, accessToken)

	// Build organizations map with namespace types
	organizationsMap := make(map[string]sqldb.NamespaceType)
	for _, org := range organizations {
		organizationsMap[org] = sqldb.NamespaceTypeGithubOrg
	}
	// Add user's own username as GITHUB_USER namespace
	organizationsMap[username] = sqldb.NamespaceTypeGithubUser

	// Create GitHubAuthContext from authentication data
	githubAuthCtx := auth.NewGitHubAuthContext(ctx, providerSourceName, username, organizationsMap)

	// Create session with 24 hour TTL (matching Python default)
	ttl := 24 * time.Hour
	err = h.sessionManagementService.CreateSessionAndCookie(
		ctx, w, githubAuthCtx.GetProviderType(), githubAuthCtx.GetUsername(),
		githubAuthCtx.IsAdmin(), githubAuthCtx.GetUserGroupNames(),
		githubAuthCtx.GetAllNamespacePermissions(), githubAuthCtx.GetProviderData(), &ttl,
	)
	if err != nil {
		log.Error().Err(err).Str("provider_source", providerSourceName).Str("username", username).Msg("Failed to create session")
		h.writeError(w, http.StatusInternalServerError, "Failed to create session")
		return
	}

	log.Info().
		Str("provider_source", providerSourceName).
		Str("username", username).
		Msg("GitHub OAuth login successful")

	// Redirect to homepage (matching Python behavior)
	http.Redirect(w, r, "/", http.StatusFound)
}

// HandleAuthStatus handles GET /{provider_source}/auth/status
// Returns authentication status for the provider source
func (h *ProviderSourceHandler) HandleAuthStatus(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Extract provider_source from URL
	providerSourceName := chi.URLParam(r, "provider_source")
	if providerSourceName == "" {
		h.writeError(w, http.StatusBadRequest, "provider_source is required")
		return
	}

	// Get provider source from factory
	providerSource, err := h.providerSourceFactory.GetProviderSourceByApiName(ctx, providerSourceName)
	if err != nil || providerSource == nil {
		log.Error().Err(err).Str("provider_source", providerSourceName).Msg("Failed to get provider source")
		h.writeError(w, http.StatusNotFound, "Provider source not found")
		return
	}

	// Get authentication status from request context
	// The middleware should have set the auth context
	response := map[string]interface{}{
		"provider_type": string(providerSource.Type()),
	}

	// Try to get auth context from middleware
	authCtx := middleware.GetSessionData(r.Context())
	if authCtx != nil && authCtx.IsAuthenticated() {
		response["authenticated"] = true
		response["username"] = authCtx.GetUsername()

		// Only include auth_method for provider source authentication types (not built-in admin)
		authMethod := string(authCtx.GetProviderType())
		if authMethod != "" && authMethod != string(auth.AuthMethodNotAuthenticated) {
			response["auth_method"] = authMethod
		}
	} else {
		response["authenticated"] = false
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// writeError writes an error response
func (h *ProviderSourceHandler) writeError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"error": message,
	})
}
