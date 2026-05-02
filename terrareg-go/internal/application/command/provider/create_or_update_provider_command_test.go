package provider

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/module/model"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/provider"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/shared/types"
	"github.com/matthewjohn/terrareg/terrareg-go/test/testutils/mocks"
)

// setupTestNamespace creates a test namespace using the real model constructor
func setupTestNamespace(t *testing.T, name string) *model.Namespace {
	namespace, err := model.NewNamespace(types.NamespaceName(name), nil, model.NamespaceTypeNone)
	require.NoError(t, err)
	// Use ReconstructNamespace to set ID (ID is private, no SetID method)
	namespace = model.ReconstructNamespace(1, types.NamespaceName(name), nil, model.NamespaceTypeNone)
	return namespace
}

// setupTestProvider creates a test provider using the real model constructor
func setupTestProvider(t *testing.T, id, namespaceID int, name string) *provider.Provider {
	testProvider := provider.NewProvider(
		namespaceID,
		name,
		nil,         // description
		"community", // tier
		nil,         // categoryID
		nil,         // repositoryID
		false,       // useProviderSourceAuth
	)
	testProvider.SetID(id)
	return testProvider
}

func TestCreateOrUpdateProviderCommand_CreateNew_Success_CallsAuditService(t *testing.T) {
	ctx := context.Background()
	mockNSRepo := new(mocks.MockNamespaceRepository)
	mockProviderRepo := new(mocks.MockProviderRepository)
	mockAuditService := new(mocks.MockProviderAuditService)

	namespaceName := "test-ns"
	providerName := "test-provider"

	// Setup namespace mock
	testNamespace := setupTestNamespace(t, namespaceName)
	mockNSRepo.On("FindByName", ctx, types.NamespaceName(namespaceName)).Return(testNamespace, nil).Once()

	// Setup provider mock - provider doesn't exist (creating new)
	mockProviderRepo.On("FindByNamespaceAndName", ctx, namespaceName, providerName).Return(nil, nil).Once()
	mockProviderRepo.On("Save", ctx, mock.AnythingOfType("*provider.Provider")).Return(nil).Once()

	// Setup audit service mock - expect LogProviderCreate call
	mockAuditService.On("LogProviderCreate", ctx, providerName, namespaceName).Return(nil)

	cmd := NewCreateOrUpdateProviderCommand(mockProviderRepo, mockNSRepo, mockAuditService)

	req := CreateOrUpdateProviderRequest{
		Namespace: namespaceName,
		Name:      providerName,
		Tier:      "community",
	}

	result, err := cmd.Execute(ctx, req)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, providerName, result.Name())
	assert.Equal(t, "community", result.Tier())

	// Verify mocks were called (synchronous - no sleep needed)
	mockNSRepo.AssertExpectations(t)
	mockProviderRepo.AssertExpectations(t)
	mockAuditService.AssertExpectations(t)
}

func TestCreateOrUpdateProviderCommand_UpdateExisting_Success_NoAuditCall(t *testing.T) {
	ctx := context.Background()
	mockNSRepo := new(mocks.MockNamespaceRepository)
	mockProviderRepo := new(mocks.MockProviderRepository)
	mockAuditService := new(mocks.MockProviderAuditService)

	namespaceName := "test-ns"
	providerName := "test-provider"

	// Setup namespace mock
	testNamespace := setupTestNamespace(t, namespaceName)
	mockNSRepo.On("FindByName", ctx, types.NamespaceName(namespaceName)).Return(testNamespace, nil).Once()

	// Setup provider mock - provider exists (updating)
	testProvider := setupTestProvider(t, 1, 1, providerName)
	mockProviderRepo.On("FindByNamespaceAndName", ctx, namespaceName, providerName).Return(testProvider, nil).Once()
	mockProviderRepo.On("Save", ctx, mock.AnythingOfType("*provider.Provider")).Return(nil).Once()

	// Audit service should NOT be called for updates
	// (no mock setup for audit service)

	cmd := NewCreateOrUpdateProviderCommand(mockProviderRepo, mockNSRepo, mockAuditService)

	req := CreateOrUpdateProviderRequest{
		Namespace: namespaceName,
		Name:      providerName,
		Tier:      "community",
	}

	result, err := cmd.Execute(ctx, req)

	require.NoError(t, err)
	require.NotNil(t, result)

	// Verify mocks were called
	mockNSRepo.AssertExpectations(t)
	mockProviderRepo.AssertExpectations(t)

	// Verify audit service was NOT called (updates don't audit)
	mockAuditService.AssertNotCalled(t, "LogProviderCreate", ctx, mock.Anything, mock.Anything)
}

func TestCreateOrUpdateProviderCommand_NamespaceNotFound_NoAuditCall(t *testing.T) {
	ctx := context.Background()
	mockNSRepo := new(mocks.MockNamespaceRepository)
	mockProviderRepo := new(mocks.MockProviderRepository)
	mockAuditService := new(mocks.MockProviderAuditService)

	namespaceName := "test-ns"
	providerName := "test-provider"

	// Setup namespace mock - return error (not found)
	mockNSRepo.On("FindByName", ctx, types.NamespaceName(namespaceName)).Return(nil, errors.New("namespace not found")).Once()

	// Audit service should NOT be called
	// (no mock setup for audit service)

	cmd := NewCreateOrUpdateProviderCommand(mockProviderRepo, mockNSRepo, mockAuditService)

	req := CreateOrUpdateProviderRequest{
		Namespace: namespaceName,
		Name:      providerName,
		Tier:      "community",
	}

	result, err := cmd.Execute(ctx, req)

	require.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "namespace not found")

	// Verify mocks were called
	mockNSRepo.AssertExpectations(t)

	// Verify audit service was NOT called (failed before audit)
	mockAuditService.AssertNotCalled(t, "LogProviderCreate", ctx, mock.Anything, mock.Anything)
}

func TestCreateOrUpdateProviderCommand_SaveError_NoAuditCall(t *testing.T) {
	ctx := context.Background()
	mockNSRepo := new(mocks.MockNamespaceRepository)
	mockProviderRepo := new(mocks.MockProviderRepository)
	mockAuditService := new(mocks.MockProviderAuditService)

	namespaceName := "test-ns"
	providerName := "test-provider"

	// Setup namespace mock
	testNamespace := setupTestNamespace(t, namespaceName)
	mockNSRepo.On("FindByName", ctx, types.NamespaceName(namespaceName)).Return(testNamespace, nil).Once()

	// Setup provider mock - creating new, but save fails
	mockProviderRepo.On("FindByNamespaceAndName", ctx, namespaceName, providerName).Return(nil, nil).Once()
	mockProviderRepo.On("Save", ctx, mock.AnythingOfType("*provider.Provider")).Return(errors.New("database error")).Once()

	// Audit service should NOT be called
	// (no mock setup for audit service)

	cmd := NewCreateOrUpdateProviderCommand(mockProviderRepo, mockNSRepo, mockAuditService)

	req := CreateOrUpdateProviderRequest{
		Namespace: namespaceName,
		Name:      providerName,
		Tier:      "community",
	}

	result, err := cmd.Execute(ctx, req)

	require.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to save provider")

	// Verify mocks were called
	mockNSRepo.AssertExpectations(t)
	mockProviderRepo.AssertExpectations(t)

	// Verify audit service was NOT called (failed before audit)
	mockAuditService.AssertNotCalled(t, "LogProviderCreate", ctx, mock.Anything, mock.Anything)
}

func TestCreateOrUpdateProviderCommand_NilAuditService_NoPanic(t *testing.T) {
	ctx := context.Background()
	mockNSRepo := new(mocks.MockNamespaceRepository)
	mockProviderRepo := new(mocks.MockProviderRepository)

	namespaceName := "test-ns"
	providerName := "test-provider"

	// Setup namespace mock
	testNamespace := setupTestNamespace(t, namespaceName)
	mockNSRepo.On("FindByName", ctx, types.NamespaceName(namespaceName)).Return(testNamespace, nil).Once()

	// Setup provider mock
	mockProviderRepo.On("FindByNamespaceAndName", ctx, namespaceName, providerName).Return(nil, nil).Once()
	mockProviderRepo.On("Save", ctx, mock.AnythingOfType("*provider.Provider")).Return(nil).Once()

	// Pass nil for audit service
	cmd := NewCreateOrUpdateProviderCommand(mockProviderRepo, mockNSRepo, nil)

	req := CreateOrUpdateProviderRequest{
		Namespace: namespaceName,
		Name:      providerName,
		Tier:      "community",
	}

	// Should not panic
	result, err := cmd.Execute(ctx, req)

	require.NoError(t, err)
	require.NotNil(t, result)

	// Verify mocks were called
	mockNSRepo.AssertExpectations(t)
	mockProviderRepo.AssertExpectations(t)
}

func TestCreateOrUpdateProviderCommand_DefaultTierApplied(t *testing.T) {
	ctx := context.Background()
	mockNSRepo := new(mocks.MockNamespaceRepository)
	mockProviderRepo := new(mocks.MockProviderRepository)
	mockAuditService := new(mocks.MockProviderAuditService)

	namespaceName := "test-ns"
	providerName := "test-provider"

	// Setup namespace mock
	testNamespace := setupTestNamespace(t, namespaceName)
	mockNSRepo.On("FindByName", ctx, types.NamespaceName(namespaceName)).Return(testNamespace, nil).Once()

	// Setup provider mock - provider doesn't exist
	mockProviderRepo.On("FindByNamespaceAndName", ctx, namespaceName, providerName).Return(nil, nil).Once()
	mockProviderRepo.On("Save", ctx, mock.AnythingOfType("*provider.Provider")).Return(nil).Once()

	// Setup audit service mock
	mockAuditService.On("LogProviderCreate", ctx, providerName, namespaceName).Return(nil)

	cmd := NewCreateOrUpdateProviderCommand(mockProviderRepo, mockNSRepo, mockAuditService)

	req := CreateOrUpdateProviderRequest{
		Namespace: namespaceName,
		Name:      providerName,
		Tier:      "", // Empty tier should default to "community"
	}

	result, err := cmd.Execute(ctx, req)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, "community", result.Tier())

	// Verify mocks were called
	mockNSRepo.AssertExpectations(t)
	mockProviderRepo.AssertExpectations(t)
	mockAuditService.AssertExpectations(t)
}
