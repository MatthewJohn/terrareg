package auth

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestNewSessionAuthContext tests the constructor
func TestNewSessionAuthContext(t *testing.T) {
	ctx := context.Background()
	userID := 123
	username := "testuser"
	email := "test@example.com"
	sessionID := "session-123"

	authCtx := NewSessionAuthContext(ctx, userID, username, email, sessionID)

	assert.NotNil(t, authCtx)
	assert.Equal(t, userID, authCtx.userID)
	assert.Equal(t, username, authCtx.username)
	assert.Equal(t, email, authCtx.email)
	assert.Equal(t, sessionID, authCtx.sessionID)
	assert.Equal(t, AuthMethodAdminSession, authCtx.GetProviderType())
}

// TestSessionAuthContext_IsAuthenticated tests authentication status
func TestSessionAuthContext_IsAuthenticated(t *testing.T) {
	t.Run("authenticated with valid session", func(t *testing.T) {
		authCtx := NewSessionAuthContext(context.Background(), 1, "user", "email@test.com", "session-id")
		assert.True(t, authCtx.IsAuthenticated())
		assert.True(t, authCtx.IsEnabled())
		assert.True(t, authCtx.CheckAuthState())
	})

	t.Run("not authenticated with empty session ID", func(t *testing.T) {
		authCtx := NewSessionAuthContext(context.Background(), 1, "user", "email@test.com", "")
		assert.False(t, authCtx.IsAuthenticated())
	})

	t.Run("not authenticated with empty username", func(t *testing.T) {
		authCtx := NewSessionAuthContext(context.Background(), 1, "", "email@test.com", "session-id")
		assert.False(t, authCtx.IsAuthenticated())
	})

	t.Run("not authenticated with both empty", func(t *testing.T) {
		authCtx := NewSessionAuthContext(context.Background(), 1, "", "email@test.com", "")
		assert.False(t, authCtx.IsAuthenticated())
	})
}

// TestSessionAuthContext_IsAdmin tests admin status
func TestSessionAuthContext_IsAdmin(t *testing.T) {
	t.Run("not admin by default", func(t *testing.T) {
		authCtx := NewSessionAuthContext(context.Background(), 1, "user", "email@test.com", "session-id")
		assert.False(t, authCtx.IsAdmin())
		assert.False(t, authCtx.IsBuiltInAdmin())
	})

	t.Run("becomes admin when site admin group added", func(t *testing.T) {
		authCtx := NewSessionAuthContext(context.Background(), 1, "user", "email@test.com", "session-id")
		adminGroup := &UserGroup{ID: 1, Name: "admins", SiteAdmin: true}
		authCtx.AddUserGroup(adminGroup)

		assert.True(t, authCtx.IsAdmin())
		assert.False(t, authCtx.IsBuiltInAdmin()) // Still not built-in admin
	})

	t.Run("remains not admin with non-site admin group", func(t *testing.T) {
		authCtx := NewSessionAuthContext(context.Background(), 1, "user", "email@test.com", "session-id")
		regularGroup := &UserGroup{ID: 1, Name: "users", SiteAdmin: false}
		authCtx.AddUserGroup(regularGroup)

		assert.False(t, authCtx.IsAdmin())
		assert.False(t, authCtx.IsBuiltInAdmin())
	})
}

// TestSessionAuthContext_RequiresCSRF tests CSRF requirement
func TestSessionAuthContext_RequiresCSRF(t *testing.T) {
	authCtx := NewSessionAuthContext(context.Background(), 1, "user", "email@test.com", "session-id")
	assert.True(t, authCtx.RequiresCSRF(), "Session auth should require CSRF")
}

// TestSessionAuthContext_AddUserGroup tests adding user groups
func TestSessionAuthContext_AddUserGroup(t *testing.T) {
	authCtx := NewSessionAuthContext(context.Background(), 1, "user", "email@test.com", "session-id")

	group1 := &UserGroup{ID: 1, Name: "group1", SiteAdmin: false}
	group2 := &UserGroup{ID: 2, Name: "group2", SiteAdmin: false}
	adminGroup := &UserGroup{ID: 3, Name: "admins", SiteAdmin: true}

	authCtx.AddUserGroup(group1)
	authCtx.AddUserGroup(group2)
	authCtx.AddUserGroup(adminGroup)

	groupNames := authCtx.GetUserGroupNames()
	assert.Equal(t, 3, len(groupNames))
	assert.Contains(t, groupNames, "group1")
	assert.Contains(t, groupNames, "group2")
	assert.Contains(t, groupNames, "admins")
	assert.True(t, authCtx.IsAdmin(), "Should be admin after adding site admin group")
}

// TestSessionAuthContext_SetPermission tests setting namespace permissions
func TestSessionAuthContext_SetPermission(t *testing.T) {
	authCtx := NewSessionAuthContext(context.Background(), 1, "user", "email@test.com", "session-id")

	authCtx.SetPermission("namespace1", "FULL")
	authCtx.SetPermission("namespace2", "MODIFY")
	authCtx.SetPermission("namespace3", "READ")

	permissions := authCtx.GetAllNamespacePermissions()
	assert.Equal(t, 3, len(permissions))
	assert.Equal(t, "FULL", permissions["namespace1"])
	assert.Equal(t, "MODIFY", permissions["namespace2"])
	assert.Equal(t, "READ", permissions["namespace3"])
}

// TestSessionAuthContext_CanPublishModuleVersion tests publish permission
func TestSessionAuthContext_CanPublishModuleVersion(t *testing.T) {
	t.Run("admin can publish", func(t *testing.T) {
		authCtx := NewSessionAuthContext(context.Background(), 1, "user", "email@test.com", "session-id")
		adminGroup := &UserGroup{ID: 1, Name: "admins", SiteAdmin: true}
		authCtx.AddUserGroup(adminGroup)

		assert.True(t, authCtx.CanPublishModuleVersion("any-namespace"))
	})

	t.Run("user with FULL permission can publish", func(t *testing.T) {
		authCtx := NewSessionAuthContext(context.Background(), 1, "user", "email@test.com", "session-id")
		authCtx.SetPermission("ns1", "FULL")

		assert.True(t, authCtx.CanPublishModuleVersion("ns1"))
	})

	t.Run("user with MODIFY permission can publish", func(t *testing.T) {
		authCtx := NewSessionAuthContext(context.Background(), 1, "user", "email@test.com", "session-id")
		authCtx.SetPermission("ns1", "MODIFY")

		assert.True(t, authCtx.CanPublishModuleVersion("ns1"))
	})

	t.Run("user with PUBLISH permission can publish", func(t *testing.T) {
		authCtx := NewSessionAuthContext(context.Background(), 1, "user", "email@test.com", "session-id")
		authCtx.SetPermission("ns1", "PUBLISH")

		assert.True(t, authCtx.CanPublishModuleVersion("ns1"))
	})

	t.Run("user with READ permission cannot publish", func(t *testing.T) {
		authCtx := NewSessionAuthContext(context.Background(), 1, "user", "email@test.com", "session-id")
		authCtx.SetPermission("ns1", "READ")

		assert.False(t, authCtx.CanPublishModuleVersion("ns1"))
	})

	t.Run("user with no permission cannot publish", func(t *testing.T) {
		authCtx := NewSessionAuthContext(context.Background(), 1, "user", "email@test.com", "session-id")

		assert.False(t, authCtx.CanPublishModuleVersion("ns1"))
	})
}

// TestSessionAuthContext_CanUploadModuleVersion tests upload permission
func TestSessionAuthContext_CanUploadModuleVersion(t *testing.T) {
	t.Run("admin can upload", func(t *testing.T) {
		authCtx := NewSessionAuthContext(context.Background(), 1, "user", "email@test.com", "session-id")
		adminGroup := &UserGroup{ID: 1, Name: "admins", SiteAdmin: true}
		authCtx.AddUserGroup(adminGroup)

		assert.True(t, authCtx.CanUploadModuleVersion("any-namespace"))
	})

	t.Run("user with FULL permission can upload", func(t *testing.T) {
		authCtx := NewSessionAuthContext(context.Background(), 1, "user", "email@test.com", "session-id")
		authCtx.SetPermission("ns1", "FULL")

		assert.True(t, authCtx.CanUploadModuleVersion("ns1"))
	})

	t.Run("user with MODIFY permission can upload", func(t *testing.T) {
		authCtx := NewSessionAuthContext(context.Background(), 1, "user", "email@test.com", "session-id")
		authCtx.SetPermission("ns1", "MODIFY")

		assert.True(t, authCtx.CanUploadModuleVersion("ns1"))
	})

	t.Run("user with UPLOAD permission can upload", func(t *testing.T) {
		authCtx := NewSessionAuthContext(context.Background(), 1, "user", "email@test.com", "session-id")
		authCtx.SetPermission("ns1", "UPLOAD")

		assert.True(t, authCtx.CanUploadModuleVersion("ns1"))
	})

	t.Run("user with PUBLISH permission cannot upload", func(t *testing.T) {
		authCtx := NewSessionAuthContext(context.Background(), 1, "user", "email@test.com", "session-id")
		authCtx.SetPermission("ns1", "PUBLISH")

		assert.False(t, authCtx.CanUploadModuleVersion("ns1"))
	})

	t.Run("user with READ permission cannot upload", func(t *testing.T) {
		authCtx := NewSessionAuthContext(context.Background(), 1, "user", "email@test.com", "session-id")
		authCtx.SetPermission("ns1", "READ")

		assert.False(t, authCtx.CanUploadModuleVersion("ns1"))
	})

	t.Run("user with no permission cannot upload", func(t *testing.T) {
		authCtx := NewSessionAuthContext(context.Background(), 1, "user", "email@test.com", "session-id")

		assert.False(t, authCtx.CanUploadModuleVersion("ns1"))
	})
}

// TestSessionAuthContext_CheckNamespaceAccess tests namespace access checking
func TestSessionAuthContext_CheckNamespaceAccess(t *testing.T) {
	t.Run("admin has all access", func(t *testing.T) {
		authCtx := NewSessionAuthContext(context.Background(), 1, "user", "email@test.com", "session-id")
		adminGroup := &UserGroup{ID: 1, Name: "admins", SiteAdmin: true}
		authCtx.AddUserGroup(adminGroup)

		assert.True(t, authCtx.CheckNamespaceAccess("READ", "ns1"))
		assert.True(t, authCtx.CheckNamespaceAccess("MODIFY", "ns1"))
		assert.True(t, authCtx.CheckNamespaceAccess("FULL", "ns1"))
		assert.True(t, authCtx.CheckNamespaceAccess("UPLOAD", "ns1"))
		assert.True(t, authCtx.CheckNamespaceAccess("PUBLISH", "ns1"))
	})

	t.Run("permission hierarchy - FULL grants all", func(t *testing.T) {
		authCtx := NewSessionAuthContext(context.Background(), 1, "user", "email@test.com", "session-id")
		authCtx.SetPermission("ns1", "FULL")

		assert.True(t, authCtx.CheckNamespaceAccess("READ", "ns1"))
		assert.True(t, authCtx.CheckNamespaceAccess("MODIFY", "ns1"))
		assert.True(t, authCtx.CheckNamespaceAccess("UPLOAD", "ns1"))
		assert.True(t, authCtx.CheckNamespaceAccess("PUBLISH", "ns1"))
		assert.True(t, authCtx.CheckNamespaceAccess("FULL", "ns1"))
	})

	t.Run("permission hierarchy - MODIFY grants READ and MODIFY", func(t *testing.T) {
		authCtx := NewSessionAuthContext(context.Background(), 1, "user", "email@test.com", "session-id")
		authCtx.SetPermission("ns1", "MODIFY")

		assert.True(t, authCtx.CheckNamespaceAccess("READ", "ns1"))
		assert.True(t, authCtx.CheckNamespaceAccess("MODIFY", "ns1"))
		assert.False(t, authCtx.CheckNamespaceAccess("FULL", "ns1"))
	})

	t.Run("permission hierarchy - READ grants only READ", func(t *testing.T) {
		authCtx := NewSessionAuthContext(context.Background(), 1, "user", "email@test.com", "session-id")
		authCtx.SetPermission("ns1", "READ")

		assert.True(t, authCtx.CheckNamespaceAccess("READ", "ns1"))
		assert.False(t, authCtx.CheckNamespaceAccess("MODIFY", "ns1"))
		assert.False(t, authCtx.CheckNamespaceAccess("FULL", "ns1"))
	})

	t.Run("no permission returns false", func(t *testing.T) {
		authCtx := NewSessionAuthContext(context.Background(), 1, "user", "email@test.com", "session-id")

		assert.False(t, authCtx.CheckNamespaceAccess("READ", "ns1"))
		assert.False(t, authCtx.CheckNamespaceAccess("MODIFY", "ns1"))
	})
}

// TestSessionAuthContext_GetAllNamespacePermissions tests getting all permissions
func TestSessionAuthContext_GetAllNamespacePermissions(t *testing.T) {
	authCtx := NewSessionAuthContext(context.Background(), 1, "user", "email@test.com", "session-id")

	authCtx.SetPermission("ns1", "FULL")
	authCtx.SetPermission("ns2", "MODIFY")
	authCtx.SetPermission("ns3", "READ")

	permissions := authCtx.GetAllNamespacePermissions()
	assert.Equal(t, 3, len(permissions))
	assert.Equal(t, "FULL", permissions["ns1"])
	assert.Equal(t, "MODIFY", permissions["ns2"])
	assert.Equal(t, "READ", permissions["ns3"])
}

// TestSessionAuthContext_GetAllNamespacePermissions_WithSiteAdmin tests permissions with site admin
func TestSessionAuthContext_GetAllNamespacePermissions_WithSiteAdmin(t *testing.T) {
	authCtx := NewSessionAuthContext(context.Background(), 1, "user", "email@test.com", "session-id")
	adminGroup := &UserGroup{ID: 1, Name: "admins", SiteAdmin: true}
	authCtx.AddUserGroup(adminGroup)

	permissions := authCtx.GetAllNamespacePermissions()
	assert.Contains(t, permissions, "*")
	assert.Equal(t, "FULL", permissions["*"])
}

// TestSessionAuthContext_GetUserGroupNames tests getting user group names
func TestSessionAuthContext_GetUserGroupNames(t *testing.T) {
	authCtx := NewSessionAuthContext(context.Background(), 1, "user", "email@test.com", "session-id")

	group1 := &UserGroup{ID: 1, Name: "group1", SiteAdmin: false}
	group2 := &UserGroup{ID: 2, Name: "group2", SiteAdmin: false}
	group3 := &UserGroup{ID: 3, Name: "group3", SiteAdmin: false}

	authCtx.AddUserGroup(group1)
	authCtx.AddUserGroup(group2)
	authCtx.AddUserGroup(group3)

	names := authCtx.GetUserGroupNames()
	assert.Equal(t, 3, len(names))
	assert.Contains(t, names, "group1")
	assert.Contains(t, names, "group2")
	assert.Contains(t, names, "group3")
}

// TestSessionAuthContext_CanAccessReadAPI tests read API access
func TestSessionAuthContext_CanAccessReadAPI(t *testing.T) {
	t.Run("authenticated user can access read API", func(t *testing.T) {
		authCtx := NewSessionAuthContext(context.Background(), 1, "user", "email@test.com", "session-id")
		assert.True(t, authCtx.CanAccessReadAPI())
	})

	t.Run("unauthenticated user cannot access read API", func(t *testing.T) {
		authCtx := NewSessionAuthContext(context.Background(), 1, "user", "email@test.com", "")
		assert.False(t, authCtx.CanAccessReadAPI())
	})
}

// TestSessionAuthContext_CanAccessTerraformAPI tests Terraform API access
func TestSessionAuthContext_CanAccessTerraformAPI(t *testing.T) {
	t.Run("authenticated user can access Terraform API", func(t *testing.T) {
		authCtx := NewSessionAuthContext(context.Background(), 1, "user", "email@test.com", "session-id")
		assert.True(t, authCtx.CanAccessTerraformAPI())
	})

	t.Run("unauthenticated user cannot access Terraform API", func(t *testing.T) {
		authCtx := NewSessionAuthContext(context.Background(), 1, "user", "email@test.com", "")
		assert.False(t, authCtx.CanAccessTerraformAPI())
	})
}

// TestSessionAuthContext_GetTerraformAuthToken tests Terraform token retrieval
func TestSessionAuthContext_GetTerraformAuthToken(t *testing.T) {
	authCtx := NewSessionAuthContext(context.Background(), 1, "user", "email@test.com", "session-id")
	assert.Empty(t, authCtx.GetTerraformAuthToken())
}

// TestSessionAuthContext_GetProviderData tests provider data retrieval
func TestSessionAuthContext_GetProviderData(t *testing.T) {
	userID := 123
	username := "testuser"
	email := "test@example.com"
	sessionID := "session-abc"

	authCtx := NewSessionAuthContext(context.Background(), userID, username, email, sessionID)

	providerData := authCtx.GetProviderData()

	assert.NotNil(t, providerData)
	assert.Equal(t, userID, providerData["user_id"])
	assert.Equal(t, username, providerData["username"])
	assert.Equal(t, email, providerData["email"])
	assert.Equal(t, sessionID, providerData["session_id"])
	assert.Equal(t, false, providerData["is_admin"])
	assert.Equal(t, string(AuthMethodAdminSession), providerData["auth_method"])
}

// TestSessionAuthContext_GetProviderData_WithAdmin tests provider data with admin user
func TestSessionAuthContext_GetProviderData_WithAdmin(t *testing.T) {
	authCtx := NewSessionAuthContext(context.Background(), 1, "admin", "admin@test.com", "session-id")
	adminGroup := &UserGroup{ID: 1, Name: "admins", SiteAdmin: true}
	authCtx.AddUserGroup(adminGroup)

	providerData := authCtx.GetProviderData()
	assert.Equal(t, true, providerData["is_admin"])
}

// TestSessionAuthContext_GetUsername tests username retrieval
func TestSessionAuthContext_GetUsername(t *testing.T) {
	username := "testuser"
	authCtx := NewSessionAuthContext(context.Background(), 1, username, "email@test.com", "session-id")
	assert.Equal(t, username, authCtx.GetUsername())
}

// TestSessionAuthContext_GetProviderType tests provider type retrieval
func TestSessionAuthContext_GetProviderType(t *testing.T) {
	authCtx := NewSessionAuthContext(context.Background(), 1, "user", "email@test.com", "session-id")
	assert.Equal(t, AuthMethodAdminSession, authCtx.GetProviderType())
}

// TestSessionAuthContext_PermissionHierarchy tests the permission hierarchy
func TestSessionAuthContext_PermissionHierarchy(t *testing.T) {
	tests := []struct {
		stored     string
		required   string
		shouldPass bool
	}{
		// FULL grants everything
		{"FULL", "READ", true},
		{"FULL", "MODIFY", true},
		{"FULL", "UPLOAD", true},
		{"FULL", "PUBLISH", true},
		{"FULL", "FULL", true},
		// PUBLISH grants READ, MODIFY, UPLOAD, PUBLISH
		{"PUBLISH", "READ", true},
		{"PUBLISH", "MODIFY", true},
		{"PUBLISH", "UPLOAD", true},
		{"PUBLISH", "PUBLISH", true},
		{"PUBLISH", "FULL", false},
		// UPLOAD grants READ, MODIFY, UPLOAD
		{"UPLOAD", "READ", true},
		{"UPLOAD", "MODIFY", true},
		{"UPLOAD", "UPLOAD", true},
		{"UPLOAD", "PUBLISH", false},
		{"UPLOAD", "FULL", false},
		// MODIFY grants READ, MODIFY
		{"MODIFY", "READ", true},
		{"MODIFY", "MODIFY", true},
		{"MODIFY", "UPLOAD", false},
		{"MODIFY", "PUBLISH", false},
		{"MODIFY", "FULL", false},
		// READ grants only READ
		{"READ", "READ", true},
		{"READ", "MODIFY", false},
		{"READ", "UPLOAD", false},
		{"READ", "PUBLISH", false},
		{"READ", "FULL", false},
	}

	for _, tt := range tests {
		t.Run(tt.stored+"_vs_"+tt.required, func(t *testing.T) {
			authCtx := NewSessionAuthContext(context.Background(), 1, "user", "email@test.com", "session-id")
			authCtx.SetPermission("test-ns", tt.stored)

			result := authCtx.CheckNamespaceAccess(tt.required, "test-ns")
			assert.Equal(t, tt.shouldPass, result, "Permission %v should grant access to %v: %v", tt.stored, tt.required, tt.shouldPass)
		})
	}
}

// TestSessionAuthContext_MultipleNamespaces tests multiple namespaces with different permissions
func TestSessionAuthContext_MultipleNamespaces(t *testing.T) {
	authCtx := NewSessionAuthContext(context.Background(), 1, "user", "email@test.com", "session-id")

	authCtx.SetPermission("ns-full", "FULL")
	authCtx.SetPermission("ns-modify", "MODIFY")
	authCtx.SetPermission("ns-read", "READ")

	// Can read from all
	assert.True(t, authCtx.CheckNamespaceAccess("READ", "ns-full"))
	assert.True(t, authCtx.CheckNamespaceAccess("READ", "ns-modify"))
	assert.True(t, authCtx.CheckNamespaceAccess("READ", "ns-read"))

	// Can modify only full and modify
	assert.True(t, authCtx.CheckNamespaceAccess("MODIFY", "ns-full"))
	assert.True(t, authCtx.CheckNamespaceAccess("MODIFY", "ns-modify"))
	assert.False(t, authCtx.CheckNamespaceAccess("MODIFY", "ns-read"))

	// Can publish only full
	assert.True(t, authCtx.CheckNamespaceAccess("PUBLISH", "ns-full"))
	assert.False(t, authCtx.CheckNamespaceAccess("PUBLISH", "ns-modify"))
	assert.False(t, authCtx.CheckNamespaceAccess("PUBLISH", "ns-read"))
}

// TestSessionAuthContext_EdgeCases tests edge cases
func TestSessionAuthContext_EdgeCases(t *testing.T) {
	t.Run("empty namespace with permission", func(t *testing.T) {
		authCtx := NewSessionAuthContext(context.Background(), 1, "user", "email@test.com", "session-id")
		authCtx.SetPermission("", "FULL")

		assert.True(t, authCtx.CheckNamespaceAccess("FULL", ""))
	})

	t.Run("zero user ID", func(t *testing.T) {
		authCtx := NewSessionAuthContext(context.Background(), 0, "user", "email@test.com", "session-id")
		assert.True(t, authCtx.IsAuthenticated())
	})

	t.Run("negative user ID", func(t *testing.T) {
		authCtx := NewSessionAuthContext(context.Background(), -1, "user", "email@test.com", "session-id")
		assert.True(t, authCtx.IsAuthenticated())
	})

	t.Run("empty email", func(t *testing.T) {
		authCtx := NewSessionAuthContext(context.Background(), 1, "user", "", "session-id")
		assert.True(t, authCtx.IsAuthenticated())
	})

	t.Run("multiple groups with same name", func(t *testing.T) {
		authCtx := NewSessionAuthContext(context.Background(), 1, "user", "email@test.com", "session-id")
		group1 := &UserGroup{ID: 1, Name: "group1", SiteAdmin: false}
		group2 := &UserGroup{ID: 2, Name: "group1", SiteAdmin: false}

		authCtx.AddUserGroup(group1)
		authCtx.AddUserGroup(group2)

		names := authCtx.GetUserGroupNames()
		assert.Equal(t, 2, len(names))
	})

	t.Run("setting same permission multiple times", func(t *testing.T) {
		authCtx := NewSessionAuthContext(context.Background(), 1, "user", "email@test.com", "session-id")
		authCtx.SetPermission("ns1", "READ")
		authCtx.SetPermission("ns1", "MODIFY")
		authCtx.SetPermission("ns1", "FULL")

		permissions := authCtx.GetAllNamespacePermissions()
		assert.Equal(t, "FULL", permissions["ns1"])
		assert.Equal(t, 1, len(permissions))
	})
}
