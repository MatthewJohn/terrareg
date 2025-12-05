package module

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/module/model"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/persistence/sqldb"
)

// ModuleVersionRepositoryImpl implements the module version repository using GORM
type ModuleVersionRepositoryImpl struct {
	db *gorm.DB
}

// NewModuleVersionRepository creates a new module version repository
func NewModuleVersionRepository(db *gorm.DB) *ModuleVersionRepositoryImpl {
	return &ModuleVersionRepositoryImpl{db: db}
}

// FindByModuleProvider retrieves module versions for a specific module provider
func (r *ModuleVersionRepositoryImpl) FindByModuleProvider(ctx context.Context, moduleProviderID int, includeBeta, includeUnpublished bool) ([]*model.ModuleVersion, error) {
	var moduleVersionDBs []sqldb.ModuleVersionDB
	query := r.db.WithContext(ctx).Where("module_provider_id = ?", moduleProviderID)

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
func (r *ModuleVersionRepositoryImpl) Save(ctx context.Context, moduleVersion *model.ModuleVersion) error {
	if moduleVersion == nil {
		return fmt.Errorf("module version cannot be nil")
	}

	dbVersion, err := r.mapToPersistenceModel(moduleVersion)
	if err != nil {
		return fmt.Errorf("failed to map module version: %w", err)
	}

	return r.db.WithContext(ctx).Save(dbVersion).Error
}

// FindByID retrieves a module version by ID
func (r *ModuleVersionRepositoryImpl) FindByID(ctx context.Context, id int) (*model.ModuleVersion, error) {
	var dbVersion sqldb.ModuleVersionDB
	err := r.db.WithContext(ctx).First(&dbVersion, id).Error
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

// mapToDomainModel converts persistence model to domain model
func (r *ModuleVersionRepositoryImpl) mapToDomainModel(dbVersion sqldb.ModuleVersionDB) (*model.ModuleVersion, error) {
	// Create simple module details for initial setup
	details := model.NewModuleDetails([]byte{}) // Empty readme for initial setup

	// Determine if version is published
	published := dbVersion.Published != nil && *dbVersion.Published

	// Handle nil PublishedAt pointer
	var publishedAt time.Time
	if dbVersion.PublishedAt != nil {
		publishedAt = *dbVersion.PublishedAt
	}

	return model.ReconstructModuleVersion(
		dbVersion.ID,
		dbVersion.Version,
		details,
		dbVersion.Beta,
		dbVersion.Internal,
		published,
		dbVersion.PublishedAt,
		dbVersion.GitSHA,
		dbVersion.GitPath,
		dbVersion.ArchiveGitPath,
		dbVersion.RepoBaseURLTemplate,
		dbVersion.RepoCloneURLTemplate,
		dbVersion.RepoBrowseURLTemplate,
		dbVersion.Owner,
		dbVersion.Description,
		dbVersion.VariableTemplate,
		dbVersion.ExtractionVersion,
		publishedAt, // Use publishedAt as createdAt for simplicity
		publishedAt, // Use publishedAt as updatedAt for simplicity
	)
}

// mapToPersistenceModel converts domain model to persistence model
func (r *ModuleVersionRepositoryImpl) mapToPersistenceModel(moduleVersion *model.ModuleVersion) (sqldb.ModuleVersionDB, error) {
	dbVersion := sqldb.ModuleVersionDB{
		ID:                   moduleVersion.ID(),
		ModuleProviderID:     moduleVersion.ModuleProvider().ID(),
		Version:              moduleVersion.Version().String(),
		Beta:                 moduleVersion.IsBeta(),
		Internal:             moduleVersion.IsInternal(),
		ArchiveGitPath:       false, // Default value
		VariableTemplate:     nil,  // Default value
	}

	// Set published field
	published := moduleVersion.IsPublished()
	dbVersion.Published = &published

	// Set timestamps (simplified for initial setup)
	if moduleVersion.PublishedAt() != nil {
		dbVersion.PublishedAt = moduleVersion.PublishedAt()
	}

	return dbVersion, nil
}