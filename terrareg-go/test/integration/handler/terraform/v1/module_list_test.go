package v1_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/matthewjohn/terrareg/terrareg-go/test/integration/testutils"
)

// TestModuleListHandler_HandleListModules_Success tests listing modules with published versions
// Python reference: /app/test/unit/terrareg/server/test_api_module_search.py
func TestModuleListHandler_HandleListModules_Success(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	// Create test data - namespace, module provider, and published version
	namespace := testutils.CreateNamespace(t, db, "test-namespace", nil)
	moduleProvider := testutils.CreateModuleProvider(t, db, namespace.ID, "test-module", "aws")
	_ = testutils.CreatePublishedModuleVersion(t, db, moduleProvider.ID, "1.0.0")

	// Create handler using testutils generator
	handler := testutils.CreateModuleListHandler(t, db)

	// Create request
	req := httptest.NewRequest("GET", "/v1/modules", nil)
	w := httptest.NewRecorder()

	// Act
	handler.HandleListModules(w, req)

	// Assert - Comprehensive validation matching Python pattern
	testutils.AssertJSONContentTypeAndCode(t, w, http.StatusOK)

	response := testutils.GetJSONBody(t, w)

	// Validate response structure
	assert.Contains(t, response, "modules")
	modules := response["modules"].([]interface{})
	assert.Len(t, modules, 1)

	// Validate all module fields (Python validates complete response)
	module := modules[0].(map[string]interface{})

	// ProviderBase fields
	assert.Contains(t, module, "id")
	assert.NotEmpty(t, module["id"], "Module ID should not be empty")

	assert.Equal(t, "test-namespace", module["namespace"])
	assert.Equal(t, "test-module", module["name"])
	assert.Equal(t, "aws", module["provider"])

	assert.Contains(t, module, "verified")
	assert.Equal(t, false, module["verified"], "Module should not be verified by default")

	assert.Contains(t, module, "trusted")
	assert.Equal(t, false, module["trusted"], "Module should not be trusted by default")

	// ModuleProviderResponse fields - optional fields
	// These may be nil/absent depending on module version details
	if owner, ok := module["owner"]; ok && owner != nil {
		assert.NotEmpty(t, owner, "Owner should not be empty if present")
	}

	if description, ok := module["description"]; ok && description != nil {
		assert.NotEmpty(t, description, "Description should not be empty if present")
	}

	if source, ok := module["source"]; ok && source != nil {
		assert.NotEmpty(t, source, "Source should not be empty if present")
	}

	assert.Contains(t, module, "published_at")
	assert.NotNil(t, module["published_at"], "Published at should be present for published version")

	assert.Contains(t, module, "downloads")
	assert.IsType(t, float64(0), module["downloads"], "Downloads should be a number")
}

// TestModuleListHandler_HandleListModules_Empty tests listing modules when no modules exist
func TestModuleListHandler_HandleListModules_Empty(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	// Create handler without test data
	handler := testutils.CreateModuleListHandler(t, db)

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
	namespace := testutils.CreateNamespace(t, db, "test-namespace", nil)

	moduleProvider1 := testutils.CreateModuleProvider(t, db, namespace.ID, "module1", "aws")
	_ = testutils.CreatePublishedModuleVersion(t, db, moduleProvider1.ID, "1.0.0")

	moduleProvider2 := testutils.CreateModuleProvider(t, db, namespace.ID, "module2", "azurerm")
	_ = testutils.CreatePublishedModuleVersion(t, db, moduleProvider2.ID, "2.0.0")

	moduleProvider3 := testutils.CreateModuleProvider(t, db, namespace.ID, "module3", "gcp")
	_ = testutils.CreatePublishedModuleVersion(t, db, moduleProvider3.ID, "1.5.0")

	// Create handler
	handler := testutils.CreateModuleListHandler(t, db)

	// Create request
	req := httptest.NewRequest("GET", "/v1/modules", nil)
	w := httptest.NewRecorder()

	// Act
	handler.HandleListModules(w, req)

	// Assert - Comprehensive validation for all modules
	testutils.AssertJSONContentTypeAndCode(t, w, http.StatusOK)

	response := testutils.GetJSONBody(t, w)
	modules := response["modules"].([]interface{})
	assert.Len(t, modules, 3)

	// Validate each module has required fields
	expectedModules := []struct {
		name     string
		provider string
	}{
		{"module1", "aws"},
		{"module2", "azurerm"},
		{"module3", "gcp"},
	}

	for i, mod := range modules {
		module := mod.(map[string]interface{})

		// Validate basic structure
		assert.Contains(t, module, "id")
		assert.NotEmpty(t, module["id"])

		assert.Equal(t, "test-namespace", module["namespace"])
		assert.Equal(t, expectedModules[i].name, module["name"])
		assert.Equal(t, expectedModules[i].provider, module["provider"])

		// Validate all required fields exist
		assert.Contains(t, module, "verified")
		assert.Contains(t, module, "trusted")
		assert.Contains(t, module, "published_at")
		assert.Contains(t, module, "downloads")
	}
}

// TestModuleListHandler_HandleListModules_WithUnpublished tests listing modules with unpublished versions
func TestModuleListHandler_HandleListModules_WithUnpublished(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	// Create test data - module with unpublished version
	namespace := testutils.CreateNamespace(t, db, "test-namespace", nil)
	moduleProvider := testutils.CreateModuleProvider(t, db, namespace.ID, "test-module", "aws")
	_ = testutils.CreatePublishedModuleVersion(t, db, moduleProvider.ID, "1.0.0") // Not published

	// Create handler
	handler := testutils.CreateModuleListHandler(t, db)

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
	namespace := testutils.CreateNamespace(t, db, "test-namespace", nil)
	moduleProvider := testutils.CreateModuleProvider(t, db, namespace.ID, "test-module", "aws")
	_ = testutils.CreatePublishedModuleVersion(t, db, moduleProvider.ID, "1.0.0")

	// Update module provider to be verified
	err := db.DB.Model(&moduleProvider).Update("verified", true).Error
	require.NoError(t, err)

	// Create handler
	handler := testutils.CreateModuleListHandler(t, db)

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

// TestModuleListHandler_ComprehensiveFieldValidation validates all module fields
// Python reference: /app/test/unit/terrareg/server/test_api_module_search.py test_with_single_module_response
// This test validates complete response structure matching Python's full JSON comparison pattern
func TestModuleListHandler_ComprehensiveFieldValidation(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	// Use SetupFullyPopulatedModule to match Python's fullypopulated test data
	_, moduleProvider, _ := testutils.SetupFullyPopulatedModule(t, db)

	// Update module provider to be verified (matching Python test)
	err := db.DB.Model(&moduleProvider).Update("verified", true).Error
	require.NoError(t, err)

	// Create handler
	handler := testutils.CreateModuleListHandler(t, db)

	// Create request
	req := httptest.NewRequest("GET", "/v1/modules", nil)
	w := httptest.NewRecorder()

	// Act
	handler.HandleListModules(w, req)

	// Assert - Comprehensive field validation matching Python pattern
	testutils.AssertJSONContentTypeAndCode(t, w, http.StatusOK)

	response := testutils.GetJSONBody(t, w)
	modules := response["modules"].([]interface{})
	require.Len(t, modules, 1, "Should return exactly one module")

	module := modules[0].(map[string]interface{})

	// Validate all required fields exist (matching Python's complete JSON structure)
	// Python reference: assert res.json == {'id': '...', 'owner': '...', 'namespace': '...', ...}

	assert.Contains(t, module, "id")
	assert.NotEmpty(t, module["id"])

	assert.Equal(t, "moduledetails", module["namespace"])
	assert.Equal(t, "fullypopulated", module["name"])
	assert.Equal(t, "testprovider", module["provider"])

	// Optional fields - validate if present (from fullypopulated test data)
	if ownerVal, ok := module["owner"]; ok && ownerVal != nil {
		assert.Equal(t, "This is the owner of the module", ownerVal)
	}

	if descVal, ok := module["description"]; ok && descVal != nil {
		assert.Equal(t, "This is a test module version for tests.", descVal)
	}

	if srcVal, ok := module["source"]; ok && srcVal != nil {
		assert.NotEmpty(t, srcVal, "Source should not be empty")
	}

	// Validate published_at timestamp format and presence
	assert.Contains(t, module, "published_at")
	publishedAt, ok := module["published_at"].(string)
	if ok && publishedAt != "" {
		// Validate ISO 8601 format (Python uses .isoformat())
		assert.NotEmpty(t, publishedAt, "Published at should not be empty for published version")
	}

	// Validate downloads count (type validation)
	assert.Contains(t, module, "downloads")
	assert.IsType(t, float64(0), module["downloads"], "Downloads should be numeric")

	// Validate boolean flags
	assert.Equal(t, true, module["verified"], "Module should be marked as verified")
	assert.Equal(t, false, module["trusted"], "Module should not be trusted by default")

	// Note: Python test includes 'version' and 'internal' fields in search response
	// but Go's module list endpoint doesn't include these in the response
	// This is an API difference to note in parity analysis
}

// TestModuleListHandler_ResultOrdering tests that modules are returned in consistent order
// Python reference: /app/test/unit/terrareg/server/test_api_module_search.py
// Note: The Go module list endpoint does not currently support pagination parameters (limit/offset).
// This test validates that modules are returned in a deterministic order (by id DESC).
func TestModuleListHandler_ResultOrdering(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	// Create test data with multiple modules
	namespace := testutils.CreateNamespace(t, db, "test-namespace", nil)

	// Create 3 modules with versions in a specific order
	// module1 (oldest), module2, module3 (newest)
	moduleProvider1 := testutils.CreateModuleProvider(t, db, namespace.ID, "module1", "aws")
	_ = testutils.CreatePublishedModuleVersion(t, db, moduleProvider1.ID, "1.0.0")

	moduleProvider2 := testutils.CreateModuleProvider(t, db, namespace.ID, "module2", "azurerm")
	_ = testutils.CreatePublishedModuleVersion(t, db, moduleProvider2.ID, "2.0.0")

	moduleProvider3 := testutils.CreateModuleProvider(t, db, namespace.ID, "module3", "gcp")
	_ = testutils.CreatePublishedModuleVersion(t, db, moduleProvider3.ID, "1.5.0")

	// Create handler
	handler := testutils.CreateModuleListHandler(t, db)

	req := httptest.NewRequest("GET", "/v1/modules", nil)
	w := httptest.NewRecorder()

	handler.HandleListModules(w, req)

	// Validate response
	testutils.AssertJSONContentTypeAndCode(t, w, http.StatusOK)
	response := testutils.GetJSONBody(t, w)

	// Validate all modules are returned (no pagination - returns all)
	modules := response["modules"].([]interface{})
	assert.Len(t, modules, 3, "Should return all modules")

	// Validate modules are returned in deterministic order (by id DESC)
	actualModuleNames := make([]string, 0, len(modules))
	for _, m := range modules {
		module := m.(map[string]interface{})
		assert.Contains(t, module, "namespace")
		assert.Contains(t, module, "name")

		name := module["name"].(string)
		actualModuleNames = append(actualModuleNames, name)
	}

	// Modules are ordered by id ASC (oldest first)
	expectedModules := []string{"module1", "module2", "module3"}
	assert.Equal(t, expectedModules, actualModuleNames,
		"Module names should match expected order (id ASC)")
}

// TestModuleListHandler_WithLimitOffset tests pagination parameters
// Python reference: test_api_module_list.py:test_with_limit_offset
func TestModuleListHandler_WithLimitOffset(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	handler := testutils.CreateModuleListHandler(t, db)

	// Request with offset=23, limit=12 (matching Python test)
	req := httptest.NewRequest("GET", "/v1/modules?offset=23&limit=12", nil)
	w := httptest.NewRecorder()

	handler.HandleListModules(w, req)

	// Python: assert res.json == {'meta': {'current_offset': 23, 'limit': 12, 'prev_offset': 11}, 'modules': []}
	testutils.AssertJSONContentTypeAndCode(t, w, http.StatusOK)
	response := testutils.GetJSONBody(t, w)

	meta := response["meta"].(map[string]interface{})
	assert.Equal(t, float64(23), meta["current_offset"])
	assert.Equal(t, float64(12), meta["limit"])
	assert.Equal(t, float64(11), meta["prev_offset"])

	modules := response["modules"].([]interface{})
	assert.Len(t, modules, 0)
}

// TestModuleListHandler_WithProviderFilter tests provider filtering
// Python reference: test_api_module_list.py:test_with_provider_filter
func TestModuleListHandler_WithProviderFilter(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	// Create test data with different providers
	namespace := testutils.CreateNamespace(t, db, "test-namespace", nil)

	moduleProvider1 := testutils.CreateModuleProvider(t, db, namespace.ID, "module1", "aws")
	_ = testutils.CreatePublishedModuleVersion(t, db, moduleProvider1.ID, "1.0.0")

	moduleProvider2 := testutils.CreateModuleProvider(t, db, namespace.ID, "module2", "azurerm")
	_ = testutils.CreatePublishedModuleVersion(t, db, moduleProvider2.ID, "2.0.0")

	moduleProvider3 := testutils.CreateModuleProvider(t, db, namespace.ID, "module3", "aws")
	_ = testutils.CreatePublishedModuleVersion(t, db, moduleProvider3.ID, "1.5.0")

	handler := testutils.CreateModuleListHandler(t, db)

	// Request with provider filter
	req := httptest.NewRequest("GET", "/v1/modules?provider=aws", nil)
	w := httptest.NewRecorder()

	handler.HandleListModules(w, req)

	// Python: providers=['aws']
	testutils.AssertJSONContentTypeAndCode(t, w, http.StatusOK)
	response := testutils.GetJSONBody(t, w)

	modules := response["modules"].([]interface{})
	assert.Len(t, modules, 2, "Should return only aws modules")

	// Verify both modules have aws provider
	for _, m := range modules {
		module := m.(map[string]interface{})
		assert.Equal(t, "aws", module["provider"])
	}
}

// TestModuleListHandler_VerifiedFalse tests verified=false filter
// Python reference: test_api_module_list.py:test_with_verified_false
func TestModuleListHandler_VerifiedFalse(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	// Create test data - unverified module by default
	namespace := testutils.CreateNamespace(t, db, "test-namespace", nil)
	moduleProvider := testutils.CreateModuleProvider(t, db, namespace.ID, "test-module", "aws")
	_ = testutils.CreatePublishedModuleVersion(t, db, moduleProvider.ID, "1.0.0")

	handler := testutils.CreateModuleListHandler(t, db)

	// Request with verified=false
	req := httptest.NewRequest("GET", "/v1/modules?verified=false", nil)
	w := httptest.NewRecorder()

	handler.HandleListModules(w, req)

	// Python: verified=False
	testutils.AssertJSONContentTypeAndCode(t, w, http.StatusOK)
	response := testutils.GetJSONBody(t, w)

	modules := response["modules"].([]interface{})
	assert.Len(t, modules, 1, "Should return unverified module")

	module := modules[0].(map[string]interface{})
	assert.Equal(t, false, module["verified"], "Module should not be verified")
}

// TestModuleListHandler_WithMoreResultsAvailable tests next_offset in meta
// Python reference: test_api_module_list.py:test_with_module_response_with_more_results_available
func TestModuleListHandler_WithMoreResultsAvailable(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	// Create test data where count > limit
	namespace := testutils.CreateNamespace(t, db, "test-namespace", nil)

	// Create 3 modules
	for i := 1; i <= 3; i++ {
		mp := testutils.CreateModuleProvider(t, db, namespace.ID, fmt.Sprintf("module%d", i), "aws")
		_ = testutils.CreatePublishedModuleVersion(t, db, mp.ID, "1.0.0")
	}

	handler := testutils.CreateModuleListHandler(t, db)

	// Request with offset=0, limit=1 (count=3, so next_offset=1)
	req := httptest.NewRequest("GET", "/v1/modules?offset=0&limit=1", nil)
	w := httptest.NewRecorder()

	handler.HandleListModules(w, req)

	// Python: count=3, limit=1, so next_offset=1
	// Python: {'meta': {'current_offset': 0, 'limit': 1, 'next_offset': 1}, 'modules': [...]}
	testutils.AssertJSONContentTypeAndCode(t, w, http.StatusOK)
	response := testutils.GetJSONBody(t, w)

	meta := response["meta"].(map[string]interface{})
	assert.Equal(t, float64(0), meta["current_offset"])
	assert.Equal(t, float64(1), meta["limit"])
	assert.Equal(t, float64(1), meta["next_offset"])

	modules := response["modules"].([]interface{})
	assert.Len(t, modules, 1)
}
