package provider

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/persistence/sqldb"
	providerrepo "github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/persistence/sqldb/provider"
	testutils "github.com/matthewjohn/terrareg/terrareg-go/test/integration/testutils"
)

// TestProviderVersionBinary_CreateAndRetrieve tests creating and retrieving provider binaries
// for various OS and architecture combinations
// Python reference: test_provider_version_binary.py::TestProviderVersionBinary::test_create
func TestProviderVersionBinary_CreateAndRetrieve(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	ctx := context.Background()
	providerRepo := providerrepo.NewProviderRepository(db.DB)

	// Create test data
	namespace := testutils.CreateNamespace(t, db, "test-binary-create")
	category := testutils.CreateProviderCategory(t, db, "Test Category", "test-category", true)
	gpgKey := testutils.CreateGPGKeyWithNamespace(t, db, "test-key", namespace.ID, "test-key-id")

	provider := testutils.CreateProvider(t, db, namespace.ID, "test-provider", nil, sqldb.ProviderTierCommunity, &category.ID)
	versionDB := testutils.CreateProviderVersion(t, db, provider.ID, "6.4.1", gpgKey.ID, false, nil)

	testCases := []struct {
		name             string
		os               sqldb.ProviderBinaryOperatingSystemType
		arch             sqldb.ProviderBinaryArchitectureType
		expectedFileName string
	}{
		{"terraform-provider-test-provider_6.4.1_linux_amd64.zip", sqldb.OSLinux, sqldb.ArchAMD64, "terraform-provider-test-provider_6.4.1_linux_amd64.zip"},
		{"terraform-provider-test-provider_6.4.1_linux_arm.zip", sqldb.OSLinux, sqldb.ArchARM, "terraform-provider-test-provider_6.4.1_linux_arm.zip"},
		{"terraform-provider-test-provider_6.4.1_linux_arm64.zip", sqldb.OSLinux, sqldb.ArchARM64, "terraform-provider-test-provider_6.4.1_linux_arm64.zip"},
		{"terraform-provider-test-provider_6.4.1_linux_386.zip", sqldb.OSLinux, sqldb.Arch386, "terraform-provider-test-provider_6.4.1_linux_386.zip"},

		{"terraform-provider-test-provider_6.4.1_windows_amd64.zip", sqldb.OSWindows, sqldb.ArchAMD64, "terraform-provider-test-provider_6.4.1_windows_amd64.zip"},
		{"terraform-provider-test-provider_6.4.1_windows_arm.zip", sqldb.OSWindows, sqldb.ArchARM, "terraform-provider-test-provider_6.4.1_windows_arm.zip"},
		{"terraform-provider-test-provider_6.4.1_windows_arm64.zip", sqldb.OSWindows, sqldb.ArchARM64, "terraform-provider-test-provider_6.4.1_windows_arm64.zip"},
		{"terraform-provider-test-provider_6.4.1_windows_386.zip", sqldb.OSWindows, sqldb.Arch386, "terraform-provider-test-provider_6.4.1_windows_386.zip"},

		{"terraform-provider-test-provider_6.4.1_darwin_amd64.zip", sqldb.OSDarwin, sqldb.ArchAMD64, "terraform-provider-test-provider_6.4.1_darwin_amd64.zip"},
		{"terraform-provider-test-provider_6.4.1_darwin_arm.zip", sqldb.OSDarwin, sqldb.ArchARM, "terraform-provider-test-provider_6.4.1_darwin_arm.zip"},
		{"terraform-provider-test-provider_6.4.1_darwin_arm64.zip", sqldb.OSDarwin, sqldb.ArchARM64, "terraform-provider-test-provider_6.4.1_darwin_arm64.zip"},
		{"terraform-provider-test-provider_6.4.1_darwin_386.zip", sqldb.OSDarwin, sqldb.Arch386, "terraform-provider-test-provider_6.4.1_darwin_386.zip"},

		{"terraform-provider-test-provider_6.4.1_freebsd_amd64.zip", sqldb.OSFreeBSD, sqldb.ArchAMD64, "terraform-provider-test-provider_6.4.1_freebsd_amd64.zip"},
		{"terraform-provider-test-provider_6.4.1_freebsd_arm.zip", sqldb.OSFreeBSD, sqldb.ArchARM, "terraform-provider-test-provider_6.4.1_freebsd_arm.zip"},
		{"terraform-provider-test-provider_6.4.1_freebsd_arm64.zip", sqldb.OSFreeBSD, sqldb.ArchARM64, "terraform-provider-test-provider_6.4.1_freebsd_arm64.zip"},
		{"terraform-provider-test-provider_6.4.1_freebsd_386.zip", sqldb.OSFreeBSD, sqldb.Arch386, "terraform-provider-test-provider_6.4.1_freebsd_386.zip"},
	}

	for _, tc := range testCases {
		t.Run(string(tc.os)+"_"+string(tc.arch), func(t *testing.T) {
			checksum := "c27f1263ae06f263d59eb1f172c7fe39f6d7a06771544d869cc272d94ed301d1"

			// Create binary
			binaryDB := testutils.CreateProviderVersionBinary(t, db, versionDB.ID, tc.name, tc.os, tc.arch, checksum)

			// Verify binary was created correctly
			assert.Equal(t, versionDB.ID, binaryDB.ProviderVersionID)
			assert.Equal(t, tc.name, binaryDB.Name)
			assert.Equal(t, checksum, binaryDB.Checksum)
			assert.Equal(t, tc.os, binaryDB.OperatingSystem)
			assert.Equal(t, tc.arch, binaryDB.Architecture)

			// Retrieve and verify using domain model
			binary, err := providerRepo.FindBinaryByPlatform(ctx, versionDB.ID, string(tc.os), string(tc.arch))
			require.NoError(t, err)
			require.NotNil(t, binary)
			assert.Equal(t, tc.expectedFileName, binary.FileName())
			assert.Equal(t, checksum, binary.FileHash())
			assert.Equal(t, string(tc.os), binary.OperatingSystem())
			assert.Equal(t, string(tc.arch), binary.Architecture())
		})
	}
}

// TestProviderVersionBinary_FindByPlatform tests finding binaries by platform
// Python reference: test_provider_version_binary.py::TestProviderVersionBinary::test_get
func TestProviderVersionBinary_FindByPlatform(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	ctx := context.Background()
	providerRepo := providerrepo.NewProviderRepository(db.DB)

	// Create test data
	namespace := testutils.CreateNamespace(t, db, "initial-providers")
	category := testutils.CreateProviderCategory(t, db, "Test Category", "test-category", true)
	gpgKey := testutils.CreateGPGKeyWithNamespace(t, db, "test-key", namespace.ID, "test-key-id")

	provider := testutils.CreateProvider(t, db, namespace.ID, "multiple-versions", nil, sqldb.ProviderTierCommunity, &category.ID)
	versionDB := testutils.CreateProviderVersion(t, db, provider.ID, "1.5.0", gpgKey.ID, false, nil)

	// Create test binaries
	checksum1 := "a26d0401981bf2749c129ab23b3037e82bd200582ff7489e0da2a967b50daa98"
	checksum2 := "bda5d57cf68ab142f5d0c9a5a0739577e24444d4e8fe4a096ab9f4935bec9e9a"
	checksum3 := "c2d859efacc3fbe1662bef92c80ce32c966834597625914592891eaf564af4bf"
	checksum4 := "e8bc51e741c45feed8d9d7eb1133ac0107152cab3c1db12e74495d4b4ec75a0c"

	testutils.CreateProviderVersionBinary(t, db, versionDB.ID, "terraform-provider-multiple-versions_1.5.0_linux_amd64.zip", sqldb.OSLinux, sqldb.ArchAMD64, checksum1)
	testutils.CreateProviderVersionBinary(t, db, versionDB.ID, "terraform-provider-multiple-versions_1.5.0_linux_arm64.zip", sqldb.OSLinux, sqldb.ArchARM64, checksum2)
	testutils.CreateProviderVersionBinary(t, db, versionDB.ID, "terraform-provider-multiple-versions_1.5.0_windows_amd64.zip", sqldb.OSWindows, sqldb.ArchAMD64, checksum3)
	testutils.CreateProviderVersionBinary(t, db, versionDB.ID, "terraform-provider-multiple-versions_1.5.0_darwin_amd64.zip", sqldb.OSDarwin, sqldb.ArchAMD64, checksum4)

	t.Run("Find existing binary", func(t *testing.T) {
		binary, err := providerRepo.FindBinaryByPlatform(ctx, versionDB.ID, string(sqldb.OSWindows), string(sqldb.ArchAMD64))
		require.NoError(t, err)
		require.NotNil(t, binary)
		assert.Equal(t, checksum3, binary.FileHash())
		assert.Equal(t, string(sqldb.OSWindows), binary.OperatingSystem())
		assert.Equal(t, string(sqldb.ArchAMD64), binary.Architecture())
	})

	t.Run("Non-existent binary returns nil", func(t *testing.T) {
		binary, err := providerRepo.FindBinaryByPlatform(ctx, versionDB.ID, string(sqldb.OSFreeBSD), string(sqldb.ArchARM64))
		require.NoError(t, err)
		assert.Nil(t, binary)
	})
}

// TestProviderVersionBinary_FindByVersion tests finding all binaries for a version
// Python reference: test_provider_version_binary.py::TestProviderVersionBinary::test_get_by_provider_version
func TestProviderVersionBinary_FindByVersion(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	ctx := context.Background()
	providerRepo := providerrepo.NewProviderRepository(db.DB)

	// Create test data
	namespace := testutils.CreateNamespace(t, db, "initial-providers")
	category := testutils.CreateProviderCategory(t, db, "Test Category", "test-category", true)
	gpgKey := testutils.CreateGPGKeyWithNamespace(t, db, "test-key", namespace.ID, "test-key-id")

	provider := testutils.CreateProvider(t, db, namespace.ID, "multiple-versions", nil, sqldb.ProviderTierCommunity, &category.ID)
	versionDB := testutils.CreateProviderVersion(t, db, provider.ID, "1.5.0", gpgKey.ID, false, nil)

	// Create test binaries with specific checksums
	checksum1 := "a26d0401981bf2749c129ab23b3037e82bd200582ff7489e0da2a967b50daa98"
	checksum2 := "bda5d57cf68ab142f5d0c9a5a0739577e24444d4e8fe4a096ab9f4935bec9e9a"
	checksum3 := "c2d859efacc3fbe1662bef92c80ce32c966834597625914592891eaf564af4bf"
	checksum4 := "e8bc51e741c45feed8d9d7eb1133ac0107152cab3c1db12e74495d4b4ec75a0c"

	testutils.CreateProviderVersionBinary(t, db, versionDB.ID, "terraform-provider-multiple-versions_1.5.0_linux_amd64.zip", sqldb.OSLinux, sqldb.ArchAMD64, checksum1)
	testutils.CreateProviderVersionBinary(t, db, versionDB.ID, "terraform-provider-multiple-versions_1.5.0_linux_arm64.zip", sqldb.OSLinux, sqldb.ArchARM64, checksum2)
	testutils.CreateProviderVersionBinary(t, db, versionDB.ID, "terraform-provider-multiple-versions_1.5.0_windows_amd64.zip", sqldb.OSWindows, sqldb.ArchAMD64, checksum3)
	testutils.CreateProviderVersionBinary(t, db, versionDB.ID, "terraform-provider-multiple-versions_1.5.0_darwin_amd64.zip", sqldb.OSDarwin, sqldb.ArchAMD64, checksum4)

	// Get all binaries for the version
	binaries, err := providerRepo.FindBinariesByVersion(ctx, versionDB.ID)
	require.NoError(t, err)
	assert.Len(t, binaries, 4)

	// Verify binaries have correct OS/Arch combinations
	osArch := make(map[string]bool)
	for _, binary := range binaries {
		key := binary.OperatingSystem() + "/" + binary.Architecture()
		osArch[key] = true
	}

	assert.True(t, osArch["linux/amd64"])
	assert.True(t, osArch["linux/arm64"])
	assert.True(t, osArch["windows/amd64"])
	assert.True(t, osArch["darwin/amd64"])
}

// TestProviderVersionBinary_Properties tests various binary properties
// Python reference: test_provider_version_binary.py::TestProviderVersionBinary::test_name, test_architecture, etc.
func TestProviderVersionBinary_Properties(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	ctx := context.Background()
	providerRepo := providerrepo.NewProviderRepository(db.DB)

	// Create test data
	namespace := testutils.CreateNamespace(t, db, "test-properties")
	category := testutils.CreateProviderCategory(t, db, "Test Category", "test-category", true)
	gpgKey := testutils.CreateGPGKeyWithNamespace(t, db, "test-key", namespace.ID, "test-key-id")

	provider := testutils.CreateProvider(t, db, namespace.ID, "test-provider", nil, sqldb.ProviderTierCommunity, &category.ID)
	versionDB := testutils.CreateProviderVersion(t, db, provider.ID, "1.5.0", gpgKey.ID, false, nil)

	t.Run("FileName property", func(t *testing.T) {
		expectedName := "terraform-provider-test-provider_1.5.0_windows_amd64.zip"
		checksum := "c2d859efacc3fbe1662bef92c80ce32c966834597625914592891eaf564af4bf"
		_ = testutils.CreateProviderVersionBinary(t, db, versionDB.ID, expectedName, sqldb.OSWindows, sqldb.ArchAMD64, checksum)

		binary, err := providerRepo.FindBinaryByPlatform(ctx, versionDB.ID, string(sqldb.OSWindows), string(sqldb.ArchAMD64))
		require.NoError(t, err)
		require.NotNil(t, binary)
		assert.Equal(t, expectedName, binary.FileName())

		// Test Filename alias
		assert.Equal(t, expectedName, binary.Filename())
	})

	t.Run("OperatingSystem property", func(t *testing.T) {
		checksum := "c2d859efacc3fbe1662bef92c80ce32c966834597625914592891eaf564af4bf"
		_ = testutils.CreateProviderVersionBinary(t, db, versionDB.ID, "terraform-provider-test-provider_1.5.0_freebsd_amd64.zip", sqldb.OSFreeBSD, sqldb.ArchAMD64, checksum)

		binary, err := providerRepo.FindBinaryByPlatform(ctx, versionDB.ID, string(sqldb.OSFreeBSD), string(sqldb.ArchAMD64))
		require.NoError(t, err)
		require.NotNil(t, binary)
		assert.Equal(t, string(sqldb.OSFreeBSD), binary.OperatingSystem())

		// Test OS alias
		assert.Equal(t, string(sqldb.OSFreeBSD), binary.OS())
	})

	t.Run("Architecture property", func(t *testing.T) {
		checksum := "c2d859efacc3fbe1662bef92c80ce32c966834597625914592891eaf564af4bf"
		_ = testutils.CreateProviderVersionBinary(t, db, versionDB.ID, "terraform-provider-test-provider_1.5.0_linux_arm64.zip", sqldb.OSLinux, sqldb.ArchARM64, checksum)

		binary, err := providerRepo.FindBinaryByPlatform(ctx, versionDB.ID, string(sqldb.OSLinux), string(sqldb.ArchARM64))
		require.NoError(t, err)
		require.NotNil(t, binary)
		assert.Equal(t, string(sqldb.ArchARM64), binary.Architecture())

		// Test Arch alias
		assert.Equal(t, string(sqldb.ArchARM64), binary.Arch())
	})

	t.Run("FileHash property", func(t *testing.T) {
		checksum := "c2d859efacc3fbe1662bef92c80ce32c966834597625914592891eaf564af4bf"
		_ = testutils.CreateProviderVersionBinary(t, db, versionDB.ID, "terraform-provider-test-provider_1.5.0_linux_amd64.zip", sqldb.OSLinux, sqldb.ArchAMD64, checksum)

		binary, err := providerRepo.FindBinaryByPlatform(ctx, versionDB.ID, string(sqldb.OSLinux), string(sqldb.ArchAMD64))
		require.NoError(t, err)
		require.NotNil(t, binary)
		assert.Equal(t, checksum, binary.FileHash())
	})

	t.Run("VersionID property", func(t *testing.T) {
		checksum := "c2d859efacc3fbe1662bef92c80ce32c966834597625914592891eaf564af4bf"
		_ = testutils.CreateProviderVersionBinary(t, db, versionDB.ID, "terraform-provider-test-provider_1.5.0_darwin_amd64.zip", sqldb.OSDarwin, sqldb.ArchAMD64, checksum)

		binary, err := providerRepo.FindBinaryByPlatform(ctx, versionDB.ID, string(sqldb.OSDarwin), string(sqldb.ArchAMD64))
		require.NoError(t, err)
		require.NotNil(t, binary)
		assert.Equal(t, versionDB.ID, binary.VersionID())
	})
}

// TestProviderVersionBinary_EmptyResults tests behavior with no binaries
// Python reference: test_provider_version_binary.py (implicit empty result handling)
func TestProviderVersionBinary_EmptyResults(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	ctx := context.Background()
	providerRepo := providerrepo.NewProviderRepository(db.DB)

	// Create test data
	namespace := testutils.CreateNamespace(t, db, "test-empty")
	category := testutils.CreateProviderCategory(t, db, "Test Category", "test-category", true)
	gpgKey := testutils.CreateGPGKeyWithNamespace(t, db, "test-key", namespace.ID, "test-key-id")

	provider := testutils.CreateProvider(t, db, namespace.ID, "test-provider", nil, sqldb.ProviderTierCommunity, &category.ID)
	versionDB := testutils.CreateProviderVersion(t, db, provider.ID, "1.0.0", gpgKey.ID, false, nil)

	t.Run("No binaries returns empty list", func(t *testing.T) {
		binaries, err := providerRepo.FindBinariesByVersion(ctx, versionDB.ID)
		require.NoError(t, err)
		assert.Empty(t, binaries)
	})

	t.Run("Non-existent binary returns nil", func(t *testing.T) {
		binary, err := providerRepo.FindBinaryByPlatform(ctx, versionDB.ID, string(sqldb.OSLinux), string(sqldb.ArchAMD64))
		require.NoError(t, err)
		assert.Nil(t, binary)
	})
}

// TestProviderVersionBinary_MultipleVersions tests binaries across multiple versions
// Python reference: test_provider_version_binary.py (multi-version behavior)
func TestProviderVersionBinary_MultipleVersions(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	ctx := context.Background()
	providerRepo := providerrepo.NewProviderRepository(db.DB)

	// Create test data
	namespace := testutils.CreateNamespace(t, db, "test-multi")
	category := testutils.CreateProviderCategory(t, db, "Test Category", "test-category", true)
	gpgKey := testutils.CreateGPGKeyWithNamespace(t, db, "test-key", namespace.ID, "test-key-id")

	provider := testutils.CreateProvider(t, db, namespace.ID, "test-provider", nil, sqldb.ProviderTierCommunity, &category.ID)

	// Create two versions
	version1DB := testutils.CreateProviderVersion(t, db, provider.ID, "1.0.0", gpgKey.ID, false, nil)
	version2DB := testutils.CreateProviderVersion(t, db, provider.ID, "2.0.0", gpgKey.ID, false, nil)

	// Create binaries for each version
	checksum1 := "checksum100000000000000000000000000000000000000000000000001"
	checksum2 := "checksum200000000000000000000000000000000000000000000000002"

	testutils.CreateProviderVersionBinary(t, db, version1DB.ID, "terraform-provider-test-provider_1.0.0_linux_amd64.zip", sqldb.OSLinux, sqldb.ArchAMD64, checksum1)
	testutils.CreateProviderVersionBinary(t, db, version2DB.ID, "terraform-provider-test-provider_2.0.0_linux_amd64.zip", sqldb.OSLinux, sqldb.ArchAMD64, checksum2)

	t.Run("Each version has correct binary", func(t *testing.T) {
		// Version 1 binary
		binary1, err := providerRepo.FindBinaryByPlatform(ctx, version1DB.ID, string(sqldb.OSLinux), string(sqldb.ArchAMD64))
		require.NoError(t, err)
		require.NotNil(t, binary1)
		assert.Equal(t, checksum1, binary1.FileHash())

		// Version 2 binary
		binary2, err := providerRepo.FindBinaryByPlatform(ctx, version2DB.ID, string(sqldb.OSLinux), string(sqldb.ArchAMD64))
		require.NoError(t, err)
		require.NotNil(t, binary2)
		assert.Equal(t, checksum2, binary2.FileHash())

		// Verify they're different
		assert.NotEqual(t, binary1.ID(), binary2.ID())
	})

	t.Run("Each version returns only its binaries", func(t *testing.T) {
		// Add another binary to version 2
		checksum3 := "checksum300000000000000000000000000000000000000000000000003"
		testutils.CreateProviderVersionBinary(t, db, version2DB.ID, "terraform-provider-test-provider_2.0.0_windows_amd64.zip", sqldb.OSWindows, sqldb.ArchAMD64, checksum3)

		// Version 1 should have 1 binary
		binaries1, err := providerRepo.FindBinariesByVersion(ctx, version1DB.ID)
		require.NoError(t, err)
		assert.Len(t, binaries1, 1)

		// Version 2 should have 2 binaries
		binaries2, err := providerRepo.FindBinariesByVersion(ctx, version2DB.ID)
		require.NoError(t, err)
		assert.Len(t, binaries2, 2)
	})
}
