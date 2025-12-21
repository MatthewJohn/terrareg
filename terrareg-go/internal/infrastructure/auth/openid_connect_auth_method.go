package auth

import (
	"context"
	"fmt"
	"time"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/auth"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/auth/model"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/config"
)

// OpenidConnectAuthMethod implements immutable OpenID Connect authentication
type OpenidConnectAuthMethod struct {
	config *config.InfrastructureConfig
}

// NewOpenidConnectAuthMethod creates a new immutable OpenID Connect auth method
func NewOpenidConnectAuthMethod(config *config.InfrastructureConfig) *OpenidConnectAuthMethod {
	return &OpenidConnectAuthMethod{
		config: config,
	}
}

// Authenticate validates OpenID Connect session and returns an adapter with authentication state
func (o *OpenidConnectAuthMethod) Authenticate(ctx context.Context, sessionData map[string]interface{}) (auth.AuthMethod, error) {
	// Check if OpenID Connect is enabled
	if !o.IsEnabled() {
		return nil, fmt.Errorf("OpenID Connect authentication not supported")
	}

	// Check session expiry
	expiresAtFloat, hasExpiry := sessionData["openid_connect_expires_at"]
	if !hasExpiry {
		return model.NewAuthContextBuilder(ctx).Build(), nil
	}

	expiresAt, ok := expiresAtFloat.(float64)
	if !ok {
		return model.NewAuthContextBuilder(ctx).Build(), nil
	}

	// Check if session has expired
	if time.Now().After(time.Unix(int64(expiresAt), 0)) {
		return model.NewAuthContextBuilder(ctx).Build(), nil
	}

	// Validate ID token if present
	idToken, hasToken := sessionData["openid_connect_id_token"]
	if !hasToken || idToken == "" {
		return model.NewAuthContextBuilder(ctx).Build(), nil
	}

	// In a real implementation, validate the token here
	// For now, assume validation passes if token exists

	username, hasUsername := sessionData["openid_username"]
	groups, _ := sessionData["openid_groups"]

	if !hasUsername || username == "" || idToken == "" {
		return model.NewAuthContextBuilder(ctx).Build(), nil
	}

	// Create permission checking adapter with user data
	adapter := &model.PermissionCheckingAdapter{
		BaseAdapter: model.NewBaseAdapter(o, ctx),
		isAdmin:     false, // OpenID users are not admins by default
	}

	// TODO: Add namespace permissions based on groups when user/group management is implemented

	return adapter, nil
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