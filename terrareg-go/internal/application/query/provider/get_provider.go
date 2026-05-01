package provider

import (
	"context"
	"fmt"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/provider"
	providerRepo "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/provider/repository"
)

// GetProviderQuery handles retrieving a single provider
type GetProviderQuery struct {
	providerRepo providerRepo.ProviderRepository
}

// NewGetProviderQuery creates a new get provider query
func NewGetProviderQuery(providerRepo providerRepo.ProviderRepository) *GetProviderQuery {
	return &GetProviderQuery{
		providerRepo: providerRepo,
	}
}

// Execute retrieves a provider by namespace and name
// Returns shared.ErrNotFound (wrapped) if provider doesn't exist
// Python reference: /app/terrareg/models.py - ModuleProvider.get()
func (q *GetProviderQuery) Execute(ctx context.Context, namespace, providerName string) (*provider.Provider, error) {
	p, err := q.providerRepo.FindByNamespaceAndName(ctx, namespace, providerName)
	if err != nil {
		return nil, fmt.Errorf("failed to get provider: %w", err)
	}
	if p == nil {
		return nil, fmt.Errorf("provider %s/%s not found", namespace, providerName)
	}
	return p, nil
}
