package module

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/module/repository"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/persistence/sqldb"
	baserepo "github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/persistence/sqldb/repository"
)

// SubmoduleRepositoryImpl implements SubmoduleRepository
type SubmoduleRepositoryImpl struct {
	*baserepo.BaseRepository
}

// NewSubmoduleRepository creates a new SubmoduleRepository
// Returns an error if db is nil
func NewSubmoduleRepository(db *gorm.DB) (repository.SubmoduleRepository, error) {
	baseRepo, err := baserepo.NewBaseRepository(db)
	if err != nil {
		return nil, fmt.Errorf("failed to create base repository: %w", err)
	}
	return &SubmoduleRepositoryImpl{
		BaseRepository: baseRepo,
	}, nil
}

// Save saves a submodule to the database
func (r *SubmoduleRepositoryImpl) Save(ctx context.Context, parentModuleVersionID int, submodule *sqldb.SubmoduleDB) (*sqldb.SubmoduleDB, error) {
	if submodule == nil {
		return nil, fmt.Errorf("submodule cannot be nil")
	}

	// Set parent module version ID
	submodule.ParentModuleVersion = parentModuleVersionID

	db := r.GetDBFromContext(ctx)
	if err := db.Create(submodule).Error; err != nil {
		return nil, fmt.Errorf("failed to save submodule: %w", err)
	}

	return submodule, nil
}

// SaveWithDetails saves a submodule with module details
func (r *SubmoduleRepositoryImpl) SaveWithDetails(ctx context.Context, parentModuleVersionID int, submodule *sqldb.SubmoduleDB, moduleDetailsID int) (*sqldb.SubmoduleDB, error) {
	if submodule == nil {
		return nil, fmt.Errorf("submodule cannot be nil")
	}

	// Set parent module version ID and module details ID
	submodule.ParentModuleVersion = parentModuleVersionID
	submodule.ModuleDetailsID = &moduleDetailsID

	db := r.GetDBFromContext(ctx)
	if err := db.Create(submodule).Error; err != nil {
		return nil, fmt.Errorf("failed to save submodule with details: %w", err)
	}

	return submodule, nil
}

// FindByParentModuleVersion finds all submodules for a module version
func (r *SubmoduleRepositoryImpl) FindByParentModuleVersion(ctx context.Context, moduleVersionID int) ([]sqldb.SubmoduleDB, error) {
	var submodules []sqldb.SubmoduleDB

	db := r.GetDBFromContext(ctx)
	if err := db.Where("parent_module_version = ?", moduleVersionID).
		Preload("ModuleDetails").
		Find(&submodules).Error; err != nil {
		return nil, fmt.Errorf("failed to find submodules for module version %d: %w", moduleVersionID, err)
	}

	return submodules, nil
}

// FindByPath finds a submodule by parent module version and path
func (r *SubmoduleRepositoryImpl) FindByPath(ctx context.Context, moduleVersionID int, path string) (*sqldb.SubmoduleDB, error) {
	var submodule sqldb.SubmoduleDB

	db := r.GetDBFromContext(ctx)
	if err := db.Where("parent_module_version = ? AND path = ?", moduleVersionID, path).
		Preload("ModuleDetails").
		First(&submodule).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("submodule not found for path %s", path)
		}
		return nil, fmt.Errorf("failed to find submodule by path %s: %w", path, err)
	}

	return &submodule, nil
}

// Delete deletes submodules for a module version
func (r *SubmoduleRepositoryImpl) DeleteByParentModuleVersion(ctx context.Context, moduleVersionID int) error {
	db := r.GetDBFromContext(ctx)
	if err := db.Where("parent_module_version = ?", moduleVersionID).Delete(&sqldb.SubmoduleDB{}).Error; err != nil {
		return fmt.Errorf("failed to delete submodules for module version %d: %w", moduleVersionID, err)
	}

	return nil
}

// UpdateModuleDetailsID updates the module details ID for a submodule
func (r *SubmoduleRepositoryImpl) UpdateModuleDetailsID(ctx context.Context, submoduleID int, moduleDetailsID int) error {
	db := r.GetDBFromContext(ctx)
	result := db.Model(&sqldb.SubmoduleDB{}).
		Where("id = ?", submoduleID).
		Update("module_details_id", moduleDetailsID)

	if result.Error != nil {
		return fmt.Errorf("failed to update module details ID for submodule %d: %w", submoduleID, result.Error)
	}

	return nil
}
