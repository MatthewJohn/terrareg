package repository

import (
	"context"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/persistence/sqldb"
)

// SubmoduleRepository defines operations for submodule persistence
type SubmoduleRepository interface {
	// Save saves a submodule to the database
	Save(ctx context.Context, parentModuleVersionID int, submodule *sqldb.SubmoduleDB) (*sqldb.SubmoduleDB, error)

	// SaveWithDetails saves a submodule with module details
	SaveWithDetails(ctx context.Context, parentModuleVersionID int, submodule *sqldb.SubmoduleDB, moduleDetailsID int) (*sqldb.SubmoduleDB, error)

	// FindByParentModuleVersion finds all submodules for a module version
	FindByParentModuleVersion(ctx context.Context, moduleVersionID int) ([]sqldb.SubmoduleDB, error)

	// FindByPath finds a submodule by parent module version and path
	FindByPath(ctx context.Context, moduleVersionID int, path string) (*sqldb.SubmoduleDB, error)

	// Delete deletes submodules for a module version
	DeleteByParentModuleVersion(ctx context.Context, moduleVersionID int) error
}
