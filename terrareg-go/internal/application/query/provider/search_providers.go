package provider

import (
	"context"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/provider"
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

// Execute searches for providers matching the query
func (q *SearchProvidersQuery) Execute(ctx context.Context, query string, offset, limit int) ([]*provider.Provider, int, error) {
	return q.providerRepo.Search(ctx, query, offset, limit)
}
