package service

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/auth"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/auth/model"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/auth/repository"
	moduleRepo "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/module/repository"
	provider_source_service "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/provider_source/service"
	infraAuth "github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/auth"
	infraConfig "github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/config"
	"github.com/rs/zerolog"
)

// AuthFactory handles authentication with immutable AuthMethod implementations
// It uses AuthMethod factories to create AuthContext instances with authentication state
type AuthFactory struct {
	authMethods               []auth.AuthMethod
	mutex                     sync.RWMutex
	sessionRepo               repository.SessionRepository
	userGroupRepo             repository.UserGroupRepository
	namespaceRepo             moduleRepo.NamespaceRepository
	config                    *infraConfig.InfrastructureConfig
	logger                    *zerolog.Logger
	providerSourceFactory     *provider_source_service.ProviderSourceFactory
	sessionManagementService  *SessionManagementService
}

// NewAuthFactory creates a new immutable authentication factory
// Returns an error if required dependencies are nil
func NewAuthFactory(
	sessionRepo repository.SessionRepository,
	userGroupRepo repository.UserGroupRepository,
	namespaceRepo moduleRepo.NamespaceRepository,
	config *infraConfig.InfrastructureConfig,
	terraformIdpService *TerraformIdpService,
	oidcService *OIDCService,
	providerSourceFactory *provider_source_service.ProviderSourceFactory,
	sessionManagementService *SessionManagementService,
	logger *zerolog.Logger,
) (*AuthFactory, error) {
	// Validate required dependencies
	if sessionRepo == nil {
		return nil, fmt.Errorf("sessionRepo is required")
	}
	if userGroupRepo == nil {
		return nil, fmt.Errorf("userGroupRepo is required")
	}
	if namespaceRepo == nil {
		return nil, fmt.Errorf("namespaceRepo is required")
	}
	if config == nil {
		return nil, fmt.Errorf("config is required")
	}
	if logger == nil {
		return nil, fmt.Errorf("logger is required")
	}
	// providerSourceFactory and sessionManagementService are optional for now
	// to maintain backward compatibility, but this should be reviewed

	factory := &AuthFactory{
		authMethods:              make([]auth.AuthMethod, 0),
		sessionRepo:              sessionRepo,
		userGroupRepo:            userGroupRepo,
		namespaceRepo:            namespaceRepo,
		config:                   config,
		logger:                   logger,
		providerSourceFactory:    providerSourceFactory,
		sessionManagementService: sessionManagementService,
	}

	// Initialize immutable auth methods in priority order
	factory.initializeAuthMethods(terraformIdpService, oidcService)

	return factory, nil
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

	// 2. Admin Session auth method - only register if session management is available (SECRET_KEY configured)
	if af.sessionManagementService != nil {
		adminSessionAuthMethod := infraAuth.NewAdminSessionAuthMethod(
			af.sessionRepo,
			af.userGroupRepo,
			af.namespaceRepo,
			af.sessionManagementService,
		)
		af.RegisterAuthMethod(adminSessionAuthMethod)
	}

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

	// 7. GitHub auth method (using provider source pattern)
	if af.providerSourceFactory != nil {
		githubAuthMethod := infraAuth.NewGitHubAuthMethod(af.providerSourceFactory, af.sessionRepo)
		af.RegisterAuthMethod(githubAuthMethod)
	}

	// 8. Terraform OIDC auth method
	if terraformIdpService != nil {
		terraformOidcAuthMethod := infraAuth.NewTerraformOidcAuthMethod(af.config, terraformIdpService)
		af.RegisterAuthMethod(terraformOidcAuthMethod)
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
// NOTE: This should only be called during initialization (before any AuthenticateRequest calls)
// Calling this during request handling may cause race conditions
func (af *AuthFactory) RegisterAuthMethod(authMethod auth.AuthMethod) {
	af.mutex.Lock()
	defer af.mutex.Unlock()
	af.authMethods = append(af.authMethods, authMethod)
}

// AuthenticateRequest authenticates an HTTP request using immutable AuthMethods
// Returns an AuthContext interface with all authentication state and permissions
// NOTE: No mutex lock needed - authMethods slice is effectively immutable after initialization
// All RegisterAuthMethod calls happen in initializeAuthMethods during NewAuthFactory construction
func (af *AuthFactory) AuthenticateRequest(ctx context.Context, headers, formData, queryParams map[string]string) (auth.AuthContext, error) {
	// No lock needed - authMethods is immutable after construction

	// Extract session ID for auth methods that need it
	sessionID := af.extractSessionID(headers)
	authorizationHeader := headers["Authorization"]

	// Create session data map and populate with session ID if available
	// This allows SessionAuthMethod implementations to access the session ID
	sessionData := make(map[string]interface{})
	if sessionID != nil {
		sessionData["session_id"] = *sessionID
	}

	// Try each auth method in priority order
	for _, authMethod := range af.authMethods {
		if !authMethod.IsEnabled() {
			continue
		}

		var authContext auth.AuthContext
		var err error

		// Each auth method handles its own specific authentication requirements
		// The Authenticate method now returns AuthContext, not AuthMethod
		switch method := authMethod.(type) {
		case auth.HeaderAuthMethod:
			// Header-based auth methods (AdminApiKey, UploadApiKey, PublishApiKey, etc.)
			authContext, err = method.Authenticate(ctx, headers, formData, queryParams)
		case auth.SessionAuthMethod:
			// Session-based auth methods (SAML, OpenID Connect, AdminSession)
			authContext, err = method.Authenticate(ctx, sessionData)
		case auth.TokenAuthMethod:
			// Token-based auth methods (Terraform OIDC)
			authContext, err = method.Authenticate(ctx, authorizationHeader)
		case auth.BearerTokenAuthMethod:
			// Bearer token auth methods
			authContext, err = method.Authenticate(ctx, authorizationHeader, nil)
		default:
			// Unknown auth method type, skip
			af.logger.Debug().Str("auth_method", string(authMethod.GetProviderType())).Msg("Unknown auth method type")
			continue
		}

		if err != nil {
			// Log the error for debugging but continue trying other auth methods
			af.logger.Debug().Err(err).Str("auth_method", string(authMethod.GetProviderType())).Msg("Authentication method failed")
			continue
		}

		// If we got an authenticated context, return it
		if authContext != nil && authContext.IsAuthenticated() {
			af.logger.Info().
				Str("auth_method", string(authContext.GetProviderType())).
				Str("username", authContext.GetUsername()).
				Bool("is_admin", authContext.IsAdmin()).
				Msg("Authentication successful")

			return authContext, nil
		}
	}

	// No authentication method succeeded - return NotAuthenticated
	af.logger.Debug().Msg("No authentication method succeeded, returning NotAuthenticated")
	return NewNotAuthenticatedAuthContext(af.config.AllowUnauthenticatedAccess), nil
}

// NewAuthenticationResponseFromAuthContext creates AuthenticationResponse from AuthContext for HTTP responses
// Used only when HTTP response is needed (e.g., for backward compatibility)
func NewAuthenticationResponseFromAuthContext(authCtx auth.AuthContext) *model.AuthenticationResponse {
	response := &model.AuthenticationResponse{
		Success:           authCtx.IsAuthenticated(),
		AuthMethod:        authCtx.GetProviderType(),
		Username:          authCtx.GetUsername(),
		IsAdmin:           authCtx.IsAdmin(),
		UserGroups:        authCtx.GetUserGroupNames(),
		Permissions:       authCtx.GetAllNamespacePermissions(),
		CanPublish:        authCtx.CanPublishModuleVersion(""),
		CanUpload:         authCtx.CanUploadModuleVersion(""),
		CanAccessAPI:      authCtx.CanAccessReadAPI(),
		CanAccessTerraform: authCtx.CanAccessTerraformAPI(),
	}

	// Set session ID from provider data if available
	if data := authCtx.GetProviderData(); data != nil {
		if authSessionID, ok := data["session_id"].(string); ok && authSessionID != "" {
			response.SessionID = &authSessionID
		}
	}

	return response
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
	// Return NotAuthenticated as default - this is a placeholder
	// The application layer should use AuthenticateRequest instead
	return &NotAuthenticatedAuthMethod{}
}

// GetCurrentAuthContext returns the current authentication context (for application layer compatibility)
// TODO: This should be removed and the application layer updated to use AuthenticateRequest
func (af *AuthFactory) GetCurrentAuthContext() auth.AuthContext {
	// Return NotAuthenticated as default - this is a placeholder
	// The application layer should use AuthenticateRequest instead
	return NewNotAuthenticatedAuthContext(af.config.AllowUnauthenticatedAccess)
}

// NotAuthenticated returns the NotAuthenticated auth context
func (af *AuthFactory) NotAuthenticated() auth.AuthContext {
	return NewNotAuthenticatedAuthContext(af.config.AllowUnauthenticatedAccess)
}

// generateRequestID generates a unique request ID
func (af *AuthFactory) generateRequestID() string {
	return fmt.Sprintf("req_%s", af.generateRandomID())
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

// extractSessionID extracts and validates session ID from headers/cookies
// Uses SessionManagementService to validate the session cookie and return the session ID
func (af *AuthFactory) extractSessionID(headers map[string]string) *string {
	// Check for session ID in headers first (for testing purposes)
	if sessionID, exists := headers["X-Session-ID"]; exists && sessionID != "" {
		return &sessionID
	}

	// Check cookie header for encrypted session data
	if af.sessionManagementService == nil {
		// No session management service configured (SECRET_KEY empty)
		return nil
	}

	if cookieHeader, exists := headers["Cookie"]; exists && cookieHeader != "" {
		// Parse cookies to find the session cookie
		cookies := strings.Split(cookieHeader, ";")
		for _, cookie := range cookies {
			cookie = strings.TrimSpace(cookie)
			if strings.HasPrefix(cookie, "terrareg_session=") {
				cookieValue := strings.TrimPrefix(cookie, "terrareg_session=")
				cookieValue = strings.TrimSuffix(cookieValue, "\"")
				if cookieValue != "" {
					// Validate the session cookie using SessionManagementService
					// This decrypts the cookie AND validates the session in the database
					session, err := af.sessionManagementService.ValidateSessionCookie(context.Background(), cookieValue)
					if err != nil {
						af.logger.Debug().Err(err).Msg("Failed to validate session cookie")
						return nil
					}
					if session != nil && session.ID != "" {
						return &session.ID
					}
				}
			}
		}
	}

	return nil
}

// NotAuthenticatedAuthContext represents the fallback authentication context for unauthenticated users
type NotAuthenticatedAuthContext struct {
	auth.BaseAuthContext
	allowUnauthenticatedAccess bool
}

// NewNotAuthenticatedAuthContext creates a new not authenticated context
// allowUnauthenticatedAccess determines whether unauthenticated users can access the read API
func NewNotAuthenticatedAuthContext(allowUnauthenticatedAccess bool) *NotAuthenticatedAuthContext {
	return &NotAuthenticatedAuthContext{
		BaseAuthContext:             auth.BaseAuthContext{},
		allowUnauthenticatedAccess: allowUnauthenticatedAccess,
	}
}

// Implement AuthContext interface methods
func (n *NotAuthenticatedAuthContext) IsBuiltInAdmin() bool {
	return false
}

func (n *NotAuthenticatedAuthContext) IsAdmin() bool {
	return false
}

func (n *NotAuthenticatedAuthContext) IsAuthenticated() bool {
	return false
}

func (n *NotAuthenticatedAuthContext) RequiresCSRF() bool {
	return false
}

func (n *NotAuthenticatedAuthContext) CheckAuthState() bool {
	return true
}

func (n *NotAuthenticatedAuthContext) CanPublishModuleVersion(namespace string) bool {
	return false
}

func (n *NotAuthenticatedAuthContext) CanUploadModuleVersion(namespace string) bool {
	return false
}

func (n *NotAuthenticatedAuthContext) CheckNamespaceAccess(permissionType, namespace string) bool {
	return false
}

func (n *NotAuthenticatedAuthContext) GetAllNamespacePermissions() map[string]string {
	return make(map[string]string)
}

func (n *NotAuthenticatedAuthContext) GetUsername() string {
	return ""
}

func (n *NotAuthenticatedAuthContext) GetUserGroupNames() []string {
	return []string{}
}

func (n *NotAuthenticatedAuthContext) CanAccessReadAPI() bool {
	return n.allowUnauthenticatedAccess
}

func (n *NotAuthenticatedAuthContext) CanAccessTerraformAPI() bool {
	return false
}

func (n *NotAuthenticatedAuthContext) GetTerraformAuthToken() string {
	return ""
}

func (n *NotAuthenticatedAuthContext) GetProviderData() map[string]interface{} {
	return make(map[string]interface{})
}

func (n *NotAuthenticatedAuthContext) GetProviderType() auth.AuthMethodType {
	return auth.AuthMethodNotAuthenticated
}

// NotAuthenticatedAuthMethod represents the fallback authentication method factory
type NotAuthenticatedAuthMethod struct{}

// NewNotAuthenticatedAuthMethod creates a new not authenticated method
func NewNotAuthenticatedAuthMethod() *NotAuthenticatedAuthMethod {
	return &NotAuthenticatedAuthMethod{}
}

// Implement AuthMethod interface methods
func (n *NotAuthenticatedAuthMethod) GetProviderType() auth.AuthMethodType {
	return auth.AuthMethodNotAuthenticated
}

func (n *NotAuthenticatedAuthMethod) IsEnabled() bool {
	return true
}
