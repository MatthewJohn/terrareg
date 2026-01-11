package v1_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	moduleQuery "github.com/matthewjohn/terrareg/terrareg-go/internal/application/query/module"
	moduleRepo "github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/persistence/sqldb/module"
	v1 "github.com/matthewjohn/terrareg/terrareg-go/internal/interfaces/http/handler/terraform/v1"
	"github.com/matthewjohn/terrareg/terrareg-go/test/integration/testutils"
)

// TestModuleListHandler_HandleListModules_Success tests listing modules with published versions
func TestModuleListHandler_HandleListModules_Success(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	// Create test data - namespace, module provider, and published version
	namespace := testutils.CreateNamespace(t, db, "test-namespace")
	moduleProvider := testutils.CreateModuleProvider(t, db, namespace.ID, "test-module", "aws")
	_ = testutils.CreatePublishedModuleVersion(t, db, moduleProvider.ID, "1.0.0")

	// Create handler with required dependencies
	namespaceRepository := moduleRepo.NewNamespaceRepository(db.DB)
	moduleProviderRepository := moduleRepo.NewModuleProviderRepository(db.DB, namespaceRepository, nil)
	listModulesQuery := moduleQuery.NewListModulesQuery(moduleProviderRepository)
	handler := v1.NewModuleListHandler(listModulesQuery)

	// Create request
	req := httptest.NewRequest("GET", "/v1/modules", nil)
	w := httptest.NewRecorder()

	// Act
	handler.HandleListModules(w, req)

	// Assert
	testutils.AssertJSONContentTypeAndCode(t, w, http.StatusOK)

	response := testutils.GetJSONBody(t, w)
	assert.Contains(t, response, "modules")

	modules := response["modules"].([]interface{})
	assert.Len(t, modules, 1)

	module := modules[0].(map[string]interface{})
	assert.Equal(t, "test-namespace", module["namespace"])
	assert.Equal(t, "test-module", module["name"])
	assert.Equal(t, "aws", module["provider"])
}

// TestModuleListHandler_HandleListModules_Empty tests listing modules when no modules exist
func TestModuleListHandler_HandleListModules_Empty(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	// Create handler without test data
	namespaceRepository := moduleRepo.NewNamespaceRepository(db.DB)
	moduleProviderRepository := moduleRepo.NewModuleProviderRepository(db.DB, namespaceRepository, nil)
	listModulesQuery := moduleQuery.NewListModulesQuery(moduleProviderRepository)
	handler := v1.NewModuleListHandler(listModulesQuery)

	// Create request
	req := httptest.NewRequest("GET", "/v1/modules", nil)
	w := httptest.NewRecorder()

	// Act
	handler.HandleListModules(w, req)

	// Assert
	testutils.AssertJSONContentTypeAndCode(t, w, http.StatusOK)

	response := testutils.GetJSONBody(t, w)
	assert.Contains(t, response, "modules")

	modules := response["modules"].([]interface{})
	assert.Len(t, modules, 0)
}

// TestModuleListHandler_HandleListModules_MultipleModules tests listing multiple modules
func TestModuleListHandler_HandleListModules_MultipleModules(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	// Create test data with multiple modules
	namespace := testutils.CreateNamespace(t, db, "test-namespace")

	moduleProvider1 := testutils.CreateModuleProvider(t, db, namespace.ID, "module1", "aws")
	_ = testutils.CreatePublishedModuleVersion(t, db, moduleProvider1.ID, "1.0.0")

	moduleProvider2 := testutils.CreateModuleProvider(t, db, namespace.ID, "module2", "azurerm")
	_ = testutils.CreatePublishedModuleVersion(t, db, moduleProvider2.ID, "2.0.0")

	moduleProvider3 := testutils.CreateModuleProvider(t, db, namespace.ID, "module3", "gcp")
	_ = testutils.CreatePublishedModuleVersion(t, db, moduleProvider3.ID, "1.5.0")

	// Create handler
	namespaceRepository := moduleRepo.NewNamespaceRepository(db.DB)
	moduleProviderRepository := moduleRepo.NewModuleProviderRepository(db.DB, namespaceRepository, nil)
	listModulesQuery := moduleQuery.NewListModulesQuery(moduleProviderRepository)
	handler := v1.NewModuleListHandler(listModulesQuery)

	// Create request
	req := httptest.NewRequest("GET", "/v1/modules", nil)
	w := httptest.NewRecorder()

	// Act
	handler.HandleListModules(w, req)

	// Assert
	testutils.AssertJSONContentTypeAndCode(t, w, http.StatusOK)

	response := testutils.GetJSONBody(t, w)
	modules := response["modules"].([]interface{})
	assert.Len(t, modules, 3)
}

// TestModuleListHandler_HandleListModules_WithUnpublished tests listing modules with unpublished versions
func TestModuleListHandler_HandleListModules_WithUnpublished(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	// Create test data - module with unpublished version
	namespace := testutils.CreateNamespace(t, db, "test-namespace")
	moduleProvider := testutils.CreateModuleProvider(t, db, namespace.ID, "test-module", "aws")
	_ = testutils.CreateModuleVersion(t, db, moduleProvider.ID, "1.0.0") // Not published

	// Create handler
	namespaceRepository := moduleRepo.NewNamespaceRepository(db.DB)
	moduleProviderRepository := moduleRepo.NewModuleProviderRepository(db.DB, namespaceRepository, nil)
	listModulesQuery := moduleQuery.NewListModulesQuery(moduleProviderRepository)
	handler := v1.NewModuleListHandler(listModulesQuery)

	// Create request
	req := httptest.NewRequest("GET", "/v1/modules", nil)
	w := httptest.NewRecorder()

	// Act
	handler.HandleListModules(w, req)

	// Assert - modules with unpublished versions are included
	// (implementation checks for any version existence, not published status)
	testutils.AssertJSONContentTypeAndCode(t, w, http.StatusOK)

	response := testutils.GetJSONBody(t, w)
	modules := response["modules"].([]interface{})
	assert.Len(t, modules, 1, "Modules with unpublished versions should be included")

	module := modules[0].(map[string]interface{})
	assert.Equal(t, "test-namespace", module["namespace"])
	assert.Equal(t, "test-module", module["name"])
	assert.Equal(t, "aws", module["provider"])
}

// TestModuleListHandler_HandleListModules_Verified tests verified module flag
func TestModuleListHandler_HandleListModules_Verified(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	// Create test data
	namespace := testutils.CreateNamespace(t, db, "test-namespace")
	moduleProvider := testutils.CreateModuleProvider(t, db, namespace.ID, "test-module", "aws")
	_ = testutils.CreatePublishedModuleVersion(t, db, moduleProvider.ID, "1.0.0")

	// Update module provider to be verified
	err := db.DB.Model(&moduleProvider).Update("verified", true).Error
	require.NoError(t, err)

	// Create handler
	namespaceRepository := moduleRepo.NewNamespaceRepository(db.DB)
	moduleProviderRepository := moduleRepo.NewModuleProviderRepository(db.DB, namespaceRepository, nil)
	listModulesQuery := moduleQuery.NewListModulesQuery(moduleProviderRepository)
	handler := v1.NewModuleListHandler(listModulesQuery)

	// Create request
	req := httptest.NewRequest("GET", "/v1/modules", nil)
	w := httptest.NewRecorder()

	// Act
	handler.HandleListModules(w, req)

	// Assert
	response := testutils.GetJSONBody(t, w)
	modules := response["modules"].([]interface{})
	require.Len(t, modules, 1)

	module := modules[0].(map[string]interface{})
	assert.Equal(t, true, module["verified"], "Module should be marked as verified")
}
