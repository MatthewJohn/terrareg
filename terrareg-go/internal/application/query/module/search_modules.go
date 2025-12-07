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
	Query                  string
	Namespaces             []string // Change from *string to []string for multiple values
	Provider               *string  // Keep for backward compatibility
	Providers              []string // New: Multiple provider support
	Verified               *bool
	TrustedNamespaces      *bool   // New: Filter for trusted namespaces only
	Contributed            *bool   // New: Filter for contributed modules only
	TargetTerraformVersion *string // New: Check compatibility with specific Terraform version
	Limit                  int
	Offset                 int
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

	// Handle backward compatibility - if single Provider is specified, add to Providers array
	providers := params.Providers
	if params.Provider != nil {
		providers = append(providers, *params.Provider)
	}

	searchQuery := repository.ModuleSearchQuery{
		Query:                  params.Query,
		Namespaces:             params.Namespaces,
		Providers:              providers,
		Verified:               params.Verified,
		TrustedNamespaces:      params.TrustedNamespaces,
		Contributed:            params.Contributed,
		TargetTerraformVersion: params.TargetTerraformVersion,
		Limit:                  params.Limit,
		Offset:                 params.Offset,
		OrderBy:                "id", // Default ordering
		OrderDir:               "DESC",
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
