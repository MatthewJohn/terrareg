package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/auth"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/auth/repository"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/provider_source/model"
	provider_source_service "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/provider_source/service"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/persistence/sqldb"
)

// GitHubAuthMethod implements GitHub OAuth authentication
// Python reference: auth/github_auth_method.py::GithubAuthMethod
type GitHubAuthMethod struct {
	providerSourceFactory *provider_source_service.ProviderSourceFactory
	sessionRepo           repository.SessionRepository
}

// NewGitHubAuthMethod creates a new GitHub auth method
func NewGitHubAuthMethod(
	providerSourceFactory *provider_source_service.ProviderSourceFactory,
	sessionRepo repository.SessionRepository,
) *GitHubAuthMethod {
	return &GitHubAuthMethod{
		providerSourceFactory: providerSourceFactory,
		sessionRepo:           sessionRepo,
	}
}

// Authenticate validates GitHub session from provider_source_auth in session
// Python reference: github_auth_method.py::check_session
func (g *GitHubAuthMethod) Authenticate(ctx context.Context, sessionData map[string]interface{}) (auth.AuthContext, error) {
	// Get session ID from sessionData map (populated by auth factory from cookies/headers)
	sessionIDInterface, exists := sessionData["session_id"]
	if !exists {
		return nil, nil // No session ID, let other auth methods try
	}

	sessionID, ok := sessionIDInterface.(string)
	if !ok || sessionID == "" || strings.TrimSpace(sessionID) == "" {
		return nil, nil // Invalid session ID, let other auth methods try
	}

	// Find session in database
	session, err := g.sessionRepo.FindByID(ctx, sessionID)
	if err != nil || session == nil {
		return nil, nil // Let other auth methods try
	}

	// Check if session is expired
	if session.IsExpired() {
		return nil, nil // Let other auth methods try
	}

	// Parse provider_source_auth JSON to get GitHub user information
	// Python reference: flask.session contains github_username, provider_source, organisations
	githubData, err := g.parseGitHubProviderData(session.ProviderSourceAuth)
	if err != nil {
		return nil, nil // Let other auth methods try
	}

	// Verify provider source exists and is a GitHub provider source
	if g.providerSourceFactory != nil {
		providerSource, err := g.providerSourceFactory.GetProviderSourceByApiName(ctx, githubData.ProviderSource)
		if err != nil || providerSource == nil {
			return nil, nil // Let other auth methods try
		}

		// Verify it's a GitHub provider source
		if providerSource.Type() != model.ProviderSourceTypeGithub {
			return nil, fmt.Errorf("provider source is not a GitHub provider: %s", providerSource.Type())
		}
	}

	// Create GitHubAuthContext
	return auth.NewGitHubAuthContext(ctx, githubData.ProviderSource, githubData.Username, githubData.Organizations), nil
}

// GitHubProviderData represents the GitHub provider data stored in session
type GitHubProviderData struct {
	ProviderSource string                         `json:"provider_source"`
	Username       string                         `json:"github_username"`
	Organizations  map[string]sqldb.NamespaceType `json:"organisations"`
}

// parseGitHubProviderData parses the provider_source_auth JSON to extract GitHub user information
func (g *GitHubAuthMethod) parseGitHubProviderData(providerSourceAuth []byte) (*GitHubProviderData, error) {
	if len(providerSourceAuth) == 0 {
		return nil, fmt.Errorf("no provider data")
	}

	var data GitHubProviderData
	err := json.Unmarshal(providerSourceAuth, &data)
	if err != nil {
		return nil, fmt.Errorf("failed to parse provider data: %w", err)
	}

	// If no organizations in session, add the user's own username
	// Python reference: github.py adds username with GITHUB_USER type
	if len(data.Organizations) == 0 && data.Username != "" {
		if data.Organizations == nil {
			data.Organizations = make(map[string]sqldb.NamespaceType)
		}
		data.Organizations[data.Username] = sqldb.NamespaceTypeGithubUser
	}

	return &data, nil
}

// GetProviderType returns the authentication method type
func (g *GitHubAuthMethod) GetProviderType() auth.AuthMethodType {
	return auth.AuthMethodGitHub
}

// IsEnabled checks if GitHub authentication is enabled
// Python reference: github_auth_method.py::is_enabled
func (g *GitHubAuthMethod) IsEnabled() bool {
	// GitHub authentication is always enabled if provider source factory is available
	// The actual validation happens when checking if a provider source exists
	return g.providerSourceFactory != nil
}
