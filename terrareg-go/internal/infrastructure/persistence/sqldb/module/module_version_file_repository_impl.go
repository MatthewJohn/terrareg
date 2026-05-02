package module

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/module/model"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/persistence/sqldb"
)

// ModuleVersionFileRepositoryImpl implements the module version file repository using GORM
type ModuleVersionFileRepositoryImpl struct {
	db *gorm.DB
}

// NewModuleVersionFileRepository creates a new module version file repository
func NewModuleVersionFileRepository(db *gorm.DB) *ModuleVersionFileRepositoryImpl {
	return &ModuleVersionFileRepositoryImpl{db: db}
}

// FindByPath retrieves a module version file by its path
func (r *ModuleVersionFileRepositoryImpl) FindByPath(ctx context.Context, moduleVersionID int, path string) (*model.ModuleVersionFile, error) {
	var fileDB sqldb.ModuleVersionFileDB
	err := r.db.WithContext(ctx).
		Where("module_version_id = ? AND path = ?", moduleVersionID, path).
		First(&fileDB).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("module version file not found: %w", model.ErrFileNotFound)
		}
		return nil, fmt.Errorf("failed to query module version file: %w", err)
	}

	// Convert to domain model
	return r.mapToDomainModel(&fileDB)
}

// FindByModuleVersionID retrieves all files for a module version
func (r *ModuleVersionFileRepositoryImpl) FindByModuleVersionID(ctx context.Context, moduleVersionID int) ([]*model.ModuleVersionFile, error) {
	var fileDBs []sqldb.ModuleVersionFileDB
	err := r.db.WithContext(ctx).
		Where("module_version_id = ?", moduleVersionID).
		Order("path ASC").
		Find(&fileDBs).Error
	if err != nil {
		return nil, fmt.Errorf("failed to query module version files: %w", err)
	}

	// Convert to domain models
	files := make([]*model.ModuleVersionFile, 0, len(fileDBs))
	for _, fileDB := range fileDBs {
		file, err := r.mapToDomainModel(&fileDB)
		if err != nil {
			return nil, fmt.Errorf("failed to map module version file: %w", err)
		}
		files = append(files, file)
	}

	return files, nil
}

// Save saves a module version file
func (r *ModuleVersionFileRepositoryImpl) Save(ctx context.Context, file *model.ModuleVersionFile) error {
	fileDB := &sqldb.ModuleVersionFileDB{
		ModuleVersionID: file.ModuleVersion().ID(),
		Path:            file.Path(),
		Content:         []byte(file.Content()),
	}

	// Check if file already exists
	var existingFile sqldb.ModuleVersionFileDB
	err := r.db.WithContext(ctx).
		Where("module_version_id = ? AND path = ?", file.ModuleVersion().ID(), file.Path()).
		First(&existingFile).Error

	if err == nil {
		// Update existing file
		fileDB.ID = existingFile.ID
		err = r.db.WithContext(ctx).Save(fileDB).Error
		if err != nil {
			return fmt.Errorf("failed to update module version file: %w", err)
		}
		// Update ID in domain model
		fileID := fileDB.ID
		*file = *model.NewModuleVersionFile(fileID, file.ModuleVersion(), file.Path(), file.Content())
	} else if err == gorm.ErrRecordNotFound {
		// Create new file
		err = r.db.WithContext(ctx).Create(fileDB).Error
		if err != nil {
			return fmt.Errorf("failed to create module version file: %w", err)
		}
		// Update ID in domain model
		fileID := fileDB.ID
		*file = *model.NewModuleVersionFile(fileID, file.ModuleVersion(), file.Path(), file.Content())
	} else {
		return fmt.Errorf("failed to check existing module version file: %w", err)
	}

	return nil
}

// Delete deletes a module version file
func (r *ModuleVersionFileRepositoryImpl) Delete(ctx context.Context, id int) error {
	err := r.db.WithContext(ctx).Delete(&sqldb.ModuleVersionFileDB{}, id).Error
	if err != nil {
		return fmt.Errorf("failed to delete module version file: %w", err)
	}
	return nil
}

// mapToDomainModel converts a database model to domain model
func (r *ModuleVersionFileRepositoryImpl) mapToDomainModel(fileDB *sqldb.ModuleVersionFileDB) (*model.ModuleVersionFile, error) {
	// Convert content from bytes to string
	content := string(fileDB.Content)

	// Create module version file domain model
	// Note: We don't have the module version here, so we create with nil and let the service handle it
	file := model.NewModuleVersionFile(fileDB.ID, nil, fileDB.Path, content)
	return file, nil
}
