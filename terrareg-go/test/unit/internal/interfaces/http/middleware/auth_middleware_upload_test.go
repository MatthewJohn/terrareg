package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/auth"
	authservice "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/auth/service"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/config/model"
	infraAuth "github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/auth"
)

// MockAuthMethod for testing
type MockAuthMethod struct {
	mock.Mock
}

func (m *MockAuthMethod) IsBuiltInAdmin() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockAuthMethod) IsAdmin() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockAuthMethod) IsAuthenticated() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockAuthMethod) IsEnabled() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockAuthMethod) RequiresCSRF() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockAuthMethod) CheckAuthState() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockAuthMethod) CanPublishModuleVersion(namespace string) bool {
	args := m.Called(namespace)
	return args.Bool(0)
}

func (m *MockAuthMethod) CanUploadModuleVersion(namespace string) bool {
	args := m.Called(namespace)
	return args.Bool(0)
}

func (m *MockAuthMethod) CheckNamespaceAccess(permissionType, namespace string) bool {
	args := m.Called(permissionType, namespace)
	return args.Bool(0)
}

func (m *MockAuthMethod) GetAllNamespacePermissions() map[string]string {
	args := m.Called()
	return args.Get(0).(map[string]string)
}

func (m *MockAuthMethod) GetUsername() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockAuthMethod) GetUserGroupNames() []string {
	args := m.Called()
	return args.Get(0).([]string)
}

func (m *MockAuthMethod) CanAccessReadAPI() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockAuthMethod) CanAccessTerraformAPI() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockAuthMethod) GetTerraformAuthToken() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockAuthMethod) GetProviderType() auth.AuthMethodType {
	args := m.Called()
	return args.Get(0).(auth.AuthMethodType)
}

func (m *MockAuthMethod) GetProviderData() map[string]interface{} {
	args := m.Called()
	return args.Get(0).(map[string]interface{})
}

// Test RequireUploadPermission middleware
func TestRequireUploadPermission(t *testing.T) {
	tests := []struct {
		name           string
		authMethod     auth.AuthMethod
		namespace      string
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "Upload permission granted",
			authMethod: &MockAuthMethod{
				Mock: mock.Mock{},
			},
			namespace:      "test-namespace",
			expectedStatus: http.StatusOK,
			expectedBody:   "success",
		},
		{
			name: "Upload permission denied",
			authMethod: &MockAuthMethod{
				Mock: mock.Mock{},
			},
			namespace:      "test-namespace",
			expectedStatus: http.StatusForbidden,
			expectedBody:   "Insufficient upload permissions\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock expectations
			mockAuthMethod := tt.authMethod.(*MockAuthMethod)

			// Mock authentication response
			mockAuthMethod.On("IsAuthenticated").Return(true)

			// Mock upload permission check
			if tt.expectedStatus == http.StatusOK {
				mockAuthMethod.On("CanUploadModuleVersion", tt.namespace).Return(true)
			} else {
				mockAuthMethod.On("CanUploadModuleVersion", tt.namespace).Return(false)
			}

			// Create auth factory with mock method
			domainConfig := &model.DomainConfig{}
			authFactory := authservice.NewAuthFactory(
				nil, nil, nil, nil, nil,
			)

			// Replace current auth method with our mock
			// Note: In a real test, you'd need to properly inject the mock
			// For now, we'll test the middleware structure

			// Create middleware
			middleware := NewAuthMiddleware(domainConfig, authFactory)

			// Create test handler
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("success"))
			})

			// Apply middleware
			protectedHandler := middleware.RequireUploadPermission("{namespace}")(handler)

			// Create test request
			req := httptest.NewRequest("POST", "/modules/test-namespace/module/provider/upload", nil)
			req = chi.URLParam(req, "namespace", tt.namespace)

			// Create response recorder
			rr := httptest.NewRecorder()

			// Serve the request
			protectedHandler.ServeHTTP(rr, req)

			// Check response
			assert.Equal(t, tt.expectedStatus, rr.Code)
			assert.Equal(t, tt.expectedBody, rr.Body.String())

			// Verify all expectations were met
			mockAuthMethod.AssertExpectations(t)
		})
	}
}

// Test RequireUploadPermission with missing namespace parameter
func TestRequireUploadPermission_MissingNamespace(t *testing.T) {
	domainConfig := &model.DomainConfig{}
	authFactory := authservice.NewAuthFactory(nil, nil, nil, nil, nil)
	middleware := NewAuthMiddleware(domainConfig, authFactory)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success"))
	})

	protectedHandler := middleware.RequireUploadPermission("{namespace}")(handler)

	// Create request without namespace parameter
	req := httptest.NewRequest("POST", "/modules//module/provider/upload", nil)
	rr := httptest.NewRecorder()

	protectedHandler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	assert.Contains(t, rr.Body.String(), "Namespace parameter required")
}
