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
			path:   "/v1/health",
		},
		{
			name:   "version",
			method: "GET",
			path:   "/v1/version",
		},
		{
			name:   "analytics",
			method: "POST",
			path:   "/v1/terraform-modules/v1/analytics",
		},
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
				req.Header.Set("X-Terrareg-PublishKey", apiKey)
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

	req := httptest.NewRequest("GET", "/v1/health", nil)
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

	req := httptest.NewRequest("GET", "/v1/version", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	// Version endpoint should return version information
	assert.Contains(t, response, "version")
}

// TestPublicEndpoints_AnalyticsAcceptsUnauthenticated verifies analytics endpoint accepts unauthenticated requests
func TestPublicEndpoints_AnalyticsAcceptsUnauthenticated(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	cont := testutils.CreateTestContainerWithConfig(t, db,
		testutils.WithAllowUnauthenticatedAccess(false))
	router := cont.Server.Router()

	// Create analytics payload
	payload := map[string]interface{}{
		"version": "1.0.0",
		"platform": "linux_amd64",
	}
	body, _ := json.Marshal(payload)
	reqBody := bytes.NewReader(body)

	req := httptest.NewRequest("POST", "/v1/terraform-modules/v1/analytics", reqBody)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Analytics endpoint should accept unauthenticated requests even with ALLOW_UNAUTHENTICATED_ACCESS=false
	// This is required for Terraform CLI telemetry
	assert.Equal(t, http.StatusOK, w.Code)
}
