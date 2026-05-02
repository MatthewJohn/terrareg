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

// setupTestNamespaceForPublish creates a test namespace using the real model constructor
func setupTestNamespaceForPublish(t *testing.T, name string) *model.Namespace {
	namespace, err := model.NewNamespace(types.NamespaceName(name), nil, model.NamespaceTypeNone)
	require.NoError(t, err)
	// Use ReconstructNamespace to set ID (ID is private, no SetID method)
	namespace = model.ReconstructNamespace(1, types.NamespaceName(name), nil, model.NamespaceTypeNone)
	return namespace
}

// setupTestProviderWithVersion creates a test provider with a version using the real model constructor
func setupTestProviderWithVersion(t *testing.T, id, namespaceID int, name string) *provider.Provider {
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

func TestPublishProviderVersionCommand_Success_CallsAuditService(t *testing.T) {
	ctx := context.Background()
	mockNSRepo := new(mocks.MockNamespaceRepository)
	mockProviderRepo := new(mocks.MockProviderRepository)
	mockAuditService := new(mocks.MockProviderAuditService)

	namespaceName := "test-ns"
	providerName := "test-provider"
	version := "1.0.0"

	// Setup provider mock
	testProvider := setupTestProviderWithVersion(t, 1, 1, providerName)
	mockProviderRepo.On("FindByNamespaceAndName", ctx, namespaceName, providerName).Return(testProvider, nil).Once()
	mockProviderRepo.On("Save", ctx, mock.AnythingOfType("*provider.Provider")).Return(nil).Once()

	// Setup audit service mock - expect LogProviderVersionIndex call
	mockAuditService.On("LogProviderVersionIndex", ctx, providerName, namespaceName, version).Return(nil)

	cmd := NewPublishProviderVersionCommand(mockProviderRepo, mockNSRepo, mockAuditService)

	req := PublishProviderVersionRequest{
		Namespace:    namespaceName,
		ProviderName: providerName,
		Version:      version,
		Protocol:     "5.0",
		IsBeta:       false,
	}

	result, err := cmd.Execute(ctx, req)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, version, result.Version())

	// Verify mocks were called (synchronous - no sleep needed)
	mockProviderRepo.AssertExpectations(t)
	mockAuditService.AssertExpectations(t)
}

func TestPublishProviderVersionCommand_ProviderNotFound_NoAuditCall(t *testing.T) {
	ctx := context.Background()
	mockNSRepo := new(mocks.MockNamespaceRepository)
	mockProviderRepo := new(mocks.MockProviderRepository)
	mockAuditService := new(mocks.MockProviderAuditService)

	namespaceName := "test-ns"
	providerName := "test-provider"
	version := "1.0.0"

	// Setup provider mock - provider not found
	mockProviderRepo.On("FindByNamespaceAndName", ctx, namespaceName, providerName).Return(nil, errors.New("provider not found")).Once()

	// Audit service should NOT be called
	// (no mock setup for audit service)

	cmd := NewPublishProviderVersionCommand(mockProviderRepo, mockNSRepo, mockAuditService)

	req := PublishProviderVersionRequest{
		Namespace:    namespaceName,
		ProviderName: providerName,
		Version:      version,
		Protocol:     "5.0",
		IsBeta:       false,
	}

	result, err := cmd.Execute(ctx, req)

	require.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "provider not found")

	// Verify mocks were called
	mockProviderRepo.AssertExpectations(t)

	// Verify audit service was NOT called (failed before audit)
	mockAuditService.AssertNotCalled(t, "LogProviderVersionIndex", ctx, mock.Anything, mock.Anything, mock.Anything)
}

func TestPublishProviderVersionCommand_SaveError_NoAuditCall(t *testing.T) {
	ctx := context.Background()
	mockNSRepo := new(mocks.MockNamespaceRepository)
	mockProviderRepo := new(mocks.MockProviderRepository)
	mockAuditService := new(mocks.MockProviderAuditService)

	namespaceName := "test-ns"
	providerName := "test-provider"
	version := "1.0.0"

	// Setup provider mock
	testProvider := setupTestProviderWithVersion(t, 1, 1, providerName)
	mockProviderRepo.On("FindByNamespaceAndName", ctx, namespaceName, providerName).Return(testProvider, nil).Once()
	mockProviderRepo.On("Save", ctx, mock.AnythingOfType("*provider.Provider")).Return(errors.New("database error")).Once()

	// Audit service should NOT be called
	// (no mock setup for audit service)

	cmd := NewPublishProviderVersionCommand(mockProviderRepo, mockNSRepo, mockAuditService)

	req := PublishProviderVersionRequest{
		Namespace:    namespaceName,
		ProviderName: providerName,
		Version:      version,
		Protocol:     "5.0",
		IsBeta:       false,
	}

	result, err := cmd.Execute(ctx, req)

	require.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to save provider")

	// Verify mocks were called
	mockProviderRepo.AssertExpectations(t)

	// Verify audit service was NOT called (failed before audit)
	mockAuditService.AssertNotCalled(t, "LogProviderVersionIndex", ctx, mock.Anything, mock.Anything, mock.Anything)
}

func TestPublishProviderVersionCommand_NilAuditService_NoPanic(t *testing.T) {
	ctx := context.Background()
	mockNSRepo := new(mocks.MockNamespaceRepository)
	mockProviderRepo := new(mocks.MockProviderRepository)

	namespaceName := "test-ns"
	providerName := "test-provider"
	version := "1.0.0"

	// Setup provider mock
	testProvider := setupTestProviderWithVersion(t, 1, 1, providerName)
	mockProviderRepo.On("FindByNamespaceAndName", ctx, namespaceName, providerName).Return(testProvider, nil).Once()
	mockProviderRepo.On("Save", ctx, mock.AnythingOfType("*provider.Provider")).Return(nil).Once()

	// Pass nil for audit service
	cmd := NewPublishProviderVersionCommand(mockProviderRepo, mockNSRepo, nil)

	req := PublishProviderVersionRequest{
		Namespace:    namespaceName,
		ProviderName: providerName,
		Version:      version,
		Protocol:     "5.0",
		IsBeta:       false,
	}

	// Should not panic
	result, err := cmd.Execute(ctx, req)

	require.NoError(t, err)
	require.NotNil(t, result)

	// Verify mocks were called
	mockProviderRepo.AssertExpectations(t)
}
