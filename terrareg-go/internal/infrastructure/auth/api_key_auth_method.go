package auth

import (
	"context"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/auth"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/auth/model"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/config"
)

// ApiKeyPermissionLevel represents the permission level for an API key
type ApiKeyPermissionLevel int

const (
	// ApiKeyPermissionUpload allows uploading modules
	ApiKeyPermissionUpload ApiKeyPermissionLevel = iota
	// ApiKeyPermissionPublish allows publishing modules (includes upload)
	ApiKeyPermissionPublish
	// ApiKeyPermissionAdmin allows full admin access
	ApiKeyPermissionAdmin
)

// ApiKeyConfig holds configuration for an API key type
type ApiKeyConfig struct {
	Keys       []string
	Permission ApiKeyPermissionLevel
	Username   string
	AuthMethod auth.AuthMethodType
}

// ApiKeyAuthMethod implements generic API key authentication
// Consolidates AdminApiKeyAuthMethod, UploadApiKeyAuthMethod, and PublishApiKeyAuthMethod
// It implements the auth.AuthMethod interface
type ApiKeyAuthMethod struct {
	config        *config.InfrastructureConfig
	apiKeyConfigs []ApiKeyConfig
	// Store authentication state
	currentKey   string
	currentConfig *ApiKeyConfig
	authenticated bool
}

// NewApiKeyAuthMethod creates a new generic API key auth method
func NewApiKeyAuthMethod(config *config.InfrastructureConfig) *ApiKeyAuthMethod {
	apiKeyConfigs := []ApiKeyConfig{
		{
			Keys:       []string{config.AdminAuthenticationToken},
			Permission: ApiKeyPermissionAdmin,
			Username:   "Admin",
			AuthMethod: auth.AuthMethodAdminApiKey,
		},
		{
			Keys:       config.UploadApiKeys,
			Permission: ApiKeyPermissionUpload,
			Username:   "Upload API Key",
			AuthMethod: auth.AuthMethodUploadApiKey,
		},
		{
			Keys:       config.PublishApiKeys,
			Permission: ApiKeyPermissionPublish,
			Username:   "Publish API Key",
			AuthMethod: auth.AuthMethodPublishApiKey,
		},
	}

	return &ApiKeyAuthMethod{
		config:        config,
		apiKeyConfigs: apiKeyConfigs,
	}
}

// Authenticate validates the API key and returns an AuthContext
func (a *ApiKeyAuthMethod) Authenticate(ctx context.Context, apiKey string) (*auth.AuthContext, error) {
	if apiKey == "" {
		return nil, model.ErrInvalidCredentials
	}

	// Check against all configured API keys
	for _, config := range a.apiKeyConfigs {
		for _, validKey := range config.Keys {
			if validKey != "" && apiKey == validKey {
				// Store the authenticated state
				a.currentKey = apiKey
				a.currentConfig = &config
				a.authenticated = true

				// Create AuthContext with this auth method
				return auth.NewAuthContext(a), nil
			}
		}
	}

	// Clear any previous authentication state
	a.currentKey = ""
	a.currentConfig = nil
	a.authenticated = false

	return nil, model.ErrInvalidCredentials
}

// AuthMethod interface implementation

// IsBuiltInAdmin returns true if this is a built-in admin key
func (a *ApiKeyAuthMethod) IsBuiltInAdmin() bool {
	return a.authenticated && a.currentConfig != nil && a.currentConfig.Permission == ApiKeyPermissionAdmin
}

// IsAdmin returns true if the authenticated key has admin privileges
func (a *ApiKeyAuthMethod) IsAdmin() bool {
	return a.IsBuiltInAdmin()
}

// IsAuthenticated returns true if a valid API key was provided
func (a *ApiKeyAuthMethod) IsAuthenticated() bool {
	return a.authenticated
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

// RequiresCSRF returns false for API key authentication
func (a *ApiKeyAuthMethod) RequiresCSRF() bool {
	return false
}

// CheckAuthState returns false for API key authentication (no state to check)
func (a *ApiKeyAuthMethod) CheckAuthState() bool {
	return false
}

// CanPublishModuleVersion checks if the key can publish modules
func (a *ApiKeyAuthMethod) CanPublishModuleVersion(namespace string) bool {
	if !a.authenticated || a.currentConfig == nil {
		return false
	}

	// Admin can publish anything
	if a.currentConfig.Permission == ApiKeyPermissionAdmin {
		return true
	}

	// Publish keys can publish
	if a.currentConfig.Permission == ApiKeyPermissionPublish {
		return true
	}

	// Upload keys cannot publish
	return false
}

// CanUploadModuleVersion checks if the key can upload modules
func (a *ApiKeyAuthMethod) CanUploadModuleVersion(namespace string) bool {
	if !a.authenticated || a.currentConfig == nil {
		return false
	}

	// Admin can upload anything
	if a.currentConfig.Permission == ApiKeyPermissionAdmin {
		return true
	}

	// Publish and upload keys can upload
	if a.currentConfig.Permission == ApiKeyPermissionPublish || a.currentConfig.Permission == ApiKeyPermissionUpload {
		return true
	}

	return false
}

// CheckNamespaceAccess checks namespace access based on permission level
func (a *ApiKeyAuthMethod) CheckNamespaceAccess(permissionType, namespace string) bool {
	if !a.authenticated || a.currentConfig == nil {
		return false
	}

	// Admin has access to everything
	if a.currentConfig.Permission == ApiKeyPermissionAdmin {
		return true
	}

	// For publish/upload keys, they have access to all namespaces with their respective permissions
	switch permissionType {
	case "READ":
		return true // All authenticated keys can read
	case "UPLOAD":
		return a.currentConfig.Permission == ApiKeyPermissionUpload || a.currentConfig.Permission == ApiKeyPermissionPublish
	case "PUBLISH", "MODIFY":
		return a.currentConfig.Permission == ApiKeyPermissionPublish
	case "FULL":
		return a.currentConfig.Permission == ApiKeyPermissionAdmin
	}

	return false
}

// GetAllNamespacePermissions returns permissions for all namespaces
func (a *ApiKeyAuthMethod) GetAllNamespacePermissions() map[string]string {
	if !a.authenticated || a.currentConfig == nil {
		return make(map[string]string)
	}

	// API keys have the same permissions across all namespaces
	// Return empty map to indicate no namespace-specific restrictions
	return make(map[string]string)
}

// GetUsername returns the username for the authenticated key
func (a *ApiKeyAuthMethod) GetUsername() string {
	if !a.authenticated || a.currentConfig == nil {
		return ""
	}
	return a.currentConfig.Username
}

// GetUserGroupNames returns empty list for API key authentication
func (a *ApiKeyAuthMethod) GetUserGroupNames() []string {
	return []string{}
}

// CanAccessReadAPI returns true for all authenticated API keys
func (a *ApiKeyAuthMethod) CanAccessReadAPI() bool {
	return a.authenticated
}

// CanAccessTerraformAPI returns true for all authenticated API keys
func (a *ApiKeyAuthMethod) CanAccessTerraformAPI() bool {
	return a.authenticated
}

// GetTerraformAuthToken returns empty string for API key authentication
func (a *ApiKeyAuthMethod) GetTerraformAuthToken() string {
	return ""
}

// GetProviderType returns the authentication provider type
// Use the specific auth method type for the current key
func (a *ApiKeyAuthMethod) GetProviderType() auth.AuthMethodType {
	if !a.authenticated || a.currentConfig == nil {
		return auth.AuthMethodNotAuthenticated
	}
	return a.currentConfig.AuthMethod
}

// GetProviderData returns provider-specific data
func (a *ApiKeyAuthMethod) GetProviderData() map[string]interface{} {
	data := make(map[string]interface{})
	if a.authenticated && a.currentConfig != nil {
		data["permission_level"] = a.currentConfig.Permission
		data["auth_method_type"] = string(a.currentConfig.AuthMethod)
	}
	return data
}

// GetPermissionLevel returns the permission level for a given API key
func (a *ApiKeyAuthMethod) GetPermissionLevel(apiKey string) (ApiKeyPermissionLevel, bool) {
	if apiKey == "" {
		return -1, false
	}

	for _, config := range a.apiKeyConfigs {
		for _, validKey := range config.Keys {
			if validKey != "" && apiKey == validKey {
				return config.Permission, true
			}
		}
	}

	return -1, false
}