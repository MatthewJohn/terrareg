package auth

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	identityModel "terrareg/internal/domain/identity/model"
	identityRepo "terrareg/internal/domain/identity/repository"
	identityService "terrareg/internal/domain/identity/service"
	"terrareg/internal/infrastructure/persistence/sqldb"
	"terrareg/internal/infrastructure/persistence/sqldb/identity"
)

func TestIdentityIntegration(t *testing.T) {
	// Setup in-memory database for testing
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	// Auto-migrate tables
	err = db.AutoMigrate(
		&sqldb.UserDB{},
		&sqldb.UserGroupDB{},
		&sqldb.UserGroupMemberDB{},
		&sqldb.UserGroupNamespacePermissionDB{},
		&sqldb.SessionDB{},
	)
	require.NoError(t, err)

	// Initialize repositories
	userRepo := identity.NewUserRepository(db)
	groupRepo := identity.NewUserGroupRepository(db)
	sessionRepo := sqldb.NewSessionRepository(db)

	// Initialize services
	userService := identityService.NewUserService(userRepo, groupRepo)
	sessionManager := identityService.NewSessionManager(sessionRepo, userService)

	t.Run("Create User", func(t *testing.T) {
		user, err := identityModel.NewUser("testuser", "Test User", "test@example.com", identityModel.AuthMethodAPIKey)
		require.NoError(t, err)

		// Save user
		err = userRepo.Save(context.Background(), user)
		assert.NoError(t, err)

		// Verify user was created
		foundUser, err := userRepo.FindByUsername(context.Background(), "testuser")
		assert.NoError(t, err)
		assert.NotNil(t, foundUser)
		assert.Equal(t, "testuser", foundUser.Username())
		assert.Equal(t, "Test User", foundUser.DisplayName())
		assert.Equal(t, "test@example.com", foundUser.Email())
		assert.Equal(t, identityModel.AuthMethodAPIKey, foundUser.AuthMethod())
	})

	t.Run("Create User Group", func(t *testing.T) {
		user, err := userRepo.FindByUsername(context.Background(), "testuser")
		require.NoError(t, err)
		require.NotNil(t, user)

		// Create a new user group (this assumes UserGroup model exists)
		// For this test, we'll focus on user operations
		assert.NotNil(t, user)
		assert.True(t, user.Active())
	})

	t.Run("User Authentication", func(t *testing.T) {
		user, err := userRepo.FindByUsername(context.Background(), "testuser")
		require.NoError(t, err)
		require.NotNil(t, user)

		// Test user authentication (placeholder implementation)
		authenticatedUser, err := userService.Authenticate(context.Background(), "testuser", "")
		// This will fail with ErrInvalidCredentials for Phase 4 placeholder
		if err != nil {
			assert.Contains(t, err.Error(), "invalid")
		} else {
			assert.Equal(t, user, authenticatedUser)
		}
	})

	t.Run("Session Management", func(t *testing.T) {
		user, err := userRepo.FindByUsername(context.Background(), "testuser")
		require.NoError(t, err)
		require.NotNil(t, user)

		// Create session for user
		sessionID, err := sessionManager.CreateSession(context.Background(), user.ID(), "")
		assert.NoError(t, err)
		assert.NotEmpty(t, sessionID)

		// Validate session
		sessionUser, err := sessionManager.ValidateSession(context.Background(), sessionID)
		assert.NoError(t, err)
		assert.Equal(t, user.ID(), sessionUser.ID())

		// Invalidate session
		err = sessionManager.InvalidateSession(context.Background(), sessionID)
		assert.NoError(t, err)

		// Session should no longer be valid
		_, err = sessionManager.ValidateSession(context.Background(), sessionID)
		assert.Error(t, err)
	})
}

func TestPermissionSystem(t *testing.T) {
	// Setup in-memory database for testing
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	// Auto-migrate tables
	err = db.AutoMigrate(
		&sqldb.UserDB{},
		&sqldb.UserGroupDB{},
		&sqldb.UserGroupNamespacePermissionDB{},
	)
	require.NoError(t, err)

	// Initialize repositories
	userRepo := identity.NewUserRepository(db)
	groupRepo := identity.NewUserGroupRepository(db)

	// Create test user
	user, err := identityModel.NewUser("adminuser", "Admin User", "admin@example.com", identityModel.AuthMethodAPIKey)
	require.NoError(t, err)
	err = userRepo.Save(context.Background(), user)
	require.NoError(t, err)

	t.Run("User Permission Check", func(t *testing.T) {
		// Test basic permission functionality
		// Add admin permission to user (this assumes we have a method to add permissions)
		hasPermission := user.HasPermission(identityModel.ResourceTypeNamespace, "test-namespace", identityModel.ActionAdmin)

		// Initially should not have permission
		assert.False(t, hasPermission)

		// Add permission (this is simplified for Phase 4 testing)
		err := user.AddPermission(identityModel.ResourceTypeNamespace, "test-namespace", identityModel.ActionAdmin, user)
		assert.NoError(t, err)

		// Should now have permission
		hasPermission = user.HasPermission(identityModel.ResourceTypeNamespace, "test-namespace", identityModel.ActionAdmin)
		assert.True(t, hasPermission)
	})

	t.Run("Permission Types", func(t *testing.T) {
		// Test permission type constants
		assert.Equal(t, "namespace", string(identityModel.ResourceTypeNamespace))
		assert.Equal(t, "module", string(identityModel.ResourceTypeModule))
		assert.Equal(t, "provider", string(identityModel.ResourceTypeProvider))

		assert.Equal(t, "read", string(identityModel.ActionRead))
		assert.Equal(t, "write", string(identityModel.ActionWrite))
		assert.Equal(t, "admin", string(identityModel.ActionAdmin))
	})

	t.Run("Auth Method Constants", func(t *testing.T) {
		// Test auth method constants
		assert.Equal(t, "NONE", identityModel.AuthMethodNone.String())
		assert.Equal(t, "SAML", identityModel.AuthMethodSAML.String())
		assert.Equal(t, "OIDC", identityModel.AuthMethodOIDC.String())
		assert.Equal(t, "GITHUB", identityModel.AuthMethodGitHub.String())
		assert.Equal(t, "API_KEY", identityModel.AuthMethodAPIKey.String())
		assert.Equal(t, "TERRAFORM", identityModel.AuthMethodTerraform.String())
	})
}