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
