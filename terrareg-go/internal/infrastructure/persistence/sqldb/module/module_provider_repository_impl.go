package module

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/config"
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
	cfg           *config.Config
}

// NewModuleProviderRepository creates a new module provider repository
func NewModuleProviderRepository(db *gorm.DB, namespaceRepo repository.NamespaceRepository, cfg *config.Config) repository.ModuleProviderRepository {
	return &ModuleProviderRepositoryImpl{
		db:            db,
		namespaceRepo: namespaceRepo,
		cfg:           cfg,
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
	// Use custom SQL for scoring
	var sql string
	var args []interface{}

	// Base query with joins
	sql = `
		SELECT
			module_provider.id as module_provider_id,
			module_provider.namespace_id as module_provider_namespace_id,
			module_provider.module as module_provider_module,
			module_provider.provider as module_provider_provider,
			module_provider.repo_base_url_template as module_provider_repo_base_url_template,
			module_provider.repo_clone_url_template as module_provider_repo_clone_url_template,
			module_provider.repo_browse_url_template as module_provider_repo_browse_url_template,
			module_provider.git_tag_format as module_provider_git_tag_format,
			module_provider.git_path as module_provider_git_path,
			module_provider.archive_git_path as module_provider_archive_git_path,
			module_provider.verified as module_provider_verified,
			module_provider.git_provider_id as module_provider_git_provider_id,
			module_provider.latest_version_id as module_provider_latest_version_id,
			namespace.id as namespace_id,
			namespace.namespace as namespace_namespace,
			namespace.display_name as namespace_display_name,
			namespace.namespace_type as namespace_type`

	// Add scoring if query provided
	if query.Query != "" {
		sql += `,
			SUM(
				CASE
					WHEN LOWER(module_provider.module) = LOWER(?) THEN 20
					WHEN LOWER(namespace.namespace) = LOWER(?) THEN 18
					WHEN LOWER(module_provider.provider) = LOWER(?) THEN 14
					WHEN LOWER(module_version.description) = LOWER(?) THEN 13
					WHEN LOWER(module_version.owner) = LOWER(?) THEN 12
					WHEN LOWER(module_provider.module) LIKE LOWER(?) THEN 5
					WHEN LOWER(module_version.description) LIKE LOWER(?) THEN 4
					WHEN LOWER(module_version.owner) LIKE LOWER(?) THEN 3
					WHEN LOWER(namespace.namespace) LIKE LOWER(?) THEN 2
					ELSE 0
				END
			) as relevance_score`
		args = append(args,
			query.Query, query.Query, query.Query, query.Query, query.Query, // Exact matches
			"%"+query.Query+"%", "%"+query.Query+"%", "%"+query.Query+"%", "%"+query.Query+"%") // Partial matches
	}

	sql += `
		FROM module_provider
		JOIN namespace ON namespace.id = module_provider.namespace_id
		LEFT JOIN module_version ON module_version.module_provider_id = module_provider.id
			AND module_version.published = true
			AND module_version.beta = false
			AND module_version.internal = false`

	// Apply WHERE conditions
	whereConditions := []string{}
	whereArgs := []interface{}{}

	// Query filter
	if query.Query != "" {
		whereConditions = append(whereConditions, "(module_provider.module LIKE ? OR namespace.namespace LIKE ? OR module_version.description LIKE ? OR module_version.owner LIKE ?)")
		queryLower := strings.ToLower(query.Query)
		whereArgs = append(whereArgs, "%"+queryLower+"%", "%"+queryLower+"%", "%"+queryLower+"%", "%"+queryLower+"%")
	}

	// Namespace filters
	if len(query.Namespaces) > 0 {
		whereConditions = append(whereConditions, "namespace.namespace IN ?")
		whereArgs = append(whereArgs, query.Namespaces)
	}

	// Module filter
	if query.Module != nil {
		whereConditions = append(whereConditions, "module_provider.module = ?")
		whereArgs = append(whereArgs, *query.Module)
	}

	// Provider filters
	if len(query.Providers) > 0 {
		whereConditions = append(whereConditions, "module_provider.provider IN ?")
		whereArgs = append(whereArgs, query.Providers)
	}

	// Verified filter
	if query.Verified != nil {
		if *query.Verified {
			whereConditions = append(whereConditions, "module_provider.verified = ?")
			whereArgs = append(whereArgs, true)
		} else {
			whereConditions = append(whereConditions, "(module_provider.verified = ? OR module_provider.verified IS NULL)")
			whereArgs = append(whereArgs, false)
		}
	}

	// Trusted/Contributed namespace filter
	if query.TrustedNamespaces != nil || query.Contributed != nil {
		if query.TrustedNamespaces != nil && *query.TrustedNamespaces {
			// Only trusted namespaces
			if len(r.cfg.TrustedNamespaces) > 0 {
				placeholders := make([]string, len(r.cfg.TrustedNamespaces))
				for i := range r.cfg.TrustedNamespaces {
					placeholders[i] = "?"
				}
				whereConditions = append(whereConditions, "namespace.namespace IN ("+strings.Join(placeholders, ",")+")")
				for _, ns := range r.cfg.TrustedNamespaces {
					whereArgs = append(whereArgs, ns)
				}
			} else {
				// No trusted namespaces configured, return empty
				whereConditions = append(whereConditions, "1 = 0")
			}
		} else if query.Contributed != nil && *query.Contributed {
			// Only contributed (non-trusted) namespaces
			if len(r.cfg.TrustedNamespaces) > 0 {
				placeholders := make([]string, len(r.cfg.TrustedNamespaces))
				for i := range r.cfg.TrustedNamespaces {
					placeholders[i] = "?"
				}
				whereConditions = append(whereConditions, "namespace.namespace NOT IN ("+strings.Join(placeholders, ",")+")")
				for _, ns := range r.cfg.TrustedNamespaces {
					whereArgs = append(whereArgs, ns)
				}
			}
			// If no trusted namespaces configured, all are contributed
		}
	}

	// Combine WHERE conditions
	if len(whereConditions) > 0 {
		sql += " WHERE " + strings.Join(whereConditions, " AND ")
	}

	// Add GROUP BY for scoring
	if query.Query != "" {
		sql += " GROUP BY module_provider.id, namespace.id"
	}

	// Ordering
	if query.Query != "" {
		sql += " ORDER BY relevance_score DESC, module_provider.module ASC, module_provider.provider ASC"
	} else {
		orderBy := "module_provider.module"
		if query.OrderBy != "" {
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
		sql += fmt.Sprintf(" ORDER BY %s %s", orderBy, orderDir)
	}

	// Count total query
	countSQL := "SELECT COUNT(DISTINCT module_provider.id) as total FROM module_provider JOIN namespace ON namespace.id = module_provider.namespace_id LEFT JOIN module_version ON module_version.module_provider_id = module_provider.id AND module_version.published = true AND module_version.beta = false AND module_version.internal = false"
	if len(whereConditions) > 0 {
		countSQL += " WHERE " + strings.Join(whereConditions, " AND ")
	}

	// Prepare arguments for main query
	finalArgs := append(append([]interface{}{}, whereArgs...), args...)

	// Apply pagination to main query only
	if query.Limit > 0 {
		sql += " LIMIT ?"
		finalArgs = append(finalArgs, query.Limit)
	}
	if query.Offset > 0 {
		sql += " OFFSET ?"
		finalArgs = append(finalArgs, query.Offset)
	}

	// Execute count query
	var totalCount struct {
		Total int64 `json:"total"`
	}
	countArgs := append(append([]interface{}{}, whereArgs...), args...)
	if err := r.db.WithContext(ctx).Raw(countSQL, countArgs...).Scan(&totalCount).Error; err != nil {
		return nil, fmt.Errorf("failed to count module providers: %w", err)
	}

	// Execute main query
	var results []sqldb.ModuleProviderSearchResult
	if err := r.db.WithContext(ctx).Raw(sql, finalArgs...).Scan(&results).Error; err != nil {
		return nil, fmt.Errorf("failed to search module providers: %w", err)
	}

	// Convert to domain models
	providers := make([]*model.ModuleProvider, len(results))
	for i, result := range results {
		// Create namespace from result
		namespaceDB := &sqldb.NamespaceDB{
			ID:            result.NamespaceID,
			Namespace:     result.NamespaceName,
			DisplayName:   &result.NamespaceDisplayName,
			NamespaceType: sqldb.NamespaceType(result.NamespaceType),
		}
		namespace := fromDBNamespace(namespaceDB)

		// Create module provider DB from result
		moduleProviderDB := &sqldb.ModuleProviderDB{
			ID:                    result.ID,
			NamespaceID:           result.NamespaceID,
			Module:                result.Module,
			Provider:              result.Provider,
			RepoBaseURLTemplate:   result.RepoBaseURLTemplate,
			RepoCloneURLTemplate:  result.RepoCloneURLTemplate,
			RepoBrowseURLTemplate: result.RepoBrowseURLTemplate,
			GitTagFormat:          result.GitTagFormat,
			GitPath:               result.GitPath,
			ArchiveGitPath:        result.ArchiveGitPath,
			Verified:              result.Verified,
			GitProviderID:         result.GitProviderID,
			LatestVersionID:       result.LatestVersionID,
		}
		mp := fromDBModuleProvider(moduleProviderDB, namespace)

		// Set relevance score if available
		if query.Query != "" {
			mp.SetRelevanceScore(&result.RelevanceScore)
		}

		// Load versions
		versions, err := r.loadVersions(ctx, result.ID)
		if err != nil {
			return nil, err
		}
		mp.SetVersions(versions)

		providers[i] = mp
	}

	return &repository.ModuleSearchResult{
		Modules:    providers,
		TotalCount: int(totalCount.Total),
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
