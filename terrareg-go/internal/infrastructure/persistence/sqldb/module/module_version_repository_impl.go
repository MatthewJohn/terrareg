package module

import (
	"context"
	"fmt"

	"gorm.io/gorm"
	"github.com/rs/zerolog"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/module/model"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/persistence/sqldb"
	baserepo "github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/persistence/sqldb/repository"
)

// ModuleVersionRepositoryImpl implements the module version repository using GORM
type ModuleVersionRepositoryImpl struct {
	*baserepo.BaseRepository
	submoduleLoader *SubmoduleLoader
}

// NewModuleVersionRepository creates a new module version repository
func NewModuleVersionRepository(db *gorm.DB) *ModuleVersionRepositoryImpl {
	return &ModuleVersionRepositoryImpl{
		BaseRepository: baserepo.NewBaseRepository(db),
		submoduleLoader: NewSubmoduleLoader(db),
	}
}


// FindByModuleProvider retrieves module versions for a specific module provider
func (r *ModuleVersionRepositoryImpl) FindByModuleProvider(ctx context.Context, moduleProviderID int, includeBeta, includeUnpublished bool) ([]*model.ModuleVersion, error) {
	var moduleVersionDBs []sqldb.ModuleVersionDB
	db := r.GetDBFromContext(ctx)
	query := db.Where("module_provider_id = ?", moduleProviderID)

	// Apply filters based on parameters
	if !includeBeta {
		query = query.Where("beta = ?", false)
	}
	if !includeUnpublished {
		query = query.Where("published = ?", true)
	}

	// Order by version (newest first) - for simplicity, we'll order by ID
	query = query.Order("id DESC")

	if err := query.Find(&moduleVersionDBs).Error; err != nil {
		return nil, fmt.Errorf("failed to query module versions: %w", err)
	}

	// Convert to domain models
	moduleVersions := make([]*model.ModuleVersion, 0, len(moduleVersionDBs))
	for _, dbVersion := range moduleVersionDBs {
		moduleVersion, err := r.mapToDomainModel(dbVersion)
		if err != nil {
			return nil, fmt.Errorf("failed to map module version: %w", err)
		}
		moduleVersions = append(moduleVersions, moduleVersion)
	}

	return moduleVersions, nil
}

// Save persists a module version and returns the updated version with database-assigned ID
// Note: This should NOT create its own transaction - it should participate in a transaction
// created by the service layer
func (r *ModuleVersionRepositoryImpl) Save(ctx context.Context, moduleVersion *model.ModuleVersion) (*model.ModuleVersion, error) {
	if moduleVersion == nil {
		return nil, fmt.Errorf("module version cannot be nil")
	}

	// CRITICAL DEBUG: Log the ID to understand what's happening
	logger := zerolog.Ctx(ctx)
	logger.Debug().
		Int("module_version_id", moduleVersion.ID()).
		Str("version", moduleVersion.Version().String()).
		Bool("is_new_record", moduleVersion.ID() == 0).
		Msg("ModuleVersion Save: Before mapping to persistence model")

	dbVersion, err := r.mapToPersistenceModel(ctx, moduleVersion)
	if err != nil {
		return nil, fmt.Errorf("failed to map module version: %w", err)
	}

	// CRITICAL DEBUG: Log the persistence model ID
	logger.Debug().
		Int("db_version_id", dbVersion.ID).
		Int("module_provider_id", dbVersion.ModuleProviderID).
		Msg("ModuleVersion Save: After mapping to persistence model")

	// Get the database instance from context (participate in existing transaction) or use default
	db := r.GetDBFromContext(ctx)

	// CRITICAL FIX: Always use Create() when ID is 0, regardless of whether a record exists
	// This follows Python's delete-then-create pattern exactly
	if moduleVersion.ID() == 0 {
		logger.Debug().Msg("ModuleVersion Save: Using CREATE operation (new record)")
		if err := db.Create(dbVersion).Error; err != nil {
			return nil, fmt.Errorf("failed to create module version: %w", err)
		}
	} else {
		logger.Debug().Int("existing_id", moduleVersion.ID()).Msg("ModuleVersion Save: Using UPDATE operation (existing record)")
		if err := db.Save(dbVersion).Error; err != nil {
			return nil, fmt.Errorf("failed to update module version: %w", err)
		}
	}

	// After successful save, return the updated domain model with proper database-assigned values
	updatedModuleVersion, err := r.mapToDomainModel(*dbVersion)
	if err != nil {
		return nil, fmt.Errorf("failed to map updated module version: %w", err)
	}

	return updatedModuleVersion, nil
}

// FindByID retrieves a module version by ID
func (r *ModuleVersionRepositoryImpl) FindByID(ctx context.Context, id int) (*model.ModuleVersion, error) {
	var dbVersion sqldb.ModuleVersionDB
	db := r.GetDBFromContext(ctx)
	err := db.First(&dbVersion, id).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find module version: %w", err)
	}

	return r.mapToDomainModel(dbVersion)
}

// FindByModuleProviderAndVersion retrieves a specific module version
func (r *ModuleVersionRepositoryImpl) FindByModuleProviderAndVersion(ctx context.Context, moduleProviderID int, version string) (*model.ModuleVersion, error) {
	var dbVersion sqldb.ModuleVersionDB
	err := r.GetDBFromContext(ctx).
		Where("module_provider_id = ? AND version = ?", moduleProviderID, version).
		First(&dbVersion).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find module version: %w", err)
	}

	return r.mapToDomainModel(dbVersion)
}

// Delete removes a module version
func (r *ModuleVersionRepositoryImpl) Delete(ctx context.Context, id int) error {
	return r.GetDBFromContext(ctx).Delete(&sqldb.ModuleVersionDB{}, id).Error
}

// Exists checks if a module version exists
func (r *ModuleVersionRepositoryImpl) Exists(ctx context.Context, moduleProviderID int, version string) (bool, error) {
	var count int64
	err := r.GetDBFromContext(ctx).
		Model(&sqldb.ModuleVersionDB{}).
		Where("module_provider_id = ? AND version = ?", moduleProviderID, version).
		Count(&count).Error

	if err != nil {
		return false, fmt.Errorf("failed to check if module version exists: %w", err)
	}

	return count > 0, nil
}

// Helper methods for mapping between domain and persistence models

// mapToDomainModel converts persistence model to domain model using centralized mapper
func (r *ModuleVersionRepositoryImpl) mapToDomainModel(dbVersion sqldb.ModuleVersionDB) (*model.ModuleVersion, error) {
	// Load module details if available
	var details *model.ModuleDetails
	if dbVersion.ModuleDetailsID != nil {
		var detailsDB sqldb.ModuleDetailsDB
		err := r.GetDB().First(&detailsDB, *dbVersion.ModuleDetailsID).Error
		if err != nil && err != gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("failed to load module details: %w", err)
		}
		if err == nil {
			details = fromDBModuleDetails(&detailsDB)
		}
	}
	if details == nil {
		details = model.NewModuleDetails([]byte{})
	}

	// Create basic module version
	moduleVersion, err := fromDBModuleVersion(&dbVersion, details)
	if err != nil {
		return nil, err
	}

	// IMPORTANT: Restore the module provider relationship if module_provider_id exists
	// This fixes the issue where domain models lose their parent relationships during reconstruction
	if dbVersion.ModuleProviderID > 0 {
		// Load the module provider and its namespace to restore the relationship
		var moduleProviderDB sqldb.ModuleProviderDB
		err := r.GetDB().Preload("Namespace").First(&moduleProviderDB, dbVersion.ModuleProviderID).Error
		if err == nil {
			// Convert namespace database model to domain model
			namespace := fromDBNamespace(&moduleProviderDB.Namespace)

			// Convert module provider database model to domain model
			moduleProvider := fromDBModuleProvider(&moduleProviderDB, namespace)
			if moduleProvider != nil {
				// Use the SetVersions method which properly establishes parent-child relationships
				moduleProvider.SetVersions([]*model.ModuleVersion{moduleVersion})
			}
		}
	}

	// Load submodules and examples using shared service
	if err := r.submoduleLoader.LoadSubmodulesAndExamples(moduleVersion, dbVersion.ID); err != nil {
		return nil, err
	}

	return moduleVersion, nil
}

// mapToPersistenceModel converts domain model to persistence model using centralized mapper
func (r *ModuleVersionRepositoryImpl) mapToPersistenceModel(ctx context.Context, moduleVersion *model.ModuleVersion) (*sqldb.ModuleVersionDB, error) {
	// For existing records, fetch current database record to preserve module_details_id
	var existingDetailsID *int
	if moduleVersion.ID() > 0 {
		var existingDBVersion sqldb.ModuleVersionDB
		db := r.GetDBFromContext(ctx)
		if err := db.First(&existingDBVersion, moduleVersion.ID()).Error; err == nil {
			existingDetailsID = existingDBVersion.ModuleDetailsID
		}
	}

	dbVersion := toDBModuleVersion(moduleVersion)

	// CRITICAL FIX: Preserve existing module_details_id from database
	// The domain model doesn't track the module_details_id, so we need to preserve it from the database
	if existingDetailsID != nil {
		dbVersion.ModuleDetailsID = existingDetailsID
	}

	// For existing records, if the domain model has lost the module provider relationship,
	// preserve the original module_provider_id from the database and skip update to prevent corruption
	if moduleVersion.ID() > 0 && dbVersion.ModuleProviderID == 0 {
		return nil, fmt.Errorf("cannot update module version %d: domain model has lost module provider relationship", moduleVersion.ID())
	}

	return &dbVersion, nil
}

// UpdateModuleDetailsID updates the module details ID for a module version
func (r *ModuleVersionRepositoryImpl) UpdateModuleDetailsID(ctx context.Context, moduleVersionID int, moduleDetailsID int) error {
	db := r.GetDBFromContext(ctx)

	// CRITICAL DEBUG: Log before update
	logger := zerolog.Ctx(ctx)
	logger.Info().
		Int("module_version_id", moduleVersionID).
		Int("module_details_id", moduleDetailsID).
		Msg("UpdateModuleDetailsID: About to execute UPDATE")

	result := db.Model(&sqldb.ModuleVersionDB{}).
		Where("id = ?", moduleVersionID).
		Update("module_details_id", moduleDetailsID)

	if result.Error != nil {
		logger.Error().
			Int("module_version_id", moduleVersionID).
			Int("module_details_id", moduleDetailsID).
			Err(result.Error).
			Msg("UpdateModuleDetailsID: Database UPDATE failed")
		return fmt.Errorf("failed to update module details ID for module version %d: %w", moduleVersionID, result.Error)
	}

	// CRITICAL DEBUG: Log after update with rows affected
	logger.Info().
		Int("module_version_id", moduleVersionID).
		Int("module_details_id", moduleDetailsID).
		Int64("rows_affected", result.RowsAffected).
		Msg("UpdateModuleDetailsID: UPDATE completed successfully")

	if result.RowsAffected == 0 {
		logger.Warn().
			Int("module_version_id", moduleVersionID).
			Int("module_details_id", moduleDetailsID).
			Msg("UpdateModuleDetailsID: No rows affected - module version not found")
		return fmt.Errorf("module version %d not found", moduleVersionID)
	}

	return nil
}
