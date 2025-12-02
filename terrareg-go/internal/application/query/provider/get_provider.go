package provider

import (
	"context"

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
func (q *GetProviderQuery) Execute(ctx context.Context, namespace, providerName string) (*provider.Provider, error) {
	return q.providerRepo.FindByNamespaceAndName(ctx, namespace, providerName)
}
