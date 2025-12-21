package model

import (
	"context"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/auth"
)

// notAuthenticatedAuthMethod is a simple auth method that represents no authentication
type notAuthenticatedAuthMethod struct{}

// Implement AuthMethod interface
func (n *notAuthenticatedAuthMethod) IsBuiltInAdmin() bool               { return false }
func (n *notAuthenticatedAuthMethod) IsAdmin() bool                     { return false }
func (n *notAuthenticatedAuthMethod) IsAuthenticated() bool              { return false }
func (n *notAuthenticatedAuthMethod) IsEnabled() bool                    { return true }
func (n *notAuthenticatedAuthMethod) RequiresCSRF() bool                  { return false }
func (n *notAuthenticatedAuthMethod) CheckAuthState() bool                 { return true }
func (n *notAuthenticatedAuthMethod) CanPublishModuleVersion(string) bool { return false }
func (n *notAuthenticatedAuthMethod) CanUploadModuleVersion(string) bool  { return false }
func (n *notAuthenticatedAuthMethod) CheckNamespaceAccess(string, string) bool { return false }
func (n *notAuthenticatedAuthMethod) GetAllNamespacePermissions() map[string]string { return make(map[string]string) }
func (n *notAuthenticatedAuthMethod) GetUsername() string                { return "" }
func (n *notAuthenticatedAuthMethod) GetUserGroupNames() []string       { return []string{} }
func (n *notAuthenticatedAuthMethod) CanAccessReadAPI() bool             { return false }
func (n *notAuthenticatedAuthMethod) CanAccessTerraformAPI() bool       { return false }
func (n *notAuthenticatedAuthMethod) GetTerraformAuthToken() string     { return "" }
func (n *notAuthenticatedAuthMethod) GetProviderData() map[string]interface{} { return make(map[string]interface{}) }
func (n *notAuthenticatedAuthMethod) GetProviderType() auth.AuthMethodType { return auth.AuthMethodNotAuthenticated }

// AuthMethodAdapter is an interface that wraps an AuthMethod with additional state
// This allows AuthMethod implementations to remain immutable while still providing
// request-specific authentication information
type AuthMethodAdapter interface {
	auth.AuthMethod
	// GetWrappedAuthMethod returns the underlying immutable AuthMethod
	GetWrappedAuthMethod() auth.AuthMethod

	// GetContext returns the context associated with this authentication
	GetContext() context.Context

	// GetBaseAdapter returns the base adapter for method chaining
	GetBaseAdapter() AuthMethodAdapter

	// SetBaseAdapter sets the base adapter for method chaining
	SetBaseAdapter(AuthMethodAdapter)
}

// BaseAdapter provides common functionality for AuthMethod adapters
type BaseAdapter struct {
	wrappedMethod auth.AuthMethod
	ctx           context.Context
	baseAdapter    AuthMethodAdapter
}

// NewBaseAdapter creates a new base adapter
func NewBaseAdapter(method auth.AuthMethod, ctx context.Context) *BaseAdapter {
	return &BaseAdapter{
		wrappedMethod: method,
		ctx:           ctx,
	}
}

// GetWrappedAuthMethod returns the wrapped AuthMethod
func (a *BaseAdapter) GetWrappedAuthMethod() auth.AuthMethod {
	return a.wrappedMethod
}

// GetContext returns the context
func (a *BaseAdapter) GetContext() context.Context {
	return a.ctx
}

// GetBaseAdapter returns the base adapter
func (a *BaseAdapter) GetBaseAdapter() AuthMethodAdapter {
	return a.baseAdapter
}

// SetBaseAdapter sets the base adapter for method chaining
func (a *BaseAdapter) SetBaseAdapter(base AuthMethodAdapter) {
	a.baseAdapter = base
}

// Forward wrapped method calls to the underlying AuthMethod
func (a *BaseAdapter) GetProviderType() auth.AuthMethodType {
	return a.wrappedMethod.GetProviderType()
}

func (a *BaseAdapter) IsBuiltInAdmin() bool {
	return a.wrappedMethod.IsBuiltInAdmin()
}

func (a *BaseAdapter) IsAdmin() bool {
	return a.wrappedMethod.IsAdmin()
}

func (a *BaseAdapter) IsAuthenticated() bool {
	return a.wrappedMethod.IsAuthenticated()
}

func (a *BaseAdapter) IsEnabled() bool {
	return a.wrappedMethod.IsEnabled()
}

func (a *BaseAdapter) RequiresCSRF() bool {
	return a.wrappedMethod.RequiresCSRF()
}

func (a *BaseAdapter) CheckAuthState() bool {
	return a.wrappedMethod.CheckAuthState()
}

func (a *BaseAdapter) CanPublishModuleVersion(namespace string) bool {
	return a.wrappedMethod.CanPublishModuleVersion(namespace)
}

func (a *BaseAdapter) CanUploadModuleVersion(namespace string) bool {
	return a.wrappedMethod.CanUploadModuleVersion(namespace)
}

func (a *BaseAdapter) CheckNamespaceAccess(permissionType, namespace string) bool {
	return a.wrappedMethod.CheckNamespaceAccess(permissionType, namespace)
}

func (a *BaseAdapter) GetAllNamespacePermissions() map[string]string {
	return a.wrappedMethod.GetAllNamespacePermissions()
}

func (a *BaseAdapter) GetUsername() string {
	return a.wrappedMethod.GetUsername()
}

func (a *BaseAdapter) GetUserGroupNames() []string {
	return a.wrappedMethod.GetUserGroupNames()
}

func (a *BaseAdapter) CanAccessReadAPI() bool {
	return a.wrappedMethod.CanAccessReadAPI()
}

func (a *BaseAdapter) CanAccessTerraformAPI() bool {
	return a.wrappedMethod.CanAccessTerraformAPI()
}

func (a *BaseAdapter) GetTerraformAuthToken() string {
	return a.wrappedMethod.GetTerraformAuthToken()
}

func (a *BaseAdapter) GetProviderData() map[string]interface{} {
	return a.wrappedMethod.GetProviderData()
}

// SessionStateAdapter wraps an AuthMethod with session-specific state
type SessionStateAdapter struct {
	*BaseAdapter
	sessionID      *string
	isAuthenticated bool
	username       string
	email          string
	userGroups     []*auth.UserGroup
	permissions    map[string]string
}

// NewSessionStateAdapter creates a new session state adapter
func NewSessionStateAdapter(method auth.AuthMethod, ctx context.Context) *SessionStateAdapter {
	return &SessionStateAdapter{
		BaseAdapter: NewBaseAdapter(method, ctx),
		permissions: make(map[string]string),
		userGroups:  make([]*auth.UserGroup, 0),
	}
}

// SetSessionData sets session-specific data
func (a *SessionStateAdapter) SetSessionData(sessionID *string, authenticated bool, username, email string) {
	a.sessionID = sessionID
	a.isAuthenticated = authenticated
	a.username = username
	a.email = email
}

// AddUserGroup adds a user group
func (a *SessionStateAdapter) AddUserGroup(group *auth.UserGroup) {
	a.userGroups = append(a.userGroups, group)
}

// SetPermission sets a namespace permission
func (a *SessionStateAdapter) SetPermission(namespace, permission string) {
	if a.permissions == nil {
		a.permissions = make(map[string]string)
	}
	a.permissions[namespace] = permission
}

// SetAdmin sets admin status (not stored, just computed)
func (a *SessionStateAdapter) SetAdmin(isAdmin bool) {
	// Admin status is computed from groups, not stored
}

// Override methods to use session-specific state
func (a *SessionStateAdapter) IsAuthenticated() bool {
	return a.isAuthenticated
}

func (a *SessionStateAdapter) GetUsername() string {
	return a.username
}

func (a *SessionStateAdapter) IsAdmin() bool {
	// Check if any group has admin rights
	for _, group := range a.userGroups {
		if group.SiteAdmin {
			return true
		}
	}
	return a.wrappedMethod.IsAdmin()
}

func (a *SessionStateAdapter) GetUserGroupNames() []string {
	names := make([]string, len(a.userGroups))
	for i, group := range a.userGroups {
		names[i] = group.Name
	}
	return names
}

func (a *SessionStateAdapter) GetAllNamespacePermissions() map[string]string {
	return a.permissions
}

// SetBaseAdapter implements AuthMethodAdapter interface
func (a *SessionStateAdapter) SetBaseAdapter(base AuthMethodAdapter) {
	a.baseAdapter = base
}

func (a *SessionStateAdapter) CheckNamespaceAccess(permissionType, namespace string) bool {
	storedPermission, exists := a.permissions[namespace]
	if !exists {
		return false
	}

	// Check permission hierarchy
	switch permissionType {
	case "READ":
		return storedPermission == "READ" || storedPermission == "MODIFY" || storedPermission == "FULL"
	case "MODIFY":
		return storedPermission == "MODIFY" || storedPermission == "FULL"
	case "FULL":
		return storedPermission == "FULL"
	default:
		return false
	}
}

// PermissionCheckingAdapter wraps an AuthMethod with computed permissions
type PermissionCheckingAdapter struct {
	*BaseAdapter
	permissions map[string]string
	isAdmin     bool
}

// NewPermissionCheckingAdapter creates a new permission checking adapter
func NewPermissionCheckingAdapter(method auth.AuthMethod, ctx context.Context) *PermissionCheckingAdapter {
	return &PermissionCheckingAdapter{
		BaseAdapter: NewBaseAdapter(method, ctx),
		permissions: make(map[string]string),
	}
}

// SetAdmin sets admin status
func (a *PermissionCheckingAdapter) SetAdmin(isAdmin bool) {
	a.isAdmin = isAdmin
}

// SetPermissions sets namespace permissions
func (a *PermissionCheckingAdapter) SetPermissions(permissions map[string]string) {
	a.permissions = permissions
}

// AddPermission adds a namespace permission
func (a *PermissionCheckingAdapter) AddPermission(namespace, permission string) {
	if a.permissions == nil {
		a.permissions = make(map[string]string)
	}
	a.permissions[namespace] = permission
}

// Override methods to use adapter state
func (a *PermissionCheckingAdapter) IsAdmin() bool {
	return a.isAdmin || a.wrappedMethod.IsAdmin()
}

func (a *PermissionCheckingAdapter) GetAllNamespacePermissions() map[string]string {
	// Merge wrapped method permissions with adapter permissions
	merged := make(map[string]string)

	// Add wrapped method permissions
	for k, v := range a.wrappedMethod.GetAllNamespacePermissions() {
		merged[k] = v
	}

	// Override with adapter permissions
	for k, v := range a.permissions {
		merged[k] = v
	}

	return merged
}

func (a *PermissionCheckingAdapter) CheckNamespaceAccess(permissionType, namespace string) bool {
	// If admin, allow everything
	if a.IsAdmin() {
		return true
	}

	// Check permissions
	return a.HasPermissionHierarchy(a.permissions[namespace], permissionType)
}

// HasPermissionHierarchy checks if the stored permission meets or exceeds the required permission
func (a *PermissionCheckingAdapter) HasPermissionHierarchy(stored, required string) bool {
	switch required {
	case "READ":
		return stored == "READ" || stored == "MODIFY" || stored == "FULL"
	case "MODIFY":
		return stored == "MODIFY" || stored == "FULL"
	case "FULL":
		return stored == "FULL"
	default:
		return false
	}
}

// SetBaseAdapter implements AuthMethodAdapter interface
func (a *PermissionCheckingAdapter) SetBaseAdapter(base AuthMethodAdapter) {
	a.baseAdapter = base
}

// AuthContextBuilder helps build AuthContext instances with adapters
type AuthContextBuilder struct {
	ctx        context.Context
	adapters   []AuthMethodAdapter
}

// NewAuthContextBuilder creates a new builder
func NewAuthContextBuilder(ctx context.Context) *AuthContextBuilder {
	return &AuthContextBuilder{
		ctx:      ctx,
		adapters: make([]AuthMethodAdapter, 0),
	}
}

// WithAdapter adds an adapter to the chain
func (b *AuthContextBuilder) WithAdapter(adapter AuthMethodAdapter) *AuthContextBuilder {
	adapter.SetBaseAdapter(b.getLastAdapter())
	b.adapters = append(b.adapters, adapter)
	return b
}

// getLastAdapter returns the last adapter in the chain
func (b *AuthContextBuilder) getLastAdapter() AuthMethodAdapter {
	if len(b.adapters) == 0 {
		return nil
	}
	return b.adapters[len(b.adapters)-1]
}

// Build builds the final AuthContext
func (b *AuthContextBuilder) Build() auth.AuthMethod {
	if len(b.adapters) == 0 {
		// Create a simple not authenticated adapter
		notAuthMethod := &notAuthenticatedAuthMethod{}

		return &PermissionCheckingAdapter{
			BaseAdapter: NewBaseAdapter(notAuthMethod, b.ctx),
			permissions: make(map[string]string),
			isAdmin:     false,
		}
	}

	// Return the top adapter (which has all the chain)
	return b.adapters[len(b.adapters)-1]
}