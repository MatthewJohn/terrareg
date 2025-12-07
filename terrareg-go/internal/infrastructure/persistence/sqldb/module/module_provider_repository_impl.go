package module

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/module/model"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/module/repository"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/shared"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/persistence/sqldb"
	gitmapper "github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/persistence/sqldb/git"
)

// ModuleProviderRepositoryImpl implements ModuleProviderRepository using GORM
type ModuleProviderRepositoryImpl struct {
	db            *gorm.DB
	namespaceRepo repository.NamespaceRepository
}

// NewModuleProviderRepository creates a new module provider repository
func NewModuleProviderRepository(db *gorm.DB, namespaceRepo repository.NamespaceRepository) repository.ModuleProviderRepository {
	return &ModuleProviderRepositoryImpl{
		db:            db,
		namespaceRepo: namespaceRepo,
	}
}

// Save persists a module provider (aggregate root)
func (r *ModuleProviderRepositoryImpl) Save(ctx context.Context, mp *model.ModuleProvider) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		dbModel := toDBModuleProvider(mp)

		var err error
		if mp.ID() == 0 {
			// Create
			err = tx.Create(&dbModel).Error
			if err != nil {
				return fmt.Errorf("failed to create module provider: %w", err)
			}
		} else {
			// Update
			err = tx.Save(&dbModel).Error
			if err != nil {
				return fmt.Errorf("failed to update module provider: %w", err)
			}
		}

		// Save versions (entities within aggregate)
		for _, version := range mp.GetAllVersions() {
			if err := r.saveVersion(tx, version); err != nil {
				return err
			}
		}

		return nil
	})
}

// saveVersion saves a module version
func (r *ModuleProviderRepositoryImpl) saveVersion(tx *gorm.DB, version *model.ModuleVersion) error {
	// First save the details if present
	var detailsID *int
	if version.Details() != nil {
		dbDetails := toDBModuleDetails(version.Details())

		if version.ID() != 0 {
			// Try to find existing details
			var existing sqldb.ModuleVersionDB
			tx.Select("module_details_id").Where("id = ?", version.ID()).First(&existing)
			if existing.ModuleDetailsID != nil {
				dbDetails.ID = *existing.ModuleDetailsID
			}
		}

		if err := tx.Save(&dbDetails).Error; err != nil {
			return fmt.Errorf("failed to save module details: %w", err)
		}
		detailsID = &dbDetails.ID
	}

	// Save the version
	dbVersion := toDBModuleVersion(version)
	dbVersion.ModuleDetailsID = detailsID

	if version.ID() == 0 {
		if err := tx.Create(&dbVersion).Error; err != nil {
			return fmt.Errorf("failed to create module version: %w", err)
		}
	} else {
		if err := tx.Save(&dbVersion).Error; err != nil {
			return fmt.Errorf("failed to update module version: %w", err)
		}
	}

	return nil
}

// FindByID retrieves a module provider by ID
func (r *ModuleProviderRepositoryImpl) FindByID(ctx context.Context, id int) (*model.ModuleProvider, error) {
	var dbModel sqldb.ModuleProviderDB

	err := r.db.WithContext(ctx).
		Preload("Namespace").
		Preload("GitProvider").
		First(&dbModel, id).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, shared.ErrNotFound
		}
		return nil, fmt.Errorf("failed to find module provider: %w", err)
	}

	return r.toDomain(&dbModel)
}

// FindByNamespaceModuleProvider retrieves a module provider by namespace/module/provider
func (r *ModuleProviderRepositoryImpl) FindByNamespaceModuleProvider(ctx context.Context, namespace, module, provider string) (*model.ModuleProvider, error) {
	var dbModel sqldb.ModuleProviderDB

	err := r.db.WithContext(ctx).
		Joins("JOIN namespace ON namespace.id = module_provider.namespace_id").
		Where("namespace.namespace = ? AND module_provider.module = ? AND module_provider.provider = ?", namespace, module, provider).
		Preload("Namespace").
		Preload("GitProvider").
		First(&dbModel).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, shared.ErrNotFound
		}
		return nil, fmt.Errorf("failed to find module provider: %w", err)
	}

	return r.toDomain(&dbModel)
}

// FindByNamespace retrieves all module providers in a namespace
func (r *ModuleProviderRepositoryImpl) FindByNamespace(ctx context.Context, namespace string) ([]*model.ModuleProvider, error) {
	var dbModels []sqldb.ModuleProviderDB

	err := r.db.WithContext(ctx).
		Joins("JOIN namespace ON namespace.id = module_provider.namespace_id").
		Where("namespace.namespace = ?", namespace).
		Preload("Namespace").
		Order("module_provider.module ASC, module_provider.provider ASC").
		Find(&dbModels).Error

	if err != nil {
		return nil, fmt.Errorf("failed to find module providers: %w", err)
	}

	providers := make([]*model.ModuleProvider, len(dbModels))
	for i, dbModel := range dbModels {
		mp, err := r.toDomain(&dbModel)
		if err != nil {
			return nil, err
		}
		providers[i] = mp
	}

	return providers, nil
}

// Search searches for module providers
func (r *ModuleProviderRepositoryImpl) Search(ctx context.Context, query repository.ModuleSearchQuery) (*repository.ModuleSearchResult, error) {
	db := r.db.WithContext(ctx).
		Joins("JOIN namespace ON namespace.id = module_provider.namespace_id").
		Preload("Namespace").
		Preload("GitProvider")

	// Apply filters
	if query.Query != "" {
		db = db.Where("module_provider.module LIKE ? OR namespace.namespace LIKE ?",
			"%"+query.Query+"%", "%"+query.Query+"%")
	}

	// Apply multiple namespace filters
	if len(query.Namespaces) > 0 {
		db = db.Where("namespace.namespace IN ?", query.Namespaces)
	}

	if query.Module != nil {
		db = db.Where("module_provider.module = ?", *query.Module)
	}

	// Apply multiple provider filters
	if len(query.Providers) > 0 {
		db = db.Where("module_provider.provider IN ?", query.Providers)
	}

	if query.Verified != nil {
		if *query.Verified {
			// Filter for verified modules (verified = 1)
			db = db.Where("module_provider.verified = ?", true)
		} else {
			// Filter for unverified modules (verified = 0 or NULL)
			db = db.Where("module_provider.verified = ? OR module_provider.verified IS NULL", false)
		}
	}

	// Note: trusted/contributed filtering will be handled at the application layer
	// since trusted status is not stored in the database but configured via environment

	// Count total - use proper table qualification to avoid ambiguous column error
	countDB := r.db.WithContext(ctx).
		Table("module_provider").
		Joins("JOIN namespace ON namespace.id = module_provider.namespace_id")

	// Apply all the same filters to countDB
	if query.Query != "" {
		countDB = countDB.Where("module_provider.module LIKE ? OR namespace.namespace LIKE ?",
			"%"+query.Query+"%", "%"+query.Query+"%")
	}

	// Apply multiple namespace filters to count
	if len(query.Namespaces) > 0 {
		countDB = countDB.Where("namespace.namespace IN ?", query.Namespaces)
	}

	if query.Module != nil {
		countDB = countDB.Where("module_provider.module = ?", *query.Module)
	}

	// Apply multiple provider filters to count
	if len(query.Providers) > 0 {
		countDB = countDB.Where("module_provider.provider IN ?", query.Providers)
	}

	if query.Verified != nil {
		if *query.Verified {
			// Filter for verified modules (verified = 1)
			countDB = countDB.Where("module_provider.verified = ?", true)
		} else {
			// Filter for unverified modules (verified = 0 or NULL)
			countDB = countDB.Where("module_provider.verified = ? OR module_provider.verified IS NULL", false)
		}
	}

	// Note: trusted/contributed filtering handled in application layer
	// since it depends on environment configuration, not database schema

	// Now count with proper qualification
	var total int64
	if err := countDB.Count(&total).Error; err != nil {
		return nil, fmt.Errorf("failed to count module providers: %w", err)
	}

	// Apply ordering
	orderBy := "module_provider.module"
	if query.OrderBy != "" {
		// Ensure column names are qualified with table name to avoid ambiguity
		if query.OrderBy == "id" {
			orderBy = "module_provider.id"
		} else if query.OrderBy == "namespace" {
			orderBy = "namespace.namespace"
		} else {
			orderBy = query.OrderBy
		}
	}
	orderDir := "ASC"
	if query.OrderDir != "" {
		orderDir = query.OrderDir
	}
	db = db.Order(fmt.Sprintf("%s %s", orderBy, orderDir))

	// Apply pagination
	if query.Limit > 0 {
		db = db.Limit(query.Limit)
	}
	if query.Offset > 0 {
		db = db.Offset(query.Offset)
	}

	var dbModels []sqldb.ModuleProviderDB
	if err := db.Find(&dbModels).Error; err != nil {
		return nil, fmt.Errorf("failed to search module providers: %w", err)
	}

	providers := make([]*model.ModuleProvider, len(dbModels))
	for i, dbModel := range dbModels {
		mp, err := r.toDomain(&dbModel)
		if err != nil {
			return nil, err
		}
		providers[i] = mp
	}

	return &repository.ModuleSearchResult{
		Modules:    providers,
		TotalCount: int(total),
	}, nil
}

// Delete removes a module provider
func (r *ModuleProviderRepositoryImpl) Delete(ctx context.Context, id int) error {
	result := r.db.WithContext(ctx).Delete(&sqldb.ModuleProviderDB{}, id)
	if result.Error != nil {
		return fmt.Errorf("failed to delete module provider: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return shared.ErrNotFound
	}

	return nil
}

// Exists checks if a module provider exists
func (r *ModuleProviderRepositoryImpl) Exists(ctx context.Context, namespace, module, provider string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&sqldb.ModuleProviderDB{}).
		Joins("JOIN namespace ON namespace.id = module_provider.namespace_id").
		Where("namespace.namespace = ? AND module_provider.module = ? AND module_provider.provider = ?", namespace, module, provider).
		Count(&count).Error

	if err != nil {
		return false, fmt.Errorf("failed to check module provider existence: %w", err)
	}

	return count > 0, nil
}

// toDomain converts database model to domain model
func (r *ModuleProviderRepositoryImpl) toDomain(db *sqldb.ModuleProviderDB) (*model.ModuleProvider, error) {
	namespace := fromDBNamespace(&db.Namespace)

	// Use the new mapper function
	mp := fromDBModuleProvider(db, namespace)

	if db.GitProvider != nil {
		gitProvider := gitmapper.FromDBGitProvider(db.GitProvider)
		mp.SetGitProvider(gitProvider)
	}

	// Load versions
	versions, err := r.loadVersions(context.Background(), db.ID)
	if err != nil {
		return nil, err
	}
	mp.SetVersions(versions)

	return mp, nil
}

// loadVersions loads all versions for a module provider
func (r *ModuleProviderRepositoryImpl) loadVersions(ctx context.Context, moduleProviderID int) ([]*model.ModuleVersion, error) {
	var dbVersions []sqldb.ModuleVersionDB

	err := r.db.WithContext(ctx).
		Where("module_provider_id = ?", moduleProviderID).
		Preload("ModuleDetails").
		Order("version DESC").
		Find(&dbVersions).Error

	if err != nil {
		return nil, fmt.Errorf("failed to load versions: %w", err)
	}

	versions := make([]*model.ModuleVersion, len(dbVersions))
	for i, dbVersion := range dbVersions {
		v, err := r.toVersionDomain(&dbVersion)
		if err != nil {
			return nil, err
		}
		versions[i] = v
	}

	return versions, nil
}

// toVersionDomain converts database version to domain version
func (r *ModuleProviderRepositoryImpl) toVersionDomain(db *sqldb.ModuleVersionDB) (*model.ModuleVersion, error) {
	var details *model.ModuleDetails
	if db.ModuleDetails != nil {
		details = fromDBModuleDetails(db.ModuleDetails)
	}

	// Use current time for now since database doesn't store timestamps
	now := time.Now()
	return model.ReconstructModuleVersion(
		db.ID,
		db.Version,
		details,
		db.Beta,
		db.Internal,
		db.Published != nil && *db.Published,
		db.PublishedAt,
		db.GitSHA,
		db.GitPath,
		db.ArchiveGitPath,
		db.RepoBaseURLTemplate,
		db.RepoCloneURLTemplate,
		db.RepoBrowseURLTemplate,
		db.Owner,
		db.Description,
		db.VariableTemplate,
		db.ExtractionVersion,
		now,
		now,
	)
}
