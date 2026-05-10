//go:build integration
// +build integration

// Python reference: /app/test/unit/terrareg/server/test_api_terrareg_module_provider_settings.py (audit tests)
// Python reference: /app/test/unit/terrareg/server/test_api_terrareg_namespace_details.py (audit tests)

package audit

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/application/command/module"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/application/command/namespace"
	auditService "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/audit/service"
	providerSourceService "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/provider_source/service"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/shared/types"
	auditRepo "github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/persistence/sqldb/audit"
	"github.com/matthewjohn/terrareg/terrareg-go/test/integration/testutils"
)

// TestModuleProviderAudit_SetProviderSource tests audit trail for setting provider source
func TestModuleProviderAudit_SetProviderSource(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	ctx := testutils.GetTestContext(t)

	// Create test data
	namespaceDB := testutils.CreateNamespace(t, db, "test-namespace-audit-mp", nil)
	moduleProviderDB := testutils.CreateModuleProvider(t, db, namespaceDB.ID, "testmodule", "testprovider")
	providerSource := testutils.CreateTestProviderSource(t, db, "test-ps-audit-mp")

	// Create application services
	repos := testutils.CreateTestRepositories(t, db)
	auditHistoryRepo, _ := auditRepo.NewAuditHistoryRepository(db.DB)
	moduleAuditService := auditService.NewModuleAuditService(auditHistoryRepo)
	providerSourceFactory := providerSourceService.NewProviderSourceFactory(repos.ProviderSource)
	cmd := module.NewUpdateModuleProviderSettingsCommand(repos.ModuleProvider, providerSourceFactory, moduleAuditService)

	// Execute command to set provider source
	providerSourceName := providerSource.Name
	err := cmd.Execute(ctx, module.UpdateModuleProviderSettingsRequest{
		Namespace:      namespaceDB.Namespace,
		Module:         "testmodule",
		Provider:       "testprovider",
		ProviderSource: &providerSourceName,
	})
	require.NoError(t, err)

	// Verify audit entry was created
	auditEntries := testutils.GetAuditEntriesForObject(t, db, "ModuleProvider", testutils.IntToString(moduleProviderDB.ID))
	assert.Greater(t, len(auditEntries), 0, "Expected audit entry to be created")

	// Find the provider source update audit entry
	var providerSourceAuditEntry *testutils.AuditHistoryEntry
	for _, entry := range auditEntries {
		if entry.Action == "module_provider_update_provider_source" {
			providerSourceAuditEntry = &entry
			break
		}
	}
	require.NotNil(t, providerSourceAuditEntry, "Expected audit entry for provider source update")
	assert.Equal(t, "module_provider_update_provider_source", providerSourceAuditEntry.Action)
	assert.Equal(t, "ModuleProvider", providerSourceAuditEntry.ObjectType)
	assert.Equal(t, testutils.IntToString(moduleProviderDB.ID), providerSourceAuditEntry.ObjectID)
	assert.Nil(t, providerSourceAuditEntry.OldValue, "Old value should be nil for first set")
	assert.NotNil(t, providerSourceAuditEntry.NewValue)
	assert.Equal(t, providerSourceName, *providerSourceAuditEntry.NewValue)
}

// TestModuleProviderAudit_UpdateProviderSourceFromValue tests audit trail when changing provider source
func TestModuleProviderAudit_UpdateProviderSourceFromValue(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	ctx := testutils.GetTestContext(t)

	// Create test data
	namespaceDB := testutils.CreateNamespace(t, db, "test-namespace-audit-change", nil)
	moduleProviderDB := testutils.CreateModuleProvider(t, db, namespaceDB.ID, "testmodule", "testprovider")

	// Create two provider sources
	providerSource1 := testutils.CreateTestProviderSource(t, db, "test-ps-audit-1")
	providerSource2 := testutils.CreateTestProviderSource(t, db, "test-ps-audit-2")

	// Create application services
	repos := testutils.CreateTestRepositories(t, db)
	auditHistoryRepo, _ := auditRepo.NewAuditHistoryRepository(db.DB)
	moduleAuditService := auditService.NewModuleAuditService(auditHistoryRepo)
	providerSourceFactory := providerSourceService.NewProviderSourceFactory(repos.ProviderSource)
	cmd := module.NewUpdateModuleProviderSettingsCommand(repos.ModuleProvider, providerSourceFactory, moduleAuditService)

	// Set first provider source
	providerSourceName1 := providerSource1.Name
	err := cmd.Execute(ctx, module.UpdateModuleProviderSettingsRequest{
		Namespace:      namespaceDB.Namespace,
		Module:         "testmodule",
		Provider:       "testprovider",
		ProviderSource: &providerSourceName1,
	})
	require.NoError(t, err)

	// Update to second provider source
	providerSourceName2 := providerSource2.Name
	err = cmd.Execute(ctx, module.UpdateModuleProviderSettingsRequest{
		Namespace:      namespaceDB.Namespace,
		Module:         "testmodule",
		Provider:       "testprovider",
		ProviderSource: &providerSourceName2,
	})
	require.NoError(t, err)

	// Verify audit entries
	auditEntries := testutils.GetAuditEntriesForObject(t, db, "ModuleProvider", testutils.IntToString(moduleProviderDB.ID))
	assert.GreaterOrEqual(t, len(auditEntries), 2, "Expected at least 2 audit entries")

	// Find the second update audit entry
	var updateAuditEntry *testutils.AuditHistoryEntry
	for _, entry := range auditEntries {
		if entry.Action == "module_provider_update_provider_source" {
			if entry.NewValue != nil && *entry.NewValue == providerSourceName2 {
				updateAuditEntry = &entry
				break
			}
		}
	}
	require.NotNil(t, updateAuditEntry, "Expected audit entry for second provider source update")
	assert.Equal(t, providerSourceName1, *updateAuditEntry.OldValue)
	assert.Equal(t, providerSourceName2, *updateAuditEntry.NewValue)
}

// TestModuleProviderAudit_UnsetProviderSource tests audit trail for unsetting provider source
func TestModuleProviderAudit_UnsetProviderSource(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	ctx := testutils.GetTestContext(t)

	// Create test data
	namespaceDB := testutils.CreateNamespace(t, db, "test-namespace-audit-unset", nil)
	moduleProviderDB := testutils.CreateModuleProvider(t, db, namespaceDB.ID, "testmodule", "testprovider")
	providerSource := testutils.CreateTestProviderSource(t, db, "test-ps-audit-unset")

	// Create application services
	repos := testutils.CreateTestRepositories(t, db)
	auditHistoryRepo, _ := auditRepo.NewAuditHistoryRepository(db.DB)
	moduleAuditService := auditService.NewModuleAuditService(auditHistoryRepo)
	providerSourceFactory := providerSourceService.NewProviderSourceFactory(repos.ProviderSource)
	cmd := module.NewUpdateModuleProviderSettingsCommand(repos.ModuleProvider, providerSourceFactory, moduleAuditService)

	// Set provider source first
	providerSourceName := providerSource.Name
	err := cmd.Execute(ctx, module.UpdateModuleProviderSettingsRequest{
		Namespace:      namespaceDB.Namespace,
		Module:         "testmodule",
		Provider:       "testprovider",
		ProviderSource: &providerSourceName,
	})
	require.NoError(t, err)

	// Clear audit entries for this test
	db.DB.Exec("DELETE FROM audit_history WHERE object_type = 'ModuleProvider' AND object_id = ?", testutils.IntToString(moduleProviderDB.ID))

	// Unset provider source by passing empty string
	emptyString := ""
	err = cmd.Execute(ctx, module.UpdateModuleProviderSettingsRequest{
		Namespace:      namespaceDB.Namespace,
		Module:         "testmodule",
		Provider:       "testprovider",
		ProviderSource: &emptyString,
	})
	require.NoError(t, err)

	// Verify audit entry was created
	auditEntries := testutils.GetAuditEntriesForObject(t, db, "ModuleProvider", testutils.IntToString(moduleProviderDB.ID))
	assert.Greater(t, len(auditEntries), 0, "Expected audit entry to be created")

	// Find the unset audit entry
	var unsetAuditEntry *testutils.AuditHistoryEntry
	for _, entry := range auditEntries {
		if entry.Action == "module_provider_update_provider_source" {
			unsetAuditEntry = &entry
			break
		}
	}
	require.NotNil(t, unsetAuditEntry, "Expected audit entry for unsetting provider source")
	assert.Equal(t, providerSourceName, *unsetAuditEntry.OldValue)
	assert.Nil(t, unsetAuditEntry.NewValue, "New value should be nil when unsetting")
}

// TestModuleProviderAudit_SetInheritanceDisabled tests audit trail for disabling provider source inheritance
func TestModuleProviderAudit_SetInheritanceDisabled(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	ctx := testutils.GetTestContext(t)

	// Create test data
	namespaceDB := testutils.CreateNamespace(t, db, "test-namespace-audit-inherit", nil)
	moduleProviderDB := testutils.CreateModuleProvider(t, db, namespaceDB.ID, "testmodule", "testprovider")

	// Create application services
	repos := testutils.CreateTestRepositories(t, db)
	auditHistoryRepo, _ := auditRepo.NewAuditHistoryRepository(db.DB)
	moduleAuditService := auditService.NewModuleAuditService(auditHistoryRepo)
	providerSourceFactory := providerSourceService.NewProviderSourceFactory(repos.ProviderSource)
	cmd := module.NewUpdateModuleProviderSettingsCommand(repos.ModuleProvider, providerSourceFactory, moduleAuditService)

	// Disable provider source inheritance
	inheritanceDisabled := true
	err := cmd.Execute(ctx, module.UpdateModuleProviderSettingsRequest{
		Namespace:                         namespaceDB.Namespace,
		Module:                            "testmodule",
		Provider:                          "testprovider",
		ProviderSourceInheritanceDisabled: &inheritanceDisabled,
	})
	require.NoError(t, err)

	// Verify audit entry was created
	auditEntries := testutils.GetAuditEntriesForObject(t, db, "ModuleProvider", testutils.IntToString(moduleProviderDB.ID))
	assert.Greater(t, len(auditEntries), 0, "Expected audit entry to be created")

	// Find the inheritance disabled audit entry
	var inheritanceAuditEntry *testutils.AuditHistoryEntry
	for _, entry := range auditEntries {
		if entry.Action == "module_provider_update_provider_source_inheritance_disabled" {
			inheritanceAuditEntry = &entry
			break
		}
	}
	require.NotNil(t, inheritanceAuditEntry, "Expected audit entry for inheritance disabled")
	assert.Equal(t, "false", *inheritanceAuditEntry.OldValue)
	assert.Equal(t, "true", *inheritanceAuditEntry.NewValue)
}

// TestModuleProviderAudit_SetInheritanceDisabledEnable tests enabling inheritance
func TestModuleProviderAudit_SetInheritanceDisabledEnable(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	ctx := testutils.GetTestContext(t)

	// Create test data
	namespaceDB := testutils.CreateNamespace(t, db, "test-namespace-audit-enable", nil)
	moduleProviderDB := testutils.CreateModuleProvider(t, db, namespaceDB.ID, "testmodule", "testprovider")

	// Create application services
	repos := testutils.CreateTestRepositories(t, db)
	auditHistoryRepo, _ := auditRepo.NewAuditHistoryRepository(db.DB)
	moduleAuditService := auditService.NewModuleAuditService(auditHistoryRepo)
	providerSourceFactory := providerSourceService.NewProviderSourceFactory(repos.ProviderSource)
	cmd := module.NewUpdateModuleProviderSettingsCommand(repos.ModuleProvider, providerSourceFactory, moduleAuditService)

	// First disable inheritance
	inheritanceDisabled := true
	err := cmd.Execute(ctx, module.UpdateModuleProviderSettingsRequest{
		Namespace:                         namespaceDB.Namespace,
		Module:                            "testmodule",
		Provider:                          "testprovider",
		ProviderSourceInheritanceDisabled: &inheritanceDisabled,
	})
	require.NoError(t, err)

	// Clear audit entries
	db.DB.Exec("DELETE FROM audit_history WHERE object_type = 'ModuleProvider' AND object_id = ?", testutils.IntToString(moduleProviderDB.ID))

	// Enable inheritance
	inheritanceDisabled = false
	err = cmd.Execute(ctx, module.UpdateModuleProviderSettingsRequest{
		Namespace:                         namespaceDB.Namespace,
		Module:                            "testmodule",
		Provider:                          "testprovider",
		ProviderSourceInheritanceDisabled: &inheritanceDisabled,
	})
	require.NoError(t, err)

	// Verify audit entry was created
	auditEntries := testutils.GetAuditEntriesForObject(t, db, "ModuleProvider", testutils.IntToString(moduleProviderDB.ID))
	assert.Greater(t, len(auditEntries), 0, "Expected audit entry to be created")

	// Find the inheritance enabled audit entry
	var inheritanceAuditEntry *testutils.AuditHistoryEntry
	for _, entry := range auditEntries {
		if entry.Action == "module_provider_update_provider_source_inheritance_disabled" {
			inheritanceAuditEntry = &entry
			break
		}
	}
	require.NotNil(t, inheritanceAuditEntry, "Expected audit entry for inheritance enabled")
	assert.Equal(t, "true", *inheritanceAuditEntry.OldValue)
	assert.Equal(t, "false", *inheritanceAuditEntry.NewValue)
}

// TestModuleProviderAudit_SetInheritanceDisabledNoChange tests that no audit is created when value doesn't change
func TestModuleProviderAudit_SetInheritanceDisabledNoChange(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	ctx := testutils.GetTestContext(t)

	// Create test data
	namespaceDB := testutils.CreateNamespace(t, db, "test-namespace-audit-nochange", nil)
	moduleProviderDB := testutils.CreateModuleProvider(t, db, namespaceDB.ID, "testmodule", "testprovider")

	// Create application services
	repos := testutils.CreateTestRepositories(t, db)
	auditHistoryRepo, _ := auditRepo.NewAuditHistoryRepository(db.DB)
	moduleAuditService := auditService.NewModuleAuditService(auditHistoryRepo)
	providerSourceFactory := providerSourceService.NewProviderSourceFactory(repos.ProviderSource)
	cmd := module.NewUpdateModuleProviderSettingsCommand(repos.ModuleProvider, providerSourceFactory, moduleAuditService)

	// Set inheritance to disabled
	inheritanceDisabled := true
	err := cmd.Execute(ctx, module.UpdateModuleProviderSettingsRequest{
		Namespace:                         namespaceDB.Namespace,
		Module:                            "testmodule",
		Provider:                          "testprovider",
		ProviderSourceInheritanceDisabled: &inheritanceDisabled,
	})
	require.NoError(t, err)

	// Get current count of audit entries
	auditEntriesBefore := testutils.GetAuditEntriesForObject(t, db, "ModuleProvider", testutils.IntToString(moduleProviderDB.ID))
	countBefore := len(auditEntriesBefore)

	// Try to set to same value
	err = cmd.Execute(ctx, module.UpdateModuleProviderSettingsRequest{
		Namespace:                         namespaceDB.Namespace,
		Module:                            "testmodule",
		Provider:                          "testprovider",
		ProviderSourceInheritanceDisabled: &inheritanceDisabled,
	})
	require.NoError(t, err)

	// Verify no new audit entry was created
	auditEntriesAfter := testutils.GetAuditEntriesForObject(t, db, "ModuleProvider", testutils.IntToString(moduleProviderDB.ID))
	countAfter := len(auditEntriesAfter)
	assert.Equal(t, countBefore, countAfter, "No new audit entry should be created when value doesn't change")
}

// TestNamespaceAudit_SetDefaultProviderSource tests audit trail for setting namespace default provider source
func TestNamespaceAudit_SetDefaultProviderSource(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	ctx := testutils.GetTestContext(t)

	// Create test namespace
	namespaceDB := testutils.CreateNamespace(t, db, "test-ns-audit-set", nil)
	providerSource := testutils.CreateTestProviderSource(t, db, "test-ps-ns-audit-set")

	// Create application services
	repos := testutils.CreateTestRepositories(t, db)
	auditHistoryRepo, _ := auditRepo.NewAuditHistoryRepository(db.DB)
	_ = auditService.NewModuleAuditService(auditHistoryRepo)
	appServices := testutils.CreateTestApplicationServices(t, db, repos)

	// Execute command to set default provider source
	providerSourceName := providerSource.Name
	response, err := appServices.UpdateNamespace.Execute(ctx, types.NamespaceName(namespaceDB.Namespace), namespace.UpdateNamespaceRequest{
		DefaultProviderSource: &providerSourceName,
	})
	require.NoError(t, err)
	require.NotNil(t, response)

	// Verify audit entry was created
	auditEntries := testutils.GetAuditEntriesForObject(t, db, "Namespace", namespaceDB.Namespace)
	assert.Greater(t, len(auditEntries), 0, "Expected audit entry to be created")

	// Find the provider source update audit entry
	var providerSourceAuditEntry *testutils.AuditHistoryEntry
	for _, entry := range auditEntries {
		if entry.Action == "namespace_modify_default_provider_source" {
			providerSourceAuditEntry = &entry
			break
		}
	}
	require.NotNil(t, providerSourceAuditEntry, "Expected audit entry for default provider source update")
	assert.Nil(t, providerSourceAuditEntry.OldValue, "Old value should be nil for first set")
	assert.NotNil(t, providerSourceAuditEntry.NewValue)
	assert.Equal(t, providerSourceName, *providerSourceAuditEntry.NewValue)
}

// TestNamespaceAudit_UpdateDefaultProviderSourceFromValue tests audit trail when changing default provider source
func TestNamespaceAudit_UpdateDefaultProviderSourceFromValue(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	ctx := testutils.GetTestContext(t)

	// Create test namespace
	namespaceDB := testutils.CreateNamespace(t, db, "test-ns-audit-change", nil)

	// Create two provider sources
	providerSource1 := testutils.CreateTestProviderSource(t, db, "test-ps-ns-audit-1")
	providerSource2 := testutils.CreateTestProviderSource(t, db, "test-ps-ns-audit-2")

	// Create application services
	repos := testutils.CreateTestRepositories(t, db)
	auditHistoryRepo, _ := auditRepo.NewAuditHistoryRepository(db.DB)
	_ = auditService.NewModuleAuditService(auditHistoryRepo)
	appServices := testutils.CreateTestApplicationServices(t, db, repos)

	// Set first default provider source
	providerSourceName1 := providerSource1.Name
	_, err := appServices.UpdateNamespace.Execute(ctx, types.NamespaceName(namespaceDB.Namespace), namespace.UpdateNamespaceRequest{
		DefaultProviderSource: &providerSourceName1,
	})
	require.NoError(t, err)

	// Update to second default provider source
	providerSourceName2 := providerSource2.Name
	_, err = appServices.UpdateNamespace.Execute(ctx, types.NamespaceName(namespaceDB.Namespace), namespace.UpdateNamespaceRequest{
		DefaultProviderSource: &providerSourceName2,
	})
	require.NoError(t, err)

	// Verify audit entries
	auditEntries := testutils.GetAuditEntriesForObject(t, db, "Namespace", namespaceDB.Namespace)
	assert.GreaterOrEqual(t, len(auditEntries), 2, "Expected at least 2 audit entries")

	// Find the second update audit entry
	var updateAuditEntry *testutils.AuditHistoryEntry
	for _, entry := range auditEntries {
		if entry.Action == "namespace_modify_default_provider_source" {
			if entry.NewValue != nil && *entry.NewValue == providerSourceName2 {
				updateAuditEntry = &entry
				break
			}
		}
	}
	require.NotNil(t, updateAuditEntry, "Expected audit entry for second default provider source update")
	assert.Equal(t, providerSourceName1, *updateAuditEntry.OldValue)
	assert.Equal(t, providerSourceName2, *updateAuditEntry.NewValue)
}

// TestNamespaceAudit_UnsetDefaultProviderSource tests audit trail for unsetting default provider source
func TestNamespaceAudit_UnsetDefaultProviderSource(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	ctx := testutils.GetTestContext(t)

	// Create test namespace
	namespaceDB := testutils.CreateNamespace(t, db, "test-ns-audit-unset", nil)
	providerSource := testutils.CreateTestProviderSource(t, db, "test-ps-ns-audit-unset")

	// Create application services
	repos := testutils.CreateTestRepositories(t, db)
	auditHistoryRepo, _ := auditRepo.NewAuditHistoryRepository(db.DB)
	_ = auditService.NewModuleAuditService(auditHistoryRepo)
	appServices := testutils.CreateTestApplicationServices(t, db, repos)

	// Set default provider source first
	providerSourceName := providerSource.Name
	_, err := appServices.UpdateNamespace.Execute(ctx, types.NamespaceName(namespaceDB.Namespace), namespace.UpdateNamespaceRequest{
		DefaultProviderSource: &providerSourceName,
	})
	require.NoError(t, err)

	// Clear audit entries
	db.DB.Exec("DELETE FROM audit_history WHERE object_type = 'Namespace' AND object_id = ?", namespaceDB.Namespace)

	// Unset default provider source by passing empty string
	emptyString := ""
	_, err = appServices.UpdateNamespace.Execute(ctx, types.NamespaceName(namespaceDB.Namespace), namespace.UpdateNamespaceRequest{
		DefaultProviderSource: &emptyString,
	})
	require.NoError(t, err)

	// Verify audit entry was created
	auditEntries := testutils.GetAuditEntriesForObject(t, db, "Namespace", namespaceDB.Namespace)
	assert.Greater(t, len(auditEntries), 0, "Expected audit entry to be created")

	// Find the unset audit entry
	var unsetAuditEntry *testutils.AuditHistoryEntry
	for _, entry := range auditEntries {
		if entry.Action == "namespace_modify_default_provider_source" {
			unsetAuditEntry = &entry
			break
		}
	}
	require.NotNil(t, unsetAuditEntry, "Expected audit entry for unsetting default provider source")
	assert.Equal(t, providerSourceName, *unsetAuditEntry.OldValue)
	assert.Nil(t, unsetAuditEntry.NewValue, "New value should be nil when unsetting")
}

// TestAuditTrail_Interoperability_PythonToGo verifies that Go can read audit entries created by Python
func TestAuditTrail_Interoperability_PythonToGo(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	// Simulate Python creating an audit entry
	namespace := testutils.CreateNamespace(t, db, "test-ns-python-compat", nil)
	moduleProviderDB := testutils.CreateModuleProvider(t, db, namespace.ID, "testmodule", "testprovider")
	providerSource := testutils.CreateTestProviderSource(t, db, "test-ps-python-compat")

	// Manually insert audit entry in Python format
	auditEntry := map[string]interface{}{
		"username":    "python-user",
		"action":      "module_provider_update_provider_source",
		"object_type": "ModuleProvider",
		"object_id":   testutils.IntToString(moduleProviderDB.ID),
		"old_value":   nil,
		"new_value":   providerSource.Name,
	}
	err := db.DB.Table("audit_history").Create(&auditEntry).Error
	require.NoError(t, err)

	// Verify Go can read the Python-created audit entry
	auditEntries := testutils.GetAuditEntriesForObject(t, db, "ModuleProvider", testutils.IntToString(moduleProviderDB.ID))
	assert.Greater(t, len(auditEntries), 0, "Expected to find Python-created audit entry")

	entry := auditEntries[0]
	assert.Equal(t, "python-user", entry.Username)
	assert.Equal(t, "module_provider_update_provider_source", entry.Action)
	assert.Equal(t, "ModuleProvider", entry.ObjectType)
	assert.Equal(t, testutils.IntToString(moduleProviderDB.ID), entry.ObjectID)
	assert.Nil(t, entry.OldValue)
	assert.NotNil(t, entry.NewValue)
	assert.Equal(t, providerSource.Name, *entry.NewValue)
}

// TestAuditTrail_Interoperability_GoToPython verifies that Go creates audit entries in Python-compatible format
func TestAuditTrail_Interoperability_GoToPython(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	ctx := testutils.GetTestContext(t)

	// Create test data
	namespaceDB := testutils.CreateNamespace(t, db, "test-ns-go-compat", nil)
	providerSource := testutils.CreateTestProviderSource(t, db, "test-ps-go-compat")

	// Create application services
	repos := testutils.CreateTestRepositories(t, db)
	auditHistoryRepo, _ := auditRepo.NewAuditHistoryRepository(db.DB)
	_ = auditService.NewModuleAuditService(auditHistoryRepo)
	appServices := testutils.CreateTestApplicationServices(t, db, repos)

	// Set provider source using Go
	providerSourceName := providerSource.Name
	_, err := appServices.UpdateNamespace.Execute(ctx, types.NamespaceName(namespaceDB.Namespace), namespace.UpdateNamespaceRequest{
		DefaultProviderSource: &providerSourceName,
	})
	require.NoError(t, err)

	// Verify Go-created audit entry matches Python format
	auditEntries := testutils.GetAuditEntriesForObject(t, db, "Namespace", namespaceDB.Namespace)
	assert.Greater(t, len(auditEntries), 0, "Expected Go to create audit entry")

	entry := auditEntries[0]
	assert.Equal(t, "namespace_modify_default_provider_source", entry.Action)
	assert.Equal(t, "Namespace", entry.ObjectType)
	assert.Equal(t, namespaceDB.Namespace, entry.ObjectID)
	assert.Nil(t, entry.OldValue, "Old value should be nil for first set")
	assert.NotNil(t, entry.NewValue)
	assert.Equal(t, providerSourceName, *entry.NewValue)
	assert.NotEmpty(t, entry.Username)
	assert.NotEmpty(t, entry.Timestamp)
}
