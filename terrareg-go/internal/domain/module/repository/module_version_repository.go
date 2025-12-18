package repository

import (
	"context"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/module/model"
)

// ModuleVersionRepository defines the interface for module version persistence
type ModuleVersionRepository interface {
	// FindByModuleProvider retrieves module versions for a specific module provider
	FindByModuleProvider(ctx context.Context, moduleProviderID int, includeBeta, includeUnpublished bool) ([]*model.ModuleVersion, error)

	// Save persists a module version and returns the updated version with database-assigned ID
	Save(ctx context.Context, moduleVersion *model.ModuleVersion) (*model.ModuleVersion, error)

	// FindByID retrieves a module version by ID
	FindByID(ctx context.Context, id int) (*model.ModuleVersion, error)

	// FindByModuleProviderAndVersion retrieves a specific module version
	FindByModuleProviderAndVersion(ctx context.Context, moduleProviderID int, version string) (*model.ModuleVersion, error)

	// Delete removes a module version
	Delete(ctx context.Context, id int) error

	// Exists checks if a module version exists
	Exists(ctx context.Context, moduleProviderID int, version string) (bool, error)

	// UpdateModuleDetailsID updates the module details ID for a module version
	UpdateModuleDetailsID(ctx context.Context, moduleVersionID int, moduleDetailsID int) error
}
