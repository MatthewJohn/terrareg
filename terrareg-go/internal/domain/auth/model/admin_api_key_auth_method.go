package model

import (
	"net/http"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/config"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/auth"
)

// AdminApiKeyAuthMethod implements authentication via X-Terrareg-ApiKey header
// Matches Python's AdminApiKeyAuthMethod behavior
type AdminApiKeyAuthMethod struct {
	*auth.BaseAuthMethod
	config      *config.Config
	apiKey      string // The X-Terrareg-ApiKey header value
	isValidated bool   // Whether the API key has been validated
}

// NewAdminApiKeyAuthMethod creates a new admin API key authentication method
func NewAdminApiKeyAuthMethod(config *config.Config) *AdminApiKeyAuthMethod {
	return &AdminApiKeyAuthMethod{
		BaseAuthMethod: auth.NewBaseAuthMethod(),
		config:         config,
	}
}

// SetAPIKey sets the API key from the X-Terrareg-ApiKey header
func (a *AdminApiKeyAuthMethod) SetAPIKey(apiKey string) {
	a.apiKey = apiKey
}

// IsBuiltInAdmin returns true for built-in admin authentication
func (a *AdminApiKeyAuthMethod) IsBuiltInAdmin() bool {
	return a.CheckAuthState()
}

// IsAdmin returns true if this is an admin authentication method
func (a *AdminApiKeyAuthMethod) IsAdmin() bool {
	return a.IsBuiltInAdmin()
}

// IsAuthenticated returns true if the API key is valid
func (a *AdminApiKeyAuthMethod) IsAuthenticated() bool {
	return a.CheckAuthState()
}

// IsEnabled returns true if admin authentication is configured
func (a *AdminApiKeyAuthMethod) IsEnabled() bool {
	return a.config.AdminAuthenticationToken != ""
}

// RequiresCSRF returns false for admin API key authentication
// API key authentication doesn't require CSRF protection
func (a *AdminApiKeyAuthMethod) RequiresCSRF() bool {
	return false
}

// CheckAuthState validates the API key against the configured admin token
func (a *AdminApiKeyAuthMethod) CheckAuthState() bool {
	if !a.IsEnabled() {
		return false
	}

	// Validate API key against configured admin authentication token
	// This matches Python's _check_api_key implementation
	if a.apiKey == "" {
		return false
	}

	return a.apiKey == a.config.AdminAuthenticationToken
}

// CanPublishModuleVersion returns true for admin authentication
func (a *AdminApiKeyAuthMethod) CanPublishModuleVersion(namespace string) bool {
	return a.IsBuiltInAdmin()
}

// CanUploadModuleVersion returns true for admin authentication
func (a *AdminApiKeyAuthMethod) CanUploadModuleVersion(namespace string) bool {
	return a.IsBuiltInAdmin()
}

// CheckNamespaceAccess returns true for admin authentication (access to all namespaces)
func (a *AdminApiKeyAuthMethod) CheckNamespaceAccess(permissionType, namespace string) bool {
	return a.IsBuiltInAdmin()
}

// GetAllNamespacePermissions returns full permissions for all namespaces for admin
func (a *AdminApiKeyAuthMethod) GetAllNamespacePermissions() map[string]string {
	if !a.IsBuiltInAdmin() {
		return make(map[string]string)
	}

	// Admin has full access to all namespaces
	// This matches Python behavior where built-in admin has all permissions
	return map[string]string{
		"*": "full",
	}
}

// GetUsername returns "admin" for admin authentication
func (a *AdminApiKeyAuthMethod) GetUsername() string {
	if a.IsBuiltInAdmin() {
		return "admin"
	}
	return ""
}

// GetUserGroupNames returns empty for admin authentication
func (a *AdminApiKeyAuthMethod) GetUserGroupNames() []string {
	return []string{}
}

// CanAccessReadAPI returns true for admin authentication
func (a *AdminApiKeyAuthMethod) CanAccessReadAPI() bool {
	return true
}

// CanAccessTerraformAPI returns true for admin authentication
func (a *AdminApiKeyAuthMethod) CanAccessTerraformAPI() bool {
	return true
}

// GetTerraformAuthToken returns the API key for Terraform authentication
func (a *AdminApiKeyAuthMethod) GetTerraformAuthToken() string {
	if a.IsBuiltInAdmin() {
		return a.apiKey
	}
	return ""
}

// GetProviderType returns the authentication method type
func (a *AdminApiKeyAuthMethod) GetProviderType() auth.AuthMethodType {
	return auth.AuthMethodAdminApiKey
}

// GetProviderData returns provider-specific data
func (a *AdminApiKeyAuthMethod) GetProviderData() map[string]interface{} {
	data := make(map[string]interface{})
	data["api_key"] = a.apiKey
	data["is_built_in_admin"] = a.IsBuiltInAdmin()
	return data
}

// ExtractFromRequest extracts the API key from HTTP request headers
// This should be called before using the authentication method
func (a *AdminApiKeyAuthMethod) ExtractFromRequest(r *http.Request) bool {
	apiKey := r.Header.Get("X-Terrareg-ApiKey")
	if apiKey == "" {
		a.isValidated = false
		return false
	}

	a.SetAPIKey(apiKey)
	a.isValidated = true
	return true
}
