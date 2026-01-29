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

func TestCreateUserGroupCommand_Success(t *testing.T) {
	tests := []struct {
		name      string
		groupName string
		siteAdmin bool
	}{
		{
			name:      "Create with site admin true",
			groupName: "testgroup",
			siteAdmin: true,
		},
		{
			name:      "Create with site admin false",
			groupName: "testgroup",
			siteAdmin: false,
		},
		{
			name:      "Create with hyphens in name",
			groupName: "test-group",
			siteAdmin: true,
		},
		{
			name:      "Create with underscores in name",
			groupName: "test_group",
			siteAdmin: false,
		},
		{
			name:      "Create with spaces in name",
			groupName: "test group",
			siteAdmin: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			mockRepo := new(mocks.MockUserGroupRepository)

			// Set up mock expectations
			mockRepo.On("FindByName", ctx, tt.groupName).Return((*auth.UserGroup)(nil), nil).Once()
			mockRepo.On("Save", ctx, mock.AnythingOfType("*auth.UserGroup")).Return(1, nil).Once()

			// Create command
			command, err := NewCreateUserGroupCommand(mockRepo, nil)
			require.NoError(t, err)

			// Execute
			siteAdmin := tt.siteAdmin
			req := CreateUserGroupRequest{
				Name:      tt.groupName,
				SiteAdmin: &siteAdmin,
			}

			result, err := command.Execute(ctx, req)

			// Assert
			require.NoError(t, err)
			require.NotNil(t, result)
			assert.Equal(t, tt.groupName, result.Name)
			assert.Equal(t, tt.siteAdmin, result.SiteAdmin)

			// Verify mocks were called
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestCreateUserGroupCommand_InvalidName(t *testing.T) {
	tests := []struct {
		name      string
		groupName string
		siteAdmin bool
	}{
		{
			name:      "Create with @ symbol",
			groupName: "test@group",
			siteAdmin: true,
		},
		{
			name:      "Create with # symbol",
			groupName: "test#group",
			siteAdmin: false,
		},
		{
			name:      "Create with quote",
			groupName: `test"group"`,
			siteAdmin: true,
		},
		{
			name:      "Create with apostrophe",
			groupName: "test'group'",
			siteAdmin: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			mockRepo := new(mocks.MockUserGroupRepository)

			// Create command
			command, err := NewCreateUserGroupCommand(mockRepo, nil)
			require.NoError(t, err)

			// Execute
			siteAdmin := tt.siteAdmin
			req := CreateUserGroupRequest{
				Name:      tt.groupName,
				SiteAdmin: &siteAdmin,
			}

			result, err := command.Execute(ctx, req)

			// Assert
			assert.Error(t, err)
			assert.Nil(t, result)
			assert.True(t, errors.Is(err, ErrInvalidUserGroupName))

			// Verify FindByName was not called
			mockRepo.AssertNotCalled(t, "FindByName", ctx, tt.groupName)
		})
	}
}

func TestCreateUserGroupCommand_SiteAdminNil(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(mocks.MockUserGroupRepository)

	// Create command
	command, err := NewCreateUserGroupCommand(mockRepo, nil)
	require.NoError(t, err)

	// Execute with nil site_admin
	req := CreateUserGroupRequest{
		Name:      "testgroup",
		SiteAdmin: nil,
	}

	result, err := command.Execute(ctx, req)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.True(t, errors.Is(err, ErrInvalidSiteAdminValue))
}

func TestCreateUserGroupCommand_AlreadyExists(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(mocks.MockUserGroupRepository)

	groupName := "existing-group"
	existingGroup := &auth.UserGroup{
		ID:        1,
		Name:      groupName,
		SiteAdmin: true,
	}

	// Set up mock expectations
	mockRepo.On("FindByName", ctx, groupName).Return(existingGroup, nil).Once()

	// Create command
	command, err := NewCreateUserGroupCommand(mockRepo, nil)
	require.NoError(t, err)

	// Execute
	siteAdmin := true
	req := CreateUserGroupRequest{
		Name:      groupName,
		SiteAdmin: &siteAdmin,
	}

	result, err := command.Execute(ctx, req)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.True(t, errors.Is(err, ErrUserGroupAlreadyExists))

	// Verify mocks were called
	mockRepo.AssertExpectations(t)
}

func TestCreateUserGroupCommand_FindByNameError(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(mocks.MockUserGroupRepository)

	groupName := "testgroup"

	// Set up mock expectations - repository returns error
	mockRepo.On("FindByName", ctx, groupName).Return((*auth.UserGroup)(nil), errors.New("database error")).Once()

	// Create command
	command, err := NewCreateUserGroupCommand(mockRepo, nil)
	require.NoError(t, err)

	// Execute
	siteAdmin := true
	req := CreateUserGroupRequest{
		Name:      groupName,
		SiteAdmin: &siteAdmin,
	}

	result, err := command.Execute(ctx, req)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to check user group existence")

	// Verify mocks were called
	mockRepo.AssertExpectations(t)
}

func TestCreateUserGroupCommand_SaveError(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(mocks.MockUserGroupRepository)

	groupName := "testgroup"

	// Set up mock expectations
	mockRepo.On("FindByName", ctx, groupName).Return((*auth.UserGroup)(nil), nil).Once()
	mockRepo.On("Save", ctx, mock.AnythingOfType("*auth.UserGroup")).Return(0, errors.New("database error")).Once()

	// Create command
	command, err := NewCreateUserGroupCommand(mockRepo, nil)
	require.NoError(t, err)

	// Execute
	siteAdmin := true
	req := CreateUserGroupRequest{
		Name:      groupName,
		SiteAdmin: &siteAdmin,
	}

	result, err := command.Execute(ctx, req)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to save user group")

	// Verify mocks were called
	mockRepo.AssertExpectations(t)
}

func TestCreateUserGroupCommand_Success_CallsAuditService(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(mocks.MockUserGroupRepository)
	mockAuditService := new(mocks.MockUserGroupAuditService)

	groupName := "testgroup"

	// Set up mock expectations
	mockRepo.On("FindByName", ctx, groupName).Return((*auth.UserGroup)(nil), nil).Once()
	mockRepo.On("Save", ctx, mock.AnythingOfType("*auth.UserGroup")).Return(1, nil).Once()

	// Set up audit service mock - expect LogUserGroupCreate call
	mockAuditService.On("LogUserGroupCreate", ctx, groupName).Return(nil)

	// Create command with audit service
	command, err := NewCreateUserGroupCommand(mockRepo, mockAuditService)
	require.NoError(t, err)

	// Execute
	siteAdmin := true
	req := CreateUserGroupRequest{
		Name:      groupName,
		SiteAdmin: &siteAdmin,
	}

	result, err := command.Execute(ctx, req)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, groupName, result.Name)

	// Verify mocks were called (synchronous - no sleep needed)
	mockRepo.AssertExpectations(t)
	mockAuditService.AssertExpectations(t)
}

func TestCreateUserGroupCommand_SaveError_NoAuditCall(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(mocks.MockUserGroupRepository)
	mockAuditService := new(mocks.MockUserGroupAuditService)

	groupName := "testgroup"

	// Set up mock expectations - save returns error
	mockRepo.On("FindByName", ctx, groupName).Return((*auth.UserGroup)(nil), nil).Once()
	mockRepo.On("Save", ctx, mock.AnythingOfType("*auth.UserGroup")).Return(0, errors.New("database error")).Once()

	// Create command with audit service
	command, err := NewCreateUserGroupCommand(mockRepo, mockAuditService)
	require.NoError(t, err)

	// Execute
	siteAdmin := true
	req := CreateUserGroupRequest{
		Name:      groupName,
		SiteAdmin: &siteAdmin,
	}

	result, err := command.Execute(ctx, req)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)

	// Verify audit service was NOT called (since operation failed)
	mockAuditService.AssertNotCalled(t, "LogUserGroupCreate", ctx, mock.Anything)
}
