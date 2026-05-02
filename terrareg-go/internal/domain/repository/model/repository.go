package model

// Repository represents a repository from a provider source (e.g., GitHub)
// Python reference: repository_model.py::Repository
type Repository struct {
	// ID is the database primary key
	ID int
	// ProviderID is the ID from the provider (e.g., GitHub repository ID)
	ProviderID string
	// Name is the repository name
	Name string
	// Owner is the namespace/owner of the repository (this is the key for namespace permission checks)
	Owner string
	// Description is the repository description
	Description *string
	// CloneURL is the URL to clone the repository
	CloneURL string
	// LogoURL is the URL to the repository logo
	LogoURL *string
	// ProviderSourceName is the name of the provider source (e.g., "github")
	ProviderSourceName string
}
