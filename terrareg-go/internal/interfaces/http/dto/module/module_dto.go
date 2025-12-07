package module

import "github.com/matthewjohn/terrareg/terrareg-go/internal/interfaces/http/dto"

// ModuleProviderResponse for Terraform registry API (equivalent to get_api_outline)
type ModuleProviderResponse struct {
	ProviderBase
	Description *string `json:"description"`
	Owner       *string `json:"owner"`
	Source      *string `json:"source"`
	PublishedAt *string `json:"published_at"`
	Downloads   int     `json:"downloads"`
}

// ModuleProviderDetailsResponse for detailed provider view (equivalent to get_api_details)
type ModuleProviderDetailsResponse struct {
	ProviderDetails
	Description *string `json:"description"`
	Owner       *string `json:"owner"`
	Source      *string `json:"source"`
	PublishedAt *string `json:"published_at"`
	Downloads   int     `json:"downloads"`
}

// TerraregModuleProviderResponse for terrareg UI (equivalent to get_terrareg_api_details)
type TerraregModuleProviderResponse struct {
	TerraregProviderDetails
	Description *string `json:"description"`
	Owner       *string `json:"owner"`
	Source      *string `json:"source"`
	PublishedAt *string `json:"published_at"`
	Downloads   int     `json:"downloads"`
}

// ModuleVersionResponse for Terraform registry API (equivalent to get_api_outline)
type ModuleVersionResponse struct {
	VersionBase
}

// ModuleVersionDetailsResponse for detailed version view (equivalent to get_api_details)
type ModuleVersionDetailsResponse struct {
	VersionDetails
}

// TerraregModuleVersionResponse for terrareg UI (equivalent to get_terrareg_api_details)
type TerraregModuleVersionResponse struct {
	TerraregVersionDetails
}

// SubmoduleResponse for submodules (uses ModuleSpecs)
type SubmoduleResponse struct {
	ModuleSpecs
	// Add any submodule-specific fields if needed
}

// ExampleResponse for examples (extends submodule with cost analysis)
type ExampleResponse struct {
	SubmoduleResponse
	CostAnalysis *CostAnalysis `json:"cost_analysis"`
}

// CostAnalysis for example cost breakdown
type CostAnalysis struct {
	Monthly  float64 `json:"monthly"`
	Hourly   float64 `json:"hourly"`
	Currency string  `json:"currency"`
}

// ModuleListResponse represents a list of module providers
type ModuleListResponse struct {
	Modules []ModuleProviderResponse `json:"modules"`
	Meta    *dto.PaginationMeta      `json:"meta"`
}

// ModuleSearchResponse represents search results
type ModuleSearchResponse struct {
	Modules []ModuleProviderResponse `json:"modules"`
	Meta    dto.PaginationMeta        `json:"meta"`
}

// ModuleProviderCreateRequest represents a request to create a module provider
type ModuleProviderCreateRequest struct {
	Namespace string `json:"namespace" binding:"required"`
	Module    string `json:"module" binding:"required"`
	Provider  string `json:"provider" binding:"required"`
}

// ModuleVersionPublishRequest represents a request to publish a module version
type ModuleVersionPublishRequest struct {
	Version     string  `json:"version" binding:"required"`
	Beta        bool    `json:"beta"`
	Description *string `json:"description"`
	Owner       *string `json:"owner"`
}

// ModuleProviderSettingsRequest represents a request to update module provider settings
type ModuleProviderSettingsRequest struct {
	GitProviderID         *int    `json:"git_provider_id"`
	RepoBaseURLTemplate   *string `json:"repo_base_url_template"`
	RepoCloneURLTemplate  *string `json:"repo_clone_url_template"`
	RepoBrowseURLTemplate *string `json:"repo_browse_url_template"`
	GitTagFormat          *string `json:"git_tag_format"`
	GitPath               *string `json:"git_path"`
	ArchiveGitPath        *bool   `json:"archive_git_path"`
	Verified              *bool   `json:"verified"`
}

// ModuleProviderSettingsResponse represents module provider settings in API responses
type ModuleProviderSettingsResponse struct {
	Namespace             string  `json:"namespace"`
	Module                string  `json:"module"`
	Provider              string  `json:"provider"`
	GitProviderID         *int    `json:"git_provider_id"`
	RepoBaseURLTemplate   *string `json:"repo_base_url_template"`
	RepoCloneURLTemplate  *string `json:"repo_clone_url_template"`
	RepoBrowseURLTemplate *string `json:"repo_browse_url_template"`
	GitTagFormat          *string `json:"git_tag_format"`
	GitPath               *string `json:"git_path"`
	ArchiveGitPath        bool    `json:"archive_git_path"`
	Verified              bool    `json:"verified"`
}

