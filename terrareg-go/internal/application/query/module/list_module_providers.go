package module

import (
	"context"
	"fmt"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/module/model"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/module/repository"
)

// ListModuleProvidersQuery handles listing all providers for a specific module
type ListModuleProvidersQuery struct {
	moduleProviderRepo repository.ModuleProviderRepository
}

// NewListModuleProvidersQuery creates a new list module providers query
func NewListModuleProvidersQuery(moduleProviderRepo repository.ModuleProviderRepository) *ListModuleProvidersQuery {
	return &ListModuleProvidersQuery{
		moduleProviderRepo: moduleProviderRepo,
	}
}

// Execute executes the query
func (q *ListModuleProvidersQuery) Execute(ctx context.Context, namespace, module string) ([]*model.ModuleProvider, error) {
	// Build search query to find all providers for this namespace/module combination
	searchQuery := repository.ModuleSearchQuery{
		Namespaces: []string{namespace},
		Module:     &module,
		Limit:      1000, // Get all providers
		Offset:     0,
		OrderBy:    "provider",
		OrderDir:   "ASC",
	}

	result, err := q.moduleProviderRepo.Search(ctx, searchQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to list module providers: %w", err)
	}

	return result.Modules, nil
}
