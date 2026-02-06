package auth

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestNewPublishApiKeyAuthContext tests the constructor
func TestNewPublishApiKeyAuthContext(t *testing.T) {
	ctx := context.Background()
	apiKey := "test-publish-key"

	authCtx := NewPublishApiKeyAuthContext(ctx, apiKey)

	assert.NotNil(t, authCtx)
	assert.Equal(t, AuthMethodPublishApiKey, authCtx.GetProviderType())
	assert.Equal(t, "publish-api-key", authCtx.GetUsername())
}

// TestPublishApiKeyAuthContext_IsAuthenticated tests authentication status
func TestPublishApiKeyAuthContext_IsAuthenticated(t *testing.T) {
	t.Run("authenticated with valid API key", func(t *testing.T) {
		authCtx := NewPublishApiKeyAuthContext(context.Background(), "valid-key")
		assert.True(t, authCtx.IsAuthenticated())
	})

	t.Run("not authenticated with empty API key", func(t *testing.T) {
		authCtx := NewPublishApiKeyAuthContext(context.Background(), "")
		assert.False(t, authCtx.IsAuthenticated())
	})
}

// TestPublishApiKeyAuthContext_IsAdmin tests admin status
func TestPublishApiKeyAuthContext_IsAdmin(t *testing.T) {
	authCtx := NewPublishApiKeyAuthContext(context.Background(), "any-key")

	assert.False(t, authCtx.IsAdmin(), "Publish API key should not be admin")
	assert.False(t, authCtx.IsBuiltInAdmin(), "Publish API key should not be built-in admin")
}

// TestPublishApiKeyAuthContext_RequiresCSRF tests CSRF requirement
func TestPublishApiKeyAuthContext_RequiresCSRF(t *testing.T) {
	authCtx := NewPublishApiKeyAuthContext(context.Background(), "any-key")
	assert.False(t, authCtx.RequiresCSRF(), "API key auth should not require CSRF")
}

// TestPublishApiKeyAuthContext_CheckAuthState tests auth state check
func TestPublishApiKeyAuthContext_CheckAuthState(t *testing.T) {
	authCtx := NewPublishApiKeyAuthContext(context.Background(), "any-key")
	assert.True(t, authCtx.CheckAuthState(), "API key auth should always pass auth state check")
}

// TestPublishApiKeyAuthContext_CanPublishModuleVersion tests publish permission
func TestPublishApiKeyAuthContext_CanPublishModuleVersion(t *testing.T) {
	authCtx := NewPublishApiKeyAuthContext(context.Background(), "any-key")

	// Publish API keys can publish to any module
	assert.True(t, authCtx.CanPublishModuleVersion("any-module"))
	assert.True(t, authCtx.CanPublishModuleVersion(""))
	assert.True(t, authCtx.CanPublishModuleVersion("specific-namespace/module"))
}

// TestPublishApiKeyAuthContext_CanUploadModuleVersion tests upload permission
func TestPublishApiKeyAuthContext_CanUploadModuleVersion(t *testing.T) {
	authCtx := NewPublishApiKeyAuthContext(context.Background(), "any-key")

	// Publish API keys can also upload to any module
	assert.True(t, authCtx.CanUploadModuleVersion("any-module"))
	assert.True(t, authCtx.CanUploadModuleVersion(""))
	assert.True(t, authCtx.CanUploadModuleVersion("specific-namespace/module"))
}

// TestPublishApiKeyAuthContext_CheckNamespaceAccess tests namespace access
func TestPublishApiKeyAuthContext_CheckNamespaceAccess(t *testing.T) {
	authCtx := NewPublishApiKeyAuthContext(context.Background(), "any-key")

	// Publish API keys don't have namespace permissions
	assert.False(t, authCtx.CheckNamespaceAccess("READ", "any-namespace"))
	assert.False(t, authCtx.CheckNamespaceAccess("MODIFY", "any-namespace"))
	assert.False(t, authCtx.CheckNamespaceAccess("FULL", "any-namespace"))
	assert.False(t, authCtx.CheckNamespaceAccess("UPLOAD", "any-namespace"))
	assert.False(t, authCtx.CheckNamespaceAccess("PUBLISH", "any-namespace"))
}

// TestPublishApiKeyAuthContext_GetAllNamespacePermissions tests getting all permissions
func TestPublishApiKeyAuthContext_GetAllNamespacePermissions(t *testing.T) {
	authCtx := NewPublishApiKeyAuthContext(context.Background(), "any-key")

	permissions := authCtx.GetAllNamespacePermissions()
	assert.Empty(t, permissions, "Publish API keys should have no namespace permissions")
}

// TestPublishApiKeyAuthContext_GetUserGroupNames tests user group names
func TestPublishApiKeyAuthContext_GetUserGroupNames(t *testing.T) {
	authCtx := NewPublishApiKeyAuthContext(context.Background(), "any-key")

	groups := authCtx.GetUserGroupNames()
	assert.Empty(t, groups, "Publish API keys should not belong to any user groups")
}

// TestPublishApiKeyAuthContext_CanAccessReadAPI tests read API access
func TestPublishApiKeyAuthContext_CanAccessReadAPI(t *testing.T) {
	authCtx := NewPublishApiKeyAuthContext(context.Background(), "any-key")
	assert.False(t, authCtx.CanAccessReadAPI(), "Publish API keys cannot access read API")
}

// TestPublishApiKeyAuthContext_CanAccessTerraformAPI tests Terraform API access
func TestPublishApiKeyAuthContext_CanAccessTerraformAPI(t *testing.T) {
	authCtx := NewPublishApiKeyAuthContext(context.Background(), "any-key")
	assert.False(t, authCtx.CanAccessTerraformAPI(), "Publish API keys cannot access Terraform API")
}

// TestPublishApiKeyAuthContext_GetTerraformAuthToken tests Terraform token
func TestPublishApiKeyAuthContext_GetTerraformAuthToken(t *testing.T) {
	authCtx := NewPublishApiKeyAuthContext(context.Background(), "any-key")
	assert.Empty(t, authCtx.GetTerraformAuthToken(), "Publish API keys should not have Terraform token")
}

// TestPublishApiKeyAuthContext_GetProviderData tests provider data
func TestPublishApiKeyAuthContext_GetProviderData(t *testing.T) {
	apiKey := "test-publish-key"
	authCtx := NewPublishApiKeyAuthContext(context.Background(), apiKey)

	providerData := authCtx.GetProviderData()

	assert.NotNil(t, providerData)
	assert.Equal(t, string(AuthMethodPublishApiKey), providerData["auth_method"])
	assert.Equal(t, false, providerData["is_admin"])
	assert.Equal(t, true, providerData["can_publish"])
	assert.Equal(t, true, providerData["can_upload"])
}

// TestPublishApiKeyAuthContext_IsEnabled tests enabled status
func TestPublishApiKeyAuthContext_IsEnabled(t *testing.T) {
	authCtx := NewPublishApiKeyAuthContext(context.Background(), "any-key")
	assert.True(t, authCtx.IsEnabled(), "Publish API key context should always be enabled")
}

// TestPublishApiKeyAuthContext_GetUsername tests username retrieval
func TestPublishApiKeyAuthContext_GetUsername(t *testing.T) {
	authCtx := NewPublishApiKeyAuthContext(context.Background(), "any-key")
	assert.Equal(t, "publish-api-key", authCtx.GetUsername())
}

// TestPublishApiKeyAuthContext_GetProviderType tests provider type
func TestPublishApiKeyAuthContext_GetProviderType(t *testing.T) {
	authCtx := NewPublishApiKeyAuthContext(context.Background(), "any-key")
	assert.Equal(t, AuthMethodPublishApiKey, authCtx.GetProviderType())
}

// TestPublishApiKeyAuthContext_EdgeCases tests edge cases
func TestPublishApiKeyAuthContext_EdgeCases(t *testing.T) {
	t.Run("empty API key", func(t *testing.T) {
		authCtx := NewPublishApiKeyAuthContext(context.Background(), "")
		assert.False(t, authCtx.IsAuthenticated())
		// Should still have can publish/upload (key validation happens at auth method level)
		assert.True(t, authCtx.CanPublishModuleVersion("any"))
		assert.True(t, authCtx.CanUploadModuleVersion("any"))
	})

	t.Run("very long API key", func(t *testing.T) {
		longKey := string(make([]byte, 10000))
		authCtx := NewPublishApiKeyAuthContext(context.Background(), longKey)
		assert.True(t, authCtx.IsAuthenticated())
	})

	t.Run("API key with special characters", func(t *testing.T) {
		specialKey := "key-with-!@#$%^&*()_+-=[]{}|;':\",./<>?"
		authCtx := NewPublishApiKeyAuthContext(context.Background(), specialKey)
		assert.True(t, authCtx.IsAuthenticated())
	})

	t.Run("API key with unicode", func(t *testing.T) {
		unicodeKey := "key-with-中文-😀"
		authCtx := NewPublishApiKeyAuthContext(context.Background(), unicodeKey)
		assert.True(t, authCtx.IsAuthenticated())
	})
}

// TestPublishApiKeyAuthContext_AllModuleNames tests with various module names
func TestPublishApiKeyAuthContext_AllModuleNames(t *testing.T) {
	authCtx := NewPublishApiKeyAuthContext(context.Background(), "any-key")

	moduleNames := []string{
		"simple",
		"with-hyphen",
		"with_underscore",
		"with.dot",
		"with/slash",
		"namespace/module/provider",
		"",
		"very-long-module-name-" + string(make([]byte, 100)),
	}

	for _, moduleName := range moduleNames {
		t.Run("module_"+moduleName, func(t *testing.T) {
			assert.True(t, authCtx.CanPublishModuleVersion(moduleName))
			assert.True(t, authCtx.CanUploadModuleVersion(moduleName))
		})
	}
}
