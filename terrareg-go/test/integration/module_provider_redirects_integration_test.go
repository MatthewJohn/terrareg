package integration

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	modulemodel "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/module/model"
	moduleRepo "github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/persistence/sqldb/module"
	"github.com/matthewjohn/terrareg/terrareg-go/test/integration/testutils"
)

func TestModuleProviderRedirectsIntegrationSimple(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer func() {
		require.NoError(t, db.Close())
	}()

	// Setup repositories with correct import paths
	namespaceRepo := moduleRepo.NewNamespaceRepository(db.DB)
	moduleProviderRepo := moduleRepo.NewModuleProviderRepository(db.DB, namespaceRepo, nil)
	redirectRepo := moduleRepo.NewModuleProviderRedirectRepository(db.DB)

	ctx := context.Background()

	// Create namespace
	namespace, err := modulemodel.NewNamespace("redirect-test", nil, modulemodel.NamespaceTypeNone)
	require.NoError(t, err)

	err = namespaceRepo.Save(ctx, namespace)
	require.NoError(t, err)

	t.Run("Create module providers for redirect", func(t *testing.T) {
		// Create source module provider (provider names must be lowercase alphanumeric only)
		sourceProvider, err := modulemodel.NewModuleProvider(namespace, "old-module", "aws")
		require.NoError(t, err)

		err = moduleProviderRepo.Save(ctx, sourceProvider)
		require.NoError(t, err)

		// Verify we can find the module provider after saving
		foundSource, err := moduleProviderRepo.FindByNamespaceModuleProvider(ctx, namespace.Name(), "old-module", "aws")
		require.NoError(t, err)
		require.NotNil(t, foundSource)

		// Create target module provider
		targetProvider, err := modulemodel.NewModuleProvider(namespace, "new-module", "gcp")
		require.NoError(t, err)

		err = moduleProviderRepo.Save(ctx, targetProvider)
		require.NoError(t, err)

		// Verify we can find the module provider after saving
		foundTarget, err := moduleProviderRepo.FindByNamespaceModuleProvider(ctx, namespace.Name(), "new-module", "gcp")
		require.NoError(t, err)
		require.NotNil(t, foundTarget)
	})

	t.Run("Test repository creation", func(t *testing.T) {
		// Verify repositories are properly created
		assert.NotNil(t, namespaceRepo)
		assert.NotNil(t, moduleProviderRepo)
		assert.NotNil(t, redirectRepo)
	})
}
