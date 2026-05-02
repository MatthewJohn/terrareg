package terrareg_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/persistence/sqldb"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/interfaces/http/handler/terrareg"
	"github.com/matthewjohn/terrareg/terrareg-go/test/integration/testutils"
)

// TestGitProvidersHandler_ServeHTTP_Success tests listing git providers successfully
func TestGitProvidersHandler_ServeHTTP_Success(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	// Create a test git provider
	gitProvider := sqldb.GitProviderDB{
		Name:            "test-git",
		GitPathTemplate: "https://github.com/{namespace}/{module}.git",
	}
	err := db.DB.Create(&gitProvider).Error
	require.NoError(t, err)

	// Create handler
	cont := testutils.CreateTestContainer(t, db)
	handler := terrareg.NewGitProvidersHandler(cont.GitProviderFactory)

	// Create request
	req := httptest.NewRequest("GET", "/v1/terrareg/git-providers", nil)
	w := httptest.NewRecorder()

	// Act
	handler.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	testutils.AssertContentType(t, w, "application/json")

	var response []map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	// Should have at least one git provider
	assert.NotEmpty(t, response)

	// Verify response structure
	for _, provider := range response {
		assert.Contains(t, provider, "id")
		assert.Contains(t, provider, "name")
		assert.Contains(t, provider, "git_path_template")
	}
}

// TestGitProvidersHandler_ServeHTTP_MultipleProviders tests listing multiple git providers
func TestGitProvidersHandler_ServeHTTP_MultipleProviders(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	// Create additional git providers in database
	gitProvider1 := sqldb.GitProviderDB{
		Name:            "custom-git1",
		GitPathTemplate: "https://custom1.com/{path}",
	}
	err := db.DB.Create(&gitProvider1).Error
	require.NoError(t, err)

	gitProvider2 := sqldb.GitProviderDB{
		Name:            "custom-git2",
		GitPathTemplate: "https://custom2.com/{path}",
	}
	err = db.DB.Create(&gitProvider2).Error
	require.NoError(t, err)

	// Create handler
	cont := testutils.CreateTestContainer(t, db)
	handler := terrareg.NewGitProvidersHandler(cont.GitProviderFactory)

	// Create request
	req := httptest.NewRequest("GET", "/v1/terrareg/git-providers", nil)
	w := httptest.NewRecorder()

	// Act
	handler.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)

	var response []map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	// Should have at least 2 providers
	assert.GreaterOrEqual(t, len(response), 2)
}

// TestGitProvidersHandler_ServeHTTP_Empty tests listing git providers when database is empty
func TestGitProvidersHandler_ServeHTTP_Empty(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	// Create handler
	cont := testutils.CreateTestContainer(t, db)
	handler := terrareg.NewGitProvidersHandler(cont.GitProviderFactory)

	// Create request
	req := httptest.NewRequest("GET", "/v1/terrareg/git-providers", nil)
	w := httptest.NewRecorder()

	// Act
	handler.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)

	var response []map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	// Empty list is valid
	assert.NotNil(t, response)
}

// TestGitProvidersHandler_ServeHTTP_ResponseStructure tests the response structure matches expected format
func TestGitProvidersHandler_ServeHTTP_ResponseStructure(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	// Create a test git provider
	gitProvider := sqldb.GitProviderDB{
		Name:            "test-git",
		GitPathTemplate: "https://github.com/{namespace}/{module}.git",
	}
	err := db.DB.Create(&gitProvider).Error
	require.NoError(t, err)

	// Create handler
	cont := testutils.CreateTestContainer(t, db)
	handler := terrareg.NewGitProvidersHandler(cont.GitProviderFactory)

	// Create request
	req := httptest.NewRequest("GET", "/v1/terrareg/git-providers", nil)
	w := httptest.NewRecorder()

	// Act
	handler.ServeHTTP(w, req)

	// Assert
	var response []map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	// Find our test provider in the response
	var testProvider *map[string]interface{}
	for _, p := range response {
		if p["name"] == "test-git" {
			testProvider = &p
			break
		}
	}

	require.NotNil(t, testProvider, "Test provider should be in response")

	// Verify all fields are present
	assert.Contains(t, *testProvider, "id")
	assert.Equal(t, "test-git", (*testProvider)["name"])
	assert.Equal(t, "https://github.com/{namespace}/{module}.git", (*testProvider)["git_path_template"])
}

// TestGitProvidersHandler_ServeHTTP_WithAllTemplates tests git provider with all template fields
func TestGitProvidersHandler_ServeHTTP_WithAllTemplates(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	// Create a test git provider with all templates
	gitProvider := sqldb.GitProviderDB{
		Name:              "full-git",
		BaseURLTemplate:   "https://github.com",
		CloneURLTemplate:  "https://github.com/{namespace}/{module}.git",
		BrowseURLTemplate: "https://github.com/{namespace}/{module}",
		GitPathTemplate:   "{namespace}/{module}",
	}
	err := db.DB.Create(&gitProvider).Error
	require.NoError(t, err)

	// Create handler
	cont := testutils.CreateTestContainer(t, db)
	handler := terrareg.NewGitProvidersHandler(cont.GitProviderFactory)

	// Create request
	req := httptest.NewRequest("GET", "/v1/terrareg/git-providers", nil)
	w := httptest.NewRecorder()

	// Act
	handler.ServeHTTP(w, req)

	// Assert
	var response []map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	// Find our test provider in the response
	var testProvider *map[string]interface{}
	for _, p := range response {
		if p["name"] == "full-git" {
			testProvider = &p
			break
		}
	}

	require.NotNil(t, testProvider, "Test provider should be in response")

	// Verify git_path_template is present (this is what the handler returns)
	assert.Equal(t, "{namespace}/{module}", (*testProvider)["git_path_template"])
}
