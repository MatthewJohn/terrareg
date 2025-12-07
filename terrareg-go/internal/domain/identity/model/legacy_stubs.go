package model

import (
	"context"
	"errors"
	"strconv"
	"time"
)

// Temporary stub implementations for legacy user-based services
// These should be removed as part of the migration to group-based auth

// User represents a user (legacy - should be removed)
// TODO: Remove this as part of migration to group-based authentication
type User struct {
	id             int        `json:"id"`
	Username       string     `json:"username"`
	Email          string     `json:"email"`
	Active         bool       `json:"active"`
	ExternalID     string     `json:"external_id"`
	AuthProviderID AuthMethod `json:"auth_provider_id"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

// User methods for legacy service compatibility
func (u *User) ID() string {
	return strconv.Itoa(u.id)
}

func (u *User) IsActive() bool {
	return u.Active
}

func (u *User) Authenticate(accessToken, refreshToken string, expiry *time.Time) error {
	if !u.Active {
		return ErrUserInactive
	}
	// Basic auth logic for legacy compatibility
	return nil
}

// GetAuthMethod returns the authentication method for this user
func (u *User) GetAuthMethod() AuthMethod {
	return u.AuthProviderID
}

// SetExternalID sets the external ID for the user
func (u *User) SetExternalID(externalID string) {
	u.ExternalID = externalID
	u.UpdatedAt = time.Now()
}

// SetAuthProviderID sets the authentication provider ID for the user
func (u *User) SetAuthProviderID(authProviderID AuthMethod) {
	u.AuthProviderID = authProviderID
	u.UpdatedAt = time.Now()
}

// SetActive sets the active status for the user
func (u *User) SetActive(active bool) {
	u.Active = active
	u.UpdatedAt = time.Now()
}

// AuthMethod represents authentication methods (legacy)
// TODO: Remove this as part of migration to auth.AuthMethodType
type AuthMethod string

// Legacy AuthMethod constants
const (
	AuthMethodNotAuthenticated                AuthMethod = "NOT_AUTHENTICATED"
	AuthMethodAdminSession                    AuthMethod = "ADMIN_SESSION"
	AuthMethodAdminApiKey                     AuthMethod = "ADMIN_API_KEY"
	AuthMethodSAML                            AuthMethod = "SAML"
	AuthMethodOpenIDConnect                   AuthMethod = "OPENID_CONNECT"
	AuthMethodGitHub                          AuthMethod = "GITHUB"
	AuthMethodTerraformOIDC                   AuthMethod = "TERRAFORM_OIDC"
	AuthMethodTerraformAnalyticsAuthKey       AuthMethod = "TERRAFORM_ANALYTICS_AUTH_KEY"
	AuthMethodTerraformIgnoreAnalyticsAuthKey AuthMethod = "TERRAFORM_IGNORE_ANALYTICS_AUTH_KEY"
	AuthMethodTerraformInternalExtraction     AuthMethod = "TERRAFORM_INTERNAL_EXTRACTION"
	AuthMethodUploadApiKey                    AuthMethod = "UPLOAD_API_KEY"
	AuthMethodPublishApiKey                   AuthMethod = "PUBLISH_API_KEY"
	AuthMethodAPIKey                          AuthMethod = "API_KEY" // Generic API key type
)

// Permission represents permissions (legacy)
// TODO: Remove this as part of migration to auth.PermissionType
type Permission struct {
	ID           int       `json:"id"`
	UserID       int       `json:"user_id"`
	ResourceType string    `json:"resource_type"`
	ResourceID   string    `json:"resource_id"`
	Action       string    `json:"action"`
	CreatedAt    time.Time `json:"created_at"`
}

// AuthResult represents authentication result (legacy)
type AuthResult struct {
	User      *User                  `json:"user"`
	Token     string                 `json:"token"`
	ExpiresAt time.Time              `json:"expires_at"`
	Metadata  map[string]interface{} `json:"metadata"`
}

// AuthProvider represents auth provider (legacy)
type AuthProvider interface {
	Authenticate(ctx context.Context, token string) (*AuthResult, error)
	GetName() string
	GetType() AuthMethod
}

// Session represents a user session (legacy - should be removed)
// TODO: Remove this as part of migration to auth domain sessions
type Session struct {
	ID        string     `json:"id"`
	UserID    string     `json:"user_id"`
	Token     string     `json:"token"`
	ExpiresAt time.Time  `json:"expires_at"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt *time.Time `json:"updated_at,omitempty"`
}

// UserGroup represents user group (legacy)
// TODO: Remove this as part of migration to auth domain user groups
type UserGroup struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ResourceType represents resource types (legacy)
type ResourceType string

// Action represents actions (legacy)
type Action string

// IDPAccessToken represents IDP access token (legacy)
type IDPAccessToken struct {
	ID        int       `json:"id"`
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
}

// Legacy error constants (should be replaced with auth domain errors)
var (
	ErrUserNotFound         = errors.New("user not found")
	ErrUserInactive         = errors.New("user is inactive")
	ErrInvalidAPIKey        = errors.New("invalid API key")
	ErrAPIKeyNotFound       = errors.New("API key not found")
	ErrInvalidAuthMethod    = errors.New("invalid authentication method")
	ErrInvalidCredentials   = errors.New("invalid credentials")
	ErrAuthenticationFailed = errors.New("authentication failed")
	ErrNotFound             = errors.New("resource not found")
	ErrSessionInvalid       = errors.New("session is invalid")
	ErrSessionNotFound      = errors.New("session not found")
	ErrSessionExpired       = errors.New("session has expired")
	ErrUserAlreadyExists    = errors.New("user already exists")
)

// NewUser creates a new user instance
func NewUser(username, email string, externalID string, authMethod AuthMethod) *User {
	return &User{
		id:             0, // Will be set by repository
		Username:       username,
		Email:          email,
		Active:         true,
		ExternalID:     externalID,
		AuthProviderID: authMethod,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
}

// NewSession creates a new session instance
func NewSession(id, userID, token string, expiresAt time.Time) *Session {
	return &Session{
		ID:        id,
		UserID:    userID,
		Token:     token,
		ExpiresAt: expiresAt,
		CreatedAt: time.Now(),
	}
}
