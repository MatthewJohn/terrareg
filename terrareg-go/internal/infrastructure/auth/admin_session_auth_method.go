package auth

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/auth"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/auth/repository"
	moduleRepo "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/module/repository"
)

// AdminSessionAuthMethod implements immutable authentication for admin users via session cookies
type AdminSessionAuthMethod struct {
	sessionRepo     repository.SessionRepository
	userGroupRepo   repository.UserGroupRepository
	namespaceRepo   moduleRepo.NamespaceRepository
	sessionManager  auth.SessionManager
}

// NewAdminSessionAuthMethod creates a new immutable admin session authentication method
func NewAdminSessionAuthMethod(
	sessionRepo repository.SessionRepository,
	userGroupRepo repository.UserGroupRepository,
	namespaceRepo moduleRepo.NamespaceRepository,
	sessionManager auth.SessionManager,
) *AdminSessionAuthMethod {
	return &AdminSessionAuthMethod{
		sessionRepo:     sessionRepo,
		userGroupRepo:   userGroupRepo,
		namespaceRepo:   namespaceRepo,
		sessionManager:  sessionManager,
	}
}

// GetProviderType returns the authentication provider type
func (a *AdminSessionAuthMethod) GetProviderType() auth.AuthMethodType {
	return auth.AuthMethodAdminSession
}

// IsEnabled returns whether this authentication method is enabled
// Admin session auth requires session management to be available (SECRET_KEY configured)
func (a *AdminSessionAuthMethod) IsEnabled() bool {
	// Simply check if sessionManager is set
	// The actual availability (SECRET_KEY configured) is checked by the session manager itself
	// when it's actually used, but we need to avoid calling IsAvailable() here because
	// it may panic if called on a nil receiver
	return a.sessionManager != nil
}

// Authenticate authenticates a request and returns an AdminSessionAuthContext
// This implements the SessionAuthMethod interface, which receives sessionData from the auth factory
func (a *AdminSessionAuthMethod) Authenticate(ctx context.Context, sessionData map[string]interface{}) (auth.AuthContext, error) {
	// Get session ID from sessionData map (populated by auth factory from cookies/headers)
	sessionIDInterface, exists := sessionData["session_id"]
	if !exists {
		return nil, nil // No session ID, let other auth methods try
	}

	sessionID, ok := sessionIDInterface.(string)
	if !ok || sessionID == "" || strings.TrimSpace(sessionID) == "" {
		return nil, nil // Invalid session ID, let other auth methods try
	}

	// Find session in database
	session, err := a.sessionRepo.FindByID(ctx, sessionID)
	if err != nil || session == nil {
		return nil, nil // Let other auth methods try
	}

	// Check if session is expired
	if session.IsExpired() {
		return nil, nil // Let other auth methods try
	}

	// Parse provider source auth to get user information
	userInfo, err := a.parseProviderSourceAuth(session.ProviderSourceAuth)
	if err != nil {
		return nil, nil // Let other auth methods try
	}

	// Create AdminSessionAuthContext with session state (convert string UserID to int for compatibility)
	// For now, use 0 as placeholder since we don't have proper user ID conversion
	userID := 0 // TODO: Convert userInfo.UserID from string to int when user ID system is defined
	authContext := auth.NewAdminSessionAuthContext(ctx, userID, userInfo.Username, userInfo.Email, sessionID)

	// Set admin status based on site_admin or is_admin flag from session
	// Support both field names for backward compatibility
	isAdmin := userInfo.SiteAdmin || userInfo.IsAdmin
	authContext.SetAdmin(isAdmin)

	// Get user permissions from user groups
	permissions, err := a.getUserPermissions(ctx, userInfo.UserGroups)
	if err == nil && permissions != nil {
		for namespace, permission := range permissions {
			authContext.SetPermission(namespace, permission)
		}
	}

	return authContext, nil
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
// Returns a map of namespace name -> permission type string
func (a *AdminSessionAuthMethod) getUserPermissions(ctx context.Context, userGroups []string) (map[string]string, error) {
	permissions := make(map[string]string)

	// Query the database for user group permissions
	for _, groupName := range userGroups {
		group, err := a.userGroupRepo.FindByName(ctx, groupName)
		if err != nil || group == nil {
			continue
		}

		// Get namespace permissions for this user group from repository
		groupPermissions, err := a.userGroupRepo.GetNamespacePermissions(ctx, group.GetID())
		if err != nil {
			continue
		}

		// Convert namespace permissions to string map
		for _, perm := range groupPermissions {
			namespaceName := a.getNamespaceName(ctx, perm.GetNamespaceID())
			if namespaceName == "" {
				continue
			}
			permissionType := string(perm.GetPermissionType())

			// Keep the highest permission level (READ < MODIFY < FULL)
			existingPermission, exists := permissions[namespaceName]
			if !exists || a.isHigherPermission(permissionType, existingPermission) {
				permissions[namespaceName] = permissionType
			}
		}
	}

	return permissions, nil
}

// getNamespaceName gets namespace name from ID using a lookup
func (a *AdminSessionAuthMethod) getNamespaceName(ctx context.Context, namespaceID int) string {
	if a.namespaceRepo == nil {
		return ""
	}

	namespace, err := a.namespaceRepo.FindByID(ctx, namespaceID)
	if err != nil || namespace == nil {
		return ""
	}

	return string(namespace.Name())
}

// isHigherPermission returns true if newPermission is higher than existingPermission
func (a *AdminSessionAuthMethod) isHigherPermission(newPermission, existingPermission string) bool {
	permissionOrder := map[string]int{
		"READ":   1,
		"MODIFY": 2,
		"FULL":   3,
	}

	newLevel := permissionOrder[newPermission]
	existingLevel := permissionOrder[existingPermission]

	return newLevel > existingLevel
}

// AuthMethod interface implementation for the base AdminSessionAuthMethod
// These return default values since the actual auth state is in the AdminSessionAuthContext
// @TODO Can these be removed? The auth Method shouldn't need these, since everything uses AuthContext
func (a *AdminSessionAuthMethod) IsBuiltInAdmin() bool                     { return false }
func (a *AdminSessionAuthMethod) IsAdmin() bool                            { return false }
func (a *AdminSessionAuthMethod) IsAuthenticated() bool                    { return false }
func (a *AdminSessionAuthMethod) RequiresCSRF() bool                       { return true }
func (a *AdminSessionAuthMethod) CheckAuthState() bool                     { return false }
func (a *AdminSessionAuthMethod) CanPublishModuleVersion(string) bool      { return false }
func (a *AdminSessionAuthMethod) CanUploadModuleVersion(string) bool       { return false }
func (a *AdminSessionAuthMethod) CheckNamespaceAccess(string, string) bool { return false }
func (a *AdminSessionAuthMethod) GetAllNamespacePermissions() map[string]string {
	return make(map[string]string)
}
func (a *AdminSessionAuthMethod) GetUsername() string           { return "" }
func (a *AdminSessionAuthMethod) GetUserGroupNames() []string   { return []string{} }
func (a *AdminSessionAuthMethod) CanAccessReadAPI() bool        { return false }
func (a *AdminSessionAuthMethod) CanAccessTerraformAPI() bool   { return false }
func (a *AdminSessionAuthMethod) GetTerraformAuthToken() string { return "" }
func (a *AdminSessionAuthMethod) GetProviderData() map[string]interface{} {
	return make(map[string]interface{})
}

// UserInfo represents user information extracted from session
type UserInfo struct {
	UserID     string   `json:"user_id"`
	Username   string   `json:"username"`
	Email      string   `json:"email"`
	UserGroups []string `json:"user_groups"`
	SiteAdmin  bool     `json:"site_admin"`
	// IsAdmin is an alias for SiteAdmin to support both JSON field names
	// Some code uses "is_admin" while other uses "site_admin"
	IsAdmin bool `json:"is_admin"`
}
