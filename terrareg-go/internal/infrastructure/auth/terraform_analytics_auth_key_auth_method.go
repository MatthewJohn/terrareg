package auth

import (
	"context"
	"strings"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/auth"
)

// TerraformAnalyticsAuthKeyAuthMethod implements authentication for Terraform analytics via API keys
type TerraformAnalyticsAuthKeyAuthMethod struct {
	auth.BaseAuthMethod
	authKey        string
	isValid        bool
	canAccessAll   bool
	allowedModules []string
}

// NewTerraformAnalyticsAuthKeyAuthMethod creates a new Terraform analytics auth key authentication method
func NewTerraformAnalyticsAuthKeyAuthMethod() *TerraformAnalyticsAuthKeyAuthMethod {
	return &TerraformAnalyticsAuthKeyAuthMethod{
		isValid: false,
	}
}

// GetProviderType returns the authentication provider type
func (t *TerraformAnalyticsAuthKeyAuthMethod) GetProviderType() auth.AuthMethodType {
	return auth.AuthMethodTerraformAnalyticsAuthKey
}

// CheckAuthState validates the current authentication state
func (t *TerraformAnalyticsAuthKeyAuthMethod) CheckAuthState() bool {
	return t.isValid
}

// IsBuiltInAdmin returns whether this is a built-in admin method
func (t *TerraformAnalyticsAuthKeyAuthMethod) IsBuiltInAdmin() bool {
	return false // Analytics auth keys are not admin
}

// IsAuthenticated returns whether the current request is authenticated
func (t *TerraformAnalyticsAuthKeyAuthMethod) IsAuthenticated() bool {
	return t.isValid
}

// IsAdmin returns whether the authenticated user has admin privileges
func (t *TerraformAnalyticsAuthKeyAuthMethod) IsAdmin() bool {
	return false // Analytics auth keys are not admin
}

// IsEnabled returns whether this authentication method is enabled
func (t *TerraformAnalyticsAuthKeyAuthMethod) IsEnabled() bool {
	return true
}

// RequiresCSRF returns whether this authentication method requires CSRF protection
func (t *TerraformAnalyticsAuthKeyAuthMethod) RequiresCSRF() bool {
	return false // API key auth doesn't need CSRF
}

// CanPublishModuleVersion checks if the user can publish module versions to the given namespace
func (t *TerraformAnalyticsAuthKeyAuthMethod) CanPublishModuleVersion(namespace string) bool {
	// Analytics auth keys are read-only - no publishing access
	return false
}

// CanUploadModuleVersion checks if the user can upload module versions to the given namespace
func (t *TerraformAnalyticsAuthKeyAuthMethod) CanUploadModuleVersion(namespace string) bool {
	// Analytics auth keys are read-only - no uploading access
	return false
}

// CheckNamespaceAccess checks if the user has the specified permission for a namespace
func (t *TerraformAnalyticsAuthKeyAuthMethod) CheckNamespaceAccess(permissionType, namespace string) bool {
	if !t.isValid {
		return false
	}

	// Analytics auth keys only have READ access
	switch auth.PermissionType(permissionType) {
	case auth.PermissionRead:
		return t.canAccessAll || t.canAccessModule(namespace)
	default:
		return false
	}
}

// GetAllNamespacePermissions returns all namespace permissions for the user
func (t *TerraformAnalyticsAuthKeyAuthMethod) GetAllNamespacePermissions() map[string]string {
	permissions := make(map[string]string)

	if !t.isValid {
		return permissions
	}

	if t.canAccessAll {
		// Analytics keys with broad access can read all namespaces
		return map[string]string{"*": string(auth.PermissionRead)}
	}

	// Analytics keys with limited access
	for _, module := range t.allowedModules {
		// Extract namespace from module name (format: namespace/module-name)
		if parts := strings.Split(module, "/"); len(parts) >= 1 {
			namespace := parts[0]
			permissions[namespace] = string(auth.PermissionRead)
		}
	}

	return permissions
}

// GetUsername returns the authenticated username
func (t *TerraformAnalyticsAuthKeyAuthMethod) GetUsername() string {
	return "Terraform Analytics Key"
}

// GetUserGroupNames returns the names of all user groups
func (t *TerraformAnalyticsAuthKeyAuthMethod) GetUserGroupNames() []string {
	// API key auth doesn't use traditional user groups
	return []string{"analytics-api-key"}
}

// CanAccessReadAPI returns whether the user can access read APIs
func (t *TerraformAnalyticsAuthKeyAuthMethod) CanAccessReadAPI() bool {
	// Analytics auth keys can access read APIs for analytics data
	return t.isValid
}

// CanAccessTerraformAPI returns whether the user can access Terraform APIs
func (t *TerraformAnalyticsAuthKeyAuthMethod) CanAccessTerraformAPI() bool {
	// Analytics auth keys can access Terraform APIs for analytics purposes
	return t.isValid
}

// GetTerraformAuthToken returns the Terraform authentication token
func (t *TerraformAnalyticsAuthKeyAuthMethod) GetTerraformAuthToken() string {
	// Analytics auth keys don't provide Terraform tokens
	return ""
}

// GetProviderData returns provider-specific data
func (t *TerraformAnalyticsAuthKeyAuthMethod) GetProviderData() map[string]interface{} {
	data := make(map[string]interface{})
	data["auth_key"] = t.authKey
	data["can_access_all"] = t.canAccessAll
	data["allowed_modules"] = t.allowedModules
	return data
}

// Authenticate authenticates a request using analytics auth key from headers
func (t *TerraformAnalyticsAuthKeyAuthMethod) Authenticate(ctx context.Context, headers map[string]string, cookies map[string]string) error {
	// Look for auth key in Authorization header
	authHeader, exists := headers["Authorization"]
	if !exists {
		// Also check for X-Terraform-Analytics-Key header (common for analytics APIs)
		authHeader = headers["X-Terraform-Analytics-Key"]
	}

	if authHeader == "" {
		return terraformAnalyticsErr("missing authorization header or X-Terraform-Analytics-Key")
	}

	// Handle different header formats
	var authKey string
	if strings.HasPrefix(authHeader, "Bearer ") {
		authKey = strings.TrimSpace(authHeader[7:]) // Remove "Bearer " prefix
	} else if strings.HasPrefix(authHeader, "AnalyticsKey ") {
		authKey = strings.TrimSpace(authHeader[13:]) // Remove "AnalyticsKey " prefix
	} else {
		// Use raw header value if no prefix
		authKey = strings.TrimSpace(authHeader)
	}

	if authKey == "" {
		return terraformAnalyticsErr("empty analytics auth key")
	}

	// Validate analytics auth key
	keyInfo, err := t.validateAnalyticsAuthKey(authKey)
	if err != nil {
		return err
	}

	t.authKey = authKey
	t.isValid = true
	t.canAccessAll = keyInfo.CanAccessAll
	t.allowedModules = keyInfo.AllowedModules

	return nil
}

// canAccessModule checks if the key can access a specific module
func (t *TerraformAnalyticsAuthKeyAuthMethod) canAccessModule(namespace string) bool {
	if t.canAccessAll {
		return true
	}

	// Check if any allowed modules belong to this namespace
	for _, module := range t.allowedModules {
		if parts := strings.Split(module, "/"); len(parts) >= 1 && parts[0] == namespace {
			return true
		}
	}

	return false
}

// TerraformAnalyticsAuthKeyInfo represents information about an analytics auth key
type TerraformAnalyticsAuthKeyInfo struct {
	CanAccessAll   bool
	AllowedModules []string
}

// validateAnalyticsAuthKey validates an analytics auth key and returns key information
// This is a placeholder implementation - in a real system, you would validate against a database
func (t *TerraformAnalyticsAuthKeyAuthMethod) validateAnalyticsAuthKey(authKey string) (*TerraformAnalyticsAuthKeyInfo, error) {
	// Simple validation for demonstration
	// In production, this would check against a database or configuration
	validKeys := map[string]*TerraformAnalyticsAuthKeyInfo{
		"analytics-key-global": {
			CanAccessAll:   true,
			AllowedModules: []string{},
		},
		"analytics-key-example": {
			CanAccessAll:   false,
			AllowedModules: []string{"example/aws-vpc", "example/aws-ec2"},
		},
		"analytics-key-mycompany": {
			CanAccessAll:   false,
			AllowedModules: []string{"mycompany/*"}, // Wildcard support
		},
		"analytics-key-public": {
			CanAccessAll:   false,
			AllowedModules: []string{"hashicorp/aws", "hashicorp/azure", "hashicorp/gcp"},
		},
	}

	keyInfo, exists := validKeys[authKey]
	if !exists {
		return nil, terraformAnalyticsErr("invalid analytics auth key")
	}

	return keyInfo, nil
}

// terraformAnalyticsErr creates a formatted error message for Terraform analytics auth key authentication
func terraformAnalyticsErr(message string) error {
	return &TerraformAnalyticsAuthKeyError{Message: message}
}

// TerraformAnalyticsAuthKeyError represents a Terraform analytics auth key authentication error
type TerraformAnalyticsAuthKeyError struct {
	Message string
}

func (e *TerraformAnalyticsAuthKeyError) Error() string {
	return "Terraform Analytics Auth Key authentication failed: " + e.Message
}
