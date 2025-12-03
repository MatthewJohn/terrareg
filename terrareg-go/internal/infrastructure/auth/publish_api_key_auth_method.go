package auth

import (
	"context"
	"strings"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/auth"
)

// PublishApiKeyAuthMethod implements authentication for module publishing via API keys
type PublishApiKeyAuthMethod struct {
	auth.BaseAuthMethod
	apiKey           string
	isValid          bool
	username         string
	allowedNamespace string
}

// NewPublishApiKeyAuthMethod creates a new publish API key authentication method
func NewPublishApiKeyAuthMethod() *PublishApiKeyAuthMethod {
	return &PublishApiKeyAuthMethod{
		isValid: false,
	}
}

// GetProviderType returns the authentication provider type
func (p *PublishApiKeyAuthMethod) GetProviderType() auth.AuthMethodType {
	return auth.AuthMethodPublishApiKey
}

// CheckAuthState validates the current authentication state
func (p *PublishApiKeyAuthMethod) CheckAuthState() bool {
	return p.isValid
}

// IsBuiltInAdmin returns whether this is a built-in admin method
func (p *PublishApiKeyAuthMethod) IsBuiltInAdmin() bool {
	return false // Publish API keys are not built-in admin
}

// IsAuthenticated returns whether the current request is authenticated
func (p *PublishApiKeyAuthMethod) IsAuthenticated() bool {
	return p.isValid
}

// IsAdmin returns whether the authenticated user has admin privileges
func (p *PublishApiKeyAuthMethod) IsAdmin() bool {
	return false // Publish API keys are not admin
}

// IsEnabled returns whether this authentication method is enabled
func (p *PublishApiKeyAuthMethod) IsEnabled() bool {
	return true
}

// RequiresCSRF returns whether this authentication method requires CSRF protection
func (p *PublishApiKeyAuthMethod) RequiresCSRF() bool {
	return false // API key auth doesn't need CSRF
}

// CanPublishModuleVersion checks if the user can publish module versions to the given namespace
func (p *PublishApiKeyAuthMethod) CanPublishModuleVersion(namespace string) bool {
	// Publish API keys can publish to their allowed namespace only
	return p.isValid && p.allowedNamespace == namespace
}

// CanUploadModuleVersion checks if the user can upload module versions to the given namespace
func (p *PublishApiKeyAuthMethod) CanUploadModuleVersion(namespace string) bool {
	// Publish API keys can upload to their allowed namespace only
	return p.isValid && p.allowedNamespace == namespace
}

// CheckNamespaceAccess checks if the user has the specified permission for a namespace
func (p *PublishApiKeyAuthMethod) CheckNamespaceAccess(permissionType, namespace string) bool {
	if !p.isValid {
		return false
	}

	// Publish API keys only work with their allowed namespace
	if p.allowedNamespace != namespace {
		return false
	}

	// Check permission hierarchy - publish keys have FULL access to their namespace
	switch auth.PermissionType(permissionType) {
	case auth.PermissionRead, auth.PermissionModify, auth.PermissionFull:
		return true
	default:
		return false
	}
}

// GetAllNamespacePermissions returns all namespace permissions for the user
func (p *PublishApiKeyAuthMethod) GetAllNamespacePermissions() map[string]string {
	// Publish API keys only have permissions for their specific namespace
	if p.isValid && p.allowedNamespace != "" {
		return map[string]string{
			p.allowedNamespace: string(auth.PermissionFull),
		}
	}
	return map[string]string{}
}

// GetUsername returns the authenticated username
func (p *PublishApiKeyAuthMethod) GetUsername() string {
	return p.username
}

// GetUserGroupNames returns the names of all user groups
func (p *PublishApiKeyAuthMethod) GetUserGroupNames() []string {
	// API key auth doesn't use traditional user groups
	return []string{"publish-api-key"}
}

// CanAccessReadAPI returns whether the user can access read APIs
func (p *PublishApiKeyAuthMethod) CanAccessReadAPI() bool {
	// Publish API keys can access read APIs for their allowed namespace
	return p.isValid
}

// CanAccessTerraformAPI returns whether the user can access Terraform APIs
func (p *PublishApiKeyAuthMethod) CanAccessTerraformAPI() bool {
	return false // Publish API keys are for module publishing, not Terraform CLI access
}

// GetTerraformAuthToken returns the Terraform authentication token
func (p *PublishApiKeyAuthMethod) GetTerraformAuthToken() string {
	// Publish API keys don't provide Terraform tokens
	return ""
}

// GetProviderData returns provider-specific data
func (p *PublishApiKeyAuthMethod) GetProviderData() map[string]interface{} {
	data := make(map[string]interface{})
	data["api_key"] = p.apiKey
	data["username"] = p.username
	data["allowed_namespace"] = p.allowedNamespace
	return data
}

// Authenticate authenticates a request using API key from headers
func (p *PublishApiKeyAuthMethod) Authenticate(ctx context.Context, headers map[string]string, cookies map[string]string) error {
	// Look for API key in Authorization header
	authHeader, exists := headers["Authorization"]
	if !exists {
		return publishApiKeyErr("missing authorization header")
	}

	if !strings.HasPrefix(authHeader, "Bearer ") {
		return publishApiKeyErr("invalid authorization header format")
	}

	apiKey := strings.TrimSpace(authHeader[7:]) // Remove "Bearer " prefix
	if apiKey == "" {
		return publishApiKeyErr("empty API key")
	}

	// Validate publish API key
	keyInfo, err := p.validatePublishAPIKey(apiKey)
	if err != nil {
		return err
	}

	p.apiKey = apiKey
	p.isValid = true
	p.username = keyInfo.Username
	p.allowedNamespace = keyInfo.AllowedNamespace

	return nil
}

// PublishAPIKeyInfo represents information about a publish API key
type PublishAPIKeyInfo struct {
	Username         string
	AllowedNamespace string
}

// validatePublishAPIKey validates a publish API key and returns key information
// This is a placeholder implementation - in a real system, you would validate against a database
func (p *PublishApiKeyAuthMethod) validatePublishAPIKey(apiKey string) (*PublishAPIKeyInfo, error) {
	// Simple validation for demonstration
	// In production, this would check against a database or configuration
	validKeys := map[string]*PublishAPIKeyInfo{
		"publish-api-key-12345": {
			Username:         "Publish User 1",
			AllowedNamespace: "example",
		},
		"publish-api-key-67890": {
			Username:         "Publish User 2",
			AllowedNamespace: "my-company",
		},
		"publish-api-key-abcdef": {
			Username:         "Publish User 3",
			AllowedNamespace: "terraform-modules",
		},
	}

	keyInfo, exists := validKeys[apiKey]
	if !exists {
		return nil, publishApiKeyErr("invalid publish API key")
	}

	return keyInfo, nil
}

// publishApiKeyErr creates a formatted error message for publish API key authentication
func publishApiKeyErr(message string) error {
	return &PublishApiKeyError{Message: message}
}

// PublishApiKeyError represents a publish API key authentication error
type PublishApiKeyError struct {
	Message string
}

func (e *PublishApiKeyError) Error() string {
	return "Publish API Key authentication failed: " + e.Message
}