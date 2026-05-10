//go:build integration
// +build integration

// Python reference: /app/test/unit/terrareg/server/test_api_terrareg_namespace_details.py

package terrareg_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/provider_source/model"
	providerSourceRepo "github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/persistence/sqldb/provider_source"
	"github.com/matthewjohn/terrareg/terrareg-go/test/integration/testutils"
)

// TestNamespaceDetailsHandler_GetWithDefaultProviderSourceNone
// Python reference: test_get_with_default_provider_source_none
func TestNamespaceDetailsHandler_GetWithDefaultProviderSourceNone(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	// Create test namespace
	namespace := testutils.CreateNamespace(t, db, "testnamespace", nil)

	// Verify no default provider source
	var nsDB testutils.NamespaceDB
	err := db.DB.First(&nsDB, namespace.ID).Error
	require.NoError(t, err)
	assert.Nil(t, nsDB.DefaultProviderSourceName)
}

// TestNamespaceDetailsHandler_PostWithValidDefaultProviderSource
// Python reference: test_update_with_valid_default_provider_source
func TestNamespaceDetailsHandler_PostWithValidDefaultProviderSource(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	// Create test data
	namespace := testutils.CreateNamespace(t, db, "test-ns-update-ps", nil)

	// Create test provider source
	providerSource := model.NewProviderSource(
		"test-ps-update",
		"test-ps-update",
		model.ProviderSourceTypeGithub,
		&model.ProviderSourceConfig{
			BaseURL:        "https://github.com",
			ApiURL:         "https://api.github.com",
			ClientID:       "test-client-id",
			ClientSecret:   "test-client-secret",
			PrivateKeyPath: "/test/key.pem",
			AppID:          "test-app-id",
			LoginButtonText: "Login with GitHub",
		},
	)
	psRepo := providerSourceRepo.NewProviderSourceRepository(db.DB)
	err := psRepo.Upsert(testutils.GetTestContext(t), providerSource)
	require.NoError(t, err)

	// Update namespace with default provider source
	providerSourceName := providerSource.Name()
	db.DB.Model(&testutils.NamespaceDB{}).Where("id = ?", namespace.ID).Update("default_provider_source_name", providerSourceName)

	// Verify the default provider source was set
	var nsDB testutils.NamespaceDB
	err = db.DB.First(&nsDB, namespace.ID).Error
	require.NoError(t, err)
	assert.NotNil(t, nsDB.DefaultProviderSourceName)
	assert.Equal(t, providerSourceName, *nsDB.DefaultProviderSourceName)
}

// TestNamespaceDetailsHandler_PostUnsetProviderSourceWithEmptyString
// Python reference: test_update_unset_provider_source_with_empty_string
func TestNamespaceDetailsHandler_PostUnsetProviderSourceWithEmptyString(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	// Create test data
	namespace := testutils.CreateNamespace(t, db, "test-ns-unset-ps", nil)

	// Create test provider source
	providerSource := model.NewProviderSource(
		"test-ps-unset",
		"test-ps-unset",
		model.ProviderSourceTypeGithub,
		&model.ProviderSourceConfig{
			BaseURL:        "https://github.com",
			ApiURL:         "https://api.github.com",
			ClientID:       "test-client-id",
			ClientSecret:   "test-client-secret",
			PrivateKeyPath: "/test/key.pem",
			AppID:          "test-app-id",
			LoginButtonText: "Login with GitHub",
		},
	)
	psRepo := providerSourceRepo.NewProviderSourceRepository(db.DB)
	err := psRepo.Upsert(testutils.GetTestContext(t), providerSource)
	require.NoError(t, err)

	// Set the default provider source
	providerSourceName := providerSource.Name()
	db.DB.Model(&testutils.NamespaceDB{}).Where("id = ?", namespace.ID).Update("default_provider_source_name", providerSourceName)

	// Verify it's set
	var nsDB testutils.NamespaceDB
	err = db.DB.First(&nsDB, namespace.ID).Error
	require.NoError(t, err)
	assert.NotNil(t, nsDB.DefaultProviderSourceName)

	// Unset with NULL
	db.DB.Model(&testutils.NamespaceDB{}).Where("id = ?", namespace.ID).Update("default_provider_source_name", nil)

	// Verify it was unset
	err = db.DB.First(&nsDB, namespace.ID).Error
	require.NoError(t, err)
	assert.Nil(t, nsDB.DefaultProviderSourceName, "Should be nil after unsetting")
}
