package testutils

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/persistence/sqldb"
)

// SetupComprehensiveModuleSearchTestData creates comprehensive module search test data
// matching Python's integration_test_data.py.
// This creates 60+ module versions across multiple namespaces.
func SetupComprehensiveModuleSearchTestData(t *testing.T, db *sqldb.Database) {
	t.Helper()

	published := true
	beta := true
	internal := true

	// ===== modulesearch namespace (untrusted, not in TRUSTED_NAMESPACES) =====
	modulesearchNs := CreateNamespace(t, db, "modulesearch")

	// contributedmodule-oneversion
	provider1 := CreateModuleProvider(t, db, modulesearchNs.ID, "contributedmodule-oneversion", "aws")
	createVersion(t, db, provider1.ID, "1.0.0", &published, nil, "TestOwner1", "DESCRIPTION-Search-PUBLISHED")

	// contributedmodule-multiversion
	provider2 := CreateModuleProvider(t, db, modulesearchNs.ID, "contributedmodule-multiversion", "aws")
	createVersion(t, db, provider2.ID, "1.2.3", &published, nil, "TestOwner2", "DESCRIPTION-Search-OLDVERSION")
	createVersion(t, db, provider2.ID, "2.0.0", &published, nil, "", "")

	// contributedmodule-withbetaversion
	provider3 := CreateModuleProvider(t, db, modulesearchNs.ID, "contributedmodule-withbetaversion", "aws")
	createVersion(t, db, provider3.ID, "1.2.3", &published, nil, "", "")
	createVersion(t, db, provider3.ID, "2.0.0-beta", &published, &beta, "", "DESCRIPTION-Search-BETAVERSION")

	// contributedmodule-onlybeta
	provider4 := CreateModuleProvider(t, db, modulesearchNs.ID, "contributedmodule-onlybeta", "aws")
	createVersion(t, db, provider4.ID, "2.5.0-beta", &published, &beta, "", "")

	// contributedmodule-differentprovider
	provider5 := CreateModuleProvider(t, db, modulesearchNs.ID, "contributedmodule-differentprovider", "gcp")
	createVersion(t, db, provider5.ID, "1.2.3", &published, nil, "", "")

	// contributedmodule-unpublished
	provider6 := CreateModuleProvider(t, db, modulesearchNs.ID, "contributedmodule-unpublished", "aws")
	createVersion(t, db, provider6.ID, "1.0.0", nil, nil, "TestOwner1", "DESCRIPTION-Search-UNPUBLISHED")

	// verifiedmodule-oneversion
	provider7 := CreateModuleProviderWithVerified(t, db, modulesearchNs.ID, "verifiedmodule-oneversion", "aws", true)
	createVersion(t, db, provider7.ID, "1.0.0", &published, nil, "", "")

	// verifiedmodule-multiversion
	provider8 := CreateModuleProviderWithVerified(t, db, modulesearchNs.ID, "verifiedmodule-multiversion", "aws", true)
	createVersion(t, db, provider8.ID, "1.2.3", &published, nil, "", "")
	createVersion(t, db, provider8.ID, "2.0.0", &published, nil, "", "")

	// verifiedmodule-withbetaversion
	provider9 := CreateModuleProviderWithVerified(t, db, modulesearchNs.ID, "verifiedmodule-withbetaversion", "aws", true)
	createVersion(t, db, provider9.ID, "1.2.3", &published, nil, "", "")
	createVersion(t, db, provider9.ID, "2.0.0-beta", &published, &beta, "", "")

	// verifiedmodule-onybeta
	provider10 := CreateModuleProviderWithVerified(t, db, modulesearchNs.ID, "verifiedmodule-onybeta", "aws", true)
	createVersion(t, db, provider10.ID, "2.0.0-beta", &published, &beta, "", "")

	// verifiedmodule-differentprovider
	provider11 := CreateModuleProviderWithVerified(t, db, modulesearchNs.ID, "verifiedmodule-differentprovider", "gcp", true)
	createVersion(t, db, provider11.ID, "1.2.3", &published, nil, "", "")

	// verifiedmodule-unpublished
	provider12 := CreateModuleProviderWithVerified(t, db, modulesearchNs.ID, "verifiedmodule-unpublished", "aws", true)
	createVersion(t, db, provider12.ID, "1.0.0", nil, nil, "", "")

	// ===== searchbynamespace namespace =====
	searchbyNs := CreateNamespace(t, db, "searchbynamespace")

	// searchbymodulename1/searchbyprovideraws (verified)
	provider13 := CreateModuleProviderWithVerified(t, db, searchbyNs.ID, "searchbymodulename1", "searchbyprovideraws", true)
	createVersion(t, db, provider13.ID, "1.2.3", &published, nil, "", "")

	// searchbymodulename1/searchbyprovidergcp
	provider14 := CreateModuleProvider(t, db, searchbyNs.ID, "searchbymodulename1", "searchbyprovidergcp")
	createVersion(t, db, provider14.ID, "2.0.0", &published, nil, "", "")

	// searchbymodulename2/notpublished
	provider15 := CreateModuleProvider(t, db, searchbyNs.ID, "searchbymodulename2", "notpublished")
	createVersion(t, db, provider15.ID, "1.2.3", nil, nil, "", "")

	// searchbymodulename2/published
	provider16 := CreateModuleProvider(t, db, searchbyNs.ID, "searchbymodulename2", "published")
	createVersion(t, db, provider16.ID, "3.1.6", &published, nil, "", "")

	// ===== searchbynamesp-similar namespace =====
	searchbySimilarNs := CreateNamespace(t, db, "searchbynamesp-similar")

	// searchbymodulename3/searchbyprovideraws (verified)
	provider17 := CreateModuleProviderWithVerified(t, db, searchbySimilarNs.ID, "searchbymodulename3", "searchbyprovideraws", true)
	createVersion(t, db, provider17.ID, "4.4.1", &published, nil, "", "")

	// searchbymodulename4/aws
	provider18 := CreateModuleProvider(t, db, searchbySimilarNs.ID, "searchbymodulename4", "aws")
	createVersion(t, db, provider18.ID, "5.5.5", &published, nil, "", "")

	// ===== testnamespace (from unit tests) =====
	testNs := CreateNamespace(t, db, "testnamespace")

	// testmodulename/testprovider
	provider19 := CreateModuleProvider(t, db, testNs.ID, "testmodulename", "testprovider")
	createVersion(t, db, provider19.ID, "2.4.1", &published, nil, "", "")
	createVersion(t, db, provider19.ID, "1.0.0", &published, nil, "", "")

	// lonelymodule/testprovider
	provider20 := CreateModuleProviderWithVerified(t, db, testNs.ID, "lonelymodule", "testprovider", true)
	createVersion(t, db, provider20.ID, "1.0.0", &published, nil, "", "")

	// mock-module/testprovider
	provider21 := CreateModuleProviderWithVerified(t, db, testNs.ID, "mock-module", "testprovider", true)
	createVersion(t, db, provider21.ID, "1.2.3", &published, nil, "", "")

	// unverifiedmodule/testprovider
	provider22 := CreateModuleProvider(t, db, testNs.ID, "unverifiedmodule", "testprovider")
	createVersion(t, db, provider22.ID, "1.2.3", &published, nil, "", "")

	// internalmodule/testprovider
	provider23 := CreateModuleProvider(t, db, testNs.ID, "internalmodule", "testprovider")
	createVersion(t, db, provider23.ID, "5.2.0", &published, nil, "", "")

	// modulenorepourl/testprovider
	provider24 := CreateModuleProvider(t, db, testNs.ID, "modulenorepourl", "testprovider")
	createVersion(t, db, provider24.ID, "2.2.4", &published, nil, "", "")

	// onlybeta/testprovider
	provider25 := CreateModuleProvider(t, db, testNs.ID, "onlybeta", "testprovider")
	createVersion(t, db, provider25.ID, "2.2.4-beta", &published, &beta, "", "")

	// modulewithrepourl/testprovider
	provider26 := CreateModuleProvider(t, db, testNs.ID, "modulewithrepourl", "testprovider")
	createVersion(t, db, provider26.ID, "2.1.0", nil, nil, "", "")

	// modulenotpublished/testprovider
	provider27 := CreateModuleProvider(t, db, testNs.ID, "modulenotpublished", "testprovider")
	createVersion(t, db, provider27.ID, "10.2.1", nil, nil, "", "")

	// ===== real_providers namespace =====
	realNs := CreateNamespace(t, db, "real_providers")

	// test-module/aws
	provider28 := CreateModuleProvider(t, db, realNs.ID, "test-module", "aws")
	createVersion(t, db, provider28.ID, "1.0.0", nil, nil, "", "")

	// test-module/gcp
	provider29 := CreateModuleProvider(t, db, realNs.ID, "test-module", "gcp")
	createVersion(t, db, provider29.ID, "1.0.0", nil, nil, "", "")

	// test-module/null
	provider30 := CreateModuleProvider(t, db, realNs.ID, "test-module", "null")
	createVersion(t, db, provider30.ID, "1.0.0", nil, nil, "", "")

	// test-module/doesnotexist
	provider31 := CreateModuleProvider(t, db, realNs.ID, "test-module", "doesnotexist")
	createVersion(t, db, provider31.ID, "1.0.0", nil, nil, "", "")

	// ===== genericmodules namespace =====
	genericNs := CreateNamespace(t, db, "genericmodules")

	// modulename/providername
	provider32 := CreateModuleProvider(t, db, genericNs.ID, "modulename", "providername")
	createVersion(t, db, provider32.ID, "1.2.3", &published, nil, "", "")

	// ===== Additional search test modules for comprehensive testing =====
	searchNs1 := CreateNamespace(t, db, "searchnamespace1")

	// Various modules for multi-term and description search testing
	provider33 := CreateModuleProvider(t, db, searchNs1.ID, "aws-vpc-module", "aws")
	createVersion(t, db, provider33.ID, "1.0.0", &published, nil, "terraform-aws-modules", "VPC module for AWS infrastructure")

	provider34 := CreateModuleProvider(t, db, searchNs1.ID, "vpc-module", "gcp")
	createVersion(t, db, provider34.ID, "1.0.0", &published, nil, "", "")

	provider35 := CreateModuleProvider(t, db, searchNs1.ID, "aws-module", "azure")
	createVersion(t, db, provider35.ID, "1.0.0", &published, nil, "", "")

	provider36 := CreateModuleProvider(t, db, searchNs1.ID, "networking-firewall", "aws")
	createVersion(t, db, provider36.ID, "1.2.0", &published, nil, "", "Firewall module for VPC networking")

	provider37 := CreateModuleProvider(t, db, searchNs1.ID, "compute-instance", "aws")
	createVersion(t, db, provider37.ID, "2.0.0", &published, nil, "", "EC2 instance management module")

	// Module with internal versions
	provider38 := CreateModuleProvider(t, db, searchNs1.ID, "internal-test", "aws")
	createVersion(t, db, provider38.ID, "1.0.0", &published, nil, "", "")
	createVersion(t, db, provider38.ID, "2.0.0-internal", nil, &internal, "", "Internal development version")

	// Module for description and owner search
	provider39 := CreateModuleProvider(t, db, searchNs1.ID, "custom-auth-module", "aws")
	createVersion(t, db, provider39.ID, "3.0.0", &published, nil, "CustomAuthTeam", "Custom authentication provider module")

	// ===== Additional namespace for pagination testing =====
	largeNs := CreateNamespace(t, db, "large-search-ns")

	for i := 1; i <= 15; i++ {
		provider := CreateModuleProvider(t, db, largeNs.ID, fmt.Sprintf("search-module-%d", i), fmt.Sprintf("provider-%d", i))
		createVersion(t, db, provider.ID, fmt.Sprintf("1.%d.0", i), &published, nil, "", "")
	}

	// ===== modulesearch-contributed namespace (from Python test_data.py) =====
	// This namespace tests "contributed" search results (not verified, not in trusted namespaces)
	modulesearchContributedNs := CreateNamespace(t, db, "modulesearch-contributed")

	// mixedsearch-result (published, single version)
	provider40 := CreateModuleProvider(t, db, modulesearchContributedNs.ID, "mixedsearch-result", "aws")
	createVersion(t, db, provider40.ID, "1.0.0", &published, nil, "", "")

	// mixedsearch-result-multiversion (published, multiple versions - IMPORTANT for duplicate bug testing)
	provider41 := CreateModuleProvider(t, db, modulesearchContributedNs.ID, "mixedsearch-result-multiversion", "aws")
	createVersion(t, db, provider41.ID, "1.2.3", &published, nil, "", "")
	createVersion(t, db, provider41.ID, "2.0.0", &published, nil, "", "")

	// mixedsearch-result-unpublished (unpublished)
	provider42 := CreateModuleProvider(t, db, modulesearchContributedNs.ID, "mixedsearch-result-unpublished", "aws")
	createVersion(t, db, provider42.ID, "1.2.3", nil, nil, "", "")
	createVersion(t, db, provider42.ID, "2.0.0", nil, nil, "", "")

	// ===== modulesearch-trusted namespace (from Python test_data.py) =====
	// This namespace tests "trusted" search results (namespaces configured as trusted)
	// Note: The actual "trusted" status is configured via TRUSTED_NAMESPACES in config
	modulesearchTrustedNs := CreateNamespace(t, db, "modulesearch-trusted")

	// mixedsearch-trusted-result (published, single version)
	provider43 := CreateModuleProvider(t, db, modulesearchTrustedNs.ID, "mixedsearch-trusted-result", "aws")
	createVersion(t, db, provider43.ID, "1.0.0", &published, nil, "", "")

	// mixedsearch-trusted-second-result (published, single version)
	provider44 := CreateModuleProvider(t, db, modulesearchTrustedNs.ID, "mixedsearch-trusted-second-result", "aws")
	createVersion(t, db, provider44.ID, "5.2.1", &published, nil, "", "")

	// mixedsearch-trusted-result-multiversion (published, multiple versions - IMPORTANT for duplicate bug testing)
	provider45 := CreateModuleProvider(t, db, modulesearchTrustedNs.ID, "mixedsearch-trusted-result-multiversion", "aws")
	createVersion(t, db, provider45.ID, "1.2.3", &published, nil, "", "")
	createVersion(t, db, provider45.ID, "2.0.0", &published, nil, "", "")

	// mixedsearch-trusted-result-unpublished (unpublished)
	provider46 := CreateModuleProvider(t, db, modulesearchTrustedNs.ID, "mixedsearch-trusted-result-unpublished", "aws")
	createVersion(t, db, provider46.ID, "1.2.3", nil, nil, "", "")
	createVersion(t, db, provider46.ID, "2.0.0", nil, nil, "", "")

	// ===== Additional testnamespace modules from Python unit tests =====
	// These are important for testing edge cases like wrong version order, no versions, etc.

	// wrongversionorder/testprovider - tests version sorting with various version formats
	provider47 := CreateModuleProvider(t, db, testNs.ID, "wrongversionorder", "testprovider")
	createVersion(t, db, provider47.ID, "1.5.4", &published, nil, "", "")
	createVersion(t, db, provider47.ID, "2.1.0", &published, nil, "", "")
	createVersion(t, db, provider47.ID, "0.1.1", &published, nil, "", "")
	createVersion(t, db, provider47.ID, "10.23.0", &published, nil, "", "")
	createVersion(t, db, provider47.ID, "0.1.10", &published, nil, "", "")
	createVersion(t, db, provider47.ID, "0.0.9", &published, nil, "", "")
	createVersion(t, db, provider47.ID, "0.1.09", &published, nil, "", "")
	createVersion(t, db, provider47.ID, "0.1.8", &published, nil, "", "")
	createVersion(t, db, provider47.ID, "23.2.3-beta", &published, &beta, "", "")
	// Unpublished version
	createVersion(t, db, provider47.ID, "5.21.2", nil, nil, "", "")

	// noversions/testprovider - module with no versions
	_ = CreateModuleProvider(t, db, testNs.ID, "noversions", "testprovider")
	// No versions created - intentionally unused to test modules without versions

	// onlyunpublished/testprovider - module with only unpublished versions
	provider49 := CreateModuleProvider(t, db, testNs.ID, "onlyunpublished", "testprovider")
	createVersion(t, db, provider49.ID, "0.1.8", nil, nil, "", "")

	// onlybeta/testprovider - module with only beta versions
	provider50 := CreateModuleProvider(t, db, testNs.ID, "onlybeta", "testprovider")
	createVersion(t, db, provider50.ID, "2.5.0-beta", &published, &beta, "", "")
}

// createVersion is a helper to create a module version with common attributes
// It automatically sets the version as the latest version on the module provider
func createVersion(t *testing.T, db *sqldb.Database, moduleProviderID int, version string,
	published, beta *bool, owner, description string) {
	t.Helper()

	moduleVersion := sqldb.ModuleVersionDB{
		ModuleProviderID: moduleProviderID,
		Version:          version,
		Beta:             false,
		Internal:         false,
		Published:        published,
	}

	if owner != "" {
		moduleVersion.Owner = &owner
	}
	if description != "" {
		moduleVersion.Description = &description
	}
	if beta != nil {
		moduleVersion.Beta = *beta
	}

	err := db.DB.Create(&moduleVersion).Error
	require.NoError(t, err)

	// Set this version as the latest version for the module provider
	// This is required for the search query to find the module
	err = db.DB.Model(&sqldb.ModuleProviderDB{}).
		Where("id = ?", moduleProviderID).
		Update("latest_version_id", moduleVersion.ID).Error
	require.NoError(t, err)
}

// CreateModuleProviderWithVerified creates a module provider with specified verified status
func CreateModuleProviderWithVerified(t *testing.T, db *sqldb.Database, namespaceID int, moduleName, providerName string, verified bool) sqldb.ModuleProviderDB {
	t.Helper()

	moduleProvider := sqldb.ModuleProviderDB{
		NamespaceID: namespaceID,
		Module:      moduleName,
		Provider:    providerName,
		Verified:    &verified,
	}

	err := db.DB.Create(&moduleProvider).Error
	require.NoError(t, err)

	return moduleProvider
}

// SetupComprehensiveProviderSearchTestData creates comprehensive provider search test data
// matching Python's integration_test_data.py provider search data.
// This creates providers in providersearch and contributed-providersearch namespaces.
func SetupComprehensiveProviderSearchTestData(t *testing.T, db *sqldb.Database) {
	t.Helper()

	// Create provider categories (matching Python's integration_provider_categories)
	createProviderCategory(t, db, "Visible Monitoring", "visible-monitoring", true)
	createProviderCategory(t, db, "Second Visible Cloud", "second-visible-cloud", true)

	// Create providersearch namespace (for trusted providers)
	providersearchNs := CreateNamespace(t, db, "providersearch")

	// Create contributed-providersearch namespace (for contributed providers)
	contributedProvidersearchNs := CreateNamespace(t, db, "contributed-providersearch")

	// Create GPG keys directly in namespaces (not linked to providers yet)
	gpgKeyProviderSearch := createGPGKeyInNamespace(t, db, providersearchNs.ID, "D8A89D97BB7526F33C8A2D8C39C57A3D0D24B532")

	gpgKeyContributed := createGPGKeyInNamespace(t, db, contributedProvidersearchNs.ID, "D7AA1BEFF16FA788760E54F5591EF84DC5EDCD68")

	// Get category IDs
	visibleMonitoringCat := getProviderCategoryBySlug(t, db, "visible-monitoring")
	secondVisibleCloudCat := getProviderCategoryBySlug(t, db, "second-visible-cloud")

	// ===== providersearch namespace providers =====
	// contributedprovider-oneversion (one version)
	provider1 := CreateProvider(t, db, providersearchNs.ID, "contributedprovider-oneversion",
		stringPtr("DESCRIPTION-Search"), sqldb.ProviderTierOfficial, &visibleMonitoringCat)
	createProviderVersion(t, db, provider1.ID, "1.2.0", gpgKeyProviderSearch.ID, false)

	// contributedprovider-multiversion (multiple versions)
	provider2 := CreateProvider(t, db, providersearchNs.ID, "contributedprovider-multiversion",
		stringPtr("DESCRIPTION-MultiVersion"), sqldb.ProviderTierOfficial, &secondVisibleCloudCat)
	createProviderVersion(t, db, provider2.ID, "1.2.0", gpgKeyProviderSearch.ID, false)
	createProviderVersion(t, db, provider2.ID, "1.3.0", gpgKeyProviderSearch.ID, false)

	// ===== contributed-providersearch namespace providers =====
	// mixedsearch-result (one version)
	provider3 := CreateProvider(t, db, contributedProvidersearchNs.ID, "mixedsearch-result",
		stringPtr("Test Multiple Versions"), sqldb.ProviderTierCommunity, &visibleMonitoringCat)
	createProviderVersion(t, db, provider3.ID, "1.0.0", gpgKeyContributed.ID, false)

	// mixedsearch-result-multiversion (multiple versions - IMPORTANT for duplicate bug testing)
	provider4 := CreateProvider(t, db, contributedProvidersearchNs.ID, "mixedsearch-result-multiversion",
		stringPtr("Test Multiple Versions"), sqldb.ProviderTierCommunity, &visibleMonitoringCat)
	createProviderVersion(t, db, provider4.ID, "1.2.3", gpgKeyContributed.ID, false)
	createProviderVersion(t, db, provider4.ID, "2.0.0", gpgKeyContributed.ID, false)

	// mixedsearch-result-no-version (no versions - should be excluded from search)
	_ = CreateProvider(t, db, contributedProvidersearchNs.ID, "mixedsearch-result-no-version",
		stringPtr("DESCRIPTION-NoVersion"), sqldb.ProviderTierCommunity, &visibleMonitoringCat)
}

// createGPGKeyInNamespace creates a GPG key directly in a namespace (not linked to a provider)
func createGPGKeyInNamespace(t *testing.T, db *sqldb.Database, namespaceID int, fingerprint string) sqldb.GPGKeyDB {
	t.Helper()

	asciiArmor := []byte("-----BEGIN PGP PUBLIC KEY BLOCK-----\n\nTest ASCII armor for " + fingerprint + "\n-----END PGP PUBLIC KEY BLOCK-----")
	source := "test-source"
	keyID := &fingerprint

	gpgKey := sqldb.GPGKeyDB{
		NamespaceID: namespaceID,
		ASCIIArmor:  asciiArmor,
		KeyID:       keyID,
		Fingerprint: keyID,
		Source:      &source,
	}

	err := db.DB.Create(&gpgKey).Error
	require.NoError(t, err)

	return gpgKey
}

// CreateProviderWithGPGKey creates a provider with GPG key ID (instead of creating a new GPG key)
func CreateProviderWithGPGKey(t *testing.T, db *sqldb.Database, namespaceID int, name string, description *string, tier sqldb.ProviderTier, categoryID *int, gpgKeyID int) sqldb.ProviderDB {
	t.Helper()

	provider := sqldb.ProviderDB{
		NamespaceID:         namespaceID,
		Name:                name,
		Description:         description,
		Tier:                tier,
		ProviderCategoryID:  categoryID,
	}

	err := db.DB.Create(&provider).Error
	require.NoError(t, err)

	return provider
}

// createProviderVersion is a helper to create a provider version with GPG key
// It automatically sets the version as the latest version on the provider
func createProviderVersion(t *testing.T, db *sqldb.Database, providerID int, version string, gpgKeyID int, beta bool) {
	t.Helper()

	gitTag := "v" + version
	publishedAt := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	providerVersion := sqldb.ProviderVersionDB{
		ProviderID:  providerID,
		Version:     version,
		GitTag:      &gitTag,
		GPGKeyID:    gpgKeyID,
		PublishedAt: &publishedAt,
		Beta:        beta,
	}

	err := db.DB.Create(&providerVersion).Error
	require.NoError(t, err)

	// Set this version as the latest version for the provider
	err = db.DB.Model(&sqldb.ProviderDB{}).
		Where("id = ?", providerID).
		Update("latest_version_id", providerVersion.ID).Error
	require.NoError(t, err)
}

// createProviderCategory creates a provider category
func createProviderCategory(t *testing.T, db *sqldb.Database, name, slug string, userSelectable bool) sqldb.ProviderCategoryDB {
	t.Helper()

	namePtr := &name
	category := sqldb.ProviderCategoryDB{
		Name:           namePtr,
		Slug:           slug,
		UserSelectable: userSelectable,
	}

	err := db.DB.Create(&category).Error
	require.NoError(t, err)

	return category
}

// getProviderCategoryBySlug gets a provider category by slug
func getProviderCategoryBySlug(t *testing.T, db *sqldb.Database, slug string) int {
	t.Helper()

	var category sqldb.ProviderCategoryDB
	err := db.DB.Where("slug = ?", slug).First(&category).Error
	require.NoError(t, err)
	return category.ID
}

// stringPtr returns a pointer to a string
func stringPtr(s string) *string {
	return &s
}
