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

// TestRefreshNamespaceRepositories tests the RefreshNamespaceRepositories method
// Python reference: test_refresh_namespace_repositories
func TestRefreshNamespaceRepositories(t *testing.T) {
	tests := []struct {
		name              string
		namespaceType     string
		resultCount       int
		expectedAddRepos  int
		expectedEndpoint  string
	}{
		{
			name:             "organisation with no results",
			namespaceType:    "GITHUB_ORGANISATION",
			resultCount:      0,
			expectedAddRepos: 0,
			expectedEndpoint: "/orgs/test-namespace/repos",
		},
		{
			name:             "user with one result",
			namespaceType:    "GITHUB_USER",
			resultCount:      1,
			expectedAddRepos: 1,
			expectedEndpoint: "/users/test-namespace/repos",
		},
		{
			name:             "organisation with 100 results",
			namespaceType:    "GITHUB_ORGANISATION",
			resultCount:      100,
			expectedAddRepos: 100,
			expectedEndpoint: "/orgs/test-namespace/repos",
		},
		{
			name:             "user with 101 results - triggers second page",
			namespaceType:    "GITHUB_USER",
			resultCount:      101,
			expectedAddRepos: 101,
			expectedEndpoint: "/users/test-namespace/repos",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			currentPage := 0
			entityTypeChecked := false
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if strings.Contains(r.URL.Path, "/users/") && !strings.Contains(r.URL.Path, "/repos") {
					// IsEntityOrgOrUser check - /users/{identity} endpoint
					entityTypeChecked = true
					w.WriteHeader(http.StatusOK)
					// Map GITHUB_USER -> "User" and GITHUB_ORGANISATION -> "Organization"
					var githubType string
					if tt.namespaceType == "GITHUB_USER" {
						githubType = "User"
					} else if tt.namespaceType == "GITHUB_ORGANISATION" {
						githubType = "Organization"
					}
					json.NewEncoder(w).Encode(map[string]interface{}{
						"type": githubType,
					})
				} else if strings.Contains(r.URL.Path, "/repos") {
					// Handle repository listing
					startIdx := currentPage * 100
					endIdx := min(tt.resultCount, startIdx+100)
					count := max(0, endIdx-startIdx)

					var results []map[string]interface{}
					for i := 0; i < count; i++ {
						repoID := startIdx + i
						results = append(results, map[string]interface{}{
							"id":        float64(repoID),
							"name":      fmt.Sprintf("terraform-provider-test-%d", repoID),
							"owner": map[string]interface{}{
								"login": "test-owner",
							},
							"clone_url": fmt.Sprintf("https://github.com/test-owner/terraform-provider-test-%d.git", repoID),
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
				BaseURL:             server.URL,
				ApiURL:              server.URL,
				ClientID:            "test-client-id",
				ClientSecret:        "test-client-secret",
				LoginButtonText:     "Test Login",
				PrivateKeyPath:      "",
				AppID:               "123",
				DefaultAccessToken:  "test-default-token",
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

			err = gh.RefreshNamespaceRepositories(context.Background(), database, "test-namespace")

			assert.NoError(t, err)
			assert.True(t, entityTypeChecked, "Entity type should have been checked")

			// Verify repository count
			var count int64
			db.Model(&sqldb.RepositoryDB{}).Count(&count)
			assert.Equal(t, int64(tt.expectedAddRepos), count)
		})
	}
}

// TestRefreshNamespaceRepositoriesNoAccessToken tests RefreshNamespaceRepositories with no access token
// Python reference: test_refresh_namespace_repositories_no_access_token
func TestRefreshNamespaceRepositoriesNoAccessToken(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/users/") && !strings.Contains(r.URL.Path, "/repos") {
			// IsEntityOrgOrUser check - this shouldn't be called since there's no access token
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"type": "Organization",
			})
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	expectedConfig := &provider_source_model.ProviderSourceConfig{
		BaseURL:            server.URL,
		ApiURL:             server.URL,
		ClientID:           "test-client-id",
		ClientSecret:       "test-client-secret",
		LoginButtonText:    "Test Login",
		PrivateKeyPath:     "",
		AppID:              "123",
		DefaultAccessToken: "", // Empty - no default access token
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

	err = gh.RefreshNamespaceRepositories(context.Background(), database, "test-namespace")

	// Should get an error about no default access token
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not been configured")

	// Verify no repositories were added
	var count int64
	db.Model(&sqldb.RepositoryDB{}).Count(&count)
	assert.Equal(t, int64(0), count)
}

// TestRefreshNamespaceRepositoriesNoType tests RefreshNamespaceRepositories when entity type cannot be determined
// Python reference: test_refresh_namespace_repositories_no_type
func TestRefreshNamespaceRepositoriesNoType(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/users/") && !strings.Contains(r.URL.Path, "/repos") {
			// Return 404 to indicate entity doesn't exist
			w.WriteHeader(http.StatusNotFound)
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	expectedConfig := &provider_source_model.ProviderSourceConfig{
		BaseURL:             server.URL,
		ApiURL:              server.URL,
		ClientID:            "test-client-id",
		ClientSecret:        "test-client-secret",
		LoginButtonText:     "Test Login",
		PrivateKeyPath:      "",
		AppID:               "123",
		DefaultAccessToken:  "test-default-token",
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

	err = gh.RefreshNamespaceRepositories(context.Background(), database, "test-namespace")

	// Should get an error about not finding namespace entity
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "could not find namespace entity")

	// Verify no repositories were added
	var count int64
	db.Model(&sqldb.RepositoryDB{}).Count(&count)
	assert.Equal(t, int64(0), count)
}

// TestRefreshNamespaceRepositoriesInvalidResponse tests RefreshNamespaceRepositories with invalid API response
// Python reference: test_refresh_namespace_repositories_invalid_response_code
func TestRefreshNamespaceRepositoriesInvalidResponse(t *testing.T) {
	entityTypeChecked := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/users/") && !strings.Contains(r.URL.Path, "/repos") {
			// IsEntityOrgOrUser check - return org type
			entityTypeChecked = true
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"type": "Organization",
			})
		} else {
			// Repository listing - return invalid response code
			w.WriteHeader(http.StatusInternalServerError)
		}
	}))
	defer server.Close()

	expectedConfig := &provider_source_model.ProviderSourceConfig{
		BaseURL:             server.URL,
		ApiURL:              server.URL,
		ClientID:            "test-client-id",
		ClientSecret:        "test-client-secret",
		LoginButtonText:     "Test Login",
		PrivateKeyPath:      "",
		AppID:               "123",
		DefaultAccessToken:  "test-default-token",
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

	err = gh.RefreshNamespaceRepositories(context.Background(), database, "test-namespace")

	// Should get an error about invalid response code
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid response code")
	assert.True(t, entityTypeChecked, "Entity type should have been checked")

	// Verify no repositories were added
	var count int64
	db.Model(&sqldb.RepositoryDB{}).Count(&count)
	assert.Equal(t, int64(0), count)
}
