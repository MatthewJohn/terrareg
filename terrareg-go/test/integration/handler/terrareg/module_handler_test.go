// Package terrareg_test provides integration tests for the terrareg HTTP handlers
package terrareg_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	moduleQuery "github.com/matthewjohn/terrareg/terrareg-go/internal/application/query/module"
	namespaceService "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/module/service"
	analyticsRepo "github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/persistence/sqldb/analytics"
	moduleRepo "github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/persistence/sqldb/module"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/interfaces/http/handler/terrareg"
	"github.com/matthewjohn/terrareg/terrareg-go/test/integration/testutils"
	sqldb "github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/persistence/sqldb"
)

// TestModuleHandler_HandleModuleList_Success tests the module list endpoint
func TestModuleHandler_HandleModuleList_Success(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	// Create test data
	namespace := testutils.CreateNamespace(t, db, "testnamespace", nil)
	moduleProvider := testutils.CreateModuleProvider(t, db, namespace.ID, "testmodule", "aws")
	moduleVersion := testutils.CreateModuleVersion(t, db, moduleProvider.ID, "1.0.0")

	// Update version to published
	published := true
	moduleVersion.Published = &published
	db.DB.Save(&moduleVersion)

	// Set this version as the latest version for the module provider
	// This is required for the search query to find the module
	err := db.DB.Model(&sqldb.ModuleProviderDB{}).
		Where("id = ?", moduleProvider.ID).
		Update("latest_version_id", moduleVersion.ID).Error
	require.NoError(t, err)

	// Create handler
	namespaceRepository := moduleRepo.NewNamespaceRepository(db.DB)
	domainConfig := testutils.CreateTestDomainConfig(t)
	moduleProviderRepository, err := moduleRepo.NewModuleProviderRepository(db.DB, namespaceRepository, domainConfig)
	require.NoError(t, err)
	listModulesQuery := moduleQuery.NewListModulesQuery(moduleProviderRepository)
	namespaceSvc := namespaceService.NewNamespaceService(domainConfig)
	analyticsRepository, err := analyticsRepo.NewAnalyticsRepository(db.DB, namespaceRepository, namespaceSvc)
	require.NoError(t, err)

	handler := terrareg.NewModuleReadHandlerForTesting(
		listModulesQuery,
		nil, // search not used
		nil, // get provider not used
		nil, // list providers not used
		analyticsRepository,
		nil, // versionPresenter not needed
	)

	// Create request
	req := httptest.NewRequest("GET", "/v1/modules", nil)
	w := httptest.NewRecorder()

	// Act
	handler.HandleModuleList(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Contains(t, response, "modules")
	modules := response["modules"].([]interface{})
	assert.GreaterOrEqual(t, len(modules), 1)

	// Check that our test module is in the list
	found := false
	for _, m := range modules {
		module := m.(map[string]interface{})
		if module["namespace"] == "testnamespace" && module["name"] == "testmodule" {
			found = true
			assert.Equal(t, "aws", module["provider"])
			break
		}
	}
	assert.True(t, found, "test module not found in response")
}

// TestModuleHandler_HandleModuleList_Empty tests module list with empty results
func TestModuleHandler_HandleModuleList_Empty(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	// Don't create any test data

	// Create handler
	namespaceRepository := moduleRepo.NewNamespaceRepository(db.DB)
	domainConfig := testutils.CreateTestDomainConfig(t)
	moduleProviderRepository, err := moduleRepo.NewModuleProviderRepository(db.DB, namespaceRepository, domainConfig)
	require.NoError(t, err)
	listModulesQuery := moduleQuery.NewListModulesQuery(moduleProviderRepository)
	namespaceSvc := namespaceService.NewNamespaceService(domainConfig)
	analyticsRepository, err := analyticsRepo.NewAnalyticsRepository(db.DB, namespaceRepository, namespaceSvc)
	require.NoError(t, err)

	handler := terrareg.NewModuleReadHandlerForTesting(
		listModulesQuery,
		nil,
		nil,
		nil,
		analyticsRepository,
		nil, // versionPresenter not needed
	)

	// Create request
	req := httptest.NewRequest("GET", "/v1/modules", nil)
	w := httptest.NewRecorder()

	// Act
	handler.HandleModuleList(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	modules := response["modules"].([]interface{})
	assert.Len(t, modules, 0)
}

// TestModuleHandler_HandleNamespaceModules_Success tests the namespace modules endpoint
func TestModuleHandler_HandleNamespaceModules_Success(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	// Create test data
	namespace := testutils.CreateNamespace(t, db, "mynamespace", nil)
	moduleProvider := testutils.CreateModuleProvider(t, db, namespace.ID, "mymodule", "aws")
	moduleVersion := testutils.CreateModuleVersion(t, db, moduleProvider.ID, "1.0.0")
	// Update version to published
	published := true
	moduleVersion.Published = &published
	db.DB.Save(&moduleVersion)

		// Set this version as the latest version for the module provider
		err := db.DB.Model(&sqldb.ModuleProviderDB{}).
			Where("id = ?", moduleProvider.ID).
			Update("latest_version_id", moduleVersion.ID).Error
		require.NoError(t, err)

	// Create handler
	namespaceRepository := moduleRepo.NewNamespaceRepository(db.DB)
	domainConfig := testutils.CreateTestDomainConfig(t)
	moduleProviderRepository, err := moduleRepo.NewModuleProviderRepository(db.DB, namespaceRepository, domainConfig)
	require.NoError(t, err)
	listModulesQuery := moduleQuery.NewListModulesQuery(moduleProviderRepository)
	namespaceSvc := namespaceService.NewNamespaceService(domainConfig)
	analyticsRepository, err := analyticsRepo.NewAnalyticsRepository(db.DB, namespaceRepository, namespaceSvc)
	require.NoError(t, err)

	handler := terrareg.NewModuleReadHandlerForTesting(
		listModulesQuery,
		nil,
		nil,
		nil,
		analyticsRepository,
		nil, // versionPresenter not needed
	)

	// Create request with chi context
	req := httptest.NewRequest("GET", "/v1/modules/mynamespace", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("namespace", "mynamespace")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()

	// Act
	handler.HandleNamespaceModules(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Contains(t, response, "modules")
	modules := response["modules"].([]interface{})
	assert.GreaterOrEqual(t, len(modules), 1)

	// Verify all modules are from the requested namespace
	for _, m := range modules {
		module := m.(map[string]interface{})
		assert.Equal(t, "mynamespace", module["namespace"])
	}
}

// TestModuleHandler_HandleNamespaceModules_NotFound tests namespace modules with non-existent namespace
func TestModuleHandler_HandleNamespaceModules_NotFound(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	// Create handler (no test data)
	namespaceRepository := moduleRepo.NewNamespaceRepository(db.DB)
	domainConfig := testutils.CreateTestDomainConfig(t)
	moduleProviderRepository, err := moduleRepo.NewModuleProviderRepository(db.DB, namespaceRepository, domainConfig)
	require.NoError(t, err)
	listModulesQuery := moduleQuery.NewListModulesQuery(moduleProviderRepository)
	namespaceSvc := namespaceService.NewNamespaceService(domainConfig)
	analyticsRepository, err := analyticsRepo.NewAnalyticsRepository(db.DB, namespaceRepository, namespaceSvc)
	require.NoError(t, err)

	handler := terrareg.NewModuleReadHandlerForTesting(
		listModulesQuery,
		nil,
		nil,
		nil,
		analyticsRepository,
		nil, // versionPresenter not needed
	)

	// Create request with chi context
	req := httptest.NewRequest("GET", "/v1/modules/nonexistent", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("namespace", "nonexistent")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()

	// Act
	handler.HandleNamespaceModules(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	modules := response["modules"].([]interface{})
	assert.Len(t, modules, 0)
}

// TestModuleHandler_HandleModuleDetails_Success tests the module details endpoint
// Python reference: /app/test/unit/terrareg/server/test_api_module_details.py - test_existing_module
func TestModuleHandler_HandleModuleDetails_Success(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	// Create test data
	namespace := testutils.CreateNamespace(t, db, "testns", nil)
	moduleProvider1 := testutils.CreateModuleProvider(t, db, namespace.ID, "testmodule", "aws")
	_ = testutils.CreatePublishedModuleVersion(t, db, moduleProvider1.ID, "1.0.0")
	moduleProvider2 := testutils.CreateModuleProvider(t, db, namespace.ID, "testmodule", "azure")
	_ = testutils.CreatePublishedModuleVersion(t, db, moduleProvider2.ID, "2.0.0")

	// Create handler
	namespaceRepository := moduleRepo.NewNamespaceRepository(db.DB)
	domainConfig := testutils.CreateTestDomainConfig(t)
	moduleProviderRepository, err := moduleRepo.NewModuleProviderRepository(db.DB, namespaceRepository, domainConfig)
	require.NoError(t, err)
	listModuleProvidersQuery := moduleQuery.NewListModuleProvidersQuery(moduleProviderRepository)
	namespaceSvc := namespaceService.NewNamespaceService(domainConfig)
	analyticsRepository, err := analyticsRepo.NewAnalyticsRepository(db.DB, namespaceRepository, namespaceSvc)
	require.NoError(t, err)

	handler := terrareg.NewModuleReadHandlerForTesting(
		nil, // list not used
		nil, // search not used
		nil, // get provider not used
		listModuleProvidersQuery,
		analyticsRepository,
		nil, // versionPresenter not needed
	)

	// Create request with chi context
	req := httptest.NewRequest("GET", "/v1/modules/testns/testmodule", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("namespace", "testns")
	rctx.URLParams.Add("name", "testmodule")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()

	// Act
	handler.HandleModuleDetails(w, req)

	// Assert - Comprehensive validation matching Python pattern
	// Python reference: assert res.json == {'meta': {'limit': 10, 'current_offset': 0}, 'modules': [...]}
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	// Validate meta structure (Python validates pagination metadata)
	assert.Contains(t, response, "meta")
	meta := response["meta"].(map[string]interface{})
	assert.Contains(t, meta, "limit")
	assert.Contains(t, meta, "current_offset")

	// Validate modules array exists and has expected providers
	assert.Contains(t, response, "modules")
	modules := response["modules"].([]interface{})
	assert.Len(t, modules, 2, "Should return exactly two module providers (aws and azure)")

	// Validate all module fields (Python validates complete response)
	// Python reference: {'id': 'testnamespace/lonelymodule/testprovider/1.0.0', 'owner': 'Mock Owner', ...}
	expectedProviders := []string{"aws", "azure"}
	for i, m := range modules {
		module := m.(map[string]interface{})

		// Validate all required fields exist (matching Python's complete JSON structure)
		assert.Contains(t, module, "id")
		assert.NotEmpty(t, module["id"], "Module ID should not be empty")

		assert.Equal(t, "testns", module["namespace"])
		assert.Equal(t, "testmodule", module["name"])
		assert.Equal(t, expectedProviders[i], module["provider"])

		assert.Contains(t, module, "verified")
		assert.IsType(t, false, module["verified"], "Verified should be a boolean")

		assert.Contains(t, module, "trusted")
		assert.IsType(t, false, module["trusted"], "Trusted should be a boolean")

		// Optional fields - validate if present
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
}

// TestModuleHandler_HandleModuleProviderDetails_Success tests the module provider details endpoint
// Python reference: /app/test/unit/terrareg/server/test_api_terrareg_module_provider_details.py - test_existing_module_provider_no_custom_urls
func TestModuleHandler_HandleModuleProviderDetails_Success(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	// Create test data
	namespace := testutils.CreateNamespace(t, db, "hashicorp", nil)
	moduleProvider := testutils.CreateModuleProvider(t, db, namespace.ID, "consul", "aws")
	_ = testutils.CreatePublishedModuleVersion(t, db, moduleProvider.ID, "1.0.0")

	// Create handler - use CreateTerraregModuleDetailsHandler for full response with internal, root, etc.
	handler := testutils.CreateTerraregModuleDetailsHandler(t, db)

	// Create request with chi context
	req := httptest.NewRequest("GET", "/v1/modules/hashicorp/consul/aws", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("namespace", "hashicorp")
	rctx.URLParams.Add("name", "consul")
	rctx.URLParams.Add("provider", "aws")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()

	// Act - Use HandleTerraregModuleProviderDetails for full response with internal, root, etc.
	handler.HandleTerraregModuleProviderDetails(w, req)

	// Assert - Comprehensive validation matching Python pattern
	// Python reference: assert res.json == {'id': 'testnamespace/lonelymodule/testprovider/1.0.0', 'owner': 'Mock Owner', ...}
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	// Validate all required top-level fields (Python validates complete response)
	assert.Contains(t, response, "id")
	assert.NotEmpty(t, response["id"], "Module provider ID should not be empty")

	assert.Equal(t, "hashicorp", response["namespace"])
	assert.Equal(t, "consul", response["name"])
	assert.Equal(t, "aws", response["provider"])

	assert.Contains(t, response, "verified")
	assert.IsType(t, false, response["verified"], "Verified should be a boolean")

	assert.Contains(t, response, "trusted")
	assert.IsType(t, false, response["trusted"], "Trusted should be a boolean")

	assert.Contains(t, response, "internal")
	assert.IsType(t, false, response["internal"], "Internal should be a boolean")

	// Validate root object structure (Python validates root specs)
	assert.Contains(t, response, "root")
	root := response["root"].(map[string]interface{})
	assert.Contains(t, root, "path")
	assert.Contains(t, root, "readme")
	assert.Contains(t, root, "empty")
	assert.IsType(t, false, root["empty"], "Empty should be a boolean")

	// Validate root arrays (Python validates these are arrays)
	assert.Contains(t, root, "inputs")
	assert.IsType(t, []interface{}{}, root["inputs"], "Inputs should be an array")

	assert.Contains(t, root, "outputs")
	assert.IsType(t, []interface{}{}, root["outputs"], "Outputs should be an array")

	assert.Contains(t, root, "dependencies")
	assert.IsType(t, []interface{}{}, root["dependencies"], "Dependencies should be an array")

	assert.Contains(t, root, "provider_dependencies")
	assert.IsType(t, []interface{}{}, root["provider_dependencies"], "Provider dependencies should be an array")

	assert.Contains(t, root, "resources")
	assert.IsType(t, []interface{}{}, root["resources"], "Resources should be an array")

	assert.Contains(t, root, "modules")
	assert.IsType(t, []interface{}{}, root["modules"], "Modules should be an array")

	// Validate submodules array (Python validates submodules)
	assert.Contains(t, response, "submodules")
	assert.IsType(t, []interface{}{}, response["submodules"], "Submodules should be an array")

	// Validate providers array (Python validates providers list)
	assert.Contains(t, response, "providers")
	assert.IsType(t, []interface{}{}, response["providers"], "Providers should be an array")

	// Validate versions array (Python validates versions list)
	assert.Contains(t, response, "versions")
	assert.IsType(t, []interface{}{}, response["versions"], "Versions should be an array")

	// Optional fields - validate if present
	if owner, ok := response["owner"]; ok && owner != nil {
		assert.NotEmpty(t, owner, "Owner should not be empty if present")
	}

	if description, ok := response["description"]; ok && description != nil {
		assert.NotEmpty(t, description, "Description should not be empty if present")
	}

	if source, ok := response["source"]; ok && source != nil {
		assert.NotEmpty(t, source, "Source should not be empty if present")
	}

	assert.Contains(t, response, "published_at")
	assert.NotNil(t, response["published_at"], "Published at should be present for published version")

	assert.Contains(t, response, "downloads")
	assert.IsType(t, float64(0), response["downloads"], "Downloads should be a number")

	// Validate module_provider_id field (Python validates this)
	assert.Contains(t, response, "module_provider_id")
	assert.Equal(t, "hashicorp/consul/aws", response["module_provider_id"])

	// Validate terraform example version string (Python validates this)
	assert.Contains(t, response, "terraform_example_version_string")
	assert.NotEmpty(t, response["terraform_example_version_string"], "Terraform example version string should not be empty")

	// Validate beta flag (Python validates this)
	assert.Contains(t, response, "beta")
	assert.IsType(t, false, response["beta"], "Beta should be a boolean")

	// Validate published flag (Python validates this)
	assert.Contains(t, response, "published")
	assert.IsType(t, true, response["published"], "Published should be a boolean")

	// Validate security failures (Python validates this)
	assert.Contains(t, response, "security_failures")
	assert.IsType(t, float64(0), response["security_failures"], "Security failures should be a number")

	// Validate git_tag_format (Python validates this)
	assert.Contains(t, response, "git_tag_format")
	assert.NotEmpty(t, response["git_tag_format"], "Git tag format should not be empty")

	// Validate graph URL (Python validates this)
	assert.Contains(t, response, "graph_url")
	assert.NotEmpty(t, response["graph_url"], "Graph URL should not be empty")

	// Validate module extraction status (Python validates this)
	assert.Contains(t, response, "module_extraction_up_to_date")
	assert.IsType(t, true, response["module_extraction_up_to_date"], "Module extraction up to date should be a boolean")

	// Validate usage example (Python validates complete terraform source)
	assert.Contains(t, response, "usage_example")
	assert.NotEmpty(t, response["usage_example"], "Usage example should not be empty")
	usageExample := response["usage_example"].(string)
	assert.Contains(t, usageExample, "module \"consul\"", "Usage example should contain module name")
	assert.Contains(t, usageExample, "source", "Usage example should contain source attribute")
}

// TestModuleHandler_HandleModuleProviderDetails_NotFound tests provider not found
func TestModuleHandler_HandleModuleProviderDetails_NotFound(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	// Create handler (no test data)
	namespaceRepository := moduleRepo.NewNamespaceRepository(db.DB)
	domainConfig := testutils.CreateTestDomainConfig(t)
	moduleProviderRepository, err := moduleRepo.NewModuleProviderRepository(db.DB, namespaceRepository, domainConfig)
	require.NoError(t, err)
	getModuleProviderQuery := moduleQuery.NewGetModuleProviderQuery(moduleProviderRepository)
	namespaceSvc := namespaceService.NewNamespaceService(domainConfig)
	analyticsRepository, err := analyticsRepo.NewAnalyticsRepository(db.DB, namespaceRepository, namespaceSvc)
	require.NoError(t, err)

	handler := terrareg.NewModuleReadHandlerForTesting(
		nil,
		nil,
		getModuleProviderQuery,
		nil,
		analyticsRepository,
		nil, // versionPresenter not needed
	)

	// Create request with chi context
	req := httptest.NewRequest("GET", "/v1/modules/unknown/module/provider", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("namespace", "unknown")
	rctx.URLParams.Add("name", "module")
	rctx.URLParams.Add("provider", "provider")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()

	// Act
	handler.HandleModuleProviderDetails(w, req)

	// Assert
	assert.Equal(t, http.StatusNotFound, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Contains(t, response, "message")
	assert.Contains(t, response["message"], "not found")
}

// TestModuleHandler_HandleModuleSearch_Success tests the module search endpoint
func TestModuleHandler_HandleModuleSearch_Success(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	// Create test data
	namespace := testutils.CreateNamespace(t, db, "searchns", nil)
	moduleProvider := testutils.CreateModuleProvider(t, db, namespace.ID, "networking-module", "aws")
	moduleVersion := testutils.CreateModuleVersion(t, db, moduleProvider.ID, "1.0.0")
	// Update version to published
	published := true
	moduleVersion.Published = &published
	db.DB.Save(&moduleVersion)

	// Create handler
	namespaceRepository := moduleRepo.NewNamespaceRepository(db.DB)
	domainConfig := testutils.CreateTestDomainConfig(t)
	moduleProviderRepository, err := moduleRepo.NewModuleProviderRepository(db.DB, namespaceRepository, domainConfig)
	require.NoError(t, err)
	searchModulesQuery, err := moduleQuery.NewSearchModulesQuery(moduleProviderRepository)
	require.NoError(t, err)
	namespaceSvc := namespaceService.NewNamespaceService(domainConfig)
	analyticsRepository, err := analyticsRepo.NewAnalyticsRepository(db.DB, namespaceRepository, namespaceSvc)
	require.NoError(t, err)

	handler := terrareg.NewModuleReadHandlerForTesting(
		nil,
		searchModulesQuery,
		nil,
		nil,
		analyticsRepository,
		nil, // versionPresenter not needed
	)

	// Create request with query parameters
	req := httptest.NewRequest("GET", "/v1/modules/search?q=networking", nil)
	w := httptest.NewRecorder()

	// Act
	handler.HandleModuleSearch(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Contains(t, response, "modules")
	assert.Contains(t, response, "meta")

	meta := response["meta"].(map[string]interface{})

	assert.Equal(t, float64(20), meta["limit"])
	assert.Equal(t, float64(0), meta["current_offset"])
}

// TestModuleHandler_HandleModuleSearch_WithFilters tests search with filters
func TestModuleHandler_HandleModuleSearch_WithFilters(t *testing.T) {
	tests := []struct {
		name        string
		queryString string
	}{
		{
			name:        "with namespace filter",
			queryString: "?q=test&namespace=searchns",
		},
		{
			name:        "with provider filter",
			queryString: "?q=test&provider=aws",
		},
		{
			name:        "with verified filter",
			queryString: "?q=test&verified=true",
		},
		{
			name:        "with custom pagination",
			queryString: "?q=test&limit=10&offset=5",
		},
		{
			name:        "with multiple filters",
			queryString: "?q=test&namespace=testns&provider=aws&verified=true&limit=10&offset=0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := testutils.SetupTestDatabase(t)
			defer testutils.CleanupTestDatabase(t, db)

			// Create handler
			namespaceRepository := moduleRepo.NewNamespaceRepository(db.DB)
			domainConfig := testutils.CreateTestDomainConfig(t)
			moduleProviderRepository, err := moduleRepo.NewModuleProviderRepository(db.DB, namespaceRepository, domainConfig)
			require.NoError(t, err)
			searchModulesQuery, err := moduleQuery.NewSearchModulesQuery(moduleProviderRepository)
			require.NoError(t, err)
			namespaceSvc := namespaceService.NewNamespaceService(domainConfig)
			analyticsRepository, err := analyticsRepo.NewAnalyticsRepository(db.DB, namespaceRepository, namespaceSvc)
			require.NoError(t, err)

			handler := terrareg.NewModuleReadHandlerForTesting(
				nil,
				searchModulesQuery,
				nil,
				nil,
				analyticsRepository,
				nil, // versionPresenter not needed
			)

			// Create request with query parameters
			req := httptest.NewRequest("GET", "/v1/modules/search"+tt.queryString, nil)
			w := httptest.NewRecorder()

			// Act
			handler.HandleModuleSearch(w, req)

			// Assert
			assert.Equal(t, http.StatusOK, w.Code)

			var response map[string]interface{}
			err = json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)

			assert.Contains(t, response, "modules")
			assert.Contains(t, response, "meta")
		})
	}
}

// TestModuleHandler_HandleModuleSearch_EmptyResults tests search with no matching results
func TestModuleHandler_HandleModuleSearch_EmptyResults(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	// Create handler (no test data)
	namespaceRepository := moduleRepo.NewNamespaceRepository(db.DB)
	domainConfig := testutils.CreateTestDomainConfig(t)
	moduleProviderRepository, err := moduleRepo.NewModuleProviderRepository(db.DB, namespaceRepository, domainConfig)
	require.NoError(t, err)
	searchModulesQuery, err := moduleQuery.NewSearchModulesQuery(moduleProviderRepository)
	require.NoError(t, err)
	namespaceSvc := namespaceService.NewNamespaceService(domainConfig)
	analyticsRepository, err := analyticsRepo.NewAnalyticsRepository(db.DB, namespaceRepository, namespaceSvc)
	require.NoError(t, err)

	handler := terrareg.NewModuleReadHandlerForTesting(
		nil,
		searchModulesQuery,
		nil,
		nil,
		analyticsRepository,
		nil, // versionPresenter not needed
	)

	// Create request
	req := httptest.NewRequest("GET", "/v1/modules/search?q=nonexistentxyz123", nil)
	w := httptest.NewRecorder()

	// Act
	handler.HandleModuleSearch(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	modules := response["modules"].([]interface{})
	assert.Len(t, modules, 0)
}

// TestModuleHandler_HandleModuleProviderDetails_WithVersion tests provider details with versions
func TestModuleHandler_HandleModuleProviderDetails_WithVersion(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	// Create test data with version
	namespace := testutils.CreateNamespace(t, db, "versionns", nil)
	moduleProvider := testutils.CreateModuleProvider(t, db, namespace.ID, "versionmodule", "aws")
	moduleVersion := testutils.CreateModuleVersion(t, db, moduleProvider.ID, "1.0.0")
	// Update version to published
	published := true
	moduleVersion.Published = &published
	db.DB.Save(&moduleVersion)

	// Create handler
	namespaceRepository := moduleRepo.NewNamespaceRepository(db.DB)
	domainConfig := testutils.CreateTestDomainConfig(t)
	moduleProviderRepository, err := moduleRepo.NewModuleProviderRepository(db.DB, namespaceRepository, domainConfig)
	require.NoError(t, err)
	getModuleProviderQuery := moduleQuery.NewGetModuleProviderQuery(moduleProviderRepository)
	namespaceSvc := namespaceService.NewNamespaceService(domainConfig)
	analyticsRepository, err := analyticsRepo.NewAnalyticsRepository(db.DB, namespaceRepository, namespaceSvc)
	require.NoError(t, err)

	handler := terrareg.NewModuleReadHandlerForTesting(
		nil,
		nil,
		getModuleProviderQuery,
		nil,
		analyticsRepository,
		nil, // versionPresenter not needed
	)

	// Create request with chi context
	req := httptest.NewRequest("GET", "/v1/modules/versionns/versionmodule/aws", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("namespace", "versionns")
	rctx.URLParams.Add("name", "versionmodule")
	rctx.URLParams.Add("provider", "aws")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()

	// Act
	handler.HandleModuleProviderDetails(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "versionns/versionmodule/aws", response["id"])
	// Should contain version information
	assert.Contains(t, response, "published_at")
}

// TestModuleHandler_HandleModuleVersionDetails_UnverifiedModuleVersion tests unverified module version
// Python reference: /app/test/unit/terrareg/server/test_api_module_version_details.py:test_unverified_module_version
func TestModuleHandler_HandleModuleVersionDetails_UnverifiedModuleVersion(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	// Create test data with unverified module (verified not set to true)
	namespace := testutils.CreateNamespace(t, db, "testnamespace", nil)
	moduleProvider := testutils.CreateModuleProvider(t, db, namespace.ID, "unverifiedmodule", "testprovider")
	moduleVersion := testutils.CreateModuleVersion(t, db, moduleProvider.ID, "1.2.3")
	// Update version to published (but not verified)
	published := true
	moduleVersion.Published = &published
	db.DB.Save(&moduleVersion)

	// Create handler using helper
	handler := testutils.CreateModuleVersionDetailsHandler(t, db)

	// Create request with chi context - note: using version-specific URL
	req := httptest.NewRequest("GET", "/v1/modules/testnamespace/unverifiedmodule/testprovider/1.2.3", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("namespace", "testnamespace")
	rctx.URLParams.Add("name", "unverifiedmodule")
	rctx.URLParams.Add("provider", "testprovider")
	rctx.URLParams.Add("version", "1.2.3")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()

	// Act
	handler.HandleModuleVersionDetails(w, req)

	// Assert - Python expects 200 with verified=False
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	// Python: assert res.json == {..., 'verified': False, ...}
	assert.Equal(t, false, response["verified"])
	assert.Equal(t, "1.2.3", response["version"])
	assert.Equal(t, "testnamespace/unverifiedmodule/testprovider/1.2.3", response["id"])
}

// TestModuleHandler_HandleModuleVersionDetails_InternalModuleVersion tests internal module version
// Python reference: /app/test/unit/terrareg/server/test_api_module_version_details.py:test_internal_module_version
func TestModuleHandler_HandleModuleVersionDetails_InternalModuleVersion(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	// Create test data with internal module version
	namespace := testutils.CreateNamespace(t, db, "testnamespace", nil)
	moduleProvider := testutils.CreateModuleProvider(t, db, namespace.ID, "internalmodule", "testprovider")
	moduleVersion := testutils.CreateModuleVersion(t, db, moduleProvider.ID, "5.2.0")
	// Update version to published and internal
	published := true
	moduleVersion.Published = &published
	internal := true
	moduleVersion.Internal = internal
	db.DB.Save(&moduleVersion)

	// Create handler using helper
	handler := testutils.CreateModuleVersionDetailsHandler(t, db)

	// Create request with chi context
	req := httptest.NewRequest("GET", "/v1/modules/testnamespace/internalmodule/testprovider/5.2.0", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("namespace", "testnamespace")
	rctx.URLParams.Add("name", "internalmodule")
	rctx.URLParams.Add("provider", "testprovider")
	rctx.URLParams.Add("version", "5.2.0")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()

	// Act
	handler.HandleModuleVersionDetails(w, req)

	// Assert - Python expects 200 with internal=True
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	// Python: assert res.json == {..., 'internal': True, ...}
	assert.Equal(t, true, response["internal"])
	assert.Equal(t, "5.2.0", response["version"])
	assert.Equal(t, "testnamespace/internalmodule/testprovider/5.2.0", response["id"])
}

// TestModuleHandler_MultipleProviders tests handler with multiple providers for the same module
func TestModuleHandler_MultipleProviders(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	// Create test data with multiple providers
	namespace := testutils.CreateNamespace(t, db, "multins", nil)
	provider1 := testutils.CreateModuleProvider(t, db, namespace.ID, "multimodule", "aws")
	provider2 := testutils.CreateModuleProvider(t, db, namespace.ID, "multimodule", "azure")
	provider3 := testutils.CreateModuleProvider(t, db, namespace.ID, "multimodule", "gcp")

	// Create published versions for each provider (required for search to find them)
	testutils.CreatePublishedModuleVersion(t, db, provider1.ID, "1.0.0")
	testutils.CreatePublishedModuleVersion(t, db, provider2.ID, "1.0.0")
	testutils.CreatePublishedModuleVersion(t, db, provider3.ID, "1.0.0")

	// Create handler
	namespaceRepository := moduleRepo.NewNamespaceRepository(db.DB)
	domainConfig := testutils.CreateTestDomainConfig(t)
	moduleProviderRepository, err := moduleRepo.NewModuleProviderRepository(db.DB, namespaceRepository, domainConfig)
	require.NoError(t, err)
	listModuleProvidersQuery := moduleQuery.NewListModuleProvidersQuery(moduleProviderRepository)
	namespaceSvc := namespaceService.NewNamespaceService(domainConfig)
	analyticsRepository, err := analyticsRepo.NewAnalyticsRepository(db.DB, namespaceRepository, namespaceSvc)
	require.NoError(t, err)

	handler := terrareg.NewModuleReadHandlerForTesting(
		nil,
		nil,
		nil,
		listModuleProvidersQuery,
		analyticsRepository,
		nil, // versionPresenter not needed
	)

	// Create request
	req := httptest.NewRequest("GET", "/v1/modules/multins/multimodule", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("namespace", "multins")
	rctx.URLParams.Add("name", "multimodule")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()

	// Act
	handler.HandleModuleDetails(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	modules := response["modules"].([]interface{})
	assert.GreaterOrEqual(t, len(modules), 3)

	providers := make([]string, 0)
	for _, m := range modules {
		module := m.(map[string]interface{})
		providers = append(providers, module["provider"].(string))
	}
	assert.Contains(t, providers, "aws")
	assert.Contains(t, providers, "azure")
	assert.Contains(t, providers, "gcp")
}

// TestModuleHandler_HandleTerraregModuleProviders_Success tests the terrareg module providers endpoint
// Python reference: /app/test/unit/terrareg/server/test_api_module_details.py
func TestModuleHandler_HandleTerraregModuleProviders_Success(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	// Create test data with multiple providers for same module
	namespace := testutils.CreateNamespace(t, db, "testns", nil)
	moduleProvider1 := testutils.CreateModuleProvider(t, db, namespace.ID, "testmodule", "aws")
	_ = testutils.CreatePublishedModuleVersion(t, db, moduleProvider1.ID, "1.0.0")
	moduleProvider2 := testutils.CreateModuleProvider(t, db, namespace.ID, "testmodule", "azure")
	_ = testutils.CreatePublishedModuleVersion(t, db, moduleProvider2.ID, "2.0.0")
	// Module provider without versions
	moduleProvider3 := testutils.CreateModuleProvider(t, db, namespace.ID, "testmodule", "gcp")

	// Create handler
	namespaceRepository := moduleRepo.NewNamespaceRepository(db.DB)
	domainConfig := testutils.CreateTestDomainConfig(t)
	moduleProviderRepository, err := moduleRepo.NewModuleProviderRepository(db.DB, namespaceRepository, domainConfig)
	require.NoError(t, err)
	listModuleProvidersQuery := moduleQuery.NewListModuleProvidersQuery(moduleProviderRepository)
	namespaceSvc := namespaceService.NewNamespaceService(domainConfig)
	analyticsRepository, err := analyticsRepo.NewAnalyticsRepository(db.DB, namespaceRepository, namespaceSvc)
	require.NoError(t, err)

	handler := terrareg.NewModuleReadHandlerForTesting(
		nil,
		nil,
		nil,
		listModuleProvidersQuery,
		analyticsRepository,
		nil, // versionPresenter not needed
	)

	// Create request for terrareg module providers endpoint
	req := httptest.NewRequest("GET", "/v1/terrareg/modules/testns/testmodule", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("namespace", "testns")
	rctx.URLParams.Add("name", "testmodule")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()

	// Act - HandleModuleDetails is used for both /v1/modules and /v1/terrareg/modules endpoints
	handler.HandleModuleDetails(w, req)

	// Assert - Validate response format matches Python's ApiTerraregModuleProviders
	// Python reference: /app/test/unit/terrareg/server/test_api_module_details.py::test_existing_module
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	// Validate meta structure (Python validates pagination metadata)
	assert.Contains(t, response, "meta")
	meta := response["meta"].(map[string]interface{})
	assert.Contains(t, meta, "limit")
	assert.Contains(t, meta, "current_offset")

	// Validate modules array exists and has expected providers
	assert.Contains(t, response, "modules")
	modules := response["modules"].([]interface{})
	assert.Len(t, modules, 3, "Should return exactly three module providers")

	// Validate each module provider
	// Python reference: modules with versions have full fields, modules without versions have minimal fields
	providers := make(map[string]map[string]interface{})
	for _, m := range modules {
		module := m.(map[string]interface{})
		providerName := module["provider"].(string)
		providers[providerName] = module

		// Common fields for all providers
		assert.Contains(t, module, "id")
		assert.NotEmpty(t, module["id"], "Module ID should not be empty")
		assert.Equal(t, "testns", module["namespace"])
		assert.Equal(t, "testmodule", module["name"])
		assert.Contains(t, module, "provider")

		assert.Contains(t, module, "verified")
		assert.IsType(t, false, module["verified"], "Verified should be a boolean")

		assert.Contains(t, module, "trusted")
		assert.IsType(t, false, module["trusted"], "Trusted should be a boolean")
	}

	// Validate aws provider (with version)
	awsModule := providers["aws"]
	assert.Contains(t, awsModule, "version")
	assert.Equal(t, "1.0.0", awsModule["version"])
	assert.Contains(t, awsModule, "published_at")

	// Validate azure provider (with version)
	azureModule := providers["azure"]
	assert.Contains(t, azureModule, "version")
	assert.Equal(t, "2.0.0", azureModule["version"])
	assert.Contains(t, azureModule, "published_at")

	// Validate gcp provider (without version) - should have minimal fields
	gcpModule := providers["gcp"]
	assert.NotContains(t, gcpModule, "version", "Provider without versions should not have version field")
	assert.NotContains(t, gcpModule, "published_at", "Provider without versions should not have published_at field")
}

// TestModuleHandler_HandleTerraregModuleProviders_NotFound tests the terrareg module providers endpoint with non-existent module
// Python reference: /app/test/unit/terrareg/server/test_api_module_details.py::test_non_existent_module
func TestModuleHandler_HandleTerraregModuleProviders_NotFound(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	// Create handler (no test data - module doesn't exist)
	namespaceRepository := moduleRepo.NewNamespaceRepository(db.DB)
	domainConfig := testutils.CreateTestDomainConfig(t)
	moduleProviderRepository, err := moduleRepo.NewModuleProviderRepository(db.DB, namespaceRepository, domainConfig)
	require.NoError(t, err)
	listModuleProvidersQuery := moduleQuery.NewListModuleProvidersQuery(moduleProviderRepository)
	namespaceSvc := namespaceService.NewNamespaceService(domainConfig)
	analyticsRepository, err := analyticsRepo.NewAnalyticsRepository(db.DB, namespaceRepository, namespaceSvc)
	require.NoError(t, err)

	handler := terrareg.NewModuleReadHandlerForTesting(
		nil,
		nil,
		nil,
		listModuleProvidersQuery,
		analyticsRepository,
		nil,
	)

	// Create request for non-existent module
	req := httptest.NewRequest("GET", "/v1/terrareg/modules/doesnotexist/nonexistent", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("namespace", "doesnotexist")
	rctx.URLParams.Add("name", "nonexistent")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()

	// Act
	handler.HandleModuleDetails(w, req)

	// Assert - Should return 404 like Python
	// Python reference: assert res.json == {'errors': ['Not Found']}
	assert.Equal(t, http.StatusNotFound, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Contains(t, response, "errors")
	errors := response["errors"].([]interface{})
	assert.NotEmpty(t, errors, "Should have errors array")
}

// TestModuleHandler_HandleTerraregModuleProviders_AnalyticsToken tests that analytics tokens are NOT converted for terrareg endpoint
// Python reference: /app/test/unit/terrareg/server/test_api_terrareg_namespace_modules.py::test_analytics_token_not_converted
func TestModuleHandler_HandleTerraregModuleProviders_AnalyticsToken(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	// Create namespace with analytics token-like name
	// Note: Analytics tokens are in format "token__namespace", but for terrareg endpoint
	// the token should NOT be converted (unlike /v1/modules endpoint)
	namespace := testutils.CreateNamespace(t, db, "test_token-name__testnamespace", nil)
	moduleProvider := testutils.CreateModuleProvider(t, db, namespace.ID, "testmodule", "aws")
	_ = testutils.CreatePublishedModuleVersion(t, db, moduleProvider.ID, "1.0.0")

	// Create handler
	namespaceRepository := moduleRepo.NewNamespaceRepository(db.DB)
	domainConfig := testutils.CreateTestDomainConfig(t)
	moduleProviderRepository, err := moduleRepo.NewModuleProviderRepository(db.DB, namespaceRepository, domainConfig)
	require.NoError(t, err)
	listModuleProvidersQuery := moduleQuery.NewListModuleProvidersQuery(moduleProviderRepository)
	namespaceSvc := namespaceService.NewNamespaceService(domainConfig)
	analyticsRepository, err := analyticsRepo.NewAnalyticsRepository(db.DB, namespaceRepository, namespaceSvc)
	require.NoError(t, err)

	handler := terrareg.NewModuleReadHandlerForTesting(
		nil,
		nil,
		nil,
		listModuleProvidersQuery,
		analyticsRepository,
		nil,
	)

	// Create request with analytics token in namespace name
	req := httptest.NewRequest("GET", "/v1/terrareg/modules/test_token-name__testnamespace/testmodule", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("namespace", "test_token-name__testnamespace")
	rctx.URLParams.Add("name", "testmodule")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()

	// Act
	handler.HandleModuleDetails(w, req)

	// Assert - Should return 404 because the namespace with analytics token is not converted
	// Python reference: Analytics tokens are NOT converted for /v1/terrareg/modules/ endpoint
	assert.Equal(t, http.StatusNotFound, w.Code)
}
