package auth

import (
	"context"

	infraConfig "github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/config"
)

// SessionAuthContext implements AuthContext for session-based authentication
// It holds the authentication state and permission logic for session users
// This handles all session-based auth: built-in admin login, SAML, OIDC, GitHub
type SessionAuthContext struct {
	BaseAuthContext
	userID       int
	username     string
	email        string
	sessionID    string
	userGroups   []*UserGroup
	permissions  map[string]string
	isAdmin      bool
	config       *infraConfig.InfrastructureConfig
}

// NewSessionAuthContext creates a new session auth context
func NewSessionAuthContext(ctx context.Context, userID int, username, email, sessionID string, config *infraConfig.InfrastructureConfig) *SessionAuthContext {
	return &SessionAuthContext{
		BaseAuthContext: BaseAuthContext{ctx: ctx},
		userID:          userID,
		username:        username,
		email:           email,
		sessionID:       sessionID,
		userGroups:      make([]*UserGroup, 0),
		permissions:     make(map[string]string),
		isAdmin:         false,
		config:          config,
	}
}

// AddUserGroup adds a user group to the admin session context
func (a *SessionAuthContext) AddUserGroup(group *UserGroup) {
	a.userGroups = append(a.userGroups, group)

	// Update admin status if any group has admin rights
	if group.SiteAdmin {
		a.isAdmin = true
	}
}

// SetPermission sets a namespace permission
func (a *SessionAuthContext) SetPermission(namespace, permission string) {
	if a.permissions == nil {
		a.permissions = make(map[string]string)
	}
	a.permissions[namespace] = permission
}

// SetAdmin sets admin status
func (a *SessionAuthContext) SetAdmin(isAdmin bool) {
	a.isAdmin = isAdmin
}

// GetProviderType returns the authentication method type
func (a *SessionAuthContext) GetProviderType() AuthMethodType {
	return AuthMethodSession
}

// GetUsername returns the username from the admin session
func (a *SessionAuthContext) GetUsername() string {
	// @TODO Should always be admin
	return a.username
}

// IsAuthenticated returns true if the admin session is valid
func (a *SessionAuthContext) IsAuthenticated() bool {
	// @TODO is always true - if an admin auth context is returned in Authenticate method, then it is authenticated.
	return a.sessionID != "" && a.username != ""
}

// IsAdmin returns true if the user is an admin
// When ENABLE_ACCESS_CONTROLS is disabled, all session-based users are admins
// When ENABLE_ACCESS_CONTROLS is enabled, only users with admin permissions are admins
func (a *SessionAuthContext) IsAdmin() bool {
	// If RBAC is disabled, all session-based users are treated as admins
	// This matches Python's BaseSsoAuthMethod.is_admin() behavior
	if a.config != nil && !a.config.EnableAccessControls {
		return true
	}
	return a.isAdmin
}

// IsBuiltInAdmin returns false for session-based users
func (a *SessionAuthContext) IsBuiltInAdmin() bool {
	return true
}

// RequiresCSRF returns true for session-based authentication
func (a *SessionAuthContext) RequiresCSRF() bool {
	return true
}

// IsEnabled returns true if the admin session is valid
// @TODO Can these be removed? IsEnabled should be for the AuthMethod
func (a *SessionAuthContext) IsEnabled() bool {
	return a.IsAuthenticated()
}

// CheckAuthState returns true if the admin session is in a valid state
func (a *SessionAuthContext) CheckAuthState() bool {
	return a.IsAuthenticated()
}

// CanPublishModuleVersion checks if the user can publish to a namespace
// Matches Python's BaseSsoAuthMethod.can_publish_module_version()
func (a *SessionAuthContext) CanPublishModuleVersion(namespace string) bool {
	// If RBAC is disabled AND PUBLISH API keys are not enabled, allow publishing
	// This matches Python: ((not ENABLE_ACCESS_CONTROLS) and (not PublishApiKeyAuthMethod.is_enabled())) or check_namespace_access(...)
	if a.config != nil && !a.config.EnableAccessControls && len(a.config.PublishApiKeys) == 0 {
		return true
	}

	// Otherwise, check for admin or namespace access
	if a.IsAdmin() {
		return true
	}

	// Check namespace permissions
	if permission, exists := a.permissions[namespace]; exists {
		return permission == "FULL" || permission == "MODIFY" || permission == "PUBLISH"
	}

	// Check group permissions
	for _, group := range a.userGroups {
		if group.SiteAdmin {
			return true
		}
	}

	return false
}

// CanUploadModuleVersion checks if the user can upload to a namespace
// Matches Python's BaseSsoAuthMethod.can_upload_module_version()
func (a *SessionAuthContext) CanUploadModuleVersion(namespace string) bool {
	// If RBAC is disabled AND UPLOAD API keys are not enabled, allow uploading
	// This matches Python: ((not ENABLE_ACCESS_CONTROLS) and (not UploadApiKeyAuthMethod.is_enabled())) or check_namespace_access(...)
	if a.config != nil && !a.config.EnableAccessControls && len(a.config.UploadApiKeys) == 0 {
		return true
	}

	// Otherwise, check for admin or namespace access
	if a.IsAdmin() {
		return true
	}

	// Check namespace permissions
	if permission, exists := a.permissions[namespace]; exists {
		return permission == "FULL" || permission == "MODIFY" || permission == "UPLOAD"
	}

	// Check group permissions
	for _, group := range a.userGroups {
		if group.SiteAdmin {
			return true
		}
	}

	return false
}

// CheckNamespaceAccess checks if the admin user has access to a namespace
func (a *SessionAuthContext) CheckNamespaceAccess(permissionType, namespace string) bool {
	// @TODO This function should always just return true for admin
	if a.IsAdmin() {
		return true
	}

	// Check namespace permissions
	storedPermission, exists := a.permissions[namespace]
	if exists {
		return a.hasPermissionHierarchy(storedPermission, permissionType)
	}

	// Check group permissions
	for _, group := range a.userGroups {
		if group.SiteAdmin {
			return true
		}
	}

	return false
}

// GetAllNamespacePermissions returns all namespace permissions for the admin user
func (a *SessionAuthContext) GetAllNamespacePermissions() map[string]string {
	// @TODO Return empty map as there should be no depdency on this, because all permissions
	// return true for admin
	result := make(map[string]string)

	// Add direct namespace permissions
	for k, v := range a.permissions {
		result[k] = v
	}

	// Add group permissions
	for _, group := range a.userGroups {
		if group.SiteAdmin {
			// Site admins get full access to all namespaces
			result["*"] = "FULL"
			break
		}
	}

	return result
}

// GetUserGroupNames returns the names of all user groups
func (a *SessionAuthContext) GetUserGroupNames() []string {
	// @TODO return empty, as not needed
	names := make([]string, len(a.userGroups))
	for i, group := range a.userGroups {
		names[i] = group.Name
	}
	return names
}

// CanAccessReadAPI returns true if the admin user can access the read API
func (a *SessionAuthContext) CanAccessReadAPI() bool {
	// Retrurn true
	return a.IsAuthenticated()
}

// CanAccessTerraformAPI returns true if the admin user can access Terraform API
func (a *SessionAuthContext) CanAccessTerraformAPI() bool {
	// Retrurn true
	return a.IsAuthenticated()
}

// GetTerraformAuthToken returns empty string for admin sessions
func (a *SessionAuthContext) GetTerraformAuthToken() string {
	return ""
}

// GetProviderData returns provider-specific data for the admin session
func (a *SessionAuthContext) GetProviderData() map[string]interface{} {
	return map[string]interface{}{
		"user_id":     a.userID,
		"username":    a.username,
		"email":       a.email,
		"session_id":  a.sessionID,
		"is_admin":    a.isAdmin,
		"auth_method": string(AuthMethodSession),
	}
}

// hasPermissionHierarchy checks if the stored permission meets or exceeds the required permission
func (a *SessionAuthContext) hasPermissionHierarchy(stored, required string) bool {
	// Retrurn Is this needed?
	switch required {
	case "READ":
		return stored == "READ" || stored == "MODIFY" || stored == "FULL" || stored == "UPLOAD" || stored == "PUBLISH"
	case "MODIFY":
		return stored == "MODIFY" || stored == "FULL" || stored == "UPLOAD" || stored == "PUBLISH"
	case "UPLOAD":
		return stored == "UPLOAD" || stored == "FULL" || stored == "PUBLISH"
	case "PUBLISH":
		return stored == "PUBLISH" || stored == "FULL"
	case "FULL":
		return stored == "FULL"
	default:
		return false
	}
}
