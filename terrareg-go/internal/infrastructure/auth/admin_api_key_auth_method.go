package auth

import (
	"context"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/auth"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/config"
)

// AdminApiKeyAuthMethod implements immutable authentication for admin API keys
type AdminApiKeyAuthMethod struct {
	config *config.InfrastructureConfig
}

// NewAdminApiKeyAuthMethod creates a new immutable admin API key authentication method
func NewAdminApiKeyAuthMethod(config *config.InfrastructureConfig) *AdminApiKeyAuthMethod {
	return &AdminApiKeyAuthMethod{
		config: config,
	}
}

// GetProviderType returns the authentication provider type
func (a *AdminApiKeyAuthMethod) GetProviderType() auth.AuthMethodType {
	return auth.AuthMethodAdminApiKey
}

// IsEnabled returns whether this authentication method is enabled
func (a *AdminApiKeyAuthMethod) IsEnabled() bool {
	// Admin API key is enabled if ADMIN_AUTHENTICATION_TOKEN is configured
	return a.config.AdminAuthenticationToken != ""
}

// Authenticate authenticates an admin API key request and returns an AdminApiKeyAuthContext
func (a *AdminApiKeyAuthMethod) Authenticate(ctx context.Context, headers, formData, queryParams map[string]string) (auth.AuthMethod, error) {
	// Check if admin API key is configured
	if !a.IsEnabled() {
		return nil, nil // Let other auth methods try
	}

	// Extract API key from X-Terrareg-ApiKey header (NOT Authorization header)
	apiKey, exists := headers["X-Terrareg-ApiKey"]
	if !exists {
		return nil, nil // Let other auth methods try
	}

	// Validate API key against configured token
	if apiKey != a.config.AdminAuthenticationToken {
		return nil, nil // Let other auth methods try
	}

	// Create AdminApiKeyAuthContext with authentication state
	authContext := auth.NewAdminApiKeyAuthContext(ctx, apiKey)

	return authContext, nil
}

// AuthMethod interface implementation for the base AdminApiKeyAuthMethod
// These return default values since the actual auth state is in the AdminApiKeyAuthContext

func (a *AdminApiKeyAuthMethod) IsBuiltInAdmin() bool               { return false }
func (a *AdminApiKeyAuthMethod) IsAdmin() bool                     { return false }
func (a *AdminApiKeyAuthMethod) IsAuthenticated() bool              { return false }
func (a *AdminApiKeyAuthMethod) RequiresCSRF() bool                   { return false }
func (a *AdminApiKeyAuthMethod) CheckAuthState() bool                  { return false }
func (a *AdminApiKeyAuthMethod) CanPublishModuleVersion(string) bool { return false }
func (a *AdminApiKeyAuthMethod) CanUploadModuleVersion(string) bool  { return false }
func (a *AdminApiKeyAuthMethod) CheckNamespaceAccess(string, string) bool { return false }
func (a *AdminApiKeyAuthMethod) GetAllNamespacePermissions() map[string]string { return make(map[string]string) }
func (a *AdminApiKeyAuthMethod) GetUsername() string                { return "" }
func (a *AdminApiKeyAuthMethod) GetUserGroupNames() []string       { return []string{} }
func (a *AdminApiKeyAuthMethod) CanAccessReadAPI() bool             { return false }
func (a *AdminApiKeyAuthMethod) CanAccessTerraformAPI() bool       { return false }
func (a *AdminApiKeyAuthMethod) GetTerraformAuthToken() string     { return "" }
func (a *AdminApiKeyAuthMethod) GetProviderData() map[string]interface{} { return make(map[string]interface{}) }