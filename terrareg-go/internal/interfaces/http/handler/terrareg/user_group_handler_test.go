package terrareg

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	userGroupCmd "github.com/matthewjohn/terrareg/terrareg-go/internal/application/command/user_group"
	userGroupQuery "github.com/matthewjohn/terrareg/terrareg-go/internal/application/query/user_group"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/auth"
	usergroupdto "github.com/matthewjohn/terrareg/terrareg-go/internal/interfaces/http/dto/user_group_dto"
)

// MockListUserGroupsQuery is a mock for ListUserGroupsQuery
type MockListUserGroupsQuery struct {
	mock.Mock
}

func (m *MockListUserGroupsQuery) Execute(ctx context.Context) ([]*auth.UserGroup, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*auth.UserGroup), args.Error(1)
}

// MockCreateUserGroupCommand is a mock for CreateUserGroupCommand
type MockCreateUserGroupCommand struct {
	mock.Mock
}

func (m *MockCreateUserGroupCommand) Execute(ctx context.Context, req userGroupCmd.CreateUserGroupRequest) (*auth.UserGroup, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*auth.UserGroup), args.Error(1)
}

// MockGetUserGroupDetailsQuery is a mock for GetUserGroupDetailsQuery
type MockGetUserGroupDetailsQuery struct {
	mock.Mock
}

func (m *MockGetUserGroupDetailsQuery) Execute(ctx context.Context, groupID int) (*auth.UserGroup, error) {
	args := m.Called(ctx, groupID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*auth.UserGroup), args.Error(1)
}

// MockUpdateUserGroupCommand is a mock for UpdateUserGroupCommand
type MockUpdateUserGroupCommand struct {
	mock.Mock
}

func (m *MockUpdateUserGroupCommand) Execute(ctx context.Context, req userGroupCmd.UpdateUserGroupRequest) (*auth.UserGroup, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*auth.UserGroup), args.Error(1)
}

// MockDeleteUserGroupCommand is a mock for DeleteUserGroupCommand
type MockDeleteUserGroupCommand struct {
	mock.Mock
}

func (m *MockDeleteUserGroupCommand) Execute(ctx context.Context, req userGroupCmd.DeleteUserGroupRequest) error {
	args := m.Called(ctx, req)
	return args.Error(0)
}

// MockGetGroupPermissionsQuery is a mock for GetGroupPermissionsQuery
type MockGetGroupPermissionsQuery struct {
	mock.Mock
}

func (m *MockGetGroupPermissionsQuery) Execute(ctx context.Context, groupID int) ([]*auth.NamespacePermissionData, error) {
	args := m.Called(ctx, groupID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*auth.NamespacePermissionData), args.Error(1)
}

// MockUpdateGroupPermissionsCommand is a mock for UpdateGroupPermissionsCommand
type MockUpdateGroupPermissionsCommand struct {
	mock.Mock
}

func (m *MockUpdateGroupPermissionsCommand) Execute(ctx context.Context, req userGroupCmd.UpdateGroupPermissionsRequest) error {
	args := m.Called(ctx, req)
	return args.Error(0)
}

func TestUserGroupHandler_HandleListUserGroups(t *testing.T) {
	tests := []struct {
		name           string
		queryResult    []*auth.UserGroup
		queryError     error
		expectedStatus int
	}{
		{
			name: "success",
			queryResult: []*auth.UserGroup{
				{
					ID:        1,
					Name:      "admin",
					SiteAdmin: true,
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				},
				{
					ID:        2,
					Name:      "developers",
					SiteAdmin: false,
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				},
			},
			queryError:     nil,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "query error",
			queryResult:    nil,
			queryError:     errors.New("database error"),
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			mockQuery := new(MockListUserGroupsQuery)
			mockQuery.On("Execute", mock.Anything).Return(tt.queryResult, tt.queryError)

			handler := &UserGroupHandler{
				listUserGroupsQuery: mockQuery,
			}

			// Create request
			req := httptest.NewRequest(http.MethodGet, "/v1/terrareg/user-groups", nil)
			w := httptest.NewRecorder()

			// Execute handler
			handler.HandleListUserGroups(w, req)

			// Assert response
			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedStatus == http.StatusOK {
				var response usergroupdto.UserGroupListResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.Len(t, response.Groups, len(tt.queryResult))
			}

			mockQuery.AssertExpectations(t)
		})
	}
}

func TestUserGroupHandler_HandleCreateUserGroup(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    interface{}
		cmdResult      *auth.UserGroup
		cmdError       error
		expectedStatus int
	}{
		{
			name: "success",
			requestBody: map[string]interface{}{
				"name":       "test-group",
				"site_admin": false,
			},
			cmdResult: &auth.UserGroup{
				ID:        1,
				Name:      "test-group",
				SiteAdmin: false,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			cmdError:       nil,
			expectedStatus: http.StatusCreated,
		},
		{
			name:           "invalid request body",
			requestBody:    "invalid",
			cmdResult:      nil,
			cmdError:       nil,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "command error",
			requestBody: map[string]interface{}{
				"name": "test-group",
			},
			cmdResult:      nil,
			cmdError:       errors.New("creation failed"),
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			mockCmd := new(MockCreateUserGroupCommand)

			if tt.requestBody != "invalid" {
				mockCmd.On("Execute", mock.Anything, mock.AnythingOfType("user_group.CreateUserGroupRequest")).Return(tt.cmdResult, tt.cmdError)
			}

			handler := &UserGroupHandler{
				createUserGroupCmd: mockCmd,
			}

			// Create request
			var body []byte
			if tt.requestBody != nil {
				body, _ = json.Marshal(tt.requestBody)
			}
			req := httptest.NewRequest(http.MethodPost, "/v1/terrareg/user-groups", bytes.NewReader(body))
			w := httptest.NewRecorder()

			// Execute handler
			handler.HandleCreateUserGroup(w, req)

			// Assert response
			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedStatus == http.StatusCreated {
				var response usergroupdto.UserGroupResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.Equal(t, tt.cmdResult.ID, response.ID)
				assert.Equal(t, tt.cmdResult.Name, response.Name)
			}

			if tt.requestBody != "invalid" {
				mockCmd.AssertExpectations(t)
			}
		})
	}
}

func TestUserGroupHandler_HandleGetUserGroupDetails(t *testing.T) {
	tests := []struct {
		name           string
		groupID        string
		queryResult    *auth.UserGroup
		queryError     error
		expectedStatus int
	}{
		{
			name:    "success",
			groupID: "1",
			queryResult: &auth.UserGroup{
				ID:        1,
				Name:      "test-group",
				SiteAdmin: false,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			queryError:     nil,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "invalid group ID",
			groupID:        "invalid",
			queryResult:    nil,
			queryError:     nil,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:        "group not found",
			groupID:     "999",
			queryResult: nil,
			queryError:  errors.New("group not found"),
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			mockQuery := new(MockGetUserGroupDetailsQuery)

			if tt.groupID != "invalid" {
				groupID, _ := strconv.Atoi(tt.groupID)
				mockQuery.On("Execute", mock.Anything, groupID).Return(tt.queryResult, tt.queryError)
			}

			handler := &UserGroupHandler{
				getUserGroupDetailsQuery: mockQuery,
			}

			// Create request with chi router context for URL parameter
			req := httptest.NewRequest(http.MethodGet, "/v1/terrareg/user-groups/"+tt.groupID, nil)
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("groupID", tt.groupID)
			ctx := context.WithValue(req.Context(), chi.RouteCtxKey, rctx)
			req = req.WithContext(ctx)
			w := httptest.NewRecorder()

			// Execute handler
			handler.HandleGetUserGroupDetails(w, req)

			// Assert response
			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedStatus == http.StatusOK {
				var response usergroupdto.UserGroupResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.Equal(t, tt.queryResult.ID, response.ID)
				assert.Equal(t, tt.queryResult.Name, response.Name)
			}

			if tt.groupID != "invalid" {
				mockQuery.AssertExpectations(t)
			}
		})
	}
}

func TestUserGroupHandler_HandleUpdateUserGroup(t *testing.T) {
	tests := []struct {
		name           string
		groupID        string
		requestBody    interface{}
		cmdResult      *auth.UserGroup
		cmdError       error
		expectedStatus int
	}{
		{
			name:    "success",
			groupID: "1",
			requestBody: map[string]interface{}{
				"name":       "updated-group",
				"site_admin": true,
			},
			cmdResult: &auth.UserGroup{
				ID:        1,
				Name:      "updated-group",
				SiteAdmin: true,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			cmdError:       nil,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "invalid group ID",
			groupID:        "invalid",
			requestBody:    map[string]interface{}{"name": "updated-group"},
			cmdResult:      nil,
			cmdError:       nil,
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			mockCmd := new(MockUpdateUserGroupCommand)

			if tt.groupID != "invalid" {
				mockCmd.On("Execute", mock.Anything, mock.AnythingOfType("user_group.UpdateUserGroupRequest")).Return(tt.cmdResult, tt.cmdError)
			}

			handler := &UserGroupHandler{
				updateUserGroupCmd: mockCmd,
			}

			// Create request
			var body []byte
			if tt.requestBody != nil {
				body, _ = json.Marshal(tt.requestBody)
			}
			req := httptest.NewRequest(http.MethodPut, "/v1/terrareg/user-groups/"+tt.groupID, bytes.NewReader(body))
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("groupID", tt.groupID)
			ctx := context.WithValue(req.Context(), chi.RouteCtxKey, rctx)
			req = req.WithContext(ctx)
			w := httptest.NewRecorder()

			// Execute handler
			handler.HandleUpdateUserGroup(w, req)

			// Assert response
			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedStatus == http.StatusOK {
				var response usergroupdto.UserGroupResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.Equal(t, tt.cmdResult.ID, response.ID)
				assert.Equal(t, tt.cmdResult.Name, response.Name)
			}

			if tt.groupID != "invalid" {
				mockCmd.AssertExpectations(t)
			}
		})
	}
}

func TestUserGroupHandler_HandleDeleteUserGroup(t *testing.T) {
	tests := []struct {
		name           string
		groupID        string
		cmdError       error
		expectedStatus int
	}{
		{
			name:           "success",
			groupID:        "1",
			cmdError:       nil,
			expectedStatus: http.StatusNoContent,
		},
		{
			name:           "invalid group ID",
			groupID:        "invalid",
			cmdError:       nil,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "command error",
			groupID:        "1",
			cmdError:       errors.New("delete failed"),
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			mockCmd := new(MockDeleteUserGroupCommand)

			if tt.groupID != "invalid" {
				mockCmd.On("Execute", mock.Anything, mock.AnythingOfType("user_group.DeleteUserGroupRequest")).Return(tt.cmdError)
			}

			handler := &UserGroupHandler{
				deleteUserGroupCmd: mockCmd,
			}

			// Create request
			req := httptest.NewRequest(http.MethodDelete, "/v1/terrareg/user-groups/"+tt.groupID, nil)
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("groupID", tt.groupID)
			ctx := context.WithValue(req.Context(), chi.RouteCtxKey, rctx)
			req = req.WithContext(ctx)
			w := httptest.NewRecorder()

			// Execute handler
			handler.HandleDeleteUserGroup(w, req)

			// Assert response
			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.groupID != "invalid" {
				mockCmd.AssertExpectations(t)
			}
		})
	}
}

func TestUserGroupHandler_HandleGetGroupPermissions(t *testing.T) {
	tests := []struct {
		name           string
		groupID        string
		queryResult    []*auth.NamespacePermissionData
		queryError     error
		expectedStatus int
	}{
		{
			name:    "success",
			groupID: "1",
			queryResult: []*auth.NamespacePermissionData{
				{
					NamespaceID:   1,
					NamespaceName: "namespace1",
					Permission:    auth.PermissionRead,
					GroupName:     "test-group",
				},
				{
					NamespaceID:   2,
					NamespaceName: "namespace2",
					Permission:    auth.PermissionModify,
					GroupName:     "test-group",
				},
			},
			queryError:     nil,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "invalid group ID",
			groupID:        "invalid",
			queryResult:    nil,
			queryError:     nil,
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			mockQuery := new(MockGetGroupPermissionsQuery)

			if tt.groupID != "invalid" {
				groupID, _ := strconv.Atoi(tt.groupID)
				mockQuery.On("Execute", mock.Anything, groupID).Return(tt.queryResult, tt.queryError)
			}

			handler := &UserGroupHandler{
				getGroupPermissionsQuery: mockQuery,
			}

			// Create request
			req := httptest.NewRequest(http.MethodGet, "/v1/terrareg/user-groups/"+tt.groupID+"/permissions", nil)
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("groupID", tt.groupID)
			ctx := context.WithValue(req.Context(), chi.RouteCtxKey, rctx)
			req = req.WithContext(ctx)
			w := httptest.NewRecorder()

			// Execute handler
			handler.HandleGetGroupPermissions(w, req)

			// Assert response
			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedStatus == http.StatusOK {
				var response usergroupdto.NamespacePermissionResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.Equal(t, tt.groupID, strconv.Itoa(response.UserGroupID))
				assert.Len(t, response.Permissions, len(tt.queryResult))
			}

			if tt.groupID != "invalid" {
				mockQuery.AssertExpectations(t)
			}
		})
	}
}