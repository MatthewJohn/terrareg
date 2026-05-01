package service

import (
	"context"
	"testing"

	provider_source_model "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/provider_source/model"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/provider_source/repository"
)

// TestNewProviderSourceFactory tests the constructor
func TestNewProviderSourceFactory(t *testing.T) {
	repo := &MockProviderSourceRepository{}
	factory := NewProviderSourceFactory(repo)

	if factory == nil {
		t.Fatal("NewProviderSourceFactory returned nil")
	}

	if factory.repo != repo {
		t.Error("repo not set correctly")
	}

	if factory.classMapping == nil {
		t.Error("classMapping not initialized")
	}
}

// TestProviderSourceFactory_RegisterProviderSourceClass tests registering provider source classes
func TestProviderSourceFactory_RegisterProviderSourceClass(t *testing.T) {
	factory := NewProviderSourceFactory(&MockProviderSourceRepository{})

	// Create a mock provider source class for testing
	class := &MockProviderSourceClass{
		type_: provider_source_model.ProviderSourceTypeGithub,
	}

	factory.RegisterProviderSourceClass(class)

	classes := factory.GetProviderClasses()
	if len(classes) != 1 {
		t.Errorf("GetProviderClasses() returned %d classes, want 1", len(classes))
	}

	if _, exists := classes[provider_source_model.ProviderSourceTypeGithub]; !exists {
		t.Error("GitHub provider source class not registered")
	}
}

// NOTE: GitHub provider source wrapper tests have been removed as they reference
// internal implementation details (githubProviderSourceWrapper) that are not
// exported. The actual OAuth flow is tested in github_provider_source_test.go.

// MockProviderSourceRepository for testing
type MockProviderSourceRepository struct{}

func (m *MockProviderSourceRepository) FindByName(ctx context.Context, name string) (*provider_source_model.ProviderSource, error) {
	return nil, nil
}

func (m *MockProviderSourceRepository) FindByApiName(ctx context.Context, apiName string) (*provider_source_model.ProviderSource, error) {
	return nil, nil
}

func (m *MockProviderSourceRepository) FindAll(ctx context.Context) ([]*provider_source_model.ProviderSource, error) {
	return nil, nil
}

func (m *MockProviderSourceRepository) Upsert(ctx context.Context, source *provider_source_model.ProviderSource) error {
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

// MockProviderSourceClass is a mock for ProviderSourceClass
type MockProviderSourceClass struct {
	type_ provider_source_model.ProviderSourceType
}

func (m *MockProviderSourceClass) Type() provider_source_model.ProviderSourceType {
	return m.type_
}

func (m *MockProviderSourceClass) GenerateDBConfigFromSourceConfig(sourceConfig map[string]interface{}) (*provider_source_model.ProviderSourceConfig, error) {
	return &provider_source_model.ProviderSourceConfig{}, nil
}

func (m *MockProviderSourceClass) ValidateConfig(config *provider_source_model.ProviderSourceConfig) error {
	return nil
}

func (m *MockProviderSourceClass) CreateInstance(name string, repo repository.ProviderSourceRepository, db interface{}) (ProviderSourceInstance, error) {
	return nil, nil
}
