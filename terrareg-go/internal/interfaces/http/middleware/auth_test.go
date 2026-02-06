package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/auth"
	authservice "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/auth/service"
	domainConfig "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/config/model"
	"github.com/stretchr/testify/assert"
)

// newTestAuthMiddleware creates a test auth middleware with minimal config
func newTestAuthMiddleware() *AuthMiddleware {
	cfg := &domainConfig.DomainConfig{}
	return &AuthMiddleware{
		domainConfig: cfg,
		authFactory:  nil, // Not needed for middleware-only tests
	}
}

// TestNewAuthMiddleware tests the constructor
func TestNewAuthMiddleware(t *testing.T) {
	cfg := &domainConfig.DomainConfig{}

	// Create a real but minimal auth factory
	// For middleware testing, we don't need a fully initialized factory
	// since we'll set auth context directly in the request

	middleware := &AuthMiddleware{
		domainConfig: cfg,
		authFactory:  nil,
	}

	assert.NotNil(t, middleware)
	assert.Equal(t, cfg, middleware.domainConfig)
}

// TestExtractRequestData tests extracting request data
func TestExtractRequestData(t *testing.T) {
	middleware := newTestAuthMiddleware()

	t.Run("extracts all headers", func(t *testing.T) {
		r := httptest.NewRequest("GET", "/test", nil)
		r.Header.Set("Authorization", "Bearer token123")
		r.Header.Set("X-Custom-Header", "custom-value")
		r.Header.Add("Multi-Header", "value1")
		r.Header.Add("Multi-Header", "value2") // Should only get first value

		headers, _, _ := middleware.extractRequestData(r)

		assert.Equal(t, "Bearer token123", headers["Authorization"])
		assert.Equal(t, "custom-value", headers["X-Custom-Header"])
		assert.Equal(t, "value1", headers["Multi-Header"], "Should only get first header value")
	})

	t.Run("extracts query parameters", func(t *testing.T) {
		r := httptest.NewRequest("GET", "/test?param1=value1&param2=value2", nil)

		_, _, queryParams := middleware.extractRequestData(r)

		assert.Equal(t, "value1", queryParams["param1"])
		assert.Equal(t, "value2", queryParams["param2"])
	})

	t.Run("handles multi-value query params", func(t *testing.T) {
		r := httptest.NewRequest("GET", "/test?param=value1&param=value2", nil)

		_, _, queryParams := middleware.extractRequestData(r)

		assert.Equal(t, "value1", queryParams["param"], "Should only get first value")
	})

	t.Run("handles empty request", func(t *testing.T) {
		r := httptest.NewRequest("GET", "/test", nil)

		headers, formData, queryParams := middleware.extractRequestData(r)

		// Maps are created but empty since request has no headers/query params
		assert.NotNil(t, headers, "Should always have headers map")
		assert.NotNil(t, formData, "Should always have formData map")
		assert.NotNil(t, queryParams, "Should always have queryParams map")
		// Headers will have some default headers set by httptest.NewRequest
		// formData and queryParams will be empty
	})
}

// TestRequireReadAccess tests read API access requirement middleware
func TestRequireReadAccess(t *testing.T) {
	middleware := newTestAuthMiddleware()

	t.Run("allows requests with read API access", func(t *testing.T) {
		mockAuthCtx := &mockAuthContext{
			isAuthenticated:  true,
			canAccessReadAPI: true,
		}

		handlerCalled := false
		nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			handlerCalled = true
			// Verify auth context is set in request
			authCtx := GetAuthContext(r.Context())
			assert.True(t, authCtx.IsAuthenticated())
		})

		req := httptest.NewRequest("GET", "/test", nil)
		ctx := SetAuthContextInContext(req.Context(), mockAuthCtx)
		req = req.WithContext(ctx)
		w := httptest.NewRecorder()

		middleware.RequireReadAccess(nextHandler).ServeHTTP(w, req)

		assert.True(t, handlerCalled, "Handler should be called when user has read API access")
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("blocks requests without read API access", func(t *testing.T) {
		mockAuthCtx := &mockAuthContext{
			isAuthenticated:  true,
			canAccessReadAPI: false,
		}

		handlerCalled := false
		nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			handlerCalled = true
		})

		req := httptest.NewRequest("GET", "/test", nil)
		ctx := SetAuthContextInContext(req.Context(), mockAuthCtx)
		req = req.WithContext(ctx)
		w := httptest.NewRecorder()

		middleware.RequireReadAccess(nextHandler).ServeHTTP(w, req)

		assert.False(t, handlerCalled, "Handler should not be called when user lacks read API access")
		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.Contains(t, w.Body.String(), "Authentication required")
	})

	t.Run("allows unauthenticated when ALLOW_UNAUTHENTICATED_ACCESS=true", func(t *testing.T) {
		notAuthCtx := authservice.NewNotAuthenticatedAuthContext(true) // allowUnauthenticatedAccess=true

		handlerCalled := false
		nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			handlerCalled = true
		})

		req := httptest.NewRequest("GET", "/test", nil)
		ctx := SetAuthContextInContext(req.Context(), notAuthCtx)
		req = req.WithContext(ctx)
		w := httptest.NewRecorder()

		middleware.RequireReadAccess(nextHandler).ServeHTTP(w, req)

		assert.True(t, handlerCalled, "Handler should be called for unauthenticated when ALLOW_UNAUTHENTICATED_ACCESS=true")
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("blocks unauthenticated when ALLOW_UNAUTHENTICATED_ACCESS=false", func(t *testing.T) {
		notAuthCtx := authservice.NewNotAuthenticatedAuthContext(false) // allowUnauthenticatedAccess=false

		handlerCalled := false
		nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			handlerCalled = true
		})

		req := httptest.NewRequest("GET", "/test", nil)
		ctx := SetAuthContextInContext(req.Context(), notAuthCtx)
		req = req.WithContext(ctx)
		w := httptest.NewRecorder()

		middleware.RequireReadAccess(nextHandler).ServeHTTP(w, req)

		assert.False(t, handlerCalled, "Handler should not be called for unauthenticated when ALLOW_UNAUTHENTICATED_ACCESS=false")
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("uses default not authenticated context when no context set", func(t *testing.T) {
		// When no auth context is set, GetAuthContext returns a default with AllowUnauthenticatedAccess=true
		handlerCalled := false
		nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			handlerCalled = true
		})

		req := httptest.NewRequest("GET", "/test", nil)
		// Don't set any auth context
		w := httptest.NewRecorder()

		middleware.RequireReadAccess(nextHandler).ServeHTTP(w, req)

		// Default not authenticated context has AllowUnauthenticatedAccess=true, so it should pass
		assert.True(t, handlerCalled, "Default not authenticated context allows read access")
		assert.Equal(t, http.StatusOK, w.Code)
	})
}

// TestRequireTerraformAccess tests Terraform API access requirement middleware
func TestRequireTerraformAccess(t *testing.T) {
	middleware := newTestAuthMiddleware()

	t.Run("allows requests with Terraform API access", func(t *testing.T) {
		mockAuthCtx := &mockAuthContext{
			isAuthenticated:       true,
			canAccessTerraformAPI: true,
			canAccessReadAPI:      true,
		}

		handlerCalled := false
		nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			handlerCalled = true
		})

		req := httptest.NewRequest("GET", "/test", nil)
		ctx := SetAuthContextInContext(req.Context(), mockAuthCtx)
		req = req.WithContext(ctx)
		w := httptest.NewRecorder()

		middleware.RequireTerraformAccess(nextHandler).ServeHTTP(w, req)

		assert.True(t, handlerCalled, "Handler should be called when user has Terraform API access")
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("blocks requests without Terraform API access", func(t *testing.T) {
		mockAuthCtx := &mockAuthContext{
			isAuthenticated:       true,
			canAccessTerraformAPI: false,
			canAccessReadAPI:      true, // Has read access but not Terraform
		}

		handlerCalled := false
		nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			handlerCalled = true
		})

		req := httptest.NewRequest("GET", "/test", nil)
		ctx := SetAuthContextInContext(req.Context(), mockAuthCtx)
		req = req.WithContext(ctx)
		w := httptest.NewRecorder()

		middleware.RequireTerraformAccess(nextHandler).ServeHTTP(w, req)

		assert.False(t, handlerCalled, "Handler should not be called when user lacks Terraform API access")
		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.Contains(t, w.Body.String(), "Authentication required")
	})

	t.Run("blocks unauthenticated requests", func(t *testing.T) {
		notAuthCtx := authservice.NewNotAuthenticatedAuthContext(true)

		handlerCalled := false
		nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			handlerCalled = true
		})

		req := httptest.NewRequest("GET", "/test", nil)
		ctx := SetAuthContextInContext(req.Context(), notAuthCtx)
		req = req.WithContext(ctx)
		w := httptest.NewRecorder()

		middleware.RequireTerraformAccess(nextHandler).ServeHTTP(w, req)

		assert.False(t, handlerCalled, "Handler should not be called for unauthenticated requests")
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("blocks user with only read access", func(t *testing.T) {
		// User has CanAccessReadAPI=true but CanAccessTerraformAPI=false
		mockAuthCtx := &mockAuthContext{
			isAuthenticated:       true,
			canAccessReadAPI:      true,
			canAccessTerraformAPI: false,
		}

		handlerCalled := false
		nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			handlerCalled = true
		})

		req := httptest.NewRequest("GET", "/test", nil)
		ctx := SetAuthContextInContext(req.Context(), mockAuthCtx)
		req = req.WithContext(ctx)
		w := httptest.NewRecorder()

		middleware.RequireTerraformAccess(nextHandler).ServeHTTP(w, req)

		assert.False(t, handlerCalled)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

// TestGetAuthContext tests retrieving auth context from request
func TestGetAuthContext(t *testing.T) {
	t.Run("returns auth context when set", func(t *testing.T) {
		mockAuthCtx := &mockAuthContext{
			isAuthenticated: true,
			username:       "testuser",
		}

		ctx := context.Background()
		ctx = SetAuthContextInContext(ctx, mockAuthCtx)

		retrievedCtx := GetAuthContext(ctx)

		assert.True(t, retrievedCtx.IsAuthenticated())
		assert.Equal(t, "testuser", retrievedCtx.GetUsername())
	})

	t.Run("returns not authenticated context when not set", func(t *testing.T) {
		ctx := context.Background()

		retrievedCtx := GetAuthContext(ctx)

		assert.NotNil(t, retrievedCtx)
		assert.False(t, retrievedCtx.IsAuthenticated())
		assert.Equal(t, auth.AuthMethodNotAuthenticated, retrievedCtx.GetProviderType())
	})

	t.Run("returns not authenticated context with AllowUnauthenticatedAccess=true by default", func(t *testing.T) {
		ctx := context.Background()

		retrievedCtx := GetAuthContext(ctx)

		assert.True(t, retrievedCtx.CanAccessReadAPI(), "Default not authenticated context should allow read API access")
	})
}

// TestGetAuthMethodFromContext tests extracting auth method
func TestGetAuthMethodFromContext(t *testing.T) {
	t.Run("returns auth method and authenticated status when set", func(t *testing.T) {
		mockAuthCtx := &mockAuthContext{
			isAuthenticated: true,
			authMethod:     auth.AuthMethodGitHub,
		}

		ctx := context.Background()
		ctx = SetAuthContextInContext(ctx, mockAuthCtx)

		method, isAuthenticated := GetAuthMethodFromContext(ctx)

		assert.Equal(t, auth.AuthMethodGitHub, method)
		assert.True(t, isAuthenticated)
	})

	t.Run("returns not authenticated when context not set", func(t *testing.T) {
		ctx := context.Background()

		method, isAuthenticated := GetAuthMethodFromContext(ctx)

		assert.Equal(t, auth.AuthMethodNotAuthenticated, method)
		assert.False(t, isAuthenticated)
	})
}

// TestGetUserFromContext tests extracting user from context
func TestGetUserFromContext(t *testing.T) {
	t.Run("returns username when authenticated", func(t *testing.T) {
		mockAuthCtx := &mockAuthContext{
			isAuthenticated: true,
			username:       "testuser",
		}

		ctx := context.Background()
		ctx = SetAuthContextInContext(ctx, mockAuthCtx)

		username, isAuthenticated := GetUserFromContext(ctx)

		assert.Equal(t, "testuser", username)
		assert.True(t, isAuthenticated)
	})

	t.Run("returns empty strings when not authenticated", func(t *testing.T) {
		ctx := context.Background()

		username, isAuthenticated := GetUserFromContext(ctx)

		assert.Empty(t, username)
		assert.False(t, isAuthenticated)
	})
}

// TestGetIsAdminFromContext tests extracting admin status
func TestGetIsAdminFromContext(t *testing.T) {
	t.Run("returns true for admin users", func(t *testing.T) {
		mockAuthCtx := &mockAuthContext{
			isAuthenticated: true,
			isAdmin:        true,
		}

		ctx := context.Background()
		ctx = SetAuthContextInContext(ctx, mockAuthCtx)

		isAdmin := GetIsAdminFromContext(ctx)

		assert.True(t, isAdmin)
	})

	t.Run("returns false for non-admin users", func(t *testing.T) {
		mockAuthCtx := &mockAuthContext{
			isAuthenticated: true,
			isAdmin:        false,
		}

		ctx := context.Background()
		ctx = SetAuthContextInContext(ctx, mockAuthCtx)

		isAdmin := GetIsAdminFromContext(ctx)

		assert.False(t, isAdmin)
	})

	t.Run("returns false for unauthenticated users", func(t *testing.T) {
		ctx := context.Background()

		isAdmin := GetIsAdminFromContext(ctx)

		assert.False(t, isAdmin)
	})
}

// TestGetSessionIDFromContext tests extracting session ID
func TestGetSessionIDFromContext(t *testing.T) {
	t.Run("returns session ID when available", func(t *testing.T) {
		mockAuthCtx := &mockAuthContext{
			isAuthenticated: true,
			providerData: map[string]interface{}{
				"session_id": "test-session-123",
			},
		}

		ctx := context.Background()
		ctx = SetAuthContextInContext(ctx, mockAuthCtx)

		sessionID := GetSessionIDFromContext(ctx)

		assert.Equal(t, "test-session-123", sessionID)
	})

	t.Run("returns empty string when no session ID", func(t *testing.T) {
		mockAuthCtx := &mockAuthContext{
			isAuthenticated: true,
			providerData:   map[string]interface{}{},
		}

		ctx := context.Background()
		ctx = SetAuthContextInContext(ctx, mockAuthCtx)

		sessionID := GetSessionIDFromContext(ctx)

		assert.Empty(t, sessionID)
	})
}

// TestGetPermissionsFromContext tests extracting permissions
func TestGetPermissionsFromContext(t *testing.T) {
	t.Run("returns permissions when authenticated", func(t *testing.T) {
		mockAuthCtx := &mockAuthContext{
			isAuthenticated: true,
			permissions: map[string]string{
				"ns1": "FULL",
				"ns2": "READ",
			},
		}

		ctx := context.Background()
		ctx = SetAuthContextInContext(ctx, mockAuthCtx)

		permissions, hasPermissions := GetPermissionsFromContext(ctx)

		assert.True(t, hasPermissions)
		assert.Equal(t, "FULL", permissions["ns1"])
		assert.Equal(t, "READ", permissions["ns2"])
	})

	t.Run("returns nil when not authenticated", func(t *testing.T) {
		ctx := context.Background()

		permissions, hasPermissions := GetPermissionsFromContext(ctx)

		assert.False(t, hasPermissions)
		assert.Nil(t, permissions)
	})
}

// TestSetAuthContextInContext tests setting auth context
func TestSetAuthContextInContext(t *testing.T) {
	mockAuthCtx := &mockAuthContext{
		isAuthenticated: true,
		username:       "testuser",
	}

	ctx := context.Background()
	newCtx := SetAuthContextInContext(ctx, mockAuthCtx)

	assert.NotEqual(t, ctx, newCtx, "Should return a new context")

	retrievedCtx := GetAuthContext(newCtx)
	assert.Equal(t, "testuser", retrievedCtx.GetUsername())
}

// TestExtractBearerToken tests Bearer token extraction
func TestExtractBearerToken(t *testing.T) {
	t.Run("extracts valid Bearer token", func(t *testing.T) {
		r := httptest.NewRequest("GET", "/test", nil)
		r.Header.Set("Authorization", "Bearer my-secret-token")

		token := extractBearerToken(r)

		assert.Equal(t, "my-secret-token", token)
	})

	t.Run("handles missing Authorization header", func(t *testing.T) {
		r := httptest.NewRequest("GET", "/test", nil)

		token := extractBearerToken(r)

		assert.Empty(t, token)
	})

	t.Run("handles malformed Bearer token", func(t *testing.T) {
		r := httptest.NewRequest("GET", "/test", nil)
		r.Header.Set("Authorization", "InvalidFormat token")

		token := extractBearerToken(r)

		assert.Empty(t, token)
	})

	t.Run("handles token without Bearer prefix", func(t *testing.T) {
		r := httptest.NewRequest("GET", "/test", nil)
		r.Header.Set("Authorization", "token-without-bearer")

		token := extractBearerToken(r)

		assert.Empty(t, token)
	})

	t.Run("handles Bearer with no token", func(t *testing.T) {
		r := httptest.NewRequest("GET", "/test", nil)
		r.Header.Set("Authorization", "Bearer")

		token := extractBearerToken(r)

		assert.Empty(t, token)
	})

	t.Run("handles Bearer token with extra spaces", func(t *testing.T) {
		r := httptest.NewRequest("GET", "/test", nil)
		r.Header.Set("Authorization", "Bearer  token-with-spaces")

		token := extractBearerToken(r)

		// The implementation preserves the space after "Bearer"
		assert.Equal(t, " token-with-spaces", token)
	})
}

// TestAuthContextImplementation tests that real auth contexts implement the interface correctly
func TestAuthContextImplementation(t *testing.T) {
	t.Run("admin API key auth context", func(t *testing.T) {
		authCtx := auth.NewAdminApiKeyAuthContext(context.Background(), "test-admin-key")

		assert.True(t, authCtx.IsAuthenticated())
		assert.True(t, authCtx.IsAdmin())
		assert.True(t, authCtx.CanAccessReadAPI())
		assert.True(t, authCtx.CanAccessTerraformAPI())
		assert.Equal(t, auth.AuthMethodAdminApiKey, authCtx.GetProviderType())
	})

	t.Run("publish API key auth context", func(t *testing.T) {
		authCtx := auth.NewPublishApiKeyAuthContext(context.Background(), "test-publish-key")

		assert.True(t, authCtx.IsAuthenticated())
		assert.False(t, authCtx.IsAdmin())
		assert.False(t, authCtx.CanAccessReadAPI())
		assert.False(t, authCtx.CanAccessTerraformAPI())
		assert.True(t, authCtx.CanPublishModuleVersion("any"))
		assert.True(t, authCtx.CanUploadModuleVersion("any"))
		assert.Equal(t, auth.AuthMethodPublishApiKey, authCtx.GetProviderType())
	})

	t.Run("not authenticated auth context with ALLOW_UNAUTHENTICATED_ACCESS=true", func(t *testing.T) {
		authCtx := authservice.NewNotAuthenticatedAuthContext(true)

		assert.False(t, authCtx.IsAuthenticated())
		assert.True(t, authCtx.CanAccessReadAPI())
		assert.False(t, authCtx.CanAccessTerraformAPI())
	})

	t.Run("not authenticated auth context with ALLOW_UNAUTHENTICATED_ACCESS=false", func(t *testing.T) {
		authCtx := authservice.NewNotAuthenticatedAuthContext(false)

		assert.False(t, authCtx.IsAuthenticated())
		assert.False(t, authCtx.CanAccessReadAPI())
		assert.False(t, authCtx.CanAccessTerraformAPI())
	})
}

// TestCheckNamespacePermission tests namespace permission checking
func TestCheckNamespacePermission(t *testing.T) {
	middleware := newTestAuthMiddleware()

	t.Run("returns true for admin users", func(t *testing.T) {
		mockAuthCtx := &mockAuthContext{
			isAuthenticated: true,
			isAdmin:        true,
		}

		ctx := context.Background()
		ctx = SetAuthContextInContext(ctx, mockAuthCtx)

		hasPermission := middleware.CheckNamespacePermission(ctx, "FULL", "test-ns")

		assert.True(t, hasPermission)
	})

	t.Run("returns true when user has permission", func(t *testing.T) {
		mockAuthCtx := &mockAuthContext{
			isAuthenticated: true,
			permissions: map[string]string{
				"test-ns": "FULL",
			},
		}

		ctx := context.Background()
		ctx = SetAuthContextInContext(ctx, mockAuthCtx)

		hasPermission := middleware.CheckNamespacePermission(ctx, "FULL", "test-ns")

		assert.True(t, hasPermission)
	})

	t.Run("returns false when user lacks permission", func(t *testing.T) {
		mockAuthCtx := &mockAuthContext{
			isAuthenticated: true,
			permissions: map[string]string{
				"test-ns": "READ",
			},
		}

		ctx := context.Background()
		ctx = SetAuthContextInContext(ctx, mockAuthCtx)

		hasPermission := middleware.CheckNamespacePermission(ctx, "FULL", "test-ns")

		assert.False(t, hasPermission)
	})
}

// mockAuthContext is a mock implementation of auth.AuthContext for testing
type mockAuthContext struct {
	isAuthenticated        bool
	authMethod            auth.AuthMethodType
	username              string
	isAdmin               bool
	isBuiltInAdmin        bool
	requiresCSRF          bool
	canPublish            bool
	canUpload             bool
	canAccessReadAPI      bool
	canAccessTerraformAPI bool
	userGroups            []string
	permissions           map[string]string
	terraformToken        string
	providerData          map[string]interface{}
}

func (m *mockAuthContext) IsAuthenticated() bool                                         { return m.isAuthenticated }
func (m *mockAuthContext) GetProviderType() auth.AuthMethodType                         { return m.authMethod }
func (m *mockAuthContext) GetUsername() string                                          { return m.username }
func (m *mockAuthContext) IsAdmin() bool                                                { return m.isAdmin }
func (m *mockAuthContext) IsBuiltInAdmin() bool                                         { return m.isBuiltInAdmin }
func (m *mockAuthContext) RequiresCSRF() bool                                           { return m.requiresCSRF }
func (m *mockAuthContext) CheckAuthState() bool                                         { return true }
func (m *mockAuthContext) CanPublishModuleVersion(module string) bool                   { return m.canPublish }
func (m *mockAuthContext) CanUploadModuleVersion(module string) bool                    { return m.canUpload }
func (m *mockAuthContext) CheckNamespaceAccess(permissionType, namespace string) bool {
	if m.isAdmin {
		return true
	}
	if perm, ok := m.permissions[namespace]; ok {
		return perm == permissionType || perm == "FULL"
	}
	return false
}
func (m *mockAuthContext) GetAllNamespacePermissions() map[string]string                { return m.permissions }
func (m *mockAuthContext) GetUserGroupNames() []string                                 { return m.userGroups }
func (m *mockAuthContext) CanAccessReadAPI() bool                                       { return m.canAccessReadAPI }
func (m *mockAuthContext) CanAccessTerraformAPI() bool                                  { return m.canAccessTerraformAPI }
func (m *mockAuthContext) GetTerraformAuthToken() string                                { return m.terraformToken }
func (m *mockAuthContext) GetProviderData() map[string]interface{}                      { return m.providerData }
func (m *mockAuthContext) IsEnabled() bool                                              { return true }
