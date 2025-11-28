package module

import (
	"context"

	"github.com/terrareg/terrareg/internal/domain/module/model"
	"github.com/terrareg/terrareg/internal/domain/module/repository"
)

// ListModulesQuery handles listing all module providers
type ListModulesQuery struct {
	moduleProviderRepo repository.ModuleProviderRepository
}

// NewListModulesQuery creates a new list modules query
func NewListModulesQuery(moduleProviderRepo repository.ModuleProviderRepository) *ListModulesQuery {
	return &ListModulesQuery{
		moduleProviderRepo: moduleProviderRepo,
	}
}

// Execute executes the query with optional namespace filter
func (q *ListModulesQuery) Execute(ctx context.Context, namespace string) ([]*model.ModuleProvider, error) {
	if namespace != "" {
		return q.moduleProviderRepo.FindByNamespace(ctx, namespace)
	}

	// For listing all modules, we can use search with empty query
	result, err := q.moduleProviderRepo.Search(ctx, repository.ModuleSearchQuery{
		Limit:  100, // Default limit
		Offset: 0,
	})
	if err != nil {
		return nil, err
	}

	return result.Modules, nil
}
