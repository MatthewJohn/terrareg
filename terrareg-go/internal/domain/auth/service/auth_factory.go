package service

import (
	"context"
	"fmt"
	"sync"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/auth"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/auth/model"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/auth/repository"
	infraAuth "github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/auth"
	infraConfig "github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/config"
	"github.com/rs/zerolog"
)

// TerraformIDPServiceAdapter wraps the domain service to implement the infrastructure interface
type TerraformIDPServiceAdapter struct {
	service interface{}
}

// ValidateToken implements the infrastructure interface
func (a *TerraformIDPServiceAdapter) ValidateToken(ctx context.Context, token string) (interface{}, error) {
	// Use type assertion to call the domain service method
	if validator, ok := a.service.(interface {
		ValidateToken(ctx context.Context, token string) (interface{}, error)
	}); ok {
		return validator.ValidateToken(ctx, token)
	}
	return nil, fmt.Errorf("service does not implement ValidateToken")
}

// AuthFactory implements the factory pattern for authentication methods
// Matches Python's AuthFactory behavior with priority-ordered discovery
type AuthFactory struct {
	authMethods         []auth.AuthMethod
	currentAuthMethod   auth.AuthMethod
	currentAuthCtx      *auth.AuthContext
	mutex               sync.RWMutex
	sessionRepo         repository.SessionRepository
	userGroupRepo       repository.UserGroupRepository
	config              *infraConfig.InfrastructureConfig
	terraformIDPService interface{}
	logger              *zerolog.Logger
}

// NewAuthFactory creates a new authentication factory
func NewAuthFactory(
	sessionRepo repository.SessionRepository,
	userGroupRepo repository.UserGroupRepository,
	config *infraConfig.InfrastructureConfig,
	terraformIDPService interface{},
	logger *zerolog.Logger,
) *AuthFactory {
	factory := &AuthFactory{
		authMethods:         make([]auth.AuthMethod, 0),
		sessionRepo:         sessionRepo,
		userGroupRepo:       userGroupRepo,
		config:              config,
		terraformIDPService: terraformIDPService,
		logger:              logger,
	}

	// Initialize auth methods in priority order (matching Python)
	factory.initializeAuthMethods()

	return factory
}

// initializeAuthMethods sets up authentication methods in priority order
func (af *AuthFactory) initializeAuthMethods() {
	// Priority order from Python terrareg
	// 1. AdminApiKeyAuthMethod (legacy)
	// 2. AuthenticationTokenAuthMethod (admin tokens)
	// 3. AuthenticationTokenAuthMethod (upload tokens)
	// 4. AuthenticationTokenAuthMethod (publish tokens)
	// 5. AdminSessionAuthMethod
	// 6. SamlAuthMethod
	// 7. OpenidConnectAuthMethod
	// 8. GithubAuthMethod
	// 9. TerraformOidcAuthMethod
	// 10. TerraformAnalyticsAuthKeyAuthMethod
	// 11. NotAuthenticated (fallback)

	// Register AdminApiKeyAuthMethod (legacy admin token - highest priority)
	if af.config.AdminAuthenticationToken != "" {
		adminApiKeyAuthMethod := infraAuth.NewAdminApiKeyAuthMethod(af.config)
		af.RegisterAuthMethod(adminApiKeyAuthMethod)
	}

	// Register UploadApiKeyAuthMethod for upload tokens
	if len(af.config.UploadApiKeys) > 0 {
		uploadApiKeyAuthMethod := infraAuth.NewUploadApiKeyAuthMethod(af.config)
		af.RegisterAuthMethod(uploadApiKeyAuthMethod)
	}

	// Register PublishApiKeyAuthMethod for publish tokens
	if len(af.config.PublishApiKeys) > 0 {
		publishApiKeyAuthMethod := infraAuth.NewPublishApiKeyAuthMethod(af.config)
		af.RegisterAuthMethod(publishApiKeyAuthMethod)
	}

	// AuthenticationTokenAuthMethod is not used as we follow Python's approach
	// of using environment variables for API keys

	// Register AdminSessionAuthMethod (5th priority)
	adminSessionAuthMethod := infraAuth.NewAdminSessionAuthMethod(af.sessionRepo, af.userGroupRepo)
	af.RegisterAuthMethod(adminSessionAuthMethod)

	// Register SamlAuthMethod (3rd priority) - only if SAML is configured
	if af.config.SAML2IDPMetadataURL != "" && af.config.SAML2IssuerEntityID != "" {
		samlConfig := &infraAuth.SAMLConfig{
			EntityID:       af.config.SAML2EntityID,
			SSOURL:         af.config.SAML2IDPMetadataURL,
			SLOURL:         "", // Optional - can be added if needed
			Certificate:    af.config.SAML2PublicKey,
			PrivateKey:     af.config.SAML2PrivateKey,
			GroupAttribute: af.config.SAML2GroupAttribute,
			// Note: Debug field not available in SAMLConfig struct - would need to be added if needed
		}
		samlAuthMethod := infraAuth.NewSamlAuthMethod(samlConfig, af.userGroupRepo)
		af.RegisterAuthMethod(samlAuthMethod)
	}

	// Register OpenIDConnectAuthMethod (4th priority) - only if OIDC is configured
	if af.config.OpenIDConnectClientID != "" && af.config.OpenIDConnectIssuer != "" {
		// Use configured scopes or fall back to default scopes
		scopes := af.config.OpenIDConnectScopes
		if len(scopes) == 0 {
			scopes = []string{"openid", "profile", "email"} // Default scopes
		}

		oidcConfig := &infraAuth.OIDCConfig{
			ClientID:     af.config.OpenIDConnectClientID,
			ClientSecret: af.config.OpenIDConnectClientSecret,
			IssuerURL:    af.config.OpenIDConnectIssuer,
			RedirectURI:  "", // TODO: Add redirect URI to config if needed
			Scopes:       scopes,
			// Note: Debug field not available in OIDCConfig struct - would need to be added if needed
		}
		oidcAuthMethod := infraAuth.NewOpenIDConnectAuthMethod(oidcConfig, af.userGroupRepo)
		af.RegisterAuthMethod(oidcAuthMethod)
	}

	// Register TerraformOidcAuthMethod (9th priority)
	// This provides OIDC authentication for Terraform CLI
	if af.terraformIDPService != nil {
		// Wrap the service in an adapter to bridge the interface gap
		adapter := &TerraformIDPServiceAdapter{
			service: af.terraformIDPService,
		}
		terraformIDP := infraAuth.NewTerraformIDP(
			adapter,
			af.logger,
			true, // Enable by default when service is available
		)
		terraformOidcAuthMethod := infraAuth.NewTerraformOidcAuthMethod(terraformIDP)
		af.RegisterAuthMethod(terraformOidcAuthMethod)
	}

	// Register fallback method
	af.RegisterAuthMethod(NewNotAuthenticatedAuthMethod())
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
	return NewNotAuthenticatedAuthMethod()
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

	// Extract API key from X-Terrareg-ApiKey header (matches Python)
	apiKey := headers["X-Terrareg-ApiKey"]

	// Try each auth method in priority order
	for _, authMethod := range af.authMethods {
		if !authMethod.IsEnabled() {
			continue
		}

		// Check authentication state
		isAuthenticated := false

		// Special handling for API key auth methods
		switch authMethod.GetProviderType() {
		case auth.AuthMethodAdminApiKey:
			if adminAuthMethod, ok := authMethod.(*infraAuth.AdminApiKeyAuthMethod); ok {
				isAuthenticated = adminAuthMethod.CheckAuthStateWithKey(apiKey)
			} else {
				isAuthenticated = authMethod.CheckAuthState()
			}
		case auth.AuthMethodUploadApiKey:
			if uploadAuthMethod, ok := authMethod.(*infraAuth.UploadApiKeyAuthMethod); ok {
				isAuthenticated = uploadAuthMethod.CheckAuthStateWithKey(apiKey)
			} else {
				isAuthenticated = authMethod.CheckAuthState()
			}
		case auth.AuthMethodPublishApiKey:
			if publishAuthMethod, ok := authMethod.(*infraAuth.PublishApiKeyAuthMethod); ok {
				isAuthenticated = publishAuthMethod.CheckAuthStateWithKey(apiKey)
			} else {
				isAuthenticated = authMethod.CheckAuthState()
			}
		default:
			isAuthenticated = authMethod.CheckAuthState()
		}

		if isAuthenticated {
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

// NotAuthenticatedAuthMethod represents the fallback authentication method
type NotAuthenticatedAuthMethod struct {
	auth.BaseAuthMethod
}

// NewNotAuthenticatedAuthMethod creates a new not authenticated method
func NewNotAuthenticatedAuthMethod() *NotAuthenticatedAuthMethod {
	return &NotAuthenticatedAuthMethod{}
}

// GetType returns the authentication method type
func (n *NotAuthenticatedAuthMethod) GetType() string {
	return "NotAuthenticated"
}

// Implement AuthMethod interface methods
func (n *NotAuthenticatedAuthMethod) IsBuiltInAdmin() bool {
	return false
}

func (n *NotAuthenticatedAuthMethod) IsAdmin() bool {
	return false
}

func (n *NotAuthenticatedAuthMethod) IsAuthenticated() bool {
	return false
}

func (n *NotAuthenticatedAuthMethod) IsEnabled() bool {
	return true
}

func (n *NotAuthenticatedAuthMethod) RequiresCSRF() bool {
	return false
}

func (n *NotAuthenticatedAuthMethod) CheckAuthState() bool {
	return true
}

func (n *NotAuthenticatedAuthMethod) CanPublishModuleVersion(namespace string) bool {
	return false
}

func (n *NotAuthenticatedAuthMethod) CanUploadModuleVersion(namespace string) bool {
	return false
}

func (n *NotAuthenticatedAuthMethod) CheckNamespaceAccess(permissionType, namespace string) bool {
	return false
}

func (n *NotAuthenticatedAuthMethod) GetAllNamespacePermissions() map[string]string {
	return make(map[string]string)
}

func (n *NotAuthenticatedAuthMethod) GetUsername() string {
	return ""
}

func (n *NotAuthenticatedAuthMethod) GetUserGroupNames() []string {
	return []string{}
}

func (n *NotAuthenticatedAuthMethod) CanAccessReadAPI() bool {
	return true
}

func (n *NotAuthenticatedAuthMethod) CanAccessTerraformAPI() bool {
	return true
}

func (n *NotAuthenticatedAuthMethod) GetTerraformAuthToken() string {
	return ""
}

func (n *NotAuthenticatedAuthMethod) GetProviderType() auth.AuthMethodType {
	return auth.AuthMethodNotAuthenticated
}

// loadUserGroupsForAuthMethod loads user groups for an authentication method
func (af *AuthFactory) loadUserGroupsForAuthMethod(ctx context.Context, authMethod auth.AuthMethod) ([]*auth.UserGroup, error) {
	// Get username from the auth method if available
	username := authMethod.GetUsername()
	if username == "" {
		// If no username is available, return empty groups list
		// This is the case for API key authentication methods and NotAuthenticated
		return []*auth.UserGroup{}, nil
	}

	// Get user groups for the authenticated user
	userGroups, err := af.userGroupRepo.GetGroupsForUser(ctx, username)
	if err != nil {
		// If user groups can't be loaded, don't fail authentication
		// Return empty groups list instead
		return []*auth.UserGroup{}, nil
	}

	return userGroups, nil
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
