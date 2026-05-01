package selenium

import (
	"context"

	analyticsCmd "github.com/matthewjohn/terrareg/terrareg-go/internal/application/command/analytics"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/shared/types"
)

// MockAnalyticsRepository is a mock implementation of AnalyticsRepository
// It uses a decorator pattern to wrap the real repository and only override
// download-related methods, matching Python's behavior of mocking get_total_downloads
type MockAnalyticsRepository struct {
	// Total downloads to return (default: 2005 to match Python)
	TotalDownloads int
	// Real repository for non-mocked methods
	realRepo analyticsCmd.AnalyticsRepository
}

// NewMockAnalyticsRepository creates a mock with the specified total downloads
func NewMockAnalyticsRepository(totalDownloads int, realRepo analyticsCmd.AnalyticsRepository) *MockAnalyticsRepository {
	return &MockAnalyticsRepository{
		TotalDownloads: totalDownloads,
		realRepo:       realRepo,
	}
}

// RecordDownload delegates to real repo
func (m *MockAnalyticsRepository) RecordDownload(ctx context.Context, event analyticsCmd.AnalyticsEvent) error {
	return m.realRepo.RecordDownload(ctx, event)
}

// RecordProviderDownload delegates to real repo
func (m *MockAnalyticsRepository) RecordProviderDownload(ctx context.Context, event analyticsCmd.ProviderDownloadEvent) error {
	return m.realRepo.RecordProviderDownload(ctx, event)
}

// GetDownloadStats returns mocked download stats
// This is the key mocked method matching Python's get_total_downloads mock
// Matches Python: AnalyticsEngine.get_module_provider_download_stats()
func (m *MockAnalyticsRepository) GetDownloadStats(ctx context.Context, namespace types.NamespaceName, module types.ModuleName, provider types.ModuleProviderName) (*analyticsCmd.DownloadStats, error) {
	// Return mocked download stats matching Python format
	// Python returns: week, month, year, total
	total := m.TotalDownloads
	// Approximate stats: week ~ 1/20 of total, month ~ 1/4 of total, year ~ 1/2 of total
	week := total / 20
	if week < 1 {
		week = 1
	}
	month := total / 4
	if month < 1 {
		month = 1
	}
	year := total / 2
	if year < 1 {
		year = 1
	}
	return &analyticsCmd.DownloadStats{
		TotalDownloads: total,
		Week:           week,
		Month:          month,
		Year:           year,
	}, nil
}

// GetTotalDownloads returns the mocked total downloads count
// This matches Python's AnalyticsEngine.get_total_downloads() mock
func (m *MockAnalyticsRepository) GetTotalDownloads(ctx context.Context) (int, error) {
	return m.TotalDownloads, nil
}

// GetDownloadsByVersionID returns mocked download count
func (m *MockAnalyticsRepository) GetDownloadsByVersionID(ctx context.Context, moduleVersionID int) (int, error) {
	downloads := m.TotalDownloads / 100
	if downloads < 1 {
		downloads = 1
	}
	return downloads, nil // Distribute across versions, no error
}

// GetMostRecentlyPublished delegates to real repo
func (m *MockAnalyticsRepository) GetMostRecentlyPublished(ctx context.Context) (*analyticsCmd.ModuleVersionInfo, error) {
	return m.realRepo.GetMostRecentlyPublished(ctx)
}

// GetMostDownloadedThisWeek delegates to real repo
func (m *MockAnalyticsRepository) GetMostDownloadedThisWeek(ctx context.Context) (*analyticsCmd.ModuleProviderInfo, error) {
	return m.realRepo.GetMostDownloadedThisWeek(ctx)
}

// GetModuleProviderID delegates to real repo
func (m *MockAnalyticsRepository) GetModuleProviderID(ctx context.Context, namespace types.NamespaceName, module types.ModuleName, provider types.ModuleProviderName) (int, error) {
	return m.realRepo.GetModuleProviderID(ctx, namespace, module, provider)
}

// GetLatestTokenVersions delegates to real repo
func (m *MockAnalyticsRepository) GetLatestTokenVersions(ctx context.Context, moduleProviderID int) (map[string]analyticsCmd.TokenVersionInfo, error) {
	return m.realRepo.GetLatestTokenVersions(ctx, moduleProviderID)
}
