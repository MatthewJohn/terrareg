package auth

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/auth"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/auth/model"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/auth/repository"
)

// AdminSessionAuthMethod implements authentication for admin users via session cookies
type AdminSessionAuthMethod struct {
	auth.BaseAuthMethod
	sessionRepo     repository.SessionRepository
	userGroupRepo   repository.UserGroupRepository
	currentSession  *auth.Session
	isAuthenticated bool
	isAdmin         bool
	username        string
	userPermissions map[string]string
	userGroups      []*auth.UserGroup
}

// NewAdminSessionAuthMethod creates a new admin session authentication method
func NewAdminSessionAuthMethod(
	sessionRepo repository.SessionRepository,
	userGroupRepo repository.UserGroupRepository,
) *AdminSessionAuthMethod {
	return &AdminSessionAuthMethod{
		sessionRepo:     sessionRepo,
		userGroupRepo:   userGroupRepo,
		userPermissions: make(map[string]string),
		userGroups:      make([]*auth.UserGroup, 0),
	}
}

// GetProviderType returns the authentication provider type
func (a *AdminSessionAuthMethod) GetProviderType() auth.AuthMethodType {
	return auth.AuthMethodAdminSession
}

// CheckAuthState validates the current authentication state
func (a *AdminSessionAuthMethod) CheckAuthState() bool {
	return a.isAuthenticated
}

// IsBuiltInAdmin returns whether this is a built-in admin method
func (a *AdminSessionAuthMethod) IsBuiltInAdmin() bool {
	return a.isAdmin
}

// IsAuthenticated returns whether the current request is authenticated
func (a *AdminSessionAuthMethod) IsAuthenticated() bool {
	return a.isAuthenticated
}

// IsAdmin returns whether the authenticated user has admin privileges
func (a *AdminSessionAuthMethod) IsAdmin() bool {
	return true
}

// IsEnabled returns whether this authentication method is enabled
func (a *AdminSessionAuthMethod) IsEnabled() bool {
	return true
}

// RequiresCSRF returns whether this authentication method requires CSRF protection
func (a *AdminSessionAuthMethod) RequiresCSRF() bool {
	return true
}

// CanPublishModuleVersion checks if the user can publish module versions to the given namespace
func (a *AdminSessionAuthMethod) CanPublishModuleVersion(namespace string) bool {
	if a.isAdmin {
		return true
	}
	return a.CheckNamespaceAccess("FULL", namespace)
}

// CanUploadModuleVersion checks if the user can upload module versions to the given namespace
func (a *AdminSessionAuthMethod) CanUploadModuleVersion(namespace string) bool {
	if a.isAdmin {
		return true
	}
	return a.CheckNamespaceAccess("FULL", namespace) || a.CheckNamespaceAccess("MODIFY", namespace)
}

// CheckNamespaceAccess checks if the user has the specified permission for a namespace
func (a *AdminSessionAuthMethod) CheckNamespaceAccess(permissionType, namespace string) bool {
	if a.isAdmin {
		return true
	}

	storedPermission, exists := a.userPermissions[namespace]
	if !exists {
		return false
	}

	// Check permission hierarchy
	switch auth.PermissionType(permissionType) {
	case auth.PermissionRead:
		return storedPermission == string(auth.PermissionRead) ||
			storedPermission == string(auth.PermissionModify) ||
			storedPermission == string(auth.PermissionFull)
	case auth.PermissionModify:
		return storedPermission == string(auth.PermissionModify) ||
			storedPermission == string(auth.PermissionFull)
	case auth.PermissionFull:
		return storedPermission == string(auth.PermissionFull)
	default:
		return false
	}
}

// GetAllNamespacePermissions returns all namespace permissions for the user
func (a *AdminSessionAuthMethod) GetAllNamespacePermissions() map[string]string {
	return a.userPermissions
}

// GetUsername returns the authenticated username
func (a *AdminSessionAuthMethod) GetUsername() string {
	return a.username
}

// GetUserGroupNames returns the names of all user groups
func (a *AdminSessionAuthMethod) GetUserGroupNames() []string {
	names := make([]string, len(a.userGroups))
	for i, group := range a.userGroups {
		names[i] = group.GetName()
	}
	return names
}

// CanAccessReadAPI returns whether the user can access read APIs
func (a *AdminSessionAuthMethod) CanAccessReadAPI() bool {
	return a.isAuthenticated
}

// CanAccessTerraformAPI returns whether the user can access Terraform APIs
func (a *AdminSessionAuthMethod) CanAccessTerraformAPI() bool {
	return a.isAdmin
}

// GetTerraformAuthToken returns the Terraform authentication token
func (a *AdminSessionAuthMethod) GetTerraformAuthToken() string {
	// Session-based auth doesn't typically provide Terraform tokens
	return ""
}

// GetProviderData returns provider-specific data
func (a *AdminSessionAuthMethod) GetProviderData() map[string]interface{} {
	data := make(map[string]interface{})
	data["session_id"] = a.currentSession.ID
	data["expires_at"] = a.currentSession.Expiry
	data["user_groups"] = a.GetUserGroupNames()
	return data
}

// Authenticate authenticates a request and updates the auth method state
func (a *AdminSessionAuthMethod) Authenticate(ctx context.Context, headers map[string]string, cookies map[string]string) error {
	// Look for session cookie
	sessionID, exists := cookies["session_id"]
	if !exists {
		return model.ErrAuthenticationFailed
	}

	if strings.TrimSpace(sessionID) == "" {
		return model.ErrAuthenticationFailed
	}

	// Find session in database
	session, err := a.sessionRepo.FindByID(ctx, sessionID)
	if err != nil {
		return model.ErrAuthenticationFailed
	}

	// Check if session is expired
	if session.IsExpired() {
		return model.ErrSessionExpired
	}

	// Store current session
	a.currentSession = session
	a.isAuthenticated = true

	// Parse provider source auth to get user information
	userInfo, err := a.parseProviderSourceAuth(session.ProviderSourceAuth)
	if err != nil {
		return model.ErrAuthenticationFailed
	}

	a.username = userInfo.Username

	// Get user groups and check admin status
	userGroups, err := a.userGroupRepo.GetGroupsForUser(ctx, userInfo.UserID)
	if err == nil {
		a.userGroups = userGroups
	}

	// Check if user is admin
	isAdmin := false
	for _, group := range a.userGroups {
		if group.IsSiteAdmin() {
			isAdmin = true
			break
		}
	}
	a.isAdmin = isAdmin

	// Get user permissions
	permissions, err := a.getUserPermissions(ctx, userInfo.UserID)
	if err == nil {
		a.userPermissions = permissions
	}

	return nil
}

// parseProviderSourceAuth parses the provider_source_auth JSON to extract user information
func (a *AdminSessionAuthMethod) parseProviderSourceAuth(providerSourceAuth []byte) (*UserInfo, error) {
	if len(providerSourceAuth) == 0 {
		// Default admin user info if no auth data
		return &UserInfo{
			UserID:   "admin",
			Username: "Admin User",
		}, nil
	}

	var userInfo UserInfo
	err := json.Unmarshal(providerSourceAuth, &userInfo)
	if err != nil {
		// If parsing fails, return default admin info
		return &UserInfo{
			UserID:   "admin",
			Username: "Admin User",
		}, nil
	}

	return &userInfo, nil
}

// getUserPermissions gets the user's permissions across all namespaces
func (a *AdminSessionAuthMethod) getUserPermissions(ctx context.Context, userID string) (map[string]string, error) {
	permissions := make(map[string]string)

	// Check if any group is admin
	if a.isAdmin {
		// Admin users get access to all namespaces - return empty map to signify admin
		return permissions, nil
	}

	// Get namespace permissions for each group
	for _, group := range a.userGroups {
		groupPermissions, err := a.userGroupRepo.GetNamespacePermissions(ctx, group.GetID())
		if err != nil {
			continue
		}

		for _, perm := range groupPermissions {
			namespaceName := a.getNamespaceName(perm.GetNamespaceID())
			if namespaceName == "" {
				continue
			}

			// Use the highest permission level if multiple permissions exist
			current, exists := permissions[namespaceName]
			permType := string(perm.GetPermissionType())
			if !exists || a.isHigherPermission(permType, current) {
				permissions[namespaceName] = permType
			}
		}
	}

	return permissions, nil
}

// getNamespaceName would get the namespace name from ID
// This is a placeholder - in a real implementation, you'd query the namespace repository
func (a *AdminSessionAuthMethod) getNamespaceName(namespaceID int) string {
	// Placeholder implementation
	return ""
}

// isHigherPermission checks if permission1 is higher level than permission2
func (a *AdminSessionAuthMethod) isHigherPermission(perm1, perm2 string) bool {
	permLevels := map[string]int{
		"READ":   1,
		"MODIFY": 2,
		"FULL":   3,
	}

	level1, exists1 := permLevels[perm1]
	level2, exists2 := permLevels[perm2]

	if !exists1 {
		return false
	}
	if !exists2 {
		return true
	}

	return level1 > level2
}

// UserInfo represents user information extracted from session
type UserInfo struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	Email    string `json:"email"`
}
