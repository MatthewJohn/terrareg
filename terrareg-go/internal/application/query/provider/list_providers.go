package provider

import (
	"context"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/provider"
	providerRepo "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/provider/repository"
)

// ListProvidersQuery handles listing all providers
type ListProvidersQuery struct {
	providerRepo providerRepo.ProviderRepository
}

// NewListProvidersQuery creates a new list providers query
func NewListProvidersQuery(providerRepo providerRepo.ProviderRepository) *ListProvidersQuery {
	return &ListProvidersQuery{
		providerRepo: providerRepo,
	}
}

// Execute retrieves all providers with pagination, including namespace names and version data
func (q *ListProvidersQuery) Execute(ctx context.Context, offset, limit int) ([]*provider.Provider, map[int]string, map[int]providerRepo.VersionData, int, error) {
	return q.providerRepo.FindAll(ctx, offset, limit)
}
