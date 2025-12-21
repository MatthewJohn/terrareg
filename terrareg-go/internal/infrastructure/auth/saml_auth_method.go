package auth

import (
	"context"
	"fmt"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/auth"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/auth/model"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/config"
)

// SamlAuthMethod implements immutable SAML authentication
type SamlAuthMethod struct {
	config *config.InfrastructureConfig
}

// NewSamlAuthMethod creates a new immutable SAML auth method
func NewSamlAuthMethod(config *config.InfrastructureConfig) *SamlAuthMethod {
	return &SamlAuthMethod{
		config: config,
	}
}

// Authenticate validates SAML session and returns an adapter with authentication state
func (s *SamlAuthMethod) Authenticate(ctx context.Context, sessionData map[string]interface{}) (auth.AuthMethod, error) {
	// Check if SAML is enabled
	if !s.IsEnabled() {
		return nil, fmt.Errorf("SAML authentication not supported")
	}

	// Extract SAML session data
	samlUserdata, exists := sessionData["samlUserdata"]
	if !exists {
		return model.NewAuthContextBuilder(ctx).Build(), nil
	}

	samlNameId, hasNameId := sessionData["samlNameId"]

	// Create adapter with authentication state
	adapter := model.NewSessionStateAdapter(s, ctx)

	authenticated := exists && hasNameId && samlNameId != ""

	// Set user-specific data in adapter
	if authenticated {
		if nameIdStr, ok := samlNameId.(string); ok {
			adapter.SetSessionData(nil, true, nameIdStr, "")

			// Extract groups if available
			if userdata, ok := samlUserdata.(map[string]interface{}); ok {
				if groups, exists := userdata["samlGroups"]; exists { // Using generic key for now
					if groupList, ok := groups.([]string); ok {
						// Convert string groups to UserGroup objects if needed
						// For now, skip group setting as it would require UserGroup objects
					}
				}
			}
		}
	}

	return adapter, nil
}

// AuthMethod interface implementation for the base SamlAuthMethod

func (s *SamlAuthMethod) IsBuiltInAdmin() bool               { return false }
func (s *SamlAuthMethod) IsAdmin() bool                     { return false }
func (s *SamlAuthMethod) IsAuthenticated() bool              { return false }
func (s *SamlAuthMethod) RequiresCSRF() bool                   { return true }
func (s *SamlAuthMethod) CheckAuthState() bool                  { return false }
func (s *SamlAuthMethod) CanPublishModuleVersion(string) bool { return false }
func (s *SamlAuthMethod) CanUploadModuleVersion(string) bool  { return false }
func (s *SamlAuthMethod) CheckNamespaceAccess(string, string) bool { return false }
func (s *SamlAuthMethod) GetAllNamespacePermissions() map[string]string { return make(map[string]string) }
func (s *SamlAuthMethod) GetUsername() string                { return "" }
func (s *SamlAuthMethod) GetUserGroupNames() []string       { return []string{} }
func (s *SamlAuthMethod) CanAccessReadAPI() bool             { return false }
func (s *SamlAuthMethod) CanAccessTerraformAPI() bool       { return false }
func (s *SamlAuthMethod) GetTerraformAuthToken() string     { return "" }
func (s *SamlAuthMethod) GetProviderData() map[string]interface{} { return make(map[string]interface{}) }

func (s *SamlAuthMethod) GetProviderType() auth.AuthMethodType {
	return auth.AuthMethodSAML
}

func (s *SamlAuthMethod) IsEnabled() bool {
	// Check if SAML is enabled in config
	// For now, assume it's enabled if IDP Metadata URL is set
	return s.config.SAML2IDPMetadataURL != ""
}