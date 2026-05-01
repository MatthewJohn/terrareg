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

// TestNamespaceCreate_Authentication tests namespace creation with RequireAuth middleware
func TestNamespaceCreate_Authentication(t *testing.T) {
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
			name: "unauthenticated request returns 401",
			setupAuth: func(t *testing.T, db *sqldb.Database) *http.Request {
				return testutils.BuildUnauthenticatedRequest(t, "POST", "/v1/terrareg/namespaces")
			},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name: "authenticated regular user can create namespace",
			setupAuth: func(t *testing.T, db *sqldb.Database) *http.Request {
				req, _ := testutils.BuildAuthenticatedRequestWithSession(t, db, "POST", "/v1/terrareg/namespaces", "regular-user", false)
				// Add request body for namespace creation
				req.Body = io.NopCloser(strings.NewReader(`{"name":"test-namespace-auth","type":"NONE"}`))
				req.Header.Set("Content-Type", "application/json")
				return req
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "authenticated admin user can create namespace",
			setupAuth: func(t *testing.T, db *sqldb.Database) *http.Request {
				req, _ := testutils.BuildAdminRequest(t, db, "POST", "/v1/terrareg/namespaces")
				// Add request body for namespace creation
				req.Body = io.NopCloser(strings.NewReader(`{"name":"admin-test-namespace","type":"NONE"}`))
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

// TestNamespaceUpdate_Authentication tests namespace update with RequireNamespacePermission(FULL) middleware
func TestNamespaceUpdate_Authentication(t *testing.T) {
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
				req := testutils.BuildUnauthenticatedRequest(t, "POST", "/v1/terrareg/namespaces/test-namespace")
				return testutils.AddChiContext(t, req, map[string]string{"namespace": "test-namespace"})
			},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name: "user with READ permission returns 403",
			setupAuth: func(t *testing.T, db *sqldb.Database) *http.Request {
				req, _ := testutils.BuildAuthenticatedRequestWithNamespacePermission(
					t, db, "POST", "/v1/terrareg/namespaces/test-namespace",
					"readonly-user", "test-namespace", sqldb.PermissionTypeRead,
				)
				req = testutils.AddChiContext(t, req, map[string]string{"namespace": "test-namespace"})
				// Add request body for namespace update
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
					t, db, "POST", "/v1/terrareg/namespaces/test-namespace",
					"modify-user", "test-namespace", sqldb.PermissionTypeModify,
				)
				req = testutils.AddChiContext(t, req, map[string]string{"namespace": "test-namespace"})
				// Add request body for namespace update
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
					t, db, "POST", "/v1/terrareg/namespaces/test-namespace",
					"full-user", "test-namespace", sqldb.PermissionTypeFull,
				)
				req = testutils.AddChiContext(t, req, map[string]string{"namespace": "test-namespace"})
				// Add request body for namespace update
				req.Body = io.NopCloser(strings.NewReader(`{"csrf_token":"test-token"}`))
				req.Header.Set("Content-Type", "application/json")
				return req
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "admin user can update any namespace",
			setupAuth: func(t *testing.T, db *sqldb.Database) *http.Request {
				req, _ := testutils.BuildAdminRequest(t, db, "POST", "/v1/terrareg/namespaces/test-namespace")
				req = testutils.AddChiContext(t, req, map[string]string{"namespace": "test-namespace"})
				// Add request body for namespace update
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

// TestNamespaceDelete_Authentication tests namespace deletion with RequireNamespacePermission(FULL) middleware
func TestNamespaceDelete_Authentication(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	// Enable RBAC for this test so permission checking works properly
	cont := testutils.CreateTestServerWithConfig(t, db, testutils.WithEnableAccessControls(true))
	router := cont.Router

	// Create test namespace
	_ = testutils.CreateNamespace(t, db, "delete-test-namespace", nil)

	tests := []struct {
		name           string
		setupAuth      func(*testing.T, *sqldb.Database) *http.Request
		expectedStatus int
	}{
		{
			name: "unauthenticated request returns 401",
			setupAuth: func(t *testing.T, db *sqldb.Database) *http.Request {
				req := testutils.BuildUnauthenticatedRequest(t, "DELETE", "/v1/terrareg/namespaces/delete-test-namespace")
				return testutils.AddChiContext(t, req, map[string]string{"namespace": "delete-test-namespace"})
			},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name: "user with READ permission returns 403",
			setupAuth: func(t *testing.T, db *sqldb.Database) *http.Request {
				req, _ := testutils.BuildAuthenticatedRequestWithNamespacePermission(
					t, db, "DELETE", "/v1/terrareg/namespaces/delete-test-namespace",
					"readonly-user", "delete-test-namespace", sqldb.PermissionTypeRead,
				)
				return testutils.AddChiContext(t, req, map[string]string{"namespace": "delete-test-namespace"})
			},
			expectedStatus: http.StatusForbidden,
		},
		{
			name: "user with MODIFY permission returns 403",
			setupAuth: func(t *testing.T, db *sqldb.Database) *http.Request {
				req, _ := testutils.BuildAuthenticatedRequestWithNamespacePermission(
					t, db, "DELETE", "/v1/terrareg/namespaces/delete-test-namespace",
					"modify-user", "delete-test-namespace", sqldb.PermissionTypeModify,
				)
				return testutils.AddChiContext(t, req, map[string]string{"namespace": "delete-test-namespace"})
			},
			expectedStatus: http.StatusForbidden,
		},
		{
			name: "user with FULL permission can delete namespace",
			setupAuth: func(t *testing.T, db *sqldb.Database) *http.Request {
				req, _ := testutils.BuildAuthenticatedRequestWithNamespacePermission(
					t, db, "DELETE", "/v1/terrareg/namespaces/delete-test-namespace",
					"full-user", "delete-test-namespace", sqldb.PermissionTypeFull,
				)
				return testutils.AddChiContext(t, req, map[string]string{"namespace": "delete-test-namespace"})
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "admin user can delete any namespace",
			setupAuth: func(t *testing.T, db *sqldb.Database) *http.Request {
				// Recreate namespace since it may have been deleted by previous test
				_ = testutils.CreateNamespace(t, db, "delete-test-namespace", nil)
				req, _ := testutils.BuildAdminRequest(t, db, "DELETE", "/v1/terrareg/namespaces/delete-test-namespace")
				return testutils.AddChiContext(t, req, map[string]string{"namespace": "delete-test-namespace"})
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

// TestNamespaceGet_AllAuthMethods tests GET namespace endpoint with OptionalAuth
// All authentication states should return 200
func TestNamespaceGet_AllAuthMethods(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	// Enable RBAC for this test so permission checking works properly
	cont := testutils.CreateTestServerWithConfig(t, db, testutils.WithEnableAccessControls(true))
	router := cont.Router

	// Create test namespace
	_ = testutils.CreateNamespace(t, db, "get-test-namespace", nil)

	tests := []struct {
		name           string
		setupAuth      func(*testing.T, *sqldb.Database) *http.Request
		expectedStatus int
	}{
		{
			name: "unauthenticated request returns 200",
			setupAuth: func(t *testing.T, db *sqldb.Database) *http.Request {
				req := testutils.BuildUnauthenticatedRequest(t, "GET", "/v1/terrareg/namespaces/get-test-namespace")
				return testutils.AddChiContext(t, req, map[string]string{"namespace": "get-test-namespace"})
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "authenticated regular user returns 200",
			setupAuth: func(t *testing.T, db *sqldb.Database) *http.Request {
				req, _ := testutils.BuildAuthenticatedRequestWithSession(t, db, "GET", "/v1/terrareg/namespaces/get-test-namespace", "regular-user", false)
				return testutils.AddChiContext(t, req, map[string]string{"namespace": "get-test-namespace"})
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "authenticated admin user returns 200",
			setupAuth: func(t *testing.T, db *sqldb.Database) *http.Request {
				req, _ := testutils.BuildAdminRequest(t, db, "GET", "/v1/terrareg/namespaces/get-test-namespace")
				return testutils.AddChiContext(t, req, map[string]string{"namespace": "get-test-namespace"})
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

// TestNamespaceList_AllAuthMethods tests GET namespace list endpoint with OptionalAuth
// All authentication states should return 200
func TestNamespaceList_AllAuthMethods(t *testing.T) {
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
			name: "unauthenticated request returns 200",
			setupAuth: func(t *testing.T, db *sqldb.Database) *http.Request {
				return testutils.BuildUnauthenticatedRequest(t, "GET", "/v1/terrareg/namespaces")
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "authenticated regular user returns 200",
			setupAuth: func(t *testing.T, db *sqldb.Database) *http.Request {
				req, _ := testutils.BuildAuthenticatedRequestWithSession(t, db, "GET", "/v1/terrareg/namespaces", "regular-user", false)
				return req
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "authenticated admin user returns 200",
			setupAuth: func(t *testing.T, db *sqldb.Database) *http.Request {
				req, _ := testutils.BuildAdminRequest(t, db, "GET", "/v1/terrareg/namespaces")
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
