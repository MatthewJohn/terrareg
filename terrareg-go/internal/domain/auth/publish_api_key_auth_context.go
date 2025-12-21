package auth

import (
	"context"
)

// PublishApiKeyAuthContext implements AuthContext for publish API key authentication
// It holds the authentication state and permission logic for publish API keys
type PublishApiKeyAuthContext struct {
	BaseAuthContext
	apiKey string
}

// NewPublishApiKeyAuthContext creates a new publish API key auth context
func NewPublishApiKeyAuthContext(ctx context.Context, apiKey string) *PublishApiKeyAuthContext {
	return &PublishApiKeyAuthContext{
		BaseAuthContext: BaseAuthContext{ctx: ctx},
		apiKey:         apiKey,
	}
}

// GetProviderType returns the authentication method type
func (p *PublishApiKeyAuthContext) GetProviderType() AuthMethodType {
	return AuthMethodPublishApiKey
}

// GetUsername returns the username for publish API key
func (p *PublishApiKeyAuthContext) GetUsername() string {
	return "publish-api-key"
}

// IsAuthenticated returns true if the API key is valid
func (p *PublishApiKeyAuthContext) IsAuthenticated() bool {
	return p.apiKey != ""
}

// IsAdmin returns false for publish API keys
func (p *PublishApiKeyAuthContext) IsAdmin() bool {
	return false
}

// IsBuiltInAdmin returns false for publish API keys
func (p *PublishApiKeyAuthContext) IsBuiltInAdmin() bool {
	return false
}

// RequiresCSRF returns false for API key authentication
func (p *PublishApiKeyAuthContext) RequiresCSRF() bool {
	return false
}

// CheckAuthState returns true for API key authentication
func (p *PublishApiKeyAuthContext) CheckAuthState() bool {
	return true
}

// CanPublishModuleVersion returns true for publish API keys
func (p *PublishApiKeyAuthContext) CanPublishModuleVersion(module string) bool {
	return true
}

// CanUploadModuleVersion returns true for publish API keys (can also upload)
func (p *PublishApiKeyAuthContext) CanUploadModuleVersion(module string) bool {
	return true
}

// CheckNamespaceAccess returns true for publish API keys (all namespaces for publishing)
func (p *PublishApiKeyAuthContext) CheckNamespaceAccess(namespace, permission string) bool {
	// Publish keys can access all namespaces for publish/upload operations
	return true
}

// GetAllNamespacePermissions returns publish permissions for all namespaces
func (p *PublishApiKeyAuthContext) GetAllNamespacePermissions() map[string]string {
	return map[string]string{}
}

// GetUserGroupNames returns empty list for publish API keys
func (p *PublishApiKeyAuthContext) GetUserGroupNames() []string {
	return []string{}
}

// CanAccessReadAPI returns false for publish API keys
func (p *PublishApiKeyAuthContext) CanAccessReadAPI() bool {
	return false
}

// CanAccessTerraformAPI returns false for publish API keys
func (p *PublishApiKeyAuthContext) CanAccessTerraformAPI() bool {
	return false
}

// GetTerraformAuthToken returns empty string for publish API keys
func (p *PublishApiKeyAuthContext) GetTerraformAuthToken() string {
	return ""
}

// GetProviderData returns provider-specific data for publish API key
func (p *PublishApiKeyAuthContext) GetProviderData() map[string]interface{} {
	return map[string]interface{}{
		"auth_method": string(AuthMethodPublishApiKey),
		"is_admin":    false,
		"can_publish": true,
		"can_upload":  true,
	}
}

// IsEnabled returns true for publish API key context (always enabled when created)
func (p *PublishApiKeyAuthContext) IsEnabled() bool {
	return true
}