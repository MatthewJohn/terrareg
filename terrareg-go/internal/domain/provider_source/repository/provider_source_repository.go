package repository

import (
	"context"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/provider_source/model"
)

// ProviderSourceRepository defines the persistence contract for provider sources
// Python reference: provider_source/factory.py (implicitly uses Database.get())
type ProviderSourceRepository interface {
	// FindByName retrieves a provider source by its display name
	// Returns nil if not found (no error)
	FindByName(ctx context.Context, name string) (*model.ProviderSource, error)

	// FindByApiName retrieves a provider source by its API-friendly name
	// Returns nil if not found (no error)
	FindByApiName(ctx context.Context, apiName string) (*model.ProviderSource, error)

	// FindAll retrieves all provider sources from the database
	// Returns empty slice if none exist
	FindAll(ctx context.Context) ([]*model.ProviderSource, error)

	// Upsert creates a new provider source or updates an existing one
	// Python reference: factory.py::initialise_from_config() - uses insert or update
	Upsert(ctx context.Context, source *model.ProviderSource) error

	// Delete removes a provider source by name
	// Returns nil if source doesn't exist (idempotent)
	Delete(ctx context.Context, name string) error

	// Exists checks if a provider source with the given name exists
	Exists(ctx context.Context, name string) (bool, error)

	// ExistsByApiName checks if a provider source with the given API name exists
	ExistsByApiName(ctx context.Context, apiName string) (bool, error)
}
