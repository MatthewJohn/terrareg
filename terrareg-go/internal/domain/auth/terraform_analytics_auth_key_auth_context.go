package auth

import (
	"context"
)

// TerraformAnalyticsAuthKeyAuthContext implements AuthContext for Terraform analytics via API keys
// It holds the authentication state and permission logic for Terraform analytics
type TerraformAnalyticsAuthKeyAuthContext struct {
	BaseAuthContext
	authKey        string
	canAccessAll   bool
	allowedModules []string
}

// NewTerraformAnalyticsAuthKeyAuthContext creates a new Terraform analytics auth key auth context
func NewTerraformAnalyticsAuthKeyAuthContext(ctx context.Context, authKey string) *TerraformAnalyticsAuthKeyAuthContext {
	return &TerraformAnalyticsAuthKeyAuthContext{
		BaseAuthContext: BaseAuthContext{ctx: ctx},
		authKey:         authKey,
		canAccessAll:    false,
		allowedModules:  make([]string, 0),
	}
}

// SetCanAccessAll sets whether the key can access all modules
func (t *TerraformAnalyticsAuthKeyAuthContext) SetCanAccessAll(canAccessAll bool) {
	t.canAccessAll = canAccessAll
}

// AddAllowedModule adds a module that the key can access
func (t *TerraformAnalyticsAuthKeyAuthContext) AddAllowedModule(module string) {
	t.allowedModules = append(t.allowedModules, module)
}

// CanAccessModule checks if the key can access a specific module
func (t *TerraformAnalyticsAuthKeyAuthContext) CanAccessModule(module string) bool {
	if t.canAccessAll {
		return true
	}

	for _, allowedModule := range t.allowedModules {
		if allowedModule == module {
			return true
		}
	}

	return false
}

// GetProviderType returns the authentication method type
func (t *TerraformAnalyticsAuthKeyAuthContext) GetProviderType() AuthMethodType {
	return AuthMethodTerraformAnalyticsAuthKey
}

// GetUsername returns the username for analytics (matches Python implementation)
func (t *TerraformAnalyticsAuthKeyAuthContext) GetUsername() string {
	return "Terraform deployment analytics token"
}

// IsAuthenticated returns true if the auth key is valid
func (t *TerraformAnalyticsAuthKeyAuthContext) IsAuthenticated() bool {
	return t.authKey != ""
}

// IsAdmin returns false for analytics auth keys
func (t *TerraformAnalyticsAuthKeyAuthContext) IsAdmin() bool {
	return false
}

// IsBuiltInAdmin returns false for analytics auth keys
func (t *TerraformAnalyticsAuthKeyAuthContext) IsBuiltInAdmin() bool {
	return false
}

// RequiresCSRF returns false for analytics auth keys
func (t *TerraformAnalyticsAuthKeyAuthContext) RequiresCSRF() bool {
	return false
}

// IsEnabled returns true if the analytics auth key is valid
func (t *TerraformAnalyticsAuthKeyAuthContext) IsEnabled() bool {
	return t.IsAuthenticated()
}

// CheckAuthState returns true if the analytics auth context is in a valid state
func (t *TerraformAnalyticsAuthKeyAuthContext) CheckAuthState() bool {
	return t.IsAuthenticated()
}

// CanPublishModuleVersion returns false for analytics auth keys
func (t *TerraformAnalyticsAuthKeyAuthContext) CanPublishModuleVersion(namespace string) bool {
	return false
}

// CanUploadModuleVersion returns false for analytics auth keys
func (t *TerraformAnalyticsAuthKeyAuthContext) CanUploadModuleVersion(namespace string) bool {
	return false
}

// CheckNamespaceAccess returns false for analytics auth keys
func (t *TerraformAnalyticsAuthKeyAuthContext) CheckNamespaceAccess(permissionType, namespace string) bool {
	return false
}

// GetAllNamespacePermissions returns empty permissions for analytics auth keys
func (t *TerraformAnalyticsAuthKeyAuthContext) GetAllNamespacePermissions() map[string]string {
	return make(map[string]string)
}

// GetUserGroupNames returns empty slice for analytics auth keys
func (t *TerraformAnalyticsAuthKeyAuthContext) GetUserGroupNames() []string {
	return []string{}
}

// CanAccessReadAPI returns true for analytics auth keys
func (t *TerraformAnalyticsAuthKeyAuthContext) CanAccessReadAPI() bool {
	return t.IsAuthenticated()
}

// CanAccessTerraformAPI returns false for analytics auth keys
func (t *TerraformAnalyticsAuthKeyAuthContext) CanAccessTerraformAPI() bool {
	return false
}

// GetTerraformAuthToken returns empty string for analytics auth keys
func (t *TerraformAnalyticsAuthKeyAuthContext) GetTerraformAuthToken() string {
	return ""
}

// GetProviderData returns provider-specific data for Terraform analytics auth
func (t *TerraformAnalyticsAuthKeyAuthContext) GetProviderData() map[string]interface{} {
	return map[string]interface{}{
		"auth_key":        t.authKey,
		"can_access_all":  t.canAccessAll,
		"allowed_modules": t.allowedModules,
		"auth_method":     string(AuthMethodTerraformAnalyticsAuthKey),
	}
}
