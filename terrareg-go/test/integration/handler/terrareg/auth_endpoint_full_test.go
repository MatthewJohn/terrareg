package terrareg

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/persistence/sqldb"
	"github.com/matthewjohn/terrareg/terrareg-go/test/integration/testutils"
)

// TestFullEndpoints_AllAuthMethods tests that FULL permission endpoints work correctly
// Only FULL permission or admin should allow these operations
func TestFullEndpoints_AllAuthMethods(t *testing.T) {
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
					req.Header.Set("X-Terrareg-UploadKey", apiKey)
				}
			},
			expectedStatus: map[string]int{"": http.StatusOK},
		},
		{
			name: "publish_api_key",
			setup: func(t *testing.T, db *sqldb.Database, authHelper *testutils.AuthHelper) func(*http.Request) {
				return func(req *http.Request) {
					apiKey := os.Getenv("PUBLISH_AUTH_TOKEN")
					if apiKey == "" {
						apiKey = "test-publish-key"
					}
					req.Header.Set("X-Terrareg-PublishKey", apiKey)
				}
			},
			expectedStatus: map[string]int{"": http.StatusOK},
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
			// READ permission is NOT sufficient for FULL endpoints
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
			// MODIFY permission is NOT sufficient for FULL endpoints
			expectedStatus: map[string]int{"": http.StatusForbidden},
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
			// FULL permission is required
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
			// MODIFY permission is NOT sufficient for FULL endpoints
			expectedStatus: map[string]int{"": http.StatusForbidden},
		},
		{
			name: "terraform_idp_with_full_permission",
			setup: func(t *testing.T, db *sqldb.Database, authHelper *testutils.AuthHelper) func(*http.Request) {
				return func(req *http.Request) {
					token := authHelper.CreateTerraformIDPToken("tf-full", map[string]string{"test-ns": "FULL"})
					req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
				}
			},
			expectedStatus: map[string]int{"": http.StatusOK},
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

			authHelper := testutils.NewAuthHelper(t, db, &testutils.TestServer{})
			setupFunc := authMethod.setup(t, db, authHelper)

			endpoints := []struct {
				name   string
				method string
				path   string
				body   []byte
			}{
				{
					name:   "delete_module_provider",
					method: "DELETE",
					path:   "/v1/terrareg/modules/test-ns/test-mod/test-prov",
					body:   nil,
				},
			}

			for _, endpoint := range endpoints {
				t.Run(endpoint.name, func(t *testing.T) {
					cont := testutils.CreateTestContainer(t, db)
					router := cont.Server.Router()

					var req *http.Request
					if endpoint.body != nil {
						req = httptest.NewRequest(endpoint.method, endpoint.path, bytes.NewReader(endpoint.body))
					} else {
						req = httptest.NewRequest(endpoint.method, endpoint.path, nil)
					}
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

// TestFullEndpoints_ModifyPermissionDenied tests that MODIFY permission is denied for FULL endpoints
func TestFullEndpoints_ModifyPermissionDenied(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	// Create test namespace and module
	namespace := testutils.CreateNamespace(t, db, "test-ns", nil)
	moduleProvider := testutils.CreateModuleProvider(t, db, namespace.ID, "test-mod", "test-prov")
	_ = testutils.CreatePublishedModuleVersion(t, db, moduleProvider.ID, "1.0.0")

	authHelper := testutils.NewAuthHelper(t, db, &testutils.TestServer{})

	// Create user with MODIFY permission
	authHelper.SetupUserGroupWithPermissions("modify-group", false, map[string]string{"test-ns": "MODIFY"})
	cookie := authHelper.CreateSessionForUser("modifyuser", false, []string{"modify-group"}, nil)

	cont := testutils.CreateTestContainer(t, db)
	router := cont.Server.Router()

	// Try to delete module provider (requires FULL permission)
	req := httptest.NewRequest("DELETE", "/v1/terrareg/modules/test-ns/test-mod/test-prov", nil)
	req.Header.Set("Cookie", cookie)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// MODIFY permission should be denied for FULL endpoints
	assert.Equal(t, http.StatusForbidden, w.Code, "MODIFY permission should be denied for FULL endpoints")
}

// TestFullEndpoints_FullPermissionAllowed tests that FULL permission allows operations
func TestFullEndpoints_FullPermissionAllowed(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	// Create test namespace and module
	namespace := testutils.CreateNamespace(t, db, "test-ns", nil)
	moduleProvider := testutils.CreateModuleProvider(t, db, namespace.ID, "test-mod", "test-prov")
	_ = testutils.CreatePublishedModuleVersion(t, db, moduleProvider.ID, "1.0.0")

	authHelper := testutils.NewAuthHelper(t, db, &testutils.TestServer{})

	// Create user with FULL permission
	authHelper.SetupUserGroupWithPermissions("full-group", false, map[string]string{"test-ns": "FULL"})
	cookie := authHelper.CreateSessionForUser("fulluser", false, []string{"full-group"}, nil)

	cont := testutils.CreateTestContainer(t, db)
	router := cont.Server.Router()

	// Try to delete module version (requires FULL permission)
	req := httptest.NewRequest("DELETE", "/v1/terrareg/modules/test-ns/test-mod/test-prov/1.0.0", nil)
	req.Header.Set("Cookie", cookie)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// FULL permission should be allowed
	assert.Equal(t, http.StatusOK, w.Code, "FULL permission should allow deletion")
}

// TestFullEndpoints_AdminBypass tests that admin bypass works for FULL endpoints
func TestFullEndpoints_AdminBypass(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	// Create test namespace and module
	namespace := testutils.CreateNamespace(t, db, "random-ns", nil)
	moduleProvider := testutils.CreateModuleProvider(t, db, namespace.ID, "test-mod", "test-prov")
	_ = testutils.CreatePublishedModuleVersion(t, db, moduleProvider.ID, "1.0.0")

	cont := testutils.CreateTestContainer(t, db)
	router := cont.Server.Router()

	// Test with admin API key - should work for any namespace
	req := httptest.NewRequest("DELETE", "/v1/terrareg/modules/random-ns/test-mod/test-prov", nil)

	apiKey := os.Getenv("ADMIN_AUTH_TOKEN")
	if apiKey == "" {
		apiKey = "test-admin-api-key"
	}
	req.Header.Set("X-Terrareg-ApiKey", apiKey)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code, "Admin should be able to delete any module provider")
}

// TestFullEndpoints_UploadApiKeyWorks tests that upload API key works for create operations
func TestFullEndpoints_UploadApiKeyWorks(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	// Create test namespace
	_ = testutils.CreateNamespace(t, db, "test-ns", nil)

	cont := testutils.CreateTestContainer(t, db)
	router := cont.Server.Router()

	// Create module provider using upload API key
	body := map[string]interface{}{
		"repository_url": "https://github.com/example/test-mod",
	}
	bodyBytes, _ := json.Marshal(body)
	req := httptest.NewRequest("POST", "/v1/terrareg/modules/test-ns/test-mod/test-prov", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	apiKey := os.Getenv("UPLOAD_AUTH_TOKEN")
	if apiKey == "" {
		apiKey = "test-upload-key"
	}
	req.Header.Set("X-Terrareg-UploadKey", apiKey)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Upload API key should allow module provider creation
	assert.Equal(t, http.StatusOK, w.Code, "Upload API key should allow module provider creation")
}
