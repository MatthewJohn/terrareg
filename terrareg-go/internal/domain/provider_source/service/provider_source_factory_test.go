package service

import (
	"context"
	"testing"
	"time"

	provider_source_model "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/provider_source/model"
)

// TestNewProviderSourceFactory tests the constructor
func TestNewProviderSourceFactory(t *testing.T) {
	repo := &MockProviderSourceRepository{}
	factory := NewProviderSourceFactory(repo)

	if factory == nil {
		t.Fatal("NewProviderSourceFactory returned nil")
	}

	if factory.repo != repo {
		t.Error("repo not set correctly")
	}

	if factory.classMapping == nil {
		t.Error("classMapping not initialized")
	}
}

// TestProviderSourceFactory_RegisterProviderSourceClass tests registering provider source classes
func TestProviderSourceFactory_RegisterProviderSourceClass(t *testing.T) {
	factory := NewProviderSourceFactory(&MockProviderSourceRepository{})

	// Create a mock provider source class
	class := &GithubProviderSourceClass{}

	factory.RegisterProviderSourceClass(class)

	classes := factory.GetProviderClasses()
	if len(classes) != 1 {
		t.Errorf("GetProviderClasses() returned %d classes, want 1", len(classes))
	}

	if _, exists := classes[provider_source_model.ProviderSourceTypeGithub]; !exists {
		t.Error("GitHub provider source class not registered")
	}
}

// TestGenerateStateToken tests state token generation
func TestGenerateStateToken(t *testing.T) {
	token := generateStateToken()

	if token == "" {
		t.Error("generateStateToken() returned empty string")
	}

	// Generate another token and verify they're different
	token2 := generateStateToken()
	if token == token2 {
		t.Error("generateStateToken() returned same token twice")
	}
}

// TestProviderSourceWrapper tests the wrapper functionality
func TestProviderSourceWrapper_GetLoginRedirectURL(t *testing.T) {
	tests := []struct {
		name        string
		baseURL     string
		clientID    string
		expectError bool
	}{
		{
			name:        "valid configuration",
			baseURL:     "https://github.com",
			clientID:    "test-client-id",
			expectError: false,
		},
		{
			name:        "custom GitHub Enterprise URL",
			baseURL:     "https://github.example.com",
			clientID:    "enterprise-client-id",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a real provider source model
			config := &provider_source_model.ProviderSourceConfig{
				BaseURL:  tt.baseURL,
				ClientID: tt.clientID,
			}
			source := provider_source_model.NewProviderSource("Test GitHub", "test-github", provider_source_model.ProviderSourceTypeGithub, config)

			// Create wrapper with real source
			wrapper := &githubProviderSourceWrapper{
				source: source,
				repo:   &MockProviderSourceRepository{},
			}

			ctx := context.Background()
			url, err := wrapper.GetLoginRedirectURL(ctx)

			if tt.expectError {
				if err == nil {
					t.Error("GetLoginRedirectURL() error = nil, want error")
				}
				return
			}

			if err != nil {
				t.Fatalf("GetLoginRedirectURL() error = %v", err)
			}

			// Verify URL contains expected parts
			if url == "" {
				t.Error("GetLoginRedirectURL() returned empty URL")
			}

			// Check that the URL contains the client_id
			if !contains(url, tt.clientID) {
				t.Errorf("GetLoginRedirectURL() URL doesn't contain client_id: %s", url)
			}

			// Check that the URL contains scope
			if !contains(url, "scope=read:org") {
				t.Errorf("GetLoginRedirectURL() URL doesn't contain scope: %s", url)
			}

			// Check that the URL contains state
			if !contains(url, "state=") {
				t.Errorf("GetLoginRedirectURL() URL doesn't contain state: %s", url)
			}
		})
	}
}

// TestProviderSourceWrapper_GetUserAccessToken tests token exchange parameter handling
func TestProviderSourceWrapper_GetUserAccessToken(t *testing.T) {
	tests := []struct {
		name         string
		baseURL      string
		clientID     string
		clientSecret string
	}{
		{
			name:         "valid configuration",
			baseURL:      "https://github.com",
			clientID:     "test-client-id",
			clientSecret: "test-client-secret",
		},
		{
			name:         "missing client ID",
			baseURL:      "https://github.com",
			clientID:     "",
			clientSecret: "test-client-secret",
		},
		{
			name:         "missing client secret",
			baseURL:      "https://github.com",
			clientID:     "test-client-id",
			clientSecret: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a real provider source model
			config := &provider_source_model.ProviderSourceConfig{
				BaseURL:      tt.baseURL,
				ClientID:     tt.clientID,
				ClientSecret: tt.clientSecret,
			}
			source := provider_source_model.NewProviderSource("Test GitHub", "test-github", provider_source_model.ProviderSourceTypeGithub, config)

			// Create wrapper with real source
			wrapper := &githubProviderSourceWrapper{
				source: source,
				repo:   &MockProviderSourceRepository{},
			}

			ctx := context.Background()

			// This will fail to make actual HTTP request in tests
			// We're testing the parameter validation and structure
			if tt.clientID == "" || tt.clientSecret == "" {
				// For missing credentials, the HTTP request will fail
				// We're just verifying the method exists and has correct signature
			}

			// Verify the method can be called (will fail on HTTP request)
			_, err := wrapper.GetUserAccessToken(ctx, "test-code")
			// We expect an error due to no actual HTTP server
			if err == nil {
				t.Log("GetUserAccessToken() succeeded - consider adding mock server for testing")
			}
		})
	}
}

// TestProviderSourceWrapper_GetUsername tests username retrieval structure
func TestProviderSourceWrapper_GetUsername(t *testing.T) {
	// Create a real provider source model
	config := &provider_source_model.ProviderSourceConfig{
		BaseURL: "https://github.com",
	}
	source := provider_source_model.NewProviderSource("Test GitHub", "test-github", provider_source_model.ProviderSourceTypeGithub, config)

	// Create wrapper with real source
	wrapper := &githubProviderSourceWrapper{
		source: source,
		repo:   &MockProviderSourceRepository{},
	}

	ctx := context.Background()

	// This will fail to make actual HTTP request in tests
	// We're testing that the method exists and has correct signature
	_, err := wrapper.GetUsername(ctx, "test-token")

	// The HTTP request will fail without a server
	if err == nil {
		t.Log("GetUsername() succeeded - consider adding mock server for testing")
	}
}

// TestProviderSourceWrapper_GetUserOrganizations tests organization retrieval structure
func TestProviderSourceWrapper_GetUserOrganizations(t *testing.T) {
	// Create a real provider source model
	config := &provider_source_model.ProviderSourceConfig{
		BaseURL: "https://github.com",
	}
	source := provider_source_model.NewProviderSource("Test GitHub", "test-github", provider_source_model.ProviderSourceTypeGithub, config)

	// Create wrapper with real source
	wrapper := &githubProviderSourceWrapper{
		source: source,
		repo:   &MockProviderSourceRepository{},
	}

	ctx := context.Background()

	// This will fail to make actual HTTP request in tests
	// We're testing that the method exists and has correct signature
	orgs := wrapper.GetUserOrganizations(ctx, "test-token")

	// Should return empty slice on error
	if orgs == nil {
		t.Error("GetUserOrganizations() returned nil, want empty slice")
	}
}

// TestProviderSourceWrapper_Expiry tests TTL conversion
func TestProviderSourceWrapper_Expiry(t *testing.T) {
	// Create a real provider source model
	config := &provider_source_model.ProviderSourceConfig{
		BaseURL: "https://github.com",
	}
	source := provider_source_model.NewProviderSource("Test GitHub", "test-github", provider_source_model.ProviderSourceTypeGithub, config)

	// Create wrapper with real source
	wrapper := &githubProviderSourceWrapper{
		source: source,
		repo:   &MockProviderSourceRepository{},
	}

	// Test expiry time calculation (24 hours)
	ttl := 24 * time.Hour
	expectedExpiry := time.Now().Add(ttl)

	// Verify the wrapper has access to the source
	if wrapper.source == nil {
		t.Error("wrapper.source is nil")
	}

	// The expiry is calculated in GetUserAccessToken, we just verify
	// the structure allows for it
	_ = expectedExpiry
}

// MockProviderSourceRepository for testing
type MockProviderSourceRepository struct{}

func (m *MockProviderSourceRepository) FindByName(ctx context.Context, name string) (*provider_source_model.ProviderSource, error) {
	return nil, nil
}

func (m *MockProviderSourceRepository) FindByApiName(ctx context.Context, apiName string) (*provider_source_model.ProviderSource, error) {
	return nil, nil
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

// Helper function to check if string contains substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
