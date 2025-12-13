package auth

import (
	"context"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/auth"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/auth/model"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/config"
)

// PublishApiKeyAuthMethod implements publish API key authentication
// Matches Python's PublishApiKeyAuthMethod which uses X-Terrareg-ApiKey header
type PublishApiKeyAuthMethod struct {
	config *config.InfrastructureConfig
}

// NewPublishApiKeyAuthMethod creates a new publish API key auth method
func NewPublishApiKeyAuthMethod(config *config.InfrastructureConfig) *PublishApiKeyAuthMethod {
	return &PublishApiKeyAuthMethod{
		config: config,
	}
}

// GetProviderType returns the authentication provider type
func (a *PublishApiKeyAuthMethod) GetProviderType() auth.AuthMethodType {
	return auth.AuthMethodPublishApiKey
}

// IsEnabled checks if publish authentication is enabled
func (a *PublishApiKeyAuthMethod) IsEnabled() bool {
	return len(a.config.PublishApiKeys) > 0
}

// Authenticate validates the publish API key from X-Terrareg-ApiKey header
func (a *PublishApiKeyAuthMethod) Authenticate(ctx context.Context, apiKey string) (*auth.AuthContext, error) {
	if apiKey == "" {
		return nil, model.ErrInvalidCredentials
	}

	// Check against publish tokens from config
	for _, validKey := range a.config.PublishApiKeys {
		if validKey != "" && apiKey == validKey {
			authCtx := auth.NewAuthContext(a)
			authCtx.IsAuthenticated = true
			authCtx.Username = "Publish API Key"
			return authCtx, nil
		}
	}

	return nil, model.ErrInvalidCredentials
}

// CheckAuthState checks if publish API key is provided
func (a *PublishApiKeyAuthMethod) CheckAuthState() bool {
	return len(a.config.PublishApiKeys) > 0
}

// CheckAuthStateWithKey checks if publish API key is provided (helper method)
func (a *PublishApiKeyAuthMethod) CheckAuthStateWithKey(apiKey string) bool {
	if apiKey == "" || len(a.config.PublishApiKeys) == 0 {
		return false
	}
	for _, validKey := range a.config.PublishApiKeys {
		if validKey != "" && apiKey == validKey {
			return true
		}
	}
	return false
}

// IsAuthenticated returns whether the current request is authenticated
func (a *PublishApiKeyAuthMethod) IsAuthenticated() bool {
	return true // Always returns true since this method is only used when authenticated
}

// IsAdmin returns whether the authenticated user has admin privileges
func (a *PublishApiKeyAuthMethod) IsAdmin() bool {
	return false // Publish API keys are not admin
}

// IsBuiltInAdmin returns whether this is a built-in admin method
func (a *PublishApiKeyAuthMethod) IsBuiltInAdmin() bool {
	return false // Publish API key is not admin
}

// RequiresCSRF returns whether this authentication method requires CSRF protection
func (a *PublishApiKeyAuthMethod) RequiresCSRF() bool {
	return false // API key auth doesn't need CSRF (matches Python)
}

// CanAccessReadAPI returns whether the user can access read APIs
// API keys cannot access read APIs (matches Python)
func (a *PublishApiKeyAuthMethod) CanAccessReadAPI() bool {
	return false
}

// CanAccessTerraformAPI returns whether the user can access Terraform APIs
func (a *PublishApiKeyAuthMethod) CanAccessTerraformAPI() bool {
	return true // Publish API keys can access Terraform APIs
}

// CanPublishModuleVersion checks if the user can publish module versions to the given namespace
func (a *PublishApiKeyAuthMethod) CanPublishModuleVersion(namespace string) bool {
	// Publish API keys can publish to any namespace (matches Python)
	return true
}

// CanUploadModuleVersion checks if the user can upload module versions to the given namespace
func (a *PublishApiKeyAuthMethod) CanUploadModuleVersion(namespace string) bool {
	// Publish API keys can only publish, not upload (matches Python)
	return false
}

// CheckNamespaceAccess checks if the user has the specified permission for a namespace
func (a *PublishApiKeyAuthMethod) CheckNamespaceAccess(permissionType, namespace string) bool {
	// Publish API keys don't have namespace permissions (matches Python)
	return false
}

// GetAllNamespacePermissions returns all namespace permissions for the user
func (a *PublishApiKeyAuthMethod) GetAllNamespacePermissions() map[string]string {
	// Publish API keys don't have namespace permissions (matches Python)
	return map[string]string{}
}

// GetUsername returns the authenticated username
func (a *PublishApiKeyAuthMethod) GetUsername() string {
	return "Publish API Key" // Matches Python's get_username()
}

// GetUserGroupNames returns the names of all user groups
func (a *PublishApiKeyAuthMethod) GetUserGroupNames() []string {
	// API key auth doesn't use traditional user groups
	return []string{}
}

// GetTerraformAuthToken returns the Terraform authentication token
func (a *PublishApiKeyAuthMethod) GetTerraformAuthToken() string {
	// Return first publish key (for simplicity)
	if len(a.config.PublishApiKeys) > 0 {
		return a.config.PublishApiKeys[0]
	}
	return ""
}

// GetProviderData returns provider-specific data
func (a *PublishApiKeyAuthMethod) GetProviderData() map[string]interface{} {
	data := make(map[string]interface{})
	data["is_admin"] = false
	data["username"] = "Publish API Key"
	data["can_upload"] = false
	data["can_publish"] = true
	return data
}
