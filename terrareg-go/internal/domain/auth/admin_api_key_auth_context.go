package auth

import (
	"context"
)

// AdminApiKeyAuthContext implements AuthContext for admin API key authentication
// It holds the authentication state and permission logic for admin API keys
type AdminApiKeyAuthContext struct {
	BaseAuthContext
	apiKey string
}

// NewAdminApiKeyAuthContext creates a new admin API key auth context
func NewAdminApiKeyAuthContext(ctx context.Context, apiKey string) *AdminApiKeyAuthContext {
	return &AdminApiKeyAuthContext{
		BaseAuthContext: BaseAuthContext{ctx: ctx},
		apiKey:         apiKey,
	}
}

// GetProviderType returns the authentication method type
func (a *AdminApiKeyAuthContext) GetProviderType() AuthMethodType {
	return AuthMethodAdminApiKey
}

// GetUsername returns the username for admin API key (matches Python implementation)
func (a *AdminApiKeyAuthContext) GetUsername() string {
	return "admin-api-key"
}

// IsAuthenticated returns true if the API key is valid
func (a *AdminApiKeyAuthContext) IsAuthenticated() bool {
	return a.apiKey != ""
}

// IsAdmin returns true for admin API keys
func (a *AdminApiKeyAuthContext) IsAdmin() bool {
	return true
}

// IsBuiltInAdmin returns true for admin API keys
func (a *AdminApiKeyAuthContext) IsBuiltInAdmin() bool {
	return true
}

// RequiresCSRF returns false for API key authentication
func (a *AdminApiKeyAuthContext) RequiresCSRF() bool {
	return false
}

// CheckAuthState returns true for API key authentication
func (a *AdminApiKeyAuthContext) CheckAuthState() bool {
	return true
}

// CanPublishModuleVersion returns true for admin API keys
func (a *AdminApiKeyAuthContext) CanPublishModuleVersion(module string) bool {
	return true
}

// CanUploadModuleVersion returns true for admin API keys
func (a *AdminApiKeyAuthContext) CanUploadModuleVersion(module string) bool {
	return true
}

// CheckNamespaceAccess returns true for admin API keys (all namespaces)
func (a *AdminApiKeyAuthContext) CheckNamespaceAccess(namespace, permission string) bool {
	return true
}

// GetAllNamespacePermissions returns all permissions for admin API keys
func (a *AdminApiKeyAuthContext) GetAllNamespacePermissions() map[string]string {
	return map[string]string{}
}

// GetUserGroupNames returns empty list for admin API keys
func (a *AdminApiKeyAuthContext) GetUserGroupNames() []string {
	return []string{}
}

// CanAccessReadAPI returns true for admin API keys
func (a *AdminApiKeyAuthContext) CanAccessReadAPI() bool {
	return true
}

// CanAccessTerraformAPI returns true for admin API keys
func (a *AdminApiKeyAuthContext) CanAccessTerraformAPI() bool {
	return true
}

// GetTerraformAuthToken returns empty string for admin API keys
func (a *AdminApiKeyAuthContext) GetTerraformAuthToken() string {
	return ""
}

// GetProviderData returns provider-specific data for admin API key
func (a *AdminApiKeyAuthContext) GetProviderData() map[string]interface{} {
	return map[string]interface{}{
		"auth_method": string(AuthMethodAdminApiKey),
		"is_admin":    true,
	}
}

// IsEnabled returns true for admin API key context (always enabled when created)
func (a *AdminApiKeyAuthContext) IsEnabled() bool {
	return true
}