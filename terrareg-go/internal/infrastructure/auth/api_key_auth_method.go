package auth

import (
	"context"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/auth"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/config"
)

// ApiKeyConfig holds configuration for an API key type
type ApiKeyConfig struct {
	Keys       []string
	Permission auth.ApiKeyPermissionLevel
	Username   string
	AuthMethod auth.AuthMethodType
}

// ApiKeyAuthMethod implements immutable API key authentication
// It never changes its internal state after creation
type ApiKeyAuthMethod struct {
	auth.BaseAuthMethod
	config        *config.InfrastructureConfig
	apiKeyConfigs []ApiKeyConfig
}

// NewApiKeyAuthMethod creates a new immutable API key auth method
func NewApiKeyAuthMethod(config *config.InfrastructureConfig) *ApiKeyAuthMethod {
	apiKeyConfigs := []ApiKeyConfig{
		{
			Keys:       []string{config.AdminAuthenticationToken},
			Permission: auth.ApiKeyPermissionAdmin,
			Username:   "Admin",
			AuthMethod: auth.AuthMethodAdminApiKey,
		},
		{
			Keys:       config.UploadApiKeys,
			Permission: auth.ApiKeyPermissionUpload,
			Username:   "Upload API Key",
			AuthMethod: auth.AuthMethodUploadApiKey,
		},
		{
			Keys:       config.PublishApiKeys,
			Permission: auth.ApiKeyPermissionPublish,
			Username:   "Publish API Key",
			AuthMethod: auth.AuthMethodPublishApiKey,
		},
	}

	return &ApiKeyAuthMethod{
		config:        config,
		apiKeyConfigs: apiKeyConfigs,
	}
}

// GetProviderType returns the provider type (for compatibility)
func (a *ApiKeyAuthMethod) GetProviderType() auth.AuthMethodType {
	return auth.AuthMethodAdminApiKey // Default, will be overridden by adapter
}

// IsEnabled checks if any API key authentication is enabled
func (a *ApiKeyAuthMethod) IsEnabled() bool {
	for _, config := range a.apiKeyConfigs {
		for _, key := range config.Keys {
			if key != "" {
				return true
			}
		}
	}
	return false
}

// Authenticate validates the API key and returns an ApiKeyAuthContext with authentication state
// NOTE: This method does NOT modify the AuthMethod itself
func (a *ApiKeyAuthMethod) Authenticate(ctx context.Context, apiKey string) (auth.AuthMethod, error) {
	if apiKey == "" {
		// No API key provided, return nil to let other auth methods try
		return nil, nil
	}

	// Check against all configured API keys
	for _, config := range a.apiKeyConfigs {
		for _, validKey := range config.Keys {
			if validKey != "" && apiKey == validKey {
				// Create ApiKeyAuthContext with authentication state
				authContext := auth.NewApiKeyAuthContext(ctx, config.Permission, config.Username, config.AuthMethod)
				return authContext, nil
			}
		}
	}

	// No valid key found - return nil to let other auth methods try
	return nil, nil
}

// The AuthMethod interface methods return default values for the base method
// Actual authentication state is provided by the adapter
func (a *ApiKeyAuthMethod) IsBuiltInAdmin() bool               { return false }
func (a *ApiKeyAuthMethod) IsAdmin() bool                     { return false }
func (a *ApiKeyAuthMethod) IsAuthenticated() bool              { return false }
func (a *ApiKeyAuthMethod) RequiresCSRF() bool                   { return false }
func (a *ApiKeyAuthMethod) CheckAuthState() bool                  { return false }
func (a *ApiKeyAuthMethod) CanPublishModuleVersion(string) bool { return false }
func (a *ApiKeyAuthMethod) CanUploadModuleVersion(string) bool  { return false }
func (a *ApiKeyAuthMethod) CheckNamespaceAccess(string, string) bool { return false }
func (a *ApiKeyAuthMethod) GetAllNamespacePermissions() map[string]string { return make(map[string]string) }
func (a *ApiKeyAuthMethod) GetUsername() string                { return "" }
func (a *ApiKeyAuthMethod) GetUserGroupNames() []string       { return []string{} }
func (a *ApiKeyAuthMethod) CanAccessReadAPI() bool             { return false }
func (a *ApiKeyAuthMethod) CanAccessTerraformAPI() bool       { return false }
func (a *ApiKeyAuthMethod) GetTerraformAuthToken() string     { return "" }
func (a *ApiKeyAuthMethod) GetProviderData() map[string]interface{} { return make(map[string]interface{}) }