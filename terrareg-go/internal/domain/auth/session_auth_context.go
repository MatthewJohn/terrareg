package auth

import (
	"context"
)

// SessionAuthContext implements AuthContext for session-based authentication
// It holds the authentication state and permission logic for authenticated users
type SessionAuthContext struct {
	BaseAuthContext
	userID      int
	username    string
	email       string
	sessionID   string
	userGroups  []*UserGroup
	permissions map[string]string
	isAdmin     bool
}

// NewSessionAuthContext creates a new session auth context
func NewSessionAuthContext(ctx context.Context, userID int, username, email, sessionID string) *SessionAuthContext {
	return &SessionAuthContext{
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

// AddUserGroup adds a user group to the session context
func (s *SessionAuthContext) AddUserGroup(group *UserGroup) {
	s.userGroups = append(s.userGroups, group)

	// Update admin status if any group has admin rights
	if group.SiteAdmin {
		s.isAdmin = true
	}
}

// SetPermission sets a namespace permission
func (s *SessionAuthContext) SetPermission(namespace, permission string) {
	if s.permissions == nil {
		s.permissions = make(map[string]string)
	}
	s.permissions[namespace] = permission
}

// GetProviderType returns the authentication method type
func (s *SessionAuthContext) GetProviderType() AuthMethodType {
	return AuthMethodAdminSession
}

// GetUsername returns the username from the session
func (s *SessionAuthContext) GetUsername() string {
	return s.username
}

// IsAuthenticated returns true if the session is valid
func (s *SessionAuthContext) IsAuthenticated() bool {
	return s.sessionID != "" && s.username != ""
}

// IsAdmin returns true if the user is an admin
func (s *SessionAuthContext) IsAdmin() bool {
	return s.isAdmin
}

// IsBuiltInAdmin returns false for session-based users
func (s *SessionAuthContext) IsBuiltInAdmin() bool {
	return false
}

// RequiresCSRF returns true for session-based authentication
func (s *SessionAuthContext) RequiresCSRF() bool {
	return true
}

// IsEnabled returns true if the session is valid
func (s *SessionAuthContext) IsEnabled() bool {
	return s.IsAuthenticated()
}

// CheckAuthState returns true if the session is in a valid state
func (s *SessionAuthContext) CheckAuthState() bool {
	return s.IsAuthenticated()
}

// CanPublishModuleVersion checks if the user can publish to a namespace
func (s *SessionAuthContext) CanPublishModuleVersion(namespace string) bool {
	if s.IsAdmin() {
		return true
	}

	// Check namespace permissions
	if permission, exists := s.permissions[namespace]; exists {
		return permission == "FULL" || permission == "MODIFY" || permission == "PUBLISH"
	}

	// Check if any user group is site admin
	for _, group := range s.userGroups {
		if group.SiteAdmin {
			return true
		}
	}

	return false
}

// CanUploadModuleVersion checks if the user can upload to a namespace
func (s *SessionAuthContext) CanUploadModuleVersion(namespace string) bool {
	if s.IsAdmin() {
		return true
	}

	// Check namespace permissions
	if permission, exists := s.permissions[namespace]; exists {
		return permission == "FULL" || permission == "MODIFY" || permission == "UPLOAD"
	}

	// Check if any user group is site admin
	for _, group := range s.userGroups {
		if group.SiteAdmin {
			return true
		}
	}

	return false
}

// CheckNamespaceAccess checks if the user has access to a namespace
func (s *SessionAuthContext) CheckNamespaceAccess(permissionType, namespace string) bool {
	if s.IsAdmin() {
		return true
	}

	// Check namespace permissions
	storedPermission, exists := s.permissions[namespace]
	if exists {
		return s.hasPermissionHierarchy(storedPermission, permissionType)
	}

	// Check if any user group is site admin
	for _, group := range s.userGroups {
		if group.SiteAdmin {
			return true
		}
	}

	return false
}

// GetAllNamespacePermissions returns all namespace permissions for the user
func (s *SessionAuthContext) GetAllNamespacePermissions() map[string]string {
	result := make(map[string]string)

	// Add direct namespace permissions
	for k, v := range s.permissions {
		result[k] = v
	}

	// Add basic permissions for user groups
	for _, group := range s.userGroups {
		if group.SiteAdmin {
			// Site admins get full access to all namespaces
			result["*"] = "FULL"
			break
		}
	}

	return result
}

// GetUserGroupNames returns the names of all user groups
func (s *SessionAuthContext) GetUserGroupNames() []string {
	names := make([]string, len(s.userGroups))
	for i, group := range s.userGroups {
		names[i] = group.Name
	}
	return names
}

// CanAccessReadAPI returns true if the user can access the read API
func (s *SessionAuthContext) CanAccessReadAPI() bool {
	return s.IsAuthenticated()
}

// CanAccessTerraformAPI returns true if the user can access the Terraform API
func (s *SessionAuthContext) CanAccessTerraformAPI() bool {
	return s.IsAuthenticated()
}

// GetTerraformAuthToken returns empty string for session-based auth
func (s *SessionAuthContext) GetTerraformAuthToken() string {
	return ""
}

// GetProviderData returns provider-specific data for the session
func (s *SessionAuthContext) GetProviderData() map[string]interface{} {
	return map[string]interface{}{
		"user_id":    s.userID,
		"username":   s.username,
		"email":      s.email,
		"session_id": s.sessionID,
		"is_admin":   s.isAdmin,
		"auth_method": string(AuthMethodAdminSession),
	}
}

// hasPermissionHierarchy checks if the stored permission meets or exceeds the required permission
func (s *SessionAuthContext) hasPermissionHierarchy(stored, required string) bool {
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

// comparePermissionLevel compares two permission levels, returns 1 if a > b, -1 if a < b, 0 if equal
func (s *SessionAuthContext) comparePermissionLevel(a, b string) int {
	levelA := s.getPermissionLevel(a)
	levelB := s.getPermissionLevel(b)

	if levelA > levelB {
		return 1
	} else if levelA < levelB {
		return -1
	}
	return 0
}

// getPermissionLevel returns a numeric value for permission comparison
func (s *SessionAuthContext) getPermissionLevel(permission string) int {
	switch permission {
	case "READ":
		return 1
	case "MODIFY":
		return 2
	case "UPLOAD":
		return 3
	case "PUBLISH":
		return 4
	case "FULL":
		return 5
	default:
		return 0
	}
}