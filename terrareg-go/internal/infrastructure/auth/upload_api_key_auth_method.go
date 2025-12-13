package auth

import (
	"context"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/auth"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/auth/model"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/config"
)

// UploadApiKeyAuthMethod implements upload API key authentication
// Matches Python's UploadApiKeyAuthMethod which uses X-Terrareg-ApiKey header
type UploadApiKeyAuthMethod struct {
	config *config.InfrastructureConfig
}

// NewUploadApiKeyAuthMethod creates a new upload API key auth method
func NewUploadApiKeyAuthMethod(config *config.InfrastructureConfig) *UploadApiKeyAuthMethod {
	return &UploadApiKeyAuthMethod{
		config: config,
	}
}

// GetProviderType returns the authentication provider type
func (a *UploadApiKeyAuthMethod) GetProviderType() auth.AuthMethodType {
	return auth.AuthMethodUploadApiKey
}

// IsEnabled checks if upload authentication is enabled
func (a *UploadApiKeyAuthMethod) IsEnabled() bool {
	return len(a.config.UploadApiKeys) > 0
}

// Authenticate validates the upload API key from X-Terrareg-ApiKey header
func (a *UploadApiKeyAuthMethod) Authenticate(ctx context.Context, apiKey string) (*auth.AuthContext, error) {
	if apiKey == "" {
		return nil, model.ErrInvalidCredentials
	}

	// Check against upload tokens from config
	for _, validKey := range a.config.UploadApiKeys {
		if validKey != "" && apiKey == validKey {
			authCtx := auth.NewAuthContext(a)
			authCtx.IsAuthenticated = true
			authCtx.Username = "Upload API Key"
			return authCtx, nil
		}
	}

	return nil, model.ErrInvalidCredentials
}

// CheckAuthState checks if upload API key is provided
func (a *UploadApiKeyAuthMethod) CheckAuthState() bool {
	return len(a.config.UploadApiKeys) > 0
}

// CheckAuthStateWithKey checks if upload API key is provided (helper method)
func (a *UploadApiKeyAuthMethod) CheckAuthStateWithKey(apiKey string) bool {
	if apiKey == "" || len(a.config.UploadApiKeys) == 0 {
		return false
	}
	for _, validKey := range a.config.UploadApiKeys {
		if validKey != "" && apiKey == validKey {
			return true
		}
	}
	return false
}

// IsAuthenticated returns whether the current request is authenticated
func (a *UploadApiKeyAuthMethod) IsAuthenticated() bool {
	return true // Always returns true since this method is only used when authenticated
}

// IsAdmin returns whether the authenticated user has admin privileges
func (a *UploadApiKeyAuthMethod) IsAdmin() bool {
	return false // Upload API keys are not admin
}

// IsBuiltInAdmin returns whether this is a built-in admin method
func (a *UploadApiKeyAuthMethod) IsBuiltInAdmin() bool {
	return false // Upload API key is not admin
}

// RequiresCSRF returns whether this authentication method requires CSRF protection
func (a *UploadApiKeyAuthMethod) RequiresCSRF() bool {
	return false // API key auth doesn't need CSRF (matches Python)
}

// CanAccessReadAPI returns whether the user can access read APIs
// API keys cannot access read APIs (matches Python)
func (a *UploadApiKeyAuthMethod) CanAccessReadAPI() bool {
	return false
}

// CanAccessTerraformAPI returns whether the user can access Terraform APIs
func (a *UploadApiKeyAuthMethod) CanAccessTerraformAPI() bool {
	return true // Upload API keys can access Terraform APIs
}

// CanPublishModuleVersion checks if the user can publish module versions to the given namespace
func (a *UploadApiKeyAuthMethod) CanPublishModuleVersion(namespace string) bool {
	// Upload API keys can only upload, not publish (matches Python)
	return false
}

// CanUploadModuleVersion checks if the user can upload module versions to the given namespace
func (a *UploadApiKeyAuthMethod) CanUploadModuleVersion(namespace string) bool {
	// Upload API keys can upload to any namespace (matches Python)
	return true
}

// CheckNamespaceAccess checks if the user has the specified permission for a namespace
func (a *UploadApiKeyAuthMethod) CheckNamespaceAccess(permissionType, namespace string) bool {
	// Upload API keys don't have namespace permissions (matches Python)
	return false
}

// GetAllNamespacePermissions returns all namespace permissions for the user
func (a *UploadApiKeyAuthMethod) GetAllNamespacePermissions() map[string]string {
	// Upload API keys don't have namespace permissions (matches Python)
	return map[string]string{}
}

// GetUsername returns the authenticated username
func (a *UploadApiKeyAuthMethod) GetUsername() string {
	return "Upload API Key" // Matches Python's get_username()
}

// GetUserGroupNames returns the names of all user groups
func (a *UploadApiKeyAuthMethod) GetUserGroupNames() []string {
	// API key auth doesn't use traditional user groups
	return []string{}
}

// GetTerraformAuthToken returns the Terraform authentication token
func (a *UploadApiKeyAuthMethod) GetTerraformAuthToken() string {
	// Return first upload key (for simplicity)
	if len(a.config.UploadApiKeys) > 0 {
		return a.config.UploadApiKeys[0]
	}
	return ""
}

// GetProviderData returns provider-specific data
func (a *UploadApiKeyAuthMethod) GetProviderData() map[string]interface{} {
	data := make(map[string]interface{})
	data["is_admin"] = false
	data["username"] = "Upload API Key"
	data["can_upload"] = true
	data["can_publish"] = false
	return data
}
