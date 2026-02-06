package terrareg_test

import (
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/persistence/sqldb"
	"github.com/matthewjohn/terrareg/terrareg-go/test/integration/testutils"
)

// TestModuleProviderCreate_Authentication tests module provider creation with RequireNamespacePermission(FULL) middleware
func TestModuleProviderCreate_Authentication(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	// Enable RBAC for this test so permission checking works properly
	cont := testutils.CreateTestServerWithConfig(t, db, testutils.WithEnableAccessControls(true))
	router := cont.Router

	// Create test namespace
	_ = testutils.CreateNamespace(t, db, "test-namespace", nil)

	tests := []struct {
		name           string
		setupAuth      func(*testing.T, *sqldb.Database) *http.Request
		expectedStatus int
	}{
		{
			name: "unauthenticated request returns 401",
			setupAuth: func(t *testing.T, db *sqldb.Database) *http.Request {
				req := testutils.BuildUnauthenticatedRequest(t, "POST", "/v1/terrareg/modules/test-namespace/testmod/testprovider/create")
				return testutils.AddChiContext(t, req, map[string]string{"namespace": "test-namespace", "name": "testmod", "provider": "testprovider"})
			},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name: "user with READ permission returns 403",
			setupAuth: func(t *testing.T, db *sqldb.Database) *http.Request {
				req, _ := testutils.BuildAuthenticatedRequestWithNamespacePermission(
					t, db, "POST", "/v1/terrareg/modules/test-namespace/testmod/testprovider/create",
					"readonly-user", "test-namespace", sqldb.PermissionTypeRead,
				)
				return testutils.AddChiContext(t, req, map[string]string{"namespace": "test-namespace", "name": "testmod", "provider": "testprovider"})
			},
			expectedStatus: http.StatusForbidden,
		},
		{
			name: "user with MODIFY permission returns 403",
			setupAuth: func(t *testing.T, db *sqldb.Database) *http.Request {
				req, _ := testutils.BuildAuthenticatedRequestWithNamespacePermission(
					t, db, "POST", "/v1/terrareg/modules/test-namespace/testmod/testprovider/create",
					"modify-user", "test-namespace", sqldb.PermissionTypeModify,
				)
				return testutils.AddChiContext(t, req, map[string]string{"namespace": "test-namespace", "name": "testmod", "provider": "testprovider"})
			},
			expectedStatus: http.StatusForbidden,
		},
		{
			name: "user with FULL permission can create module provider",
			setupAuth: func(t *testing.T, db *sqldb.Database) *http.Request {
				req, _ := testutils.BuildAuthenticatedRequestWithNamespacePermission(
					t, db, "POST", "/v1/terrareg/modules/test-namespace/testmod-full/testprovider/create",
					"full-user", "test-namespace", sqldb.PermissionTypeFull,
				)
				return testutils.AddChiContext(t, req, map[string]string{"namespace": "test-namespace", "name": "testmod-full", "provider": "testprovider"})
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name: "admin user can create any module provider",
			setupAuth: func(t *testing.T, db *sqldb.Database) *http.Request {
				req, _ := testutils.BuildAdminRequest(t, db, "POST", "/v1/terrareg/modules/test-namespace/testmod-admin/testprovider/create")
				return testutils.AddChiContext(t, req, map[string]string{"namespace": "test-namespace", "name": "testmod-admin", "provider": "testprovider"})
			},
			expectedStatus: http.StatusCreated,
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

// TestModuleProviderDelete_Authentication tests module provider deletion with RequireNamespacePermission(FULL) middleware
func TestModuleProviderDelete_Authentication(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	// Enable RBAC for this test so permission checking works properly
	cont := testutils.CreateTestServerWithConfig(t, db, testutils.WithEnableAccessControls(true))
	router := cont.Router

	// Create test namespace and module provider
	namespace := testutils.CreateNamespace(t, db, "delete-namespace", nil)
	_ = testutils.CreateModuleProvider(t, db, namespace.ID, "testmod", "testprovider")

	tests := []struct {
		name           string
		setupAuth      func(*testing.T, *sqldb.Database) *http.Request
		expectedStatus int
	}{
		{
			name: "unauthenticated request returns 401",
			setupAuth: func(t *testing.T, db *sqldb.Database) *http.Request {
				req := testutils.BuildUnauthenticatedRequest(t, "DELETE", "/v1/terrareg/modules/delete-namespace/testmod/testprovider/delete")
				return testutils.AddChiContext(t, req, map[string]string{"namespace": "delete-namespace", "name": "testmod", "provider": "testprovider"})
			},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name: "user with READ permission returns 403",
			setupAuth: func(t *testing.T, db *sqldb.Database) *http.Request {
				req, _ := testutils.BuildAuthenticatedRequestWithNamespacePermission(
					t, db, "DELETE", "/v1/terrareg/modules/delete-namespace/testmod/testprovider/delete",
					"readonly-user", "delete-namespace", sqldb.PermissionTypeRead,
				)
				return testutils.AddChiContext(t, req, map[string]string{"namespace": "delete-namespace", "name": "testmod", "provider": "testprovider"})
			},
			expectedStatus: http.StatusForbidden,
		},
		{
			name: "user with MODIFY permission returns 403",
			setupAuth: func(t *testing.T, db *sqldb.Database) *http.Request {
				req, _ := testutils.BuildAuthenticatedRequestWithNamespacePermission(
					t, db, "DELETE", "/v1/terrareg/modules/delete-namespace/testmod/testprovider/delete",
					"modify-user", "delete-namespace", sqldb.PermissionTypeModify,
				)
				return testutils.AddChiContext(t, req, map[string]string{"namespace": "delete-namespace", "name": "testmod", "provider": "testprovider"})
			},
			expectedStatus: http.StatusForbidden,
		},
		{
			name: "user with FULL permission can delete module provider",
			setupAuth: func(t *testing.T, db *sqldb.Database) *http.Request {
				req, _ := testutils.BuildAuthenticatedRequestWithNamespacePermission(
					t, db, "DELETE", "/v1/terrareg/modules/delete-namespace/testmod/testprovider/delete",
					"full-user", "delete-namespace", sqldb.PermissionTypeFull,
				)
				return testutils.AddChiContext(t, req, map[string]string{"namespace": "delete-namespace", "name": "testmod", "provider": "testprovider"})
			},
			expectedStatus: http.StatusNoContent,
		},
		{
			name: "admin user can delete any module provider",
			setupAuth: func(t *testing.T, db *sqldb.Database) *http.Request {
				// Recreate module provider since it was deleted by previous test
				var namespaceDB sqldb.NamespaceDB
				err := db.DB.Where("namespace = ?", "delete-namespace").First(&namespaceDB).Error
				if err == nil {
					testutils.CreateModuleProvider(t, db, namespaceDB.ID, "testmod", "testprovider")
				}
				req, _ := testutils.BuildAdminRequest(t, db, "DELETE", "/v1/terrareg/modules/delete-namespace/testmod/testprovider/delete")
				return testutils.AddChiContext(t, req, map[string]string{"namespace": "delete-namespace", "name": "testmod", "provider": "testprovider"})
			},
			expectedStatus: http.StatusNoContent,
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

// TestModuleProviderSettingsUpdate_Authentication tests module settings update with RequireNamespacePermission(MODIFY) middleware
func TestModuleProviderSettingsUpdate_Authentication(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	// Enable RBAC for this test so permission checking works properly
	cont := testutils.CreateTestServerWithConfig(t, db, testutils.WithEnableAccessControls(true))
	router := cont.Router

	// Create test namespace and module provider
	namespace := testutils.CreateNamespace(t, db, "settings-namespace", nil)
	_ = testutils.CreateModuleProvider(t, db, namespace.ID, "testmod", "testprovider")

	tests := []struct {
		name           string
		setupAuth      func(*testing.T, *sqldb.Database) *http.Request
		expectedStatus int
	}{
		{
			name: "unauthenticated request returns 401",
			setupAuth: func(t *testing.T, db *sqldb.Database) *http.Request {
				req := testutils.BuildUnauthenticatedRequest(t, "POST", "/v1/terrareg/modules/settings-namespace/testmod/testprovider/settings")
				return testutils.AddChiContext(t, req, map[string]string{"namespace": "settings-namespace", "name": "testmod", "provider": "testprovider"})
			},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name: "user with READ permission returns 403",
			setupAuth: func(t *testing.T, db *sqldb.Database) *http.Request {
				req, _ := testutils.BuildAuthenticatedRequestWithNamespacePermission(
					t, db, "POST", "/v1/terrareg/modules/settings-namespace/testmod/testprovider/settings",
					"readonly-user", "settings-namespace", sqldb.PermissionTypeRead,
				)
				return testutils.AddChiContext(t, req, map[string]string{"namespace": "settings-namespace", "name": "testmod", "provider": "testprovider"})
			},
			expectedStatus: http.StatusForbidden,
		},
		{
			name: "user with MODIFY permission can update settings",
			setupAuth: func(t *testing.T, db *sqldb.Database) *http.Request {
				req, _ := testutils.BuildAuthenticatedRequestWithNamespacePermission(
					t, db, "POST", "/v1/terrareg/modules/settings-namespace/testmod/testprovider/settings",
					"modify-user", "settings-namespace", sqldb.PermissionTypeModify,
				)
				// Add request body for settings update
				req.Body = io.NopCloser(strings.NewReader(`{"csrf_token":"test-token"}`))
				req.Header.Set("Content-Type", "application/json")
				return testutils.AddChiContext(t, req, map[string]string{"namespace": "settings-namespace", "name": "testmod", "provider": "testprovider"})
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "user with FULL permission can update settings",
			setupAuth: func(t *testing.T, db *sqldb.Database) *http.Request {
				req, _ := testutils.BuildAuthenticatedRequestWithNamespacePermission(
					t, db, "POST", "/v1/terrareg/modules/settings-namespace/testmod/testprovider/settings",
					"full-user", "settings-namespace", sqldb.PermissionTypeFull,
				)
				// Add request body for settings update
				req.Body = io.NopCloser(strings.NewReader(`{"csrf_token":"test-token"}`))
				req.Header.Set("Content-Type", "application/json")
				return testutils.AddChiContext(t, req, map[string]string{"namespace": "settings-namespace", "name": "testmod", "provider": "testprovider"})
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "admin user can update any module settings",
			setupAuth: func(t *testing.T, db *sqldb.Database) *http.Request {
				req, _ := testutils.BuildAdminRequest(t, db, "POST", "/v1/terrareg/modules/settings-namespace/testmod/testprovider/settings")
				// Add request body for settings update
				req.Body = io.NopCloser(strings.NewReader(`{"csrf_token":"test-token"}`))
				req.Header.Set("Content-Type", "application/json")
				return testutils.AddChiContext(t, req, map[string]string{"namespace": "settings-namespace", "name": "testmod", "provider": "testprovider"})
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

// TestModuleVersionDelete_Authentication tests module version deletion with RequireNamespacePermission(FULL) middleware
// This matches the Python test: test/unit/terrareg/server/test_api_terrareg_module_version_delete.py
func TestModuleVersionDelete_Authentication(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	// Enable RBAC for this test so permission checking works properly
	cont := testutils.CreateTestServerWithConfig(t, db, testutils.WithEnableAccessControls(true))
	router := cont.Router

	// Create test data: namespace -> module provider -> version
	namespace := testutils.CreateNamespace(t, db, "version-delete-namespace", nil)
	moduleProvider := testutils.CreateModuleProvider(t, db, namespace.ID, "testmodulename", "testprovider")
	_ = testutils.CreateModuleVersion(t, db, moduleProvider.ID, "2.4.1")

	tests := []struct {
		name           string
		setupAuth      func(*testing.T, *sqldb.Database) *http.Request
		expectedStatus int
	}{
		{
			name: "unauthenticated request returns 401",
			setupAuth: func(t *testing.T, db *sqldb.Database) *http.Request {
				req := testutils.BuildUnauthenticatedRequest(t, "DELETE", "/v1/terrareg/modules/version-delete-namespace/testmodulename/testprovider/2.4.1/delete")
				return testutils.AddChiContext(t, req, map[string]string{"namespace": "version-delete-namespace", "name": "testmodulename", "provider": "testprovider", "version": "2.4.1"})
			},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name: "user with READ permission returns 403",
			setupAuth: func(t *testing.T, db *sqldb.Database) *http.Request {
				req, _ := testutils.BuildAuthenticatedRequestWithNamespacePermission(
					t, db, "DELETE", "/v1/terrareg/modules/version-delete-namespace/testmodulename/testprovider/2.4.1/delete",
					"readonly-user", "version-delete-namespace", sqldb.PermissionTypeRead,
				)
				return testutils.AddChiContext(t, req, map[string]string{"namespace": "version-delete-namespace", "name": "testmodulename", "provider": "testprovider", "version": "2.4.1"})
			},
			expectedStatus: http.StatusForbidden,
		},
		{
			name: "user with MODIFY permission returns 403",
			setupAuth: func(t *testing.T, db *sqldb.Database) *http.Request {
				req, _ := testutils.BuildAuthenticatedRequestWithNamespacePermission(
					t, db, "DELETE", "/v1/terrareg/modules/version-delete-namespace/testmodulename/testprovider/2.4.1/delete",
					"modify-user", "version-delete-namespace", sqldb.PermissionTypeModify,
				)
				return testutils.AddChiContext(t, req, map[string]string{"namespace": "version-delete-namespace", "name": "testmodulename", "provider": "testprovider", "version": "2.4.1"})
			},
			expectedStatus: http.StatusForbidden,
		},
		{
			name: "user with FULL permission can delete module version",
			setupAuth: func(t *testing.T, db *sqldb.Database) *http.Request {
				// Ensure module version exists before attempting delete
				var namespaceDB sqldb.NamespaceDB
				err := db.DB.Where("namespace = ?", "version-delete-namespace").First(&namespaceDB).Error
				if err == nil {
					var moduleProviderDB sqldb.ModuleProviderDB
					err = db.DB.Where("namespace_id = ? AND module = ? AND provider = ?", namespaceDB.ID, "testmodulename", "testprovider").First(&moduleProviderDB).Error
					if err == nil {
						// Check if version already exists, if not create it
						var versionCount int64
						db.DB.Model(&sqldb.ModuleVersionDB{}).Where("module_provider_id = ? AND version = ?", moduleProviderDB.ID, "2.4.1").Count(&versionCount)
						if versionCount == 0 {
							testutils.CreateModuleVersion(t, db, moduleProviderDB.ID, "2.4.1")
						}
					}
				}
				req, _ := testutils.BuildAuthenticatedRequestWithNamespacePermission(
					t, db, "DELETE", "/v1/terrareg/modules/version-delete-namespace/testmodulename/testprovider/2.4.1/delete",
					"full-user", "version-delete-namespace", sqldb.PermissionTypeFull,
				)
				return testutils.AddChiContext(t, req, map[string]string{"namespace": "version-delete-namespace", "name": "testmodulename", "provider": "testprovider", "version": "2.4.1"})
			},
			expectedStatus: http.StatusOK, // Python returns 200 with {'status': 'Success'}
		},
		{
			name: "admin user can delete any module version",
			setupAuth: func(t *testing.T, db *sqldb.Database) *http.Request {
				// Recreate module version since it was deleted by previous test
				var namespaceDB sqldb.NamespaceDB
				err := db.DB.Where("namespace = ?", "version-delete-namespace").First(&namespaceDB).Error
				if err == nil {
					var moduleProviderDB sqldb.ModuleProviderDB
					err = db.DB.Where("namespace_id = ? AND module = ? AND provider = ?", namespaceDB.ID, "testmodulename", "testprovider").First(&moduleProviderDB).Error
					if err == nil {
						testutils.CreateModuleVersion(t, db, moduleProviderDB.ID, "2.4.1")
					}
				}
				req, _ := testutils.BuildAdminRequest(t, db, "DELETE", "/v1/terrareg/modules/version-delete-namespace/testmodulename/testprovider/2.4.1/delete")
				return testutils.AddChiContext(t, req, map[string]string{"namespace": "version-delete-namespace", "name": "testmodulename", "provider": "testprovider", "version": "2.4.1"})
			},
			expectedStatus: http.StatusOK, // Python returns 200 with {'status': 'Success'}
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

// TestModuleVersionPublish_Authentication tests module version publish with RequireAuth middleware
func TestModuleVersionPublish_Authentication(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	// Enable RBAC for this test so permission checking works properly
	cont := testutils.CreateTestServerWithConfig(t, db, testutils.WithEnableAccessControls(true))
	router := cont.Router

	// Create test data
	namespace := testutils.CreateNamespace(t, db, "publish-namespace", nil)
	moduleProvider := testutils.CreateModuleProvider(t, db, namespace.ID, "testmod", "testprovider")
	_ = testutils.CreateModuleVersion(t, db, moduleProvider.ID, "1.0.0")

	tests := []struct {
		name           string
		setupAuth      func(*testing.T, *sqldb.Database) *http.Request
		expectedStatus int
	}{
		{
			name: "unauthenticated request returns 401",
			setupAuth: func(t *testing.T, db *sqldb.Database) *http.Request {
				req := testutils.BuildUnauthenticatedRequest(t, "POST", "/v1/terrareg/modules/publish-namespace/testmod/testprovider/1.0.0/publish")
				return testutils.AddChiContext(t, req, map[string]string{"namespace": "publish-namespace", "name": "testmod", "provider": "testprovider", "version": "1.0.0"})
			},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name: "authenticated user can publish module version",
			setupAuth: func(t *testing.T, db *sqldb.Database) *http.Request {
				req, _ := testutils.BuildAuthenticatedRequestWithSession(t, db, "POST", "/v1/terrareg/modules/publish-namespace/testmod/testprovider/1.0.0/publish", "regular-user", false)
				return testutils.AddChiContext(t, req, map[string]string{"namespace": "publish-namespace", "name": "testmod", "provider": "testprovider", "version": "1.0.0"})
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name: "admin user can publish module version",
			setupAuth: func(t *testing.T, db *sqldb.Database) *http.Request {
				req, _ := testutils.BuildAdminRequest(t, db, "POST", "/v1/terrareg/modules/publish-namespace/testmod/testprovider/1.0.0/publish")
				return testutils.AddChiContext(t, req, map[string]string{"namespace": "publish-namespace", "name": "testmod", "provider": "testprovider", "version": "1.0.0"})
			},
			expectedStatus: http.StatusCreated,
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

// TestModuleDetails_AllAuthMethods tests GET module details endpoint with OptionalAuth
// All authentication states should return 200
func TestModuleDetails_AllAuthMethods(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	// Enable RBAC for this test so permission checking works properly
	cont := testutils.CreateTestServerWithConfig(t, db, testutils.WithEnableAccessControls(true))
	router := cont.Router

	// Create test data
	namespace := testutils.CreateNamespace(t, db, "details-namespace", nil)
	_ = testutils.CreateModuleProvider(t, db, namespace.ID, "testmod", "testprovider")

	tests := []struct {
		name           string
		setupAuth      func(*testing.T, *sqldb.Database) *http.Request
		expectedStatus int
	}{
		{
			name: "unauthenticated request returns 200",
			setupAuth: func(t *testing.T, db *sqldb.Database) *http.Request {
				req := testutils.BuildUnauthenticatedRequest(t, "GET", "/v1/terrareg/modules/details-namespace/testmod/testprovider")
				return testutils.AddChiContext(t, req, map[string]string{"namespace": "details-namespace", "name": "testmod", "provider": "testprovider"})
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "authenticated regular user returns 200",
			setupAuth: func(t *testing.T, db *sqldb.Database) *http.Request {
				req, _ := testutils.BuildAuthenticatedRequestWithSession(t, db, "GET", "/v1/terrareg/modules/details-namespace/testmod/testprovider", "regular-user", false)
				return testutils.AddChiContext(t, req, map[string]string{"namespace": "details-namespace", "name": "testmod", "provider": "testprovider"})
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "authenticated admin user returns 200",
			setupAuth: func(t *testing.T, db *sqldb.Database) *http.Request {
				req, _ := testutils.BuildAdminRequest(t, db, "GET", "/v1/terrareg/modules/details-namespace/testmod/testprovider")
				return testutils.AddChiContext(t, req, map[string]string{"namespace": "details-namespace", "name": "testmod", "provider": "testprovider"})
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
