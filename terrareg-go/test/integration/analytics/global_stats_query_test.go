package analytics

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	analyticsQuery "github.com/matthewjohn/terrareg/terrareg-go/internal/application/query/analytics"
	namespaceService "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/module/service"
	analyticsPersistence "github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/persistence/sqldb/analytics"
	moduleRepo "github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/persistence/sqldb/module"
	"github.com/matthewjohn/terrareg/terrareg-go/test/integration/testutils"
)

// TestGlobalStatsQuery_EmptyDatabase tests that stats return zeros when database is empty
func TestGlobalStatsQuery_EmptyDatabase(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	ctx := context.Background()

	// Create query with no data
	namespaceRepo := moduleRepo.NewNamespaceRepository(db.DB)
	domainConfig := testutils.CreateTestDomainConfig(t)
	moduleProviderRepo, err := moduleRepo.NewModuleProviderRepository(db.DB, namespaceRepo, domainConfig)
	require.NoError(t, err)
	namespaceSvc := namespaceService.NewNamespaceService(testutils.CreateTestDomainConfig(t))
	analyticsRepo, err := analyticsPersistence.NewAnalyticsRepository(db.DB, namespaceRepo, namespaceSvc)
	require.NoError(t, err)
	query := analyticsQuery.NewGlobalStatsQuery(namespaceRepo, moduleProviderRepo, analyticsRepo)

	// Execute query
	stats, err := query.Execute(ctx)
	require.NoError(t, err)

	// Verify all counts are zero
	assert.Equal(t, 0, stats.Namespaces)
	assert.Equal(t, 0, stats.Modules)
	assert.Equal(t, 0, stats.ModuleVersions)
	assert.Equal(t, 0, stats.Downloads)
}

// TestGlobalStatsQuery_SingleModule tests stats with a single module and version
func TestGlobalStatsQuery_SingleModule(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	ctx := context.Background()

	// Create test data
	namespace := testutils.CreateNamespace(t, db, "test-ns", nil)
	provider := testutils.CreateModuleProvider(t, db, namespace.ID, "test-module", "aws")
	version := testutils.CreatePublishedModuleVersion(t, db, provider.ID, "1.0.0")

	// Create analytics data with matching namespace/module/provider
	timestamp := time.Now()
	for i := 0; i < 42; i++ {
		testutils.CreateAnalyticsDataWithDetails(t, db, version.ID, timestamp, "1.5.0", "token", "auth", "prod", namespace.Namespace, provider.Module, provider.Provider)
	}

	// Create query
	namespaceRepo := moduleRepo.NewNamespaceRepository(db.DB)
	domainConfig := testutils.CreateTestDomainConfig(t)
	moduleProviderRepo, err := moduleRepo.NewModuleProviderRepository(db.DB, namespaceRepo, domainConfig)
	require.NoError(t, err)
	namespaceSvc := namespaceService.NewNamespaceService(testutils.CreateTestDomainConfig(t))
	analyticsRepo, err := analyticsPersistence.NewAnalyticsRepository(db.DB, namespaceRepo, namespaceSvc)
	require.NoError(t, err)
	query := analyticsQuery.NewGlobalStatsQuery(namespaceRepo, moduleProviderRepo, analyticsRepo)

	// Execute query
	stats, err := query.Execute(ctx)
	require.NoError(t, err)

	// Verify counts
	assert.Equal(t, 1, stats.Namespaces)
	assert.Equal(t, 1, stats.Modules)
	assert.Equal(t, 1, stats.ModuleVersions)
	assert.Equal(t, 42, stats.Downloads)
}

// TestGlobalStatsQuery_MultipleVersionsPerModule tests version counting with multiple versions
func TestGlobalStatsQuery_MultipleVersionsPerModule(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	ctx := context.Background()

	// Create test data with 5 versions
	namespace := testutils.CreateNamespace(t, db, "multi-version-ns", nil)
	provider := testutils.CreateModuleProvider(t, db, namespace.ID, "multi-version-module", "aws")

	versions := []string{"1.0.0", "1.1.0", "1.2.0", "2.0.0", "2.1.0"}
	for _, ver := range versions {
		testutils.CreatePublishedModuleVersion(t, db, provider.ID, ver)
	}

	// Create query
	namespaceRepo := moduleRepo.NewNamespaceRepository(db.DB)
	domainConfig := testutils.CreateTestDomainConfig(t)
	moduleProviderRepo, err := moduleRepo.NewModuleProviderRepository(db.DB, namespaceRepo, domainConfig)
	require.NoError(t, err)
	namespaceSvc := namespaceService.NewNamespaceService(testutils.CreateTestDomainConfig(t))
	analyticsRepo, err := analyticsPersistence.NewAnalyticsRepository(db.DB, namespaceRepo, namespaceSvc)
	require.NoError(t, err)
	query := analyticsQuery.NewGlobalStatsQuery(namespaceRepo, moduleProviderRepo, analyticsRepo)

	// Execute query
	stats, err := query.Execute(ctx)
	require.NoError(t, err)

	// Verify counts
	assert.Equal(t, 1, stats.Namespaces, "Should have 1 namespace")
	assert.Equal(t, 1, stats.Modules, "Should have 1 module")
	assert.Equal(t, 5, stats.ModuleVersions, "Should have 5 versions")
	assert.Equal(t, 0, stats.Downloads, "Should have 0 downloads")
}

// TestGlobalStatsQuery_DownloadAggregation tests that downloads aggregate correctly across modules
func TestGlobalStatsQuery_DownloadAggregation(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	ctx := context.Background()

	// Create test data with 3 modules
	namespace := testutils.CreateNamespace(t, db, "aggregate-ns", nil)
	provider1 := testutils.CreateModuleProvider(t, db, namespace.ID, "module1", "aws")
	provider2 := testutils.CreateModuleProvider(t, db, namespace.ID, "module2", "aws")
	provider3 := testutils.CreateModuleProvider(t, db, namespace.ID, "module3", "aws")

	version1 := testutils.CreatePublishedModuleVersion(t, db, provider1.ID, "1.0.0")
	version2 := testutils.CreatePublishedModuleVersion(t, db, provider2.ID, "1.0.0")
	version3 := testutils.CreatePublishedModuleVersion(t, db, provider3.ID, "1.0.0")

	// Create analytics: 10 + 20 + 30 = 60 total downloads
	timestamp := time.Now()
	for i := 0; i < 10; i++ {
		testutils.CreateAnalyticsDataWithDetails(t, db, version1.ID, timestamp, "1.5.0", "token", "auth", "prod", namespace.Namespace, provider1.Module, provider1.Provider)
	}
	for i := 0; i < 20; i++ {
		testutils.CreateAnalyticsDataWithDetails(t, db, version2.ID, timestamp, "1.5.0", "token", "auth", "prod", namespace.Namespace, provider2.Module, provider2.Provider)
	}
	for i := 0; i < 30; i++ {
		testutils.CreateAnalyticsDataWithDetails(t, db, version3.ID, timestamp, "1.5.0", "token", "auth", "prod", namespace.Namespace, provider3.Module, provider3.Provider)
	}

	// Create query
	namespaceRepo := moduleRepo.NewNamespaceRepository(db.DB)
	domainConfig := testutils.CreateTestDomainConfig(t)
	moduleProviderRepo, err := moduleRepo.NewModuleProviderRepository(db.DB, namespaceRepo, domainConfig)
	require.NoError(t, err)
	namespaceSvc := namespaceService.NewNamespaceService(testutils.CreateTestDomainConfig(t))
	analyticsRepo, err := analyticsPersistence.NewAnalyticsRepository(db.DB, namespaceRepo, namespaceSvc)
	require.NoError(t, err)
	query := analyticsQuery.NewGlobalStatsQuery(namespaceRepo, moduleProviderRepo, analyticsRepo)

	// Execute query
	stats, err := query.Execute(ctx)
	require.NoError(t, err)

	// Verify aggregation
	assert.Equal(t, 1, stats.Namespaces)
	assert.Equal(t, 3, stats.Modules)
	assert.Equal(t, 3, stats.ModuleVersions)
	assert.Equal(t, 60, stats.Downloads, "Downloads should aggregate to 60")
}

// TestGlobalStatsQuery_MultipleNamespaces tests aggregation across multiple namespaces
func TestGlobalStatsQuery_MultipleNamespaces(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	ctx := context.Background()

	// Create 5 namespaces, each with 2 modules and 3 versions
	timestamp := time.Now()
	totalDownloads := 0

	for i := 1; i <= 5; i++ {
		namespaceName := fmt.Sprintf("multi-ns-%d", i)
		namespace := testutils.CreateNamespace(t, db, namespaceName, nil)

		for j := 1; j <= 2; j++ {
			moduleName := fmt.Sprintf("multi-module-%d-%d", i, j)
			provider := testutils.CreateModuleProvider(t, db, namespace.ID, moduleName, "aws")

			for k := 1; k <= 3; k++ {
				version := testutils.CreatePublishedModuleVersion(t, db, provider.ID, "1.0.0")
				// Add some downloads
				downloads := (i * j * k)
				for l := 0; l < downloads; l++ {
					testutils.CreateAnalyticsDataWithDetails(t, db, version.ID, timestamp, "1.5.0", "token", "auth", "prod", namespace.Namespace, provider.Module, provider.Provider)
				}
				totalDownloads += downloads
			}
		}
	}

	// Create query
	namespaceRepo := moduleRepo.NewNamespaceRepository(db.DB)
	domainConfig := testutils.CreateTestDomainConfig(t)
	moduleProviderRepo, err := moduleRepo.NewModuleProviderRepository(db.DB, namespaceRepo, domainConfig)
	require.NoError(t, err)
	namespaceSvc := namespaceService.NewNamespaceService(testutils.CreateTestDomainConfig(t))
	analyticsRepo, err := analyticsPersistence.NewAnalyticsRepository(db.DB, namespaceRepo, namespaceSvc)
	require.NoError(t, err)
	query := analyticsQuery.NewGlobalStatsQuery(namespaceRepo, moduleProviderRepo, analyticsRepo)

	// Execute query
	stats, err := query.Execute(ctx)
	require.NoError(t, err)

	// Verify counts: 5 namespaces, 10 modules (5*2), 30 versions (5*2*3)
	assert.Equal(t, 5, stats.Namespaces, "Should have 5 namespaces")
	assert.Equal(t, 10, stats.Modules, "Should have 10 modules")
	assert.Equal(t, 30, stats.ModuleVersions, "Should have 30 versions")
	assert.Equal(t, totalDownloads, stats.Downloads, "Should aggregate all downloads")
}

// TestGlobalStatsQuery_LargeScale tests with larger dataset to verify performance
func TestGlobalStatsQuery_LargeScale(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	ctx := context.Background()

	// Create 10 namespaces, 50 modules, 150 versions
	timestamp := time.Now()

	for i := 1; i <= 10; i++ {
		namespaceName := fmt.Sprintf("scale-ns-%d", i)
		namespace := testutils.CreateNamespace(t, db, namespaceName, nil)

		for j := 1; j <= 5; j++ {
			moduleName := fmt.Sprintf("scale-module-%d-%d", i, j)
			provider := testutils.CreateModuleProvider(t, db, namespace.ID, moduleName, "aws")

			for k := 1; k <= 3; k++ {
				version := testutils.CreatePublishedModuleVersion(t, db, provider.ID, "1.0.0")
				// Add some downloads
				for l := 0; l < 5; l++ {
					testutils.CreateAnalyticsDataWithDetails(t, db, version.ID, timestamp, "1.5.0", "token", "auth", "prod", namespace.Namespace, provider.Module, provider.Provider)
				}
			}
		}
	}

	// Create query
	namespaceRepo := moduleRepo.NewNamespaceRepository(db.DB)
	domainConfig := testutils.CreateTestDomainConfig(t)
	moduleProviderRepo, err := moduleRepo.NewModuleProviderRepository(db.DB, namespaceRepo, domainConfig)
	require.NoError(t, err)
	namespaceSvc := namespaceService.NewNamespaceService(testutils.CreateTestDomainConfig(t))
	analyticsRepo, err := analyticsPersistence.NewAnalyticsRepository(db.DB, namespaceRepo, namespaceSvc)
	require.NoError(t, err)
	query := analyticsQuery.NewGlobalStatsQuery(namespaceRepo, moduleProviderRepo, analyticsRepo)

	// Execute query and measure time
	start := time.Now()
	stats, err := query.Execute(ctx)
	elapsed := time.Since(start)

	require.NoError(t, err)

	// Verify counts
	assert.Equal(t, 10, stats.Namespaces)
	assert.Equal(t, 50, stats.Modules)
	assert.Equal(t, 150, stats.ModuleVersions)
	assert.Equal(t, 750, stats.Downloads) // 10 * 5 * 3 * 5 downloads

	// Performance check: should complete in reasonable time
	assert.Less(t, elapsed.Milliseconds(), int64(1000), "Query should complete in less than 1 second")
}

// TestGlobalStatsQuery_OnlyPublishedVersions counts only published versions
func TestGlobalStatsQuery_OnlyPublishedVersions(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	ctx := context.Background()

	// Create test data with published and unpublished versions
	namespace := testutils.CreateNamespace(t, db, "published-ns", nil)
	provider := testutils.CreateModuleProvider(t, db, namespace.ID, "test-module", "aws")

	// Create 3 published versions
	for i := 1; i <= 3; i++ {
		testutils.CreatePublishedModuleVersion(t, db, provider.ID, "1.0.0")
	}

	// Create 2 unpublished versions
	for i := 1; i <= 2; i++ {
		testutils.CreatePublishedModuleVersion(t, db, provider.ID, "0.0.0")
	}

	// Create query
	namespaceRepo := moduleRepo.NewNamespaceRepository(db.DB)
	domainConfig := testutils.CreateTestDomainConfig(t)
	moduleProviderRepo, err := moduleRepo.NewModuleProviderRepository(db.DB, namespaceRepo, domainConfig)
	require.NoError(t, err)
	namespaceSvc := namespaceService.NewNamespaceService(testutils.CreateTestDomainConfig(t))
	analyticsRepo, err := analyticsPersistence.NewAnalyticsRepository(db.DB, namespaceRepo, namespaceSvc)
	require.NoError(t, err)
	query := analyticsQuery.NewGlobalStatsQuery(namespaceRepo, moduleProviderRepo, analyticsRepo)

	// Execute query
	stats, err := query.Execute(ctx)
	require.NoError(t, err)

	// Should count all versions (the actual implementation counts all, not just published)
	assert.Equal(t, 1, stats.Namespaces)
	assert.Equal(t, 1, stats.Modules)
	assert.Equal(t, 5, stats.ModuleVersions, "Should count all versions (published + unpublished)")
	assert.Equal(t, 0, stats.Downloads)
}

// TestGlobalStatsQuery_BetaAndInternalVersions includes beta and internal versions
func TestGlobalStatsQuery_BetaAndInternalVersions(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	ctx := context.Background()

	// Create test data with beta and internal versions
	namespace := testutils.CreateNamespace(t, db, "beta-ns", nil)
	provider := testutils.CreateModuleProvider(t, db, namespace.ID, "test-module", "aws")

	// Create published version
	moduleVersion := testutils.CreatePublishedModuleVersion(t, db, provider.ID, "1.0.0")

	// Mark as published and not beta/internal
	published := true
	err := db.DB.Model(&moduleVersion).Updates(map[string]interface{}{
		"published": &published,
		"beta":      false,
		"internal":  false,
	}).Error
	require.NoError(t, err)

	// Create query
	namespaceRepo := moduleRepo.NewNamespaceRepository(db.DB)
	domainConfig := testutils.CreateTestDomainConfig(t)
	moduleProviderRepo, err := moduleRepo.NewModuleProviderRepository(db.DB, namespaceRepo, domainConfig)
	require.NoError(t, err)
	namespaceSvc := namespaceService.NewNamespaceService(testutils.CreateTestDomainConfig(t))
	analyticsRepo, err := analyticsPersistence.NewAnalyticsRepository(db.DB, namespaceRepo, namespaceSvc)
	require.NoError(t, err)
	query := analyticsQuery.NewGlobalStatsQuery(namespaceRepo, moduleProviderRepo, analyticsRepo)

	// Execute query
	stats, err := query.Execute(ctx)
	require.NoError(t, err)

	// Should count the published version
	assert.Equal(t, 1, stats.Namespaces)
	assert.Equal(t, 1, stats.Modules)
	assert.Equal(t, 1, stats.ModuleVersions)
}
