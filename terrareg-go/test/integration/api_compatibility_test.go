package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/application/command/module"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/application/query/module"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/auth/model"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/config/model"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/module/model"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/persistence/sqldb"
	terraformHandler "github.com/matthewjohn/terrareg/terrareg-go/internal/interfaces/http/handler/terraform/v1"
	moduleHandler "github.com/matthewjohn/terrareg/terrareg-go/internal/interfaces/http/handler/terrareg"
	"github.com/matthewjohn/terrareg/terrareg-go/test/integration/testutils"
)

func TestAPICompatibility(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer func() {
		require.NoError(t, db.Close())
	}()

	// Setup repositories
	namespaceRepo := sqldb.NewNamespaceRepository(db.DB)
	moduleProviderRepo := sqldb.NewModuleProviderRepository(db.DB)
	moduleVersionRepo := sqldb.NewModuleVersionRepository(db.DB)
	namespaceRedirectRepo := sqldb.NewNamespaceRedirectRepository(db.DB)
	moduleProviderRedirectRepo := sqldb.NewModuleProviderRedirectRepository(db.DB)

	// Create domain config
	domainConfig := &model.DomainConfig{
		AllowModuleHosting: model.ModuleHostingModeAllow,
	}

	t.Run("Terraform Registry API - Module List", func(t *testing.T) {
		// Create test data
		namespace, err := namespaceRepo.Create(context.Background(), "testnamespace", false)
		require.NoError(t, err)

		moduleProvider := model.NewModuleProvider(namespace, "testmodule", "aws")
		require.NoError(t, err)
		err = moduleProviderRepo.Save(context.Background(), moduleProvider)
		require.NoError(t, err)

		// Create handler
		listModulesQuery := query.NewListModulesQuery(moduleProviderRepo)
		handler := terraformHandler.NewModuleHandler(listModulesQuery, nil, nil)

		// Test Terraform v1 API endpoint
		req := httptest.NewRequest("GET", "/v1/modules", nil)
		w := httptest.NewRecorder()

		handler.HandleModuleList(w, req)
		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		// Verify Terraform Registry API format
		assert.Contains(t, response, "modules")
		modules := response["modules"].([]interface{})
		assert.Len(t, modules, 1)

		module := modules[0].(map[string]interface{})
		assert.Equal(t, "testnamespace", module["namespace"])
		assert.Equal(t, "testmodule", module["name"])
		assert.Equal(t, "aws", module["provider"])
		assert.Equal(t, "https://www.terraform.io", module["source"])
	})

	t.Run("Terraform Registry API - Module Details", func(t *testing.T) {
		// Create test data
		namespace, err := namespaceRepo.Create(context.Background(), "testnamespace", false)
		require.NoError(t, err)

		moduleProvider := model.NewModuleProvider(namespace, "testmodule", "aws")
		require.NoError(t, err)
		err = moduleProviderRepo.Save(context.Background(), moduleProvider)
		require.NoError(t, err)

		// Create module version
		version, err := shared.ParseVersion("1.0.0")
		require.NoError(t, err)

		details := &model.ModuleDetails{
			Owner:       "testowner",
			Description: "Test module",
		}

		moduleVersion, err := model.NewModuleVersion(version.String(), details, false)
		require.NoError(t, err)
		moduleVersion.SetModuleProvider(moduleProvider)
		err = moduleVersionRepo.Save(context.Background(), moduleVersion)
		require.NoError(t, err)

		// Create handler
		getModuleVersionQuery := query.NewGetModuleVersionQuery(moduleVersionRepo)
		handler := terraformHandler.NewModuleHandler(nil, nil, getModuleVersionQuery)

		// Test Terraform v1 API endpoint
		req := httptest.NewRequest("GET", "/v1/modules/testnamespace/testmodule/aws/1.0.0", nil)
		w := httptest.NewRecorder()

		handler.HandleModuleVersion(w, req)
		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		// Verify Terraform Registry API format
		assert.Equal(t, "testnamespace", response["namespace"])
		assert.Equal(t, "testmodule", response["name"])
		assert.Equal(t, "aws", response["provider"])
		assert.Equal(t, "1.0.0", response["version"])
		assert.Equal(t, "testowner", response["owner"])
		assert.Equal(t, "Test module", response["description"])
		assert.Equal(t, "https://www.terraform.io", response["source"])
		assert.Equal(t, "https://registry.terraform.io", response["published_at"])
	})

	t.Run("Terrareg API - Module Provider Details", func(t *testing.T) {
		// Create test data
		namespace, err := namespaceRepo.Create(context.Background(), "testnamespace", false)
		require.NoError(t, err)

		moduleProvider := model.NewModuleProvider(namespace, "testmodule", "aws")
		require.NoError(t, err)
		err = moduleProviderRepo.Save(context.Background(), moduleProvider)
		require.NoError(t, err)

		// Create module version
		version, err := shared.ParseVersion("1.0.0")
		require.NoError(t, err)

		details := &model.ModuleDetails{
			Owner:       "testowner",
			Description: "Test module",
		}

		moduleVersion, err := model.NewModuleVersion(version.String(), details, false)
		require.NoError(t, err)
		moduleVersion.SetModuleProvider(moduleProvider)
		err = moduleVersionRepo.Save(context.Background(), moduleVersion)
		require.NoError(t, err)

		// Create handler
		getModuleProviderQuery := query.NewGetModuleProviderQuery(moduleProviderRepo)
		getModuleVersionQuery := query.NewGetModuleVersionQuery(moduleVersionRepo)
		handler := moduleHandler.NewModuleHandler(
			nil, // listModulesQuery
			nil, // searchModulesQuery
			getModuleProviderQuery,
			nil, // listModuleProvidersQuery
			getModuleVersionQuery,
			nil, // getModuleDownloadQuery
			nil, // getModuleProviderSettingsQuery
			nil, // getReadmeHTMLQuery
			nil, // getSubmodulesQuery
			nil, // getExamplesQuery
			nil, // getIntegrationsQuery
			nil, // createModuleProviderCmd
			nil, // publishModuleVersionCmd
			nil, // updateModuleProviderSettingsCmd
			nil, // deleteModuleProviderCmd
			nil, // uploadModuleVersionCmd
			nil, // importModuleVersionCmd
			nil, // getModuleVersionFileCmd
			nil, // deleteModuleVersionCmd
			nil, // generateModuleSourceCmd
			nil, // getVariableTemplateQuery
			nil, // createModuleProviderRedirectCmd
			nil, // deleteModuleProviderRedirectCmd
			nil, // getModuleProviderRedirectsQuery
			domainConfig,
			nil, // namespaceService
			nil, // analyticsRepo
		)

		// Test Terrareg API endpoint
		req := httptest.NewRequest("GET", "/v1/terrareg/modules/testnamespace/testmodule/aws", nil)
		w := httptest.NewRecorder()

		handler.HandleTerraregModuleProviderDetails(w, req)
		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		// Verify Terrareg API format
		assert.Equal(t, "testnamespace", response["namespace"])
		assert.Equal(t, "testmodule", response["name"])
		assert.Equal(t, "aws", response["provider"])
		assert.Equal(t, "testowner", response["owner"])
		assert.Equal(t, "Test module", response["description"])
		assert.Equal(t, "1.0.0", response["version"])
		assert.Contains(t, response, "versions")
	})

	t.Run("Module Provider Redirects API", func(t *testing.T) {
		// Create test data
		namespace1, err := namespaceRepo.Create(context.Background(), "fromnamespace", false)
		require.NoError(t, err)

		namespace2, err := namespaceRepo.Create(context.Background(), "tonamespace", false)
		require.NoError(t, err)

		fromProvider := model.NewModuleProvider(namespace1, "frommodule", "aws")
		require.NoError(t, err)
		err = moduleProviderRepo.Save(context.Background(), fromProvider)
		require.NoError(t, err)

		toProvider := model.NewModuleProvider(namespace2, "tomodule", "aws")
		require.NoError(t, err)
		err = moduleProviderRepo.Save(context.Background(), toProvider)
		require.NoError(t, err)

		// Create redirect
		redirect := &model.ModuleProviderRedirect{
			FromNamespace:      namespace1.Name(),
			FromModule:         fromProvider.Module(),
			FromProvider:       fromProvider.Provider(),
			ToModuleProviderID: toProvider.ID(),
		}

		err = moduleProviderRedirectRepo.Create(context.Background(), redirect)
		require.NoError(t, err)

		// Create handler
		getModuleProviderRedirectsQuery := query.NewGetModuleProviderRedirectsQuery(moduleProviderRedirectRepo)
		handler := moduleHandler.NewModuleHandler(
			nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil,
			nil, nil, nil, nil, nil, nil, nil, nil,
			nil, nil, nil,
			domainConfig,
			nil,
			nil,
		)

		// Test get all redirects
		req := httptest.NewRequest("GET", "/v1/terrareg/modules/redirects", nil)
		w := httptest.NewRecorder()

		handler.HandleModuleProviderRedirectsGet(w, req)
		assert.Equal(t, http.StatusOK, w.Code)

		var response []map[string]interface{}
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Len(t, response, 1)

		redirectResp := response[0]
		assert.Equal(t, namespace1.Name(), redirectResp["from_namespace"])
		assert.Equal(t, fromProvider.Module(), redirectResp["from_module"])
		assert.Equal(t, fromProvider.Provider(), redirectResp["from_provider"])
		assert.Equal(t, toProvider.ID(), redirectResp["to_module_provider_id"])
	})

	t.Run("Graph API - Get Graph Data", func(t *testing.T) {
		// Create graph handler
		handler := moduleHandler.NewGraphHandler()

		// Test graph data endpoint
		req := httptest.NewRequest("GET", "/v1/terrareg/graph/data?include-beta=true&namespace=test", nil)
		w := httptest.NewRecorder()

		handler.HandleGraphDataGet(w, req)
		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		// Verify graph API format
		assert.Contains(t, response, "nodes")
		assert.Contains(t, response, "edges")
		assert.Contains(t, response, "metadata")

		nodes := response["nodes"].([]interface{})
		assert.NotEmpty(t, nodes)

		metadata := response["metadata"].(map[string]interface{})
		assert.Equal(t, true, metadata["include_beta"])
		assert.Equal(t, "test", metadata["namespace"])
	})

	t.Run("Graph API - Module Dependency Graph", func(t *testing.T) {
		// Create graph handler
		handler := moduleHandler.NewGraphHandler()

		// Test module dependency graph endpoint
		req := httptest.NewRequest(
			"GET",
			"/v1/terrareg/modules/testnamespace/testmodule/aws/1.0.0/graph?include-beta=true&include-optional=true",
			nil,
		)
		w := httptest.NewRecorder()

		handler.HandleModuleDependencyGraph(w, req)
		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		// Verify dependency graph format
		assert.Contains(t, response, "module")
		assert.Contains(t, response, "dependencies")
		assert.Contains(t, response, "modules")
		assert.Contains(t, response, "metadata")

		module := response["module"].(map[string]interface{})
		assert.Equal(t, "testnamespace", module["namespace"])
		assert.Equal(t, "testmodule", module["name"])
		assert.Equal(t, "aws", module["provider"])
		assert.Equal(t, "1.0.0", module["version"])

		metadata := response["metadata"].(map[string]interface{})
		assert.Equal(t, true, metadata["include_beta"])
		assert.Equal(t, true, metadata["include_optional"])
	})

	t.Run("Graph API - Export Graph", func(t *testing.T) {
		// Create graph handler
		handler := moduleHandler.NewGraphHandler()

		// Test export in DOT format
		req := httptest.NewRequest("GET", "/v1/terrareg/graph/export?format=dot", nil)
		w := httptest.NewRecorder()

		handler.HandleGraphExport(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "text/plain", w.Header().Get("Content-Type"))
		assert.Contains(t, w.Body.String(), "digraph module_dependencies")

		// Test export in JSON format
		req = httptest.NewRequest("GET", "/v1/terrareg/graph/export?format=json", nil)
		w = httptest.NewRecorder()

		handler.HandleGraphExport(w, req)
		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Contains(t, response, "nodes")

		// Test export in SVG format
		req = httptest.NewRequest("GET", "/v1/terrareg/graph/export?format=svg", nil)
		w = httptest.NewRecorder()

		handler.HandleGraphExport(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "image/svg+xml", w.Header().Get("Content-Type"))
		assert.Contains(t, w.Body.String(), "<svg")

		// Test invalid format
		req = httptest.NewRequest("GET", "/v1/terrareg/graph/export?format=invalid", nil)
		w = httptest.NewRecorder()

		handler.HandleGraphExport(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}
