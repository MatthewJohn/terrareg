package provider_source

import (
	"context"
	"testing"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/provider_source/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockProviderSourceRepository is a mock for testing
type MockProviderSourceRepository struct {
	findByNameFunc func(ctx context.Context, name string) (*model.ProviderSource, error)
}

func (m *MockProviderSourceRepository) FindByName(ctx context.Context, name string) (*model.ProviderSource, error) {
	if m.findByNameFunc != nil {
		return m.findByNameFunc(ctx, name)
	}
	return nil, nil
}

func (m *MockProviderSourceRepository) FindByApiName(ctx context.Context, apiName string) (*model.ProviderSource, error) {
	return nil, nil
}

func (m *MockProviderSourceRepository) FindAll(ctx context.Context) ([]*model.ProviderSource, error) {
	return nil, nil
}

func (m *MockProviderSourceRepository) Upsert(ctx context.Context, source *model.ProviderSource) error {
	return nil
}

func (m *MockProviderSourceRepository) Delete(ctx context.Context, name string) error {
	return nil
}

func (m *MockProviderSourceRepository) Exists(ctx context.Context, name string) (bool, error) {
	return false, nil
}

func (m *MockProviderSourceRepository) ExistsByApiName(ctx context.Context, apiName string) (bool, error) {
	return false, nil
}

// MockProviderSourceClass is a mock for testing
type MockProviderSourceClass struct {
	typeFunc       func() model.ProviderSourceType
	generateDBFunc func(sourceConfig map[string]interface{}) (*model.ProviderSourceConfig, error)
	validateFunc   func(config *model.ProviderSourceConfig) error
}

func (m *MockProviderSourceClass) Type() model.ProviderSourceType {
	if m.typeFunc != nil {
		return m.typeFunc()
	}
	return model.ProviderSourceTypeGithub
}

func (m *MockProviderSourceClass) GenerateDBConfigFromSourceConfig(sourceConfig map[string]interface{}) (*model.ProviderSourceConfig, error) {
	if m.generateDBFunc != nil {
		return m.generateDBFunc(sourceConfig)
	}
	return &model.ProviderSourceConfig{}, nil
}

func (m *MockProviderSourceClass) ValidateConfig(config *model.ProviderSourceConfig) error {
	if m.validateFunc != nil {
		return m.validateFunc(config)
	}
	return nil
}

// TestBaseProviderSource_Name tests the Name method
func TestBaseProviderSource_Name(t *testing.T) {
	mockRepo := &MockProviderSourceRepository{}
	mockClass := &MockProviderSourceClass{}

	base := NewBaseProviderSource("test-name", mockRepo, mockClass)

	assert.Equal(t, "test-name", base.Name())
}

// TestBaseProviderSource_ApiName_NotFound tests ApiName when provider source is not found
func TestBaseProviderSource_ApiName_NotFound(t *testing.T) {
	mockRepo := &MockProviderSourceRepository{
		findByNameFunc: func(ctx context.Context, name string) (*model.ProviderSource, error) {
			return nil, nil // Not found
		},
	}
	mockClass := &MockProviderSourceClass{}

	base := NewBaseProviderSource("test-name", mockRepo, mockClass)

	apiName, err := base.ApiName(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, "", apiName)
}

// TestBaseProviderSource_ApiName_Found tests ApiName when provider source exists
func TestBaseProviderSource_ApiName_Found(t *testing.T) {
	expectedApiName := "test-api-name"

	mockRepo := &MockProviderSourceRepository{
		findByNameFunc: func(ctx context.Context, name string) (*model.ProviderSource, error) {
			// Return a mock provider source using constructor
			return model.NewProviderSource(
				name,
				expectedApiName,
				model.ProviderSourceTypeGithub,
				&model.ProviderSourceConfig{},
			), nil
		},
	}
	mockClass := &MockProviderSourceClass{}

	base := NewBaseProviderSource("test-name", mockRepo, mockClass)

	apiName, err := base.ApiName(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, expectedApiName, apiName)
}

// TestBaseProviderSource_Config_NotFound tests Config when provider source is not found
func TestBaseProviderSource_Config_NotFound(t *testing.T) {
	mockRepo := &MockProviderSourceRepository{
		findByNameFunc: func(ctx context.Context, name string) (*model.ProviderSource, error) {
			return nil, nil // Not found
		},
	}
	mockClass := &MockProviderSourceClass{}

	base := NewBaseProviderSource("test-name", mockRepo, mockClass)

	config, err := base.Config(context.Background())
	assert.NoError(t, err)
	assert.Nil(t, config)
}

// TestBaseProviderSource_Config_Found tests Config when provider source exists
func TestBaseProviderSource_Config_Found(t *testing.T) {
	expectedConfig := &model.ProviderSourceConfig{
		BaseURL:         "https://github.com",
		ApiURL:          "https://api.github.com",
		ClientID:        "test-client-id",
		ClientSecret:    "test-client-secret",
		LoginButtonText: "Login with GitHub",
	}

	mockRepo := &MockProviderSourceRepository{
		findByNameFunc: func(ctx context.Context, name string) (*model.ProviderSource, error) {
			// Return a mock provider source using constructor
			ps := model.NewProviderSource(
				name,
				"test-api-name",
				model.ProviderSourceTypeGithub,
				expectedConfig,
			)
			return ps, nil
		},
	}
	mockClass := &MockProviderSourceClass{}

	base := NewBaseProviderSource("test-name", mockRepo, mockClass)

	config, err := base.Config(context.Background())
	require.NoError(t, err)
	require.NotNil(t, config)

	assert.Equal(t, expectedConfig.BaseURL, config.BaseURL)
	assert.Equal(t, expectedConfig.ApiURL, config.ApiURL)
	assert.Equal(t, expectedConfig.ClientID, config.ClientID)
	assert.Equal(t, expectedConfig.ClientSecret, config.ClientSecret)
}
