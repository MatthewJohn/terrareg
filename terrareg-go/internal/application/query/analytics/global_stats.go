package analytics

import (
	"context"

	analyticsCmd "github.com/matthewjohn/terrareg/terrareg-go/internal/application/command/analytics"
	moduleRepo "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/module/repository"
)

// GlobalStatsQuery handles getting global statistics
type GlobalStatsQuery struct {
	namespaceRepo      moduleRepo.NamespaceRepository
	moduleProviderRepo moduleRepo.ModuleProviderRepository
	analyticsRepo      analyticsCmd.AnalyticsRepository
}

// NewGlobalStatsQuery creates a new global stats query
func NewGlobalStatsQuery(
	namespaceRepo moduleRepo.NamespaceRepository,
	moduleProviderRepo moduleRepo.ModuleProviderRepository,
	analyticsRepo analyticsCmd.AnalyticsRepository,
) *GlobalStatsQuery {
	return &GlobalStatsQuery{
		namespaceRepo:      namespaceRepo,
		moduleProviderRepo: moduleProviderRepo,
		analyticsRepo:      analyticsRepo,
	}
}

// GlobalStats represents global statistics
type GlobalStats struct {
	Namespaces     int `json:"namespaces"`
	Modules        int `json:"modules"`
	ModuleVersions int `json:"module_versions"`
	Downloads      int `json:"downloads"`
}

// Execute executes the query
func (q *GlobalStatsQuery) Execute(ctx context.Context) (*GlobalStats, error) {
	// Get namespace count (no pagination needed)
	namespaces, _, err := q.namespaceRepo.List(ctx, nil)
	if err != nil {
		return nil, err
	}

	// Get all module providers to count modules
	moduleProviders, err := q.moduleProviderRepo.Search(ctx, moduleRepo.ModuleSearchQuery{})
	if err != nil {
		return nil, err
	}

	// Get total downloads using the GetTotalDownloads method
	// This matches Python's AnalyticsEngine.get_total_downloads()
	totalDownloads, err := q.analyticsRepo.GetTotalDownloads(ctx)
	if err != nil {
		return nil, err
	}

	// Count module versions by getting versions from each module provider
	moduleVersions := 0
	for _, mp := range moduleProviders.Modules {
		versions := mp.GetAllVersions()
		moduleVersions += len(versions)
	}

	stats := &GlobalStats{
		Namespaces:     len(namespaces),
		Modules:        len(moduleProviders.Modules),
		ModuleVersions: moduleVersions,
		Downloads:      totalDownloads,
	}

	return stats, nil
}
