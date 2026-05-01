package terrareg_test

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/persistence/sqldb"
	"github.com/matthewjohn/terrareg/terrareg-go/test/integration/testutils"
)

// TestProviderSourceOrganizations_Authentication tests provider source organizations endpoint with RequireAuth middleware
func TestProviderSourceOrganizations_Authentication(t *testing.T) {
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
				req := testutils.BuildUnauthenticatedRequest(t, "GET", "/github/organizations")
				return testutils.AddChiContext(t, req, map[string]string{"provider_source": "github"})
			},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name: "authenticated user gets 404 when provider source doesn't exist",
			setupAuth: func(t *testing.T, db *sqldb.Database) *http.Request {
				req, _ := testutils.BuildAuthenticatedRequestWithSession(t, db, "GET", "/github/organizations", "regular-user", false)
				return testutils.AddChiContext(t, req, map[string]string{"provider_source": "github"})
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name: "admin user gets 404 when provider source doesn't exist",
			setupAuth: func(t *testing.T, db *sqldb.Database) *http.Request {
				req, _ := testutils.BuildAdminRequest(t, db, "GET", "/github/organizations")
				return testutils.AddChiContext(t, req, map[string]string{"provider_source": "github"})
			},
			expectedStatus: http.StatusNotFound,
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

// TestProviderSourceRepositories_Authentication tests provider source repositories endpoint with RequireAuth middleware
func TestProviderSourceRepositories_Authentication(t *testing.T) {
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
				req := testutils.BuildUnauthenticatedRequest(t, "GET", "/github/repositories")
				return testutils.AddChiContext(t, req, map[string]string{"provider_source": "github"})
			},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name: "authenticated user gets 404 when provider source doesn't exist",
			setupAuth: func(t *testing.T, db *sqldb.Database) *http.Request {
				req, _ := testutils.BuildAuthenticatedRequestWithSession(t, db, "GET", "/github/repositories", "regular-user", false)
				return testutils.AddChiContext(t, req, map[string]string{"provider_source": "github"})
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name: "admin user gets 404 when provider source doesn't exist",
			setupAuth: func(t *testing.T, db *sqldb.Database) *http.Request {
				req, _ := testutils.BuildAdminRequest(t, db, "GET", "/github/repositories")
				return testutils.AddChiContext(t, req, map[string]string{"provider_source": "github"})
			},
			expectedStatus: http.StatusNotFound,
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

// TestProviderSourceRefreshNamespace_Authentication tests provider source refresh namespace endpoint with RequireAuth middleware
func TestProviderSourceRefreshNamespace_Authentication(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	// Enable RBAC for this test so permission checking works properly
	cont := testutils.CreateTestServerWithConfig(t, db, testutils.WithEnableAccessControls(true))
	router := cont.Router

	// Create test namespace
	_ = testutils.CreateNamespace(t, db, "refresh-namespace", nil)

	tests := []struct {
		name           string
		setupAuth      func(*testing.T, *sqldb.Database) *http.Request
		expectedStatus int
	}{
		{
			name: "unauthenticated request returns 403 (requires admin)",
			setupAuth: func(t *testing.T, db *sqldb.Database) *http.Request {
				req := testutils.BuildUnauthenticatedRequest(t, "POST", "/github/refresh-namespace")
				return testutils.AddChiContext(t, req, map[string]string{"provider_source": "github"})
			},
			expectedStatus: http.StatusForbidden,
		},
		{
			name: "authenticated user gets 403 (requires admin)",
			setupAuth: func(t *testing.T, db *sqldb.Database) *http.Request {
				reqBody, _ := json.Marshal(map[string]string{"namespace": "refresh-namespace"})
				req, _ := testutils.BuildAuthenticatedRequestWithSession(t, db, "POST", "/github/refresh-namespace", "regular-user", false)
				req.Body = io.NopCloser(strings.NewReader(string(reqBody)))
				return testutils.AddChiContext(t, req, map[string]string{"provider_source": "github"})
			},
			expectedStatus: http.StatusForbidden,
		},
		{
			name: "admin user gets 404 when provider source doesn't exist",
			setupAuth: func(t *testing.T, db *sqldb.Database) *http.Request {
				reqBody, _ := json.Marshal(map[string]string{"namespace": "refresh-namespace"})
				req, _ := testutils.BuildAdminRequest(t, db, "POST", "/github/refresh-namespace")
				req.Body = io.NopCloser(strings.NewReader(string(reqBody)))
				return testutils.AddChiContext(t, req, map[string]string{"provider_source": "github"})
			},
			expectedStatus: http.StatusNotFound,
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

// TestProviderSourcePublishProvider_Authentication tests provider source publish provider endpoint
// Python reference: github_repository_publish_provider.py - requires GitHub auth OR admin
func TestProviderSourcePublishProvider_Authentication(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	// Enable RBAC for this test so permission checking works properly
	cont := testutils.CreateTestServerWithConfig(t, db, testutils.WithEnableAccessControls(true))
	router := cont.Router

	// Create test namespace
	_ = testutils.CreateNamespace(t, db, "publish-provider-namespace", nil)

	tests := []struct {
		name           string
		setupAuth      func(*testing.T, *sqldb.Database) *http.Request
		expectedStatus int
	}{
		{
			name: "unauthenticated request returns 401",
			setupAuth: func(t *testing.T, db *sqldb.Database) *http.Request {
				req := testutils.BuildUnauthenticatedRequest(t, "POST", "/github/repositories/123/publish-provider")
				return testutils.AddChiContext(t, req, map[string]string{"provider_source": "github", "repo_id": "123"})
			},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name: "non-admin session password user returns 403 (requires GitHub auth or admin)",
			setupAuth: func(t *testing.T, db *sqldb.Database) *http.Request {
				req, _ := testutils.BuildAuthenticatedRequestWithSession(t, db, "POST", "/github/repositories/123/publish-provider", "regular-user", false)
				return testutils.AddChiContext(t, req, map[string]string{"provider_source": "github", "repo_id": "123"})
			},
			expectedStatus: http.StatusForbidden,
		},
		{
			name: "admin user can publish provider",
			setupAuth: func(t *testing.T, db *sqldb.Database) *http.Request {
				req, _ := testutils.BuildAdminRequest(t, db, "POST", "/github/repositories/123/publish-provider")
				return testutils.AddChiContext(t, req, map[string]string{"provider_source": "github", "repo_id": "123"})
			},
			expectedStatus: http.StatusNotFound, // 404 because repository doesn't exist (checked before form data)
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
