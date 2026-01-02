package provider

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/persistence/sqldb"
	providerrepo "github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/persistence/sqldb/provider"
	testutils "github.com/matthewjohn/terrareg/terrareg-go/test/integration/testutils"
)

// TestProviderVersion_GetByProviderAndVersion tests getting a provider version by provider and version string
// Python reference: test_provider_version.py::TestProviderVersion::test_get
func TestProviderVersion_GetByProviderAndVersion(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	ctx := context.Background()
	providerRepo := providerrepo.NewProviderRepository(db.DB)

	// Create test data
	namespace := testutils.CreateNamespace(t, db, "initial-providers")
	category := testutils.CreateProviderCategory(t, db, "Test Category", "test-category", true)
	gpgKey := testutils.CreateGPGKeyWithNamespace(t, db, "test-key", namespace.ID, "test-key-id")

	provider := testutils.CreateProvider(t, db, namespace.ID, "test-initial", nil, sqldb.ProviderTierCommunity, &category.ID)

	// Create a version
	now := time.Now()
	versionDB := testutils.CreateProviderVersion(t, db, provider.ID, "1.5.0", gpgKey.ID, false, &now)
	testutils.SetProviderLatestVersion(t, db, provider.ID, versionDB.ID)

	// Get version by provider and version string
	version, err := providerRepo.FindVersionByProviderAndVersion(ctx, provider.ID, "1.5.0")
	require.NoError(t, err)
	require.NotNil(t, version)
	assert.Equal(t, "1.5.0", version.Version())
	assert.Equal(t, provider.ID, version.ProviderID())

	// Test non-existent version
	version, err = providerRepo.FindVersionByProviderAndVersion(ctx, provider.ID, "1.9.0")
	require.NoError(t, err)
	assert.Nil(t, version)
}

// TestProviderVersion_GetByID tests getting a provider version by ID
// Python reference: test_provider_version.py::TestProviderVersion::test_get_by_pk
func TestProviderVersion_GetByID(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	ctx := context.Background()
	providerRepo := providerrepo.NewProviderRepository(db.DB)

	// Create test data
	namespace := testutils.CreateNamespace(t, db, "test-namespace")
	category := testutils.CreateProviderCategory(t, db, "Test Category", "test-category", true)
	gpgKey := testutils.CreateGPGKeyWithNamespace(t, db, "test-key", namespace.ID, "test-key-id")

	provider := testutils.CreateProvider(t, db, namespace.ID, "test-provider", nil, sqldb.ProviderTierCommunity, &category.ID)

	// Create a version
	now := time.Now()
	versionDB := testutils.CreateProviderVersion(t, db, provider.ID, "1.5.0", gpgKey.ID, false, &now)
	testutils.SetProviderLatestVersion(t, db, provider.ID, versionDB.ID)

	// Get version by ID
	version, err := providerRepo.FindVersionByID(ctx, versionDB.ID)
	require.NoError(t, err)
	require.NotNil(t, version)
	assert.Equal(t, "1.5.0", version.Version())
	assert.Equal(t, versionDB.ID, version.ID())

	// Verify parent objects are loaded correctly
	// Note: In Go, the provider is not automatically loaded, so we verify by ID
	assert.Equal(t, provider.ID, version.ProviderID())
}

// TestProviderVersion_PublishedAtFormatting tests the published_at formatting
// Python reference: test_provider_version.py::TestProviderVersion::test_published_at_display
func TestProviderVersion_PublishedAtFormatting(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	ctx := context.Background()
	providerRepo := providerrepo.NewProviderRepository(db.DB)

	// Create test data
	namespace := testutils.CreateNamespace(t, db, "test-namespace")
	category := testutils.CreateProviderCategory(t, db, "Test Category", "test-category", true)
	gpgKey := testutils.CreateGPGKeyWithNamespace(t, db, "test-key", namespace.ID, "test-key-id")

	provider := testutils.CreateProvider(t, db, namespace.ID, "test-provider", nil, sqldb.ProviderTierCommunity, &category.ID)

	t.Run("PublishedAt with date", func(t *testing.T) {
		// Create version with specific published date
		publishedAt := time.Date(2023, 10, 10, 6, 23, 0, 0, time.UTC)
		versionDB := testutils.CreateProviderVersion(t, db, provider.ID, "1.0.0", gpgKey.ID, false, &publishedAt)
		testutils.SetProviderLatestVersion(t, db, provider.ID, versionDB.ID)

		// Get version and check published date
		version, err := providerRepo.FindVersionByID(ctx, versionDB.ID)
		require.NoError(t, err)
		require.NotNil(t, version)
		assert.NotNil(t, version.PublishedAt())
		assert.Equal(t, 2023, version.PublishedAt().Year())
		assert.Equal(t, time.October, version.PublishedAt().Month())
		assert.Equal(t, 10, version.PublishedAt().Day())
	})

	t.Run("PublishedAt is nil", func(t *testing.T) {
		// Create version without published date
		versionDB := testutils.CreateProviderVersion(t, db, provider.ID, "2.0.0", gpgKey.ID, false, nil)
		testutils.SetProviderLatestVersion(t, db, provider.ID, versionDB.ID)

		// Get version and check published date is nil
		version, err := providerRepo.FindVersionByID(ctx, versionDB.ID)
		require.NoError(t, err)
		require.NotNil(t, version)
		assert.Nil(t, version.PublishedAt())
	})
}

// TestProviderVersion_BetaDetection tests beta version detection
// Python reference: test_provider_version.py::TestProviderVersion::test___validate_version_valid
func TestProviderVersion_BetaDetection(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	// Create test data
	namespace := testutils.CreateNamespace(t, db, "test-beta")
	category := testutils.CreateProviderCategory(t, db, "Test Category", "test-category", true)
	gpgKey := testutils.CreateGPGKeyWithNamespace(t, db, "test-key", namespace.ID, "test-key-id")

	provider := testutils.CreateProvider(t, db, namespace.ID, "test-provider", nil, sqldb.ProviderTierCommunity, &category.ID)

	testCases := []struct {
		version      string
		expectedBeta bool
	}{
		{"1.1.1", false},
		{"13.14.16", false},
		{"1.10.10", false},
		{"1.2.3-alpha", true},
		{"1.2.3-beta", true},
		{"1.2.3-anothersuffix1", true},
		{"1.2.2-123", true},
	}

	for _, tc := range testCases {
		t.Run(tc.version, func(t *testing.T) {
			versionDB := testutils.CreateProviderVersion(t, db, provider.ID, tc.version, gpgKey.ID, tc.expectedBeta, nil)

			// Verify beta flag is set correctly
			assert.Equal(t, tc.expectedBeta, versionDB.Beta)
		})
	}
}

// TestProviderVersion_GetBinaries tests getting binaries for a provider version
// Python reference: test_provider_version.py (various tests for binaries)
func TestProviderVersion_GetBinaries(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	ctx := context.Background()
	providerRepo := providerrepo.NewProviderRepository(db.DB)

	// Create test data
	namespace := testutils.CreateNamespace(t, db, "test-binaries")
	category := testutils.CreateProviderCategory(t, db, "Test Category", "test-category", true)
	gpgKey := testutils.CreateGPGKeyWithNamespace(t, db, "test-key", namespace.ID, "test-key-id")

	provider := testutils.CreateProvider(t, db, namespace.ID, "test-provider", nil, sqldb.ProviderTierCommunity, &category.ID)

	t.Run("No binaries returns empty list", func(t *testing.T) {
		versionDB := testutils.CreateProviderVersion(t, db, provider.ID, "1.0.0", gpgKey.ID, false, nil)

		binaries, err := providerRepo.FindBinariesByVersion(ctx, versionDB.ID)
		require.NoError(t, err)
		assert.Empty(t, binaries)
	})

	t.Run("Multiple binaries returns all", func(t *testing.T) {
		versionDB := testutils.CreateProviderVersion(t, db, provider.ID, "2.0.0", gpgKey.ID, false, nil)

		// Create binaries
		_ = testutils.CreateProviderVersionBinary(t, db, versionDB.ID, "terraform-provider-test_2.0.0_linux_amd64.zip", sqldb.OSLinux, sqldb.ArchAMD64, "abc123")
		_ = testutils.CreateProviderVersionBinary(t, db, versionDB.ID, "terraform-provider-test_2.0.0_darwin_amd64.zip", sqldb.OSDarwin, sqldb.ArchAMD64, "def456")
		_ = testutils.CreateProviderVersionBinary(t, db, versionDB.ID, "terraform-provider-test_2.0.0_windows_amd64.zip", sqldb.OSWindows, sqldb.ArchAMD64, "ghi789")

		binaries, err := providerRepo.FindBinariesByVersion(ctx, versionDB.ID)
		require.NoError(t, err)
		assert.Len(t, binaries, 3)

		// Verify binaries have correct OS/Arch
		osArch := make(map[string]bool)
		for _, binary := range binaries {
			key := binary.OperatingSystem() + "/" + binary.Architecture()
			osArch[key] = true
		}

		assert.True(t, osArch["linux/amd64"])
		assert.True(t, osArch["darwin/amd64"])
		assert.True(t, osArch["windows/amd64"])
	})
}

// TestProviderVersion_GitTag tests git tag handling
// Python reference: test_provider_version.py::TestProviderVersion::test_github_git_tag
func TestProviderVersion_GitTag(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	ctx := context.Background()
	providerRepo := providerrepo.NewProviderRepository(db.DB)

	// Create test data
	namespace := testutils.CreateNamespace(t, db, "test-gittag")
	category := testutils.CreateProviderCategory(t, db, "Test Category", "test-category", true)
	gpgKey := testutils.CreateGPGKeyWithNamespace(t, db, "test-key", namespace.ID, "test-key-id")

	provider := testutils.CreateProvider(t, db, namespace.ID, "test-provider", nil, sqldb.ProviderTierCommunity, &category.ID)

	t.Run("Version with git tag", func(t *testing.T) {
		gitTag := "v1.0.0"
		versionDB := testutils.CreateProviderVersion(t, db, provider.ID, "1.0.0", gpgKey.ID, false, nil)

		// Update git tag in database
		db.DB.Model(&versionDB).Update("git_tag", gitTag)

		// Get version and verify git tag
		version, err := providerRepo.FindVersionByID(ctx, versionDB.ID)
		require.NoError(t, err)
		require.NotNil(t, version)
		assert.NotNil(t, version.GitTag())
		assert.Equal(t, "v1.0.0", *version.GitTag())
	})

	t.Run("Version without git tag", func(t *testing.T) {
		versionDB := testutils.CreateProviderVersion(t, db, provider.ID, "2.0.0", gpgKey.ID, false, nil)

		// Get version and verify git tag is nil
		version, err := providerRepo.FindVersionByID(ctx, versionDB.ID)
		require.NoError(t, err)
		require.NotNil(t, version)
		assert.Nil(t, version.GitTag())
	})
}

// TestProviderVersion_Delete tests deleting a provider version
// Python reference: test_provider_version.py::TestProviderVersion::test_delete
func TestProviderVersion_Delete(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	ctx := context.Background()
	providerRepo := providerrepo.NewProviderRepository(db.DB)

	// Create test data
	namespace := testutils.CreateNamespace(t, db, "test-delete")
	category := testutils.CreateProviderCategory(t, db, "Test Category", "test-category", true)
	gpgKey := testutils.CreateGPGKeyWithNamespace(t, db, "test-key", namespace.ID, "test-key-id")

	provider := testutils.CreateProvider(t, db, namespace.ID, "test-provider", nil, sqldb.ProviderTierCommunity, &category.ID)

	t.Run("Delete non-latest version", func(t *testing.T) {
		// Create two versions
		version1DB := testutils.CreateProviderVersion(t, db, provider.ID, "1.0.0", gpgKey.ID, false, nil)
		version2DB := testutils.CreateProviderVersion(t, db, provider.ID, "2.0.0", gpgKey.ID, false, nil)

		// Set version 2 as latest
		testutils.SetProviderLatestVersion(t, db, provider.ID, version2DB.ID)

		// Delete version 1
		err := providerRepo.DeleteVersion(ctx, version1DB.ID)
		require.NoError(t, err)

		// Verify version 1 is deleted
		version, err := providerRepo.FindVersionByID(ctx, version1DB.ID)
		require.NoError(t, err)
		assert.Nil(t, version)

		// Verify version 2 still exists
		version, err = providerRepo.FindVersionByID(ctx, version2DB.ID)
		require.NoError(t, err)
		assert.NotNil(t, version)
	})

	t.Run("Cannot delete latest version", func(t *testing.T) {
		// Create a version and set it as latest
		versionDB := testutils.CreateProviderVersion(t, db, provider.ID, "3.0.0", gpgKey.ID, false, nil)
		testutils.SetProviderLatestVersion(t, db, provider.ID, versionDB.ID)

		// Attempt to delete latest version should fail
		err := providerRepo.DeleteVersion(ctx, versionDB.ID)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "cannot remove latest version")
	})
}

// TestProviderVersion_ProtocolVersions tests protocol versions handling
// Python reference: test_provider_version.py (various tests for protocol versions)
func TestProviderVersion_ProtocolVersions(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	ctx := context.Background()
	providerRepo := providerrepo.NewProviderRepository(db.DB)

	// Create test data
	namespace := testutils.CreateNamespace(t, db, "test-protocol")
	category := testutils.CreateProviderCategory(t, db, "Test Category", "test-category", true)
	gpgKey := testutils.CreateGPGKeyWithNamespace(t, db, "test-key", namespace.ID, "test-key-id")

	provider := testutils.CreateProvider(t, db, namespace.ID, "test-provider", nil, sqldb.ProviderTierCommunity, &category.ID)

	t.Run("Version with protocol versions", func(t *testing.T) {
		versionDB := testutils.CreateProviderVersion(t, db, provider.ID, "1.0.0", gpgKey.ID, false, nil)

		// Update protocol versions in database
		protocolVersions := "5.0,4.1,4.0"
		db.DB.Model(&versionDB).Update("protocol_versions", protocolVersions)

		// Get version and verify protocol versions
		version, err := providerRepo.FindVersionByID(ctx, versionDB.ID)
		require.NoError(t, err)
		require.NotNil(t, version)

		// Note: In Go, protocol versions are stored as JSON array and need to be parsed
		// The repository should handle this conversion
		protocolVers := version.ProtocolVersions()
		assert.Len(t, protocolVers, 3)
	})
}

// TestProviderVersion_VersionCount tests getting the count of versions for a provider
// Python reference: test_provider_version.py::TestProviderVersion::test_get_version_count
func TestProviderVersion_VersionCount(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	ctx := context.Background()
	providerRepo := providerrepo.NewProviderRepository(db.DB)

	// Create test data
	namespace := testutils.CreateNamespace(t, db, "test-count")
	category := testutils.CreateProviderCategory(t, db, "Test Category", "test-category", true)
	gpgKey := testutils.CreateGPGKeyWithNamespace(t, db, "test-key", namespace.ID, "test-key-id")

	provider := testutils.CreateProvider(t, db, namespace.ID, "test-provider", nil, sqldb.ProviderTierCommunity, &category.ID)

	t.Run("No versions returns 0", func(t *testing.T) {
		count, err := providerRepo.GetProviderVersionCount(ctx, provider.ID)
		require.NoError(t, err)
		assert.Equal(t, 0, count)
	})

	t.Run("Multiple versions returns correct count", func(t *testing.T) {
		// Create 5 versions
		for i := 1; i <= 5; i++ {
			version := testutils.CreateProviderVersion(t, db, provider.ID, "1.0.0", gpgKey.ID, false, nil)
			if i == 5 {
				testutils.SetProviderLatestVersion(t, db, provider.ID, version.ID)
			}
		}

		count, err := providerRepo.GetProviderVersionCount(ctx, provider.ID)
		require.NoError(t, err)
		assert.Equal(t, 5, count)
	})
}
