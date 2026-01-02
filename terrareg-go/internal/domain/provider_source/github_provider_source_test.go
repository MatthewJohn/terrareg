package provider_source

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
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
