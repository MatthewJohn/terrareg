package gpgkey

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	gpgkeyModel "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/gpgkey/model"
	gpgkeyService "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/gpgkey/service"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/shared/types"
)

// MockGPGKeyService is a mock for the GPGKeyServiceInterface
type MockGPGKeyService struct {
	mock.Mock
}

// Ensure MockGPGKeyService implements the interface at compile time
var _ gpgkeyService.GPGKeyServiceInterface = (*MockGPGKeyService)(nil)

func (m *MockGPGKeyService) CreateGPGKey(ctx context.Context, req gpgkeyService.CreateGPGKeyRequest) (*gpgkeyModel.GPGKey, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*gpgkeyModel.GPGKey), args.Error(1)
}

func (m *MockGPGKeyService) DeleteGPGKey(ctx context.Context, namespace, keyID string) error {
	args := m.Called(ctx, namespace, keyID)
	return args.Error(0)
}

// MockGpgKeyAuditService is a mock for the GpgKeyAuditService
type MockGpgKeyAuditService struct {
	mock.Mock
}

func (m *MockGpgKeyAuditService) LogGpgKeyCreate(ctx context.Context, keyID, namespace string) error {
	args := m.Called(ctx, keyID, namespace)
	return args.Error(0)
}

func (m *MockGpgKeyAuditService) LogGpgKeyDelete(ctx context.Context, keyID, namespace string) error {
	args := m.Called(ctx, keyID, namespace)
	return args.Error(0)
}

// setupTestGPGKey creates a test GPG key using the real model constructor
func setupTestGPGKey(t *testing.T, keyID, namespace string) *gpgkeyModel.GPGKey {
	// Create a real GPG key using the model's constructor
	gpgKey, err := gpgkeyModel.NewGPGKey(
		1, // namespaceID
		"-----BEGIN PGP PUBLIC KEY BLOCK-----\ntest-armor\n-----END PGP PUBLIC KEY BLOCK-----",
		keyID,
		"1234567890ABCDEF1234567890ABCDEF12345678", // fingerprint
	)
	require.NoError(t, err)

	// Set the namespace
	gpgKey.SetNamespace(gpgkeyModel.NewNamespace(1, types.NamespaceName(namespace)))

	// Set ID for testing
	gpgKey.SetID(1)

	return gpgKey
}

func TestManageGPGKeyCommand_CreateGPGKey_Success_CallsAuditService(t *testing.T) {
	ctx := context.Background()
	mockGPGService := new(MockGPGKeyService)
	mockAuditService := new(MockGpgKeyAuditService)

	testGPGKey := setupTestGPGKey(t, "ABC123", "test-namespace")

	// Setup GPG service mock
	mockGPGService.On("CreateGPGKey", ctx, mock.MatchedBy(func(req gpgkeyService.CreateGPGKeyRequest) bool {
		return req.Namespace == "test-namespace" && req.ASCIILArmor == "test-armor"
	})).Return(testGPGKey, nil)

	// Setup audit service mock - expect LogGpgKeyCreate call
	mockAuditService.On("LogGpgKeyCreate", ctx, "ABC123", "test-namespace").Return(nil)

	cmd := NewManageGPGKeyCommand(mockGPGService, mockAuditService)

	req := CreateGPGKeyRequest{
		Namespace:   "test-namespace",
		ASCIILArmor: "test-armor",
	}

	result, err := cmd.CreateGPGKey(ctx, req)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, "gpg-keys", result.Type)
	assert.Equal(t, "ABC123", result.ID)

	// Verify mocks were called (synchronous - no sleep needed)
	mockGPGService.AssertExpectations(t)
	mockAuditService.AssertExpectations(t)
}

func TestManageGPGKeyCommand_CreateGPGKey_ServiceError_NoAuditCall(t *testing.T) {
	ctx := context.Background()
	mockGPGService := new(MockGPGKeyService)
	mockAuditService := new(MockGpgKeyAuditService)

	// Setup GPG service mock to return error
	mockGPGService.On("CreateGPGKey", ctx, mock.AnythingOfType("service.CreateGPGKeyRequest")).
		Return(nil, errors.New("service error"))

	cmd := NewManageGPGKeyCommand(mockGPGService, mockAuditService)

	req := CreateGPGKeyRequest{
		Namespace:   "test-namespace",
		ASCIILArmor: "test-armor",
	}

	result, err := cmd.CreateGPGKey(ctx, req)

	require.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to create GPG key")

	// Verify audit service was NOT called (since operation failed)
	mockAuditService.AssertNotCalled(t, "LogGpgKeyCreate", ctx, mock.Anything, mock.Anything)
}

func TestManageGPGKeyCommand_DeleteGPGKey_Success_CallsAuditService(t *testing.T) {
	ctx := context.Background()
	mockGPGService := new(MockGPGKeyService)
	mockAuditService := new(MockGpgKeyAuditService)

	// Setup GPG service mock - successful deletion
	mockGPGService.On("DeleteGPGKey", ctx, "test-namespace", "ABC123").Return(nil)

	// Setup audit service mock - expect LogGpgKeyDelete call
	mockAuditService.On("LogGpgKeyDelete", ctx, "ABC123", "test-namespace").Return(nil)

	cmd := NewManageGPGKeyCommand(mockGPGService, mockAuditService)

	req := DeleteGPGKeyRequest{
		Namespace: "test-namespace",
		KeyID:     "ABC123",
	}

	err := cmd.DeleteGPGKey(ctx, req)

	require.NoError(t, err)

	// Verify mocks were called (synchronous - no sleep needed)
	mockGPGService.AssertExpectations(t)
	mockAuditService.AssertExpectations(t)
}

func TestManageGPGKeyCommand_DeleteGPGKey_NotFound_NoAuditCall(t *testing.T) {
	ctx := context.Background()
	mockGPGService := new(MockGPGKeyService)
	mockAuditService := new(MockGpgKeyAuditService)

	// Setup GPG service mock - key not found
	mockGPGService.On("DeleteGPGKey", ctx, "test-namespace", "ABC123").Return(gpgkeyModel.ErrGPGKeyNotFound)

	cmd := NewManageGPGKeyCommand(mockGPGService, mockAuditService)

	req := DeleteGPGKeyRequest{
		Namespace: "test-namespace",
		KeyID:     "ABC123",
	}

	err := cmd.DeleteGPGKey(ctx, req)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "GPG key not found")

	// Verify audit service was NOT called (since operation failed)
	mockAuditService.AssertNotCalled(t, "LogGpgKeyDelete", ctx, mock.Anything, mock.Anything)
}

func TestManageGPGKeyCommand_DeleteGPGKey_InUse_NoAuditCall(t *testing.T) {
	ctx := context.Background()
	mockGPGService := new(MockGPGKeyService)
	mockAuditService := new(MockGpgKeyAuditService)

	// Setup GPG service mock - key in use
	mockGPGService.On("DeleteGPGKey", ctx, "test-namespace", "ABC123").Return(gpgkeyModel.ErrGPGKeyInUse)

	cmd := NewManageGPGKeyCommand(mockGPGService, mockAuditService)

	req := DeleteGPGKeyRequest{
		Namespace: "test-namespace",
		KeyID:     "ABC123",
	}

	err := cmd.DeleteGPGKey(ctx, req)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "cannot delete GPG key that is in use")

	// Verify audit service was NOT called (since operation failed)
	mockAuditService.AssertNotCalled(t, "LogGpgKeyDelete", ctx, mock.Anything, mock.Anything)
}
