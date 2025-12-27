package auth

import (
	"context"
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

// ApiKeyAuthContext implements AuthContext for API key authentication
// It holds the authentication state and permission logic for API keys
type ApiKeyAuthContext struct {
	BaseAuthContext
	permissionLevel ApiKeyPermissionLevel
	username        string
	authMethodType  AuthMethodType
}

// NewApiKeyAuthContext creates a new API key auth context
func NewApiKeyAuthContext(ctx context.Context, permissionLevel ApiKeyPermissionLevel, username string, authMethodType AuthMethodType) *ApiKeyAuthContext {
	return &ApiKeyAuthContext{
		BaseAuthContext: BaseAuthContext{ctx: ctx},
		permissionLevel: permissionLevel,
		username:        username,
		authMethodType:  authMethodType,
	}
}

// GetProviderType returns the authentication method type
func (a *ApiKeyAuthContext) GetProviderType() AuthMethodType {
	return a.authMethodType
}

// GetUsername returns the username associated with the API key
func (a *ApiKeyAuthContext) GetUsername() string {
	return a.username
}

// IsAuthenticated returns true if the API key is valid
func (a *ApiKeyAuthContext) IsAuthenticated() bool {
	return a.permissionLevel >= ApiKeyPermissionUpload
}

// IsAdmin returns true if the API key has admin privileges
func (a *ApiKeyAuthContext) IsAdmin() bool {
	return a.permissionLevel == ApiKeyPermissionAdmin
}

// IsBuiltInAdmin returns false for API keys (they're not built-in admin)
func (a *ApiKeyAuthContext) IsBuiltInAdmin() bool {
	return false
}

// RequiresCSRF returns false for API key authentication
func (a *ApiKeyAuthContext) RequiresCSRF() bool {
	return false
}

// IsEnabled returns true if the API key is valid
func (a *ApiKeyAuthContext) IsEnabled() bool {
	return a.IsAuthenticated()
}

// CheckAuthState returns true if the context is in a valid state
func (a *ApiKeyAuthContext) CheckAuthState() bool {
	return a.IsAuthenticated()
}

// CanPublishModuleVersion checks if the API key can publish modules
func (a *ApiKeyAuthContext) CanPublishModuleVersion(namespace string) bool {
	return a.permissionLevel >= ApiKeyPermissionPublish
}

// CanUploadModuleVersion checks if the API key can upload modules
func (a *ApiKeyAuthContext) CanUploadModuleVersion(namespace string) bool {
	return a.permissionLevel >= ApiKeyPermissionUpload
}

// CheckNamespaceAccess checks if the API key has access to a namespace
// Admin API keys have access to all namespaces
func (a *ApiKeyAuthContext) CheckNamespaceAccess(permissionType, namespace string) bool {
	if a.IsAdmin() {
		return true
	}

	// Non-admin API keys have no namespace restrictions
	// They can upload/publish based on their permission level
	switch permissionType {
	case "READ":
		return true // All API keys can read
	case "MODIFY":
		return a.CanUploadModuleVersion(namespace)
	case "FULL":
		return a.CanPublishModuleVersion(namespace)
	default:
		return false
	}
}

// GetAllNamespacePermissions returns all namespace permissions for the API key
// Admin keys get full access to all namespaces, others get based on permission level
func (a *ApiKeyAuthContext) GetAllNamespacePermissions() map[string]string {
	permissions := make(map[string]string)

	if a.IsAdmin() {
		// Admin keys have full access to all namespaces
		// We don't pre-populate all namespaces, but CheckNamespaceAccess will return true
		permissions["*"] = "FULL"
	} else if a.permissionLevel == ApiKeyPermissionPublish {
		// Publish keys can modify any namespace
		permissions["*"] = "MODIFY"
	} else if a.permissionLevel == ApiKeyPermissionUpload {
		// Upload keys can read any namespace
		permissions["*"] = "READ"
	}

	return permissions
}

// GetUserGroupNames returns empty slice for API keys (they don't have groups)
func (a *ApiKeyAuthContext) GetUserGroupNames() []string {
	return []string{}
}

// CanAccessReadAPI returns true if the API key can access the read API
func (a *ApiKeyAuthContext) CanAccessReadAPI() bool {
	return a.IsAuthenticated()
}

// CanAccessTerraformAPI returns true if the API key can access the Terraform API
func (a *ApiKeyAuthContext) CanAccessTerraformAPI() bool {
	return a.IsAuthenticated()
}

// GetTerraformAuthToken returns empty string for API keys
func (a *ApiKeyAuthContext) GetTerraformAuthToken() string {
	return ""
}

// GetProviderData returns provider-specific data for the API key
func (a *ApiKeyAuthContext) GetProviderData() map[string]interface{} {
	return map[string]interface{}{
		"permission_level": int(a.permissionLevel),
		"username":       a.username,
		"auth_method":    string(a.authMethodType),
	}
}