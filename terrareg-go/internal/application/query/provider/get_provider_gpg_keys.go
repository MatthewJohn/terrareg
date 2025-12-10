package provider

import (
	"context"
	"fmt"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/provider"
	providerRepo "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/provider/repository"
	namespaceRepo "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/module/repository"
)

// GetProviderGPGKeysQuery handles retrieving GPG keys for a provider
type GetProviderGPGKeysQuery struct {
	providerRepo  providerRepo.ProviderRepository
	namespaceRepo namespaceRepo.NamespaceRepository
}

// NewGetProviderGPGKeysQuery creates a new get provider GPG keys query
func NewGetProviderGPGKeysQuery(
	providerRepo providerRepo.ProviderRepository,
	namespaceRepo namespaceRepo.NamespaceRepository,
) *GetProviderGPGKeysQuery {
	return &GetProviderGPGKeysQuery{
		providerRepo:  providerRepo,
		namespaceRepo: namespaceRepo,
	}
}

// Execute retrieves all GPG keys for a provider
func (q *GetProviderGPGKeysQuery) Execute(ctx context.Context, namespace, providerName string) ([]*provider.GPGKey, error) {
	// Get provider
	_, err := q.providerRepo.FindByNamespaceAndName(ctx, namespace, providerName)
	if err != nil {
		return nil, fmt.Errorf("provider not found: %w", err)
	}

	// Return GPG keys from provider domain model
	// TODO: Implement when GPG keys are properly stored in provider domain model
	// For now, return empty slice
	return []*provider.GPGKey{}, nil
}