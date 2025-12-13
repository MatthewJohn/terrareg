package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/application/command/module"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/application/query/module"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/container"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/persistence/sqldb"
	testutils "github.com/matthewjohn/terrareg/terrareg-go/test/integration/testutils"
)

func TestModuleProviderRedirectsIntegration(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer func() {
		require.NoError(t, db.Close())
	}()

	// Create test configuration
	domainConfig := testutils.CreateTestDomainConfig(t)
	infraConfig := testutils.CreateTestInfraConfig(t)

	// Create container
	container, err := container.NewContainer(domainConfig, infraConfig, nil, testutils.GetTestLogger(), db)
	require.NoError(t, err)

	// Setup test data
	ctx := context.Background()

	// Create test namespace
	namespace := &sqldb.Namespace{
		Name: "test-namespace",
	}
	err = db.DB.Create(namespace).Error
	require.NoError(t, err)

	// Create source module provider (to redirect from)
	sourceProvider := &sqldb.ModuleProvider{
		NamespaceID: namespace.ID,
		Module:      "old-module",
		Provider:    "old-provider",
		Description: "Source module provider",
	}
	err = db.DB.Create(sourceProvider).Error
	require.NoError(t, err)

	// Create target module provider (to redirect to)
	targetProvider := &sqldb.ModuleProvider{
		NamespaceID: namespace.ID,
		Module:      "new-module",
		Provider:    "new-provider",
		Description: "Target module provider",
	}
	err = db.DB.Create(targetProvider).Error
	require.NoError(t, err)

	server := container.Server
	router := server.GetRouter()

	t.Run("Create Module Provider Redirect", func(t *testing.T) {
		redirectReq := map[string]interface{}{
			"from_namespace":      "test-namespace",
			"from_module":         "old-module",
			"from_provider":       "old-provider",
			"to_module_provider_id": targetProvider.ID,
		}
		reqBody, _ := json.Marshal(redirectReq)

		req := httptest.NewRequest("PUT", "/v1/terrareg/modules/test-namespace/old-module/old-provider/redirect", bytes.NewReader(reqBody))
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "Redirect created successfully", response["message"])
	})

	t.Run("Get All Module Provider Redirects", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/v1/terrareg/modules/redirects", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response []map[string]interface{}
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Len(t, response, 1)

		redirect := response[0]
		assert.Equal(t, "test-namespace", redirect["namespace_name"])
		assert.Equal(t, "old-module", redirect["module"])
		assert.Equal(t, "old-provider", redirect["provider"])
	})

	t.Run("Delete Module Provider Redirect", func(t *testing.T) {
		req := httptest.NewRequest("DELETE", "/v1/terrareg/modules/test-namespace/old-module/old-provider/redirect", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "Redirect deleted successfully", response["message"])
	})

	t.Run("Verify Redirect Deleted", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/v1/terrareg/modules/redirects", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response []map[string]interface{}
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Len(t, response, 0)
	})

	t.Run("Create Redirect with Invalid Target", func(t *testing.T) {
		redirectReq := map[string]interface{}{
			"from_namespace":      "test-namespace",
			"from_module":         "invalid-module",
			"from_provider":       "invalid-provider",
			"to_module_provider_id": 99999, // Non-existent ID
		}
		reqBody, _ := json.Marshal(redirectReq)

		req := httptest.NewRequest("PUT", "/v1/terrareg/modules/test-namespace/invalid-module/invalid-provider/redirect", bytes.NewReader(reqBody))
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)

		var response map[string]interface{}
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Contains(t, response["error"], "failed to create redirect")
	})

	t.Run("Delete Non-existent Redirect", func(t *testing.T) {
		req := httptest.NewRequest("DELETE", "/v1/terrareg/modules/test-namespace/non-existent-module/non-existent-provider/redirect", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		// Should still return 200 even if redirect doesn't exist (idempotent)
		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "Redirect deleted successfully", response["message"])
	})
}

func TestModuleProviderRedirectsDirectAPI(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer func() {
		require.NoError(t, db.Close())
	}()

	// Create test repositories
	namespaceRepo := sqldb.NewNamespaceRepository(db.DB, nil, nil)
	moduleProviderRepo := sqldb.NewModuleProviderRepository(db.DB, namespaceRepo, nil)
	redirectRepo := sqldb.NewModuleProviderRedirectRepository(db.DB)

	// Create commands and queries
	createRedirectCmd := module.NewCreateModuleProviderRedirectCommand(redirectRepo)
	deleteRedirectCmd := module.NewDeleteModuleProviderRedirectCommand(redirectRepo)
	getRedirectsQuery := module.NewGetModuleProviderRedirectsQuery(redirectRepo)

	ctx := context.Background()

	// Create test namespace
	namespace := &sqldb.Namespace{
		Name: "direct-api-namespace",
	}
	err := db.DB.Create(namespace).Error
	require.NoError(t, err)

	// Create source and target module providers
	sourceProvider := &sqldb.ModuleProvider{
		NamespaceID: namespace.ID,
		Module:      "source-module",
		Provider:    "source-provider",
		Description: "Source module provider",
	}
	err = db.DB.Create(sourceProvider).Error
	require.NoError(t, err)

	targetProvider := &sqldb.ModuleProvider{
		NamespaceID: namespace.ID,
		Module:      "target-module",
		Provider:    "target-provider",
		Description: "Target module provider",
	}
	err = db.DB.Create(targetProvider).Error
	require.NoError(t, err)

	t.Run("Create Redirect via Command", func(t *testing.T) {
		req := module.CreateModuleProviderRedirectRequest{
			FromNamespace:      "direct-api-namespace",
			FromModule:         "source-module",
			FromProvider:       "source-provider",
			ToModuleProviderID: targetProvider.ID,
		}

		err = createRedirectCmd.Execute(ctx, req)
		require.NoError(t, err)
	})

	t.Run("Get Redirects via Query", func(t *testing.T) {
		redirects, err := getRedirectsQuery.Execute(ctx)
		require.NoError(t, err)
		assert.Len(t, redirects, 1)

		redirect := redirects[0]
		assert.Equal(t, "direct-api-namespace", redirect.FromNamespace)
		assert.Equal(t, "source-module", redirect.FromModule)
		assert.Equal(t, "source-provider", redirect.FromProvider)
		assert.Equal(t, targetProvider.ID, redirect.ToModuleProviderID)
	})

	t.Run("Get Redirect by From Fields", func(t *testing.T) {
		redirect, err := getRedirectsQuery.ExecuteByFrom(ctx, "direct-api-namespace", "source-module", "source-provider")
		require.NoError(t, err)
		assert.NotNil(t, redirect)

		assert.Equal(t, "direct-api-namespace", redirect.FromNamespace)
		assert.Equal(t, "source-module", redirect.FromModule)
		assert.Equal(t, "source-provider", redirect.FromProvider)
		assert.Equal(t, targetProvider.ID, redirect.ToModuleProviderID)
	})

	t.Run("Delete Redirect via Command", func(t *testing.T) {
		err = deleteRedirectCmd.Execute(ctx, "direct-api-namespace", "source-module", "source-provider")
		require.NoError(t, err)
	})

	t.Run("Verify Redirect Deleted", func(t *testing.T) {
		redirects, err := getRedirectsQuery.Execute(ctx)
		require.NoError(t, err)
		assert.Len(t, redirects, 0)
	})
}

func TestModuleProviderRedirectsEdgeCases(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer func() {
		require.NoError(t, db.Close())
	}()

	namespaceRepo := sqldb.NewNamespaceRepository(db.DB, nil, nil)
	redirectRepo := sqldb.NewModuleProviderRedirectRepository(db.DB)

	createRedirectCmd := module.NewCreateModuleProviderRedirectCommand(redirectRepo)
	getRedirectsQuery := module.NewGetModuleProviderRedirectsQuery(redirectRepo)

	ctx := context.Background()

	// Create test namespace
	namespace := &sqldb.Namespace{
		Name: "edge-case-namespace",
	}
	err := db.DB.Create(namespace).Error
	require.NoError(t, err)

	// Create target module provider
	targetProvider := &sqldb.ModuleProvider{
		NamespaceID: namespace.ID,
		Module:      "target-module",
		Provider:    "target-provider",
		Description: "Target module provider",
	}
	err = db.DB.Create(targetProvider).Error
	require.NoError(t, err)

	t.Run("Duplicate Redirect Creation", func(t *testing.T) {
		req := module.CreateModuleProviderRedirectRequest{
			FromNamespace:      "edge-case-namespace",
			FromModule:         "duplicate-module",
			FromProvider:       "duplicate-provider",
			ToModuleProviderID: targetProvider.ID,
		}

		// Create first redirect
		err = createRedirectCmd.Execute(ctx, req)
		require.NoError(t, err)

		// Try to create duplicate redirect (should update existing)
		err = createRedirectCmd.Execute(ctx, req)
		require.NoError(t, err)

		// Verify only one redirect exists
		redirects, err := getRedirectsQuery.Execute(ctx)
		require.NoError(t, err)
		assert.Len(t, redirects, 1)
	})

	t.Run("Case Insensitive Lookup", func(t *testing.T) {
		// Create redirect with mixed case
		req := module.CreateModuleProviderRedirectRequest{
			FromNamespace:      "Edge-Case-Namespace",
			FromModule:         "Mixed-Case-Module",
			FromProvider:       "Mixed-Case-Provider",
			ToModuleProviderID: targetProvider.ID,
		}

		err = createRedirectCmd.Execute(ctx, req)
		require.NoError(t, err)

		// Try to lookup with different case
		redirect, err := getRedirectsQuery.ExecuteByFrom(ctx, "edge-case-namespace", "mixed-case-module", "mixed-case-provider")
		require.NoError(t, err)
		assert.NotNil(t, redirect)
	})

	t.Run("Special Characters in Names", func(t *testing.T) {
		req := module.CreateModuleProviderRedirectRequest{
			FromNamespace:      "edge-case-namespace",
			FromModule:         "module-with-dashes_and_underscores",
			FromProvider:       "provider-with.dots",
			ToModuleProviderID: targetProvider.ID,
		}

		err = createRedirectCmd.Execute(ctx, req)
		require.NoError(t, err)

		redirects, err := getRedirectsQuery.Execute(ctx)
		require.NoError(t, err)
		assert.Len(t, redirects, 3) // Should have 3 redirects total now

		// Find the redirect with special characters
		var found bool
		for _, r := range redirects {
			if r.FromModule == "module-with-dashes_and_underscores" &&
				r.FromProvider == "provider-with.dots" {
				found = true
				break
			}
		}
		assert.True(t, found, "Should find redirect with special characters")
	})
}