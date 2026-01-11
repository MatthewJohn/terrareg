package auth

import (
	"context"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/persistence/sqldb"
)

// GitHubAuthContext implements AuthContext for GitHub OAuth authentication
// Python reference: auth/github_auth_method.py::GithubAuthMethod
type GitHubAuthContext struct {
	BaseAuthContext
	providerSourceName string
	username           string
	organizations      map[string]sqldb.NamespaceType // org name -> namespace type
}

// NewGitHubAuthContext creates a new GitHub auth context
func NewGitHubAuthContext(
	ctx context.Context,
	providerSourceName string,
	username string,
	organizations map[string]sqldb.NamespaceType,
) *GitHubAuthContext {
	return &GitHubAuthContext{
		BaseAuthContext:     BaseAuthContext{ctx: ctx},
		providerSourceName: providerSourceName,
		username:           username,
		organizations:      organizations,
	}
}

// GetProviderType returns the authentication method type
func (g *GitHubAuthContext) GetProviderType() AuthMethodType {
	return AuthMethodGitHub
}

// GetUsername returns the GitHub username
func (g *GitHubAuthContext) GetUsername() string {
	return g.username
}

// GetGroupMemberships returns the list of organizations the user belongs to
// Python reference: github_auth_method.py::get_group_memberships
func (g *GitHubAuthContext) GetGroupMemberships() []string {
	groups := make([]string, 0, len(g.organizations))
	for org := range g.organizations {
		groups = append(groups, org)
	}
	return groups
}

// IsAuthenticated returns true if the user is authenticated
func (g *GitHubAuthContext) IsAuthenticated() bool {
	return g.username != "" && g.providerSourceName != ""
}

// IsAdmin returns false for GitHub users (not site admins)
func (g *GitHubAuthContext) IsAdmin() bool {
	return false
}

// IsBuiltInAdmin returns false for GitHub users
func (g *GitHubAuthContext) IsBuiltInAdmin() bool {
	return false
}

// RequiresCSRF returns true for session-based authentication
func (g *GitHubAuthContext) RequiresCSRF() bool {
	return true
}

// IsEnabled returns true if the context is valid
func (g *GitHubAuthContext) IsEnabled() bool {
	return g.IsAuthenticated()
}

// CheckAuthState returns true if the context is in a valid state
func (g *GitHubAuthContext) CheckAuthState() bool {
	return g.IsAuthenticated()
}

// GetProviderSourceName returns the provider source name
func (g *GitHubAuthContext) GetProviderSourceName() string {
	return g.providerSourceName
}

// CheckNamespaceAccess checks if the user has access to a namespace
// Python reference: github_auth_method.py::check_namespace_access
func (g *GitHubAuthContext) CheckNamespaceAccess(permissionType, namespace string) bool {
	// Check if the namespace matches one of the user's organizations
	// Uses case-insensitive matching as per Python implementation
	namespaceLower := toLower(namespace)
	for org := range g.organizations {
		if toLower(org) == namespaceLower {
			return true
		}
	}

	return false
}

// GetAllNamespacePermissions returns all namespace permissions for the user
// Python reference: github_auth_method.py::get_all_namespace_permissions
func (g *GitHubAuthContext) GetAllNamespacePermissions() map[string]string {
	result := make(map[string]string)

	// Add permissions for each organization
	// Auto-generate namespaces means the user gets FULL access to their orgs
	for org := range g.organizations {
		result[org] = "FULL"
	}

	// Add the user's own username as a namespace
	if g.username != "" {
		result[g.username] = "FULL"
	}

	return result
}

// GetProviderData returns provider-specific data for the session
func (g *GitHubAuthContext) GetProviderData() map[string]interface{} {
	return map[string]interface{}{
		"provider_source": g.providerSourceName,
		"github_username": g.username,
		"organisations":   g.organizations,
		"auth_method":     string(AuthMethodGitHub),
	}
}

// CanPublishModuleVersion checks if the user can publish to a namespace
func (g *GitHubAuthContext) CanPublishModuleVersion(namespace string) bool {
	return g.CheckNamespaceAccess("FULL", namespace)
}

// CanUploadModuleVersion checks if the user can upload to a namespace
func (g *GitHubAuthContext) CanUploadModuleVersion(namespace string) bool {
	return g.CheckNamespaceAccess("FULL", namespace)
}

// CanAccessReadAPI returns true for authenticated users
func (g *GitHubAuthContext) CanAccessReadAPI() bool {
	return g.IsAuthenticated()
}

// CanAccessTerraformAPI returns true for authenticated users
func (g *GitHubAuthContext) CanAccessTerraformAPI() bool {
	return g.IsAuthenticated()
}

// GetTerraformAuthToken returns empty string for GitHub OAuth
func (g *GitHubAuthContext) GetTerraformAuthToken() string {
	return ""
}

// GetUserGroupNames returns the list of organizations
func (g *GitHubAuthContext) GetUserGroupNames() []string {
	return g.GetGroupMemberships()
}

// toLower converts a string to lowercase for case-insensitive comparison
func toLower(s string) string {
	// Simple lowercase conversion
	result := make([]rune, len(s))
	for i, r := range s {
		if r >= 'A' && r <= 'Z' {
			result[i] = r + 32
		} else {
			result[i] = r
		}
	}
	return string(result)
}
