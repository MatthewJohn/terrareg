package mocks

import (
	"context"

	auditservice "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/audit/service"
	"github.com/stretchr/testify/mock"
)

// MockProviderAuditService is a mock for ProviderAuditServiceInterface
// It provides thread-safe mocking for audit service methods
type MockProviderAuditService struct {
	mock.Mock
}

// Ensure MockProviderAuditService implements the interface at compile time
var _ auditservice.ProviderAuditServiceInterface = (*MockProviderAuditService)(nil)

// LogProviderCreate mocks the method
func (m *MockProviderAuditService) LogProviderCreate(ctx context.Context, providerName, namespace string) error {
	args := m.Called(ctx, providerName, namespace)
	return args.Error(0)
}

// LogProviderDelete mocks the method
func (m *MockProviderAuditService) LogProviderDelete(ctx context.Context, providerName, namespace string) error {
	args := m.Called(ctx, providerName, namespace)
	return args.Error(0)
}

// LogProviderVersionIndex mocks the method
func (m *MockProviderAuditService) LogProviderVersionIndex(ctx context.Context, providerName, namespace, version string) error {
	args := m.Called(ctx, providerName, namespace, version)
	return args.Error(0)
}

// LogProviderVersionDelete mocks the method
func (m *MockProviderAuditService) LogProviderVersionDelete(ctx context.Context, providerName, namespace, version string) error {
	args := m.Called(ctx, providerName, namespace, version)
	return args.Error(0)
}
