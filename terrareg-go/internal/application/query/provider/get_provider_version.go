package provider

import (
	"context"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/provider"
	providerRepo "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/provider/repository"
)

// GetProviderVersionQuery handles retrieving a specific provider version
type GetProviderVersionQuery struct {
	providerRepo providerRepo.ProviderRepository
}

// NewGetProviderVersionQuery creates a new get provider version query
func NewGetProviderVersionQuery(providerRepo providerRepo.ProviderRepository) *GetProviderVersionQuery {
	return &GetProviderVersionQuery{
		providerRepo: providerRepo,
	}
}

// Execute retrieves a specific provider version by provider ID and version
func (q *GetProviderVersionQuery) Execute(ctx context.Context, providerID int, version string) (*provider.ProviderVersion, error) {
	return q.providerRepo.FindVersionByProviderAndVersion(ctx, providerID, version)
}