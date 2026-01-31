package auth

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/auth"
	provider_source_model "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/provider_source/model"
	provider_source_service "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/provider_source/service"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/persistence/sqldb"
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
	// For testing, return nil for "nonexistent-github" to test provider source not found scenario
	if apiName == "nonexistent-github" {
		return nil, nil
	}
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

// MockSessionRepository is a mock implementation of SessionRepository for testing
type MockSessionRepository struct {
	session *auth.Session
}

func (m *MockSessionRepository) Create(ctx context.Context, session *auth.Session) error {
	return nil
}

func (m *MockSessionRepository) FindByID(ctx context.Context, id string) (*auth.Session, error) {
	return m.session, nil
}

func (m *MockSessionRepository) Delete(ctx context.Context, sessionID string) error {
	return nil
}

func (m *MockSessionRepository) CleanupExpired(ctx context.Context) error {
	return nil
}

func (m *MockSessionRepository) UpdateProviderSourceAuth(ctx context.Context, sessionID string, data []byte) error {
	return nil
}

// createTestGitHubProviderData creates the provider data JSON that would be stored in a session
func createTestGitHubProviderData(username string, providerSource string, organizations map[string]sqldb.NamespaceType) []byte {
	data := map[string]interface{}{
		"provider_source": providerSource,
		"github_username": username,
		"organisations":   organizations,
		"auth_method":     string(auth.AuthMethodGitHub),
	}
	jsonData, _ := json.Marshal(data)
	return jsonData
}

// TestNewGitHubAuthMethod tests the constructor
func TestNewGitHubAuthMethod(t *testing.T) {
	repo := &MockProviderSourceRepository{}
	factory := provider_source_service.NewProviderSourceFactory(repo)
	sessionRepo := &MockSessionRepository{}
	method := NewGitHubAuthMethod(factory, sessionRepo)

	if method == nil {
		t.Fatal("NewGitHubAuthMethod returned nil")
	}

	if method.providerSourceFactory != factory {
		t.Error("providerSourceFactory not set correctly")
	}

	if method.sessionRepo != sessionRepo {
		t.Error("sessionRepo not set correctly")
	}
}

// TestGitHubAuthMethod_GetProviderType tests GetProviderType
func TestGitHubAuthMethod_GetProviderType(t *testing.T) {
	repo := &MockProviderSourceRepository{}
	factory := provider_source_service.NewProviderSourceFactory(repo)
	sessionRepo := &MockSessionRepository{}
	method := NewGitHubAuthMethod(factory, sessionRepo)
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
		sessionRepo := &MockSessionRepository{}
		method := NewGitHubAuthMethod(factory, sessionRepo)
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

// TestGitHubAuthMethod_Authenticate_Success tests successful authentication
func TestGitHubAuthMethod_Authenticate_Success(t *testing.T) {
	expiry := time.Now().Add(24 * time.Hour)
	orgs := map[string]sqldb.NamespaceType{
		"test-user":  sqldb.NamespaceTypeGithubUser,
		"test-org-1": sqldb.NamespaceTypeGithubOrg,
		"test-org-2": sqldb.NamespaceTypeGithubOrg,
	}
	session := &auth.Session{
		ID:                 "test-session-id",
		Expiry:             expiry,
		ProviderSourceAuth: createTestGitHubProviderData("test-user", "test-github", orgs),
	}

	sessionRepo := &MockSessionRepository{session: session}
	repo := &MockProviderSourceRepository{}
	factory := provider_source_service.NewProviderSourceFactory(repo)
	method := NewGitHubAuthMethod(factory, sessionRepo)

	sessionData := map[string]interface{}{
		"session_id": "test-session-id",
	}
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

// TestGitHubAuthMethod_Authenticate_MissingSessionID tests authentication without session_id
func TestGitHubAuthMethod_Authenticate_MissingSessionID(t *testing.T) {
	sessionRepo := &MockSessionRepository{}
	repo := &MockProviderSourceRepository{}
	factory := provider_source_service.NewProviderSourceFactory(repo)
	method := NewGitHubAuthMethod(factory, sessionRepo)

	sessionData := map[string]interface{}{} // No session_id
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

// TestGitHubAuthMethod_Authenticate_SessionNotFound tests authentication when session not found
func TestGitHubAuthMethod_Authenticate_SessionNotFound(t *testing.T) {
	sessionRepo := &MockSessionRepository{session: nil} // Session not found
	repo := &MockProviderSourceRepository{}
	factory := provider_source_service.NewProviderSourceFactory(repo)
	method := NewGitHubAuthMethod(factory, sessionRepo)

	sessionData := map[string]interface{}{
		"session_id": "nonexistent-session",
	}
	ctx := context.Background()

	authContext, err := method.Authenticate(ctx, sessionData)

	// Returns nil, nil to indicate session not found (not an error)
	if err != nil {
		t.Errorf("Authenticate() error = %v, want nil", err)
	}

	if authContext != nil {
		t.Error("Authenticate() returned non-nil authContext, want nil")
	}
}

// TestGitHubAuthMethod_Authenticate_ExpiredSession tests authentication with expired session
func TestGitHubAuthMethod_Authenticate_ExpiredSession(t *testing.T) {
	expiry := time.Now().Add(-1 * time.Hour) // Expired
	orgs := map[string]sqldb.NamespaceType{
		"test-user": sqldb.NamespaceTypeGithubUser,
	}
	session := &auth.Session{
		ID:                 "test-session-id",
		Expiry:             expiry,
		ProviderSourceAuth: createTestGitHubProviderData("test-user", "test-github", orgs),
	}

	sessionRepo := &MockSessionRepository{session: session}
	repo := &MockProviderSourceRepository{}
	factory := provider_source_service.NewProviderSourceFactory(repo)
	method := NewGitHubAuthMethod(factory, sessionRepo)

	sessionData := map[string]interface{}{
		"session_id": "test-session-id",
	}
	ctx := context.Background()

	authContext, err := method.Authenticate(ctx, sessionData)

	// Returns nil, nil for expired session (not an error)
	if err != nil {
		t.Errorf("Authenticate() error = %v, want nil", err)
	}

	if authContext != nil {
		t.Error("Authenticate() returned non-nil authContext for expired session, want nil")
	}
}

// TestGitHubAuthMethod_Authenticate_MissingOrganizations tests authentication without organizations
func TestGitHubAuthMethod_Authenticate_MissingOrganizations(t *testing.T) {
	expiry := time.Now().Add(24 * time.Hour)
	// Provider data without organizations (should add username as fallback)
	data := map[string]interface{}{
		"provider_source": "test-github",
		"github_username": "test-user",
		"auth_method":     string(auth.AuthMethodGitHub),
	}
	jsonData, _ := json.Marshal(data)

	session := &auth.Session{
		ID:                 "test-session-id",
		Expiry:             expiry,
		ProviderSourceAuth: jsonData,
	}

	sessionRepo := &MockSessionRepository{session: session}
	repo := &MockProviderSourceRepository{}
	factory := provider_source_service.NewProviderSourceFactory(repo)
	method := NewGitHubAuthMethod(factory, sessionRepo)

	sessionData := map[string]interface{}{
		"session_id": "test-session-id",
	}
	ctx := context.Background()

	authContext, err := method.Authenticate(ctx, sessionData)

	// Should still succeed with empty organizations (username added as fallback)
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

// TestGitHubAuthMethod_Authenticate_CheckNamespacePermissions tests namespace permission checking
func TestGitHubAuthMethod_Authenticate_CheckNamespacePermissions(t *testing.T) {
	tests := []struct {
		name             string
		namespace        string
		shouldHaveAccess bool
	}{
		{
			name:             "user has access to own namespace",
			namespace:        "test-user",
			shouldHaveAccess: true,
		},
		{
			name:             "user has access to organization namespace",
			namespace:        "test-org-1",
			shouldHaveAccess: true,
		},
		{
			name:             "user does not have access to other namespace",
			namespace:        "other-org",
			shouldHaveAccess: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expiry := time.Now().Add(24 * time.Hour)
			orgs := map[string]sqldb.NamespaceType{
				"test-user":  sqldb.NamespaceTypeGithubUser,
				"test-org-1": sqldb.NamespaceTypeGithubOrg,
			}
			session := &auth.Session{
				ID:                 "test-session-id",
				Expiry:             expiry,
				ProviderSourceAuth: createTestGitHubProviderData("test-user", "test-github", orgs),
			}

			sessionRepo := &MockSessionRepository{session: session}
			repo := &MockProviderSourceRepository{}
			factory := provider_source_service.NewProviderSourceFactory(repo)
			method := NewGitHubAuthMethod(factory, sessionRepo)

			sessionData := map[string]interface{}{
				"session_id": "test-session-id",
			}
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
	expiry := time.Now().Add(24 * time.Hour)
	orgs := map[string]sqldb.NamespaceType{
		"test-user":  sqldb.NamespaceTypeGithubUser,
		"test-org-1": sqldb.NamespaceTypeGithubOrg,
		"test-org-2": sqldb.NamespaceTypeGithubOrg,
	}
	session := &auth.Session{
		ID:                 "test-session-id",
		Expiry:             expiry,
		ProviderSourceAuth: createTestGitHubProviderData("test-user", "test-github", orgs),
	}

	sessionRepo := &MockSessionRepository{session: session}
	repo := &MockProviderSourceRepository{}
	factory := provider_source_service.NewProviderSourceFactory(repo)
	method := NewGitHubAuthMethod(factory, sessionRepo)

	sessionData := map[string]interface{}{
		"session_id": "test-session-id",
	}
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
		name             string
		namespace        string
		shouldHaveAccess bool
	}{
		{
			name:             "user can publish to own namespace",
			namespace:        "test-user",
			shouldHaveAccess: true,
		},
		{
			name:             "user can publish to organization namespace",
			namespace:        "test-org-1",
			shouldHaveAccess: true,
		},
		{
			name:             "user cannot publish to other namespace",
			namespace:        "other-org",
			shouldHaveAccess: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expiry := time.Now().Add(24 * time.Hour)
			orgs := map[string]sqldb.NamespaceType{
				"test-user":  sqldb.NamespaceTypeGithubUser,
				"test-org-1": sqldb.NamespaceTypeGithubOrg,
			}
			session := &auth.Session{
				ID:                 "test-session-id",
				Expiry:             expiry,
				ProviderSourceAuth: createTestGitHubProviderData("test-user", "test-github", orgs),
			}

			sessionRepo := &MockSessionRepository{session: session}
			repo := &MockProviderSourceRepository{}
			factory := provider_source_service.NewProviderSourceFactory(repo)
			method := NewGitHubAuthMethod(factory, sessionRepo)

			sessionData := map[string]interface{}{
				"session_id": "test-session-id",
			}
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
	expiry := time.Now().Add(24 * time.Hour)
	orgs := map[string]sqldb.NamespaceType{
		"test-user":  sqldb.NamespaceTypeGithubUser,
		"test-org-1": sqldb.NamespaceTypeGithubOrg,
	}
	session := &auth.Session{
		ID:                 "test-session-id",
		Expiry:             expiry,
		ProviderSourceAuth: createTestGitHubProviderData("test-user", "test-github", orgs),
	}

	sessionRepo := &MockSessionRepository{session: session}
	repo := &MockProviderSourceRepository{}
	factory := provider_source_service.NewProviderSourceFactory(repo)
	method := NewGitHubAuthMethod(factory, sessionRepo)

	sessionData := map[string]interface{}{
		"session_id": "test-session-id",
	}
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
	expiry := time.Now().Add(24 * time.Hour)
	orgs := map[string]sqldb.NamespaceType{
		"test-user": sqldb.NamespaceTypeGithubUser,
	}
	session := &auth.Session{
		ID:                 "test-session-id",
		Expiry:             expiry,
		ProviderSourceAuth: createTestGitHubProviderData("test-user", "test-github", orgs),
	}

	sessionRepo := &MockSessionRepository{session: session}
	repo := &MockProviderSourceRepository{}
	factory := provider_source_service.NewProviderSourceFactory(repo)
	method := NewGitHubAuthMethod(factory, sessionRepo)

	sessionData := map[string]interface{}{
		"session_id": "test-session-id",
	}
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

// TestGitHubAuthMethod_Authenticate_ProviderSourceNotFound tests when provider source doesn't exist
func TestGitHubAuthMethod_Authenticate_ProviderSourceNotFound(t *testing.T) {
	expiry := time.Now().Add(24 * time.Hour)
	// Provider data with non-existent provider source
	data := map[string]interface{}{
		"provider_source": "nonexistent-github",
		"github_username": "test-user",
		"auth_method":     string(auth.AuthMethodGitHub),
	}
	jsonData, _ := json.Marshal(data)

	session := &auth.Session{
		ID:                 "test-session-id",
		Expiry:             expiry,
		ProviderSourceAuth: jsonData,
	}

	sessionRepo := &MockSessionRepository{session: session}
	repo := &MockProviderSourceRepository{} // Will return not found for nonexistent-github
	factory := provider_source_service.NewProviderSourceFactory(repo)
	method := NewGitHubAuthMethod(factory, sessionRepo)

	sessionData := map[string]interface{}{
		"session_id": "test-session-id",
	}
	ctx := context.Background()

	authContext, err := method.Authenticate(ctx, sessionData)

	// Should return nil, nil when provider source not found (lets other auth methods try)
	if err != nil {
		t.Errorf("Authenticate() error = %v, want nil", err)
	}

	if authContext != nil {
		t.Error("Authenticate() returned non-nil authContext when provider source not found")
	}
}

// TestGitHubAuthMethod_Authenticate_NilFactory tests authentication with nil factory
func TestGitHubAuthMethod_Authenticate_NilFactory(t *testing.T) {
	expiry := time.Now().Add(24 * time.Hour)
	orgs := map[string]sqldb.NamespaceType{
		"test-user": sqldb.NamespaceTypeGithubUser,
	}
	session := &auth.Session{
		ID:                 "test-session-id",
		Expiry:             expiry,
		ProviderSourceAuth: createTestGitHubProviderData("test-user", "test-github", orgs),
	}

	sessionRepo := &MockSessionRepository{session: session}
	method := &GitHubAuthMethod{sessionRepo: sessionRepo}

	sessionData := map[string]interface{}{
		"session_id": "test-session-id",
	}
	ctx := context.Background()

	authContext, err := method.Authenticate(ctx, sessionData)

	// With nil factory, it should skip provider source validation and succeed
	if err != nil {
		t.Errorf("Authenticate() error = %v, want nil", err)
	}

	if authContext == nil {
		t.Error("Authenticate() returned nil authContext")
	}

	if authContext != nil && !authContext.IsAuthenticated() {
		t.Error("AuthContext should be authenticated")
	}
}

// TestGitHubAuthContext_CaseInsensitiveNamespaceMatching tests case-insensitive namespace matching
func TestGitHubAuthContext_CaseInsensitiveNamespaceMatching(t *testing.T) {
	tests := []struct {
		name             string
		namespace        string
		organizations    map[string]sqldb.NamespaceType
		shouldHaveAccess bool
	}{
		{
			name:      "exact case match - user",
			namespace: "test-user",
			organizations: map[string]sqldb.NamespaceType{
				"test-user": sqldb.NamespaceTypeGithubUser,
			},
			shouldHaveAccess: true,
		},
		{
			name:      "lowercase namespace match - user",
			namespace: "test-user",
			organizations: map[string]sqldb.NamespaceType{
				"Test-User": sqldb.NamespaceTypeGithubUser,
			},
			shouldHaveAccess: true,
		},
		{
			name:      "uppercase namespace match - user",
			namespace: "TEST-USER",
			organizations: map[string]sqldb.NamespaceType{
				"test-user": sqldb.NamespaceTypeGithubUser,
			},
			shouldHaveAccess: true,
		},
		{
			name:      "mixed case org match",
			namespace: "My-Org",
			organizations: map[string]sqldb.NamespaceType{
				"my-org": sqldb.NamespaceTypeGithubOrg,
			},
			shouldHaveAccess: true,
		},
		{
			name:      "case mismatch no access",
			namespace: "other-org",
			organizations: map[string]sqldb.NamespaceType{
				"test-user": sqldb.NamespaceTypeGithubUser,
			},
			shouldHaveAccess: false,
		},
		{
			name:      "multiple orgs with different cases",
			namespace: "My-Company",
			organizations: map[string]sqldb.NamespaceType{
				"test-user":   sqldb.NamespaceTypeGithubUser,
				"my-company":  sqldb.NamespaceTypeGithubOrg,
				"Another-Org": sqldb.NamespaceTypeGithubOrg,
			},
			shouldHaveAccess: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			authCtx := auth.NewGitHubAuthContext(
				context.Background(),
				"test-github",
				"test-user",
				tt.organizations,
			)

			hasAccess := authCtx.CheckNamespaceAccess("MODIFY", tt.namespace)
			if hasAccess != tt.shouldHaveAccess {
				t.Errorf("CheckNamespaceAccess(MODIFY, %s) = %v, want %v", tt.namespace, hasAccess, tt.shouldHaveAccess)
			}
		})
	}
}
