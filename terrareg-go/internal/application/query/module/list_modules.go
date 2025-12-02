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

// Execute executes the query to list all module providers
func (q *ListModulesQuery) Execute(ctx context.Context) ([]*model.ModuleProvider, error) {
	// For simplicity, this lists all module providers.
	// In a real scenario, this might take filters or pagination.
	// The repository.ModuleSearchQuery could be used here.
	result, err := q.moduleProviderRepo.Search(ctx, repository.ModuleSearchQuery{})
	if err != nil {
		return nil, err
	}
	return result.Modules, nil
}