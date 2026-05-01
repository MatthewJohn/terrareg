package terrareg_test

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/persistence/sqldb"
	"github.com/matthewjohn/terrareg/terrareg-go/test/integration/testutils"
)

// TestModuleVersionUpload_Authentication tests module upload with RequireUploadPermission middleware
func TestModuleVersionUpload_Authentication(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	// Enable RBAC for this test so permission checking works properly
	cont := testutils.CreateTestServerWithConfig(t, db, testutils.WithEnableAccessControls(true))
	router := cont.Router

	// Create test namespace
	_ = testutils.CreateNamespace(t, db, "upload-namespace", nil)

	tests := []struct {
		name           string
		setupAuth      func(*testing.T, *sqldb.Database) *http.Request
		expectedStatus int
	}{
		{
			name: "unauthenticated request returns 401",
			setupAuth: func(t *testing.T, db *sqldb.Database) *http.Request {
				req := testutils.BuildUnauthenticatedRequest(t, "POST", "/v1/terrareg/modules/upload-namespace/testmod/testprovider/1.0.0/upload")
				return testutils.AddChiContext(t, req, map[string]string{"namespace": "upload-namespace", "name": "testmod", "provider": "testprovider", "version": "1.0.0"})
			},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name: "user with READ permission returns 403",
			setupAuth: func(t *testing.T, db *sqldb.Database) *http.Request {
				req, _ := testutils.BuildAuthenticatedRequestWithNamespacePermission(
					t, db, "POST", "/v1/terrareg/modules/upload-namespace/testmod/testprovider/1.0.0/upload",
					"readonly-user", "upload-namespace", sqldb.PermissionTypeRead,
				)
				return testutils.AddChiContext(t, req, map[string]string{"namespace": "upload-namespace", "name": "testmod", "provider": "testprovider", "version": "1.0.0"})
			},
			expectedStatus: http.StatusForbidden,
		},
		{
			name: "user with MODIFY permission can upload",
			setupAuth: func(t *testing.T, db *sqldb.Database) *http.Request {
				req, _ := testutils.BuildAuthenticatedRequestWithNamespacePermission(
					t, db, "POST", "/v1/terrareg/modules/upload-namespace/testmod/testprovider/1.0.0/upload",
					"modify-user", "upload-namespace", sqldb.PermissionTypeModify,
				)
				return testutils.AddChiContext(t, req, map[string]string{"namespace": "upload-namespace", "name": "testmod", "provider": "testprovider", "version": "1.0.0"})
			},
			expectedStatus: http.StatusBadRequest, // Auth passes but handler returns 400 due to missing file
		},
		{
			name: "user with FULL permission can upload",
			setupAuth: func(t *testing.T, db *sqldb.Database) *http.Request {
				req, _ := testutils.BuildAuthenticatedRequestWithNamespacePermission(
					t, db, "POST", "/v1/terrareg/modules/upload-namespace/testmod/testprovider/1.0.0/upload",
					"full-user", "upload-namespace", sqldb.PermissionTypeFull,
				)
				return testutils.AddChiContext(t, req, map[string]string{"namespace": "upload-namespace", "name": "testmod", "provider": "testprovider", "version": "1.0.0"})
			},
			expectedStatus: http.StatusBadRequest, // Auth passes but handler returns 400 due to missing file
		},
		{
			name: "admin user can upload to any namespace",
			setupAuth: func(t *testing.T, db *sqldb.Database) *http.Request {
				req, _ := testutils.BuildAdminRequest(t, db, "POST", "/v1/terrareg/modules/upload-namespace/testmod/testprovider/1.0.0/upload")
				return testutils.AddChiContext(t, req, map[string]string{"namespace": "upload-namespace", "name": "testmod", "provider": "testprovider", "version": "1.0.0"})
			},
			expectedStatus: http.StatusBadRequest, // Auth passes but handler returns 400 due to missing file
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

// TestModuleImport_Authentication tests module import with RequireUploadPermission middleware
func TestModuleImport_Authentication(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	// Enable RBAC for this test so permission checking works properly
	cont := testutils.CreateTestServerWithConfig(t, db, testutils.WithEnableAccessControls(true))
	router := cont.Router

	// Create test namespace
	_ = testutils.CreateNamespace(t, db, "import-namespace", nil)

	tests := []struct {
		name           string
		setupAuth      func(*testing.T, *sqldb.Database) *http.Request
		expectedStatus int
	}{
		{
			name: "unauthenticated request returns 401",
			setupAuth: func(t *testing.T, db *sqldb.Database) *http.Request {
				req := testutils.BuildUnauthenticatedRequest(t, "POST", "/v1/terrareg/modules/import-namespace/testmod/testprovider/import")
				return testutils.AddChiContext(t, req, map[string]string{"namespace": "import-namespace", "name": "testmod", "provider": "testprovider"})
			},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name: "user with READ permission returns 403",
			setupAuth: func(t *testing.T, db *sqldb.Database) *http.Request {
				req, _ := testutils.BuildAuthenticatedRequestWithNamespacePermission(
					t, db, "POST", "/v1/terrareg/modules/import-namespace/testmod/testprovider/import",
					"readonly-user", "import-namespace", sqldb.PermissionTypeRead,
				)
				return testutils.AddChiContext(t, req, map[string]string{"namespace": "import-namespace", "name": "testmod", "provider": "testprovider"})
			},
			expectedStatus: http.StatusForbidden,
		},
		{
			name: "user with MODIFY permission can import",
			setupAuth: func(t *testing.T, db *sqldb.Database) *http.Request {
				req, _ := testutils.BuildAuthenticatedRequestWithNamespacePermission(
					t, db, "POST", "/v1/terrareg/modules/import-namespace/testmod/testprovider/import",
					"modify-user", "import-namespace", sqldb.PermissionTypeModify,
				)
				return testutils.AddChiContext(t, req, map[string]string{"namespace": "import-namespace", "name": "testmod", "provider": "testprovider"})
			},
			expectedStatus: http.StatusBadRequest, // Auth passes but handler returns 400 due to missing data
		},
		{
			name: "user with FULL permission can import",
			setupAuth: func(t *testing.T, db *sqldb.Database) *http.Request {
				req, _ := testutils.BuildAuthenticatedRequestWithNamespacePermission(
					t, db, "POST", "/v1/terrareg/modules/import-namespace/testmod/testprovider/import",
					"full-user", "import-namespace", sqldb.PermissionTypeFull,
				)
				return testutils.AddChiContext(t, req, map[string]string{"namespace": "import-namespace", "name": "testmod", "provider": "testprovider"})
			},
			expectedStatus: http.StatusBadRequest, // Auth passes but handler returns 400 due to missing data
		},
		{
			name: "admin user can import to any namespace",
			setupAuth: func(t *testing.T, db *sqldb.Database) *http.Request {
				req, _ := testutils.BuildAdminRequest(t, db, "POST", "/v1/terrareg/modules/import-namespace/testmod/testprovider/import")
				return testutils.AddChiContext(t, req, map[string]string{"namespace": "import-namespace", "name": "testmod", "provider": "testprovider"})
			},
			expectedStatus: http.StatusBadRequest, // Auth passes but handler returns 400 due to missing data
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
