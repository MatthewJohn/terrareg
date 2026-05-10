//go:build integration
// +build integration

// Python reference: /app/test/unit/terrareg/server/test_api_terrareg_namespace_details.py

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

// TestNamespaceDefaultProviderSource_Authentication tests authentication for namespace default provider source
func TestNamespaceDefaultProviderSource_Authentication(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	// Create test namespace
	_ = testutils.CreateNamespace(t, db, "test-namespace-dps", nil)

	// Enable RBAC for this test so permission checking works properly
	cont := testutils.CreateTestServerWithConfig(t, db, testutils.WithEnableAccessControls(true))
	router := cont.Router

	tests := []struct {
		name           string
		setupAuth      func(*testing.T, *sqldb.Database) *http.Request
		expectedStatus int
	}{
		{
			name: "unauthenticated request returns 401",
			setupAuth: func(t *testing.T, db *sqldb.Database) *http.Request {
				req := testutils.BuildUnauthenticatedRequest(t, "POST", "/v1/terrareg/namespaces/test-namespace-dps")
				return testutils.AddChiContext(t, req, map[string]string{"namespace": "test-namespace-dps"})
			},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name: "user with READ permission returns 403",
			setupAuth: func(t *testing.T, db *sqldb.Database) *http.Request {
				req, _ := testutils.BuildAuthenticatedRequestWithNamespacePermission(
					t, db, "POST", "/v1/terrareg/namespaces/test-namespace-dps",
					"readonly-user", "test-namespace-dps", sqldb.PermissionTypeRead,
				)
				req = testutils.AddChiContext(t, req, map[string]string{"namespace": "test-namespace-dps"})
				req.Body = io.NopCloser(strings.NewReader(`{"csrf_token":"test-token"}`))
				req.Header.Set("Content-Type", "application/json")
				return req
			},
			expectedStatus: http.StatusForbidden,
		},
		{
			name: "user with MODIFY permission returns 403",
			setupAuth: func(t *testing.T, db *sqldb.Database) *http.Request {
				req, _ := testutils.BuildAuthenticatedRequestWithNamespacePermission(
					t, db, "POST", "/v1/terrareg/namespaces/test-namespace-dps",
					"modify-user", "test-namespace-dps", sqldb.PermissionTypeModify,
				)
				req = testutils.AddChiContext(t, req, map[string]string{"namespace": "test-namespace-dps"})
				req.Body = io.NopCloser(strings.NewReader(`{"csrf_token":"test-token"}`))
				req.Header.Set("Content-Type", "application/json")
				return req
			},
			expectedStatus: http.StatusForbidden,
		},
		{
			name: "user with FULL permission can update namespace",
			setupAuth: func(t *testing.T, db *sqldb.Database) *http.Request {
				req, _ := testutils.BuildAuthenticatedRequestWithNamespacePermission(
					t, db, "POST", "/v1/terrareg/namespaces/test-namespace-dps",
					"full-user", "test-namespace-dps", sqldb.PermissionTypeFull,
				)
				req = testutils.AddChiContext(t, req, map[string]string{"namespace": "test-namespace-dps"})
				req.Body = io.NopCloser(strings.NewReader(`{"csrf_token":"test-token"}`))
				req.Header.Set("Content-Type", "application/json")
				return req
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "admin user can update any namespace",
			setupAuth: func(t *testing.T, db *sqldb.Database) *http.Request {
				req, _ := testutils.BuildAdminRequest(t, db, "POST", "/v1/terrareg/namespaces/test-namespace-dps")
				req = testutils.AddChiContext(t, req, map[string]string{"namespace": "test-namespace-dps"})
				req.Body = io.NopCloser(strings.NewReader(`{"csrf_token":"test-token"}`))
				req.Header.Set("Content-Type", "application/json")
				return req
			},
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := tt.setupAuth(t, db)
			w := testutils.ServeHTTP(router, req)
			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

// TestNamespaceDefaultProviderSource_SetValid tests setting a valid default provider source
func TestNamespaceDefaultProviderSource_SetValid(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	// Create test namespace
	namespace := testutils.CreateNamespace(t, db, "test-ns-set-dps", nil)

	// Create test provider source
	providerSource := model.NewProviderSource(
		"test-ps-set-dps",
		"test-ps-set-dps",
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
	req, _ := testutils.BuildAdminRequest(t, db, "POST", "/v1/terrareg/namespaces/test-ns-set-dps")
	req = testutils.AddChiContext(t, req, map[string]string{"namespace": "test-ns-set-dps"})

	// Set request body with default provider source
	requestBody := map[string]interface{}{
		"csrf_token":             "test-token",
		"default_provider_source": "test-ps-set-dps",
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

	// Verify default_provider_source was set in response
	assert.Contains(t, response, "default_provider_source")
	assert.Equal(t, "test-ps-set-dps", response["default_provider_source"])

	// Verify it was actually set in the database
	var nsDB testutils.NamespaceDB
	err = db.DB.First(&nsDB, namespace.ID).Error
	require.NoError(t, err)
	require.NotNil(t, nsDB.DefaultProviderSourceName)
	assert.Equal(t, "test-ps-set-dps", *nsDB.DefaultProviderSourceName)
}

// TestNamespaceDefaultProviderSource_Unset tests unsetting a default provider source
func TestNamespaceDefaultProviderSource_Unset(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	// Create test namespace
	namespace := testutils.CreateNamespace(t, db, "test-ns-unset-dps", nil)

	// Create and set test provider source
	providerSource := model.NewProviderSource(
		"test-ps-unset-dps",
		"test-ps-unset-dps",
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

	// Set default provider source in database
	db.DB.Model(&testutils.NamespaceDB{}).Where("id = ?", namespace.ID).Update("default_provider_source_name", providerSource.Name())

	// Enable RBAC
	cont := testutils.CreateTestServerWithConfig(t, db, testutils.WithEnableAccessControls(true))
	router := cont.Router

	// Create authenticated admin request
	req, _ := testutils.BuildAdminRequest(t, db, "POST", "/v1/terrareg/namespaces/test-ns-unset-dps")
	req = testutils.AddChiContext(t, req, map[string]string{"namespace": "test-ns-unset-dps"})

	// Set request body with empty string to clear default provider source
	requestBody := map[string]interface{}{
		"csrf_token":             "test-token",
		"default_provider_source": "",
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

	// Verify default_provider_source is null in response
	assert.Contains(t, response, "default_provider_source")
	assert.Nil(t, response["default_provider_source"])

	// Verify it was actually unset in the database
	var nsDB testutils.NamespaceDB
	err = db.DB.First(&nsDB, namespace.ID).Error
	require.NoError(t, err)
	assert.Nil(t, nsDB.DefaultProviderSourceName)
}

// TestNamespaceDefaultProviderSource_InvalidProviderSource tests error handling for invalid provider source
func TestNamespaceDefaultProviderSource_InvalidProviderSource(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	// Create test namespace
	_ = testutils.CreateNamespace(t, db, "test-ns-invalid-dps", nil)

	// Enable RBAC
	cont := testutils.CreateTestServerWithConfig(t, db, testutils.WithEnableAccessControls(true))
	router := cont.Router

	// Create authenticated admin request
	req, _ := testutils.BuildAdminRequest(t, db, "POST", "/v1/terrareg/namespaces/test-ns-invalid-dps")
	req = testutils.AddChiContext(t, req, map[string]string{"namespace": "test-ns-invalid-dps"})

	// Set request body with non-existent provider source
	requestBody := map[string]interface{}{
		"csrf_token":             "test-token",
		"default_provider_source": "non-existent-provider-source",
	}
	bodyBytes, _ := json.Marshal(requestBody)
	req.Body = io.NopCloser(strings.NewReader(string(bodyBytes)))
	req.Header.Set("Content-Type", "application/json")

	// Act
	w := testutils.ServeHTTP(router, req)

	// Assert - should get 400 error for invalid provider source
	assert.Equal(t, http.StatusBadRequest, w.Code)
}
