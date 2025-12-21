package auth

import (
	"context"
	"strings"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/auth"
)

// TerraformAnalyticsAuthKeyAuthMethod implements immutable authentication for Terraform analytics via API keys
type TerraformAnalyticsAuthKeyAuthMethod struct {
	// No mutable state - this is immutable
}

// NewTerraformAnalyticsAuthKeyAuthMethod creates a new immutable Terraform analytics auth key authentication method
func NewTerraformAnalyticsAuthKeyAuthMethod() *TerraformAnalyticsAuthKeyAuthMethod {
	return &TerraformAnalyticsAuthKeyAuthMethod{}
}

// GetProviderType returns the authentication provider type
func (t *TerraformAnalyticsAuthKeyAuthMethod) GetProviderType() auth.AuthMethodType {
	return auth.AuthMethodTerraformAnalyticsAuthKey
}

// IsEnabled returns whether this authentication method is enabled
func (t *TerraformAnalyticsAuthKeyAuthMethod) IsEnabled() bool {
	return true
}

// Authenticate authenticates a request and returns a TerraformAnalyticsAuthKeyAuthContext
func (t *TerraformAnalyticsAuthKeyAuthMethod) Authenticate(ctx context.Context, headers, formData, queryParams map[string]string) (auth.AuthMethod, error) {
	// Look for auth key in Authorization header
	authHeader, exists := headers["Authorization"]
	if !exists {
		// Also check for X-Terraform-Analytics-Key header (common for analytics APIs)
		authHeader = headers["X-Terraform-Analytics-Key"]
	}

	if authHeader == "" {
		return nil, nil // Let other auth methods try
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
		return nil, nil // Let other auth methods try
	}

	// Validate analytics auth key
	keyInfo, err := t.validateAnalyticsAuthKey(authKey)
	if err != nil {
		return nil, nil // Let other auth methods try
	}

	// Create TerraformAnalyticsAuthKeyAuthContext with authentication state
	authContext := auth.NewTerraformAnalyticsAuthKeyAuthContext(ctx, authKey)

	// Set permissions based on key info
	authContext.SetCanAccessAll(keyInfo.CanAccessAll)

	// Add allowed modules
	for _, module := range keyInfo.AllowedModules {
		authContext.AddAllowedModule(module)
	}

	return authContext, nil
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
		return nil, &TerraformAnalyticsAuthKeyError{Message: "invalid analytics auth key"}
	}

	return keyInfo, nil
}

// AuthMethod interface implementation for the base TerraformAnalyticsAuthKeyAuthMethod
// These return default values since the actual auth state is in the TerraformAnalyticsAuthKeyAuthContext

func (t *TerraformAnalyticsAuthKeyAuthMethod) IsBuiltInAdmin() bool               { return false }
func (t *TerraformAnalyticsAuthKeyAuthMethod) IsAdmin() bool                     { return false }
func (t *TerraformAnalyticsAuthKeyAuthMethod) IsAuthenticated() bool              { return false }
func (t *TerraformAnalyticsAuthKeyAuthMethod) RequiresCSRF() bool                   { return false }
func (t *TerraformAnalyticsAuthKeyAuthMethod) CheckAuthState() bool                  { return false }
func (t *TerraformAnalyticsAuthKeyAuthMethod) CanPublishModuleVersion(string) bool { return false }
func (t *TerraformAnalyticsAuthKeyAuthMethod) CanUploadModuleVersion(string) bool  { return false }
func (t *TerraformAnalyticsAuthKeyAuthMethod) CheckNamespaceAccess(string, string) bool { return false }
func (t *TerraformAnalyticsAuthKeyAuthMethod) GetAllNamespacePermissions() map[string]string { return make(map[string]string) }
func (t *TerraformAnalyticsAuthKeyAuthMethod) GetUsername() string                { return "" }
func (t *TerraformAnalyticsAuthKeyAuthMethod) GetUserGroupNames() []string       { return []string{} }
func (t *TerraformAnalyticsAuthKeyAuthMethod) CanAccessReadAPI() bool             { return false }
func (t *TerraformAnalyticsAuthKeyAuthMethod) CanAccessTerraformAPI() bool       { return false }
func (t *TerraformAnalyticsAuthKeyAuthMethod) GetTerraformAuthToken() string     { return "" }
func (t *TerraformAnalyticsAuthKeyAuthMethod) GetProviderData() map[string]interface{} { return make(map[string]interface{}) }

// TerraformAnalyticsAuthKeyError represents a Terraform analytics auth key authentication error
type TerraformAnalyticsAuthKeyError struct {
	Message string
}

func (e *TerraformAnalyticsAuthKeyError) Error() string {
	return "Terraform Analytics Auth Key authentication failed: " + e.Message
}