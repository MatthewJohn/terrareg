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
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/module/model"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/shared"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/persistence/sqldb"
	moduleHandler "github.com/matthewjohn/terrareg/terrareg-go/internal/interfaces/http/handler/terrareg"
	"github.com/matthewjohn/terrareg/terrareg-go/test/integration/testutils"
)

func TestExampleIntegration(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer func() {
		require.NoError(t, db.Close())
	}()

	// Setup repositories
	namespaceRepo := sqldb.NewNamespaceRepository(db.DB)
	moduleProviderRepo := sqldb.NewModuleProviderRepository(db.DB)
	moduleVersionRepo := sqldb.NewModuleVersionRepository(db.DB)
	exampleRepo := sqldb.NewExampleFileRepository(db.DB)
	moduleVersionFileRepo := sqldb.NewModuleVersionFileRepository(db.DB)

	// Create namespace
	namespace, err := namespaceRepo.Create(context.Background(), "testnamespace", false)
	require.NoError(t, err)

	// Create module provider
	moduleProvider := model.NewModuleProvider(namespace, "testmodule", "aws")
	require.NoError(t, err)

	// Save module provider to get ID
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

	// Save module version
	err = moduleVersionRepo.Save(context.Background(), moduleVersion)
	require.NoError(t, err)

	// Add example files
	exampleFiles := map[string]map[string][]byte{
		"basic": {
			"main.tf":      []byte(`module "test" { source = "./../" }`),
			"variables.tf": []byte(`variable "region" { description = "AWS region" }`),
			"outputs.tf":   []byte(`output "id" { value = "example" }`),
		},
		"advanced": {
			"main.tf":     []byte(`module "test_advanced" { source = "./../" count = 2 }`),
			"README.md":   []byte("# Advanced Example\n\nThis shows advanced usage."),
		},
		"minimal": {
			"main.tf": []byte(`module "minimal" { source = "./../" }`),
		},
	}

	for exampleName, files := range exampleFiles {
		for path, content := range files {
			exampleFile := &model.ExampleFile{
				Path:    path,
				Content: content,
			}
			err = exampleRepo.Save(context.Background(), moduleVersion, exampleName, exampleFile)
			require.NoError(t, err)
		}
	}

	// Create handlers and queries
	listModulesQuery := query.NewListModulesQuery(moduleProviderRepo)
	getModuleProviderQuery := query.NewGetModuleProviderQuery(moduleProviderRepo)
	getExamplesQuery := query.NewGetExamplesQuery(exampleRepo)

	handler := moduleHandler.NewModuleHandler(
		listModulesQuery,
		nil, // searchModulesQuery
		getModuleProviderQuery,
		nil, // listModuleProvidersQuery
		nil, // getModuleVersionQuery
		nil, // getModuleDownloadQuery
		nil, // getModuleProviderSettingsQuery
		nil, // getReadmeHTMLQuery
		nil, // getSubmodulesQuery
		getExamplesQuery,
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
		nil, // domainConfig
		nil, // namespaceService
		nil, // analyticsRepo
	)

	t.Run("Get examples", func(t *testing.T) {
		// Create request
		req := httptest.NewRequest(
			"GET",
			"/v1/terrareg/modules/testnamespace/testmodule/aws/1.0.0/examples",
			nil,
		)
		w := httptest.NewRecorder()

		// Handle request
		handler.HandleGetExamples(w, req)

		// Check response
		assert.Equal(t, http.StatusOK, w.Code)

		var response []map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		// Verify examples
		assert.Len(t, response, 3)

		// Check basic example
		basic := findExample(response, "basic")
		require.NotNil(t, basic)
		assert.Equal(t, "basic", basic["path"])
		assert.Contains(t, basic, "files")

		files := basic["files"].([]interface{})
		assert.Len(t, files, 3) // main.tf, variables.tf, outputs.tf

		// Check advanced example
		advanced := findExample(response, "advanced")
		require.NotNil(t, advanced)
		assert.Equal(t, "advanced", advanced["path"])

		advancedFiles := advanced["files"].([]interface{})
		assert.Len(t, advancedFiles, 2) // main.tf, README.md

		// Check minimal example
		minimal := findExample(response, "minimal")
		require.NotNil(t, minimal)
		assert.Equal(t, "minimal", minimal["path"])

		minimalFiles := minimal["files"].([]interface{})
		assert.Len(t, minimalFiles, 1) // main.tf
	})

	t.Run("Get examples for non-existent module", func(t *testing.T) {
		// Create request for non-existent module
		req := httptest.NewRequest(
			"GET",
			"/v1/terrareg/modules/nonexistent/module/aws/1.0.0/examples",
			nil,
		)
		w := httptest.NewRecorder()

		// Handle request
		handler.HandleGetExamples(w, req)

		// Should return error
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("Get examples with missing path parameters", func(t *testing.T) {
		// Create request without version
		req := httptest.NewRequest(
			"GET",
			"/v1/terrareg/modules/testnamespace/testmodule/aws/examples",
			nil,
		)
		w := httptest.NewRecorder()

		// Handle request
		handler.HandleGetExamples(w, req)

		// Should return bad request
		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Contains(t, response["error"], "Missing required path parameters")
	})

	t.Run("Get examples for module with no examples", func(t *testing.T) {
		// Create another module version without examples
		newVersion, err := shared.ParseVersion("2.0.0")
		require.NoError(t, err)

		moduleVersion2, err := model.NewModuleVersion(newVersion.String(), details, false)
		require.NoError(t, err)
		moduleVersion2.SetModuleProvider(moduleProvider)

		// Save module version
		err = moduleVersionRepo.Save(context.Background(), moduleVersion2)
		require.NoError(t, err)

		// Create request
		req := httptest.NewRequest(
			"GET",
			"/v1/terrareg/modules/testnamespace/testmodule/aws/2.0.0/examples",
			nil,
		)
		w := httptest.NewRecorder()

		// Handle request
		handler.HandleGetExamples(w, req)

		// Check response - should return empty array
		assert.Equal(t, http.StatusOK, w.Code)

		var response []map[string]interface{}
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Len(t, response, 0)
	})
}

func findExample(examples []map[string]interface{}, path string) map[string]interface{} {
	for _, example := range examples {
		if p, ok := example["path"].(string); ok && p == path {
			return example
		}
	}
	return nil
}