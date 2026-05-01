package dto

// RefreshNamespaceRequest represents a request to refresh namespace repositories
type RefreshNamespaceRequest struct {
	// Namespace is the namespace to refresh
	Namespace string `json:"namespace" binding:"required"`
	// CsrfToken is the CSRF token for protection
	CsrfToken string `json:"csrf_token" binding:"required"`
}

// PublishProviderRequest represents a request to publish a provider from a repository
// This is sent as form data (multipart/form-data)
type PublishProviderRequest struct {
	// CategoryID is the provider category ID
	CategoryID int `form:"category_id" binding:"required"`
	// CsrfToken is the CSRF token for protection
	CsrfToken string `form:"csrf_token"`
}
