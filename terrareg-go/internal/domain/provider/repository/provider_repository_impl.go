package repository

import (
	"context"
	"fmt"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/provider"
)

// ProviderRepositoryImpl implements the provider repository
type ProviderRepositoryImpl struct {
	// For now, this is a placeholder implementation
	// In a real implementation, this would use GORM or another ORM
	// For the purpose of this demonstration, we'll use in-memory storage
	providers map[int]*provider.Provider
	versions  map[int][]*provider.ProviderVersion
	nextID    int
}

// NewProviderRepository creates a new provider repository implementation
func NewProviderRepository() ProviderRepository {
	return &ProviderRepositoryImpl{
		providers: make(map[int]*provider.Provider),
		versions:  make(map[int][]*provider.ProviderVersion),
		nextID:    1,
	}
}

// Save persists a provider aggregate to the database
func (r *ProviderRepositoryImpl) Save(ctx context.Context, providerEntity *provider.Provider) error {
	if providerEntity.ID() == 0 {
		// New provider, assign ID
		providerEntity.SetID(r.nextID)
		r.nextID++
	}

	// Store provider
	r.providers[providerEntity.ID()] = providerEntity

	return nil
}

// FindAll retrieves all providers with pagination
func (r *ProviderRepositoryImpl) FindAll(ctx context.Context, offset, limit int) ([]*provider.Provider, int, error) {
	allProviders := make([]*provider.Provider, 0, len(r.providers))
	for _, p := range r.providers {
		allProviders = append(allProviders, p)
	}

	total := len(allProviders)

	// Apply pagination
	if offset >= total {
		return []*provider.Provider{}, total, nil
	}

	end := offset + limit
	if end > total {
		end = total
	}

	return allProviders[offset:end], total, nil
}

// Search searches for providers by query
func (r *ProviderRepositoryImpl) Search(ctx context.Context, query string, offset, limit int) ([]*provider.Provider, int, error) {
	// Simple implementation - in reality this would use database search
	allProviders := make([]*provider.Provider, 0, len(r.providers))
	for _, p := range r.providers {
		// Simple name matching
		if fmt.Sprintf("%s", p.Name()) == query {
			allProviders = append(allProviders, p)
		}
	}

	total := len(allProviders)

	// Apply pagination
	if offset >= total {
		return []*provider.Provider{}, total, nil
	}

	end := offset + limit
	if end > total {
		end = total
	}

	return allProviders[offset:end], total, nil
}

// FindByNamespaceAndName retrieves a provider by namespace and name
func (r *ProviderRepositoryImpl) FindByNamespaceAndName(ctx context.Context, namespace, providerName string) (*provider.Provider, error) {
	for _, p := range r.providers {
		// This is a simplified implementation
		// In reality, you'd need to resolve namespace ID and compare
		if p.Name() == providerName {
			return p, nil
		}
	}

	return nil, fmt.Errorf("provider not found: %s/%s", namespace, providerName)
}

// FindVersionsByProvider retrieves all versions for a provider
func (r *ProviderRepositoryImpl) FindVersionsByProvider(ctx context.Context, providerID int) ([]*provider.ProviderVersion, error) {
	versions, exists := r.versions[providerID]
	if !exists {
		return []*provider.ProviderVersion{}, nil
	}

	// Return a copy to avoid modification
	result := make([]*provider.ProviderVersion, len(versions))
	copy(result, versions)

	return result, nil
}

// FindVersionByProviderAndVersion retrieves a specific version
func (r *ProviderRepositoryImpl) FindVersionByProviderAndVersion(ctx context.Context, providerID int, version string) (*provider.ProviderVersion, error) {
	versions, exists := r.versions[providerID]
	if !exists {
		return nil, fmt.Errorf("no versions found for provider ID %d", providerID)
	}

	for _, v := range versions {
		if v.Version() == version {
			return v, nil
		}
	}

	return nil, fmt.Errorf("version %s not found for provider ID %d", version, providerID)
}