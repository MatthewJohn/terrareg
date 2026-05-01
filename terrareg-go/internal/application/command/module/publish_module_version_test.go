package module_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/application/command/module"
	modulemodel "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/module/model"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/module/repository"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/shared"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/shared/types"
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

func (m *MockModuleProviderRepository) FindByNamespaceModuleProvider(ctx context.Context, namespace types.NamespaceName, module types.ModuleName, provider types.ModuleProviderName) (*modulemodel.ModuleProvider, error) {
	key := string(namespace) + "/" + string(module) + "/" + string(provider)
	if mp, exists := m.moduleProviders[key]; exists {
		return mp, nil
	}
	return nil, shared.ErrNotFound
}

func (m *MockModuleProviderRepository) Save(ctx context.Context, moduleProvider *modulemodel.ModuleProvider) error {
	key := string(moduleProvider.Namespace().Name()) + "/" + string(moduleProvider.Module()) + "/" + string(moduleProvider.Provider())
	m.moduleProviders[key] = moduleProvider
	return nil
}

func (m *MockModuleProviderRepository) Create(ctx context.Context, moduleProvider *modulemodel.ModuleProvider) error {
	key := string(moduleProvider.Namespace().Name()) + "/" + string(moduleProvider.Module()) + "/" + string(moduleProvider.Provider())
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

func (m *MockModuleProviderRepository) FindByNamespace(ctx context.Context, namespace types.NamespaceName) ([]*modulemodel.ModuleProvider, error) {
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

func (m *MockModuleProviderRepository) Exists(ctx context.Context, namespace types.NamespaceName, module types.ModuleName, provider types.ModuleProviderName) (bool, error) {
	key := string(namespace) + "/" + string(module) + "/" + string(provider)
	_, exists := m.moduleProviders[key]
	return exists, nil
}

// TestPublishModuleVersion_Success tests publishing an existing module version
// Python reference: test/unit/terrareg/server/test_api_terrareg_module_version_publish.py
func TestPublishModuleVersion_Success(t *testing.T) {
	// Setup
	ctx := context.Background()
	mockRepo := NewMockModuleProviderRepository()
	mockAuditService := new(mocks.MockModuleAuditService)

	// Set up mock expectations
	mockAuditService.On("LogModuleVersionPublish", mock.Anything, mock.Anything, types.NamespaceName("testns"), types.ModuleName("testmod"), types.ModuleProviderName("aws"), types.ModuleVersion("1.0.0")).Return(nil)

	// Create test data
	namespace, err := modulemodel.NewNamespace(types.NamespaceName("testns"), nil, modulemodel.NamespaceTypeNone)
	require.NoError(t, err)

	moduleProvider, err := modulemodel.NewModuleProvider(namespace, types.ModuleName("testmod"), types.ModuleProviderName("aws"))
	require.NoError(t, err)

	// Create an existing unpublished version (simulating upload/index)
	existingDetails := modulemodel.NewModuleDetails(nil)
	existingVersion, err := moduleProvider.PublishVersion("1.0.0", existingDetails, false)
	require.NoError(t, err)

	// Verify version is not yet published
	assert.False(t, existingVersion.IsPublished())

	// Save to mock repository
	err = mockRepo.Save(ctx, moduleProvider)
	require.NoError(t, err)

	// Create command
	command, err := module.NewPublishModuleVersionCommand(mockRepo, mockAuditService)
	require.NoError(t, err)

	// Execute - publish the existing version
	err = command.Execute(ctx, "testns", "testmod", "aws", "1.0.0")

	// Assert
	require.NoError(t, err)

	// Verify the version is now published
	assert.True(t, existingVersion.IsPublished())

	// Verify mock was called
	mockAuditService.AssertExpectations(t)
}

// TestPublishModuleVersion_IdempotentRepublish tests that re-publishing is idempotent
// Python reference: test/unit/terrareg/models/test_module_version.py::test_publish
func TestPublishModuleVersion_IdempotentRepublish(t *testing.T) {
	// Setup
	ctx := context.Background()
	mockRepo := NewMockModuleProviderRepository()
	mockAuditService := new(mocks.MockModuleAuditService)

	// Create test data with existing published version
	namespace, err := modulemodel.NewNamespace(types.NamespaceName("testns"), nil, modulemodel.NamespaceTypeNone)
	require.NoError(t, err)

	moduleProvider, err := modulemodel.NewModuleProvider(namespace, types.ModuleName("testmod"), types.ModuleProviderName("aws"))
	require.NoError(t, err)

	// Add existing version and mark as published
	existingDetails := modulemodel.NewModuleDetails(nil)
	existingVersion, err := moduleProvider.PublishVersion("1.0.0", existingDetails, false)
	require.NoError(t, err)
	err = existingVersion.Publish()
	require.NoError(t, err)

	// Verify version is already published
	assert.True(t, existingVersion.IsPublished())

	// Save to mock repository
	err = mockRepo.Save(ctx, moduleProvider)
	require.NoError(t, err)

	// Create command (no audit expectations - already published versions return early)
	command, err := module.NewPublishModuleVersionCommand(mockRepo, mockAuditService)
	require.NoError(t, err)

	// Execute - trying to publish same version (idempotent)
	err = command.Execute(ctx, "testns", "testmod", "aws", "1.0.0")

	// Assert - operation is idempotent, should succeed without error
	require.NoError(t, err)

	// Version should still be published
	assert.True(t, existingVersion.IsPublished())
}

// TestPublishModuleVersion_ModuleProviderNotFound tests error when module provider doesn't exist
func TestPublishModuleVersion_ModuleProviderNotFound(t *testing.T) {
	// Setup
	ctx := context.Background()
	mockRepo := NewMockModuleProviderRepository()
	mockAuditService := new(mocks.MockModuleAuditService)

	// Create command
	command, err := module.NewPublishModuleVersionCommand(mockRepo, mockAuditService)
	require.NoError(t, err)

	// Execute - non-existent module provider
	err = command.Execute(ctx, "nonexistent", "testmod", "aws", "1.0.0")

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "module provider nonexistent/testmod/aws not found")
}

// TestPublishModuleVersion_VersionNotFound tests error when version doesn't exist
// Python behavior: Would fail when getting module_version from database
func TestPublishModuleVersion_VersionNotFound(t *testing.T) {
	// Setup
	ctx := context.Background()
	mockRepo := NewMockModuleProviderRepository()
	mockAuditService := new(mocks.MockModuleAuditService)

	// Create test data - module provider exists but no versions
	namespace, err := modulemodel.NewNamespace(types.NamespaceName("testns"), nil, modulemodel.NamespaceTypeNone)
	require.NoError(t, err)

	moduleProvider, err := modulemodel.NewModuleProvider(namespace, types.ModuleName("testmod"), types.ModuleProviderName("aws"))
	require.NoError(t, err)

	// Save to mock repository (no versions added)
	err = mockRepo.Save(ctx, moduleProvider)
	require.NoError(t, err)

	// Create command
	command, err := module.NewPublishModuleVersionCommand(mockRepo, mockAuditService)
	require.NoError(t, err)

	// Execute - version doesn't exist
	err = command.Execute(ctx, "testns", "testmod", "aws", "1.0.0")

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "version 1.0.0 not found")
}

// TestPublishModuleVersion_NonBetaVersionUpdatesLatestVersionId tests that publishing a non-beta version updates latest_version_id
// Python reference: test/unit/terrareg/models/test_module_provider.py::test_calculate_latest_version
func TestPublishModuleVersion_NonBetaVersionUpdatesLatestVersionId(t *testing.T) {
	// Setup
	ctx := context.Background()
	mockRepo := NewMockModuleProviderRepository()
	mockAuditService := new(mocks.MockModuleAuditService)

	// Set up mock expectations
	mockAuditService.On("LogModuleVersionPublish", mock.Anything, mock.Anything, types.NamespaceName("testns"), types.ModuleName("testmod"), types.ModuleProviderName("aws"), types.ModuleVersion("1.0.0")).Return(nil)

	// Create test data
	namespace, err := modulemodel.NewNamespace(types.NamespaceName("testns"), nil, modulemodel.NamespaceTypeNone)
	require.NoError(t, err)

	moduleProvider, err := modulemodel.NewModuleProvider(namespace, types.ModuleName("testmod"), types.ModuleProviderName("aws"))
	require.NoError(t, err)

	// Create an existing unpublished non-beta version
	existingDetails := modulemodel.NewModuleDetails(nil)
	existingVersion, err := moduleProvider.PublishVersion("1.0.0", existingDetails, false)
	require.NoError(t, err)

	// Save to mock repository
	err = mockRepo.Save(ctx, moduleProvider)
	require.NoError(t, err)

	// Create command
	command, err := module.NewPublishModuleVersionCommand(mockRepo, mockAuditService)
	require.NoError(t, err)

	// Execute
	err = command.Execute(ctx, "testns", "testmod", "aws", "1.0.0")
	require.NoError(t, err)

	// Assert - version is published and is the latest
	assert.True(t, existingVersion.IsPublished())
	latestVersion := moduleProvider.GetLatestVersion()
	assert.NotNil(t, latestVersion)
	assert.Equal(t, "1.0.0", latestVersion.Version().String())
}

// TestPublishModuleVersion_BetaVersionDoesNotBecomeLatest tests that publishing a beta version doesn't update latest_version_id
// Python reference: test/unit/terrareg/models/test_module_provider.py::test_calculate_latest_version_with_beta
func TestPublishModuleVersion_BetaVersionDoesNotBecomeLatest(t *testing.T) {
	// Setup
	ctx := context.Background()
	mockRepo := NewMockModuleProviderRepository()
	mockAuditService := new(mocks.MockModuleAuditService)

	// Set up mock expectations
	mockAuditService.On("LogModuleVersionPublish", mock.Anything, mock.Anything, types.NamespaceName("testns"), types.ModuleName("testmod"), types.ModuleProviderName("aws"), types.ModuleVersion("1.0.0-beta")).Return(nil)

	// Create test data
	namespace, err := modulemodel.NewNamespace(types.NamespaceName("testns"), nil, modulemodel.NamespaceTypeNone)
	require.NoError(t, err)

	moduleProvider, err := modulemodel.NewModuleProvider(namespace, types.ModuleName("testmod"), types.ModuleProviderName("aws"))
	require.NoError(t, err)

	// Create an existing unpublished beta version
	existingDetails := modulemodel.NewModuleDetails(nil)
	existingVersion, err := moduleProvider.PublishVersion("1.0.0-beta", existingDetails, true)
	require.NoError(t, err)

	// Save to mock repository
	err = mockRepo.Save(ctx, moduleProvider)
	require.NoError(t, err)

	// Create command
	command, err := module.NewPublishModuleVersionCommand(mockRepo, mockAuditService)
	require.NoError(t, err)

	// Execute
	err = command.Execute(ctx, "testns", "testmod", "aws", "1.0.0-beta")
	require.NoError(t, err)

	// Assert - beta version is published but NOT the latest
	assert.True(t, existingVersion.IsPublished())
	latestVersion := moduleProvider.GetLatestVersion()
	assert.Nil(t, latestVersion, "Beta versions should not become the latest version")
}
