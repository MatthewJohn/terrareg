package service

import (
	"context"
	"sync"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/auth"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/auth/model"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/auth/repository"
	infraConfig "github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/config"
)

// AuthFactory implements the factory pattern for authentication methods
// Matches Python's AuthFactory behavior with priority-ordered discovery
type AuthFactory struct {
	authMethods       []auth.AuthMethod
	currentAuthMethod auth.AuthMethod
	currentAuthCtx    *auth.AuthContext
	mutex             sync.RWMutex
	sessionRepo       repository.SessionRepository
	userGroupRepo     repository.UserGroupRepository
	config            *infraConfig.InfrastructureConfig
}

// NewAuthFactory creates a new authentication factory
func NewAuthFactory(
	sessionRepo repository.SessionRepository,
	userGroupRepo repository.UserGroupRepository,
	config *infraConfig.InfrastructureConfig,
) *AuthFactory {
	factory := &AuthFactory{
		authMethods:   make([]auth.AuthMethod, 0),
		sessionRepo:   sessionRepo,
		userGroupRepo: userGroupRepo,
		config:        config,
	}

	// Initialize auth methods in priority order (matching Python)
	factory.initializeAuthMethods()

	return factory
}

// initializeAuthMethods sets up authentication methods in priority order
func (af *AuthFactory) initializeAuthMethods() {
	// Priority order from Python terrareg
	// 1. AdminApiKeyAuthMethod
	// 2. AdminSessionAuthMethod
	// 3. SamlAuthMethod
	// 4. OpenidConnectAuthMethod
	// 5. GithubAuthMethod
	// 6. TerraformOidcAuthMethod
	// 7. TerraformAnalyticsAuthKeyAuthMethod
	// 8. NotAuthenticated (fallback)

	// Register AdminApiKeyAuthMethod (highest priority)
	if af.config.AdminAuthenticationToken != "" {
		adminApiKeyAuthMethod := model.NewAdminApiKeyAuthMethod(af.config)
		af.RegisterAuthMethod(adminApiKeyAuthMethod)
	}

	// Register other auth methods (to be implemented in future phases)
	// TODO: Add AdminSessionAuthMethod, SamlAuthMethod, etc.

	// Register fallback method
	af.RegisterAuthMethod(&NotAuthenticatedAuthMethod{})
}

// RegisterAuthMethod registers an authentication method with the factory
func (af *AuthFactory) RegisterAuthMethod(authMethod auth.AuthMethod) {
	af.mutex.Lock()
	defer af.mutex.Unlock()

	af.authMethods = append(af.authMethods, authMethod)
}

// GetCurrentAuthMethod returns the currently authenticated auth method
func (af *AuthFactory) GetCurrentAuthMethod() auth.AuthMethod {
	af.mutex.RLock()
	defer af.mutex.RUnlock()

	if af.currentAuthMethod != nil {
		return af.currentAuthMethod
	}

	// Fallback to NotAuthenticated if no method is current
	return &NotAuthenticatedAuthMethod{}
}

// GetCurrentAuthContext returns the current authentication context
func (af *AuthFactory) GetCurrentAuthContext() *auth.AuthContext {
	af.mutex.RLock()
	defer af.mutex.RUnlock()

	if af.currentAuthCtx != nil {
		return af.currentAuthCtx
	}

	// Create context for NotAuthenticated
	return auth.NewAuthContext(&NotAuthenticatedAuthMethod{})
}

// AuthenticateRequest authenticates an HTTP request
func (af *AuthFactory) AuthenticateRequest(ctx context.Context, headers, formData, queryParams map[string]string) (*model.AuthenticationResponse, error) {
	af.mutex.Lock()
	defer af.mutex.Unlock()

	request := model.NewAuthenticationRequest(ctx, auth.AuthMethodNotAuthenticated, headers, formData, queryParams)
	request.Context = ctx

	// Try each auth method in priority order
	for _, authMethod := range af.authMethods {
		if authMethod.IsEnabled() && authMethod.CheckAuthState() {
			af.currentAuthMethod = authMethod
			af.currentAuthCtx = af.buildAuthContext(ctx, authMethod)

			response := model.NewAuthenticationResponse(request.RequestID, true, authMethod.GetProviderType())
			response.AuthMethod = authMethod.GetProviderType()
			response.Username = authMethod.GetUsername()
			response.IsAdmin = authMethod.IsAdmin()
			response.UserGroups = authMethod.GetUserGroupNames()
			response.Permissions = authMethod.GetAllNamespacePermissions()
			response.CanPublish = authMethod.CanPublishModuleVersion("")
			response.CanUpload = authMethod.CanUploadModuleVersion("")
			response.CanAccessAPI = authMethod.CanAccessReadAPI()
			response.CanAccessTerraform = authMethod.CanAccessTerraformAPI()
			response.TerraformToken = af.getStringPtr(authMethod.GetTerraformAuthToken())

			// Set session ID if available
			if sessionID := af.extractSessionID(headers); sessionID != nil {
				response.SessionID = sessionID
			}

			return response, nil
		}
	}

	// No authentication method succeeded - return NotAuthenticated
	fallbackMethod := &NotAuthenticatedAuthMethod{}
	af.currentAuthMethod = fallbackMethod
	af.currentAuthCtx = auth.NewAuthContext(fallbackMethod)

	response := model.NewAuthenticationResponse(request.RequestID, false, auth.AuthMethodNotAuthenticated)
	response.AuthMethod = auth.AuthMethodNotAuthenticated
	response.Username = fallbackMethod.GetUsername()
	response.CanAccessAPI = fallbackMethod.CanAccessReadAPI()

	return response, nil
}

// CheckNamespacePermission checks if the current auth context has permission for a namespace
func (af *AuthFactory) CheckNamespacePermission(namespace, permissionType string) bool {
	authCtx := af.GetCurrentAuthContext()

	// Admin users have access to everything
	if authCtx.AuthMethod.IsAdmin() {
		return true
	}

	// Check permissions in context
	return authCtx.HasPermission(namespace, permissionType)
}

// CanPublishModuleVersion checks if current user can publish to a namespace
func (af *AuthFactory) CanPublishModuleVersion(namespace string) bool {
	authMethod := af.GetCurrentAuthMethod()
	return authMethod.CanPublishModuleVersion(namespace)
}

// CanUploadModuleVersion checks if current user can upload to a namespace
func (af *AuthFactory) CanUploadModuleVersion(namespace string) bool {
	authMethod := af.GetCurrentAuthMethod()
	return authMethod.CanUploadModuleVersion(namespace)
}

// CanAccessReadAPI checks if current user can access the read API
func (af *AuthFactory) CanAccessReadAPI() bool {
	authMethod := af.GetCurrentAuthMethod()
	return authMethod.CanAccessReadAPI()
}

// CanAccessTerraformAPI checks if current user can access Terraform API
func (af *AuthFactory) CanAccessTerraformAPI() bool {
	authMethod := af.GetCurrentAuthMethod()
	return authMethod.CanAccessTerraformAPI()
}

// InvalidateAuthentication clears the current authentication state
func (af *AuthFactory) InvalidateAuthentication() {
	af.mutex.Lock()
	defer af.mutex.Unlock()

	af.currentAuthMethod = nil
	af.currentAuthCtx = nil
}

// buildAuthContext builds an authentication context for an auth method
func (af *AuthFactory) buildAuthContext(ctx context.Context, authMethod auth.AuthMethod) *auth.AuthContext {
	authCtx := auth.NewAuthContext(authMethod)

	// Load user groups for the authentication method
	userGroups, err := af.loadUserGroupsForAuthMethod(ctx, authMethod)
	if err == nil {
		for _, group := range userGroups {
			authCtx.AddUserGroup(group)
		}
	}

	return authCtx
}

// loadUserGroupsForAuthMethod loads user groups for an authentication method
func (af *AuthFactory) loadUserGroupsForAuthMethod(ctx context.Context, authMethod auth.AuthMethod) ([]*auth.UserGroup, error) {
	// This will be implemented when we add concrete auth methods
	// For now, return empty list
	return []*auth.UserGroup{}, nil
}

// extractSessionID extracts session ID from headers
func (af *AuthFactory) extractSessionID(headers map[string]string) *string {
	// Check for session cookie or header
	if cookie, exists := headers["Cookie"]; exists {
		// Parse session from cookie
		// Implementation depends on cookie format
		return af.parseSessionFromCookie(cookie)
	}

	if sessionID, exists := headers["X-Session-ID"]; exists {
		return &sessionID
	}

	return nil
}

// parseSessionFromCookie parses session ID from cookie string
func (af *AuthFactory) parseSessionFromCookie(cookie string) *string {
	// Simple implementation - in practice, this would parse cookie format properly
	// This matches Python's session cookie handling
	if cookie == "" {
		return nil
	}

	// For now, just return the cookie as session ID
	// Real implementation would parse "session_id=<value>" format
	return &cookie
}

// getStringPtr returns a pointer to a string or nil if empty
func (af *AuthFactory) getStringPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// NotAuthenticatedAuthMethod is the fallback authentication method
type NotAuthenticatedAuthMethod struct {
	auth.BaseAuthMethod
}

func (n *NotAuthenticatedAuthMethod) IsEnabled() bool {
	return true // Always enabled as fallback
}

func (n *NotAuthenticatedAuthMethod) CheckAuthState() bool {
	return true // Always succeeds as fallback
}

func (n *NotAuthenticatedAuthMethod) IsAuthenticated() bool {
	return false // Not actually authenticated
}

func (n *NotAuthenticatedAuthMethod) GetProviderType() auth.AuthMethodType {
	return auth.AuthMethodNotAuthenticated
}

func (n *NotAuthenticatedAuthMethod) GetUsername() string {
	return "Anonymous User"
}

func (n *NotAuthenticatedAuthMethod) GetUserGroupNames() []string {
	return []string{}
}

func (n *NotAuthenticatedAuthMethod) RequiresCSRF() bool {
	return false
}
