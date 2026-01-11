package module_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/application/command/module"
	modulemodel "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/module/model"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/module/repository"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/shared"
	"github.com/matthewjohn/terrareg/terrareg-go/test/testutils/mocks"
)

// MockModuleProviderRepository is a mock for testing
type MockModuleProviderRepository struct {
	moduleProviders map[string]*modulemodel.ModuleProvider
}

func NewMockModuleProviderRepository() *MockModuleProviderRepository {
	return &MockModuleProviderRepository{
		moduleProviders: make(map[string]*modulemodel.ModuleProvider),
	}
}

func (m *MockModuleProviderRepository) FindByNamespaceModuleProvider(ctx context.Context, namespace, module, provider string) (*modulemodel.ModuleProvider, error) {
	key := namespace + "/" + module + "/" + provider
	if mp, exists := m.moduleProviders[key]; exists {
		return mp, nil
	}
	return nil, shared.ErrNotFound
}

func (m *MockModuleProviderRepository) Save(ctx context.Context, moduleProvider *modulemodel.ModuleProvider) error {
	key := moduleProvider.Namespace().Name() + "/" + moduleProvider.Module() + "/" + moduleProvider.Provider()
	m.moduleProviders[key] = moduleProvider
	return nil
}

func (m *MockModuleProviderRepository) Create(ctx context.Context, moduleProvider *modulemodel.ModuleProvider) error {
	key := moduleProvider.Namespace().Name() + "/" + moduleProvider.Module() + "/" + moduleProvider.Provider()
	m.moduleProviders[key] = moduleProvider
	return nil
}

func (m *MockModuleProviderRepository) Delete(ctx context.Context, id int) error {
	// For simplicity, we'll just clear the mock in this test
	// In a real implementation, you would delete by ID
	m.moduleProviders = make(map[string]*modulemodel.ModuleProvider)
	return nil
}

func (m *MockModuleProviderRepository) FindByID(ctx context.Context, id int) (*modulemodel.ModuleProvider, error) {
	// Simple implementation that returns the first provider (for testing)
	for _, mp := range m.moduleProviders {
		if mp.ID() == id {
			return mp, nil
		}
	}
	return nil, shared.ErrNotFound
}

func (m *MockModuleProviderRepository) FindByNamespace(ctx context.Context, namespace string) ([]*modulemodel.ModuleProvider, error) {
	var providers []*modulemodel.ModuleProvider
	for _, mp := range m.moduleProviders {
		if mp.Namespace().Name() == namespace {
			providers = append(providers, mp)
		}
	}
	return providers, nil
}

func (m *MockModuleProviderRepository) Search(ctx context.Context, query repository.ModuleSearchQuery) (*repository.ModuleSearchResult, error) {
	return &repository.ModuleSearchResult{
		Modules:    []*modulemodel.ModuleProvider{},
		TotalCount: 0,
	}, nil
}

func (m *MockModuleProviderRepository) Exists(ctx context.Context, namespace, module, provider string) (bool, error) {
	key := namespace + "/" + module + "/" + provider
	_, exists := m.moduleProviders[key]
	return exists, nil
}

func TestPublishModuleVersion_Success(t *testing.T) {
	// Setup
	ctx := context.Background()
	mockRepo := NewMockModuleProviderRepository()
	mockAuditService := new(mocks.MockModuleAuditService)

	// Set up mock expectations - the audit service is called in a goroutine
	mockAuditService.On("LogModuleVersionIndex", mock.Anything, mock.Anything, "testns", "testmod", "aws", "1.0.0").Return(nil)
	mockAuditService.On("LogModuleVersionPublish", mock.Anything, mock.Anything, "testns", "testmod", "aws", "1.0.0").Return(nil)

	// Create test data
	namespace, err := modulemodel.NewNamespace("testns", nil, modulemodel.NamespaceTypeNone)
	require.NoError(t, err)

	moduleProvider, err := modulemodel.NewModuleProvider(namespace, "testmod", "aws")
	require.NoError(t, err)

	// Save to mock repository
	err = mockRepo.Save(ctx, moduleProvider)
	require.NoError(t, err)

	// Create command with proper mock
	command := module.NewPublishModuleVersionCommand(mockRepo, mockAuditService)

	// Test data
	req := module.PublishModuleVersionRequest{
		Namespace:   "testns",
		Module:      "testmod",
		Provider:    "aws",
		Version:     "1.0.0",
		Beta:        false,
		Description: stringPtr("Test version"),
		Owner:       stringPtr("testowner"),
	}

	// Execute
	result, err := command.Execute(ctx, req)

	// Wait for goroutines to complete (audit logging happens in background)
	time.Sleep(10 * time.Millisecond)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, result)

	assert.Equal(t, "1.0.0", result.Version().String())
	assert.False(t, result.IsBeta())
	assert.True(t, result.IsPublished())
	assert.Equal(t, "testowner", *result.Owner())
	assert.Equal(t, "Test version", *result.Description())

	// Verify mock was called
	mockAuditService.AssertExpectations(t)
}

func TestPublishModuleVersion_VersionAlreadyExists(t *testing.T) {
	// Setup
	ctx := context.Background()
	mockRepo := NewMockModuleProviderRepository()
	mockAuditService := new(mocks.MockModuleAuditService)

	// Create test data with existing version
	namespace, err := modulemodel.NewNamespace("testns", nil, modulemodel.NamespaceTypeNone)
	require.NoError(t, err)

	moduleProvider, err := modulemodel.NewModuleProvider(namespace, "testmod", "aws")
	require.NoError(t, err)

	// Add existing version
	existingDetails := modulemodel.NewModuleDetails(nil)
	_, err = moduleProvider.PublishVersion("1.0.0", existingDetails, false)
	require.NoError(t, err)

	// Save to mock repository
	err = mockRepo.Save(ctx, moduleProvider)
	require.NoError(t, err)

	// Create command with proper mock (no expectations since version exists and errors before audit)
	command := module.NewPublishModuleVersionCommand(mockRepo, mockAuditService)

	// Test data - trying to publish same version
	req := module.PublishModuleVersionRequest{
		Namespace: "testns",
		Module:    "testmod",
		Provider:  "aws",
		Version:   "1.0.0", // Same as existing
		Beta:      false,
	}

	// Execute
	result, err := command.Execute(ctx, req)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "version 1.0.0 already exists")
}

func TestPublishModuleVersion_ModuleProviderNotFound(t *testing.T) {
	// Setup
	ctx := context.Background()
	mockRepo := NewMockModuleProviderRepository()
	mockAuditService := new(mocks.MockModuleAuditService)

	// Create command with proper mock (no expectations since provider not found)
	command := module.NewPublishModuleVersionCommand(mockRepo, mockAuditService)

	// Test data - non-existent module provider
	req := module.PublishModuleVersionRequest{
		Namespace: "nonexistent",
		Module:    "testmod",
		Provider:  "aws",
		Version:   "1.0.0",
		Beta:      false,
	}

	// Execute
	result, err := command.Execute(ctx, req)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "module provider nonexistent/testmod/aws not found")
}

func TestPublishModuleVersion_WithBetaVersion(t *testing.T) {
	// Setup
	ctx := context.Background()
	mockRepo := NewMockModuleProviderRepository()
	mockAuditService := new(mocks.MockModuleAuditService)

	// Set up mock expectations (using valid version format per updated regex)
	mockAuditService.On("LogModuleVersionIndex", mock.Anything, mock.Anything, "testns", "testmod", "aws", "1.0.0-betal").Return(nil)
	mockAuditService.On("LogModuleVersionPublish", mock.Anything, mock.Anything, "testns", "testmod", "aws", "1.0.0-betal").Return(nil)

	// Create test data
	namespace, err := modulemodel.NewNamespace("testns", nil, modulemodel.NamespaceTypeNone)
	require.NoError(t, err)

	moduleProvider, err := modulemodel.NewModuleProvider(namespace, "testmod", "aws")
	require.NoError(t, err)

	// Save to mock repository
	err = mockRepo.Save(ctx, moduleProvider)
	require.NoError(t, err)

	// Create command with proper mock
	command := module.NewPublishModuleVersionCommand(mockRepo, mockAuditService)

	// Test data
	req := module.PublishModuleVersionRequest{
		Namespace:   "testns",
		Module:      "testmod",
		Provider:    "aws",
		Version:     "1.0.0-betal",
		Beta:        true,
		Description: stringPtr("Beta version"),
		Owner:       stringPtr("testowner"),
	}

	// Execute
	result, err := command.Execute(ctx, req)

	// Wait for goroutines to complete (audit logging happens in background)
	time.Sleep(10 * time.Millisecond)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, result)

	assert.Equal(t, "1.0.0-betal", result.Version().String())
	assert.True(t, result.IsBeta())
	assert.True(t, result.IsPublished())

	// Verify mock was called
	mockAuditService.AssertExpectations(t)
}

// Helper function for string pointers
func stringPtr(s string) *string {
	return &s
}
