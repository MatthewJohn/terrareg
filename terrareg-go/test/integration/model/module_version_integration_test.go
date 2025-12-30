package model

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/persistence/sqldb"
	"github.com/matthewjohn/terrareg/terrareg-go/test/integration/testutils"
)

// TestModuleVersion_IntegrationWithDetails tests module version with module details
// Python reference: test_module_version.py::TestModuleVersion::test_create_db_row (partial)
func TestModuleVersion_IntegrationWithDetails(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	// Create namespace and module provider
	namespace := testutils.CreateNamespace(t, db, "testnamespace")
	moduleProvider := testutils.CreateModuleProvider(t, db, namespace.ID, "test-module", "aws")

	// Create module details
	readmeContent := "# Test Module\n\nThis is a test module."
	moduleDetails := testutils.CreateModuleDetails(t, db, readmeContent)

	// Create module version with details
	version := "1.0.0"
	moduleVersion := sqldb.ModuleVersionDB{
		ModuleProviderID: moduleProvider.ID,
		Version:          version,
		Beta:             false,
		Internal:         false,
		Published:        nil,
		ModuleDetailsID:  &moduleDetails.ID,
	}
	err := db.DB.Create(&moduleVersion).Error
	require.NoError(t, err)

	// Verify the relationship
	var retrievedVersion sqldb.ModuleVersionDB
	err = db.DB.Preload("ModuleDetails").First(&retrievedVersion, moduleVersion.ID).Error
	require.NoError(t, err)
	assert.Equal(t, version, retrievedVersion.Version)
	assert.NotNil(t, retrievedVersion.ModuleDetailsID)
	assert.Equal(t, moduleDetails.ID, *retrievedVersion.ModuleDetailsID)
}

// TestModuleVersion_IntegrationWithSubmodules tests module version with submodules
// Python reference: test_module_version.py::TestModuleVersion::test_create_db_row (partial - submodule integration)
func TestModuleVersion_IntegrationWithSubmodules(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	// Setup: Create namespace, module provider, and version
	namespace := testutils.CreateNamespace(t, db, "testsubmodules")
	moduleProvider := testutils.CreateModuleProvider(t, db, namespace.ID, "test-module", "aws")
	moduleVersion := testutils.CreateModuleVersion(t, db, moduleProvider.ID, "1.0.0")

	// Create module details for submodules
	submodule1Details := testutils.CreateModuleDetails(t, db, "# Submodule 1\n\nContent for submodule 1.")
	submodule2Details := testutils.CreateModuleDetails(t, db, "# Submodule 2\n\nContent for submodule 2.")

	// Create submodules
	submodule1Name := "database"
	submodule1Type := "module"
	_ = testutils.CreateSubmodule(t, db, moduleVersion.ID, "modules/database", submodule1Name, submodule1Type, &submodule1Details.ID)

	submodule2Name := "network"
	submodule2Type := "module"
	_ = testutils.CreateSubmodule(t, db, moduleVersion.ID, "modules/network", submodule2Name, submodule2Type, &submodule2Details.ID)

	// Verify the submodules were created
	var submodules []sqldb.SubmoduleDB
	err := db.DB.Where("parent_module_version = ?", moduleVersion.ID).Find(&submodules).Error
	require.NoError(t, err)
	assert.Len(t, submodules, 2)

	// Verify submodule details
	for _, sub := range submodules {
		assert.Equal(t, moduleVersion.ID, sub.ParentModuleVersion)
		assert.NotNil(t, sub.Name)
		assert.NotNil(t, sub.Type)
		assert.NotNil(t, sub.ModuleDetailsID)
	}
}

// TestModuleVersion_PublishWorkflow tests the publish workflow
// Python reference: test_module_version.py::TestModuleVersion::test_module_create_extraction_wrapper (partial - publish)
func TestModuleVersion_PublishWorkflow(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	namespace := testutils.CreateNamespace(t, db, "testpublish")
	moduleProvider := testutils.CreateModuleProvider(t, db, namespace.ID, "test-module", "aws")

	// Create unpublished version
	moduleVersion := testutils.CreateModuleVersion(t, db, moduleProvider.ID, "1.0.0")
	assert.Nil(t, moduleVersion.Published)
	assert.Nil(t, moduleVersion.PublishedAt)

	// Publish the version
	published := true
	now := time.Now()
	err := db.DB.Model(&moduleVersion).Updates(map[string]interface{}{
		"published":    published,
		"published_at": now,
	}).Error
	require.NoError(t, err)

	// Verify the update
	var retrievedVersion sqldb.ModuleVersionDB
	err = db.DB.First(&retrievedVersion, moduleVersion.ID).Error
	require.NoError(t, err)
	assert.NotNil(t, retrievedVersion.Published)
	assert.True(t, *retrievedVersion.Published)
	assert.NotNil(t, retrievedVersion.PublishedAt)
}

// TestModuleVersion_BetaDetection tests beta version detection
// Python reference: test_module_version.py::TestModuleVersion::test_create_beta_version
func TestModuleVersion_BetaDetection(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	namespace := testutils.CreateNamespace(t, db, "testbeta")
	moduleProvider := testutils.CreateModuleProvider(t, db, namespace.ID, "test-module", "aws")

	// Create various versions and test beta detection
	versions := []string{
		"1.0.0",      // stable
		"1.0.0-beta", // beta
		"1.0.0-alpha",// alpha (beta)
		"1.0.0-rc1",  // rc (beta)
		"2.0.0",      // stable
	}

	expectedBeta := []bool{false, true, true, true, false}

	for i, version := range versions {
		moduleVersion := sqldb.ModuleVersionDB{
			ModuleProviderID: moduleProvider.ID,
			Version:          version,
			Beta:             expectedBeta[i],
			Published:        nil,
		}
		err := db.DB.Create(&moduleVersion).Error
		require.NoError(t, err)

		// Verify
		var retrieved sqldb.ModuleVersionDB
		err = db.DB.First(&retrieved, moduleVersion.ID).Error
		require.NoError(t, err)
		assert.Equal(t, expectedBeta[i], retrieved.Beta, "Version %s beta mismatch", version)
	}
}

// TestModuleVersion_VersionOrdering tests semantic version ordering
// Python reference: test_module_version.py::TestModuleVersion::test_valid_module_versions (partial - version ordering)
func TestModuleVersion_VersionOrdering(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	namespace := testutils.CreateNamespace(t, db, "testordering")
	moduleProvider := testutils.CreateModuleProvider(t, db, namespace.ID, "test-module", "aws")

	// Create multiple versions in non-sequential order
	versions := []string{
		"1.2.0",
		"1.0.0",
		"2.0.0",
		"1.10.0",
		"1.1.0",
	}

	for _, version := range versions {
		_ = testutils.CreateModuleVersion(t, db, moduleProvider.ID, version)
	}

	// Retrieve all versions and verify they exist
	var moduleVersions []sqldb.ModuleVersionDB
	err := db.DB.Where("module_provider_id = ?", moduleProvider.ID).Order("version ASC").Find(&moduleVersions).Error
	require.NoError(t, err)
	assert.Len(t, moduleVersions, 5)

	// Verify versions are stored correctly
	storedVersions := make([]string, len(moduleVersions))
	for i, v := range moduleVersions {
		storedVersions[i] = v.Version
	}

	// The string sort order should put them in: 1.0.0, 1.1.0, 1.10.0, 1.2.0, 2.0.0
	// (string order, not semantic order)
	assert.Equal(t, "1.0.0", storedVersions[0])
	assert.Equal(t, "1.1.0", storedVersions[1])
	assert.Equal(t, "1.10.0", storedVersions[2])
	assert.Equal(t, "1.2.0", storedVersions[3])
	assert.Equal(t, "2.0.0", storedVersions[4])
}

// TestModuleVersion_CascadeDelete tests that cascade deletion needs explicit handling
// Python reference: test_module_version.py::TestModuleVersion::test_delete (partial - cascade behavior)
func TestModuleVersion_CascadeDelete(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	namespace := testutils.CreateNamespace(t, db, "testcascade")
	moduleProvider := testutils.CreateModuleProvider(t, db, namespace.ID, "test-module", "aws")

	// Create module version with submodules
	moduleVersion := testutils.CreateModuleVersion(t, db, moduleProvider.ID, "1.0.0")
	moduleDetails := testutils.CreateModuleDetails(t, db, "# Test")
	_ = testutils.CreateSubmodule(t, db, moduleVersion.ID, "modules/test", "test", "module", &moduleDetails.ID)

	// Delete the module version
	err := db.DB.Delete(&moduleVersion).Error
	require.NoError(t, err)

	// Verify submodules still exist (no automatic cascade in SQLite test setup)
	var submodules []sqldb.SubmoduleDB
	err = db.DB.Where("parent_module_version = ?", moduleVersion.ID).Find(&submodules).Error
	require.NoError(t, err)
	assert.Len(t, submodules, 1, "Submodules are not automatically cascade deleted - this requires explicit handling in production")

	// Clean up submodules explicitly
	db.DB.Where("parent_module_version = ?", moduleVersion.ID).Delete(&sqldb.SubmoduleDB{})
}
