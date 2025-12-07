package module

import (
	"encoding/json"
	"testing"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/interfaces/http/dto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestModuleProviderResponse_JSONSerialization(t *testing.T) {
	response := ModuleProviderResponse{
		ProviderBase: ProviderBase{
			ID:        "example/aws/2.0.0",
			Namespace: "example",
			Name:      "aws",
			Provider:  "aws",
			Verified:  true,
			Trusted:   false,
		},
		Description: &[]string{"AWS provider for Terraform"}[0],
		Downloads:   1500,
	}

	// Test JSON serialization
	jsonData, err := json.Marshal(response)
	require.NoError(t, err)
	assert.JSONEq(t, `{"id":"example/aws/2.0.0","namespace":"example","name":"aws","provider":"aws","verified":true,"trusted":false,"description":"AWS provider for Terraform","owner":null,"source":null,"published_at":null,"downloads":1500}`, string(jsonData))

	// Test JSON deserialization
	var unmarshaled ModuleProviderResponse
	err = json.Unmarshal(jsonData, &unmarshaled)
	require.NoError(t, err)
	assert.Equal(t, response, unmarshaled)
}

func TestModuleVersionResponse_JSONSerialization(t *testing.T) {
	response := ModuleVersionResponse{
		VersionBase: VersionBase{
			ProviderBase: ProviderBase{
				ID:        "example/aws/2.1.0",
				Namespace: "example",
				Name:      "aws",
				Provider:  "aws",
				Verified:  true,
				Trusted:   false,
			},
			Version:     "2.1.0",
			Owner:       &[]string{"example-team"}[0],
			Description: &[]string{"AWS provider version 2.1.0"}[0],
			Source:      &[]string{"https://github.com/example/aws/tree/v2.1.0"}[0],
			PublishedAt: &[]string{"2023-02-01T15:45:00Z"}[0],
			Downloads:   320,
			Internal:    false,
		},
	}

	// Test JSON serialization
	jsonData, err := json.Marshal(response)
	require.NoError(t, err)
	assert.Contains(t, string(jsonData), `"id":"example/aws/2.1.0"`)
	assert.Contains(t, string(jsonData), `"version":"2.1.0"`)
	assert.Contains(t, string(jsonData), `"downloads":320`)
	assert.Contains(t, string(jsonData), `"internal":false`)
}

func TestModuleSearchResponse_PaginationMeta(t *testing.T) {
	modules := []ModuleProviderResponse{
		{
			ProviderBase: ProviderBase{
				ID:        "example/aws/2.0.0",
				Namespace: "example",
				Name:      "aws",
				Provider:  "aws",
				Verified:  true,
				Trusted:   false,
			},
		},
	}

	response := ModuleSearchResponse{
		Modules: modules,
		Meta: dto.PaginationMeta{
			Limit:      20,
			Offset:     0,
			TotalCount: 2,
		},
	}

	// Test JSON serialization
	jsonData, err := json.Marshal(response)
	require.NoError(t, err)

	// Test JSON deserialization
	var unmarshaled ModuleSearchResponse
	err = json.Unmarshal(jsonData, &unmarshaled)
	require.NoError(t, err)
	assert.Equal(t, response, unmarshaled)

	// Verify pagination fields
	assert.Equal(t, 20, unmarshaled.Meta.Limit)
	assert.Equal(t, 0, unmarshaled.Meta.Offset)
	assert.Equal(t, 2, unmarshaled.Meta.TotalCount)
	assert.Len(t, unmarshaled.Modules, 1)
}

func TestTerraregModuleVersionResponse_UIFields(t *testing.T) {
	response := TerraregModuleVersionResponse{
		TerraregVersionDetails: TerraregVersionDetails{
			VersionDetails: VersionDetails{
				VersionBase: VersionBase{
					ProviderBase: ProviderBase{
						ID:        "example/aws/2.0.0",
						Namespace: "example",
						Name:      "aws",
						Provider:  "aws",
						Verified:  true,
						Trusted:   false,
					},
					Version: "2.0.0",
				},
			},
			Beta:                       false,
			Published:                  true,
			TerraformVersionConstraint: &[]string{">= 0.14"}[0],
			SecurityFailures:           1,
			DisplaySourceURL:           &[]string{"https://github.com/example/aws"}[0],
			VersionCompatibility:       &[]string{"Fully Compatible"}[0],
		},
	}

	// Test JSON serialization
	jsonData, err := json.Marshal(response)
	require.NoError(t, err)
	assert.Contains(t, string(jsonData), `"beta":false`)
	assert.Contains(t, string(jsonData), `"published":true`)
	assert.Contains(t, string(jsonData), `"terraform_version_constraint":`)
	assert.Contains(t, string(jsonData), `"security_failures":1`)
	assert.Contains(t, string(jsonData), `"version_compatibility":"Fully Compatible"`)
}

func TestModuleSpecs_TerraformFields(t *testing.T) {
	specs := ModuleSpecs{
		Path:   "submodules/example",
		Readme: "# Example Submodule\n\nThis is an example submodule.",
		Empty:  false,
		Inputs: []TerraformInput{
			{
				Name:        "name",
				Type:        "string",
				Description: &[]string{"Name of the resource"}[0],
				Required:    true,
			},
			{
				Name:        "tags",
				Type:        "map(string)",
				Description: &[]string{"Tags to apply"}[0],
				Default:     map[string]string{"Environment": "dev"},
				Required:    false,
			},
		},
		Outputs: []TerraformOutput{
			{
				Name:        "id",
				Description: &[]string{"ID of the created resource"}[0],
			},
		},
		Dependencies: []TerraformDependency{
			{
				Name:    "example-network",
				Source:  "example/network",
				Version: ">= 1.0",
			},
		},
		ProviderDependencies: []TerraformProvider{
			{
				Name:    "aws",
				Version: ">= 4.0",
				Source:  "hashicorp/aws",
				Configuration: map[string]interface{}{
					"region": "us-west-2",
				},
			},
		},
		Resources: []TerraformResource{
			{
				Type:     "aws_s3_bucket",
				Name:     "example",
				Provider: "aws",
				Mode:     "managed",
				Version:  "4.0",
				Config: map[string]interface{}{
					"bucket_prefix": "example-",
					"force_destroy": true,
				},
			},
		},
		Modules: []TerraformModule{
			{
				Name:    "example-submodule",
				Source:  "./submodule",
				Version: "1.0.0",
			},
		},
	}

	// Test JSON serialization
	jsonData, err := json.Marshal(specs)
	require.NoError(t, err)
	assert.Contains(t, string(jsonData), `"path":"submodules/example"`)
	assert.Contains(t, string(jsonData), `"name":"name"`)
	assert.Contains(t, string(jsonData), `"type":"string"`)
	assert.Contains(t, string(jsonData), `"required":true`)
}

func TestCostAnalysis(t *testing.T) {
	tests := []struct {
		name     string
		analysis CostAnalysis
		expected string
	}{
		{
			name: "Complete cost analysis",
			analysis: CostAnalysis{
				Monthly:  123.45,
				Hourly:   0.17,
				Currency: "USD",
			},
			expected: `{"monthly":123.45,"hourly":0.17,"currency":"USD"}`,
		},
		{
			name: "Zero cost analysis",
			analysis: CostAnalysis{
				Currency: "USD",
			},
			expected: `{"monthly":0,"hourly":0,"currency":"USD"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test JSON serialization
			jsonData, err := json.Marshal(tt.analysis)
			require.NoError(t, err)
			assert.JSONEq(t, tt.expected, string(jsonData))

			// Test JSON deserialization
			var unmarshaled CostAnalysis
			err = json.Unmarshal(jsonData, &unmarshaled)
			require.NoError(t, err)
			assert.Equal(t, tt.analysis, unmarshaled)
		})
	}
}

func TestModuleVersionPublishRequest(t *testing.T) {
	tests := []struct {
		name     string
		request  ModuleVersionPublishRequest
		expected string
	}{
		{
			name: "Complete request",
			request: ModuleVersionPublishRequest{
				Version:     "1.0.0",
				Beta:        false,
				Description: &[]string{"Initial release"}[0],
				Owner:       &[]string{"example-team"}[0],
			},
			expected: `{"version":"1.0.0","beta":false,"description":"Initial release","owner":"example-team"}`,
		},
		{
			name: "Beta release",
			request: ModuleVersionPublishRequest{
				Version: "2.0.0-beta",
				Beta:    true,
				Owner:   &[]string{"dev-team"}[0],
			},
			expected: `{"version":"2.0.0-beta","beta":true,"description":null,"owner":"dev-team"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test JSON serialization
			jsonData, err := json.Marshal(tt.request)
			require.NoError(t, err)
			assert.JSONEq(t, tt.expected, string(jsonData))

			// Test JSON deserialization
			var unmarshaled ModuleVersionPublishRequest
			err = json.Unmarshal(jsonData, &unmarshaled)
			require.NoError(t, err)
			assert.Equal(t, tt.request, unmarshaled)
		})
	}
}
