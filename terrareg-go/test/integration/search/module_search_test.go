package search

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	modulequery "github.com/matthewjohn/terrareg/terrareg-go/internal/application/query/module"
	configModel "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/config/model"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/persistence/sqldb"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/persistence/sqldb/module"
	testutils "github.com/matthewjohn/terrareg/terrareg-go/test/integration/testutils"
)

// TestModuleSearch_BasicSearch tests basic search functionality
func TestModuleSearch_BasicSearch(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	ctx := context.Background()

	// Setup repositories
	namespaceRepo := module.NewNamespaceRepository(db.DB)
	domainConfig := testutils.CreateTestDomainConfig(t)
	moduleProviderRepo := module.NewModuleProviderRepository(db.DB, namespaceRepo, domainConfig)
	searchQuery := modulequery.NewSearchModulesQuery(moduleProviderRepo)

	// Create test namespaces
	namespace1 := testutils.CreateNamespace(t, db, "search-ns1")
	namespace2 := testutils.CreateNamespace(t, db, "search-ns2")

	// Create module providers with a pattern
	provider1 := testutils.CreateModuleProvider(t, db, namespace1.ID, "searchmodule", "aws")
	provider2 := testutils.CreateModuleProvider(t, db, namespace1.ID, "searchmodule", "gcp")
	provider3 := testutils.CreateModuleProvider(t, db, namespace2.ID, "othermodule", "aws")

	// Create versions for each provider
	version1 := testutils.CreatePublishedModuleVersion(t, db, provider1.ID, "1.0.0")
	version2 := testutils.CreatePublishedModuleVersion(t, db, provider2.ID, "1.0.0")
	version3 := testutils.CreatePublishedModuleVersion(t, db, provider3.ID, "1.0.0")

	// Publish versions
	_ = version1
	_ = version2
	_ = version3

	t.Run("Search by partial name", func(t *testing.T) {
		params := modulequery.SearchParams{
			Query: "searchmod",
			Limit: 10,
		}

		result, err := searchQuery.Execute(ctx, params)
		require.NoError(t, err)
		assert.Equal(t, 2, result.TotalCount)
		assert.Len(t, result.Modules, 2)
	})

	t.Run("Search by exact module name", func(t *testing.T) {
		params := modulequery.SearchParams{
			Query: "searchmodule",
			Limit: 10,
		}

		result, err := searchQuery.Execute(ctx, params)
		require.NoError(t, err)
		assert.Equal(t, 2, result.TotalCount)
	})

	t.Run("Search with no matches", func(t *testing.T) {
		params := modulequery.SearchParams{
			Query: "nonexistent",
			Limit: 10,
		}

		result, err := searchQuery.Execute(ctx, params)
		require.NoError(t, err)
		assert.Equal(t, 0, result.TotalCount)
		assert.Empty(t, result.Modules)
	})

	t.Run("Search with empty query returns all", func(t *testing.T) {
		params := modulequery.SearchParams{
			Query: "",
			Limit: 10,
		}

		result, err := searchQuery.Execute(ctx, params)
		require.NoError(t, err)
		assert.Equal(t, 3, result.TotalCount)
	})
}

// TestModuleSearch_NamespaceFilter tests filtering by namespace
func TestModuleSearch_NamespaceFilter(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	ctx := context.Background()

	namespaceRepo := module.NewNamespaceRepository(db.DB)
	domainConfig := testutils.CreateTestDomainConfig(t)
	moduleProviderRepo := module.NewModuleProviderRepository(db.DB, namespaceRepo, domainConfig)
	searchQuery := modulequery.NewSearchModulesQuery(moduleProviderRepo)

	// Create test namespaces
	namespace1 := testutils.CreateNamespace(t, db, "filter-ns1")
	namespace2 := testutils.CreateNamespace(t, db, "filter-ns2")
	namespace3 := testutils.CreateNamespace(t, db, "different-ns")

	// Create module providers
	provider1 := testutils.CreateModuleProvider(t, db, namespace1.ID, "testmodule", "aws")
	provider2 := testutils.CreateModuleProvider(t, db, namespace2.ID, "testmodule", "aws")
	provider3 := testutils.CreateModuleProvider(t, db, namespace3.ID, "testmodule", "aws")

	// Create versions
	testutils.CreatePublishedModuleVersion(t, db, provider1.ID, "1.0.0")
	testutils.CreatePublishedModuleVersion(t, db, provider2.ID, "1.0.0")
	testutils.CreatePublishedModuleVersion(t, db, provider3.ID, "1.0.0")

	t.Run("Filter by single namespace", func(t *testing.T) {
		params := modulequery.SearchParams{
			Query:      "testmodule",
			Namespaces: []string{"filter-ns1"},
			Limit:      10,
		}

		result, err := searchQuery.Execute(ctx, params)
		require.NoError(t, err)
		assert.Equal(t, 1, result.TotalCount)
	})

	t.Run("Filter by multiple namespaces", func(t *testing.T) {
		params := modulequery.SearchParams{
			Query:      "testmodule",
			Namespaces: []string{"filter-ns1", "filter-ns2"},
			Limit:      10,
		}

		result, err := searchQuery.Execute(ctx, params)
		require.NoError(t, err)
		assert.Equal(t, 2, result.TotalCount)
	})

	t.Run("Filter by non-existent namespace", func(t *testing.T) {
		params := modulequery.SearchParams{
			Query:      "testmodule",
			Namespaces: []string{"nonexistent"},
			Limit:      10,
		}

		result, err := searchQuery.Execute(ctx, params)
		require.NoError(t, err)
		assert.Equal(t, 0, result.TotalCount)
	})
}

// TestModuleSearch_ProviderFilter tests filtering by provider name
func TestModuleSearch_ProviderFilter(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	ctx := context.Background()

	namespaceRepo := module.NewNamespaceRepository(db.DB)
	domainConfig := testutils.CreateTestDomainConfig(t)
	moduleProviderRepo := module.NewModuleProviderRepository(db.DB, namespaceRepo, domainConfig)
	searchQuery := modulequery.NewSearchModulesQuery(moduleProviderRepo)

	namespace := testutils.CreateNamespace(t, db, "provider-filter-ns")

	// Create module providers with different providers
	provider1 := testutils.CreateModuleProvider(t, db, namespace.ID, "testmodule", "aws")
	provider2 := testutils.CreateModuleProvider(t, db, namespace.ID, "testmodule", "gcp")
	provider3 := testutils.CreateModuleProvider(t, db, namespace.ID, "testmodule", "azure")

	// Create versions
	testutils.CreatePublishedModuleVersion(t, db, provider1.ID, "1.0.0")
	testutils.CreatePublishedModuleVersion(t, db, provider2.ID, "1.0.0")
	testutils.CreatePublishedModuleVersion(t, db, provider3.ID, "1.0.0")

	t.Run("Filter by single provider", func(t *testing.T) {
		params := modulequery.SearchParams{
			Query:     "testmodule",
			Providers: []string{"aws"},
			Limit:     10,
		}

		result, err := searchQuery.Execute(ctx, params)
		require.NoError(t, err)
		assert.Equal(t, 1, result.TotalCount)
	})

	t.Run("Filter by multiple providers", func(t *testing.T) {
		params := modulequery.SearchParams{
			Query:     "testmodule",
			Providers: []string{"aws", "gcp"},
			Limit:     10,
		}

		result, err := searchQuery.Execute(ctx, params)
		require.NoError(t, err)
		assert.Equal(t, 2, result.TotalCount)
	})

	t.Run("Filter using Provider field for backward compatibility", func(t *testing.T) {
		azure := "azure"
		params := modulequery.SearchParams{
			Query:    "testmodule",
			Provider: &azure,
			Limit:    10,
		}

		result, err := searchQuery.Execute(ctx, params)
		require.NoError(t, err)
		assert.Equal(t, 1, result.TotalCount)
	})
}

// TestModuleSearch_VerifiedFilter tests filtering by verified flag
func TestModuleSearch_VerifiedFilter(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	ctx := context.Background()

	namespaceRepo := module.NewNamespaceRepository(db.DB)
	domainConfig := testutils.CreateTestDomainConfig(t)
	moduleProviderRepo := module.NewModuleProviderRepository(db.DB, namespaceRepo, domainConfig)
	searchQuery := modulequery.NewSearchModulesQuery(moduleProviderRepo)

	namespace := testutils.CreateNamespace(t, db, "verified-ns")

	// Create verified and unverified providers
	provider1 := testutils.CreateModuleProvider(t, db, namespace.ID, "verified-module", "aws")
	verified := true
	provider1.Verified = &verified
	db.DB.Save(&provider1)

	provider2 := testutils.CreateModuleProvider(t, db, namespace.ID, "verified-module", "gcp")

	// Create versions
	testutils.CreatePublishedModuleVersion(t, db, provider1.ID, "1.0.0")
	testutils.CreatePublishedModuleVersion(t, db, provider2.ID, "1.0.0")

	t.Run("Filter for verified only", func(t *testing.T) {
		verified := true
		params := modulequery.SearchParams{
			Query:    "verified-module",
			Verified: &verified,
			Limit:    10,
		}

		result, err := searchQuery.Execute(ctx, params)
		require.NoError(t, err)
		assert.Equal(t, 1, result.TotalCount)
	})

	t.Run("Filter for unverified only", func(t *testing.T) {
		unverified := false
		params := modulequery.SearchParams{
			Query:    "verified-module",
			Verified: &unverified,
			Limit:    10,
		}

		result, err := searchQuery.Execute(ctx, params)
		require.NoError(t, err)
		assert.Equal(t, 1, result.TotalCount)
	})

	t.Run("No verified filter returns all", func(t *testing.T) {
		params := modulequery.SearchParams{
			Query: "verified-module",
			Limit: 10,
		}

		result, err := searchQuery.Execute(ctx, params)
		require.NoError(t, err)
		assert.Equal(t, 2, result.TotalCount)
	})
}

// TestModuleSearch_OffsetAndLimit tests pagination with offset and limit
func TestModuleSearch_OffsetAndLimit(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	ctx := context.Background()

	namespaceRepo := module.NewNamespaceRepository(db.DB)
	domainConfig := testutils.CreateTestDomainConfig(t)
	moduleProviderRepo := module.NewModuleProviderRepository(db.DB, namespaceRepo, domainConfig)
	searchQuery := modulequery.NewSearchModulesQuery(moduleProviderRepo)

	namespace := testutils.CreateNamespace(t, db, "pagination-ns")

	// Create multiple module providers with unique provider names
	providers := make([]sqldb.ModuleProviderDB, 5)
	for i := 1; i <= 5; i++ {
		providerName := fmt.Sprintf("provider-%d", i)
		provider := testutils.CreateModuleProvider(t, db, namespace.ID, "pagination-module", providerName)
		testutils.CreatePublishedModuleVersion(t, db, provider.ID, "1.0.0")
		providers[i-1] = provider
	}

	t.Run("Default limit when not specified", func(t *testing.T) {
		params := modulequery.SearchParams{
			Query: "pagination-module",
			// No limit specified - should default to 20
		}

		result, err := searchQuery.Execute(ctx, params)
		require.NoError(t, err)
		assert.Equal(t, 5, result.TotalCount)
		assert.Len(t, result.Modules, 5)
	})

	t.Run("Limit with offset", func(t *testing.T) {
		params := modulequery.SearchParams{
			Query:  "pagination-module",
			Limit:  2,
			Offset: 0,
		}

		result, err := searchQuery.Execute(ctx, params)
		require.NoError(t, err)
		assert.Equal(t, 5, result.TotalCount)
		assert.Len(t, result.Modules, 2)
	})

	t.Run("Offset beyond results", func(t *testing.T) {
		params := modulequery.SearchParams{
			Query:  "pagination-module",
			Limit:  2,
			Offset: 10,
		}

		result, err := searchQuery.Execute(ctx, params)
		require.NoError(t, err)
		assert.Equal(t, 5, result.TotalCount)
		assert.Empty(t, result.Modules)
	})

	t.Run("Offset in middle of results", func(t *testing.T) {
		params := modulequery.SearchParams{
			Query:  "pagination-module",
			Limit:  2,
			Offset: 2,
		}

		result, err := searchQuery.Execute(ctx, params)
		require.NoError(t, err)
		assert.Equal(t, 5, result.TotalCount)
		assert.Len(t, result.Modules, 2)
	})
}

// TestModuleSearch_ExcludeModulesWithoutLatestVersion tests that modules without latest versions are excluded
func TestModuleSearch_ExcludeModulesWithoutLatestVersion(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	ctx := context.Background()

	namespaceRepo := module.NewNamespaceRepository(db.DB)
	domainConfig := testutils.CreateTestDomainConfig(t)
	moduleProviderRepo := module.NewModuleProviderRepository(db.DB, namespaceRepo, domainConfig)
	searchQuery := modulequery.NewSearchModulesQuery(moduleProviderRepo)

	namespace := testutils.CreateNamespace(t, db, "latest-version-ns")

	// Create a provider with a published version (has latest)
	provider1 := testutils.CreateModuleProvider(t, db, namespace.ID, "has-latest", "aws")
	version1 := testutils.CreatePublishedModuleVersion(t, db, provider1.ID, "1.0.0")
	published := true
	version1.Published = &published
	db.DB.Save(&version1)

	// Create a provider without any versions (no latest)
	_ = testutils.CreateModuleProvider(t, db, namespace.ID, "no-latest", "aws")

	params := modulequery.SearchParams{
		Query: "", // Search all
		Limit: 10,
	}

	result, err := searchQuery.Execute(ctx, params)
	require.NoError(t, err)
	// TotalCount should be 2 (both providers match the query)
	// But only the provider with a published version should be in the results
	assert.Equal(t, 2, result.TotalCount)
	assert.Len(t, result.Modules, 1)
	if len(result.Modules) > 0 {
		assert.Contains(t, result.Modules[0].Module(), "has-latest")
	}
}

// TestModuleSearch_CaseInsensitiveSearch tests that search is case insensitive
func TestModuleSearch_CaseInsensitiveSearch(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	ctx := context.Background()

	namespaceRepo := module.NewNamespaceRepository(db.DB)
	domainConfig := testutils.CreateTestDomainConfig(t)
	moduleProviderRepo := module.NewModuleProviderRepository(db.DB, namespaceRepo, domainConfig)
	searchQuery := modulequery.NewSearchModulesQuery(moduleProviderRepo)

	namespace := testutils.CreateNamespace(t, db, "case-ns")

	provider := testutils.CreateModuleProvider(t, db, namespace.ID, "MixedCaseModule", "aws")
	testutils.CreatePublishedModuleVersion(t, db, provider.ID, "1.0.0")

	testCases := []struct {
		name     string
		query    string
		expected int
	}{
		{"Lowercase search", "mixedcasemodule", 1},
		{"Uppercase search", "MIXEDCASEMODULE", 1},
		{"Mixed case search", "MixedCaseModule", 1},
		{"Partial lowercase", "mixedcase", 1},
		{"No match", "nomatch", 0},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			params := modulequery.SearchParams{
				Query: tc.query,
				Limit: 10,
			}

			result, err := searchQuery.Execute(ctx, params)
			require.NoError(t, err)
			assert.Equal(t, tc.expected, result.TotalCount)
		})
	}
}

// TestModuleSearch_CombinedFilters tests combining multiple filters
func TestModuleSearch_CombinedFilters(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	ctx := context.Background()

	namespaceRepo := module.NewNamespaceRepository(db.DB)
	domainConfig := testutils.CreateTestDomainConfig(t)
	moduleProviderRepo := module.NewModuleProviderRepository(db.DB, namespaceRepo, domainConfig)
	searchQuery := modulequery.NewSearchModulesQuery(moduleProviderRepo)

	namespace1 := testutils.CreateNamespace(t, db, "combined-ns1")
	namespace2 := testutils.CreateNamespace(t, db, "combined-ns2")

	// Create various providers
	provider1 := testutils.CreateModuleProvider(t, db, namespace1.ID, "combined-module", "aws")
	verified1 := true
	provider1.Verified = &verified1
	db.DB.Save(&provider1)

	provider2 := testutils.CreateModuleProvider(t, db, namespace1.ID, "combined-module", "gcp")
	verified2 := true
	provider2.Verified = &verified2
	db.DB.Save(&provider2)

	provider3 := testutils.CreateModuleProvider(t, db, namespace2.ID, "combined-module", "aws")

	// Create versions
	testutils.CreatePublishedModuleVersion(t, db, provider1.ID, "1.0.0")
	testutils.CreatePublishedModuleVersion(t, db, provider2.ID, "1.0.0")
	testutils.CreatePublishedModuleVersion(t, db, provider3.ID, "1.0.0")

	t.Run("Namespace + Provider + Verified", func(t *testing.T) {
		verified := true
		params := modulequery.SearchParams{
			Query:      "combined-module",
			Namespaces: []string{"combined-ns1"},
			Providers:  []string{"aws"},
			Verified:   &verified,
			Limit:      10,
		}

		result, err := searchQuery.Execute(ctx, params)
		require.NoError(t, err)
		assert.Equal(t, 1, result.TotalCount)
	})

	t.Run("Namespace + Multiple Providers", func(t *testing.T) {
		params := modulequery.SearchParams{
			Query:      "combined-module",
			Namespaces: []string{"combined-ns1"},
			Providers:  []string{"aws", "gcp"},
			Limit:      10,
		}

		result, err := searchQuery.Execute(ctx, params)
		require.NoError(t, err)
		assert.Equal(t, 2, result.TotalCount)
	})
}

// TestModuleSearch_TrustedNamespaceFilter tests filtering by trusted namespace
func TestModuleSearch_TrustedNamespaceFilter(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	ctx := context.Background()

	namespaceRepo := module.NewNamespaceRepository(db.DB)

	// Create custom domain config with "trusted-ns" as trusted namespace
	domainConfig := &configModel.DomainConfig{
		TrustedNamespaces:        []string{"trusted-ns"},
		VerifiedModuleNamespaces: []string{"verified"},
		AllowModuleHosting:       configModel.ModuleHostingModeAllow,
		SecretKeySet:             true,
	}

	moduleProviderRepo := module.NewModuleProviderRepository(db.DB, namespaceRepo, domainConfig)
	searchQuery := modulequery.NewSearchModulesQuery(moduleProviderRepo)

	namespace1 := testutils.CreateNamespace(t, db, "trusted-ns")
	namespace2 := testutils.CreateNamespace(t, db, "untrusted-ns")

	provider1 := testutils.CreateModuleProvider(t, db, namespace1.ID, "trust-module", "aws")
	provider2 := testutils.CreateModuleProvider(t, db, namespace2.ID, "trust-module", "aws")

	// Create and publish versions for search to find them
	version1 := testutils.CreatePublishedModuleVersion(t, db, provider1.ID, "1.0.0")
	published := true
	version1.Published = &published
	db.DB.Save(&version1)

	version2 := testutils.CreatePublishedModuleVersion(t, db, provider2.ID, "1.0.0")
	version2.Published = &published
	db.DB.Save(&version2)

	t.Run("Filter for trusted namespaces only", func(t *testing.T) {
		trusted := true
		params := modulequery.SearchParams{
			Query:             "trust-module",
			TrustedNamespaces: &trusted,
			Limit:             10,
		}

		result, err := searchQuery.Execute(ctx, params)
		require.NoError(t, err)
		assert.Equal(t, 1, result.TotalCount)
	})

	t.Run("Filter for contributed (untrusted) namespaces only", func(t *testing.T) {
		contributed := true
		params := modulequery.SearchParams{
			Query:       "trust-module",
			Contributed: &contributed,
			Limit:       10,
		}

		result, err := searchQuery.Execute(ctx, params)
		require.NoError(t, err)
		assert.Equal(t, 1, result.TotalCount)
	})
}
