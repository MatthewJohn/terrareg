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

func TestExampleIntegration(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer func() {
		require.NoError(t, db.Close())
	}()

	// Setup repositories with correct import paths
	namespaceRepo := moduleRepo.NewNamespaceRepository(db.DB)
	moduleProviderRepo := moduleRepo.NewModuleProviderRepository(db.DB, namespaceRepo, nil)
	moduleVersionRepo := moduleRepo.NewModuleVersionRepository(db.DB)
	exampleRepo := moduleRepo.NewExampleFileRepository(db.DB)
	moduleVersionFileRepo := moduleRepo.NewModuleVersionFileRepository(db.DB)
	submoduleRepo := moduleRepo.NewSubmoduleRepository(db.DB)

	ctx := context.Background()

	// Create namespace
	namespace, err := modulemodel.NewNamespace("testnamespace", nil, modulemodel.NamespaceTypeNone)
	require.NoError(t, err)

	err = namespaceRepo.Save(ctx, namespace)
	require.NoError(t, err)

	t.Run("Create module provider and version", func(t *testing.T) {
		// Create module provider - returns 2 values
		moduleProvider, err := modulemodel.NewModuleProvider(namespace, "testmodule", "aws")
		require.NoError(t, err)

		// Save module provider to get ID
		err = moduleProviderRepo.Save(ctx, moduleProvider)
		require.NoError(t, err)

		// Create module version
		version, err := shared.ParseVersion("1.0.0")
		require.NoError(t, err)

		moduleVersion, err := modulemodel.NewModuleVersion(version.String(), nil, false)
		require.NoError(t, err)

		// Add version to module provider
		err = moduleProvider.AddVersion(moduleVersion)
		require.NoError(t, err)

		// Save module version - returns 2 values
		_, err = moduleVersionRepo.Save(ctx, moduleVersion)
		require.NoError(t, err)

		assert.Equal(t, "1.0.0", moduleVersion.Version().String())
	})

	t.Run("Test repository creation", func(t *testing.T) {
		// Verify repositories are properly created
		assert.NotNil(t, namespaceRepo)
		assert.NotNil(t, moduleProviderRepo)
		assert.NotNil(t, moduleVersionRepo)
		assert.NotNil(t, exampleRepo)
		assert.NotNil(t, moduleVersionFileRepo)
		assert.NotNil(t, submoduleRepo)
	})

	_ = moduleVersionFileRepo // Avoid unused variable
}

func TestSubmoduleIntegrationSimple(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer func() {
		require.NoError(t, db.Close())
	}()

	// Setup repositories
	namespaceRepo := moduleRepo.NewNamespaceRepository(db.DB)
	moduleProviderRepo := moduleRepo.NewModuleProviderRepository(db.DB, namespaceRepo, nil)
	moduleVersionRepo := moduleRepo.NewModuleVersionRepository(db.DB)
	submoduleRepo := moduleRepo.NewSubmoduleRepository(db.DB)

	ctx := context.Background()

	// Create namespace
	namespace, err := modulemodel.NewNamespace("submodule-test", nil, modulemodel.NamespaceTypeNone)
	require.NoError(t, err)

	err = namespaceRepo.Save(ctx, namespace)
	require.NoError(t, err)

	t.Run("Create submodule model", func(t *testing.T) {
		// Create module provider
		moduleProvider, err := modulemodel.NewModuleProvider(namespace, "parent-module", "aws")
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

		// Create submodule using the correct constructor
		path := "examples"
		name := stringPtr("examples")
		subType := stringPtr("terraform")

		submodule := modulemodel.NewSubmodule(path, name, subType, nil)

		assert.Equal(t, "examples", submodule.Path())
		// Name() and Type() return *string, so dereference them
		assert.Equal(t, "examples", *submodule.Name())
		assert.Equal(t, "terraform", *submodule.Type())
	})

	_ = submoduleRepo
}

func stringPtr(s string) *string {
	return &s
}
