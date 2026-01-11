package testutils

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/config/model"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/config"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/container"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/persistence/sqldb"
)

// AssertHTTPError asserts HTTP error status and error response
func AssertHTTPError(t *testing.T, w *httptest.ResponseRecorder, statusCode int, containsMessage string) {
	assert.Equal(t, statusCode, w.Code, "Expected status code %d", statusCode)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err, "Response should be valid JSON")

	if containsMessage != "" {
		if errorMsg, ok := response["error"].(string); ok {
			assert.Contains(t, errorMsg, containsMessage, "Error message should contain expected text")
		} else if message, ok := response["message"].(string); ok {
			assert.Contains(t, message, containsMessage, "Message should contain expected text")
		}
	}
}

// AssertJSONSuccess asserts successful JSON response
func AssertJSONSuccess(t *testing.T, w *httptest.ResponseRecorder) {
	assert.Equal(t, http.StatusOK, w.Code, "Expected status code 200")

	// Verify response is valid JSON
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err, "Response should be valid JSON")
}

// CreateTestRequest creates a test request with headers and optional auth
func CreateTestRequest(t *testing.T, method, url string, body io.Reader, headers map[string]string) *http.Request {
	req := httptest.NewRequest(method, url, body)
	for key, value := range headers {
		req.Header.Set(key, value)
	}
	return req
}

// SetupIntegrationTest creates database, config, and server for integration tests
// Returns cleanup function
func SetupIntegrationTest(t *testing.T) (*sqldb.Database, *container.Container, func()) {
	db := SetupTestDatabase(t)

	domainConfig := CreateTestDomainConfig(t)
	infraConfig := CreateTestInfraConfig(t)

	container, err := container.NewContainer(domainConfig, infraConfig, nil, GetTestLogger(), db)
	require.NoError(t, err, "Failed to create container")

	cleanup := func() {
		CleanupTestDatabase(t, db)
	}

	return db, container, cleanup
}

// MakeAuthenticatedRequest makes an HTTP request with authentication
func MakeAuthenticatedRequest(t *testing.T, handler http.Handler, method, url string, body io.Reader, authConfig MockAuthConfig) *httptest.ResponseRecorder {
	req := CreateAuthenticatedRequest(httptest.NewRequest(method, url, body), authConfig)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	return w
}

// PerformRequest performs an HTTP request and returns the recorder
func PerformRequest(t *testing.T, handler http.Handler, method, url string, body io.Reader) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, url, body)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	return w
}

// AssertJSONResponse asserts response is valid JSON and contains expected data
func AssertJSONResponse(t *testing.T, w *httptest.ResponseRecorder, expectedStatusCode int, expectedBody map[string]interface{}) {
	assert.Equal(t, expectedStatusCode, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	for key, expectedValue := range expectedBody {
		actualValue, ok := response[key]
		assert.True(t, ok, "Response should contain key: %s", key)
		assert.Equal(t, expectedValue, actualValue, "Value for key %s should match", key)
	}
}

// GetJSONBody parses JSON response body into a map
func GetJSONBody(t *testing.T, w *httptest.ResponseRecorder) map[string]interface{} {
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err, "Response should be valid JSON")
	return response
}

// CreateJSONBody creates a JSON reader for request body
func CreateJSONBody(t *testing.T, data interface{}) io.Reader {
	body, err := json.Marshal(data)
	require.NoError(t, err, "Failed to marshal JSON body")
	return bytes.NewReader(body)
}

// CreateTestConfig creates test configuration for integration tests
func CreateTestConfig() (*model.DomainConfig, *config.InfrastructureConfig) {
	domainConfig := CreateTestDomainConfig(nil)
	infraConfig := CreateTestInfraConfig(nil)
	return domainConfig, infraConfig
}

// CreateTestServerWithAuth creates a test server with authentication middleware
func CreateTestServerWithAuth(t *testing.T, handler http.Handler, authConfig MockAuthConfig) *httptest.Server {
	authHandler := SetupMockAuth(authConfig)(handler)
	return httptest.NewServer(authHandler)
}

// AssertModuleProviderInResponse asserts that module provider data exists in response
func AssertModuleProviderInResponse(t *testing.T, response map[string]interface{}, namespace, module, provider string) {
	data, ok := response["data"].(map[string]interface{})
	require.True(t, ok, "Response should have 'data' field")

	// Check for module provider fields
	if ns, ok := data["namespace"].(string); ok {
		assert.Equal(t, namespace, ns, "Namespace should match")
	}
	if mod, ok := data["module"].(string); ok {
		assert.Equal(t, module, mod, "Module should match")
	}
	if prov, ok := data["provider"].(string); ok {
		assert.Equal(t, provider, prov, "Provider should match")
	}
}

// AssertErrorContains asserts response contains an error with the expected message
func AssertErrorContains(t *testing.T, w *httptest.ResponseRecorder, expectedMessage string) {
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err, "Response should be valid JSON")

	errorMsg, ok := response["error"].(string)
	if ok {
		assert.Contains(t, errorMsg, expectedMessage, "Error should contain expected message")
		return
	}

	message, ok := response["message"].(string)
	if ok {
		assert.Contains(t, message, expectedMessage, "Message should contain expected text")
		return
	}

	t.Fatalf("Response should contain 'error' or 'message' field")
}

// WithAdminHeaders adds admin API key headers to request
func WithAdminHeaders(req *http.Request, token string) {
	req.Header.Set("X-Terrareg-ApiKey", token)
}

// WithUploadKeyHeaders adds upload API key headers to request
func WithUploadKeyHeaders(req *http.Request, token string) {
	req.Header.Set("X-Terrareg-Upload-Key", token)
}

// WithPublishKeyHeaders adds publish API key headers to request
func WithPublishKeyHeaders(req *http.Request, token string) {
	req.Header.Set("X-Terrareg-Publish-Key", token)
}

// AssertModuleVersionsInResponse asserts module versions are present in response
func AssertModuleVersionsInResponse(t *testing.T, response map[string]interface{}, expectedVersions []string) {
	data, ok := response["data"].([]interface{})
	require.True(t, ok, "Response should have 'data' array")

	actualVersions := make([]string, 0, len(data))
	for _, item := range data {
		version, ok := item.(map[string]interface{})["version"].(string)
		if ok {
			actualVersions = append(actualVersions, version)
		}
	}

	assert.Equal(t, expectedVersions, actualVersions, "Versions should match expected list")
}

// AssertPaginatedResponse asserts paginated response structure
func AssertPaginatedResponse(t *testing.T, w *httptest.ResponseRecorder, expectedPage, expectedLimit, expectedTotal int) {
	response := GetJSONBody(t, w)

	meta, ok := response["meta"].(map[string]interface{})
	require.True(t, ok, "Response should have 'meta' field")

	assert.Equal(t, float64(expectedPage), meta["page"], "Page should match")
	assert.Equal(t, float64(expectedLimit), meta["limit"], "Limit should match")
	assert.Equal(t, float64(expectedTotal), meta["total"], "Total should match")
}

// Now returns a fixed time for consistent test results
func Now() time.Time {
	return time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
}
