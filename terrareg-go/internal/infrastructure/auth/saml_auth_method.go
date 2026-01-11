package auth

import (
	"context"
	"fmt"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/auth"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/config"
)

// SamlAuthMethod implements immutable SAML authentication factory
// This is a factory that creates SAML auth contexts with actual permission logic
type SamlAuthMethod struct {
	config *config.InfrastructureConfig
}

// NewSamlAuthMethod creates a new immutable SAML auth method factory
func NewSamlAuthMethod(config *config.InfrastructureConfig) *SamlAuthMethod {
	return &SamlAuthMethod{
		config: config,
	}
}

// Authenticate validates SAML session and returns a SamlAuthContext
func (s *SamlAuthMethod) Authenticate(ctx context.Context, sessionData map[string]interface{}) (auth.AuthContext, error) {
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

// AuthMethod interface implementation

func (s *SamlAuthMethod) GetProviderType() auth.AuthMethodType {
	return auth.AuthMethodSAML
}

func (s *SamlAuthMethod) IsEnabled() bool {
	// Check if SAML is enabled in config
	// For now, assume it's enabled if IDP Metadata URL is set
	return s.config.SAML2IDPMetadataURL != ""
}
