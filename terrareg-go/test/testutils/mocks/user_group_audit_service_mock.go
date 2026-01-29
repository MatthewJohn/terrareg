package mocks

import (
	"context"

	auditservice "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/audit/service"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/shared/types"
	"github.com/stretchr/testify/mock"
)

// MockUserGroupAuditService is a mock for UserGroupAuditServiceInterface
// It provides thread-safe mocking for audit service methods
type MockUserGroupAuditService struct {
	mock.Mock
}

// Ensure MockUserGroupAuditService implements the interface at compile time
var _ auditservice.UserGroupAuditServiceInterface = (*MockUserGroupAuditService)(nil)

// LogUserGroupCreate mocks the method
func (m *MockUserGroupAuditService) LogUserGroupCreate(ctx context.Context, groupName string) error {
	args := m.Called(ctx, groupName)
	return args.Error(0)
}

// LogUserGroupDelete mocks the method
func (m *MockUserGroupAuditService) LogUserGroupDelete(ctx context.Context, groupName string) error {
	args := m.Called(ctx, groupName)
	return args.Error(0)
}

// LogUserGroupNamespacePermissionAdd mocks the method
func (m *MockUserGroupAuditService) LogUserGroupNamespacePermissionAdd(ctx context.Context, groupName string, namespace types.NamespaceName, permissionType string) error {
	args := m.Called(ctx, groupName, namespace, permissionType)
	return args.Error(0)
}

// LogUserGroupNamespacePermissionModify mocks the method
func (m *MockUserGroupAuditService) LogUserGroupNamespacePermissionModify(ctx context.Context, groupName string, namespace types.NamespaceName, oldPermission, newPermission string) error {
	args := m.Called(ctx, groupName, namespace, oldPermission, newPermission)
	return args.Error(0)
}

// LogUserGroupNamespacePermissionDelete mocks the method
func (m *MockUserGroupAuditService) LogUserGroupNamespacePermissionDelete(ctx context.Context, groupName string, namespace types.NamespaceName, permissionType string) error {
	args := m.Called(ctx, groupName, namespace, permissionType)
	return args.Error(0)
}
