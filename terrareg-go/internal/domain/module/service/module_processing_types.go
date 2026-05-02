package service

import "context"

// ModuleProcessorService defines the interface for processing module directories
// and extracting metadata, submodules, and examples.
type ModuleProcessorService interface {
	// ProcessModule processes a module directory and extracts all metadata
	ProcessModule(ctx context.Context, moduleDir string, metadata *ModuleProcessingMetadata) (*ModuleProcessingResult, error)

	// ValidateModuleStructure validates that a module directory has the required structure
	ValidateModuleStructure(ctx context.Context, moduleDir string) error

	// ExtractMetadata extracts metadata from a module directory
	ExtractMetadata(ctx context.Context, moduleDir string) (*ModuleMetadata, error)
}

// ModuleProcessingMetadata contains metadata about the module being processed
type ModuleProcessingMetadata struct {
	ModuleVersionID int    // Database ID of the module version
	GitTag          string // Git tag associated with this version
	GitURL          string // Git repository URL
	GitPath         string // Path to module in git repository
	CommitSHA       string // Git commit SHA
}

// ModuleProcessingResult contains the results of processing a module
type ModuleProcessingResult struct {
	ModuleMetadata   *ModuleMetadata // Extracted module metadata
	Submodules       []SubmoduleInfo // Detected and processed submodules
	Examples         []ExampleInfo   // Detected and processed examples
	ReadmeContent    string          // README file content
	VariableTemplate string          // JSON template of variables
	ProcessedFiles   []string        // List of processed files
}

// ModuleMetadata contains comprehensive metadata about a module
type ModuleMetadata struct {
	Name         string           // Module name
	Description  string           // Module description
	Version      string           // Module version
	Providers    []ProviderInfo   // Required providers
	Variables    []VariableInfo   // Input variables
	Outputs      []OutputInfo     // Output values
	Resources    []ResourceInfo   // Resources used in module
	Dependencies []DependencyInfo // Module dependencies
}

// VariableInfo represents a Terraform input variable
type VariableInfo struct {
	Name        string      // Variable name
	Type        string      // Variable type
	Description string      // Variable description
	Default     interface{} // Default value (can be nil)
	Required    bool        // Whether variable is required
}

// OutputInfo represents a Terraform output value
type OutputInfo struct {
	Name        string      // Output name
	Description string      // Output description
	Value       interface{} // Output value (nil during extraction)
	Sensitive   bool        // Whether output is sensitive
}

// ProviderInfo represents a Terraform provider requirement
type ProviderInfo struct {
	Name    string // Provider name
	Version string // Provider version constraint
	Source  string // Provider source (e.g., hashicorp/aws)
}

// ResourceInfo represents a Terraform resource used in the module
type ResourceInfo struct {
	Type string // Resource type (e.g., aws_s3_bucket)
	Name string // Resource name in configuration
}

// DependencyInfo represents a module dependency
type DependencyInfo struct {
	Source  string // Dependency source
	Version string // Dependency version constraint
}

// SubmoduleInfo represents a detected submodule
// Python: terrareg.models.Submodule (partial - path only, source/version are not applicable for submodules)
type SubmoduleInfo struct {
	Path string // Submodule path (relative to module root)
}

// ExampleInfo represents a detected example
type ExampleInfo struct {
	Name        string   // Example name
	Description string   // Example description
	Files       []string // List of files in the example
}
