// Package terrareg_test provides integration tests for module version publish database state validation
package terrareg_test

import (
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/persistence/sqldb"
	"github.com/matthewjohn/terrareg/terrareg-go/test/integration/testutils"
)

// TestPublish_Database_NonBetaVersionSetsLatestVersionId tests that publishing a non-beta version
// updates the latest_version_id column in the database
// Python reference: test/integration/terrareg/models/test_module_provider.py::test_calculate_latest_version
func TestPublish_Database_NonBetaVersionSetsLatestVersionId(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	// Create test server
	cont := testutils.CreateTestServer(t, db)
	authHelper := testutils.NewAuthHelper(t, db, cont)
	cookie := authHelper.CreateSessionForUser("testuser", false, []string{}, nil)

	// Create test data - namespace, module provider, and an unpublished version
	namespace := testutils.CreateNamespace(t, db, "testns", nil)
	moduleProvider := testutils.CreateModuleProvider(t, db, namespace.ID, "testmod", "aws")
	moduleVersion := testutils.CreateModuleVersion(t, db, moduleProvider.ID, "1.0.0")

	// Verify version is not published and latest_version_id is not set
	var dbModuleProvider sqldb.ModuleProviderDB
	err := db.DB.First(&dbModuleProvider, moduleProvider.ID).Error
	require.NoError(t, err)
	assert.Nil(t, dbModuleProvider.LatestVersionID, "latest_version_id should be nil initially")

	var dbVersion sqldb.ModuleVersionDB
	err = db.DB.First(&dbVersion, moduleVersion.ID).Error
	require.NoError(t, err)
	assert.False(t, dbVersion.Published != nil && *dbVersion.Published, "version should not be published initially")

	// Publish the version via HTTP endpoint
	req := httptest.NewRequest("POST", "/v1/terrareg/modules/testns/testmod/aws/1.0.0/publish", nil)
	req.Header.Set("Cookie", cookie)
	w := httptest.NewRecorder()
	cont.Router.ServeHTTP(w, req)

	// Verify the response
	assert.Equal(t, 200, w.Code)

	// Verify database state - version should be published
	err = db.DB.First(&dbVersion, moduleVersion.ID).Error
	require.NoError(t, err)
	assert.True(t, dbVersion.Published != nil && *dbVersion.Published, "version should be published")

	// Verify database state - latest_version_id should be set
	err = db.DB.First(&dbModuleProvider, moduleProvider.ID).Error
	require.NoError(t, err)
	assert.NotNil(t, dbModuleProvider.LatestVersionID, "latest_version_id should be set")
	assert.Equal(t, moduleVersion.ID, *dbModuleProvider.LatestVersionID, "latest_version_id should point to the published version")
}

// TestPublish_Database_BetaVersionDoesNotSetLatestVersionId tests that publishing a beta version
// does NOT update the latest_version_id column in the database
// Python reference: test/integration/terrareg/models/test_module_provider.py::test_calculate_latest_version_with_beta
func TestPublish_Database_BetaVersionDoesNotSetLatestVersionId(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	// Create test server
	cont := testutils.CreateTestServer(t, db)
	authHelper := testutils.NewAuthHelper(t, db, cont)
	cookie := authHelper.CreateSessionForUser("testuser", false, []string{}, nil)

	// Create test data - namespace, module provider, and an unpublished beta version
	namespace := testutils.CreateNamespace(t, db, "testns", nil)
	moduleProvider := testutils.CreateModuleProvider(t, db, namespace.ID, "testmod", "aws")
	moduleVersion := testutils.CreateModuleVersion(t, db, moduleProvider.ID, "1.0.0-beta")

	// Mark as beta version
	moduleVersion.Beta = true
	err := db.DB.Save(&moduleVersion).Error
	require.NoError(t, err)

	// Publish the version via HTTP endpoint
	req := httptest.NewRequest("POST", "/v1/terrareg/modules/testns/testmod/aws/1.0.0-beta/publish", nil)
	req.Header.Set("Cookie", cookie)
	w := httptest.NewRecorder()
	cont.Router.ServeHTTP(w, req)

	// Verify the response
	assert.Equal(t, 200, w.Code)

	// Verify database state - version should be published
	var dbVersion sqldb.ModuleVersionDB
	err = db.DB.First(&dbVersion, moduleVersion.ID).Error
	require.NoError(t, err)
	assert.True(t, dbVersion.Published != nil && *dbVersion.Published, "beta version should be published")
	assert.True(t, dbVersion.Beta, "version should still be marked as beta")

	// Verify database state - latest_version_id should NOT be set for beta versions
	var dbModuleProvider sqldb.ModuleProviderDB
	err = db.DB.First(&dbModuleProvider, moduleProvider.ID).Error
	require.NoError(t, err)
	assert.Nil(t, dbModuleProvider.LatestVersionID, "latest_version_id should remain nil for beta versions")
}

// TestPublish_Database_MultipleVersionsUpdatesToLatest tests that publishing multiple versions
// updates latest_version_id to point to the newest non-beta published version
// Python reference: test/integration/terrareg/models/test_module_provider.py::test_calculate_latest_version_ordering
func TestPublish_Database_MultipleVersionsUpdatesToLatest(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	// Create test server
	cont := testutils.CreateTestServer(t, db)
	authHelper := testutils.NewAuthHelper(t, db, cont)
	cookie := authHelper.CreateSessionForUser("testuser", false, []string{}, nil)

	// Create test data
	namespace := testutils.CreateNamespace(t, db, "testns", nil)
	moduleProvider := testutils.CreateModuleProvider(t, db, namespace.ID, "testmod", "aws")

	// Create and publish version 1.0.0
	version1 := testutils.CreateModuleVersion(t, db, moduleProvider.ID, "1.0.0")
	req1 := httptest.NewRequest("POST", "/v1/terrareg/modules/testns/testmod/aws/1.0.0/publish", nil)
	req1.Header.Set("Cookie", cookie)
	w1 := httptest.NewRecorder()
	cont.Router.ServeHTTP(w1, req1)
	assert.Equal(t, 200, w1.Code)

	// Verify latest_version_id points to version 1.0.0
	var dbModuleProvider sqldb.ModuleProviderDB
	err := db.DB.First(&dbModuleProvider, moduleProvider.ID).Error
	require.NoError(t, err)
	assert.Equal(t, version1.ID, *dbModuleProvider.LatestVersionID, "latest_version_id should point to 1.0.0")

	// Create and publish version 1.1.0
	version2 := testutils.CreateModuleVersion(t, db, moduleProvider.ID, "1.1.0")
	req2 := httptest.NewRequest("POST", "/v1/terrareg/modules/testns/testmod/aws/1.1.0/publish", nil)
	req2.Header.Set("Cookie", cookie)
	w2 := httptest.NewRecorder()
	cont.Router.ServeHTTP(w2, req2)
	assert.Equal(t, 200, w2.Code)

	// Verify latest_version_id now points to version 1.1.0 (the newer version)
	err = db.DB.First(&dbModuleProvider, moduleProvider.ID).Error
	require.NoError(t, err)
	assert.Equal(t, version2.ID, *dbModuleProvider.LatestVersionID, "latest_version_id should update to newest version")
}

// TestPublish_Database_IdempotentRepublishDoesNotChangeLatestVersionId tests that re-publishing
// an already latest version doesn't change the latest_version_id
func TestPublish_Database_IdempotentRepublishDoesNotChangeLatestVersionId(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	// Create test server
	cont := testutils.CreateTestServer(t, db)
	authHelper := testutils.NewAuthHelper(t, db, cont)
	cookie := authHelper.CreateSessionForUser("testuser", false, []string{}, nil)

	// Create test data
	namespace := testutils.CreateNamespace(t, db, "testns", nil)
	moduleProvider := testutils.CreateModuleProvider(t, db, namespace.ID, "testmod", "aws")

	// Create and publish version 1.0.0
	version1 := testutils.CreateModuleVersion(t, db, moduleProvider.ID, "1.0.0")
	req1 := httptest.NewRequest("POST", "/v1/terrareg/modules/testns/testmod/aws/1.0.0/publish", nil)
	req1.Header.Set("Cookie", cookie)
	w1 := httptest.NewRecorder()
	cont.Router.ServeHTTP(w1, req1)
	assert.Equal(t, 200, w1.Code)

	// Get the latest_version_id after first publish
	var dbModuleProvider sqldb.ModuleProviderDB
	err := db.DB.First(&dbModuleProvider, moduleProvider.ID).Error
	require.NoError(t, err)
	firstLatestVersionID := dbModuleProvider.LatestVersionID
	assert.NotNil(t, firstLatestVersionID)
	assert.Equal(t, version1.ID, *firstLatestVersionID)

	// Publish the same version again (idempotent)
	req2 := httptest.NewRequest("POST", "/v1/terrareg/modules/testns/testmod/aws/1.0.0/publish", nil)
	req2.Header.Set("Cookie", cookie)
	w2 := httptest.NewRecorder()
	cont.Router.ServeHTTP(w2, req2)
	assert.Equal(t, 200, w2.Code)

	// Verify latest_version_id hasn't changed
	err = db.DB.First(&dbModuleProvider, moduleProvider.ID).Error
	require.NoError(t, err)
	require.NotNil(t, dbModuleProvider.LatestVersionID, "latest_version_id should not be nil after republish")
	assert.Equal(t, *firstLatestVersionID, *dbModuleProvider.LatestVersionID, "latest_version_id should not change on idempotent republish")
}
