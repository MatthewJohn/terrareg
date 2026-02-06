package auth

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestNewAdminApiKeyAuthContext tests the constructor
func TestNewAdminApiKeyAuthContext(t *testing.T) {
	ctx := context.Background()
	apiKey := "test-admin-key"

	authCtx := NewAdminApiKeyAuthContext(ctx, apiKey)

	assert.NotNil(t, authCtx)
	assert.Equal(t, AuthMethodAdminApiKey, authCtx.GetProviderType())
	assert.Equal(t, "admin-api-key", authCtx.GetUsername())
}

// TestAdminApiKeyAuthContext_IsAuthenticated tests authentication status
func TestAdminApiKeyAuthContext_IsAuthenticated(t *testing.T) {
	t.Run("authenticated with valid API key", func(t *testing.T) {
		authCtx := NewAdminApiKeyAuthContext(context.Background(), "valid-key")
		assert.True(t, authCtx.IsAuthenticated())
	})

	t.Run("not authenticated with empty API key", func(t *testing.T) {
		authCtx := NewAdminApiKeyAuthContext(context.Background(), "")
		assert.False(t, authCtx.IsAuthenticated())
	})
}

// TestAdminApiKeyAuthContext_IsAdmin tests admin status
func TestAdminApiKeyAuthContext_IsAdmin(t *testing.T) {
	authCtx := NewAdminApiKeyAuthContext(context.Background(), "any-key")

	assert.True(t, authCtx.IsAdmin(), "Admin API key should be admin")
	assert.True(t, authCtx.IsBuiltInAdmin(), "Admin API key should be built-in admin")
}

// TestAdminApiKeyAuthContext_RequiresCSRF tests CSRF requirement
func TestAdminApiKeyAuthContext_RequiresCSRF(t *testing.T) {
	authCtx := NewAdminApiKeyAuthContext(context.Background(), "any-key")
	assert.False(t, authCtx.RequiresCSRF(), "API key auth should not require CSRF")
}

// TestAdminApiKeyAuthContext_CheckAuthState tests auth state check
func TestAdminApiKeyAuthContext_CheckAuthState(t *testing.T) {
	authCtx := NewAdminApiKeyAuthContext(context.Background(), "any-key")
	assert.True(t, authCtx.CheckAuthState(), "API key auth should always pass auth state check")
}

// TestAdminApiKeyAuthContext_CanPublishModuleVersion tests publish permission
func TestAdminApiKeyAuthContext_CanPublishModuleVersion(t *testing.T) {
	authCtx := NewAdminApiKeyAuthContext(context.Background(), "any-key")

	// Admin API keys can publish to any module
	assert.True(t, authCtx.CanPublishModuleVersion("any-module"))
	assert.True(t, authCtx.CanPublishModuleVersion(""))
	assert.True(t, authCtx.CanPublishModuleVersion("specific-namespace/module"))
}

// TestAdminApiKeyAuthContext_CanUploadModuleVersion tests upload permission
func TestAdminApiKeyAuthContext_CanUploadModuleVersion(t *testing.T) {
	authCtx := NewAdminApiKeyAuthContext(context.Background(), "any-key")

	// Admin API keys can upload to any module
	assert.True(t, authCtx.CanUploadModuleVersion("any-module"))
	assert.True(t, authCtx.CanUploadModuleVersion(""))
	assert.True(t, authCtx.CanUploadModuleVersion("specific-namespace/module"))
}

// TestAdminApiKeyAuthContext_CheckNamespaceAccess tests namespace access
func TestAdminApiKeyAuthContext_CheckNamespaceAccess(t *testing.T) {
	authCtx := NewAdminApiKeyAuthContext(context.Background(), "any-key")

	// Admin API keys have access to all namespaces and all permission types
	permissions := []string{"READ", "MODIFY", "FULL", "UPLOAD", "PUBLISH", "DELETE", "CREATE_MODULE"}
	namespaces := []string{"any-namespace", "", "specific-ns", "ns-with/slash"}

	for _, permission := range permissions {
		t.Run("permission_"+permission, func(t *testing.T) {
			for _, namespace := range namespaces {
				assert.True(t, authCtx.CheckNamespaceAccess(permission, namespace),
					"Admin should have "+permission+" access to "+namespace)
			}
		})
	}
}

// TestAdminApiKeyAuthContext_GetAllNamespacePermissions tests getting all permissions
func TestAdminApiKeyAuthContext_GetAllNamespacePermissions(t *testing.T) {
	authCtx := NewAdminApiKeyAuthContext(context.Background(), "any-key")

	permissions := authCtx.GetAllNamespacePermissions()
	// Admin API keys return empty map (access is granted via CheckNamespaceAccess returning true)
	assert.Empty(t, permissions)
}

// TestAdminApiKeyAuthContext_GetUserGroupNames tests user group names
func TestAdminApiKeyAuthContext_GetUserGroupNames(t *testing.T) {
	authCtx := NewAdminApiKeyAuthContext(context.Background(), "any-key")

	groups := authCtx.GetUserGroupNames()
	assert.Empty(t, groups, "Admin API keys should not belong to any user groups")
}

// TestAdminApiKeyAuthContext_CanAccessReadAPI tests read API access
func TestAdminApiKeyAuthContext_CanAccessReadAPI(t *testing.T) {
	authCtx := NewAdminApiKeyAuthContext(context.Background(), "any-key")
	assert.True(t, authCtx.CanAccessReadAPI(), "Admin API keys can access read API")
}

// TestAdminApiKeyAuthContext_CanAccessTerraformAPI tests Terraform API access
func TestAdminApiKeyAuthContext_CanAccessTerraformAPI(t *testing.T) {
	authCtx := NewAdminApiKeyAuthContext(context.Background(), "any-key")
	assert.True(t, authCtx.CanAccessTerraformAPI(), "Admin API keys can access Terraform API")
}

// TestAdminApiKeyAuthContext_GetTerraformAuthToken tests Terraform token
func TestAdminApiKeyAuthContext_GetTerraformAuthToken(t *testing.T) {
	authCtx := NewAdminApiKeyAuthContext(context.Background(), "any-key")
	assert.Empty(t, authCtx.GetTerraformAuthToken(), "Admin API keys should not have Terraform token")
}

// TestAdminApiKeyAuthContext_GetProviderData tests provider data
func TestAdminApiKeyAuthContext_GetProviderData(t *testing.T) {
	apiKey := "test-admin-key"
	authCtx := NewAdminApiKeyAuthContext(context.Background(), apiKey)

	providerData := authCtx.GetProviderData()

	assert.NotNil(t, providerData)
	assert.Equal(t, string(AuthMethodAdminApiKey), providerData["auth_method"])
	assert.Equal(t, true, providerData["is_admin"])
}

// TestAdminApiKeyAuthContext_IsEnabled tests enabled status
func TestAdminApiKeyAuthContext_IsEnabled(t *testing.T) {
	authCtx := NewAdminApiKeyAuthContext(context.Background(), "any-key")
	assert.True(t, authCtx.IsEnabled(), "Admin API key context should always be enabled")
}

// TestAdminApiKeyAuthContext_GetUsername tests username retrieval
func TestAdminApiKeyAuthContext_GetUsername(t *testing.T) {
	authCtx := NewAdminApiKeyAuthContext(context.Background(), "any-key")
	assert.Equal(t, "admin-api-key", authCtx.GetUsername())
}

// TestAdminApiKeyAuthContext_GetProviderType tests provider type
func TestAdminApiKeyAuthContext_GetProviderType(t *testing.T) {
	authCtx := NewAdminApiKeyAuthContext(context.Background(), "any-key")
	assert.Equal(t, AuthMethodAdminApiKey, authCtx.GetProviderType())
}

// TestAdminApiKeyAuthContext_EdgeCases tests edge cases
func TestAdminApiKeyAuthContext_EdgeCases(t *testing.T) {
	t.Run("empty API key", func(t *testing.T) {
		authCtx := NewAdminApiKeyAuthContext(context.Background(), "")
		assert.False(t, authCtx.IsAuthenticated())
		// Even with empty key, admin context still grants access (validation happens at auth method level)
		assert.True(t, authCtx.IsAdmin())
		assert.True(t, authCtx.CheckNamespaceAccess("FULL", "any"))
	})

	t.Run("very long API key", func(t *testing.T) {
		longKey := string(make([]byte, 10000))
		authCtx := NewAdminApiKeyAuthContext(context.Background(), longKey)
		assert.True(t, authCtx.IsAuthenticated())
	})

	t.Run("API key with special characters", func(t *testing.T) {
		specialKey := "key-with-!@#$%^&*()_+-=[]{}|;':\",./<>?"
		authCtx := NewAdminApiKeyAuthContext(context.Background(), specialKey)
		assert.True(t, authCtx.IsAuthenticated())
	})

	t.Run("API key with unicode", func(t *testing.T) {
		unicodeKey := "key-with-中文-😀"
		authCtx := NewAdminApiKeyAuthContext(context.Background(), unicodeKey)
		assert.True(t, authCtx.IsAuthenticated())
	})
}

// TestAdminApiKeyAuthContext_FullAccess tests that admin has full access everywhere
func TestAdminApiKeyAuthContext_FullAccess(t *testing.T) {
	authCtx := NewAdminApiKeyAuthContext(context.Background(), "admin-key")

	// Admin should have full access to everything
	assert.True(t, authCtx.IsAdmin())
	assert.True(t, authCtx.IsBuiltInAdmin())
	assert.True(t, authCtx.CanAccessReadAPI())
	assert.True(t, authCtx.CanAccessTerraformAPI())
	assert.True(t, authCtx.CanPublishModuleVersion("any"))
	assert.True(t, authCtx.CanUploadModuleVersion("any"))
	assert.True(t, authCtx.CheckNamespaceAccess("ANY", "any-namespace"))
}

// TestAdminApiKeyAuthContext_AllPermissionTypes tests all permission types
func TestAdminApiKeyAuthContext_AllPermissionTypes(t *testing.T) {
	authCtx := NewAdminApiKeyAuthContext(context.Background(), "admin-key")

	permissionTypes := []string{
		"READ",
		"MODIFY",
		"FULL",
		"UPLOAD",
		"PUBLISH",
		"DELETE",
		"CREATE_MODULE",
		"DELETE_MODULE",
		"UPDATE_MODULE",
		"CREATE_MODULE_VERSION",
		"DELETE_MODULE_VERSION",
		"UPDATE_MODULE_VERSION",
		"CREATE_INTEGRATION",
		"DELETE_INTEGRATION",
	}

	for _, permType := range permissionTypes {
		t.Run("permission_"+permType, func(t *testing.T) {
			assert.True(t, authCtx.CheckNamespaceAccess(permType, "any-namespace"),
				"Admin should have "+permType+" permission")
		})
	}
}

// TestAdminApiKeyAuthContext_CompareWithPublishKey compares admin vs publish API keys
func TestAdminApiKeyAuthContext_CompareWithPublishKey(t *testing.T) {
	adminCtx := NewAdminApiKeyAuthContext(context.Background(), "admin-key")
	publishCtx := NewPublishApiKeyAuthContext(context.Background(), "publish-key")

	// Admin key can access read and terraform API, publish key cannot
	assert.True(t, adminCtx.CanAccessReadAPI())
	assert.False(t, publishCtx.CanAccessReadAPI())

	assert.True(t, adminCtx.CanAccessTerraformAPI())
	assert.False(t, publishCtx.CanAccessTerraformAPI())

	// Admin key has namespace access, publish key does not
	assert.True(t, adminCtx.CheckNamespaceAccess("READ", "any"))
	assert.False(t, publishCtx.CheckNamespaceAccess("READ", "any"))

	// Both can publish and upload
	assert.True(t, adminCtx.CanPublishModuleVersion("any"))
	assert.True(t, publishCtx.CanPublishModuleVersion("any"))

	assert.True(t, adminCtx.CanUploadModuleVersion("any"))
	assert.True(t, publishCtx.CanUploadModuleVersion("any"))

	// Admin is admin, publish key is not
	assert.True(t, adminCtx.IsAdmin())
	assert.False(t, publishCtx.IsAdmin())

	assert.True(t, adminCtx.IsBuiltInAdmin())
	assert.False(t, publishCtx.IsBuiltInAdmin())
}
