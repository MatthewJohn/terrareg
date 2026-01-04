package integration

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/auth"
	authRepo "github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/persistence/sqldb/auth"
	moduleRepo "github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/persistence/sqldb/module"
	"github.com/matthewjohn/terrareg/terrareg-go/test/integration/testutils"
)

func TestAuthenticationProvidersIntegration(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer func() {
		require.NoError(t, db.Close())
	}()

	// Setup repositories
	sessionRepo := authRepo.NewSessionRepository(db.DB)
	userGroupRepo := authRepo.NewUserGroupRepository(db.DB)
	namespaceRepo := moduleRepo.NewNamespaceRepository(db.DB)

	ctx := context.Background()
	_ = sessionRepo

	t.Run("Create user group", func(t *testing.T) {
		// Test user group repository - UserGroup is a simple struct
		userGroup := &auth.UserGroup{
			Name:      "test-group",
			SiteAdmin: false,
		}

		err := userGroupRepo.Save(ctx, userGroup)
		require.NoError(t, err)
		assert.Equal(t, "test-group", userGroup.Name)
		assert.False(t, userGroup.SiteAdmin)
		assert.Greater(t, userGroup.ID, 0)
	})

	t.Run("Create site admin user group", func(t *testing.T) {
		// Test creating a site admin user group
		userGroup := &auth.UserGroup{
			Name:      "admin-group",
			SiteAdmin: true,
		}

		err := userGroupRepo.Save(ctx, userGroup)
		require.NoError(t, err)
		assert.Equal(t, "admin-group", userGroup.Name)
		assert.True(t, userGroup.SiteAdmin)
	})

	t.Run("Find user group by name", func(t *testing.T) {
		// First create a user group
		userGroup := &auth.UserGroup{
			Name:      "find-test-group",
			SiteAdmin: false,
		}

		err := userGroupRepo.Save(ctx, userGroup)
		require.NoError(t, err)

		// Now find it
		foundGroup, err := userGroupRepo.FindByName(ctx, "find-test-group")
		require.NoError(t, err)
		require.NotNil(t, foundGroup)
		assert.Equal(t, "find-test-group", foundGroup.Name)
		assert.False(t, foundGroup.SiteAdmin)
	})

	t.Run("Find non-existent user group", func(t *testing.T) {
		// Try to find a user group that doesn't exist
		foundGroup, err := userGroupRepo.FindByName(ctx, "non-existent-group")
		require.NoError(t, err)
		assert.Nil(t, foundGroup)
	})

	t.Run("List user groups", func(t *testing.T) {
		// Create a couple of user groups
		group1 := &auth.UserGroup{
			Name:      "list-group-1",
			SiteAdmin: false,
		}
		err := userGroupRepo.Save(ctx, group1)
		require.NoError(t, err)

		group2 := &auth.UserGroup{
			Name:      "list-group-2",
			SiteAdmin: false,
		}
		err = userGroupRepo.Save(ctx, group2)
		require.NoError(t, err)

		// List all user groups (offset=0, limit=100)
		groups, err := userGroupRepo.List(ctx, 0, 100)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(groups), 2)
	})

	t.Run("Count user groups", func(t *testing.T) {
		// Get count of user groups
		count, err := userGroupRepo.Count(ctx)
		require.NoError(t, err)
		assert.Greater(t, count, int64(0))
	})

	t.Run("Update user group", func(t *testing.T) {
		// Create a user group
		userGroup := &auth.UserGroup{
			Name:      "update-test-group",
			SiteAdmin: false,
		}
		err := userGroupRepo.Save(ctx, userGroup)
		require.NoError(t, err)

		// Update it to be a site admin
		userGroup.SiteAdmin = true
		err = userGroupRepo.Update(ctx, userGroup)
		require.NoError(t, err)

		// Verify the update
		foundGroup, err := userGroupRepo.FindByName(ctx, "update-test-group")
		require.NoError(t, err)
		require.NotNil(t, foundGroup)
		assert.True(t, foundGroup.SiteAdmin)
	})

	t.Run("Delete user group", func(t *testing.T) {
		// Create a user group
		userGroup := &auth.UserGroup{
			Name:      "delete-test-group",
			SiteAdmin: false,
		}
		err := userGroupRepo.Save(ctx, userGroup)
		require.NoError(t, err)

		// Delete it
		err = userGroupRepo.Delete(ctx, userGroup.ID)
		require.NoError(t, err)

		// Verify it's deleted
		foundGroup, err := userGroupRepo.FindByName(ctx, "delete-test-group")
		require.NoError(t, err)
		assert.Nil(t, foundGroup)
	})

	t.Run("Find site admin groups", func(t *testing.T) {
		// Create a site admin group
		userGroup := &auth.UserGroup{
			Name:      "site-admin-test",
			SiteAdmin: true,
		}
		err := userGroupRepo.Save(ctx, userGroup)
		require.NoError(t, err)

		// Find all site admin groups
		adminGroups, err := userGroupRepo.FindSiteAdminGroups(ctx)
		require.NoError(t, err)
		assert.Greater(t, len(adminGroups), 0)
	})

	_ = namespaceRepo // Avoid unused variable
}
