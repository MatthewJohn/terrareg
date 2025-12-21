package auth

import (
	"context"
	"fmt"
	"time"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/auth"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/config"
)

// OpenidConnectAuthMethod implements immutable OpenID Connect authentication factory
// This is a factory that creates OpenID Connect auth contexts with actual permission logic
type OpenidConnectAuthMethod struct {
	config     *config.InfrastructureConfig
	oidcService auth.OIDCValidator
}

// NewOpenidConnectAuthMethod creates a new immutable OpenID Connect auth method factory
func NewOpenidConnectAuthMethod(config *config.InfrastructureConfig, oidcService auth.OIDCValidator) *OpenidConnectAuthMethod {
	return &OpenidConnectAuthMethod{
		config:     config,
		oidcService: oidcService,
	}
}

// Authenticate validates OpenID Connect session and returns an OpenidConnectAuthContext
func (o *OpenidConnectAuthMethod) Authenticate(ctx context.Context, sessionData map[string]interface{}) (auth.AuthContext, error) {
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
	username := userInfo.Name
	if username == "" {
		if userInfo.Email != "" {
			username = userInfo.Email
		} else {
			username = userInfo.Sub
		}
	}

	// Create OpenidConnectAuthContext with authentication state
	authContext := auth.NewOpenidConnectAuthContext(ctx, userInfo.Sub, make(map[string]interface{}))

	// Extract user details from claims (username, email, name, etc.)
	authContext.ExtractUserDetails()

	return authContext, nil
}

// AuthMethodFactory interface implementation

func (o *OpenidConnectAuthMethod) GetProviderType() auth.AuthMethodType {
	return auth.AuthMethodOpenIDConnect
}

func (o *OpenidConnectAuthMethod) IsEnabled() bool {
	// Check if OpenID Connect is enabled in config
	// For now, assume it's enabled if ClientID is set
	return o.config.OpenIDConnectClientID != ""
}