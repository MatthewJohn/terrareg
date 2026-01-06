package provider

import (
	"context"

	providerRepo "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/provider/repository"
)

// SearchProvidersQuery handles searching for providers
type SearchProvidersQuery struct {
	providerRepo providerRepo.ProviderRepository
}

// NewSearchProvidersQuery creates a new search providers query
func NewSearchProvidersQuery(providerRepo providerRepo.ProviderRepository) *SearchProvidersQuery {
	return &SearchProvidersQuery{
		providerRepo: providerRepo,
	}
}

// Execute searches for providers matching the query with filters
func (q *SearchProvidersQuery) Execute(ctx context.Context, params providerRepo.ProviderSearchQuery) (*providerRepo.ProviderSearchResult, error) {
	return q.providerRepo.Search(ctx, params)
}
