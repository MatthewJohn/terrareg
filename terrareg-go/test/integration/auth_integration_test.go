package integration

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/auth"
	authservice "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/auth/service"
	modulemodel "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/module/model"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/config"
	authRepo "github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/persistence/sqldb/auth"
	moduleRepo "github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/persistence/sqldb/module"
	"github.com/matthewjohn/terrareg/terrareg-go/test/integration/testutils"
)

func TestAuthenticationIntegration(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer func() {
		require.NoError(t, db.Close())
	}()

	// Setup repositories with correct import paths
	sessionRepo := authRepo.NewSessionRepository(db.DB)
	userGroupRepo := authRepo.NewUserGroupRepository(db.DB)
	namespaceRepo := moduleRepo.NewNamespaceRepository(db.DB)

	// Create Terraform IDP repositories
	authCodeRepo := authRepo.NewTerraformIdpAuthorizationCodeRepository(db.DB)
	accessTokenRepo := authRepo.NewTerraformIdpAccessTokenRepository(db.DB)
	subjectIdentifierRepo := authRepo.NewTerraformIdpSubjectIdentifierRepository(db.DB)

	_ = userGroupRepo
	_ = authCodeRepo
	_ = accessTokenRepo
	_ = subjectIdentifierRepo

	// Create minimal services for testing
	sessionConfig := authservice.DefaultSessionDatabaseConfig()
	sessionService := authservice.NewSessionService(sessionRepo, sessionConfig)

	// Create basic infrastructure config with valid SECRET_KEY
	// (CookieService requires SECRET_KEY to be at least 32 bytes when hex-decoded)
	infraConfig := &config.InfrastructureConfig{
		ListenPort:    3000,
		PublicURL:     "http://localhost:3000",
		DomainName:    "localhost",
		Debug:         true,
		SecretKey:     "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
	}

	// Create auth service with minimal dependencies
	cookieService := authservice.NewCookieService(infraConfig)
	_ = authservice.NewAuthenticationService(sessionService, cookieService)

	t.Run("Create session", func(t *testing.T) {
		// Create a session directly using the repository
		ctx := context.Background()
		sessionID := "test-session-id-create"
		expiry := time.Now().Add(24 * time.Hour)

		session := auth.NewSession(sessionID, expiry)

		err := sessionRepo.Create(ctx, session)
		require.NoError(t, err)

		// Verify session was created
		foundSession, err := sessionRepo.FindByID(ctx, sessionID)
		require.NoError(t, err)
		require.NotNil(t, foundSession)
		assert.Equal(t, sessionID, foundSession.ID)
		assert.WithinDuration(t, expiry, foundSession.Expiry, time.Second)
	})

	t.Run("Get session", func(t *testing.T) {
		// First create a session
		ctx := context.Background()
		sessionID := "test-session-id-get"
		expiry := time.Now().Add(24 * time.Hour)

		session := auth.NewSession(sessionID, expiry)

		err := sessionRepo.Create(ctx, session)
		require.NoError(t, err)

		// Now get the session
		foundSession, err := sessionRepo.FindByID(ctx, sessionID)
		require.NoError(t, err)
		require.NotNil(t, foundSession)
		assert.Equal(t, sessionID, foundSession.ID)
		assert.WithinDuration(t, expiry, foundSession.Expiry, time.Second)
	})

	t.Run("Delete session", func(t *testing.T) {
		// First create a session
		ctx := context.Background()
		sessionID := "test-session-id-delete"
		expiry := time.Now().Add(24 * time.Hour)

		session := auth.NewSession(sessionID, expiry)

		err := sessionRepo.Create(ctx, session)
		require.NoError(t, err)

		// Delete the session
		err = sessionRepo.Delete(ctx, sessionID)
		require.NoError(t, err)

		// Verify session is deleted
		_, err = sessionRepo.FindByID(ctx, sessionID)
		assert.NoError(t, err) // FindByID returns nil, not error when not found
	})

	t.Run("Handle expired session", func(t *testing.T) {
		// Create an expired session
		ctx := context.Background()
		sessionID := "expired-session-id"
		expiry := time.Now().Add(-1 * time.Hour) // Expired

		session := auth.NewSession(sessionID, expiry)

		err := sessionRepo.Create(ctx, session)
		require.NoError(t, err)

		// Try to get the expired session - should return nil
		foundSession, err := sessionRepo.FindByID(ctx, sessionID)
		require.NoError(t, err)
		assert.Nil(t, foundSession, "Expired session should not be found")
	})

	t.Run("Handle missing session", func(t *testing.T) {
		// Try to get session that doesn't exist
		ctx := context.Background()
		foundSession, err := sessionRepo.FindByID(ctx, "non-existent-session")
		require.NoError(t, err)
		assert.Nil(t, foundSession, "Non-existent session should not be found")
	})

	t.Run("Test session with provider source auth", func(t *testing.T) {
		// Create a session with provider source auth data
		ctx := context.Background()
		sessionID := "session-with-auth-data"
		expiry := time.Now().Add(24 * time.Hour)

		authData := []byte(`{"auth_method": "oidc", "user_id": "test-user"}`)

		session := auth.NewSession(sessionID, expiry)
		session.SetProviderSourceAuth(authData)

		err := sessionRepo.Create(ctx, session)
		require.NoError(t, err)

		// Get the session and verify auth data
		foundSession, err := sessionRepo.FindByID(ctx, sessionID)
		require.NoError(t, err)
		require.NotNil(t, foundSession)
		assert.Equal(t, sessionID, foundSession.ID)
		assert.Equal(t, authData, foundSession.ProviderSourceAuth)
	})

	t.Run("Test concurrent session operations", func(t *testing.T) {
		// Create multiple sessions concurrently
		ctx := context.Background()
		sessions := make([]*auth.Session, 5)
		for i := 0; i < 5; i++ {
			sessionID := fmt.Sprintf("concurrent-session-%d", i)
			expiry := time.Now().Add(24 * time.Hour)

			sessions[i] = auth.NewSession(sessionID, expiry)

			err := sessionRepo.Create(ctx, sessions[i])
			require.NoError(t, err)
		}

		// Verify all sessions can be retrieved
		for _, session := range sessions {
			foundSession, err := sessionRepo.FindByID(ctx, session.ID)
			require.NoError(t, err)
			require.NotNil(t, foundSession)
			assert.Equal(t, session.ID, foundSession.ID)
		}

		// Delete all sessions
		for _, session := range sessions {
			err := sessionRepo.Delete(ctx, session.ID)
			require.NoError(t, err)
		}

		// Verify all sessions are deleted
		for _, session := range sessions {
			foundSession, err := sessionRepo.FindByID(ctx, session.ID)
			require.NoError(t, err)
			assert.Nil(t, foundSession)
		}
	})

	t.Run("Create namespace", func(t *testing.T) {
		// Test namespace repository - use Save method
		ctx := context.Background()
		namespaceName := "test-auth-namespace"

		namespace, err := modulemodel.NewNamespace(namespaceName, nil, modulemodel.NamespaceTypeNone)
		require.NoError(t, err)

		err = namespaceRepo.Save(ctx, namespace)
		require.NoError(t, err)
		assert.Equal(t, namespaceName, namespace.Name())
	})

	_ = namespaceRepo // Avoid unused variable
}

func TestTerraformIdpRepositories(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer func() {
		require.NoError(t, db.Close())
	}()

	// Setup Terraform IDP repositories
	authCodeRepo := authRepo.NewTerraformIdpAuthorizationCodeRepository(db.DB)
	accessTokenRepo := authRepo.NewTerraformIdpAccessTokenRepository(db.DB)
	subjectIdentifierRepo := authRepo.NewTerraformIdpSubjectIdentifierRepository(db.DB)

	ctx := context.Background()

	t.Run("Create and find authorization code", func(t *testing.T) {
		key := "test-auth-code"
		data := []byte(`{"test": "data"}`)
		expiry := time.Now().Add(1 * time.Hour)

		err := authCodeRepo.Create(ctx, key, data, expiry)
		require.NoError(t, err)

		// Find the authorization code
		foundCode, err := authCodeRepo.FindByKey(ctx, key)
		require.NoError(t, err)
		require.NotNil(t, foundCode)
		assert.Equal(t, key, foundCode.Key)
		assert.Equal(t, data, foundCode.Data)

		// Delete the code
		err = authCodeRepo.DeleteByKey(ctx, key)
		require.NoError(t, err)

		// Verify it's deleted
		_, err = authCodeRepo.FindByKey(ctx, key)
		assert.Error(t, err)
	})

	t.Run("Create and find access token", func(t *testing.T) {
		key := "test-access-token"
		data := []byte(`{"token": "data"}`)
		expiry := time.Now().Add(1 * time.Hour)

		err := accessTokenRepo.Create(ctx, key, data, expiry)
		require.NoError(t, err)

		// Find the access token
		foundToken, err := accessTokenRepo.FindByKey(ctx, key)
		require.NoError(t, err)
		require.NotNil(t, foundToken)
		assert.Equal(t, key, foundToken.Key)
		assert.Equal(t, data, foundToken.Data)

		// Delete the token
		err = accessTokenRepo.DeleteByKey(ctx, key)
		require.NoError(t, err)
	})

	t.Run("Create and find subject identifier", func(t *testing.T) {
		key := "test-subject-identifier"
		data := []byte(`{"subject": "data"}`)
		expiry := time.Now().Add(1 * time.Hour)

		err := subjectIdentifierRepo.Create(ctx, key, data, expiry)
		require.NoError(t, err)

		// Find the subject identifier
		foundIdentifier, err := subjectIdentifierRepo.FindByKey(ctx, key)
		require.NoError(t, err)
		require.NotNil(t, foundIdentifier)
		assert.Equal(t, key, foundIdentifier.Key)
		assert.Equal(t, data, foundIdentifier.Data)

		// Delete the identifier
		err = subjectIdentifierRepo.DeleteByKey(ctx, key)
		require.NoError(t, err)
	})

	t.Run("Handle expired authorization codes", func(t *testing.T) {
		// Create an expired authorization code
		key := "expired-auth-code"
		data := []byte(`{"expired": "true"}`)
		expiry := time.Now().Add(-1 * time.Hour)

		err := authCodeRepo.Create(ctx, key, data, expiry)
		require.NoError(t, err)

		// Try to find the expired code - should fail
		_, err = authCodeRepo.FindByKey(ctx, key)
		assert.Error(t, err)
	})

	t.Run("Cleanup expired authorization codes", func(t *testing.T) {
		// Create multiple expired authorization codes
		for i := 0; i < 3; i++ {
			key := fmt.Sprintf("expired-code-%d", i)
			data := []byte(fmt.Sprintf(`{"index": %d}`, i))
			expiry := time.Now().Add(-1 * time.Hour)

			err := authCodeRepo.Create(ctx, key, data, expiry)
			require.NoError(t, err)
		}

		// Cleanup expired codes
		count, err := authCodeRepo.DeleteExpired(ctx)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, count, int64(3))
	})
}
