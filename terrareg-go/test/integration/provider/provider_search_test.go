package provider

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	providerquery "github.com/matthewjohn/terrareg/terrareg-go/internal/application/query/provider"
	providerrepo "github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/persistence/sqldb/provider"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/persistence/sqldb"
	testutils "github.com/matthewjohn/terrareg/terrareg-go/test/integration/testutils"
)

// TestProviderSearch_BasicSearch tests basic provider search functionality
func TestProviderSearch_BasicSearch(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	ctx := context.Background()

	// Setup repository
	providerRepo := providerrepo.NewProviderRepository(db.DB)
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
		providers, count, err := searchQuery.Execute(ctx, "testprovider", 0, 10)
		require.NoError(t, err)
		assert.Equal(t, 2, count)
		assert.Len(t, providers, 2)
	})

	t.Run("Search by exact name", func(t *testing.T) {
		providers, count, err := searchQuery.Execute(ctx, "testprovider-aws", 0, 10)
		require.NoError(t, err)
		assert.Equal(t, 1, count)
		if len(providers) > 0 {
			assert.Equal(t, "testprovider-aws", providers[0].Name())
		}
	})

	t.Run("Search with no matches", func(t *testing.T) {
		providers, count, err := searchQuery.Execute(ctx, "nonexistent", 0, 10)
		require.NoError(t, err)
		assert.Equal(t, 0, count)
		assert.Empty(t, providers)
	})

	t.Run("Search with empty query returns all", func(t *testing.T) {
		providers, count, err := searchQuery.Execute(ctx, "", 0, 10)
		require.NoError(t, err)
		assert.Equal(t, 3, count)
		assert.Len(t, providers, 3)
	})
}

// TestProviderSearch_SearchInDescription tests searching in provider description
func TestProviderSearch_SearchInDescription(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	ctx := context.Background()

	providerRepo := providerrepo.NewProviderRepository(db.DB)
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
		providers, count, err := searchQuery.Execute(ctx, "DESCRIPTION-Search", 0, 10)
		require.NoError(t, err)
		assert.Equal(t, 1, count)
		if len(providers) > 0 {
			assert.Equal(t, "searchdescprovider", providers[0].Name())
		}
	})
}

// TestProviderSearch_CaseInsensitiveSearch tests that search is case-insensitive
func TestProviderSearch_CaseInsensitiveSearch(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	ctx := context.Background()

	providerRepo := providerrepo.NewProviderRepository(db.DB)
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
			providers, count, err := searchQuery.Execute(ctx, tc.query, 0, 10)
			require.NoError(t, err)
			assert.Equal(t, tc.expected, count)
			if tc.expected > 0 {
				assert.Len(t, providers, tc.expected)
			}
		})
	}
}

// TestProviderSearch_OffsetAndLimit tests pagination with offset and limit
func TestProviderSearch_OffsetAndLimit(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	ctx := context.Background()

	providerRepo := providerrepo.NewProviderRepository(db.DB)
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
		providers, count, err := searchQuery.Execute(ctx, "pagination-provider", 0, 2)
		require.NoError(t, err)
		assert.Equal(t, 5, count)
		assert.Len(t, providers, 2)
	})

	t.Run("Offset beyond results", func(t *testing.T) {
		providers, count, err := searchQuery.Execute(ctx, "pagination-provider", 10, 2)
		require.NoError(t, err)
		assert.Equal(t, 5, count)
		assert.Empty(t, providers)
	})

	t.Run("Offset in middle of results", func(t *testing.T) {
		providers, count, err := searchQuery.Execute(ctx, "pagination-provider", 2, 2)
		require.NoError(t, err)
		assert.Equal(t, 5, count)
		assert.Len(t, providers, 2)
	})
}

// TestProviderSearch_ExcludeProvidersWithoutLatestVersion tests that providers without latest versions are excluded
func TestProviderSearch_ExcludeProvidersWithoutLatestVersion(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	ctx := context.Background()

	providerRepo := providerrepo.NewProviderRepository(db.DB)
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
		providers, count, err := searchQuery.Execute(ctx, "", 0, 10)
		require.NoError(t, err)
		// Only the provider with a version should be included
		assert.Equal(t, 1, count)
		if len(providers) > 0 {
			assert.Contains(t, providers[0].Name(), "has-latest")
		}
	})
}

// TestProviderSearch_MultipleProvidersSameNamespace tests multiple providers in same namespace
func TestProviderSearch_MultipleProvidersSameNamespace(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	ctx := context.Background()

	providerRepo := providerrepo.NewProviderRepository(db.DB)
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
		providers, count, err := searchQuery.Execute(ctx, "multi-provider", 0, 10)
		require.NoError(t, err)
		assert.Equal(t, 2, count)
		assert.Len(t, providers, 2)
	})
}

// TestProviderSearch_BetaVersionProviders tests that providers with beta versions are included
func TestProviderSearch_BetaVersionProviders(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	ctx := context.Background()

	providerRepo := providerrepo.NewProviderRepository(db.DB)
	searchQuery := providerquery.NewSearchProvidersQuery(providerRepo)

	namespace := testutils.CreateNamespace(t, db, "beta-ns")

	description := "Beta provider"
	provider := testutils.CreateProvider(t, db, namespace.ID, "beta-provider", &description, sqldb.ProviderTierCommunity, nil)

	gpgKey := testutils.CreateGPGKey(t, db, "key", provider.ID, "BETAKEY")
	now := time.Now()
	betaVersion := testutils.CreateProviderVersion(t, db, provider.ID, "1.0.0-beta", gpgKey.ID, true, &now)
	testutils.SetProviderLatestVersion(t, db, provider.ID, betaVersion.ID)

	t.Run("Search includes provider with beta version", func(t *testing.T) {
		providers, count, err := searchQuery.Execute(ctx, "beta-provider", 0, 10)
		require.NoError(t, err)
		assert.Equal(t, 1, count)
		if len(providers) > 0 {
			assert.Equal(t, "beta-provider", providers[0].Name())
		}
	})
}

// TestProviderSearch_WithProviderCategory tests search with provider categories
func TestProviderSearch_WithProviderCategory(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	ctx := context.Background()

	providerRepo := providerrepo.NewProviderRepository(db.DB)
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
		providers, count, err := searchQuery.Execute(ctx, "cloud-provider", 0, 10)
		require.NoError(t, err)
		assert.Equal(t, 1, count)
		if len(providers) > 0 {
			assert.Equal(t, "cloud-provider", providers[0].Name())
		}
	})
}
