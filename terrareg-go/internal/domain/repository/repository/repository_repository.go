package repository

import (
	"context"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/repository/model"
)

// RepositoryRepository defines the persistence contract for repositories
// Python reference: repository_model.py::Repository class methods
type RepositoryRepository interface {
	// FindByID retrieves a repository by its database primary key
	// Returns nil if not found (no error)
	// Python reference: repository_model.py::Repository.get_by_pk()
	FindByID(ctx context.Context, id int) (*model.Repository, error)

	// FindByProviderSourceAndProviderID retrieves a repository by provider source name and provider ID
	// Returns nil if not found (no error)
	// Python reference: repository_model.py::Repository.get_by_provider_source_and_provider_id()
	FindByProviderSourceAndProviderID(ctx context.Context, providerSourceName string, providerID string) (*model.Repository, error)

	// FindByOwnerList retrieves repositories matching any of the given owners
	// Returns empty slice if none found
	// Python reference: repository_model.py::Repository.get_repositories_by_owner_list()
	FindByOwnerList(ctx context.Context, owners []string) ([]*model.Repository, error)

	// Create creates a new repository
	// Returns the created repository with ID set, or error if creation fails
	// Python reference: repository_model.py::Repository.create()
	Create(ctx context.Context, repository *model.Repository) (*model.Repository, error)

	// Update updates an existing repository
	// Python reference: repository_model.py::Repository.update_attributes()
	Update(ctx context.Context, repository *model.Repository) error

	// Delete removes a repository by its ID
	// Returns nil if repository doesn't exist (idempotent)
	// Python reference: implicit (used in tests)
	Delete(ctx context.Context, id int) error

	// Exists checks if a repository exists by provider source name and provider ID
	// Python reference: used in Repository.create() before insertion
	Exists(ctx context.Context, providerSourceName string, providerID string) (bool, error)
}
