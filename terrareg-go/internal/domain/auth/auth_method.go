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

// BaseAuthContextImplementation provides a concrete implementation of AuthContext interface
type BaseAuthContextImplementation struct {
	authMethod      AuthMethod
	sessionID       *string
	userGroups      []*UserGroup
	isAuthenticated bool
	username        string
	permissions     map[string]string // namespace -> permission_type
}

// NewBaseAuthContextImplementation creates a new authentication context implementation
func NewBaseAuthContextImplementation(authMethod AuthMethod) *BaseAuthContextImplementation {
	return &BaseAuthContextImplementation{
		authMethod:      authMethod,
		userGroups:      make([]*UserGroup, 0),
		isAuthenticated: true, // Default to authenticated for basic implementation
		username:        "base-user",
		permissions:     make(map[string]string),
	}
}

// HasPermission checks if the context has permission for a namespace
func (ac *BaseAuthContextImplementation) HasPermission(namespace, permissionType string) bool {
	if ac.IsAdmin() {
		return true
	}

	storedPermission, exists := ac.permissions[namespace]
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
func (ac *BaseAuthContextImplementation) GetUserGroupNames() []string {
	names := make([]string, len(ac.userGroups))
	for i, group := range ac.userGroups {
		names[i] = group.GetName()
	}
	return names
}

// AddUserGroup adds a user group to the context
func (ac *BaseAuthContextImplementation) AddUserGroup(group *UserGroup) {
	ac.userGroups = append(ac.userGroups, group)
}

// AuthContext interface implementation

// IsBuiltInAdmin returns true if this is a built-in admin
func (ac *BaseAuthContextImplementation) IsBuiltInAdmin() bool {
	return false // Default implementation
}

// IsAdmin returns true if user has admin privileges
func (ac *BaseAuthContextImplementation) IsAdmin() bool {
	return false // Default implementation - specific auth contexts should override
}

// IsAuthenticated returns true if user is authenticated
func (ac *BaseAuthContextImplementation) IsAuthenticated() bool {
	return ac.isAuthenticated
}

// RequiresCSRF returns true if CSRF protection is required
func (ac *BaseAuthContextImplementation) RequiresCSRF() bool {
	return true // Default implementation
}

// CheckAuthState returns true if auth state is valid
func (ac *BaseAuthContextImplementation) CheckAuthState() bool {
	return ac.isAuthenticated
}

// CanPublishModuleVersion returns true if user can publish to namespace
func (ac *BaseAuthContextImplementation) CanPublishModuleVersion(namespace string) bool {
	return false // Default implementation - specific auth contexts should override
}

// CanUploadModuleVersion returns true if user can upload to namespace
func (ac *BaseAuthContextImplementation) CanUploadModuleVersion(namespace string) bool {
	return false // Default implementation - specific auth contexts should override
}

// CheckNamespaceAccess returns true if user has access to namespace
func (ac *BaseAuthContextImplementation) CheckNamespaceAccess(permissionType, namespace string) bool {
	return false // Default implementation - specific auth contexts should override
}

// GetAllNamespacePermissions returns all namespace permissions
func (ac *BaseAuthContextImplementation) GetAllNamespacePermissions() map[string]string {
	return ac.permissions
}

// GetUsername returns the username
func (ac *BaseAuthContextImplementation) GetUsername() string {
	return ac.username
}

// CanAccessReadAPI returns true if user can access read API
func (ac *BaseAuthContextImplementation) CanAccessReadAPI() bool {
	return true // Default implementation - most authenticated users can access read API
}

// CanAccessTerraformAPI returns true if user can access Terraform API
func (ac *BaseAuthContextImplementation) CanAccessTerraformAPI() bool {
	return false // Default implementation - specific auth contexts should override
}

// GetTerraformAuthToken returns the Terraform auth token
func (ac *BaseAuthContextImplementation) GetTerraformAuthToken() string {
	return "" // Default implementation - specific auth contexts should override
}

// GetProviderType returns the auth method provider type
func (ac *BaseAuthContextImplementation) GetProviderType() AuthMethodType {
	return ac.authMethod.GetProviderType()
}

// GetProviderData returns provider-specific data
func (ac *BaseAuthContextImplementation) GetProviderData() map[string]interface{} {
	return make(map[string]interface{}) // Default implementation
}
