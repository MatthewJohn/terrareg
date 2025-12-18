package module

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/module/model"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/module/repository"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/persistence/sqldb"
	baserepo "github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/persistence/sqldb/repository"
)

// ModuleDetailsRepositoryImpl implements the ModuleDetailsRepository interface
type ModuleDetailsRepositoryImpl struct {
	*baserepo.BaseRepository
}

// NewModuleDetailsRepository creates a new ModuleDetails repository implementation
func NewModuleDetailsRepository(db *gorm.DB) repository.ModuleDetailsRepository {
	return &ModuleDetailsRepositoryImpl{
		BaseRepository: baserepo.NewBaseRepository(db),
	}
}

// Save saves a new module details entity to the database
func (r *ModuleDetailsRepositoryImpl) Save(ctx context.Context, details *model.ModuleDetails) (*model.ModuleDetails, error) {
	db := r.GetDBFromContext(ctx)

	// Convert domain model to database model
	dbDetails := &sqldb.ModuleDetailsDB{
		ReadmeContent:    details.ReadmeContent(),
		TerraformDocs:    details.TerraformDocs(),
		Tfsec:            details.Tfsec(),
		Infracost:        details.Infracost(),
		TerraformGraph:   details.TerraformGraph(),
		TerraformModules: details.TerraformModules(),
		TerraformVersion: []byte(details.TerraformVersion()),
	}

	// Create the record in the database
	if err := db.Create(dbDetails).Error; err != nil {
		return nil, fmt.Errorf("failed to create module details: %w", err)
	}

	// Convert back to domain model
	return fromDBModuleDetails(dbDetails), nil
}

// FindByID retrieves a module details entity by its ID
func (r *ModuleDetailsRepositoryImpl) FindByID(ctx context.Context, id int) (*model.ModuleDetails, error) {
	db := r.GetDBFromContext(ctx)

	var dbDetails sqldb.ModuleDetailsDB
	err := db.First(&dbDetails, id).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find module details by ID %d: %w", id, err)
	}

	return fromDBModuleDetails(&dbDetails), nil
}

// Update updates an existing module details entity
func (r *ModuleDetailsRepositoryImpl) Update(ctx context.Context, id int, details *model.ModuleDetails) (*model.ModuleDetails, error) {
	db := r.GetDBFromContext(ctx)

	// Convert domain model to database model
	dbDetails := &sqldb.ModuleDetailsDB{
		ID:               id,
		ReadmeContent:    details.ReadmeContent(),
		TerraformDocs:    details.TerraformDocs(),
		Tfsec:            details.Tfsec(),
		Infracost:        details.Infracost(),
		TerraformGraph:   details.TerraformGraph(),
		TerraformModules: details.TerraformModules(),
		TerraformVersion: []byte(details.TerraformVersion()),
	}

	// Update the record in the database
	if err := db.Save(dbDetails).Error; err != nil {
		return nil, fmt.Errorf("failed to update module details: %w", err)
	}

	// Convert back to domain model
	return fromDBModuleDetails(dbDetails), nil
}

// Delete removes a module details entity by its ID
func (r *ModuleDetailsRepositoryImpl) Delete(ctx context.Context, id int) error {
	db := r.GetDBFromContext(ctx)

	if err := db.Delete(&sqldb.ModuleDetailsDB{}, id).Error; err != nil {
		return fmt.Errorf("failed to delete module details with ID %d: %w", id, err)
	}

	return nil
}

// FindByModuleVersionID retrieves module details associated with a specific module version
func (r *ModuleDetailsRepositoryImpl) FindByModuleVersionID(ctx context.Context, moduleVersionID int) (*model.ModuleDetails, error) {
	db := r.GetDBFromContext(ctx)

	var dbDetails sqldb.ModuleDetailsDB
	err := db.
		Joins("JOIN module_version ON module_details.id = module_version.module_details_id").
		Where("module_version.id = ?", moduleVersionID).
		First(&dbDetails).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find module details for module version %d: %w", moduleVersionID, err)
	}

	return fromDBModuleDetails(&dbDetails), nil
}
// SaveAndReturnID saves a new module details entity and returns the database ID
func (r *ModuleDetailsRepositoryImpl) SaveAndReturnID(ctx context.Context, details *model.ModuleDetails) (int, error) {
	db := r.GetDBFromContext(ctx)

	// Convert domain model to database model
	dbDetails := &sqldb.ModuleDetailsDB{
		ReadmeContent:    details.ReadmeContent(),
		TerraformDocs:    details.TerraformDocs(),
		Tfsec:            details.Tfsec(),
		Infracost:        details.Infracost(),
		TerraformGraph:   details.TerraformGraph(),
		TerraformModules: details.TerraformModules(),
		TerraformVersion: []byte(details.TerraformVersion()),
	}

	// Create the record in the database
	if err := db.Create(dbDetails).Error; err != nil {
		return 0, fmt.Errorf("failed to create module details: %w", err)
	}

	return dbDetails.ID, nil
}
