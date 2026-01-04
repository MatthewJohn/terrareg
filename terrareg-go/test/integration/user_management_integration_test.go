package integration

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/auth"
	modulemodel "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/module/model"
	authRepo "github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/persistence/sqldb/auth"
	moduleRepo "github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/persistence/sqldb/module"
	"github.com/matthewjohn/terrareg/terrareg-go/test/integration/testutils"
)

func TestUserManagementIntegrationSimple(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer func() {
		require.NoError(t, db.Close())
	}()

	// Setup repositories with correct import paths
	sessionRepo := authRepo.NewSessionRepository(db.DB)
	userGroupRepo := authRepo.NewUserGroupRepository(db.DB)
	namespaceRepo := moduleRepo.NewNamespaceRepository(db.DB)

	ctx := context.Background()

	t.Run("Create user group", func(t *testing.T) {
		// Test user group repository - UserGroup is a simple struct
		userGroup := &auth.UserGroup{
			Name:      "test-user-group",
			SiteAdmin: false,
		}

		err := userGroupRepo.Save(ctx, userGroup)
		require.NoError(t, err)
		assert.Equal(t, "test-user-group", userGroup.Name)
		assert.False(t, userGroup.SiteAdmin)
		assert.Greater(t, userGroup.ID, 0)
	})

	t.Run("Create namespace for user group permissions", func(t *testing.T) {
		// Test namespace repository
		namespace, err := modulemodel.NewNamespace("user-permission-ns", nil, modulemodel.NamespaceTypeNone)
		require.NoError(t, err)

		err = namespaceRepo.Save(ctx, namespace)
		require.NoError(t, err)
		assert.Equal(t, "user-permission-ns", namespace.Name())
	})

	t.Run("Test repository creation", func(t *testing.T) {
		// Verify repositories are properly created
		assert.NotNil(t, sessionRepo)
		assert.NotNil(t, userGroupRepo)
		assert.NotNil(t, namespaceRepo)
	})
}
