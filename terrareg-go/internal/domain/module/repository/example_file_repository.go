package repository

import (
	"context"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/persistence/sqldb"
)

// ExampleFileRepository defines operations for example file persistence
type ExampleFileRepository interface {
	// Save saves an example file to the database
	Save(ctx context.Context, exampleFile *sqldb.ExampleFileDB) (*sqldb.ExampleFileDB, error)

	// SaveBatch saves multiple example files in a single transaction
	SaveBatch(ctx context.Context, exampleFiles []*sqldb.ExampleFileDB) ([]*sqldb.ExampleFileDB, error)

	// FindBySubmoduleID finds all example files for a submodule
	FindBySubmoduleID(ctx context.Context, submoduleID int) ([]sqldb.ExampleFileDB, error)

	// DeleteBySubmoduleID deletes all example files for a submodule
	DeleteBySubmoduleID(ctx context.Context, submoduleID int) error

	// DeleteByModuleVersion deletes all example files for a module version
	DeleteByModuleVersion(ctx context.Context, moduleVersionID int) error
}