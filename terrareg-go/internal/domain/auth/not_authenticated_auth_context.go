package auth

import (
	"context"
)

// NotAuthenticatedAuthContext implements AuthContext for unauthenticated access
// It represents the default state when no authentication is provided or valid
type NotAuthenticatedAuthContext struct {
	BaseAuthContext
}

// NewNotAuthenticatedAuthContext creates a new not authenticated auth context
func NewNotAuthenticatedAuthContext(ctx context.Context) *NotAuthenticatedAuthContext {
	return &NotAuthenticatedAuthContext{
		BaseAuthContext: BaseAuthContext{ctx: ctx},
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

// CanPublishModuleVersion returns false for unauthenticated users
func (n *NotAuthenticatedAuthContext) CanPublishModuleVersion(namespace string) bool {
	return false
}

// CanUploadModuleVersion returns false for unauthenticated users
func (n *NotAuthenticatedAuthContext) CanUploadModuleVersion(namespace string) bool {
	return false
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

// CanAccessReadAPI returns false for unauthenticated users
func (n *NotAuthenticatedAuthContext) CanAccessReadAPI() bool {
	return false
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
