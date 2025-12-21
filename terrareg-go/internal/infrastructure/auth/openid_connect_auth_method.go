package auth

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/auth"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/config"
)

// OpenidConnectAuthMethod implements immutable OpenID Connect authentication
type OpenidConnectAuthMethod struct {
	config     *config.InfrastructureConfig
	oidcService auth.OIDCValidator
}

// NewOpenidConnectAuthMethod creates a new immutable OpenID Connect auth method
func NewOpenidConnectAuthMethod(config *config.InfrastructureConfig, oidcService auth.OIDCValidator) *OpenidConnectAuthMethod {
	return &OpenidConnectAuthMethod{
		config:     config,
		oidcService: oidcService,
	}
}

// Authenticate validates OpenID Connect session and returns an OpenidConnectAuthContext
func (o *OpenidConnectAuthMethod) Authenticate(ctx context.Context, sessionData map[string]interface{}) (auth.AuthMethod, error) {
	// Check if OpenID Connect is enabled
	if !o.IsEnabled() || o.oidcService == nil {
		return nil, nil // Let other auth methods try
	}

	// Check session expiry
	expiresAtFloat, hasExpiry := sessionData["openid_connect_expires_at"]
	if !hasExpiry {
		return nil, nil // Let other auth methods try
	}

	expiresAt, ok := expiresAtFloat.(float64)
	if !ok {
		return nil, nil // Let other auth methods try
	}

	// Check if session has expired
	if time.Now().After(time.Unix(int64(expiresAt), 0)) {
		return nil, nil // Let other auth methods try
	}

	// Validate ID token if present
	idTokenInterface, hasToken := sessionData["openid_connect_id_token"]
	if !hasToken {
		return nil, nil // Let other auth methods try
	}

	idToken, ok := idTokenInterface.(string)
	if !ok || idToken == "" {
		return nil, nil // Let other auth methods try
	}

	// Validate ID token signature using OIDCService
	userInfo, err := o.oidcService.VerifyIDToken(ctx, idToken)
	if err != nil {
		return nil, fmt.Errorf("ID token validation failed: %w", err)
	}

	// Extract username from validated user info
	username := userInfo.Username
	if username == "" {
		if userInfo.Name != "" {
			username = userInfo.Name
		} else if userInfo.Email != "" {
			username = userInfo.Email
		} else {
			username = userInfo.Sub
		}
	}

	// Create OpenidConnectAuthContext with authentication state
	authContext := auth.NewOpenidConnectAuthContext(ctx, userInfo.Sub, userInfo.RawClaims)

	// Extract user details from claims (username, email, name, etc.)
	authContext.ExtractUserDetails()

	return authContext, nil
}

// AuthMethod interface implementation for the base OpenidConnectAuthMethod

func (o *OpenidConnectAuthMethod) IsBuiltInAdmin() bool               { return false }
func (o *OpenidConnectAuthMethod) IsAdmin() bool                     { return false }
func (o *OpenidConnectAuthMethod) IsAuthenticated() bool              { return false }
func (o *OpenidConnectAuthMethod) RequiresCSRF() bool                   { return true }
func (o *OpenidConnectAuthMethod) CheckAuthState() bool                  { return false }
func (o *OpenidConnectAuthMethod) CanPublishModuleVersion(string) bool { return false }
func (o *OpenidConnectAuthMethod) CanUploadModuleVersion(string) bool  { return false }
func (o *OpenidConnectAuthMethod) CheckNamespaceAccess(string, string) bool { return false }
func (o *OpenidConnectAuthMethod) GetAllNamespacePermissions() map[string]string { return make(map[string]string) }
func (o *OpenidConnectAuthMethod) GetUsername() string                { return "" }
func (o *OpenidConnectAuthMethod) GetUserGroupNames() []string       { return []string{} }
func (o *OpenidConnectAuthMethod) CanAccessReadAPI() bool             { return false }
func (o *OpenidConnectAuthMethod) CanAccessTerraformAPI() bool       { return false }
func (o *OpenidConnectAuthMethod) GetTerraformAuthToken() string     { return "" }
func (o *OpenidConnectAuthMethod) GetProviderData() map[string]interface{} { return make(map[string]interface{}) }

func (o *OpenidConnectAuthMethod) GetProviderType() auth.AuthMethodType {
	return auth.AuthMethodOpenIDConnect
}

func (o *OpenidConnectAuthMethod) IsEnabled() bool {
	// Check if OpenID Connect is enabled in config
	// For now, assume it's enabled if ClientID is set
	return o.config.OpenIDConnectClientID != ""
}