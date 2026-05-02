package user_group

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/auth"
	"github.com/matthewjohn/terrareg/terrareg-go/test/testutils/mocks"
)

func TestDeleteUserGroupCommand_Success(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(mocks.MockUserGroupRepository)

	groupName := "testgroup"
	existingGroup := &auth.UserGroup{
		ID:        1,
		Name:      groupName,
		SiteAdmin: true,
	}

	// Set up mock expectations
	mockRepo.On("FindByName", ctx, groupName).Return(existingGroup, nil).Once()
	mockRepo.On("Delete", ctx, existingGroup.ID).Return(nil).Once()

	// Create command
	command := NewDeleteUserGroupCommand(mockRepo, nil)

	// Execute
	err := command.Execute(ctx, groupName)

	// Assert
	require.NoError(t, err)

	// Verify mocks were called
	mockRepo.AssertExpectations(t)
}

func TestDeleteUserGroupCommand_NotFound(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(mocks.MockUserGroupRepository)

	groupName := "nonexistent"

	// Set up mock expectations
	mockRepo.On("FindByName", ctx, groupName).Return((*auth.UserGroup)(nil), nil).Once()

	// Create command
	command := NewDeleteUserGroupCommand(mockRepo, nil)

	// Execute
	err := command.Execute(ctx, groupName)

	// Assert
	assert.Error(t, err)
	assert.True(t, errors.Is(err, ErrUserGroupNotFound))

	// Verify mocks were called
	mockRepo.AssertExpectations(t)

	// Verify Delete was not called
	mockRepo.AssertNotCalled(t, "Delete", mock.Anything, mock.Anything)
}

func TestDeleteUserGroupCommand_FindByNameError(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(mocks.MockUserGroupRepository)

	groupName := "testgroup"

	// Set up mock expectations - repository returns error
	mockRepo.On("FindByName", ctx, groupName).Return((*auth.UserGroup)(nil), errors.New("database error")).Once()

	// Create command
	command := NewDeleteUserGroupCommand(mockRepo, nil)

	// Execute
	err := command.Execute(ctx, groupName)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to find user group")

	// Verify mocks were called
	mockRepo.AssertExpectations(t)

	// Verify Delete was not called
	mockRepo.AssertNotCalled(t, "Delete", mock.Anything, mock.Anything)
}

func TestDeleteUserGroupCommand_DeleteError(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(mocks.MockUserGroupRepository)

	groupName := "testgroup"
	existingGroup := &auth.UserGroup{
		ID:        1,
		Name:      groupName,
		SiteAdmin: true,
	}

	// Set up mock expectations
	mockRepo.On("FindByName", ctx, groupName).Return(existingGroup, nil).Once()
	mockRepo.On("Delete", ctx, existingGroup.ID).Return(errors.New("database error")).Once()

	// Create command
	command := NewDeleteUserGroupCommand(mockRepo, nil)

	// Execute
	err := command.Execute(ctx, groupName)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to delete user group")

	// Verify mocks were called
	mockRepo.AssertExpectations(t)
}

func TestDeleteUserGroupCommand_DeleteSiteAdminGroup(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(mocks.MockUserGroupRepository)

	groupName := "site-admins"
	existingGroup := &auth.UserGroup{
		ID:        1,
		Name:      groupName,
		SiteAdmin: true,
	}

	// Set up mock expectations
	mockRepo.On("FindByName", ctx, groupName).Return(existingGroup, nil).Once()
	mockRepo.On("Delete", ctx, existingGroup.ID).Return(nil).Once()

	// Create command
	command := NewDeleteUserGroupCommand(mockRepo, nil)

	// Execute
	err := command.Execute(ctx, groupName)

	// Assert
	require.NoError(t, err)

	// Verify mocks were called
	mockRepo.AssertExpectations(t)
}

func TestDeleteUserGroupCommand_DeleteNonSiteAdminGroup(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(mocks.MockUserGroupRepository)

	groupName := "regular-users"
	existingGroup := &auth.UserGroup{
		ID:        2,
		Name:      groupName,
		SiteAdmin: false,
	}

	// Set up mock expectations
	mockRepo.On("FindByName", ctx, groupName).Return(existingGroup, nil).Once()
	mockRepo.On("Delete", ctx, existingGroup.ID).Return(nil).Once()

	// Create command
	command := NewDeleteUserGroupCommand(mockRepo, nil)

	// Execute
	err := command.Execute(ctx, groupName)

	// Assert
	require.NoError(t, err)

	// Verify mocks were called
	mockRepo.AssertExpectations(t)
}
