package auth

import (
	"context"
)

// TerraformInternalExtractionAuthContext implements AuthContext for internal Terraform extraction processes
// It holds the authentication state and permission logic for internal Terraform operations
type TerraformInternalExtractionAuthContext struct {
	BaseAuthContext
	serviceName string
}

// NewTerraformInternalExtractionAuthContext creates a new Terraform internal extraction auth context
func NewTerraformInternalExtractionAuthContext(ctx context.Context, serviceName string) *TerraformInternalExtractionAuthContext {
	return &TerraformInternalExtractionAuthContext{
		BaseAuthContext: BaseAuthContext{ctx: ctx},
		serviceName:     serviceName,
	}
}

// GetProviderType returns the authentication method type
func (t *TerraformInternalExtractionAuthContext) GetProviderType() AuthMethodType {
	return AuthMethodTerraformInternalExtraction
}

// GetUsername returns the username for internal extraction (matches Python implementation)
func (t *TerraformInternalExtractionAuthContext) GetUsername() string {
	return "Terraform internal extraction"
}

// IsAuthenticated returns true for internal extraction (always authenticated)
func (t *TerraformInternalExtractionAuthContext) IsAuthenticated() bool {
	return true
}

// IsAdmin returns true for internal extraction (internal processes have elevated privileges)
func (t *TerraformInternalExtractionAuthContext) IsAdmin() bool {
	return true
}

// IsBuiltInAdmin returns true for internal extraction
func (t *TerraformInternalExtractionAuthContext) IsBuiltInAdmin() bool {
	return true
}

// RequiresCSRF returns false for internal extraction
func (t *TerraformInternalExtractionAuthContext) RequiresCSRF() bool {
	return false
}

// IsEnabled returns true for internal extraction
func (t *TerraformInternalExtractionAuthContext) IsEnabled() bool {
	return true
}

// CheckAuthState returns true for internal extraction
func (t *TerraformInternalExtractionAuthContext) CheckAuthState() bool {
	return true
}

// CanPublishModuleVersion returns true for internal extraction
func (t *TerraformInternalExtractionAuthContext) CanPublishModuleVersion(namespace string) bool {
	return true
}

// CanUploadModuleVersion returns true for internal extraction
func (t *TerraformInternalExtractionAuthContext) CanUploadModuleVersion(namespace string) bool {
	return true
}

// CheckNamespaceAccess returns true for internal extraction (access to all namespaces)
func (t *TerraformInternalExtractionAuthContext) CheckNamespaceAccess(permissionType, namespace string) bool {
	return true
}

// GetAllNamespacePermissions returns full permissions for all namespaces
func (t *TerraformInternalExtractionAuthContext) GetAllNamespacePermissions() map[string]string {
	return map[string]string{
		"*": "FULL",
	}
}

// GetUserGroupNames returns empty slice for internal extraction
func (t *TerraformInternalExtractionAuthContext) GetUserGroupNames() []string {
	return []string{}
}

// CanAccessReadAPI returns true for internal extraction
func (t *TerraformInternalExtractionAuthContext) CanAccessReadAPI() bool {
	return true
}

// CanAccessTerraformAPI returns true for internal extraction
func (t *TerraformInternalExtractionAuthContext) CanAccessTerraformAPI() bool {
	return true
}

// GetTerraformAuthToken returns empty string for internal extraction
func (t *TerraformInternalExtractionAuthContext) GetTerraformAuthToken() string {
	return ""
}

// ShouldRecordTerraformAnalytics returns false for internal extraction (matches Python)
func (t *TerraformInternalExtractionAuthContext) ShouldRecordTerraformAnalytics() bool {
	return false
}

// GetProviderData returns provider-specific data for internal extraction
func (t *TerraformInternalExtractionAuthContext) GetProviderData() map[string]interface{} {
	return map[string]interface{}{
		"service_name": t.serviceName,
		"is_internal":  true,
		"auth_method":  string(AuthMethodTerraformInternalExtraction),
	}
}