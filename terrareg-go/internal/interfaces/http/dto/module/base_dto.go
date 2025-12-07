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
	Versions []string `json:"versions,omitempty"` // List of version strings
}

// TerraregProviderDetails extends ProviderDetails with UI-specific fields
// Similar to Python's get_terrareg_api_details()
type TerraregProviderDetails struct {
	ProviderDetails
	ModuleProviderID     *int    `json:"module_provider_id,omitempty"`
	GitProviderID        *int    `json:"git_provider_id,omitempty"`
	GitTagFormat         *string `json:"git_tag_format,omitempty"`
	GitPath              *string `json:"git_path,omitempty"`
	ArchiveGitPath       bool    `json:"archive_git_path,omitempty"`
	RepoBaseURLTemplate  *string `json:"repo_base_url_template,omitempty"`
	RepoCloneURLTemplate *string `json:"repo_clone_url_template,omitempty"`
	RepoBrowseURLTemplate *string `json:"repo_browse_url_template,omitempty"`
}

// VersionBase extends ProviderBase with version-specific fields
// Matches Python's ModuleVersion.get_api_outline() additions
type VersionBase struct {
	ProviderBase
	Version     string  `json:"version"`      // Version string
	Owner       *string `json:"owner,omitempty"`
	Description *string `json:"description,omitempty"`
	Source      *string `json:"source,omitempty"`      // From get_source_base_url()
	PublishedAt *string `json:"published_at,omitempty"` // ISO format from .isoformat()
	Downloads   int     `json:"downloads"`
	Internal    bool    `json:"internal"`
}

// VersionDetails extends VersionBase with Terraform specifications
// Similar to Python's get_api_details()
type VersionDetails struct {
	VersionBase
	Root       ModuleSpecs `json:"root"`                    // Root module specs
	Submodules []ModuleSpecs `json:"submodules,omitempty"` // Submodule specs
	Providers  []string     `json:"providers,omitempty"`    // List of provider names
}

// TerraregVersionDetails extends VersionDetails with UI-specific fields
// Similar to Python's get_terrareg_api_details()
type TerraregVersionDetails struct {
	VersionDetails
	Beta                      bool        `json:"beta"`
	Published                 bool        `json:"published"`
	TerraformVersionConstraint *string   `json:"terraform_version_constraint,omitempty"`
	SecurityFailures          int         `json:"security_failures"`
	SecurityResults           interface{} `json:"security_results,omitempty"`
	DisplaySourceURL          *string     `json:"display_source_url,omitempty"`
	GraphURL                  *string     `json:"graph_url,omitempty"`
	UsageExample              *string     `json:"usage_example,omitempty"`
	VersionCompatibility       *string     `json:"version_compatibility,omitempty"` // Optional Terraform compatibility
}

// ModuleSpecs contains Terraform module specification data
// Matches Python's get_api_module_specs()
type ModuleSpecs struct {
	Path                 string                `json:"path"`
	Readme               string                `json:"readme"`
	Empty                bool                  `json:"empty"`
	Inputs               []TerraformInput      `json:"inputs,omitempty"`
	Outputs              []TerraformOutput     `json:"outputs,omitempty"`
	Dependencies         []TerraformDependency `json:"dependencies,omitempty"`
	ProviderDependencies []TerraformProvider   `json:"provider_dependencies,omitempty"`
	Resources            []TerraformResource   `json:"resources,omitempty"`
	Modules              []TerraformModule     `json:"modules,omitempty"`
}

// Terraform specification types
type TerraformInput struct {
	Name        string      `json:"name"`
	Type        string      `json:"type"`
	Description *string     `json:"description,omitempty"`
	Default     interface{} `json:"default,omitempty"`
	Required    bool        `json:"required"`
}

type TerraformOutput struct {
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`
}

type TerraformDependency struct {
	Name    string `json:"name"`
	Source  string `json:"source"`
	Version string `json:"version,omitempty"`
}

type TerraformProvider struct {
	Name         string                 `json:"name"`
	Version      string                 `json:"version,omitempty"`
	Source       string                 `json:"source,omitempty"`
	Configuration map[string]interface{} `json:"configuration,omitempty"`
}

type TerraformResource struct {
	Type     string                 `json:"type"`
	Name     string                 `json:"name"`
	Provider string                 `json:"provider"`
	Mode     string                 `json:"mode"` // "managed" or "data"
	Version  string                 `json:"version,omitempty"`
	Config   map[string]interface{} `json:"config,omitempty"`
}

type TerraformModule struct {
	Name    string `json:"name"`
	Source  string `json:"source"`
	Version string `json:"version,omitempty"`
}