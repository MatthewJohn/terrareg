package provider_source

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/provider_source/model"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/provider_source/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestGithubProviderSource_Config tests that the config is loaded correctly
func TestGithubProviderSource_Config(t *testing.T) {
	expectedConfig := &model.ProviderSourceConfig{
		BaseURL:                  "https://github.example-test.com",
		ApiURL:                   "https://api.github.example-test.com",
		ClientID:                 "unittest-github-client-id",
		ClientSecret:             "unittest-github-client-secret",
		LoginButtonText:          "Unit Test Github Login",
		PrivateKeyPath:           "./unittest-path-to-private-key.pem",
		AppID:                    "954956",
		DefaultAccessToken:        "",
		DefaultInstallationID:     "",
		AutoGenerateNamespaces:   false,
	}

	mockPSRepo := &MockProviderSourceRepository{
		findByNameFunc: func(ctx context.Context, name string) (*model.ProviderSource, error) {
			return model.NewProviderSource(
				name,
				"test-provider-source",
				model.ProviderSourceTypeGithub,
				expectedConfig,
			), nil
		},
	}
	ghClass := service.NewGithubProviderSourceClass()

	gh := NewGithubProviderSource("test-name", mockPSRepo, ghClass)

	config, err := gh.Config(context.Background())
	require.NoError(t, err)
	require.NotNil(t, config)

	assert.Equal(t, expectedConfig.BaseURL, config.BaseURL)
	assert.Equal(t, expectedConfig.ApiURL, config.ApiURL)
	assert.Equal(t, expectedConfig.ClientID, config.ClientID)
	assert.Equal(t, expectedConfig.ClientSecret, config.ClientSecret)
	assert.Equal(t, expectedConfig.LoginButtonText, config.LoginButtonText)
}

// TestGithubProviderSource_LoginButtonText tests the LoginButtonText method
func TestGithubProviderSource_LoginButtonText(t *testing.T) {
	expectedConfig := &model.ProviderSourceConfig{
		BaseURL:                  "https://github.example-test.com",
		ApiURL:                   "https://api.github.example-test.com",
		ClientID:                 "unittest-github-client-id",
		ClientSecret:             "unittest-github-client-secret",
		LoginButtonText:          "Unit Test Github Login",
		PrivateKeyPath:           "./unittest-path-to-private-key.pem",
		AppID:                    "954956",
	}

	mockPSRepo := &MockProviderSourceRepository{
		findByNameFunc: func(ctx context.Context, name string) (*model.ProviderSource, error) {
			return model.NewProviderSource(
				name,
				"test-provider-source",
				model.ProviderSourceTypeGithub,
				expectedConfig,
			), nil
		},
	}
	ghClass := service.NewGithubProviderSourceClass()

	gh := NewGithubProviderSource("test-name", mockPSRepo, ghClass)

	text, err := gh.LoginButtonText(context.Background())
	require.NoError(t, err)
	assert.Equal(t, "Unit Test Github Login", text)
}

// TestGithubProviderSource_GetLoginRedirectURL tests the GetLoginRedirectURL method
func TestGithubProviderSource_GetLoginRedirectURL(t *testing.T) {
	expectedConfig := &model.ProviderSourceConfig{
		BaseURL:         "https://github.example-test.com",
		ApiURL:          "https://api.github.example-test.com",
		ClientID:        "unittest-github-client-id",
		ClientSecret:    "unittest-github-client-secret",
		LoginButtonText: "Unit Test Github Login",
		PrivateKeyPath:  "./unittest-path-to-private-key.pem",
		AppID:           "954956",
	}

	mockPSRepo := &MockProviderSourceRepository{
		findByNameFunc: func(ctx context.Context, name string) (*model.ProviderSource, error) {
			return model.NewProviderSource(
				name,
				"test-provider-source",
				model.ProviderSourceTypeGithub,
				expectedConfig,
			), nil
		},
	}
	ghClass := service.NewGithubProviderSourceClass()

	gh := NewGithubProviderSource("test-name", mockPSRepo, ghClass)

	url, err := gh.GetLoginRedirectURL(context.Background())
	require.NoError(t, err)
	assert.Equal(t, "https://github.example-test.com/login/oauth/authorize?client_id=unittest-github-client-id", url)
}

// TestGithubProviderSource_AutoGenerateGithubOrganisationNamespaces tests the AutoGenerateGithubOrganisationNamespaces method
func TestGithubProviderSource_AutoGenerateGithubOrganisationNamespaces(t *testing.T) {
	tests := []struct {
		name     string
		config   *model.ProviderSourceConfig
		expected bool
	}{
		{
			name: "enabled true",
			config: &model.ProviderSourceConfig{
				BaseURL:                  "https://github.example-test.com",
				ApiURL:                   "https://api.github.example-test.com",
				ClientID:                 "unittest-github-client-id",
				ClientSecret:             "unittest-github-client-secret",
				LoginButtonText:          "Unit Test Github Login",
				PrivateKeyPath:           "./unittest-path-to-private-key.pem",
				AppID:                    "954956",
				AutoGenerateNamespaces:   true,
			},
			expected: true,
		},
		{
			name: "enabled false",
			config: &model.ProviderSourceConfig{
				BaseURL:                  "https://github.example-test.com",
				ApiURL:                   "https://api.github.example-test.com",
				ClientID:                 "unittest-github-client-id",
				ClientSecret:             "unittest-github-client-secret",
				LoginButtonText:          "Unit Test Github Login",
				PrivateKeyPath:           "./unittest-path-to-private-key.pem",
				AppID:                    "954956",
				AutoGenerateNamespaces:   false,
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockPSRepo := &MockProviderSourceRepository{
				findByNameFunc: func(ctx context.Context, name string) (*model.ProviderSource, error) {
					return model.NewProviderSource(
						name,
						"test-provider-source",
						model.ProviderSourceTypeGithub,
						tt.config,
					), nil
				},
			}
			ghClass := service.NewGithubProviderSourceClass()

			gh := NewGithubProviderSource("test-name", mockPSRepo, ghClass)

			result, err := gh.AutoGenerateGithubOrganisationNamespaces(context.Background())
			require.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestGithubProviderSource_IsEnabled tests the IsEnabled method
func TestGithubProviderSource_IsEnabled(t *testing.T) {
	tests := []struct {
		name     string
		config   *model.ProviderSourceConfig
		expected bool
	}{
		{
			name: "all required fields present",
			config: &model.ProviderSourceConfig{
				BaseURL:         "https://github.example-test.com",
				ApiURL:          "https://api.github.example-test.com",
				ClientID:        "unittest-github-client-id",
				ClientSecret:    "unittest-github-client-secret",
				LoginButtonText: "Unit Test Github Login",
				PrivateKeyPath:  "./unittest-path-to-private-key.pem",
				AppID:           "954956",
			},
			expected: true,
		},
		{
			name: "missing client_id",
			config: &model.ProviderSourceConfig{
				BaseURL:         "https://github.example-test.com",
				ApiURL:          "https://api.github.example-test.com",
				ClientSecret:    "unittest-github-client-secret",
				LoginButtonText: "Unit Test Github Login",
				PrivateKeyPath:  "./unittest-path-to-private-key.pem",
				AppID:           "954956",
			},
			expected: false,
		},
		{
			name: "missing client_secret",
			config: &model.ProviderSourceConfig{
				BaseURL:         "https://github.example-test.com",
				ApiURL:          "https://api.github.example-test.com",
				ClientID:        "unittest-github-client-id",
				LoginButtonText: "Unit Test Github Login",
				PrivateKeyPath:  "./unittest-path-to-private-key.pem",
				AppID:           "954956",
			},
			expected: false,
		},
		{
			name: "missing base_url",
			config: &model.ProviderSourceConfig{
				ApiURL:          "https://api.github.example-test.com",
				ClientID:        "unittest-github-client-id",
				ClientSecret:    "unittest-github-client-secret",
				LoginButtonText: "Unit Test Github Login",
				PrivateKeyPath:  "./unittest-path-to-private-key.pem",
				AppID:           "954956",
			},
			expected: false,
		},
		{
			name: "missing api_url",
			config: &model.ProviderSourceConfig{
				BaseURL:         "https://github.example-test.com",
				ClientID:        "unittest-github-client-id",
				ClientSecret:    "unittest-github-client-secret",
				LoginButtonText: "Unit Test Github Login",
				PrivateKeyPath:  "./unittest-path-to-private-key.pem",
				AppID:           "954956",
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockPSRepo := &MockProviderSourceRepository{
				findByNameFunc: func(ctx context.Context, name string) (*model.ProviderSource, error) {
					return model.NewProviderSource(
						name,
						"test-provider-source",
						model.ProviderSourceTypeGithub,
						tt.config,
					), nil
				},
			}
			ghClass := service.NewGithubProviderSourceClass()

			gh := NewGithubProviderSource("test-name", mockPSRepo, ghClass)

			result, err := gh.IsEnabled(context.Background())
			require.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestGithubProviderSource_GetPublicSourceURL tests the GetPublicSourceURL method
func TestGithubProviderSource_GetPublicSourceURL(t *testing.T) {
	expectedConfig := &model.ProviderSourceConfig{
		BaseURL:         "https://github.example-test.com",
		ApiURL:          "https://api.github.example-test.com",
		ClientID:        "unittest-github-client-id",
		ClientSecret:    "unittest-github-client-secret",
		LoginButtonText: "Unit Test Github Login",
		PrivateKeyPath:  "./unittest-path-to-private-key.pem",
		AppID:           "954956",
	}

	mockPSRepo := &MockProviderSourceRepository{
		findByNameFunc: func(ctx context.Context, name string) (*model.ProviderSource, error) {
			return model.NewProviderSource(
				name,
				"test-provider-source",
				model.ProviderSourceTypeGithub,
				expectedConfig,
			), nil
		},
	}
	ghClass := service.NewGithubProviderSourceClass()

	gh := NewGithubProviderSource("test-name", mockPSRepo, ghClass)

	// Mock repository
	mockRepo := &mockRepositoryImpl{
		owner: "some-organisation",
		name:  "terraform-provider-test",
	}

	url, err := gh.GetPublicSourceURL(context.Background(), mockRepo)
	require.NoError(t, err)
	assert.Equal(t, "https://github.example-test.com/some-organisation/terraform-provider-test", url)
}

// TestGithubProviderSource_GetPublicArtifactDownloadURL tests the GetPublicArtifactDownloadURL method
func TestGithubProviderSource_GetPublicArtifactDownloadURL(t *testing.T) {
	expectedConfig := &model.ProviderSourceConfig{
		BaseURL:         "https://github.example-test.com",
		ApiURL:          "https://api.github.example-test.com",
		ClientID:        "unittest-github-client-id",
		ClientSecret:    "unittest-github-client-secret",
		LoginButtonText: "Unit Test Github Login",
		PrivateKeyPath:  "./unittest-path-to-private-key.pem",
		AppID:           "954956",
	}

	mockPSRepo := &MockProviderSourceRepository{
		findByNameFunc: func(ctx context.Context, name string) (*model.ProviderSource, error) {
			return model.NewProviderSource(
				name,
				"test-provider-source",
				model.ProviderSourceTypeGithub,
				expectedConfig,
			), nil
		},
	}
	ghClass := service.NewGithubProviderSourceClass()

	gh := NewGithubProviderSource("test-name", mockPSRepo, ghClass)

	// Mock repository and provider version
	testRepo := &mockRepositoryImpl{
		owner: "some-organisation",
		name:  "terraform-provider-test",
	}
	mockPV := &mockProviderVersionImpl{
		repo:   testRepo,
		gitTag: "v2.3.1",
	}

	url, err := gh.GetPublicArtifactDownloadURL(context.Background(), mockPV, "test-artifact.tar.gz")
	require.NoError(t, err)
	assert.Equal(t, "https://github.example-test.com/some-organisation/terraform-provider-test/releases/download/v2.3.1/test-artifact.tar.gz", url)
}

// mockRepositoryImpl is a simple mock of Repository for testing
type mockRepositoryImpl struct {
	owner string
	name  string
}

func (m *mockRepositoryImpl) Owner() string { return m.owner }
func (m *mockRepositoryImpl) Name() string  { return m.name }

// mockProviderVersionImpl is a simple mock of ProviderVersion for testing
type mockProviderVersionImpl struct {
	repo   Repository
	gitTag string
}

func (m *mockProviderVersionImpl) Repository() (Repository, error) { return m.repo, nil }
func (m *mockProviderVersionImpl) GitTag() (string, error)        { return m.gitTag, nil }

// TestGithubProviderSource_GetUserAccessToken tests the GetUserAccessToken method
// Python reference: test_get_user_access_token
func TestGithubProviderSource_GetUserAccessToken(t *testing.T) {
	tests := []struct {
		name          string
		responseCode  int
		responseBody  string
		expectedToken string
	}{
		{
			name:          "valid token exchange",
			responseCode:  200,
			responseBody:  "first_param=125132&access_token=unittest-access-token&someotherparam=123",
			expectedToken: "unittest-access-token",
		},
		{
			name:          "invalid response code",
			responseCode:  400,
			responseBody:  "",
			expectedToken: "",
		},
		{
			name:          "invalid response data",
			responseCode:  200,
			responseBody:  "first_param=125132&someotherparam=123",
			expectedToken: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Mock HTTP server - must be created before setting config
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Verify request is to base URL, not API URL
				assert.Contains(t, r.URL.String(), "/login/oauth/access_token")
				assert.Equal(t, "POST", r.Method)

				// Verify form data (basic check)
				body, _ := io.ReadAll(r.Body)
				assert.Contains(t, string(body), "client_id=unittest-github-client-id")
				assert.Contains(t, string(body), "client_secret=unittest-github-client-secret")
				assert.Contains(t, string(body), "code=abcdef-inputcode")

				w.WriteHeader(tt.responseCode)
				w.Write([]byte(tt.responseBody))
			}))
			defer server.Close()

			// Set baseURL to test server URL
			expectedConfig := &model.ProviderSourceConfig{
				BaseURL:         server.URL,
				ApiURL:          "https://api.github.example-test.com",
				ClientID:        "unittest-github-client-id",
				ClientSecret:    "unittest-github-client-secret",
				LoginButtonText: "Unit Test Github Login",
				PrivateKeyPath:   "./unittest-path-to-private-key.pem",
				AppID:           "954956",
			}

			mockPSRepo := &MockProviderSourceRepository{
				findByNameFunc: func(ctx context.Context, name string) (*model.ProviderSource, error) {
					return model.NewProviderSource(
						name,
						"test-provider-source",
						model.ProviderSourceTypeGithub,
						expectedConfig,
					), nil
				},
			}
			ghClass := service.NewGithubProviderSourceClass()

			gh := NewGithubProviderSource("test-name", mockPSRepo, ghClass)

			token, err := gh.GetUserAccessToken(context.Background(), "abcdef-inputcode")
			require.NoError(t, err)
			assert.Equal(t, tt.expectedToken, token)
		})
	}
}

// TestGithubProviderSource_GetUsername tests the GetUsername method
// Python reference: test_get_username
func TestGithubProviderSource_GetUsername(t *testing.T) {
	tests := []struct {
		name           string
		accessToken    string
		responseCode   int
		responseData   map[string]interface{}
		expectedResult string
	}{
		{
			name:        "valid username",
			accessToken: "test-token",
			responseCode: 200,
			responseData: map[string]interface{}{
				"login": "testuser",
			},
			expectedResult: "testuser",
		},
		{
			name:           "empty access token",
			accessToken:    "",
			responseCode:   0,
			responseData:   nil,
			expectedResult: "",
		},
		{
			name:           "401 unauthorized",
			accessToken:    "invalid-token",
			responseCode:   401,
			responseData:   nil,
			expectedResult: "",
		},
		{
			name:        "missing login in response",
			accessToken: "test-token",
			responseCode: 200,
			responseData: map[string]interface{}{
				"id": 12345,
			},
			expectedResult: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Mock HTTP server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if tt.accessToken == "" {
					// No request should be made for empty token
					t.Fatal("Unexpected request for empty access token")
				}
				w.WriteHeader(tt.responseCode)
				if tt.responseData != nil {
					json.NewEncoder(w).Encode(tt.responseData)
				}
			}))
			defer server.Close()

			// Set apiURL to test server URL
			expectedConfig := &model.ProviderSourceConfig{
				BaseURL:         "https://github.example-test.com",
				ApiURL:          server.URL,
				ClientID:        "unittest-github-client-id",
				ClientSecret:    "unittest-github-client-secret",
				LoginButtonText: "Unit Test Github Login",
				PrivateKeyPath:   "./unittest-path-to-private-key.pem",
				AppID:           "954956",
			}

			mockPSRepo := &MockProviderSourceRepository{
				findByNameFunc: func(ctx context.Context, name string) (*model.ProviderSource, error) {
					return model.NewProviderSource(
						name,
						"test-provider-source",
						model.ProviderSourceTypeGithub,
						expectedConfig,
					), nil
				},
			}
			ghClass := service.NewGithubProviderSourceClass()

			gh := NewGithubProviderSource("test-name", mockPSRepo, ghClass)

			username, err := gh.GetUsername(context.Background(), tt.accessToken)
			require.NoError(t, err)
			assert.Equal(t, tt.expectedResult, username)
		})
	}
}

// TestGithubProviderSource_GetUserOrganisations tests the GetUserOrganisations method
// Python reference: test_get_user_organisations
func TestGithubProviderSource_GetUserOrganisations(t *testing.T) {
	tests := []struct {
		name         string
		accessToken  string
		responseCode int
		responseData []map[string]interface{}
		expectedOrgs []string
	}{
		{
			name:        "multiple organisations with active admin membership",
			accessToken: "test-token",
			responseCode: 200,
			responseData: []map[string]interface{}{
				{
					"state": "active",
					"role":  "admin",
					"organization": map[string]interface{}{
						"login": "org1",
					},
				},
				{
					"state": "active",
					"role":  "admin",
					"organization": map[string]interface{}{
						"login": "org2",
					},
				},
				{
					"state": "inactive",
					"role":  "admin",
					"organization": map[string]interface{}{
						"login": "org3",
					},
				},
				{
					"state": "active",
					"role":  "member",
					"organization": map[string]interface{}{
						"login": "org4",
					},
				},
			},
			expectedOrgs: []string{"org1", "org2"},
		},
		{
			name:         "empty access token",
			accessToken:  "",
			responseCode: 0,
			responseData: nil,
			expectedOrgs: []string{},
		},
		{
			name:         "no organisations",
			accessToken:  "test-token",
			responseCode: 200,
			responseData: []map[string]interface{}{},
			expectedOrgs: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Mock HTTP server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if tt.accessToken == "" {
					t.Fatal("Unexpected request for empty access token")
				}
				w.WriteHeader(tt.responseCode)
				if tt.responseData != nil {
					json.NewEncoder(w).Encode(tt.responseData)
				}
			}))
			defer server.Close()

			// Set apiURL to test server URL
			expectedConfig := &model.ProviderSourceConfig{
				BaseURL:         "https://github.example-test.com",
				ApiURL:          server.URL,
				ClientID:        "unittest-github-client-id",
				ClientSecret:    "unittest-github-client-secret",
				LoginButtonText: "Unit Test Github Login",
				PrivateKeyPath:   "./unittest-path-to-private-key.pem",
				AppID:           "954956",
			}

			mockPSRepo := &MockProviderSourceRepository{
				findByNameFunc: func(ctx context.Context, name string) (*model.ProviderSource, error) {
					return model.NewProviderSource(
						name,
						"test-provider-source",
						model.ProviderSourceTypeGithub,
						expectedConfig,
					), nil
				},
			}
			ghClass := service.NewGithubProviderSourceClass()

			gh := NewGithubProviderSource("test-name", mockPSRepo, ghClass)

			orgs, err := gh.GetUserOrganisations(context.Background(), tt.accessToken)
			require.NoError(t, err)
			assert.Equal(t, tt.expectedOrgs, orgs)
		})
	}
}

// TestGithubProviderSource_GetDefaultAccessToken tests the GetDefaultAccessToken method
// Python reference: test__get_default_access_token
func TestGithubProviderSource_GetDefaultAccessToken(t *testing.T) {
	tests := []struct {
		name                  string
		defaultInstallationID string
		defaultAccessToken    string
		expectedToken         string
	}{
		{
			name:                  "prefer default_installation_id",
			defaultInstallationID: "test-installation-id",
			defaultAccessToken:    "fallback-token",
			expectedToken:         "", // Will be from GenerateAppInstallationToken - returns "" if no key
		},
		{
			name:                  "fallback to default_access_token",
			defaultInstallationID: "",
			defaultAccessToken:    "fallback-token",
			expectedToken:         "fallback-token",
		},
		{
			name:                  "both empty",
			defaultInstallationID: "",
			defaultAccessToken:    "",
			expectedToken:         "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expectedConfig := &model.ProviderSourceConfig{
				BaseURL:               "https://github.example-test.com",
				ApiURL:                "https://api.github.example-test.com",
				ClientID:              "unittest-github-client-id",
				ClientSecret:          "unittest-github-client-secret",
				LoginButtonText:       "Unit Test Github Login",
				PrivateKeyPath:        "./unittest-path-to-private-key.pem",
				AppID:                 "954956",
				DefaultInstallationID: tt.defaultInstallationID,
				DefaultAccessToken:     tt.defaultAccessToken,
			}

			mockPSRepo := &MockProviderSourceRepository{
				findByNameFunc: func(ctx context.Context, name string) (*model.ProviderSource, error) {
					return model.NewProviderSource(
						name,
						"test-provider-source",
						model.ProviderSourceTypeGithub,
						expectedConfig,
					), nil
				},
			}
			ghClass := service.NewGithubProviderSourceClass()

			gh := NewGithubProviderSource("test-name", mockPSRepo, ghClass)

			token, err := gh.GetDefaultAccessToken(context.Background())
			require.NoError(t, err)
			assert.Equal(t, tt.expectedToken, token)
		})
	}
}

// TestGithubProviderSource_GenerateAppInstallationToken tests the GenerateAppInstallationToken method
// Python reference: test_generate_app_installation_token
func TestGithubProviderSource_GenerateAppInstallationToken(t *testing.T) {
	tests := []struct {
		name                string
		installationID      string
		responseCode        int
		responseBody        map[string]interface{}
		providePrivateKey   bool
		expectedToken       string
		expectedError       bool
	}{
		{
			name:           "valid token generation",
			installationID: "unittest-installation-id1",
			responseCode:   201,
			responseBody:   map[string]interface{}{"token": "unittest-access-token"},
			providePrivateKey: true,
			expectedToken:  "unittest-access-token",
		},
		{
			name:           "no token in response",
			installationID: "unittest-installation-id2",
			responseCode:   201,
			responseBody:   map[string]interface{}{},
			providePrivateKey: true,
			expectedToken:  "",
		},
		{
			name:           "invalid response code",
			installationID: "unittest-installation-id3",
			responseCode:   403,
			responseBody:   map[string]interface{}{},
			providePrivateKey: true,
			expectedError:  true,
		},
		{
			name:           "no installation ID",
			installationID: "",
			responseCode:   201,
			responseBody:   map[string]interface{}{"token": "unittest-access-token"},
			providePrivateKey: true,
			expectedToken:  "",
		},
		{
			name:           "no private key",
			installationID: "unittest-installation-id4",
			responseCode:   201,
			responseBody:   map[string]interface{}{"token": "unittest-access-token"},
			providePrivateKey: false,
			expectedToken:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "POST", r.Method)
				assert.Equal(t, "/app/installations/"+tt.installationID+"/access_tokens", r.URL.Path)
				assert.Equal(t, "2022-11-28", r.Header.Get("X-GitHub-Api-Version"))
				assert.Equal(t, "application/vnd.github+json", r.Header.Get("Accept"))
				// Check Authorization header contains Bearer token (JWT)
				authHeader := r.Header.Get("Authorization")
				if tt.installationID != "" {
					assert.True(t, len(authHeader) > 0 && authHeader[:7] == "Bearer ")
				}
				w.WriteHeader(tt.responseCode)
				json.NewEncoder(w).Encode(tt.responseBody)
			}))
			defer server.Close()

			privateKeyPath := ""
			if tt.providePrivateKey {
				privateKeyPath = "/tmp/test_github_private_key.pem"
			}

			expectedConfig := &model.ProviderSourceConfig{
				BaseURL:         server.URL,
				ApiURL:          server.URL,
				ClientID:        "test-client-id",
				ClientSecret:    "test-client-secret",
				LoginButtonText: "Test Login",
				PrivateKeyPath:  privateKeyPath,
				AppID:           "954956",
			}

			mockPSRepo := &MockProviderSourceRepository{
				findByNameFunc: func(ctx context.Context, name string) (*model.ProviderSource, error) {
					return model.NewProviderSource(
						name,
						"test-api-name",
						model.ProviderSourceTypeGithub,
						expectedConfig,
					), nil
				},
			}
			ghClass := service.NewGithubProviderSourceClass()

			gh := NewGithubProviderSource("test-name", mockPSRepo, ghClass)

			token, err := gh.GenerateAppInstallationToken(context.Background(), tt.installationID)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedToken, token)
			}
		})
	}
}

// TestGithubProviderSource_GenerateJWT tests the generateJWT method
// Python reference: test__generate_jwt
func TestGithubProviderSource_GenerateJWT(t *testing.T) {
	tests := []struct {
		name              string
		providePrivateKey bool
		expectError       bool
	}{
		{
			name:              "valid JWT generation",
			providePrivateKey: true,
			expectError:       false,
		},
		{
			name:              "no private key",
			providePrivateKey: false,
			expectError:       false, // Returns empty string, not error
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			privateKeyPath := ""
			if tt.providePrivateKey {
				privateKeyPath = "/tmp/test_github_private_key.pem"
			}

			expectedConfig := &model.ProviderSourceConfig{
				BaseURL:         "https://github.example-test.com",
				ApiURL:          "https://api.github.example-test.com",
				ClientID:        "test-client-id",
				ClientSecret:    "test-client-secret",
				LoginButtonText: "Test Login",
				PrivateKeyPath:  privateKeyPath,
				AppID:           "954956",
			}

			mockPSRepo := &MockProviderSourceRepository{
				findByNameFunc: func(ctx context.Context, name string) (*model.ProviderSource, error) {
					return model.NewProviderSource(
						name,
						"test-api-name",
						model.ProviderSourceTypeGithub,
						expectedConfig,
					), nil
				},
			}
			ghClass := service.NewGithubProviderSourceClass()

			gh := NewGithubProviderSource("test-name", mockPSRepo, ghClass)

			jwt, err := gh.generateJWT(context.Background())

			assert.NoError(t, err)

			if tt.providePrivateKey {
				assert.NotEmpty(t, jwt)
				// JWT should have three parts separated by dots
				parts := strings.Split(jwt, ".")
				assert.Len(t, parts, 3)
			} else {
				assert.Empty(t, jwt)
			}
		})
	}
}

// TestGithubProviderSource_GetAppMetadata tests the GetAppMetadata method
// Python reference: test__get_app_metadata
func TestGithubProviderSource_GetAppMetadata(t *testing.T) {
	tests := []struct {
		name              string
		responseCode      int
		responseBody      map[string]interface{}
		providePrivateKey bool
		expectedError     bool
		expectedID        string
		expectedName      string
	}{
		{
			name:         "valid metadata",
			responseCode: 200,
			responseBody: map[string]interface{}{
				"id":   "abcd-1234",
				"name": "test-org",
			},
			providePrivateKey: true,
			expectedError:    false,
			expectedID:       "abcd-1234",
			expectedName:     "test-org",
		},
		{
			name:         "invalid response code",
			responseCode: 403,
			responseBody: map[string]interface{}{
				"id":   "abcd-1234",
				"name": "test-org",
			},
			providePrivateKey: true,
			expectedError:     true,
		},
		{
			name:         "no private key",
			responseCode: 200,
			responseBody: map[string]interface{}{
				"id":   "abcd-1234",
				"name": "test-org",
			},
			providePrivateKey: false,
			expectedError:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "GET", r.Method)
				assert.Equal(t, "/app", r.URL.Path)
				assert.Equal(t, "2022-11-28", r.Header.Get("X-GitHub-Api-Version"))
				assert.Equal(t, "application/json", r.Header.Get("Accept"))
				w.WriteHeader(tt.responseCode)
				json.NewEncoder(w).Encode(tt.responseBody)
			}))
			defer server.Close()

			privateKeyPath := ""
			if tt.providePrivateKey {
				privateKeyPath = "/tmp/test_github_private_key.pem"
			}

			expectedConfig := &model.ProviderSourceConfig{
				BaseURL:         server.URL,
				ApiURL:          server.URL,
				ClientID:        "test-client-id",
				ClientSecret:    "test-client-secret",
				LoginButtonText: "Test Login",
				PrivateKeyPath:  privateKeyPath,
				AppID:           "954956",
			}

			mockPSRepo := &MockProviderSourceRepository{
				findByNameFunc: func(ctx context.Context, name string) (*model.ProviderSource, error) {
					return model.NewProviderSource(
						name,
						"test-api-name",
						model.ProviderSourceTypeGithub,
						expectedConfig,
					), nil
				},
			}
			ghClass := service.NewGithubProviderSourceClass()

			gh := NewGithubProviderSource("test-name", mockPSRepo, ghClass)

			metadata, err := gh.GetAppMetadata(context.Background())

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, metadata)
				assert.Equal(t, tt.expectedID, metadata["id"])
				assert.Equal(t, tt.expectedName, metadata["name"])
			}
		})
	}
}

// TestGithubProviderSource_GetAppInstallationURL tests the GetAppInstallationURL method
// Python reference: test_get_app_installation_url
func TestGithubProviderSource_GetAppInstallationURL(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/app", r.URL.Path)
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"html_url": "https://example.github.com/apps/my-special-app",
		})
	}))
	defer server.Close()

	expectedConfig := &model.ProviderSourceConfig{
		BaseURL:         server.URL,
		ApiURL:          server.URL,
		ClientID:        "test-client-id",
		ClientSecret:    "test-client-secret",
		LoginButtonText: "Test Login",
		PrivateKeyPath:  "/tmp/test_github_private_key.pem",
		AppID:           "954956",
	}

	mockPSRepo := &MockProviderSourceRepository{
		findByNameFunc: func(ctx context.Context, name string) (*model.ProviderSource, error) {
			return model.NewProviderSource(
				name,
				"test-api-name",
				model.ProviderSourceTypeGithub,
				expectedConfig,
			), nil
		},
	}
	ghClass := service.NewGithubProviderSourceClass()

	gh := NewGithubProviderSource("test-name", mockPSRepo, ghClass)

	url, err := gh.GetAppInstallationURL(context.Background())

	require.NoError(t, err)
	assert.Equal(t, "https://example.github.com/apps/my-special-app/installations/new", url)
}

// TestGithubProviderSource_GetGithubAppInstallationID tests the GetGithubAppInstallationID method
// Python reference: test_get_github_app_installation_id
func TestGithubProviderSource_GetGithubAppInstallationID(t *testing.T) {
	tests := []struct {
		name              string
		namespace         string
		orgResponseCode   int
		userResponseCode  int
		responseBody      map[string]interface{}
		providePrivateKey bool
		expectedID        string
		expectedError     bool
	}{
		{
			name:      "organization installation",
			namespace: "some-organisation",
			orgResponseCode: 200,
			userResponseCode: 404,
			responseBody: map[string]interface{}{
				"id": float64(12345), // Must be float64 for JSON unmarshaling to int
			},
			providePrivateKey: true,
			expectedID:        "12345",
		},
		{
			name:      "user installation (org fails, user succeeds)",
			namespace: "some-organisation",
			orgResponseCode: 404,
			userResponseCode: 200,
			responseBody: map[string]interface{}{
				"id": float64(67890),
			},
			providePrivateKey: true,
			expectedID:        "67890",
		},
		{
			name:      "404 not found",
			namespace: "some-organisation",
			orgResponseCode: 404,
			userResponseCode: 404,
			responseBody: map[string]interface{}{},
			providePrivateKey: true,
			expectedError:     true,
		},
		{
			name:      "500 server error",
			namespace: "some-organisation",
			orgResponseCode: 500,
			userResponseCode: 500,
			responseBody: map[string]interface{}{},
			providePrivateKey: true,
			expectedError:     true,
		},
		{
			name:      "invalid response data (id is nil)",
			namespace: "some-organisation",
			orgResponseCode: 200,
			userResponseCode: 200,
			responseBody: map[string]interface{}{
				"id": nil,
			},
			providePrivateKey: true,
			expectedError:     true,
		},
		{
			name:      "invalid response data (missing id)",
			namespace: "some-organisation",
			orgResponseCode: 200,
			userResponseCode: 200,
			responseBody: map[string]interface{}{},
			providePrivateKey: true,
			expectedError:     true,
		},
		{
			name:      "no private key",
			namespace: "some-organisation",
			orgResponseCode: 200,
			userResponseCode: 200,
			responseBody: map[string]interface{}{
				"id": float64(12345),
			},
			providePrivateKey: false,
			expectedError:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Try organization first
				if strings.HasPrefix(r.URL.Path, "/orgs/") {
					w.WriteHeader(tt.orgResponseCode)
					json.NewEncoder(w).Encode(tt.responseBody)
					return
				}
				// Try user lookup
				if strings.HasPrefix(r.URL.Path, "/users/") {
					w.WriteHeader(tt.userResponseCode)
					json.NewEncoder(w).Encode(tt.responseBody)
					return
				}
				w.WriteHeader(http.StatusNotFound)
				json.NewEncoder(w).Encode(map[string]interface{}{})
			}))
			defer server.Close()

			privateKeyPath := ""
			if tt.providePrivateKey {
				privateKeyPath = "/tmp/test_github_private_key.pem"
			}

			expectedConfig := &model.ProviderSourceConfig{
				BaseURL:         server.URL,
				ApiURL:          server.URL,
				ClientID:        "test-client-id",
				ClientSecret:    "test-client-secret",
				LoginButtonText: "Test Login",
				PrivateKeyPath:  privateKeyPath,
				AppID:           "954956",
			}

			mockPSRepo := &MockProviderSourceRepository{
				findByNameFunc: func(ctx context.Context, name string) (*model.ProviderSource, error) {
					return model.NewProviderSource(
						name,
						"test-api-name",
						model.ProviderSourceTypeGithub,
						expectedConfig,
					), nil
				},
			}
			ghClass := service.NewGithubProviderSourceClass()

			gh := NewGithubProviderSource("test-name", mockPSRepo, ghClass)

			id, err := gh.GetGithubAppInstallationID(context.Background(), tt.namespace)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedID, id)
			}
		})
	}
}

// TestGithubProviderSource_GetAccessTokenForProvider tests the GetAccessTokenForProvider method
// Python reference: test__get_access_token_for_provider
func TestGithubProviderSource_GetAccessTokenForProvider(t *testing.T) {
	tests := []struct {
		name                        string
		namespace                   string
		orgResponseCode             int
		userResponseCode            int
		providePrivateKey           bool
		defaultAccessToken          string
		expectedToken               string
	}{
		{
			name:              "installation token succeeds",
			namespace:         "test-namespace",
			orgResponseCode:   200,
			userResponseCode:  404,
			providePrivateKey: true,
			defaultAccessToken: "",
			expectedToken:     "unittest-installation-token",
		},
		{
			name:              "fallback to default access token",
			namespace:         "test-namespace",
			orgResponseCode:   404,
			userResponseCode:  404,
			providePrivateKey: false,
			defaultAccessToken: "default-auth-access-token",
			expectedToken:     "default-auth-access-token",
		},
		{
			name:              "no installation and no default token",
			namespace:         "test-namespace",
			orgResponseCode:   404,
			userResponseCode:  404,
			providePrivateKey: false,
			defaultAccessToken: "",
			expectedToken:     "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if strings.HasPrefix(r.URL.Path, "/orgs/") {
					w.WriteHeader(tt.orgResponseCode)
					if tt.orgResponseCode == 200 {
						json.NewEncoder(w).Encode(map[string]interface{}{"id": float64(12345)})
					} else {
						json.NewEncoder(w).Encode(map[string]interface{}{})
					}
					return
				}
				if strings.HasPrefix(r.URL.Path, "/users/") {
					w.WriteHeader(tt.userResponseCode)
					if tt.userResponseCode == 200 {
						json.NewEncoder(w).Encode(map[string]interface{}{"id": float64(12345)})
					} else {
						json.NewEncoder(w).Encode(map[string]interface{}{})
					}
					return
				}
				// Installation token endpoint
				w.WriteHeader(201)
				json.NewEncoder(w).Encode(map[string]interface{}{"token": "unittest-installation-token"})
			}))
			defer server.Close()

			privateKeyPath := ""
			if tt.providePrivateKey {
				privateKeyPath = "/tmp/test_github_private_key.pem"
			}

			expectedConfig := &model.ProviderSourceConfig{
				BaseURL:             server.URL,
				ApiURL:              server.URL,
				ClientID:            "test-client-id",
				ClientSecret:        "test-client-secret",
				LoginButtonText:     "Test Login",
				PrivateKeyPath:      privateKeyPath,
				AppID:               "954956",
				DefaultAccessToken:  tt.defaultAccessToken,
			}

			mockPSRepo := &MockProviderSourceRepository{
				findByNameFunc: func(ctx context.Context, name string) (*model.ProviderSource, error) {
					return model.NewProviderSource(
						name,
						"test-api-name",
						model.ProviderSourceTypeGithub,
						expectedConfig,
					), nil
				},
			}
			ghClass := service.NewGithubProviderSourceClass()

			gh := NewGithubProviderSource("test-name", mockPSRepo, ghClass)

			token, err := gh.GetAccessTokenForProvider(context.Background(), tt.namespace)

			assert.NoError(t, err)
			assert.Equal(t, tt.expectedToken, token)
		})
	}
}

// TestGithubProviderSource_IsEntityOrgOrUser tests the IsEntityOrgOrUser method
// Python reference: test__is_entity_org_or_user
func TestGithubProviderSource_IsEntityOrgOrUser(t *testing.T) {
	tests := []struct {
		name         string
		identity     string
		responseCode int
		responseBody map[string]interface{}
		expectedType string
	}{
		{
			name:     "user type",
			identity: "unit-test-identity-name",
			responseCode: 200,
			responseBody: map[string]interface{}{
				"type": "User",
			},
			expectedType: "GITHUB_USER",
		},
		{
			name:     "organization type",
			identity: "unit-test-identity-name",
			responseCode: 200,
			responseBody: map[string]interface{}{
				"type": "Organization",
			},
			expectedType: "GITHUB_ORGANISATION",
		},
		{
			name:     "invalid response code",
			identity: "unit-test-identity-name",
			responseCode: 404,
			responseBody: map[string]interface{}{
				"type": "Organization",
			},
			expectedType: "",
		},
		{
			name:     "invalid type",
			identity: "unit-test-identity-name",
			responseCode: 200,
			responseBody: map[string]interface{}{
				"type": "SomethingElse",
			},
			expectedType: "",
		},
		{
			name:     "type is nil",
			identity: "unit-test-identity-name",
			responseCode: 200,
			responseBody: map[string]interface{}{
				"type": nil,
			},
			expectedType: "",
		},
		{
			name:         "missing type field",
			identity:     "unit-test-identity-name",
			responseCode: 200,
			responseBody: map[string]interface{}{},
			expectedType: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "GET", r.Method)
				assert.Equal(t, "/users/"+tt.identity, r.URL.Path)
				assert.Equal(t, "2022-11-28", r.Header.Get("X-GitHub-Api-Version"))
				assert.Equal(t, "application/json", r.Header.Get("Accept"))
				assert.Equal(t, "Bearer unittest-access-token", r.Header.Get("Authorization"))
				w.WriteHeader(tt.responseCode)
				json.NewEncoder(w).Encode(tt.responseBody)
			}))
			defer server.Close()

			expectedConfig := &model.ProviderSourceConfig{
				BaseURL:         server.URL,
				ApiURL:          server.URL,
				ClientID:        "test-client-id",
				ClientSecret:    "test-client-secret",
				LoginButtonText: "Test Login",
				PrivateKeyPath:  "",
				AppID:           "954956",
			}

			mockPSRepo := &MockProviderSourceRepository{
				findByNameFunc: func(ctx context.Context, name string) (*model.ProviderSource, error) {
					return model.NewProviderSource(
						name,
						"test-api-name",
						model.ProviderSourceTypeGithub,
						expectedConfig,
					), nil
				},
			}
			ghClass := service.NewGithubProviderSourceClass()

			gh := NewGithubProviderSource("test-name", mockPSRepo, ghClass)

			entityType, err := gh.IsEntityOrgOrUser(context.Background(), tt.identity, "unittest-access-token")

			assert.NoError(t, err)
			assert.Equal(t, tt.expectedType, entityType)
		})
	}
}
