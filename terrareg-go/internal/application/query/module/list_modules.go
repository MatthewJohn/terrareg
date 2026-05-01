package module

import (
	"context"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/module/model"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/module/repository"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/shared/types"
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

// ListModulesInput represents input parameters for listing modules
type ListModulesInput struct {
	Offset       int
	Limit        int
	Providers    []types.ModuleProviderName
	Verified     *bool
	IncludeCount bool
}

// Execute executes the query to list all module providers with optional filters
// Python reference: /app/test/unit/terrareg/server/test_api_module_list.py
func (q *ListModulesQuery) Execute(ctx context.Context, input ListModulesInput) ([]*model.ModuleProvider, int, error) {
	searchQuery := repository.ModuleSearchQuery{
		Offset:    input.Offset,
		Limit:     input.Limit,
		Providers: input.Providers,
		Verified:  input.Verified,
	}

	// Default limit to 10 if not specified (matching Python behavior)
	if searchQuery.Limit == 0 {
		searchQuery.Limit = 10
	}

	// Execute search
	result, err := q.moduleProviderRepo.Search(ctx, searchQuery)
	if err != nil {
		return nil, 0, err
	}

	// Return count if requested
	count := 0
	if input.IncludeCount {
		count = result.TotalCount
	}

	return result.Modules, count, nil
}
