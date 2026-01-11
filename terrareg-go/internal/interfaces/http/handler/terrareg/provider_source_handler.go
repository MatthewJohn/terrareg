package terrareg

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog/log"

	authservice "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/auth/service"
	provider_source_service "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/provider_source/service"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/persistence/sqldb"
)

// ProviderSourceFactoryInterface defines the interface for provider source factory
// This allows for mocking in tests
type ProviderSourceFactoryInterface interface {
	GetProviderSourceByApiName(ctx context.Context, apiName string) (provider_source_service.ProviderSourceInstance, error)
}

// AuthenticationServiceInterface defines the interface for authentication service
// This allows for mocking in tests
type AuthenticationServiceInterface interface {
	CreateAuthenticatedSession(ctx context.Context, w http.ResponseWriter, authMethod string, providerData map[string]interface{}, ttl *time.Duration) error
	ValidateRequest(ctx context.Context, r *http.Request) (*authservice.AuthenticationContext, error)
}

// ProviderSourceHandler handles provider source OAuth flow
// Python reference: server/api/github/github_login_callback.py
type ProviderSourceHandler struct {
	providerSourceFactory ProviderSourceFactoryInterface
	authService           AuthenticationServiceInterface
}

// NewProviderSourceHandler creates a new provider source handler
func NewProviderSourceHandler(
	providerSourceFactory ProviderSourceFactoryInterface,
	authService AuthenticationServiceInterface,
) *ProviderSourceHandler {
	return &ProviderSourceHandler{
		providerSourceFactory: providerSourceFactory,
		authService:           authService,
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
	orgsMap := make(map[string]string)
	for _, org := range organizations {
		orgsMap[org] = string(sqldb.NamespaceTypeGithubOrg)
	}
	// Add user's own username as GITHUB_USER namespace
	orgsMap[username] = string(sqldb.NamespaceTypeGithubUser)

	// Build provider data for session
	// This matches the Python implementation structure
	providerData := map[string]interface{}{
		"provider_source": providerSourceName,
		"github_username": username,
		"organisations":   orgsMap,
		"provider_source_auth": map[string]interface{}{
			"github_access_token": accessToken,
		},
	}

	// Create session with 24 hour TTL (matching Python default)
	ttl := 24 * time.Hour
	err = h.authService.CreateAuthenticatedSession(ctx, w, "GITHUB", providerData, &ttl)
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

	// Check authentication status
	authCtx, err := h.authService.ValidateRequest(ctx, r)
	if err != nil {
		log.Error().Err(err).Msg("Failed to validate request")
		h.writeError(w, http.StatusInternalServerError, "Failed to validate authentication")
		return
	}

	response := map[string]interface{}{
		"authenticated": authCtx.IsAuthenticated,
		"provider_type": string(providerSource.Type()),
	}

	if authCtx.IsAuthenticated {
		response["username"] = authCtx.Username
		response["auth_method"] = authCtx.AuthMethod
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
