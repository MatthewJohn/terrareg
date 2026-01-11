package auth

import (
	"context"
	"testing"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/auth"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/persistence/sqldb"
	provider_source_service "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/provider_source/service"
	provider_source_model "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/provider_source/model"
)

// MockProviderSourceRepository is a mock implementation of ProviderSourceRepository
type MockProviderSourceRepository struct{}

func (m *MockProviderSourceRepository) FindByName(ctx context.Context, name string) (*provider_source_model.ProviderSource, error) {
	// Return a valid GitHub provider source for testing
	config := &provider_source_model.ProviderSourceConfig{
		BaseURL:      "https://github.com",
		ClientID:     "test-client-id",
		ClientSecret: "test-client-secret",
	}
	ps := provider_source_model.NewProviderSource("Test GitHub", "test-github", provider_source_model.ProviderSourceTypeGithub, config)
	return ps, nil
}

func (m *MockProviderSourceRepository) FindByApiName(ctx context.Context, apiName string) (*provider_source_model.ProviderSource, error) {
	// Return a valid GitHub provider source for testing
	config := &provider_source_model.ProviderSourceConfig{
		BaseURL:      "https://github.com",
		ClientID:     "test-client-id",
		ClientSecret: "test-client-secret",
	}
	ps := provider_source_model.NewProviderSource("Test GitHub", apiName, provider_source_model.ProviderSourceTypeGithub, config)
	return ps, nil
}

func (m *MockProviderSourceRepository) FindAll(ctx context.Context) ([]*provider_source_model.ProviderSource, error) {
	return nil, nil
}

func (m *MockProviderSourceRepository) Upsert(ctx context.Context, source *provider_source_model.ProviderSource) error {
	return nil
}

func (m *MockProviderSourceRepository) Delete(ctx context.Context, name string) error {
	return nil
}

func (m *MockProviderSourceRepository) Exists(ctx context.Context, name string) (bool, error) {
	return false, nil
}

func (m *MockProviderSourceRepository) ExistsByApiName(ctx context.Context, apiName string) (bool, error) {
	return false, nil
}

// TestNewGitHubAuthMethod tests the constructor
func TestNewGitHubAuthMethod(t *testing.T) {
	repo := &MockProviderSourceRepository{}
	factory := provider_source_service.NewProviderSourceFactory(repo)
	method := NewGitHubAuthMethod(factory)

	if method == nil {
		t.Fatal("NewGitHubAuthMethod returned nil")
	}

	if method.providerSourceFactory != factory {
		t.Error("providerSourceFactory not set correctly")
	}
}

// TestGitHubAuthMethod_GetProviderType tests GetProviderType
func TestGitHubAuthMethod_GetProviderType(t *testing.T) {
	repo := &MockProviderSourceRepository{}
	factory := provider_source_service.NewProviderSourceFactory(repo)
	method := NewGitHubAuthMethod(factory)
	expected := auth.AuthMethodGitHub
	actual := method.GetProviderType()

	if actual != expected {
		t.Errorf("GetProviderType() = %v, want %v", actual, expected)
	}
}

// TestGitHubAuthMethod_IsEnabled tests IsEnabled
func TestGitHubAuthMethod_IsEnabled(t *testing.T) {
	t.Run("enabled with factory", func(t *testing.T) {
		repo := &MockProviderSourceRepository{}
		factory := provider_source_service.NewProviderSourceFactory(repo)
		method := NewGitHubAuthMethod(factory)
		if !method.IsEnabled() {
			t.Error("IsEnabled() = false, want true")
		}
	})

	t.Run("disabled without factory", func(t *testing.T) {
		method := &GitHubAuthMethod{}
		if method.IsEnabled() {
			t.Error("IsEnabled() = true, want false")
		}
	})
}

// TestGitHubAuthMethod_Authenticate tests successful authentication
func TestGitHubAuthMethod_Authenticate_Success(t *testing.T) {
	sessionData := map[string]interface{}{
		"provider_source": "test-github",
		"github_username":  "test-user",
		"organisations": map[string]string{
			"test-user":   string(sqldb.NamespaceTypeGithubUser),
			"test-org-1":  string(sqldb.NamespaceTypeGithubOrg),
			"test-org-2":  string(sqldb.NamespaceTypeGithubOrg),
		},
	}

	repo := &MockProviderSourceRepository{}
	factory := provider_source_service.NewProviderSourceFactory(repo)
	method := NewGitHubAuthMethod(factory)
	ctx := context.Background()

	authContext, err := method.Authenticate(ctx, sessionData)

	if err != nil {
		t.Fatalf("Authenticate() error = %v", err)
	}

	if authContext == nil {
		t.Fatal("Authenticate() returned nil authContext")
	}

	if !authContext.IsAuthenticated() {
		t.Error("AuthContext.IsAuthenticated() = false, want true")
	}

	if authContext.GetUsername() != "test-user" {
		t.Errorf("AuthContext.GetUsername() = %v, want test-user", authContext.GetUsername())
	}

	if authContext.GetProviderType() != auth.AuthMethodGitHub {
		t.Errorf("AuthContext.GetProviderType() = %v, want %v", authContext.GetProviderType(), auth.AuthMethodGitHub)
	}
}

// TestGitHubAuthMethod_Authenticate_MissingProviderSource tests authentication without provider_source
func TestGitHubAuthMethod_Authenticate_MissingProviderSource(t *testing.T) {
	sessionData := map[string]interface{}{
		"github_username": "test-user",
		"organisations": map[string]string{
			"test-user": string(sqldb.NamespaceTypeGithubUser),
		},
	}

	repo := &MockProviderSourceRepository{}
	factory := provider_source_service.NewProviderSourceFactory(repo)
	method := NewGitHubAuthMethod(factory)
	ctx := context.Background()

	authContext, err := method.Authenticate(ctx, sessionData)

	// Returns nil, nil to indicate this is not a GitHub session (not an error)
	if err != nil {
		t.Errorf("Authenticate() error = %v, want nil", err)
	}

	if authContext != nil {
		t.Error("Authenticate() returned non-nil authContext, want nil")
	}
}

// TestGitHubAuthMethod_Authenticate_MissingUsername tests authentication without username
func TestGitHubAuthMethod_Authenticate_MissingUsername(t *testing.T) {
	sessionData := map[string]interface{}{
		"provider_source": "test-github",
		"organisations": map[string]string{
			"test-user": string(sqldb.NamespaceTypeGithubUser),
		},
	}

	repo := &MockProviderSourceRepository{}
	factory := provider_source_service.NewProviderSourceFactory(repo)
	method := NewGitHubAuthMethod(factory)
	ctx := context.Background()

	authContext, err := method.Authenticate(ctx, sessionData)

	// Returns nil, nil to indicate this is not a valid GitHub session (not an error)
	if err != nil {
		t.Errorf("Authenticate() error = %v, want nil", err)
	}

	if authContext != nil {
		t.Error("Authenticate() returned non-nil authContext, want nil")
	}
}

// TestGitHubAuthMethod_Authenticate_MissingOrganizations tests authentication without organizations
func TestGitHubAuthMethod_Authenticate_MissingOrganizations(t *testing.T) {
	sessionData := map[string]interface{}{
		"provider_source": "test-github",
		"github_username":  "test-user",
	}

	repo := &MockProviderSourceRepository{}
	factory := provider_source_service.NewProviderSourceFactory(repo)
	method := NewGitHubAuthMethod(factory)
	ctx := context.Background()

	authContext, err := method.Authenticate(ctx, sessionData)

	// Should still succeed with empty organizations
	if err != nil {
		t.Fatalf("Authenticate() error = %v", err)
	}

	if authContext == nil {
		t.Fatal("Authenticate() returned nil authContext")
	}

	// Verify username was added to organizations
	permissions := authContext.GetAllNamespacePermissions()
	if _, exists := permissions["test-user"]; !exists {
		t.Error("Username not added to organizations when organizations map was missing")
	}
}

// TestGitHubAuthMethod_Authenticate_InvalidOrganizationsType tests authentication with invalid organizations type
func TestGitHubAuthMethod_Authenticate_InvalidOrganizationsType(t *testing.T) {
	sessionData := map[string]interface{}{
		"provider_source": "test-github",
		"github_username":  "test-user",
		"organisations":   "not-a-map", // Invalid type
	}

	repo := &MockProviderSourceRepository{}
	factory := provider_source_service.NewProviderSourceFactory(repo)
	method := NewGitHubAuthMethod(factory)
	ctx := context.Background()

	authContext, err := method.Authenticate(ctx, sessionData)

	// Should still succeed with just the username in organizations (fallback behavior)
	if err != nil {
		t.Errorf("Authenticate() error = %v, want nil", err)
	}

	if authContext == nil {
		t.Fatal("Authenticate() returned nil authContext")
	}

	// Verify username was added as fallback
	permissions := authContext.GetAllNamespacePermissions()
	if _, exists := permissions["test-user"]; !exists {
		t.Error("Username not added to organizations when organizations type was invalid")
	}
}

// TestGitHubAuthMethod_Authenticate_EmptyOrganizationsMap tests authentication with empty organizations map
func TestGitHubAuthMethod_Authenticate_EmptyOrganizationsMap(t *testing.T) {
	sessionData := map[string]interface{}{
		"provider_source": "test-github",
		"github_username":  "test-user",
		"organisations":   map[string]string{},
	}

	repo := &MockProviderSourceRepository{}
	factory := provider_source_service.NewProviderSourceFactory(repo)
	method := NewGitHubAuthMethod(factory)
	ctx := context.Background()

	authContext, err := method.Authenticate(ctx, sessionData)

	if err != nil {
		t.Fatalf("Authenticate() error = %v", err)
	}

	if authContext == nil {
		t.Fatal("Authenticate() returned nil authContext")
	}

	// Verify username was added to organizations when map was empty
	permissions := authContext.GetAllNamespacePermissions()
	if _, exists := permissions["test-user"]; !exists {
		t.Error("Username not added to organizations when organizations map was empty")
	}
}

// TestGitHubAuthMethod_Authenticate_CheckNamespacePermissions tests namespace permission checking
func TestGitHubAuthMethod_Authenticate_CheckNamespacePermissions(t *testing.T) {
	tests := []struct {
		name            string
		namespace       string
		shouldHaveAccess bool
	}{
		{
			name:            "user has access to own namespace",
			namespace:       "test-user",
			shouldHaveAccess: true,
		},
		{
			name:            "user has access to organization namespace",
			namespace:       "test-org-1",
			shouldHaveAccess: true,
		},
		{
			name:            "user does not have access to other namespace",
			namespace:       "other-org",
			shouldHaveAccess: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sessionData := map[string]interface{}{
				"provider_source": "test-github",
				"github_username":  "test-user",
				"organisations": map[string]interface{}{
					"test-user":  string(sqldb.NamespaceTypeGithubUser),
					"test-org-1": string(sqldb.NamespaceTypeGithubOrg),
				},
			}

			repo := &MockProviderSourceRepository{}
			factory := provider_source_service.NewProviderSourceFactory(repo)
			method := NewGitHubAuthMethod(factory)
			ctx := context.Background()

			authContext, err := method.Authenticate(ctx, sessionData)

			if err != nil {
				t.Fatalf("Authenticate() error = %v", err)
			}

			hasAccess := authContext.CheckNamespaceAccess("MODIFY", tt.namespace)
			if hasAccess != tt.shouldHaveAccess {
				t.Errorf("CheckNamespaceAccess() = %v, want %v for namespace %s", hasAccess, tt.shouldHaveAccess, tt.namespace)
			}
		})
	}
}

// TestGitHubAuthMethod_Authenticate_GetAllNamespacePermissions tests getting all namespace permissions
func TestGitHubAuthMethod_Authenticate_GetAllNamespacePermissions(t *testing.T) {
	sessionData := map[string]interface{}{
		"provider_source": "test-github",
		"github_username":  "test-user",
		"organisations": map[string]interface{}{
			"test-user":  string(sqldb.NamespaceTypeGithubUser),
			"test-org-1": string(sqldb.NamespaceTypeGithubOrg),
			"test-org-2": string(sqldb.NamespaceTypeGithubOrg),
		},
	}

	repo := &MockProviderSourceRepository{}
	factory := provider_source_service.NewProviderSourceFactory(repo)
	method := NewGitHubAuthMethod(factory)
	ctx := context.Background()

	authContext, err := method.Authenticate(ctx, sessionData)

	if err != nil {
		t.Fatalf("Authenticate() error = %v", err)
	}

	permissions := authContext.GetAllNamespacePermissions()

	// GetAllNamespacePermissions returns "FULL" for all accessible namespaces
	// because GitHub users get full access to their own namespaces and orgs
	expectedPermissions := 3
	if len(permissions) != expectedPermissions {
		t.Errorf("GetAllNamespacePermissions() returned %d permissions, want %d", len(permissions), expectedPermissions)
	}

	// Verify each organization has FULL access permission
	if permissions["test-user"] != "FULL" {
		t.Errorf("GetAllNamespacePermissions()[test-user] = %v, want FULL", permissions["test-user"])
	}

	if permissions["test-org-1"] != "FULL" {
		t.Errorf("GetAllNamespacePermissions()[test-org-1] = %v, want FULL", permissions["test-org-1"])
	}

	if permissions["test-org-2"] != "FULL" {
		t.Errorf("GetAllNamespacePermissions()[test-org-2] = %v, want FULL", permissions["test-org-2"])
	}
}

// TestGitHubAuthMethod_Authenticate_CanPublishModuleVersion tests publish permissions
func TestGitHubAuthMethod_Authenticate_CanPublishModuleVersion(t *testing.T) {
	tests := []struct {
		name            string
		namespace       string
		shouldHaveAccess bool
	}{
		{
			name:            "user can publish to own namespace",
			namespace:       "test-user",
			shouldHaveAccess: true,
		},
		{
			name:            "user can publish to organization namespace",
			namespace:       "test-org-1",
			shouldHaveAccess: true,
		},
		{
			name:            "user cannot publish to other namespace",
			namespace:       "other-org",
			shouldHaveAccess: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sessionData := map[string]interface{}{
				"provider_source": "test-github",
				"github_username":  "test-user",
				"organisations": map[string]interface{}{
					"test-user":  string(sqldb.NamespaceTypeGithubUser),
					"test-org-1": string(sqldb.NamespaceTypeGithubOrg),
				},
			}

			repo := &MockProviderSourceRepository{}
			factory := provider_source_service.NewProviderSourceFactory(repo)
			method := NewGitHubAuthMethod(factory)
			ctx := context.Background()

			authContext, err := method.Authenticate(ctx, sessionData)

			if err != nil {
				t.Fatalf("Authenticate() error = %v", err)
			}

			canPublish := authContext.CanPublishModuleVersion(tt.namespace)
			if canPublish != tt.shouldHaveAccess {
				t.Errorf("CanPublishModuleVersion() = %v, want %v for namespace %s", canPublish, tt.shouldHaveAccess, tt.namespace)
			}
		})
	}
}

// TestGitHubAuthMethod_Authenticate_ProviderData tests provider data retrieval
func TestGitHubAuthMethod_Authenticate_ProviderData(t *testing.T) {
	sessionData := map[string]interface{}{
		"provider_source": "test-github",
		"github_username":  "test-user",
		"organisations": map[string]string{
			"test-user":  string(sqldb.NamespaceTypeGithubUser),
			"test-org-1": string(sqldb.NamespaceTypeGithubOrg),
		},
	}

	repo := &MockProviderSourceRepository{}
	factory := provider_source_service.NewProviderSourceFactory(repo)
	method := NewGitHubAuthMethod(factory)
	ctx := context.Background()

	authContext, err := method.Authenticate(ctx, sessionData)

	if err != nil {
		t.Fatalf("Authenticate() error = %v", err)
	}

	providerData := authContext.GetProviderData()
	if providerData == nil {
		t.Fatal("GetProviderData() returned nil")
	}

	// Verify provider data contains expected fields
	if providerData["provider_source"] != "test-github" {
		t.Errorf("GetProviderData()[provider_source] = %v, want test-github", providerData["provider_source"])
	}

	if providerData["github_username"] != "test-user" {
		t.Errorf("GetProviderData()[github_username] = %v, want test-user", providerData["github_username"])
	}

	if providerData["auth_method"] != string(auth.AuthMethodGitHub) {
		t.Errorf("GetProviderData()[auth_method] = %v, want %v", providerData["auth_method"], string(auth.AuthMethodGitHub))
	}
}

// TestGitHubAuthMethod_Authenticate_NotAdmin tests that GitHub auth is not admin by default
func TestGitHubAuthMethod_Authenticate_NotAdmin(t *testing.T) {
	sessionData := map[string]interface{}{
		"provider_source": "test-github",
		"github_username":  "test-user",
		"organisations": map[string]string{
			"test-user": string(sqldb.NamespaceTypeGithubUser),
		},
	}

	repo := &MockProviderSourceRepository{}
	factory := provider_source_service.NewProviderSourceFactory(repo)
	method := NewGitHubAuthMethod(factory)
	ctx := context.Background()

	authContext, err := method.Authenticate(ctx, sessionData)

	if err != nil {
		t.Fatalf("Authenticate() error = %v", err)
	}

	if authContext.IsAdmin() {
		t.Error("GitHub auth should not be admin by default")
	}

	if authContext.IsBuiltInAdmin() {
		t.Error("GitHub auth should not be built-in admin")
	}
}
