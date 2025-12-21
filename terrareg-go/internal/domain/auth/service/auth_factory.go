package service

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
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
	factory.initializeAuthMethods()

	return factory
}

// initializeAuthMethods sets up immutable authentication methods in priority order
func (af *AuthFactory) initializeAuthMethods() {
	// Priority order from Python terrareg
	// 1. ImmutableApiKeyAuthMethod (consolidated admin/upload/publish API keys)
	// 2. ImmutableSessionAuthMethod
	// 3. Other auth methods...

	// Register consolidated API key auth method
	apiKeyAuthMethod := infraAuth.NewApiKeyAuthMethod(af.config)
	if apiKeyAuthMethod.IsEnabled() {
		af.RegisterAuthMethod(apiKeyAuthMethod)
	}

	// Register session auth method
	sessionAuthConfig := &infraAuth.SessionAuthMethodConfig{
		SessionCookieName: "terrareg_session",
		SessionSecure:     true,
		SessionHTTPOnly:   true,
		SessionSameSite:   "Lax",
		SessionMaxAge:     86400, // 24 hours
	}
	sessionAuthMethod := infraAuth.NewSessionAuthMethod(
		af.sessionRepo,
		af.userGroupRepo,
		sessionAuthConfig,
	)
	af.RegisterAuthMethod(sessionAuthMethod)

	// Register SAML auth method
	samlAuthMethod := infraAuth.NewSamlAuthMethod(af.config)
	if samlAuthMethod.IsEnabled() {
		af.RegisterAuthMethod(samlAuthMethod)
	}

	// Register OpenID Connect auth method
	openidAuthMethod := infraAuth.NewOpenidConnectAuthMethod(af.config)
	if openidAuthMethod.IsEnabled() {
		af.RegisterAuthMethod(openidAuthMethod)
	}

	// Register Admin Session auth method
	adminSessionAuthMethod := infraAuth.NewAdminSessionAuthMethod(
		af.sessionRepo,
		af.userGroupRepo,
	)
	af.RegisterAuthMethod(adminSessionAuthMethod)

	// Register Terraform OIDC auth method
	terraformOidcAuthMethod := infraAuth.NewTerraformOidcAuthMethod(af.config)
	if terraformOidcAuthMethod.IsEnabled() {
		af.RegisterAuthMethod(terraformOidcAuthMethod)
	}

	// Register Terraform Analytics Auth Key auth method
	terraformAnalyticsAuthMethod := infraAuth.NewTerraformAnalyticsAuthKeyAuthMethod()
	af.RegisterAuthMethod(terraformAnalyticsAuthMethod)

	// Register Terraform Internal Extraction auth method
	terraformInternalExtractionAuthMethod := infraAuth.NewTerraformInternalExtractionAuthMethod("terraform-internal-extraction")
	af.RegisterAuthMethod(terraformInternalExtractionAuthMethod)
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

	// Extract API key from X-Terrareg-ApiKey header (matches Python)
	apiKey := headers["X-Terrareg-ApiKey"]

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
		case *infraAuth.ApiKeyAuthMethod:
			authenticatedAdapter, err = method.Authenticate(ctx, apiKey)
		case *infraAuth.SessionAuthMethod:
			authenticatedAdapter, err = method.Authenticate(ctx, sessionID)
		case *infraAuth.SamlAuthMethod:
			authenticatedAdapter, err = method.Authenticate(ctx, sessionData)
		case *infraAuth.OpenidConnectAuthMethod:
			authenticatedAdapter, err = method.Authenticate(ctx, sessionData)
		case *infraAuth.AdminSessionAuthMethod:
			authenticatedAdapter, err = method.Authenticate(ctx, headers, formData, queryParams)
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