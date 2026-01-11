package auth

import (
	"context"
	"fmt"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/auth"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/provider_source/model"
	provider_source_service "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/provider_source/service"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/persistence/sqldb"
)

// GitHubAuthMethod implements GitHub OAuth authentication
// Python reference: auth/github_auth_method.py::GithubAuthMethod
type GitHubAuthMethod struct {
	providerSourceFactory *provider_source_service.ProviderSourceFactory
}

// NewGitHubAuthMethod creates a new GitHub auth method
func NewGitHubAuthMethod(providerSourceFactory *provider_source_service.ProviderSourceFactory) *GitHubAuthMethod {
	return &GitHubAuthMethod{
		providerSourceFactory: providerSourceFactory,
	}
}

// Authenticate validates GitHub session from provider_source_auth in session
// Python reference: github_auth_method.py::check_session
func (g *GitHubAuthMethod) Authenticate(ctx context.Context, sessionData map[string]interface{}) (auth.AuthContext, error) {
	// Extract provider_source from sessionData
	providerSourceName, hasProviderSource := sessionData["provider_source"].(string)
	if !hasProviderSource || providerSourceName == "" {
		return nil, nil // Not a GitHub session
	}

	// Extract github_username from sessionData
	username, hasUsername := sessionData["github_username"].(string)
	if !hasUsername || username == "" {
		return nil, nil // Invalid session
	}

	// Verify provider source exists and is a GitHub provider source
	if g.providerSourceFactory != nil {
		providerSource, err := g.providerSourceFactory.GetProviderSourceByName(ctx, providerSourceName)
		if err != nil || providerSource == nil {
			return nil, fmt.Errorf("provider source not found: %s", providerSourceName)
		}

		// Verify it's a GitHub provider source
		if providerSource.Type() != model.ProviderSourceTypeGithub {
			return nil, fmt.Errorf("provider source is not a GitHub provider: %s", providerSource.Type())
		}
	}

	// Extract organisations from sessionData
	// Python reference: github.py stores organisations as a dict of org_name -> namespace_type
	organizations := make(map[string]sqldb.NamespaceType)
	if orgsData, hasOrgs := sessionData["organisations"].(map[string]interface{}); hasOrgs {
		for orgName, namespaceType := range orgsData {
			// Convert namespace type string to NamespaceType
			if namespaceTypeStr, ok := namespaceType.(string); ok {
				// Python stores as string values like "GITHUB_ORGANISATION" or "GITHUB_USER"
				organizations[orgName] = sqldb.NamespaceType(namespaceTypeStr)
			}
		}
	}

	// If no organizations in session, add the user's own username
	// Python reference: github.py adds username with GITHUB_USER type
	if len(organizations) == 0 && username != "" {
		organizations[username] = sqldb.NamespaceTypeGithubUser
	}

	// Create GitHubAuthContext
	return auth.NewGitHubAuthContext(ctx, providerSourceName, username, organizations), nil
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
