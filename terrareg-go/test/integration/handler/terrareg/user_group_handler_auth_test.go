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

// TestUserGroupList_Authentication tests user group list with RequireAdmin middleware
// This matches the Python test: test/unit/terrareg/server/test_api_terrareg_auth_user_groups.py::test_get
func TestUserGroupList_Authentication(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	// Enable RBAC for this test so permission checking works properly
	cont := testutils.CreateTestServerWithConfig(t, db, testutils.WithEnableAccessControls(true))
	router := cont.Router

	tests := []struct {
		name           string
		setupAuth      func(*testing.T, *sqldb.Database) *http.Request
		expectedStatus int
	}{
		{
			name: "unauthenticated request returns 403",
			setupAuth: func(t *testing.T, db *sqldb.Database) *http.Request {
				return testutils.BuildUnauthenticatedRequest(t, "GET", "/v1/terrareg/user-groups")
			},
			expectedStatus: http.StatusForbidden,
		},
		{
			name: "regular authenticated user returns 403",
			setupAuth: func(t *testing.T, db *sqldb.Database) *http.Request {
				req, _ := testutils.BuildAuthenticatedRequestWithSession(t, db, "GET", "/v1/terrareg/user-groups", "regular-user", false)
				return req
			},
			expectedStatus: http.StatusForbidden,
		},
		{
			name: "admin user can access user groups",
			setupAuth: func(t *testing.T, db *sqldb.Database) *http.Request {
				req, _ := testutils.BuildAdminRequest(t, db, "GET", "/v1/terrareg/user-groups")
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

// TestUserGroupCreate_Authentication tests user group creation with RequireAdmin middleware
func TestUserGroupCreate_Authentication(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	// Enable RBAC for this test so permission checking works properly
	cont := testutils.CreateTestServerWithConfig(t, db, testutils.WithEnableAccessControls(true))
	router := cont.Router

	tests := []struct {
		name           string
		setupAuth      func(*testing.T, *sqldb.Database) *http.Request
		expectedStatus int
	}{
		{
			name: "unauthenticated request returns 403",
			setupAuth: func(t *testing.T, db *sqldb.Database) *http.Request {
				return testutils.BuildUnauthenticatedRequest(t, "POST", "/v1/terrareg/user-groups")
			},
			expectedStatus: http.StatusForbidden,
		},
		{
			name: "regular authenticated user returns 403",
			setupAuth: func(t *testing.T, db *sqldb.Database) *http.Request {
				req, _ := testutils.BuildAuthenticatedRequestWithSession(t, db, "POST", "/v1/terrareg/user-groups", "regular-user", false)
				return req
			},
			expectedStatus: http.StatusForbidden,
		},
		{
			name: "admin user can create user groups",
			setupAuth: func(t *testing.T, db *sqldb.Database) *http.Request {
				req, _ := testutils.BuildAdminRequest(t, db, "POST", "/v1/terrareg/user-groups")
				// Add request body for user group creation
				req.Body = io.NopCloser(strings.NewReader(`{"name":"test-admin-group","site_admin":false}`))
				req.Header.Set("Content-Type", "application/json")
				return req
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

// TestUserGroupDelete_Authentication tests user group deletion with RequireAdmin middleware
func TestUserGroupDelete_Authentication(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	// Enable RBAC for this test so permission checking works properly
	cont := testutils.CreateTestServerWithConfig(t, db, testutils.WithEnableAccessControls(true))
	router := cont.Router

	// Create test user group
	userGroup := testutils.CreateTestAuthUserGroup(t, db, "test-group", false)

	tests := []struct {
		name           string
		setupAuth      func(*testing.T, *sqldb.Database) *http.Request
		expectedStatus int
	}{
		{
			name: "unauthenticated request returns 403",
			setupAuth: func(t *testing.T, db *sqldb.Database) *http.Request {
				req := testutils.BuildUnauthenticatedRequest(t, "DELETE", "/v1/terrareg/user-groups/test-group")
				return testutils.AddChiContext(t, req, map[string]string{"group": "test-group"})
			},
			expectedStatus: http.StatusForbidden,
		},
		{
			name: "regular authenticated user returns 403",
			setupAuth: func(t *testing.T, db *sqldb.Database) *http.Request {
				req, _ := testutils.BuildAuthenticatedRequestWithSession(t, db, "DELETE", "/v1/terrareg/user-groups/test-group", "regular-user", false)
				return testutils.AddChiContext(t, req, map[string]string{"group": "test-group"})
			},
			expectedStatus: http.StatusForbidden,
		},
		{
			name: "admin user can delete user groups",
			setupAuth: func(t *testing.T, db *sqldb.Database) *http.Request {
				req, _ := testutils.BuildAdminRequest(t, db, "DELETE", "/v1/terrareg/user-groups/test-group")
				return testutils.AddChiContext(t, req, map[string]string{"group": "test-group"})
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

	_ = userGroup
}

// TestUserGroupNamespacePermissionsCreate_Authentication tests namespace permission creation with RequireAdmin middleware
func TestUserGroupNamespacePermissionsCreate_Authentication(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	// Enable RBAC for this test so permission checking works properly
	cont := testutils.CreateTestServerWithConfig(t, db, testutils.WithEnableAccessControls(true))
	router := cont.Router

	// Create test user group and namespace
	userGroup := testutils.CreateTestAuthUserGroup(t, db, "perm-test-group", false)
	_ = testutils.CreateNamespace(t, db, "perm-test-namespace", nil)

	tests := []struct {
		name           string
		setupAuth      func(*testing.T, *sqldb.Database) *http.Request
		expectedStatus int
	}{
		{
			name: "unauthenticated request returns 403",
			setupAuth: func(t *testing.T, db *sqldb.Database) *http.Request {
				req := testutils.BuildUnauthenticatedRequest(t, "POST", "/v1/terrareg/user-groups/perm-test-group/permissions/perm-test-namespace")
				return testutils.AddChiContext(t, req, map[string]string{"group": "perm-test-group", "namespace": "perm-test-namespace"})
			},
			expectedStatus: http.StatusForbidden,
		},
		{
			name: "regular authenticated user returns 403",
			setupAuth: func(t *testing.T, db *sqldb.Database) *http.Request {
				req, _ := testutils.BuildAuthenticatedRequestWithSession(t, db, "POST", "/v1/terrareg/user-groups/perm-test-group/permissions/perm-test-namespace", "regular-user", false)
				return testutils.AddChiContext(t, req, map[string]string{"group": "perm-test-group", "namespace": "perm-test-namespace"})
			},
			expectedStatus: http.StatusForbidden,
		},
		{
			name: "admin user can create namespace permissions",
			setupAuth: func(t *testing.T, db *sqldb.Database) *http.Request {
				req, _ := testutils.BuildAdminRequest(t, db, "POST", "/v1/terrareg/user-groups/perm-test-group/permissions/perm-test-namespace")
				// Add request body for namespace permission creation
				req.Body = io.NopCloser(strings.NewReader(`{"permission_type":"FULL","csrf_token":"test-token"}`))
				req.Header.Set("Content-Type", "application/json")
				return testutils.AddChiContext(t, req, map[string]string{"group": "perm-test-group", "namespace": "perm-test-namespace"})
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

	_ = userGroup
}

// TestUserGroupNamespacePermissionsDelete_Authentication tests namespace permission deletion with RequireAdmin middleware
func TestUserGroupNamespacePermissionsDelete_Authentication(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	// Enable RBAC for this test so permission checking works properly
	cont := testutils.CreateTestServerWithConfig(t, db, testutils.WithEnableAccessControls(true))
	router := cont.Router

	// Create test user group and namespace with permission
	userGroup := testutils.CreateTestAuthUserGroup(t, db, "perm-delete-group", false)
	namespace := testutils.CreateNamespace(t, db, "perm-delete-namespace", nil)
	_ = testutils.CreateTestNamespacePermission(t, db, userGroup.ID, namespace.ID, sqldb.PermissionTypeFull)

	tests := []struct {
		name           string
		setupAuth      func(*testing.T, *sqldb.Database) *http.Request
		expectedStatus int
	}{
		{
			name: "unauthenticated request returns 403",
			setupAuth: func(t *testing.T, db *sqldb.Database) *http.Request {
				req := testutils.BuildUnauthenticatedRequest(t, "DELETE", "/v1/terrareg/user-groups/perm-delete-group/permissions/perm-delete-namespace")
				return testutils.AddChiContext(t, req, map[string]string{"group": "perm-delete-group", "namespace": "perm-delete-namespace"})
			},
			expectedStatus: http.StatusForbidden,
		},
		{
			name: "regular authenticated user returns 403",
			setupAuth: func(t *testing.T, db *sqldb.Database) *http.Request {
				req, _ := testutils.BuildAuthenticatedRequestWithSession(t, db, "DELETE", "/v1/terrareg/user-groups/perm-delete-group/permissions/perm-delete-namespace", "regular-user", false)
				return testutils.AddChiContext(t, req, map[string]string{"group": "perm-delete-group", "namespace": "perm-delete-namespace"})
			},
			expectedStatus: http.StatusForbidden,
		},
		{
			name: "admin user can delete namespace permissions",
			setupAuth: func(t *testing.T, db *sqldb.Database) *http.Request {
				req, _ := testutils.BuildAdminRequest(t, db, "DELETE", "/v1/terrareg/user-groups/perm-delete-group/permissions/perm-delete-namespace")
				return testutils.AddChiContext(t, req, map[string]string{"group": "perm-delete-group", "namespace": "perm-delete-namespace"})
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
