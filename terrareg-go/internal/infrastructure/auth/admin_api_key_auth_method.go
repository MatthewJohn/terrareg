package auth

import (
	"context"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/auth"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/auth/model"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/config"
)

// AdminApiKeyAuthMethod implements admin API key authentication
// Matches Python's AdminApiKeyAuthMethod which uses X-Terrareg-ApiKey header
type AdminApiKeyAuthMethod struct {
	config *config.InfrastructureConfig
}

// NewAdminApiKeyAuthMethod creates a new admin API key auth method
func NewAdminApiKeyAuthMethod(config *config.InfrastructureConfig) *AdminApiKeyAuthMethod {
	return &AdminApiKeyAuthMethod{
		config: config,
	}
}

// GetProviderType returns the authentication provider type (matches interface)
func (a *AdminApiKeyAuthMethod) GetProviderType() auth.AuthMethodType {
	return auth.AuthMethodAdminApiKey
}

// GetAuthMethodType returns the auth method type (alias for consistency)
func (a *AdminApiKeyAuthMethod) GetAuthMethodType() auth.AuthMethodType {
	return auth.AuthMethodAdminApiKey
}

// IsEnabled checks if admin authentication is enabled
func (a *AdminApiKeyAuthMethod) IsEnabled() bool {
	return a.config.AdminAuthenticationToken != ""
}

// Authenticate validates the admin API key from X-Terrareg-ApiKey header
func (a *AdminApiKeyAuthMethod) Authenticate(ctx context.Context, apiKey string) (*auth.AuthContext, error) {
	if apiKey == "" {
		return nil, model.ErrInvalidCredentials
	}

	// Check against admin token from config
	if a.config.AdminAuthenticationToken != "" && apiKey == a.config.AdminAuthenticationToken {
		authCtx := auth.NewAuthContext(a)
		authCtx.IsAuthenticated = true
		authCtx.Username = "Admin"
		return authCtx, nil
	}

	return nil, model.ErrInvalidCredentials
}

// CheckAuthState checks if admin API key is provided
func (a *AdminApiKeyAuthMethod) CheckAuthState() bool {
	return a.config.AdminAuthenticationToken != ""
}

// CheckAuthStateWithKey checks if admin API key is provided (helper method)
func (a *AdminApiKeyAuthMethod) CheckAuthStateWithKey(apiKey string) bool {
	if apiKey == "" {
		return false
	}
	return a.config.AdminAuthenticationToken != "" && apiKey == a.config.AdminAuthenticationToken
}

// IsAuthenticated returns whether the current request is authenticated
func (a *AdminApiKeyAuthMethod) IsAuthenticated() bool {
	return true // Always returns true since this method is only used when authenticated
}

// IsAdmin returns whether the authenticated user has admin privileges
func (a *AdminApiKeyAuthMethod) IsAdmin() bool {
	return true // Admin API key always has admin privileges
}

// IsBuiltInAdmin returns whether this is a built-in admin method
func (a *AdminApiKeyAuthMethod) IsBuiltInAdmin() bool {
	return true // Admin API key method is always considered built-in admin
}

// RequiresCSRF returns whether this authentication method requires CSRF protection
func (a *AdminApiKeyAuthMethod) RequiresCSRF() bool {
	return false // API key auth doesn't need CSRF (matches Python)
}

// CanAccessReadAPI returns whether the user can access read APIs
// API keys cannot access read APIs (matches Python)
func (a *AdminApiKeyAuthMethod) CanAccessReadAPI() bool {
	return false
}

// CanAccessTerraformAPI returns whether the user can access Terraform APIs
func (a *AdminApiKeyAuthMethod) CanAccessTerraformAPI() bool {
	return true // Admin API keys can access Terraform APIs
}

// CanPublishModuleVersion checks if the user can publish module versions to the given namespace
func (a *AdminApiKeyAuthMethod) CanPublishModuleVersion(namespace string) bool {
	// Admin API keys have full access to all namespaces
	return true
}

// CanUploadModuleVersion checks if the user can upload module versions to the given namespace
func (a *AdminApiKeyAuthMethod) CanUploadModuleVersion(namespace string) bool {
	// Admin API keys have full access to all namespaces
	return true
}

// CheckNamespaceAccess checks if the user has the specified permission for a namespace
func (a *AdminApiKeyAuthMethod) CheckNamespaceAccess(permissionType, namespace string) bool {
	// Admin API keys have full access to all namespaces
	return true
}

// GetAllNamespacePermissions returns all namespace permissions for the user
func (a *AdminApiKeyAuthMethod) GetAllNamespacePermissions() map[string]string {
	// Admin API keys have full permissions to all namespaces
	// Return empty map to signify admin access
	return map[string]string{}
}

// GetUsername returns the authenticated username
func (a *AdminApiKeyAuthMethod) GetUsername() string {
	return "Admin"
}

// GetUserGroupNames returns the names of all user groups
func (a *AdminApiKeyAuthMethod) GetUserGroupNames() []string {
	// API key auth doesn't use traditional user groups
	return []string{"admin"}
}

// GetTerraformAuthToken returns the Terraform authentication token
func (a *AdminApiKeyAuthMethod) GetTerraformAuthToken() string {
	return a.config.AdminAuthenticationToken
}

// GetProviderData returns provider-specific data
func (a *AdminApiKeyAuthMethod) GetProviderData() map[string]interface{} {
	data := make(map[string]interface{})
	data["is_admin"] = true
	data["username"] = "Admin"
	return data
}
