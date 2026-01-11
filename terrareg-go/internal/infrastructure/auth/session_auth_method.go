package auth

import (
	"context"
	"net/http"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/auth"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/auth/repository"
)

// SessionAuthMethodConfig holds configuration for session authentication
type SessionAuthMethodConfig struct {
	SessionCookieName string
	SessionSecure     bool
	SessionHTTPOnly   bool
	SessionSameSite   string
	SessionMaxAge     int
}

// SessionAuthMethod implements immutable session-based authentication
type SessionAuthMethod struct {
	auth.BaseAuthMethod
	config        *SessionAuthMethodConfig
	sessionRepo   repository.SessionRepository
	userGroupRepo repository.UserGroupRepository
	httpClient    *http.Client
}

// NewSessionAuthMethod creates a new immutable session auth method
func NewSessionAuthMethod(
	sessionRepo repository.SessionRepository,
	userGroupRepo repository.UserGroupRepository,
	config *SessionAuthMethodConfig,
) *SessionAuthMethod {
	return &SessionAuthMethod{
		config:        config,
		sessionRepo:   sessionRepo,
		userGroupRepo: userGroupRepo,
		httpClient:    &http.Client{},
	}
}

// GetProviderType returns the provider type
func (a *SessionAuthMethod) GetProviderType() auth.AuthMethodType {
	return auth.AuthMethodAdminSession
}

// IsEnabled returns whether this auth method is enabled
func (a *SessionAuthMethod) IsEnabled() bool {
	return true
}

// Authenticate validates session and returns a SessionAuthContext with authentication state
func (a *SessionAuthMethod) Authenticate(ctx context.Context, sessionID string) (auth.AuthMethod, error) {
	if sessionID == "" {
		// No session ID provided, return nil to let other auth methods try
		return nil, nil
	}

	// Find session in repository
	session, err := a.sessionRepo.FindByID(ctx, sessionID)
	if err != nil || session == nil {
		// Session not found, return nil to let other auth methods try
		return nil, nil
	}

	// Check if session is expired
	if session.IsExpired() {
		// Session expired, return nil to let other auth methods try
		return nil, nil
	}

	// Create SessionAuthContext with session state
	// For now, use session ID as username since the session doesn't have user details
	// TODO: Update this when user session details are available
	authContext := auth.NewSessionAuthContext(ctx, 0, sessionID, "", sessionID)

	// TODO: Load user groups and permissions when repository methods are available
	// For now, just return the basic session context

	return authContext, nil
}

// Default AuthMethod implementations (returning false/empty values)
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
