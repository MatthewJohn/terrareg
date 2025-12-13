package terrareg_test

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

	userCmd "github.com/matthewjohn/terrareg/terrareg-go/internal/application/command/user"
	userQuery "github.com/matthewjohn/terrareg/terrareg-go/internal/application/query/user"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/identity/model"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/interfaces/http/dto/userdto"
)

// MockGetUserProfileQuery is a mock for GetUserProfileQuery
type MockGetUserProfileQuery struct {
	mock.Mock
}

func (m *MockGetUserProfileQuery) Execute(ctx context.Context, userID int) (*model.User, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

// MockUpdateUserProfileCommand is a mock for UpdateUserProfileCommand
type MockUpdateUserProfileCommand struct {
	mock.Mock
}

func (m *MockUpdateUserProfileCommand) Execute(ctx context.Context, req userCmd.UpdateUserProfileRequest) (*model.User, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

// MockListUserAccessTokensQuery is a mock for ListUserAccessTokensQuery
type MockListUserAccessTokensQuery struct {
	mock.Mock
}

func (m *MockListUserAccessTokensQuery) Execute(ctx context.Context, userID int) ([]*model.AccessToken, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.AccessToken), args.Error(1)
}

// MockCreateUserAccessTokenCommand is a mock for CreateUserAccessTokenCommand
type MockCreateUserAccessTokenCommand struct {
	mock.Mock
}

func (m *MockCreateUserAccessTokenCommand) Execute(ctx context.Context, req userCmd.CreateUserAccessTokenRequest) (*model.AccessToken, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.AccessToken), args.Error(1)
}

// MockDeleteUserAccessTokenCommand is a mock for DeleteUserAccessTokenCommand
type MockDeleteUserAccessTokenCommand struct {
	mock.Mock
}

func (m *MockDeleteUserAccessTokenCommand) Execute(ctx context.Context, req userCmd.DeleteUserAccessTokenRequest) error {
	args := m.Called(ctx, req)
	return args.Error(0)
}

// MockGetUserSiteAdminQuery is a mock for GetUserSiteAdminQuery
type MockGetUserSiteAdminQuery struct {
	mock.Mock
}

func (m *MockGetUserSiteAdminQuery) Execute(ctx context.Context, userID int) (bool, error) {
	args := m.Called(ctx, userID)
	return args.Bool(0), args.Error(1)
}

// MockGetUserPermissionsQuery is a mock for GetUserPermissionsQuery
type MockGetUserPermissionsQuery struct {
	mock.Mock
}

func (m *MockGetUserPermissionsQuery) Execute(ctx context.Context, userID int) (*userQuery.UserPermissions, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*userQuery.UserPermissions), args.Error(1)
}

func TestUserHandler_HandleGetUserProfile(t *testing.T) {
	tests := []struct {
		name           string
		userIDInCtx    int
		queryError     error
		expectedStatus int
	}{
		{
			name:           "success",
			userIDInCtx:    1,
			queryError:     nil,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "user not authenticated",
			userIDInCtx:    0,
			queryError:     nil,
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "query error",
			userIDInCtx:    1,
			queryError:     errors.New("database error"),
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			mockQuery := new(MockGetUserProfileQuery)

			if tt.userIDInCtx > 0 && tt.queryError == nil {
				user := &model.User{
					ID:           tt.userIDInCtx,
					Username:     "testuser",
					Email:        "test@example.com",
					Active:       true,
					ExternalID:   "12345",
					AuthProvider: model.AuthMethodOpenIDConnect,
					CreatedAt:    time.Now(),
					UpdatedAt:    time.Now(),
				}
				mockQuery.On("Execute", mock.Anything, tt.userIDInCtx).Return(user, nil)
			} else if tt.userIDInCtx > 0 && tt.queryError != nil {
				mockQuery.On("Execute", mock.Anything, tt.userIDInCtx).Return(nil, tt.queryError)
			}

			handler := &UserHandler{
				getUserProfileQuery: mockQuery,
			}

			// Create request
			req := httptest.NewRequest(http.MethodGet, "/v1/terrareg/user/profile", nil)
			ctx := context.WithValue(req.Context(), contextKey("user_id"), tt.userIDInCtx)
			req = req.WithContext(ctx)
			w := httptest.NewRecorder()

			// Execute handler
			handler.HandleGetUserProfile(w, req)

			// Assert response
			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedStatus == http.StatusOK {
				var response userdto.UserProfileResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.Equal(t, tt.userIDInCtx, response.ID)
				assert.Equal(t, "testuser", response.Username)
			}

			mockQuery.AssertExpectations(t)
		})
	}
}

func TestUserHandler_HandleUpdateUserProfile(t *testing.T) {
	tests := []struct {
		name           string
		userIDInCtx    int
		requestBody    interface{}
		cmdError       error
		expectedStatus int
	}{
		{
			name:           "success",
			userIDInCtx:    1,
			requestBody:    map[string]interface{}{"username": "newuser", "email": "new@example.com"},
			cmdError:       nil,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "invalid request body",
			userIDInCtx:    1,
			requestBody:    "invalid",
			cmdError:       nil,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "command error",
			userIDInCtx:    1,
			requestBody:    map[string]interface{}{"username": "newuser"},
			cmdError:       errors.New("update failed"),
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			mockCmd := new(MockUpdateUserProfileCommand)

			if tt.userIDInCtx > 0 && tt.cmdError == nil && tt.requestBody != "invalid" {
				updatedUser := &model.User{
					ID:           tt.userIDInCtx,
					Username:     "newuser",
					Email:        "new@example.com",
					Active:       true,
					ExternalID:   "12345",
					AuthProvider: model.AuthMethodOpenIDConnect,
					CreatedAt:    time.Now(),
					UpdatedAt:    time.Now(),
				}
				mockCmd.On("Execute", mock.Anything, mock.AnythingOfType("user.UpdateUserProfileRequest")).Return(updatedUser, nil)
			} else if tt.userIDInCtx > 0 && tt.cmdError != nil && tt.requestBody != "invalid" {
				mockCmd.On("Execute", mock.Anything, mock.AnythingOfType("user.UpdateUserProfileRequest")).Return(nil, tt.cmdError)
			}

			handler := &UserHandler{
				updateUserProfileCmd: mockCmd,
			}

			// Create request
			var body []byte
			if tt.requestBody != nil {
				body, _ = json.Marshal(tt.requestBody)
			}
			req := httptest.NewRequest(http.MethodPut, "/v1/terrareg/user/profile", bytes.NewReader(body))
			ctx := context.WithValue(req.Context(), contextKey("user_id"), tt.userIDInCtx)
			req = req.WithContext(ctx)
			w := httptest.NewRecorder()

			// Execute handler
			handler.HandleUpdateUserProfile(w, req)

			// Assert response
			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedStatus == http.StatusOK {
				var response userdto.UserProfileResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.Equal(t, "newuser", response.Username)
				assert.Equal(t, "new@example.com", response.Email)
			}

			if tt.requestBody != "invalid" {
				mockCmd.AssertExpectations(t)
			}
		})
	}
}

func TestUserHandler_HandleCreateUserAccessToken(t *testing.T) {
	tests := []struct {
		name           string
		userIDInCtx    int
		requestBody    interface{}
		cmdError       error
		expectedStatus int
	}{
		{
			name:        "success",
			userIDInCtx: 1,
			requestBody: map[string]interface{}{
				"name":        "test-token",
				"description": "Test token",
			},
			cmdError:       nil,
			expectedStatus: http.StatusCreated,
		},
		{
			name:           "invalid request body",
			userIDInCtx:    1,
			requestBody:    "invalid",
			cmdError:       nil,
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			mockCmd := new(MockCreateUserAccessTokenCommand)

			if tt.userIDInCtx > 0 && tt.cmdError == nil && tt.requestBody != "invalid" {
				token := &model.AccessToken{
					ID:          1,
					Token:       "test-token-value",
					Name:        "test-token",
					Description: "Test token",
					CreatedAt:   time.Now(),
				}
				mockCmd.On("Execute", mock.Anything, mock.AnythingOfType("user.CreateUserAccessTokenRequest")).Return(token, nil)
			}

			handler := &UserHandler{
				createUserAccessTokenCmd: mockCmd,
			}

			// Create request
			var body []byte
			if tt.requestBody != nil {
				body, _ = json.Marshal(tt.requestBody)
			}
			req := httptest.NewRequest(http.MethodPost, "/v1/terrareg/user/access-tokens", bytes.NewReader(body))
			ctx := context.WithValue(req.Context(), contextKey("user_id"), tt.userIDInCtx)
			req = req.WithContext(ctx)
			w := httptest.NewRecorder()

			// Execute handler
			handler.HandleCreateUserAccessToken(w, req)

			// Assert response
			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedStatus == http.StatusCreated {
				var response userdto.UserAccessTokenResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.Equal(t, "test-token", response.Name)
				assert.Equal(t, "Test token", response.Description)
			}

			if tt.requestBody != "invalid" {
				mockCmd.AssertExpectations(t)
			}
		})
	}
}

// contextKey is used to avoid context key collisions
type contextKey string

func TestUserHandler_HandleDeleteUserAccessToken(t *testing.T) {
	tests := []struct {
		name           string
		userIDInCtx    int
		tokenID        string
		cmdError       error
		expectedStatus int
	}{
		{
			name:           "success",
			userIDInCtx:    1,
			tokenID:        "123",
			cmdError:       nil,
			expectedStatus: http.StatusNoContent,
		},
		{
			name:           "invalid token ID",
			userIDInCtx:    1,
			tokenID:        "invalid",
			cmdError:       nil,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "command error",
			userIDInCtx:    1,
			tokenID:        "123",
			cmdError:       errors.New("delete failed"),
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			mockCmd := new(MockDeleteUserAccessTokenCommand)

			if tt.tokenID != "invalid" && tt.cmdError == nil {
				mockCmd.On("Execute", mock.Anything, mock.AnythingOfType("user.DeleteUserAccessTokenRequest")).Return(nil)
			} else if tt.tokenID != "invalid" && tt.cmdError != nil {
				mockCmd.On("Execute", mock.Anything, mock.AnythingOfType("user.DeleteUserAccessTokenRequest")).Return(tt.cmdError)
			}

			handler := &UserHandler{
				deleteUserAccessTokenCmd: mockCmd,
			}

			// Create request with chi router context for URL parameter
			req := httptest.NewRequest(http.MethodDelete, "/v1/terrareg/user/access-tokens/"+tt.tokenID, nil)
			ctx := context.WithValue(req.Context(), contextKey("user_id"), tt.userIDInCtx)

			// Add chi URL param to context
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("tokenID", tt.tokenID)
			ctx = context.WithValue(ctx, chi.RouteCtxKey, rctx)

			req = req.WithContext(ctx)
			w := httptest.NewRecorder()

			// Execute handler
			handler.HandleDeleteUserAccessToken(w, req)

			// Assert response
			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.tokenID != "invalid" {
				mockCmd.AssertExpectations(t)
			}
		})
	}
}

func TestUserHandler_HandleGetUserSiteAdmin(t *testing.T) {
	tests := []struct {
		name           string
		userIDInCtx    int
		isSiteAdmin    bool
		queryError     error
		expectedStatus int
	}{
		{
			name:           "site admin true",
			userIDInCtx:    1,
			isSiteAdmin:    true,
			queryError:     nil,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "site admin false",
			userIDInCtx:    1,
			isSiteAdmin:    false,
			queryError:     nil,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "user not authenticated",
			userIDInCtx:    0,
			isSiteAdmin:    false,
			queryError:     nil,
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			mockQuery := new(MockGetUserSiteAdminQuery)

			if tt.userIDInCtx > 0 && tt.queryError == nil {
				mockQuery.On("Execute", mock.Anything, tt.userIDInCtx).Return(tt.isSiteAdmin, nil)
			}

			handler := &UserHandler{
				getUserSiteAdminQuery: mockQuery,
			}

			// Create request
			req := httptest.NewRequest(http.MethodGet, "/v1/terrareg/user/site-admin", nil)
			ctx := context.WithValue(req.Context(), contextKey("user_id"), tt.userIDInCtx)
			req = req.WithContext(ctx)
			w := httptest.NewRecorder()

			// Execute handler
			handler.HandleGetUserSiteAdmin(w, req)

			// Assert response
			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedStatus == http.StatusOK {
				var response userdto.UserSiteAdminResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.Equal(t, tt.isSiteAdmin, response.IsSiteAdmin)
			}

			if tt.userIDInCtx > 0 {
				mockQuery.AssertExpectations(t)
			}
		})
	}
}

func TestUserHandler_HandleGetUserPermissions(t *testing.T) {
	tests := []struct {
		name           string
		userIDInCtx    int
		permissions    *userQuery.UserPermissions
		queryError     error
		expectedStatus int
	}{
		{
			name:        "success",
			userIDInCtx: 1,
			permissions: &userQuery.UserPermissions{
				ReadAccess: true,
				SiteAdmin:  false,
				NamespacePermissions: map[string]string{
					"namespace1": "READ",
					"namespace2": "MODIFY",
				},
			},
			queryError:     nil,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "user not authenticated",
			userIDInCtx:    0,
			permissions:    nil,
			queryError:     nil,
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			mockQuery := new(MockGetUserPermissionsQuery)

			if tt.userIDInCtx > 0 && tt.queryError == nil {
				mockQuery.On("Execute", mock.Anything, tt.userIDInCtx).Return(tt.permissions, nil)
			}

			handler := &UserHandler{
				getUserPermissionsQuery: mockQuery,
			}

			// Create request
			req := httptest.NewRequest(http.MethodGet, "/v1/terrareg/user/permissions", nil)
			ctx := context.WithValue(req.Context(), contextKey("user_id"), tt.userIDInCtx)
			req = req.WithContext(ctx)
			w := httptest.NewRecorder()

			// Execute handler
			handler.HandleGetUserPermissions(w, req)

			// Assert response
			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedStatus == http.StatusOK {
				var response userdto.UserPermissionsResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.Equal(t, tt.permissions.ReadAccess, response.ReadAccess)
				assert.Equal(t, tt.permissions.SiteAdmin, response.SiteAdmin)
				assert.Equal(t, tt.permissions.NamespacePermissions, response.NamespacePermissions)
			}

			if tt.userIDInCtx > 0 {
				mockQuery.AssertExpectations(t)
			}
		})
	}
}
