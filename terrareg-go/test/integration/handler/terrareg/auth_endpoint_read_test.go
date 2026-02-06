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
	"github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/persistence/sqldb"
)

// TestReadEndpoints_AllAuthMethods tests that read access endpoints work correctly
// with all authentication methods and both ALLOW_UNAUTHENTICATED_ACCESS states
func TestReadEndpoints_AllAuthMethods(t *testing.T) {
	endpoints := []struct {
		name                 string
		method               string
		path                 string
		requiresTerraformAPI bool // true = requires can_access_terraform_api, false = requires can_access_read_api
	}{
		{
			name:                 "module_list",
			method:               "GET",
			path:                 "/v1/modules",
			requiresTerraformAPI: false, // requires can_access_read_api
		},
		{
			name:                 "module_versions",
			method:               "GET",
			path:                 "/v1/modules/test-ns/test-mod/test-prov/versions",
			requiresTerraformAPI: true, // requires can_access_terraform_api
		},
	}

	authMethods := []struct {
		name                 string
		setup                func(t *testing.T, db *sqldb.Database, cont *testutils.TestContainer, req *http.Request)
		hasReadAPIAccess     bool // true = has can_access_read_api
		hasTerraformAPIAccess bool // true = has can_access_terraform_api
		configOptions        []testutils.InfraConfigOption // config options needed for this auth method
		requiresSigningKey   bool // true = needs a signing key file to be generated for Terraform OIDC
	}{
		{
			name:                 "unauthenticated",
			setup:                func(t *testing.T, db *sqldb.Database, cont *testutils.TestContainer, req *http.Request) {},
			hasReadAPIAccess:     false, // controlled by ALLOW_UNAUTHENTICATED_ACCESS config
			hasTerraformAPIAccess: false,
			configOptions:        nil,
			requiresSigningKey:   false,
		},
		{
			name: "admin_api_key",
			setup: func(t *testing.T, db *sqldb.Database, cont *testutils.TestContainer, req *http.Request) {
				apiKey := os.Getenv("ADMIN_AUTH_TOKEN")
				if apiKey == "" {
					apiKey = "test-admin-api-key"
				}
				req.Header.Set("X-Terrareg-ApiKey", apiKey)
			},
			hasReadAPIAccess:     true,
			hasTerraformAPIAccess: true,
			configOptions:        nil,
			requiresSigningKey:   false,
		},
		{
			name: "upload_api_key",
			setup: func(t *testing.T, db *sqldb.Database, cont *testutils.TestContainer, req *http.Request) {
				apiKey := os.Getenv("UPLOAD_AUTH_TOKEN")
				if apiKey == "" {
					apiKey = "test-upload-key"
				}
				// Python uses X-Terrareg-ApiKey for all API key types (admin, upload, publish)
				req.Header.Set("X-Terrareg-ApiKey", apiKey)
			},
			hasReadAPIAccess:     false, // upload API keys don't have read API access (Python: base_api_key_auth_method.py)
			hasTerraformAPIAccess: false,
			configOptions:        nil,
			requiresSigningKey:   false,
		},
		{
			name: "publish_api_key",
			setup: func(t *testing.T, db *sqldb.Database, cont *testutils.TestContainer, req *http.Request) {
				apiKey := os.Getenv("PUBLISH_AUTH_TOKEN")
				if apiKey == "" {
					apiKey = "test-publish-key"
				}
				// Python uses X-Terrareg-ApiKey for all API key types (admin, upload, publish)
				req.Header.Set("X-Terrareg-ApiKey", apiKey)
			},
			hasReadAPIAccess:     false, // publish API keys don't have read API access (Python: base_api_key_auth_method.py)
			hasTerraformAPIAccess: false,
			configOptions:        nil,
			requiresSigningKey:   false,
		},
		{
			name: "user_session",
			setup: func(t *testing.T, db *sqldb.Database, cont *testutils.TestContainer, req *http.Request) {
				authHelper := testutils.NewAuthHelper(t, db, cont)
				cookie := authHelper.CreateSessionForUser("testuser", false, []string{}, nil)
				req.Header.Set("Cookie", cookie)
			},
			hasReadAPIAccess:     true, // session auth has read API access (Python: base_session_auth_method.py)
			hasTerraformAPIAccess: true, // inherits from read API access
			configOptions:        nil,
			requiresSigningKey:   false,
		},
		{
			name: "user_with_read_permission",
			setup: func(t *testing.T, db *sqldb.Database, cont *testutils.TestContainer, req *http.Request) {
				authHelper := testutils.NewAuthHelper(t, db, cont)
				// Create a group with READ permission for test-ns
				authHelper.SetupUserGroupWithPermissions("read-group", false, map[string]string{"test-ns": "READ"})
				cookie := authHelper.CreateSessionForUser("readuser", false, []string{"read-group"}, nil)
				req.Header.Set("Cookie", cookie)
			},
			hasReadAPIAccess:     true, // session auth has read API access
			hasTerraformAPIAccess: true, // inherits from read API access
			configOptions:        nil,
			requiresSigningKey:   false,
		},
		{
			name: "terraform_idp_token",
			setup: func(t *testing.T, db *sqldb.Database, cont *testutils.TestContainer, req *http.Request) {
				authHelper := testutils.NewAuthHelper(t, db, cont)
				token := authHelper.CreateTerraformIDPToken("test-subject", nil)
				req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
			},
			hasReadAPIAccess:     false, // terraform IDP tokens don't have read API access (Python: base_terraform_static_token.py)
			hasTerraformAPIAccess: true,  // but they do have terraform API access
			configOptions:        nil,
			requiresSigningKey:   true, // needs a signing key file to be generated
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

						// Build config options - enable RBAC so permission checking works properly
						opts := []testutils.InfraConfigOption{
							testutils.WithAllowUnauthenticatedAccess(true),
							testutils.WithEnableAccessControls(true),
						}
						if authMethod.configOptions != nil {
							opts = append(opts, authMethod.configOptions...)
						}

						// Generate signing key if needed for Terraform OIDC
						var keyCleanup func()
						if authMethod.requiresSigningKey {
							keyPath, cleanup := testutils.CreateTestTerraformOIDCSigningKey(t)
							opts = append(opts, testutils.WithTerraformOIDCConfig(keyPath))
							defer cleanup()
							keyCleanup = cleanup
						}

						cont := testutils.CreateTestContainerWithConfig(t, db, opts...)
						testServer := &testutils.TestContainer{
							Container: cont,
							Router:    cont.Server.Router(),
						}
						router := testServer.Router

						req := httptest.NewRequest(endpoint.method, endpoint.path, nil)
						authMethod.setup(t, db, testServer, req)

						// Clean up signing key if it was generated
						if keyCleanup != nil {
							keyCleanup()
						}

						w := httptest.NewRecorder()
						router.ServeHTTP(w, req)

						// With ALLOW_UNAUTHENTICATED_ACCESS=true, unauthenticated users have read access
						// Check if the auth method should have access based on endpoint type
						var hasAccess bool
						if endpoint.requiresTerraformAPI {
							hasAccess = authMethod.hasTerraformAPIAccess
						} else {
							// For read endpoints, check read API access
							// Unauthenticated users have read access when ALLOW_UNAUTHENTICATED_ACCESS=true
							if authMethod.name == "unauthenticated" {
								hasAccess = true
							} else {
								hasAccess = authMethod.hasReadAPIAccess
							}
						}

						expectedCode := http.StatusOK
						if !hasAccess {
							expectedCode = http.StatusUnauthorized
						}

						assert.Equal(t, expectedCode, w.Code,
							fmt.Sprintf("Endpoint %s with auth %s and ALLOW_UNAUTHENTICATED_ACCESS=true should return %d",
								endpoint.path, authMethod.name, expectedCode))
					})

					// Test with ALLOW_UNAUTHENTICATED_ACCESS=false
					t.Run("ALLOW_UNAUTHENTICATED_ACCESS=false", func(t *testing.T) {
						db := testutils.SetupTestDatabase(t)
						defer testutils.CleanupTestDatabase(t, db)

						// Recreate test data for this test
						namespace := testutils.CreateNamespace(t, db, "test-ns", nil)
						moduleProvider := testutils.CreateModuleProvider(t, db, namespace.ID, "test-mod", "test-prov")
						_ = testutils.CreatePublishedModuleVersion(t, db, moduleProvider.ID, "1.0.0")

						// Build config options - enable RBAC so permission checking works properly
						opts := []testutils.InfraConfigOption{
							testutils.WithAllowUnauthenticatedAccess(false),
							testutils.WithEnableAccessControls(true),
						}
						if authMethod.configOptions != nil {
							opts = append(opts, authMethod.configOptions...)
						}

						// Generate signing key if needed for Terraform OIDC
						var keyCleanup func()
						if authMethod.requiresSigningKey {
							keyPath, cleanup := testutils.CreateTestTerraformOIDCSigningKey(t)
							opts = append(opts, testutils.WithTerraformOIDCConfig(keyPath))
							defer cleanup()
							keyCleanup = cleanup
						}

						cont := testutils.CreateTestContainerWithConfig(t, db, opts...)
						testServer := &testutils.TestContainer{
							Container: cont,
							Router:    cont.Server.Router(),
						}
						router := testServer.Router

						req := httptest.NewRequest(endpoint.method, endpoint.path, nil)
						authMethod.setup(t, db, testServer, req)

						// Clean up signing key if it was generated
						if keyCleanup != nil {
							keyCleanup()
						}

						w := httptest.NewRecorder()
						router.ServeHTTP(w, req)

						// Check if the auth method should have access based on endpoint type
						var hasAccess bool
						if endpoint.requiresTerraformAPI {
							hasAccess = authMethod.hasTerraformAPIAccess
						} else {
							// For read endpoints, unauthenticated users don't have access when config is false
							hasAccess = authMethod.hasReadAPIAccess
						}

						expectedCode := http.StatusOK
						if !hasAccess {
							expectedCode = http.StatusUnauthorized
						}

						assert.Equal(t, expectedCode, w.Code,
							fmt.Sprintf("Endpoint %s with auth %s and ALLOW_UNAUTHENTICATED_ACCESS=false should return %d",
								endpoint.path, authMethod.name, expectedCode))
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

	// Enable RBAC so permission checking works properly
	cont := testutils.CreateTestServerWithConfig(t, db, testutils.WithEnableAccessControls(true))
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
			// Enable RBAC so permission checking works properly
			cont := testutils.CreateTestContainerWithConfig(t, db,
				testutils.WithAllowUnauthenticatedAccess(tt.allowUnauthenticated),
				testutils.WithEnableAccessControls(true))
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

	// Enable RBAC so permission checking works properly
	cont := testutils.CreateTestServerWithConfig(t, db, testutils.WithEnableAccessControls(true))
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
