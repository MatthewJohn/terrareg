package mocks

import (
	"context"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/auth"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/auth/repository"
	"github.com/stretchr/testify/mock"
)

// MockUserGroupRepository is a mock for UserGroupRepository
type MockUserGroupRepository struct {
	mock.Mock
}

// Ensure MockUserGroupRepository implements the interface at compile time
var _ repository.UserGroupRepository = (*MockUserGroupRepository)(nil)

// Save mocks saving a user group
func (m *MockUserGroupRepository) Save(ctx context.Context, userGroup *auth.UserGroup) error {
	args := m.Called(ctx, userGroup)
	// Set ID on the user group if provided
	if id, ok := args.Get(0).(int); ok && id > 0 {
		userGroup.ID = id
	}
	return args.Error(1)
}

// FindByID mocks finding a user group by ID
func (m *MockUserGroupRepository) FindByID(ctx context.Context, id int) (*auth.UserGroup, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*auth.UserGroup), args.Error(1)
}

// FindByName mocks finding a user group by name
func (m *MockUserGroupRepository) FindByName(ctx context.Context, name string) (*auth.UserGroup, error) {
	args := m.Called(ctx, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*auth.UserGroup), args.Error(1)
}

// Update mocks updating a user group
func (m *MockUserGroupRepository) Update(ctx context.Context, userGroup *auth.UserGroup) error {
	args := m.Called(ctx, userGroup)
	return args.Error(0)
}

// Delete mocks deleting a user group
func (m *MockUserGroupRepository) Delete(ctx context.Context, id int) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// List mocks listing user groups
func (m *MockUserGroupRepository) List(ctx context.Context, offset, limit int) ([]*auth.UserGroup, error) {
	args := m.Called(ctx, offset, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*auth.UserGroup), args.Error(1)
}

// Count mocks counting user groups
func (m *MockUserGroupRepository) Count(ctx context.Context) (int64, error) {
	args := m.Called(ctx)
	return args.Get(0).(int64), args.Error(1)
}

// AddNamespacePermission mocks adding a namespace permission
func (m *MockUserGroupRepository) AddNamespacePermission(ctx context.Context, userGroupID, namespaceID int, permissionType auth.PermissionType) error {
	args := m.Called(ctx, userGroupID, namespaceID, permissionType)
	return args.Error(0)
}

// RemoveNamespacePermission mocks removing a namespace permission
func (m *MockUserGroupRepository) RemoveNamespacePermission(ctx context.Context, userGroupID, namespaceID int) error {
	args := m.Called(ctx, userGroupID, namespaceID)
	return args.Error(0)
}

// HasNamespacePermission mocks checking if a user group has a namespace permission
func (m *MockUserGroupRepository) HasNamespacePermission(ctx context.Context, userGroupID, namespaceID int, permissionType auth.PermissionType) (bool, error) {
	args := m.Called(ctx, userGroupID, namespaceID, permissionType)
	return args.Bool(0), args.Error(1)
}

// GetNamespacePermissions mocks getting namespace permissions for a user group
func (m *MockUserGroupRepository) GetNamespacePermissions(ctx context.Context, userGroupID int) ([]auth.NamespacePermission, error) {
	args := m.Called(ctx, userGroupID)
	if args.Get(0) == nil {
		return []auth.NamespacePermission{}, args.Error(1)
	}
	return args.Get(0).([]auth.NamespacePermission), args.Error(1)
}

// GetHighestNamespacePermission mocks getting the highest permission for a namespace
func (m *MockUserGroupRepository) GetHighestNamespacePermission(ctx context.Context, userGroupID, namespaceID int) (auth.PermissionType, error) {
	args := m.Called(ctx, userGroupID, namespaceID)
	return args.Get(0).(auth.PermissionType), args.Error(1)
}

// FindSiteAdminGroups mocks finding site admin groups
func (m *MockUserGroupRepository) FindSiteAdminGroups(ctx context.Context) ([]*auth.UserGroup, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*auth.UserGroup), args.Error(1)
}

// FindGroupsByNamespace mocks finding groups by namespace
func (m *MockUserGroupRepository) FindGroupsByNamespace(ctx context.Context, namespaceID int) ([]*auth.UserGroup, error) {
	args := m.Called(ctx, namespaceID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*auth.UserGroup), args.Error(1)
}

// SearchByName mocks searching user groups by name
func (m *MockUserGroupRepository) SearchByName(ctx context.Context, query string, offset, limit int) ([]*auth.UserGroup, error) {
	args := m.Called(ctx, query, offset, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*auth.UserGroup), args.Error(1)
}

// GetGroupsForUser mocks getting groups for a user
func (m *MockUserGroupRepository) GetGroupsForUser(ctx context.Context, userID string) ([]*auth.UserGroup, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*auth.UserGroup), args.Error(1)
}
