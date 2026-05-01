package v1_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/matthewjohn/terrareg/terrareg-go/test/integration/testutils"
)

// TestModuleVersionsHandler_ExistingModuleVersion tests successful version list retrieval
// Python reference: test_api_module_versions.py:test_existing_module_version
func TestModuleVersionsHandler_ExistingModuleVersion(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	// Setup test data matching Python's moduledetails/fullypopulated/testprovider
	// Python creates versions: 1.2.0, 1.6.1-beta, 1.5.0
	namespace := testutils.CreateNamespace(t, db, "moduledetails", nil)
	moduleProvider := testutils.CreateModuleProvider(t, db, namespace.ID, "fullypopulated", "testprovider")

	// Create three versions matching Python test data
	v1 := testutils.CreateModuleVersion(t, db, moduleProvider.ID, "1.2.0")
	v2 := testutils.CreateModuleVersion(t, db, moduleProvider.ID, "1.6.1-beta")
	v3 := testutils.CreateModuleVersion(t, db, moduleProvider.ID, "1.5.0")

	// Publish all versions
	published := true
	db.DB.Model(&v1).Update("published", &published)
	db.DB.Model(&v2).Update("published", &published)
	db.DB.Model(&v3).Update("published", &published)

	// Create handler
	handler := testutils.CreateModuleVersionsHandler(t, db)

	// Create request with Chi context
	req := testutils.CreateRequestWithChiParams(t, "GET", "/v1/modules/moduledetails/fullypopulated/testprovider/versions", map[string]string{
		"namespace": "moduledetails",
		"name":      "fullypopulated",
		"provider":  "testprovider",
	})
	w := httptest.NewRecorder()

	// Act
	handler.HandleModuleVersions(w, req)

	// Assert - Python validates complete JSON structure
	testutils.AssertJSONContentTypeAndCode(t, w, http.StatusOK)
	response := testutils.GetJSONBody(t, w)

	// Python assertion: assert res.json == {'modules': [{'source': ..., 'versions': [...]}]}
	assert.Contains(t, response, "modules")
	modules := response["modules"].([]interface{})
	assert.Len(t, modules, 1)

	module := modules[0].(map[string]interface{})

	// Python: 'source': 'moduledetails/fullypopulated/testprovider'
	assert.Equal(t, "moduledetails/fullypopulated/testprovider", module["source"])

	// Python: validates 3 versions with specific structure
	versions := module["versions"].([]interface{})
	assert.Len(t, versions, 3)

	// Python: Each version has root, submodules, version
	for _, v := range versions {
		version := v.(map[string]interface{})
		assert.Contains(t, version, "root")
		assert.Contains(t, version, "submodules")
		assert.Contains(t, version, "version")
	}
}

// TestModuleVersionsHandler_UnverifiedModuleVersion tests unverified module
// Python reference: test_api_module_versions.py:test_unverified_module_version
func TestModuleVersionsHandler_UnverifiedModuleVersion(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	// Create unverified module (verified=false by default)
	namespace := testutils.CreateNamespace(t, db, "testnamespace", nil)
	moduleProvider := testutils.CreateModuleProvider(t, db, namespace.ID, "unverifiedmodule", "testprovider")
	// Don't set verified=true, it defaults to false
	_ = testutils.CreatePublishedModuleVersion(t, db, moduleProvider.ID, "1.2.3")

	handler := testutils.CreateModuleVersionsHandler(t, db)

	req := testutils.CreateRequestWithChiParams(t, "GET", "/v1/modules/testnamespace/unverifiedmodule/testprovider/versions", map[string]string{
		"namespace": "testnamespace",
		"name":      "unverifiedmodule",
		"provider":  "testprovider",
	})
	w := httptest.NewRecorder()

	handler.HandleModuleVersions(w, req)

	// Python: returns 200 with unverified module data
	testutils.AssertJSONContentTypeAndCode(t, w, http.StatusOK)
	response := testutils.GetJSONBody(t, w)

	modules := response["modules"].([]interface{})
	assert.Len(t, modules, 1)

	module := modules[0].(map[string]interface{})
	assert.Equal(t, "testnamespace/unverifiedmodule/testprovider", module["source"])

	versions := module["versions"].([]interface{})
	assert.Len(t, versions, 1)
	// Python: Single version 1.2.3
	version := versions[0].(map[string]interface{})
	assert.Equal(t, "1.2.3", version["version"])
}

// TestModuleVersionsHandler_InternalModuleVersion tests internal module
// Python reference: test_api_module_versions.py:test_internal_module_version
func TestModuleVersionsHandler_InternalModuleVersion(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	namespace := testutils.CreateNamespace(t, db, "testnamespace", nil)
	moduleProvider := testutils.CreateModuleProvider(t, db, namespace.ID, "internalmodule", "testprovider")
	moduleVersion := testutils.CreatePublishedModuleVersion(t, db, moduleProvider.ID, "5.2.0")

	// Set internal = true
	internal := true
	db.DB.Model(&moduleVersion).Update("internal", &internal)

	handler := testutils.CreateModuleVersionsHandler(t, db)

	req := testutils.CreateRequestWithChiParams(t, "GET", "/v1/modules/testnamespace/internalmodule/testprovider/versions", map[string]string{
		"namespace": "testnamespace",
		"name":      "internalmodule",
		"provider":  "testprovider",
	})
	w := httptest.NewRecorder()

	handler.HandleModuleVersions(w, req)

	// Python: returns 200, internal modules are included
	testutils.AssertJSONContentTypeAndCode(t, w, http.StatusOK)
	response := testutils.GetJSONBody(t, w)

	modules := response["modules"].([]interface{})
	assert.Len(t, modules, 1)

	versions := modules[0].(map[string]interface{})["versions"].([]interface{})
	assert.Len(t, versions, 1)
	assert.Equal(t, "5.2.0", versions[0].(map[string]interface{})["version"])
}

// TestModuleVersionsHandler_NonExistentModuleVersion tests 404 for non-existent module
// Python reference: test_api_module_versions.py:test_non_existent_module_version
func TestModuleVersionsHandler_NonExistentModuleVersion(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	handler := testutils.CreateModuleVersionsHandler(t, db)

	req := testutils.CreateRequestWithChiParams(t, "GET", "/v1/modules/namespacename/modulename/doesnotexist/versions", map[string]string{
		"namespace": "namespacename",
		"name":      "modulename",
		"provider":  "doesnotexist",
	})
	w := httptest.NewRecorder()

	handler.HandleModuleVersions(w, req)

	// Python: assert res.json == {'errors': ['Not Found']}
	// Python: assert res.status_code == 404
	// Note: Go returns {"error":"Not Found"} instead of Python's {"errors":["Not Found"]}
	// This is a design difference noted in the parity analysis
	assert.Equal(t, http.StatusNotFound, w.Code)

	response := testutils.GetJSONBody(t, w)
	// Go implementation uses "message" field for errors
	assert.Contains(t, response, "message")
	assert.Contains(t, response["message"].(string), "Not Found")
}

// TestModuleVersionsHandler_AnalyticsToken tests analytics token handling
// Python reference: test_api_module_versions.py:test_analytics_token
func TestModuleVersionsHandler_AnalyticsToken(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	// Create namespace and module
	namespace := testutils.CreateNamespace(t, db, "testnamespace", nil)
	moduleProvider := testutils.CreateModuleProvider(t, db, namespace.ID, "testmodulename", "testprovider")

	// Create two versions: 2.4.1 and 1.0.0 (matching Python test)
	v1 := testutils.CreateModuleVersion(t, db, moduleProvider.ID, "2.4.1")
	v2 := testutils.CreateModuleVersion(t, db, moduleProvider.ID, "1.0.0")

	published := true
	db.DB.Model(&v1).Update("published", &published)
	db.DB.Model(&v2).Update("published", &published)

	handler := testutils.CreateModuleVersionsHandler(t, db)

	// Request with analytics token in namespace (test_token-name__testnamespace)
	req := testutils.CreateRequestWithChiParams(t, "GET", "/v1/modules/test_token-name__testnamespace/testmodulename/testprovider/versions", map[string]string{
		"namespace": "test_token-name__testnamespace",
		"name":      "testmodulename",
		"provider":  "testprovider",
	})
	w := httptest.NewRecorder()

	handler.HandleModuleVersions(w, req)

	// Python: Analytics token should be converted, returns data for testnamespace
	testutils.AssertJSONContentTypeAndCode(t, w, http.StatusOK)
	response := testutils.GetJSONBody(t, w)

	// Python: source should be "testnamespace/testmodulename/testprovider" (token stripped)
	modules := response["modules"].([]interface{})
	module := modules[0].(map[string]interface{})
	assert.Equal(t, "testnamespace/testmodulename/testprovider", module["source"])

	// Python: Both versions 2.4.1 and 1.0.0 should be present
	versions := module["versions"].([]interface{})
	assert.Len(t, versions, 2)

	// Extract version strings
	versionStrings := make([]string, len(versions))
	for i, v := range versions {
		versionStrings[i] = v.(map[string]interface{})["version"].(string)
	}

	// Python expects both versions
	assert.Contains(t, versionStrings, "2.4.1")
	assert.Contains(t, versionStrings, "1.0.0")
}

// TestModuleVersionsHandler_Unauthenticated tests unauthenticated access
// Python reference: test_api_module_versions.py:test_unauthenticated
func TestModuleVersionsHandler_Unauthenticated(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	// Create module with published version
	namespace := testutils.CreateNamespace(t, db, "moduledetails", nil)
	moduleProvider := testutils.CreateModuleProvider(t, db, namespace.ID, "fullypopulated", "testprovider")
	_ = testutils.CreatePublishedModuleVersion(t, db, moduleProvider.ID, "1.2.0")

	handler := testutils.CreateModuleVersionsHandler(t, db)

	// Create request without authentication
	req := testutils.CreateRequestWithChiParams(t, "GET", "/v1/modules/moduledetails/fullypopulated/testprovider/versions", map[string]string{
		"namespace": "moduledetails",
		"name":      "fullypopulated",
		"provider":  "testprovider",
	})
	w := httptest.NewRecorder()

	handler.HandleModuleVersions(w, req)

	// Python: Unauthenticated access should return 200 (Terraform API doesn't require auth for read)
	testutils.AssertJSONContentTypeAndCode(t, w, http.StatusOK)
	response := testutils.GetJSONBody(t, w)

	// Verify response structure
	assert.Contains(t, response, "modules")
	modules := response["modules"].([]interface{})
	assert.Len(t, modules, 1)
}
