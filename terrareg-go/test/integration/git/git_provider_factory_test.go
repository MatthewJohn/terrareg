package git

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/git/service"
	gitrepo "github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/persistence/sqldb/git"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/persistence/sqldb"
	testutils "github.com/matthewjohn/terrareg/terrareg-go/test/integration/testutils"
)

// setupTestFactory creates a test factory with database cleanup
func setupTestFactory(t *testing.T) (*service.GitProviderFactory, *sqldb.Database) {
	db := testutils.SetupTestDatabase(t)

	// Clean git_provider table
	db.DB.Exec("DELETE FROM git_provider")

	// Create repository
	repo := gitrepo.NewGitProviderRepository(db)

	// Create factory
	factory := service.NewGitProviderFactory(repo)

	return factory, db
}

// TestInitialiseFromConfig tests basic config loading
// Python reference: test_git_provider.py::test_initialise_from_config
func TestGitProviderFactory_InitialiseFromConfig(t *testing.T) {
	factory, testDB := setupTestFactory(t)
	defer testutils.CleanupTestDatabase(t, testDB)

	ctx := context.Background()

	configJSON := `[{
		"name": "Test One",
		"base_url": "https://example.com/{namespace}/{module}",
		"clone_url": "ssh://git@example.com/{namespace}/{module}.git",
		"browse_url": "https://example.com/{namespace}/{module}/tree/{tag}/{path}"
	}, {
		"name": "Test Two",
		"base_url": "https://example.com/{namespace}/modules",
		"clone_url": "ssh://git@example.com/{namespace}/modules.git",
		"browse_url": "https://example.com/{namespace}/modules/tree/{tag}/{path}",
		"git_path": "/{module}/{provider}"
	}]`

	err := factory.InitialiseFromConfig(ctx, configJSON)
	require.NoError(t, err)

	// Verify providers were created
	providers, err := factory.GetAll(ctx)
	require.NoError(t, err)
	assert.Len(t, providers, 2)

	// Sort by name for consistent ordering
	if providers[0].Name > providers[1].Name {
		providers[0], providers[1] = providers[1], providers[0]
	}

	// Check first provider
	assert.Equal(t, "Test One", providers[0].Name)
	assert.Equal(t, "https://example.com/{namespace}/{module}", providers[0].BaseURLTemplate)
	assert.Equal(t, "ssh://git@example.com/{namespace}/{module}.git", providers[0].CloneURLTemplate)
	assert.Equal(t, "https://example.com/{namespace}/{module}/tree/{tag}/{path}", providers[0].BrowseURLTemplate)
	assert.Equal(t, "", providers[0].GitPathTemplate)

	// Check second provider
	assert.Equal(t, "Test Two", providers[1].Name)
	assert.Equal(t, "https://example.com/{namespace}/modules", providers[1].BaseURLTemplate)
	assert.Equal(t, "ssh://git@example.com/{namespace}/modules.git", providers[1].CloneURLTemplate)
	assert.Equal(t, "https://example.com/{namespace}/modules/tree/{tag}/{path}", providers[1].BrowseURLTemplate)
	assert.Equal(t, "/{module}/{provider}", providers[1].GitPathTemplate)
}

// TestInitialiseFromConfig_MissingPlaceholders tests placeholder validation
// Python reference: test_git_provider.py::test_initialise_from_config_missing_placeholders
func TestGitProviderFactory_MissingPlaceholders(t *testing.T) {
	tests := []struct {
		name        string
		urlSuffix   string
		gitPath     string
		expectError string
	}{
		{
			name:      "with provider",
			urlSuffix: "{namespace}/{module}",
			gitPath:   "",
		},
		{
			name:      "with provider and provider",
			urlSuffix: "{namespace}/{module}/{provider}",
			gitPath:   "",
		},
		{
			name:      "provider in git path",
			urlSuffix: "{namespace}/{module}",
			gitPath:   "{provider}",
		},
		{
			name:      "without module - module in git path",
			urlSuffix: "{namespace}",
			gitPath:   "{module}",
		},
		{
			name:      "without namespace - namespace in git path",
			urlSuffix: "{module}",
			gitPath:   "{namespace}",
		},
		{
			name:      "both namespace and module in git path",
			urlSuffix: "blah",
			gitPath:   "{namespace}-{module}",
		},
		{
			name:        "missing namespace and module",
			urlSuffix:   "blah",
			gitPath:     "",
			expectError: "Namespace placeholder not present in URL",
		},
		{
			name:        "invalid placeholder in url",
			urlSuffix:   "{namespace}-{module}-{somethingelse}",
			gitPath:     "",
			expectError: "Template contains unknown placeholder",
		},
		{
			name:        "invalid placeholder in git path",
			urlSuffix:   "{namespace}-{module}",
			gitPath:     "{somethingelse}",
			expectError: "Template contains unknown placeholder",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			factory, testDB := setupTestFactory(t)
			defer testutils.CleanupTestDatabase(t, testDB)

			ctx := context.Background()

			configJSON := fmt.Sprintf(`[{
				"name": "Test One",
				"base_url": "https://example.com/%s",
				"clone_url": "ssh://git@example.com/%s",
				"browse_url": "https://example.com/%s/tree/{tag}/{path}",
				"git_path": "%s"
			}]`, tt.urlSuffix, tt.urlSuffix, tt.urlSuffix, tt.gitPath)

			err := factory.InitialiseFromConfig(ctx, configJSON)

			if tt.expectError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectError)
			} else {
				require.NoError(t, err)

				// Verify provider was created
				providers, err := factory.GetAll(ctx)
				require.NoError(t, err)
				assert.Len(t, providers, 1)
				assert.Equal(t, "Test One", providers[0].Name)
			}
		})
	}
}

// TestGetByName tests getting a git provider by name
func TestGitProviderFactory_GetByName(t *testing.T) {
	factory, testDB := setupTestFactory(t)
	defer testutils.CleanupTestDatabase(t, testDB)

	ctx := context.Background()

	// Create a test provider
	configJSON := `[{
		"name": "Test Provider",
		"base_url": "https://example.com/{namespace}/{module}",
		"clone_url": "ssh://git@example.com/{namespace}/{module}.git",
		"browse_url": "https://example.com/{namespace}/{module}/tree/{tag}/{path}"
	}]`

	err := factory.InitialiseFromConfig(ctx, configJSON)
	require.NoError(t, err)

	// Test getting by name
	provider, err := factory.GetByName(ctx, "Test Provider")
	require.NoError(t, err)
	require.NotNil(t, provider)

	assert.Equal(t, "Test Provider", provider.Name)
	assert.Equal(t, "https://example.com/{namespace}/{module}", provider.BaseURLTemplate)

	// Test non-existent provider
	provider, err = factory.GetByName(ctx, "Non Existent")
	require.NoError(t, err)
	assert.Nil(t, provider)
}

// TestGetAll tests getting all git providers
func TestGitProviderFactory_GetAll(t *testing.T) {
	factory, testDB := setupTestFactory(t)
	defer testutils.CleanupTestDatabase(t, testDB)

	ctx := context.Background()

	// Initially should be empty
	providers, err := factory.GetAll(ctx)
	require.NoError(t, err)
	assert.Len(t, providers, 0)

	// Create two test providers
	configJSON := `[{
		"name": "Provider A",
		"base_url": "https://a.com/{namespace}/{module}",
		"clone_url": "ssh://git@a.com/{namespace}/{module}.git",
		"browse_url": "https://a.com/{namespace}/{module}/tree/{tag}/{path}"
	}, {
		"name": "Provider B",
		"base_url": "https://b.com/{namespace}/{module}",
		"clone_url": "ssh://git@b.com/{namespace}/{module}.git",
		"browse_url": "https://b.com/{namespace}/{module}/tree/{tag}/{path}"
	}]`

	err = factory.InitialiseFromConfig(ctx, configJSON)
	require.NoError(t, err)

	// Get all providers - should be sorted by name
	providers, err = factory.GetAll(ctx)
	require.NoError(t, err)
	assert.Len(t, providers, 2)
	assert.Equal(t, "Provider A", providers[0].Name)
	assert.Equal(t, "Provider B", providers[1].Name)
}

// TestInitialiseFromConfig_UpdateExisting tests upsert functionality
func TestGitProviderFactory_UpdateExisting(t *testing.T) {
	factory, testDB := setupTestFactory(t)
	defer testutils.CleanupTestDatabase(t, testDB)

	ctx := context.Background()

	// Create initial provider
	configJSON := `[{
		"name": "Test Provider",
		"base_url": "https://old.com/{namespace}/{module}",
		"clone_url": "ssh://git@old.com/{namespace}/{module}.git",
		"browse_url": "https://old.com/{namespace}/{module}/tree/{tag}/{path}"
	}]`

	err := factory.InitialiseFromConfig(ctx, configJSON)
	require.NoError(t, err)

	// Verify initial values
	provider, err := factory.GetByName(ctx, "Test Provider")
	require.NoError(t, err)
	assert.Equal(t, "https://old.com/{namespace}/{module}", provider.BaseURLTemplate)

	// Update with new config
	configJSON = `[{
		"name": "Test Provider",
		"base_url": "https://new.com/{namespace}/{module}",
		"clone_url": "ssh://git@new.com/{namespace}/{module}.git",
		"browse_url": "https://new.com/{namespace}/{module}/tree/{tag}/{path}"
	}]`

	err = factory.InitialiseFromConfig(ctx, configJSON)
	require.NoError(t, err)

	// Verify updated values
	provider, err = factory.GetByName(ctx, "Test Provider")
	require.NoError(t, err)
	assert.Equal(t, "https://new.com/{namespace}/{module}", provider.BaseURLTemplate)

	// Verify only one provider exists
	providers, err := factory.GetAll(ctx)
	require.NoError(t, err)
	assert.Len(t, providers, 1)
}

// TestInitialiseFromConfig_InvalidConfig tests invalid configurations
func TestGitProviderFactory_InvalidConfig(t *testing.T) {
	tests := []struct {
		name           string
		configJSON     string
		expectedError  string
	}{
		{
			name: "missing name",
			configJSON: `[{
				"base_url": "https://example.com/{namespace}/{module}",
				"clone_url": "ssh://git@example.com/{namespace}/{module}.git",
				"browse_url": "https://example.com/{namespace}/{module}/tree/{tag}/{path}"
			}]`,
			expectedError: "Git provider config does not contain required attribute: name",
		},
		{
			name: "missing base_url",
			configJSON: `[{
				"name": "Test",
				"clone_url": "ssh://git@example.com/{namespace}/{module}.git",
				"browse_url": "https://example.com/{namespace}/{module}/tree/{tag}/{path}"
			}]`,
			expectedError: "Git provider config does not contain required attribute: base_url",
		},
		{
			name: "missing clone_url",
			configJSON: `[{
				"name": "Test",
				"base_url": "https://example.com/{namespace}/{module}",
				"browse_url": "https://example.com/{namespace}/{module}/tree/{tag}/{path}"
			}]`,
			expectedError: "Git provider config does not contain required attribute: clone_url",
		},
		{
			name: "missing browse_url",
			configJSON: `[{
				"name": "Test",
				"base_url": "https://example.com/{namespace}/{module}",
				"clone_url": "ssh://git@example.com/{namespace}/{module}.git"
			}]`,
			expectedError: "Git provider config does not contain required attribute: browse_url",
		},
		{
			name:           "invalid JSON",
			configJSON:     `{invalid JSON}`,
			expectedError:  "not valid JSON",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			factory, testDB := setupTestFactory(t)
			defer testutils.CleanupTestDatabase(t, testDB)

			ctx := context.Background()

			err := factory.InitialiseFromConfig(ctx, tt.configJSON)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.expectedError)
		})
	}
}
