package auth

import (
	"context"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/auth"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/config"
)

// PublishApiKeyAuthMethod implements immutable authentication for publish API keys
type PublishApiKeyAuthMethod struct {
	config *config.InfrastructureConfig
}

// NewPublishApiKeyAuthMethod creates a new immutable publish API key authentication method
func NewPublishApiKeyAuthMethod(config *config.InfrastructureConfig) *PublishApiKeyAuthMethod {
	return &PublishApiKeyAuthMethod{
		config: config,
	}
}

// GetProviderType returns the authentication provider type
func (p *PublishApiKeyAuthMethod) GetProviderType() auth.AuthMethodType {
	return auth.AuthMethodPublishApiKey
}

// IsEnabled returns whether this authentication method is enabled
func (p *PublishApiKeyAuthMethod) IsEnabled() bool {
	// Publish API key is enabled if PUBLISH_API_KEYS is configured and not empty
	return len(p.config.PublishApiKeys) > 0
}

// Authenticate authenticates a publish API key request and returns a PublishApiKeyAuthContext
func (p *PublishApiKeyAuthMethod) Authenticate(ctx context.Context, headers, formData, queryParams map[string]string) (auth.AuthMethod, error) {
	// Check if publish API keys are configured
	if !p.IsEnabled() {
		return nil, nil // Let other auth methods try
	}

	// Extract API key from X-Terrareg-ApiKey header (NOT Authorization header)
	apiKey, exists := headers["X-Terrareg-ApiKey"]
	if !exists {
		return nil, nil // Let other auth methods try
	}

	// Validate API key against configured publish keys
	validKey := false
	for _, configuredKey := range p.config.PublishApiKeys {
		if apiKey == configuredKey {
			validKey = true
			break
		}
	}

	if !validKey {
		return nil, nil // Let other auth methods try
	}

	// Create PublishApiKeyAuthContext with authentication state
	authContext := auth.NewPublishApiKeyAuthContext(ctx, apiKey)

	return authContext, nil
}

// AuthMethod interface implementation for the base PublishApiKeyAuthMethod
// These return default values since the actual auth state is in the PublishApiKeyAuthContext

func (p *PublishApiKeyAuthMethod) IsBuiltInAdmin() bool                     { return false }
func (p *PublishApiKeyAuthMethod) IsAdmin() bool                            { return false }
func (p *PublishApiKeyAuthMethod) IsAuthenticated() bool                    { return false }
func (p *PublishApiKeyAuthMethod) RequiresCSRF() bool                       { return false }
func (p *PublishApiKeyAuthMethod) CheckAuthState() bool                     { return false }
func (p *PublishApiKeyAuthMethod) CanPublishModuleVersion(string) bool      { return false }
func (p *PublishApiKeyAuthMethod) CanUploadModuleVersion(string) bool       { return false }
func (p *PublishApiKeyAuthMethod) CheckNamespaceAccess(string, string) bool { return false }
func (p *PublishApiKeyAuthMethod) GetAllNamespacePermissions() map[string]string {
	return make(map[string]string)
}
func (p *PublishApiKeyAuthMethod) GetUsername() string           { return "" }
func (p *PublishApiKeyAuthMethod) GetUserGroupNames() []string   { return []string{} }
func (p *PublishApiKeyAuthMethod) CanAccessReadAPI() bool        { return false }
func (p *PublishApiKeyAuthMethod) CanAccessTerraformAPI() bool   { return false }
func (p *PublishApiKeyAuthMethod) GetTerraformAuthToken() string { return "" }
func (p *PublishApiKeyAuthMethod) GetProviderData() map[string]interface{} {
	return make(map[string]interface{})
}
