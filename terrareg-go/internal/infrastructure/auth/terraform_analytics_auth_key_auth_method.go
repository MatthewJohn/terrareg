package auth

import (
	"context"
	"strings"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/auth"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/config"
)

// TerraformAnalyticsAuthKeyAuthMethod implements immutable authentication for Terraform analytics via API keys
type TerraformAnalyticsAuthKeyAuthMethod struct {
	config *config.InfrastructureConfig
}

// NewTerraformAnalyticsAuthKeyAuthMethod creates a new immutable Terraform analytics auth key authentication method
func NewTerraformAnalyticsAuthKeyAuthMethod(config *config.InfrastructureConfig) *TerraformAnalyticsAuthKeyAuthMethod {
	return &TerraformAnalyticsAuthKeyAuthMethod{
		config: config,
	}
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
// Python implementation uses ANALYTICS_AUTH_KEYS config with token:environment format
func (t *TerraformAnalyticsAuthKeyAuthMethod) validateAnalyticsAuthKey(authKey string) (*TerraformAnalyticsAuthKeyInfo, error) {
	// Handle token:environment format - only validate part before colon
	actualToken := authKey
	if colonIndex := strings.Index(authKey, ":"); colonIndex != -1 {
		actualToken = authKey[:colonIndex]
	}

	// Check against ANALYTICS_AUTH_KEYS configuration
	for _, configuredKey := range t.config.AnalyticsAuthKeys {
		// Configured keys might also be in token:environment format
		configToken := configuredKey
		if colonIndex := strings.Index(configuredKey, ":"); colonIndex != -1 {
			configToken = configuredKey[:colonIndex]
		}

		if actualToken == configToken {
			// For analytics keys, allow access to all modules
			// Python implementation doesn't restrict modules for analytics
			return &TerraformAnalyticsAuthKeyInfo{
				CanAccessAll:   true,
				AllowedModules: []string{}, // Empty means all modules
			}, nil
		}
	}

	return nil, &TerraformAnalyticsAuthKeyError{Message: "invalid analytics auth key"}
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