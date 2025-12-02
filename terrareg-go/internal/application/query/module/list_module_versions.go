package module

import (
	"context"
	"errors"
	"fmt"

	"github.com/terrareg/terrareg/internal/domain/module/model"
	"github.com/terrareg/terrareg/internal/domain/module/repository"
	"github.com/terrareg/terrareg/internal/domain/shared"
)

// ListModuleVersionsQuery handles retrieving all versions for a specific module provider
type ListModuleVersionsQuery struct {
	moduleProviderRepo repository.ModuleProviderRepository
}

// NewListModuleVersionsQuery creates a new list module versions query
func NewListModuleVersionsQuery(moduleProviderRepo repository.ModuleProviderRepository) *ListModuleVersionsQuery {
	return &ListModuleVersionsQuery{
		moduleProviderRepo: moduleProviderRepo,
	}
}

// Execute executes the query
func (q *ListModuleVersionsQuery) Execute(ctx context.Context, namespace, module, provider string) ([]*model.ModuleVersion, error) {
	moduleProvider, err := q.moduleProviderRepo.FindByNamespaceModuleProvider(ctx, namespace, module, provider)
	if err != nil {
		if errors.Is(err, shared.ErrNotFound) {
			return nil, fmt.Errorf("module provider %s/%s/%s not found", namespace, module, provider)
		}
		return nil, fmt.Errorf("failed to get module provider for versions: %w", err)
	}

	return moduleProvider.GetAllVersions(), nil
}
