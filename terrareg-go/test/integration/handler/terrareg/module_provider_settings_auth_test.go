//go:build integration
// +build integration

// Python reference: /app/test/unit/terrareg/server/test_api_terrareg_module_provider_settings.py

package terrareg_test

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/provider_source/model"
	providerSourceRepo "github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/persistence/sqldb/provider_source"
	sqldb "github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/persistence/sqldb"
	"github.com/matthewjohn/terrareg/terrareg-go/test/integration/testutils"
)

// TestModuleProviderSettings_Authentication tests authentication for module provider settings endpoint
func TestModuleProviderSettings_Authentication(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	// Create test data
	namespace := testutils.CreateNamespace(t, db, "settings-namespace", nil)
	_ = testutils.CreateModuleProvider(t, db, namespace.ID, "testmodule", "testprovider")

	// Enable RBAC for this test so permission checking works properly
	cont := testutils.CreateTestServerWithConfig(t, db, testutils.WithEnableAccessControls(true))
	router := cont.Router

	tests := []struct {
		name           string
		setupAuth      func(*testing.T, *sqldb.Database) *http.Request
		expectedStatus int
	}{
		{
			name: "unauthenticated PUT request returns 401",
			setupAuth: func(t *testing.T, db *sqldb.Database) *http.Request {
				req := testutils.BuildUnauthenticatedRequest(t, "PUT", "/v1/terrareg/modules/settings-namespace/testmodule/testprovider/settings")
				return testutils.AddChiContext(t, req, map[string]string{
					"namespace": "settings-namespace",
					"name":      "testmodule",
					"provider":  "testprovider",
				})
			},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name: "authenticated user without permission gets 403",
			setupAuth: func(t *testing.T, db *sqldb.Database) *http.Request {
				req, _ := testutils.BuildAuthenticatedRequestWithSession(t, db, "PUT", "/v1/terrareg/modules/settings-namespace/testmodule/testprovider/settings", "regular-user", false)
				return testutils.AddChiContext(t, req, map[string]string{
					"namespace": "settings-namespace",
					"name":      "testmodule",
					"provider":  "testprovider",
				})
			},
			expectedStatus: http.StatusForbidden,
		},
		{
			name: "user with MODIFY permission can update settings",
			setupAuth: func(t *testing.T, db *sqldb.Database) *http.Request {
				req, _ := testutils.BuildAuthenticatedRequestWithNamespacePermission(
					t, db, "PUT", "/v1/terrareg/modules/settings-namespace/testmodule/testprovider/settings",
					"modify-user", "settings-namespace", sqldb.PermissionTypeModify,
				)
				return testutils.AddChiContext(t, req, map[string]string{
					"namespace": "settings-namespace",
					"name":      "testmodule",
					"provider":  "testprovider",
				})
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "admin user can update any settings",
			setupAuth: func(t *testing.T, db *sqldb.Database) *http.Request {
				req, _ := testutils.BuildAdminRequest(t, db, "PUT", "/v1/terrareg/modules/settings-namespace/testmodule/testprovider/settings")
				return testutils.AddChiContext(t, req, map[string]string{
					"namespace": "settings-namespace",
					"name":      "testmodule",
					"provider":  "testprovider",
				})
			},
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := tt.setupAuth(t, db)
			// Add minimal request body with CSRF token
			requestBody := map[string]interface{}{
				"csrf_token": "test-token",
			}
			bodyBytes, _ := json.Marshal(requestBody)
			req.Body = io.NopCloser(strings.NewReader(string(bodyBytes)))
			req.Header.Set("Content-Type", "application/json")
			w := testutils.ServeHTTP(router, req)
			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

// TestModuleProviderSettings_SetProviderSourceHTTP tests setting a provider source via HTTP
func TestModuleProviderSettings_SetProviderSourceHTTP(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	// Create test data
	namespace := testutils.CreateNamespace(t, db, "test-ns-set-ps", nil)
	moduleProvider := testutils.CreateModuleProvider(t, db, namespace.ID, "testmodule", "testprovider")

	// Create test provider source
	providerSource := model.NewProviderSource(
		"test-ps-set-mp",
		"test-ps-set-mp",
		model.ProviderSourceTypeGithub,
		&model.ProviderSourceConfig{
			BaseURL:         "https://github.com",
			ApiURL:          "https://api.github.com",
			ClientID:        "test-client-id",
			ClientSecret:    "test-client-secret",
			PrivateKeyPath:  "/test/key.pem",
			AppID:           "test-app-id",
			LoginButtonText: "Login with GitHub",
		},
	)
	psRepo := providerSourceRepo.NewProviderSourceRepository(db.DB)
	err := psRepo.Upsert(testutils.GetTestContext(t), providerSource)
	require.NoError(t, err)

	// Enable RBAC
	cont := testutils.CreateTestServerWithConfig(t, db, testutils.WithEnableAccessControls(true))
	router := cont.Router

	// Create authenticated admin request
	req, _ := testutils.BuildAdminRequest(t, db, "PUT", "/v1/terrareg/modules/test-ns-set-ps/testmodule/testprovider/settings")
	req = testutils.AddChiContext(t, req, map[string]string{
		"namespace": "test-ns-set-ps",
		"name":      "testmodule",
		"provider":  "testprovider",
	})

	// Set request body with provider source
	requestBody := map[string]interface{}{
		"csrf_token":       "test-token",
		"provider_source":  "test-ps-set-mp",
	}
	bodyBytes, _ := json.Marshal(requestBody)
	req.Body = io.NopCloser(strings.NewReader(string(bodyBytes)))
	req.Header.Set("Content-Type", "application/json")

	// Act
	w := testutils.ServeHTTP(router, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	// Verify provider_source was set in response
	assert.Contains(t, response, "provider_source")
	assert.Equal(t, "test-ps-set-mp", response["provider_source"])

	// Verify it was actually set in the database
	var mpDB testutils.ModuleProviderDB
	err = db.DB.First(&mpDB, moduleProvider.ID).Error
	require.NoError(t, err)
	require.NotNil(t, mpDB.ProviderSourceName)
	assert.Equal(t, "test-ps-set-mp", *mpDB.ProviderSourceName)
}

// TestModuleProviderSettings_UnsetProviderSourceHTTP tests unsetting a provider source via HTTP
func TestModuleProviderSettings_UnsetProviderSourceHTTP(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	// Create test data
	namespace := testutils.CreateNamespace(t, db, "test-ns-unset-ps", nil)
	moduleProvider := testutils.CreateModuleProvider(t, db, namespace.ID, "testmodule", "testprovider")

	// Create and set test provider source
	providerSource := model.NewProviderSource(
		"test-ps-unset",
		"test-ps-unset",
		model.ProviderSourceTypeGithub,
		&model.ProviderSourceConfig{
			BaseURL:         "https://github.com",
			ApiURL:          "https://api.github.com",
			ClientID:        "test-client-id",
			ClientSecret:    "test-client-secret",
			PrivateKeyPath:  "/test/key.pem",
			AppID:           "test-app-id",
			LoginButtonText: "Login with GitHub",
		},
	)
	psRepo := providerSourceRepo.NewProviderSourceRepository(db.DB)
	err := psRepo.Upsert(testutils.GetTestContext(t), providerSource)
	require.NoError(t, err)

	// Set provider source in database
	db.DB.Model(&testutils.ModuleProviderDB{}).Where("id = ?", moduleProvider.ID).Update("provider_source_name", providerSource.Name())

	// Enable RBAC
	cont := testutils.CreateTestServerWithConfig(t, db, testutils.WithEnableAccessControls(true))
	router := cont.Router

	// Create authenticated admin request
	req, _ := testutils.BuildAdminRequest(t, db, "PUT", "/v1/terrareg/modules/test-ns-unset-ps/testmodule/testprovider/settings")
	req = testutils.AddChiContext(t, req, map[string]string{
		"namespace": "test-ns-unset-ps",
		"name":      "testmodule",
		"provider":  "testprovider",
	})

	// Set request body with empty string to clear provider source
	requestBody := map[string]interface{}{
		"csrf_token":      "test-token",
		"provider_source": "",
	}
	bodyBytes, _ := json.Marshal(requestBody)
	req.Body = io.NopCloser(strings.NewReader(string(bodyBytes)))
	req.Header.Set("Content-Type", "application/json")

	// Act
	w := testutils.ServeHTTP(router, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	// Verify provider_source is null in response
	assert.Contains(t, response, "provider_source")
	assert.Nil(t, response["provider_source"])

	// Verify it was actually unset in the database
	var mpDB testutils.ModuleProviderDB
	err = db.DB.First(&mpDB, moduleProvider.ID).Error
	require.NoError(t, err)
	assert.Nil(t, mpDB.ProviderSourceName)
}

// TestModuleProviderSettings_SetInheritanceDisabledHTTP tests setting inheritance disabled via HTTP
func TestModuleProviderSettings_SetInheritanceDisabledHTTP(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	// Create test data
	namespace := testutils.CreateNamespace(t, db, "test-ns-inherit", nil)
	moduleProvider := testutils.CreateModuleProvider(t, db, namespace.ID, "testmodule", "testprovider")

	// Enable RBAC
	cont := testutils.CreateTestServerWithConfig(t, db, testutils.WithEnableAccessControls(true))
	router := cont.Router

	// Create authenticated admin request
	req, _ := testutils.BuildAdminRequest(t, db, "PUT", "/v1/terrareg/modules/test-ns-inherit/testmodule/testprovider/settings")
	req = testutils.AddChiContext(t, req, map[string]string{
		"namespace": "test-ns-inherit",
		"name":      "testmodule",
		"provider":  "testprovider",
	})

	// Set request body with inheritance disabled
	requestBody := map[string]interface{}{
		"csrf_token":                               "test-token",
		"provider_source_inheritance_disabled": true,
	}
	bodyBytes, _ := json.Marshal(requestBody)
	req.Body = io.NopCloser(strings.NewReader(string(bodyBytes)))
	req.Header.Set("Content-Type", "application/json")

	// Act
	w := testutils.ServeHTTP(router, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	// Verify provider_source_inheritance_disabled was set in response
	assert.Contains(t, response, "provider_source_inheritance_disabled")
	assert.True(t, response["provider_source_inheritance_disabled"].(bool))

	// Verify it was actually set in the database
	var mpDB testutils.ModuleProviderDB
	err = db.DB.First(&mpDB, moduleProvider.ID).Error
	require.NoError(t, err)
	assert.True(t, mpDB.ProviderSourceInheritanceDisabled)
}

// TestModuleProviderSettings_InvalidProviderSourceHTTP tests error handling for invalid provider source via HTTP
func TestModuleProviderSettings_InvalidProviderSourceHTTP(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	// Create test data
	namespace := testutils.CreateNamespace(t, db, "test-ns-invalid-ps", nil)
	_ = testutils.CreateModuleProvider(t, db, namespace.ID, "testmodule", "testprovider")

	// Enable RBAC
	cont := testutils.CreateTestServerWithConfig(t, db, testutils.WithEnableAccessControls(true))
	router := cont.Router

	// Create authenticated admin request
	req, _ := testutils.BuildAdminRequest(t, db, "PUT", "/v1/terrareg/modules/test-ns-invalid-ps/testmodule/testprovider/settings")
	req = testutils.AddChiContext(t, req, map[string]string{
		"namespace": "test-ns-invalid-ps",
		"name":      "testmodule",
		"provider":  "testprovider",
	})

	// Set request body with non-existent provider source
	requestBody := map[string]interface{}{
		"csrf_token":      "test-token",
		"provider_source": "non-existent-provider-source",
	}
	bodyBytes, _ := json.Marshal(requestBody)
	req.Body = io.NopCloser(strings.NewReader(string(bodyBytes)))
	req.Header.Set("Content-Type", "application/json")

	// Act
	w := testutils.ServeHTTP(router, req)

	// Assert - should get 400 error for invalid provider source
	assert.Equal(t, http.StatusBadRequest, w.Code)
}
