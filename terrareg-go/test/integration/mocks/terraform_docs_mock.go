package mocks

import (
	"encoding/json"
	"fmt"
)

// TerraformDocsMock provides a mock implementation of terraform-docs output
// This is used for testing module extraction without requiring the actual terraform-docs binary
type TerraformDocsMock struct {
	ExpectedOutput *TerraformDocsOutput
	ShouldError    bool
	ErrorMessage   string
}

// TerraformDocsOutput represents the structured output from terraform-docs
type TerraformDocsOutput struct {
	Header       string            `json:"header,omitempty"`
	Inputs       []VariableDoc     `json:"inputs"`
	Outputs      []OutputDoc       `json:"outputs"`
	Providers    []ProviderDoc     `json:"providers,omitempty"`
	Resources    []ResourceDoc     `json:"resources,omitempty"`
	Modules      []ModuleDoc       `json:"modules,omitempty"`
	Requirements []RequirementDoc  `json:"requirements,omitempty"`
}

// VariableDoc represents a terraform input variable from terraform-docs
type VariableDoc struct {
	Name        string      `json:"name"`
	Type        string      `json:"type"`
	Description string      `json:"description,omitempty"`
	Default     interface{} `json:"default,omitempty"`
	Required    bool        `json:"required"`
	Position     PositionDoc `json:"position,omitempty"`
}

// OutputDoc represents a terraform output from terraform-docs
type OutputDoc struct {
	Name        string      `json:"name"`
	Description string      `json:"description,omitempty"`
	Sensitive   bool        `json:"sensitive,omitempty"`
	Position    PositionDoc `json:"position,omitempty"`
}

// ProviderDoc represents a terraform provider from terraform-docs
type ProviderDoc struct {
	Name    string `json:"name"`
	Alias   string `json:"alias,omitempty"`
	Version string `json:"version,omitempty"`
}

// ResourceDoc represents a terraform resource from terraform-docs
type ResourceDoc struct {
	Type string `json:"type"`
	Name string `json:"name"`
}

// ModuleDoc represents a terraform module from terraform-docs
type ModuleDoc struct {
	Name    string `json:"name"`
	Source  string `json:"source,omitempty"`
}

// RequirementDoc represents a terraform requirement from terraform-docs
type RequirementDoc struct {
	Name    string `json:"name"`
	Version string `json:"version,omitempty"`
}

// PositionDoc represents the position of an element in source code
type PositionDoc struct {
	Filename string `json:"filename,omitempty"`
	Line     int    `json:"line,omitempty"`
}

// Generate returns the terraform-docs output as JSON or an error
func (m *TerraformDocsMock) Generate() ([]byte, error) {
	if m.ShouldError {
		if m.ErrorMessage != "" {
			return nil, fmt.Errorf("terraform-docs mock error: %s", m.ErrorMessage)
		}
		return nil, fmt.Errorf("terraform-docs mock error")
	}

	// If no expected output is set, return a default empty output
	if m.ExpectedOutput == nil {
		m.ExpectedOutput = &TerraformDocsOutput{
			Inputs:  []VariableDoc{},
			Outputs: []OutputDoc{},
		}
	}

	jsonOutput, err := json.MarshalIndent(m.ExpectedOutput, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal terraform-docs output: %w", err)
	}

	return jsonOutput, nil
}

// SetOutput sets the expected output for the mock
func (m *TerraformDocsMock) SetOutput(output *TerraformDocsOutput) {
	m.ExpectedOutput = output
}

// SetError configures the mock to return an error
func (m *TerraformDocsMock) SetError(message string) {
	m.ShouldError = true
	m.ErrorMessage = message
}

// ClearError clears the error state
func (m *TerraformDocsMock) ClearError() {
	m.ShouldError = false
	m.ErrorMessage = ""
}

// NewTerraformDocsMock creates a new mock with default test data
func NewTerraformDocsMock() *TerraformDocsMock {
	return &TerraformDocsMock{
		ExpectedOutput: &TerraformDocsOutput{
			Inputs: []VariableDoc{
				{
					Name:        "instance_type",
					Type:        "string",
					Description: "The type of instance to create",
					Default:     "t2.micro",
					Required:    false,
				},
				{
					Name:        "ami_id",
					Type:        "string",
					Description: "The AMI ID to use",
					Required:    true,
				},
			},
			Outputs: []OutputDoc{
				{
					Name:        "instance_id",
					Description: "The ID of the instance",
				},
			},
			Providers: []ProviderDoc{
				{
					Name:    "aws",
					Version: ">= 4.0",
				},
			},
		},
		ShouldError: false,
	}
}

// NewEmptyTerraformDocsMock creates a mock with empty outputs
func NewEmptyTerraformDocsMock() *TerraformDocsMock {
	return &TerraformDocsMock{
		ExpectedOutput: &TerraformDocsOutput{
			Inputs:   []VariableDoc{},
			Outputs:  []OutputDoc{},
			Providers: []ProviderDoc{},
		},
		ShouldError: false,
	}
}

// NewErrorTerraformDocsMock creates a mock that always returns an error
func NewErrorTerraformDocsMock(message string) *TerraformDocsMock {
	return &TerraformDocsMock{
		ShouldError:  true,
		ErrorMessage: message,
	}
}

// Helper functions to create common test outputs

// CreateSimpleModuleDocs creates terraform-docs output for a simple module
func CreateSimpleModuleDocs() *TerraformDocsOutput {
	return &TerraformDocsOutput{
		Inputs: []VariableDoc{
			{
				Name:        "name",
				Type:        "string",
				Description: "The name for the resource",
				Required:    true,
			},
		},
		Outputs: []OutputDoc{
			{
				Name:        "id",
				Description: "The ID of the created resource",
			},
		},
		Providers: []ProviderDoc{
			{
				Name: "aws",
			},
		},
	}
}

// CreateComplexModuleDocs creates terraform-docs output for a complex module
func CreateComplexModuleDocs() *TerraformDocsOutput {
	return &TerraformDocsOutput{
		Inputs: []VariableDoc{
			{
				Name:        "instance_type",
				Type:        "string",
				Description: "The type of instance to create",
				Default:     "t2.micro",
				Required:    false,
			},
			{
				Name:        "ami_id",
				Type:        "string",
				Description: "The AMI ID to use for the instance",
				Required:    true,
			},
			{
				Name:        "vpc_id",
				Type:        "string",
				Description: "The VPC ID where the instance will be created",
				Required:    true,
			},
			{
				Name:        "subnet_ids",
				Type:        "list(string)",
				Description: "List of subnet IDs for the instance",
				Required:    true,
			},
			{
				Name:        "enable_monitoring",
				Type:        "bool",
				Description: "Enable detailed monitoring",
				Default:     true,
				Required:    false,
			},
			{
				Name:        "tags",
				Type:        "map(string)",
				Description: "Tags to apply to the instance",
				Default:     map[string]string{},
				Required:    false,
			},
		},
		Outputs: []OutputDoc{
			{
				Name:        "instance_id",
				Description: "The ID of the created EC2 instance",
			},
			{
				Name:        "private_ip",
				Description: "The private IP address of the instance",
			},
			{
				Name:        "public_ip",
				Description: "The public IP address of the instance (if applicable)",
			},
		},
		Providers: []ProviderDoc{
			{
				Name:    "aws",
				Version: ">= 4.0",
			},
		},
		Resources: []ResourceDoc{
			{
				Type: "aws_instance",
				Name: "this",
			},
			{
				Type: "aws_security_group",
				Name: "this",
			},
		},
	}
}

// CreateModuleWithSubmodulesDocs creates terraform-docs output including submodules
func CreateModuleWithSubmodulesDocs() *TerraformDocsOutput {
	return &TerraformDocsOutput{
		Inputs: []VariableDoc{
			{
				Name:        "region",
				Type:        "string",
				Description: "AWS region",
				Default:     "us-west-2",
				Required:    false,
			},
		},
		Outputs: []OutputDoc{
			{
				Name:        "vpc_id",
				Description: "VPC ID",
			},
		},
		Modules: []ModuleDoc{
			{
				Name:   "vpc",
				Source: "./modules/vpc",
			},
			{
				Name:   "subnet",
				Source: "./modules/subnet",
			},
		},
	}
}
