package repository

import (
	"context"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/config/model"
)

// ConfigRepository defines the interface for configuration retrieval
type ConfigRepository interface {
	// GetConfig retrieves the complete configuration
	GetConfig(ctx context.Context) (*model.Config, error)

	// GetVersion retrieves the application version
	GetVersion(ctx context.Context) (string, error)

	// IsOpenIDConnectEnabled checks if OpenID Connect authentication is enabled
	IsOpenIDConnectEnabled(ctx context.Context) bool

	// IsSAMLEnabled checks if SAML authentication is enabled
	IsSAMLEnabled(ctx context.Context) bool

	// IsAdminLoginEnabled checks if admin login is enabled
	IsAdminLoginEnabled(ctx context.Context) bool

	// GetProviderSources retrieves all configured provider sources
	GetProviderSources(ctx context.Context) ([]model.ProviderSource, error)
}
