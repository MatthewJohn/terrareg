package auth

import (
	"context"
	"strings"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/auth"
)

// AdminApiKeyAuthMethod implements authentication for admin users via API keys
type AdminApiKeyAuthMethod struct {
	auth.BaseAuthMethod
	apiKey   string
	isValid  bool
	isAdmin  bool
	username string
}

// NewAdminApiKeyAuthMethod creates a new admin API key authentication method
func NewAdminApiKeyAuthMethod() *AdminApiKeyAuthMethod {
	return &AdminApiKeyAuthMethod{
		isValid: false,
		isAdmin: false,
	}
}

// GetProviderType returns the authentication provider type
func (a *AdminApiKeyAuthMethod) GetProviderType() auth.AuthMethodType {
	return auth.AuthMethodAdminApiKey
}

// CheckAuthState validates the current authentication state
func (a *AdminApiKeyAuthMethod) CheckAuthState() bool {
	return a.isValid
}

// IsBuiltInAdmin returns whether this is a built-in admin method
func (a *AdminApiKeyAuthMethod) IsBuiltInAdmin() bool {
	return a.isAdmin
}

// IsAuthenticated returns whether the current request is authenticated
func (a *AdminApiKeyAuthMethod) IsAuthenticated() bool {
	return a.isValid
}

// IsAdmin returns whether the authenticated user has admin privileges
func (a *AdminApiKeyAuthMethod) IsAdmin() bool {
	return a.isAdmin
}

// IsEnabled returns whether this authentication method is enabled
func (a *AdminApiKeyAuthMethod) IsEnabled() bool {
	return true
}

// RequiresCSRF returns whether this authentication method requires CSRF protection
func (a *AdminApiKeyAuthMethod) RequiresCSRF() bool {
	return false // API key auth doesn't need CSRF
}

// CanPublishModuleVersion checks if the user can publish module versions to the given namespace
func (a *AdminApiKeyAuthMethod) CanPublishModuleVersion(namespace string) bool {
	// Admin API keys have full access to all namespaces
	return a.isAdmin
}

// CanUploadModuleVersion checks if the user can upload module versions to the given namespace
func (a *AdminApiKeyAuthMethod) CanUploadModuleVersion(namespace string) bool {
	// Admin API keys have full access to all namespaces
	return a.isAdmin
}

// CheckNamespaceAccess checks if the user has the specified permission for a namespace
func (a *AdminApiKeyAuthMethod) CheckNamespaceAccess(permissionType, namespace string) bool {
	// Admin API keys have full access to all namespaces
	return a.isAdmin
}

// GetAllNamespacePermissions returns all namespace permissions for the user
func (a *AdminApiKeyAuthMethod) GetAllNamespacePermissions() map[string]string {
	// Admin API keys have full permissions to all namespaces
	// Return empty map to signify admin access
	return map[string]string{}
}

// GetUsername returns the authenticated username
func (a *AdminApiKeyAuthMethod) GetUsername() string {
	return a.username
}

// GetUserGroupNames returns the names of all user groups
func (a *AdminApiKeyAuthMethod) GetUserGroupNames() []string {
	// API key auth doesn't use traditional user groups
	return []string{"admin"}
}

// CanAccessReadAPI returns whether the user can access read APIs
func (a *AdminApiKeyAuthMethod) CanAccessReadAPI() bool {
	return a.isValid
}

// CanAccessTerraformAPI returns whether the user can access Terraform APIs
func (a *AdminApiKeyAuthMethod) CanAccessTerraformAPI() bool {
	return a.isAdmin
}

// GetTerraformAuthToken returns the Terraform authentication token
func (a *AdminApiKeyAuthMethod) GetTerraformAuthToken() string {
	return a.apiKey
}

// GetProviderData returns provider-specific data
func (a *AdminApiKeyAuthMethod) GetProviderData() map[string]interface{} {
	data := make(map[string]interface{})
	data["api_key"] = a.apiKey
	data["is_admin"] = a.isAdmin
	data["username"] = a.username
	return data
}

// Authenticate authenticates a request using API key from headers
func (a *AdminApiKeyAuthMethod) Authenticate(ctx context.Context, headers map[string]string, cookies map[string]string) error {
	// Look for API key in Authorization header
	authHeader, exists := headers["Authorization"]
	if !exists {
		return adminApiKeyErr("missing authorization header")
	}

	if !strings.HasPrefix(authHeader, "Bearer ") {
		return adminApiKeyErr("invalid authorization header format")
	}

	apiKey := strings.TrimSpace(authHeader[7:]) // Remove "Bearer " prefix
	if apiKey == "" {
		return adminApiKeyErr("empty API key")
	}

	// For now, validate against a simple pattern
	// In a real implementation, this would check against a database of valid admin API keys
	if !a.isValidAdminKey(apiKey) {
		return adminApiKeyErr("invalid admin API key")
	}

	a.apiKey = apiKey
	a.isValid = true
	a.isAdmin = true
	a.username = "Admin API Key User"

	return nil
}

// isValidAdminKey checks if the provided API key is a valid admin API key
// This is a placeholder implementation - in a real system, you would validate against a database
func (a *AdminApiKeyAuthMethod) isValidAdminKey(apiKey string) bool {
	// Simple validation for demonstration
	// In production, this would check against a database or configuration
	validKeys := []string{
		"admin-api-key-12345",
		"admin-api-key-67890",
		"admin-api-key-abcdef",
	}

	for _, validKey := range validKeys {
		if apiKey == validKey {
			return true
		}
	}

	return false
}

// adminApiKeyErr creates a formatted error message for admin API key authentication
func adminApiKeyErr(message string) error {
	return &AdminApiKeyError{Message: message}
}

// AdminApiKeyError represents an admin API key authentication error
type AdminApiKeyError struct {
	Message string
}

func (e *AdminApiKeyError) Error() string {
	return "Admin API Key authentication failed: " + e.Message
}
