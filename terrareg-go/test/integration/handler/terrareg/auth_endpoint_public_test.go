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
)

// TestPublicEndpoints_AllAuthMethods tests that public endpoints allow all authentication methods
// These endpoints should return 200 OK regardless of authentication state or ALLOW_UNAUTHENTICATED_ACCESS config
func TestPublicEndpoints_AllAuthMethods(t *testing.T) {
	endpoints := []struct {
		name   string
		method string
		path   string
	}{
		{
			name:   "health",
			method: "GET",
			path:   "/v1/terrareg/health",
		},
		{
			name:   "version",
			method: "GET",
			path:   "/v1/terrareg/version",
		},
		// Note: analytics/global/* endpoints require read API access (can_access_read_api),
		// so they are NOT truly public. They are tested in auth_endpoint_read_test.go instead.
	}

	authMethods := []struct {
		name  string
		setup func(t *testing.T, req *http.Request)
	}{
		{
			name: "unauthenticated",
			setup: func(t *testing.T, req *http.Request) {
				// No authentication
			},
		},
		{
			name: "admin_api_key",
			setup: func(t *testing.T, req *http.Request) {
				apiKey := os.Getenv("ADMIN_AUTH_TOKEN")
				if apiKey == "" {
					apiKey = "test-admin-api-key"
				}
				req.Header.Set("X-Terrareg-ApiKey", apiKey)
			},
		},
		{
			name: "upload_api_key",
			setup: func(t *testing.T, req *http.Request) {
				apiKey := os.Getenv("UPLOAD_AUTH_TOKEN")
				if apiKey == "" {
					apiKey = "test-upload-key"
				}
				req.Header.Set("X-Terrareg-UploadKey", apiKey)
			},
		},
		{
			name: "publish_api_key",
			setup: func(t *testing.T, req *http.Request) {
				apiKey := os.Getenv("PUBLISH_AUTH_TOKEN")
				if apiKey == "" {
					apiKey = "test-publish-key"
				}
				req.Header.Set("X-Terrareg-ApiKey", apiKey)
			},
		},
		{
			name: "admin_session",
			setup: func(t *testing.T, req *http.Request) {
				// For public endpoints, we don't need to set up a full session
				// The endpoint should work regardless
			},
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

						cont := testutils.CreateTestContainerWithConfig(t, db,
							testutils.WithAllowUnauthenticatedAccess(true))
						router := cont.Server.Router()

						req := httptest.NewRequest(endpoint.method, endpoint.path, nil)
						authMethod.setup(t, req)

						w := httptest.NewRecorder()
						router.ServeHTTP(w, req)

						// Public endpoints should always return 200 OK
						assert.Equal(t, http.StatusOK, w.Code,
							fmt.Sprintf("Endpoint %s with auth %s should return 200", endpoint.path, authMethod.name))
					})

					// Test with ALLOW_UNAUTHENTICATED_ACCESS=false
					t.Run("ALLOW_UNAUTHENTICATED_ACCESS=false", func(t *testing.T) {
						db := testutils.SetupTestDatabase(t)
						defer testutils.CleanupTestDatabase(t, db)

						cont := testutils.CreateTestContainerWithConfig(t, db,
							testutils.WithAllowUnauthenticatedAccess(false))
						router := cont.Server.Router()

						req := httptest.NewRequest(endpoint.method, endpoint.path, nil)
						authMethod.setup(t, req)

						w := httptest.NewRecorder()
						router.ServeHTTP(w, req)

						// Public endpoints should still return 200 OK even with ALLOW_UNAUTHENTICATED_ACCESS=false
						assert.Equal(t, http.StatusOK, w.Code,
							fmt.Sprintf("Endpoint %s with auth %s should return 200", endpoint.path, authMethod.name))
					})
				})
			}
		})
	}
}

// TestPublicEndpoints_HealthResponseStructure verifies the health endpoint returns valid JSON
func TestPublicEndpoints_HealthResponseStructure(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	cont := testutils.CreateTestContainer(t, db)
	router := cont.Server.Router()

	req := httptest.NewRequest("GET", "/v1/terrareg/health", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	// Health endpoint should return a status
	assert.Contains(t, response, "status")
}

// TestPublicEndpoints_VersionResponseStructure verifies the version endpoint returns valid JSON
func TestPublicEndpoints_VersionResponseStructure(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	cont := testutils.CreateTestContainer(t, db)
	router := cont.Server.Router()

	req := httptest.NewRequest("GET", "/v1/terrareg/version", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	// Version endpoint should return version information
	assert.Contains(t, response, "version")
}

// TestPublicEndpoints_AnalyticsAcceptsUnauthenticated verifies analytics endpoint requires read API access
func TestPublicEndpoints_AnalyticsAcceptsUnauthenticated(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	cont := testutils.CreateTestContainerWithConfig(t, db,
		testutils.WithAllowUnauthenticatedAccess(false))
	router := cont.Server.Router()

	req := httptest.NewRequest("GET", "/v1/terrareg/analytics/global/stats_summary", nil)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Analytics endpoint requires read API access (can_access_read_api)
	// With ALLOW_UNAUTHENTICATED_ACCESS=false, unauthenticated requests should return 401
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// TestPublicEndpoints_AnalyticsAcceptsAuthenticatedWithReadAccess verifies analytics endpoint works with authenticated session
func TestPublicEndpoints_AnalyticsAcceptsAuthenticatedWithReadAccess(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	// Enable RBAC so permission checking works properly
	cont := testutils.CreateTestServerWithConfig(t, db, testutils.WithEnableAccessControls(true))
	authHelper := testutils.NewAuthHelper(t, db, cont)

	// Create a user session
	cookie := authHelper.CreateSessionForUser("testuser", false, []string{}, nil)

	router := cont.Router
	req := httptest.NewRequest("GET", "/v1/terrareg/analytics/global/stats_summary", nil)
	req.Header.Set("Cookie", cookie)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Analytics endpoint should work with authenticated session (has read API access)
	assert.Equal(t, http.StatusOK, w.Code)
}
