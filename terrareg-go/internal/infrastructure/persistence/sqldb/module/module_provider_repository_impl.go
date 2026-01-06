package module

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"

	configModel "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/config/model"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/module/model"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/module/repository"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/shared"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/persistence/sqldb"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/persistence/sqldb/git"
	baserepo "github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/persistence/sqldb/repository"
)

// ModuleProviderRepositoryImpl implements ModuleProviderRepository using GORM
type ModuleProviderRepositoryImpl struct {
	*baserepo.BaseRepository
	namespaceRepo   repository.NamespaceRepository
	domainConfig    *configModel.DomainConfig
	submoduleLoader *SubmoduleLoader
}

// NewModuleProviderRepository creates a new module provider repository
func NewModuleProviderRepository(db *gorm.DB, namespaceRepo repository.NamespaceRepository, domainConfig *configModel.DomainConfig) repository.ModuleProviderRepository {
	return &ModuleProviderRepositoryImpl{
		BaseRepository:  baserepo.NewBaseRepository(db),
		namespaceRepo:   namespaceRepo,
		domainConfig:    domainConfig,
		submoduleLoader: NewSubmoduleLoader(db),
	}
}

// Save persists a module provider (aggregate root)
// Note: This should NOT create its own transaction - it should participate in a transaction
// created by the service layer
func (r *ModuleProviderRepositoryImpl) Save(ctx context.Context, mp *model.ModuleProvider) error {
	// Get the database instance from context (participate in existing transaction) or use default
	db := r.GetDBFromContext(ctx)

	dbModel := toDBModuleProvider(mp)

	var err error
	if mp.ID() == 0 {
		// Create
		err = db.Create(&dbModel).Error
		if err != nil {
			return fmt.Errorf("failed to create module provider: %w", err)
		}
	} else {
		// Update
		err = db.Save(&dbModel).Error
		if err != nil {
			return fmt.Errorf("failed to update module provider: %w", err)
		}
	}

	// Save versions (entities within aggregate)
	for _, version := range mp.GetAllVersions() {
		if err := r.saveVersion(db, version); err != nil {
			return err
		}
	}

	// TODO: Update the domain model with the database-generated ID if needed
	// This would require a method in the domain model to set the ID
	return nil
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

	err := r.GetDBFromContext(ctx).
		Preload("Namespace").
		Preload("GitProvider").
		First(&dbModel, id).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, shared.ErrNotFound
		}
		return nil, fmt.Errorf("failed to find module provider: %w", err)
	}

	return r.toDomain(ctx, &dbModel)
}

// FindByNamespaceModuleProvider retrieves a module provider by namespace/module/provider
func (r *ModuleProviderRepositoryImpl) FindByNamespaceModuleProvider(ctx context.Context, namespace, module, provider string) (*model.ModuleProvider, error) {
	var dbModel sqldb.ModuleProviderDB

	// DEBUG: Log context deadline
	if deadline, ok := ctx.Deadline(); ok {
		fmt.Printf("[DEBUG] FindByNamespaceModuleProvider: context deadline = %v, now = %v, time remaining = %v\n",
			deadline, time.Now(), deadline.Sub(time.Now()))
	} else {
		fmt.Printf("[DEBUG] FindByNamespaceModuleProvider: no context deadline\n")
	}

	// Use background context for DB query to avoid HTTP request timeout issues
	dbCtx := context.Background()
	if _, hasTx := sqldb.GetTxFromContext(ctx); hasTx {
		// If in a transaction, keep the original context
		dbCtx = ctx
	}

	err := r.GetDBFromContext(dbCtx).
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

	return r.toDomain(ctx, &dbModel)
}

// FindByNamespace retrieves all module providers in a namespace
func (r *ModuleProviderRepositoryImpl) FindByNamespace(ctx context.Context, namespace string) ([]*model.ModuleProvider, error) {
	var dbModels []sqldb.ModuleProviderDB

	err := r.GetDBFromContext(ctx).
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
		mp, err := r.toDomain(ctx, &dbModel)
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
	// Split query into terms (matching Python's query.split())
	queryTerms := strings.Fields(query.Query)
	if len(queryTerms) > 0 {
		// Build scoring for each term and sum them
		sql += `, (`
		for i, term := range queryTerms {
			if i > 0 {
				sql += " + "
			}
			sql += fmt.Sprintf(`(
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
			)`)
			// Add exact match and partial match arguments for this term
			args = append(args,
				term, term, term, term, term, // Exact matches
				"%"+term+"%", "%"+term+"%", "%"+term+"%", "%"+term+"%") // Partial matches
		}
		sql += `) as relevance_score`
	}

	sql += `
		FROM module_provider
		JOIN namespace ON namespace.id = module_provider.namespace_id
		LEFT JOIN module_version ON module_version.id = module_provider.latest_version_id`

	// Apply WHERE conditions
	whereConditions := []string{}
	whereArgs := []interface{}{}

	// Query filter - handle multiple query terms
	if len(queryTerms) > 0 {
		// Build OR conditions for each query term (matching Python behavior)
		termConditions := []string{}
		for _, term := range queryTerms {
			termConditions = append(termConditions, "(module_provider.module LIKE ? OR namespace.namespace LIKE ? OR module_version.description LIKE ? OR module_version.owner LIKE ?)")
			termLower := strings.ToLower(term)
			whereArgs = append(whereArgs, "%"+termLower+"%", "%"+termLower+"%", "%"+termLower+"%", "%"+termLower+"%")
		}
		// Combine all term conditions with OR - match if ANY term matches
		whereConditions = append(whereConditions, "("+strings.Join(termConditions, " OR ")+")")
	}

	// Namespace filters
	if len(query.Namespaces) > 0 {
		placeholders := make([]string, len(query.Namespaces))
		for i := range query.Namespaces {
			placeholders[i] = "?"
		}
		whereConditions = append(whereConditions, "namespace.namespace IN ("+strings.Join(placeholders, ",")+")")
		for _, ns := range query.Namespaces {
			whereArgs = append(whereArgs, ns)
		}
	}

	// Module filter
	if query.Module != nil {
		whereConditions = append(whereConditions, "module_provider.module = ?")
		whereArgs = append(whereArgs, *query.Module)
	}

	// Provider filters
	if len(query.Providers) > 0 {
		placeholders := make([]string, len(query.Providers))
		for i := range query.Providers {
			placeholders[i] = "?"
		}
		whereConditions = append(whereConditions, "module_provider.provider IN ("+strings.Join(placeholders, ",")+")")
		for _, p := range query.Providers {
			whereArgs = append(whereArgs, p)
		}
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
	if (query.TrustedNamespaces != nil && *query.TrustedNamespaces) || (query.Contributed != nil && *query.Contributed) {
		orConditions := []string{}
		if query.TrustedNamespaces != nil && *query.TrustedNamespaces && len(r.domainConfig.TrustedNamespaces) > 0 {
			placeholders := make([]string, len(r.domainConfig.TrustedNamespaces))
			for i := range r.domainConfig.TrustedNamespaces {
				placeholders[i] = "?"
			}
			orConditions = append(orConditions, "namespace.namespace IN ("+strings.Join(placeholders, ",")+")")
			for _, ns := range r.domainConfig.TrustedNamespaces {
				whereArgs = append(whereArgs, ns)
			}
		}
		if query.Contributed != nil && *query.Contributed && len(r.domainConfig.TrustedNamespaces) > 0 {
			placeholders := make([]string, len(r.domainConfig.TrustedNamespaces))
			for i := range r.domainConfig.TrustedNamespaces {
				placeholders[i] = "?"
			}
			orConditions = append(orConditions, "namespace.namespace NOT IN ("+strings.Join(placeholders, ",")+")")
			for _, ns := range r.domainConfig.TrustedNamespaces {
				whereArgs = append(whereArgs, ns)
			}
		}
		// If no trusted namespaces configured but filters are requested, handle appropriately
		if len(r.domainConfig.TrustedNamespaces) == 0 {
			if query.TrustedNamespaces != nil && *query.TrustedNamespaces {
				// No trusted namespaces configured and user wants trusted only - return empty
				whereConditions = append(whereConditions, "1 = 0")
			}
			// If contributed only with no trusted namespaces, all modules are contributed - no filter needed
		} else if len(orConditions) > 0 {
			// Combine OR conditions
			whereConditions = append(whereConditions, "("+strings.Join(orConditions, " OR ")+")")
		}
	}

	// Combine WHERE conditions
	if len(whereConditions) > 0 {
		sql += " WHERE " + strings.Join(whereConditions, " AND ")
	}

	// Add GROUP BY for scoring
	if len(queryTerms) > 0 {
		sql += " GROUP BY module_provider.id, namespace.id"
	}

	// Ordering
	if len(queryTerms) > 0 {
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
	countSQL := "SELECT COUNT(DISTINCT module_provider.id) as total FROM module_provider JOIN namespace ON namespace.id = module_provider.namespace_id LEFT JOIN module_version ON module_version.id = module_provider.latest_version_id"
	if len(whereConditions) > 0 {
		// Add condition to ensure we only count module_providers with a latest version
		// This prevents counting modules without any published non-beta versions
		countSQL += " WHERE module_version.id IS NOT NULL AND " + strings.Join(whereConditions, " AND ")
	} else {
		// Even without search filters, only count modules with a latest version
		countSQL += " WHERE module_version.id IS NOT NULL"
	}

	// Prepare arguments for main query - scoring args first (SELECT clause), then WHERE args
	finalArgs := append(append([]interface{}{}, args...), whereArgs...)

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
	// Count query only uses WHERE args, not the scoring args (since COUNT query doesn't have scoring SELECT clause)
	countArgs := append([]interface{}{}, whereArgs...)
	if err := r.GetDBFromContext(ctx).Raw(countSQL, countArgs...).Scan(&totalCount).Error; err != nil {
		return nil, fmt.Errorf("failed to count module providers: %w", err)
	}

	// Execute main query
	var results []sqldb.ModuleProviderSearchResult
	if err := r.GetDBFromContext(ctx).Raw(sql, finalArgs...).Scan(&results).Error; err != nil {
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
		if len(queryTerms) > 0 {
			mp.SetRelevanceScore(result.RelevanceScore)
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
	result := r.GetDBFromContext(ctx).Delete(&sqldb.ModuleProviderDB{}, id)
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
	err := r.GetDBFromContext(ctx).
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
func (r *ModuleProviderRepositoryImpl) toDomain(ctx context.Context, db *sqldb.ModuleProviderDB) (*model.ModuleProvider, error) {
	namespace := fromDBNamespace(&db.Namespace)

	// Use the new mapper function
	mp := fromDBModuleProvider(db, namespace)

	if db.GitProvider != nil {
		gitProvider := git.FromDBGitProvider(db.GitProvider)
		mp.SetGitProvider(gitProvider)
	}

	// Load versions
	versions, err := r.loadVersions(ctx, db.ID)
	if err != nil {
		return nil, err
	}
	mp.SetVersions(versions)

	return mp, nil
}

// loadVersions loads all versions for a module provider
func (r *ModuleProviderRepositoryImpl) loadVersions(ctx context.Context, moduleProviderID int) ([]*model.ModuleVersion, error) {
	var dbVersions []sqldb.ModuleVersionDB

	err := r.GetDBFromContext(ctx).
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
	moduleVersion, err := model.ReconstructModuleVersion(
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
	if err != nil {
		return nil, err
	}

	// Load submodules and examples using shared service
	if err := r.submoduleLoader.LoadSubmodulesAndExamples(moduleVersion, db.ID); err != nil {
		return nil, err
	}

	return moduleVersion, nil
}
