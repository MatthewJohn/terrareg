package terrareg

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/persistence/sqldb"
	"github.com/matthewjohn/terrareg/terrareg-go/test/integration/testutils"
)

// setupModifyTestContainerWithSigningKey creates a test container with Terraform OIDC signing key for terraform_idp tests
func setupModifyTestContainerWithSigningKey(t *testing.T, db *sqldb.Database, authMethodName string) *testutils.TestContainer {
	if strings.HasPrefix(authMethodName, "terraform_idp") {
		keyPath, _ := testutils.CreateTestTerraformOIDCSigningKey(t)
		cont := testutils.CreateTestContainerWithConfig(t, db, testutils.WithTerraformOIDCConfig(keyPath))
		return &testutils.TestContainer{
			Container: cont,
			Router:    cont.Server.Router(),
		}
	}
	return testutils.CreateTestServer(t, db)
}

// TestModifyEndpoints_AllAuthMethods tests that MODIFY permission endpoints work correctly
// with all authentication methods and permission levels
func TestModifyEndpoints_AllAuthMethods(t *testing.T) {
	authMethods := []struct {
		name           string
		setup          func(t *testing.T, db *sqldb.Database, authHelper *testutils.AuthHelper) func(*http.Request)
		expectedStatus map[string]int // key is permission level, value is expected status
	}{
		{
			name: "unauthenticated",
			setup: func(t *testing.T, db *sqldb.Database, authHelper *testutils.AuthHelper) func(*http.Request) {
				return func(req *http.Request) {
					// No authentication
				}
			},
			expectedStatus: map[string]int{"": http.StatusUnauthorized},
		},
		{
			name: "admin_api_key",
			setup: func(t *testing.T, db *sqldb.Database, authHelper *testutils.AuthHelper) func(*http.Request) {
				return func(req *http.Request) {
					apiKey := os.Getenv("ADMIN_AUTH_TOKEN")
					if apiKey == "" {
						apiKey = "test-admin-api-key"
					}
					req.Header.Set("X-Terrareg-ApiKey", apiKey)
				}
			},
			expectedStatus: map[string]int{"": http.StatusOK},
		},
		{
			name: "upload_api_key",
			setup: func(t *testing.T, db *sqldb.Database, authHelper *testutils.AuthHelper) func(*http.Request) {
				return func(req *http.Request) {
					apiKey := os.Getenv("UPLOAD_AUTH_TOKEN")
					if apiKey == "" {
						apiKey = "test-upload-key"
					}
					// Python uses X-Terrareg-ApiKey for all API key types
					req.Header.Set("X-Terrareg-ApiKey", apiKey)
				}
			},
			// Upload key doesn't grant MODIFY access to settings endpoints
			expectedStatus: map[string]int{"": http.StatusForbidden},
		},
		{
			name: "user_with_no_permissions",
			setup: func(t *testing.T, db *sqldb.Database, authHelper *testutils.AuthHelper) func(*http.Request) {
				return func(req *http.Request) {
					cookie := authHelper.CreateSessionForUser("nopermuser", false, []string{}, nil)
					req.Header.Set("Cookie", cookie)
				}
			},
			expectedStatus: map[string]int{"": http.StatusForbidden},
		},
		{
			name: "user_with_read_permission",
			setup: func(t *testing.T, db *sqldb.Database, authHelper *testutils.AuthHelper) func(*http.Request) {
				return func(req *http.Request) {
					authHelper.SetupUserGroupWithPermissions("read-group", false, map[string]string{"test-ns": "READ"})
					cookie := authHelper.CreateSessionForUser("readuser", false, []string{"read-group"}, nil)
					req.Header.Set("Cookie", cookie)
				}
			},
			// READ permission is not sufficient for MODIFY endpoints
			expectedStatus: map[string]int{"": http.StatusForbidden},
		},
		{
			name: "user_with_modify_permission",
			setup: func(t *testing.T, db *sqldb.Database, authHelper *testutils.AuthHelper) func(*http.Request) {
				return func(req *http.Request) {
					authHelper.SetupUserGroupWithPermissions("modify-group", false, map[string]string{"test-ns": "MODIFY"})
					cookie := authHelper.CreateSessionForUser("modifyuser", false, []string{"modify-group"}, nil)
					req.Header.Set("Cookie", cookie)
				}
			},
			// MODIFY permission is sufficient
			expectedStatus: map[string]int{"": http.StatusOK},
		},
		{
			name: "user_with_full_permission",
			setup: func(t *testing.T, db *sqldb.Database, authHelper *testutils.AuthHelper) func(*http.Request) {
				return func(req *http.Request) {
					authHelper.SetupUserGroupWithPermissions("full-group", false, map[string]string{"test-ns": "FULL"})
					cookie := authHelper.CreateSessionForUser("fulluser", false, []string{"full-group"}, nil)
					req.Header.Set("Cookie", cookie)
				}
			},
			// FULL permission is also sufficient
			expectedStatus: map[string]int{"": http.StatusOK},
		},
		{
			name: "site_admin_user",
			setup: func(t *testing.T, db *sqldb.Database, authHelper *testutils.AuthHelper) func(*http.Request) {
				return func(req *http.Request) {
					authHelper.SetupUserGroupWithPermissions("admin-group", true, nil)
					cookie := authHelper.CreateSessionForUser("siteadmin", true, []string{"admin-group"}, nil)
					req.Header.Set("Cookie", cookie)
				}
			},
			// Site admin can access anything
			expectedStatus: map[string]int{"": http.StatusOK},
		},
		{
			name: "terraform_idp_no_permissions",
			setup: func(t *testing.T, db *sqldb.Database, authHelper *testutils.AuthHelper) func(*http.Request) {
				return func(req *http.Request) {
					token := authHelper.CreateTerraformIDPToken("tf-no-perms", nil)
					req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
				}
			},
			expectedStatus: map[string]int{"": http.StatusForbidden},
		},
		{
			name: "terraform_idp_with_modify_permission",
			setup: func(t *testing.T, db *sqldb.Database, authHelper *testutils.AuthHelper) func(*http.Request) {
				return func(req *http.Request) {
					token := authHelper.CreateTerraformIDPToken("tf-modify", map[string]string{"test-ns": "MODIFY"})
					req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
				}
			},
			// Python: check_namespace_access returns False for terraform tokens
			// Namespace permissions are not supported for terraform IDP tokens
			expectedStatus: map[string]int{"": http.StatusForbidden},
		},
	}

	for _, authMethod := range authMethods {
		t.Run(authMethod.name, func(t *testing.T) {
			db := testutils.SetupTestDatabase(t)
			defer testutils.CleanupTestDatabase(t, db)

			// Create test namespace and module
			namespace := testutils.CreateNamespace(t, db, "test-ns", nil)
			moduleProvider := testutils.CreateModuleProvider(t, db, namespace.ID, "test-mod", "test-prov")
			_ = testutils.CreatePublishedModuleVersion(t, db, moduleProvider.ID, "1.0.0")

			// Create container with signing key if terraform_idp test
			cont := setupModifyTestContainerWithSigningKey(t, db, authMethod.name)
			authHelper := testutils.NewAuthHelper(t, db, cont)
			setupFunc := authMethod.setup(t, db, authHelper)

			endpoints := []struct {
				name   string
				method string
				path   string
				body   []byte
			}{
				{
					name:   "update_settings",
					method: "PUT",
					path:   "/v1/terrareg/modules/test-ns/test-mod/test-prov/settings",
					body:   []byte(`{"description": "Updated description"}`),
				},
			}

			router := cont.Router

			for _, endpoint := range endpoints {
				t.Run(endpoint.name, func(t *testing.T) {
					req := httptest.NewRequest(endpoint.method, endpoint.path, bytes.NewReader(endpoint.body))
					req.Header.Set("Content-Type", "application/json")
					setupFunc(req)

					w := httptest.NewRecorder()
					router.ServeHTTP(w, req)

					expectedStatus := authMethod.expectedStatus[""]
					assert.Equal(t, expectedStatus, w.Code,
						fmt.Sprintf("Endpoint %s with auth %s should return %d",
							endpoint.path, authMethod.name, expectedStatus))
				})
			}
		})
	}
}

// TestModifyEndpoints_CrossNamespacePermissions tests that permissions are properly scoped to namespaces
func TestModifyEndpoints_CrossNamespacePermissions(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	// Create two namespaces
	ns1 := testutils.CreateNamespace(t, db, "ns1", nil)
	ns2 := testutils.CreateNamespace(t, db, "ns2", nil)

	// Create modules in both namespaces
	mp1 := testutils.CreateModuleProvider(t, db, ns1.ID, "test-mod", "test-prov")
	mp2 := testutils.CreateModuleProvider(t, db, ns2.ID, "test-mod", "test-prov")
	_ = testutils.CreatePublishedModuleVersion(t, db, mp1.ID, "1.0.0")
	_ = testutils.CreatePublishedModuleVersion(t, db, mp2.ID, "1.0.0")

	cont := testutils.CreateTestServer(t, db)
	authHelper := testutils.NewAuthHelper(t, db, cont)

	// Create user group with MODIFY permission only for ns1
	authHelper.SetupUserGroupWithPermissions("ns1-only-group", false, map[string]string{"ns1": "MODIFY"})
	cookie := authHelper.CreateSessionForUser("ns1user", false, []string{"ns1-only-group"}, nil)

	router := cont.Router

	tests := []struct {
		name           string
		path           string
		expectedStatus int
	}{
		{
			name:           "namespace_with_permission",
			path:           "/v1/terrareg/modules/ns1/test-mod/test-prov/settings",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "namespace_without_permission",
			path:           "/v1/terrareg/modules/ns2/test-mod/test-prov/settings",
			expectedStatus: http.StatusForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body := []byte(`{"description": "Updated description"}`)
			req := httptest.NewRequest("PUT", tt.path, bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Cookie", cookie)

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

// TestModifyEndpoints_AdminBypass tests that admin users can access any namespace
func TestModifyEndpoints_AdminBypass(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	// Create a test namespace
	namespace := testutils.CreateNamespace(t, db, "random-ns", nil)
	moduleProvider := testutils.CreateModuleProvider(t, db, namespace.ID, "test-mod", "test-prov")
	_ = testutils.CreatePublishedModuleVersion(t, db, moduleProvider.ID, "1.0.0")

	cont := testutils.CreateTestContainer(t, db)
	router := cont.Server.Router()

	// Test with admin API key - should work for any namespace
	body := []byte(`{"description": "Updated by admin"}`)
	req := httptest.NewRequest("PUT", "/v1/terrareg/modules/random-ns/test-mod/test-prov/settings", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	apiKey := os.Getenv("ADMIN_AUTH_TOKEN")
	if apiKey == "" {
		apiKey = "test-admin-api-key"
	}
	req.Header.Set("X-Terrareg-ApiKey", apiKey)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code, "Admin should be able to modify any namespace")
}

// TestModifyEndpoints_RequireUploadPermission tests the RequireUploadPermission middleware
func TestModifyEndpoints_RequireUploadPermission(t *testing.T) {
	authMethods := []struct {
		name           string
		setup          func(t *testing.T, req *http.Request)
		expectAuthPass bool // true = should NOT get 401 Unauthorized
	}{
		{
			name:           "unauthenticated",
			setup:          func(t *testing.T, req *http.Request) {},
			expectAuthPass: false, // Expect 401 Unauthorized
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
			expectAuthPass: true, // Admin bypasses upload permission check
		},
		{
			name: "upload_api_key",
			setup: func(t *testing.T, req *http.Request) {
				apiKey := os.Getenv("UPLOAD_AUTH_TOKEN")
				if apiKey == "" {
					apiKey = "test-upload-key"
				}
				req.Header.Set("X-Terrareg-ApiKey", apiKey)
			},
			expectAuthPass: true, // Upload API key has upload permission
		},
		{
			name: "publish_api_key",
			setup: func(t *testing.T, req *http.Request) {
				apiKey := os.Getenv("PUBLISH_AUTH_TOKEN")
				if apiKey == "" {
					apiKey = "test-publish-key"
				}
				req.Header.Set("X-Terrareg-ApiKey", apiKey)
			},
			expectAuthPass: true, // Publish API key has upload permission
		},
		{
			name: "user_with_read_permission",
			setup: func(t *testing.T, req *http.Request) {
				db := testutils.SetupTestDatabase(t)
				defer testutils.CleanupTestDatabase(t, db)
				cont := testutils.CreateTestServer(t, db)
				authHelper := testutils.NewAuthHelper(t, db, cont)
				authHelper.SetupUserGroupWithPermissions("read-group", false, map[string]string{"test-ns": "READ"})
				cookie := authHelper.CreateSessionForUser("readuser", false, []string{"read-group"}, nil)
				req.Header.Set("Cookie", cookie)
			},
			expectAuthPass: false, // Read permission is not sufficient for upload
		},
		{
			name: "user_with_modify_permission",
			setup: func(t *testing.T, req *http.Request) {
				db := testutils.SetupTestDatabase(t)
				defer testutils.CleanupTestDatabase(t, db)
				cont := testutils.CreateTestServer(t, db)
				authHelper := testutils.NewAuthHelper(t, db, cont)
				authHelper.SetupUserGroupWithPermissions("modify-group", false, map[string]string{"test-ns": "MODIFY"})
				cookie := authHelper.CreateSessionForUser("modifyuser", false, []string{"modify-group"}, nil)
				req.Header.Set("Cookie", cookie)
			},
			expectAuthPass: true, // MODIFY permission grants upload access
		},
		{
			name: "user_with_full_permission",
			setup: func(t *testing.T, req *http.Request) {
				db := testutils.SetupTestDatabase(t)
				defer testutils.CleanupTestDatabase(t, db)
				cont := testutils.CreateTestServer(t, db)
				authHelper := testutils.NewAuthHelper(t, db, cont)
				authHelper.SetupUserGroupWithPermissions("full-group", false, map[string]string{"test-ns": "FULL"})
				cookie := authHelper.CreateSessionForUser("fulluser", false, []string{"full-group"}, nil)
				req.Header.Set("Cookie", cookie)
			},
			expectAuthPass: true, // FULL permission grants upload access
		},
	}

	for _, authMethod := range authMethods {
		t.Run(authMethod.name, func(t *testing.T) {
			db := testutils.SetupTestDatabase(t)
			defer testutils.CleanupTestDatabase(t, db)

			// Create test namespace and module
			namespace := testutils.CreateNamespace(t, db, "test-ns", nil)
			moduleProvider := testutils.CreateModuleProvider(t, db, namespace.ID, "test-mod", "test-prov")
			_ = testutils.CreatePublishedModuleVersion(t, db, moduleProvider.ID, "1.0.0")

			cont := testutils.CreateTestContainer(t, db)
			router := cont.Server.Router()

			// Test upload endpoint which uses RequireUploadPermission
			// Note: This will return 400 without proper file upload, but we're testing auth
			body := []byte("{}")
			req := httptest.NewRequest("POST", "/v1/terrareg/modules/test-ns/test-mod/test-prov/1.0.0/upload", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			authMethod.setup(t, req)

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if authMethod.expectAuthPass {
				// Should NOT get 401 Unauthorized (may get 400 for missing file, etc.)
				assert.NotEqual(t, http.StatusUnauthorized, w.Code,
					fmt.Sprintf("Upload endpoint with auth %s should not return 401 Unauthorized", authMethod.name))
			} else {
				// Should get 401 Unauthorized or 403 Forbidden
				assert.Contains(t, []int{http.StatusUnauthorized, http.StatusForbidden}, w.Code,
					fmt.Sprintf("Upload endpoint with auth %s should return 401 or 403", authMethod.name))
			}
		})
	}
}
