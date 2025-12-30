package analytics

import (
	"context"

	analyticsCmd "github.com/matthewjohn/terrareg/terrareg-go/internal/application/command/analytics"
	moduleRepo "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/module/repository"
)

// GlobalUsageStatsQuery handles getting global usage statistics
type GlobalUsageStatsQuery struct {
	moduleProviderRepo moduleRepo.ModuleProviderRepository
	analyticsRepo      analyticsCmd.AnalyticsRepository
}

// NewGlobalUsageStatsQuery creates a new global usage stats query
func NewGlobalUsageStatsQuery(
	moduleProviderRepo moduleRepo.ModuleProviderRepository,
	analyticsRepo analyticsCmd.AnalyticsRepository,
) *GlobalUsageStatsQuery {
	return &GlobalUsageStatsQuery{
		moduleProviderRepo: moduleProviderRepo,
		analyticsRepo:      analyticsRepo,
	}
}

// GlobalUsageStats represents global usage statistics
type GlobalUsageStats struct {
	ModuleProviderCount                             int            `json:"module_provider_count"`
	ModuleProviderUsageBreakdownWithAuthToken       map[string]int `json:"module_provider_usage_breakdown_with_auth_token"`
	ModuleProviderUsageCountWithAuthToken           int            `json:"module_provider_usage_count_with_auth_token"`
	ModuleProviderUsageIncludingEmptyAuthToken      map[string]int `json:"module_provider_usage_including_empty_auth_token"`
	ModuleProviderUsageCountIncludingEmptyAuthToken int            `json:"module_provider_usage_count_including_empty_auth_token"`
}

// Execute executes the query
func (q *GlobalUsageStatsQuery) Execute(ctx context.Context) (*GlobalUsageStats, error) {
	// Get all module providers
	moduleProviders, err := q.moduleProviderRepo.Search(ctx, moduleRepo.ModuleSearchQuery{})
	if err != nil {
		return nil, err
	}

	usageWithAuth := make(map[string]int)
	usageIncludingEmpty := make(map[string]int)
	totalWithAuth := 0
	totalIncludingEmpty := 0

	// Get usage stats for each module provider
	for _, mp := range moduleProviders.Modules {
		moduleKey := mp.Namespace().Name() + "/" + mp.Module() + "/" + mp.Provider()

		// Get download stats (this counts total downloads)
		stats, err := q.analyticsRepo.GetDownloadStats(ctx, mp.Namespace().Name(), mp.Module(), mp.Provider())
		if err != nil {
			continue
		}

		// For now, we'll use the same stats for both categories
		// In a full implementation, we would distinguish between authenticated and anonymous downloads
		usageWithAuth[moduleKey] = stats.TotalDownloads
		usageIncludingEmpty[moduleKey] = stats.TotalDownloads
		totalWithAuth += stats.TotalDownloads
		totalIncludingEmpty += stats.TotalDownloads
	}

	result := &GlobalUsageStats{
		ModuleProviderCount:                             len(moduleProviders.Modules),
		ModuleProviderUsageBreakdownWithAuthToken:       usageWithAuth,
		ModuleProviderUsageCountWithAuthToken:           totalWithAuth,
		ModuleProviderUsageIncludingEmptyAuthToken:      usageIncludingEmpty,
		ModuleProviderUsageCountIncludingEmptyAuthToken: totalIncludingEmpty,
	}

	return result, nil
}
