package util

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/persistence/sqldb"
)

// AuthTestHelper provides utilities for authentication testing
type AuthTestHelper struct {
	DBHelper *DatabaseTestHelper
}

// NewAuthTestHelper creates a new authentication test helper
func NewAuthTestHelper(t *testing.T) *AuthTestHelper {
	return &AuthTestHelper{
		DBHelper: NewTestDatabase(t),
	}
}

// Cleanup cleans up test resources
func (h *AuthTestHelper) Cleanup(t *testing.T) {
	h.DBHelper.Cleanup(t)
}

// CreateTestSession creates a test session
func (h *AuthTestHelper) CreateTestSession(t *testing.T, sessionID string) sqldb.SessionDB {
	session := sqldb.SessionDB{
		ID:     "test-session-" + sessionID,
		Expiry: time.Now().Add(1 * time.Hour),
		ProviderSourceAuth: []byte(`{
			"user_id": "user-` + sessionID + `",
			"username": "testuser-` + sessionID + `",
			"full_name": "Test User ` + sessionID + `",
			"email": "testuser-` + sessionID + `@example.com",
			"is_admin": false
		}`),
	}

	err := h.DBHelper.DB.DB.Create(&session).Error
	require.NoError(t, err)

	return session
}

// CreateTestAdminSession creates a test admin session
func (h *AuthTestHelper) CreateTestAdminSession(t *testing.T, sessionID string) sqldb.SessionDB {
	session := sqldb.SessionDB{
		ID:     "admin-session-" + sessionID,
		Expiry: time.Now().Add(1 * time.Hour),
		ProviderSourceAuth: []byte(`{
			"user_id": "admin-` + sessionID + `",
			"username": "admin-` + sessionID + `",
			"full_name": "Admin User ` + sessionID + `",
			"email": "admin-` + sessionID + `@example.com",
			"is_admin": true
		}`),
	}

	err := h.DBHelper.DB.DB.Create(&session).Error
	require.NoError(t, err)

	return session
}

// CreateExpiredSession creates an expired test session
func (h *AuthTestHelper) CreateExpiredSession(t *testing.T, sessionID string) sqldb.SessionDB {
	session := sqldb.SessionDB{
		ID:     "expired-session-" + sessionID,
		Expiry: time.Now().Add(-1 * time.Hour), // Expired 1 hour ago
		ProviderSourceAuth: []byte(`{
			"user_id": "expired-` + sessionID + `",
			"username": "expireduser-` + sessionID + `",
			"full_name": "Expired User ` + sessionID + `",
			"email": "expired-` + sessionID + `@example.com",
			"is_admin": false
		}`),
	}

	err := h.DBHelper.DB.DB.Create(&session).Error
	require.NoError(t, err)

	return session
}

// CreateTestUserGroup creates a test user group with specified permissions
func (h *AuthTestHelper) CreateTestUserGroup(t *testing.T, name string, siteAdmin bool) sqldb.UserGroupDB {
	userGroup := sqldb.UserGroupDB{
		Name:      name,
		SiteAdmin: siteAdmin,
	}

	err := h.DBHelper.DB.DB.Create(&userGroup).Error
	require.NoError(t, err)

	return userGroup
}

// CreateTestUserGroupNamespacePermission creates a test namespace permission for a user group
func (h *AuthTestHelper) CreateTestUserGroupNamespacePermission(
	t *testing.T,
	userGroupID int,
	namespaceID int,
	permissionType sqldb.UserGroupNamespacePermissionType,
) sqldb.UserGroupNamespacePermissionDB {
	permission := sqldb.UserGroupNamespacePermissionDB{
		UserGroupID:     userGroupID,
		NamespaceID:     namespaceID,
		PermissionType: permissionType,
	}

	err := h.DBHelper.DB.DB.Create(&permission).Error
	require.NoError(t, err)

	return permission
}

// ValidAPIKeys contains valid API keys for testing
var ValidAPIKeys = []string{
	"test-api-key-12345",
	"test-api-key-67890",
	"admin-api-key-abcdef",
}

// InvalidAPIKeys contains invalid API keys for testing
var InvalidAPIKeys = []string{
	"invalid-key",
	"malformed-key",
	"expired-key",
	"",
}

// CreateAuthContext creates a context with authentication headers
func CreateAuthContext(ctx context.Context, sessionID string) context.Context {
	return context.WithValue(ctx, "session_id", sessionID)
}

// CreateAPIKeyContext creates a context with API key authentication
func CreateAPIKeyContext(ctx context.Context, apiKey string) context.Context {
	return context.WithValue(ctx, "api_key", apiKey)
}

// GetAuthHeaders returns authentication headers for testing
func GetAuthHeaders(sessionID, csrfToken string) map[string]string {
	headers := map[string]string{
		"Cookie": "terrareg_session=" + sessionID,
	}
	if csrfToken != "" {
		headers["X-CSRF-Token"] = csrfToken
	}
	return headers
}

// GetAPIKeyHeaders returns API key authentication headers
func GetAPIKeyHeaders(apiKey string) map[string]string {
	return map[string]string{
		"Authorization": "Bearer " + apiKey,
	}
}

// AssertSessionValid checks if a session is valid and not expired
func AssertSessionValid(t *testing.T, session sqldb.SessionDB) {
	require.NotEmpty(t, session.ID)
	require.True(t, session.Expiry.After(time.Now()), "Session should not be expired")
	require.NotEmpty(t, session.ProviderSourceAuth, "Session should have auth data")
}

// AssertSessionExpired checks if a session is expired
func AssertSessionExpired(t *testing.T, session sqldb.SessionDB) {
	require.True(t, session.Expiry.Before(time.Now()), "Session should be expired")
}

// AssertAdminSession checks if a session belongs to an admin user
func AssertAdminSession(t *testing.T, session sqldb.SessionDB) {
	// Parse the JSON in ProviderSourceAuth to check admin status
	var authData map[string]interface{}
	err := json.Unmarshal(session.ProviderSourceAuth, &authData)
	require.NoError(t, err, "Session should have valid auth data")

	isAdmin, ok := authData["is_admin"].(bool)
	require.True(t, ok, "Session should have is_admin field")
	require.True(t, isAdmin, "Session should belong to an admin user")
}

// AssertUserGroupPermission checks if a user group has the expected namespace permission
func AssertUserGroupPermission(
	t *testing.T,
	permission sqldb.UserGroupNamespacePermissionDB,
	expectedNamespaceID int,
	expectedPermissionType sqldb.UserGroupNamespacePermissionType,
) {
	require.Equal(t, expectedNamespaceID, permission.NamespaceID)
	require.Equal(t, expectedPermissionType, permission.PermissionType)
}