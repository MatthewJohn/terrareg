package terrareg

import (
	"encoding/json"
	"testing"
)

func TestTerraregModuleProviderDetailsResponse_MarshalJSON(t *testing.T) {
	// Test case: Full populated response
	response := &TerraregModuleProviderDetailsResponse{
		ID:          "terraform-aws-modules/vpc/aws",
		Namespace:   "terraform-aws-modules",
		Name:        "vpc",
		Provider:    "aws",
		Verified:    true,
		Trusted:     false,
		Owner:       stringPtr("terraform-aws-modules"),
		Version:     "3.19.0",
		Description: stringPtr("Terraform AWS VPC module"),
		Internal:    false,
		Published:   true,
		Beta:        false,
		Versions:    []string{"3.19.0", "3.18.0", "3.17.0"},
		Root: TerraregModuleSpecs{
			Path:   "",
			Readme: "<h1>VPC Module</h1>\n<p>Creates VPC resources...</p>",
			Empty:  false,
			Inputs: []TerraregInput{
				{
					Name:        "name",
					Type:        "string",
					Description: stringPtr("Name to be used on all resources"),
					Required:    true,
					Default:     nil,
				},
			},
			Outputs: []TerraregOutput{
				{
					Name:        "vpc_id",
					Description: stringPtr("The ID of the VPC"),
				},
			},
		},
		Submodules: []TerraregModuleSpecs{},
		Providers:  []string{"aws"},
		CustomLinks: []TerraregCustomLink{
			{
				Text: "Documentation",
				URL:  "https://registry.terraform.io/terraform-aws-modules/vpc/latest",
			},
		},
	}

	// Marshal to JSON
	data, err := json.Marshal(response)
	if err != nil {
		t.Fatalf("Failed to marshal TerraregModuleProviderDetailsResponse: %v", err)
	}

	// Verify it produces valid JSON
	var unmarshaled map[string]interface{}
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("Failed to unmarshal generated JSON: %v", err)
	}

	// Check required fields exist
	expectedFields := []string{
		"id", "namespace", "name", "provider", "verified", "trusted",
		"owner", "version", "description", "internal", "published", "beta",
		"versions", "root", "submodules", "providers", "custom_links",
	}

	for _, field := range expectedFields {
		if _, exists := unmarshaled[field]; !exists {
			t.Errorf("Expected field '%s' to be present in JSON", field)
		}
	}
}

func TestTerraregModuleProviderDetailsResponse_WithNilFields(t *testing.T) {
	// Test case: Minimal response with many nil fields
	response := &TerraregModuleProviderDetailsResponse{
		ID:        "test/module/provider",
		Namespace: "test",
		Name:      "module",
		Provider:  "provider",
		Verified:  false,
		Trusted:   false,
		Version:  "1.0.0",
		Internal: false,
		Published: false,
		Beta:     false,
		// All optional fields left as nil/empty
	}

	// Marshal to JSON
	data, err := json.Marshal(response)
	if err != nil {
		t.Fatalf("Failed to marshal TerraregModuleProviderDetailsResponse: %v", err)
	}

	// Unmarshal and check that nil fields are handled properly
	var unmarshaled TerraregModuleProviderDetailsResponse
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("Failed to unmarshal generated JSON: %v", err)
	}

	// Verify that optional fields are nil when not set
	if unmarshaled.Owner != nil {
		t.Errorf("Expected Owner to be nil, got %v", unmarshaled.Owner)
	}
	if unmarshaled.Description != nil {
		t.Errorf("Expected Description to be nil, got %v", unmarshaled.Description)
	}
}

func TestTerraregModuleProviderDetailsResponse_WithAnalyticsToken(t *testing.T) {
	// Test case: Response with analytics token
	token := "analytics-token-123"
	response := &TerraregModuleProviderDetailsResponse{
		ID:             "namespace__token/module/provider",
		Namespace:      "namespace__token",
		Name:           "module",
		Provider:       "provider",
		Verified:       false,
		Trusted:        false,
		Version:        "1.0.0",
		Internal:       false,
		Published:      false,
		Beta:           false,
		AnalyticsToken: &token,
	}

	// Marshal to JSON
	data, err := json.Marshal(response)
	if err != nil {
		t.Fatalf("Failed to marshal TerraregModuleProviderDetailsResponse: %v", err)
	}

	// Unmarshal and check analytics token
	var unmarshaled TerraregModuleProviderDetailsResponse
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("Failed to unmarshal generated JSON: %v", err)
	}

	if unmarshaled.AnalyticsToken == nil {
		t.Error("Expected AnalyticsToken to be non-nil")
	} else if *unmarshaled.AnalyticsToken != token {
		t.Errorf("Expected AnalyticsToken to be %s, got %s", token, *unmarshaled.AnalyticsToken)
	}
}

func TestTerraregModuleSpecs_JSON(t *testing.T) {
	// Test TerraregModuleSpecs JSON serialization
	specs := TerraregModuleSpecs{
		Path:   "submodule/path",
		Readme: "Submodule readme",
		Empty:  false,
		Inputs: []TerraregInput{
			{
				Name:           "subnet_cidr",
				Type:           "string",
				Description:    stringPtr("CIDR block for subnet"),
				Required:       true,
				Default:        nil,
				AdditionalHelp: stringPtr("Must be a valid CIDR block"),
				QuoteValue:     true,
				Sensitive:      false,
			},
		},
		Outputs: []TerraregOutput{
			{
				Name:        "subnet_id",
				Description: stringPtr("ID of the created subnet"),
				Type:        stringPtr("string"),
			},
		},
		Dependencies: []TerraregDependency{
			{
				Module:  "vpc",
				Source:  "terraform-aws-modules/vpc/aws",
				Version: ">= 3.0",
			},
		},
		Modules: []TerraregModule{
			{
				Name:      "security_group",
				Source:    "./security-group",
				Version:   "1.0.0",
				Key:       "sg",
				Providers: []string{"aws"},
			},
		},
	}

	// Marshal and unmarshal
	data, err := json.Marshal(specs)
	if err != nil {
		t.Fatalf("Failed to marshal TerraregModuleSpecs: %v", err)
	}

	var unmarshaled TerraregModuleSpecs
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("Failed to unmarshal TerraregModuleSpecs: %v", err)
	}

	// Verify key fields
	if unmarshaled.Path != specs.Path {
		t.Errorf("Expected Path to be %s, got %s", specs.Path, unmarshaled.Path)
	}
	if len(unmarshaled.Modules) != len(specs.Modules) {
		t.Errorf("Expected %d modules, got %d", len(specs.Modules), len(unmarshaled.Modules))
	}
}

func TestTerraregSecurityResult_JSON(t *testing.T) {
	// Test TerraregSecurityResult JSON serialization
	result := TerraregSecurityResult{
		RuleID:      "AWS009",
		Severity:    "HIGH",
		Title:       "resource should not use plaintext credentials",
		Description: "The resource contains plaintext credentials",
		Location: TerraregSecurityLocation{
			Filename:  "main.tf",
			StartLine: 10,
			EndLine:   15,
		},
	}

	// Marshal and unmarshal
	data, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("Failed to marshal TerraregSecurityResult: %v", err)
	}

	var unmarshaled TerraregSecurityResult
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("Failed to unmarshal TerraregSecurityResult: %v", err)
	}

	// Verify key fields
	if unmarshaled.RuleID != result.RuleID {
		t.Errorf("Expected RuleID to be %s, got %s", result.RuleID, unmarshaled.RuleID)
	}
	if unmarshaled.Severity != result.Severity {
		t.Errorf("Expected Severity to be %s, got %s", result.Severity, unmarshaled.Severity)
	}
	if unmarshaled.Location.StartLine != result.Location.StartLine {
		t.Errorf("Expected StartLine to be %d, got %d", result.Location.StartLine, unmarshaled.Location.StartLine)
	}
}

func TestTerraregCustomLink_JSON(t *testing.T) {
	// Test TerraregCustomLink JSON serialization
	link := TerraregCustomLink{
		Text: "GitHub Repository",
		URL:  "https://github.com/example/repo",
	}

	// Marshal and unmarshal
	data, err := json.Marshal(link)
	if err != nil {
		t.Fatalf("Failed to marshal TerraregCustomLink: %v", err)
	}

	var unmarshaled TerraregCustomLink
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("Failed to unmarshal TerraregCustomLink: %v", err)
	}

	// Verify key fields
	if unmarshaled.Text != link.Text {
		t.Errorf("Expected Text to be %s, got %s", link.Text, unmarshaled.Text)
	}
	if unmarshaled.URL != link.URL {
		t.Errorf("Expected URL to be %s, got %s", link.URL, unmarshaled.URL)
	}
}

// Helper function to create string pointers for test data
func stringPtr(s string) *string {
	return &s
}