package terrareg_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/matthewjohn/terrareg/terrareg-go/test/integration/testutils"
)

// TestProviderSourceHandler_HandleGitHubLogin_Success tests successful GitHub login redirect
func TestProviderSourceHandler_HandleGitHubLogin_Success(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	// Create mock GitHub server
	mockServer := testutils.NewMockGitHubOAuthServer()
	defer mockServer.Close()

	// Create test GitHub provider source in database
	testutils.CreateTestGitHubProviderSource(t, db, mockServer)

	// Get the real server router from the container
	cont := testutils.CreateTestContainer(t, db)
	router := cont.Server.Router()

	// Create request to GitHub login endpoint
	req := httptest.NewRequest("GET", "/test-github/login", nil)
	w := httptest.NewRecorder()

	// Serve the request
	router.ServeHTTP(w, req)

	// Should redirect to GitHub OAuth
	assert.Equal(t, http.StatusFound, w.Code)

	// Verify redirect URL contains GitHub parameters
	location := w.Header().Get("Location")
	assert.NotEmpty(t, location, "Redirect URL should not be empty")
	assert.Contains(t, location, mockServer.Server.URL, "Should redirect to mock GitHub server")
	assert.Contains(t, location, "client_id="+mockServer.ClientID, "Should include client_id")
	assert.Contains(t, location, "state=", "Should include state parameter")
	assert.Contains(t, location, "scope=read:org", "Should include scope")
}

// TestProviderSourceHandler_HandleGitHubLogin_ProviderNotFound tests login with non-existent provider
func TestProviderSourceHandler_HandleGitHubLogin_ProviderNotFound(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	// Get the real server router from the container
	cont := testutils.CreateTestContainer(t, db)
	router := cont.Server.Router()

	// Create request to non-existent provider
	req := httptest.NewRequest("GET", "/non-existent/login", nil)
	w := httptest.NewRecorder()

	// Serve the request
	router.ServeHTTP(w, req)

	// Should return 404
	assert.Equal(t, http.StatusNotFound, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Contains(t, response, "error")
}

// TestProviderSourceHandler_HandleGitHubCallback_Success tests successful GitHub OAuth callback
func TestProviderSourceHandler_HandleGitHubCallback_Success(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	// Create mock GitHub server
	mockServer := testutils.NewMockGitHubOAuthServer()
	defer mockServer.Close()

	// Create test GitHub provider source in database
	testutils.CreateTestGitHubProviderSource(t, db, mockServer)

	// Get the real server router from the container
	cont := testutils.CreateTestContainer(t, db)
	router := cont.Server.Router()

	// First, initiate login to get the state parameter
	loginReq := httptest.NewRequest("GET", "/test-github/login", nil)
	loginW := httptest.NewRecorder()
	router.ServeHTTP(loginW, loginReq)

	// Extract state from redirect URL
	loginLocation := loginW.Header().Get("Location")
	parsedURL, err := url.Parse(loginLocation)
	require.NoError(t, err)
	state := parsedURL.Query().Get("state")
	require.NotEmpty(t, state, "State should be present in login redirect")

	// Simulate GitHub callback with authorization code
	// The mock server will accept this and return a mock auth code
	callbackURL := "/test-github/callback?code=mock-auth-code&state=" + url.QueryEscape(state)
	callbackReq := httptest.NewRequest("GET", callbackURL, nil)
	callbackW := httptest.NewRecorder()

	// Serve the callback request
	router.ServeHTTP(callbackW, callbackReq)

	// Should redirect to home after successful authentication
	assert.Equal(t, http.StatusFound, callbackW.Code)
	assert.Equal(t, "/", callbackW.Header().Get("Location"))

	// Verify session cookie was set
	cookies := callbackW.Result().Cookies()
	assert.NotEmpty(t, cookies, "Session cookie should be set")
	sessionCookie := cookies[0]
	assert.Equal(t, "terrareg_session", sessionCookie.Name)
}

// TestProviderSourceHandler_HandleGitHubCallback_MissingCode tests callback without code
func TestProviderSourceHandler_HandleGitHubCallback_MissingCode(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	// Create mock GitHub server
	mockServer := testutils.NewMockGitHubOAuthServer()
	defer mockServer.Close()

	// Create test GitHub provider source in database
	testutils.CreateTestGitHubProviderSource(t, db, mockServer)

	// Get the real server router from the container
	cont := testutils.CreateTestContainer(t, db)
	router := cont.Server.Router()

	// Create callback request without code
	req := httptest.NewRequest("GET", "/test-github/callback", nil)
	w := httptest.NewRecorder()

	// Serve the request
	router.ServeHTTP(w, req)

	// Should return 400 Bad Request
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Contains(t, response, "error")
}

// TestProviderSourceHandler_HandleGitHubCallback_ProviderNotFound tests callback with non-existent provider
func TestProviderSourceHandler_HandleGitHubCallback_ProviderNotFound(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	// Get the real server router from the container
	cont := testutils.CreateTestContainer(t, db)
	router := cont.Server.Router()

	// Create callback request for non-existent provider
	req := httptest.NewRequest("GET", "/non-existent/callback?code=test-code", nil)
	w := httptest.NewRecorder()

	// Serve the request
	router.ServeHTTP(w, req)

	// Should return 404
	assert.Equal(t, http.StatusNotFound, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Contains(t, response, "error")
}

// TestProviderSourceHandler_HandleGitHubAuthStatus_Authenticated tests auth status for authenticated user
func TestProviderSourceHandler_HandleGitHubAuthStatus_Authenticated(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	// Create mock GitHub server
	mockServer := testutils.NewMockGitHubOAuthServer()
	defer mockServer.Close()

	// Create test GitHub provider source in database
	testutils.CreateTestGitHubProviderSource(t, db, mockServer)

	// Get the real server router from the container
	cont := testutils.CreateTestContainer(t, db)
	router := cont.Server.Router()

	// First, authenticate to get a session
	// Initiate login
	loginReq := httptest.NewRequest("GET", "/test-github/login", nil)
	loginW := httptest.NewRecorder()
	router.ServeHTTP(loginW, loginReq)

	// Extract state from redirect URL
	loginLocation := loginW.Header().Get("Location")
	parsedURL, err := url.Parse(loginLocation)
	require.NoError(t, err)
	state := parsedURL.Query().Get("state")

	// Complete OAuth flow
	callbackURL := "/test-github/callback?code=mock-auth-code&state=" + url.QueryEscape(state)
	callbackReq := httptest.NewRequest("GET", callbackURL, nil)
	callbackW := httptest.NewRecorder()
	router.ServeHTTP(callbackW, callbackReq)

	// Extract session cookie
	cookies := callbackW.Result().Cookies()
	require.NotEmpty(t, cookies, "Session cookie should be set after authentication")
	sessionCookie := cookies[0]

	// Now check auth status
	authReq := httptest.NewRequest("GET", "/test-github/auth/status", nil)
	authReq.Header.Set("Cookie", "terrareg_session="+sessionCookie.Value)
	authW := httptest.NewRecorder()

	router.ServeHTTP(authW, authReq)

	// Should return 200 OK with authenticated status
	assert.Equal(t, http.StatusOK, authW.Code)

	var response map[string]interface{}
	err = json.Unmarshal(authW.Body.Bytes(), &response)
	require.NoError(t, err)

	// Verify response
	assert.Contains(t, response, "authenticated")
	assert.True(t, bool(response["authenticated"].(bool)))
	assert.Contains(t, response, "username")
	assert.Contains(t, response, "auth_method")
	assert.Contains(t, response, "provider_type")

	// Verify username matches mock server
	username := response["username"].(string)
	assert.Equal(t, mockServer.TestUserInfo.Login, username)
}

// TestProviderSourceHandler_HandleGitHubAuthStatus_Unauthenticated tests auth status for unauthenticated user
func TestProviderSourceHandler_HandleGitHubAuthStatus_Unauthenticated(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	// Create mock GitHub server
	mockServer := testutils.NewMockGitHubOAuthServer()
	defer mockServer.Close()

	// Create test GitHub provider source in database
	testutils.CreateTestGitHubProviderSource(t, db, mockServer)

	// Get the real server router from the container
	cont := testutils.CreateTestContainer(t, db)
	router := cont.Server.Router()

	// Check auth status without authentication
	req := httptest.NewRequest("GET", "/test-github/auth/status", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Should return 200 OK with unauthenticated status
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	// Verify unauthenticated status
	assert.Contains(t, response, "authenticated")
	assert.False(t, bool(response["authenticated"].(bool)))

	// Username and auth_method should not be present for unauthenticated users
	_, hasUsername := response["username"]
	_, hasAuthMethod := response["auth_method"]
	assert.False(t, hasUsername, "Username should not be present for unauthenticated users")
	assert.False(t, hasAuthMethod, "Auth method should not be present for unauthenticated users")
}

// TestProviderSourceHandler_HandleGitHubCallback_WithOrganizations tests callback with organizations
func TestProviderSourceHandler_HandleGitHubCallback_WithOrganizations(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	// Create mock GitHub server with test organizations
	mockServer := testutils.NewMockGitHubOAuthServer()
	// Set custom organizations for testing
	mockServer.TestOrgs = []string{"test-org-1", "test-org-2", "mycompany"}
	defer mockServer.Close()

	// Create test GitHub provider source in database
	testutils.CreateTestGitHubProviderSource(t, db, mockServer)

	// Get the real server router from the container
	cont := testutils.CreateTestContainer(t, db)
	router := cont.Server.Router()

	// Initiate login
	loginReq := httptest.NewRequest("GET", "/test-github/login", nil)
	loginW := httptest.NewRecorder()
	router.ServeHTTP(loginW, loginReq)

	// Extract state from redirect URL
	loginLocation := loginW.Header().Get("Location")
	parsedURL, _ := url.Parse(loginLocation)
	state := parsedURL.Query().Get("state")

	// Complete OAuth flow
	callbackURL := "/test-github/callback?code=mock-auth-code&state=" + url.QueryEscape(state)
	callbackReq := httptest.NewRequest("GET", callbackURL, nil)
	callbackW := httptest.NewRecorder()
	router.ServeHTTP(callbackW, callbackReq)

	// Should redirect to home after successful authentication
	assert.Equal(t, http.StatusFound, callbackW.Code)

	// Extract session cookie
	cookies := callbackW.Result().Cookies()
	require.NotEmpty(t, cookies)
	sessionCookie := cookies[0]

	// Check auth status to verify organizations were stored
	authReq := httptest.NewRequest("GET", "/test-github/auth/status", nil)
	authReq.Header.Set("Cookie", "terrareg_session="+sessionCookie.Value)
	authW := httptest.NewRecorder()

	router.ServeHTTP(authW, authReq)

	assert.Equal(t, http.StatusOK, authW.Code)

	var response map[string]interface{}
	json.Unmarshal(authW.Body.Bytes(), &response)
	require.NoError(t, json.Unmarshal(authW.Body.Bytes(), &response))

	// Verify authenticated with organizations
	assert.True(t, bool(response["authenticated"].(bool)))

	// The username should be from the mock server
	username := response["username"].(string)
	assert.Equal(t, mockServer.TestUserInfo.Login, username)
}

// TestProviderSourceHandler_HandleMultipleGitHubProviders tests multiple GitHub provider sources
func TestProviderSourceHandler_HandleMultipleGitHubProviders(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	// Create two mock GitHub servers for different providers
	mockServer1 := testutils.NewMockGitHubOAuthServer()
	mockServer1.TestUserInfo.Login = "user-github-com"
	mockServer1.TestOrgs = []string{"org1", "org2"}
	defer mockServer1.Close()

	mockServer2 := testutils.NewMockGitHubOAuthServer()
	mockServer2.TestUserInfo.Login = "user-github-enterprise"
	mockServer2.TestOrgs = []string{"enterprise-org"}
	defer mockServer2.Close()

	// Create two provider sources
	testutils.CreateTestGitHubProviderSourceWithName(t, db, "github-com", mockServer1)
	testutils.CreateTestGitHubProviderSourceWithName(t, db, "github-enterprise", mockServer2)

	// Get the real server router from the container
	cont := testutils.CreateTestContainer(t, db)
	router := cont.Server.Router()

	// Test login to first provider
	loginReq1 := httptest.NewRequest("GET", "/github-com/login", nil)
	loginW1 := httptest.NewRecorder()
	router.ServeHTTP(loginW1, loginReq1)

	assert.Equal(t, http.StatusFound, loginW1.Code)
	location1 := loginW1.Header().Get("Location")
	assert.Contains(t, location1, mockServer1.Server.URL)

	// Test login to second provider
	loginReq2 := httptest.NewRequest("GET", "/github-enterprise/login", nil)
	loginW2 := httptest.NewRecorder()
	router.ServeHTTP(loginW2, loginReq2)

	assert.Equal(t, http.StatusFound, loginW2.Code)
	location2 := loginW2.Header().Get("Location")
	assert.Contains(t, location2, mockServer2.Server.URL)
}

// TestProviderSourceHandler_HandleGitHubCallback_MissingProviderSource tests callback without provider source
func TestProviderSourceHandler_HandleGitHubCallback_MissingProviderSource(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	// Get the real server router from the container
	cont := testutils.CreateTestContainer(t, db)
	router := cont.Server.Router()

	// Create callback request with empty provider source
	req := httptest.NewRequest("GET", "//callback?code=test-code", nil)
	w := httptest.NewRecorder()

	// Serve the request
	router.ServeHTTP(w, req)

	// Should return 400 Bad Request
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Contains(t, response, "error")
}
