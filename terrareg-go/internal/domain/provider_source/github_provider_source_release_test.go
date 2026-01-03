package provider_source

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
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

// TestUpdateRepositoriesInvalidResponse tests the UpdateRepositories method with invalid response
// Python reference: test_update_repositories_invalid_response
func TestUpdateRepositoriesInvalidResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Return invalid response code
		w.WriteHeader(http.StatusInternalServerError)
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

	// Set up in-memory database
	db, err := gorm.Open(sqlite.Open("file::memory:"), &gorm.Config{})
	require.NoError(t, err)
	err = db.AutoMigrate(&sqldb.RepositoryDB{})
	require.NoError(t, err)

	database := &sqldb.Database{DB: db}

	// Call UpdateRepositories - should return error without adding repositories
	err = gh.UpdateRepositories(context.Background(), database, "test-token")

	// Should get an error about invalid response code
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid response code")

	// Verify no repositories were added
	var count int64
	db.Model(&sqldb.RepositoryDB{}).Count(&count)
	assert.Equal(t, int64(0), count)
}

// TestProcessRelease tests the ProcessRelease method
// Python reference: test__process_release
func TestProcessRelease(t *testing.T) {
	tests := []struct {
		name                      string
		githubReleaseMetadata     map[string]interface{}
		commitHashResponse        map[string]interface{}
		commitHashStatusCode      int
		artifactsResponse         []map[string]interface{}
		expectNilResult           bool
		expectedReleaseName       string
		expectedTag               string
		expectedArchiveURL        string
		expectedCommitHash        string
	}{
		{
			name: "valid release with semver tag",
			githubReleaseMetadata: map[string]interface{}{
				"id":         float64(12345),
				"name":       "Test Release v1.0.0",
				"tag_name":   "v1.0.0",
				"tarball_url": "https://example.com/archive.tar.gz",
			},
			commitHashResponse: map[string]interface{}{
				"object": map[string]interface{}{
					"sha": "abc123def456",
				},
			},
			commitHashStatusCode: 200,
			artifactsResponse: []map[string]interface{}{
				{"id": float64(1), "name": "artifact1.zip"},
				{"id": float64(2), "name": "artifact2.zip"},
			},
			expectNilResult:     false,
			expectedReleaseName: "Test Release v1.0.0",
			expectedTag:         "v1.0.0",
			expectedArchiveURL:  "https://example.com/archive.tar.gz",
			expectedCommitHash:  "abc123def456",
		},
		{
			name: "release missing id field",
			githubReleaseMetadata: map[string]interface{}{
				"name":       "Test Release",
				"tag_name":   "v1.0.0",
				"tarball_url": "https://example.com/archive.tar.gz",
			},
			expectNilResult: true,
		},
		{
			name: "release missing name field",
			githubReleaseMetadata: map[string]interface{}{
				"id":         float64(12345),
				"tag_name":   "v1.0.0",
				"tarball_url": "https://example.com/archive.tar.gz",
			},
			expectNilResult: true,
		},
		{
			name: "release with non-semver tag",
			githubReleaseMetadata: map[string]interface{}{
				"id":         float64(12345),
				"name":       "Test Release",
				"tag_name":   "not-a-version",
				"tarball_url": "https://example.com/archive.tar.gz",
			},
			commitHashResponse: map[string]interface{}{
				"object": map[string]interface{}{
					"sha": "abc123def456",
				},
			},
			commitHashStatusCode: 200,
			artifactsResponse:    []map[string]interface{}{},
			expectNilResult:      true,
		},
		{
			name: "release with empty commit hash",
			githubReleaseMetadata: map[string]interface{}{
				"id":         float64(12345),
				"name":       "Test Release",
				"tag_name":   "v1.0.0",
				"tarball_url": "https://example.com/archive.tar.gz",
			},
			commitHashResponse:   map[string]interface{}{},
			commitHashStatusCode: 404,
			expectNilResult:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create unified test server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Handle different request types based on the URL path
				if strings.Contains(r.URL.Path, "/git/ref/tags/") {
					// Handle GetCommitHashByRelease requests
					w.WriteHeader(tt.commitHashStatusCode)
					json.NewEncoder(w).Encode(tt.commitHashResponse)
				} else if strings.Contains(r.URL.Path, "/releases/") && strings.HasSuffix(r.URL.Path, "/assets") {
					// Handle GetReleaseArtifactsMetadata requests
					w.WriteHeader(http.StatusOK)
					json.NewEncoder(w).Encode(tt.artifactsResponse)
				} else {
					w.WriteHeader(http.StatusNotFound)
				}
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

			result, err := gh.ProcessRelease(context.Background(), repo, tt.githubReleaseMetadata, "test-token")

			assert.NoError(t, err)

			if tt.expectNilResult {
				assert.Nil(t, result)
			} else {
				require.NotNil(t, result)
				assert.Equal(t, tt.expectedReleaseName, result.Name)
				assert.Equal(t, tt.expectedTag, result.Tag)
				assert.Equal(t, tt.expectedArchiveURL, result.ArchiveURL)
				assert.Equal(t, tt.expectedCommitHash, result.CommitHash)
				assert.Len(t, result.ReleaseArtifacts, len(tt.artifactsResponse))
			}
		})
	}
}

// TestGetNewReleases tests the GetNewReleases method
// Python reference: test_get_new_releases
func TestGetNewReleases(t *testing.T) {
	tests := []struct {
		name             string
		resultCount      int
		expectedPages    int
		expectedReleases int
	}{
		{
			name:             "no results",
			resultCount:      0,
			expectedPages:    1,
			expectedReleases: 0,
		},
		{
			name:             "one result",
			resultCount:      1,
			expectedPages:    1,
			expectedReleases: 1,
		},
		{
			name:             "100 results - boundary for pagination",
			resultCount:      100,
			expectedPages:    2,
			expectedReleases: 100,
		},
		{
			name:             "101 results - triggers second page",
			resultCount:      101,
			expectedPages:    2,
			expectedReleases: 101,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			currentPage := 0
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Handle different request types based on the URL path
				// Check order matters - more specific patterns first
				if strings.Contains(r.URL.Path, "/releases/") && strings.HasSuffix(r.URL.Path, "/assets") {
					// Handle GetReleaseArtifactsMetadata requests
					w.WriteHeader(http.StatusOK)
					json.NewEncoder(w).Encode([]map[string]interface{}{})
				} else if strings.Contains(r.URL.Path, "/git/ref/tags/") {
					// Handle GetCommitHashByRelease requests
					w.WriteHeader(http.StatusOK)
					json.NewEncoder(w).Encode(map[string]interface{}{
						"object": map[string]interface{}{
							"sha": "abc123def456",
						},
					})
				} else if strings.Contains(r.URL.Path, "/releases") {
					// Handle GetNewReleases requests
					startIdx := currentPage * 100
					endIdx := min(tt.resultCount, startIdx+100)
					count := max(0, endIdx-startIdx)

					var results []map[string]interface{}
					for i := 0; i < count; i++ {
						releaseNum := startIdx + i
						results = append(results, map[string]interface{}{
							"id":          float64(releaseNum),
							"name":        "Release",
							"tag_name":    fmt.Sprintf("v1.0.%d", releaseNum),
							"tarball_url": fmt.Sprintf("https://example.com/archive%d.tar.gz", releaseNum),
						})
					}

					w.WriteHeader(http.StatusOK)
					json.NewEncoder(w).Encode(results)
					currentPage++
				} else {
					w.WriteHeader(http.StatusNotFound)
				}
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

			releases, err := gh.GetNewReleases(context.Background(), repo, "test-token")

			assert.NoError(t, err)
			assert.Len(t, releases, tt.expectedReleases)
		})
	}
}

// TestGetNewReleasesSkipInvalidReleases tests GetNewReleases with invalid releases
// Python reference: test_get_new_releases_skip_invalid_releases
func TestGetNewReleasesSkipInvalidReleases(t *testing.T) {
	currentPage := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Handle different request types based on the URL path
		if strings.Contains(r.URL.Path, "/releases") && !strings.Contains(r.URL.Path, "/releases/") {
			// Handle GetNewReleases requests
			if currentPage == 0 {
				results := []map[string]interface{}{
					// Invalid - missing id
					{"name": "Invalid Release 1", "tag_name": "v1.0.0"},
					// Valid
					{"id": float64(1), "name": "Valid Release 1", "tag_name": "v1.0.1", "tarball_url": "https://example.com/archive1.tar.gz"},
					// Invalid - non-semver tag
					{"id": float64(2), "name": "Invalid Release 2", "tag_name": "not-a-version", "tarball_url": "https://example.com/archive2.tar.gz"},
					// Valid
					{"id": float64(3), "name": "Valid Release 2", "tag_name": "v1.0.2", "tarball_url": "https://example.com/archive3.tar.gz"},
				}
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(results)
			} else {
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode([]map[string]interface{}{})
			}
			currentPage++
		} else if strings.Contains(r.URL.Path, "/git/ref/tags/") {
			// Handle GetCommitHashByRelease requests
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"object": map[string]interface{}{
					"sha": "abc123def456",
				},
			})
		} else if strings.Contains(r.URL.Path, "/releases/") && strings.HasSuffix(r.URL.Path, "/assets") {
			// Handle GetReleaseArtifactsMetadata requests
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode([]map[string]interface{}{})
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
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

	releases, err := gh.GetNewReleases(context.Background(), repo, "test-token")

	assert.NoError(t, err)
	// Should only return the 2 valid releases (invalid ones are skipped)
	assert.Len(t, releases, 2)
}

// TestGetNewReleasesInvalidResponse tests GetNewReleases with invalid API response
// Python reference: test_get_new_releases_invalid_response
func TestGetNewReleasesInvalidResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Return invalid response code
		w.WriteHeader(http.StatusInternalServerError)
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

	releases, err := gh.GetNewReleases(context.Background(), repo, "test-token")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid response code")
	assert.Nil(t, releases)
}