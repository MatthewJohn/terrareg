package auth

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/auth"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/auth/repository"
	moduleRepo "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/module/repository"
	infraConfig "github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/config"
)

// SessionAuthMethod implements immutable authentication for session-based users
type SessionAuthMethod struct {
	sessionRepo    repository.SessionRepository
	userGroupRepo  repository.UserGroupRepository
	namespaceRepo  moduleRepo.NamespaceRepository
	sessionManager auth.SessionManager
	config         *infraConfig.InfrastructureConfig
}

// NewSessionAuthMethod creates a new immutable session-based authentication method
func NewSessionAuthMethod(
	sessionRepo repository.SessionRepository,
	userGroupRepo repository.UserGroupRepository,
	namespaceRepo moduleRepo.NamespaceRepository,
	sessionManager auth.SessionManager,
	config *infraConfig.InfrastructureConfig,
) *SessionAuthMethod {
	return &SessionAuthMethod{
		sessionRepo:    sessionRepo,
		userGroupRepo:  userGroupRepo,
		namespaceRepo:  namespaceRepo,
		sessionManager: sessionManager,
		config:         config,
	}
}

// GetProviderType returns the authentication provider type
func (a *SessionAuthMethod) GetProviderType() auth.AuthMethodType {
	return auth.AuthMethodSession
}

// IsEnabled returns whether this authentication method is enabled
// Session auth requires session management to be available (SECRET_KEY configured)
func (a *SessionAuthMethod) IsEnabled() bool {
	// Simply check if sessionManager is set
	// The actual availability (SECRET_KEY configured) is checked by the session manager itself
	// when it's actually used, but we need to avoid calling IsAvailable() here because
	// it may panic if called on a nil receiver
	return a.sessionManager != nil
}

// Authenticate authenticates a request and returns an SessionAuthContext
// This implements the SessionAuthMethod interface, which receives sessionData from the auth factory
func (a *SessionAuthMethod) Authenticate(ctx context.Context, sessionData map[string]interface{}) (auth.AuthContext, error) {
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

	// Create SessionAuthContext with session state (convert string UserID to int for compatibility)
	// For now, use 0 as placeholder since we don't have proper user ID conversion
	userID := 0 // TODO: Convert userInfo.UserID from string to int when user ID system is defined
	authContext := auth.NewSessionAuthContext(ctx, userID, userInfo.Username, userInfo.Email, sessionID, a.config)

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
func (a *SessionAuthMethod) parseProviderSourceAuth(providerSourceAuth []byte) (*UserInfo, error) {
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
func (a *SessionAuthMethod) getUserPermissions(ctx context.Context, userGroups []string) (map[string]string, error) {
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
func (a *SessionAuthMethod) getNamespaceName(ctx context.Context, namespaceID int) string {
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
func (a *SessionAuthMethod) isHigherPermission(newPermission, existingPermission string) bool {
	permissionOrder := map[string]int{
		"READ":   1,
		"MODIFY": 2,
		"FULL":   3,
	}

	newLevel := permissionOrder[newPermission]
	existingLevel := permissionOrder[existingPermission]

	return newLevel > existingLevel
}

// AuthMethod interface implementation for the base SessionAuthMethod
// These return default values since the actual auth state is in the SessionAuthContext
// @TODO Can these be removed? The auth Method shouldn't need these, since everything uses AuthContext
func (a *SessionAuthMethod) IsBuiltInAdmin() bool                     { return false }
func (a *SessionAuthMethod) IsAdmin() bool                            { return false }
func (a *SessionAuthMethod) IsAuthenticated() bool                    { return false }
func (a *SessionAuthMethod) RequiresCSRF() bool                       { return true }
func (a *SessionAuthMethod) CheckAuthState() bool                     { return false }
func (a *SessionAuthMethod) CanPublishModuleVersion(string) bool      { return false }
func (a *SessionAuthMethod) CanUploadModuleVersion(string) bool       { return false }
func (a *SessionAuthMethod) CheckNamespaceAccess(string, string) bool { return false }
func (a *SessionAuthMethod) GetAllNamespacePermissions() map[string]string {
	return make(map[string]string)
}
func (a *SessionAuthMethod) GetUsername() string           { return "" }
func (a *SessionAuthMethod) GetUserGroupNames() []string   { return []string{} }
func (a *SessionAuthMethod) CanAccessReadAPI() bool        { return false }
func (a *SessionAuthMethod) CanAccessTerraformAPI() bool   { return false }
func (a *SessionAuthMethod) GetTerraformAuthToken() string { return "" }
func (a *SessionAuthMethod) GetProviderData() map[string]interface{} {
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
