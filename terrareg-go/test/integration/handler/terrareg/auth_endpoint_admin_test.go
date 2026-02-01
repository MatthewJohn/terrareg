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

// TestAdminEndpoints_AllAuthMethods tests that admin-only endpoints work correctly
// Only site admin users should be able to access these endpoints
func TestAdminEndpoints_AllAuthMethods(t *testing.T) {
	authMethods := []struct {
		name           string
		setup          func(t *testing.T, db *sqldb.Database, authHelper *testutils.AuthHelper) func(*http.Request)
		expectedStatus int
	}{
		{
			name: "unauthenticated",
			setup: func(t *testing.T, db *sqldb.Database, authHelper *testutils.AuthHelper) func(*http.Request) {
				return func(req *http.Request) {
					// No authentication
				}
			},
			expectedStatus: http.StatusForbidden,
		},
		{
			name: "regular_user_session",
			setup: func(t *testing.T, db *sqldb.Database, authHelper *testutils.AuthHelper) func(*http.Request) {
				return func(req *http.Request) {
					cookie := authHelper.CreateSessionForUser("regularuser", false, []string{}, nil)
					req.Header.Set("Cookie", cookie)
				}
			},
			expectedStatus: http.StatusForbidden,
		},
		{
			name: "user_with_full_permission_but_not_site_admin",
			setup: func(t *testing.T, db *sqldb.Database, authHelper *testutils.AuthHelper) func(*http.Request) {
				return func(req *http.Request) {
					authHelper.SetupUserGroupWithPermissions("full-group", false, map[string]string{"test-ns": "FULL"})
					cookie := authHelper.CreateSessionForUser("fulluser", false, []string{"full-group"}, nil)
					req.Header.Set("Cookie", cookie)
				}
			},
			// Even with FULL permission, non-site-admin should be denied
			expectedStatus: http.StatusForbidden,
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
			expectedStatus: http.StatusOK,
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
			expectedStatus: http.StatusOK,
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
			// Upload API key doesn't grant admin access
			expectedStatus: http.StatusForbidden,
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
			// Publish API key doesn't grant admin access
			expectedStatus: http.StatusForbidden,
		},
		{
			name: "terraform_idp_no_permissions",
			setup: func(t *testing.T, db *sqldb.Database, authHelper *testutils.AuthHelper) func(*http.Request) {
				return func(req *http.Request) {
					token := authHelper.CreateTerraformIDPToken("tf-regular", nil)
					req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
				}
			},
			expectedStatus: http.StatusForbidden,
		},
		{
			name: "terraform_idp_site_admin",
			setup: func(t *testing.T, db *sqldb.Database, authHelper *testutils.AuthHelper) func(*http.Request) {
				return func(req *http.Request) {
					// Create Terraform IDP subject with site admin
					token := authHelper.CreateTerraformIDPToken("tf-admin", nil)
					authHelper.GetUserGroupsForSubject("tf-admin", []string{"admin-group"}, true)
					req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
				}
			},
			expectedStatus: http.StatusOK,
		},
	}

	for _, authMethod := range authMethods {
		t.Run(authMethod.name, func(t *testing.T) {
			db := testutils.SetupTestDatabase(t)
			defer testutils.CleanupTestDatabase(t, db)

			authHelper := testutils.NewAuthHelper(t, db, &testutils.TestServer{})
			setupFunc := authMethod.setup(t, db, authHelper)

			endpoints := []struct {
				name   string
				method string
				path   string
				body   []byte
			}{
				{
					name:   "get_config",
					method: "GET",
					path:   "/v1/terrareg/admin/config",
				},
				{
					name:   "list_users",
					method: "GET",
					path:   "/v1/terrareg/auth/admin/users",
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

					assert.Equal(t, authMethod.expectedStatus, w.Code,
						fmt.Sprintf("Endpoint %s with auth %s should return %d",
							endpoint.path, authMethod.name, authMethod.expectedStatus))
				})
			}
		})
	}
}

// TestAdminEndpoints_FullPermissionNotSufficient tests that FULL permission doesn't grant admin access
func TestAdminEndpoints_FullPermissionNotSufficient(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	authHelper := testutils.NewAuthHelper(t, db, &testutils.TestServer{})

	// Create user with FULL permission but NOT site admin
	authHelper.SetupUserGroupWithPermissions("full-group", false, map[string]string{"test-ns": "FULL"})
	cookie := authHelper.CreateSessionForUser("fulluser", false, []string{"full-group"}, nil)

	cont := testutils.CreateTestContainer(t, db)
	router := cont.Server.Router()

	// Try to access admin config endpoint
	req := httptest.NewRequest("GET", "/v1/terrareg/admin/config", nil)
	req.Header.Set("Cookie", cookie)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// FULL permission should NOT grant admin access
	assert.Equal(t, http.StatusForbidden, w.Code, "FULL permission should not grant admin access")
}

// TestAdminEndpoints_SiteAdminSufficient tests that site admin can access admin endpoints
func TestAdminEndpoints_SiteAdminSufficient(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	authHelper := testutils.NewAuthHelper(t, db, &testutils.TestServer{})

	// Create user group with site_admin = true
	authHelper.SetupUserGroupWithPermissions("admin-group", true, nil)
	cookie := authHelper.CreateSessionForUser("siteadmin", true, []string{"admin-group"}, nil)

	cont := testutils.CreateTestContainer(t, db)
	router := cont.Server.Router()

	// Try to access admin config endpoint
	req := httptest.NewRequest("GET", "/v1/terrareg/admin/config", nil)
	req.Header.Set("Cookie", cookie)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Site admin should be able to access admin endpoints
	assert.Equal(t, http.StatusOK, w.Code, "Site admin should be able to access admin endpoints")
}

// TestAdminEndpoints_UpdateConfig tests that only site admin can update config
func TestAdminEndpoints_UpdateConfig(t *testing.T) {
	authMethods := []struct {
		name           string
		setup          func(t *testing.T, db *sqldb.Database, authHelper *testutils.AuthHelper) func(*http.Request)
		expectedStatus int
	}{
		{
			name: "unauthenticated",
			setup: func(t *testing.T, db *sqldb.Database, authHelper *testutils.AuthHelper) func(*http.Request) {
				return func(req *http.Request) {}
			},
			expectedStatus: http.StatusForbidden,
		},
		{
			name: "regular_user",
			setup: func(t *testing.T, db *sqldb.Database, authHelper *testutils.AuthHelper) func(*http.Request) {
				return func(req *http.Request) {
					cookie := authHelper.CreateSessionForUser("regularuser", false, []string{}, nil)
					req.Header.Set("Cookie", cookie)
				}
			},
			expectedStatus: http.StatusForbidden,
		},
		{
			name: "site_admin",
			setup: func(t *testing.T, db *sqldb.Database, authHelper *testutils.AuthHelper) func(*http.Request) {
				return func(req *http.Request) {
					authHelper.SetupUserGroupWithPermissions("admin-group", true, nil)
					cookie := authHelper.CreateSessionForUser("siteadmin", true, []string{"admin-group"}, nil)
					req.Header.Set("Cookie", cookie)
				}
			},
			expectedStatus: http.StatusOK,
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
			expectedStatus: http.StatusOK,
		},
	}

	for _, authMethod := range authMethods {
		t.Run(authMethod.name, func(t *testing.T) {
			db := testutils.SetupTestDatabase(t)
			defer testutils.CleanupTestDatabase(t, db)

			authHelper := testutils.NewAuthHelper(t, db, &testutils.TestServer{})
			setupFunc := authMethod.setup(t, db, authHelper)

			cont := testutils.CreateTestContainer(t, db)
			router := cont.Server.Router()

			// Try to update config
			body := map[string]interface{}{
				"some_config": "value",
			}
			bodyBytes, _ := json.Marshal(body)
			req := httptest.NewRequest("POST", "/v1/terrareg/admin/config", bytes.NewReader(bodyBytes))
			req.Header.Set("Content-Type", "application/json")
			setupFunc(req)

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, authMethod.expectedStatus, w.Code,
				fmt.Sprintf("Update config with auth %s should return %d", authMethod.name, authMethod.expectedStatus))
		})
	}
}

// TestAdminEndpoints_ApiKeysDontGrantAdminAccess tests that API keys don't grant admin endpoint access
func TestAdminEndpoints_ApiKeysDontGrantAdminAccess(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	cont := testutils.CreateTestContainer(t, db)
	router := cont.Server.Router()

	apiKeys := []struct {
		name   string
		header string
		key    string
	}{
		{
			name:   "upload_key",
			header: "X-Terrareg-UploadKey",
			key: func() string {
				if k := os.Getenv("UPLOAD_AUTH_TOKEN"); k != "" {
					return k
				}
				return "test-upload-key"
			}(),
		},
		{
			name:   "publish_key",
			header: "X-Terrareg-PublishKey",
			key: func() string {
				if k := os.Getenv("PUBLISH_AUTH_TOKEN"); k != "" {
					return k
				}
				return "test-publish-key"
			}(),
		},
	}

	for _, apiKey := range apiKeys {
		t.Run(apiKey.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/v1/terrareg/admin/config", nil)
			req.Header.Set(apiKey.header, apiKey.key)

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// API keys should NOT grant admin endpoint access
			assert.Equal(t, http.StatusForbidden, w.Code,
				fmt.Sprintf("%s should not grant admin access", apiKey.name))
		})
	}
}

// TestAdminEndpoints_LogoutAllUsers tests that only site admin can logout all users
func TestAdminEndpoints_LogoutAllUsers(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	authHelper := testutils.NewAuthHelper(t, db, &testutils.TestServer{})

	cont := testutils.CreateTestContainer(t, db)
	router := cont.Server.Router()

	tests := []struct {
		name           string
		setup          func(t *testing.T) *http.Request
		expectedStatus int
	}{
		{
			name: "unauthenticated",
			setup: func(t *testing.T) *http.Request {
				return httptest.NewRequest("POST", "/v1/terrareg/auth/admin/logout-all", nil)
			},
			expectedStatus: http.StatusForbidden,
		},
		{
			name: "regular_user",
			setup: func(t *testing.T) *http.Request {
				cookie := authHelper.CreateSessionForUser("regularuser", false, []string{}, nil)
				req := httptest.NewRequest("POST", "/v1/terrareg/auth/admin/logout-all", nil)
				req.Header.Set("Cookie", cookie)
				return req
			},
			expectedStatus: http.StatusForbidden,
		},
		{
			name: "site_admin",
			setup: func(t *testing.T) *http.Request {
				authHelper.SetupUserGroupWithPermissions("admin-group", true, nil)
				cookie := authHelper.CreateSessionForUser("siteadmin", true, []string{"admin-group"}, nil)
				req := httptest.NewRequest("POST", "/v1/terrareg/auth/admin/logout-all", nil)
				req.Header.Set("Cookie", cookie)
				return req
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "admin_api_key",
			setup: func(t *testing.T) *http.Request {
				apiKey := os.Getenv("ADMIN_AUTH_TOKEN")
				if apiKey == "" {
					apiKey = "test-admin-api-key"
				}
				req := httptest.NewRequest("POST", "/v1/terrareg/auth/admin/logout-all", nil)
				req.Header.Set("X-Terrareg-ApiKey", apiKey)
				return req
			},
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			router.ServeHTTP(w, tt.setup(t))

			assert.Equal(t, tt.expectedStatus, w.Code,
				fmt.Sprintf("Logout all users with %s should return %d", tt.name, tt.expectedStatus))
		})
	}
}
