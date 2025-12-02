package module

import (
	"context"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/module/model"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/module/repository"
)

// SearchModulesQuery handles searching for module providers
type SearchModulesQuery struct {
	moduleProviderRepo repository.ModuleProviderRepository
}

// NewSearchModulesQuery creates a new search modules query
func NewSearchModulesQuery(moduleProviderRepo repository.ModuleProviderRepository) *SearchModulesQuery {
	return &SearchModulesQuery{
		moduleProviderRepo: moduleProviderRepo,
	}
}

// SearchParams represents search parameters
type SearchParams struct {
	Query     string
	Namespace *string
	Provider  *string
	Verified  *bool
	Limit     int
	Offset    int
}

// SearchResult represents search results with pagination
type SearchResult struct {
	Modules    []*model.ModuleProvider
	TotalCount int
}

// Execute executes the search query
func (q *SearchModulesQuery) Execute(ctx context.Context, params SearchParams) (*SearchResult, error) {
	// Set defaults
	if params.Limit == 0 {
		params.Limit = 20
	}

	searchQuery := repository.ModuleSearchQuery{
		Query:     params.Query,
		Namespace: params.Namespace,
		Provider:  params.Provider,
		Verified:  params.Verified,
		Limit:     params.Limit,
		Offset:    params.Offset,
		OrderBy:   "id", // Default ordering
		OrderDir:  "DESC",
	}

	result, err := q.moduleProviderRepo.Search(ctx, searchQuery)
	if err != nil {
		return nil, err
	}

	return &SearchResult{
		Modules:    result.Modules,
		TotalCount: result.TotalCount,
	}, nil
}
