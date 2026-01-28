package module

import (
	"context"
	"fmt"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/module/model"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/module/repository"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/shared/types"
)

// SearchModulesQuery handles searching for module providers
type SearchModulesQuery struct {
	// moduleProviderRepo handles module provider persistence (required)
	moduleProviderRepo repository.ModuleProviderRepository
}

// NewSearchModulesQuery creates a new search modules query
// Returns an error if moduleProviderRepo is nil
func NewSearchModulesQuery(moduleProviderRepo repository.ModuleProviderRepository) (*SearchModulesQuery, error) {
	if moduleProviderRepo == nil {
		return nil, fmt.Errorf("moduleProviderRepo cannot be nil")
	}
	return &SearchModulesQuery{
		moduleProviderRepo: moduleProviderRepo,
	}, nil
}

// SearchParams represents search parameters
type SearchParams struct {
	Query                  string
	Namespaces             []types.NamespaceName
	Provider               *string // Keep for backward compatibility
	Providers              []types.ModuleProviderName
	Verified               *bool
	TrustedNamespaces      *bool
	Contributed            *bool
	TargetTerraformVersion *string
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
		providers = append(providers, types.ModuleProviderName(*params.Provider))
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

	// Filter out modules without latest versions (matching Python behavior)
	modulesWithLatestVersion := make([]*model.ModuleProvider, 0)
	for _, module := range result.Modules {
		if module.GetLatestVersion() != nil {
			modulesWithLatestVersion = append(modulesWithLatestVersion, module)
		}
	}

	// Note: We preserve the original TotalCount from the repository for pagination purposes.
	// The TotalCount represents the total number of records matching the query (before LIMIT/OFFSET),
	// not the number of records returned after filtering or after applying LIMIT.
	return &SearchResult{
		Modules:    modulesWithLatestVersion,
		TotalCount: result.TotalCount, // Use the original count from repository for pagination
	}, nil
}
