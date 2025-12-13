package terrareg

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	authCmd "github.com/matthewjohn/terrareg/terrareg-go/internal/application/command/auth"
	authQuery "github.com/matthewjohn/terrareg/terrareg-go/internal/application/query/auth"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/auth/service"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/config"
)

// MockOIDCLoginCommand is a mock for OIDCLoginCommand
type MockOIDCLoginCommand struct {
	mock.Mock
}

func (m *MockOIDCLoginCommand) Execute(ctx context.Context, redirectURL string) (string, string, error) {
	args := m.Called(ctx, redirectURL)
	return args.String(0), args.String(1), args.Error(2)
}

// MockOIDCCallbackCommand is a mock for OIDCCallbackCommand
type MockOIDCCallbackCommand struct {
	mock.Mock
}

func (m *MockOIDCCallbackCommand) Execute(ctx context.Context, code, state string) (string, error) {
	args := m.Called(ctx, code, state)
	return args.String(0), args.Error(1)
}

// MockSAMLLoginCommand is a mock for SAMLLoginCommand
type MockSAMLLoginCommand struct {
	mock.Mock
}

func (m *MockSAMLLoginCommand) Execute(ctx context.Context, redirectURL string) (string, error) {
	args := m.Called(ctx, redirectURL)
	return args.String(0), args.Error(1)
}

// MockSAMLMetadataCommand is a mock for SAMLMetadataCommand
type MockSAMLMetadataCommand struct {
	mock.Mock
}

func (m *MockSAMLMetadataCommand) Execute(ctx context.Context) (string, error) {
	args := m.Called(ctx)
	return args.String(0), args.Error(1)
}

// MockGitHubOAuthCommand is a mock for GitHubOAuthCommand
type MockGitHubOAuthCommand struct {
	mock.Mock
}

func (m *MockGitHubOAuthCommand) Execute(ctx context.Context, redirectURL string) (string, string, error) {
	args := m.Called(ctx, redirectURL)
	return args.String(0), args.String(1), args.Error(2)
}

// MockAuthenticationService is a mock for AuthenticationService
type MockAuthenticationService struct {
	mock.Mock
}

func (m *MockAuthenticationService) CreateSession(ctx context.Context, w http.ResponseWriter, sessionID string) error {
	args := m.Called(ctx, w, sessionID)
	return args.Error(0)
}

func (m *MockAuthenticationService) ClearSession(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	args := m.Called(ctx, w, r)
	return args.Error(0)
}

func TestAuthHandler_HandleOIDCLogin(t *testing.T) {
	tests := []struct {
		name           string
		redirectURL    string
		cmdAuthURL     string
		cmdState       string
		cmdError       error
		expectedStatus int
	}{
		{
			name:           "success with redirect URL",
			redirectURL:    "/dashboard",
			cmdAuthURL:     "https://oauth.provider.com/auth",
			cmdState:       "state123",
			cmdError:       nil,
			expectedStatus: http.StatusFound,
		},
		{
			name:           "success with default redirect",
			redirectURL:    "",
			cmdAuthURL:     "https://oauth.provider.com/auth",
			cmdState:       "state123",
			cmdError:       nil,
			expectedStatus: http.StatusFound,
		},
		{
			name:           "command error",
			redirectURL:    "/dashboard",
			cmdAuthURL:     "",
			cmdState:       "",
			cmdError:       errors.New("OIDC error"),
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			mockCmd := new(MockOIDCLoginCommand)
			mockCmd.On("Execute", mock.Anything, tt.redirectURL).Return(tt.cmdAuthURL, tt.cmdState, tt.cmdError)

			handler := &AuthHandler{
				oidcLoginCmd: mockCmd,
			}

			// Create request with query parameters
			reqURL := "/v1/terrareg/auth/oidc/login"
			if tt.redirectURL != "" {
				reqURL += "?redirect_url=" + url.QueryEscape(tt.redirectURL)
			}
			req := httptest.NewRequest(http.MethodGet, reqURL, nil)
			w := httptest.NewRecorder()

			// Execute handler
			handler.HandleOIDCLogin(w, req)

			// Assert response
			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedStatus == http.StatusFound {
				location := w.Header().Get("Location")
				assert.Equal(t, tt.cmdAuthURL, location)
			}

			mockCmd.AssertExpectations(t)
		})
	}
}

func TestAuthHandler_HandleOIDCCallback(t *testing.T) {
	tests := []struct {
		name           string
		queryParams    string
		cmdSessionID   string
		cmdError       error
		serviceError   error
		expectedStatus int
	}{
		{
			name:           "success",
			queryParams:    "code=auth123&state=state123",
			cmdSessionID:   "session123",
			cmdError:       nil,
			serviceError:   nil,
			expectedStatus: http.StatusFound,
		},
		{
			name:           "OAuth error response",
			queryParams:    "error=access_denied",
			cmdSessionID:   "",
			cmdError:       nil,
			serviceError:   nil,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "command error",
			queryParams:    "code=auth123&state=state123",
			cmdSessionID:   "",
			cmdError:       errors.New("invalid code"),
			serviceError:   nil,
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:           "service error",
			queryParams:    "code=auth123&state=state123",
			cmdSessionID:   "session123",
			cmdError:       nil,
			serviceError:   errors.New("session creation failed"),
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			mockCmd := new(MockOIDCCallbackCommand)
			mockService := new(MockAuthenticationService)

			if !strings.Contains(tt.queryParams, "error=") {
				code := "auth123"
				state := "state123"
				mockCmd.On("Execute", mock.Anything, code, state).Return(tt.cmdSessionID, tt.cmdError)

				if tt.cmdError == nil && tt.cmdSessionID != "" {
					mockService.On("CreateSession", mock.Anything, mock.AnythingOfType("*httptest.ResponseRecorder"), tt.cmdSessionID).Return(tt.serviceError)
				}
			}

			handler := &AuthHandler{
				oidcCallbackCmd: mockCmd,
				authService:     mockService,
			}

			// Create request with query parameters
			req := httptest.NewRequest(http.MethodGet, "/v1/terrareg/auth/oidc/callback?"+tt.queryParams, nil)
			w := httptest.NewRecorder()

			// Execute handler
			handler.HandleOIDCCallback(w, req)

			// Assert response
			assert.Equal(t, tt.expectedStatus, w.Code)

			if !strings.Contains(tt.queryParams, "error=") {
				mockCmd.AssertExpectations(t)
				if tt.cmdError == nil && tt.cmdSessionID != "" {
					mockService.AssertExpectations(t)
				}
			}
		})
	}
}

func TestAuthHandler_HandleSAMLLogin(t *testing.T) {
	tests := []struct {
		name           string
		redirectURL    string
		cmdAuthRequest string
		cmdError       error
		expectedStatus int
	}{
		{
			name:           "success",
			redirectURL:    "/dashboard",
			cmdAuthRequest: "<saml>AuthRequest</saml>",
			cmdError:       nil,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "command error",
			redirectURL:    "/dashboard",
			cmdAuthRequest: "",
			cmdError:       errors.New("SAML error"),
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			mockCmd := new(MockSAMLLoginCommand)
			mockCmd.On("Execute", mock.Anything, tt.redirectURL).Return(tt.cmdAuthRequest, tt.cmdError)

			handler := &AuthHandler{
				samlLoginCmd: mockCmd,
			}

			// Create request
			reqURL := "/v1/terrareg/auth/saml/login"
			if tt.redirectURL != "" {
				reqURL += "?redirect_url=" + url.QueryEscape(tt.redirectURL)
			}
			req := httptest.NewRequest(http.MethodGet, reqURL, nil)
			w := httptest.NewRecorder()

			// Execute handler
			handler.HandleSAMLLogin(w, req)

			// Assert response
			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedStatus == http.StatusOK {
				contentType := w.Header().Get("Content-Type")
				assert.Equal(t, "application/html", contentType)
				assert.Equal(t, tt.cmdAuthRequest, w.Body.String())
			}

			mockCmd.AssertExpectations(t)
		})
	}
}

func TestAuthHandler_HandleSAMLMetadata(t *testing.T) {
	tests := []struct {
		name           string
		cmdMetadata    string
		cmdError       error
		expectedStatus int
	}{
		{
			name:           "success",
			cmdMetadata:    "<?xml version=\"1.0\"?><EntityDescriptor></EntityDescriptor>",
			cmdError:       nil,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "command error",
			cmdMetadata:    "",
			cmdError:       errors.New("metadata error"),
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			mockCmd := new(MockSAMLMetadataCommand)
			mockCmd.On("Execute", mock.Anything).Return(tt.cmdMetadata, tt.cmdError)

			handler := &AuthHandler{
				samlMetadataCmd: mockCmd,
			}

			// Create request
			req := httptest.NewRequest(http.MethodGet, "/v1/terrareg/auth/saml/metadata", nil)
			w := httptest.NewRecorder()

			// Execute handler
			handler.HandleSAMLMetadata(w, req)

			// Assert response
			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedStatus == http.StatusOK {
				contentType := w.Header().Get("Content-Type")
				assert.Equal(t, "application/xml", contentType)
				assert.Equal(t, tt.cmdMetadata, w.Body.String())
			}

			mockCmd.AssertExpectations(t)
		})
	}
}

func TestAuthHandler_HandleGitHubOAuth(t *testing.T) {
	tests := []struct {
		name           string
		redirectURL    string
		cmdAuthURL     string
		cmdState       string
		cmdError       error
		expectedStatus int
	}{
		{
			name:           "success",
			redirectURL:    "/dashboard",
			cmdAuthURL:     "https://github.com/login/oauth/authorize",
			cmdState:       "state123",
			cmdError:       nil,
			expectedStatus: http.StatusFound,
		},
		{
			name:           "command error",
			redirectURL:    "/dashboard",
			cmdAuthURL:     "",
			cmdState:       "",
			cmdError:       errors.New("GitHub OAuth error"),
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			mockCmd := new(MockGitHubOAuthCommand)
			mockCmd.On("Execute", mock.Anything, tt.redirectURL).Return(tt.cmdAuthURL, tt.cmdState, tt.cmdError)

			handler := &AuthHandler{
				githubOAuthCmd: mockCmd,
			}

			// Create request
			reqURL := "/v1/terrareg/auth/github/oauth"
			if tt.redirectURL != "" {
				reqURL += "?redirect_url=" + url.QueryEscape(tt.redirectURL)
			}
			req := httptest.NewRequest(http.MethodGet, reqURL, nil)
			w := httptest.NewRecorder()

			// Execute handler
			handler.HandleGitHubOAuth(w, req)

			// Assert response
			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedStatus == http.StatusFound {
				location := w.Header().Get("Location")
				assert.Equal(t, tt.cmdAuthURL, location)
			}

			mockCmd.AssertExpectations(t)
		})
	}
}

func TestAuthHandler_HandleLogout(t *testing.T) {
	tests := []struct {
		name           string
		serviceError   error
		expectedStatus int
	}{
		{
			name:           "success",
			serviceError:   nil,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "service error",
			serviceError:   errors.New("logout failed"),
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			mockService := new(MockAuthenticationService)
			mockService.On("ClearSession", mock.Anything, mock.AnythingOfType("*httptest.ResponseRecorder"), mock.AnythingOfType("*http.Request")).Return(tt.serviceError)

			handler := &AuthHandler{
				authService: mockService,
			}

			// Create request
			req := httptest.NewRequest(http.MethodPost, "/v1/terrareg/auth/logout", nil)
			w := httptest.NewRecorder()

			// Execute handler
			handler.HandleLogout(w, req)

			// Assert response
			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedStatus == http.StatusOK {
				// Check for JSON response
				assert.Contains(t, w.Body.String(), "Successfully logged out")
			}

			mockService.AssertExpectations(t)
		})
	}
}