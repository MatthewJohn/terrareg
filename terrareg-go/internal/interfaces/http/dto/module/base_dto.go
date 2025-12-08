package module

// ProviderBase contains the 6 core fields from Python's ModuleProvider.get_api_outline()
type ProviderBase struct {
	ID        string `json:"id"`        // Provider ID
	Namespace string `json:"namespace"` // Namespace name
	Name      string `json:"name"`      // Module name (Python uses "name" not "module")
	Provider  string `json:"provider"`  // Provider name
	Verified  bool   `json:"verified"`  // Provider verification status
	Trusted   bool   `json:"trusted"`   // Namespace trusted status (computed)
}

// ProviderDetails extends ProviderBase with additional provider information
// Similar to Python's get_api_details()
type ProviderDetails struct {
	ProviderBase
	Versions []string `json:"versions"` // List of version strings
}

// TerraregProviderDetails extends ProviderDetails with UI-specific fields
// Similar to Python's get_terrareg_api_details()
type TerraregProviderDetails struct {
	ProviderDetails
	ModuleProviderID      *int    `json:"module_provider_id"`
	GitProviderID         *int    `json:"git_provider_id"`
	GitTagFormat          *string `json:"git_tag_format"`
	GitPath               *string `json:"git_path"`
	ArchiveGitPath        bool    `json:"archive_git_path"`
	RepoBaseURLTemplate   *string `json:"repo_base_url_template"`
	RepoCloneURLTemplate  *string `json:"repo_clone_url_template"`
	RepoBrowseURLTemplate *string `json:"repo_browse_url_template"`
}

// VersionBase extends ProviderBase with version-specific fields
// Matches Python's ModuleVersion.get_api_outline() additions
type VersionBase struct {
	ProviderBase
	Version     string  `json:"version"` // Version string
	Owner       *string `json:"owner"`
	Description *string `json:"description"`
	Source      *string `json:"source"`       // From get_source_base_url()
	PublishedAt *string `json:"published_at"` // ISO format from .isoformat()
	Downloads   int     `json:"downloads"`
	Internal    bool    `json:"internal"`
}

// VersionDetails extends VersionBase with Terraform specifications
// Similar to Python's get_api_details()
type VersionDetails struct {
	VersionBase
	Root       ModuleSpecs   `json:"root"`       // Root module specs
	Submodules []ModuleSpecs `json:"submodules"` // Submodule specs
	Providers  []string      `json:"providers"`  // List of provider names
}

// TerraregVersionDetails extends VersionDetails with UI-specific fields
// Similar to Python's get_terrareg_api_details()
type TerraregVersionDetails struct {
	VersionDetails
	TerraregProviderDetails
	Beta                       bool        `json:"beta"`
	Published                  bool        `json:"published"`
	TerraformVersionConstraint *string     `json:"terraform_version_constraint"`
	SecurityFailures           int         `json:"security_failures"`
	SecurityResults            interface{} `json:"security_results"`
	DisplaySourceURL           *string     `json:"display_source_url"`
	GraphURL                   *string     `json:"graph_url"`
	UsageExample               *string     `json:"usage_example"`
	VersionCompatibility       *string     `json:"version_compatibility"` // Optional Terraform compatibility
}

// ModuleSpecs contains Terraform module specification data
// Matches Python's get_api_module_specs() - NOTE: No 'modules' field in standard API
type ModuleSpecs struct {
	Path                 string                `json:"path"`
	Readme               string                `json:"readme"`
	Empty                bool                  `json:"empty"`
	Inputs               []TerraformInput      `json:"inputs"`
	Outputs              []TerraformOutput     `json:"outputs"`
	Dependencies         []TerraformDependency `json:"dependencies"`
	ProviderDependencies []TerraformProvider   `json:"provider_dependencies"`
	Resources            []TerraformResource   `json:"resources"`
}

// Terraform specification types
type TerraformInput struct {
	Name        string      `json:"name"`
	Type        string      `json:"type"`
	Description *string     `json:"description"`
	Default     interface{} `json:"default"`
	Required    bool        `json:"required"`
}

type TerraformOutput struct {
	Name        string  `json:"name"`
	Description *string `json:"description"`
}

type TerraformDependency struct {
	Module  string `json:"module"`  // Python uses "module" not "name"
	Source  string `json:"source"`
	Version string `json:"version"`
}

type TerraformProvider struct {
	Name    string `json:"name"`    // Provider name from get_providers()
	// Note: Standard API only returns provider names, not full provider details
}

type TerraformResource struct {
	Name string `json:"name"` // Resource name
	Type string `json:"type"` // Resource type
	// Note: Standard API only returns name and type
}

// TerraformModule is only used in terrareg API, not standard API
// See terrareg.TerraregModule for the terrareg-specific implementation
