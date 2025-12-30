package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/persistence/sqldb"
	testutils "github.com/matthewjohn/terrareg/terrareg-go/test/integration/testutils"
)

// TestGitProvider_Create tests creating git providers
func TestGitProvider_Create(t *testing.T) {
	t.Run("single provider", func(t *testing.T) {
		db := testutils.SetupTestDatabase(t)
		defer testutils.CleanupTestDatabase(t, db)

		// Clean git_provider table
		db.DB.Exec("DELETE FROM git_provider")

		gitProvider := sqldb.GitProviderDB{
			Name:              "Test One",
			BaseURLTemplate:   "https://example.com/{namespace}/{module}",
			CloneURLTemplate:  "ssh://git@example.com/{namespace}/{module}.git",
			BrowseURLTemplate: "https://example.com/{namespace}/{module}/tree/{tag}/{path}",
			GitPathTemplate:   "",
		}

		err := db.DB.Create(&gitProvider).Error
		require.NoError(t, err)

		assert.NotZero(t, gitProvider.ID)
		assert.Equal(t, "Test One", gitProvider.Name)
		assert.Equal(t, "https://example.com/{namespace}/{module}", gitProvider.BaseURLTemplate)
		assert.Equal(t, "ssh://git@example.com/{namespace}/{module}.git", gitProvider.CloneURLTemplate)
		assert.Equal(t, "https://example.com/{namespace}/{module}/tree/{tag}/{path}", gitProvider.BrowseURLTemplate)
		assert.Equal(t, "", gitProvider.GitPathTemplate)
	})

	t.Run("provider with git path", func(t *testing.T) {
		db := testutils.SetupTestDatabase(t)
		defer testutils.CleanupTestDatabase(t, db)

		// Clean git_provider table
		db.DB.Exec("DELETE FROM git_provider")

		gitPath := "/{module}/{provider}"
		gitProvider := sqldb.GitProviderDB{
			Name:              "Test Two",
			BaseURLTemplate:   "https://example.com/{namespace}/modules",
			CloneURLTemplate:  "ssh://git@example.com/{namespace}/modules.git",
			BrowseURLTemplate: "https://example.com/{namespace}/modules/tree/{tag}/{path}",
			GitPathTemplate:   gitPath,
		}

		err := db.DB.Create(&gitProvider).Error
		require.NoError(t, err)

		assert.NotZero(t, gitProvider.ID)
		assert.Equal(t, "Test Two", gitProvider.Name)
		assert.Equal(t, gitPath, gitProvider.GitPathTemplate)
	})

	t.Run("multiple providers", func(t *testing.T) {
		db := testutils.SetupTestDatabase(t)
		defer testutils.CleanupTestDatabase(t, db)

		// Clean git_provider table
		db.DB.Exec("DELETE FROM git_provider")

		providers := []sqldb.GitProviderDB{
			{
				Name:              "Test One",
				BaseURLTemplate:   "https://example.com/{namespace}/{module}",
				CloneURLTemplate:  "ssh://git@example.com/{namespace}/{module}.git",
				BrowseURLTemplate: "https://example.com/{namespace}/{module}/tree/{tag}/{path}",
				GitPathTemplate:   "",
			},
			{
				Name:              "Test Two",
				BaseURLTemplate:   "https://example.com/{namespace}/modules",
				CloneURLTemplate:  "ssh://git@example.com/{namespace}/modules.git",
				BrowseURLTemplate: "https://example.com/{namespace}/modules/tree/{tag}/{path}",
				GitPathTemplate:   "/{module}/{provider}",
			},
		}

		for i := range providers {
			err := db.DB.Create(&providers[i]).Error
			require.NoError(t, err)
		}

		// Query all providers
		var allProviders []sqldb.GitProviderDB
		err := db.DB.Order("name ASC").Find(&allProviders).Error
		require.NoError(t, err)

		assert.Len(t, allProviders, 2)

		// Check first provider
		assert.Equal(t, "Test One", allProviders[0].Name)
		assert.Equal(t, "https://example.com/{namespace}/{module}", allProviders[0].BaseURLTemplate)
		assert.Equal(t, "ssh://git@example.com/{namespace}/{module}.git", allProviders[0].CloneURLTemplate)
		assert.Equal(t, "https://example.com/{namespace}/{module}/tree/{tag}/{path}", allProviders[0].BrowseURLTemplate)
		assert.Equal(t, "", allProviders[0].GitPathTemplate)

		// Check second
		assert.Equal(t, "Test Two", allProviders[1].Name)
		assert.Equal(t, "https://example.com/{namespace}/modules", allProviders[1].BaseURLTemplate)
		assert.Equal(t, "ssh://git@example.com/{namespace}/modules.git", allProviders[1].CloneURLTemplate)
		assert.Equal(t, "https://example.com/{namespace}/modules/tree/{tag}/{path}", allProviders[1].BrowseURLTemplate)
		assert.Equal(t, "/{module}/{provider}", allProviders[1].GitPathTemplate)
	})
}

// TestGitProvider_GetAll tests retrieving all git providers
func TestGitProvider_GetAll(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	// Clean git_provider table
	db.DB.Exec("DELETE FROM git_provider")

	// Create test providers
	providers := []sqldb.GitProviderDB{
		{
			Name:              "Provider A",
			BaseURLTemplate:   "https://a.com/{namespace}/{module}",
			CloneURLTemplate:  "ssh://git@a.com/{namespace}/{module}.git",
			BrowseURLTemplate: "https://a.com/{namespace}/{module}/tree/{tag}/{path}",
		},
		{
			Name:              "Provider B",
			BaseURLTemplate:   "https://b.com/{namespace}/{module}",
			CloneURLTemplate:  "ssh://git@b.com/{namespace}/{module}.git",
			BrowseURLTemplate: "https://b.com/{namespace}/{module}/tree/{tag}/{path}",
		},
	}

	for i := range providers {
		err := db.DB.Create(&providers[i]).Error
		require.NoError(t, err)
	}

	// Query all providers
	var allProviders []sqldb.GitProviderDB
	err := db.DB.Order("name ASC").Find(&allProviders).Error
	require.NoError(t, err)

	assert.Len(t, allProviders, 2)
	assert.Equal(t, "Provider A", allProviders[0].Name)
	assert.Equal(t, "Provider B", allProviders[1].Name)
}

// TestGitProvider_UniqueName tests that provider names must be unique
func TestGitProvider_UniqueName(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	// Clean git_provider table
	db.DB.Exec("DELETE FROM git_provider")

	// Create first provider
	provider1 := sqldb.GitProviderDB{
		Name:              "Duplicate",
		BaseURLTemplate:   "https://example.com/{namespace}/{module}",
		CloneURLTemplate:  "ssh://git@example.com/{namespace}/{module}.git",
		BrowseURLTemplate: "https://example.com/{namespace}/{module}/tree/{tag}/{path}",
	}

	err := db.DB.Create(&provider1).Error
	require.NoError(t, err)

	// Try to create duplicate - should fail due to unique index
	provider2 := sqldb.GitProviderDB{
		Name:              "Duplicate",
		BaseURLTemplate:   "https://other.com/{namespace}/{module}",
		CloneURLTemplate:  "ssh://git@other.com/{namespace}/{module}.git",
		BrowseURLTemplate: "https://other.com/{namespace}/{module}/tree/{tag}/{path}",
	}

	err = db.DB.Create(&provider2).Error
	assert.Error(t, err, "Expected error for duplicate git provider name")
}

// TestGitProvider_Delete tests deleting git providers
func TestGitProvider_Delete(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	// Clean git_provider table
	db.DB.Exec("DELETE FROM git_provider")

	// Create provider
	provider := sqldb.GitProviderDB{
		Name:              "To Delete",
		BaseURLTemplate:   "https://example.com/{namespace}/{module}",
		CloneURLTemplate:  "ssh://git@example.com/{namespace}/{module}.git",
		BrowseURLTemplate: "https://example.com/{namespace}/{module}/tree/{tag}/{path}",
	}

	err := db.DB.Create(&provider).Error
	require.NoError(t, err)
	assert.NotZero(t, provider.ID)

	// Delete provider
	err = db.DB.Delete(&provider).Error
	require.NoError(t, err)

	// Verify deleted
	var count int64
	err = db.DB.Model(&sqldb.GitProviderDB{}).Count(&count).Error
	require.NoError(t, err)
	assert.Equal(t, int64(0), count)
}

// TestGitProvider_Update tests updating git providers
func TestGitProvider_Update(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	// Clean git_provider table
	db.DB.Exec("DELETE FROM git_provider")

	// Create provider
	provider := sqldb.GitProviderDB{
		Name:              "Original Name",
		BaseURLTemplate:   "https://example.com/{namespace}/{module}",
		CloneURLTemplate:  "ssh://git@example.com/{namespace}/{module}.git",
		BrowseURLTemplate: "https://example.com/{namespace}/{module}/tree/{tag}/{path}",
	}

	err := db.DB.Create(&provider).Error
	require.NoError(t, err)

	// Update provider
	provider.Name = "Updated Name"
	provider.BaseURLTemplate = "https://updated.com/{namespace}/{module}"
	err = db.DB.Save(&provider).Error
	require.NoError(t, err)

	// Verify updated
	var updated sqldb.GitProviderDB
	err = db.DB.First(&updated, provider.ID).Error
	require.NoError(t, err)

	assert.Equal(t, "Updated Name", updated.Name)
	assert.Equal(t, "https://updated.com/{namespace}/{module}", updated.BaseURLTemplate)
}
