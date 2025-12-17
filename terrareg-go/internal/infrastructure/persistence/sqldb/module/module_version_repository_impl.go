package module

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/module/model"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/persistence/sqldb"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/persistence/sqldb/transaction"
)

// ModuleVersionRepositoryImpl implements the module version repository using GORM
type ModuleVersionRepositoryImpl struct {
	db              *gorm.DB
	submoduleLoader *SubmoduleLoader
}

// NewModuleVersionRepository creates a new module version repository
func NewModuleVersionRepository(db *gorm.DB) *ModuleVersionRepositoryImpl {
	return &ModuleVersionRepositoryImpl{
		db:              db,
		submoduleLoader: NewSubmoduleLoader(db),
	}
}

// getDBFromContext returns the database instance from context or the default db
// This allows repositories to participate in transactions created by the service layer
func (r *ModuleVersionRepositoryImpl) getDBFromContext(ctx context.Context) *gorm.DB {
	if tx, exists := ctx.Value(transaction.TransactionDBKey).(*gorm.DB); exists {
		return tx
	}
	return r.db.WithContext(ctx)
}

// FindByModuleProvider retrieves module versions for a specific module provider
func (r *ModuleVersionRepositoryImpl) FindByModuleProvider(ctx context.Context, moduleProviderID int, includeBeta, includeUnpublished bool) ([]*model.ModuleVersion, error) {
	var moduleVersionDBs []sqldb.ModuleVersionDB
	db := r.getDBFromContext(ctx)
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

// Save persists a module version
// Note: This should NOT create its own transaction - it should participate in a transaction
// created by the service layer
func (r *ModuleVersionRepositoryImpl) Save(ctx context.Context, moduleVersion *model.ModuleVersion) error {
	if moduleVersion == nil {
		return fmt.Errorf("module version cannot be nil")
	}

	dbVersion, err := r.mapToPersistenceModel(moduleVersion)
	if err != nil {
		return fmt.Errorf("failed to map module version: %w", err)
	}

	// Get the database instance from context (participate in existing transaction) or use default
	db := r.getDBFromContext(ctx)

	return db.Save(dbVersion).Error
}

// FindByID retrieves a module version by ID
func (r *ModuleVersionRepositoryImpl) FindByID(ctx context.Context, id int) (*model.ModuleVersion, error) {
	var dbVersion sqldb.ModuleVersionDB
	db := r.getDBFromContext(ctx)
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
	err := r.db.WithContext(ctx).
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
	return r.db.WithContext(ctx).Delete(&sqldb.ModuleVersionDB{}, id).Error
}

// Exists checks if a module version exists
func (r *ModuleVersionRepositoryImpl) Exists(ctx context.Context, moduleProviderID int, version string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
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
		err := r.db.First(&detailsDB, *dbVersion.ModuleDetailsID).Error
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

	// Load submodules and examples using shared service
	if err := r.submoduleLoader.LoadSubmodulesAndExamples(moduleVersion, dbVersion.ID); err != nil {
		return nil, err
	}

	return moduleVersion, nil
}

// mapToPersistenceModel converts domain model to persistence model using centralized mapper
func (r *ModuleVersionRepositoryImpl) mapToPersistenceModel(moduleVersion *model.ModuleVersion) (sqldb.ModuleVersionDB, error) {
	return toDBModuleVersion(moduleVersion), nil
}
