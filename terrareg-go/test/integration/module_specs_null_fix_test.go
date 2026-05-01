package integration

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"testing"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/interfaces/http/dto/terrareg"
)

func TestModuleSpecs_JSON_EmptyArrays(t *testing.T) {
	// Test that TerraregModuleSpecs serializes empty slices as [] rather than null
	specs := terrareg.TerraregModuleSpecs{
		Path:                 "/test",
		Readme:               "",
		Empty:                true,
		Inputs:               []terrareg.TerraregInput{},
		Outputs:              []terrareg.TerraregOutput{},
		Dependencies:         []terrareg.TerraregDependency{},
		ProviderDependencies: []terrareg.TerraregProviderDep{},
		Resources:            []terrareg.TerraregResource{},
		Modules:              []terrareg.TerraregModule{},
	}

	// Serialize to JSON
	jsonBytes, err := json.Marshal(specs)
	if err != nil {
		t.Fatalf("Failed to marshal TerraregModuleSpecs: %v", err)
	}

	// Parse back to verify structure
	var unmarshaled terrareg.TerraregModuleSpecs
	err = json.Unmarshal(jsonBytes, &unmarshaled)
	if err != nil {
		t.Fatalf("Failed to unmarshal TerraregModuleSpecs: %v", err)
	}

	// Verify that all slice fields are empty arrays, not null
	jsonStr := string(jsonBytes)
	t.Logf("Serialized JSON: %s", jsonStr)

	// Check that slice fields appear as empty arrays
	assert.Contains(t, jsonStr, `"inputs":[]`, "Inputs should be empty array")
	assert.Contains(t, jsonStr, `"outputs":[]`, "Outputs should be empty array")
	assert.Contains(t, jsonStr, `"dependencies":[]`, "Dependencies should be empty array")
	assert.Contains(t, jsonStr, `"provider_dependencies":[]`, "ProviderDependencies should be empty array")
	assert.Contains(t, jsonStr, `"resources":[]`, "Resources should be empty array")
	assert.Contains(t, jsonStr, `"modules":[]`, "Modules should be empty array")

	// Verify the unmarshaled values are properly initialized
	assert.NotNil(t, unmarshaled.Inputs, "Inputs should not be nil")
	assert.NotNil(t, unmarshaled.Outputs, "Outputs should not be nil")
	assert.NotNil(t, unmarshaled.Dependencies, "Dependencies should not be nil")
	assert.NotNil(t, unmarshaled.ProviderDependencies, "ProviderDependencies should not be nil")
	assert.NotNil(t, unmarshaled.Resources, "Resources should not be nil")
	assert.NotNil(t, unmarshaled.Modules, "Modules should not be nil")

	assert.Empty(t, unmarshaled.Inputs, "Inputs should be empty")
	assert.Empty(t, unmarshaled.Outputs, "Outputs should be empty")
	assert.Empty(t, unmarshaled.Dependencies, "Dependencies should be empty")
	assert.Empty(t, unmarshaled.ProviderDependencies, "ProviderDependencies should be empty")
	assert.Empty(t, unmarshaled.Resources, "Resources should be empty")
	assert.Empty(t, unmarshaled.Modules, "Modules should be empty")
}

func TestModuleProviderDetailsResponse_JSON_EmptyArrays(t *testing.T) {
	// Test that the full response structure also properly handles empty arrays
	response := terrareg.TerraregModuleProviderDetailsResponse{
		ID:        "test/test/test/1.0.0",
		Namespace: "test",
		Name:      "test",
		Provider:  "test",
		Version:   "1.0.0",
		Root: terrareg.TerraregModuleSpecs{
			Path:                 "/test",
			Readme:               "",
			Empty:                true,
			Inputs:               []terrareg.TerraregInput{},
			Outputs:              []terrareg.TerraregOutput{},
			Dependencies:         []terrareg.TerraregDependency{},
			ProviderDependencies: []terrareg.TerraregProviderDep{},
			Resources:            []terrareg.TerraregResource{},
			Modules:              []terrareg.TerraregModule{},
		},
		Submodules:  []terrareg.TerraregModuleSpecs{},
		Providers:   []string{},
		CustomLinks: []terrareg.TerraregCustomLink{},
		// Initialize other slice fields that could be null
		AdditionalTabFiles: map[string]string{},
		SecurityResults:    []terrareg.TerraregSecurityResult{},
	}

	// Serialize to JSON
	jsonBytes, err := json.Marshal(response)
	if err != nil {
		t.Fatalf("Failed to marshal TerraregModuleProviderDetailsResponse: %v", err)
	}

	jsonStr := string(jsonBytes)
	t.Logf("Serialized JSON: %s", jsonStr)

	// Verify that slice fields appear as empty arrays, not null
	assert.Contains(t, jsonStr, `"inputs":[]`, "Root inputs should be empty array")
	assert.Contains(t, jsonStr, `"outputs":[]`, "Root outputs should be empty array")
	assert.Contains(t, jsonStr, `"dependencies":[]`, "Root dependencies should be empty array")
	assert.Contains(t, jsonStr, `"provider_dependencies":[]`, "Root provider dependencies should be empty array")
	assert.Contains(t, jsonStr, `"resources":[]`, "Root resources should be empty array")
	assert.Contains(t, jsonStr, `"modules":[]`, "Root modules should be empty array")
	assert.Contains(t, jsonStr, `"submodules":[]`, "Submodules should be empty array")
	assert.Contains(t, jsonStr, `"providers":[]`, "Providers should be empty array")
	assert.Contains(t, jsonStr, `"custom_links":[]`, "Custom links should be empty array")
	assert.Contains(t, jsonStr, `"additional_tab_files":{}`, "Additional tab files should be empty object")
	assert.Contains(t, jsonStr, `"security_results":[]`, "Security results should be empty array")
}
