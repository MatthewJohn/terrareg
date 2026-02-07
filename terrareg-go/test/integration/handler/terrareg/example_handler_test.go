package terrareg_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	moduleQuery "github.com/matthewjohn/terrareg/terrareg-go/internal/application/query/module"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/url/service"
	moduleRepo "github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/persistence/sqldb/module"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/interfaces/http/handler/terrareg"
	"github.com/matthewjohn/terrareg/terrareg-go/test/integration/testutils"
)

// TestExampleHandler_HandleExampleDetails_Success tests getting example details successfully
func TestExampleHandler_HandleExampleDetails_Success(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	// Create test data - examples are stored as submodules
	namespace := testutils.CreateNamespace(t, db, "test-namespace", nil)
	moduleProvider := testutils.CreateModuleProvider(t, db, namespace.ID, "test-module", "aws")
	moduleVersion := testutils.CreatePublishedModuleVersion(t, db, moduleProvider.ID, "1.0.0")
	// Create example as a submodule (examples are just submodules in the database)
	_ = testutils.CreateSubmodule(t, db, moduleVersion.ID, "examples/test-example", "Test Example", "example", nil)

	// Create handler with repositories
	namespaceRepo := moduleRepo.NewNamespaceRepository(db.DB)
	domainConfig := testutils.CreateTestDomainConfig(t)
	moduleProviderRepository, err := moduleRepo.NewModuleProviderRepository(db.DB, namespaceRepo, domainConfig)
	require.NoError(t, err)
	moduleVersionRepository, err := moduleRepo.NewModuleVersionRepository(db.DB)
	require.NoError(t, err)
	infraConfig := testutils.CreateTestInfraConfig(t)
	urlService, err := service.NewURLService(infraConfig)
	require.NoError(t, err)
	getExampleDetailsQuery := moduleQuery.NewGetExampleDetailsQuery(moduleProviderRepository, moduleVersionRepository, urlService)
	handler := terrareg.NewExampleHandler(getExampleDetailsQuery, nil, nil, nil)

	// Create request with chi context
	req := httptest.NewRequest("GET", "/v1/terrareg/modules/test-namespace/test-module/aws/1.0.0/examples/details/examples/test-example", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("namespace", "test-namespace")
	rctx.URLParams.Add("name", "test-module")
	rctx.URLParams.Add("provider", "aws")
	rctx.URLParams.Add("version", "1.0.0")
	rctx.URLParams.Add("*", "examples/test-example")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()

	// Act
	handler.HandleExampleDetails(w, req)

	// Assert - Should return 200 with example details
	assert.Equal(t, http.StatusOK, w.Code)
	response := testutils.GetJSONBody(t, w)
	assert.Contains(t, response, "path")

	// Verify usage_example contains the full URL with domain and port (HTTP mode)
	usageExample, ok := response["usage_example"].(string)
	assert.True(t, ok, "usage_example should be a string")
	assert.Contains(t, usageExample, "localhost:5000/modules/", "usage_example should contain full URL with domain and port")
	assert.Contains(t, usageExample, "test-namespace/test-module/aws/1.0.0//examples/test-example", "usage_example should contain provider, version and path")
}

// TestExampleHandler_HandleExampleDetails_MissingParameters tests missing required path parameters
func TestExampleHandler_HandleExampleDetails_MissingParameters(t *testing.T) {
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
	getExampleDetailsQuery := moduleQuery.NewGetExampleDetailsQuery(moduleProviderRepository, moduleVersionRepository, urlService)
	handler := terrareg.NewExampleHandler(getExampleDetailsQuery, nil, nil, nil)

	tests := []struct {
		name       string
		namespace  string
		moduleName string
		provider   string
		version    string
		example    string
	}{
		{"missing namespace", "", "module", "aws", "1.0.0", "example"},
		{"missing module", "ns", "", "aws", "1.0.0", "example"},
		{"missing provider", "ns", "module", "", "1.0.0", "example"},
		{"missing version", "ns", "module", "aws", "", "example"},
		{"missing example", "ns", "module", "aws", "1.0.0", ""},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/v1/terrareg/modules/", nil)
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("namespace", tc.namespace)
			rctx.URLParams.Add("name", tc.moduleName)
			rctx.URLParams.Add("provider", tc.provider)
			rctx.URLParams.Add("version", tc.version)
			rctx.URLParams.Add("*", tc.example)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			w := httptest.NewRecorder()
			handler.HandleExampleDetails(w, req)

			// Enhanced error validation - validate specific error message
			assert.Equal(t, http.StatusBadRequest, w.Code)
			testutils.AssertErrorContains(t, w, "Missing required path parameters")
		})
	}
}

// TestExampleHandler_HandleExampleDetails_NotFound tests example not found
func TestExampleHandler_HandleExampleDetails_NotFound(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	// Create test data
	namespace := testutils.CreateNamespace(t, db, "test-namespace", nil)
	moduleProvider := testutils.CreateModuleProvider(t, db, namespace.ID, "test-module", "aws")
	_ = testutils.CreatePublishedModuleVersion(t, db, moduleProvider.ID, "1.0.0")
	// Don't create an example

	// Create handler
	namespaceRepo := moduleRepo.NewNamespaceRepository(db.DB)
	domainConfig := testutils.CreateTestDomainConfig(t)
	moduleProviderRepository, err := moduleRepo.NewModuleProviderRepository(db.DB, namespaceRepo, domainConfig)
	require.NoError(t, err)
	moduleVersionRepository, err := moduleRepo.NewModuleVersionRepository(db.DB)
	require.NoError(t, err)
	infraConfig := testutils.CreateTestInfraConfig(t)
	urlService, err := service.NewURLService(infraConfig)
	require.NoError(t, err)
	getExampleDetailsQuery := moduleQuery.NewGetExampleDetailsQuery(moduleProviderRepository, moduleVersionRepository, urlService)
	handler := terrareg.NewExampleHandler(getExampleDetailsQuery, nil, nil, nil)

	// Create request with chi context
	req := httptest.NewRequest("GET", "/v1/terrareg/modules/test-namespace/test-module/aws/1.0.0/examples/details/nonexistent", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("namespace", "test-namespace")
	rctx.URLParams.Add("name", "test-module")
	rctx.URLParams.Add("provider", "aws")
	rctx.URLParams.Add("version", "1.0.0")
	rctx.URLParams.Add("*", "nonexistent")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()

	// Act
	handler.HandleExampleDetails(w, req)

	// Assert - Should return 404 for non-existent examples (like Python)
	// Enhanced error validation - validate specific error message
	assert.Equal(t, http.StatusNotFound, w.Code)
	testutils.AssertErrorContains(t, w, "example not found")
}

// TestExampleHandler_HandleExampleReadmeHTML_Success tests getting example readme HTML successfully
func TestExampleHandler_HandleExampleReadmeHTML_Success(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	// Create test data
	namespace := testutils.CreateNamespace(t, db, "test-namespace", nil)
	moduleProvider := testutils.CreateModuleProvider(t, db, namespace.ID, "test-module", "aws")
	moduleVersion := testutils.CreatePublishedModuleVersion(t, db, moduleProvider.ID, "1.0.0")
	// Create example as a submodule
	_ = testutils.CreateSubmodule(t, db, moduleVersion.ID, "examples/test-example", "Test Example", "example", nil)

	// Create handler
	namespaceRepo := moduleRepo.NewNamespaceRepository(db.DB)
	domainConfig := testutils.CreateTestDomainConfig(t)
	moduleProviderRepository, err := moduleRepo.NewModuleProviderRepository(db.DB, namespaceRepo, domainConfig)
	require.NoError(t, err)
	moduleVersionRepository, err := moduleRepo.NewModuleVersionRepository(db.DB)
	require.NoError(t, err)
	getExampleReadmeHTMLQuery := moduleQuery.NewGetExampleReadmeHTMLQuery(moduleProviderRepository, moduleVersionRepository)
	handler := terrareg.NewExampleHandler(nil, getExampleReadmeHTMLQuery, nil, nil)

	// Create request with chi context
	req := httptest.NewRequest("GET", "/v1/terrareg/modules/test-namespace/test-module/aws/1.0.0/examples/readme_html/examples/test-example", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("namespace", "test-namespace")
	rctx.URLParams.Add("name", "test-module")
	rctx.URLParams.Add("provider", "aws")
	rctx.URLParams.Add("version", "1.0.0")
	rctx.URLParams.Add("*", "examples/test-example")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()

	// Act
	handler.HandleExampleReadmeHTML(w, req)

	// Assert - Should return 200 with HTML content
	assert.Equal(t, http.StatusOK, w.Code)
	assert.NotEmpty(t, w.Body.String())
	assert.Equal(t, "text/html; charset=utf-8", w.Header().Get("Content-Type"))
}

// TestExampleHandler_HandleExampleReadmeHTML_MissingParameters tests missing required path parameters
func TestExampleHandler_HandleExampleReadmeHTML_MissingParameters(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	namespaceRepo := moduleRepo.NewNamespaceRepository(db.DB)
	domainConfig := testutils.CreateTestDomainConfig(t)
	moduleProviderRepository, err := moduleRepo.NewModuleProviderRepository(db.DB, namespaceRepo, domainConfig)
	require.NoError(t, err)
	moduleVersionRepository, err := moduleRepo.NewModuleVersionRepository(db.DB)
	require.NoError(t, err)
	getExampleReadmeHTMLQuery := moduleQuery.NewGetExampleReadmeHTMLQuery(moduleProviderRepository, moduleVersionRepository)
	handler := terrareg.NewExampleHandler(nil, getExampleReadmeHTMLQuery, nil, nil)

	tests := []struct {
		name       string
		namespace  string
		moduleName string
		provider   string
		version    string
		example    string
	}{
		{"missing namespace", "", "module", "aws", "1.0.0", "example"},
		{"missing module", "ns", "", "aws", "1.0.0", "example"},
		{"missing provider", "ns", "module", "", "1.0.0", "example"},
		{"missing version", "ns", "module", "aws", "", "example"},
		{"missing example", "ns", "module", "aws", "1.0.0", ""},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/v1/terrareg/modules/", nil)
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("namespace", tc.namespace)
			rctx.URLParams.Add("name", tc.moduleName)
			rctx.URLParams.Add("provider", tc.provider)
			rctx.URLParams.Add("version", tc.version)
			rctx.URLParams.Add("*", tc.example)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			w := httptest.NewRecorder()
			handler.HandleExampleReadmeHTML(w, req)

			// Enhanced error validation - validate specific error message
			assert.Equal(t, http.StatusBadRequest, w.Code)
			testutils.AssertErrorContains(t, w, "Missing required path parameters")
		})
	}
}

// TestExampleHandler_HandleExampleFileList_Success tests getting example file list successfully
func TestExampleHandler_HandleExampleFileList_Success(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	// Create test data
	namespace := testutils.CreateNamespace(t, db, "test-namespace", nil)
	moduleProvider := testutils.CreateModuleProvider(t, db, namespace.ID, "test-module", "aws")
	moduleVersion := testutils.CreatePublishedModuleVersion(t, db, moduleProvider.ID, "1.0.0")
	// Create example as a submodule
	_ = testutils.CreateSubmodule(t, db, moduleVersion.ID, "examples/test-example", "Test Example", "example", nil)

	// Create handler
	namespaceRepo := moduleRepo.NewNamespaceRepository(db.DB)
	domainConfig := testutils.CreateTestDomainConfig(t)
	moduleProviderRepository, err := moduleRepo.NewModuleProviderRepository(db.DB, namespaceRepo, domainConfig)
	require.NoError(t, err)
	moduleVersionRepository, err := moduleRepo.NewModuleVersionRepository(db.DB)
	require.NoError(t, err)
	getExampleFileListQuery := moduleQuery.NewGetExampleFileListQuery(moduleProviderRepository, moduleVersionRepository)
	handler := terrareg.NewExampleHandler(nil, nil, getExampleFileListQuery, nil)

	// Create request with chi context
	req := httptest.NewRequest("GET", "/v1/terrareg/modules/test-namespace/test-module/aws/1.0.0/examples/filelist/examples/test-example", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("namespace", "test-namespace")
	rctx.URLParams.Add("name", "test-module")
	rctx.URLParams.Add("provider", "aws")
	rctx.URLParams.Add("version", "1.0.0")
	rctx.URLParams.Add("*", "examples/test-example")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()

	// Act
	handler.HandleExampleFileList(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)

	// Parse JSON array response
	var response []interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err, "Response should be valid JSON array")
	assert.IsType(t, []interface{}{}, response, "Response should be an array of files")
}

// TestExampleHandler_HandleExampleFileList_MissingParameters tests missing required path parameters
func TestExampleHandler_HandleExampleFileList_MissingParameters(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	namespaceRepo := moduleRepo.NewNamespaceRepository(db.DB)
	domainConfig := testutils.CreateTestDomainConfig(t)
	moduleProviderRepository, err := moduleRepo.NewModuleProviderRepository(db.DB, namespaceRepo, domainConfig)
	require.NoError(t, err)
	moduleVersionRepository, err := moduleRepo.NewModuleVersionRepository(db.DB)
	require.NoError(t, err)
	getExampleFileListQuery := moduleQuery.NewGetExampleFileListQuery(moduleProviderRepository, moduleVersionRepository)
	handler := terrareg.NewExampleHandler(nil, nil, getExampleFileListQuery, nil)

	tests := []struct {
		name       string
		namespace  string
		moduleName string
		provider   string
		version    string
		example    string
	}{
		{"missing namespace", "", "module", "aws", "1.0.0", "example"},
		{"missing module", "ns", "", "aws", "1.0.0", "example"},
		{"missing provider", "ns", "module", "", "1.0.0", "example"},
		{"missing version", "ns", "module", "aws", "", "example"},
		{"missing example", "ns", "module", "aws", "1.0.0", ""},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/v1/terrareg/modules/", nil)
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("namespace", tc.namespace)
			rctx.URLParams.Add("name", tc.moduleName)
			rctx.URLParams.Add("provider", tc.provider)
			rctx.URLParams.Add("version", tc.version)
			rctx.URLParams.Add("*", tc.example)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			w := httptest.NewRecorder()
			handler.HandleExampleFileList(w, req)

			// Enhanced error validation - validate specific error message
			assert.Equal(t, http.StatusBadRequest, w.Code)
			testutils.AssertErrorContains(t, w, "Missing required path parameters")
		})
	}
}

// TestExampleHandler_HandleExampleFile_Success tests getting example file successfully
func TestExampleHandler_HandleExampleFile_Success(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	// Create test data
	namespace := testutils.CreateNamespace(t, db, "test-namespace", nil)
	moduleProvider := testutils.CreateModuleProvider(t, db, namespace.ID, "test-module", "aws")
	moduleVersion := testutils.CreatePublishedModuleVersion(t, db, moduleProvider.ID, "1.0.0")
	// Create example as a submodule
	_ = testutils.CreateSubmodule(t, db, moduleVersion.ID, "examples/test-example", "Test Example", "example", nil)

	// Create handler
	namespaceRepo := moduleRepo.NewNamespaceRepository(db.DB)
	domainConfig := testutils.CreateTestDomainConfig(t)
	moduleProviderRepository, err := moduleRepo.NewModuleProviderRepository(db.DB, namespaceRepo, domainConfig)
	require.NoError(t, err)
	moduleVersionRepository, err := moduleRepo.NewModuleVersionRepository(db.DB)
	require.NoError(t, err)
	getExampleFileQuery := moduleQuery.NewGetExampleFileQuery(moduleProviderRepository, moduleVersionRepository)
	handler := terrareg.NewExampleHandler(nil, nil, nil, getExampleFileQuery)

	// Create request with chi context
	// The file parameter combines example path and file path
	req := httptest.NewRequest("GET", "/v1/terrareg/modules/test-namespace/test-module/aws/1.0.0/examples/file/examples/test-example/main.tf", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("namespace", "test-namespace")
	rctx.URLParams.Add("name", "test-module")
	rctx.URLParams.Add("provider", "aws")
	rctx.URLParams.Add("version", "1.0.0")
	rctx.URLParams.Add("file", "examples/test-example/main.tf")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()

	// Act
	handler.HandleExampleFile(w, req)

	// Assert - Should return file content or error (implementation-specific)
	assert.NotEqual(t, http.StatusBadRequest, w.Code, "Should not return bad request for valid parameters")
}

// TestExampleHandler_HandleExampleFile_MissingFileParam tests missing file parameter
func TestExampleHandler_HandleExampleFile_MissingFileParam(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	namespaceRepo := moduleRepo.NewNamespaceRepository(db.DB)
	domainConfig := testutils.CreateTestDomainConfig(t)
	moduleProviderRepository, err := moduleRepo.NewModuleProviderRepository(db.DB, namespaceRepo, domainConfig)
	require.NoError(t, err)
	moduleVersionRepository, err := moduleRepo.NewModuleVersionRepository(db.DB)
	require.NoError(t, err)
	getExampleFileQuery := moduleQuery.NewGetExampleFileQuery(moduleProviderRepository, moduleVersionRepository)
	handler := terrareg.NewExampleHandler(nil, nil, nil, getExampleFileQuery)

	// Create request without file parameter
	req := httptest.NewRequest("GET", "/v1/terrareg/modules/ns/module/aws/1.0.0/examples/file/", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("namespace", "ns")
	rctx.URLParams.Add("name", "module")
	rctx.URLParams.Add("provider", "aws")
	rctx.URLParams.Add("version", "1.0.0")
	rctx.URLParams.Add("file", "")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()

	// Act
	handler.HandleExampleFile(w, req)

	// Assert - Enhanced error validation - validate specific error message
	assert.Equal(t, http.StatusBadRequest, w.Code)
	testutils.AssertErrorContains(t, w, "Missing required path parameters")
}

// TestExampleHandler_HandleExampleFile_MissingOtherParameters tests missing other required parameters
func TestExampleHandler_HandleExampleFile_MissingOtherParameters(t *testing.T) {
	getExampleFileQuery := moduleQuery.NewGetExampleFileQuery(nil, nil)
	handler := terrareg.NewExampleHandler(nil, nil, nil, getExampleFileQuery)

	tests := []struct {
		name       string
		namespace  string
		moduleName string
		provider   string
		version    string
		file       string
	}{
		{"missing namespace", "", "module", "aws", "1.0.0", "example/file.tf"},
		{"missing module", "ns", "", "aws", "1.0.0", "example/file.tf"},
		{"missing provider", "ns", "module", "", "1.0.0", "example/file.tf"},
		{"missing version", "ns", "module", "aws", "", "example/file.tf"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/v1/terrareg/modules/", nil)
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("namespace", tc.namespace)
			rctx.URLParams.Add("name", tc.moduleName)
			rctx.URLParams.Add("provider", tc.provider)
			rctx.URLParams.Add("version", tc.version)
			rctx.URLParams.Add("file", tc.file)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			w := httptest.NewRecorder()
			handler.HandleExampleFile(w, req)

			// Enhanced error validation - validate specific error message
			assert.Equal(t, http.StatusBadRequest, w.Code)
			testutils.AssertErrorContains(t, w, "Missing required path parameters")
		})
	}
}

// TestExampleHandler_HandleExampleDetails_HTTPSWithNonStandardPort tests terraform source URL
// with HTTPS and non-standard port (e.g., 5000 instead of 443)
func TestExampleHandler_HandleExampleDetails_HTTPSWithNonStandardPort(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	// Create test data - examples are stored as submodules
	namespace := testutils.CreateNamespace(t, db, "test-namespace", nil)
	moduleProvider := testutils.CreateModuleProvider(t, db, namespace.ID, "test-module", "aws")
	moduleVersion := testutils.CreatePublishedModuleVersion(t, db, moduleProvider.ID, "1.0.0")
	// Create example as a submodule (examples are just submodules in the database)
	_ = testutils.CreateSubmodule(t, db, moduleVersion.ID, "examples/test-example", "Test Example", "example", nil)

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
	getExampleDetailsQuery := moduleQuery.NewGetExampleDetailsQuery(moduleProviderRepository, moduleVersionRepository, urlService)
	handler := terrareg.NewExampleHandler(getExampleDetailsQuery, nil, nil, nil)

	// Create request with chi context
	req := httptest.NewRequest("GET", "/v1/terrareg/modules/test-namespace/test-module/aws/1.0.0/examples/details/examples/test-example", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("namespace", "test-namespace")
	rctx.URLParams.Add("name", "test-module")
	rctx.URLParams.Add("provider", "aws")
	rctx.URLParams.Add("version", "1.0.0")
	rctx.URLParams.Add("*", "examples/test-example")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()

	// Act
	handler.HandleExampleDetails(w, req)

	// Assert - Should return 200 with example details
	assert.Equal(t, http.StatusOK, w.Code)
	response := testutils.GetJSONBody(t, w)
	assert.Contains(t, response, "path")

	// Verify usage_example contains the domain:port format (HTTPS with non-standard port)
	usageExample, ok := response["usage_example"].(string)
	assert.True(t, ok, "usage_example should be a string")
	// For HTTPS with non-standard port, format should be: domain:port/providerID//modulePath
	// e.g., local-dev.dock.studio:5000/test-namespace/test-module/aws//examples/test-example
	assert.Contains(t, usageExample, "local-dev.dock.studio:5000/", "usage_example should contain domain with port for HTTPS with non-standard port")
	assert.Contains(t, usageExample, "test-namespace/test-module/aws//examples/test-example", "usage_example should contain provider and path")
	// Verify no protocol in HTTPS URL
	assert.NotContains(t, usageExample, "http://", "HTTPS URL should not contain http:// protocol")
	assert.NotContains(t, usageExample, "https://", "HTTPS URL should not contain https:// protocol")
	// Verify version is NOT in the URL for HTTPS (it should only be after the source, as a separate attribute)
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
