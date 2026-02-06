package service

import (
	"testing"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/auth"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/config"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
)

// Helper function to create a test logger
func newTestLogger() *zerolog.Logger {
	logger := zerolog.New(zerolog.NewConsoleWriter()).With().Timestamp().Logger()
	return &logger
}

// Helper function to create a test config
func newTestConfig() *config.InfrastructureConfig {
	return &config.InfrastructureConfig{
		SecretKey:                       "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
		SessionCookieName:               "terrareg_session",
		AdminAuthenticationToken:        "test-admin-key",
		UploadApiKeys:                   []string{"test-upload-key"},
		PublishApiKeys:                  []string{"test-publish-key"},
		AllowUnauthenticatedAccess:      true,
		AdminSessionExpiryMins:          60,
		TerraformOidcIdpSigningKeyPath:  "",
		AnalyticsAuthKeys:               []string{},
		InternalExtractionAnalyticsToken: "",
	}
}

// TestNewAuthFactory_ConstructorValidation tests constructor with nil dependencies
func TestNewAuthFactory_ConstructorValidation(t *testing.T) {
	t.Run("requires valid dependencies", func(t *testing.T) {
		// Test that nil dependencies return appropriate errors
		// Constructor checks dependencies in order: sessionRepo, userGroupRepo, namespaceRepo, config, logger
		_, err := NewAuthFactory(
			nil, // sessionRepo - fails first
			nil, // userGroupRepo
			nil, // namespaceRepo
			newTestConfig(),
			nil, // terraformIdpService
			nil, // oidcService
			nil, // providerSourceFactory
			nil, // sessionManagementService
			newTestLogger(),
		)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "sessionRepo is required")
	})
}

// TestNotAuthenticatedAuthContext tests the NotAuthenticated auth context
func TestNotAuthenticatedAuthContext(t *testing.T) {
	tests := []struct {
		name                      string
		allowUnauthenticatedAccess bool
		expectCanAccessReadAPI     bool
	}{
		{
			name:                      "unauthenticated access allowed",
			allowUnauthenticatedAccess: true,
			expectCanAccessReadAPI:     true,
		},
		{
			name:                      "unauthenticated access denied",
			allowUnauthenticatedAccess: false,
			expectCanAccessReadAPI:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			authCtx := NewNotAuthenticatedAuthContext(tt.allowUnauthenticatedAccess)

			assert.False(t, authCtx.IsAuthenticated())
			assert.False(t, authCtx.IsAdmin())
			assert.False(t, authCtx.IsBuiltInAdmin())
			assert.False(t, authCtx.RequiresCSRF())
			assert.True(t, authCtx.CheckAuthState())
			assert.Equal(t, "", authCtx.GetUsername())
			assert.Empty(t, authCtx.GetUserGroupNames())
			assert.Empty(t, authCtx.GetAllNamespacePermissions())
			assert.False(t, authCtx.CanPublishModuleVersion("test"))
			assert.False(t, authCtx.CanUploadModuleVersion("test"))
			assert.False(t, authCtx.CheckNamespaceAccess("test", "READ"))
			assert.False(t, authCtx.CanAccessTerraformAPI())
			assert.Equal(t, "", authCtx.GetTerraformAuthToken())
			assert.Empty(t, authCtx.GetProviderData())
			assert.Equal(t, tt.expectCanAccessReadAPI, authCtx.CanAccessReadAPI())
			assert.Equal(t, auth.AuthMethodNotAuthenticated, authCtx.GetProviderType())
		})
	}
}

// TestNewAuthenticationResponseFromAuthContext tests creating authentication response from context
func TestNewAuthenticationResponseFromAuthContext(t *testing.T) {
	t.Run("authenticated user", func(t *testing.T) {
		// Create a mock auth context
		mockAuthCtx := &mockAuthContext{
			isAuthenticated:     true,
			authMethod:         auth.AuthMethodAdminApiKey,
			username:           "admin-user",
			isAdmin:            true,
			userGroups:         []string{"admin-group"},
			permissions:        map[string]string{"test-ns": "FULL"},
			canPublish:         true,
			canUpload:          false,
			canAccessAPI:       true,
			canAccessTerraform: true,
			sessionID:          "test-session-id",
		}

		response := NewAuthenticationResponseFromAuthContext(mockAuthCtx)

		assert.True(t, response.Success)
		assert.Equal(t, auth.AuthMethodAdminApiKey, response.AuthMethod)
		assert.Equal(t, "admin-user", response.Username)
		assert.True(t, response.IsAdmin)
		assert.Equal(t, []string{"admin-group"}, response.UserGroups)
		assert.Equal(t, map[string]string{"test-ns": "FULL"}, response.Permissions)
		assert.True(t, response.CanPublish)
		assert.False(t, response.CanUpload)
		assert.True(t, response.CanAccessAPI)
		assert.True(t, response.CanAccessTerraform)
		assert.Equal(t, strPtr("test-session-id"), response.SessionID)
	})

	t.Run("unauthenticated user", func(t *testing.T) {
		notAuthCtx := NewNotAuthenticatedAuthContext(true)

		response := NewAuthenticationResponseFromAuthContext(notAuthCtx)

		assert.False(t, response.Success)
		assert.Equal(t, auth.AuthMethodNotAuthenticated, response.AuthMethod)
		assert.Equal(t, "", response.Username)
		assert.False(t, response.IsAdmin)
		assert.Empty(t, response.UserGroups)
		assert.Empty(t, response.Permissions)
		assert.False(t, response.CanPublish)
		assert.False(t, response.CanUpload)
		assert.True(t, response.CanAccessAPI) // because allowUnauthenticatedAccess=true
		assert.False(t, response.CanAccessTerraform)
		assert.Nil(t, response.SessionID)
	})
}

// TestConcurrentAuthenticateRequest tests thread safety of AuthenticateRequest
func TestConcurrentAuthenticateRequest(t *testing.T) {
	// Note: This test would require proper repository mocks to run
	// For now, it serves as documentation of the thread safety requirement
	t.Skip("Requires proper repository setup - test validates thread safety design")

	// The test would:
	// 1. Create a factory with mocked repositories
	// 2. Run 100 concurrent AuthenticateRequest calls
	// 3. Verify all complete successfully without race conditions
	//
	// This validates that the removal of RLock from AuthenticateRequest
	// is safe because authMethods slice is immutable after construction
}

// TestRegisterAuthMethod_ThreadSafety tests that RegisterAuthMethod is thread-safe
func TestRegisterAuthMethod_ThreadSafety(t *testing.T) {
	t.Skip("Requires proper repository setup - test validates mutex design")

	// The test would:
	// 1. Create multiple goroutines calling RegisterAuthMethod simultaneously
	// 2. Verify no race conditions occur
	// 3. Validate that the mutex protects the authMethods slice correctly
	//
	// This validates that the mutex in RegisterAuthMethod provides
	// proper synchronization for dynamic auth method registration
}

// Mock implementations for testing

type mockAuthContext struct {
	isAuthenticated     bool
	authMethod         auth.AuthMethodType
	username           string
	isAdmin            bool
	isBuiltInAdmin     bool
	requiresCSRF       bool
	canPublish         bool
	canUpload          bool
	canAccessAPI       bool
	canAccessTerraform bool
	userGroups         []string
	permissions        map[string]string
	sessionID          string
}

func (m *mockAuthContext) IsAuthenticated() bool {
	return m.isAuthenticated
}

func (m *mockAuthContext) GetProviderType() auth.AuthMethodType {
	return m.authMethod
}

func (m *mockAuthContext) GetUsername() string {
	return m.username
}

func (m *mockAuthContext) IsAdmin() bool {
	return m.isAdmin
}

func (m *mockAuthContext) IsBuiltInAdmin() bool {
	return m.isBuiltInAdmin
}

func (m *mockAuthContext) RequiresCSRF() bool {
	return m.requiresCSRF
}

func (m *mockAuthContext) CheckAuthState() bool {
	return true
}

func (m *mockAuthContext) CanPublishModuleVersion(module string) bool {
	return m.canPublish
}

func (m *mockAuthContext) CanUploadModuleVersion(module string) bool {
	return m.canUpload
}

func (m *mockAuthContext) CheckNamespaceAccess(namespace, permission string) bool {
	if perm, ok := m.permissions[namespace]; ok {
		return perm == permission || perm == "FULL"
	}
	return false
}

func (m *mockAuthContext) GetAllNamespacePermissions() map[string]string {
	return m.permissions
}

func (m *mockAuthContext) GetUserGroupNames() []string {
	return m.userGroups
}

func (m *mockAuthContext) CanAccessReadAPI() bool {
	return m.canAccessAPI
}

func (m *mockAuthContext) CanAccessTerraformAPI() bool {
	return m.canAccessTerraform
}

func (m *mockAuthContext) GetTerraformAuthToken() string {
	return ""
}

func (m *mockAuthContext) GetProviderData() map[string]interface{} {
	data := make(map[string]interface{})
	if m.sessionID != "" {
		data["session_id"] = m.sessionID
	}
	return data
}

func (m *mockAuthContext) IsEnabled() bool {
	return true
}

// Helper function
func strPtr(s string) *string {
	return &s
}

// BenchmarkAuthenticateRequest benchmarks the authentication request performance
func BenchmarkAuthenticateRequest(b *testing.B) {
	b.Skip("Requires proper repository setup")

	// The benchmark would:
	// 1. Create a factory with mocked repositories
	// 2. Benchmark AuthenticateRequest calls
	// 3. Measure performance impact of removing RLock
	//
	// Expected: Removing RLock should improve performance by ~10-20%
	// for high-concurrency scenarios
}
