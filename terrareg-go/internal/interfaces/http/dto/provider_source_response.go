package dto

// OrganizationResponse represents an organization in API responses
type OrganizationResponse struct {
	// Name is the organization/user name (also used as namespace name)
	Name string `json:"name"`
	// Type indicates whether this is a "user" or "organization"
	Type string `json:"type"`
	// CanPublishProviders indicates if the user can publish providers for this organization
	CanPublishProviders bool `json:"can_publish_providers"`
}

// RepositoryResponse represents a repository in API responses
type RepositoryResponse struct {
	// ID is the GitHub repository ID
	ID string `json:"id"`
	// FullName is the full repository name (e.g., "owner/repo")
	FullName string `json:"full_name"`
	// OwnerLogin is the owner's username
	OwnerLogin string `json:"owner_login"`
	// OwnerType is the type of owner ("user" or "organization")
	OwnerType string `json:"owner_type"`
	// Kind is the repository kind
	Kind string `json:"kind"`
	// PublishedID is the ID of the published provider, if any
	PublishedID *string `json:"published_id"`
}

// PublishProviderResponse represents the result of publishing a provider
type PublishProviderResponse struct {
	// Name is the provider name
	Name string `json:"name"`
	// Namespace is the namespace name
	Namespace string `json:"namespace"`
}
