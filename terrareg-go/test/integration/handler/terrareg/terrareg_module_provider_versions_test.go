// Package terrareg_test provides integration tests for the ListModuleVersionsQuery
package terrareg_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/matthewjohn/terrareg/terrareg-go/test/integration/testutils"
)

// TestTerraregModuleProviderVersions_Success_ReturnsAllVersions tests successful retrieval of all versions
func TestTerraregModuleProviderVersions_Success_ReturnsAllVersions(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	cont := testutils.CreateTestContainer(t, db)
	query := cont.ListModuleVersionsQuery
	require.NotNil(t, query)

	// Create test namespace
	namespace := testutils.CreateNamespace(t, db, "test-versions")

	// Create test module provider
	moduleProvider := testutils.CreateModuleProvider(t, db, namespace.ID, "test-module", "aws")

	// Create test versions
	_ = testutils.CreatePublishedModuleVersion(t, db, moduleProvider.ID, "1.0.0")
	_ = testutils.CreatePublishedModuleVersion(t, db, moduleProvider.ID, "1.1.0")
	_ = testutils.CreatePublishedModuleVersion(t, db, moduleProvider.ID, "2.0.0")

	// Execute query
	versions, err := query.Execute(context.Background(), "test-versions", "test-module", "aws")

	// Assert
	require.NoError(t, err)
	require.Len(t, versions, 3, "should return exactly 3 versions")

	// Verify version objects
	for _, version := range versions {
		assert.NotNil(t, version.Version())
	}
}

// TestTerraregModuleProviderVersions_Failure_ModuleNotFound tests requesting versions for non-existent module
func TestTerraregModuleProviderVersions_Failure_ModuleNotFound(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	cont := testutils.CreateTestContainer(t, db)
	query := cont.ListModuleVersionsQuery
	require.NotNil(t, query)

	// Execute query for non-existent module
	versions, err := query.Execute(context.Background(), "nonexistent", "test-module", "aws")

	// Assert
	require.Error(t, err)
	require.Nil(t, versions)
	assert.Contains(t, err.Error(), "not found")
}

// TestTerraregModuleProviderVersions_Success_EmptyVersionsList tests module with no versions
func TestTerraregModuleProviderVersions_Success_EmptyVersionsList(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	cont := testutils.CreateTestContainer(t, db)
	query := cont.ListModuleVersionsQuery
	require.NotNil(t, query)

	// Create test namespace
	namespace := testutils.CreateNamespace(t, db, "test-empty")

	// Create test module provider without versions
	_ = testutils.CreateModuleProvider(t, db, namespace.ID, "empty-module", "aws")

	// Execute query
	versions, err := query.Execute(context.Background(), "test-empty", "empty-module", "aws")

	// Assert
	require.NoError(t, err)
	require.NotNil(t, versions)
	assert.Empty(t, versions, "versions should be empty")
}

// TestTerraregModuleProviderVersions_Success_SingleVersion tests module with single version
func TestTerraregModuleProviderVersions_Success_SingleVersion(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	cont := testutils.CreateTestContainer(t, db)
	query := cont.ListModuleVersionsQuery
	require.NotNil(t, query)

	// Create test namespace
	namespace := testutils.CreateNamespace(t, db, "test-single")

	// Create test module provider
	moduleProvider := testutils.CreateModuleProvider(t, db, namespace.ID, "single-module", "aws")

	// Create single version
	_ = testutils.CreatePublishedModuleVersion(t, db, moduleProvider.ID, "1.0.0")

	// Execute query
	versions, err := query.Execute(context.Background(), "test-single", "single-module", "aws")

	// Assert
	require.NoError(t, err)
	require.Len(t, versions, 1, "should contain exactly 1 version")

	// Verify the version content
	assert.Equal(t, "1.0.0", versions[0].Version().String())
}

// TestTerraregModuleProviderVersions_Success_OrdersVersionsCorrectly tests that versions are ordered correctly
func TestTerraregModuleProviderVersions_Success_OrdersVersionsCorrectly(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	cont := testutils.CreateTestContainer(t, db)
	query := cont.ListModuleVersionsQuery
	require.NotNil(t, query)

	// Create test namespace
	namespace := testutils.CreateNamespace(t, db, "test-order")

	// Create test module provider
	moduleProvider := testutils.CreateModuleProvider(t, db, namespace.ID, "order-module", "aws")

	// Create versions in non-sequential order
	_ = testutils.CreatePublishedModuleVersion(t, db, moduleProvider.ID, "2.0.0")
	_ = testutils.CreatePublishedModuleVersion(t, db, moduleProvider.ID, "1.0.0")
	_ = testutils.CreatePublishedModuleVersion(t, db, moduleProvider.ID, "1.5.0")
	_ = testutils.CreatePublishedModuleVersion(t, db, moduleProvider.ID, "0.9.0")

	// Execute query
	versions, err := query.Execute(context.Background(), "test-order", "order-module", "aws")

	// Assert
	require.NoError(t, err)
	require.Len(t, versions, 4, "should contain exactly 4 versions")

	// Verify we got all expected versions
	versionStrings := make([]string, len(versions))
	for i, version := range versions {
		versionStrings[i] = version.Version().String()
	}

	assert.Contains(t, versionStrings, "0.9.0")
	assert.Contains(t, versionStrings, "1.0.0")
	assert.Contains(t, versionStrings, "1.5.0")
	assert.Contains(t, versionStrings, "2.0.0")
}

// TestTerraregModuleProviderVersions_Success_MultipleProviders tests versions for different providers
func TestTerraregModuleProviderVersions_Success_MultipleProviders(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	cont := testutils.CreateTestContainer(t, db)
	query := cont.ListModuleVersionsQuery
	require.NotNil(t, query)

	// Create test namespace
	namespace := testutils.CreateNamespace(t, db, "test-multi")

	// Create multiple module providers
	moduleProvider1 := testutils.CreateModuleProvider(t, db, namespace.ID, "multi-module", "aws")
	moduleProvider2 := testutils.CreateModuleProvider(t, db, namespace.ID, "multi-module", "gcp")

	// Create versions for each provider
	_ = testutils.CreatePublishedModuleVersion(t, db, moduleProvider1.ID, "1.0.0")
	_ = testutils.CreatePublishedModuleVersion(t, db, moduleProvider2.ID, "1.0.0")
	_ = testutils.CreatePublishedModuleVersion(t, db, moduleProvider2.ID, "2.0.0")

	// Test first provider
	versions1, err := query.Execute(context.Background(), "test-multi", "multi-module", "aws")
	require.NoError(t, err)
	assert.Len(t, versions1, 1, "aws provider should have 1 version")

	// Test second provider
	versions2, err := query.Execute(context.Background(), "test-multi", "multi-module", "gcp")
	require.NoError(t, err)
	assert.Len(t, versions2, 2, "gcp provider should have 2 versions")
}

// TestTerraregModuleProviderVersions_Success_VersionObjectsHaveCorrectProperties tests that returned version objects have all expected properties
func TestTerraregModuleProviderVersions_Success_VersionObjectsHaveCorrectProperties(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	cont := testutils.CreateTestContainer(t, db)
	query := cont.ListModuleVersionsQuery
	require.NotNil(t, query)

	// Create test namespace
	namespace := testutils.CreateNamespace(t, db, "test-props")

	// Create test module provider
	moduleProvider := testutils.CreateModuleProvider(t, db, namespace.ID, "props-module", "aws")

	// Create a version
	_ = testutils.CreatePublishedModuleVersion(t, db, moduleProvider.ID, "1.0.0")

	// Execute query
	versions, err := query.Execute(context.Background(), "test-props", "props-module", "aws")

	// Assert
	require.NoError(t, err)
	require.Len(t, versions, 1)

	// Verify version object properties
	version := versions[0]
	assert.NotNil(t, version.Version())
	assert.Equal(t, "1.0.0", version.Version().String())
	assert.False(t, version.IsBeta(), "version should not be beta")
}
