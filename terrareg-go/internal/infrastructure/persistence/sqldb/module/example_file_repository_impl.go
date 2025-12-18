package module

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/module/repository"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/persistence/sqldb"
)

// ExampleFileRepositoryImpl implements ExampleFileRepository
type ExampleFileRepositoryImpl struct {
	db *gorm.DB
}

// NewExampleFileRepository creates a new ExampleFileRepository
func NewExampleFileRepository(db *gorm.DB) repository.ExampleFileRepository {
	return &ExampleFileRepositoryImpl{db: db}
}

// Save saves an example file to the database
func (r *ExampleFileRepositoryImpl) Save(ctx context.Context, exampleFile *sqldb.ExampleFileDB) (*sqldb.ExampleFileDB, error) {
	if exampleFile == nil {
		return nil, fmt.Errorf("example file cannot be nil")
	}

	db := r.getDBFromContext(ctx)
	if err := db.Create(exampleFile).Error; err != nil {
		return nil, fmt.Errorf("failed to save example file: %w", err)
	}

	return exampleFile, nil
}

// SaveBatch saves multiple example files in a single transaction
func (r *ExampleFileRepositoryImpl) SaveBatch(ctx context.Context, exampleFiles []*sqldb.ExampleFileDB) ([]*sqldb.ExampleFileDB, error) {
	if len(exampleFiles) == 0 {
		return exampleFiles, nil
	}

	db := r.getDBFromContext(ctx)
	if err := db.CreateInBatches(exampleFiles, 100).Error; err != nil {
		return nil, fmt.Errorf("failed to save example files batch: %w", err)
	}

	return exampleFiles, nil
}

// FindBySubmoduleID finds all example files for a submodule
func (r *ExampleFileRepositoryImpl) FindBySubmoduleID(ctx context.Context, submoduleID int) ([]sqldb.ExampleFileDB, error) {
	var exampleFiles []sqldb.ExampleFileDB

	db := r.getDBFromContext(ctx)
	if err := db.Where("submodule_id = ?", submoduleID).Find(&exampleFiles).Error; err != nil {
		return nil, fmt.Errorf("failed to find example files for submodule %d: %w", submoduleID, err)
	}

	return exampleFiles, nil
}

// DeleteBySubmoduleID deletes all example files for a submodule
func (r *ExampleFileRepositoryImpl) DeleteBySubmoduleID(ctx context.Context, submoduleID int) error {
	db := r.getDBFromContext(ctx)
	if err := db.Where("submodule_id = ?", submoduleID).Delete(&sqldb.ExampleFileDB{}).Error; err != nil {
		return fmt.Errorf("failed to delete example files for submodule %d: %w", submoduleID, err)
	}

	return nil
}

// DeleteByModuleVersion deletes all example files for a module version
func (r *ExampleFileRepositoryImpl) DeleteByModuleVersion(ctx context.Context, moduleVersionID int) error {
	db := r.getDBFromContext(ctx)

	// First find all submodules for the module version
	var submoduleIDs []int
	if err := db.Table("submodule").
		Select("id").
		Where("parent_module_version = ?", moduleVersionID).
		Scan(&submoduleIDs).Error; err != nil {
		return fmt.Errorf("failed to find submodules for module version %d: %w", moduleVersionID, err)
	}

	// Delete example files for those submodules
	if len(submoduleIDs) > 0 {
		if err := db.Where("submodule_id IN ?", submoduleIDs).Delete(&sqldb.ExampleFileDB{}).Error; err != nil {
			return fmt.Errorf("failed to delete example files for module version %d: %w", moduleVersionID, err)
		}
	}

	return nil
}

// getDBFromContext gets the database connection from context
func (r *ExampleFileRepositoryImpl) getDBFromContext(ctx context.Context) *gorm.DB {
	if db, ok := ctx.Value("tx").(*gorm.DB); ok {
		return db
	}
	return r.db
}