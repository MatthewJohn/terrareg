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
	analyticsRepo "github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/persistence/sqldb/analytics"
	moduleRepo "github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/persistence/sqldb/module"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/interfaces/http/handler/terrareg"
	"github.com/matthewjohn/terrareg/terrareg-go/test/integration/testutils"
)

// TestModuleHandler_HandleModuleList_Success tests the module list endpoint
func TestModuleHandler_HandleModuleList_Success(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	// Create test data
	namespace := testutils.CreateNamespace(t, db, "testnamespace")
	moduleProvider := testutils.CreateModuleProvider(t, db, namespace.ID, "testmodule", "aws")
	moduleVersion := testutils.CreateModuleVersion(t, db, moduleProvider.ID, "1.0.0")

	// Update version to published
	published := true
	moduleVersion.Published = &published
	db.DB.Save(&moduleVersion)

	// Create handler
	namespaceRepository := moduleRepo.NewNamespaceRepository(db.DB)
	moduleProviderRepository := moduleRepo.NewModuleProviderRepository(db.DB, namespaceRepository, nil)
	listModulesQuery := moduleQuery.NewListModulesQuery(moduleProviderRepository)
	analyticsRepository := analyticsRepo.NewAnalyticsRepository(db.DB, namespaceRepository, nil)

	handler := terrareg.NewModuleReadHandlerForTesting(
		listModulesQuery,
		nil, // search not used
		nil, // get provider not used
		nil, // list providers not used
		analyticsRepository,
	)

	// Create request
	req := httptest.NewRequest("GET", "/v1/modules", nil)
	w := httptest.NewRecorder()

	// Act
	handler.HandleModuleList(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
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
	moduleProviderRepository := moduleRepo.NewModuleProviderRepository(db.DB, namespaceRepository, nil)
	listModulesQuery := moduleQuery.NewListModulesQuery(moduleProviderRepository)
	analyticsRepository := analyticsRepo.NewAnalyticsRepository(db.DB, namespaceRepository, nil)

	handler := terrareg.NewModuleReadHandlerForTesting(
		listModulesQuery,
		nil,
		nil,
		nil,
		analyticsRepository,
	)

	// Create request
	req := httptest.NewRequest("GET", "/v1/modules", nil)
	w := httptest.NewRecorder()

	// Act
	handler.HandleModuleList(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	modules := response["modules"].([]interface{})
	assert.Len(t, modules, 0)
}

// TestModuleHandler_HandleNamespaceModules_Success tests the namespace modules endpoint
func TestModuleHandler_HandleNamespaceModules_Success(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	// Create test data
	namespace := testutils.CreateNamespace(t, db, "mynamespace")
	moduleProvider := testutils.CreateModuleProvider(t, db, namespace.ID, "mymodule", "aws")
	moduleVersion := testutils.CreateModuleVersion(t, db, moduleProvider.ID, "1.0.0")
	// Update version to published
	published := true
	moduleVersion.Published = &published
	db.DB.Save(&moduleVersion)

	// Create handler
	namespaceRepository := moduleRepo.NewNamespaceRepository(db.DB)
	moduleProviderRepository := moduleRepo.NewModuleProviderRepository(db.DB, namespaceRepository, nil)
	listModulesQuery := moduleQuery.NewListModulesQuery(moduleProviderRepository)
	analyticsRepository := analyticsRepo.NewAnalyticsRepository(db.DB, namespaceRepository, nil)

	handler := terrareg.NewModuleReadHandlerForTesting(
		listModulesQuery,
		nil,
		nil,
		nil,
		analyticsRepository,
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
	err := json.Unmarshal(w.Body.Bytes(), &response)
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
	moduleProviderRepository := moduleRepo.NewModuleProviderRepository(db.DB, namespaceRepository, nil)
	listModulesQuery := moduleQuery.NewListModulesQuery(moduleProviderRepository)
	analyticsRepository := analyticsRepo.NewAnalyticsRepository(db.DB, namespaceRepository, nil)

	handler := terrareg.NewModuleReadHandlerForTesting(
		listModulesQuery,
		nil,
		nil,
		nil,
		analyticsRepository,
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
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	modules := response["modules"].([]interface{})
	assert.Len(t, modules, 0)
}

// TestModuleHandler_HandleModuleDetails_Success tests the module details endpoint
func TestModuleHandler_HandleModuleDetails_Success(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	// Create test data
	namespace := testutils.CreateNamespace(t, db, "testns")
	testutils.CreateModuleProvider(t, db, namespace.ID, "testmodule", "aws")
	testutils.CreateModuleProvider(t, db, namespace.ID, "testmodule", "azure")

	// Create handler
	namespaceRepository := moduleRepo.NewNamespaceRepository(db.DB)
	moduleProviderRepository := moduleRepo.NewModuleProviderRepository(db.DB, namespaceRepository, nil)
	listModuleProvidersQuery := moduleQuery.NewListModuleProvidersQuery(moduleProviderRepository)
	analyticsRepository := analyticsRepo.NewAnalyticsRepository(db.DB, namespaceRepository, nil)

	handler := terrareg.NewModuleReadHandlerForTesting(
		nil, // list not used
		nil, // search not used
		nil, // get provider not used
		listModuleProvidersQuery,
		analyticsRepository,
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

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Contains(t, response, "modules")
	modules := response["modules"].([]interface{})
	assert.GreaterOrEqual(t, len(modules), 2)

	// Verify all modules match the requested namespace and name
	for _, m := range modules {
		module := m.(map[string]interface{})
		assert.Equal(t, "testns", module["namespace"])
		assert.Equal(t, "testmodule", module["name"])
	}
}

// TestModuleHandler_HandleModuleProviderDetails_Success tests the module provider details endpoint
func TestModuleHandler_HandleModuleProviderDetails_Success(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	// Create test data
	namespace := testutils.CreateNamespace(t, db, "hashicorp")
	moduleProvider := testutils.CreateModuleProvider(t, db, namespace.ID, "consul", "aws")
	moduleVersion := testutils.CreateModuleVersion(t, db, moduleProvider.ID, "1.0.0")
	// Update version to published
	published := true
	moduleVersion.Published = &published
	db.DB.Save(&moduleVersion)

	// Create handler
	namespaceRepository := moduleRepo.NewNamespaceRepository(db.DB)
	moduleProviderRepository := moduleRepo.NewModuleProviderRepository(db.DB, namespaceRepository, nil)
	getModuleProviderQuery := moduleQuery.NewGetModuleProviderQuery(moduleProviderRepository)
	analyticsRepository := analyticsRepo.NewAnalyticsRepository(db.DB, namespaceRepository, nil)

	handler := terrareg.NewModuleReadHandlerForTesting(
		nil,
		nil,
		getModuleProviderQuery,
		nil,
		analyticsRepository,
	)

	// Create request with chi context
	req := httptest.NewRequest("GET", "/v1/modules/hashicorp/consul/aws", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("namespace", "hashicorp")
	rctx.URLParams.Add("name", "consul")
	rctx.URLParams.Add("provider", "aws")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()

	// Act
	handler.HandleModuleProviderDetails(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "hashicorp/consul/aws", response["id"])
	assert.Equal(t, "hashicorp", response["namespace"])
	assert.Equal(t, "consul", response["name"])
	assert.Equal(t, "aws", response["provider"])
	assert.Contains(t, response, "downloads")
}

// TestModuleHandler_HandleModuleProviderDetails_NotFound tests provider not found
func TestModuleHandler_HandleModuleProviderDetails_NotFound(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	// Create handler (no test data)
	namespaceRepository := moduleRepo.NewNamespaceRepository(db.DB)
	moduleProviderRepository := moduleRepo.NewModuleProviderRepository(db.DB, namespaceRepository, nil)
	getModuleProviderQuery := moduleQuery.NewGetModuleProviderQuery(moduleProviderRepository)
	analyticsRepository := analyticsRepo.NewAnalyticsRepository(db.DB, namespaceRepository, nil)

	handler := terrareg.NewModuleReadHandlerForTesting(
		nil,
		nil,
		getModuleProviderQuery,
		nil,
		analyticsRepository,
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
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Contains(t, response, "error")
	assert.Contains(t, response["error"], "not found")
}

// TestModuleHandler_HandleModuleSearch_Success tests the module search endpoint
func TestModuleHandler_HandleModuleSearch_Success(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	// Create test data
	namespace := testutils.CreateNamespace(t, db, "searchns")
	moduleProvider := testutils.CreateModuleProvider(t, db, namespace.ID, "networking-module", "aws")
	moduleVersion := testutils.CreateModuleVersion(t, db, moduleProvider.ID, "1.0.0")
	// Update version to published
	published := true
	moduleVersion.Published = &published
	db.DB.Save(&moduleVersion)

	// Create handler
	namespaceRepository := moduleRepo.NewNamespaceRepository(db.DB)
	moduleProviderRepository := moduleRepo.NewModuleProviderRepository(db.DB, namespaceRepository, nil)
	searchModulesQuery := moduleQuery.NewSearchModulesQuery(moduleProviderRepository)
	analyticsRepository := analyticsRepo.NewAnalyticsRepository(db.DB, namespaceRepository, nil)

	handler := terrareg.NewModuleReadHandlerForTesting(
		nil,
		searchModulesQuery,
		nil,
		nil,
		analyticsRepository,
	)

	// Create request with query parameters
	req := httptest.NewRequest("GET", "/v1/modules/search?q=networking", nil)
	w := httptest.NewRecorder()

	// Act
	handler.HandleModuleSearch(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
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
			moduleProviderRepository := moduleRepo.NewModuleProviderRepository(db.DB, namespaceRepository, nil)
			searchModulesQuery := moduleQuery.NewSearchModulesQuery(moduleProviderRepository)
			analyticsRepository := analyticsRepo.NewAnalyticsRepository(db.DB, namespaceRepository, nil)

			handler := terrareg.NewModuleReadHandlerForTesting(
				nil,
				searchModulesQuery,
				nil,
				nil,
				analyticsRepository,
			)

			// Create request with query parameters
			req := httptest.NewRequest("GET", "/v1/modules/search"+tt.queryString, nil)
			w := httptest.NewRecorder()

			// Act
			handler.HandleModuleSearch(w, req)

			// Assert
			assert.Equal(t, http.StatusOK, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
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
	moduleProviderRepository := moduleRepo.NewModuleProviderRepository(db.DB, namespaceRepository, nil)
	searchModulesQuery := moduleQuery.NewSearchModulesQuery(moduleProviderRepository)
	analyticsRepository := analyticsRepo.NewAnalyticsRepository(db.DB, namespaceRepository, nil)

	handler := terrareg.NewModuleReadHandlerForTesting(
		nil,
		searchModulesQuery,
		nil,
		nil,
		analyticsRepository,
	)

	// Create request
	req := httptest.NewRequest("GET", "/v1/modules/search?q=nonexistentxyz123", nil)
	w := httptest.NewRecorder()

	// Act
	handler.HandleModuleSearch(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	modules := response["modules"].([]interface{})
	assert.Len(t, modules, 0)
}

// TestModuleHandler_HandleModuleProviderDetails_WithVersion tests provider details with versions
func TestModuleHandler_HandleModuleProviderDetails_WithVersion(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	// Create test data with version
	namespace := testutils.CreateNamespace(t, db, "versionns")
	moduleProvider := testutils.CreateModuleProvider(t, db, namespace.ID, "versionmodule", "aws")
	moduleVersion := testutils.CreateModuleVersion(t, db, moduleProvider.ID, "1.0.0")
	// Update version to published
	published := true
	moduleVersion.Published = &published
	db.DB.Save(&moduleVersion)

	// Create handler
	namespaceRepository := moduleRepo.NewNamespaceRepository(db.DB)
	moduleProviderRepository := moduleRepo.NewModuleProviderRepository(db.DB, namespaceRepository, nil)
	getModuleProviderQuery := moduleQuery.NewGetModuleProviderQuery(moduleProviderRepository)
	analyticsRepository := analyticsRepo.NewAnalyticsRepository(db.DB, namespaceRepository, nil)

	handler := terrareg.NewModuleReadHandlerForTesting(
		nil,
		nil,
		getModuleProviderQuery,
		nil,
		analyticsRepository,
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
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "versionns/versionmodule/aws", response["id"])
	// Should contain version information
	assert.Contains(t, response, "published_at")
}

// TestModuleHandler_MultipleProviders tests handler with multiple providers for the same module
func TestModuleHandler_MultipleProviders(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	// Create test data with multiple providers
	namespace := testutils.CreateNamespace(t, db, "multins")
	testutils.CreateModuleProvider(t, db, namespace.ID, "multimodule", "aws")
	testutils.CreateModuleProvider(t, db, namespace.ID, "multimodule", "azure")
	testutils.CreateModuleProvider(t, db, namespace.ID, "multimodule", "gcp")

	// Create handler
	namespaceRepository := moduleRepo.NewNamespaceRepository(db.DB)
	moduleProviderRepository := moduleRepo.NewModuleProviderRepository(db.DB, namespaceRepository, nil)
	listModuleProvidersQuery := moduleQuery.NewListModuleProvidersQuery(moduleProviderRepository)
	analyticsRepository := analyticsRepo.NewAnalyticsRepository(db.DB, namespaceRepository, nil)

	handler := terrareg.NewModuleReadHandlerForTesting(
		nil,
		nil,
		nil,
		listModuleProvidersQuery,
		analyticsRepository,
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
	err := json.Unmarshal(w.Body.Bytes(), &response)
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
