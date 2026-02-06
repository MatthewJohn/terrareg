package terrareg_test

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/persistence/sqldb"
	"github.com/matthewjohn/terrareg/terrareg-go/test/integration/testutils"
)

// TestAuditHistory_Authentication tests audit history endpoint with RequireAdmin middleware
func TestAuditHistory_Authentication(t *testing.T) {
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
				return testutils.BuildUnauthenticatedRequest(t, "GET", "/v1/terrareg/audit-history")
			},
			expectedStatus: http.StatusForbidden,
		},
		{
			name: "regular authenticated user returns 403",
			setupAuth: func(t *testing.T, db *sqldb.Database) *http.Request {
				req, _ := testutils.BuildAuthenticatedRequestWithSession(t, db, "GET", "/v1/terrareg/audit-history", "regular-user", false)
				return req
			},
			expectedStatus: http.StatusForbidden,
		},
		{
			name: "admin user can access audit history",
			setupAuth: func(t *testing.T, db *sqldb.Database) *http.Request {
				req, _ := testutils.BuildAdminRequest(t, db, "GET", "/v1/terrareg/audit-history")
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
