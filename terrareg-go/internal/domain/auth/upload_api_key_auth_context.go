package auth

import (
	"context"
)

// UploadApiKeyAuthContext implements AuthContext for upload API key authentication
// It holds the authentication state and permission logic for upload API keys
type UploadApiKeyAuthContext struct {
	BaseAuthContext
	apiKey string
}

// NewUploadApiKeyAuthContext creates a new upload API key auth context
func NewUploadApiKeyAuthContext(ctx context.Context, apiKey string) *UploadApiKeyAuthContext {
	return &UploadApiKeyAuthContext{
		BaseAuthContext: BaseAuthContext{ctx: ctx},
		apiKey:          apiKey,
	}
}

// GetProviderType returns the authentication method type
func (u *UploadApiKeyAuthContext) GetProviderType() AuthMethodType {
	return AuthMethodUploadApiKey
}

// GetUsername returns the username for upload API key
func (u *UploadApiKeyAuthContext) GetUsername() string {
	return "upload-api-key"
}

// IsAuthenticated returns true if the API key is valid
func (u *UploadApiKeyAuthContext) IsAuthenticated() bool {
	return u.apiKey != ""
}

// IsAdmin returns false for upload API keys
func (u *UploadApiKeyAuthContext) IsAdmin() bool {
	return false
}

// IsBuiltInAdmin returns false for upload API keys
func (u *UploadApiKeyAuthContext) IsBuiltInAdmin() bool {
	return false
}

// RequiresCSRF returns false for API key authentication
func (u *UploadApiKeyAuthContext) RequiresCSRF() bool {
	return false
}

// CheckAuthState returns true for API key authentication
func (u *UploadApiKeyAuthContext) CheckAuthState() bool {
	return true
}

// CanPublishModuleVersion returns false for upload API keys (cannot publish)
func (u *UploadApiKeyAuthContext) CanPublishModuleVersion(module string) bool {
	return false
}

// CanUploadModuleVersion returns true for upload API keys
func (u *UploadApiKeyAuthContext) CanUploadModuleVersion(module string) bool {
	return true
}

// CheckNamespaceAccess returns true for upload API keys (all namespaces for uploading)
func (u *UploadApiKeyAuthContext) CheckNamespaceAccess(namespace, permission string) bool {
	// Upload keys can access all namespaces for upload operations
	return true
}

// GetAllNamespacePermissions returns upload permissions for all namespaces
func (u *UploadApiKeyAuthContext) GetAllNamespacePermissions() map[string]string {
	return map[string]string{}
}

// GetUserGroupNames returns empty list for upload API keys
func (u *UploadApiKeyAuthContext) GetUserGroupNames() []string {
	return []string{}
}

// CanAccessReadAPI returns false for upload API keys
func (u *UploadApiKeyAuthContext) CanAccessReadAPI() bool {
	return false
}

// CanAccessTerraformAPI returns false for upload API keys
func (u *UploadApiKeyAuthContext) CanAccessTerraformAPI() bool {
	return false
}

// GetTerraformAuthToken returns empty string for upload API keys
func (u *UploadApiKeyAuthContext) GetTerraformAuthToken() string {
	return ""
}

// GetProviderData returns provider-specific data for upload API key
func (u *UploadApiKeyAuthContext) GetProviderData() map[string]interface{} {
	return map[string]interface{}{
		"auth_method": string(AuthMethodUploadApiKey),
		"is_admin":    false,
		"can_upload":  true,
		"can_publish": false,
	}
}

// IsEnabled returns true for upload API key context (always enabled when created)
func (u *UploadApiKeyAuthContext) IsEnabled() bool {
	return true
}
