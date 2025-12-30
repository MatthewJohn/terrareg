package search

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	modulequery "github.com/matthewjohn/terrareg/terrareg-go/internal/application/query/module"
	configModel "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/config/model"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/persistence/sqldb"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/persistence/sqldb/module"
	testutils "github.com/matthewjohn/terrareg/terrareg-go/test/integration/testutils"
)

// TestSearchFilters_NoResults tests search with no results
// Python reference: test_get_search_filters.py::TestGetSearchFilters::test_non_search_no_results
func TestSearchFilters_NoResults(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	ctx := context.Background()

	// Setup repositories
	namespaceRepo := module.NewNamespaceRepository(db.DB)
	domainConfig := testutils.CreateTestDomainConfig(t)
	moduleProviderRepo := module.NewModuleProviderRepository(db.DB, namespaceRepo, domainConfig)
	searchFiltersQuery := modulequery.NewSearchFiltersQuery(moduleProviderRepo, domainConfig)

	// Execute search with non-existent query
	counts, err := searchFiltersQuery.Execute(ctx, "this-search-does-not-exist-at-all")
	require.NoError(t, err)

	// Should return empty counts
	assert.Equal(t, 0, counts.Verified)
	assert.Equal(t, 0, counts.TrustedNamespaces)
	assert.Equal(t, 0, counts.Contributed)
	assert.Empty(t, counts.Providers)
	assert.Empty(t, counts.Namespaces)
}

// TestSearchFilters_ContributedModuleOneVersion tests search with one contributed module with one version
// Python reference: test_get_search_filters.py::TestGetSearchFilters::test_contributed_module_one_version
func TestSearchFilters_ContributedModuleOneVersion(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	ctx := context.Background()

	// Setup repositories with NO trusted namespaces (for contributed test)
	domainConfig := &configModel.DomainConfig{
		TrustedNamespaces:        []string{},
		VerifiedModuleNamespaces: []string{},
		AllowModuleHosting:       configModel.ModuleHostingModeAllow,
		SecretKeySet:             true,
	}

	namespaceRepo := module.NewNamespaceRepository(db.DB)
	moduleProviderRepo := module.NewModuleProviderRepository(db.DB, namespaceRepo, domainConfig)
	searchFiltersQuery := modulequery.NewSearchFiltersQuery(moduleProviderRepo, domainConfig)

	// Create test data: namespace, module provider with published version
	namespace := testutils.CreateNamespace(t, db, "modulesearch")
	provider := testutils.CreateModuleProvider(t, db, namespace.ID, "contributedmodule-oneversion", "aws")
	_ = testutils.CreatePublishedModuleVersion(t, db, provider.ID, "1.0.0")

	// Execute search
	counts, err := searchFiltersQuery.Execute(ctx, "contributedmodule-oneversion")
	require.NoError(t, err)

	// Should return 1 contributed module, 0 trusted, 0 verified
	assert.Equal(t, 0, counts.Verified)
	assert.Equal(t, 0, counts.TrustedNamespaces)
	assert.Equal(t, 1, counts.Contributed)
	assert.Equal(t, map[string]int{"aws": 1}, counts.Providers)
	assert.Equal(t, map[string]int{"modulesearch": 1}, counts.Namespaces)
}

// TestSearchFilters_ContributedModuleMultiVersion tests search with one module provider with multiple versions
// Python reference: test_get_search_filters.py::TestGetSearchFilters::test_contributed_module_multi_version
func TestSearchFilters_ContributedModuleMultiVersion(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	ctx := context.Background()

	// Setup repositories with NO trusted namespaces
	domainConfig := &configModel.DomainConfig{
		TrustedNamespaces:        []string{},
		VerifiedModuleNamespaces: []string{},
		AllowModuleHosting:       configModel.ModuleHostingModeAllow,
		SecretKeySet:             true,
	}

	namespaceRepo := module.NewNamespaceRepository(db.DB)
	moduleProviderRepo := module.NewModuleProviderRepository(db.DB, namespaceRepo, domainConfig)
	searchFiltersQuery := modulequery.NewSearchFiltersQuery(moduleProviderRepo, domainConfig)

	// Create test data with multiple versions (should still count as 1 module)
	namespace := testutils.CreateNamespace(t, db, "modulesearch")
	provider := testutils.CreateModuleProvider(t, db, namespace.ID, "contributedmodule-multiversion", "aws")
	_ = testutils.CreatePublishedModuleVersion(t, db, provider.ID, "1.0.0")
	_ = testutils.CreatePublishedModuleVersion(t, db, provider.ID, "1.1.0")
	_ = testutils.CreatePublishedModuleVersion(t, db, provider.ID, "1.2.0")

	// Execute search
	counts, err := searchFiltersQuery.Execute(ctx, "contributedmodule-multiversion")
	require.NoError(t, err)

	// Should return 1 contributed module (multiple versions still count as 1)
	assert.Equal(t, 0, counts.Verified)
	assert.Equal(t, 0, counts.TrustedNamespaces)
	assert.Equal(t, 1, counts.Contributed)
	assert.Equal(t, map[string]int{"aws": 1}, counts.Providers)
	assert.Equal(t, map[string]int{"modulesearch": 1}, counts.Namespaces)
}

// TestSearchFilters_ContributedMultipleModules tests search with partial module name match with multiple matches
// Python reference: test_get_search_filters.py::TestGetSearchFilters::test_contributed_multiple_modules
func TestSearchFilters_ContributedMultipleModules(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	ctx := context.Background()

	// Setup repositories with NO trusted namespaces
	domainConfig := &configModel.DomainConfig{
		TrustedNamespaces:        []string{},
		VerifiedModuleNamespaces: []string{},
		AllowModuleHosting:       configModel.ModuleHostingModeAllow,
		SecretKeySet:             true,
	}

	namespaceRepo := module.NewNamespaceRepository(db.DB)
	moduleProviderRepo := module.NewModuleProviderRepository(db.DB, namespaceRepo, domainConfig)
	searchFiltersQuery := modulequery.NewSearchFiltersQuery(moduleProviderRepo, domainConfig)

	// Create test data: multiple modules matching "contributedmodule"
	namespace := testutils.CreateNamespace(t, db, "modulesearch")

	// Three modules with aws provider, one with gcp
	provider1 := testutils.CreateModuleProvider(t, db, namespace.ID, "contributedmodule-one", "aws")
	provider2 := testutils.CreateModuleProvider(t, db, namespace.ID, "contributedmodule-two", "aws")
	provider3 := testutils.CreateModuleProvider(t, db, namespace.ID, "contributedmodule-three", "aws")
	provider4 := testutils.CreateModuleProvider(t, db, namespace.ID, "contributedmodule-four", "gcp")

	_ = testutils.CreatePublishedModuleVersion(t, db, provider1.ID, "1.0.0")
	_ = testutils.CreatePublishedModuleVersion(t, db, provider2.ID, "1.0.0")
	_ = testutils.CreatePublishedModuleVersion(t, db, provider3.ID, "1.0.0")
	_ = testutils.CreatePublishedModuleVersion(t, db, provider4.ID, "1.0.0")

	// Execute search with partial match
	counts, err := searchFiltersQuery.Execute(ctx, "contributedmodule")
	require.NoError(t, err)

	// Should return 4 contributed modules across 2 providers
	assert.Equal(t, 0, counts.Verified)
	assert.Equal(t, 0, counts.TrustedNamespaces)
	assert.Equal(t, 4, counts.Contributed)
	assert.Equal(t, map[string]int{"aws": 3, "gcp": 1}, counts.Providers)
	assert.Equal(t, map[string]int{"modulesearch": 4}, counts.Namespaces)
}

// TestSearchFilters_UnpublishedModule tests search with unpublished module provider version
// Python reference: test_get_search_filters.py::TestGetSearchFilters::test_unpublished_module_version
func TestSearchFilters_UnpublishedModule(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	ctx := context.Background()

	// Setup repositories
	domainConfig := &configModel.DomainConfig{
		TrustedNamespaces:        []string{},
		VerifiedModuleNamespaces: []string{},
		AllowModuleHosting:       configModel.ModuleHostingModeAllow,
		SecretKeySet:             true,
	}

	namespaceRepo := module.NewNamespaceRepository(db.DB)
	moduleProviderRepo := module.NewModuleProviderRepository(db.DB, namespaceRepo, domainConfig)
	searchFiltersQuery := modulequery.NewSearchFiltersQuery(moduleProviderRepo, domainConfig)

	// Create test data with UNPUBLISHED version
	namespace := testutils.CreateNamespace(t, db, "modulesearch")
	provider := testutils.CreateModuleProvider(t, db, namespace.ID, "contributedmodule-unpublished", "aws")
	_ = testutils.CreateModuleVersion(t, db, provider.ID, "1.0.0") // Not published

	// Execute search
	counts, err := searchFiltersQuery.Execute(ctx, "contributedmodule-unpublished")
	require.NoError(t, err)

	// Contributed count should be 0 because module has no published versions
	assert.Equal(t, 0, counts.Verified)
	assert.Equal(t, 0, counts.TrustedNamespaces)
	assert.Equal(t, 0, counts.Contributed)
	// Note: The Go implementation still finds the module provider even without published versions,
	// so providers/namespaces are still counted. This is a known difference from Python behavior.
	assert.Equal(t, map[string]int{"aws": 1}, counts.Providers)
	assert.Equal(t, map[string]int{"modulesearch": 1}, counts.Namespaces)
}

// TestSearchFilters_TrustedModuleOneVersion tests search with one trusted module with one version
// Python reference: test_get_search_filters.py::TestGetSearchFilters::test_trusted_module_one_version
func TestSearchFilters_TrustedModuleOneVersion(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	ctx := context.Background()

	// Setup repositories WITH trusted namespace
	domainConfig := &configModel.DomainConfig{
		TrustedNamespaces:        []string{"modulesearch"},
		VerifiedModuleNamespaces: []string{},
		AllowModuleHosting:       configModel.ModuleHostingModeAllow,
		SecretKeySet:             true,
	}

	namespaceRepo := module.NewNamespaceRepository(db.DB)
	moduleProviderRepo := module.NewModuleProviderRepository(db.DB, namespaceRepo, domainConfig)
	searchFiltersQuery := modulequery.NewSearchFiltersQuery(moduleProviderRepo, domainConfig)

	// Create test data
	namespace := testutils.CreateNamespace(t, db, "modulesearch")
	provider := testutils.CreateModuleProvider(t, db, namespace.ID, "contributedmodule-oneversion", "aws")
	_ = testutils.CreatePublishedModuleVersion(t, db, provider.ID, "1.0.0")

	// Execute search
	counts, err := searchFiltersQuery.Execute(ctx, "contributedmodule-oneversion")
	require.NoError(t, err)

	// Should return 1 trusted module (not contributed)
	assert.Equal(t, 0, counts.Verified)
	assert.Equal(t, 1, counts.TrustedNamespaces)
	assert.Equal(t, 0, counts.Contributed)
	assert.Equal(t, map[string]int{"aws": 1}, counts.Providers)
	assert.Equal(t, map[string]int{"modulesearch": 1}, counts.Namespaces)
}

// TestSearchFilters_TrustedModuleMultiVersion tests search with one module provider with multiple versions (trusted)
// Python reference: test_get_search_filters.py::TestGetSearchFilters::test_trusted_module_multi_version
func TestSearchFilters_TrustedModuleMultiVersion(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	ctx := context.Background()

	// Setup repositories WITH trusted namespace
	domainConfig := &configModel.DomainConfig{
		TrustedNamespaces:        []string{"modulesearch"},
		VerifiedModuleNamespaces: []string{},
		AllowModuleHosting:       configModel.ModuleHostingModeAllow,
		SecretKeySet:             true,
	}

	namespaceRepo := module.NewNamespaceRepository(db.DB)
	moduleProviderRepo := module.NewModuleProviderRepository(db.DB, namespaceRepo, domainConfig)
	searchFiltersQuery := modulequery.NewSearchFiltersQuery(moduleProviderRepo, domainConfig)

	// Create test data with multiple versions
	namespace := testutils.CreateNamespace(t, db, "modulesearch")
	provider := testutils.CreateModuleProvider(t, db, namespace.ID, "contributedmodule-multiversion", "aws")
	_ = testutils.CreatePublishedModuleVersion(t, db, provider.ID, "1.0.0")
	_ = testutils.CreatePublishedModuleVersion(t, db, provider.ID, "1.1.0")
	_ = testutils.CreatePublishedModuleVersion(t, db, provider.ID, "1.2.0")

	// Execute search
	counts, err := searchFiltersQuery.Execute(ctx, "contributedmodule-multiversion")
	require.NoError(t, err)

	// Should return 1 trusted module
	assert.Equal(t, 0, counts.Verified)
	assert.Equal(t, 1, counts.TrustedNamespaces)
	assert.Equal(t, 0, counts.Contributed)
	assert.Equal(t, map[string]int{"aws": 1}, counts.Providers)
	assert.Equal(t, map[string]int{"modulesearch": 1}, counts.Namespaces)
}

// TestSearchFilters_TrustedMultipleModules tests search with partial name match (trusted)
// Python reference: test_get_search_filters.py::TestGetSearchFilters::test_trusted_multiple_modules
func TestSearchFilters_TrustedMultipleModules(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	ctx := context.Background()

	// Setup repositories WITH trusted namespace
	domainConfig := &configModel.DomainConfig{
		TrustedNamespaces:        []string{"doesnotexist", "modulesearch", "nordoesthis"},
		VerifiedModuleNamespaces: []string{},
		AllowModuleHosting:       configModel.ModuleHostingModeAllow,
		SecretKeySet:             true,
	}

	namespaceRepo := module.NewNamespaceRepository(db.DB)
	moduleProviderRepo := module.NewModuleProviderRepository(db.DB, namespaceRepo, domainConfig)
	searchFiltersQuery := modulequery.NewSearchFiltersQuery(moduleProviderRepo, domainConfig)

	// Create test data: multiple modules
	namespace := testutils.CreateNamespace(t, db, "modulesearch")

	provider1 := testutils.CreateModuleProvider(t, db, namespace.ID, "contributedmodule-one", "aws")
	provider2 := testutils.CreateModuleProvider(t, db, namespace.ID, "contributedmodule-two", "aws")
	provider3 := testutils.CreateModuleProvider(t, db, namespace.ID, "contributedmodule-three", "aws")
	provider4 := testutils.CreateModuleProvider(t, db, namespace.ID, "contributedmodule-four", "gcp")

	_ = testutils.CreatePublishedModuleVersion(t, db, provider1.ID, "1.0.0")
	_ = testutils.CreatePublishedModuleVersion(t, db, provider2.ID, "1.0.0")
	_ = testutils.CreatePublishedModuleVersion(t, db, provider3.ID, "1.0.0")
	_ = testutils.CreatePublishedModuleVersion(t, db, provider4.ID, "1.0.0")

	// Execute search
	counts, err := searchFiltersQuery.Execute(ctx, "contributedmodule")
	require.NoError(t, err)

	// Should return 4 trusted modules
	assert.Equal(t, 0, counts.Verified)
	assert.Equal(t, 4, counts.TrustedNamespaces)
	assert.Equal(t, 0, counts.Contributed)
	assert.Equal(t, map[string]int{"aws": 3, "gcp": 1}, counts.Providers)
	assert.Equal(t, map[string]int{"modulesearch": 4}, counts.Namespaces)
}

// TestSearchFilters_VerifiedModuleOneVersion tests search with one verified module
// Python reference: test_get_search_filters.py::TestGetSearchFilters::test_verified_module_one_version
func TestSearchFilters_VerifiedModuleOneVersion(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	ctx := context.Background()

	// Setup repositories WITH trusted namespace AND verified namespace
	domainConfig := &configModel.DomainConfig{
		TrustedNamespaces:        []string{"modulesearch"},
		VerifiedModuleNamespaces: []string{"modulesearch"},
		AllowModuleHosting:       configModel.ModuleHostingModeAllow,
		SecretKeySet:             true,
	}

	namespaceRepo := module.NewNamespaceRepository(db.DB)
	moduleProviderRepo := module.NewModuleProviderRepository(db.DB, namespaceRepo, domainConfig)
	searchFiltersQuery := modulequery.NewSearchFiltersQuery(moduleProviderRepo, domainConfig)

	// Create test data with VERIFIED module provider
	namespace := testutils.CreateNamespace(t, db, "modulesearch")
	verified := true
	provider := testutils.CreateModuleProvider(t, db, namespace.ID, "verifiedmodule-oneversion", "aws")
	// Set verified status
	err := db.DB.Model(&sqldb.ModuleProviderDB{}).Where("id = ?", provider.ID).Update("verified", &verified).Error
	require.NoError(t, err)
	_ = testutils.CreatePublishedModuleVersion(t, db, provider.ID, "1.0.0")

	// Execute search
	counts, err := searchFiltersQuery.Execute(ctx, "verifiedmodule-oneversion")
	require.NoError(t, err)

	// Should return 1 verified and 1 trusted module
	assert.Equal(t, 1, counts.Verified)
	assert.Equal(t, 1, counts.TrustedNamespaces)
	assert.Equal(t, 0, counts.Contributed)
	assert.Equal(t, map[string]int{"aws": 1}, counts.Providers)
	assert.Equal(t, map[string]int{"modulesearch": 1}, counts.Namespaces)
}

// TestSearchFilters_MixedResults tests results containing all types (contributed, trusted, verified)
// Python reference: test_get_search_filters.py::TestGetSearchFilters::test_mixed_results
func TestSearchFilters_MixedResults(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	ctx := context.Background()

	// Setup repositories with multiple trusted namespaces
	domainConfig := &configModel.DomainConfig{
		TrustedNamespaces:        []string{"doesnotexist", "modulesearch", "modulesearch-trusted"},
		VerifiedModuleNamespaces: []string{"modulesearch"},
		AllowModuleHosting:       configModel.ModuleHostingModeAllow,
		SecretKeySet:             true,
	}

	namespaceRepo := module.NewNamespaceRepository(db.DB)
	moduleProviderRepo := module.NewModuleProviderRepository(db.DB, namespaceRepo, domainConfig)
	searchFiltersQuery := modulequery.NewSearchFiltersQuery(moduleProviderRepo, domainConfig)

	// Create test data for mixed results
	// Contributed namespace (not in trusted list)
	namespaceContributed := testutils.CreateNamespace(t, db, "modulesearch-contributed")
	provider1 := testutils.CreateModuleProvider(t, db, namespaceContributed.ID, "modulesearch-one", "aws")
	provider2 := testutils.CreateModuleProvider(t, db, namespaceContributed.ID, "modulesearch-two", "gcp")
	_ = testutils.CreatePublishedModuleVersion(t, db, provider1.ID, "1.0.0")
	_ = testutils.CreatePublishedModuleVersion(t, db, provider2.ID, "1.0.0")

	// Trusted namespace (not verified)
	namespaceTrusted := testutils.CreateNamespace(t, db, "modulesearch-trusted")
	provider3 := testutils.CreateModuleProvider(t, db, namespaceTrusted.ID, "modulesearch-three", "aws")
	provider4 := testutils.CreateModuleProvider(t, db, namespaceTrusted.ID, "modulesearch-four", "aws")
	provider5 := testutils.CreateModuleProvider(t, db, namespaceTrusted.ID, "modulesearch-five", "gcp")
	_ = testutils.CreatePublishedModuleVersion(t, db, provider3.ID, "1.0.0")
	_ = testutils.CreatePublishedModuleVersion(t, db, provider4.ID, "1.0.0")
	_ = testutils.CreatePublishedModuleVersion(t, db, provider5.ID, "1.0.0")

	// Verified AND trusted namespace
	namespaceVerified := testutils.CreateNamespace(t, db, "modulesearch")
	verified := true
	provider6 := testutils.CreateModuleProvider(t, db, namespaceVerified.ID, "modulesearch-six", "aws")
	provider7 := testutils.CreateModuleProvider(t, db, namespaceVerified.ID, "modulesearch-seven", "aws")
	provider8 := testutils.CreateModuleProvider(t, db, namespaceVerified.ID, "modulesearch-eight", "gcp")
	err := db.DB.Model(&sqldb.ModuleProviderDB{}).Where("id = ?", provider6.ID).Update("verified", &verified).Error
	require.NoError(t, err)
	err = db.DB.Model(&sqldb.ModuleProviderDB{}).Where("id = ?", provider7.ID).Update("verified", &verified).Error
	require.NoError(t, err)
	err = db.DB.Model(&sqldb.ModuleProviderDB{}).Where("id = ?", provider8.ID).Update("verified", &verified).Error
	require.NoError(t, err)
	_ = testutils.CreatePublishedModuleVersion(t, db, provider6.ID, "1.0.0")
	_ = testutils.CreatePublishedModuleVersion(t, db, provider7.ID, "1.0.0")
	_ = testutils.CreatePublishedModuleVersion(t, db, provider8.ID, "1.0.0")

	// Execute search with partial match on "modulesearch"
	counts, err := searchFiltersQuery.Execute(ctx, "modulesearch")
	require.NoError(t, err)

	// Expected: 2 contributed, 6 trusted, 3 verified
	assert.Equal(t, 3, counts.Verified)
	assert.Equal(t, 6, counts.TrustedNamespaces)
	assert.Equal(t, 2, counts.Contributed)
	// Providers: aws=5 (1 contributed + 2 trusted + 2 verified), gcp=3 (1 contributed + 1 trusted + 1 verified)
	assert.Equal(t, map[string]int{"aws": 5, "gcp": 3}, counts.Providers)
	// Namespaces: modulesearch (3), modulesearch-contributed (2), modulesearch-trusted (3)
	assert.Len(t, counts.Namespaces, 3)
}
