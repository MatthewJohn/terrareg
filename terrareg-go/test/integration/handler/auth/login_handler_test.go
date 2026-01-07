// Package auth_test provides integration tests for the auth HTTP handlers
package auth_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	authCmd "github.com/matthewjohn/terrareg/terrareg-go/internal/application/command/auth"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/interfaces/http/handler/auth"
	"github.com/matthewjohn/terrareg/terrareg-go/test/integration/testutils"
)

// TestLoginHandler_HandleLogin_Success_ValidCredentials tests successful admin login
func TestLoginHandler_HandleLogin_Success_ValidCredentials(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	// Create handler with real dependencies
	cookieSessionService, cont := testutils.CreateTestCookieSessionService(t, db)
	createSessionCmd := authCmd.NewCreateSessionCommand(cookieSessionService)
	handler := auth.NewLoginHandler(createSessionCmd, cookieSessionService, testutils.TestLogger)

	// Create request with valid credentials
	loginReq := auth.LoginRequest{
		Username: "test-admin",
		Password: "test-password",
	}
	reqBody, err := json.Marshal(loginReq)
	require.NoError(t, err)

	req := httptest.NewRequest("POST", "/v1/terrareg/auth/admin/login", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Act
	handler.HandleLogin(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

	var response auth.LoginResponse
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.True(t, response.Success)

	// Verify session cookie was set
	cookies := w.Result().Cookies()
	assert.NotEmpty(t, cookies, "Session cookie should be set")

	var sessionCookie *http.Cookie
	for _, c := range cookies {
		if c.Name == "terrareg_session" {
			sessionCookie = c
			break
		}
	}
	require.NotNil(t, sessionCookie, "Session cookie should be set")
	assert.NotEmpty(t, sessionCookie.Value, "Session cookie value should not be empty")

	// Decrypt the cookie to get session data
	// We need to use the CookieService directly to decrypt the cookie
	// Use the container from CreateTestCookieSessionService to ensure SECRET_KEY consistency
	decryptedSessionData, err := cont.CookieService.DecryptSession(sessionCookie.Value)
	require.NoError(t, err, "Session cookie should be decryptable")
	require.NotNil(t, decryptedSessionData)

	// Now validate the session using the session ID from the decrypted data
	sessionData, err := cookieSessionService.ValidateSession(req.Context(), decryptedSessionData.SessionID)
	require.NoError(t, err)
	assert.NotNil(t, sessionData)
}

// TestLoginHandler_HandleLogin_Failure_InvalidUsername tests login with empty username
func TestLoginHandler_HandleLogin_Failure_InvalidUsername(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	// Create handler
	cookieSessionService, _ := testutils.CreateTestCookieSessionService(t, db)
	createSessionCmd := authCmd.NewCreateSessionCommand(cookieSessionService)
	handler := auth.NewLoginHandler(createSessionCmd, cookieSessionService, testutils.TestLogger)

	// Create request with empty username
	loginReq := auth.LoginRequest{
		Username: "",
		Password: "test-password",
	}
	reqBody, err := json.Marshal(loginReq)
	require.NoError(t, err)

	req := httptest.NewRequest("POST", "/v1/terrareg/auth/admin/login", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Act
	handler.HandleLogin(w, req)

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response auth.LoginResponse
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.False(t, response.Success)
	assert.Equal(t, "Username and password are required", response.Message)
}

// TestLoginHandler_HandleLogin_Failure_InvalidPassword tests login with empty password
func TestLoginHandler_HandleLogin_Failure_InvalidPassword(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	// Create handler
	cookieSessionService, _ := testutils.CreateTestCookieSessionService(t, db)
	createSessionCmd := authCmd.NewCreateSessionCommand(cookieSessionService)
	handler := auth.NewLoginHandler(createSessionCmd, cookieSessionService, testutils.TestLogger)

	// Create request with empty password
	loginReq := auth.LoginRequest{
		Username: "test-admin",
		Password: "",
	}
	reqBody, err := json.Marshal(loginReq)
	require.NoError(t, err)

	req := httptest.NewRequest("POST", "/v1/terrareg/auth/admin/login", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Act
	handler.HandleLogin(w, req)

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response auth.LoginResponse
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.False(t, response.Success)
	assert.Equal(t, "Username and password are required", response.Message)
}

// TestLoginHandler_HandleLogin_Failure_EmptyUsername tests login with missing username
func TestLoginHandler_HandleLogin_Failure_EmptyUsername(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	// Create handler
	cookieSessionService, _ := testutils.CreateTestCookieSessionService(t, db)
	createSessionCmd := authCmd.NewCreateSessionCommand(cookieSessionService)
	handler := auth.NewLoginHandler(createSessionCmd, cookieSessionService, testutils.TestLogger)

	// Create request with no username field
	loginReq := map[string]string{
		"password": "test-password",
	}
	reqBody, err := json.Marshal(loginReq)
	require.NoError(t, err)

	req := httptest.NewRequest("POST", "/v1/terrareg/auth/admin/login", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Act
	handler.HandleLogin(w, req)

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// TestLoginHandler_HandleLogin_Failure_EmptyPassword tests login with missing password
func TestLoginHandler_HandleLogin_Failure_EmptyPassword(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	// Create handler
	cookieSessionService, _ := testutils.CreateTestCookieSessionService(t, db)
	createSessionCmd := authCmd.NewCreateSessionCommand(cookieSessionService)
	handler := auth.NewLoginHandler(createSessionCmd, cookieSessionService, testutils.TestLogger)

	// Create request with no password field
	loginReq := map[string]string{
		"username": "test-admin",
	}
	reqBody, err := json.Marshal(loginReq)
	require.NoError(t, err)

	req := httptest.NewRequest("POST", "/v1/terrareg/auth/admin/login", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Act
	handler.HandleLogin(w, req)

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// TestLoginHandler_HandleLogin_CreatesSessionInDatabase verifies session is created in DB
func TestLoginHandler_HandleLogin_CreatesSessionInDatabase(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	// Create handler
	cookieSessionService, cont := testutils.CreateTestCookieSessionService(t, db)
	createSessionCmd := authCmd.NewCreateSessionCommand(cookieSessionService)
	handler := auth.NewLoginHandler(createSessionCmd, cookieSessionService, testutils.TestLogger)

	// Create request
	loginReq := auth.LoginRequest{
		Username: "test-admin",
		Password: "test-password",
	}
	reqBody, err := json.Marshal(loginReq)
	require.NoError(t, err)

	req := httptest.NewRequest("POST", "/v1/terrareg/auth/admin/login", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Act
	handler.HandleLogin(w, req)

	// Assert - verify session was created in database
	cookies := w.Result().Cookies()
	require.NotEmpty(t, cookies)

	var sessionCookie *http.Cookie
	for _, c := range cookies {
		if c.Name == "terrareg_session" {
			sessionCookie = c
			break
		}
	}
	require.NotNil(t, sessionCookie)

	// Decrypt cookie to get session ID
	decryptedSessionData, err := cont.CookieService.DecryptSession(sessionCookie.Value)
	require.NoError(t, err)
	require.NotNil(t, decryptedSessionData)

	// Verify session exists in database
	testutils.AssertSessionExists(t, db, decryptedSessionData.SessionID)
}

// TestLoginHandler_HandleLogin_SetsEncryptedCookie verifies cookie encryption
func TestLoginHandler_HandleLogin_SetsEncryptedCookie(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	// Create handler
	cookieSessionService, _ := testutils.CreateTestCookieSessionService(t, db)
	createSessionCmd := authCmd.NewCreateSessionCommand(cookieSessionService)
	handler := auth.NewLoginHandler(createSessionCmd, cookieSessionService, testutils.TestLogger)

	// Create request
	loginReq := auth.LoginRequest{
		Username: "test-admin",
		Password: "test-password",
	}
	reqBody, err := json.Marshal(loginReq)
	require.NoError(t, err)

	req := httptest.NewRequest("POST", "/v1/terrareg/auth/admin/login", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Act
	handler.HandleLogin(w, req)

	// Assert - verify cookie is encrypted (not plain text)
	cookies := w.Result().Cookies()
	require.NotEmpty(t, cookies)

	var sessionCookie *http.Cookie
	for _, c := range cookies {
		if c.Name == "terrareg_session" {
			sessionCookie = c
			break
		}
	}
	require.NotNil(t, sessionCookie)

	// Cookie value should be base64-encoded (encrypted)
	// Plain session ID would be like "test-session-test-admin"
	// Encrypted value should be longer and base64-encoded
	assert.NotContains(t, sessionCookie.Value, "test-admin", "Cookie should be encrypted, not plain text")
	assert.Greater(t, len(sessionCookie.Value), 20, "Encrypted cookie should be longer than plain session ID")
}

// TestLoginHandler_HandleLogout_Success_WithSession tests logout with active session
func TestLoginHandler_HandleLogout_Success_WithSession(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	// Create handler
	cookieSessionService, _ := testutils.CreateTestCookieSessionService(t, db)
	createSessionCmd := authCmd.NewCreateSessionCommand(cookieSessionService)
	handler := auth.NewLoginHandler(createSessionCmd, cookieSessionService, testutils.TestLogger)

	// Create request
	req := httptest.NewRequest("POST", "/v1/terrareg/auth/admin/logout", nil)
	w := httptest.NewRecorder()

	// Act
	handler.HandleLogout(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

	var response auth.LoginResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.True(t, response.Success)

	// Verify session cookie was cleared
	cookies := w.Result().Cookies()
	var sessionCookie *http.Cookie
	for _, c := range cookies {
		if c.Name == "terrareg_session" {
			sessionCookie = c
			break
		}
	}
	require.NotNil(t, sessionCookie, "Session cookie should be cleared")
	assert.Equal(t, "", sessionCookie.Value, "Session cookie value should be empty")
	assert.Equal(t, -1, sessionCookie.MaxAge, "Session cookie MaxAge should be -1 (deleted)")
}

// TestLoginHandler_HandleLogout_Success_ClearsCookie verifies cookie is cleared
func TestLoginHandler_HandleLogout_Success_ClearsCookie(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	// Create handler
	cookieSessionService, _ := testutils.CreateTestCookieSessionService(t, db)
	createSessionCmd := authCmd.NewCreateSessionCommand(cookieSessionService)
	handler := auth.NewLoginHandler(createSessionCmd, cookieSessionService, testutils.TestLogger)

	// Create request
	req := httptest.NewRequest("POST", "/v1/terrareg/auth/admin/logout", nil)
	w := httptest.NewRecorder()

	// Act
	handler.HandleLogout(w, req)

	// Assert - verify all cookie clearing properties
	cookies := w.Result().Cookies()
	assert.NotEmpty(t, cookies, "Cookies should be set in response")

	var sessionCookie *http.Cookie
	for _, c := range cookies {
		if c.Name == "terrareg_session" {
			sessionCookie = c
			break
		}
	}
	require.NotNil(t, sessionCookie)
	assert.Empty(t, sessionCookie.Value, "Cookie value should be empty")
	assert.Equal(t, -1, sessionCookie.MaxAge, "Cookie MaxAge should be -1")
}

// TestLoginHandler_HandleLogout_Success_WithoutSession tests logout without active session
func TestLoginHandler_HandleLogout_Success_WithoutSession(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	// Create handler
	cookieSessionService, _ := testutils.CreateTestCookieSessionService(t, db)
	createSessionCmd := authCmd.NewCreateSessionCommand(cookieSessionService)
	handler := auth.NewLoginHandler(createSessionCmd, cookieSessionService, testutils.TestLogger)

	// Create request without any session cookie
	req := httptest.NewRequest("POST", "/v1/terrareg/auth/admin/logout", nil)
	w := httptest.NewRecorder()

	// Act
	handler.HandleLogout(w, req)

	// Assert - logout should still succeed even without a session
	assert.Equal(t, http.StatusOK, w.Code)

	var response auth.LoginResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.True(t, response.Success)
}

// TestLoginHandler_HandleLogin_InvalidRequestMethod tests login with GET instead of POST
func TestLoginHandler_HandleLogin_InvalidRequestMethod(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	// Create handler
	cookieSessionService, _ := testutils.CreateTestCookieSessionService(t, db)
	createSessionCmd := authCmd.NewCreateSessionCommand(cookieSessionService)
	handler := auth.NewLoginHandler(createSessionCmd, cookieSessionService, testutils.TestLogger)

	// Create request with GET method instead of POST
	req := httptest.NewRequest("GET", "/v1/terrareg/auth/admin/login", nil)
	w := httptest.NewRecorder()

	// Act
	handler.HandleLogin(w, req)

	// Assert
	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
}

// TestLoginHandler_HandleLogin_InvalidRequestBody tests login with invalid JSON
func TestLoginHandler_HandleLogin_InvalidRequestBody(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	// Create handler
	cookieSessionService, _ := testutils.CreateTestCookieSessionService(t, db)
	createSessionCmd := authCmd.NewCreateSessionCommand(cookieSessionService)
	handler := auth.NewLoginHandler(createSessionCmd, cookieSessionService, testutils.TestLogger)

	// Create request with invalid JSON
	req := httptest.NewRequest("POST", "/v1/terrareg/auth/admin/login", bytes.NewBuffer([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Act
	handler.HandleLogin(w, req)

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response auth.LoginResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.False(t, response.Success)
	assert.Equal(t, "Invalid request body", response.Message)
}
