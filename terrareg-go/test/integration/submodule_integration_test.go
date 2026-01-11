package integration

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	modulemodel "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/module/model"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/shared"
	moduleRepo "github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/persistence/sqldb/module"
	"github.com/matthewjohn/terrareg/terrareg-go/test/integration/testutils"
)

func TestSubmoduleRepositoryIntegration(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer func() {
		require.NoError(t, db.Close())
	}()

	// Setup repositories with correct import paths
	namespaceRepo := moduleRepo.NewNamespaceRepository(db.DB)
	moduleProviderRepo := moduleRepo.NewModuleProviderRepository(db.DB, namespaceRepo, nil)
	moduleVersionRepo := moduleRepo.NewModuleVersionRepository(db.DB)
	exampleRepo := moduleRepo.NewExampleFileRepository(db.DB)
	submoduleRepo := moduleRepo.NewSubmoduleRepository(db.DB)
	moduleVersionFileRepo := moduleRepo.NewModuleVersionFileRepository(db.DB)

	ctx := context.Background()

	// Create namespace
	namespace, err := modulemodel.NewNamespace("submodule-repo-test", nil, modulemodel.NamespaceTypeNone)
	require.NoError(t, err)

	err = namespaceRepo.Save(ctx, namespace)
	require.NoError(t, err)

	t.Run("Create submodule model", func(t *testing.T) {
		// Create module provider
		moduleProvider, err := modulemodel.NewModuleProvider(namespace, "submodule-parent", "aws")
		require.NoError(t, err)

		err = moduleProviderRepo.Save(ctx, moduleProvider)
		require.NoError(t, err)

		// Create module version
		version, err := shared.ParseVersion("1.0.0")
		require.NoError(t, err)

		moduleVersion, err := modulemodel.NewModuleVersion(version.String(), nil, false)
		require.NoError(t, err)

		err = moduleProvider.AddVersion(moduleVersion)
		require.NoError(t, err)

		_, err = moduleVersionRepo.Save(ctx, moduleVersion)
		require.NoError(t, err)

		// Create submodules using the correct constructor
		name := stringPtr("examples")
		subType := stringPtr("terraform")

		submodule1 := modulemodel.NewSubmodule("examples1", name, subType, nil)
		assert.Equal(t, "examples1", submodule1.Path())
		assert.Equal(t, "examples", *submodule1.Name())
		assert.Equal(t, "terraform", *submodule1.Type())

		submodule2 := modulemodel.NewSubmodule("examples2", name, subType, nil)
		assert.Equal(t, "examples2", submodule2.Path())
	})

	t.Run("Create submodule with details", func(t *testing.T) {
		// Create module provider
		moduleProvider, err := modulemodel.NewModuleProvider(namespace, "submodule-with-details", "aws")
		require.NoError(t, err)

		err = moduleProviderRepo.Save(ctx, moduleProvider)
		require.NoError(t, err)

		// Create module version
		version, err := shared.ParseVersion("2.0.0")
		require.NoError(t, err)

		moduleVersion, err := modulemodel.NewModuleVersion(version.String(), nil, false)
		require.NoError(t, err)

		err = moduleProvider.AddVersion(moduleVersion)
		require.NoError(t, err)

		_, err = moduleVersionRepo.Save(ctx, moduleVersion)
		require.NoError(t, err)

		// Create submodule with details
		details := modulemodel.NewModuleDetails(nil)

		name := stringPtr("examples")
		subType := stringPtr("terraform")

		submodule := modulemodel.NewSubmodule("examples", name, subType, details)

		assert.Equal(t, "examples", submodule.Path())
		assert.NotNil(t, submodule.Details())
	})

	t.Run("Test repository creation", func(t *testing.T) {
		// Verify repositories are properly created
		assert.NotNil(t, namespaceRepo)
		assert.NotNil(t, moduleProviderRepo)
		assert.NotNil(t, moduleVersionRepo)
		assert.NotNil(t, exampleRepo)
		assert.NotNil(t, submoduleRepo)
		assert.NotNil(t, moduleVersionFileRepo)
	})

	_ = moduleVersionFileRepo // Avoid unused variable
}

