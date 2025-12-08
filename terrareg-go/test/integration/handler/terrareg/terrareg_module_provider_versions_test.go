package terrareg

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/module/model"
)

func TestTerraregModuleProviderVersions_EndpointLogic(t *testing.T) {
	// Setup test data
	namespace, err := model.NewNamespace("versions-test", nil, model.NamespaceTypeNone)
	require.NoError(t, err)
	require.NotNil(t, namespace)

	moduleProvider, err := model.NewModuleProvider(namespace, "versions-module", "aws")
	require.NoError(t, err)

	// Create multiple versions
	moduleDetails := model.NewModuleDetails([]byte("# Versions Test Module"))
	require.NotNil(t, moduleDetails)

	version1, err := model.NewModuleVersion("1.0.0", moduleDetails, false)
	require.NoError(t, err)

	version2, err := model.NewModuleVersion("1.1.0", moduleDetails, false)
	require.NoError(t, err)

	version3, err := model.NewModuleVersion("2.0.0", moduleDetails, true) // beta version
	require.NoError(t, err)

	version4, err := model.NewModuleVersion("2.1.0", moduleDetails, true) // published beta version
	require.NoError(t, err)

	// Add versions to module provider
	moduleProvider.AddVersion(version1)
	moduleProvider.AddVersion(version3) // Add beta version
	moduleProvider.AddVersion(version2)
	moduleProvider.AddVersion(version4) // Add published beta version

	// Publish only some versions
	err = version1.Publish()
	require.NoError(t, err)
	err = version2.Publish()
	require.NoError(t, err)
	err = version4.Publish() // Publish the beta version
	require.NoError(t, err)
	// version3 remains unpublished (beta)

	// Create test router with the actual handler logic
	r := chi.NewRouter()

	// Mock route that simulates the handler logic
	r.Get("/v1/terrareg/modules/{namespace}/{name}/{provider}/versions", func(w http.ResponseWriter, r *http.Request) {
		// Extract query parameters
		includeBeta := r.URL.Query().Get("include-beta") == "true"
		includeUnpublished := r.URL.Query().Get("include-unpublished") == "true"

		// Get all versions from module provider
		allVersions := moduleProvider.GetAllVersions()

		var versions []map[string]interface{}
		for _, version := range allVersions {
			if !includeBeta && version.IsBeta() {
				continue
			}
			if !includeUnpublished && !version.IsPublished() {
				continue
			}

			versions = append(versions, map[string]interface{}{
				"version":   version.Version().String(),
				"published": version.IsPublished(),
				"beta":      version.IsBeta(),
			})
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		// Encode and send response
		encoder := json.NewEncoder(w)
		encoder.SetEscapeHTML(false)
		if err := encoder.Encode(versions); err != nil {
			http.Error(w, "Failed to encode response", http.StatusInternalServerError)
			return
		}
	})

	// Test 1: Get all versions (default behavior - published only, no beta)
	req := httptest.NewRequest("GET", "/v1/terrareg/modules/versions-test/versions-module/aws/versions", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response []map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	// Should only contain published, non-beta versions
	assert.Len(t, response, 2, "Should contain exactly 2 published non-beta versions")

	// Verify version objects have correct structure
	for _, versionObj := range response {
		assert.Contains(t, versionObj, "version", "Version object should have 'version' field")
		assert.Contains(t, versionObj, "published", "Version object should have 'published' field")
		assert.Contains(t, versionObj, "beta", "Version object should have 'beta' field")

		version, ok := versionObj["version"].(string)
		assert.True(t, ok, "Version should be a string")
		assert.NotEmpty(t, version, "Version should not be empty")

		published, ok := versionObj["published"].(bool)
		assert.True(t, ok, "Published should be a boolean")

		beta, ok := versionObj["beta"].(bool)
		assert.True(t, ok, "Beta should be a boolean")

		// Verify default filtering: only published, non-beta versions
		assert.True(t, published, "Default behavior should only return published versions")
		assert.False(t, beta, "Default behavior should not return beta versions")
	}

	// Test 2: Include beta versions (but not unpublished) - should include published beta versions only
	req = httptest.NewRequest("GET", "/v1/terrareg/modules/versions-test/versions-module/aws/versions?include-beta=true", nil)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	
	// Should contain published versions + published beta versions (not unpublished beta)
	assert.Len(t, response, 3, "Should contain 3 versions when including beta (published non-beta + published beta)")

	// Verify published beta version is included
	foundPublishedBeta := false
	for _, versionObj := range response {
		if version, ok := versionObj["version"].(string); ok && version == "2.1.0" {
			foundPublishedBeta = true
			assert.True(t, versionObj["published"].(bool), "Published beta version should be published")
			assert.True(t, versionObj["beta"].(bool), "Version 2.1.0 should be marked as beta")
			break
		}
	}
	assert.True(t, foundPublishedBeta, "Published beta version 2.1.0 should be included when include-beta=true")

	// Test 3: Include unpublished versions (but not beta)
	req = httptest.NewRequest("GET", "/v1/terrareg/modules/versions-test/versions-module/aws/versions?include-unpublished=true", nil)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	// Should contain published non-beta versions + unpublished non-beta versions (beta versions filtered out)
	assert.Len(t, response, 2, "Should contain published non-beta versions + unpublished non-beta versions (beta versions filtered out)")

	// Test 4: Include both beta and unpublished versions
	req = httptest.NewRequest("GET", "/v1/terrareg/modules/versions-test/versions-module/aws/versions?include-beta=true&include-unpublished=true", nil)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	// Should contain all versions
	assert.Len(t, response, 4, "Should contain all 4 versions when including both beta and unpublished")

	// Verify all versions are present
	versionMap := make(map[string]map[string]interface{})
	for _, versionObj := range response {
		if version, ok := versionObj["version"].(string); ok {
			versionMap[version] = versionObj
		}
	}

	// Check specific versions
	assert.Contains(t, versionMap, "1.0.0", "Version 1.0.0 should be present")
	assert.Equal(t, true, versionMap["1.0.0"]["published"], "1.0.0 should be published")
	assert.Equal(t, false, versionMap["1.0.0"]["beta"], "1.0.0 should not be beta")

	assert.Contains(t, versionMap, "1.1.0", "Version 1.1.0 should be present")
	assert.Equal(t, true, versionMap["1.1.0"]["published"], "1.1.0 should be published")
	assert.Equal(t, false, versionMap["1.1.0"]["beta"], "1.1.0 should not be beta")

	assert.Contains(t, versionMap, "2.0.0", "Version 2.0.0 should be present")
	assert.Equal(t, false, versionMap["2.0.0"]["published"], "2.0.0 should be unpublished")
	assert.Equal(t, true, versionMap["2.0.0"]["beta"], "2.0.0 should be beta")

	assert.Contains(t, versionMap, "2.1.0", "Version 2.1.0 should be present")
	assert.Equal(t, true, versionMap["2.1.0"]["published"], "2.1.0 should be published")
	assert.Equal(t, true, versionMap["2.1.0"]["beta"], "2.1.0 should be beta")
}