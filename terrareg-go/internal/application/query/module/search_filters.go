package module

import (
	"context"
	"fmt"

	configModel "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/config/model"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/module/repository"
)

type SearchFiltersQuery struct {
	moduleProviderRepo repository.ModuleProviderRepository
	domainConfig       *configModel.DomainConfig
}

type SearchFilterCounts struct {
	Verified         int            `json:"verified"`
	TrustedNamespaces int            `json:"trusted_namespaces"`
	Contributed      int            `json:"contributed"`
	Providers        map[string]int `json:"providers"`
	Namespaces       map[string]int `json:"namespaces"`
}

func NewSearchFiltersQuery(moduleProviderRepo repository.ModuleProviderRepository, domainConfig *configModel.DomainConfig) *SearchFiltersQuery {
	return &SearchFiltersQuery{
		moduleProviderRepo: moduleProviderRepo,
		domainConfig:       domainConfig,
	}
}

func (q *SearchFiltersQuery) Execute(ctx context.Context, queryString string) (*SearchFilterCounts, error) {
	// Execute base search query
	searchQuery := repository.ModuleSearchQuery{
		Query:  queryString,
		Limit:  0, // No limit for counts
		Offset: 0,
	}

	result, err := q.moduleProviderRepo.Search(ctx, searchQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to execute base search: %w", err)
	}

	// Count with filters
	counts := &SearchFilterCounts{
		Providers:  make(map[string]int),
		Namespaces: make(map[string]int),
	}

	// Get verified count - apply same filtering as main search (modules with latest versions only)
	verifiedQuery := searchQuery
	verifiedQuery.Verified = boolPtr(true)
	verifiedResult, _ := q.moduleProviderRepo.Search(ctx, verifiedQuery)
	// Filter out modules without latest versions
	verifiedModulesWithLatestVersion := 0
	for _, module := range verifiedResult.Modules {
		if module.GetLatestVersion() != nil {
			verifiedModulesWithLatestVersion++
		}
	}
	counts.Verified = verifiedModulesWithLatestVersion

	// Get trusted namespaces count
	if len(q.domainConfig.TrustedNamespaces) > 0 {
		trustedQuery := searchQuery
		trustedQuery.TrustedNamespaces = boolPtr(true)
		trustedResult, _ := q.moduleProviderRepo.Search(ctx, trustedQuery)
		// Filter out modules without latest versions
		trustedModulesWithLatestVersion := 0
		for _, module := range trustedResult.Modules {
			if module.GetLatestVersion() != nil {
				trustedModulesWithLatestVersion++
			}
		}
		counts.TrustedNamespaces = trustedModulesWithLatestVersion

		contributedQuery := searchQuery
		contributedQuery.Contributed = boolPtr(true)
		contributedResult, _ := q.moduleProviderRepo.Search(ctx, contributedQuery)
		// Filter out modules without latest versions
		contributedModulesWithLatestVersion := 0
		for _, module := range contributedResult.Modules {
			if module.GetLatestVersion() != nil {
				contributedModulesWithLatestVersion++
			}
		}
		counts.Contributed = contributedModulesWithLatestVersion
	}

	// Count providers and namespaces
	for _, module := range result.Modules {
		provider := module.Provider()
		counts.Providers[provider]++

		namespace := module.Namespace().Name()
		counts.Namespaces[namespace]++
	}

	return counts, nil
}

func boolPtr(b bool) *bool {
	return &b
}