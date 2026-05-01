package provider

import (
	"context"
	"fmt"

	configModel "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/config/model"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/module/repository"
	providerRepo "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/provider/repository"
)

// SearchFiltersQuery handles getting search filter counts for providers
type SearchFiltersQuery struct {
	providerRepo  providerRepo.ProviderRepository
	namespaceRepo repository.NamespaceRepository
	domainConfig  *configModel.DomainConfig
}

// SearchFilterCounts represents the filter counts for provider search
type SearchFilterCounts struct {
	Contributed        int            `json:"contributed"`
	TrustedNamespaces  int            `json:"trusted_namespaces"`
	Namespaces         map[string]int `json:"namespaces"`
	ProviderCategories map[string]int `json:"provider_categories"`
}

// NewSearchFiltersQuery creates a new search filters query
func NewSearchFiltersQuery(providerRepo providerRepo.ProviderRepository, namespaceRepo repository.NamespaceRepository, domainConfig *configModel.DomainConfig) *SearchFiltersQuery {
	return &SearchFiltersQuery{
		providerRepo:  providerRepo,
		namespaceRepo: namespaceRepo,
		domainConfig:  domainConfig,
	}
}

// Execute gets search filter counts for the given query string
func (q *SearchFiltersQuery) Execute(ctx context.Context, queryString string) (*SearchFilterCounts, error) {
	// Get search filters from repository, passing trusted namespaces from domain config
	filters, err := q.providerRepo.GetSearchFilters(ctx, queryString, q.domainConfig.TrustedNamespaces)
	if err != nil {
		return nil, fmt.Errorf("failed to get search filters: %w", err)
	}

	counts := &SearchFilterCounts{
		Namespaces:         make(map[string]int),
		ProviderCategories: make(map[string]int),
	}

	// Copy all counts from filters
	counts.TrustedNamespaces = filters.TrustedNamespaces
	counts.Contributed = filters.Contributed

	// Copy namespace and category counts from filters
	for ns, count := range filters.Namespaces {
		counts.Namespaces[ns] = count
	}
	for cat, count := range filters.ProviderCategories {
		counts.ProviderCategories[cat] = count
	}

	return counts, nil
}

// isTrustedNamespace checks if a namespace is in the trusted namespaces list
func (q *SearchFiltersQuery) isTrustedNamespace(namespace string) bool {
	for _, trusted := range q.domainConfig.TrustedNamespaces {
		if trusted == namespace {
			return true
		}
	}
	return false
}
