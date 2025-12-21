package service

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/auth"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/auth/model"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/auth/repository"
	infraAuth "github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/auth"
	infraConfig "github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/config"
	"github.com/rs/zerolog"
)

// AuthFactory handles authentication with immutable AuthMethod implementations
// It creates adapters that hold authentication state without modifying the AuthMethods themselves
type AuthFactory struct {
	authMethods []auth.AuthMethod
	mutex       sync.RWMutex
	sessionRepo repository.SessionRepository
	userGroupRepo repository.UserGroupRepository
	config      *infraConfig.InfrastructureConfig
	logger      *zerolog.Logger
}

// NewAuthFactory creates a new immutable authentication factory
func NewAuthFactory(
	sessionRepo repository.SessionRepository,
	userGroupRepo repository.UserGroupRepository,
	config *infraConfig.InfrastructureConfig,
	terraformIdpService *TerraformIdpService,
	oidcService *OIDCService,
	logger *zerolog.Logger,
) *AuthFactory {
	factory := &AuthFactory{
		authMethods:   make([]auth.AuthMethod, 0),
		sessionRepo:   sessionRepo,
		userGroupRepo: userGroupRepo,
		config:        config,
		logger:        logger,
	}

	// Initialize immutable auth methods in priority order
	factory.initializeAuthMethods(terraformIdpService, oidcService)

	return factory
}

// initializeAuthMethods sets up immutable authentication methods in priority order
// Priority order matches Python terrareg:
// 1. AdminApiKey
// 2. AdminSession
// 3. UploadApiKey
// 4. PublishApiKey
// 5. SAML
// 6. OpenID Connect
// 7. GitHub (not implemented yet)
// 8. Terraform OIDC
// 9. Terraform Analytics
// 10. Terraform Internal Extraction
// 11. NotAuthenticated (fallback)
func (af *AuthFactory) initializeAuthMethods(terraformIdpService *TerraformIdpService, oidcService *OIDCService) {
	// 1. Admin API Key auth method
	adminApiKeyAuthMethod := infraAuth.NewAdminApiKeyAuthMethod(af.config)
	if adminApiKeyAuthMethod.IsEnabled() {
		af.RegisterAuthMethod(adminApiKeyAuthMethod)
	}

	// 2. Admin Session auth method
	adminSessionAuthMethod := infraAuth.NewAdminSessionAuthMethod(
		af.sessionRepo,
		af.userGroupRepo,
	)
	af.RegisterAuthMethod(adminSessionAuthMethod)

	// 3. Upload API Key auth method
	uploadApiKeyAuthMethod := infraAuth.NewUploadApiKeyAuthMethod(af.config)
	if uploadApiKeyAuthMethod.IsEnabled() {
		af.RegisterAuthMethod(uploadApiKeyAuthMethod)
	}

	// 4. Publish API Key auth method
	publishApiKeyAuthMethod := infraAuth.NewPublishApiKeyAuthMethod(af.config)
	if publishApiKeyAuthMethod.IsEnabled() {
		af.RegisterAuthMethod(publishApiKeyAuthMethod)
	}

	// 5. SAML auth method
	samlAuthMethod := infraAuth.NewSamlAuthMethod(af.config)
	if samlAuthMethod.IsEnabled() {
		af.RegisterAuthMethod(samlAuthMethod)
	}

	// 6. OpenID Connect auth method
	if oidcService != nil {
		// For now, skip OpenID Connect until we fix the interface issue
		// TODO: Fix OpenID Connect auth method interface compatibility
	}

	// 7. GitHub auth method (TODO: implement)

	// 8. Terraform OIDC auth method
	if terraformIdpService != nil {
		// For now, skip Terraform OIDC until we fix the interface issue
		// TODO: Fix Terraform OIDC auth method interface compatibility
	}

	// 9. Terraform Analytics auth key method
	analyticsAuthKeyAuthMethod := infraAuth.NewTerraformAnalyticsAuthKeyAuthMethod(af.config)
	if analyticsAuthKeyAuthMethod.IsEnabled() {
		af.RegisterAuthMethod(analyticsAuthKeyAuthMethod)
	}

	// 10. Terraform Internal Extraction auth method
	internalExtractionAuthMethod := infraAuth.NewTerraformInternalExtractionAuthMethod("terraform-internal", af.config)
	if internalExtractionAuthMethod.IsEnabled() {
		af.RegisterAuthMethod(internalExtractionAuthMethod)
	}
}

// RegisterAuthMethod registers an authentication method
func (af *AuthFactory) RegisterAuthMethod(authMethod auth.AuthMethod) {
	af.mutex.Lock()
	defer af.mutex.Unlock()
	af.authMethods = append(af.authMethods, authMethod)
}

// AuthenticateRequest authenticates an HTTP request using immutable AuthMethods
func (af *AuthFactory) AuthenticateRequest(ctx context.Context, headers, formData, queryParams map[string]string) (*model.AuthenticationResponse, error) {
	af.mutex.Lock()
	defer af.mutex.Unlock()

	// API key extraction is now handled by individual auth methods

	// Extract session ID from cookies
	sessionID := af.extractSessionID(headers)

	// Extract session data for SAML/OpenID (in a real implementation, this would come from the session store)
	sessionData := make(map[string]interface{})
	if sessionID != nil {
		// TODO: Load session data from session repository
		// For now, assume session data is extracted from headers/cookies
		sessionData["samlNameId"] = headers["X-SAML-NameID"]
		sessionData["samlUserdata"] = nil // Would be loaded from session
		sessionData["openid_connect_expires_at"] = 0.0
		sessionData["openid_connect_id_token"] = ""
		sessionData["openid_username"] = headers["X-OpenID-Username"]
		sessionData["openid_groups"] = []string{} // Would be loaded from session
	}

	// Extract authorization header for Terraform OIDC
	authorizationHeader := headers["Authorization"]

	// Extract request body for Terraform OIDC (if needed)
	var requestBody []byte
	// In a real implementation, this would be extracted from the request

	// Try each auth method in priority order
	for _, authMethod := range af.authMethods {
		if !authMethod.IsEnabled() {
			continue
		}

		var authenticatedAdapter auth.AuthMethod
		var err error

		// Handle different authentication methods
		switch method := authMethod.(type) {
		case *infraAuth.AdminApiKeyAuthMethod:
			authenticatedAdapter, err = method.Authenticate(ctx, headers, formData, queryParams)
		case *infraAuth.AdminSessionAuthMethod:
			authenticatedAdapter, err = method.Authenticate(ctx, headers, formData, queryParams)
		case *infraAuth.UploadApiKeyAuthMethod:
			authenticatedAdapter, err = method.Authenticate(ctx, headers, formData, queryParams)
		case *infraAuth.PublishApiKeyAuthMethod:
			authenticatedAdapter, err = method.Authenticate(ctx, headers, formData, queryParams)
		case *infraAuth.SamlAuthMethod:
			authenticatedAdapter, err = method.Authenticate(ctx, sessionData)
		case *infraAuth.OpenidConnectAuthMethod:
			authenticatedAdapter, err = method.Authenticate(ctx, sessionData)
		case *infraAuth.TerraformOidcAuthMethod:
			authenticatedAdapter, err = method.Authenticate(ctx, authorizationHeader, requestBody)
		case *infraAuth.TerraformAnalyticsAuthKeyAuthMethod:
			authenticatedAdapter, err = method.Authenticate(ctx, headers, formData, queryParams)
		case *infraAuth.TerraformInternalExtractionAuthMethod:
			authenticatedAdapter, err = method.Authenticate(ctx, headers, formData, queryParams)
		default:
			// For legacy auth methods, fall back to CheckAuthState
			if authMethod.CheckAuthState() {
				authenticatedAdapter = authMethod
			}
		}

		if err != nil {
			continue // Try next auth method
		}

		// If we got an authenticated adapter, build response
		if authenticatedAdapter != nil && authenticatedAdapter.IsAuthenticated() {
			// Build response using the adapter
			response := model.NewAuthenticationResponse(
				af.generateRequestID(),
				true,
				authenticatedAdapter.GetProviderType(),
			)
			response.AuthMethod = authenticatedAdapter.GetProviderType()
			response.Username = authenticatedAdapter.GetUsername()
			response.IsAdmin = authenticatedAdapter.IsAdmin()
			response.UserGroups = authenticatedAdapter.GetUserGroupNames()
			response.Permissions = authenticatedAdapter.GetAllNamespacePermissions()
			response.CanPublish = authenticatedAdapter.CanPublishModuleVersion("")
			response.CanUpload = authenticatedAdapter.CanUploadModuleVersion("")
			response.CanAccessAPI = authenticatedAdapter.CanAccessReadAPI()
			response.CanAccessTerraform = authenticatedAdapter.CanAccessTerraformAPI()
			response.TerraformToken = af.getStringPtr(authenticatedAdapter.GetTerraformAuthToken())

			// Set session ID if available
			if adapter, ok := authenticatedAdapter.(*model.SessionStateAdapter); ok {
				response.SessionID = adapter.GetContext().Value("session_id").(*string)
			} else if sessionID := af.extractSessionID(headers); sessionID != nil {
				response.SessionID = sessionID
			}

			return response, nil
		}
	}

	// No authentication method succeeded - return NotAuthenticated
	fallbackMethod := &NotAuthenticatedAuthMethod{}
	response := model.NewAuthenticationResponse(
		af.generateRequestID(),
		false,
		auth.AuthMethodNotAuthenticated,
	)
	response.AuthMethod = auth.AuthMethodNotAuthenticated
	response.Username = fallbackMethod.GetUsername()
	response.CanAccessAPI = fallbackMethod.CanAccessReadAPI()

	return response, nil
}

// CheckNamespacePermission checks if the current auth context has permission for a namespace
func (af *AuthFactory) CheckNamespacePermission(namespace, permissionType string) bool {
	// This method would need to be called with the current AuthContext
	// For now, return false as we don't store the current context
	return false
}

// CanPublishModuleVersion checks if current user can publish to a namespace
func (af *AuthFactory) CanPublishModuleVersion(namespace string) bool {
	// This method would need to be called with the current AuthContext
	return false
}

// CanUploadModuleVersion checks if current user can upload to a namespace
func (af *AuthFactory) CanUploadModuleVersion(namespace string) bool {
	// This method would need to be called with the current AuthContext
	return false
}

// CanAccessReadAPI checks if current user can access the read API
func (af *AuthFactory) CanAccessReadAPI() bool {
	// This method would need to be called with the current AuthContext
	return false
}

// CanAccessTerraformAPI checks if current user can access Terraform API
func (af *AuthFactory) CanAccessTerraformAPI() bool {
	// This method would need to be called with the current AuthContext
	return false
}

// InvalidateAuthentication clears the current authentication state
func (af *AuthFactory) InvalidateAuthentication() {
	// With immutable methods, there's no state to invalidate
	// The authentication state is managed per-request
}

// GetCurrentAuthMethod returns the current authentication method (for application layer compatibility)
// TODO: This should be removed and the application layer updated to use AuthenticateRequest
func (af *AuthFactory) GetCurrentAuthMethod() auth.AuthMethod {
	// Return NotAuthenticated as default
	return af.NotAuthenticated()
}

// GetCurrentAuthContext returns the current authentication context (for application layer compatibility)
// TODO: This should be removed and the application layer updated to use AuthenticateRequest
func (af *AuthFactory) GetCurrentAuthContext() *AuthContextAdapter {
	// Return an adapter that provides the expected interface
	return &AuthContextAdapter{
		authMethod: af.GetCurrentAuthMethod(),
	}
}

// NotAuthenticated returns the NotAuthenticated auth method
func (af *AuthFactory) NotAuthenticated() auth.AuthMethod {
	return &NotAuthenticatedAuthMethod{}
}

// generateRequestID generates a unique request ID
func (af *AuthFactory) generateRequestID() string {
	return fmt.Sprintf("req_%d", af.generateRandomID())
}

// generateRandomID generates a random ID (simplified)
func (af *AuthFactory) generateRandomID() string {
	// In a real implementation, use crypto/rand or similar
	return fmt.Sprintf("%d", 0) // Placeholder
}

// getStringPtr returns a pointer to the string, or nil if empty
func (af *AuthFactory) getStringPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// extractSessionID extracts session ID from headers/cookies
func (af *AuthFactory) extractSessionID(headers map[string]string) *string {
	// Check for session ID in headers first
	if sessionID, exists := headers["X-Session-ID"]; exists && sessionID != "" {
		return &sessionID
	}

	// Check cookie header for session ID
	if cookieHeader, exists := headers["Cookie"]; exists && cookieHeader != "" {
		// Parse cookies
		cookies := strings.Split(cookieHeader, ";")
		for _, cookie := range cookies {
			cookie = strings.TrimSpace(cookie)
			if strings.HasPrefix(cookie, "terrareg_session=") {
				sessionID := strings.TrimPrefix(cookie, "terrareg_session=")
				sessionID = strings.TrimSuffix(sessionID, "\"")
				if sessionID != "" {
					return &sessionID
				}
			}
		}
	}

	return nil
}

// AuthContextAdapter provides the interface expected by the application layer
type AuthContextAdapter struct {
	authMethod auth.AuthMethod
}

// Permissions returns permissions (stub for application layer compatibility)
func (a *AuthContextAdapter) Permissions() map[string]string {
	return a.authMethod.GetAllNamespacePermissions()
}

// SessionID returns session ID (stub for application layer compatibility)
func (a *AuthContextAdapter) SessionID() *string {
	// Extract session ID from provider data if available
	if data := a.authMethod.GetProviderData(); data != nil {
		if sessionID, ok := data["session_id"].(string); ok {
			return &sessionID
		}
	}
	return nil
}

// Username returns username
func (a *AuthContextAdapter) Username() string {
	return a.authMethod.GetUsername()
}

// HasPermission checks if the context has a specific permission (stub)
func (a *AuthContextAdapter) HasPermission(permission string) bool {
	// For now, return false - this would need to be implemented in auth contexts
	return false
}

// NotAuthenticatedAuthMethod represents the fallback authentication method
type NotAuthenticatedAuthMethod struct {
	auth.BaseAuthMethod
}

// NewNotAuthenticatedAuthMethod creates a new not authenticated method
func NewNotAuthenticatedAuthMethod() *NotAuthenticatedAuthMethod {
	return &NotAuthenticatedAuthMethod{}
}

// Implement AuthMethod interface methods
func (n *NotAuthenticatedAuthMethod) GetType() string {
	return "NotAuthenticated"
}

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
	return false
}

func (n *NotAuthenticatedAuthMethod) CanAccessTerraformAPI() bool {
	return false
}

func (n *NotAuthenticatedAuthMethod) GetTerraformAuthToken() string {
	return ""
}

func (n *NotAuthenticatedAuthMethod) GetProviderData() map[string]interface{} {
	return make(map[string]interface{})
}

func (n *NotAuthenticatedAuthMethod) GetProviderType() auth.AuthMethodType {
	return auth.AuthMethodNotAuthenticated
}