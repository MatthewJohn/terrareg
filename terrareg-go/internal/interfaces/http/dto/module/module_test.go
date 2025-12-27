package module

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestModuleSpecs_StandardAPIFields(t *testing.T) {
	// Test that standard ModuleSpecs has correct fields (no 'modules' field)
	specs := ModuleSpecs{
		Path:    "",
		Readme:  "# Test Readme",
		Empty:   false,
		Inputs:  []TerraformInput{},
		Outputs: []TerraformOutput{},
		Dependencies: []TerraformDependency{
			{
				Module:  "vpc",
				Source:  "terraform-aws-modules/vpc/aws",
				Version: ">= 3.0",
			},
		},
		ProviderDependencies: []TerraformProvider{
			{
				Name: "aws",
			},
		},
		Resources: []TerraformResource{
			{
				Name: "vpc_id",
				Type: "string",
			},
		},
	}

	// Test JSON serialization
	jsonData, err := json.Marshal(specs)
	assert.NoError(t, err)

	var parsed ModuleSpecs
	err = json.Unmarshal(jsonData, &parsed)
	assert.NoError(t, err)

	assert.Equal(t, specs.Path, parsed.Path)
	assert.Equal(t, specs.Readme, parsed.Readme)
	assert.Equal(t, specs.Empty, parsed.Empty)
	assert.Len(t, parsed.Dependencies, 1)
	assert.Equal(t, "vpc", parsed.Dependencies[0].Module)
	assert.Equal(t, "aws", parsed.ProviderDependencies[0].Name)
	assert.Equal(t, "vpc_id", parsed.Resources[0].Name)
}

func TestTerraformDependency_CorrectFieldNames(t *testing.T) {
	// Test that TerraformDependency uses 'Module' field (not 'name') like Python
	dep := TerraformDependency{
		Module:  "vpc",
		Source:  "terraform-aws-modules/vpc/aws",
		Version: ">= 3.0",
	}

	jsonData, err := json.Marshal(dep)
	assert.NoError(t, err)

	var parsed TerraformDependency
	err = json.Unmarshal(jsonData, &parsed)
	assert.NoError(t, err)

	assert.Equal(t, "vpc", parsed.Module) // Should be 'Module' not 'name'
	assert.Equal(t, "terraform-aws-modules/vpc/aws", parsed.Source)
	assert.Equal(t, ">= 3.0", parsed.Version)
}
