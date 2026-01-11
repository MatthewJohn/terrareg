package auth

import (
	"context"
	"testing"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/persistence/sqldb"
)

// TestNewGitHubAuthContext tests the constructor
func TestNewGitHubAuthContext(t *testing.T) {
	ctx := context.Background()
	providerSourceName := "test-github"
	username := "test-user"
	organizations := map[string]sqldb.NamespaceType{
		"test-user":  sqldb.NamespaceTypeGithubUser,
		"test-org-1": sqldb.NamespaceTypeGithubOrg,
	}

	authCtx := NewGitHubAuthContext(ctx, providerSourceName, username, organizations)

	if authCtx == nil {
		t.Fatal("NewGitHubAuthContext returned nil")
	}

	if authCtx.GetUsername() != username {
		t.Errorf("GetUsername() = %v, want %v", authCtx.GetUsername(), username)
	}

	if authCtx.IsAuthenticated() != true {
		t.Error("IsAuthenticated() = false, want true")
	}

	if authCtx.GetProviderType() != AuthMethodGitHub {
		t.Errorf("GetProviderType() = %v, want %v", authCtx.GetProviderType(), AuthMethodGitHub)
	}

	// Check that the provider source name is set correctly
	if authCtx.GetProviderSourceName() != providerSourceName {
		t.Errorf("GetProviderSourceName() = %v, want %v", authCtx.GetProviderSourceName(), providerSourceName)
	}
}

// TestGitHubAuthContext_CheckNamespaceAccess tests namespace access checking
func TestGitHubAuthContext_CheckNamespaceAccess(t *testing.T) {
	tests := []struct {
		name          string
		namespace     string
		accessType    string
		shouldHaveAccess bool
	}{
		{
			name:          "user has MODIFY access to own namespace",
			namespace:     "test-user",
			accessType:    "MODIFY",
			shouldHaveAccess: true,
		},
		{
			name:          "user has FULL access to own namespace",
			namespace:     "test-user",
			accessType:    "FULL",
			shouldHaveAccess: true,
		},
		{
			name:          "user has MODIFY access to organization namespace",
			namespace:     "test-org-1",
			accessType:    "MODIFY",
			shouldHaveAccess: true,
		},
		{
			name:          "user has FULL access to organization namespace",
			namespace:     "test-org-1",
			accessType:    "FULL",
			shouldHaveAccess: true,
		},
		{
			name:          "user does not have access to other namespace",
			namespace:     "other-org",
			accessType:    "MODIFY",
			shouldHaveAccess: false,
		},
		{
			name:          "user does not have FULL access to other namespace",
			namespace:     "other-org",
			accessType:    "FULL",
			shouldHaveAccess: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			organizations := map[string]sqldb.NamespaceType{
				"test-user":  sqldb.NamespaceTypeGithubUser,
				"test-org-1": sqldb.NamespaceTypeGithubOrg,
			}

			authCtx := NewGitHubAuthContext(
				context.Background(),
				"test-github",
				"test-user",
				organizations,
			)

			hasAccess := authCtx.CheckNamespaceAccess(tt.accessType, tt.namespace)
			if hasAccess != tt.shouldHaveAccess {
				t.Errorf("CheckNamespaceAccess(%s, %s) = %v, want %v", tt.accessType, tt.namespace, hasAccess, tt.shouldHaveAccess)
			}
		})
	}
}

// TestGitHubAuthContext_CheckNamespaceAccess_AllAccessLevels tests all access levels
func TestGitHubAuthContext_CheckNamespaceAccess_AllAccessLevels(t *testing.T) {
	accessLevels := []string{
		"READ",
		"MODIFY",
		"FULL",
		"DELETE",
		"UPDATE_MODULE",
		"CREATE_MODULE",
		"DELETE_MODULE",
		"UPDATE_MODULE_VERSION",
		"CREATE_MODULE_VERSION",
		"DELETE_MODULE_VERSION",
		"CREATE_INTEGRATION",
		"DELETE_INTEGRATION",
	}

	organizations := map[string]sqldb.NamespaceType{
		"test-user": sqldb.NamespaceTypeGithubUser,
	}

	authCtx := NewGitHubAuthContext(
		context.Background(),
		"test-github",
		"test-user",
		organizations,
	)

	for _, accessLevel := range accessLevels {
		t.Run(accessLevel, func(t *testing.T) {
			hasAccess := authCtx.CheckNamespaceAccess(accessLevel, "test-user")
			if !hasAccess {
				t.Errorf("CheckNamespaceAccess(%s, test-user) = false, want true", accessLevel)
			}
		})
	}
}

// TestGitHubAuthContext_CanPublishModuleVersion tests module version publish permissions
func TestGitHubAuthContext_CanPublishModuleVersion(t *testing.T) {
	tests := []struct {
		name          string
		namespace     string
		shouldHaveAccess bool
	}{
		{
			name:          "user can publish to own namespace",
			namespace:     "test-user",
			shouldHaveAccess: true,
		},
		{
			name:          "user can publish to organization namespace",
			namespace:     "test-org-1",
			shouldHaveAccess: true,
		},
		{
			name:          "user cannot publish to other namespace",
			namespace:     "other-org",
			shouldHaveAccess: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			organizations := map[string]sqldb.NamespaceType{
				"test-user":  sqldb.NamespaceTypeGithubUser,
				"test-org-1": sqldb.NamespaceTypeGithubOrg,
			}

			authCtx := NewGitHubAuthContext(
				context.Background(),
				"test-github",
				"test-user",
				organizations,
			)

			canPublish := authCtx.CanPublishModuleVersion(tt.namespace)
			if canPublish != tt.shouldHaveAccess {
				t.Errorf("CanPublishModuleVersion(%s) = %v, want %v", tt.namespace, canPublish, tt.shouldHaveAccess)
			}
		})
	}
}

// TestGitHubAuthContext_GetAllNamespacePermissions tests getting all namespace permissions
func TestGitHubAuthContext_GetAllNamespacePermissions(t *testing.T) {
	organizations := map[string]sqldb.NamespaceType{
		"test-user":        sqldb.NamespaceTypeGithubUser,
		"test-org-1":       sqldb.NamespaceTypeGithubOrg,
		"test-org-2":       sqldb.NamespaceTypeGithubOrg,
		"auto-generated-org": sqldb.NamespaceTypeGithubOrg,
	}

	authCtx := NewGitHubAuthContext(
		context.Background(),
		"test-github",
		"test-user",
		organizations,
	)

	permissions := authCtx.GetAllNamespacePermissions()

	// Verify all organizations are returned
	expectedCount := 4
	if len(permissions) != expectedCount {
		t.Errorf("GetAllNamespacePermissions() returned %d permissions, want %d", len(permissions), expectedCount)
	}

	// GetAllNamespacePermissions returns permission levels (FULL), not namespace types
	// The implementation returns "FULL" for all accessible namespaces
	if permissions["test-user"] != "FULL" {
		t.Errorf("GetAllNamespacePermissions()[test-user] = %v, want FULL", permissions["test-user"])
	}

	if permissions["test-org-1"] != "FULL" {
		t.Errorf("GetAllNamespacePermissions()[test-org-1] = %v, want FULL", permissions["test-org-1"])
	}

	if permissions["test-org-2"] != "FULL" {
		t.Errorf("GetAllNamespacePermissions()[test-org-2] = %v, want FULL", permissions["test-org-2"])
	}

	if permissions["auto-generated-org"] != "FULL" {
		t.Errorf("GetAllNamespacePermissions()[auto-generated-org] = %v, want FULL", permissions["auto-generated-org"])
	}
}

// TestGitHubAuthContext_GetAllNamespacePermissions_EmptyOrganizations tests with empty organizations
func TestGitHubAuthContext_GetAllNamespacePermissions_EmptyOrganizations(t *testing.T) {
	organizations := map[string]sqldb.NamespaceType{}

	authCtx := NewGitHubAuthContext(
		context.Background(),
		"test-github",
		"test-user",
		organizations,
	)

	permissions := authCtx.GetAllNamespacePermissions()

	// Even with empty organizations, the implementation adds the username
	// So we expect 1 permission: the username with FULL access
	expectedCount := 1
	if len(permissions) != expectedCount {
		t.Errorf("GetAllNamespacePermissions() returned %d permissions, want %d", len(permissions), expectedCount)
	}

	if permissions["test-user"] != "FULL" {
		t.Errorf("GetAllNamespacePermissions()[test-user] = %v, want FULL", permissions["test-user"])
	}
}

// TestGitHubAuthContext_GetProviderData tests provider data retrieval
func TestGitHubAuthContext_GetProviderData(t *testing.T) {
	providerSourceName := "test-github"
	username := "test-user"
	organizations := map[string]sqldb.NamespaceType{
		"test-user":  sqldb.NamespaceTypeGithubUser,
		"test-org-1": sqldb.NamespaceTypeGithubOrg,
	}

	authCtx := NewGitHubAuthContext(
		context.Background(),
		providerSourceName,
		username,
		organizations,
	)

	providerData := authCtx.GetProviderData()

	if providerData == nil {
		t.Fatal("GetProviderData() returned nil")
	}

	// Verify provider data contains expected fields
	if providerData["provider_source"] != providerSourceName {
		t.Errorf("GetProviderData()[provider_source] = %v, want %v", providerData["provider_source"], providerSourceName)
	}

	if providerData["github_username"] != username {
		t.Errorf("GetProviderData()[github_username] = %v, want %v", providerData["github_username"], username)
	}

	if providerData["auth_method"] != string(AuthMethodGitHub) {
		t.Errorf("GetProviderData()[auth_method] = %v, want %v", providerData["auth_method"], string(AuthMethodGitHub))
	}

	// Verify organizations are included
	orgs, ok := providerData["organisations"].(map[string]sqldb.NamespaceType)
	if !ok {
		t.Fatal("GetProviderData()[organisations] is not a map[string]NamespaceType")
	}

	if len(orgs) != 2 {
		t.Errorf("GetProviderData()[organisations] has %d entries, want 2", len(orgs))
	}

	if orgs["test-user"] != sqldb.NamespaceTypeGithubUser {
		t.Errorf("GetProviderData()[organisations][test-user] = %v, want %v", orgs["test-user"], sqldb.NamespaceTypeGithubUser)
	}

	if orgs["test-org-1"] != sqldb.NamespaceTypeGithubOrg {
		t.Errorf("GetProviderData()[organisations][test-org-1] = %v, want %v", orgs["test-org-1"], sqldb.NamespaceTypeGithubOrg)
	}
}

// TestGitHubAuthContext_IsAdmin tests admin status
func TestGitHubAuthContext_IsAdmin(t *testing.T) {
	organizations := map[string]sqldb.NamespaceType{
		"test-user": sqldb.NamespaceTypeGithubUser,
	}

	authCtx := NewGitHubAuthContext(
		context.Background(),
		"test-github",
		"test-user",
		organizations,
	)

	// GitHub auth is not admin by default
	if authCtx.IsAdmin() {
		t.Error("IsAdmin() = true, want false")
	}

	// GitHub auth is not built-in admin
	if authCtx.IsBuiltInAdmin() {
		t.Error("IsBuiltInAdmin() = true, want false")
	}
}

// TestGitHubAuthContext_GetTerraformAuthToken tests terraform token retrieval (should be empty for GitHub)
func TestGitHubAuthContext_GetTerraformAuthToken(t *testing.T) {
	organizations := map[string]sqldb.NamespaceType{
		"test-user": sqldb.NamespaceTypeGithubUser,
	}

	authCtx := NewGitHubAuthContext(
		context.Background(),
		"test-github",
		"test-user",
		organizations,
	)

	// GitHub auth doesn't use terraform tokens
	token := authCtx.GetTerraformAuthToken()
	if token != "" {
		t.Errorf("GetTerraformAuthToken() = %v, want empty string", token)
	}
}

// TestGitHubAuthContext_GetProviderSourceName tests provider source name retrieval
func TestGitHubAuthContext_GetProviderSourceName(t *testing.T) {
	ctx := context.Background()
	organizations := map[string]sqldb.NamespaceType{
		"test-user": sqldb.NamespaceTypeGithubUser,
	}

	authCtx := NewGitHubAuthContext(
		ctx,
		"test-github",
		"test-user",
		organizations,
	)

	if authCtx.GetProviderSourceName() != "test-github" {
		t.Errorf("GetProviderSourceName() = %v, want test-github", authCtx.GetProviderSourceName())
	}
}

// TestGitHubAuthContext_GetProviderType tests provider type retrieval
func TestGitHubAuthContext_GetProviderType(t *testing.T) {
	organizations := map[string]sqldb.NamespaceType{
		"test-user": sqldb.NamespaceTypeGithubUser,
	}

	authCtx := NewGitHubAuthContext(
		context.Background(),
		"test-github",
		"test-user",
		organizations,
	)

	if authCtx.GetProviderType() != AuthMethodGitHub {
		t.Errorf("GetProviderType() = %v, want %v", authCtx.GetProviderType(), AuthMethodGitHub)
	}
}

// TestGitHubAuthContext_WithMultipleOrganizations tests with multiple organizations
func TestGitHubAuthContext_WithMultipleOrganizations(t *testing.T) {
	username := "test-user"
	organizations := map[string]sqldb.NamespaceType{
		username:              sqldb.NamespaceTypeGithubUser,
		"my-org-1":            sqldb.NamespaceTypeGithubOrg,
		"my-org-2":            sqldb.NamespaceTypeGithubOrg,
		"my-company-frontend": sqldb.NamespaceTypeGithubOrg,
		"my-company-backend":  sqldb.NamespaceTypeGithubOrg,
	}

	authCtx := NewGitHubAuthContext(
		context.Background(),
		"test-github",
		username,
		organizations,
	)

	permissions := authCtx.GetAllNamespacePermissions()

	// Verify all organizations are accessible
	for org := range organizations {
		if !authCtx.CheckNamespaceAccess("MODIFY", org) {
			t.Errorf("CheckNamespaceAccess(MODIFY, %s) = false, want true", org)
		}
		if !authCtx.CheckNamespaceAccess("FULL", org) {
			t.Errorf("CheckNamespaceAccess(FULL, %s) = false, want true", org)
		}
	}

	// Verify permissions count
	if len(permissions) != len(organizations) {
		t.Errorf("GetAllNamespacePermissions() returned %d permissions, want %d", len(permissions), len(organizations))
	}
}

// TestGitHubAuthContext_GetUsername tests username retrieval
func TestGitHubAuthContext_GetUsername(t *testing.T) {
	tests := []struct {
		name     string
		username string
	}{
		{
			name:     "simple username",
			username: "testuser",
		},
		{
			name:     "username with numbers",
			username: "user123",
		},
		{
			name:     "username with hyphen",
			username: "test-user",
		},
		{
			name:     "empty username",
			username: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			organizations := map[string]sqldb.NamespaceType{
				tt.username: sqldb.NamespaceTypeGithubUser,
			}

			authCtx := NewGitHubAuthContext(
				context.Background(),
				"test-github",
				tt.username,
				organizations,
			)

			if authCtx.GetUsername() != tt.username {
				t.Errorf("GetUsername() = %v, want %v", authCtx.GetUsername(), tt.username)
			}
		})
	}
}
