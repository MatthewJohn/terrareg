package terrareg

import "github.com/matthewjohn/terrareg/terrareg-go/internal/interfaces/http/dto"

// TerraregModuleProviderDetailsResponse represents the complete response for
// the terrareg-specific module provider details endpoint.
// This matches Python's ModuleVersion.get_terrareg_api_details() exactly.
type TerraregModuleProviderDetailsResponse struct {
	// Base provider info (from ModuleProvider.get_terrareg_api_details)
	ID                    string   `json:"id"`
	Namespace             string   `json:"namespace"`
	Name                  string   `json:"name"` // Python uses "name" not "module"
	Provider              string   `json:"provider"`
	Verified              bool     `json:"verified"`
	Trusted               bool     `json:"trusted"`
	ModuleProviderID      string   `json:"module_provider_id"`
	GitProviderID         *int     `json:"git_provider_id"`
	GitTagFormat          *string  `json:"git_tag_format"`
	GitPath               *string  `json:"git_path"`
	ArchiveGitPath        bool     `json:"archive_git_path"`
	RepoBaseURLTemplate   *string  `json:"repo_base_url_template"`
	RepoCloneURLTemplate  *string  `json:"repo_clone_url_template"`
	RepoBrowseURLTemplate *string  `json:"repo_browse_url_template"`
	Versions              []string `json:"versions"` // Limited versions list

	// Module version details (from ModuleVersion.get_api_details)
	Owner       *string `json:"owner"`
	Version     string  `json:"version"`
	Description *string `json:"description"`
	Source      *string `json:"source"`
	PublishedAt *string `json:"published_at"`
	Downloads   int     `json:"downloads"`
	Internal    bool    `json:"internal"`

	// Terraform module specifications (from ModuleVersion.get_api_module_specs)
	Root       TerraregModuleSpecs   `json:"root"`
	Submodules []TerraregModuleSpecs `json:"submodules"`
	Providers  []string              `json:"providers"`

	// UI-specific terrareg fields (additional fields not in standard API)
	PublishedAtDisplay             *string                  `json:"published_at_display"`
	DisplaySourceURL               *string                  `json:"display_source_url"`
	TerraformExampleVersionString  *string                  `json:"terraform_example_version_string"`
	TerraformExampleVersionComment []string                 `json:"terraform_example_version_comment"`
	SecurityFailures               int                      `json:"security_failures"`
	SecurityResults                []TerraregSecurityResult `json:"security_results"`
	Beta                           bool                     `json:"beta"`
	Published                      bool                     `json:"published"`
	AdditionalTabFiles             map[string]string        `json:"additional_tab_files"`
	CustomLinks                    []TerraregCustomLink     `json:"custom_links"`
	GraphURL                       *string                  `json:"graph_url"`
	TerraformVersionConstraint     *string                  `json:"terraform_version_constraint"`
	ModuleExtractionUpToDate       bool                     `json:"module_extraction_up_to_date"`
	UsageExample                   *string                  `json:"usage_example"`

	// Analytics token (extracted from namespace)
	AnalyticsToken *string `json:"analytics_token,omitempty"`
}

// PaginationResponse wraps terrareg responses with pagination metadata
type PaginationResponse struct {
	Meta dto.PaginationMeta `json:"meta"`
}

// ListResponse represents a paginated list response
type ListResponse[T any] struct {
	Data T                  `json:"data"`
	Meta dto.PaginationMeta `json:"meta"`
}
