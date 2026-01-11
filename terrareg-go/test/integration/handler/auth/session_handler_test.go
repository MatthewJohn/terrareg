// Package auth_test provides integration tests for the session HTTP handlers
package auth_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/interfaces/http/handler/auth"
	"github.com/matthewjohn/terrareg/terrareg-go/test/integration/testutils"
)

// TestSessionHandler_HandleGetSession_Success_WithValidSession tests getting session info with valid session
func TestSessionHandler_HandleGetSession_Success_WithValidSession(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	// Create handler
	cookieSessionService, _ := testutils.CreateTestCookieSessionService(t, db)
	handler := auth.NewSessionHandler(cookieSessionService, testutils.TestLogger)

	// Create authenticated request with session
	req, _ := testutils.CreateAuthenticatedRequestWithSession(t, db, "GET", "/v1/terrareg/auth/session", "testuser", true)
	w := httptest.NewRecorder()

	// Act
	handler.HandleGetSession(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

	var response auth.GetSessionResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.True(t, response.Authenticated)
	assert.Equal(t, "testuser", response.Username)
	assert.True(t, response.IsAdmin)
}

// TestSessionHandler_HandleGetSession_Failure_NoSession tests getting session without authentication
func TestSessionHandler_HandleGetSession_Failure_NoSession(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	// Create handler
	cookieSessionService, _ := testutils.CreateTestCookieSessionService(t, db)
	handler := auth.NewSessionHandler(cookieSessionService, testutils.TestLogger)

	// Create request without session
	req := httptest.NewRequest("GET", "/v1/terrareg/auth/session", nil)
	w := httptest.NewRecorder()

	// Act
	handler.HandleGetSession(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)

	var response auth.GetSessionResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.False(t, response.Authenticated)
	assert.Empty(t, response.Username)
}

// TestSessionHandler_HandleGetSession_Success_IncludesUserData tests that user data is included
func TestSessionHandler_HandleGetSession_Success_IncludesUserData(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	// Create handler
	cookieSessionService, _ := testutils.CreateTestCookieSessionService(t, db)
	handler := auth.NewSessionHandler(cookieSessionService, testutils.TestLogger)

	// Create authenticated request with user groups
	req, _ := testutils.CreateAuthenticatedRequestWithSession(t, db, "GET", "/v1/terrareg/auth/session", "testuser", true)
	w := httptest.NewRecorder()

	// Act
	handler.HandleGetSession(w, req)

	// Assert
	var response auth.GetSessionResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	// Verify all user fields are present
	assert.True(t, response.Authenticated)
	assert.NotEmpty(t, response.UserID)
	assert.NotEmpty(t, response.Username)
	assert.NotEmpty(t, response.AuthMethod)
	// IsAdmin is a bool, so we check the value
	assert.True(t, response.IsAdmin)
	// UserGroups may be nil if empty (due to omitempty JSON tag)
	// This is expected behavior - empty slices are omitted from JSON
}

// TestSessionHandler_HandleDeleteSession_Success_WithValidSession tests deleting an active session
func TestSessionHandler_HandleDeleteSession_Success_WithValidSession(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	// Create handler
	cookieSessionService, _ := testutils.CreateTestCookieSessionService(t, db)
	handler := auth.NewSessionHandler(cookieSessionService, testutils.TestLogger)

	// Create authenticated request with session
	req, _ := testutils.CreateAuthenticatedRequestWithSession(t, db, "DELETE", "/v1/terrareg/auth/session", "testuser", true)
	w := httptest.NewRecorder()

	// Get session ID before deletion
	// For this test, we know the session ID format

	// Act
	handler.HandleDeleteSession(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]bool
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.True(t, response["success"])

	// Verify cookie was cleared
	cookies := w.Result().Cookies()
	var clearedCookie *http.Cookie
	for _, c := range cookies {
		if c.Name == "terrareg_session" {
			clearedCookie = c
			break
		}
	}
	require.NotNil(t, clearedCookie, "Session cookie should be cleared")
	assert.Equal(t, "", clearedCookie.Value, "Cookie value should be empty")
}

// TestSessionHandler_HandleDeleteSession_Failure_NoSession tests deleting session without authentication
func TestSessionHandler_HandleDeleteSession_Failure_NoSession(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	// Create handler
	cookieSessionService, _ := testutils.CreateTestCookieSessionService(t, db)
	handler := auth.NewSessionHandler(cookieSessionService, testutils.TestLogger)

	// Create request without session
	req := httptest.NewRequest("DELETE", "/v1/terrareg/auth/session", nil)
	w := httptest.NewRecorder()

	// Act
	handler.HandleDeleteSession(w, req)

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// TestSessionHandler_HandleDeleteSession_Success_ClearsCookie tests that cookie is cleared on logout
func TestSessionHandler_HandleDeleteSession_Success_ClearsCookie(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	// Create handler
	cookieSessionService, _ := testutils.CreateTestCookieSessionService(t, db)
	handler := auth.NewSessionHandler(cookieSessionService, testutils.TestLogger)

	// Create authenticated request
	req, _ := testutils.CreateAuthenticatedRequestWithSession(t, db, "DELETE", "/v1/terrareg/auth/session", "testuser", true)
	w := httptest.NewRecorder()

	// Act
	handler.HandleDeleteSession(w, req)

	// Assert - verify cookie clearing properties
	cookies := w.Result().Cookies()
	assert.NotEmpty(t, cookies, "Cookies should be set in response")

	var clearedCookie *http.Cookie
	for _, c := range cookies {
		if c.Name == "terrareg_session" {
			clearedCookie = c
			break
		}
	}
	require.NotNil(t, clearedCookie, "Session cookie should be cleared")
	assert.Empty(t, clearedCookie.Value, "Cookie value should be empty")
	assert.Equal(t, -1, clearedCookie.MaxAge, "Cookie MaxAge should be -1")
}

// TestSessionHandler_HandleRefreshSession_Success_ValidSession tests refreshing a session
func TestSessionHandler_HandleRefreshSession_Success_ValidSession(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	// Create handler
	cookieSessionService, _ := testutils.CreateTestCookieSessionService(t, db)
	handler := auth.NewSessionHandler(cookieSessionService, testutils.TestLogger)

	// Create authenticated request
	req, _ := testutils.CreateAuthenticatedRequestWithSession(t, db, "POST", "/v1/terrareg/auth/session/refresh", "testuser", true)
	w := httptest.NewRecorder()

	// Act
	handler.HandleRefreshSession(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]bool
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.True(t, response["success"])

	// Verify cookie was refreshed (new cookie should be set)
	cookies := w.Result().Cookies()
	var refreshedCookie *http.Cookie
	for _, c := range cookies {
		if c.Name == "terrareg_session" {
			refreshedCookie = c
			break
		}
	}
	require.NotNil(t, refreshedCookie, "Session cookie should be refreshed")
	assert.NotEmpty(t, refreshedCookie.Value, "Cookie value should not be empty")
}

// TestSessionHandler_HandleRefreshSession_Failure_NoSession tests refreshing without authentication
func TestSessionHandler_HandleRefreshSession_Failure_NoSession(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	// Create handler
	cookieSessionService, _ := testutils.CreateTestCookieSessionService(t, db)
	handler := auth.NewSessionHandler(cookieSessionService, testutils.TestLogger)

	// Create request without session
	req := httptest.NewRequest("POST", "/v1/terrareg/auth/session/refresh", nil)
	w := httptest.NewRecorder()

	// Act
	handler.HandleRefreshSession(w, req)

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// TestSessionHandler_HandleRefreshSession_ReturnsUpdatedSession tests that refresh returns updated session
func TestSessionHandler_HandleRefreshSession_ReturnsUpdatedSession(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	// Create handler
	cookieSessionService, _ := testutils.CreateTestCookieSessionService(t, db)
	handler := auth.NewSessionHandler(cookieSessionService, testutils.TestLogger)

	// Create authenticated request
	req, _ := testutils.CreateAuthenticatedRequestWithSession(t, db, "POST", "/v1/terrareg/auth/session/refresh", "testuser", true)
	w := httptest.NewRecorder()

	// Act
	handler.HandleRefreshSession(w, req)

	// Assert - verify response indicates success
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]bool
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.True(t, response["success"])

	// Verify cookie was set with refreshed value
	cookies := w.Result().Cookies()
	assert.NotEmpty(t, cookies, "New cookie should be set after refresh")
}
