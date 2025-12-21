package auth

import (
	"context"
)

// TerraformOidcAuthContext implements AuthContext for Terraform OIDC-based authentication
// It holds the authentication state and permission logic for Terraform workflows
type TerraformOidcAuthContext struct {
	BaseAuthContext
	subject     string
	permissions map[string]bool // Terraform-specific permissions like "read", "download"
}

// NewTerraformOidcAuthContext creates a new Terraform OIDC auth context
func NewTerraformOidcAuthContext(ctx context.Context, subject string) *TerraformOidcAuthContext {
	return &TerraformOidcAuthContext{
		BaseAuthContext: BaseAuthContext{ctx: ctx},
		subject:         subject,
		permissions:     make(map[string]bool),
	}
}

// AddPermission adds a Terraform-specific permission
func (t *TerraformOidcAuthContext) AddPermission(permission string) {
	if t.permissions == nil {
		t.permissions = make(map[string]bool)
	}
	t.permissions[permission] = true
}

// HasPermission checks if the context has a specific Terraform permission
func (t *TerraformOidcAuthContext) HasPermission(permission string) bool {
	if t.permissions == nil {
		return false
	}
	return t.permissions[permission]
}

// GetProviderType returns the authentication method type
func (t *TerraformOidcAuthContext) GetProviderType() AuthMethodType {
	return AuthMethodTerraformOIDC
}

// GetUsername returns the subject from the Terraform OIDC token
func (t *TerraformOidcAuthContext) GetUsername() string {
	return t.subject
}

// IsAuthenticated returns true if Terraform OIDC authentication was successful
func (t *TerraformOidcAuthContext) IsAuthenticated() bool {
	return t.subject != ""
}

// IsAdmin returns false for Terraform OIDC (service accounts are not admins)
func (t *TerraformOidcAuthContext) IsAdmin() bool {
	return false
}

// IsBuiltInAdmin returns false for Terraform OIDC users
func (t *TerraformOidcAuthContext) IsBuiltInAdmin() bool {
	return false
}

// RequiresCSRF returns false for Terraform OIDC (uses bearer tokens)
func (t *TerraformOidcAuthContext) RequiresCSRF() bool {
	return false
}

// IsEnabled returns true if the Terraform OIDC authentication is valid
func (t *TerraformOidcAuthContext) IsEnabled() bool {
	return t.IsAuthenticated()
}

// CheckAuthState returns true if the Terraform OIDC context is in a valid state
func (t *TerraformOidcAuthContext) CheckAuthState() bool {
	return t.IsAuthenticated()
}

// CanPublishModuleVersion returns false for Terraform OIDC (service accounts cannot publish)
func (t *TerraformOidcAuthContext) CanPublishModuleVersion(namespace string) bool {
	return false
}

// CanUploadModuleVersion returns false for Terraform OIDC (service accounts cannot upload)
func (t *TerraformOidcAuthContext) CanUploadModuleVersion(namespace string) bool {
	return false
}

// CheckNamespaceAccess returns false for Terraform OIDC (service accounts don't have namespace permissions)
func (t *TerraformOidcAuthContext) CheckNamespaceAccess(permissionType, namespace string) bool {
	return false
}

// GetAllNamespacePermissions returns empty permissions for Terraform OIDC
func (t *TerraformOidcAuthContext) GetAllNamespacePermissions() map[string]string {
	return make(map[string]string)
}

// GetUserGroupNames returns empty slice for Terraform OIDC (service accounts don't have groups)
func (t *TerraformOidcAuthContext) GetUserGroupNames() []string {
	return []string{}
}

// CanAccessReadAPI returns true if the service account has read permission
func (t *TerraformOidcAuthContext) CanAccessReadAPI() bool {
	return t.HasPermission("read")
}

// CanAccessTerraformAPI returns true for Terraform OIDC (this is specifically for Terraform)
func (t *TerraformOidcAuthContext) CanAccessTerraformAPI() bool {
	return true
}

// GetTerraformAuthToken returns the bearer token for Terraform
func (t *TerraformOidcAuthContext) GetTerraformAuthToken() string {
	// This would be set during authentication
	return ""
}

// SetTerraformAuthToken sets the bearer token for Terraform
func (t *TerraformOidcAuthContext) SetTerraformAuthToken(token string) {
	// Store token if needed
}

// GetProviderData returns provider-specific data for Terraform OIDC
func (t *TerraformOidcAuthContext) GetProviderData() map[string]interface{} {
	return map[string]interface{}{
		"subject":     t.subject,
		"permissions": t.permissions,
		"auth_method": string(AuthMethodTerraformOIDC),
	}
}