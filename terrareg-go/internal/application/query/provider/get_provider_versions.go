package provider

import (
	"context"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/provider"
	providerRepo "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/provider/repository"
)

// GetProviderVersionsQuery handles retrieving provider versions
type GetProviderVersionsQuery struct {
	providerRepo providerRepo.ProviderRepository
}

// NewGetProviderVersionsQuery creates a new get provider versions query
func NewGetProviderVersionsQuery(providerRepo providerRepo.ProviderRepository) *GetProviderVersionsQuery {
	return &GetProviderVersionsQuery{
		providerRepo: providerRepo,
	}
}

// Execute retrieves all versions for a provider
func (q *GetProviderVersionsQuery) Execute(ctx context.Context, providerID int) ([]*provider.ProviderVersion, error) {
	return q.providerRepo.FindVersionsByProvider(ctx, providerID)
}
