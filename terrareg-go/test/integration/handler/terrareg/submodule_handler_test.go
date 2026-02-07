package terrareg_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	moduleQuery "github.com/matthewjohn/terrareg/terrareg-go/internal/application/query/module"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/url/service"
	urlService "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/url/service"
	moduleRepo "github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/persistence/sqldb/module"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/interfaces/http/handler/terrareg"
	"github.com/matthewjohn/terrareg/terrareg-go/test/integration/testutils"
)

// TestSubmoduleHandler_HandleSubmoduleDetails_Success tests getting submodule details successfully
// Python reference: /app/test/unit/terrareg/server/test_api_terrareg_module_version_details.py - test_submodules
func TestSubmoduleHandler_HandleSubmoduleDetails_Success(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	// Create test data
	namespace := testutils.CreateNamespace(t, db, "test-namespace", nil)
	moduleProvider := testutils.CreateModuleProvider(t, db, namespace.ID, "test-module", "aws")
	moduleVersion := testutils.CreatePublishedModuleVersion(t, db, moduleProvider.ID, "1.0.0")
	_ = testutils.CreateSubmodule(t, db, moduleVersion.ID, "submodules/terraform-aws-modules/submodule", "Submodule description", "", nil)

	// Create handler with repositories
	namespaceRepo := moduleRepo.NewNamespaceRepository(db.DB)
	domainConfig := testutils.CreateTestDomainConfig(t)
	moduleProviderRepository, err := moduleRepo.NewModuleProviderRepository(db.DB, namespaceRepo, domainConfig)
	require.NoError(t, err)
	moduleVersionRepository, err := moduleRepo.NewModuleVersionRepository(db.DB)
	require.NoError(t, err)
	infraConfig := testutils.CreateTestInfraConfig(t)
	urlService, err := urlService.NewURLService(infraConfig)
	require.NoError(t, err)
	getSubmoduleDetailsQuery := moduleQuery.NewGetSubmoduleDetailsQuery(moduleProviderRepository, moduleVersionRepository, urlService)
	handler := terrareg.NewSubmoduleHandler(getSubmoduleDetailsQuery, nil)

	// Create request with chi context
	req := httptest.NewRequest("GET", "/v1/terrareg/modules/test-namespace/test-module/aws/1.0.0/submodules/details/submodules/terraform-aws-modules/submodule", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("namespace", "test-namespace")
	rctx.URLParams.Add("name", "test-module")
	rctx.URLParams.Add("provider", "aws")
	rctx.URLParams.Add("version", "1.0.0")
	rctx.URLParams.Add("*", "submodules/terraform-aws-modules/submodule")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()

	// Act
	handler.HandleSubmoduleDetails(w, req)

	// Assert - Comprehensive validation matching Python pattern
	// Python reference: validates root["modules"] with name, source, version, description
	assert.Equal(t, http.StatusOK, w.Code)
	response := testutils.GetJSONBody(t, w)

	// Validate all required fields (Python validates complete response)
	assert.Contains(t, response, "path")
	assert.Equal(t, "submodules/terraform-aws-modules/submodule", response["path"])

	// Validate readme field (Python validates readme)
	assert.Contains(t, response, "readme")
	// Readme may be empty if not provided during submodule creation
	_, ok := response["readme"].(string)
	assert.True(t, ok, "Readme should be a string")

	// Validate empty flag (Python validates this)
	assert.Contains(t, response, "empty")
	assert.IsType(t, false, response["empty"], "Empty should be a boolean")

	// Validate all array fields (Python validates these are arrays)
	assert.Contains(t, response, "inputs")
	assert.IsType(t, []interface{}{}, response["inputs"], "Inputs should be an array")

	assert.Contains(t, response, "outputs")
	assert.IsType(t, []interface{}{}, response["outputs"], "Outputs should be an array")

	assert.Contains(t, response, "dependencies")
	assert.IsType(t, []interface{}{}, response["dependencies"], "Dependencies should be an array")

	assert.Contains(t, response, "provider_dependencies")
	assert.IsType(t, []interface{}{}, response["provider_dependencies"], "Provider dependencies should be an array")

	assert.Contains(t, response, "resources")
	assert.IsType(t, []interface{}{}, response["resources"], "Resources should be an array")

	// Validate modules array (Python validates this contains submodule/module references)
	// Python reference: root["modules"] == [{name, source, version, description}, ...]
	assert.Contains(t, response, "modules")
	assert.IsType(t, []interface{}{}, response["modules"], "Modules should be an array")

	// Validate usage_example contains the full URL with domain and port (HTTP mode)
	usageExample, ok := response["usage_example"].(string)
	assert.True(t, ok, "usage_example should be a string")
	assert.Contains(t, usageExample, "localhost:5000/modules/", "usage_example should contain full URL with domain and port")
	assert.Contains(t, usageExample, "test-namespace/test-module/aws/1.0.0//submodules/terraform-aws-modules/submodule", "usage_example should contain provider, version and path")
}

// TestSubmoduleHandler_HandleSubmoduleDetails_MissingParameters tests missing required path parameters
func TestSubmoduleHandler_HandleSubmoduleDetails_MissingParameters(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	namespaceRepo := moduleRepo.NewNamespaceRepository(db.DB)
	domainConfig := testutils.CreateTestDomainConfig(t)
	moduleProviderRepository, err := moduleRepo.NewModuleProviderRepository(db.DB, namespaceRepo, domainConfig)
	require.NoError(t, err)
	moduleVersionRepository, err := moduleRepo.NewModuleVersionRepository(db.DB)
	require.NoError(t, err)
	infraConfig := testutils.CreateTestInfraConfig(t)
	urlService, err := service.NewURLService(infraConfig)
	require.NoError(t, err)
	getSubmoduleDetailsQuery := moduleQuery.NewGetSubmoduleDetailsQuery(moduleProviderRepository, moduleVersionRepository, urlService)
	handler := terrareg.NewSubmoduleHandler(getSubmoduleDetailsQuery, nil)

	tests := []struct {
		name       string
		namespace  string
		moduleName string
		provider   string
		version    string
		submodule  string
	}{
		{"missing namespace", "", "module", "aws", "1.0.0", "submodule"},
		{"missing module", "ns", "", "aws", "1.0.0", "submodule"},
		{"missing provider", "ns", "module", "", "1.0.0", "submodule"},
		{"missing version", "ns", "module", "aws", "", "submodule"},
		{"missing submodule", "ns", "module", "aws", "1.0.0", ""},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/v1/terrareg/modules/", nil)
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("namespace", tc.namespace)
			rctx.URLParams.Add("name", tc.moduleName)
			rctx.URLParams.Add("provider", tc.provider)
			rctx.URLParams.Add("version", tc.version)
			rctx.URLParams.Add("*", tc.submodule)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			w := httptest.NewRecorder()
			handler.HandleSubmoduleDetails(w, req)

			assert.Equal(t, http.StatusBadRequest, w.Code)
			response := testutils.GetJSONBody(t, w)
			assert.Contains(t, response, "error")
		})
	}
}

// TestSubmoduleHandler_HandleSubmoduleReadmeHTML_Success tests getting submodule readme HTML successfully
func TestSubmoduleHandler_HandleSubmoduleReadmeHTML_Success(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	// Create test data
	namespace := testutils.CreateNamespace(t, db, "test-namespace", nil)
	moduleProvider := testutils.CreateModuleProvider(t, db, namespace.ID, "test-module", "aws")
	moduleVersion := testutils.CreatePublishedModuleVersion(t, db, moduleProvider.ID, "1.0.0")
	_ = testutils.CreateSubmodule(t, db, moduleVersion.ID, "submodules/test-submodule", "Test submodule", "", nil)

	// Create handler with repositories
	namespaceRepo := moduleRepo.NewNamespaceRepository(db.DB)
	domainConfig := testutils.CreateTestDomainConfig(t)
	moduleProviderRepository, err := moduleRepo.NewModuleProviderRepository(db.DB, namespaceRepo, domainConfig)
	require.NoError(t, err)
	moduleVersionRepository, err := moduleRepo.NewModuleVersionRepository(db.DB)
	require.NoError(t, err)
	getSubmoduleReadmeHTMLQuery := moduleQuery.NewGetSubmoduleReadmeHTMLQuery(moduleProviderRepository, moduleVersionRepository)
	handler := terrareg.NewSubmoduleHandler(nil, getSubmoduleReadmeHTMLQuery)

	// Create request with chi context
	req := httptest.NewRequest("GET", "/v1/terrareg/modules/test-namespace/test-module/aws/1.0.0/submodules/readme_html/submodules/test-submodule", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("namespace", "test-namespace")
	rctx.URLParams.Add("name", "test-module")
	rctx.URLParams.Add("provider", "aws")
	rctx.URLParams.Add("version", "1.0.0")
	rctx.URLParams.Add("*", "submodules/test-submodule")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()

	// Act
	handler.HandleSubmoduleReadmeHTML(w, req)

	// Assert - Should return 200 with HTML content
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "README")
}

// TestSubmoduleHandler_HandleSubmoduleReadmeHTML_NoReadme tests getting submodule readme when none exists
func TestSubmoduleHandler_HandleSubmoduleReadmeHTML_NoReadme(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	// Create test data - module without submodule
	namespace := testutils.CreateNamespace(t, db, "test-namespace", nil)
	moduleProvider := testutils.CreateModuleProvider(t, db, namespace.ID, "test-module", "aws")
	_ = testutils.CreatePublishedModuleVersion(t, db, moduleProvider.ID, "1.0.0")
	// Don't create a submodule - testing the case where it doesn't exist

	// Create handler with repositories
	namespaceRepo := moduleRepo.NewNamespaceRepository(db.DB)
	domainConfig := testutils.CreateTestDomainConfig(t)
	moduleProviderRepository, err := moduleRepo.NewModuleProviderRepository(db.DB, namespaceRepo, domainConfig)
	require.NoError(t, err)
	moduleVersionRepository, err := moduleRepo.NewModuleVersionRepository(db.DB)
	require.NoError(t, err)
	getSubmoduleReadmeHTMLQuery := moduleQuery.NewGetSubmoduleReadmeHTMLQuery(moduleProviderRepository, moduleVersionRepository)
	handler := terrareg.NewSubmoduleHandler(nil, getSubmoduleReadmeHTMLQuery)

	// Create request with chi context - using non-existent submodule
	req := httptest.NewRequest("GET", "/v1/terrareg/modules/test-namespace/test-module/aws/1.0.0/submodules/readme_html/nonexistent", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("namespace", "test-namespace")
	rctx.URLParams.Add("name", "test-module")
	rctx.URLParams.Add("provider", "aws")
	rctx.URLParams.Add("version", "1.0.0")
	rctx.URLParams.Add("*", "nonexistent")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()

	// Act
	handler.HandleSubmoduleReadmeHTML(w, req)

	// Assert - Should return 200 with HTML content (either README or warning message)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.NotEmpty(t, w.Body.String(), "Response should have body content")
}

// TestSubmoduleHandler_HandleSubmoduleReadmeHTML_MissingParameters tests missing required path parameters
func TestSubmoduleHandler_HandleSubmoduleReadmeHTML_MissingParameters(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	namespaceRepo := moduleRepo.NewNamespaceRepository(db.DB)
	domainConfig := testutils.CreateTestDomainConfig(t)
	moduleProviderRepository, err := moduleRepo.NewModuleProviderRepository(db.DB, namespaceRepo, domainConfig)
	require.NoError(t, err)
	moduleVersionRepository, err := moduleRepo.NewModuleVersionRepository(db.DB)
	require.NoError(t, err)
	getSubmoduleReadmeHTMLQuery := moduleQuery.NewGetSubmoduleReadmeHTMLQuery(moduleProviderRepository, moduleVersionRepository)
	handler := terrareg.NewSubmoduleHandler(nil, getSubmoduleReadmeHTMLQuery)

	tests := []struct {
		name       string
		namespace  string
		moduleName string
		provider   string
		version    string
		submodule  string
	}{
		{"missing namespace", "", "module", "aws", "1.0.0", "submodule"},
		{"missing module", "ns", "", "aws", "1.0.0", "submodule"},
		{"missing provider", "ns", "module", "", "1.0.0", "submodule"},
		{"missing version", "ns", "module", "aws", "", "submodule"},
		{"missing submodule", "ns", "module", "aws", "1.0.0", ""},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/v1/terrareg/modules/", nil)
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("namespace", tc.namespace)
			rctx.URLParams.Add("name", tc.moduleName)
			rctx.URLParams.Add("provider", tc.provider)
			rctx.URLParams.Add("version", tc.version)
			rctx.URLParams.Add("*", tc.submodule)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			w := httptest.NewRecorder()
			handler.HandleSubmoduleReadmeHTML(w, req)

			assert.Equal(t, http.StatusBadRequest, w.Code)
			response := testutils.GetJSONBody(t, w)
			assert.Contains(t, response, "error")
		})
	}
}

// TestSubmoduleHandler_HandleSubmoduleDetails_HTTPSWithNonStandardPort tests terraform source URL
// with HTTPS and non-standard port (e.g., 5000 instead of 443)
func TestSubmoduleHandler_HandleSubmoduleDetails_HTTPSWithNonStandardPort(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	// Create test data
	namespace := testutils.CreateNamespace(t, db, "test-namespace", nil)
	moduleProvider := testutils.CreateModuleProvider(t, db, namespace.ID, "test-module", "aws")
	moduleVersion := testutils.CreatePublishedModuleVersion(t, db, moduleProvider.ID, "1.0.0")
	_ = testutils.CreateSubmodule(t, db, moduleVersion.ID, "modules/example-submodule1", "Submodule description", "", nil)

	// Create handler with repositories
	namespaceRepo := moduleRepo.NewNamespaceRepository(db.DB)
	domainConfig := testutils.CreateTestDomainConfig(t)
	moduleProviderRepository, err := moduleRepo.NewModuleProviderRepository(db.DB, namespaceRepo, domainConfig)
	require.NoError(t, err)
	moduleVersionRepository, err := moduleRepo.NewModuleVersionRepository(db.DB)
	require.NoError(t, err)

	// Create infra config with HTTPS and non-standard port (5000)
	infraConfig := testutils.CreateTestInfraConfigWithPublicURL(t, "https://local-dev.dock.studio:5000")
	urlService, err := service.NewURLService(infraConfig)
	require.NoError(t, err)
	getSubmoduleDetailsQuery := moduleQuery.NewGetSubmoduleDetailsQuery(moduleProviderRepository, moduleVersionRepository, urlService)
	handler := terrareg.NewSubmoduleHandler(getSubmoduleDetailsQuery, nil)

	// Create request with chi context
	req := httptest.NewRequest("GET", "/v1/terrareg/modules/test-namespace/test-module/aws/1.0.0/submodules/details/modules/example-submodule1", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("namespace", "test-namespace")
	rctx.URLParams.Add("name", "test-module")
	rctx.URLParams.Add("provider", "aws")
	rctx.URLParams.Add("version", "1.0.0")
	rctx.URLParams.Add("*", "modules/example-submodule1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()

	// Act
	handler.HandleSubmoduleDetails(w, req)

	// Assert - Should return 200 with submodule details
	assert.Equal(t, http.StatusOK, w.Code)
	response := testutils.GetJSONBody(t, w)
	assert.Contains(t, response, "path")
	assert.Equal(t, "modules/example-submodule1", response["path"])

	// Verify usage_example contains the domain:port format (HTTPS with non-standard port)
	usageExample, ok := response["usage_example"].(string)
	assert.True(t, ok, "usage_example should be a string")
	// For HTTPS with non-standard port, format should be: domain:port/providerID//modulePath
	// e.g., local-dev.dock.studio:5000/test-namespace/test-module/aws//modules/example-submodule1
	assert.Contains(t, usageExample, "local-dev.dock.studio:5000/", "usage_example should contain domain with port for HTTPS with non-standard port")
	assert.Contains(t, usageExample, "test-namespace/test-module/aws//modules/example-submodule1", "usage_example should contain provider and path")
	// Verify no protocol in HTTPS URL
	assert.NotContains(t, usageExample, "http://", "HTTPS URL should not contain http:// protocol")
	assert.NotContains(t, usageExample, "https://", "HTTPS URL should not contain https:// protocol")
	// Verify version is NOT in the URL for HTTPS (it should only be after the source, as a separate attribute)
	// Check that the source line doesn't contain /1.0.0 (the version would be on a separate version line)
	sourceLineIdx := strings.Index(usageExample, "source =")
	if sourceLineIdx != -1 {
		// Extract just the source line and verify version is not in it
		sourceLineEnd := strings.Index(usageExample[sourceLineIdx:], "\n")
		if sourceLineEnd == -1 {
			sourceLineEnd = len(usageExample)
		}
		sourceLine := usageExample[sourceLineIdx : sourceLineIdx+sourceLineEnd]
		assert.NotContains(t, sourceLine, "/1.0.0", "HTTPS source URL should not contain version in the source attribute")
	}
	// For HTTPS with non-standard port, version should be present as a separate attribute
	assert.Contains(t, usageExample, "version =", "HTTPS URL should have version as a separate attribute")
}
