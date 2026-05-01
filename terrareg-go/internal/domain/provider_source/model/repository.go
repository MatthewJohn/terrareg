package model

// Repository represents a GitHub repository that can be used to publish a provider
type Repository struct {
	// ID is the GitHub repository ID (as string for compatibility)
	ID string `json:"id"`
	// FullName is the full repository name (e.g., "owner/repo")
	FullName string `json:"full_name"`
	// OwnerLogin is the owner's username
	OwnerLogin string `json:"owner_login"`
	// OwnerType is the type of owner ("user" or "organization")
	OwnerType string `json:"owner_type"`
	// Kind is the repository kind (e.g., "repository_kind_value")
	Kind string `json:"kind"`
	// PublishedID is the ID of the published provider, if any
	PublishedID *string `json:"published_id"`
}

// NewRepository creates a new repository model
func NewRepository(id, fullName, ownerLogin, ownerType, kind string, publishedID *string) *Repository {
	return &Repository{
		ID:          id,
		FullName:    fullName,
		OwnerLogin:  ownerLogin,
		OwnerType:   ownerType,
		Kind:        kind,
		PublishedID: publishedID,
	}
}

// GetID returns the repository ID
func (r *Repository) GetID() string {
	return r.ID
}

// GetFullName returns the full repository name
func (r *Repository) GetFullName() string {
	return r.FullName
}

// GetOwnerLogin returns the owner's username
func (r *Repository) GetOwnerLogin() string {
	return r.OwnerLogin
}

// GetOwnerType returns the owner type
func (r *Repository) GetOwnerType() string {
	return r.OwnerType
}

// GetKind returns the repository kind
func (r *Repository) GetKind() string {
	return r.Kind
}

// IsPublished returns whether this repository has been published
func (r *Repository) IsPublished() bool {
	return r.PublishedID != nil
}

// GetPublishedID returns the published provider ID
func (r *Repository) GetPublishedID() *string {
	return r.PublishedID
}
