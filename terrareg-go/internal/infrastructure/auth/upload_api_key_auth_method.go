package auth

import (
	"context"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/auth"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/config"
)

// UploadApiKeyAuthMethod implements immutable authentication for upload API keys
type UploadApiKeyAuthMethod struct {
	config *config.InfrastructureConfig
}

// NewUploadApiKeyAuthMethod creates a new immutable upload API key authentication method
func NewUploadApiKeyAuthMethod(config *config.InfrastructureConfig) *UploadApiKeyAuthMethod {
	return &UploadApiKeyAuthMethod{
		config: config,
	}
}

// GetProviderType returns the authentication provider type
func (u *UploadApiKeyAuthMethod) GetProviderType() auth.AuthMethodType {
	return auth.AuthMethodUploadApiKey
}

// IsEnabled returns whether this authentication method is enabled
func (u *UploadApiKeyAuthMethod) IsEnabled() bool {
	// Upload API key is enabled if UPLOAD_API_KEYS is configured and not empty
	return len(u.config.UploadApiKeys) > 0
}

// Authenticate authenticates an upload API key request and returns an UploadApiKeyAuthContext
func (u *UploadApiKeyAuthMethod) Authenticate(ctx context.Context, headers, formData, queryParams map[string]string) (auth.AuthMethod, error) {
	// Check if upload API keys are configured
	if !u.IsEnabled() {
		return nil, nil // Let other auth methods try
	}

	// Extract API key from X-Terrareg-ApiKey header (NOT Authorization header)
	apiKey, exists := headers["X-Terrareg-ApiKey"]
	if !exists {
		return nil, nil // Let other auth methods try
	}

	// Validate API key against configured upload keys
	validKey := false
	for _, configuredKey := range u.config.UploadApiKeys {
		if apiKey == configuredKey {
			validKey = true
			break
		}
	}

	if !validKey {
		return nil, nil // Let other auth methods try
	}

	// Create UploadApiKeyAuthContext with authentication state
	authContext := auth.NewUploadApiKeyAuthContext(ctx, apiKey)

	return authContext, nil
}

// AuthMethod interface implementation for the base UploadApiKeyAuthMethod
// These return default values since the actual auth state is in the UploadApiKeyAuthContext

func (u *UploadApiKeyAuthMethod) IsBuiltInAdmin() bool                     { return false }
func (u *UploadApiKeyAuthMethod) IsAdmin() bool                            { return false }
func (u *UploadApiKeyAuthMethod) IsAuthenticated() bool                    { return false }
func (u *UploadApiKeyAuthMethod) RequiresCSRF() bool                       { return false }
func (u *UploadApiKeyAuthMethod) CheckAuthState() bool                     { return false }
func (u *UploadApiKeyAuthMethod) CanPublishModuleVersion(string) bool      { return false }
func (u *UploadApiKeyAuthMethod) CanUploadModuleVersion(string) bool       { return false }
func (u *UploadApiKeyAuthMethod) CheckNamespaceAccess(string, string) bool { return false }
func (u *UploadApiKeyAuthMethod) GetAllNamespacePermissions() map[string]string {
	return make(map[string]string)
}
func (u *UploadApiKeyAuthMethod) GetUsername() string           { return "" }
func (u *UploadApiKeyAuthMethod) GetUserGroupNames() []string   { return []string{} }
func (u *UploadApiKeyAuthMethod) CanAccessReadAPI() bool        { return false }
func (u *UploadApiKeyAuthMethod) CanAccessTerraformAPI() bool   { return false }
func (u *UploadApiKeyAuthMethod) GetTerraformAuthToken() string { return "" }
func (u *UploadApiKeyAuthMethod) GetProviderData() map[string]interface{} {
	return make(map[string]interface{})
}
