package testutils

import (
	"context"
	"net/http"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/auth"
)

// MockAuthConfig configures mock authentication for tests
type MockAuthConfig struct {
	AuthMethod       auth.AuthMethodType
	IsAuthenticated  bool
	IsAdmin          bool
	IsBuiltInAdmin   bool
	Username         string
	UserGroups       []string
	Permissions      map[string]string // namespace -> permission
	CanAccessReadAPI bool
}

// mockAuthContext is a mock AuthContext for testing
type mockAuthContext struct {
	config MockAuthConfig
}

func (m *mockAuthContext) IsBuiltInAdmin() bool {
	return m.config.IsBuiltInAdmin
}

func (m *mockAuthContext) IsAdmin() bool {
	return m.config.IsAdmin
}

func (m *mockAuthContext) IsAuthenticated() bool {
	return m.config.IsAuthenticated
}

func (m *mockAuthContext) RequiresCSRF() bool {
	return false
}

func (m *mockAuthContext) CheckAuthState() bool {
	return true
}

func (m *mockAuthContext) CanPublishModuleVersion(namespace string) bool {
	if m.config.IsAdmin {
		return true
	}
	if perm, ok := m.config.Permissions[namespace]; ok {
		return perm == "full" || perm == "publish"
	}
	return false
}

func (m *mockAuthContext) CanUploadModuleVersion(namespace string) bool {
	if m.config.IsAdmin {
		return true
	}
	if perm, ok := m.config.Permissions[namespace]; ok {
		return perm == "full" || perm == "upload"
	}
	return false
}

func (m *mockAuthContext) CheckNamespaceAccess(permissionType, namespace string) bool {
	if m.config.IsAdmin {
		return true
	}
	if perm, ok := m.config.Permissions[namespace]; ok {
		return perm == "full" || perm == permissionType
	}
	return false
}

func (m *mockAuthContext) GetAllNamespacePermissions() map[string]string {
	return m.config.Permissions
}

func (m *mockAuthContext) GetUsername() string {
	return m.config.Username
}

func (m *mockAuthContext) GetUserGroupNames() []string {
	return m.config.UserGroups
}

func (m *mockAuthContext) CanAccessReadAPI() bool {
	return m.config.CanAccessReadAPI
}

func (m *mockAuthContext) CanAccessTerraformAPI() bool {
	return m.config.IsAuthenticated
}

func (m *mockAuthContext) GetTerraformAuthToken() string {
	return ""
}

func (m *mockAuthContext) GetProviderData() map[string]interface{} {
	return map[string]interface{}{
		"mock_auth": true,
	}
}

func (m *mockAuthContext) GetProviderType() auth.AuthMethodType {
	return m.config.AuthMethod
}

// SetupMockAuth creates a mock authentication middleware for testing
func SetupMockAuth(config MockAuthConfig) func(http.Handler) http.Handler {
	authContext := &mockAuthContext{config: config}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Store the mock auth context in the request context
			ctx := context.WithValue(r.Context(), "authContext", authContext)
			ctx = context.WithValue(ctx, "authenticated", config.IsAuthenticated)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// WithAdminAuth returns a handler with admin authentication mocked
func WithAdminAuth(handler http.Handler) http.Handler {
	return SetupMockAuth(MockAuthConfig{
		AuthMethod:       auth.AuthMethodAdminApiKey,
		IsAuthenticated:  true,
		IsAdmin:          true,
		IsBuiltInAdmin:   false,
		Username:         "admin-test-user",
		UserGroups:       []string{},
		Permissions:      map[string]string{},
		CanAccessReadAPI: true,
	})(handler)
}

// WithNoAuth returns a handler without authentication
func WithNoAuth(handler http.Handler) http.Handler {
	return SetupMockAuth(MockAuthConfig{
		AuthMethod:       auth.AuthMethodNotAuthenticated,
		IsAuthenticated:  false,
		IsAdmin:          false,
		IsBuiltInAdmin:   false,
		Username:         "",
		UserGroups:       []string{},
		Permissions:      map[string]string{},
		CanAccessReadAPI: true,
	})(handler)
}

// WithNamespacePermission returns a handler with specific namespace permissions
func WithNamespacePermission(handler http.Handler, namespace, permission string) http.Handler {
	return SetupMockAuth(MockAuthConfig{
		AuthMethod:       auth.AuthMethodAdminApiKey,
		IsAuthenticated:  true,
		IsAdmin:          false,
		IsBuiltInAdmin:   false,
		Username:         "test-user",
		UserGroups:       []string{"test-group"},
		Permissions:      map[string]string{namespace: permission},
		CanAccessReadAPI: true,
	})(handler)
}

// WithUploadAuth returns a handler with upload permissions
func WithUploadAuth(handler http.Handler) http.Handler {
	return SetupMockAuth(MockAuthConfig{
		AuthMethod:       auth.AuthMethodUploadApiKey,
		IsAuthenticated:  true,
		IsAdmin:          false,
		IsBuiltInAdmin:   false,
		Username:         "upload-api-key",
		UserGroups:       []string{},
		Permissions:      map[string]string{},
		CanAccessReadAPI: true,
	})(handler)
}

// WithPublishAuth returns a handler with publish permissions
func WithPublishAuth(handler http.Handler) http.Handler {
	return SetupMockAuth(MockAuthConfig{
		AuthMethod:       auth.AuthMethodPublishApiKey,
		IsAuthenticated:  true,
		IsAdmin:          false,
		IsBuiltInAdmin:   false,
		Username:         "publish-api-key",
		UserGroups:       []string{},
		Permissions:      map[string]string{},
		CanAccessReadAPI: true,
	})(handler)
}

// GetAuthContextFromRequest retrieves the auth context from a request
// Used in tests to verify authentication state
func GetAuthContextFromRequest(r *http.Request) auth.AuthContext {
	if authContext, ok := r.Context().Value("authContext").(auth.AuthContext); ok {
		return authContext
	}
	// Return a mock not authenticated context
	return &mockAuthContext{config: MockAuthConfig{
		IsAuthenticated:  false,
		IsAdmin:          false,
		Username:         "",
		CanAccessReadAPI: true,
	}}
}

// CreateAuthenticatedRequest creates an HTTP request with mock authentication context
func CreateAuthenticatedRequest(req *http.Request, config MockAuthConfig) *http.Request {
	authContext := &mockAuthContext{config: config}
	ctx := context.WithValue(req.Context(), "authContext", authContext)
	ctx = context.WithValue(ctx, "authenticated", config.IsAuthenticated)
	return req.WithContext(ctx)
}
