// Package terrareg_test provides integration tests for module version reindex mode behavior
// Python reference: test/integration/terrareg/models/test_module_version.py::TestModuleVersion::test_create_db_row_replace_existing
package terrareg_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	moduleModel "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/config/model"
	moduleService "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/module/service"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/persistence/sqldb"
	"github.com/matthewjohn/terrareg/terrareg-go/test/integration/testutils"
)

// TestModuleReindexMode_Legacy_PublishedVersionBecomesUnpublished tests that in legacy mode,
// re-indexing a published version creates a new unpublished version
// Python reference: test/integration/terrareg/models/test_module_version.py line 124-126
// (ModuleVersionReindexMode.LEGACY, previous_publish_state=True, config_auto_publish=False, expected_return_value=False)
func TestModuleReindexMode_Legacy_PublishedVersionBecomesUnpublished(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	// Create test server with LEGACY reindex mode
	domainConfig := testutils.CreateTestDomainConfigWithReindexMode(t, moduleModel.ModuleVersionReindexModeLegacy)
	cont := testutils.CreateTestServerWithDomainConfig(t, db, domainConfig)

	// Create test data - namespace, module provider, and a published version
	namespace := testutils.CreateNamespace(t, db, "testns", nil)
	moduleProvider := testutils.CreateModuleProvider(t, db, namespace.ID, "testmod", "aws")
	moduleVersion := testutils.CreateModuleVersion(t, db, moduleProvider.ID, "1.0.0")

	// Mark the version as published
	published := true
	moduleVersion.Published = &published
	err := db.DB.Save(&moduleVersion).Error
	require.NoError(t, err)

	// Update the module provider's latest_version_id
	err = db.DB.Model(&moduleProvider).Update("latest_version_id", moduleVersion.ID).Error
	require.NoError(t, err)

	// Now simulate a re-index (like uploading the same version again)
	// This should delete the old version and create a new one
	// In LEGACY mode, the new version should be unpublished
	req := moduleService.PrepareModuleRequest{
		Namespace:        "testns",
		ModuleName:       "testmod",
		Provider:         "aws",
		Version:          "1.0.0",
		ModuleProviderID: &moduleProvider.ID,
	}

	_, err = cont.ModuleCreationWrapper.PrepareModule(context.Background(), req)
	require.NoError(t, err)

	// Query the database to verify the new version is unpublished
	var newVersion sqldb.ModuleVersionDB
	err = db.DB.Where("module_provider_id = ? AND version = ?", moduleProvider.ID, "1.0.0").First(&newVersion).Error
	require.NoError(t, err)

	// The key assertion: in LEGACY mode, the new version should be unpublished
	// even though the old version was published
	assert.False(t, newVersion.Published != nil && *newVersion.Published, "In LEGACY mode, reindexed version should start unpublished")
}

// TestModuleReindexMode_AutoPublish_PublishedVersionStaysPublished tests that in auto-publish mode,
// re-indexing a published version preserves the published state
// Python reference: test/integration/terrareg/models/test_module_version.py line 132
// (ModuleVersionReindexMode.AUTO_PUBLISH, previous_publish_state=True, config_auto_publish=False, expected_return_value=True)
func TestModuleReindexMode_AutoPublish_PublishedVersionStaysPublished(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	// Create test server with AUTO_PUBLISH reindex mode
	domainConfig := testutils.CreateTestDomainConfigWithReindexMode(t, moduleModel.ModuleVersionReindexModeAutoPublish)
	cont := testutils.CreateTestServerWithDomainConfig(t, db, domainConfig)

	// Create test data - namespace, module provider, and a published version
	namespace := testutils.CreateNamespace(t, db, "testns", nil)
	moduleProvider := testutils.CreateModuleProvider(t, db, namespace.ID, "testmod", "aws")
	moduleVersion := testutils.CreateModuleVersion(t, db, moduleProvider.ID, "1.0.0")

	// Mark the version as published
	published := true
	moduleVersion.Published = &published
	err := db.DB.Save(&moduleVersion).Error
	require.NoError(t, err)

	// Update the module provider's latest_version_id
	err = db.DB.Model(&moduleProvider).Update("latest_version_id", moduleVersion.ID).Error
	require.NoError(t, err)

	// Now simulate a re-index (like uploading the same version again)
	// In AUTO_PUBLISH mode, the new version should preserve the published state
	req := moduleService.PrepareModuleRequest{
		Namespace:        "testns",
		ModuleName:       "testmod",
		Provider:         "aws",
		Version:          "1.0.0",
		ModuleProviderID: &moduleProvider.ID,
	}

	result, err := cont.ModuleCreationWrapper.PrepareModule(context.Background(), req)
	require.NoError(t, err)

	// The key assertion: in AUTO_PUBLISH mode, the new version should be published
	assert.True(t, result.ShouldPublish, "In AUTO_PUBLISH mode, reindexed version should preserve published state")
}

// TestModuleReindexMode_AutoPublish_UnpublishedVersionStaysUnpublished tests that in auto-publish mode,
// re-indexing an unpublished version preserves the unpublished state
// Python reference: test/integration/terrareg/models/test_module_version.py line 131
// (ModuleVersionReindexMode.AUTO_PUBLISH, previous_publish_state=False, config_auto_publish=False, expected_return_value=False)
func TestModuleReindexMode_AutoPublish_UnpublishedVersionStaysUnpublished(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	// Create test server with AUTO_PUBLISH reindex mode
	domainConfig := testutils.CreateTestDomainConfigWithReindexMode(t, moduleModel.ModuleVersionReindexModeAutoPublish)
	cont := testutils.CreateTestServerWithDomainConfig(t, db, domainConfig)

	// Create test data - namespace, module provider, and an unpublished version
	namespace := testutils.CreateNamespace(t, db, "testns", nil)
	moduleProvider := testutils.CreateModuleProvider(t, db, namespace.ID, "testmod", "aws")
	_ = testutils.CreateModuleVersion(t, db, moduleProvider.ID, "1.0.0")

	// Version is not published (default)

	// Now simulate a re-index
	// In AUTO_PUBLISH mode, the new version should preserve the unpublished state
	req := moduleService.PrepareModuleRequest{
		Namespace:        "testns",
		ModuleName:       "testmod",
		Provider:         "aws",
		Version:          "1.0.0",
		ModuleProviderID: &moduleProvider.ID,
	}

	result, err := cont.ModuleCreationWrapper.PrepareModule(context.Background(), req)
	require.NoError(t, err)

	// The key assertion: in AUTO_PUBLISH mode, the new version should stay unpublished
	assert.False(t, result.ShouldPublish, "In AUTO_PUBLISH mode, reindexed version should preserve unpublished state")
}

// TestModuleReindexMode_Prohibit_RaisesError tests that in prohibit mode,
// re-indexing an existing version raises an error
// Python reference: test/integration/terrareg/models/test_module_version.py line 138
// (ModuleVersionReindexMode.PROHIBIT, should_raise_error=True)
func TestModuleReindexMode_Prohibit_RaisesError(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	// Create test server with PROHIBIT reindex mode
	domainConfig := testutils.CreateTestDomainConfigWithReindexMode(t, moduleModel.ModuleVersionReindexModeProhibit)
	cont := testutils.CreateTestServerWithDomainConfig(t, db, domainConfig)

	// Create test data - namespace, module provider, and an existing version
	namespace := testutils.CreateNamespace(t, db, "testns", nil)
	moduleProvider := testutils.CreateModuleProvider(t, db, namespace.ID, "testmod", "aws")
	_ = testutils.CreateModuleVersion(t, db, moduleProvider.ID, "1.0.0")

	// Now simulate a re-index
	// In PROHIBIT mode, this should raise an error
	req := moduleService.PrepareModuleRequest{
		Namespace:        "testns",
		ModuleName:       "testmod",
		Provider:         "aws",
		Version:          "1.0.0",
		ModuleProviderID: &moduleProvider.ID,
	}

	_, err := cont.ModuleCreationWrapper.PrepareModule(context.Background(), req)

	// The key assertion: in PROHIBIT mode, re-indexing should raise an error
	assert.Error(t, err, "In PROHIBIT mode, re-indexing should raise an error")
	assert.Contains(t, err.Error(), "reindex mode is prohibit", "Error should mention prohibit mode")
}

// TestModuleReindexMode_Legacy_WithAutoPublishConfig tests that in legacy mode,
// the AUTO_PUBLISH_MODULE_VERSIONS config is respected
// Python reference: test/integration/terrareg/models/test_module_version.py line 128
// (ModuleVersionReindexMode.LEGACY, previous_publish_state=False, config_auto_publish=True, expected_return_value=True)
func TestModuleReindexMode_Legacy_WithAutoPublishConfig(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	// Create test server with LEGACY reindex mode and AUTO_PUBLISH enabled
	domainConfig := testutils.CreateTestDomainConfigWithReindexMode(t, moduleModel.ModuleVersionReindexModeLegacy)
	domainConfig.AutoPublishModuleVersions = true
	cont := testutils.CreateTestServerWithDomainConfig(t, db, domainConfig)

	// Create test data - namespace, module provider, and an unpublished version
	namespace := testutils.CreateNamespace(t, db, "testns", nil)
	moduleProvider := testutils.CreateModuleProvider(t, db, namespace.ID, "testmod", "aws")
	_ = testutils.CreateModuleVersion(t, db, moduleProvider.ID, "1.0.0")

	// Now simulate a re-index
	// In LEGACY mode with AUTO_PUBLISH enabled, the new version should be published
	req := moduleService.PrepareModuleRequest{
		Namespace:        "testns",
		ModuleName:       "testmod",
		Provider:         "aws",
		Version:          "1.0.0",
		ModuleProviderID: &moduleProvider.ID,
	}

	result, err := cont.ModuleCreationWrapper.PrepareModule(context.Background(), req)
	require.NoError(t, err)

	// The key assertion: in LEGACY mode with AUTO_PUBLISH enabled, new versions should be published
	assert.True(t, result.ShouldPublish, "In LEGACY mode with AUTO_PUBLISH enabled, new versions should be published")
}
