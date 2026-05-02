package repository

import (
	"context"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/repository/model"
)

// RepositoryRepository defines the interface for repository operations
// Python reference: repository_model.py::Repository
type RepositoryRepository interface {
	// FindByProviderSourceAndProviderID retrieves a repository by provider source name and provider ID
	// Python reference: repository_model.py::get_by_provider_source_and_provider_id
	FindByProviderSourceAndProviderID(ctx context.Context, providerSourceName string, providerID string) (*model.Repository, error)
}
