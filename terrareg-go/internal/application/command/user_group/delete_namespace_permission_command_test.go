package user_group

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/auth"
	types "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/shared/types"
	"github.com/matthewjohn/terrareg/terrareg-go/test/testutils/mocks"
)

func TestDeleteUserGroupNamespacePermissionCommand_Success(t *testing.T) {
	tests := []struct {
		name           string
		permissionType auth.PermissionType
	}{
		{
			name:           "Delete FULL permission",
			permissionType: auth.PermissionFull,
		},
		{
			name:           "Delete MODIFY permission",
			permissionType: auth.PermissionModify,
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
			mockNamespaceRepo.On("FindByName", ctx, namespaceName).Return(createMockNamespace(1, namespaceName), nil).Once()

			// Set up mock expectations for user group
			userGroup := &auth.UserGroup{
				ID:        1,
				Name:      userGroupName,
				SiteAdmin: false,
			}
			mockUserGroupRepo.On("FindByName", ctx, userGroupName).Return(userGroup, nil).Once()

			// Set up mock for getting existing permissions (permission exists)
			existingPermissions := []auth.NamespacePermission{
				{
					UserGroupID:    userGroup.ID,
					NamespaceID:    1,
					PermissionType: tt.permissionType,
				},
			}
			mockUserGroupRepo.On("GetNamespacePermissions", ctx, userGroup.ID).Return(existingPermissions, nil).Once()

			// Set up mock for removing permission
			mockUserGroupRepo.On("RemoveNamespacePermission", ctx, userGroup.ID, mock.AnythingOfType("int")).Return(nil).Once()

			// Create command
			command := NewDeleteUserGroupNamespacePermissionCommand(mockUserGroupRepo, mockNamespaceRepo)

			// Execute
			err := command.Execute(ctx, userGroupName, namespaceName)

			// Assert
			require.NoError(t, err)

			// Verify mocks were called
			mockUserGroupRepo.AssertExpectations(t)
			mockNamespaceRepo.AssertExpectations(t)
		})
	}
}

func TestDeleteUserGroupNamespacePermissionCommand_NamespaceNotFound(t *testing.T) {
	ctx := context.Background()
	mockUserGroupRepo := new(mocks.MockUserGroupRepository)
	mockNamespaceRepo := new(mocks.MockNamespaceRepository)

	namespaceName := types.NamespaceName("nonexistent")

	// Set up mock expectations - namespace not found
	mockNamespaceRepo.On("FindByName", ctx, namespaceName).Return(nil, nil).Once()

	// Create command
	command := NewDeleteUserGroupNamespacePermissionCommand(mockUserGroupRepo, mockNamespaceRepo)

	// Execute
	err := command.Execute(ctx, "testgroup", namespaceName)

	// Assert
	assert.Error(t, err)
	assert.True(t, errors.Is(err, ErrNamespaceNotFound))

	// Verify mocks were called
	mockNamespaceRepo.AssertExpectations(t)

	// Verify user group repository was not called
	mockUserGroupRepo.AssertNotCalled(t, "FindByName", mock.Anything, mock.Anything)
}

func TestDeleteUserGroupNamespacePermissionCommand_UserGroupNotFound(t *testing.T) {
	ctx := context.Background()
	mockUserGroupRepo := new(mocks.MockUserGroupRepository)
	mockNamespaceRepo := new(mocks.MockNamespaceRepository)

	userGroupName := "nonexistent"
	namespaceName := types.NamespaceName("testnamespace")

	// Set up mock expectations for namespace (found)
	mockNamespaceRepo.On("FindByName", ctx, namespaceName).Return(createMockNamespace(1, namespaceName), nil).Once()

	// Set up mock expectations for user group (not found)
	mockUserGroupRepo.On("FindByName", ctx, userGroupName).Return((*auth.UserGroup)(nil), nil).Once()

	// Create command
	command := NewDeleteUserGroupNamespacePermissionCommand(mockUserGroupRepo, mockNamespaceRepo)

	// Execute
	err := command.Execute(ctx, userGroupName, namespaceName)

	// Assert
	assert.Error(t, err)
	assert.True(t, errors.Is(err, ErrUserGroupNotFound))

	// Verify mocks were called
	mockUserGroupRepo.AssertExpectations(t)
	mockNamespaceRepo.AssertExpectations(t)

	// Verify RemoveNamespacePermission was not called
	mockUserGroupRepo.AssertNotCalled(t, "RemoveNamespacePermission", mock.Anything, mock.Anything, mock.Anything)
}

func TestDeleteUserGroupNamespacePermissionCommand_PermissionNotFound(t *testing.T) {
	ctx := context.Background()
	mockUserGroupRepo := new(mocks.MockUserGroupRepository)
	mockNamespaceRepo := new(mocks.MockNamespaceRepository)

	userGroupName := "testgroup"
	namespaceName := types.NamespaceName("testnamespace")

	// Set up mock expectations for namespace
	mockNamespaceRepo.On("FindByName", ctx, namespaceName).Return(createMockNamespace(1, namespaceName), nil).Once()

	// Set up mock expectations for user group
	userGroup := &auth.UserGroup{
		ID:        1,
		Name:      userGroupName,
		SiteAdmin: false,
	}
	mockUserGroupRepo.On("FindByName", ctx, userGroupName).Return(userGroup, nil).Once()

	// Set up mock for getting existing permissions (no permission for this namespace)
	existingPermissions := []auth.NamespacePermission{
		{
			UserGroupID:    userGroup.ID,
			NamespaceID:    2, // Different namespace
			PermissionType: auth.PermissionFull,
		},
	}
	mockUserGroupRepo.On("GetNamespacePermissions", ctx, userGroup.ID).Return(existingPermissions, nil).Once()

	// Create command
	command := NewDeleteUserGroupNamespacePermissionCommand(mockUserGroupRepo, mockNamespaceRepo)

	// Execute
	err := command.Execute(ctx, userGroupName, namespaceName)

	// Assert
	assert.Error(t, err)
	assert.True(t, errors.Is(err, ErrPermissionNotFound))

	// Verify mocks were called
	mockUserGroupRepo.AssertExpectations(t)
	mockNamespaceRepo.AssertExpectations(t)

	// Verify RemoveNamespacePermission was not called
	mockUserGroupRepo.AssertNotCalled(t, "RemoveNamespacePermission", mock.Anything, mock.Anything, mock.Anything)
}

func TestDeleteUserGroupNamespacePermissionCommand_NamespaceFindError(t *testing.T) {
	ctx := context.Background()
	mockUserGroupRepo := new(mocks.MockUserGroupRepository)
	mockNamespaceRepo := new(mocks.MockNamespaceRepository)

	namespaceName := types.NamespaceName("testnamespace")

	// Set up mock expectations - namespace repository returns error
	mockNamespaceRepo.On("FindByName", ctx, namespaceName).Return(nil, errors.New("database error")).Once()

	// Create command
	command := NewDeleteUserGroupNamespacePermissionCommand(mockUserGroupRepo, mockNamespaceRepo)

	// Execute
	err := command.Execute(ctx, "testgroup", namespaceName)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to find namespace")

	// Verify mocks were called
	mockNamespaceRepo.AssertExpectations(t)
}

func TestDeleteUserGroupNamespacePermissionCommand_UserGroupFindError(t *testing.T) {
	ctx := context.Background()
	mockUserGroupRepo := new(mocks.MockUserGroupRepository)
	mockNamespaceRepo := new(mocks.MockNamespaceRepository)

	userGroupName := "testgroup"
	namespaceName := types.NamespaceName("testnamespace")

	// Set up mock expectations for namespace (found)
	mockNamespaceRepo.On("FindByName", ctx, namespaceName).Return(createMockNamespace(1, namespaceName), nil).Once()

	// Set up mock expectations for user group - returns error
	mockUserGroupRepo.On("FindByName", ctx, userGroupName).Return((*auth.UserGroup)(nil), errors.New("database error")).Once()

	// Create command
	command := NewDeleteUserGroupNamespacePermissionCommand(mockUserGroupRepo, mockNamespaceRepo)

	// Execute
	err := command.Execute(ctx, userGroupName, namespaceName)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to find user group")

	// Verify mocks were called
	mockUserGroupRepo.AssertExpectations(t)
	mockNamespaceRepo.AssertExpectations(t)
}

func TestDeleteUserGroupNamespacePermissionCommand_GetPermissionsError(t *testing.T) {
	ctx := context.Background()
	mockUserGroupRepo := new(mocks.MockUserGroupRepository)
	mockNamespaceRepo := new(mocks.MockNamespaceRepository)

	userGroupName := "testgroup"
	namespaceName := types.NamespaceName("testnamespace")

	// Set up mock expectations for namespace
	mockNamespaceRepo.On("FindByName", ctx, namespaceName).Return(createMockNamespace(1, namespaceName), nil).Once()

	// Set up mock expectations for user group
	userGroup := &auth.UserGroup{
		ID:        1,
		Name:      userGroupName,
		SiteAdmin: false,
	}
	mockUserGroupRepo.On("FindByName", ctx, userGroupName).Return(userGroup, nil).Once()

	// Set up mock for getting existing permissions - returns error
	mockUserGroupRepo.On("GetNamespacePermissions", ctx, userGroup.ID).Return([]auth.NamespacePermission{}, errors.New("database error")).Once()

	// Create command
	command := NewDeleteUserGroupNamespacePermissionCommand(mockUserGroupRepo, mockNamespaceRepo)

	// Execute
	err := command.Execute(ctx, userGroupName, namespaceName)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to check existing permissions")

	// Verify mocks were called
	mockUserGroupRepo.AssertExpectations(t)
	mockNamespaceRepo.AssertExpectations(t)

	// Verify RemoveNamespacePermission was not called
	mockUserGroupRepo.AssertNotCalled(t, "RemoveNamespacePermission", mock.Anything, mock.Anything, mock.Anything)
}

func TestDeleteUserGroupNamespacePermissionCommand_RemovePermissionError(t *testing.T) {
	ctx := context.Background()
	mockUserGroupRepo := new(mocks.MockUserGroupRepository)
	mockNamespaceRepo := new(mocks.MockNamespaceRepository)

	userGroupName := "testgroup"
	namespaceName := types.NamespaceName("testnamespace")

	// Set up mock expectations for namespace
	mockNamespaceRepo.On("FindByName", ctx, namespaceName).Return(createMockNamespace(1, namespaceName), nil).Once()

	// Set up mock expectations for user group
	userGroup := &auth.UserGroup{
		ID:        1,
		Name:      userGroupName,
		SiteAdmin: false,
	}
	mockUserGroupRepo.On("FindByName", ctx, userGroupName).Return(userGroup, nil).Once()

	// Set up mock for getting existing permissions (permission exists)
	existingPermissions := []auth.NamespacePermission{
		{
			UserGroupID:    userGroup.ID,
			NamespaceID:    1,
			PermissionType: auth.PermissionFull,
		},
	}
	mockUserGroupRepo.On("GetNamespacePermissions", ctx, userGroup.ID).Return(existingPermissions, nil).Once()

	// Set up mock for removing permission - returns error
	mockUserGroupRepo.On("RemoveNamespacePermission", ctx, userGroup.ID, mock.AnythingOfType("int")).Return(errors.New("database error")).Once()

	// Create command
	command := NewDeleteUserGroupNamespacePermissionCommand(mockUserGroupRepo, mockNamespaceRepo)

	// Execute
	err := command.Execute(ctx, userGroupName, namespaceName)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to delete namespace permission")

	// Verify mocks were called
	mockUserGroupRepo.AssertExpectations(t)
	mockNamespaceRepo.AssertExpectations(t)
}

func TestDeleteUserGroupNamespacePermissionCommand_MultiplePermissionsForSameNamespace(t *testing.T) {
	ctx := context.Background()
	mockUserGroupRepo := new(mocks.MockUserGroupRepository)
	mockNamespaceRepo := new(mocks.MockNamespaceRepository)

	userGroupName := "testgroup"
	namespaceName := types.NamespaceName("testnamespace")

	// Set up mock expectations for namespace
	mockNamespaceRepo.On("FindByName", ctx, namespaceName).Return(createMockNamespace(1, namespaceName), nil).Once()

	// Set up mock expectations for user group
	userGroup := &auth.UserGroup{
		ID:        1,
		Name:      userGroupName,
		SiteAdmin: false,
	}
	mockUserGroupRepo.On("FindByName", ctx, userGroupName).Return(userGroup, nil).Once()

	// Set up mock for getting existing permissions (multiple permissions for same namespace - though this shouldn't happen in practice)
	existingPermissions := []auth.NamespacePermission{
		{
			UserGroupID:    userGroup.ID,
			NamespaceID:    1,
			PermissionType: auth.PermissionFull,
		},
		{
			UserGroupID:    userGroup.ID,
			NamespaceID:    1,
			PermissionType: auth.PermissionModify, // This shouldn't happen, but testing anyway
		},
	}
	mockUserGroupRepo.On("GetNamespacePermissions", ctx, userGroup.ID).Return(existingPermissions, nil).Once()

	// Set up mock for removing permission
	mockUserGroupRepo.On("RemoveNamespacePermission", ctx, userGroup.ID, mock.AnythingOfType("int")).Return(nil).Once()

	// Create command
	command := NewDeleteUserGroupNamespacePermissionCommand(mockUserGroupRepo, mockNamespaceRepo)

	// Execute
	err := command.Execute(ctx, userGroupName, namespaceName)

	// Assert
	require.NoError(t, err)

	// Verify mocks were called
	mockUserGroupRepo.AssertExpectations(t)
	mockNamespaceRepo.AssertExpectations(t)
}
