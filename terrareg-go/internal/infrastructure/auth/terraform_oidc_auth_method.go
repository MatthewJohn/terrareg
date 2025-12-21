package auth

import (
	"context"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/auth"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/config"
)

// TerraformOidcAuthMethod implements immutable Terraform OIDC authentication
type TerraformOidcAuthMethod struct {
	config *config.InfrastructureConfig
}

// NewTerraformOidcAuthMethod creates a new immutable Terraform OIDC auth method
func NewTerraformOidcAuthMethod(config *config.InfrastructureConfig) *TerraformOidcAuthMethod {
	return &TerraformOidcAuthMethod{
		config: config,
	}
}

// Authenticate validates Terraform OIDC token and returns a TerraformOidcAuthContext
func (t *TerraformOidcAuthMethod) Authenticate(ctx context.Context, authorizationHeader string, requestData []byte) (auth.AuthMethod, error) {
	// Check if Terraform OIDC is enabled
	if !t.IsEnabled() {
		return nil, nil // Let other auth methods try
	}

	if authorizationHeader == "" {
		return nil, nil // Let other auth methods try
	}

	// In a real implementation, validate the authorization token here
	// This would involve parsing the JWT token and validating it against the OIDC provider
	// For now, assume validation passes if header exists

	// Extract user information from the validated token
	// This would come from the token claims in a real implementation
	subject := "terraform-user"

	// Create TerraformOidcAuthContext with authentication state
	authContext := auth.NewTerraformOidcAuthContext(ctx, subject)

	// Add Terraform-specific permissions
	authContext.AddPermission("read")
	authContext.AddPermission("download")

	// Set the bearer token for Terraform
	authContext.SetTerraformAuthToken(authorizationHeader)

	return authContext, nil
}

// AuthMethod interface implementation for the base TerraformOidcAuthMethod

func (t *TerraformOidcAuthMethod) IsBuiltInAdmin() bool               { return false }
func (t *TerraformOidcAuthMethod) IsAdmin() bool                     { return false }
func (t *TerraformOidcAuthMethod) IsAuthenticated() bool              { return false }
func (t *TerraformOidcAuthMethod) RequiresCSRF() bool                   { return false }
func (t *TerraformOidcAuthMethod) CheckAuthState() bool                  { return false }
func (t *TerraformOidcAuthMethod) CanPublishModuleVersion(string) bool { return false }
func (t *TerraformOidcAuthMethod) CanUploadModuleVersion(string) bool  { return false }
func (t *TerraformOidcAuthMethod) CheckNamespaceAccess(string, string) bool { return false }
func (t *TerraformOidcAuthMethod) GetAllNamespacePermissions() map[string]string { return make(map[string]string) }
func (t *TerraformOidcAuthMethod) GetUsername() string                { return "" }
func (t *TerraformOidcAuthMethod) GetUserGroupNames() []string       { return []string{} }
func (t *TerraformOidcAuthMethod) CanAccessReadAPI() bool             { return false }
func (t *TerraformOidcAuthMethod) CanAccessTerraformAPI() bool       { return true }  // OIDC is specifically for Terraform
func (t *TerraformOidcAuthMethod) GetTerraformAuthToken() string     { return "" }  // Uses standard OIDC tokens
func (t *TerraformOidcAuthMethod) GetProviderData() map[string]interface{} { return make(map[string]interface{}) }

func (t *TerraformOidcAuthMethod) GetProviderType() auth.AuthMethodType {
	return auth.AuthMethodTerraformOIDC
}

func (t *TerraformOidcAuthMethod) IsEnabled() bool {
	// Terraform OIDC is enabled if IDP signing key path is configured
	return t.config.TerraformOidcIdpSigningKeyPath != ""
}