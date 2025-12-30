package auth

import (
	"context"
	"strings"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/auth"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/config"
)

// AdminApiKeyAuthMethod implements immutable authentication factory for admin API keys
type AdminApiKeyAuthMethod struct {
	config *config.InfrastructureConfig
}

// NewAdminApiKeyAuthMethod creates a new immutable admin API key authentication factory
func NewAdminApiKeyAuthMethod(config *config.InfrastructureConfig) *AdminApiKeyAuthMethod {
	return &AdminApiKeyAuthMethod{
		config: config,
	}
}

// GetProviderType returns the authentication provider type
func (a *AdminApiKeyAuthMethod) GetProviderType() auth.AuthMethodType {
	return auth.AuthMethodAdminApiKey
}

// IsEnabled returns whether this authentication method is enabled
func (a *AdminApiKeyAuthMethod) IsEnabled() bool {
	// Admin API key is enabled if ADMIN_AUTHENTICATION_TOKEN is configured
	return a.config.AdminAuthenticationToken != ""
}

// Authenticate authenticates an admin API key request and returns an AdminApiKeyAuthContext
func (a *AdminApiKeyAuthMethod) Authenticate(ctx context.Context, headers, formData, queryParams map[string]string) (auth.AuthContext, error) {
	// Check if admin API key is configured
	if !a.IsEnabled() {
		return nil, nil // Let other auth methods try
	}

	// Extract API key from X-Terrareg-ApiKey header (NOT Authorization header)
	// HTTP headers are case-insensitive, so check multiple variations
	apiKey := ""
	exists := false
	for key, value := range headers {
		if strings.EqualFold(key, "X-Terrareg-ApiKey") {
			apiKey = value
			exists = true
			break
		}
	}

	if !exists {
		return nil, nil // Let other auth methods try
	}

	// Validate API key against configured token
	if apiKey != a.config.AdminAuthenticationToken {
		return nil, nil // Let other auth methods try
	}

	// Create AdminApiKeyAuthContext with authentication state
	authContext := auth.NewAdminApiKeyAuthContext(ctx, apiKey)

	return authContext, nil
}

// AuthMethod interface implementation completed above
