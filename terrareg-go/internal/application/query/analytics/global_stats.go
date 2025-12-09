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
	analyticsRepo     analyticsCmd.AnalyticsRepository
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
		analyticsRepo:     analyticsRepo,
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
	// Get namespace count
	namespaces, err := q.namespaceRepo.List(ctx)
	if err != nil {
		return nil, err
	}

	// Get all module providers to count modules
	moduleProviders, err := q.moduleProviderRepo.Search(ctx, moduleRepo.ModuleSearchQuery{})
	if err != nil {
		return nil, err
	}

	// Count total downloads by summing up all module provider download stats
	totalDownloads := 0

	for _, mp := range moduleProviders.Modules {
		downloads, err := q.analyticsRepo.GetDownloadStats(ctx, mp.Namespace().Name(), mp.Module(), mp.Provider())
		if err == nil {
			totalDownloads += downloads.TotalDownloads
		}
		// If analytics fails for a specific provider, continue without crashing
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
