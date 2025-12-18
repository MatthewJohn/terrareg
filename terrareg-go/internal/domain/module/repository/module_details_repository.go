package repository

import (
	"context"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/module/model"
)

// ModuleDetailsRepository defines the interface for module details persistence operations
type ModuleDetailsRepository interface {
	// Save saves a new module details entity to the database
	Save(ctx context.Context, details *model.ModuleDetails) (*model.ModuleDetails, error)

	// FindByID retrieves a module details entity by its ID
	FindByID(ctx context.Context, id int) (*model.ModuleDetails, error)

	// Update updates an existing module details entity
	Update(ctx context.Context, id int, details *model.ModuleDetails) (*model.ModuleDetails, error)

	// Delete removes a module details entity by its ID
	Delete(ctx context.Context, id int) error

	// FindByModuleVersionID retrieves module details associated with a specific module version
	FindByModuleVersionID(ctx context.Context, moduleVersionID int) (*model.ModuleDetails, error)

	// SaveAndReturnID saves a new module details entity and returns the database ID
	SaveAndReturnID(ctx context.Context, details *model.ModuleDetails) (int, error)
}