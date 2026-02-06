package auth

import (
	"context"

	infraConfig "github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/config"
)

// NotAuthenticatedAuthContext implements AuthContext for unauthenticated access
// It represents the default state when no authentication is provided or valid
type NotAuthenticatedAuthContext struct {
	BaseAuthContext
	enableAccessControls       bool
	allowUnauthenticatedAccess bool
	publishApiKeysEnabled       bool
	uploadApiKeysEnabled        bool
}

// NewNotAuthenticatedAuthContext creates a new not authenticated auth context
func NewNotAuthenticatedAuthContext(ctx context.Context, config *infraConfig.InfrastructureConfig) *NotAuthenticatedAuthContext {
	if config == nil {
		// Default values when config is not provided
		// AllowUnauthenticatedAccess defaults to true for safety (allows read API access)
		return &NotAuthenticatedAuthContext{
			BaseAuthContext:            BaseAuthContext{ctx: ctx},
			enableAccessControls:       false, // Default RBAC disabled
			allowUnauthenticatedAccess: true, // Default to allow unauthenticated read access
			publishApiKeysEnabled:       false, // No API keys configured
			uploadApiKeysEnabled:        false, // No API keys configured
		}
	}
	return &NotAuthenticatedAuthContext{
		BaseAuthContext:            BaseAuthContext{ctx: ctx},
		enableAccessControls:       config.EnableAccessControls,
		allowUnauthenticatedAccess: config.AllowUnauthenticatedAccess,
		publishApiKeysEnabled:       len(config.PublishApiKeys) > 0,
		uploadApiKeysEnabled:        len(config.UploadApiKeys) > 0,
	}
}

// GetProviderType returns the authentication method type
func (n *NotAuthenticatedAuthContext) GetProviderType() AuthMethodType {
	return AuthMethodNotAuthenticated
}

// GetUsername returns empty string for unauthenticated users
func (n *NotAuthenticatedAuthContext) GetUsername() string {
	return ""
}

// IsAuthenticated returns false for unauthenticated users
func (n *NotAuthenticatedAuthContext) IsAuthenticated() bool {
	return false
}

// IsAdmin returns false for unauthenticated users
func (n *NotAuthenticatedAuthContext) IsAdmin() bool {
	return false
}

// IsBuiltInAdmin returns false for unauthenticated users
func (n *NotAuthenticatedAuthContext) IsBuiltInAdmin() bool {
	return false
}

// RequiresCSRF returns false for unauthenticated access
func (n *NotAuthenticatedAuthContext) RequiresCSRF() bool {
	return false
}

// IsEnabled returns true (not authenticated is always enabled)
func (n *NotAuthenticatedAuthContext) IsEnabled() bool {
	return true
}

// CheckAuthState returns true (unauthenticated is a valid state)
func (n *NotAuthenticatedAuthContext) CheckAuthState() bool {
	return true
}

// CanPublishModuleVersion checks if unauthenticated user can publish
// Matches Python's NotAuthenticated.can_publish_module_version()
func (n *NotAuthenticatedAuthContext) CanPublishModuleVersion(namespace string) bool {
	// If API key authentication is not configured for publishing modules,
	// RBAC is not enabled and unauthenticated access is enabled,
	// allow unauthenticated access
	// This matches Python: ((not ENABLE_ACCESS_CONTROLS) and (not PublishApiKeyAuthMethod.is_enabled()) and ALLOW_UNAUTHENTICATED_ACCESS)
	return !n.enableAccessControls && !n.publishApiKeysEnabled && n.allowUnauthenticatedAccess
}

// CanUploadModuleVersion checks if unauthenticated user can upload
// Matches Python's NotAuthenticated.can_upload_module_version()
func (n *NotAuthenticatedAuthContext) CanUploadModuleVersion(namespace string) bool {
	// If API key authentication is not configured for uploading modules,
	// RBAC is not enabled and unauthenticated access is enabled,
	// allow unauthenticated access
	// This matches Python: ((not ENABLE_ACCESS_CONTROLS) and (not UploadApiKeyAuthMethod.is_enabled()) and ALLOW_UNAUTHENTICATED_ACCESS)
	return !n.enableAccessControls && !n.uploadApiKeysEnabled && n.allowUnauthenticatedAccess
}

// CheckNamespaceAccess returns false for unauthenticated users
func (n *NotAuthenticatedAuthContext) CheckNamespaceAccess(permissionType, namespace string) bool {
	return false
}

// GetAllNamespacePermissions returns empty permissions for unauthenticated users
func (n *NotAuthenticatedAuthContext) GetAllNamespacePermissions() map[string]string {
	return make(map[string]string)
}

// GetUserGroupNames returns empty slice for unauthenticated users
func (n *NotAuthenticatedAuthContext) GetUserGroupNames() []string {
	return []string{}
}

// CanAccessReadAPI returns true if unauthenticated access is allowed
// Matches Python's NotAuthenticated.can_access_read_api()
func (n *NotAuthenticatedAuthContext) CanAccessReadAPI() bool {
	return n.allowUnauthenticatedAccess
}

// CanAccessTerraformAPI returns false for unauthenticated users
func (n *NotAuthenticatedAuthContext) CanAccessTerraformAPI() bool {
	return false
}

// GetTerraformAuthToken returns empty string for unauthenticated users
func (n *NotAuthenticatedAuthContext) GetTerraformAuthToken() string {
	return ""
}

// GetProviderData returns empty provider data for unauthenticated users
func (n *NotAuthenticatedAuthContext) GetProviderData() map[string]interface{} {
	return map[string]interface{}{
		"auth_method": string(AuthMethodNotAuthenticated),
	}
}
