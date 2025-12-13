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
	"github.com/matthewjohn/terrareg/terrareg-go/test/util"
)

func TestSubmoduleIntegration(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer func() {
		require.NoError(t, db.Close())
	}()

	// Setup repositories
	namespaceRepo := sqldb.NewNamespaceRepository(db.DB)
	moduleProviderRepo := sqldb.NewModuleProviderRepository(db.DB)
	moduleVersionRepo := sqldb.NewModuleVersionRepository(db.DB)
	submoduleRepo := sqldb.NewSubmoduleRepository(db.DB)
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

	// Add submodule files
	submoduleContent := map[string][]byte{
		"submodule1/main.tf":     []byte(`resource "null_resource" "example" {}`),
		"submodule1/variables.tf": []byte(`variable "name" { description = "Name" }`),
		"submodule2/main.tf":      []byte(`resource "random_pet" "example" {}`),
	}

	for path, content := range submoduleContent {
		file := &model.ModuleVersionFile{
			Path:    path,
			Content: content,
			Type:    "tf",
		}
		err = moduleVersionFileRepo.Save(context.Background(), moduleVersion, file)
		require.NoError(t, err)
	}

	// Create handlers and queries
	listModulesQuery := query.NewListModulesQuery(moduleProviderRepo)
	getModuleProviderQuery := query.NewGetModuleProviderQuery(moduleProviderRepo)
	getSubmodulesQuery := query.NewGetSubmodulesQuery(submoduleRepo)

	handler := moduleHandler.NewModuleHandler(
		listModulesQuery,
		nil, // searchModulesQuery
		getModuleProviderQuery,
		nil, // listModuleProvidersQuery
		nil, // getModuleVersionQuery
		nil, // getModuleDownloadQuery
		nil, // getModuleProviderSettingsQuery
		nil, // getReadmeHTMLQuery
		getSubmodulesQuery,
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
		nil, // domainConfig
		nil, // namespaceService
		nil, // analyticsRepo
	)

	t.Run("Get submodules", func(t *testing.T) {
		// Create request
		req := httptest.NewRequest(
			"GET",
			"/v1/terrareg/modules/testnamespace/testmodule/aws/1.0.0/submodules",
			nil,
		)
		w := httptest.NewRecorder()

		// Handle request
		handler.HandleGetSubmodules(w, req)

		// Check response
		assert.Equal(t, http.StatusOK, w.Code)

		var response []map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		// Verify submodules
		assert.Len(t, response, 2)

		// Check first submodule
		submodule1 := findSubmodule(response, "submodule1")
		require.NotNil(t, submodule1)
		assert.Equal(t, "submodule1", submodule1["path"])
		assert.Contains(t, submodule1, "terraform_modules")

		// Check second submodule
		submodule2 := findSubmodule(response, "submodule2")
		require.NotNil(t, submodule2)
		assert.Equal(t, "submodule2", submodule2["path"])
		assert.Contains(t, submodule2, "terraform_modules")
	})

	t.Run("Get submodules for non-existent module", func(t *testing.T) {
		// Create request for non-existent module
		req := httptest.NewRequest(
			"GET",
			"/v1/terrareg/modules/nonexistent/module/aws/1.0.0/submodules",
			nil,
		)
		w := httptest.NewRecorder()

		// Handle request
		handler.HandleGetSubmodules(w, req)

		// Should return error
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("Get submodules with missing path parameters", func(t *testing.T) {
		// Create request without version
		req := httptest.NewRequest(
			"GET",
			"/v1/terrareg/modules/testnamespace/testmodule/aws/submodules",
			nil,
		)
		w := httptest.NewRecorder()

		// Handle request
		handler.HandleGetSubmodules(w, req)

		// Should return bad request
		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Contains(t, response["error"], "Missing required path parameters")
	})
}

func findSubmodule(submodules []map[string]interface{}, path string) map[string]interface{} {
	for _, submodule := range submodules {
		if p, ok := submodule["path"].(string); ok && p == path {
			return submodule
		}
	}
	return nil
}