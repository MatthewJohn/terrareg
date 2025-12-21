package auth

import (
	"context"
)

// AuthMethod defines the interface for authentication method factories
// These are lightweight objects that create AuthContext instances
type AuthMethod interface {
	// Factory methods
	GetProviderType() AuthMethodType
	IsEnabled() bool
}

// SessionAuthMethod defines interface for auth methods that work with session data
type SessionAuthMethod interface {
	AuthMethod
	Authenticate(ctx context.Context, sessionData map[string]interface{}) (AuthContext, error)
}

// HeaderAuthMethod defines interface for auth methods that work with HTTP headers
type HeaderAuthMethod interface {
	AuthMethod
	Authenticate(ctx context.Context, headers, formData, queryParams map[string]string) (AuthContext, error)
}

// TokenAuthMethod defines interface for auth methods that work with tokens
type TokenAuthMethod interface {
	AuthMethod
	Authenticate(ctx context.Context, token string) (AuthContext, error)
}

// BearerTokenAuthMethod defines interface for auth methods that work with bearer tokens
type BearerTokenAuthMethod interface {
	AuthMethod
	Authenticate(ctx context.Context, authorizationHeader string, requestData []byte) (AuthContext, error)
}

// AuthContext defines the interface for authenticated contexts
// This contains the actual permission logic and user information
// Based on Python's BaseAuthMethod class
type AuthContext interface {
	// Core authentication methods
	IsBuiltInAdmin() bool
	IsAdmin() bool
	IsAuthenticated() bool

	// Configuration and state
	RequiresCSRF() bool
	CheckAuthState() bool

	// Permission methods
	CanPublishModuleVersion(namespace string) bool
	CanUploadModuleVersion(namespace string) bool
	CheckNamespaceAccess(permissionType, namespace string) bool
	GetAllNamespacePermissions() map[string]string

	// User information
	GetUsername() string
	GetUserGroupNames() []string

	// API access methods
	CanAccessReadAPI() bool
	CanAccessTerraformAPI() bool
	GetTerraformAuthToken() string

	// Provider-specific data
	GetProviderType() AuthMethodType
	GetProviderData() map[string]interface{}
}

// AuthMethodType represents different authentication method types
type AuthMethodType string

const (
	AuthMethodNotAuthenticated                AuthMethodType = "NOT_AUTHENTICATED"
	AuthMethodAdminSession                    AuthMethodType = "ADMIN_SESSION"
	AuthMethodAdminApiKey                     AuthMethodType = "ADMIN_API_KEY"
	AuthMethodSAML                            AuthMethodType = "SAML"
	AuthMethodOpenIDConnect                   AuthMethodType = "OPENID_CONNECT"
	AuthMethodGitHub                          AuthMethodType = "GITHUB"
	AuthMethodTerraformOIDC                   AuthMethodType = "TERRAFORM_OIDC"
	AuthMethodTerraformAnalyticsAuthKey       AuthMethodType = "TERRAFORM_ANALYTICS_AUTH_KEY"
	AuthMethodTerraformIgnoreAnalyticsAuthKey AuthMethodType = "TERRAFORM_IGNORE_ANALYTICS_AUTH_KEY"
	AuthMethodTerraformInternalExtraction     AuthMethodType = "TERRAFORM_INTERNAL_EXTRACTION"
	AuthMethodUploadApiKey                    AuthMethodType = "UPLOAD_API_KEY"
	AuthMethodPublishApiKey                   AuthMethodType = "PUBLISH_API_KEY"
)

// BaseAuthMethod provides common functionality for authentication methods
type BaseAuthMethod struct{}

// NewBaseAuthMethod creates a new base authentication method
func NewBaseAuthMethod() *BaseAuthMethod {
	return &BaseAuthMethod{}
}

// Default implementations for BaseAuthMethod

func (b *BaseAuthMethod) IsBuiltInAdmin() bool {
	return false
}

func (b *BaseAuthMethod) IsAdmin() bool {
	return false
}

func (b *BaseAuthMethod) IsAuthenticated() bool {
	return true
}

func (b *BaseAuthMethod) RequiresCSRF() bool {
	return true
}

func (b *BaseAuthMethod) CanPublishModuleVersion(namespace string) bool {
	return false
}

func (b *BaseAuthMethod) CanUploadModuleVersion(namespace string) bool {
	return false
}

func (b *BaseAuthMethod) CheckNamespaceAccess(permissionType, namespace string) bool {
	return false
}

func (b *BaseAuthMethod) GetAllNamespacePermissions() map[string]string {
	return make(map[string]string)
}

func (b *BaseAuthMethod) CanAccessReadAPI() bool {
	return true
}

func (b *BaseAuthMethod) CanAccessTerraformAPI() bool {
	return false
}

func (b *BaseAuthMethod) GetTerraformAuthToken() string {
	return ""
}

func (b *BaseAuthMethod) GetProviderData() map[string]interface{} {
	return make(map[string]interface{})
}

// BaseAuthContext provides common context for all auth contexts
type BaseAuthContext struct {
	ctx context.Context
}

// GetContext returns the context
func (b *BaseAuthContext) GetContext() context.Context {
	return b.ctx
}

// AuthContext represents the current authentication context
type AuthContext struct {
	AuthMethod      AuthMethod
	SessionID       *string
	UserGroups      []*UserGroup
	IsAuthenticated bool
	Username        string
	Permissions     map[string]string // namespace -> permission_type
}

// NewAuthContext creates a new authentication context
func NewAuthContext(authMethod AuthMethod) *AuthContext {
	return &AuthContext{
		AuthMethod:      authMethod,
		UserGroups:      make([]*UserGroup, 0),
		IsAuthenticated: authMethod.IsAuthenticated(),
		Username:        authMethod.GetUsername(),
		Permissions:     authMethod.GetAllNamespacePermissions(),
	}
}

// HasPermission checks if the context has permission for a namespace
func (ac *AuthContext) HasPermission(namespace, permissionType string) bool {
	if ac.AuthMethod.IsAdmin() {
		return true
	}

	storedPermission, exists := ac.Permissions[namespace]
	if !exists {
		return false
	}

	// Check permission hierarchy
	switch PermissionType(permissionType) {
	case PermissionRead:
		return storedPermission == string(PermissionRead) ||
			storedPermission == string(PermissionModify) ||
			storedPermission == string(PermissionFull)
	case PermissionModify:
		return storedPermission == string(PermissionModify) ||
			storedPermission == string(PermissionFull)
	case PermissionFull:
		return storedPermission == string(PermissionFull)
	default:
		return false
	}
}

// GetUserGroupNames returns the names of all user groups in the context
func (ac *AuthContext) GetUserGroupNames() []string {
	names := make([]string, len(ac.UserGroups))
	for i, group := range ac.UserGroups {
		names[i] = group.GetName()
	}
	return names
}

// AddUserGroup adds a user group to the context
func (ac *AuthContext) AddUserGroup(group *UserGroup) {
	ac.UserGroups = append(ac.UserGroups, group)
}
