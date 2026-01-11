package provider_source

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/provider_source/model"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/provider_source/service"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/persistence/sqldb"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/persistence/sqldb/provider_source"
	testutils "github.com/matthewjohn/terrareg/terrareg-go/test/integration/testutils"
)

// setupTestFactory creates a test factory with database cleanup
func setupTestFactory(t *testing.T) (*service.ProviderSourceFactory, *sqldb.Database) {
	db := testutils.SetupTestDatabase(t)

	// Clean provider_source table
	db.DB.Exec("DELETE FROM provider_source")

	// Create repository
	repo := provider_source.NewProviderSourceRepository(db.DB)

	// Create factory
	factory := service.NewProviderSourceFactory(repo)

	// Register GitHub provider source class
	githubClass := service.NewGithubProviderSourceClass()
	factory.RegisterProviderSourceClass(githubClass)

	return factory, db
}

// TestGetProviderClasses verifies the class mapping
// Python reference: test_provider_source_factory.py::test_get_provider_classes
func TestGetProviderClasses(t *testing.T) {
	factory, db := setupTestFactory(t)
	defer testutils.CleanupTestDatabase(t, db)

	mappings := factory.GetProviderClasses()

	// Should only have GitHub registered
	assert.Len(t, mappings, 1)
	assert.Contains(t, mappings, model.ProviderSourceTypeGithub)

	// Verify the class type
	githubClass := mappings[model.ProviderSourceTypeGithub]
	assert.Equal(t, model.ProviderSourceTypeGithub, githubClass.Type())
}

// TestGetProviderSourceClassByType tests getting provider source class by type
// Python reference: test_provider_source_factory.py::test_get_provider_source_class_by_type
func TestGetProviderSourceClassByType(t *testing.T) {
	tests := []struct {
		name         string
		type_        model.ProviderSourceType
		expectNil    bool
	}{
		{
			name:      "github type",
			type_:     model.ProviderSourceTypeGithub,
			expectNil: false,
		},
		{
			name:      "invalid type",
			type_:     model.ProviderSourceType("does_not_exist"),
			expectNil: true,
		},
		{
			name:      "empty type",
			type_:     model.ProviderSourceType(""),
			expectNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			factory, db := setupTestFactory(t)
			defer testutils.CleanupTestDatabase(t, db)

			result := factory.GetProviderSourceClassByType(tt.type_)
			if tt.expectNil {
				assert.Nil(t, result)
			} else {
				assert.NotNil(t, result)
				assert.Equal(t, tt.type_, result.Type())
			}
		})
	}
}

// TestGetProviderSourceByName tests retrieving provider source by name
// Python reference: test_provider_source_factory.py::test_get_provider_source_by_name
func TestGetProviderSourceByName(t *testing.T) {
	tests := []struct {
		name           string
		providerName   string
		callName       string
		expectFound    bool
	}{
		{
			name:         "exists",
			providerName: "test-provider-source",
			callName:     "test-provider-source",
			expectFound:  true,
		},
		{
			name:         "does not exist",
			providerName: "test-provider-source",
			callName:     "does-not-exist",
			expectFound:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			factory, testDB := setupTestFactory(t)
			defer testutils.CleanupTestDatabase(t, testDB)

			ctx := context.Background()

			// Create test provider
			name := tt.providerName
			apiName := "ut-name"
			config := &model.ProviderSourceConfig{
				BaseURL:         "https://github.com",
				ApiURL:          "https://api.github.com",
				ClientID:        "test-client-id",
				ClientSecret:    "test-client-secret",
				LoginButtonText: "Sign in with GitHub",
				PrivateKeyPath:  "/path/to/key",
				AppID:           "123456",
			}

			source := model.NewProviderSource(name, apiName, model.ProviderSourceTypeGithub, config)
			require.NoError(t, testDB.DB.Create(source.ToDBModel()).Error)

			// Test retrieval
			result, err := factory.GetProviderSourceByName(ctx, tt.callName)
			require.NoError(t, err)

			if tt.expectFound {
				require.NotNil(t, result)
				assert.Equal(t, tt.providerName, result.Name())
			} else {
				assert.Nil(t, result)
			}
		})
	}
}

// TestGetProviderSourceByApiName tests retrieving provider source by api_name
// Python reference: test_provider_source_factory.py::test_get_provider_source_by_api_name
func TestGetProviderSourceByApiName(t *testing.T) {
	tests := []struct {
		name              string
		providerApiName   string
		callApiName       string
		expectFound       bool
	}{
		{
			name:            "exists",
			providerApiName: "test-provider-source",
			callApiName:     "test-provider-source",
			expectFound:     true,
		},
		{
			name:            "does not exist",
			providerApiName: "test-provider-source",
			callApiName:     "does-not-exist",
			expectFound:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			factory, testDB := setupTestFactory(t)
			defer testutils.CleanupTestDatabase(t, testDB)

			ctx := context.Background()

			// Create test provider
			name := "Provider Name"
			apiName := tt.providerApiName
			config := &model.ProviderSourceConfig{
				BaseURL:         "https://github.com",
				ApiURL:          "https://api.github.com",
				ClientID:        "test-client-id",
				ClientSecret:    "test-client-secret",
				LoginButtonText: "Sign in with GitHub",
				PrivateKeyPath:  "/path/to/key",
				AppID:           "123456",
			}

			source := model.NewProviderSource(name, apiName, model.ProviderSourceTypeGithub, config)
			require.NoError(t, testDB.DB.Create(source.ToDBModel()).Error)

			// Test retrieval
			result, err := factory.GetProviderSourceByApiName(ctx, tt.callApiName)
			require.NoError(t, err)

			if tt.expectFound {
				require.NotNil(t, result)
				assert.Equal(t, tt.providerApiName, result.ApiName())
			} else {
				assert.Nil(t, result)
			}
		})
	}
}

// TestGetAllProviderSources tests retrieving all provider sources
// Python reference: test_provider_source_factory.py::test_get_all_provider_sources
func TestGetAllProviderSources(t *testing.T) {
	factory, testDB := setupTestFactory(t)
	defer testutils.CleanupTestDatabase(t, testDB)

	ctx := context.Background()

	// Create two test providers
	providers := []struct {
		name    string
		apiName string
	}{
		{"Test Provider 1", "prov-1"},
		{"Test Provider 2", "prov-2"},
	}

	for _, p := range providers {
		config := &model.ProviderSourceConfig{
			BaseURL:         "https://github.com",
			ApiURL:          "https://api.github.com",
			ClientID:        "test-client-id",
			ClientSecret:    "test-client-secret",
			LoginButtonText: "Sign in with GitHub",
			PrivateKeyPath:  "/path/to/key",
			AppID:           "123456",
		}
		source := model.NewProviderSource(p.name, p.apiName, model.ProviderSourceTypeGithub, config)
		require.NoError(t, testDB.DB.Create(source.ToDBModel()).Error)
	}

	// Get all provider sources
	result, err := factory.GetAllProviderSources(ctx)
	require.NoError(t, err)

	assert.Len(t, result, 2)

	names := make([]string, len(result))
	apiNames := make([]string, len(result))
	for i, ps := range result {
		names[i] = ps.Name()
		apiNames[i] = ps.ApiName()
	}

	assert.ElementsMatch(t, []string{"Test Provider 1", "Test Provider 2"}, names)
	assert.ElementsMatch(t, []string{"prov-1", "prov-2"}, apiNames)
}

// TestNameToApiName tests name to API name conversion
// Python reference: test_provider_source_factory.py::test__name_to_api_name
func TestNameToApiName(t *testing.T) {
	factory, db := setupTestFactory(t)
	defer testutils.CleanupTestDatabase(t, db)

	tests := []struct {
		name           string
		input          string
		expectedOutput string
	}{
		// Empty values
		{
			name:           "empty string",
			input:          "",
			expectedOutput: "",
		},
		// Unchanged
		{
			name:           "simple name",
			input:          "test-name",
			expectedOutput: "test-name",
		},
		// Lower case
		{
			name:           "mixed case",
			input:          "TestName",
			expectedOutput: "testname",
		},
		// Space replacement
		{
			name:           "spaces",
			input:          "test name",
			expectedOutput: "test-name",
		},
		// Special chars
		{
			name:           "special chars",
			input:          "test@name",
			expectedOutput: "testname",
		},
		// Numbers
		{
			name:           "numbers",
			input:          "testname1234",
			expectedOutput: "testname1234",
		},
		// Leading/trailing space
		{
			name:           "leading space",
			input:          " test",
			expectedOutput: "test",
		},
		{
			name:           "trailing space",
			input:          "test ",
			expectedOutput: "test",
		},
		// Leading/trailing dash
		{
			name:           "leading dash",
			input:          "-test",
			expectedOutput: "test",
		},
		{
			name:           "trailing dash",
			input:          "test-",
			expectedOutput: "test",
		},
		// Combined test
		{
			name:           "complex",
			input:          " -This 15 a Sp3c!@L Test CASE+=  ",
			expectedOutput: "this-15-a-sp3cl-test-case",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := factory.NameToApiName(tt.input)
			assert.Equal(t, tt.expectedOutput, result)
		})
	}
}

// TestInitialiseFromConfig_Create tests creating a new provider from config
// Python reference: test_provider_source_factory.py::test_initialise_from_config_create
func TestInitialiseFromConfig_Create(t *testing.T) {
	factory, testDB := setupTestFactory(t)
	defer testutils.CleanupTestDatabase(t, testDB)

	ctx := context.Background()

	configJSON := `[{
		"name": "Test Create 1",
		"type": "github",
		"base_url": "https://github.com",
		"api_url": "https://api.github.com",
		"client_id": "test-client-id",
		"client_secret": "test-client-secret",
		"login_button_text": "Sign in with GitHub",
		"private_key_path": "/path/to/key",
		"app_id": "123456"
	}]`

	err := factory.InitialiseFromConfig(ctx, configJSON)
	require.NoError(t, err)

	// Verify provider was created
	result, err := factory.GetProviderSourceByName(ctx, "Test Create 1")
	require.NoError(t, err)
	require.NotNil(t, result)

	assert.Equal(t, "Test Create 1", result.Name())
	assert.Equal(t, "test-create-1", result.ApiName())
	assert.Equal(t, model.ProviderSourceTypeGithub, result.Type())

	// Verify in database
	var dbModel sqldb.ProviderSourceDB
	err = testDB.DB.Where("name = ?", "Test Create 1").First(&dbModel).Error
	require.NoError(t, err)
	assert.Equal(t, "test-create-1", *dbModel.APIName)
}

// TestInitialiseFromConfig_InvalidConfig tests invalid configurations
// Python reference: test_provider_source_factory.py::test_initialise_from_config_invalid_config
func TestInitialiseFromConfig_InvalidConfig(t *testing.T) {
	tests := []struct {
		name           string
		configJSON     string
		expectedError  string
	}{
		{
			name: "invalid type",
			configJSON: `[{
				"name": "Test Provider 1",
				"type": "invalid Type"
			}]`,
			expectedError: "Invalid provider source type. Valid types: github",
		},
		{
			name: "no type",
			configJSON: `[{
				"name": "Test Provider 1"
			}]`,
			expectedError: "Provider source config does not contain required attribute: type",
		},
		{
			name: "missing config - base_url",
			configJSON: `[{
				"name": "Test Provider 1",
				"type": "github"
			}]`,
			expectedError: "Missing required Github provider source config: base_url",
		},
		{
			name: "invalid name",
			configJSON: `[{
				"name": "  --  ",
				"type": "github"
			}]`,
			expectedError: "Invalid provider source config: Name must contain some alphanumeric characters",
		},
		{
			name: "missing name",
			configJSON: `[{
				"type": "github"
			}]`,
			expectedError: "Provider source config does not contain required attribute: name",
		},
		{
			name:           "invalid JSON",
			configJSON:     `{"invalid JSON"`,
			expectedError:  "Provider source config is not a valid JSON list of objects",
		},
		{
			name:           "not a list",
			configJSON:     `{}`,
			expectedError:  "Provider source config is not a valid JSON list of objects",
		},
		{
			name:           "not a list of objects",
			configJSON:     `[[\"Hi\"]]`,
			expectedError:  "Provider source config is not a valid JSON list of objects",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			factory, db := setupTestFactory(t)
			defer testutils.CleanupTestDatabase(t, db)

			ctx := context.Background()

			err := factory.InitialiseFromConfig(ctx, tt.configJSON)
			require.Error(t, err)

			assert.Contains(t, err.Error(), tt.expectedError)
		})
	}
}

// TestInitialiseFromConfig_UpdateExisting tests updating an existing provider source
// Python reference: test_provider_source_factory.py::test_initialise_from_config_update_existing
func TestInitialiseFromConfig_UpdateExisting(t *testing.T) {
	factory, testDB := setupTestFactory(t)
	defer testutils.CleanupTestDatabase(t, testDB)

	ctx := context.Background()

	// Create pre-existing provider
	oldConfig := &model.ProviderSourceConfig{
		BaseURL:         "https://github.com",
		ApiURL:          "https://api.github.com",
		ClientID:        "old-client-id",
		ClientSecret:    "old-client-secret",
		LoginButtonText: "Old Button Text",
		PrivateKeyPath:  "/old/path",
		AppID:           "111111",
	}

	source := model.NewProviderSource("Test Pre-existing", "test-pre-existing", model.ProviderSourceTypeGithub, oldConfig)
	require.NoError(t, testDB.DB.Create(source.ToDBModel()).Error)

	// Update with new config
	configJSON := `[{
		"name": "Test Pre-existing",
		"type": "github",
		"base_url": "https://github.com",
		"api_url": "https://api.github.com",
		"client_id": "new-client-id",
		"client_secret": "new-client-secret",
		"login_button_text": "New Button Text",
		"private_key_path": "/new/path",
		"app_id": "222222"
	}]`

	err := factory.InitialiseFromConfig(ctx, configJSON)
	require.NoError(t, err)

	// Verify updated in database
	var dbModel sqldb.ProviderSourceDB
	err = testDB.DB.Where("api_name = ?", "test-pre-existing").First(&dbModel).Error
	require.NoError(t, err)

	assert.Equal(t, "Test Pre-existing", dbModel.Name)
	assert.Equal(t, "test-pre-existing", *dbModel.APIName)
	assert.Equal(t, sqldb.ProviderSourceTypeGithub, dbModel.ProviderSourceType)

	// Verify config was updated
	var config model.ProviderSourceConfig
	require.NoError(t, sqldb.DecodeBlob(dbModel.Config, &config))
	assert.Equal(t, "new-client-id", config.ClientID)
}

// TestInitialiseFromConfig_Duplicate tests duplicate provider detection
// Python reference: test_provider_source_factory.py::test_initialise_from_config_duplicate
func TestInitialiseFromConfig_Duplicate(t *testing.T) {
	factory, testDB := setupTestFactory(t)
	defer testutils.CleanupTestDatabase(t, testDB)

	ctx := context.Background()

	configJSON := `[{
		"name": "Test Duplicate",
		"type": "github",
		"base_url": "https://github.com",
		"api_url": "https://api.github.com",
		"client_id": "client-1",
		"client_secret": "secret-1",
		"login_button_text": "Button 1",
		"private_key_path": "/path/1",
		"app_id": "111111"
	}, {
		"name": "Test Duplicate",
		"type": "github",
		"base_url": "https://github.com",
		"api_url": "https://api.github.com",
		"client_id": "client-2",
		"client_secret": "secret-2",
		"login_button_text": "Button 2",
		"private_key_path": "/path/2",
		"app_id": "222222"
	}]`

	err := factory.InitialiseFromConfig(ctx, configJSON)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Duplicate Provider Source name found: Test Duplicate")

	// Verify only first provider was created
	var dbModel sqldb.ProviderSourceDB
	err = testDB.DB.Where("api_name = ?", "test-duplicate").First(&dbModel).Error
	require.NoError(t, err)

	var config model.ProviderSourceConfig
	require.NoError(t, sqldb.DecodeBlob(dbModel.Config, &config))
	assert.Equal(t, "client-1", config.ClientID)
}
