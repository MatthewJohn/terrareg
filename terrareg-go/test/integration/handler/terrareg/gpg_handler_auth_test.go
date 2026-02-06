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

// TestGPGKeyCreate_Authentication tests GPG key creation with RequireAdmin middleware (POST /v2/gpg-keys)
func TestGPGKeyCreate_Authentication(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	// Enable RBAC for this test so permission checking works properly
	cont := testutils.CreateTestServerWithConfig(t, db, testutils.WithEnableAccessControls(true))
	router := cont.Router

	// Create test namespace
	_ = testutils.CreateNamespace(t, db, "gpg-namespace", nil)

	tests := []struct {
		name           string
		setupAuth      func(*testing.T, *sqldb.Database) *http.Request
		expectedStatus int
	}{
		{
			name: "unauthenticated request returns 403",
			setupAuth: func(t *testing.T, db *sqldb.Database) *http.Request {
				return testutils.BuildUnauthenticatedRequest(t, "POST", "/v2/gpg-keys")
			},
			expectedStatus: http.StatusForbidden,
		},
		{
			name: "regular authenticated user returns 403",
			setupAuth: func(t *testing.T, db *sqldb.Database) *http.Request {
				req, _ := testutils.BuildAuthenticatedRequestWithSession(t, db, "POST", "/v2/gpg-keys", "regular-user", false)
				return req
			},
			expectedStatus: http.StatusForbidden,
		},
		{
			name: "admin user can create GPG keys",
			setupAuth: func(t *testing.T, db *sqldb.Database) *http.Request {
				req, _ := testutils.BuildAdminRequest(t, db, "POST", "/v2/gpg-keys")
				// Add request body for GPG key creation
				reqBody := `{
					"data": {
						"type": "gpg-keys",
						"attributes": {
							"namespace": "gpg-namespace",
							"ascii-armor": "-----BEGIN PGP PUBLIC KEY BLOCK-----\nMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEA...\n-----END PGP PUBLIC KEY BLOCK-----"
						}
					},
					"csrf_token": "test-token"
				}`
				req.Body = io.NopCloser(strings.NewReader(reqBody))
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

// TestGPGKeyDelete_Authentication tests GPG key deletion with RequireAdmin middleware (DELETE /v2/gpg-keys/{namespace}/{key_id})
func TestGPGKeyDelete_Authentication(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	// Enable RBAC for this test so permission checking works properly
	cont := testutils.CreateTestServerWithConfig(t, db, testutils.WithEnableAccessControls(true))
	router := cont.Router

	// Create test namespace
	_ = testutils.CreateNamespace(t, db, "gpg-delete-namespace", nil)

	tests := []struct {
		name           string
		setupAuth      func(*testing.T, *sqldb.Database) *http.Request
		expectedStatus int
	}{
		{
			name: "unauthenticated request returns 403",
			setupAuth: func(t *testing.T, db *sqldb.Database) *http.Request {
				req := testutils.BuildUnauthenticatedRequest(t, "DELETE", "/v2/gpg-keys/gpg-delete-namespace/test-key-id")
				return testutils.AddChiContext(t, req, map[string]string{"namespace": "gpg-delete-namespace", "key_id": "test-key-id"})
			},
			expectedStatus: http.StatusForbidden,
		},
		{
			name: "regular authenticated user returns 403",
			setupAuth: func(t *testing.T, db *sqldb.Database) *http.Request {
				req, _ := testutils.BuildAuthenticatedRequestWithSession(t, db, "DELETE", "/v2/gpg-keys/gpg-delete-namespace/test-key-id", "regular-user", false)
				return testutils.AddChiContext(t, req, map[string]string{"namespace": "gpg-delete-namespace", "key_id": "test-key-id"})
			},
			expectedStatus: http.StatusForbidden,
		},
		{
			name: "admin user can delete GPG keys",
			setupAuth: func(t *testing.T, db *sqldb.Database) *http.Request {
				// Create a GPG key to delete - first get namespace ID
				var namespaceDB sqldb.NamespaceDB
				err := db.DB.Where("namespace = ?", "gpg-delete-namespace").First(&namespaceDB).Error
				if err == nil {
					testutils.CreateGPGKeyWithNamespace(t, db, "test-source", namespaceDB.ID, "test-key-id")
				}
				req, _ := testutils.BuildAdminRequest(t, db, "DELETE", "/v2/gpg-keys/gpg-delete-namespace/test-key-id")
				return testutils.AddChiContext(t, req, map[string]string{"namespace": "gpg-delete-namespace", "key_id": "test-key-id"})
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

// TestGPGKeyList_AllAuthMethods tests GET GPG key list endpoint with OptionalAuth
// All authentication states should return 200
func TestGPGKeyList_AllAuthMethods(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	// Enable RBAC for this test so permission checking works properly
	cont := testutils.CreateTestServerWithConfig(t, db, testutils.WithEnableAccessControls(true))
	router := cont.Router

	// Create test namespace
	_ = testutils.CreateNamespace(t, db, "gpg-list-namespace", nil)

	tests := []struct {
		name           string
		setupAuth      func(*testing.T, *sqldb.Database) *http.Request
		expectedStatus int
	}{
		{
			name: "unauthenticated request returns 200",
			setupAuth: func(t *testing.T, db *sqldb.Database) *http.Request {
				return testutils.BuildUnauthenticatedRequest(t, "GET", "/v2/gpg-keys?filter[namespace]=gpg-list-namespace")
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "authenticated regular user returns 200",
			setupAuth: func(t *testing.T, db *sqldb.Database) *http.Request {
				req, _ := testutils.BuildAuthenticatedRequestWithSession(t, db, "GET", "/v2/gpg-keys?filter[namespace]=gpg-list-namespace", "regular-user", false)
				return req
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "authenticated admin user returns 200",
			setupAuth: func(t *testing.T, db *sqldb.Database) *http.Request {
				req, _ := testutils.BuildAdminRequest(t, db, "GET", "/v2/gpg-keys?filter[namespace]=gpg-list-namespace")
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
