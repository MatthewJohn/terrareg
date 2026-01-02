package provider_source

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	provider_source_model "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/provider_source/model"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/provider_source/service"
	repository_model "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/repository/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestGetCommitHashByRelease tests the GetCommitHashByRelease method
func TestGetCommitHashByRelease(t *testing.T) {
	tests := []struct {
		name           string
		responseCode   int
		responseBody   map[string]interface{}
		expectedHash   string
		expectError    bool
	}{
		{
			name:         "valid commit hash",
			responseCode: 200,
			responseBody: map[string]interface{}{
				"object": map[string]interface{}{
					"sha": "abc123def456",
				},
			},
			expectedHash: "abc123def456",
		},
		{
			name:         "404 response",
			responseCode: 404,
			responseBody: map[string]interface{}{},
			expectedHash: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))
				assert.Equal(t, "2022-11-28", r.Header.Get("X-GitHub-Api-Version"))
				assert.Equal(t, "application/vnd.github+json", r.Header.Get("Accept"))
				w.WriteHeader(tt.responseCode)
				json.NewEncoder(w).Encode(tt.responseBody)
			}))
			defer server.Close()

			expectedConfig := &provider_source_model.ProviderSourceConfig{
				BaseURL:         server.URL,
				ApiURL:          server.URL,
				ClientID:        "test-client-id",
				ClientSecret:    "test-client-secret",
				LoginButtonText: "Test Login",
				PrivateKeyPath:  "",
				AppID:           "123",
			}

			mockPSRepo := &MockProviderSourceRepository{
				findByNameFunc: func(ctx context.Context, name string) (*provider_source_model.ProviderSource, error) {
					return provider_source_model.NewProviderSource(
						name,
						"test-api-name",
						provider_source_model.ProviderSourceTypeGithub,
						expectedConfig,
					), nil
				},
			}
			ghClass := service.NewGithubProviderSourceClass()

			gh := NewGithubProviderSource("test-name", mockPSRepo, ghClass)

			repo := repository_model.ReconstructRepository(
				1,
				nil,
				"test-owner",
				"test-repo",
				nil,
				nil,
				nil,
				"test-name",
			)

			hash, err := gh.GetCommitHashByRelease(context.Background(), repo, "v1.0.0", "test-token")

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedHash, hash)
			}
		})
	}
}

// TestGetReleaseArtifactsMetadata tests the GetReleaseArtifactsMetadata method
func TestGetReleaseArtifactsMetadata(t *testing.T) {
	tests := []struct {
		name              string
		responseCode      int
		responseBody      []map[string]interface{}
		expectedArtifacts int
	}{
		{
			name:         "multiple artifacts",
			responseCode: 200,
			responseBody: []map[string]interface{}{
				{"id": float64(123), "name": "artifact1.zip"},
				{"id": float64(456), "name": "artifact2.zip"},
			},
			expectedArtifacts: 2,
		},
		{
			name:              "non-200 response",
			responseCode:      404,
			responseBody:      []map[string]interface{}{},
			expectedArtifacts: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.responseCode)
				json.NewEncoder(w).Encode(tt.responseBody)
			}))
			defer server.Close()

			expectedConfig := &provider_source_model.ProviderSourceConfig{
				BaseURL:         server.URL,
				ApiURL:          server.URL,
				ClientID:        "test-client-id",
				ClientSecret:    "test-client-secret",
				LoginButtonText: "Test Login",
				PrivateKeyPath:  "",
				AppID:           "123",
			}

			mockPSRepo := &MockProviderSourceRepository{
				findByNameFunc: func(ctx context.Context, name string) (*provider_source_model.ProviderSource, error) {
					return provider_source_model.NewProviderSource(
						name,
						"test-api-name",
						provider_source_model.ProviderSourceTypeGithub,
						expectedConfig,
					), nil
				},
			}
			ghClass := service.NewGithubProviderSourceClass()

			gh := NewGithubProviderSource("test-name", mockPSRepo, ghClass)

			repo := repository_model.ReconstructRepository(
				1,
				nil,
				"test-owner",
				"test-repo",
				nil,
				nil,
				nil,
				"test-name",
			)

			artifacts, err := gh.GetReleaseArtifactsMetadata(context.Background(), repo, 12345, "test-token")

			assert.NoError(t, err)
			assert.Len(t, artifacts, tt.expectedArtifacts)
		})
	}
}

// TestGetReleaseArtifact tests the GetReleaseArtifact method
func TestGetReleaseArtifact(t *testing.T) {
	expectedContent := []byte("test artifact content")

	tests := []struct {
		name           string
		responseCode   int
		responseBody   []byte
		expectedResult []byte
	}{
		{
			name:           "successful download",
			responseCode:   200,
			responseBody:   expectedContent,
			expectedResult: expectedContent,
		},
		{
			name:           "404 not found",
			responseCode:   404,
			responseBody:   []byte("not found"),
			expectedResult: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.responseCode)
				w.Write(tt.responseBody)
			}))
			defer server.Close()

			expectedConfig := &provider_source_model.ProviderSourceConfig{
				BaseURL:         server.URL,
				ApiURL:          server.URL,
				ClientID:        "test-client-id",
				ClientSecret:    "test-client-secret",
				LoginButtonText: "Test Login",
				PrivateKeyPath:  "",
				AppID:           "123",
			}

			mockPSRepo := &MockProviderSourceRepository{
				findByNameFunc: func(ctx context.Context, name string) (*provider_source_model.ProviderSource, error) {
					return provider_source_model.NewProviderSource(
						name,
						"test-api-name",
						provider_source_model.ProviderSourceTypeGithub,
						expectedConfig,
					), nil
				},
			}
			ghClass := service.NewGithubProviderSourceClass()

			gh := NewGithubProviderSource("test-name", mockPSRepo, ghClass)

			repo := repository_model.ReconstructRepository(
				1,
				nil,
				"test-owner",
				"test-repo",
				nil,
				nil,
				nil,
				"test-name",
			)

			artifact := provider_source_model.NewReleaseArtifactMetadata("test-artifact.zip", 123)

			result, err := gh.GetReleaseArtifact(context.Background(), repo, artifact, "test-token")

			assert.NoError(t, err)
			assert.Equal(t, tt.expectedResult, result)
		})
	}
}

// TestGetReleaseArchive tests the GetReleaseArchive method
func TestGetReleaseArchive(t *testing.T) {
	expectedContent := []byte("test archive content")

	tests := []struct {
		name            string
		responseCode    int
		responseBody    []byte
		expectedContent []byte
	}{
		{
			name:            "successful download",
			responseCode:    200,
			responseBody:    expectedContent,
			expectedContent: expectedContent,
		},
		{
			name:            "404 not found",
			responseCode:    404,
			responseBody:    []byte("not found"),
			expectedContent: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.responseCode)
				w.Write(tt.responseBody)
			}))
			defer server.Close()

			expectedConfig := &provider_source_model.ProviderSourceConfig{
				BaseURL:         server.URL,
				ApiURL:          server.URL,
				ClientID:        "test-client-id",
				ClientSecret:    "test-client-secret",
				LoginButtonText: "Test Login",
				PrivateKeyPath:  "",
				AppID:           "123",
			}

			mockPSRepo := &MockProviderSourceRepository{
				findByNameFunc: func(ctx context.Context, name string) (*provider_source_model.ProviderSource, error) {
					return provider_source_model.NewProviderSource(
						name,
						"test-api-name",
						provider_source_model.ProviderSourceTypeGithub,
						expectedConfig,
					), nil
				},
			}
			ghClass := service.NewGithubProviderSourceClass()

			gh := NewGithubProviderSource("test-name", mockPSRepo, ghClass)

			repo := repository_model.ReconstructRepository(
				1,
				nil,
				"test-owner",
				"test-repo",
				nil,
				nil,
				nil,
				"test-name",
			)

			releaseMetadata := provider_source_model.NewRepositoryReleaseMetadata(
				"Test Release",
				"v1.0.0",
				"https://example.com/archive.tar.gz",
				"abcdef1234567",
				123,
				*repo,
				[]*provider_source_model.ReleaseArtifactMetadata{},
			)

			content, archiveID, err := gh.GetReleaseArchive(context.Background(), repo, releaseMetadata, "test-token")

			assert.NoError(t, err)
			assert.Equal(t, "test-owner-test-repo-abcdef1", archiveID)
			assert.Equal(t, tt.expectedContent, content)
		})
	}
}

// TestAddRepository tests the AddRepository method
func TestAddRepository(t *testing.T) {
	tests := []struct {
		name                string
		repositoryMetadata  map[string]interface{}
		expectCreate        bool
		expectedProviderID  string
		expectedOwner       string
		expectedName        string
	}{
		{
			name: "valid repository",
			repositoryMetadata: map[string]interface{}{
				"id":          float64(12345),
				"name":        "terraform-provider-test",
				"owner": map[string]interface{}{
					"login": "test-owner",
				},
			},
			expectCreate:       true,
			expectedProviderID: "12345",
			expectedOwner:      "test-owner",
			expectedName:        "terraform-provider-test",
		},
		{
			name: "missing id",
			repositoryMetadata: map[string]interface{}{
				"name": "test-repo",
				"owner": map[string]interface{}{
					"login": "test-owner",
				},
			},
			expectCreate: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockPSRepo := &MockProviderSourceRepository{
				findByNameFunc: func(ctx context.Context, name string) (*provider_source_model.ProviderSource, error) {
					return provider_source_model.NewProviderSource(
						name,
						"test-api-name",
						provider_source_model.ProviderSourceTypeGithub,
						&provider_source_model.ProviderSourceConfig{},
					), nil
				},
			}
			ghClass := service.NewGithubProviderSourceClass()

			gh := NewGithubProviderSource("test-source-name", mockPSRepo, ghClass)

			var createdRepo *repository_model.Repository
			mockRepoRepo := &MockRepositoryRepository{
				existsFunc: func(ctx context.Context, providerSourceName string, providerID string) (bool, error) {
					return false, nil
				},
				createFunc: func(ctx context.Context, repository *repository_model.Repository) (*repository_model.Repository, error) {
					createdRepo = repository
					return repository, nil
				},
			}

			err := gh.AddRepository(context.Background(), mockRepoRepo, tt.repositoryMetadata)

			assert.NoError(t, err)

			if tt.expectCreate {
				require.NotNil(t, createdRepo)
				assert.Equal(t, tt.expectedProviderID, *createdRepo.ProviderID())
				assert.Equal(t, tt.expectedOwner, createdRepo.Owner())
				assert.Equal(t, tt.expectedName, createdRepo.Name())
			}
		})
	}
}

// MockRepositoryRepository is a mock for RepositoryRepository
type MockRepositoryRepository struct {
	existsFunc func(ctx context.Context, providerSourceName string, providerID string) (bool, error)
	createFunc func(ctx context.Context, repository *repository_model.Repository) (*repository_model.Repository, error)
}

func (m *MockRepositoryRepository) FindByID(ctx context.Context, id int) (*repository_model.Repository, error) {
	return nil, nil
}

func (m *MockRepositoryRepository) FindByProviderSourceAndProviderID(ctx context.Context, providerSourceName string, providerID string) (*repository_model.Repository, error) {
	return nil, nil
}

func (m *MockRepositoryRepository) FindByOwnerList(ctx context.Context, owners []string) ([]*repository_model.Repository, error) {
	return nil, nil
}

func (m *MockRepositoryRepository) Create(ctx context.Context, repository *repository_model.Repository) (*repository_model.Repository, error) {
	if m.createFunc != nil {
		return m.createFunc(ctx, repository)
	}
	return repository, nil
}

func (m *MockRepositoryRepository) Update(ctx context.Context, repository *repository_model.Repository) error {
	return nil
}

func (m *MockRepositoryRepository) Delete(ctx context.Context, id int) error {
	return nil
}

func (m *MockRepositoryRepository) Exists(ctx context.Context, providerSourceName string, providerID string) (bool, error) {
	if m.existsFunc != nil {
		return m.existsFunc(ctx, providerSourceName, providerID)
	}
	return false, nil
}

// TestUpdateRepositories tests the UpdateRepositories method
func TestUpdateRepositories(t *testing.T) {
	tests := []struct {
		name                string
		resultCount         int
		expectedPages       int
		expectedAddRepoCall int
	}{
		{"no results", 0, 1, 0},
		{"one result", 1, 1, 1},
		{"100 results", 100, 2, 100},
		{"101 results", 101, 2, 101},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			currentPage := 0
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				startIdx := currentPage * 100
				endIdx := min(tt.resultCount, startIdx+100)
				count := max(0, endIdx-startIdx)

				var results []map[string]interface{}
				for i := 0; i < count; i++ {
					repoID := startIdx + i
					results = append(results, map[string]interface{}{
						"id":   float64(repoID),
						"name": "terraform-provider-test",
						"owner": map[string]interface{}{
							"login": "test-owner",
						},
						"clone_url": "https://github.com/test-owner/terraform-provider-test.git",
					})
				}

				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(results)
				currentPage++
			}))
			defer server.Close()

			expectedConfig := &provider_source_model.ProviderSourceConfig{
				BaseURL:         server.URL,
				ApiURL:          server.URL,
				ClientID:        "test-client-id",
				ClientSecret:    "test-client-secret",
				LoginButtonText: "Test Login",
				PrivateKeyPath:  "",
				AppID:           "123",
			}

			mockPSRepo := &MockProviderSourceRepository{
				findByNameFunc: func(ctx context.Context, name string) (*provider_source_model.ProviderSource, error) {
					return provider_source_model.NewProviderSource(
						name,
						"test-api-name",
						provider_source_model.ProviderSourceTypeGithub,
						expectedConfig,
					), nil
				},
			}
			ghClass := service.NewGithubProviderSourceClass()

			gh := NewGithubProviderSource("test-name", mockPSRepo, ghClass)

			addRepoCallCount := 0
			mockRepoRepo := &MockRepositoryRepository{
				existsFunc: func(ctx context.Context, providerSourceName string, providerID string) (bool, error) {
					return false, nil
				},
				createFunc: func(ctx context.Context, repository *repository_model.Repository) (*repository_model.Repository, error) {
					addRepoCallCount++
					return repository, nil
				},
			}

			err := gh.UpdateRepositories(context.Background(), mockRepoRepo, "test-token")

			assert.NoError(t, err)
			assert.Equal(t, tt.expectedAddRepoCall, addRepoCallCount)
		})
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
