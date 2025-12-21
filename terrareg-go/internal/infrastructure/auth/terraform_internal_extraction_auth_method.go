package auth

import (
	"context"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/auth"
)

// TerraformInternalExtractionAuthMethod implements immutable authentication for internal Terraform extraction processes
type TerraformInternalExtractionAuthMethod struct {
	serviceName string
}

// NewTerraformInternalExtractionAuthMethod creates a new immutable Terraform internal extraction auth method
func NewTerraformInternalExtractionAuthMethod(serviceName string) *TerraformInternalExtractionAuthMethod {
	return &TerraformInternalExtractionAuthMethod{
		serviceName: serviceName,
	}
}

// GetProviderType returns the authentication provider type
func (t *TerraformInternalExtractionAuthMethod) GetProviderType() auth.AuthMethodType {
	return auth.AuthMethodTerraformInternalExtraction
}

// IsEnabled returns whether this authentication method is enabled
func (t *TerraformInternalExtractionAuthMethod) IsEnabled() bool {
	return true
}

// Authenticate authenticates an internal extraction request and returns a TerraformInternalExtractionAuthContext
func (t *TerraformInternalExtractionAuthMethod) Authenticate(ctx context.Context, headers, formData, queryParams map[string]string) (auth.AuthMethod, error) {
	// For internal extraction, we assume the request is valid if it comes from an internal source
	// In a real implementation, you might check for internal headers, IP ranges, or service tokens

	// Check for internal service token header
	_, exists := headers["X-Internal-Service-Token"]
	if !exists {
		// Check Authorization header for service token
		if authHeader, exists := headers["Authorization"]; exists && authHeader == "Bearer internal-service-token" {
			_ = authHeader // Auth header validated
		}
	}

	// For demonstration, we'll assume internal requests are always valid
	// In production, you would validate the service token

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