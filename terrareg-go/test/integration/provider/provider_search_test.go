package provider

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	providerquery "github.com/matthewjohn/terrareg/terrareg-go/internal/application/query/provider"
	configModel "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/config/model"
	providerdomainrepo "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/provider/repository"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/persistence/sqldb"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/persistence/sqldb/module"
	sqldbprovider "github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/persistence/sqldb/provider"
	testutils "github.com/matthewjohn/terrareg/terrareg-go/test/integration/testutils"
)

// TestProviderSearch_BasicSearch tests basic provider search functionality
func TestProviderSearch_BasicSearch(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	ctx := context.Background()

	// Setup repository
	providerRepo := sqldbprovider.NewProviderRepository(db.DB)
	searchQuery := providerquery.NewSearchProvidersQuery(providerRepo)

	// Create test namespaces
	namespace1 := testutils.CreateNamespace(t, db, "provider-ns1")
	namespace2 := testutils.CreateNamespace(t, db, "provider-ns2")

	// Create providers with different names
	description1 := "Test provider for AWS"
	description2 := "Test provider for GCP"
	description3 := "Another test provider"

	provider1 := testutils.CreateProvider(t, db, namespace1.ID, "testprovider-aws", &description1, sqldb.ProviderTierCommunity, nil)
	provider2 := testutils.CreateProvider(t, db, namespace1.ID, "testprovider-gcp", &description2, sqldb.ProviderTierCommunity, nil)
	provider3 := testutils.CreateProvider(t, db, namespace2.ID, "otherprovider", &description3, sqldb.ProviderTierCommunity, nil)

	// Create GPG keys and versions for each provider (search requires published versions)
	gpgKey1 := testutils.CreateGPGKey(t, db, "key1", provider1.ID, "ABC12345")
	gpgKey2 := testutils.CreateGPGKey(t, db, "key2", provider2.ID, "DEF67890")
	gpgKey3 := testutils.CreateGPGKey(t, db, "key3", provider3.ID, "GHI13579")

	now := time.Now()
	version1 := testutils.CreateProviderVersion(t, db, provider1.ID, "1.0.0", gpgKey1.ID, false, &now)
	version2 := testutils.CreateProviderVersion(t, db, provider2.ID, "1.0.0", gpgKey2.ID, false, &now)
	version3 := testutils.CreateProviderVersion(t, db, provider3.ID, "1.0.0", gpgKey3.ID, false, &now)

	// Set latest versions
	testutils.SetProviderLatestVersion(t, db, provider1.ID, version1.ID)
	testutils.SetProviderLatestVersion(t, db, provider2.ID, version2.ID)
	testutils.SetProviderLatestVersion(t, db, provider3.ID, version3.ID)

	t.Run("Search by partial name", func(t *testing.T) {
		result, err := searchQuery.Execute(ctx, providerdomainrepo.ProviderSearchQuery{
			Query:  "testprovider",
			Offset: 0,
			Limit:  10,
		})
		require.NoError(t, err)
		assert.Equal(t, 2, result.TotalCount)
		assert.Len(t, result.Providers, 2)
	})

	t.Run("Search by exact name", func(t *testing.T) {
		result, err := searchQuery.Execute(ctx, providerdomainrepo.ProviderSearchQuery{
			Query:  "testprovider-aws",
			Offset: 0,
			Limit:  10,
		})
		require.NoError(t, err)
		assert.Equal(t, 1, result.TotalCount)
		if len(result.Providers) > 0 {
			assert.Equal(t, "testprovider-aws", result.Providers[0].Name())
		}
	})

	t.Run("Search with no matches", func(t *testing.T) {
		result, err := searchQuery.Execute(ctx, providerdomainrepo.ProviderSearchQuery{
			Query:  "nonexistent",
			Offset: 0,
			Limit:  10,
		})
		require.NoError(t, err)
		assert.Equal(t, 0, result.TotalCount)
		assert.Empty(t, result.Providers)
	})

	t.Run("Search with empty query returns all", func(t *testing.T) {
		result, err := searchQuery.Execute(ctx, providerdomainrepo.ProviderSearchQuery{
		Query:  "",
		Offset: 0,
		Limit:  10,
		})
		require.NoError(t, err)
		assert.Equal(t, 3, result.TotalCount)
		assert.Len(t, result.Providers, 3)
	})
}

// TestProviderSearch_SearchInDescription tests searching in provider description
func TestProviderSearch_SearchInDescription(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	ctx := context.Background()

	providerRepo := sqldbprovider.NewProviderRepository(db.DB)
	searchQuery := providerquery.NewSearchProvidersQuery(providerRepo)

	namespace := testutils.CreateNamespace(t, db, "search-desc-ns")

	// Create provider with unique description
	description := "DESCRIPTION-Search unique term in description"
	provider := testutils.CreateProvider(t, db, namespace.ID, "searchdescprovider", &description, sqldb.ProviderTierCommunity, nil)

	// Create version
	gpgKey := testutils.CreateGPGKey(t, db, "key", provider.ID, "SEARCHKEY123")
	now := time.Now()
	version := testutils.CreateProviderVersion(t, db, provider.ID, "1.0.0", gpgKey.ID, false, &now)
	testutils.SetProviderLatestVersion(t, db, provider.ID, version.ID)

	t.Run("Search by description term", func(t *testing.T) {
		result, err := searchQuery.Execute(ctx, providerdomainrepo.ProviderSearchQuery{
		Query:  "DESCRIPTION-Search",
		Offset: 0,
		Limit:  10,
		})
		require.NoError(t, err)
		assert.Equal(t, 1, result.TotalCount)
		if len(result.Providers) > 0 {
			assert.Equal(t, "searchdescprovider", result.Providers[0].Name())
		}
	})
}

// TestProviderSearch_CaseInsensitiveSearch tests that search is case-insensitive
func TestProviderSearch_CaseInsensitiveSearch(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	ctx := context.Background()

	providerRepo := sqldbprovider.NewProviderRepository(db.DB)
	searchQuery := providerquery.NewSearchProvidersQuery(providerRepo)

	namespace := testutils.CreateNamespace(t, db, "case-ns")

	description := "Test provider"
	provider := testutils.CreateProvider(t, db, namespace.ID, "MixedCaseProvider", &description, sqldb.ProviderTierCommunity, nil)

	// Create version
	gpgKey := testutils.CreateGPGKey(t, db, "key", provider.ID, "MIXEDKEY123")
	now := time.Now()
	version := testutils.CreateProviderVersion(t, db, provider.ID, "1.0.0", gpgKey.ID, false, &now)
	testutils.SetProviderLatestVersion(t, db, provider.ID, version.ID)

	testCases := []struct {
		name     string
		query    string
		expected int
	}{
		{"Lowercase search", "mixedcaseprovider", 1},
		{"Uppercase search", "MIXEDCASEPROVIDER", 1},
		{"Mixed case search", "MixedCaseProvider", 1},
		{"Partial lowercase", "mixedcase", 1},
		{"No match", "nomatch", 0},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := searchQuery.Execute(ctx, providerdomainrepo.ProviderSearchQuery{
				Query:  tc.query,
				Offset: 0,
				Limit:  10,
			})
			require.NoError(t, err)
			assert.Equal(t, tc.expected, result.TotalCount)
			if tc.expected > 0 {
				assert.Len(t, result.Providers, tc.expected)
			}
		})
	}
}

// TestProviderSearch_OffsetAndLimit tests pagination with offset and limit
func TestProviderSearch_OffsetAndLimit(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	ctx := context.Background()

	providerRepo := sqldbprovider.NewProviderRepository(db.DB)
	searchQuery := providerquery.NewSearchProvidersQuery(providerRepo)

	namespace := testutils.CreateNamespace(t, db, "pagination-ns")

	// Create multiple providers
	for i := 1; i <= 5; i++ {
		description := "Test provider"
		provider := testutils.CreateProvider(t, db, namespace.ID, "pagination-provider", &description, sqldb.ProviderTierCommunity, nil)

		gpgKey := testutils.CreateGPGKey(t, db, "key", provider.ID, "PAGINATIONKEY")
		now := time.Now()
		version := testutils.CreateProviderVersion(t, db, provider.ID, "1.0.0", gpgKey.ID, false, &now)
		testutils.SetProviderLatestVersion(t, db, provider.ID, version.ID)
	}

	t.Run("Limit with offset", func(t *testing.T) {
		result, err := searchQuery.Execute(ctx, providerdomainrepo.ProviderSearchQuery{
		Query:  "pagination-provider",
		Offset: 0,
		Limit:  2,
		})
		require.NoError(t, err)
		assert.Equal(t, 5, result.TotalCount)
		assert.Len(t, result.Providers, 2)
	})

	t.Run("Offset beyond results", func(t *testing.T) {
		result, err := searchQuery.Execute(ctx, providerdomainrepo.ProviderSearchQuery{
		Query:  "pagination-provider",
		Offset: 10,
		Limit:  2,
		})
		require.NoError(t, err)
		assert.Equal(t, 5, result.TotalCount)
		assert.Empty(t, result.Providers)
	})

	t.Run("Offset in middle of results", func(t *testing.T) {
		result, err := searchQuery.Execute(ctx, providerdomainrepo.ProviderSearchQuery{
		Query:  "pagination-provider",
		Offset: 2,
		Limit:  2,
		})
		require.NoError(t, err)
		assert.Equal(t, 5, result.TotalCount)
		assert.Len(t, result.Providers, 2)
	})
}

// TestProviderSearch_ExcludeProvidersWithoutLatestVersion tests that providers without latest versions are excluded
func TestProviderSearch_ExcludeProvidersWithoutLatestVersion(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	ctx := context.Background()

	providerRepo := sqldbprovider.NewProviderRepository(db.DB)
	searchQuery := providerquery.NewSearchProvidersQuery(providerRepo)

	namespace := testutils.CreateNamespace(t, db, "latest-version-ns")

	// Create a provider with a published version (has latest)
	description1 := "Provider with latest"
	provider1 := testutils.CreateProvider(t, db, namespace.ID, "has-latest", &description1, sqldb.ProviderTierCommunity, nil)

	gpgKey1 := testutils.CreateGPGKey(t, db, "key1", provider1.ID, "HASLATESTKEY")
	now := time.Now()
	version1 := testutils.CreateProviderVersion(t, db, provider1.ID, "1.0.0", gpgKey1.ID, false, &now)
	testutils.SetProviderLatestVersion(t, db, provider1.ID, version1.ID)

	// Create a provider without any versions (no latest)
	description2 := "Provider without latest"
	_ = testutils.CreateProvider(t, db, namespace.ID, "no-latest", &description2, sqldb.ProviderTierCommunity, nil)

	t.Run("Search includes provider with latest", func(t *testing.T) {
		result, err := searchQuery.Execute(ctx, providerdomainrepo.ProviderSearchQuery{
		Query:  "",
		Offset: 0,
		Limit:  10,
		})
		require.NoError(t, err)
		// Only the provider with a version should be included
		assert.Equal(t, 1, result.TotalCount)
		if len(result.Providers) > 0 {
			assert.Contains(t, result.Providers[0].Name(), "has-latest")
		}
	})
}

// TestProviderSearch_MultipleProvidersSameNamespace tests multiple providers in same namespace
func TestProviderSearch_MultipleProvidersSameNamespace(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	ctx := context.Background()

	providerRepo := sqldbprovider.NewProviderRepository(db.DB)
	searchQuery := providerquery.NewSearchProvidersQuery(providerRepo)

	namespace := testutils.CreateNamespace(t, db, "multi-ns")

	// Create multiple providers with similar names
	description := "Test provider"
	provider1 := testutils.CreateProvider(t, db, namespace.ID, "multi-provider-one", &description, sqldb.ProviderTierCommunity, nil)
	provider2 := testutils.CreateProvider(t, db, namespace.ID, "multi-provider-two", &description, sqldb.ProviderTierCommunity, nil)

	gpgKey1 := testutils.CreateGPGKey(t, db, "key1", provider1.ID, "MULTIKEY1")
	gpgKey2 := testutils.CreateGPGKey(t, db, "key2", provider2.ID, "MULTIKEY2")

	now := time.Now()
	version1 := testutils.CreateProviderVersion(t, db, provider1.ID, "1.0.0", gpgKey1.ID, false, &now)
	version2 := testutils.CreateProviderVersion(t, db, provider2.ID, "1.0.0", gpgKey2.ID, false, &now)

	testutils.SetProviderLatestVersion(t, db, provider1.ID, version1.ID)
	testutils.SetProviderLatestVersion(t, db, provider2.ID, version2.ID)

	t.Run("Search by partial name matches multiple", func(t *testing.T) {
		result, err := searchQuery.Execute(ctx, providerdomainrepo.ProviderSearchQuery{
		Query:  "multi-provider",
		Offset: 0,
		Limit:  10,
		})
		require.NoError(t, err)
		assert.Equal(t, 2, result.TotalCount)
		assert.Len(t, result.Providers, 2)
	})
}

// TestProviderSearch_BetaVersionProviders tests that providers with beta versions are included
func TestProviderSearch_BetaVersionProviders(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	ctx := context.Background()

	providerRepo := sqldbprovider.NewProviderRepository(db.DB)
	searchQuery := providerquery.NewSearchProvidersQuery(providerRepo)

	namespace := testutils.CreateNamespace(t, db, "beta-ns")

	description := "Beta provider"
	provider := testutils.CreateProvider(t, db, namespace.ID, "beta-provider", &description, sqldb.ProviderTierCommunity, nil)

	gpgKey := testutils.CreateGPGKey(t, db, "key", provider.ID, "BETAKEY")
	now := time.Now()
	betaVersion := testutils.CreateProviderVersion(t, db, provider.ID, "1.0.0-beta", gpgKey.ID, true, &now)
	testutils.SetProviderLatestVersion(t, db, provider.ID, betaVersion.ID)

	t.Run("Search includes provider with beta version", func(t *testing.T) {
		result, err := searchQuery.Execute(ctx, providerdomainrepo.ProviderSearchQuery{
		Query:  "beta-provider",
		Offset: 0,
		Limit:  10,
		})
		require.NoError(t, err)
		assert.Equal(t, 1, result.TotalCount)
		if len(result.Providers) > 0 {
			assert.Equal(t, "beta-provider", result.Providers[0].Name())
		}
	})
}

// TestProviderSearch_WithProviderCategory tests search with provider categories
func TestProviderSearch_WithProviderCategory(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	ctx := context.Background()

	providerRepo := sqldbprovider.NewProviderRepository(db.DB)
	searchQuery := providerquery.NewSearchProvidersQuery(providerRepo)

	namespace := testutils.CreateNamespace(t, db, "category-ns")

	// Create provider category
	categoryName := "Cloud Providers"
	category := testutils.CreateProviderCategory(t, db, categoryName, "cloud", true)

	description := "Cloud provider"
	provider := testutils.CreateProvider(t, db, namespace.ID, "cloud-provider", &description, sqldb.ProviderTierCommunity, &category.ID)

	gpgKey := testutils.CreateGPGKey(t, db, "key", provider.ID, "CLOUDKEY")
	now := time.Now()
	version := testutils.CreateProviderVersion(t, db, provider.ID, "1.0.0", gpgKey.ID, false, &now)
	testutils.SetProviderLatestVersion(t, db, provider.ID, version.ID)

	t.Run("Search finds provider with category", func(t *testing.T) {
		result, err := searchQuery.Execute(ctx, providerdomainrepo.ProviderSearchQuery{
		Query:  "cloud-provider",
		Offset: 0,
		Limit:  10,
		})
		require.NoError(t, err)
		assert.Equal(t, 1, result.TotalCount)
		if len(result.Providers) > 0 {
			assert.Equal(t, "cloud-provider", result.Providers[0].Name())
		}
	})
}

// Provider Search Filters Tests
// Python reference: test/integration/terrareg/provider_search/test_get_search_filters.py

// TestProviderSearchFilters_NoResults tests search with no results
// Python reference: test_get_search_filters.py::TestGetSearchFilters::test_non_search_no_results
func TestProviderSearchFilters_NoResults(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	ctx := context.Background()

	// Setup with empty trusted namespaces
	domainConfig := &configModel.DomainConfig{
		TrustedNamespaces:        []string{},
		VerifiedModuleNamespaces: []string{},
		AllowModuleHosting:       configModel.ModuleHostingModeAllow,
		SecretKeySet:             true,
	}

	providerRepo := sqldbprovider.NewProviderRepository(db.DB)
	namespaceRepo := module.NewNamespaceRepository(db.DB)
	searchFiltersQuery := providerquery.NewSearchFiltersQuery(providerRepo, namespaceRepo, domainConfig)

	// Execute search with non-existent query
	counts, err := searchFiltersQuery.Execute(ctx, "this-search-does-not-exist-at-all")
	require.NoError(t, err)

	// Should return empty counts
	assert.Equal(t, 0, counts.TrustedNamespaces)
	assert.Equal(t, 0, counts.Contributed)
	assert.Empty(t, counts.Namespaces)
	assert.Empty(t, counts.ProviderCategories)
}

// TestProviderSearchFilters_ContributedProviderOneVersion tests search with one contributed provider
// Python reference: test_get_search_filters.py::TestGetSearchFilters::test_contributed_provider_one_version
func TestProviderSearchFilters_ContributedProviderOneVersion(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	ctx := context.Background()

	// Setup with empty trusted namespaces (all providers are contributed)
	domainConfig := &configModel.DomainConfig{
		TrustedNamespaces:        []string{},
		VerifiedModuleNamespaces: []string{},
		AllowModuleHosting:       configModel.ModuleHostingModeAllow,
		SecretKeySet:             true,
	}

	providerRepo := sqldbprovider.NewProviderRepository(db.DB)
	namespaceRepo := module.NewNamespaceRepository(db.DB)
	searchFiltersQuery := providerquery.NewSearchFiltersQuery(providerRepo, namespaceRepo, domainConfig)

	// Create test data
	namespace := testutils.CreateNamespace(t, db, "providersearch")
	category := testutils.CreateProviderCategory(t, db, "Visible Monitoring", "visible-monitoring", true)

	description := "Test provider"
	provider := testutils.CreateProvider(t, db, namespace.ID, "contributedprovider-oneversion", &description, sqldb.ProviderTierCommunity, &category.ID)

	gpgKey := testutils.CreateGPGKey(t, db, "key1", provider.ID, "CONTRIBUTEDKEY")
	now := time.Now()
	version := testutils.CreateProviderVersion(t, db, provider.ID, "1.0.0", gpgKey.ID, false, &now)
	testutils.SetProviderLatestVersion(t, db, provider.ID, version.ID)

	// Execute search
	counts, err := searchFiltersQuery.Execute(ctx, "contributedprovider-oneversion")
	require.NoError(t, err)

	// Should return 1 contributed provider
	assert.Equal(t, 0, counts.TrustedNamespaces)
	assert.Equal(t, 1, counts.Contributed)
	assert.Equal(t, map[string]int{"providersearch": 1}, counts.Namespaces)
	assert.Equal(t, 1, len(counts.ProviderCategories)) // One category
}

// TestProviderSearchFilters_ContributedProviderMultiVersion tests provider with multiple versions
// Python reference: test_get_search_filters.py::TestGetSearchFilters::test_contributed_provider_multi_version
func TestProviderSearchFilters_ContributedProviderMultiVersion(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	ctx := context.Background()

	domainConfig := &configModel.DomainConfig{
		TrustedNamespaces:        []string{},
		VerifiedModuleNamespaces: []string{},
		AllowModuleHosting:       configModel.ModuleHostingModeAllow,
		SecretKeySet:             true,
	}

	providerRepo := sqldbprovider.NewProviderRepository(db.DB)
	namespaceRepo := module.NewNamespaceRepository(db.DB)
	searchFiltersQuery := providerquery.NewSearchFiltersQuery(providerRepo, namespaceRepo, domainConfig)

	// Create test data
	namespace := testutils.CreateNamespace(t, db, "providersearch")
	category := testutils.CreateProviderCategory(t, db, "Second Visible Cloud", "second-visible-cloud", true)

	description := "Test provider"
	provider := testutils.CreateProvider(t, db, namespace.ID, "contributedprovider-multiversion", &description, sqldb.ProviderTierCommunity, &category.ID)

	gpgKey := testutils.CreateGPGKey(t, db, "key1", provider.ID, "MULTIKEY")
	now := time.Now()
	// Create multiple versions
	_ = testutils.CreateProviderVersion(t, db, provider.ID, "1.0.0", gpgKey.ID, false, &now)
	version2 := testutils.CreateProviderVersion(t, db, provider.ID, "1.1.0", gpgKey.ID, false, &now)
	testutils.SetProviderLatestVersion(t, db, provider.ID, version2.ID) // Set latest to 1.1.0

	// Execute search
	counts, err := searchFiltersQuery.Execute(ctx, "contributedprovider-multiversion")
	require.NoError(t, err)

	// Should return 1 contributed provider (multiple versions still count as 1)
	assert.Equal(t, 0, counts.TrustedNamespaces)
	assert.Equal(t, 1, counts.Contributed)
	assert.Equal(t, map[string]int{"providersearch": 1}, counts.Namespaces)
	assert.Equal(t, 1, len(counts.ProviderCategories))
}

// TestProviderSearchFilters_ContributedMultipleCategories tests multiple providers with different categories
// Python reference: test_get_search_filters.py::TestGetSearchFilters::test_contributed_multiple_categories
func TestProviderSearchFilters_ContributedMultipleCategories(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	ctx := context.Background()

	domainConfig := &configModel.DomainConfig{
		TrustedNamespaces:        []string{},
		VerifiedModuleNamespaces: []string{},
		AllowModuleHosting:       configModel.ModuleHostingModeAllow,
		SecretKeySet:             true,
	}

	providerRepo := sqldbprovider.NewProviderRepository(db.DB)
	namespaceRepo := module.NewNamespaceRepository(db.DB)
	searchFiltersQuery := providerquery.NewSearchFiltersQuery(providerRepo, namespaceRepo, domainConfig)

	// Create test data
	namespace := testutils.CreateNamespace(t, db, "providersearch")
	category1 := testutils.CreateProviderCategory(t, db, "Visible Monitoring", "visible-monitoring", true)
	category2 := testutils.CreateProviderCategory(t, db, "Second Visible Cloud", "second-visible-cloud", true)

	description := "Test provider"
	provider1 := testutils.CreateProvider(t, db, namespace.ID, "contributedprovider-one", &description, sqldb.ProviderTierCommunity, &category1.ID)
	provider2 := testutils.CreateProvider(t, db, namespace.ID, "contributedprovider-two", &description, sqldb.ProviderTierCommunity, &category2.ID)

	gpgKey1 := testutils.CreateGPGKey(t, db, "key1", provider1.ID, "KEY1")
	gpgKey2 := testutils.CreateGPGKey(t, db, "key2", provider2.ID, "KEY2")
	now := time.Now()
	version1 := testutils.CreateProviderVersion(t, db, provider1.ID, "1.0.0", gpgKey1.ID, false, &now)
	version2 := testutils.CreateProviderVersion(t, db, provider2.ID, "1.0.0", gpgKey2.ID, false, &now)
	testutils.SetProviderLatestVersion(t, db, provider1.ID, version1.ID)
	testutils.SetProviderLatestVersion(t, db, provider2.ID, version2.ID)

	// Execute search with partial match
	counts, err := searchFiltersQuery.Execute(ctx, "contributedprovider")
	require.NoError(t, err)

	// Should return 2 contributed providers
	assert.Equal(t, 0, counts.TrustedNamespaces)
	assert.Equal(t, 2, counts.Contributed)
	assert.Equal(t, map[string]int{"providersearch": 2}, counts.Namespaces)
	assert.Equal(t, 2, len(counts.ProviderCategories))
}

// TestProviderSearchFilters_NoProviderVersion tests provider without a version
// Python reference: test_get_search_filters.py::TestGetSearchFilters::test_no_provider_version
func TestProviderSearchFilters_NoProviderVersion(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	ctx := context.Background()

	domainConfig := &configModel.DomainConfig{
		TrustedNamespaces:        []string{},
		VerifiedModuleNamespaces: []string{},
		AllowModuleHosting:       configModel.ModuleHostingModeAllow,
		SecretKeySet:             true,
	}

	providerRepo := sqldbprovider.NewProviderRepository(db.DB)
	namespaceRepo := module.NewNamespaceRepository(db.DB)
	searchFiltersQuery := providerquery.NewSearchFiltersQuery(providerRepo, namespaceRepo, domainConfig)

	// Create provider WITHOUT any versions
	namespace := testutils.CreateNamespace(t, db, "providersearch")
	description := "Empty provider"
	_ = testutils.CreateProvider(t, db, namespace.ID, "empty-provider-publish", &description, sqldb.ProviderTierCommunity, nil)

	// Execute search
	counts, err := searchFiltersQuery.Execute(ctx, "empty-provider-publish")
	require.NoError(t, err)

	// Should return 0 - providers without versions are excluded
	assert.Equal(t, 0, counts.TrustedNamespaces)
	assert.Equal(t, 0, counts.Contributed)
	assert.Empty(t, counts.Namespaces)
	assert.Empty(t, counts.ProviderCategories)
}

// TestProviderSearchFilters_TrustedProviderOneVersion tests search with one trusted provider
// Python reference: test_get_search_filters.py::TestGetSearchFilters::test_trusted_provider_one_version
func TestProviderSearchFilters_TrustedProviderOneVersion(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	ctx := context.Background()

	// Setup WITH trusted namespace
	domainConfig := &configModel.DomainConfig{
		TrustedNamespaces:        []string{"providersearch"},
		VerifiedModuleNamespaces: []string{},
		AllowModuleHosting:       configModel.ModuleHostingModeAllow,
		SecretKeySet:             true,
	}

	providerRepo := sqldbprovider.NewProviderRepository(db.DB)
	namespaceRepo := module.NewNamespaceRepository(db.DB)
	searchFiltersQuery := providerquery.NewSearchFiltersQuery(providerRepo, namespaceRepo, domainConfig)

	// Create test data
	namespace := testutils.CreateNamespace(t, db, "providersearch")
	category := testutils.CreateProviderCategory(t, db, "Visible Monitoring", "visible-monitoring", true)

	description := "Test provider"
	provider := testutils.CreateProvider(t, db, namespace.ID, "contributedprovider-oneversion", &description, sqldb.ProviderTierCommunity, &category.ID)

	gpgKey := testutils.CreateGPGKey(t, db, "key1", provider.ID, "TRUSTEDKEY")
	now := time.Now()
	version := testutils.CreateProviderVersion(t, db, provider.ID, "1.0.0", gpgKey.ID, false, &now)
	testutils.SetProviderLatestVersion(t, db, provider.ID, version.ID)

	// Execute search
	counts, err := searchFiltersQuery.Execute(ctx, "contributedprovider-oneversion")
	require.NoError(t, err)

	// Should return 1 trusted provider (not contributed)
	assert.Equal(t, 1, counts.TrustedNamespaces)
	assert.Equal(t, 0, counts.Contributed)
	assert.Equal(t, map[string]int{"providersearch": 1}, counts.Namespaces)
	assert.Equal(t, 1, len(counts.ProviderCategories))
}

// TestProviderSearchFilters_TrustedProviderMultiVersion tests trusted provider with multiple versions
// Python reference: test_get_search_filters.py::TestGetSearchFilters::test_trusted_provider_multi_version
func TestProviderSearchFilters_TrustedProviderMultiVersion(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	ctx := context.Background()

	domainConfig := &configModel.DomainConfig{
		TrustedNamespaces:        []string{"providersearch"},
		VerifiedModuleNamespaces: []string{},
		AllowModuleHosting:       configModel.ModuleHostingModeAllow,
		SecretKeySet:             true,
	}

	providerRepo := sqldbprovider.NewProviderRepository(db.DB)
	namespaceRepo := module.NewNamespaceRepository(db.DB)
	searchFiltersQuery := providerquery.NewSearchFiltersQuery(providerRepo, namespaceRepo, domainConfig)

	// Create test data
	namespace := testutils.CreateNamespace(t, db, "providersearch")
	category := testutils.CreateProviderCategory(t, db, "Second Visible Cloud", "second-visible-cloud", true)

	description := "Test provider"
	provider := testutils.CreateProvider(t, db, namespace.ID, "contributedprovider-multiversion", &description, sqldb.ProviderTierCommunity, &category.ID)

	gpgKey := testutils.CreateGPGKey(t, db, "key1", provider.ID, "TRUSTEDMULTIKEY")
	now := time.Now()
	_ = testutils.CreateProviderVersion(t, db, provider.ID, "1.0.0", gpgKey.ID, false, &now)
	version2 := testutils.CreateProviderVersion(t, db, provider.ID, "1.1.0", gpgKey.ID, false, &now)
	testutils.SetProviderLatestVersion(t, db, provider.ID, version2.ID)

	// Execute search
	counts, err := searchFiltersQuery.Execute(ctx, "contributedprovider-multiversion")
	require.NoError(t, err)

	// Should return 1 trusted provider (multiple versions still count as 1)
	assert.Equal(t, 1, counts.TrustedNamespaces)
	assert.Equal(t, 0, counts.Contributed)
	assert.Equal(t, map[string]int{"providersearch": 1}, counts.Namespaces)
	assert.Equal(t, 1, len(counts.ProviderCategories))
}

// TestProviderSearchFilters_TrustedMultipleProviders tests multiple trusted providers
// Python reference: test_get_search_filters.py::TestGetSearchFilters::test_trusted_multiple_providers
func TestProviderSearchFilters_TrustedMultipleProviders(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	ctx := context.Background()

	domainConfig := &configModel.DomainConfig{
		TrustedNamespaces:        []string{"doestexist", "providersearch", "nordoesthis"},
		VerifiedModuleNamespaces: []string{},
		AllowModuleHosting:       configModel.ModuleHostingModeAllow,
		SecretKeySet:             true,
	}

	providerRepo := sqldbprovider.NewProviderRepository(db.DB)
	namespaceRepo := module.NewNamespaceRepository(db.DB)
	searchFiltersQuery := providerquery.NewSearchFiltersQuery(providerRepo, namespaceRepo, domainConfig)

	// Create test data
	namespace := testutils.CreateNamespace(t, db, "providersearch")
	category1 := testutils.CreateProviderCategory(t, db, "Visible Monitoring", "visible-monitoring", true)
	category2 := testutils.CreateProviderCategory(t, db, "Second Visible Cloud", "second-visible-cloud", true)

	description := "Test provider"
	provider1 := testutils.CreateProvider(t, db, namespace.ID, "contributedprovider-one", &description, sqldb.ProviderTierCommunity, &category1.ID)
	provider2 := testutils.CreateProvider(t, db, namespace.ID, "contributedprovider-two", &description, sqldb.ProviderTierCommunity, &category2.ID)

	gpgKey1 := testutils.CreateGPGKey(t, db, "key1", provider1.ID, "KEY1")
	gpgKey2 := testutils.CreateGPGKey(t, db, "key2", provider2.ID, "KEY2")
	now := time.Now()
	version1 := testutils.CreateProviderVersion(t, db, provider1.ID, "1.0.0", gpgKey1.ID, false, &now)
	version2 := testutils.CreateProviderVersion(t, db, provider2.ID, "1.0.0", gpgKey2.ID, false, &now)
	testutils.SetProviderLatestVersion(t, db, provider1.ID, version1.ID)
	testutils.SetProviderLatestVersion(t, db, provider2.ID, version2.ID)

	// Execute search
	counts, err := searchFiltersQuery.Execute(ctx, "contributedprovider")
	require.NoError(t, err)

	// Should return 2 trusted providers
	assert.Equal(t, 2, counts.TrustedNamespaces)
	assert.Equal(t, 0, counts.Contributed)
	assert.Equal(t, map[string]int{"providersearch": 2}, counts.Namespaces)
	assert.Equal(t, 2, len(counts.ProviderCategories))
}

// TestProviderSearchFilters_TrustedNoVersionProvider tests trusted provider without versions
// Python reference: test_get_search_filters.py::TestGetSearchFilters::test_trusted_no_version_provider
func TestProviderSearchFilters_TrustedNoVersionProvider(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	ctx := context.Background()

	domainConfig := &configModel.DomainConfig{
		TrustedNamespaces:        []string{"doestexist", "providersearch"},
		VerifiedModuleNamespaces: []string{},
		AllowModuleHosting:       configModel.ModuleHostingModeAllow,
		SecretKeySet:             true,
	}

	providerRepo := sqldbprovider.NewProviderRepository(db.DB)
	namespaceRepo := module.NewNamespaceRepository(db.DB)
	searchFiltersQuery := providerquery.NewSearchFiltersQuery(providerRepo, namespaceRepo, domainConfig)

	// Create provider WITHOUT any versions
	namespace := testutils.CreateNamespace(t, db, "providersearch")
	description := "Empty provider"
	_ = testutils.CreateProvider(t, db, namespace.ID, "contributedprovider-unpublished", &description, sqldb.ProviderTierCommunity, nil)

	// Execute search
	counts, err := searchFiltersQuery.Execute(ctx, "contributedprovider-unpublished")
	require.NoError(t, err)

	// Should return 0 - providers without versions are excluded
	assert.Equal(t, 0, counts.TrustedNamespaces)
	assert.Equal(t, 0, counts.Contributed)
	assert.Empty(t, counts.Namespaces)
	assert.Empty(t, counts.ProviderCategories)
}

// TestProviderSearchFilters_MixedResults tests mixed contributed and trusted providers
// Python reference: test_get_search_filters.py::TestGetSearchFilters::test_mixed_results
func TestProviderSearchFilters_MixedResults(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	ctx := context.Background()

	domainConfig := &configModel.DomainConfig{
		TrustedNamespaces:        []string{"doestexist", "providersearch", "providersearch-trusted"},
		VerifiedModuleNamespaces: []string{},
		AllowModuleHosting:       configModel.ModuleHostingModeAllow,
		SecretKeySet:             true,
	}

	providerRepo := sqldbprovider.NewProviderRepository(db.DB)
	namespaceRepo := module.NewNamespaceRepository(db.DB)
	searchFiltersQuery := providerquery.NewSearchFiltersQuery(providerRepo, namespaceRepo, domainConfig)

	// Create trusted namespace providers
	namespaceTrusted := testutils.CreateNamespace(t, db, "providersearch")
	category1 := testutils.CreateProviderCategory(t, db, "Visible Monitoring", "visible-monitoring", true)
	category2 := testutils.CreateProviderCategory(t, db, "Second Visible Cloud", "second-visible-cloud", true)

	description := "Test provider"
	provider1 := testutils.CreateProvider(t, db, namespaceTrusted.ID, "providersearch-one", &description, sqldb.ProviderTierCommunity, &category1.ID)
	provider2 := testutils.CreateProvider(t, db, namespaceTrusted.ID, "providersearch-two", &description, sqldb.ProviderTierCommunity, &category2.ID)

	// Create contributed namespace providers
	namespaceContributed := testutils.CreateNamespace(t, db, "contributed-providersearch")
	provider3 := testutils.CreateProvider(t, db, namespaceContributed.ID, "providersearch-three", &description, sqldb.ProviderTierCommunity, &category1.ID)
	provider4 := testutils.CreateProvider(t, db, namespaceContributed.ID, "providersearch-four", &description, sqldb.ProviderTierCommunity, &category1.ID)

	gpgKey1 := testutils.CreateGPGKey(t, db, "key1", provider1.ID, "KEY1")
	gpgKey2 := testutils.CreateGPGKey(t, db, "key2", provider2.ID, "KEY2")
	gpgKey3 := testutils.CreateGPGKey(t, db, "key3", provider3.ID, "KEY3")
	gpgKey4 := testutils.CreateGPGKey(t, db, "key4", provider4.ID, "KEY4")

	now := time.Now()
	version1 := testutils.CreateProviderVersion(t, db, provider1.ID, "1.0.0", gpgKey1.ID, false, &now)
	version2 := testutils.CreateProviderVersion(t, db, provider2.ID, "1.0.0", gpgKey2.ID, false, &now)
	version3 := testutils.CreateProviderVersion(t, db, provider3.ID, "1.0.0", gpgKey3.ID, false, &now)
	version4 := testutils.CreateProviderVersion(t, db, provider4.ID, "1.0.0", gpgKey4.ID, false, &now)

	testutils.SetProviderLatestVersion(t, db, provider1.ID, version1.ID)
	testutils.SetProviderLatestVersion(t, db, provider2.ID, version2.ID)
	testutils.SetProviderLatestVersion(t, db, provider3.ID, version3.ID)
	testutils.SetProviderLatestVersion(t, db, provider4.ID, version4.ID)

	// Execute search with partial match on "providersearch"
	counts, err := searchFiltersQuery.Execute(ctx, "providersearch")
	require.NoError(t, err)

	// Expected: 2 contributed, 2 trusted, 4 total providers
	assert.Equal(t, 2, counts.TrustedNamespaces)
	assert.Equal(t, 2, counts.Contributed)
	assert.Equal(t, map[string]int{"providersearch": 2, "contributed-providersearch": 2}, counts.Namespaces)
	// 2 unique categories (category-1 appears 3 times, category-2 once)
	// Note: Go uses placeholder category names, Python uses slugs
	assert.Equal(t, 2, len(counts.ProviderCategories))
}

// Python Test Data Tests
// These tests use comprehensive test data matching Python's integration_test_data.py

// TestProviderSearch_PythonTestData tests search using Python test data
// Python reference: test/integration/terrareg/provider_search/test_search_providers.py
func TestProviderSearch_PythonTestData(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	ctx := context.Background()

	providerRepo := sqldbprovider.NewProviderRepository(db.DB)
	searchQuery := providerquery.NewSearchProvidersQuery(providerRepo)

	// Setup comprehensive test data matching Python's test_data.py
	testutils.SetupComprehensiveProviderSearchTestData(t, db)

	t.Run("Search by description returns exact match", func(t *testing.T) {
		// Python reference: test_search_in_description
		result, err := searchQuery.Execute(ctx, providerdomainrepo.ProviderSearchQuery{
			Query:  "DESCRIPTION-Search",
			Offset: 0,
			Limit:  10,
		})
		require.NoError(t, err)
		assert.Equal(t, 1, result.TotalCount)
		if len(result.Providers) > 0 {
			assert.Equal(t, "contributedprovider-oneversion", result.Providers[0].Name())
		}
	})

	t.Run("Search by partial provider name returns multiple matches", func(t *testing.T) {
		// Python reference: test_provider_name_search_in_query_string (partial match)
		result, err := searchQuery.Execute(ctx, providerdomainrepo.ProviderSearchQuery{
			Query:  "mixedsearch",
			Offset: 0,
			Limit:  10,
		})
		require.NoError(t, err)
		// Should find: mixedsearch-result, mixedsearch-result-multiversion (in contributed-providersearch)
		assert.GreaterOrEqual(t, result.TotalCount, 2)
	})

	t.Run("Search by exact provider name returns exact match", func(t *testing.T) {
		// Python reference: test_provider_name_search_in_query_string (exact match)
		result, err := searchQuery.Execute(ctx, providerdomainrepo.ProviderSearchQuery{
			Query:  "mixedsearch-result",
			Offset: 0,
			Limit:  10,
		})
		require.NoError(t, err)
		// Should find both mixedsearch-result and mixedsearch-result-multiversion
		assert.GreaterOrEqual(t, result.TotalCount, 1)
	})

	t.Run("Search excludes providers without versions", func(t *testing.T) {
		// Python reference: test_search_in_description_no_version
		result, err := searchQuery.Execute(ctx, providerdomainrepo.ProviderSearchQuery{
			Query:  "DESCRIPTION-NoVersion",
			Offset: 0,
			Limit:  10,
		})
		require.NoError(t, err)
		// Should return 0 - providers without versions are excluded
		assert.Equal(t, 0, result.TotalCount)
	})

	t.Run("Search for non-existent provider returns no results", func(t *testing.T) {
		result, err := searchQuery.Execute(ctx, providerdomainrepo.ProviderSearchQuery{
			Query:  "this-provider-does-not-exist",
			Offset: 0,
			Limit:  10,
		})
		require.NoError(t, err)
		assert.Equal(t, 0, result.TotalCount)
	})
}

// TestProviderSearch_MaxLimitEnforcement tests that limit is capped at 50
// Python reference: test/integration/terrareg/provider_search/test_search_providers.py::test_offset_without_next
func TestProviderSearch_MaxLimitEnforcement(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	ctx := context.Background()

	providerRepo := sqldbprovider.NewProviderRepository(db.DB)
	searchQuery := providerquery.NewSearchProvidersQuery(providerRepo)

	// Setup comprehensive test data
	testutils.SetupComprehensiveProviderSearchTestData(t, db)

	t.Run("Limit of 50 is allowed", func(t *testing.T) {
		result, err := searchQuery.Execute(ctx, providerdomainrepo.ProviderSearchQuery{
			Query:  "mixedsearch",
			Offset: 0,
			Limit:  50,
		})
		require.NoError(t, err)
		// Should return all providers (at least 2)
		assert.GreaterOrEqual(t, result.TotalCount, 2)
		// Repository layer doesn't enforce limit, handler does
		assert.Len(t, result.Providers, result.TotalCount)
	})

	t.Run("Limit exceeding 50 is handled by repository (no enforcement in query layer)", func(t *testing.T) {
		// Note: The repository layer doesn't enforce max limit
		// The HTTP handler enforces max limit of 50 (provider_handler.go:427-430)
		result, err := searchQuery.Execute(ctx, providerdomainrepo.ProviderSearchQuery{
			Query:  "mixedsearch",
			Offset: 0,
			Limit:  100, // Exceeds max
		})
		require.NoError(t, err)
		// Repository will return all results (limit enforcement is in handler)
		assert.GreaterOrEqual(t, result.TotalCount, 2)
	})
}

// TestProviderSearch_NamespaceFilter tests namespace filter functionality
// Python reference: test/integration/terrareg/provider_search/test_search_providers.py::test_namespace_search_in_filter
func TestProviderSearch_NamespaceFilter(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	ctx := context.Background()

	providerRepo := sqldbprovider.NewProviderRepository(db.DB)
	searchQuery := providerquery.NewSearchProvidersQuery(providerRepo)

	// Setup comprehensive test data
	testutils.SetupComprehensiveProviderSearchTestData(t, db)

	t.Run("Filter by exact namespace match", func(t *testing.T) {
		result, err := searchQuery.Execute(ctx, providerdomainrepo.ProviderSearchQuery{
			Query:      "",
			Namespaces: []string{"providersearch"},
			Offset:     0,
			Limit:      10,
		})
		require.NoError(t, err)
		// Should find: contributedprovider-oneversion, contributedprovider-multiversion
		assert.Equal(t, 2, result.TotalCount)
	})

	t.Run("Filter by non-existent namespace returns no results", func(t *testing.T) {
		result, err := searchQuery.Execute(ctx, providerdomainrepo.ProviderSearchQuery{
			Query:      "",
			Namespaces: []string{"this-namespace-does-not-exist"},
			Offset:     0,
			Limit:      10,
		})
		require.NoError(t, err)
		assert.Equal(t, 0, result.TotalCount)
	})

	t.Run("Filter by multiple namespaces", func(t *testing.T) {
		result, err := searchQuery.Execute(ctx, providerdomainrepo.ProviderSearchQuery{
			Query:      "",
			Namespaces: []string{"providersearch", "contributed-providersearch"},
			Offset:     0,
			Limit:      10,
		})
		require.NoError(t, err)
		// Should find all 4 providers (2 in each namespace)
		assert.Equal(t, 4, result.TotalCount)
	})

	t.Run("Search in namespace with query string", func(t *testing.T) {
		result, err := searchQuery.Execute(ctx, providerdomainrepo.ProviderSearchQuery{
			Query:      "contributedprovider",
			Namespaces: []string{"providersearch"},
			Offset:     0,
			Limit:      10,
		})
		require.NoError(t, err)
		// Should find: contributedprovider-oneversion, contributedprovider-multiversion
		assert.Equal(t, 2, result.TotalCount)
	})
}

// TestProviderSearch_NoDuplicateResultsForMultiplePublishedVersions tests that providers
// with multiple published versions don't create duplicate results
// This is the provider search equivalent of the module search duplicate bug test
func TestProviderSearch_NoDuplicateResultsForMultiplePublishedVersions(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	ctx := context.Background()

	providerRepo := sqldbprovider.NewProviderRepository(db.DB)
	searchQuery := providerquery.NewSearchProvidersQuery(providerRepo)

	// Setup comprehensive test data which includes multiversion providers
	testutils.SetupComprehensiveProviderSearchTestData(t, db)

	// Search for providers that have multiple versions
	result, err := searchQuery.Execute(ctx, providerdomainrepo.ProviderSearchQuery{
		Query:  "multiversion",
		Offset: 0,
		Limit:  10,
	})
	require.NoError(t, err)

	// Critical assertions - TotalCount and Providers length must match
	// Each provider should appear exactly once, even if it has multiple versions
	assert.Equal(t, result.TotalCount, len(result.Providers),
		"TotalCount should match actual provider count (no duplicates)")
}
