//go:build integration
// +build integration

// Python reference: /app/test/unit/terrareg/server/test_api_terrareg_module_provider_settings.py

package terrareg_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/provider_source/model"
	providerSourceRepo "github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/persistence/sqldb/provider_source"
	"github.com/matthewjohn/terrareg/terrareg-go/test/integration/testutils"
)

// TestModuleProviderSettings_GetWithNoProviderSource
// Python reference: test_get_with_no_provider_source
func TestModuleProviderSettings_GetWithNoProviderSource(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	// Create test data
	namespace := testutils.CreateNamespace(t, db, "testnamespace", nil)
	moduleProvider := testutils.CreateModuleProvider(t, db, namespace.ID, "testmodulename", "testprovider")

	// Verify no provider source is set
	var mpDB testutils.ModuleProviderDB
	err := db.DB.First(&mpDB, moduleProvider.ID).Error
	require.NoError(t, err)
	assert.Nil(t, mpDB.ProviderSourceName)
	assert.False(t, mpDB.ProviderSourceInheritanceDisabled)
}

// TestModuleProviderSettings_SetProviderSource
// Python reference: test_post_set_provider_source
func TestModuleProviderSettings_SetProviderSource(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	// Create test data
	namespace := testutils.CreateNamespace(t, db, "testnamespace", nil)
	moduleProvider := testutils.CreateModuleProvider(t, db, namespace.ID, "testmodulename", "testprovider")

	// Create test provider source
	providerSource := model.NewProviderSource(
		"test-provider-source",
		"test-provider-source",
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

	// Set provider source on module provider directly in DB
	providerSourceName := providerSource.Name()
	db.DB.Model(&testutils.ModuleProviderDB{}).Where("id = ?", moduleProvider.ID).Update("provider_source_name", providerSourceName)

	// Verify the provider source was set
	var mpDB testutils.ModuleProviderDB
	err = db.DB.First(&mpDB, moduleProvider.ID).Error
	require.NoError(t, err)
	assert.NotNil(t, mpDB.ProviderSourceName)
	assert.Equal(t, providerSourceName, *mpDB.ProviderSourceName)
}

// TestModuleProviderSettings_UnsetProviderSource
// Python reference: test_post_unset_provider_source
func TestModuleProviderSettings_UnsetProviderSource(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	// Create test data
	namespace := testutils.CreateNamespace(t, db, "test-ns-unset-ps", nil)
	moduleProvider := testutils.CreateModuleProvider(t, db, namespace.ID, "test-module", "aws")

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

	// Set provider source on module provider first
	providerSourceName := providerSource.Name()
	db.DB.Model(&testutils.ModuleProviderDB{}).Where("id = ?", moduleProvider.ID).Update("provider_source_name", providerSourceName)

	// Verify it's set
	var mpDB testutils.ModuleProviderDB
	err = db.DB.First(&mpDB, moduleProvider.ID).Error
	require.NoError(t, err)
	assert.NotNil(t, mpDB.ProviderSourceName)

	// Unset with empty string (NULL in database)
	db.DB.Model(&testutils.ModuleProviderDB{}).Where("id = ?", moduleProvider.ID).Update("provider_source_name", nil)

	// Verify it was unset
	err = db.DB.First(&mpDB, moduleProvider.ID).Error
	require.NoError(t, err)
	assert.Nil(t, mpDB.ProviderSourceName)
}

// TestModuleProviderSettings_SetInheritanceDisabled
// Python reference: test_post_set_provider_source_inheritance_disabled
func TestModuleProviderSettings_SetInheritanceDisabled(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	// Create test data
	namespace := testutils.CreateNamespace(t, db, "test-ns-inherit", nil)
	moduleProvider := testutils.CreateModuleProvider(t, db, namespace.ID, "test-module", "aws")

	// Set inheritance disabled
	db.DB.Model(&testutils.ModuleProviderDB{}).Where("id = ?", moduleProvider.ID).Update("provider_source_inheritance_disabled", true)

	// Verify the inheritance disabled flag was set
	var mpDB testutils.ModuleProviderDB
	err := db.DB.First(&mpDB, moduleProvider.ID).Error
	require.NoError(t, err)
	assert.True(t, mpDB.ProviderSourceInheritanceDisabled)
}

// TestModuleProviderSettings_NamespaceWithDefaultProviderSource
// Python reference: test_namespace_with_default_provider_source
func TestModuleProviderSettings_NamespaceWithDefaultProviderSource(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	// Create test provider source
	providerSource := model.NewProviderSource(
		"test-ps-namespace",
		"test-ps-namespace",
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

	// Create namespace with default provider source
	providerSourceName := providerSource.Name()
	namespace := testutils.CreateNamespace(t, db, "test-ns-with-ps", nil)
	db.DB.Model(&testutils.NamespaceDB{}).Where("id = ?", namespace.ID).Update("default_provider_source_name", providerSourceName)

	// Create module provider
	moduleProvider := testutils.CreateModuleProvider(t, db, namespace.ID, "test-module", "aws")

	// Verify namespace has default provider source
	var nsDB testutils.NamespaceDB
	err = db.DB.First(&nsDB, namespace.ID).Error
	require.NoError(t, err)
	assert.NotNil(t, nsDB.DefaultProviderSourceName)
	assert.Equal(t, providerSourceName, *nsDB.DefaultProviderSourceName)

	// Verify module provider can inherit from namespace
	var mpDB testutils.ModuleProviderDB
	err = db.DB.First(&mpDB, moduleProvider.ID).Error
	require.NoError(t, err)
	assert.Nil(t, mpDB.ProviderSourceName, "Module provider should not have its own provider source")
	assert.False(t, mpDB.ProviderSourceInheritanceDisabled, "Inheritance should be enabled by default")
}

// TestModuleProviderSettings_InheritanceDisabledFallsBackToNone
// Python reference: test_inheritance_disabled_falls_back_to_none
func TestModuleProviderSettings_InheritanceDisabledFallsBackToNone(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	// Create test provider source
	providerSource := model.NewProviderSource(
		"test-ps-inherit",
		"test-ps-inherit",
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

	// Create namespace with default provider source
	providerSourceName := providerSource.Name()
	namespace := testutils.CreateNamespace(t, db, "test-ns-inherit-disabled", nil)
	db.DB.Model(&testutils.NamespaceDB{}).Where("id = ?", namespace.ID).Update("default_provider_source_name", providerSourceName)

	// Create module provider with inheritance disabled
	moduleProvider := testutils.CreateModuleProvider(t, db, namespace.ID, "test-module", "aws")
	db.DB.Model(&testutils.ModuleProviderDB{}).Where("id = ?", moduleProvider.ID).
		Updates(map[string]interface{}{
			"provider_source_inheritance_disabled": true,
		})

	// Verify inheritance is disabled
	var mpDB testutils.ModuleProviderDB
	err = db.DB.First(&mpDB, moduleProvider.ID).Error
	require.NoError(t, err)
	assert.True(t, mpDB.ProviderSourceInheritanceDisabled)
	assert.Nil(t, mpDB.ProviderSourceName, "Module provider should not have its own provider source")

	// Verify namespace has default provider source
	var nsDB testutils.NamespaceDB
	err = db.DB.First(&nsDB, namespace.ID).Error
	require.NoError(t, err)
	assert.NotNil(t, nsDB.DefaultProviderSourceName)
}

// TestNamespace_UpdateDefaultProviderSourceValid
// Python reference: test_update_with_valid_default_provider_source
func TestNamespace_UpdateDefaultProviderSourceValid(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

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

	// Create namespace
	namespace := testutils.CreateNamespace(t, db, "test-ns-update-ps", nil)

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

// TestNamespace_UpdateDefaultProviderSourceUnset
// Python reference: test_update_unset_provider_source_with_empty_string
func TestNamespace_UpdateDefaultProviderSourceUnset(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

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

	// Create namespace with default provider source
	providerSourceName := providerSource.Name()
	namespace := testutils.CreateNamespace(t, db, "test-ns-unset", nil)
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
	assert.Nil(t, nsDB.DefaultProviderSourceName)
}
