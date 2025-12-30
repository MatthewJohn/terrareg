package auth

import (
	"context"
)

// AdminSessionAuthContext implements AuthContext for admin users via session authentication
// It holds the authentication state and permission logic for admin sessions
type AdminSessionAuthContext struct {
	BaseAuthContext
	userID      int
	username    string
	email       string
	sessionID   string
	userGroups  []*UserGroup
	permissions map[string]string
	isAdmin     bool
}

// NewAdminSessionAuthContext creates a new admin session auth context
func NewAdminSessionAuthContext(ctx context.Context, userID int, username, email, sessionID string) *AdminSessionAuthContext {
	return &AdminSessionAuthContext{
		BaseAuthContext: BaseAuthContext{ctx: ctx},
		userID:          userID,
		username:        username,
		email:           email,
		sessionID:       sessionID,
		userGroups:      make([]*UserGroup, 0),
		permissions:     make(map[string]string),
		isAdmin:         false,
	}
}

// AddUserGroup adds a user group to the admin session context
func (a *AdminSessionAuthContext) AddUserGroup(group *UserGroup) {
	a.userGroups = append(a.userGroups, group)

	// Update admin status if any group has admin rights
	if group.SiteAdmin {
		a.isAdmin = true
	}
}

// SetPermission sets a namespace permission
func (a *AdminSessionAuthContext) SetPermission(namespace, permission string) {
	if a.permissions == nil {
		a.permissions = make(map[string]string)
	}
	a.permissions[namespace] = permission
}

// SetAdmin sets admin status
func (a *AdminSessionAuthContext) SetAdmin(isAdmin bool) {
	a.isAdmin = isAdmin
}

// GetProviderType returns the authentication method type
func (a *AdminSessionAuthContext) GetProviderType() AuthMethodType {
	return AuthMethodAdminSession
}

// GetUsername returns the username from the admin session
func (a *AdminSessionAuthContext) GetUsername() string {
	return a.username
}

// IsAuthenticated returns true if the admin session is valid
func (a *AdminSessionAuthContext) IsAuthenticated() bool {
	return a.sessionID != "" && a.username != ""
}

// IsAdmin returns true if the user is an admin
func (a *AdminSessionAuthContext) IsAdmin() bool {
	return a.isAdmin
}

// IsBuiltInAdmin returns false for session-based users
func (a *AdminSessionAuthContext) IsBuiltInAdmin() bool {
	return false
}

// RequiresCSRF returns true for session-based authentication
func (a *AdminSessionAuthContext) RequiresCSRF() bool {
	return true
}

// IsEnabled returns true if the admin session is valid
func (a *AdminSessionAuthContext) IsEnabled() bool {
	return a.IsAuthenticated()
}

// CheckAuthState returns true if the admin session is in a valid state
func (a *AdminSessionAuthContext) CheckAuthState() bool {
	return a.IsAuthenticated()
}

// CanPublishModuleVersion checks if the admin user can publish to a namespace
func (a *AdminSessionAuthContext) CanPublishModuleVersion(namespace string) bool {
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

// CanUploadModuleVersion checks if the admin user can upload to a namespace
func (a *AdminSessionAuthContext) CanUploadModuleVersion(namespace string) bool {
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
func (a *AdminSessionAuthContext) CheckNamespaceAccess(permissionType, namespace string) bool {
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
func (a *AdminSessionAuthContext) GetAllNamespacePermissions() map[string]string {
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
func (a *AdminSessionAuthContext) GetUserGroupNames() []string {
	names := make([]string, len(a.userGroups))
	for i, group := range a.userGroups {
		names[i] = group.Name
	}
	return names
}

// CanAccessReadAPI returns true if the admin user can access the read API
func (a *AdminSessionAuthContext) CanAccessReadAPI() bool {
	return a.IsAuthenticated()
}

// CanAccessTerraformAPI returns true if the admin user can access Terraform API
func (a *AdminSessionAuthContext) CanAccessTerraformAPI() bool {
	return a.IsAuthenticated()
}

// GetTerraformAuthToken returns empty string for admin sessions
func (a *AdminSessionAuthContext) GetTerraformAuthToken() string {
	return ""
}

// GetProviderData returns provider-specific data for the admin session
func (a *AdminSessionAuthContext) GetProviderData() map[string]interface{} {
	return map[string]interface{}{
		"user_id":    a.userID,
		"username":   a.username,
		"email":      a.email,
		"session_id": a.sessionID,
		"is_admin":   a.isAdmin,
		"auth_method": string(AuthMethodAdminSession),
	}
}

// hasPermissionHierarchy checks if the stored permission meets or exceeds the required permission
func (a *AdminSessionAuthContext) hasPermissionHierarchy(stored, required string) bool {
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