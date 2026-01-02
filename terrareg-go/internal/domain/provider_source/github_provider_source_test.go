package provider_source

import (
	"context"
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
