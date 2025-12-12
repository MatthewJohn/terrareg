package auth

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/auth"
	domainauth "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/auth"
)

// AdminAPIKeyAuthMethodTestSuite tests the Admin API Key authentication method
func TestAdminAPIKeyAuthMethod(t *testing.T) {
	t.Run("Constructor", func(t *testing.T) {
		authMethod := auth.NewAdminApiKeyAuthMethod()

		require.NotNil(t, authMethod)
		assert.Equal(t, domainauth.AuthMethodAdminApiKey, authMethod.GetProviderType())
		assert.True(t, authMethod.IsEnabled())
		assert.False(t, authMethod.RequiresCSRF())
		assert.False(t, authMethod.IsAuthenticated()) // Initially not authenticated
		assert.True(t, authMethod.IsBuiltInAdmin())    // Should always return true for admin API key auth
	})

	t.Run("Unauthenticated State", func(t *testing.T) {
		authMethod := auth.NewAdminApiKeyAuthMethod()

		// Test unauthenticated state behavior
		assert.False(t, authMethod.CheckAuthState())
		assert.False(t, authMethod.IsAuthenticated())
		assert.False(t, authMethod.IsAdmin()) // Admin status depends on authentication
		assert.Empty(t, authMethod.GetUsername())
		assert.Empty(t, authMethod.GetTerraformAuthToken())

		// Test permissions before authentication
		assert.False(t, authMethod.CanAccessReadAPI())
		assert.False(t, authMethod.CanAccessTerraformAPI())
		assert.False(t, authMethod.CanPublishModuleVersion("testns"))
		assert.False(t, authMethod.CanUploadModuleVersion("testns"))
		assert.False(t, authMethod.CheckNamespaceAccess("FULL_ACCESS", "testns"))

		// Test provider data before authentication
		providerData := authMethod.GetProviderData()
		require.NotNil(t, providerData)
		assert.False(t, providerData["is_admin"].(bool)) // Admin status only set after authentication
		assert.Empty(t, providerData["username"])
		// API key should not be present before authentication
		assert.NotContains(t, providerData, "api_key")

		// Test user groups
		userGroups := authMethod.GetUserGroupNames()
		assert.Len(t, userGroups, 1)
		assert.Equal(t, "admin", userGroups[0])
	})

	t.Run("Successful Authentication", func(t *testing.T) {
		authMethod := auth.NewAdminApiKeyAuthMethod()

		headers := map[string]string{
			"Authorization": "Bearer admin-api-key-12345",
		}
		cookies := make(map[string]string)

		err := authMethod.Authenticate(context.Background(), headers, cookies)
		require.NoError(t, err)

		// Test authenticated state
		assert.True(t, authMethod.CheckAuthState())
		assert.True(t, authMethod.IsAuthenticated())
		assert.True(t, authMethod.IsAdmin())
		assert.Equal(t, "Admin API Key User", authMethod.GetUsername())
		assert.Equal(t, "admin-api-key-12345", authMethod.GetTerraformAuthToken())

		// Test permissions after authentication
		assert.True(t, authMethod.CanAccessReadAPI())
		assert.True(t, authMethod.CanAccessTerraformAPI())
		assert.True(t, authMethod.CanPublishModuleVersion("testns"))
		assert.True(t, authMethod.CanUploadModuleVersion("testns"))
		assert.True(t, authMethod.CheckNamespaceAccess("FULL_ACCESS", "testns"))

		// Test provider data after authentication
		providerData := authMethod.GetProviderData()
		require.NotNil(t, providerData)
		assert.True(t, providerData["is_admin"].(bool))
		assert.Equal(t, "Admin API Key User", providerData["username"])
		assert.Equal(t, "admin-api-key-12345", providerData["api_key"])

		// Test user groups (should remain the same)
		userGroups := authMethod.GetUserGroupNames()
		assert.Len(t, userGroups, 1)
		assert.Equal(t, "admin", userGroups[0])
	})

	t.Run("Authentication with Valid Keys", func(t *testing.T) {
		validKeys := []string{
			"admin-api-key-12345",
			"admin-api-key-67890",
			"admin-api-key-abcdef",
		}

		for _, apiKey := range validKeys {
			t.Run("Valid key: "+apiKey, func(t *testing.T) {
				authMethod := auth.NewAdminApiKeyAuthMethod()

				headers := map[string]string{
					"Authorization": "Bearer " + apiKey,
				}

				err := authMethod.Authenticate(context.Background(), headers, make(map[string]string))
				assert.NoError(t, err)
				assert.True(t, authMethod.IsAuthenticated())
				assert.Equal(t, apiKey, authMethod.GetTerraformAuthToken())
			})
		}
	})

	t.Run("Authentication Failure Scenarios", func(t *testing.T) {
		testCases := []struct {
			name        string
			headers     map[string]string
			expectError bool
		}{
			{
				name:        "Missing Authorization header",
				headers:     map[string]string{},
				expectError: true,
			},
			{
				name: "Empty Authorization header",
				headers: map[string]string{
					"Authorization": "",
				},
				expectError: true,
			},
			{
				name: "Missing Bearer prefix",
				headers: map[string]string{
					"Authorization": "admin-api-key-12345",
				},
				expectError: true,
			},
			{
				name: "Wrong prefix",
				headers: map[string]string{
					"Authorization": "Token admin-api-key-12345",
				},
				expectError: true,
			},
			{
				name: "Empty Bearer token",
				headers: map[string]string{
					"Authorization": "Bearer ",
				},
				expectError: true,
			},
			{
				name: "Only whitespace token",
				headers: map[string]string{
					"Authorization": "Bearer   ",
				},
				expectError: true,
			},
			{
				name: "Invalid API key",
				headers: map[string]string{
					"Authorization": "Bearer invalid-key",
				},
				expectError: true,
			},
			{
				name: "Malformed key with special chars",
				headers: map[string]string{
					"Authorization": "Bearer admin-key-with;$pecial-chars",
				},
				expectError: true,
			},
		}

		for _, test := range testCases {
			t.Run(test.name, func(t *testing.T) {
				authMethod := auth.NewAdminApiKeyAuthMethod()

				err := authMethod.Authenticate(context.Background(), test.headers, make(map[string]string))

				if test.expectError {
					assert.Error(t, err)
					assert.False(t, authMethod.IsAuthenticated())
					assert.IsType(t, &auth.AdminApiKeyError{}, err)
				} else {
					assert.NoError(t, err)
					assert.True(t, authMethod.IsAuthenticated())
				}
			})
		}
	})

	t.Run("Namespace Permissions", func(t *testing.T) {
		authMethod := auth.NewAdminApiKeyAuthMethod()

		// Before authentication - should not have access
		assert.False(t, authMethod.CanPublishModuleVersion("testns"))
		assert.False(t, authMethod.CanUploadModuleVersion("testns"))
		assert.False(t, authMethod.CheckNamespaceAccess("FULL_ACCESS", "testns"))
		assert.False(t, authMethod.CheckNamespaceAccess("READ_ACCESS", "testns"))

		// Authenticate
		headers := map[string]string{
			"Authorization": "Bearer admin-api-key-12345",
		}
		err := authMethod.Authenticate(context.Background(), headers, make(map[string]string))
		require.NoError(t, err)

		// After authentication - should have full access to all namespaces
		testNamespaces := []string{"", "testns", "prod", "dev", "private-ns"}

		for _, ns := range testNamespaces {
			t.Run("Namespace: "+ns, func(t *testing.T) {
				assert.True(t, authMethod.CanPublishModuleVersion(ns), "Should be able to publish to namespace: "+ns)
				assert.True(t, authMethod.CanUploadModuleVersion(ns), "Should be able to upload to namespace: "+ns)
				assert.True(t, authMethod.CheckNamespaceAccess("FULL_ACCESS", ns), "Should have full access to namespace: "+ns)
				assert.True(t, authMethod.CheckNamespaceAccess("READ_ACCESS", ns), "Should have read access to namespace: "+ns)
				assert.True(t, authMethod.CheckNamespaceAccess("WRITE_ACCESS", ns), "Should have write access to namespace: "+ns)
				assert.True(t, authMethod.CheckNamespaceAccess("ADMIN_ACCESS", ns), "Should have admin access to namespace: "+ns)
			})
		}

		// Test GetAllNamespacePermissions
		permissions := authMethod.GetAllNamespacePermissions()
		assert.NotNil(t, permissions)
		// For admin API key, should return empty map to signify admin access
		assert.Empty(t, permissions)
	})

	t.Run("Consistent Behavior After Authentication", func(t *testing.T) {
		authMethod := auth.NewAdminApiKeyAuthMethod()

		// Authenticate
		headers := map[string]string{
			"Authorization": "Bearer admin-api-key-12345",
		}
		err := authMethod.Authenticate(context.Background(), headers, make(map[string]string))
		require.NoError(t, err)

		// Test multiple calls to ensure consistent behavior
		for i := 0; i < 5; i++ {
			assert.True(t, authMethod.IsAuthenticated(), "Should remain authenticated on call %d", i+1)
			assert.True(t, authMethod.CheckAuthState(), "Auth state should remain valid on call %d", i+1)
			assert.True(t, authMethod.IsAdmin(), "Should remain admin on call %d", i+1)
			assert.Equal(t, "admin-api-key-12345", authMethod.GetTerraformAuthToken(), "Token should remain consistent on call %d", i+1)
			assert.Equal(t, "Admin API Key User", authMethod.GetUsername(), "Username should remain consistent on call %d", i+1)
		}
	})

	t.Run("Provider Data Consistency", func(t *testing.T) {
		authMethod := auth.NewAdminApiKeyAuthMethod()

		// Test provider data before authentication
		providerDataBefore := authMethod.GetProviderData()
		require.NotNil(t, providerDataBefore)
		assert.False(t, providerDataBefore["is_admin"].(bool)) // Admin status only set after authentication
		assert.Empty(t, providerDataBefore["username"])
		assert.NotContains(t, providerDataBefore, "api_key")

		// Authenticate
		headers := map[string]string{
			"Authorization": "Bearer admin-api-key-67890",
		}
		err := authMethod.Authenticate(context.Background(), headers, make(map[string]string))
		require.NoError(t, err)

		// Test provider data after authentication
		providerDataAfter := authMethod.GetProviderData()
		require.NotNil(t, providerDataAfter)
		assert.True(t, providerDataAfter["is_admin"].(bool))
		assert.Equal(t, "Admin API Key User", providerDataAfter["username"])
		assert.Equal(t, "admin-api-key-67890", providerDataAfter["api_key"])

		// Multiple calls should return the same data
		providerDataAgain := authMethod.GetProviderData()
		assert.Equal(t, providerDataAfter, providerDataAgain)
	})

	t.Run("Context Independence", func(t *testing.T) {
		authMethod := auth.NewAdminApiKeyAuthMethod()

		// Test with different contexts
		contexts := []context.Context{
			context.Background(),
			context.TODO(),
			context.WithValue(context.Background(), "key", "value"),
		}

		for i, ctx := range contexts {
			t.Run("Context test "+string(rune('A'+i)), func(t *testing.T) {
				headers := map[string]string{
					"Authorization": "Bearer admin-api-key-12345",
				}

				err := authMethod.Authenticate(ctx, headers, make(map[string]string))
				assert.NoError(t, err)
				assert.True(t, authMethod.IsAuthenticated())

				// Reset authentication for next test
				authMethod = auth.NewAdminApiKeyAuthMethod()
			})
		}
	})

	t.Run("IsBuiltInAdmin Always Returns True", func(t *testing.T) {
		// This is a specific test for the IsBuiltInAdmin method behavior
		authMethod := auth.NewAdminApiKeyAuthMethod()

		// Should return true even before authentication
		assert.True(t, authMethod.IsBuiltInAdmin(), "IsBuiltInAdmin should always return true")

		// Authenticate and test again
		headers := map[string]string{
			"Authorization": "Bearer admin-api-key-12345",
		}
		err := authMethod.Authenticate(context.Background(), headers, make(map[string]string))
		require.NoError(t, err)

		// Should still return true after authentication
		assert.True(t, authMethod.IsBuiltInAdmin(), "IsBuiltInAdmin should still return true after authentication")
	})
}

// TestAdminAPIKeyError tests the error type
func TestAdminAPIKeyError(t *testing.T) {
	t.Run("Error Creation and Message", func(t *testing.T) {
		err := &auth.AdminApiKeyError{Message: "test error message"}

		assert.Equal(t, "Admin API Key authentication failed: test error message", err.Error())
	})

	t.Run("Error Type Checking", func(t *testing.T) {
		authMethod := auth.NewAdminApiKeyAuthMethod()

		// Try to authenticate with invalid key
		headers := map[string]string{
			"Authorization": "Bearer invalid-key",
		}

		err := authMethod.Authenticate(context.Background(), headers, make(map[string]string))
		require.Error(t, err)

		// Check error type
		var adminErr *auth.AdminApiKeyError
		assert.ErrorAs(t, err, &adminErr)
	})
}

// TestAuthenticationWithCookies tests that cookies are ignored for API key authentication
func TestAuthenticationWithCookies(t *testing.T) {
	authMethod := auth.NewAdminApiKeyAuthMethod()

	// Test with session cookies but no API key
	cookies := map[string]string{
		"terrareg_session": "valid-session-token",
		"csrf_token":       "csrf-token-123",
	}

	err := authMethod.Authenticate(context.Background(), make(map[string]string), cookies)
	assert.Error(t, err)
	assert.False(t, authMethod.IsAuthenticated())

	// Test with both cookies and valid API key (should succeed)
	headers := map[string]string{
		"Authorization": "Bearer admin-api-key-12345",
	}

	err = authMethod.Authenticate(context.Background(), headers, cookies)
	assert.NoError(t, err)
	assert.True(t, authMethod.IsAuthenticated())
}

// BenchmarkAuthentication tests the performance of authentication
func BenchmarkAuthentication(b *testing.B) {
	authMethod := auth.NewAdminApiKeyAuthMethod()
	headers := map[string]string{
		"Authorization": "Bearer admin-api-key-12345",
	}
	cookies := make(map[string]string)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Create new auth method for each iteration
		authMethod = auth.NewAdminApiKeyAuthMethod()
		_ = authMethod.Authenticate(ctx, headers, cookies)
	}
}

// BenchmarkCheckPermissions tests permission checking performance
func BenchmarkCheckPermissions(b *testing.B) {
	authMethod := auth.NewAdminApiKeyAuthMethod()
	headers := map[string]string{
		"Authorization": "Bearer admin-api-key-12345",
	}
	_ = authMethod.Authenticate(context.Background(), headers, make(map[string]string))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		authMethod.CanPublishModuleVersion("test-namespace")
		authMethod.CanUploadModuleVersion("test-namespace")
		authMethod.CheckNamespaceAccess("FULL_ACCESS", "test-namespace")
	}
}