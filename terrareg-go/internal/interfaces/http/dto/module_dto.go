package dto

// ModuleProviderResponse represents a module provider in API responses
type ModuleProviderResponse struct {
	ID          string  `json:"id"`
	Namespace   string  `json:"namespace"`
	Name        string  `json:"name"`
	Provider    string  `json:"provider"`
	Verified    bool    `json:"verified"`
	Description *string `json:"description,omitempty"`
	Owner       *string `json:"owner,omitempty"`
	Source      *string `json:"source,omitempty"`
	PublishedAt *string `json:"published_at,omitempty"`
	Downloads   int     `json:"downloads,omitempty"`
	Version     *string `json:"version,omitempty"` // Latest version
}

// ModuleListResponse represents a list of module providers
type ModuleListResponse struct {
	Modules []ModuleProviderResponse `json:"modules"`
	Meta    *PaginationMeta          `json:"meta,omitempty"`
}

// ModuleSearchResponse represents search results
type ModuleSearchResponse struct {
	Modules []ModuleProviderResponse `json:"modules"`
	Meta    PaginationMeta           `json:"meta"`
}

// PaginationMeta represents pagination metadata
type PaginationMeta struct {
	Limit      int `json:"limit"`
	Offset     int `json:"offset"`
	TotalCount int `json:"total_count"`
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

// ModuleVersionResponse represents a module version in API responses
type ModuleVersionResponse struct {
	ID          string  `json:"id"`
	Version     string  `json:"version"`
	Published   bool    `json:"published"`
	Beta        bool    `json:"beta"`
	Internal    bool    `json:"internal"`
	Description *string `json:"description,omitempty"`
	Owner       *string `json:"owner,omitempty"`
	PublishedAt *string `json:"published_at,omitempty"`
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
	GitProviderID         *int    `json:"git_provider_id,omitempty"`
	RepoBaseURLTemplate   *string `json:"repo_base_url_template,omitempty"`
	RepoCloneURLTemplate  *string `json:"repo_clone_url_template,omitempty"`
	RepoBrowseURLTemplate *string `json:"repo_browse_url_template,omitempty"`
	GitTagFormat          *string `json:"git_tag_format,omitempty"`
	GitPath               *string `json:"git_path,omitempty"`
	ArchiveGitPath        bool    `json:"archive_git_path"`
	Verified              bool    `json:"verified"`
}

// NamespaceUpdateRequest represents a request to update a namespace
type NamespaceUpdateRequest struct {
	Name        *string `json:"name,omitempty"`
	DisplayName *string `json:"display_name,omitempty"`
	CsrfToken   string  `json:"csrf_token"`
}
