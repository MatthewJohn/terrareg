package testutils

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/auth"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/auth/service"
	urlService "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/url/service"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/config"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/persistence/sqldb"
)

// BuildUnauthenticatedRequest creates a request without authentication
func BuildUnauthenticatedRequest(t *testing.T, method, url string) *http.Request {
	return httptest.NewRequest(method, url, nil)
}

// BuildAuthenticatedRequestWithSession creates a request with session authentication
// This properly creates a session in the database and returns a request with the session cookie set
// Returns the request (with cookie header set) and the cookie value
func BuildAuthenticatedRequestWithSession(t *testing.T, db *sqldb.Database, method, url, username string, isAdmin bool) (*http.Request, string) {
	// Create session in database
	sessionID := CreateTestSession(t, db, username, isAdmin)

	// Create cookie service to encrypt session cookie
	cfg := &config.InfrastructureConfig{
		SecretKey:         "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
		SessionCookieName: "terrareg_session",
	}
	urlService, err := urlService.NewURLService(cfg)
	require.NoError(t, err)
	cookieService, err := service.NewCookieService(cfg, urlService)
	require.NoError(t, err)

	// Create encrypted cookie value
	expiry := time.Now().Add(1 * time.Hour)
	sessionData := &auth.SessionData{
		SessionID:  sessionID,
		UserID:     username,
		Username:   username,
		AuthMethod: "session_password",
		IsAdmin:    isAdmin,
		SiteAdmin:  isAdmin,
		UserGroups: []string{},
		Expiry:     &expiry,
	}

	// Encrypt session data for cookie
	cookieValue, err := cookieService.EncryptSession(sessionData)
	require.NoError(t, err)

	// Create request with cookie
	req := httptest.NewRequest(method, url, nil)
	req.AddCookie(&http.Cookie{
		Name:  "terrareg_session",
		Value: cookieValue,
	})

	return req, cookieValue
}

// BuildAuthenticatedRequestWithNamespacePermission creates a request with namespace permissions
// This helper creates a user group with the specified namespace permission and returns
// a request with a session cookie that has that permission
func BuildAuthenticatedRequestWithNamespacePermission(
	t *testing.T,
	db *sqldb.Database,
	method, url, username, namespace string,
	permissionType sqldb.UserGroupNamespacePermissionType,
) (*http.Request, string) {
	// Create user group
	groupName := username + "-" + namespace + "-group"
	group := CreateTestAuthUserGroup(t, db, groupName, false)

	// Get or create namespace
	var namespaceDB sqldb.NamespaceDB
	err := db.DB.Where("namespace = ?", namespace).First(&namespaceDB).Error
	if err != nil {
		namespaceDB = CreateNamespace(t, db, namespace, nil)
	}

	// Create permission
	CreateTestNamespacePermission(t, db, group.ID, namespaceDB.ID, permissionType)

	// Create session in database
	sessionID := CreateTestSession(t, db, username, false)

	// Update session with user group info
	// First, get the existing session data to preserve username, auth_type, etc.
	var existingSession sqldb.SessionDB
	err = db.DB.Where("id = ?", sessionID).First(&existingSession).Error
	require.NoError(t, err)

	// Decode existing provider_source_auth
	var providerData map[string]interface{}
	err = json.Unmarshal(existingSession.ProviderSourceAuth, &providerData)
	require.NoError(t, err)

	// Add/update user groups and site_admin fields
	providerData["user_groups"] = []string{groupName}
	providerData["site_admin"] = false

	// Encode and update
	encodedData := sqldb.EncodeBlob(providerData)
	err = db.DB.Model(&sqldb.SessionDB{}).
		Where("id = ?", sessionID).
		Update("provider_source_auth", encodedData).Error
	require.NoError(t, err)

	// Create cookie service to encrypt session cookie
	cfg := &config.InfrastructureConfig{
		SecretKey:         "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
		SessionCookieName: "terrareg_session",
	}
	urlService, err := urlService.NewURLService(cfg)
	require.NoError(t, err)
	cookieService, err := service.NewCookieService(cfg, urlService)
	require.NoError(t, err)

	// Create encrypted cookie value with user groups
	expiry := time.Now().Add(1 * time.Hour)
	sessionData := &auth.SessionData{
		SessionID:  sessionID,
		UserID:     username,
		Username:   username,
		AuthMethod: "session_password",
		IsAdmin:    false,
		SiteAdmin:  false,
		UserGroups: []string{groupName},
		Expiry:     &expiry,
	}

	// Encrypt session data for cookie
	cookieValue, err := cookieService.EncryptSession(sessionData)
	require.NoError(t, err)

	// Create request with cookie
	req := httptest.NewRequest(method, url, nil)
	req.AddCookie(&http.Cookie{
		Name:  "terrareg_session",
		Value: cookieValue,
	})

	return req, cookieValue
}

// BuildAuthenticatedRequestWithMultipleNamespacePermissions creates a request with multiple namespace permissions
func BuildAuthenticatedRequestWithMultipleNamespacePermissions(
	t *testing.T,
	db *sqldb.Database,
	method, url, username string,
	namespacePermissions map[string]sqldb.UserGroupNamespacePermissionType,
) (*http.Request, string) {
	// Create user group
	groupName := username + "-multi-group"
	group := CreateTestAuthUserGroup(t, db, groupName, false)

	// Create permissions for each namespace and collect namespace names
	for namespace, permissionType := range namespacePermissions {
		// Get or create namespace
		var namespaceDB sqldb.NamespaceDB
		err := db.DB.Where("namespace = ?", namespace).First(&namespaceDB).Error
		if err != nil {
			namespaceDB = CreateNamespace(t, db, namespace, nil)
		}

		// Create permission
		CreateTestNamespacePermission(t, db, group.ID, namespaceDB.ID, permissionType)
	}

	// Create session in database
	sessionID := CreateTestSession(t, db, username, false)

	// Update session with user group info
	var providerData map[string]interface{}
	err := json.Unmarshal(
		[]byte(`{"user_groups":["`+groupName+`"],"site_admin":false}`),
		&providerData,
	)
	require.NoError(t, err)

	encodedData := sqldb.EncodeBlob(providerData)
	err = db.DB.Model(&sqldb.SessionDB{}).
		Where("id = ?", sessionID).
		Update("provider_source_auth", encodedData).Error
	require.NoError(t, err)

	// Create cookie service to encrypt session cookie
	cfg := &config.InfrastructureConfig{
		SecretKey:         "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
		SessionCookieName: "terrareg_session",
	}
	urlService, err := urlService.NewURLService(cfg)
	require.NoError(t, err)
	cookieService, err := service.NewCookieService(cfg, urlService)
	require.NoError(t, err)

	// Create encrypted cookie value with user groups
	expiry := time.Now().Add(1 * time.Hour)
	sessionData := &auth.SessionData{
		SessionID:  sessionID,
		UserID:     username,
		Username:   username,
		AuthMethod: "session_password",
		IsAdmin:    false,
		SiteAdmin:  false,
		UserGroups: []string{groupName},
		Expiry:     &expiry,
	}

	// Encrypt session data for cookie
	cookieValue, err := cookieService.EncryptSession(sessionData)
	require.NoError(t, err)

	// Create request with cookie
	req := httptest.NewRequest(method, url, nil)
	req.AddCookie(&http.Cookie{
		Name:  "terrareg_session",
		Value: cookieValue,
	})

	return req, cookieValue
}

// BuildAdminRequest creates an admin-authenticated request with session cookie
func BuildAdminRequest(t *testing.T, db *sqldb.Database, method, url string) (*http.Request, string) {
	return BuildAuthenticatedRequestWithSession(t, db, method, url, "admin-user", true)
}

// BuildAuthenticatedRequest is a compatibility wrapper that returns both request and AuthContext
// Note: This returns an AuthContext for reference, but the actual authentication is done via session cookie
func BuildAuthenticatedRequest(t *testing.T, db *sqldb.Database, method, url, username string, isAdmin bool) (*http.Request, auth.AuthContext) {
	req, _ := BuildAuthenticatedRequestWithSession(t, db, method, url, username, isAdmin)

	// Return test auth context for reference
	authCtx := &testAuthContext{
		username:        username,
		authMethod:      "session_password",
		isAdmin:         isAdmin,
		isAuthenticated: true,
	}

	return req, authCtx
}

// AuthTestCase represents a single authentication test scenario
type AuthTestCase struct {
	Name           string
	SetupAuth      func(*testing.T, *sqldb.Database) (*http.Request, string)
	ExpectedStatus int
}

// RunAuthTests executes multiple authentication test cases
// This is a helper for running table-driven authentication tests
func RunAuthTests(t *testing.T, handler http.HandlerFunc, tests []AuthTestCase) {
	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			db := SetupTestDatabase(t)
			defer CleanupTestDatabase(t, db)

			req, _ := tt.SetupAuth(t, db)

			w := httptest.NewRecorder()
			handler(w, req)

			require.Equal(t, tt.ExpectedStatus, w.Code, "Expected status code %d, got %d", tt.ExpectedStatus, w.Code)
		})
	}
}

// AssertStatusCode asserts that the response has the expected status code
func AssertStatusCode(t *testing.T, w *httptest.ResponseRecorder, expectedStatus int) {
	require.Equal(t, expectedStatus, w.Code, "Expected status code %d, got %d. Body: %s", expectedStatus, w.Code, w.Body.String())
}

// AssertUnauthorized asserts that the response is 401 Unauthorized
func AssertUnauthorized(t *testing.T, w *httptest.ResponseRecorder) {
	AssertStatusCode(t, w, http.StatusUnauthorized)
}

// AssertForbidden asserts that the response is 403 Forbidden
func AssertForbidden(t *testing.T, w *httptest.ResponseRecorder) {
	AssertStatusCode(t, w, http.StatusForbidden)
}

// AssertOK asserts that the response is 200 OK
func AssertOK(t *testing.T, w *httptest.ResponseRecorder) {
	AssertStatusCode(t, w, http.StatusOK)
}

// AssertCreated asserts that the response is 201 Created
func AssertCreated(t *testing.T, w *httptest.ResponseRecorder) {
	AssertStatusCode(t, w, http.StatusCreated)
}
