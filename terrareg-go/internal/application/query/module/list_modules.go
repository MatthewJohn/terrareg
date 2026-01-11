package module

import (
	"context"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/module/model"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/module/repository"
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
func (q *ListModulesQuery) Execute(ctx context.Context, namespace ...string) ([]*model.ModuleProvider, error) {
	// For simplicity, this lists all module providers.
	// In a real scenario, this might take filters or pagination.
	// The repository.ModuleSearchQuery could be used here.
	searchQuery := repository.ModuleSearchQuery{}
	if len(namespace) > 0 && namespace[0] != "" {
		searchQuery.Namespaces = []string{namespace[0]}
	}
	result, err := q.moduleProviderRepo.Search(ctx, searchQuery)
	if err != nil {
		return nil, err
	}
	return result.Modules, nil
}
