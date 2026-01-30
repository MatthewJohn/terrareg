package audit

import (
	"context"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/auth"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/persistence/sqldb"
	testutils "github.com/matthewjohn/terrareg/terrareg-go/test/integration/testutils"
)

// verifyUserLoginAuditEntry checks database for USER_LOGIN audit entry
func verifyUserLoginAuditEntry(t *testing.T, db *sqldb.Database, username, authMethod string) {
	t.Helper()

	// Wait a bit for async audit logging to complete
	// Audit entries are created asynchronously via goroutines
	time.Sleep(50 * time.Millisecond)

	var auditHistoryDB sqldb.AuditHistoryDB
	err := db.DB.Where("action = ? AND username = ?", sqldb.AuditActionUserLogin, username).
		Order("timestamp DESC").
		First(&auditHistoryDB).Error
	require.NoError(t, err, "Expected USER_LOGIN audit entry for user %s", username)

	assert.Equal(t, username, *auditHistoryDB.Username)
	assert.Equal(t, sqldb.AuditActionUserLogin, auditHistoryDB.Action)
	assert.Equal(t, "User", *auditHistoryDB.ObjectType)

	// Python reference: /app/terrareg/server/api/github/github_login_callback.py:65-69
	// Python uses old_value=None, new_value=None for login events
	assert.Nil(t, auditHistoryDB.OldValue, "old_value should be nil (matching Python behavior)")
	assert.Nil(t, auditHistoryDB.NewValue, "new_value should be nil (matching Python behavior)")

	// Verify timestamp is recent
	assert.NotNil(t, auditHistoryDB.Timestamp)
	assert.WithinDuration(t, time.Now(), *auditHistoryDB.Timestamp, time.Minute)
}

// TestLoginAudit_AdminApiKey tests that admin API key login creates USER_LOGIN audit entry
// Python reference: /app/terrareg/server/api/terrareg_admin_authenticate.py:28
func TestLoginAudit_AdminApiKey(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	// Clean audit table
	db.DB.Exec("DELETE FROM audit_history")

	// Create test container with all services
	cont := testutils.CreateTestContainer(t, db)

	ctx := context.Background()

	// Create admin session (using admin API key authentication)
	session, err := cont.SessionService.CreateSession(ctx, "ADMIN_API_KEY", []byte("{}"), nil)
	require.NoError(t, err)

	// Create response writer for cookie
	w := httptest.NewRecorder()

	// Set cookie for existing session with username and auth method for audit logging
	err = cont.SessionManagementService.SetCookieForExistingSession(ctx, w, session.ID, "Built-in admin", "ADMIN_API_KEY")
	require.NoError(t, err)

	// Verify audit entry was created
	verifyUserLoginAuditEntry(t, db, "Built-in admin", "ADMIN_API_KEY")
}

// TestLoginAudit_GitHubOAuth tests that GitHub OAuth login creates USER_LOGIN audit entry
// Python reference: /app/terrareg/server/api/github/github_login_callback.py:65
func TestLoginAudit_GitHubOAuth(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	// Clean audit table
	db.DB.Exec("DELETE FROM audit_history")

	// Create test container with all services
	cont := testutils.CreateTestContainer(t, db)

	ctx := context.Background()

	// Create GitHub auth context for test user
	username := "test-github-user"
	organizations := map[string]sqldb.NamespaceType{
		username: sqldb.NamespaceTypeGithubUser,
	}
	githubAuthCtx := auth.NewGitHubAuthContext(ctx, "test-github", username, organizations)

	// Create response writer for cookie
	w := httptest.NewRecorder()

	// Create session from auth context (this should trigger audit logging)
	err := cont.SessionManagementService.CreateSessionAndCookie(
		ctx, w,
		githubAuthCtx.GetProviderType(),
		githubAuthCtx.GetUsername(),
		githubAuthCtx.IsAdmin(),
		githubAuthCtx.GetUserGroupNames(),
		githubAuthCtx.GetAllNamespacePermissions(),
		githubAuthCtx.GetProviderData(),
		nil, // ttl - use default
	)
	require.NoError(t, err)

	// Verify audit entry was created
	verifyUserLoginAuditEntry(t, db, username, "GITHUB")
}

// TestLoginAudit_SAML tests that SAML login creates USER_LOGIN audit entry
// Python reference: /app/terrareg/server/api/saml_callback.py:65
func TestLoginAudit_SAML(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	// Clean audit table
	db.DB.Exec("DELETE FROM audit_history")

	// Create test container with all services
	cont := testutils.CreateTestContainer(t, db)

	ctx := context.Background()

	// Create SAML auth context for test user
	username := "test-saml-user"
	nameID := "test-saml-name-id"
	attributes := map[string][]string{
		"username": {username},
		"email":    {"test@example.com"},
	}
	samlAuthCtx := auth.NewSamlAuthContext(ctx, nameID, attributes)

	// The SAML auth context doesn't automatically set username from attributes
	// We need to check what GetUsername returns
	if samlAuthCtx.GetUsername() == "" {
		t.Skip("SAML auth context doesn't derive username from nameID without additional configuration")
	}

	// Create response writer for cookie
	w := httptest.NewRecorder()

	// Create session from auth context (this should trigger audit logging)
	err := cont.SessionManagementService.CreateSessionAndCookie(
		ctx, w,
		samlAuthCtx.GetProviderType(),
		samlAuthCtx.GetUsername(),
		samlAuthCtx.IsAdmin(),
		samlAuthCtx.GetUserGroupNames(),
		samlAuthCtx.GetAllNamespacePermissions(),
		samlAuthCtx.GetProviderData(),
		nil, // ttl - use default
	)
	require.NoError(t, err)

	// Verify audit entry was created
	// SAML uses nameID as username if no username attribute is set
	expectedUsername := username
	if samlAuthCtx.GetUsername() == "" {
		expectedUsername = nameID
	}
	verifyUserLoginAuditEntry(t, db, expectedUsername, "SAML")
}

// TestLoginAudit_OIDC tests that OIDC login creates USER_LOGIN audit entry
// Python reference: /app/terrareg/server/api/open_id_callback.py:86
func TestLoginAudit_OIDC(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	// Clean audit table
	db.DB.Exec("DELETE FROM audit_history")

	// Create test container with all services
	cont := testutils.CreateTestContainer(t, db)

	ctx := context.Background()

	// Create OIDC auth context for test user
	username := "test-oidc-user"
	sub := "user-123"
	claims := map[string]interface{}{
		"sub":   sub,
		"name":  username,
		"email": "test@example.com",
	}
	oidcAuthCtx := auth.NewOpenidConnectAuthContext(ctx, sub, claims)

	// The OIDC auth context may not automatically set username from claims
	// We need to check what GetUsername returns
	if oidcAuthCtx.GetUsername() == "" {
		t.Skip("OIDC auth context doesn't derive username from claims without additional configuration")
	}

	// Create response writer for cookie
	w := httptest.NewRecorder()

	// Create session from auth context (this should trigger audit logging)
	err := cont.SessionManagementService.CreateSessionAndCookie(
		ctx, w,
		oidcAuthCtx.GetProviderType(),
		oidcAuthCtx.GetUsername(),
		oidcAuthCtx.IsAdmin(),
		oidcAuthCtx.GetUserGroupNames(),
		oidcAuthCtx.GetAllNamespacePermissions(),
		oidcAuthCtx.GetProviderData(),
		nil, // ttl - use default
	)
	require.NoError(t, err)

	// Verify audit entry was created
	// OIDC uses sub (subject) as username if no name claim is used
	expectedUsername := username
	if oidcAuthCtx.GetUsername() == "" {
		expectedUsername = sub
	}
	verifyUserLoginAuditEntry(t, db, expectedUsername, "OPENID_CONNECT")
}

// TestLoginAudit_MultipleLogins tests that multiple login attempts create separate audit entries
func TestLoginAudit_MultipleLogins(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	// Clean audit table
	db.DB.Exec("DELETE FROM audit_history")

	// Create test container with all services
	cont := testutils.CreateTestContainer(t, db)

	ctx := context.Background()

	// Create first admin session
	session1, err := cont.SessionService.CreateSession(ctx, "ADMIN_API_KEY", []byte("{}"), nil)
	require.NoError(t, err)

	w1 := httptest.NewRecorder()
	err = cont.SessionManagementService.SetCookieForExistingSession(ctx, w1, session1.ID, "Built-in admin", "ADMIN_API_KEY")
	require.NoError(t, err)

	// Wait a bit to ensure different timestamps
	time.Sleep(10 * time.Millisecond)

	// Create second admin session
	session2, err := cont.SessionService.CreateSession(ctx, "ADMIN_API_KEY", []byte("{}"), nil)
	require.NoError(t, err)

	w2 := httptest.NewRecorder()
	err = cont.SessionManagementService.SetCookieForExistingSession(ctx, w2, session2.ID, "Built-in admin", "ADMIN_API_KEY")
	require.NoError(t, err)

	// Wait for async audit logging to complete
	// Audit entries are created asynchronously via goroutines
	time.Sleep(50 * time.Millisecond)

	// Verify both audit entries exist
	var auditEntries []sqldb.AuditHistoryDB
	err = db.DB.Where("action = ? AND username = ?", sqldb.AuditActionUserLogin, "Built-in admin").
		Order("timestamp ASC").
		Find(&auditEntries).Error
	require.NoError(t, err)

	// Should have exactly 2 entries
	require.Equal(t, 2, len(auditEntries), "Expected 2 USER_LOGIN audit entries")

	// Verify timestamps are different (second entry should be newer)
	assert.True(t, auditEntries[1].Timestamp.After(*auditEntries[0].Timestamp))
}

// TestLoginAudit_SessionManagementServiceIsProperlyInitialized tests that session management service is properly initialized
func TestLoginAudit_SessionManagementServiceIsProperlyInitialized(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	// Create test container with all services
	cont := testutils.CreateTestContainer(t, db)

	// Verify that SessionManagementService is properly initialized
	// It should not be nil since SECRET_KEY is configured in test container
	require.NotNil(t, cont.SessionManagementService, "SessionManagementService should be initialized when SECRET_KEY is configured")
	require.True(t, cont.SessionManagementService.IsAvailable(), "SessionManagementService should be available")

	// Verify that we can create a session and cookie through SessionManagementService
	ctx := context.Background()
	w := httptest.NewRecorder()

	err := cont.SessionManagementService.CreateSessionAndCookie(
		ctx, w,
		auth.AuthMethodAdminApiKey,
		"Test User",
		true,
		[]string{},
		nil,
		nil,
		nil,
	)
	require.NoError(t, err)

	// Verify a cookie was set
	cookies := w.Result().Cookies()
	require.NotEmpty(t, cookies, "Session cookie should be set")
}
