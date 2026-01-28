package module

import (
	"context"
	"errors"
	"fmt"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/module/model"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/module/repository"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/shared"
	types "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/shared/types"
)

// GetModuleProviderQuery handles retrieving a specific module provider
type GetModuleProviderQuery struct {
	moduleProviderRepo repository.ModuleProviderRepository
}

// NewGetModuleProviderQuery creates a new get module provider query
func NewGetModuleProviderQuery(moduleProviderRepo repository.ModuleProviderRepository) *GetModuleProviderQuery {
	return &GetModuleProviderQuery{
		moduleProviderRepo: moduleProviderRepo,
	}
}

// Execute executes the query
func (q *GetModuleProviderQuery) Execute(ctx context.Context, namespace types.NamespaceName, module types.ModuleName, provider types.ModuleProviderName) (*model.ModuleProvider, error) {
	moduleProvider, err := q.moduleProviderRepo.FindByNamespaceModuleProvider(ctx, namespace, module, provider)
	if err != nil {
		if errors.Is(err, shared.ErrNotFound) {
			return nil, fmt.Errorf("module provider %s/%s/%s not found", namespace, module, provider)
		}
		return nil, fmt.Errorf("failed to get module provider: %w", err)
	}

	return moduleProvider, nil
}
