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
	domainConfig   *configModel.DomainConfig
}

// SearchFilterCounts represents the filter counts for provider search
type SearchFilterCounts struct {
	Contributed       int            `json:"contributed"`
	TrustedNamespaces int            `json:"trusted_namespaces"`
	Namespaces        map[string]int `json:"namespaces"`
	ProviderCategories map[string]int `json:"provider_categories"`
}

// NewSearchFiltersQuery creates a new search filters query
func NewSearchFiltersQuery(providerRepo providerRepo.ProviderRepository, namespaceRepo repository.NamespaceRepository, domainConfig *configModel.DomainConfig) *SearchFiltersQuery {
	return &SearchFiltersQuery{
		providerRepo:  providerRepo,
		namespaceRepo: namespaceRepo,
		domainConfig:   domainConfig,
	}
}

// Execute gets search filter counts for the given query string
func (q *SearchFiltersQuery) Execute(ctx context.Context, queryString string) (*SearchFilterCounts, error) {
	// Get all results (no offset/limit for counts)
	allProviders, _, err := q.providerRepo.Search(ctx, queryString, 0, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to execute base search: %w", err)
	}

	counts := &SearchFilterCounts{
		Namespaces:        make(map[string]int),
		ProviderCategories: make(map[string]int),
	}

	// Build a map of namespace IDs to namespace names
	namespaceNames := make(map[int]string)
	for _, p := range allProviders {
		nsID := p.NamespaceID()
		if _, exists := namespaceNames[nsID]; !exists {
			// Look up namespace by ID
			ns, err := q.namespaceRepo.FindByID(ctx, nsID)
			if err == nil && ns != nil {
				namespaceNames[nsID] = ns.Name()
			}
		}
	}

	// Get trusted namespaces count
	trustedCount := 0
	if len(q.domainConfig.TrustedNamespaces) > 0 {
		for _, p := range allProviders {
			if nsName, ok := namespaceNames[p.NamespaceID()]; ok {
				if q.isTrustedNamespace(nsName) {
					trustedCount++
				}
			}
		}
	}
	counts.TrustedNamespaces = trustedCount

	// Get contributed count (not in trusted namespaces)
	// When there are no trusted namespaces configured, all providers are contributed
	contributedCount := 0
	for _, p := range allProviders {
		if nsName, ok := namespaceNames[p.NamespaceID()]; ok {
			if !q.isTrustedNamespace(nsName) {
				contributedCount++
			}
		}
	}
	counts.Contributed = contributedCount

	// Count namespaces and provider categories
	for _, p := range allProviders {
		if nsName, ok := namespaceNames[p.NamespaceID()]; ok {
			counts.Namespaces[nsName]++
		}

		// Count provider categories (if provider has a category)
		if p.CategoryID() != nil {
			// Use category name from provider category association
			// For now, use a placeholder since category details would need to be loaded
			categoryName := fmt.Sprintf("category-%d", *p.CategoryID())
			counts.ProviderCategories[categoryName]++
		}
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
