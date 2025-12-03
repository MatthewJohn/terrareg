package auth

import (
	"context"
	"strings"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/auth"
)

// UploadApiKeyAuthMethod implements authentication for module uploads via API keys
type UploadApiKeyAuthMethod struct {
	auth.BaseAuthMethod
	apiKey           string
	isValid          bool
	username         string
	allowedNamespace string
}

// NewUploadApiKeyAuthMethod creates a new upload API key authentication method
func NewUploadApiKeyAuthMethod() *UploadApiKeyAuthMethod {
	return &UploadApiKeyAuthMethod{
		isValid: false,
	}
}

// GetProviderType returns the authentication provider type
func (u *UploadApiKeyAuthMethod) GetProviderType() auth.AuthMethodType {
	return auth.AuthMethodUploadApiKey
}

// CheckAuthState validates the current authentication state
func (u *UploadApiKeyAuthMethod) CheckAuthState() bool {
	return u.isValid
}

// IsBuiltInAdmin returns whether this is a built-in admin method
func (u *UploadApiKeyAuthMethod) IsBuiltInAdmin() bool {
	return false // Upload API keys are not built-in admin
}

// IsAuthenticated returns whether the current request is authenticated
func (u *UploadApiKeyAuthMethod) IsAuthenticated() bool {
	return u.isValid
}

// IsAdmin returns whether the authenticated user has admin privileges
func (u *UploadApiKeyAuthMethod) IsAdmin() bool {
	return false // Upload API keys are not admin
}

// IsEnabled returns whether this authentication method is enabled
func (u *UploadApiKeyAuthMethod) IsEnabled() bool {
	return true
}

// RequiresCSRF returns whether this authentication method requires CSRF protection
func (u *UploadApiKeyAuthMethod) RequiresCSRF() bool {
	return false // API key auth doesn't need CSRF
}

// CanPublishModuleVersion checks if the user can publish module versions to the given namespace
func (u *UploadApiKeyAuthMethod) CanPublishModuleVersion(namespace string) bool {
	// Upload API keys can publish to their allowed namespace only
	return u.isValid && u.allowedNamespace == namespace
}

// CanUploadModuleVersion checks if the user can upload module versions to the given namespace
func (u *UploadApiKeyAuthMethod) CanUploadModuleVersion(namespace string) bool {
	// Upload API keys can upload to their allowed namespace only
	return u.isValid && u.allowedNamespace == namespace
}

// CheckNamespaceAccess checks if the user has the specified permission for a namespace
func (u *UploadApiKeyAuthMethod) CheckNamespaceAccess(permissionType, namespace string) bool {
	if !u.isValid {
		return false
	}

	// Upload API keys only work with their allowed namespace
	if u.allowedNamespace != namespace {
		return false
	}

	// Check permission hierarchy - upload keys have FULL access to their namespace
	switch auth.PermissionType(permissionType) {
	case auth.PermissionRead, auth.PermissionModify, auth.PermissionFull:
		return true
	default:
		return false
	}
}

// GetAllNamespacePermissions returns all namespace permissions for the user
func (u *UploadApiKeyAuthMethod) GetAllNamespacePermissions() map[string]string {
	// Upload API keys only have permissions for their specific namespace
	if u.isValid && u.allowedNamespace != "" {
		return map[string]string{
			u.allowedNamespace: string(auth.PermissionFull),
		}
	}
	return map[string]string{}
}

// GetUsername returns the authenticated username
func (u *UploadApiKeyAuthMethod) GetUsername() string {
	return u.username
}

// GetUserGroupNames returns the names of all user groups
func (u *UploadApiKeyAuthMethod) GetUserGroupNames() []string {
	// API key auth doesn't use traditional user groups
	return []string{"upload-api-key"}
}

// CanAccessReadAPI returns whether the user can access read APIs
func (u *UploadApiKeyAuthMethod) CanAccessReadAPI() bool {
	// Upload API keys can access read APIs for their allowed namespace
	return u.isValid
}

// CanAccessTerraformAPI returns whether the user can access Terraform APIs
func (u *UploadApiKeyAuthMethod) CanAccessTerraformAPI() bool {
	return false // Upload API keys are for module uploads, not Terraform CLI access
}

// GetTerraformAuthToken returns the Terraform authentication token
func (u *UploadApiKeyAuthMethod) GetTerraformAuthToken() string {
	// Upload API keys don't provide Terraform tokens
	return ""
}

// GetProviderData returns provider-specific data
func (u *UploadApiKeyAuthMethod) GetProviderData() map[string]interface{} {
	data := make(map[string]interface{})
	data["api_key"] = u.apiKey
	data["username"] = u.username
	data["allowed_namespace"] = u.allowedNamespace
	return data
}

// Authenticate authenticates a request using API key from headers
func (u *UploadApiKeyAuthMethod) Authenticate(ctx context.Context, headers map[string]string, cookies map[string]string) error {
	// Look for API key in Authorization header
	authHeader, exists := headers["Authorization"]
	if !exists {
		return uploadApiKeyErr("missing authorization header")
	}

	if !strings.HasPrefix(authHeader, "Bearer ") {
		return uploadApiKeyErr("invalid authorization header format")
	}

	apiKey := strings.TrimSpace(authHeader[7:]) // Remove "Bearer " prefix
	if apiKey == "" {
		return uploadApiKeyErr("empty API key")
	}

	// Validate upload API key
	keyInfo, err := u.validateUploadAPIKey(apiKey)
	if err != nil {
		return err
	}

	u.apiKey = apiKey
	u.isValid = true
	u.username = keyInfo.Username
	u.allowedNamespace = keyInfo.AllowedNamespace

	return nil
}

// UploadAPIKeyInfo represents information about an upload API key
type UploadAPIKeyInfo struct {
	Username         string
	AllowedNamespace string
}

// validateUploadAPIKey validates an upload API key and returns key information
// This is a placeholder implementation - in a real system, you would validate against a database
func (u *UploadApiKeyAuthMethod) validateUploadAPIKey(apiKey string) (*UploadAPIKeyInfo, error) {
	// Simple validation for demonstration
	// In production, this would check against a database or configuration
	validKeys := map[string]*UploadAPIKeyInfo{
		"upload-api-key-12345": {
			Username:         "Upload User 1",
			AllowedNamespace: "example",
		},
		"upload-api-key-67890": {
			Username:         "Upload User 2",
			AllowedNamespace: "my-company",
		},
		"upload-api-key-abcdef": {
			Username:         "Upload User 3",
			AllowedNamespace: "terraform-modules",
		},
	}

	keyInfo, exists := validKeys[apiKey]
	if !exists {
		return nil, uploadApiKeyErr("invalid upload API key")
	}

	return keyInfo, nil
}

// uploadApiKeyErr creates a formatted error message for upload API key authentication
func uploadApiKeyErr(message string) error {
	return &UploadApiKeyError{Message: message}
}

// UploadApiKeyError represents an upload API key authentication error
type UploadApiKeyError struct {
	Message string
}

func (e *UploadApiKeyError) Error() string {
	return "Upload API Key authentication failed: " + e.Message
}