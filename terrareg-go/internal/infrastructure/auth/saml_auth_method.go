package auth

import (
	"context"
	"fmt"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/auth"
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

// Authenticate validates SAML session and returns a SamlAuthContext
func (s *SamlAuthMethod) Authenticate(ctx context.Context, sessionData map[string]interface{}) (auth.AuthMethod, error) {
	// Check if SAML is enabled
	if !s.IsEnabled() {
		return nil, nil // Let other auth methods try
	}

	// Extract SAML NameID (required)
	samlNameId, hasNameId := sessionData["samlNameId"]
	if !hasNameId || samlNameId == "" {
		return nil, nil // Let other auth methods try
	}

	// Convert NameID to string
	nameIdStr, ok := samlNameId.(string)
	if !ok {
		nameIdStr = fmt.Sprintf("%v", samlNameId)
	}

	// Extract SAML attributes
	attributes := make(map[string][]string)
	samlUserdata, exists := sessionData["samlUserdata"]
	if exists {
		if userdata, ok := samlUserdata.(map[string]interface{}); ok {
			// Extract attributes from SAML userdata
			for key, value := range userdata {
				if key == "groups" || key == "memberOf" {
					// Handle groups as string arrays
					if groupList, ok := value.([]string); ok {
						attributes[key] = groupList
					} else if groupStr, ok := value.(string); ok {
						attributes[key] = []string{groupStr}
					}
				} else {
					// Handle other attributes as strings
					if attrStr, ok := value.(string); ok {
						attributes[key] = []string{attrStr}
					}
				}
			}
		}
	}

	// Create SamlAuthContext with authentication state
	authContext := auth.NewSamlAuthContext(ctx, nameIdStr, attributes)

	// Extract user details from SAML attributes
	authContext.ExtractUserDetails()

	return authContext, nil
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