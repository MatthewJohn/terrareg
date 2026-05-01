package user_group

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/auth"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/shared/types"
	"github.com/matthewjohn/terrareg/terrareg-go/test/testutils/mocks"
)

func TestListUserGroupsQuery_Success(t *testing.T) {
	tests := []struct {
		name          string
		userGroups    []*auth.UserGroup
		permissions   map[int][]auth.NamespacePermission
		expectedCount int
	}{
		{
			name:          "List with no user groups",
			userGroups:    []*auth.UserGroup{},
			permissions:   map[int][]auth.NamespacePermission{},
			expectedCount: 0,
		},
		{
			name: "List with single user group without permissions",
			userGroups: []*auth.UserGroup{
				{
					ID:        1,
					Name:      "testgroup",
					SiteAdmin: false,
				},
			},
			permissions:   map[int][]auth.NamespacePermission{},
			expectedCount: 1,
		},
		{
			name: "List with single user group with permissions",
			userGroups: []*auth.UserGroup{
				{
					ID:        1,
					Name:      "testgroup",
					SiteAdmin: true,
				},
			},
			permissions: map[int][]auth.NamespacePermission{
				1: {
					{
						UserGroupID:    1,
						NamespaceID:    1,
						PermissionType: auth.PermissionFull,
					},
					{
						UserGroupID:    1,
						NamespaceID:    2,
						PermissionType: auth.PermissionModify,
					},
				},
			},
			expectedCount: 1,
		},
		{
			name: "List with multiple user groups with mixed permissions",
			userGroups: []*auth.UserGroup{
				{
					ID:        1,
					Name:      "site-admins",
					SiteAdmin: true,
				},
				{
					ID:        2,
					Name:      "developers",
					SiteAdmin: false,
				},
				{
					ID:        3,
					Name:      "readers",
					SiteAdmin: false,
				},
			},
			permissions: map[int][]auth.NamespacePermission{
				1: {}, // Site admins with no namespace permissions
				2: {
					{
						UserGroupID:    2,
						NamespaceID:    1,
						PermissionType: auth.PermissionFull,
					},
				},
				3: {
					{
						UserGroupID:    3,
						NamespaceID:    2,
						PermissionType: auth.PermissionModify,
					},
					{
						UserGroupID:    3,
						NamespaceID:    3,
						PermissionType: auth.PermissionFull,
					},
				},
			},
			expectedCount: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			mockUserGroupRepo := new(mocks.MockUserGroupRepository)
			mockNamespaceRepo := new(mocks.MockNamespaceRepository)

			// Set up mock expectations for listing user groups
			mockUserGroupRepo.On("List", ctx, 0, -1).Return(tt.userGroups, nil).Once()

			// Set up mock expectations for getting permissions for each user group
			for _, ug := range tt.userGroups {
				permissions := tt.permissions[ug.ID]
				if permissions == nil {
					permissions = []auth.NamespacePermission{}
				}
				mockUserGroupRepo.On("GetNamespacePermissions", ctx, ug.ID).Return(permissions, nil).Once()

				// Set up mock expectations for finding namespaces by ID
				for _, perm := range permissions {
					namespaceName := "namespace"
					switch perm.NamespaceID {
					case 1:
						namespaceName = "ns1"
					case 2:
						namespaceName = "ns2"
					case 3:
						namespaceName = "ns3"
					}
					mockNamespaceRepo.On("FindByID", ctx, perm.NamespaceID).Return(createMockNamespace(perm.NamespaceID, types.NamespaceName(namespaceName)), nil).Once()
				}
			}

			// Create query
			query := NewListUserGroupsQuery(mockUserGroupRepo, mockNamespaceRepo)

			// Execute
			result, err := query.Execute(ctx)

			// Assert
			require.NoError(t, err)
			require.NotNil(t, result)
			assert.Len(t, result, tt.expectedCount)

			// Verify user group properties
			for i, ugResponse := range result {
				assert.Equal(t, tt.userGroups[i].Name, ugResponse.Name)
				assert.Equal(t, tt.userGroups[i].SiteAdmin, ugResponse.SiteAdmin)

				// Verify permissions
				expectedPermissions := tt.permissions[tt.userGroups[i].ID]
				if expectedPermissions == nil {
					expectedPermissions = []auth.NamespacePermission{}
				}
				assert.Len(t, ugResponse.NamespacePermissions, len(expectedPermissions))
			}

			// Verify mocks were called
			mockUserGroupRepo.AssertExpectations(t)
			mockNamespaceRepo.AssertExpectations(t)
		})
	}
}

func TestListUserGroupsQuery_ListError(t *testing.T) {
	ctx := context.Background()
	mockUserGroupRepo := new(mocks.MockUserGroupRepository)
	mockNamespaceRepo := new(mocks.MockNamespaceRepository)

	// Set up mock expectations - List returns error
	mockUserGroupRepo.On("List", ctx, 0, -1).Return([]*auth.UserGroup{}, assert.AnError).Once()

	// Create query
	query := NewListUserGroupsQuery(mockUserGroupRepo, mockNamespaceRepo)

	// Execute
	result, err := query.Execute(ctx)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)

	// Verify mocks were called
	mockUserGroupRepo.AssertExpectations(t)

	// Verify namespace repo was not called
	mockNamespaceRepo.AssertNotCalled(t, "FindByID", mock.Anything, mock.Anything)
}

func TestListUserGroupsQuery_GetPermissionsError(t *testing.T) {
	ctx := context.Background()
	mockUserGroupRepo := new(mocks.MockUserGroupRepository)
	mockNamespaceRepo := new(mocks.MockNamespaceRepository)

	userGroups := []*auth.UserGroup{
		{
			ID:        1,
			Name:      "testgroup",
			SiteAdmin: false,
		},
	}

	// Set up mock expectations for listing user groups
	mockUserGroupRepo.On("List", ctx, 0, -1).Return(userGroups, nil).Once()

	// Set up mock expectations for getting permissions - returns error
	mockUserGroupRepo.On("GetNamespacePermissions", ctx, userGroups[0].ID).Return([]auth.NamespacePermission{}, assert.AnError).Once()

	// Create query
	query := NewListUserGroupsQuery(mockUserGroupRepo, mockNamespaceRepo)

	// Execute
	result, err := query.Execute(ctx)

	// Assert
	// Note: Based on the implementation, if getting permissions fails, it continues with empty list
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Len(t, result, 1)
	assert.Len(t, result[0].NamespacePermissions, 0) // Empty due to error

	// Verify mocks were called
	mockUserGroupRepo.AssertExpectations(t)
}

func TestListUserGroupsQuery_NamespaceNotFound(t *testing.T) {
	ctx := context.Background()
	mockUserGroupRepo := new(mocks.MockUserGroupRepository)
	mockNamespaceRepo := new(mocks.MockNamespaceRepository)

	userGroups := []*auth.UserGroup{
		{
			ID:        1,
			Name:      "testgroup",
			SiteAdmin: false,
		},
	}

	permissions := []auth.NamespacePermission{
		{
			UserGroupID:    1,
			NamespaceID:    1,
			PermissionType: auth.PermissionFull,
		},
	}

	// Set up mock expectations for listing user groups
	mockUserGroupRepo.On("List", ctx, 0, -1).Return(userGroups, nil).Once()

	// Set up mock expectations for getting permissions
	mockUserGroupRepo.On("GetNamespacePermissions", ctx, userGroups[0].ID).Return(permissions, nil).Once()

	// Set up mock expectations for finding namespace - not found
	mockNamespaceRepo.On("FindByID", ctx, permissions[0].NamespaceID).Return(nil, nil).Once()

	// Create query
	query := NewListUserGroupsQuery(mockUserGroupRepo, mockNamespaceRepo)

	// Execute
	result, err := query.Execute(ctx)

	// Assert
	// Note: Based on the implementation, if namespace is not found, permission is skipped
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Len(t, result, 1)
	assert.Len(t, result[0].NamespacePermissions, 0) // Empty due to namespace not found

	// Verify mocks were called
	mockUserGroupRepo.AssertExpectations(t)
	mockNamespaceRepo.AssertExpectations(t)
}

func TestListUserGroupsQuery_MixedPermissionsAndNamespaces(t *testing.T) {
	ctx := context.Background()
	mockUserGroupRepo := new(mocks.MockUserGroupRepository)
	mockNamespaceRepo := new(mocks.MockNamespaceRepository)

	userGroups := []*auth.UserGroup{
		{
			ID:        1,
			Name:      "testgroup1",
			SiteAdmin: false,
		},
		{
			ID:        2,
			Name:      "testgroup2",
			SiteAdmin: true,
		},
	}

	permissions1 := []auth.NamespacePermission{
		{
			UserGroupID:    1,
			NamespaceID:    1,
			PermissionType: auth.PermissionFull,
		},
		{
			UserGroupID:    1,
			NamespaceID:    2,
			PermissionType: auth.PermissionModify,
		},
	}

	permissions2 := []auth.NamespacePermission{
		{
			UserGroupID:    2,
			NamespaceID:    3,
			PermissionType: auth.PermissionFull,
		},
	}

	// Set up mock expectations for listing user groups
	mockUserGroupRepo.On("List", ctx, 0, -1).Return(userGroups, nil).Once()

	// Set up mock expectations for getting permissions
	mockUserGroupRepo.On("GetNamespacePermissions", ctx, userGroups[0].ID).Return(permissions1, nil).Once()
	mockUserGroupRepo.On("GetNamespacePermissions", ctx, userGroups[1].ID).Return(permissions2, nil).Once()

	// Set up mock expectations for finding namespaces
	mockNamespaceRepo.On("FindByID", ctx, 1).Return(createMockNamespace(1, "ns1"), nil).Once()
	mockNamespaceRepo.On("FindByID", ctx, 2).Return(createMockNamespace(2, "ns2"), nil).Once()
	mockNamespaceRepo.On("FindByID", ctx, 3).Return(createMockNamespace(3, "ns3"), nil).Once()

	// Create query
	query := NewListUserGroupsQuery(mockUserGroupRepo, mockNamespaceRepo)

	// Execute
	result, err := query.Execute(ctx)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Len(t, result, 2)

	// Verify first user group
	assert.Equal(t, "testgroup1", result[0].Name)
	assert.False(t, result[0].SiteAdmin)
	assert.Len(t, result[0].NamespacePermissions, 2)

	// Verify second user group
	assert.Equal(t, "testgroup2", result[1].Name)
	assert.True(t, result[1].SiteAdmin)
	assert.Len(t, result[1].NamespacePermissions, 1)

	// Verify mocks were called
	mockUserGroupRepo.AssertExpectations(t)
	mockNamespaceRepo.AssertExpectations(t)
}
