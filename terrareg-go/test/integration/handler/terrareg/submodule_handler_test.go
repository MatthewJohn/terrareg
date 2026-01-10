package terrareg_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"

	moduleQuery "github.com/matthewjohn/terrareg/terrareg-go/internal/application/query/module"
	moduleRepo "github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/persistence/sqldb/module"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/interfaces/http/handler/terrareg"
	"github.com/matthewjohn/terrareg/terrareg-go/test/integration/testutils"
)

// TestSubmoduleHandler_HandleSubmoduleDetails_Success tests getting submodule details successfully
func TestSubmoduleHandler_HandleSubmoduleDetails_Success(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	// Create test data
	namespace := testutils.CreateNamespace(t, db, "test-namespace")
	moduleProvider := testutils.CreateModuleProvider(t, db, namespace.ID, "test-module", "aws")
	moduleVersion := testutils.CreatePublishedModuleVersion(t, db, moduleProvider.ID, "1.0.0")
	_ = testutils.CreateSubmodule(t, db, moduleVersion.ID, "submodules/terraform-aws-modules/submodule", "Submodule description", "", nil)

	// Create handler with repositories
	moduleProviderRepository := moduleRepo.NewModuleProviderRepository(db.DB, nil, nil)
	moduleVersionRepository := moduleRepo.NewModuleVersionRepository(db.DB)
	getSubmoduleDetailsQuery := moduleQuery.NewGetSubmoduleDetailsQuery(moduleProviderRepository, moduleVersionRepository)
	handler := terrareg.NewSubmoduleHandler(getSubmoduleDetailsQuery, nil)

	// Create request with chi context
	req := httptest.NewRequest("GET", "/v1/terrareg/modules/test-namespace/test-module/aws/1.0.0/submodules/details/submodules/terraform-aws-modules/submodule", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("namespace", "test-namespace")
	rctx.URLParams.Add("name", "test-module")
	rctx.URLParams.Add("provider", "aws")
	rctx.URLParams.Add("version", "1.0.0")
	rctx.URLParams.Add("submodule", "submodules/terraform-aws-modules/submodule")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()

	// Act
	handler.HandleSubmoduleDetails(w, req)

	// Assert - Should return 200 with submodule details
	assert.Equal(t, http.StatusOK, w.Code)
	response := testutils.GetJSONBody(t, w)
	assert.Contains(t, response, "path")
	assert.Equal(t, "submodules/terraform-aws-modules/submodule", response["path"])
}

// TestSubmoduleHandler_HandleSubmoduleDetails_MissingParameters tests missing required path parameters
func TestSubmoduleHandler_HandleSubmoduleDetails_MissingParameters(t *testing.T) {
	moduleProviderRepository := moduleRepo.NewModuleProviderRepository(nil, nil, nil)
	moduleVersionRepository := moduleRepo.NewModuleVersionRepository(nil)
	getSubmoduleDetailsQuery := moduleQuery.NewGetSubmoduleDetailsQuery(moduleProviderRepository, moduleVersionRepository)
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
			rctx.URLParams.Add("submodule", tc.submodule)
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
	namespace := testutils.CreateNamespace(t, db, "test-namespace")
	moduleProvider := testutils.CreateModuleProvider(t, db, namespace.ID, "test-module", "aws")
	moduleVersion := testutils.CreatePublishedModuleVersion(t, db, moduleProvider.ID, "1.0.0")
	_ = testutils.CreateSubmodule(t, db, moduleVersion.ID, "submodules/test-submodule", "Test submodule", "", nil)

	// Create handler with repositories
	moduleProviderRepository := moduleRepo.NewModuleProviderRepository(db.DB, nil, nil)
	moduleVersionRepository := moduleRepo.NewModuleVersionRepository(db.DB)
	getSubmoduleReadmeHTMLQuery := moduleQuery.NewGetSubmoduleReadmeHTMLQuery(moduleProviderRepository, moduleVersionRepository)
	handler := terrareg.NewSubmoduleHandler(nil, getSubmoduleReadmeHTMLQuery)

	// Create request with chi context
	req := httptest.NewRequest("GET", "/v1/terrareg/modules/test-namespace/test-module/aws/1.0.0/submodules/readme_html/submodules/test-submodule", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("namespace", "test-namespace")
	rctx.URLParams.Add("name", "test-module")
	rctx.URLParams.Add("provider", "aws")
	rctx.URLParams.Add("version", "1.0.0")
	rctx.URLParams.Add("submodule", "submodules/test-submodule")
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
	namespace := testutils.CreateNamespace(t, db, "test-namespace")
	moduleProvider := testutils.CreateModuleProvider(t, db, namespace.ID, "test-module", "aws")
	_ = testutils.CreatePublishedModuleVersion(t, db, moduleProvider.ID, "1.0.0")
	// Don't create a submodule - testing the case where it doesn't exist

	// Create handler with repositories
	moduleProviderRepository := moduleRepo.NewModuleProviderRepository(db.DB, nil, nil)
	moduleVersionRepository := moduleRepo.NewModuleVersionRepository(db.DB)
	getSubmoduleReadmeHTMLQuery := moduleQuery.NewGetSubmoduleReadmeHTMLQuery(moduleProviderRepository, moduleVersionRepository)
	handler := terrareg.NewSubmoduleHandler(nil, getSubmoduleReadmeHTMLQuery)

	// Create request with chi context - using non-existent submodule
	req := httptest.NewRequest("GET", "/v1/terrareg/modules/test-namespace/test-module/aws/1.0.0/submodules/readme_html/nonexistent", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("namespace", "test-namespace")
	rctx.URLParams.Add("name", "test-module")
	rctx.URLParams.Add("provider", "aws")
	rctx.URLParams.Add("version", "1.0.0")
	rctx.URLParams.Add("submodule", "nonexistent")
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
	moduleProviderRepository := moduleRepo.NewModuleProviderRepository(nil, nil, nil)
	moduleVersionRepository := moduleRepo.NewModuleVersionRepository(nil)
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
			rctx.URLParams.Add("submodule", tc.submodule)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			w := httptest.NewRecorder()
			handler.HandleSubmoduleReadmeHTML(w, req)

			assert.Equal(t, http.StatusBadRequest, w.Code)
			response := testutils.GetJSONBody(t, w)
			assert.Contains(t, response, "error")
		})
	}
}

// TestSubmoduleHandler_NotImplemented is a placeholder noting that full integration tests
// require the submodule query implementation to be completed
// TODO: Implement full integration tests once submodule queries are fully implemented
func TestSubmoduleHandler_NotImplemented(t *testing.T) {
	t.Skip("Submodule queries not yet fully implemented - queries return nil repository errors")
}
