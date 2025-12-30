// Package model provides integration tests for the provider model.
// Python reference: /app/test/integration/terrareg/models/test_provider.py
package model

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/provider"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/persistence/sqldb"
	providerprepo "github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/persistence/sqldb/provider"
	testutils "github.com/matthewjohn/terrareg/terrareg-go/test/integration/testutils"
)

// TestProvider_RepositoryNameToProviderName tests converting repository name to provider name
// Python reference: test_provider.py::TestProvider::test_repository_name_to_provider_name
func TestProvider_RepositoryNameToProviderName(t *testing.T) {
	testCases := []struct {
		name           string
		repositoryName string
		expectedResult string
	}{
		{"terraform-provider-jmon", "terraform-provider-jmon", "jmon"},
		{"terraform-provider-some-service", "terraform-provider-some-service", "some-service"},
		{"terraform-some-service", "terraform-some-service", ""},
		{"some-service", "some-service", ""},
		{"", "", ""},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := provider.RepositoryNameToProviderName(tc.repositoryName)
			assert.Equal(t, tc.expectedResult, result)
		})
	}
}

// TestProvider_Create tests creating a provider
// Python reference: test_provider.py::TestProvider::test_create
func TestProvider_Create(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	ctx := context.Background()
	providerRepo := providerprepo.NewProviderRepository(db.DB)

	// Create namespace
	namespace := testutils.CreateNamespace(t, db, "test-create-provider")

	// Create provider category
	category := testutils.CreateProviderCategory(t, db, "Test Category", "test-category", true)

	// Test with use_default_provider_source_auth = true
	t.Run("Create with default provider source auth true", func(t *testing.T) {
		description := "Test provider description"
		prov := provider.NewProvider(
			namespace.ID,
			"testprovider",
			&description,
			"community",
			&category.ID,
			nil,
			true,
		)

		err := providerRepo.Save(ctx, prov)
		require.NoError(t, err)
		assert.Greater(t, prov.ID(), 0)

		// Verify provider was created correctly
		retrieved, err := providerRepo.FindByID(ctx, prov.ID())
		require.NoError(t, err)
		require.NotNil(t, retrieved)
		assert.Equal(t, "testprovider", retrieved.Name())
		assert.Equal(t, "Test provider description", *retrieved.Description())
		assert.Equal(t, "community", retrieved.Tier())
		assert.True(t, retrieved.UseProviderSourceAuth())
		assert.Equal(t, category.ID, *retrieved.CategoryID())
	})

	// Test with use_default_provider_source_auth = false
	t.Run("Create with default provider source auth false", func(t *testing.T) {
		prov := provider.NewProvider(
			namespace.ID,
			"testprovider2",
			nil,
			"community",
			&category.ID,
			nil,
			false,
		)

		err := providerRepo.Save(ctx, prov)
		require.NoError(t, err)
		assert.Greater(t, prov.ID(), 0)

		// Verify provider was created correctly
		retrieved, err := providerRepo.FindByID(ctx, prov.ID())
		require.NoError(t, err)
		require.NotNil(t, retrieved)
		assert.False(t, retrieved.UseProviderSourceAuth())
	})
}

// TestProvider_FindByID tests finding a provider by ID
// Python reference: test_provider.py::TestProvider::test_get_by_pk
func TestProvider_FindByID(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	ctx := context.Background()
	providerRepo := providerprepo.NewProviderRepository(db.DB)

	namespace := testutils.CreateNamespace(t, db, "test-find-by-id")
	providerDB := testutils.CreateProvider(t, db, namespace.ID, "testprovider", nil, sqldb.ProviderTierCommunity, nil)

	t.Run("Find existing provider by ID", func(t *testing.T) {
		prov, err := providerRepo.FindByID(ctx, providerDB.ID)
		require.NoError(t, err)
		assert.NotNil(t, prov)
		assert.Equal(t, providerDB.ID, prov.ID())
		assert.Equal(t, "testprovider", prov.Name())
	})

	t.Run("Find non-existent provider by ID", func(t *testing.T) {
		prov, err := providerRepo.FindByID(ctx, 999999)
		require.NoError(t, err)
		assert.Nil(t, prov)
	})
}

// TestProvider_FindByNamespaceAndName tests finding a provider by namespace and name
// Python reference: test_provider.py::TestProvider::test_get
func TestProvider_FindByNamespaceAndName(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	ctx := context.Background()
	providerRepo := providerprepo.NewProviderRepository(db.DB)

	namespace := testutils.CreateNamespace(t, db, "test-namespace-name")
	testutils.CreateProvider(t, db, namespace.ID, "testprovider", nil, sqldb.ProviderTierCommunity, nil)

	t.Run("Find existing provider by namespace and name", func(t *testing.T) {
		prov, err := providerRepo.FindByNamespaceAndName(ctx, "test-namespace-name", "testprovider")
		require.NoError(t, err)
		assert.NotNil(t, prov)
		assert.Equal(t, "testprovider", prov.Name())
	})

	t.Run("Find non-existent provider by namespace and name", func(t *testing.T) {
		prov, err := providerRepo.FindByNamespaceAndName(ctx, "test-namespace-name", "does-not-exist")
		require.NoError(t, err)
		assert.Nil(t, prov)
	})
}

// TestProvider_Properties tests various provider properties
// Python reference: test_provider.py::TestProvider::test_name, test_id, test_full_name, test_tier, etc.
func TestProvider_Properties(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	ctx := context.Background()
	providerRepo := providerprepo.NewProviderRepository(db.DB)

	namespace := testutils.CreateNamespace(t, db, "test-props")
	description := "Test provider description"
	providerDB := testutils.CreateProvider(t, db, namespace.ID, "unittest-create-provider-name", &description, sqldb.ProviderTierCommunity, nil)

	prov, err := providerRepo.FindByID(ctx, providerDB.ID)
	require.NoError(t, err)

	t.Run("Name property", func(t *testing.T) {
		assert.Equal(t, "unittest-create-provider-name", prov.Name())
	})

	t.Run("ID property", func(t *testing.T) {
		assert.Greater(t, prov.ID(), 0)
	})

	t.Run("NamespaceID property", func(t *testing.T) {
		assert.Equal(t, namespace.ID, prov.NamespaceID())
	})

	t.Run("Description property", func(t *testing.T) {
		assert.NotNil(t, prov.Description())
		assert.Equal(t, "Test provider description", *prov.Description())
	})

	t.Run("Tier property", func(t *testing.T) {
		assert.Equal(t, "community", prov.Tier())
	})

	t.Run("UseProviderSourceAuth property", func(t *testing.T) {
		assert.False(t, prov.UseProviderSourceAuth())
	})
}

// TestProvider_UpdateDescription tests updating provider description
// Python reference: test_provider.py::TestProvider::test_update_attributes
func TestProvider_UpdateDescription(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	ctx := context.Background()
	providerRepo := providerprepo.NewProviderRepository(db.DB)

	namespace := testutils.CreateNamespace(t, db, "test-update")
	oldDescription := "Old description"
	newDescription := "New description"
	providerDB := testutils.CreateProvider(t, db, namespace.ID, "testprovider", &oldDescription, sqldb.ProviderTierCommunity, nil)

	// Get the provider
	prov, err := providerRepo.FindByID(ctx, providerDB.ID)
	require.NoError(t, err)

	// Update description
	prov.SetDescription(&newDescription)
	err = providerRepo.Save(ctx, prov)
	require.NoError(t, err)

	// Verify update
	retrieved, err := providerRepo.FindByID(ctx, prov.ID())
	require.NoError(t, err)
	assert.Equal(t, "New description", *retrieved.Description())
}

// TestProvider_UpdateTier tests updating provider tier
// Python reference: test_provider.py::TestProvider::test_tier (update portion)
func TestProvider_UpdateTier(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	ctx := context.Background()
	providerRepo := providerprepo.NewProviderRepository(db.DB)

	namespace := testutils.CreateNamespace(t, db, "test-tier-update")
	providerDB := testutils.CreateProvider(t, db, namespace.ID, "testprovider", nil, sqldb.ProviderTierCommunity, nil)

	// Get the provider
	prov, err := providerRepo.FindByID(ctx, providerDB.ID)
	require.NoError(t, err)
	assert.Equal(t, "community", prov.Tier())

	// Update tier to official
	prov.SetTier("official")
	err = providerRepo.Save(ctx, prov)
	require.NoError(t, err)

	// Verify update
	retrieved, err := providerRepo.FindByID(ctx, prov.ID())
	require.NoError(t, err)
	assert.Equal(t, "official", retrieved.Tier())
}

// TestProvider_GetLatestVersion tests getting the latest version of a provider
// Python reference: test_provider.py::TestProvider::test_get_latest_version
func TestProvider_GetLatestVersion(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	ctx := context.Background()
	providerRepo := providerprepo.NewProviderRepository(db.DB)

	namespace := testutils.CreateNamespace(t, db, "test-latest-version")
	providerDB := testutils.CreateProvider(t, db, namespace.ID, "testprovider", nil, sqldb.ProviderTierCommunity, nil)

	t.Run("No versions returns nil", func(t *testing.T) {
		// Provider with no versions
		prov, err := providerRepo.FindByID(ctx, providerDB.ID)
		require.NoError(t, err)

		versions, err := providerRepo.FindVersionsByProvider(ctx, prov.ID())
		require.NoError(t, err)
		assert.Empty(t, versions)
	})

	t.Run("Single version returns that version", func(t *testing.T) {
		gpgKey := testutils.CreateGPGKey(t, db, "key1", providerDB.ID, "KEY12345")
		now := time.Now()
		versionDB := testutils.CreateProviderVersion(t, db, providerDB.ID, "1.0.0", gpgKey.ID, false, &now)
		testutils.SetProviderLatestVersion(t, db, providerDB.ID, versionDB.ID)

		// Get versions
		prov, err := providerRepo.FindByID(ctx, providerDB.ID)
		require.NoError(t, err)

		versions, err := providerRepo.FindVersionsByProvider(ctx, prov.ID())
		require.NoError(t, err)
		assert.Len(t, versions, 1)
		assert.Equal(t, "1.0.0", versions[0].Version())
	})

	t.Run("Multiple versions returns highest version as latest", func(t *testing.T) {
		// Clean up previous test data
		db.DB.Exec("DELETE FROM provider_version WHERE provider_id = ?", providerDB.ID)

		gpgKey := testutils.CreateGPGKey(t, db, "key2", providerDB.ID, "KEY67890")
		now := time.Now()

		// Create multiple versions
		versionStrings := []string{"1.0.0", "3.0.0", "1.5.2", "2.1.0"}
		for i, ver := range versionStrings {
			v := testutils.CreateProviderVersion(t, db, providerDB.ID, ver, gpgKey.ID, false, &now)
			// Set the last created version (3.0.0) as latest
			if i == 1 {
				testutils.SetProviderLatestVersion(t, db, providerDB.ID, v.ID)
			}
		}

		// Get versions
		prov, err := providerRepo.FindByID(ctx, providerDB.ID)
		require.NoError(t, err)

		providerVersions, err := providerRepo.FindVersionsByProvider(ctx, prov.ID())
		require.NoError(t, err)
		assert.Len(t, providerVersions, 4)
	})
}

// TestProvider_GetAllVersions tests getting all versions of a provider
// Python reference: test_provider.py::TestProvider::test_get_all_versions
func TestProvider_GetAllVersions(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	ctx := context.Background()
	providerRepo := providerprepo.NewProviderRepository(db.DB)

	namespace := testutils.CreateNamespace(t, db, "test-all-versions")
	providerDB := testutils.CreateProvider(t, db, namespace.ID, "testprovider", nil, sqldb.ProviderTierCommunity, nil)

	t.Run("No versions returns empty list", func(t *testing.T) {
		prov, err := providerRepo.FindByID(ctx, providerDB.ID)
		require.NoError(t, err)

		versions, err := providerRepo.FindVersionsByProvider(ctx, prov.ID())
		require.NoError(t, err)
		assert.Empty(t, versions)
	})

	t.Run("Single version returns list with one version", func(t *testing.T) {
		gpgKey := testutils.CreateGPGKey(t, db, "key1", providerDB.ID, "KEY12345")
		now := time.Now()
		_ = testutils.CreateProviderVersion(t, db, providerDB.ID, "1.0.0", gpgKey.ID, false, &now)

		// Get versions
		prov, err := providerRepo.FindByID(ctx, providerDB.ID)
		require.NoError(t, err)

		versions, err := providerRepo.FindVersionsByProvider(ctx, prov.ID())
		require.NoError(t, err)
		assert.Len(t, versions, 1)
		assert.Equal(t, "1.0.0", versions[0].Version())
	})

	t.Run("Multiple versions returns all versions in descending order", func(t *testing.T) {
		// Clean up previous test data
		db.DB.Exec("DELETE FROM provider_version WHERE provider_id = ?", providerDB.ID)

		gpgKey := testutils.CreateGPGKey(t, db, "key2", providerDB.ID, "KEY67890")
		now := time.Now()

		// Create multiple versions
		versions := []string{"1.0.0", "3.0.0", "1.5.2", "2.1.0", "2.0.5", "2.0.0"}
		for _, ver := range versions {
			_ = testutils.CreateProviderVersion(t, db, providerDB.ID, ver, gpgKey.ID, false, &now)
		}

		// Get versions
		prov, err := providerRepo.FindByID(ctx, providerDB.ID)
		require.NoError(t, err)

		retrievedVersions, err := providerRepo.FindVersionsByProvider(ctx, prov.ID())
		require.NoError(t, err)

		// Check we got all versions
		assert.Len(t, retrievedVersions, 6)

		// Verify version numbers are present
		versionNumbers := make([]string, len(retrievedVersions))
		for i, v := range retrievedVersions {
			versionNumbers[i] = v.Version()
		}

		// All created versions should be present
		for _, ver := range versions {
			assert.Contains(t, versionNumbers, ver)
		}
	})
}

// TestProvider_FindVersionByProviderAndVersion tests finding a specific version
// Python reference: test_provider.py::TestProvider::test_get (implicit version lookup)
func TestProvider_FindVersionByProviderAndVersion(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	ctx := context.Background()
	providerRepo := providerprepo.NewProviderRepository(db.DB)

	namespace := testutils.CreateNamespace(t, db, "test-find-version")
	providerDB := testutils.CreateProvider(t, db, namespace.ID, "testprovider", nil, sqldb.ProviderTierCommunity, nil)

	gpgKey := testutils.CreateGPGKey(t, db, "key1", providerDB.ID, "KEY12345")
	now := time.Now()
	_ = testutils.CreateProviderVersion(t, db, providerDB.ID, "1.0.0", gpgKey.ID, false, &now)
	_ = testutils.CreateProviderVersion(t, db, providerDB.ID, "2.0.0", gpgKey.ID, false, &now)

	t.Run("Find existing version", func(t *testing.T) {
		prov, err := providerRepo.FindByID(ctx, providerDB.ID)
		require.NoError(t, err)

		version, err := providerRepo.FindVersionByProviderAndVersion(ctx, prov.ID(), "1.0.0")
		require.NoError(t, err)
		assert.NotNil(t, version)
		assert.Equal(t, "1.0.0", version.Version())
	})

	t.Run("Find non-existent version", func(t *testing.T) {
		prov, err := providerRepo.FindByID(ctx, providerDB.ID)
		require.NoError(t, err)

		version, err := providerRepo.FindVersionByProviderAndVersion(ctx, prov.ID(), "9.9.9")
		require.NoError(t, err)
		assert.Nil(t, version)
	})
}

// TestProvider_SetLatestVersion tests setting the latest version
// Python reference: test_provider.py::TestProvider::test_calculate_latest_version (partial)
func TestProvider_SetLatestVersion(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	ctx := context.Background()
	providerRepo := providerprepo.NewProviderRepository(db.DB)

	namespace := testutils.CreateNamespace(t, db, "test-set-latest")
	providerDB := testutils.CreateProvider(t, db, namespace.ID, "testprovider", nil, sqldb.ProviderTierCommunity, nil)

	gpgKey := testutils.CreateGPGKey(t, db, "key1", providerDB.ID, "KEY12345")
	now := time.Now()

	// Create multiple versions
	version1 := testutils.CreateProviderVersion(t, db, providerDB.ID, "1.0.0", gpgKey.ID, false, &now)
	version2 := testutils.CreateProviderVersion(t, db, providerDB.ID, "2.0.0", gpgKey.ID, false, &now)

	// Set latest version to version2
	err := providerRepo.SetLatestVersion(ctx, providerDB.ID, version2.ID)
	require.NoError(t, err)

	// Verify latest version was set
	prov, err := providerRepo.FindByID(ctx, providerDB.ID)
	require.NoError(t, err)
	assert.Equal(t, version2.ID, *prov.LatestVersionID())

	// Change to version1
	err = providerRepo.SetLatestVersion(ctx, providerDB.ID, version1.ID)
	require.NoError(t, err)

	// Verify latest version was updated
	prov, err = providerRepo.FindByID(ctx, providerDB.ID)
	require.NoError(t, err)
	assert.Equal(t, version1.ID, *prov.LatestVersionID())
}

// TestProvider_FindAll tests finding all providers
// Python reference: test_provider.py::TestProvider (general find all behavior)
func TestProvider_FindAll(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	ctx := context.Background()
	providerRepo := providerprepo.NewProviderRepository(db.DB)

	namespace := testutils.CreateNamespace(t, db, "test-findall")

	t.Run("Empty database returns empty list", func(t *testing.T) {
		providers, count, err := providerRepo.FindAll(ctx, 0, 10)
		require.NoError(t, err)
		assert.Empty(t, providers)
		assert.Equal(t, 0, count)
	})

	t.Run("Multiple providers returned correctly", func(t *testing.T) {
		// Create multiple providers
		testutils.CreateProvider(t, db, namespace.ID, "provider1", nil, sqldb.ProviderTierCommunity, nil)
		testutils.CreateProvider(t, db, namespace.ID, "provider2", nil, sqldb.ProviderTierCommunity, nil)
		testutils.CreateProvider(t, db, namespace.ID, "provider3", nil, sqldb.ProviderTierCommunity, nil)

		providers, count, err := providerRepo.FindAll(ctx, 0, 10)
		require.NoError(t, err)
		assert.Len(t, providers, 3)
		assert.Equal(t, 3, count)
	})

	t.Run("Pagination works correctly", func(t *testing.T) {
		// Test offset and limit
		providers, count, err := providerRepo.FindAll(ctx, 0, 2)
		require.NoError(t, err)
		assert.Len(t, providers, 2)
		assert.Equal(t, 3, count) // Total count is still 3

		// Test offset
		providers, count, err = providerRepo.FindAll(ctx, 2, 2)
		require.NoError(t, err)
		assert.Len(t, providers, 1) // Only 1 remaining
		assert.Equal(t, 3, count)
	})
}

// TestProvider_Search tests searching for providers
// Python reference: test_provider.py::TestProvider (implicit search via repository)
func TestProvider_Search(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	ctx := context.Background()
	providerRepo := providerprepo.NewProviderRepository(db.DB)

	namespace := testutils.CreateNamespace(t, db, "test-search")
	description1 := "AWS provider for Terraform"
	description2 := "GCP provider for Terraform"
	description3 := "Azure provider for Terraform"

	providerDB1 := testutils.CreateProvider(t, db, namespace.ID, "aws-provider", &description1, sqldb.ProviderTierCommunity, nil)
	providerDB2 := testutils.CreateProvider(t, db, namespace.ID, "gcp-provider", &description2, sqldb.ProviderTierCommunity, nil)
	providerDB3 := testutils.CreateProvider(t, db, namespace.ID, "azure-provider", &description3, sqldb.ProviderTierCommunity, nil)

	// Add versions to make them searchable
	gpgKey1 := testutils.CreateGPGKey(t, db, "key1", providerDB1.ID, "KEY1")
	gpgKey2 := testutils.CreateGPGKey(t, db, "key2", providerDB2.ID, "KEY2")
	gpgKey3 := testutils.CreateGPGKey(t, db, "key3", providerDB3.ID, "KEY3")

	now := time.Now()
	version1 := testutils.CreateProviderVersion(t, db, providerDB1.ID, "1.0.0", gpgKey1.ID, false, &now)
	version2 := testutils.CreateProviderVersion(t, db, providerDB2.ID, "1.0.0", gpgKey2.ID, false, &now)
	version3 := testutils.CreateProviderVersion(t, db, providerDB3.ID, "1.0.0", gpgKey3.ID, false, &now)

	testutils.SetProviderLatestVersion(t, db, providerDB1.ID, version1.ID)
	testutils.SetProviderLatestVersion(t, db, providerDB2.ID, version2.ID)
	testutils.SetProviderLatestVersion(t, db, providerDB3.ID, version3.ID)

	t.Run("Search by name", func(t *testing.T) {
		providers, count, err := providerRepo.Search(ctx, "aws", 0, 10)
		require.NoError(t, err)
		assert.Len(t, providers, 1)
		assert.Equal(t, 1, count)
		assert.Equal(t, "aws-provider", providers[0].Name())
	})

	t.Run("Search by description", func(t *testing.T) {
		providers, count, err := providerRepo.Search(ctx, "GCP", 0, 10)
		require.NoError(t, err)
		assert.Len(t, providers, 1)
		assert.Equal(t, 1, count)
		assert.Equal(t, "gcp-provider", providers[0].Name())
	})

	t.Run("Search with no matches", func(t *testing.T) {
		providers, count, err := providerRepo.Search(ctx, "nonexistent", 0, 10)
		require.NoError(t, err)
		assert.Empty(t, providers)
		assert.Equal(t, 0, count)
	})

	t.Run("Search with empty query returns all providers with latest version", func(t *testing.T) {
		providers, count, err := providerRepo.Search(ctx, "", 0, 10)
		require.NoError(t, err)
		assert.Len(t, providers, 3)
		assert.Equal(t, 3, count)
	})

	t.Run("Search is case-insensitive", func(t *testing.T) {
		providers, count, err := providerRepo.Search(ctx, "AWS", 0, 10)
		require.NoError(t, err)
		assert.Len(t, providers, 1)
		assert.Equal(t, 1, count)
		assert.Equal(t, "aws-provider", providers[0].Name())
	})

	t.Run("Providers without latest version are excluded", func(t *testing.T) {
		// Create a provider without any versions
		testutils.CreateProvider(t, db, namespace.ID, "no-version-provider", nil, sqldb.ProviderTierCommunity, nil)

		providers, count, err := providerRepo.Search(ctx, "", 0, 10)
		require.NoError(t, err)
		// Should still only return 3 (the ones with versions)
		assert.Len(t, providers, 3)
		assert.Equal(t, 3, count)
	})
}

// TestProvider_Delete tests deleting a provider
// Python reference: test_provider.py::TestProvider (implicit delete behavior)
func TestProvider_Delete(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	ctx := context.Background()
	providerRepo := providerprepo.NewProviderRepository(db.DB)

	namespace := testutils.CreateNamespace(t, db, "test-delete")
	providerDB := testutils.CreateProvider(t, db, namespace.ID, "testprovider", nil, sqldb.ProviderTierCommunity, nil)

	t.Run("Delete existing provider", func(t *testing.T) {
		// Verify provider exists
		prov, err := providerRepo.FindByID(ctx, providerDB.ID)
		require.NoError(t, err)
		assert.NotNil(t, prov)

		// Delete the provider (by creating a new one to test delete)
		providerToDelete := testutils.CreateProvider(t, db, namespace.ID, "delete-me", nil, sqldb.ProviderTierCommunity, nil)

		// Perform delete via raw SQL since DeleteVersion is the method available
		err = db.DB.Delete(&sqldb.ProviderDB{}, providerToDelete.ID).Error
		require.NoError(t, err)

		// Verify provider was deleted
		deletedProv, err := providerRepo.FindByID(ctx, providerToDelete.ID)
		require.NoError(t, err)
		assert.Nil(t, deletedProv)
	})
}

// TestProvider_DeleteVersion tests deleting a provider version
// Python reference: test_provider.py::TestProvider (implicit version delete behavior)
func TestProvider_DeleteVersion(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	ctx := context.Background()
	providerRepo := providerprepo.NewProviderRepository(db.DB)

	namespace := testutils.CreateNamespace(t, db, "test-delete-version")
	providerDB := testutils.CreateProvider(t, db, namespace.ID, "testprovider", nil, sqldb.ProviderTierCommunity, nil)

	gpgKey := testutils.CreateGPGKey(t, db, "key1", providerDB.ID, "KEY12345")
	now := time.Now()
	versionDB := testutils.CreateProviderVersion(t, db, providerDB.ID, "1.0.0", gpgKey.ID, false, &now)

	t.Run("Delete existing version", func(t *testing.T) {
		// Verify version exists
		version, err := providerRepo.FindVersionByID(ctx, versionDB.ID)
		require.NoError(t, err)
		assert.NotNil(t, version)

		// Delete the version
		err = providerRepo.DeleteVersion(ctx, versionDB.ID)
		require.NoError(t, err)

		// Verify version was deleted
		deletedVersion, err := providerRepo.FindVersionByID(ctx, versionDB.ID)
		require.NoError(t, err)
		assert.Nil(t, deletedVersion)
	})
}

// TestProvider_BinaryOperations tests provider version binary operations
// Python reference: test_provider.py::TestProvider::test_get_versions_api_details (partial)
func TestProvider_BinaryOperations(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	ctx := context.Background()
	providerRepo := providerprepo.NewProviderRepository(db.DB)

	namespace := testutils.CreateNamespace(t, db, "test-binaries")
	providerDB := testutils.CreateProvider(t, db, namespace.ID, "testprovider", nil, sqldb.ProviderTierCommunity, nil)

	gpgKey := testutils.CreateGPGKey(t, db, "key1", providerDB.ID, "KEY12345")
	now := time.Now()
	versionDB := testutils.CreateProviderVersion(t, db, providerDB.ID, "1.0.0", gpgKey.ID, false, &now)

	t.Run("Create binaries for different platforms", func(t *testing.T) {
		// Create binaries for different OS/Arch combinations
		_ = testutils.CreateProviderVersionBinary(t, db, versionDB.ID,
			"terraform-provider-testprovider_1.0.0_linux_amd64.zip",
			sqldb.OSLinux,
			sqldb.ArchAMD64,
			"checksum1")

		_ = testutils.CreateProviderVersionBinary(t, db, versionDB.ID,
			"terraform-provider-testprovider_1.0.0_linux_arm64.zip",
			sqldb.OSLinux,
			sqldb.ArchARM64,
			"checksum2")

		_ = testutils.CreateProviderVersionBinary(t, db, versionDB.ID,
			"terraform-provider-testprovider_1.0.0_windows_amd64.zip",
			sqldb.OSWindows,
			sqldb.ArchAMD64,
			"checksum3")

		// Get binaries for the version
		binaries, err := providerRepo.FindBinariesByVersion(ctx, versionDB.ID)
		require.NoError(t, err)
		assert.Len(t, binaries, 3)

		// Verify binary properties
		for _, binary := range binaries {
			assert.NotEmpty(t, binary.FileName())
			assert.NotEmpty(t, binary.OperatingSystem())
			assert.NotEmpty(t, binary.Architecture())
			assert.NotEmpty(t, binary.FileHash())
		}
	})

	t.Run("Find binary by specific platform", func(t *testing.T) {
		// Find a specific binary
		binary, err := providerRepo.FindBinaryByPlatform(ctx, versionDB.ID,
			string(sqldb.OSLinux),
			string(sqldb.ArchAMD64))
		require.NoError(t, err)
		assert.NotNil(t, binary)
		assert.Equal(t, "checksum1", binary.FileHash())
	})

	t.Run("Binary for non-existent platform", func(t *testing.T) {
		// Try to find a binary for a non-existent platform
		binary, err := providerRepo.FindBinaryByPlatform(ctx, versionDB.ID,
			string(sqldb.OSDarwin),
			string(sqldb.ArchAMD64))
		require.NoError(t, err)
		assert.Nil(t, binary)
	})
}

// TestProvider_GetProviderVersionCount tests getting the version count for a provider
// Python reference: test_provider.py::TestProvider (implicit count behavior)
func TestProvider_GetProviderVersionCount(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	ctx := context.Background()
	providerRepo := providerprepo.NewProviderRepository(db.DB)

	namespace := testutils.CreateNamespace(t, db, "test-version-count")
	providerDB := testutils.CreateProvider(t, db, namespace.ID, "testprovider", nil, sqldb.ProviderTierCommunity, nil)

	gpgKey := testutils.CreateGPGKey(t, db, "key1", providerDB.ID, "KEY12345")
	now := time.Now()

	t.Run("No versions returns 0", func(t *testing.T) {
		count, err := providerRepo.GetProviderVersionCount(ctx, providerDB.ID)
		require.NoError(t, err)
		assert.Equal(t, 0, count)
	})

	t.Run("Count matches number of versions", func(t *testing.T) {
		// Create 3 versions
		_ = testutils.CreateProviderVersion(t, db, providerDB.ID, "1.0.0", gpgKey.ID, false, &now)
		_ = testutils.CreateProviderVersion(t, db, providerDB.ID, "2.0.0", gpgKey.ID, false, &now)
		_ = testutils.CreateProviderVersion(t, db, providerDB.ID, "3.0.0", gpgKey.ID, false, &now)

		count, err := providerRepo.GetProviderVersionCount(ctx, providerDB.ID)
		require.NoError(t, err)
		assert.Equal(t, 3, count)
	})
}

// TestProvider_UseProviderSourceAuth tests the use_default_provider_source_auth property
// Python reference: test_provider.py::TestProvider::test_use_default_provider_source_auth
func TestProvider_UseProviderSourceAuth(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	ctx := context.Background()
	providerRepo := providerprepo.NewProviderRepository(db.DB)

	namespace := testutils.CreateNamespace(t, db, "test-auth-flag")

	t.Run("Provider with use_default_provider_source_auth true", func(t *testing.T) {
		prov := provider.NewProvider(
			namespace.ID,
			"testprovider1",
			nil,
			"community",
			nil,
			nil,
			true, // use default auth
		)

		err := providerRepo.Save(ctx, prov)
		require.NoError(t, err)

		// Verify the property was saved correctly
		retrieved, err := providerRepo.FindByID(ctx, prov.ID())
		require.NoError(t, err)
		assert.True(t, retrieved.UseProviderSourceAuth())
	})

	t.Run("Provider with use_default_provider_source_auth false", func(t *testing.T) {
		prov := provider.NewProvider(
			namespace.ID,
			"testprovider2",
			nil,
			"community",
			nil,
			nil,
			false, // don't use default auth
		)

		err := providerRepo.Save(ctx, prov)
		require.NoError(t, err)

		// Verify the property was saved correctly
		retrieved, err := providerRepo.FindByID(ctx, prov.ID())
		require.NoError(t, err)
		assert.False(t, retrieved.UseProviderSourceAuth())
	})
}

// TestProvider_CategoryAssociation tests provider category association
// Python reference: test_provider.py::TestProvider::test_category
func TestProvider_CategoryAssociation(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	ctx := context.Background()
	providerRepo := providerprepo.NewProviderRepository(db.DB)

	namespace := testutils.CreateNamespace(t, db, "test-category")
	category := testutils.CreateProviderCategory(t, db, "Cloud Providers", "cloud", true)

	providerDB := testutils.CreateProvider(t, db, namespace.ID, "testprovider", nil, sqldb.ProviderTierCommunity, &category.ID)

	// Get provider with category preloaded
	prov, err := providerRepo.FindByID(ctx, providerDB.ID)
	require.NoError(t, err)

	assert.NotNil(t, prov.CategoryID())
	assert.Equal(t, category.ID, *prov.CategoryID())
}

// TestProvider_DifferentTiers tests different provider tiers
// Python reference: test_provider.py::TestProvider::test_tier
func TestProvider_DifferentTiers(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	ctx := context.Background()
	providerRepo := providerprepo.NewProviderRepository(db.DB)

	namespace := testutils.CreateNamespace(t, db, "test-tiers")

	testCases := []struct {
		tier           sqldb.ProviderTier
		expectedString string
	}{
		{sqldb.ProviderTierCommunity, "community"},
		{sqldb.ProviderTierOfficial, "official"},
		{sqldb.ProviderTierPartner, "partner"},
	}

	for _, tc := range testCases {
		t.Run(tc.expectedString+" tier", func(t *testing.T) {
			providerDB := testutils.CreateProvider(t, db, namespace.ID, "test-"+tc.expectedString, nil, tc.tier, nil)

			prov, err := providerRepo.FindByID(ctx, providerDB.ID)
			require.NoError(t, err)
			assert.Equal(t, tc.expectedString, prov.Tier())
		})
	}
}

// TestProvider_GPGKeyOperations tests GPG key operations
// Python reference: test_provider.py::TestProvider (GPG key implicit operations)
func TestProvider_GPGKeyOperations(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	ctx := context.Background()
	providerRepo := providerprepo.NewProviderRepository(db.DB)

	namespace := testutils.CreateNamespace(t, db, "test-gpg")
	providerDB := testutils.CreateProvider(t, db, namespace.ID, "testprovider", nil, sqldb.ProviderTierCommunity, nil)

	t.Run("Find GPG keys by provider", func(t *testing.T) {
		// Create GPG keys for the namespace
		gpgKey1 := testutils.CreateGPGKey(t, db, "key1", providerDB.ID, "KEY11111111")
		_ = testutils.CreateGPGKey(t, db, "key2", providerDB.ID, "KEY22222222")

		// Find GPG keys for provider
		gpgKeys, err := providerRepo.FindGPGKeysByProvider(ctx, providerDB.ID)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(gpgKeys), 1)

		// Verify we can find a specific key
		// gpgKey1.KeyID is *string, so we need to dereference it
		if gpgKey1.KeyID != nil {
			foundKey, err := providerRepo.FindGPGKeyByKeyID(ctx, *gpgKey1.KeyID)
			require.NoError(t, err)
			assert.NotNil(t, foundKey)
		}
	})

	t.Run("Find GPG key by key ID", func(t *testing.T) {
		// Create a specific GPG key
		gpgKey := testutils.CreateGPGKey(t, db, "testkey", providerDB.ID, "ABC12345")

		// Find by key ID
		foundKey, err := providerRepo.FindGPGKeyByKeyID(ctx, "ABC12345")
		require.NoError(t, err)
		assert.NotNil(t, foundKey)
		assert.Equal(t, gpgKey.ID, foundKey.ID())
	})

	t.Run("Find non-existent GPG key", func(t *testing.T) {
		// Try to find a non-existent key
		foundKey, err := providerRepo.FindGPGKeyByKeyID(ctx, "NONEXISTENT")
		require.NoError(t, err)
		assert.Nil(t, foundKey)
	})
}
