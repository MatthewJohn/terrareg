package mocks

import (
	"context"

	auditservice "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/audit/service"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/shared/types"
	"github.com/stretchr/testify/mock"
)

// MockModuleAuditService is a mock for ModuleAuditServiceInterface
// It provides thread-safe mocking for audit service methods
type MockModuleAuditService struct {
	mock.Mock
}

// Ensure MockModuleAuditService implements the interface at compile time
var _ auditservice.ModuleAuditServiceInterface = (*MockModuleAuditService)(nil)

// LogModuleVersionIndex mocks the method
func (m *MockModuleAuditService) LogModuleVersionIndex(ctx context.Context, username types.NamespaceName, namespace types.NamespaceName, module types.ModuleName, provider types.ModuleProviderName, version types.ModuleVersion) error {
	args := m.Called(ctx, username, namespace, module, provider, version)
	return args.Error(0)
}

// LogModuleVersionPublish mocks the method
func (m *MockModuleAuditService) LogModuleVersionPublish(ctx context.Context, username types.NamespaceName, namespace types.NamespaceName, module types.ModuleName, provider types.ModuleProviderName, version types.ModuleVersion) error {
	args := m.Called(ctx, username, namespace, module, provider, version)
	return args.Error(0)
}

// LogModuleVersionDelete mocks the method
func (m *MockModuleAuditService) LogModuleVersionDelete(ctx context.Context, username types.NamespaceName, namespace types.NamespaceName, module types.ModuleName, provider types.ModuleProviderName, version types.ModuleVersion) error {
	args := m.Called(ctx, username, namespace, module, provider, version)
	return args.Error(0)
}

// LogModuleProviderCreate mocks the method
func (m *MockModuleAuditService) LogModuleProviderCreate(ctx context.Context, username types.NamespaceName, namespace types.NamespaceName, module types.ModuleName, provider types.ModuleProviderName) error {
	args := m.Called(ctx, username, namespace, module, provider)
	return args.Error(0)
}

// LogModuleProviderDelete mocks the method
func (m *MockModuleAuditService) LogModuleProviderDelete(ctx context.Context, username types.NamespaceName, namespace types.NamespaceName, module types.ModuleName, provider types.ModuleProviderName) error {
	args := m.Called(ctx, username, namespace, module, provider)
	return args.Error(0)
}
