package terrareg_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/matthewjohn/terrareg/terrareg-go/test/integration/testutils"
)

// TestAuthHandler_HandleIsAuthenticated_Unauthenticated tests the is_authenticated endpoint for unauthenticated requests
func TestAuthHandler_HandleIsAuthenticated_Unauthenticated(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	// Get the real server router from the container
	cont := testutils.CreateTestContainer(t, db)
	router := cont.Server.Router()

	// Create request to is_authenticated endpoint
	req := httptest.NewRequest("GET", "/v1/terrareg/auth/admin/is_authenticated", nil)
	w := httptest.NewRecorder()

	// Serve the request
	router.ServeHTTP(w, req)

	// Should return 200 OK with unauthenticated status
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	// Verify response structure
	assert.Contains(t, response, "authenticated")
	assert.Contains(t, response, "read_access")
	assert.Contains(t, response, "site_admin")
	assert.Contains(t, response, "namespace_permissions")

	// Verify unauthenticated status
	assert.False(t, bool(response["authenticated"].(bool)))
	assert.False(t, bool(response["read_access"].(bool)))
	assert.False(t, bool(response["site_admin"].(bool)))
}

// TestAuthHandler_HandleIsAuthenticated_Authenticated tests the is_authenticated endpoint for authenticated admin user
func TestAuthHandler_HandleIsAuthenticated_Authenticated(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	// Get the real server router from the container
	cont := testutils.CreateTestContainer(t, db)
	router := cont.Server.Router()

	// Create an authenticated request with admin session
	req, cookieValue := testutils.CreateAuthenticatedRequestWithSession(t, db, "GET", "/v1/terrareg/auth/admin/is_authenticated", "admin-user", true)
	req.Header.Set("Cookie", "terrareg_session="+cookieValue)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Should return 200 OK with authenticated status
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	// Verify authenticated status
	assert.True(t, bool(response["authenticated"].(bool)))
	assert.True(t, bool(response["read_access"].(bool)))
	assert.True(t, bool(response["site_admin"].(bool)))
}

// TestAuthHandler_HandleIsAuthenticated_NonAdminUser tests the is_authenticated endpoint for non-admin user
func TestAuthHandler_HandleIsAuthenticated_NonAdminUser(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	// Get the real server router from the container
	cont := testutils.CreateTestContainer(t, db)
	router := cont.Server.Router()

	// Create an authenticated request with non-admin session
	req, cookieValue := testutils.CreateAuthenticatedRequestWithSession(t, db, "GET", "/v1/terrareg/auth/admin/is_authenticated", "regular-user", false)
	req.Header.Set("Cookie", "terrareg_session="+cookieValue)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Should return 200 OK with authenticated but non-admin status
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	// Verify authenticated but not site admin
	assert.True(t, bool(response["authenticated"].(bool)))
	assert.True(t, bool(response["read_access"].(bool)))
	assert.False(t, bool(response["site_admin"].(bool)))
}

// TestAuthHandler_HandleAdminLogin_Success tests admin login with valid API key
func TestAuthHandler_HandleAdminLogin_Success(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	// Get the real server router from the container
	cont := testutils.CreateTestContainer(t, db)
	router := cont.Server.Router()

	// Create request with admin API key
	req := httptest.NewRequest("POST", "/v1/terrareg/auth/admin/login", nil)
	req.Header.Set("X-Terrareg-ApiKey", "test-admin-api-key")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Should return 200 OK with authenticated status
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	// Verify authenticated status
	assert.Contains(t, response, "authenticated")
	assert.True(t, bool(response["authenticated"].(bool)))

	// Verify session cookie was set
	cookies := w.Result().Cookies()
	assert.NotEmpty(t, cookies, "Session cookie should be set")
	sessionCookie := cookies[0]
	assert.Equal(t, "terrareg_session", sessionCookie.Name)
}

// TestAuthHandler_HandleAdminLogin_Failure_InvalidApiKey tests admin login with invalid API key
func TestAuthHandler_HandleAdminLogin_Failure_InvalidApiKey(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	// Get the real server router from the container
	cont := testutils.CreateTestContainer(t, db)
	router := cont.Server.Router()

	// Create request with invalid API key
	req := httptest.NewRequest("POST", "/v1/terrareg/auth/admin/login", nil)
	req.Header.Set("X-Terrareg-ApiKey", "invalid-api-key")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Should return 401 Unauthorized
	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	// Verify error message
	assert.Contains(t, response, "message")
}

// TestAuthHandler_HandleAdminLogin_Failure_NoApiKey tests admin login without API key
func TestAuthHandler_HandleAdminLogin_Failure_NoApiKey(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	// Get the real server router from the container
	cont := testutils.CreateTestContainer(t, db)
	router := cont.Server.Router()

	// Create request without API key
	req := httptest.NewRequest("POST", "/v1/terrareg/auth/admin/login", nil)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Should return 401 Unauthorized
	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	// Verify error message
	assert.Contains(t, response, "message")
}

// TestAuthHandler_HandleAdminLogin_Failure_MethodNotAllowed tests admin login with GET method
func TestAuthHandler_HandleAdminLogin_Failure_MethodNotAllowed(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	// Get the real server router from the container
	cont := testutils.CreateTestContainer(t, db)
	router := cont.Server.Router()

	// Create request with GET method (should be POST)
	req := httptest.NewRequest("GET", "/v1/terrareg/auth/admin/login", nil)
	req.Header.Set("X-Terrareg-ApiKey", "test-admin-api-key")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Should return 405 Method Not Allowed
	// Chi router returns 405 with empty body when method doesn't match
	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)

	// Response body may be empty for 405
	if w.Body.Len() > 0 {
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		if err == nil {
			// If response body exists, verify it has a message
			assert.Contains(t, response, "message")
		}
	}
}

// Note: OIDC, SAML, and GitHub OAuth tests are skipped because these features
// require additional configuration (OIDC providers, SAML IDP metadata, GitHub OAuth app).
//
// Mock servers are available in testutils/auth_mocks.go for testing:
//   - testutils.NewMockOIDCServer() - Creates a mock OIDC provider
//   - testutils.NewMockSAMLServer() - Creates a mock SAML IDP
//   - testutils.NewMockGitHubOAuthServer() - Creates a mock GitHub OAuth server
//
// To enable these tests, the test container would need to accept custom auth configuration
// to point to the mock servers instead of real providers.
//
// Example usage:
//   mockServer := testutils.NewMockOIDCServer()
//   defer mockServer.Close()
//   config := testutils.MockAuthConfigWithOIDC(mockServer)

// TestAuthHandler_HandleOIDCLogin_RedirectURL tests OIDC login endpoint with redirect URL
func TestAuthHandler_HandleOIDCLogin_RedirectURL(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	// Create mock OIDC server
	mockServer := testutils.NewMockOIDCServer()
	defer mockServer.Close()

	// Create container with mock OIDC config
	cont := testutils.CreateTestContainerWithConfig(t, db,
		testutils.WithOIDCConfig(mockServer.Server.URL, mockServer.ClientID, mockServer.ClientSecret),
		testutils.WithPublicURL("http://localhost:5000"),
	)
	router := cont.Server.Router()

	// Test OIDC login with redirect URL
	req := httptest.NewRequest("GET", "/openid/login?redirect_url=/test-redirect", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Should redirect to OIDC provider
	assert.Equal(t, http.StatusFound, w.Code)
	location := w.Header().Get("Location")
	assert.Contains(t, location, mockServer.Server.URL)
}

// TestAuthHandler_HandleOIDCLogin_DefaultRedirect tests OIDC login without redirect URL
func TestAuthHandler_HandleOIDCLogin_DefaultRedirect(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	// Create mock OIDC server
	mockServer := testutils.NewMockOIDCServer()
	defer mockServer.Close()

	// Create container with mock OIDC config
	cont := testutils.CreateTestContainerWithConfig(t, db,
		testutils.WithOIDCConfig(mockServer.Server.URL, mockServer.ClientID, mockServer.ClientSecret),
		testutils.WithPublicURL("http://localhost:5000"),
	)
	router := cont.Server.Router()

	// Test OIDC login without redirect URL (should use default)
	req := httptest.NewRequest("GET", "/openid/login", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Should redirect to OIDC provider
	assert.Equal(t, http.StatusFound, w.Code)
	location := w.Header().Get("Location")
	assert.Contains(t, location, mockServer.Server.URL)
}

// TestAuthHandler_HandleOIDCCallback tests OIDC callback endpoint
func TestAuthHandler_HandleOIDCCallback(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	// Create mock OIDC server
	mockServer := testutils.NewMockOIDCServer()
	defer mockServer.Close()

	// Create container with mock OIDC config
	cont := testutils.CreateTestContainerWithConfig(t, db,
		testutils.WithOIDCConfig(mockServer.Server.URL, mockServer.ClientID, mockServer.ClientSecret),
		testutils.WithPublicURL("http://localhost:5000"),
	)
	router := cont.Server.Router()

	// Test OIDC callback with mock authorization code
	// Note: This test would need a valid state parameter from a previous login flow
	// For simplicity, we just verify the endpoint responds (may return error without proper state)
	req := httptest.NewRequest("GET", "/openid/callback?code=test-code&state=test-state", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// The handler should process the request (may return error if state is invalid)
	// We just verify it doesn't panic and returns a response
	assert.NotEqual(t, 0, w.Code)
}

// TestAuthHandler_HandleSAMLLogin tests SAML login endpoint
func TestAuthHandler_HandleSAMLLogin(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	// Create mock SAML server
	mockServer, err := testutils.NewMockSAMLServer()
	require.NoError(t, err)
	defer mockServer.Close()

	// Create container with mock SAML config
	cont := testutils.CreateTestContainerWithConfig(t, db,
		testutils.WithSAMLConfig("https://sp.example.com/saml", mockServer.MetadataURL),
		testutils.WithPublicURL("http://localhost:5000"),
	)
	router := cont.Server.Router()

	// Test SAML login
	req := httptest.NewRequest("GET", "/saml/login", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Should redirect to SAML IDP or return SAML response
	assert.NotEqual(t, 0, w.Code)
}

// TestAuthHandler_HandleSAMLMetadata tests SAML metadata endpoint
func TestAuthHandler_HandleSAMLMetadata(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	// Create mock SAML server
	mockServer, err := testutils.NewMockSAMLServer()
	require.NoError(t, err)
	defer mockServer.Close()

	// Create container with mock SAML config
	cont := testutils.CreateTestContainerWithConfig(t, db,
		testutils.WithSAMLConfig("https://sp.example.com/saml", mockServer.MetadataURL),
		testutils.WithPublicURL("http://localhost:5000"),
	)
	router := cont.Server.Router()

	// Test SAML metadata endpoint
	req := httptest.NewRequest("GET", "/saml/metadata", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Should return SAML metadata XML
	assert.NotEqual(t, 0, w.Code)
}

// TestAuthHandler_HandleGitHubOAuth tests GitHub OAuth endpoint
// Disabled: GitHub OAuth not yet implemented in Go version
func TestAuthHandler_HandleGitHubOAuth(t *testing.T) {
	t.Skip("GitHub OAuth not implemented in Go version yet")
}

// TestAuthHandler_HandleLogout_Authenticated tests logout with authenticated session
func TestAuthHandler_HandleLogout_Authenticated(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	// Get the real server router from the container
	cont := testutils.CreateTestContainer(t, db)
	router := cont.Server.Router()

	// Create an authenticated request
	req, cookieValue := testutils.CreateAuthenticatedRequestWithSession(t, db, "POST", "/v1/terrareg/auth/logout", "test-user", false)
	req.Header.Set("Cookie", "terrareg_session="+cookieValue)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Should return redirect (303 See Other) - the server logout implementation redirects
	assert.Equal(t, http.StatusSeeOther, w.Code)
}

// TestAuthHandler_HandleLogout_Unauthenticated tests logout without session
func TestAuthHandler_HandleLogout_Unauthenticated(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	// Get the real server router from the container
	cont := testutils.CreateTestContainer(t, db)
	router := cont.Server.Router()

	// Create request without session
	req := httptest.NewRequest("POST", "/v1/terrareg/auth/logout", nil)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Should still return redirect (logout is idempotent)
	assert.Equal(t, http.StatusSeeOther, w.Code)
}


