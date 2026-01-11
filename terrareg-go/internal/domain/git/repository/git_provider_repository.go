package repository

import (
	"context"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/git/model"
)

// GitProviderRepository defines the interface for git provider persistence operations
// Python reference: models.py::GitProvider (database operations)
type GitProviderRepository interface {
	// FindByName retrieves a git provider by its name
	// Python reference: models.py::GitProvider.get_by_name()
	FindByName(ctx context.Context, name string) (*model.GitProvider, error)

	// FindAll retrieves all git providers
	// Python reference: models.py::GitProvider.get_all()
	FindAll(ctx context.Context) ([]*model.GitProvider, error)

	// Upsert creates or updates a git provider
	// Python reference: models.py::GitProvider.initialise_from_config() (upsert logic)
	Upsert(ctx context.Context, provider *model.GitProvider) error

	// Delete removes a git provider by name
	Delete(ctx context.Context, name string) error

	// Exists checks if a git provider with the given name exists
	Exists(ctx context.Context, name string) (bool, error)
}
