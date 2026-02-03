package terrareg

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/matthewjohn/terrareg/terrareg-go/test/integration/testutils"
)

// TestReadEndpoints_AllAuthMethods tests that read access endpoints work correctly
// with all authentication methods and both ALLOW_UNAUTHENTICATED_ACCESS states
func TestReadEndpoints_AllAuthMethods(t *testing.T) {
	// Setup: Create test data
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	// Create a test namespace with a published module
	namespace := testutils.CreateNamespace(t, db, "test-ns", nil)
	moduleProvider := testutils.CreateModuleProvider(t, db, namespace.ID, "test-mod", "test-prov")
	_ = testutils.CreatePublishedModuleVersion(t, db, moduleProvider.ID, "1.0.0")

	cont := testutils.CreateTestServer(t, db)
	authHelper := testutils.NewAuthHelper(t, db, cont)

	endpoints := []struct {
		name   string
		method string
		path   string
	}{
		{
			name:   "module_list",
			method: "GET",
			path:   "/v1/modules",
		},
		{
			name:   "module_versions",
			method: "GET",
			path:   "/v1/modules/test-ns/test-mod/test-prov/versions",
		},
		{
			name:   "is_authenticated",
			method: "GET",
			path:   "/v1/terrareg/auth/admin/is_authenticated",
		},
	}

	authMethods := []struct {
		name          string
		setup         func(t *testing.T, req *http.Request)
		expectsUnauth bool // true = expects unauth to work when config=true
	}{
		{
			name: "unauthenticated",
			setup: func(t *testing.T, req *http.Request) {
				// No authentication
			},
			expectsUnauth: true,
		},
		{
			name: "admin_api_key",
			setup: func(t *testing.T, req *http.Request) {
				apiKey := os.Getenv("ADMIN_AUTH_TOKEN")
				if apiKey == "" {
					apiKey = "test-admin-api-key"
				}
				req.Header.Set("X-Terrareg-ApiKey", apiKey)
			},
			expectsUnauth: false,
		},
		{
			name: "upload_api_key",
			setup: func(t *testing.T, req *http.Request) {
				apiKey := os.Getenv("UPLOAD_AUTH_TOKEN")
				if apiKey == "" {
					apiKey = "test-upload-key"
				}
				req.Header.Set("X-Terrareg-UploadKey", apiKey)
			},
			expectsUnauth: false,
		},
		{
			name: "publish_api_key",
			setup: func(t *testing.T, req *http.Request) {
				apiKey := os.Getenv("PUBLISH_AUTH_TOKEN")
				if apiKey == "" {
					apiKey = "test-publish-key"
				}
				req.Header.Set("X-Terrareg-PublishKey", apiKey)
			},
			expectsUnauth: false,
		},
		{
			name: "user_session",
			setup: func(t *testing.T, req *http.Request) {
				cookie := authHelper.CreateSessionForUser("testuser", false, []string{}, nil)
				req.Header.Set("Cookie", cookie)
			},
			expectsUnauth: false,
		},
		{
			name: "user_with_read_permission",
			setup: func(t *testing.T, req *http.Request) {
				// Create a group with READ permission for test-ns
				authHelper.SetupUserGroupWithPermissions("read-group", false, map[string]string{"test-ns": "READ"})
				cookie := authHelper.CreateSessionForUser("readuser", false, []string{"read-group"}, nil)
				req.Header.Set("Cookie", cookie)
			},
			expectsUnauth: false,
		},
		{
			name: "terraform_idp_token",
			setup: func(t *testing.T, req *http.Request) {
				token := authHelper.CreateTerraformIDPToken("test-subject", nil)
				req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
			},
			expectsUnauth: false,
		},
	}

	for _, endpoint := range endpoints {
		t.Run(endpoint.name, func(t *testing.T) {
			for _, authMethod := range authMethods {
				t.Run(authMethod.name, func(t *testing.T) {
					// Test with ALLOW_UNAUTHENTICATED_ACCESS=true
					t.Run("ALLOW_UNAUTHENTICATED_ACCESS=true", func(t *testing.T) {
						db := testutils.SetupTestDatabase(t)
						defer testutils.CleanupTestDatabase(t, db)

						// Recreate test data for this test
						namespace := testutils.CreateNamespace(t, db, "test-ns", nil)
						moduleProvider := testutils.CreateModuleProvider(t, db, namespace.ID, "test-mod", "test-prov")
						_ = testutils.CreatePublishedModuleVersion(t, db, moduleProvider.ID, "1.0.0")

						cont := testutils.CreateTestContainerWithConfig(t, db,
							testutils.WithAllowUnauthenticatedAccess(true))
						router := cont.Server.Router()

						req := httptest.NewRequest(endpoint.method, endpoint.path, nil)
						authMethod.setup(t, req)

						w := httptest.NewRecorder()
						router.ServeHTTP(w, req)

						// With ALLOW_UNAUTHENTICATED_ACCESS=true, all auth methods should work
						assert.Equal(t, http.StatusOK, w.Code,
							fmt.Sprintf("Endpoint %s with auth %s and ALLOW_UNAUTHENTICATED_ACCESS=true should return 200",
								endpoint.path, authMethod.name))
					})

					// Test with ALLOW_UNAUTHENTICATED_ACCESS=false
					t.Run("ALLOW_UNAUTHENTICATED_ACCESS=false", func(t *testing.T) {
						db := testutils.SetupTestDatabase(t)
						defer testutils.CleanupTestDatabase(t, db)

						// Recreate test data for this test
						namespace := testutils.CreateNamespace(t, db, "test-ns", nil)
						moduleProvider := testutils.CreateModuleProvider(t, db, namespace.ID, "test-mod", "test-prov")
						_ = testutils.CreatePublishedModuleVersion(t, db, moduleProvider.ID, "1.0.0")

						cont := testutils.CreateTestContainerWithConfig(t, db,
							testutils.WithAllowUnauthenticatedAccess(false))
						router := cont.Server.Router()

						req := httptest.NewRequest(endpoint.method, endpoint.path, nil)
						authMethod.setup(t, req)

						w := httptest.NewRecorder()
						router.ServeHTTP(w, req)

						if authMethod.name == "unauthenticated" {
							// Unauthenticated requests should fail with 401
							assert.Equal(t, http.StatusUnauthorized, w.Code,
								fmt.Sprintf("Endpoint %s with no auth and ALLOW_UNAUTHENTICATED_ACCESS=false should return 401",
									endpoint.path))
						} else {
							// All authenticated methods should work
							assert.Equal(t, http.StatusOK, w.Code,
								fmt.Sprintf("Endpoint %s with auth %s should return 200",
									endpoint.path, authMethod.name))
						}
					})
				})
			}
		})
	}
}

// TestReadEndpoints_IsAuthenticatedResponseStructure verifies the is_authenticated response structure
func TestReadEndpoints_IsAuthenticatedResponseStructure(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	cont := testutils.CreateTestServer(t, db)
	authHelper := testutils.NewAuthHelper(t, db, cont)

	tests := []struct {
		name                 string
		allowUnauthenticated bool
		setup                func(t *testing.T, req *http.Request)
		expectAuthenticated  bool
		expectReadAccess     bool
		expectSiteAdmin      bool
	}{
		{
			name:                 "unauthenticated_with_allow_true",
			allowUnauthenticated: true,
			setup:                func(t *testing.T, req *http.Request) {},
			expectAuthenticated:  false,
			expectReadAccess:     true, // read_access is true when ALLOW_UNAUTHENTICATED_ACCESS=true
			expectSiteAdmin:      false,
		},
		{
			name:                 "unauthenticated_with_allow_false",
			allowUnauthenticated: false,
			setup:                func(t *testing.T, req *http.Request) {},
			expectAuthenticated:  false,
			expectReadAccess:     false, // read_access is false when ALLOW_UNAUTHENTICATED_ACCESS=false
			expectSiteAdmin:      false,
		},
		{
			name:                 "admin_api_key",
			allowUnauthenticated: true,
			setup: func(t *testing.T, req *http.Request) {
				apiKey := os.Getenv("ADMIN_AUTH_TOKEN")
				if apiKey == "" {
					apiKey = "test-admin-api-key"
				}
				req.Header.Set("X-Terrareg-ApiKey", apiKey)
			},
			expectAuthenticated: true,
			expectReadAccess:    true,
			expectSiteAdmin:     true,
		},
		{
			name:                 "user_with_read_permission",
			allowUnauthenticated: true,
			setup: func(t *testing.T, req *http.Request) {
				authHelper.SetupUserGroupWithPermissions("read-group", false, map[string]string{"test-ns": "READ"})
				cookie := authHelper.CreateSessionForUser("readuser", false, []string{"read-group"}, nil)
				req.Header.Set("Cookie", cookie)
			},
			expectAuthenticated: true,
			expectReadAccess:    true,
			expectSiteAdmin:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cont := testutils.CreateTestContainerWithConfig(t, db,
				testutils.WithAllowUnauthenticatedAccess(tt.allowUnauthenticated))
			router := cont.Server.Router()

			req := httptest.NewRequest("GET", "/v1/terrareg/auth/admin/is_authenticated", nil)
			tt.setup(t, req)

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)

			assert.Equal(t, tt.expectAuthenticated, response["authenticated"].(bool))
			assert.Equal(t, tt.expectReadAccess, response["read_access"].(bool))
			assert.Equal(t, tt.expectSiteAdmin, response["site_admin"].(bool))
			assert.Contains(t, response, "namespace_permissions")
		})
	}
}

// TestReadEndpoints_NamespacePermissionsInResponse verifies namespace permissions are included in response
func TestReadEndpoints_NamespacePermissionsInResponse(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	// Create test namespaces
	_ = testutils.CreateNamespace(t, db, "ns-full", nil)
	_ = testutils.CreateNamespace(t, db, "ns-modify", nil)
	_ = testutils.CreateNamespace(t, db, "ns-read", nil)

	cont := testutils.CreateTestServer(t, db)
	authHelper := testutils.NewAuthHelper(t, db, cont)

	// Create user group with mixed permissions
	authHelper.SetupUserGroupWithPermissions("mixed-group", false, map[string]string{
		"ns-full":   "FULL",
		"ns-modify": "MODIFY",
		"ns-read":   "READ",
	})

	cookie := authHelper.CreateSessionForUser("mixeduser", false, []string{"mixed-group"}, nil)

	router := cont.Router

	req := httptest.NewRequest("GET", "/v1/terrareg/auth/admin/is_authenticated", nil)
	req.Header.Set("Cookie", cookie)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.True(t, response["authenticated"].(bool))
	assert.True(t, response["read_access"].(bool))
	assert.False(t, response["site_admin"].(bool))

	// Check namespace permissions
	permissions, ok := response["namespace_permissions"].(map[string]interface{})
	require.True(t, ok, "namespace_permissions should be a map")

	assert.Equal(t, "FULL", permissions["ns-full"])
	assert.Equal(t, "MODIFY", permissions["ns-modify"])
	assert.Equal(t, "READ", permissions["ns-read"])
}
