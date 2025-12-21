package auth

import (
	"context"
	"strings"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/auth"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/config"
)

// TerraformInternalExtractionAuthMethod implements immutable authentication for internal Terraform extraction processes
type TerraformInternalExtractionAuthMethod struct {
	serviceName string
	config      *config.InfrastructureConfig
}

// NewTerraformInternalExtractionAuthMethod creates a new immutable Terraform internal extraction auth method
func NewTerraformInternalExtractionAuthMethod(serviceName string, config *config.InfrastructureConfig) *TerraformInternalExtractionAuthMethod {
	return &TerraformInternalExtractionAuthMethod{
		serviceName: serviceName,
		config:      config,
	}
}

// GetProviderType returns the authentication provider type
func (t *TerraformInternalExtractionAuthMethod) GetProviderType() auth.AuthMethodType {
	return auth.AuthMethodTerraformInternalExtraction
}

// IsEnabled returns whether this authentication method is enabled
func (t *TerraformInternalExtractionAuthMethod) IsEnabled() bool {
	// Internal extraction is enabled if token is configured
	return t.config.InternalExtractionAnalyticsToken != ""
}

// Authenticate authenticates an internal extraction request and returns a TerraformInternalExtractionAuthContext
func (t *TerraformInternalExtractionAuthMethod) Authenticate(ctx context.Context, headers, formData, queryParams map[string]string) (auth.AuthMethod, error) {
	// Check if INTERNAL_EXTRACTION_ANALYTICS_TOKEN is configured
	if !t.IsEnabled() {
		return nil, nil // Let other auth methods try
	}

	// Look for auth token in Authorization header
	authHeader, exists := headers["Authorization"]
	if !exists {
		return nil, nil // Let other auth methods try
	}

	// Handle Bearer token format
	expectedToken := t.config.InternalExtractionAnalyticsToken
	if !strings.HasPrefix(authHeader, "Bearer ") {
		return nil, nil // Let other auth methods try
	}

	// Extract and validate token
	token := strings.TrimSpace(authHeader[7:]) // Remove "Bearer " prefix
	if token != expectedToken {
		return nil, nil // Let other auth methods try
	}

	// Create TerraformInternalExtractionAuthContext with authentication state
	authContext := auth.NewTerraformInternalExtractionAuthContext(ctx, t.serviceName)

	return authContext, nil
}

// AuthMethod interface implementation for the base TerraformInternalExtractionAuthMethod
// These return default values since the actual auth state is in the TerraformInternalExtractionAuthContext

func (t *TerraformInternalExtractionAuthMethod) IsBuiltInAdmin() bool               { return false }
func (t *TerraformInternalExtractionAuthMethod) IsAdmin() bool                     { return false }
func (t *TerraformInternalExtractionAuthMethod) IsAuthenticated() bool              { return false }
func (t *TerraformInternalExtractionAuthMethod) RequiresCSRF() bool                   { return false }
func (t *TerraformInternalExtractionAuthMethod) CheckAuthState() bool                  { return false }
func (t *TerraformInternalExtractionAuthMethod) CanPublishModuleVersion(string) bool { return false }
func (t *TerraformInternalExtractionAuthMethod) CanUploadModuleVersion(string) bool  { return false }
func (t *TerraformInternalExtractionAuthMethod) CheckNamespaceAccess(string, string) bool { return false }
func (t *TerraformInternalExtractionAuthMethod) GetAllNamespacePermissions() map[string]string { return make(map[string]string) }
func (t *TerraformInternalExtractionAuthMethod) GetUsername() string                { return "" }
func (t *TerraformInternalExtractionAuthMethod) GetUserGroupNames() []string       { return []string{} }
func (t *TerraformInternalExtractionAuthMethod) CanAccessReadAPI() bool             { return false }
func (t *TerraformInternalExtractionAuthMethod) CanAccessTerraformAPI() bool       { return false }
func (t *TerraformInternalExtractionAuthMethod) GetTerraformAuthToken() string     { return "" }
func (t *TerraformInternalExtractionAuthMethod) GetProviderData() map[string]interface{} { return make(map[string]interface{}) }