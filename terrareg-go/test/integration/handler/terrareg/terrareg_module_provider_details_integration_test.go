package terrareg

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/module/model"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/interfaces/http/dto/terrareg"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/interfaces/http/presenter"
)

func TestTerraregModuleProviderDetails_Integration_BasicFunctionality(t *testing.T) {
	// Test the basic functionality of ToTerraregProviderDetailsDTO method
	// This tests the integration between domain models and presenter

	// Setup test data
	namespace, err := model.NewNamespace("test-namespace", nil, model.NamespaceTypeNone)
	require.NoError(t, err)

	moduleProvider, err := model.NewModuleProvider(namespace, "test-module", "aws")
	require.NoError(t, err)

	moduleDetails := model.NewCompleteModuleDetails(
		[]byte("# Test Module\n\nThis is a test module with readme content."),
		nil,     // terraform docs
		nil,     // tfsec
		nil,     // infracost
		nil,     // terraform graph
		nil,     // terraform modules
		"1.0.0", // terraform version
	)
	require.NotNil(t, moduleDetails)

	moduleVersion, err := model.NewModuleVersion("1.0.0", moduleDetails, false)
	require.NoError(t, err)

	// Set up relationships using the aggregate root pattern
	moduleProvider.AddVersion(moduleVersion)

	// Publish the version
	err = moduleVersion.Publish()
	require.NoError(t, err)

	// Create presenter
	moduleVersionPresenter := presenter.NewModuleVersionPresenter()

	// Test the presenter method
	response := moduleVersionPresenter.ToTerraregProviderDetailsDTO(
		context.Background(),
		moduleVersion,
		"test-namespace",
		"test-module",
		"aws",
		"localhost:8080",
	)

	// Verify response structure
	require.NotNil(t, response)
	assert.Equal(t, "test-namespace/test-module/aws", response.ID)
	assert.Equal(t, "test-namespace", response.Namespace)
	assert.Equal(t, "test-module", response.Name)
	assert.Equal(t, "aws", response.Provider)
	assert.Equal(t, "1.0.0", response.Version)
	assert.True(t, response.Published)
	assert.False(t, response.Internal)
	assert.False(t, response.Beta)

	// Verify versions list contains the version we added
	assert.NotEmpty(t, response.Versions, "Versions list should not be empty")
	assert.Contains(t, response.Versions, "1.0.0", "Versions list should contain the version we added")

	// Verify terraform module specs
	assert.Equal(t, "", response.Root.Path)
	assert.Equal(t, "# Test Module\n\nThis is a test module with readme content.", response.Root.Readme)
	assert.False(t, response.Root.Empty)
	// Note: These fields are currently empty/nil because terraform parsing is not implemented yet
	assert.Empty(t, response.Root.Inputs)
	assert.Empty(t, response.Root.Outputs)
	assert.Empty(t, response.Root.Dependencies)
	assert.Empty(t, response.Root.Resources)
	assert.Empty(t, response.Root.Modules)

	// Verify usage example generation
	assert.NotNil(t, response.UsageExample)
	assert.Contains(t, *response.UsageExample, `module "test-module"`)
	assert.Contains(t, *response.UsageExample, "localhost:8080/test-namespace/test-module/aws")
	assert.Contains(t, *response.UsageExample, "version = \"1.0.0\"")

	// Verify graph URL generation
	assert.NotNil(t, response.GraphURL)
	assert.Equal(t, "/modules/test-namespace/test-module/graph", *response.GraphURL)

	// Verify published at display formatting
	assert.NotNil(t, response.PublishedAtDisplay)
	assert.NotEmpty(t, *response.PublishedAtDisplay)

	// Verify security results (should be empty for this test)
	assert.Equal(t, 0, response.SecurityFailures)
	assert.Empty(t, response.SecurityResults)

	// Verify custom links (should be empty for this test)
	assert.Empty(t, response.CustomLinks)

	// Verify additional tab files (should be empty for this test)
	assert.Empty(t, response.AdditionalTabFiles)

	// Verify module extraction status
	assert.False(t, response.ModuleExtractionUpToDate) // Should be false for basic module
}

func TestTerraregModuleProviderDetails_Integration_JSONSerialization(t *testing.T) {
	// Test that the response properly serializes to JSON

	namespace, err := model.NewNamespace("json-test", nil, model.NamespaceTypeNone)
	require.NoError(t, err)

	moduleProvider, err := model.NewModuleProvider(namespace, "json-module", "gcp")
	require.NoError(t, err)

	moduleDetails := model.NewModuleDetails([]byte("# JSON Test Module"))
	require.NotNil(t, moduleDetails)

	moduleVersion, err := model.NewModuleVersion("1.2.3", moduleDetails, false)
	require.NoError(t, err)

	moduleProvider.AddVersion(moduleVersion)
	err = moduleVersion.Publish()
	require.NoError(t, err)

	// Create presenter
	moduleVersionPresenter := presenter.NewModuleVersionPresenter()

	// Generate response
	response := moduleVersionPresenter.ToTerraregProviderDetailsDTO(
		context.Background(),
		moduleVersion,
		"json-test",
		"json-module",
		"gcp",
		"example.com",
	)

	// Test JSON serialization
	jsonData, err := json.Marshal(response)
	require.NoError(t, err)

	// Test that JSON is valid and can be unmarshaled back
	var unmarshaled terrareg.TerraregModuleProviderDetailsResponse
	err = json.Unmarshal(jsonData, &unmarshaled)
	require.NoError(t, err)

	// Verify key fields are preserved through serialization
	assert.Equal(t, response.ID, unmarshaled.ID)
	assert.Equal(t, response.Namespace, unmarshaled.Namespace)
	assert.Equal(t, response.Name, unmarshaled.Name)
	assert.Equal(t, response.Provider, unmarshaled.Provider)
	assert.Equal(t, response.Version, unmarshaled.Version)
	assert.Equal(t, response.Published, unmarshaled.Published)
	assert.Equal(t, response.Beta, unmarshaled.Beta)
	assert.Equal(t, response.Root.Readme, unmarshaled.Root.Readme)

	// Verify usage example is preserved
	if response.UsageExample != nil {
		assert.Equal(t, response.UsageExample, unmarshaled.UsageExample)
	}

	// Verify analytics token is preserved
	if response.AnalyticsToken != nil {
		assert.Equal(t, response.AnalyticsToken, unmarshaled.AnalyticsToken)
	}

	// Test that the JSON contains all required fields (based on Python API)
	var jsonObj map[string]interface{}
	err = json.Unmarshal(jsonData, &jsonObj)
	require.NoError(t, err)

	// Check for critical fields
	criticalFields := []string{
		"id", "namespace", "name", "provider", "verified", "trusted",
		"version", "published", "internal", "beta",
		"root", "submodules", "providers",
		"usage_example", "graph_url",
		"security_failures", "security_results",
		"downloads",
	}

	for _, field := range criticalFields {
		_, exists := jsonObj[field]
		assert.True(t, exists, "Critical field '%s' missing from JSON response", field)
	}
}

func TestTerraregModuleProviderDetails_Integration_VersionsList(t *testing.T) {
	// Test specifically that versions list is populated correctly

	namespace, err := model.NewNamespace("versions-test", nil, model.NamespaceTypeNone)
	require.NoError(t, err)

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

	// Add versions to module provider (order matters for testing)
	moduleProvider.AddVersion(version1)
	moduleProvider.AddVersion(version3) // Add beta version
	moduleProvider.AddVersion(version2)

	// Publish only some versions
	err = version1.Publish()
	require.NoError(t, err)
	err = version2.Publish()
	require.NoError(t, err)
	// version3 remains unpublished (beta)

	// Test presenter with a specific version
	moduleVersionPresenter := presenter.NewModuleVersionPresenter()
	response := moduleVersionPresenter.ToTerraregProviderDetailsDTO(
		context.Background(),
		version3, // Use the beta version (2.0.0) as the main version
		"versions-test",
		"versions-module",
		"aws",
		"localhost:8080",
	)

	// Verify all versions are in the list
	require.NotNil(t, response)
	assert.NotEmpty(t, response.Versions, "Versions list should not be empty")
	assert.Len(t, response.Versions, 3, "Should contain all 3 versions")

	// Verify specific versions are present
	assert.Contains(t, response.Versions, "1.0.0", "Should contain version 1.0.0")
	assert.Contains(t, response.Versions, "1.1.0", "Should contain version 1.1.0")
	assert.Contains(t, response.Versions, "2.0.0", "Should contain version 2.0.0")

	// The current response should be for the version we passed (2.0.0)
	assert.Equal(t, "2.0.0", response.Version)
	assert.True(t, response.Beta, "2.0.0 should be marked as beta")

	// Test that versions field would catch empty lists
	var emptyVersions []string
	assert.NotEqual(t, emptyVersions, response.Versions, "Versions field should never be empty when module provider has versions")
}
