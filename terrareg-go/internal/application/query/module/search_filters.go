package module

import (
	"context"
	"fmt"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/config"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/module/repository"
)

type SearchFiltersQuery struct {
	moduleProviderRepo repository.ModuleProviderRepository
	cfg                *config.Config
}

type SearchFilterCounts struct {
	Verified         int            `json:"verified"`
	TrustedNamespaces int            `json:"trusted_namespaces"`
	Contributed      int            `json:"contributed"`
	Providers        map[string]int `json:"providers"`
	Namespaces       map[string]int `json:"namespaces"`
}

func NewSearchFiltersQuery(moduleProviderRepo repository.ModuleProviderRepository, cfg *config.Config) *SearchFiltersQuery {
	return &SearchFiltersQuery{
		moduleProviderRepo: moduleProviderRepo,
		cfg:                cfg,
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

	// Get verified count
	verifiedQuery := searchQuery
	verifiedQuery.Verified = boolPtr(true)
	verifiedResult, _ := q.moduleProviderRepo.Search(ctx, verifiedQuery)
	counts.Verified = verifiedResult.TotalCount

	// Get trusted namespaces count
	if len(q.cfg.TrustedNamespaces) > 0 {
		trustedQuery := searchQuery
		trustedQuery.TrustedNamespaces = boolPtr(true)
		trustedResult, _ := q.moduleProviderRepo.Search(ctx, trustedQuery)
		counts.TrustedNamespaces = trustedResult.TotalCount

		contributedQuery := searchQuery
		contributedQuery.Contributed = boolPtr(true)
		contributedResult, _ := q.moduleProviderRepo.Search(ctx, contributedQuery)
		counts.Contributed = contributedResult.TotalCount
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