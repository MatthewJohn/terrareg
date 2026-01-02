package provider_source

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	provider_source_model "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/provider_source/model"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/provider_source/service"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/persistence/sqldb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
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

			owner := "test-owner"
			name := "test-repo"
			repo := &sqldb.RepositoryDB{
				Owner:              &owner,
				Name:               &name,
				ProviderSourceName: "test-name",
			}

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

			owner := "test-owner"
			name := "test-repo"
			repo := &sqldb.RepositoryDB{
				Owner:              &owner,
				Name:               &name,
				ProviderSourceName: "test-name",
			}

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

			owner := "test-owner"
			name := "test-repo"
			repo := &sqldb.RepositoryDB{
				Owner:              &owner,
				Name:               &name,
				ProviderSourceName: "test-name",
			}

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

			owner := "test-owner"
			name := "test-repo"
			cloneURL := "https://github.com/test-owner/test-repo.git"
			repo := &sqldb.RepositoryDB{
				Owner:              &owner,
				Name:               &name,
				CloneURL:           &cloneURL,
				ProviderSourceName: "test-name",
			}

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
		name               string
		repositoryMetadata map[string]interface{}
		expectCreate       bool
		expectedProviderID string
		expectedOwner      string
		expectedName       string
	}{
		{
			name: "valid repository",
			repositoryMetadata: map[string]interface{}{
				"id":          float64(12345),
				"name":        "terraform-provider-test",
				"owner": map[string]interface{}{
					"login": "test-owner",
				},
				"clone_url": "https://github.com/test-owner/terraform-provider-test.git",
			},
			expectCreate:       true,
			expectedProviderID: "12345",
			expectedOwner:      "test-owner",
			expectedName:       "terraform-provider-test",
		},
		{
			name: "missing id",
			repositoryMetadata: map[string]interface{}{
				"name": "test-repo",
				"owner": map[string]interface{}{
					"login": "test-owner",
				},
				"clone_url": "https://github.com/test-owner/test-repo.git",
			},
			expectCreate: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up in-memory database
			db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
			require.NoError(t, err)
			err = db.AutoMigrate(&sqldb.RepositoryDB{})
			require.NoError(t, err)

			database := &sqldb.Database{DB: db}

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

			err = gh.AddRepository(context.Background(), database, tt.repositoryMetadata)

			assert.NoError(t, err)

			if tt.expectCreate {
				// Verify repository was created
				var repos []sqldb.RepositoryDB
				db.Find(&repos)
				require.Len(t, repos, 1)
				assert.Equal(t, tt.expectedProviderID, *repos[0].ProviderID)
				assert.Equal(t, tt.expectedOwner, *repos[0].Owner)
				assert.Equal(t, tt.expectedName, *repos[0].Name)
			}
		})
	}
}

// TestUpdateRepositories tests the UpdateRepositories method
func TestUpdateRepositories(t *testing.T) {
	tests := []struct {
		name                string
		resultCount         int
		expectedPages       int
		expectedAddRepoCall int
	}{
		{"no_results", 0, 1, 0},
		{"one_result", 1, 1, 1},
		{"100_results", 100, 2, 100},
		{"101_results", 101, 2, 101},
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
						"id":        float64(repoID),
						"name":      "terraform-provider-test",
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

			// Set up unique in-memory database for this test
			db, err := gorm.Open(sqlite.Open("file::memory:"), &gorm.Config{})
			require.NoError(t, err)
			err = db.AutoMigrate(&sqldb.RepositoryDB{})
			require.NoError(t, err)

			database := &sqldb.Database{DB: db}

			err = gh.UpdateRepositories(context.Background(), database, "test-token")

			assert.NoError(t, err)

			// Verify repository count
			var count int64
			db.Model(&sqldb.RepositoryDB{}).Count(&count)
			assert.Equal(t, int64(tt.expectedAddRepoCall), count)
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
