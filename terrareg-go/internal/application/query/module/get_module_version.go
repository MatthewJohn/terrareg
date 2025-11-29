package module

import (
	"context"
	"errors"
	"fmt"

	"github.com/terrareg/terrareg/internal/domain/module/model"
	"github.com/terrareg/terrareg/internal/domain/module/repository"
	"github.com/terrareg/terrareg/internal/domain/shared"
)

// GetModuleVersionQuery handles retrieving a specific module version
type GetModuleVersionQuery struct {
	moduleProviderRepo repository.ModuleProviderRepository
}

// NewGetModuleVersionQuery creates a new get module version query
func NewGetModuleVersionQuery(moduleProviderRepo repository.ModuleProviderRepository) *GetModuleVersionQuery {
	return &GetModuleVersionQuery{
		moduleProviderRepo: moduleProviderRepo,
	}
}

// Execute executes the query
func (q *GetModuleVersionQuery) Execute(ctx context.Context, namespace, module, provider, version string) (*model.ModuleVersion, error) {
	// First get the module provider
	moduleProvider, err := q.moduleProviderRepo.FindByNamespaceModuleProvider(ctx, namespace, module, provider)
	if err != nil {
		if errors.Is(err, shared.ErrNotFound) {
			return nil, fmt.Errorf("module provider %s/%s/%s not found", namespace, module, provider)
		}
		return nil, fmt.Errorf("failed to get module provider: %w", err)
	}

	// Get the specific version
	moduleVersion, err := moduleProvider.GetVersion(version)
	if err != nil || moduleVersion == nil {
		return nil, fmt.Errorf("version %s not found for %s/%s/%s", version, namespace, module, provider)
	}

	return moduleVersion, nil
}
