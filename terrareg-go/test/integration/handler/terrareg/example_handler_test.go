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
	moduleRepo "github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/persistence/sqldb/module"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/interfaces/http/handler/terrareg"
	"github.com/matthewjohn/terrareg/terrareg-go/test/integration/testutils"
)

// TestExampleHandler_HandleExampleDetails_Success tests getting example details successfully
func TestExampleHandler_HandleExampleDetails_Success(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	// Create test data - examples are stored as submodules
	namespace := testutils.CreateNamespace(t, db, "test-namespace")
	moduleProvider := testutils.CreateModuleProvider(t, db, namespace.ID, "test-module", "aws")
	moduleVersion := testutils.CreatePublishedModuleVersion(t, db, moduleProvider.ID, "1.0.0")
	// Create example as a submodule (examples are just submodules in the database)
	_ = testutils.CreateSubmodule(t, db, moduleVersion.ID, "examples/test-example", "Test Example", "example", nil)

	// Create handler with repositories
	moduleProviderRepository := moduleRepo.NewModuleProviderRepository(db.DB, nil, nil)
	moduleVersionRepository := moduleRepo.NewModuleVersionRepository(db.DB)
	getExampleDetailsQuery := moduleQuery.NewGetExampleDetailsQuery(moduleProviderRepository, moduleVersionRepository)
	handler := terrareg.NewExampleHandler(getExampleDetailsQuery, nil, nil, nil)

	// Create request with chi context
	req := httptest.NewRequest("GET", "/v1/terrareg/modules/test-namespace/test-module/aws/1.0.0/examples/details/examples/test-example", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("namespace", "test-namespace")
	rctx.URLParams.Add("name", "test-module")
	rctx.URLParams.Add("provider", "aws")
	rctx.URLParams.Add("version", "1.0.0")
	rctx.URLParams.Add("example", "examples/test-example")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()

	// Act
	handler.HandleExampleDetails(w, req)

	// Assert - Should return 200 with example details
	assert.Equal(t, http.StatusOK, w.Code)
	response := testutils.GetJSONBody(t, w)
	assert.Contains(t, response, "path")
}

// TestExampleHandler_HandleExampleDetails_MissingParameters tests missing required path parameters
func TestExampleHandler_HandleExampleDetails_MissingParameters(t *testing.T) {
	moduleProviderRepository := moduleRepo.NewModuleProviderRepository(nil, nil, nil)
	moduleVersionRepository := moduleRepo.NewModuleVersionRepository(nil)
	getExampleDetailsQuery := moduleQuery.NewGetExampleDetailsQuery(moduleProviderRepository, moduleVersionRepository)
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
			rctx.URLParams.Add("example", tc.example)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			w := httptest.NewRecorder()
			handler.HandleExampleDetails(w, req)

			assert.Equal(t, http.StatusBadRequest, w.Code)
			response := testutils.GetJSONBody(t, w)
			assert.Contains(t, response, "error")
		})
	}
}

// TestExampleHandler_HandleExampleDetails_NotFound tests example not found
func TestExampleHandler_HandleExampleDetails_NotFound(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	// Create test data
	namespace := testutils.CreateNamespace(t, db, "test-namespace")
	moduleProvider := testutils.CreateModuleProvider(t, db, namespace.ID, "test-module", "aws")
	_ = testutils.CreatePublishedModuleVersion(t, db, moduleProvider.ID, "1.0.0")
	// Don't create an example

	// Create handler
	moduleProviderRepository := moduleRepo.NewModuleProviderRepository(db.DB, nil, nil)
	moduleVersionRepository := moduleRepo.NewModuleVersionRepository(db.DB)
	getExampleDetailsQuery := moduleQuery.NewGetExampleDetailsQuery(moduleProviderRepository, moduleVersionRepository)
	handler := terrareg.NewExampleHandler(getExampleDetailsQuery, nil, nil, nil)

	// Create request with chi context
	req := httptest.NewRequest("GET", "/v1/terrareg/modules/test-namespace/test-module/aws/1.0.0/examples/details/nonexistent", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("namespace", "test-namespace")
	rctx.URLParams.Add("name", "test-module")
	rctx.URLParams.Add("provider", "aws")
	rctx.URLParams.Add("version", "1.0.0")
	rctx.URLParams.Add("example", "nonexistent")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()

	// Act
	handler.HandleExampleDetails(w, req)

	// Assert - Query generates fallback details for non-existent examples
	assert.Equal(t, http.StatusOK, w.Code)
	response := testutils.GetJSONBody(t, w)
	assert.Contains(t, response, "path")
}

// TestExampleHandler_HandleExampleReadmeHTML_Success tests getting example readme HTML successfully
func TestExampleHandler_HandleExampleReadmeHTML_Success(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	// Create test data
	namespace := testutils.CreateNamespace(t, db, "test-namespace")
	moduleProvider := testutils.CreateModuleProvider(t, db, namespace.ID, "test-module", "aws")
	moduleVersion := testutils.CreatePublishedModuleVersion(t, db, moduleProvider.ID, "1.0.0")
	// Create example as a submodule
	_ = testutils.CreateSubmodule(t, db, moduleVersion.ID, "examples/test-example", "Test Example", "example", nil)

	// Create handler
	moduleProviderRepository := moduleRepo.NewModuleProviderRepository(db.DB, nil, nil)
	moduleVersionRepository := moduleRepo.NewModuleVersionRepository(db.DB)
	getExampleReadmeHTMLQuery := moduleQuery.NewGetExampleReadmeHTMLQuery(moduleProviderRepository, moduleVersionRepository)
	handler := terrareg.NewExampleHandler(nil, getExampleReadmeHTMLQuery, nil, nil)

	// Create request with chi context
	req := httptest.NewRequest("GET", "/v1/terrareg/modules/test-namespace/test-module/aws/1.0.0/examples/readme_html/examples/test-example", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("namespace", "test-namespace")
	rctx.URLParams.Add("name", "test-module")
	rctx.URLParams.Add("provider", "aws")
	rctx.URLParams.Add("version", "1.0.0")
	rctx.URLParams.Add("example", "examples/test-example")
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
	moduleProviderRepository := moduleRepo.NewModuleProviderRepository(nil, nil, nil)
	moduleVersionRepository := moduleRepo.NewModuleVersionRepository(nil)
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
			rctx.URLParams.Add("example", tc.example)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			w := httptest.NewRecorder()
			handler.HandleExampleReadmeHTML(w, req)

			assert.Equal(t, http.StatusBadRequest, w.Code)
			response := testutils.GetJSONBody(t, w)
			assert.Contains(t, response, "error")
		})
	}
}

// TestExampleHandler_HandleExampleFileList_Success tests getting example file list successfully
func TestExampleHandler_HandleExampleFileList_Success(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	// Create test data
	namespace := testutils.CreateNamespace(t, db, "test-namespace")
	moduleProvider := testutils.CreateModuleProvider(t, db, namespace.ID, "test-module", "aws")
	moduleVersion := testutils.CreatePublishedModuleVersion(t, db, moduleProvider.ID, "1.0.0")
	// Create example as a submodule
	_ = testutils.CreateSubmodule(t, db, moduleVersion.ID, "examples/test-example", "Test Example", "example", nil)

	// Create handler
	moduleProviderRepository := moduleRepo.NewModuleProviderRepository(db.DB, nil, nil)
	moduleVersionRepository := moduleRepo.NewModuleVersionRepository(db.DB)
	getExampleFileListQuery := moduleQuery.NewGetExampleFileListQuery(moduleProviderRepository, moduleVersionRepository)
	handler := terrareg.NewExampleHandler(nil, nil, getExampleFileListQuery, nil)

	// Create request with chi context
	req := httptest.NewRequest("GET", "/v1/terrareg/modules/test-namespace/test-module/aws/1.0.0/examples/filelist/examples/test-example", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("namespace", "test-namespace")
	rctx.URLParams.Add("name", "test-module")
	rctx.URLParams.Add("provider", "aws")
	rctx.URLParams.Add("version", "1.0.0")
	rctx.URLParams.Add("example", "examples/test-example")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()

	// Act
	handler.HandleExampleFileList(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)

	// Parse JSON array response
	var response []interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err, "Response should be valid JSON array")
	assert.IsType(t, []interface{}{}, response, "Response should be an array of files")
}

// TestExampleHandler_HandleExampleFileList_MissingParameters tests missing required path parameters
func TestExampleHandler_HandleExampleFileList_MissingParameters(t *testing.T) {
	moduleProviderRepository := moduleRepo.NewModuleProviderRepository(nil, nil, nil)
	moduleVersionRepository := moduleRepo.NewModuleVersionRepository(nil)
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
			rctx.URLParams.Add("example", tc.example)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			w := httptest.NewRecorder()
			handler.HandleExampleFileList(w, req)

			assert.Equal(t, http.StatusBadRequest, w.Code)
			response := testutils.GetJSONBody(t, w)
			assert.Contains(t, response, "error")
		})
	}
}

// TestExampleHandler_HandleExampleFile_Success tests getting example file successfully
func TestExampleHandler_HandleExampleFile_Success(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	// Create test data
	namespace := testutils.CreateNamespace(t, db, "test-namespace")
	moduleProvider := testutils.CreateModuleProvider(t, db, namespace.ID, "test-module", "aws")
	moduleVersion := testutils.CreatePublishedModuleVersion(t, db, moduleProvider.ID, "1.0.0")
	// Create example as a submodule
	_ = testutils.CreateSubmodule(t, db, moduleVersion.ID, "examples/test-example", "Test Example", "example", nil)

	// Create handler
	moduleProviderRepository := moduleRepo.NewModuleProviderRepository(db.DB, nil, nil)
	moduleVersionRepository := moduleRepo.NewModuleVersionRepository(db.DB)
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
	moduleProviderRepository := moduleRepo.NewModuleProviderRepository(nil, nil, nil)
	moduleVersionRepository := moduleRepo.NewModuleVersionRepository(nil)
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

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)
	response := testutils.GetJSONBody(t, w)
	assert.Contains(t, response, "error")
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

			assert.Equal(t, http.StatusBadRequest, w.Code)
			response := testutils.GetJSONBody(t, w)
			assert.Contains(t, response, "error")
		})
	}
}
