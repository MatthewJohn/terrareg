package repository

import (
	"context"

	"github.com/terrareg/terrareg/internal/domain/provider"
)

// ProviderRepository defines the interface for provider persistence
type ProviderRepository interface {
	// FindAll retrieves all providers
	FindAll(ctx context.Context, offset, limit int) ([]*provider.Provider, int, error)

	// Search searches for providers by query
	Search(ctx context.Context, query string, offset, limit int) ([]*provider.Provider, int, error)

	// FindByNamespaceAndName retrieves a provider by namespace and name
	FindByNamespaceAndName(ctx context.Context, namespace, providerName string) (*provider.Provider, error)

	// FindVersionsByProvider retrieves all versions for a provider
	FindVersionsByProvider(ctx context.Context, providerID int) ([]*provider.ProviderVersion, error)

	// FindVersionByProviderAndVersion retrieves a specific version
	FindVersionByProviderAndVersion(ctx context.Context, providerID int, version string) (*provider.ProviderVersion, error)
}
