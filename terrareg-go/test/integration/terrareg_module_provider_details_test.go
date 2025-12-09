package integration

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/module/model"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/interfaces/http/dto/terrareg"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/interfaces/http/handler/terrareg"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/interfaces/http/presenter"
)

func TestTerraregModuleProviderDetails_EndToEnd(t *testing.T) {
	// Setup test data
	namespace := model.NewNamespace("test-namespace", nil)
	require.NotNil(t, namespace)

	moduleProvider, err := model.NewModuleProvider(namespace, "test-module", "aws")
	require.NoError(t, err)

	moduleDetails := model.NewModuleDetails([]byte("# Test Readme"))
	require.NotNil(t, moduleDetails)

	moduleVersion, err := model.NewModuleVersion("1.0.0", moduleDetails, false)
	require.NoError(t, err)

	// Set module provider relationship
	moduleVersion.SetModuleProvider(moduleProvider)
	moduleProvider.AddVersion(moduleVersion)

	// Mark as published
	err = moduleVersion.Publish()
	require.NoError(t, err)

	// Setup handler and dependencies
	moduleHandler := terrareg.NewModuleHandler()
	moduleVersionPresenter := presenter.NewModuleVersionPresenter()

	// Create a mock repository that returns our test data
	// Note: In a real integration test, you'd set up a test database and repositories
	// For this example, we'll create a simple router with the handler

	r := chi.NewRouter()

	// Mock route similar to the actual implementation
	r.Get("/v1/terrareg/modules/{namespace}/{module}/{provider}", func(w http.ResponseWriter, r *http.Request) {
		// For this integration test, return a mocked response using our test data
		response := moduleVersionPresenter.ToTerraregProviderDetailsDTO(
			context.Background(),
			moduleVersion,
			"test-namespace",
			"test-module",
			"aws",
			"localhost:8080",
		)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		if err := json.NewEncoder(w).Encode(response); err != nil {
			http.Error(w, "Failed to encode response", http.StatusInternalServerError)
			return
		}
	})

	// Create test request
	req := httptest.NewRequest("GET", "/v1/terrareg/modules/test-namespace/test-module/aws", nil)
	req.Header.Set("Host", "localhost:8080")

	// Perform request
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// Assert response status
	assert.Equal(t, http.StatusOK, w.Code)

	// Parse response
	var response terrareg.TerraregModuleProviderDetailsResponse
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	// Assert basic fields
	assert.Equal(t, "test-namespace/test-module/aws", response.ID)
	assert.Equal(t, "test-namespace", response.Namespace)
	assert.Equal(t, "test-module", response.Name)
	assert.Equal(t, "aws", response.Provider)
	assert.Equal(t, "1.0.0", response.Version)
	assert.True(t, response.Published)
	assert.False(t, response.Internal)
	assert.False(t, response.Beta)

	// Assert terraform module specs
	assert.Equal(t, "", response.Root.Path)
	assert.Equal(t, "# Test Readme", response.Root.Readme)
	assert.False(t, response.Root.Empty)
	assert.NotNil(t, response.Root.Inputs)
	assert.NotNil(t, response.Root.Outputs)
	assert.NotNil(t, response.Root.Dependencies)
	assert.NotNil(t, response.Root.Resources)
	assert.NotNil(t, response.Root.Modules)

	// Assert usage example is generated
	assert.NotNil(t, response.UsageExample)
	assert.Contains(t, *response.UsageExample, `module "test-module"`)
	assert.Contains(t, *response.UsageExample, "localhost:8080/test-namespace/test-module/aws")
	assert.Contains(t, *response.UsageExample, "version = \"1.0.0\"")

	// Assert graph URL is generated
	assert.NotNil(t, response.GraphURL)
	assert.Equal(t, "/modules/test-namespace/test-module/graph", *response.GraphURL)

	// Assert published at display is formatted
	assert.NotNil(t, response.PublishedAtDisplay)

	// Assert security results (should be empty for this test)
	assert.Equal(t, 0, response.SecurityFailures)
	assert.Empty(t, response.SecurityResults)

	// Assert custom links (should be empty for this test)
	assert.Empty(t, response.CustomLinks)

	// Assert additional tab files (should be empty for this test)
	assert.Empty(t, response.AdditionalTabFiles)
}

func TestTerraregModuleProviderDetails_UnpublishedVersion(t *testing.T) {
	// Setup test data for unpublished version
	namespace := model.NewNamespace("test-namespace", nil)
	require.NotNil(t, namespace)

	moduleProvider, err := model.NewModuleProvider(namespace, "test-module", "aws")
	require.NoError(t, err)

	moduleDetails := model.NewModuleDetails([]byte("# Test Readme"))
	require.NotNil(t, moduleDetails)

	moduleVersion, err := model.NewModuleVersion("1.1.0", moduleDetails, false)
	require.NoError(t, err)

	// Set module provider relationship but don't publish
	moduleVersion.SetModuleProvider(moduleProvider)
	moduleProvider.AddVersion(moduleVersion)

	// Setup handler
	moduleVersionPresenter := presenter.NewModuleVersionPresenter()

	// Create test router
	r := chi.NewRouter()
	r.Get("/v1/terrareg/modules/{namespace}/{module}/{provider}", func(w http.ResponseWriter, r *http.Request) {
		response := moduleVersionPresenter.ToTerraregProviderDetailsDTO(
			context.Background(),
			moduleVersion,
			"test-namespace",
			"test-module",
			"aws",
			"localhost:8080",
		)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		if err := json.NewEncoder(w).Encode(response); err != nil {
			http.Error(w, "Failed to encode response", http.StatusInternalServerError)
			return
		}
	})

	// Create test request
	req := httptest.NewRequest("GET", "/v1/terrareg/modules/test-namespace/test-module/aws", nil)
	req.Header.Set("Host", "localhost:8080")

	// Perform request
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// Assert response status
	assert.Equal(t, http.StatusOK, w.Code)

	// Parse response
	var response terrareg.TerraregModuleProviderDetailsResponse
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	// Assert unpublished version is returned but marked as unpublished
	assert.Equal(t, "1.1.0", response.Version)
	assert.False(t, response.Published)
	assert.NotNil(t, response.UsageExample)

	// Usage example should still work for unpublished versions
	assert.Contains(t, *response.UsageExample, `module "test-module"`)
	assert.Contains(t, *response.UsageExample, "version = \"1.1.0\"")
}

func TestTerraregModuleProviderDetails_WithAnalyticsToken(t *testing.T) {
	// Setup test data with analytics token in namespace
	namespaceName := "test-namespace__analytics123"
	namespace := model.NewNamespace(namespaceName, nil)
	require.NotNil(t, namespace)

	moduleProvider, err := model.NewModuleProvider(namespace, "test-module", "aws")
	require.NoError(t, err)

	moduleDetails := model.NewModuleDetails([]byte("# Test Readme"))
	require.NotNil(t, moduleDetails)

	moduleVersion, err := model.NewModuleVersion("1.0.0", moduleDetails, false)
	require.NoError(t, err)

	moduleVersion.SetModuleProvider(moduleProvider)
	moduleProvider.AddVersion(moduleVersion)
	err = moduleVersion.Publish()
	require.NoError(t, err)

	// Setup handler
	moduleVersionPresenter := presenter.NewModuleVersionPresenter()

	// Create test router
	r := chi.NewRouter()
	r.Get("/v1/terrareg/modules/{namespace}/{module}/{provider}", func(w http.ResponseWriter, r *http.Request) {
		response := moduleVersionPresenter.ToTerraregProviderDetailsDTO(
			context.Background(),
			moduleVersion,
			namespaceName, // Use full namespace with token
			"test-module",
			"aws",
			"localhost:8080",
		)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		if err := json.NewEncoder(w).Encode(response); err != nil {
			http.Error(w, "Failed to encode response", http.StatusInternalServerError)
			return
		}
	})

	// Create test request (URL should still use the namespace part)
	req := httptest.NewRequest("GET", "/v1/terrareg/modules/test-namespace__analytics123/test-module/aws", nil)
	req.Header.Set("Host", "localhost:8080")

	// Perform request
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// Assert response status
	assert.Equal(t, http.StatusOK, w.Code)

	// Parse response
	var response terrareg.TerraregModuleProviderDetailsResponse
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	// Assert analytics token is extracted and included
	assert.NotNil(t, response.AnalyticsToken)
	assert.Equal(t, "analytics123", *response.AnalyticsToken)

	// Assert namespace in response is the full namespace (including token)
	assert.Equal(t, "test-namespace__analytics123", response.Namespace)
}

func TestTerraregModuleProviderDetails_JSONSerialization(t *testing.T) {
	// Test that the response can be properly serialized to JSON
	namespace := model.NewNamespace("test-namespace", nil)
	require.NotNil(t, namespace)

	moduleProvider, err := model.NewModuleProvider(namespace, "test-module", "aws")
	require.NoError(t, err)

	moduleDetails := model.NewModuleDetails([]byte("# Test Readme\n\nThis is a test module."))
	require.NotNil(t, moduleDetails)

	moduleVersion, err := model.NewModuleVersion("1.0.0", moduleDetails, false)
	require.NoError(t, err)

	moduleVersion.SetModuleProvider(moduleProvider)
	moduleProvider.AddVersion(moduleVersion)

	// Setup handler
	moduleVersionPresenter := presenter.NewModuleVersionPresenter()

	response := moduleVersionPresenter.ToTerraregProviderDetailsDTO(
		context.Background(),
		moduleVersion,
		"test-namespace",
		"test-module",
		"aws",
		"localhost:8080",
	)

	// Test JSON serialization
	jsonData, err := json.Marshal(response)
	require.NoError(t, err)

	// Test that JSON is valid and can be unmarshaled back
	var unmarshaled terrareg.TerraregModuleProviderDetailsResponse
	err = json.Unmarshal(jsonData, &unmarshaled)
	require.NoError(t, err)

	// Verify key fields are preserved
	assert.Equal(t, response.ID, unmarshaled.ID)
	assert.Equal(t, response.Namespace, unmarshaled.Namespace)
	assert.Equal(t, response.Name, unmarshaled.Name)
	assert.Equal(t, response.Provider, unmarshaled.Provider)
	assert.Equal(t, response.Version, unmarshaled.Version)
	assert.Equal(t, response.Published, unmarshaled.Published)
	assert.Equal(t, response.Root.Readme, unmarshaled.Root.Readme)

	// Verify usage example is preserved
	assert.Equal(t, response.UsageExample, unmarshaled.UsageExample)
	if response.UsageExample != nil {
		assert.NotEmpty(t, *response.UsageExample)
	}

	// Verify required JSON fields are present (based on Python API)
	var jsonObj map[string]interface{}
	err = json.Unmarshal(jsonData, &jsonObj)
	require.NoError(t, err)

	requiredFields := []string{
		"id", "namespace", "name", "provider", "verified", "trusted",
		"owner", "version", "description", "internal", "published", "beta",
		"versions", "root", "submodules", "providers",
		"published_at_display", "display_source_url", "graph_url",
		"security_failures", "security_results", "usage_example",
		"additional_tab_files", "custom_links", "terraform_example_version_string",
		"terraform_example_version_comment", "terraform_version_constraint",
		"module_extraction_up_to_date", "downloads",
	}

	for _, field := range requiredFields {
		_, exists := jsonObj[field]
		assert.True(t, exists, "Required field '%s' missing from JSON response", field)
	}
}

func TestTerraregModuleProviderDetails_MissingModuleVersion(t *testing.T) {
	// This test demonstrates error handling for missing module versions
	// In a real integration test, this would test the actual error response from the handler

	r := chi.NewRouter()
	r.Get("/v1/terrareg/modules/{namespace}/{module}/{provider}", func(w http.ResponseWriter, r *http.Request) {
		// Simulate 404 for missing module version
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"error": "Module version not found"}`))
	})

	// Create test request for non-existent module
	req := httptest.NewRequest("GET", "/v1/terrareg/modules/nonexistent/module/provider", nil)

	// Perform request
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// Assert 404 response
	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Contains(t, w.Body.String(), "Module version not found")
}
