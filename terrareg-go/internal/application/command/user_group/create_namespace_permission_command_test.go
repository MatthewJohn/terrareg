package user_group

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/auth"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/shared/types"
	"github.com/matthewjohn/terrareg/terrareg-go/test/testutils/mocks"
)

func TestCreateUserGroupNamespacePermissionCommand_Success(t *testing.T) {
	tests := []struct {
		name           string
		permissionType string
	}{
		{
			name:           "Create FULL permission",
			permissionType: "FULL",
		},
		{
			name:           "Create MODIFY permission",
			permissionType: "MODIFY",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			mockUserGroupRepo := new(mocks.MockUserGroupRepository)
			mockNamespaceRepo := new(mocks.MockNamespaceRepository)

			userGroupName := "testgroup"
			namespaceName := types.NamespaceName("testnamespace")

			// Set up mock expectations for namespace
			mockNamespaceRepo.On("FindByName", ctx, types.NamespaceName(namespaceName)).Return(createMockNamespace(1, namespaceName), nil).Once()

			// Set up mock expectations for user group
			userGroup := &auth.UserGroup{
				ID:        1,
				Name:      userGroupName,
				SiteAdmin: false,
			}
			mockUserGroupRepo.On("FindByName", ctx, userGroupName).Return(userGroup, nil).Once()

			// Set up mock for getting existing permissions (none exist)
			mockUserGroupRepo.On("GetNamespacePermissions", ctx, userGroup.ID).Return([]auth.NamespacePermission{}, nil).Once()

			// Set up mock for adding permission
			mockUserGroupRepo.On("AddNamespacePermission", ctx, userGroup.ID, mock.AnythingOfType("int"), mock.AnythingOfType("auth.PermissionType")).Return(nil).Once()

			// Create command
			command := NewCreateUserGroupNamespacePermissionCommand(mockUserGroupRepo, mockNamespaceRepo)

			// Execute
			req := CreateNamespacePermissionRequest{
				PermissionType: tt.permissionType,
			}

			result, err := command.Execute(ctx, userGroupName, namespaceName, req)

			// Assert
			require.NoError(t, err)
			require.NotNil(t, result)
			assert.Equal(t, userGroupName, result.UserGroup)
			assert.Equal(t, namespaceName, result.Namespace)
			assert.Equal(t, tt.permissionType, result.PermissionType)

			// Verify mocks were called
			mockUserGroupRepo.AssertExpectations(t)
			mockNamespaceRepo.AssertExpectations(t)
		})
	}
}

func TestCreateUserGroupNamespacePermissionCommand_InvalidPermissionType(t *testing.T) {
	tests := []struct {
		name           string
		permissionType string
	}{
		{
			name:           "Invalid permission type READ",
			permissionType: "READ",
		},
		{
			name:           "Invalid permission type INVALID",
			permissionType: "INVALID",
		},
		{
			name:           "Invalid permission type empty",
			permissionType: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			mockUserGroupRepo := new(mocks.MockUserGroupRepository)
			mockNamespaceRepo := new(mocks.MockNamespaceRepository)

			// Create command
			command := NewCreateUserGroupNamespacePermissionCommand(mockUserGroupRepo, mockNamespaceRepo)

			// Execute
			req := CreateNamespacePermissionRequest{
				PermissionType: tt.permissionType,
			}

			result, err := command.Execute(ctx, "testgroup", "testnamespace", req)

			// Assert
			assert.Error(t, err)
			assert.Nil(t, result)
			assert.True(t, errors.Is(err, ErrInvalidPermissionType))

			// Verify no repository methods were called
			mockNamespaceRepo.AssertNotCalled(t, "FindByName", mock.Anything, mock.Anything)
			mockUserGroupRepo.AssertNotCalled(t, "FindByName", mock.Anything, mock.Anything)
		})
	}
}

func TestCreateUserGroupNamespacePermissionCommand_NamespaceNotFound(t *testing.T) {
	ctx := context.Background()
	mockUserGroupRepo := new(mocks.MockUserGroupRepository)
	mockNamespaceRepo := new(mocks.MockNamespaceRepository)

	namespaceName := types.NamespaceName("nonexistent")

	// Set up mock expectations - namespace not found
	mockNamespaceRepo.On("FindByName", ctx, types.NamespaceName(namespaceName)).Return(nil, nil).Once()

	// Create command
	command := NewCreateUserGroupNamespacePermissionCommand(mockUserGroupRepo, mockNamespaceRepo)

	// Execute
	req := CreateNamespacePermissionRequest{
		PermissionType: "FULL",
	}

	result, err := command.Execute(ctx, "testgroup", namespaceName, req)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.True(t, errors.Is(err, ErrNamespaceNotFound))

	// Verify mocks were called
	mockNamespaceRepo.AssertExpectations(t)

	// Verify user group repository was not called
	mockUserGroupRepo.AssertNotCalled(t, "FindByName", mock.Anything, mock.Anything)
}

func TestCreateUserGroupNamespacePermissionCommand_UserGroupNotFound(t *testing.T) {
	ctx := context.Background()
	mockUserGroupRepo := new(mocks.MockUserGroupRepository)
	mockNamespaceRepo := new(mocks.MockNamespaceRepository)

	userGroupName := "nonexistent"
	namespaceName := types.NamespaceName("testnamespace")

	// Set up mock expectations for namespace (found)
	mockNamespaceRepo.On("FindByName", ctx, types.NamespaceName(namespaceName)).Return(createMockNamespace(1, namespaceName), nil).Once()

	// Set up mock expectations for user group (not found)
	mockUserGroupRepo.On("FindByName", ctx, userGroupName).Return((*auth.UserGroup)(nil), nil).Once()

	// Create command
	command := NewCreateUserGroupNamespacePermissionCommand(mockUserGroupRepo, mockNamespaceRepo)

	// Execute
	req := CreateNamespacePermissionRequest{
		PermissionType: "FULL",
	}

	result, err := command.Execute(ctx, userGroupName, namespaceName, req)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.True(t, errors.Is(err, ErrUserGroupNotFound))

	// Verify mocks were called
	mockUserGroupRepo.AssertExpectations(t)
	mockNamespaceRepo.AssertExpectations(t)
}

func TestCreateUserGroupNamespacePermissionCommand_PermissionAlreadyExists(t *testing.T) {
	ctx := context.Background()
	mockUserGroupRepo := new(mocks.MockUserGroupRepository)
	mockNamespaceRepo := new(mocks.MockNamespaceRepository)

	userGroupName := "testgroup"
	namespaceName := types.NamespaceName("testnamespace")

	// Set up mock expectations for namespace
	mockNamespaceRepo.On("FindByName", ctx, types.NamespaceName(namespaceName)).Return(createMockNamespace(1, namespaceName), nil).Once()

	// Set up mock expectations for user group
	userGroup := &auth.UserGroup{
		ID:        1,
		Name:      userGroupName,
		SiteAdmin: false,
	}
	mockUserGroupRepo.On("FindByName", ctx, userGroupName).Return(userGroup, nil).Once()

	// Set up mock for getting existing permissions (permission already exists)
	existingPermissions := []auth.NamespacePermission{
		{
			UserGroupID:    userGroup.ID,
			NamespaceID:    1,
			PermissionType: auth.PermissionFull,
		},
	}
	mockUserGroupRepo.On("GetNamespacePermissions", ctx, userGroup.ID).Return(existingPermissions, nil).Once()

	// Create command
	command := NewCreateUserGroupNamespacePermissionCommand(mockUserGroupRepo, mockNamespaceRepo)

	// Execute
	req := CreateNamespacePermissionRequest{
		PermissionType: "FULL",
	}

	result, err := command.Execute(ctx, userGroupName, namespaceName, req)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.True(t, errors.Is(err, ErrPermissionAlreadyExists))

	// Verify mocks were called
	mockUserGroupRepo.AssertExpectations(t)
	mockNamespaceRepo.AssertExpectations(t)

	// Verify AddNamespacePermission was not called
	mockUserGroupRepo.AssertNotCalled(t, "AddNamespacePermission", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
}

func TestCreateUserGroupNamespacePermissionCommand_NamespaceFindError(t *testing.T) {
	ctx := context.Background()
	mockUserGroupRepo := new(mocks.MockUserGroupRepository)
	mockNamespaceRepo := new(mocks.MockNamespaceRepository)

	namespaceName := types.NamespaceName("testnamespace")

	// Set up mock expectations - namespace repository returns error
	mockNamespaceRepo.On("FindByName", ctx, types.NamespaceName(namespaceName)).Return(nil, errors.New("database error")).Once()

	// Create command
	command := NewCreateUserGroupNamespacePermissionCommand(mockUserGroupRepo, mockNamespaceRepo)

	// Execute
	req := CreateNamespacePermissionRequest{
		PermissionType: "FULL",
	}

	result, err := command.Execute(ctx, "testgroup", namespaceName, req)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to find namespace")

	// Verify mocks were called
	mockNamespaceRepo.AssertExpectations(t)
}

func TestCreateUserGroupNamespacePermissionCommand_AddPermissionError(t *testing.T) {
	ctx := context.Background()
	mockUserGroupRepo := new(mocks.MockUserGroupRepository)
	mockNamespaceRepo := new(mocks.MockNamespaceRepository)

	userGroupName := "testgroup"
	namespaceName := types.NamespaceName("testnamespace")

	// Set up mock expectations for namespace
	mockNamespaceRepo.On("FindByName", ctx, types.NamespaceName(namespaceName)).Return(createMockNamespace(1, namespaceName), nil).Once()

	// Set up mock expectations for user group
	userGroup := &auth.UserGroup{
		ID:        1,
		Name:      userGroupName,
		SiteAdmin: false,
	}
	mockUserGroupRepo.On("FindByName", ctx, userGroupName).Return(userGroup, nil).Once()

	// Set up mock for getting existing permissions (none exist)
	mockUserGroupRepo.On("GetNamespacePermissions", ctx, userGroup.ID).Return([]auth.NamespacePermission{}, nil).Once()

	// Set up mock for adding permission - returns error
	mockUserGroupRepo.On("AddNamespacePermission", ctx, userGroup.ID, mock.AnythingOfType("int"), mock.AnythingOfType("auth.PermissionType")).Return(errors.New("database error")).Once()

	// Create command
	command := NewCreateUserGroupNamespacePermissionCommand(mockUserGroupRepo, mockNamespaceRepo)

	// Execute
	req := CreateNamespacePermissionRequest{
		PermissionType: "FULL",
	}

	result, err := command.Execute(ctx, userGroupName, namespaceName, req)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to create namespace permission")

	// Verify mocks were called
	mockUserGroupRepo.AssertExpectations(t)
	mockNamespaceRepo.AssertExpectations(t)
}
