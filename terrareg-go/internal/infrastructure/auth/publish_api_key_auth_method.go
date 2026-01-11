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
func (p *PublishApiKeyAuthMethod) Authenticate(ctx context.Context, headers, formData, queryParams map[string]string) (auth.AuthContext, error) {
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
