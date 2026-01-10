package provider

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/provider"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/provider/repository"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/shared"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/persistence/sqldb"
)

// ProviderRepository implements the provider repository interface using GORM
// Uses main sqldb models (ProviderDB, ProviderVersionDB, etc.) for DDD consistency
type ProviderRepository struct {
	db *gorm.DB
}

// NewProviderRepository creates a new provider repository
func NewProviderRepository(db *gorm.DB) repository.ProviderRepository {
	return &ProviderRepository{
		db: db,
	}
}

// FindAll retrieves all providers with pagination, including namespace names and version data
func (r *ProviderRepository) FindAll(ctx context.Context, offset, limit int) ([]*provider.Provider, map[int]string, map[int]repository.VersionData, int, error) {
	// Use custom SQL query similar to Search() to collect namespace names and version data
	sql := `
		SELECT
			provider.id as provider_id,
			provider.namespace_id as provider_namespace_id,
			provider.name as provider_name,
			provider.description as provider_description,
			provider.tier as provider_tier,
			provider.default_provider_source_auth as provider_default_provider_source_auth,
			provider.provider_category_id as provider_provider_category_id,
			provider.repository_id as provider_repository_id,
			provider.latest_version_id as provider_latest_version_id,
			namespace.id as namespace_id,
			namespace.namespace as namespace_namespace,
			namespace.display_name as namespace_display_name,
			namespace.namespace_type as namespace_type,
			provider_version.id as version_id,
			provider_version.version as version_version,
			provider_version.git_tag as version_git_tag,
			provider_version.published_at as version_published_at,
			repository.owner as repository_owner,
			repository.description as repository_description,
			repository.clone_url as repository_clone_url,
			repository.logo_url as repository_logo_url,
			COALESCE(analytics_counts.download_count, 0) as download_count
		FROM provider
		JOIN namespace ON namespace.id = provider.namespace_id
		LEFT JOIN provider_version ON provider_version.id = provider.latest_version_id
		LEFT JOIN repository ON repository.id = provider.repository_id
		LEFT JOIN (
			SELECT parent_module_version, COUNT(*) as download_count
			FROM analytics
			WHERE parent_module_version IS NOT NULL
			GROUP BY parent_module_version
		) analytics_counts ON analytics_counts.parent_module_version = provider.latest_version_id
		WHERE provider.latest_version_id IS NOT NULL
		  AND provider.latest_version_id > 0
		GROUP BY provider.id, namespace.id
		ORDER BY provider.id DESC`

	// Apply pagination
	if limit > 0 {
		sql += " LIMIT ?"
		if offset > 0 {
			sql += " OFFSET ?"
		}
	}

	// Count total query
	countSQL := `
		SELECT COUNT(DISTINCT provider.id) as total
		FROM provider
		JOIN namespace ON namespace.id = provider.namespace_id
		WHERE provider.latest_version_id IS NOT NULL
		  AND provider.latest_version_id > 0`

	// Execute count query
	var totalCount struct {
		Total int64 `json:"total"`
	}
	if err := r.db.WithContext(ctx).Raw(countSQL).Scan(&totalCount).Error; err != nil {
		return nil, nil, nil, 0, fmt.Errorf("failed to count providers: %w", err)
	}

	// Prepare arguments for main query
	args := []interface{}{}
	if limit > 0 {
		args = append(args, limit)
		if offset > 0 {
			args = append(args, offset)
		}
	}

	// Execute main query
	type FindAllResult struct {
		ProviderID                 int      `gorm:"column:provider_id"`
		NamespaceID                int      `gorm:"column:provider_namespace_id"`
		Name                       string   `gorm:"column:provider_name"`
		Description                *string  `gorm:"column:provider_description"`
		Tier                       string   `gorm:"column:provider_tier"`
		DefaultProviderSourceAuth *bool    `gorm:"column:provider_default_provider_source_auth"`
		ProviderCategoryID         *int     `gorm:"column:provider_provider_category_id"`
		RepositoryID               *int     `gorm:"column:provider_repository_id"`
		LatestVersionID            *int     `gorm:"column:provider_latest_version_id"`
		NamespaceName              string   `gorm:"column:namespace_namespace"`
		NamespaceDisplayName       *string  `gorm:"column:namespace_display_name"`
		NamespaceType              string   `gorm:"column:namespace_type"`
		VersionID                  *int     `gorm:"column:version_id"`
		Version                    *string  `gorm:"column:version_version"`
		GitTag                     *string  `gorm:"column:version_git_tag"`
		PublishedAt                *string  `gorm:"column:version_published_at"`
		RepositoryOwner            *string  `gorm:"column:repository_owner"`
		RepositoryDescription      *string  `gorm:"column:repository_description"`
		RepositoryCloneURL         *string  `gorm:"column:repository_clone_url"`
		RepositoryLogoURL          *string  `gorm:"column:repository_logo_url"`
		DownloadCount              int64    `gorm:"column:download_count"`
	}

	var results []FindAllResult
	if err := r.db.WithContext(ctx).Raw(sql, args...).Scan(&results).Error; err != nil {
		return nil, nil, nil, 0, fmt.Errorf("failed to find providers: %w", err)
	}

	// Convert to domain entities and collect namespace names and version data
	providers := make([]*provider.Provider, len(results))
	namespaceNames := make(map[int]string, len(results))
	versionData := make(map[int]repository.VersionData, len(results))

	for i, result := range results {
		// Store namespace name for this provider
		namespaceNames[result.ProviderID] = result.NamespaceName

		// Store version data for this provider (if available)
		if result.VersionID != nil {
			versionStr := ""
			if result.Version != nil {
				versionStr = *result.Version
			}
			versionData[result.ProviderID] = repository.VersionData{
				VersionID:             *result.VersionID,
				Version:               versionStr,
				GitTag:                result.GitTag,
				PublishedAt:           result.PublishedAt,
				RepositoryOwner:       result.RepositoryOwner,
				RepositoryDescription: result.RepositoryDescription,
				RepositoryCloneURL:    result.RepositoryCloneURL,
				RepositoryLogoURL:     result.RepositoryLogoURL,
				Downloads:             result.DownloadCount,
			}
		}

		// Handle nullable DefaultProviderSourceAuth
		defaultProviderSourceAuth := false
		if result.DefaultProviderSourceAuth != nil {
			defaultProviderSourceAuth = *result.DefaultProviderSourceAuth
		}

		// Create provider (using Reconstruct to match existing pattern)
		prov := provider.ReconstructProvider(
			result.ProviderID,
			result.NamespaceID,
			result.Name,
			result.Description,
			result.Tier,
			result.ProviderCategoryID,
			result.RepositoryID,
			result.LatestVersionID,
			defaultProviderSourceAuth,
		)
		providers[i] = prov
	}

	return providers, namespaceNames, versionData, int(totalCount.Total), nil
}

// Search searches for providers by query with filters
func (r *ProviderRepository) Search(ctx context.Context, params repository.ProviderSearchQuery) (*repository.ProviderSearchResult, error) {
	// Use custom SQL for scoring (matching Python's provider search implementation)
	var sql string
	var args []interface{}

	// Split query into terms (matching Python's query.split())
	queryTerms := strings.Fields(params.Query)

	// Base query with joins
	sql = `
		SELECT
			provider.id as provider_id,
			provider.namespace_id as provider_namespace_id,
			provider.name as provider_name,
			provider.description as provider_description,
			provider.tier as provider_tier,
			provider.default_provider_source_auth as provider_default_provider_source_auth,
			provider.provider_category_id as provider_provider_category_id,
			provider.repository_id as provider_repository_id,
			provider.latest_version_id as provider_latest_version_id,
			namespace.id as namespace_id,
			namespace.namespace as namespace_namespace,
			namespace.display_name as namespace_display_name,
			namespace.namespace_type as namespace_type,
			provider_version.id as version_id,
			provider_version.version as version_version,
			provider_version.git_tag as version_git_tag,
			provider_version.published_at as version_published_at,
			repository.owner as repository_owner,
			repository.description as repository_description,
			repository.clone_url as repository_clone_url,
			repository.logo_url as repository_logo_url,
			COALESCE(analytics_counts.download_count, 0) as download_count`

	// Add scoring if query provided
	if len(queryTerms) > 0 {
		// Build scoring for each term and sum them
		sql += `, (`
		for i, term := range queryTerms {
			if i > 0 {
				sql += " + "
			}
			sql += `(
				CASE
					WHEN LOWER(provider.name) = LOWER(?) THEN 20
					WHEN LOWER(namespace.namespace) = LOWER(?) THEN 18
					WHEN LOWER(provider.description) = LOWER(?) THEN 13
					WHEN LOWER(provider.name) LIKE LOWER(?) THEN 5
					WHEN LOWER(provider.description) LIKE LOWER(?) THEN 4
					WHEN LOWER(namespace.namespace) LIKE LOWER(?) THEN 2
					ELSE 0
				END
			)`
			args = append(args, term, term, term, "%"+term+"%", "%"+term+"%", "%"+term+"%")
		}
		sql += `) as relevance_score`
	}

	// FROM clause with joins
	sql += `
		FROM provider
		JOIN namespace ON namespace.id = provider.namespace_id
		LEFT JOIN provider_version ON provider_version.id = provider.latest_version_id
		LEFT JOIN repository ON repository.id = provider.repository_id
		LEFT JOIN (
			SELECT parent_module_version, COUNT(*) as download_count
			FROM analytics
			WHERE parent_module_version IS NOT NULL
			GROUP BY parent_module_version
		) analytics_counts ON analytics_counts.parent_module_version = provider.latest_version_id`

	// WHERE clause starts with filter for providers with latest versions
	whereConditions := []string{
		"provider.latest_version_id IS NOT NULL",
		"provider.latest_version_id > 0",
	}
	whereArgs := []interface{}{}

	// Query filter - handle multiple query terms (matching Python behavior)
	if len(queryTerms) > 0 {
		// Build OR conditions for each query term
		termConditions := []string{}
		for _, term := range queryTerms {
			termConditions = append(termConditions, "(provider.name LIKE ? OR namespace.namespace LIKE ? OR provider.description LIKE ?)")
			termLower := strings.ToLower(term)
			whereArgs = append(whereArgs, "%"+termLower+"%", "%"+termLower+"%", "%"+termLower+"%")
		}
		// Combine all term conditions with OR - match if ANY term matches
		whereConditions = append(whereConditions, "("+strings.Join(termConditions, " OR ")+")")
	}

	// Namespace filters
	if len(params.Namespaces) > 0 {
		placeholders := make([]string, len(params.Namespaces))
		for i := range params.Namespaces {
			placeholders[i] = "?"
		}
		whereConditions = append(whereConditions, "namespace.namespace IN ("+strings.Join(placeholders, ",")+")")
		for _, ns := range params.Namespaces {
			whereArgs = append(whereArgs, ns)
		}
	}

	// Category filters
	if len(params.Categories) > 0 {
		sql += ` LEFT JOIN provider_category ON provider_category.id = provider.provider_category_id`
		placeholders := make([]string, len(params.Categories))
		for i := range params.Categories {
			placeholders[i] = "?"
		}
		whereConditions = append(whereConditions, "provider_category.slug IN ("+strings.Join(placeholders, ",")+")")
		for _, cat := range params.Categories {
			whereArgs = append(whereArgs, cat)
		}
	} else {
		// Still join for category if not filtering
		sql += ` LEFT JOIN provider_category ON provider_category.id = provider.provider_category_id`
	}

	// Trusted/Contributed namespace filter
	if (params.TrustedNamespaces != nil && *params.TrustedNamespaces) || (params.Contributed != nil && *params.Contributed) {
		// Get trusted namespaces from config - for now, use empty list as placeholder
		// This should be injected from domain config
		trustedNamespaces := []string{} // TODO: Get from domain config

		if params.TrustedNamespaces != nil && *params.TrustedNamespaces && len(trustedNamespaces) > 0 {
			placeholders := make([]string, len(trustedNamespaces))
			for i := range trustedNamespaces {
				placeholders[i] = "?"
			}
			whereConditions = append(whereConditions, "namespace.namespace IN ("+strings.Join(placeholders, ",")+")")
			for _, ns := range trustedNamespaces {
				whereArgs = append(whereArgs, ns)
			}
		}
		if params.Contributed != nil && *params.Contributed && len(trustedNamespaces) > 0 {
			placeholders := make([]string, len(trustedNamespaces))
			for i := range trustedNamespaces {
				placeholders[i] = "?"
			}
			whereConditions = append(whereConditions, "namespace.namespace NOT IN ("+strings.Join(placeholders, ",")+")")
			for _, ns := range trustedNamespaces {
				whereArgs = append(whereArgs, ns)
			}
		}
	}

	// Combine WHERE conditions
	sql += " WHERE " + strings.Join(whereConditions, " AND ")

	// GROUP BY and ORDER BY
	sql += ` GROUP BY provider.id, namespace.id`
	if len(queryTerms) > 0 {
		sql += ` ORDER BY relevance_score DESC, provider.name DESC`
	} else {
		sql += ` ORDER BY provider.name DESC`
	}

	// Count total query
	countSQL := "SELECT COUNT(DISTINCT provider.id) as total FROM provider JOIN namespace ON namespace.id = provider.namespace_id LEFT JOIN provider_category ON provider_category.id = provider.provider_category_id WHERE " + strings.Join(whereConditions, " AND ")

	// Prepare arguments for main query - scoring args first (SELECT clause), then WHERE args
	finalArgs := append([]interface{}{}, args...)
	finalArgs = append(finalArgs, whereArgs...)

	// Apply pagination to main query only
	if params.Limit > 0 {
		sql += " LIMIT ?"
		finalArgs = append(finalArgs, params.Limit)
	}
	if params.Offset > 0 {
		sql += " OFFSET ?"
		finalArgs = append(finalArgs, params.Offset)
	}

	// Execute count query
	var totalCount struct {
		Total int64 `json:"total"`
	}
	// Count query only uses WHERE args, not the scoring args (since COUNT query doesn't have scoring SELECT clause)
	countArgs := append([]interface{}{}, whereArgs...)
	if err := r.db.WithContext(ctx).Raw(countSQL, countArgs...).Scan(&totalCount).Error; err != nil {
		return nil, fmt.Errorf("failed to count search results: %w", err)
	}

	// Execute main query
	type SearchResult struct {
		ProviderID                 int      `gorm:"column:provider_id"`
		NamespaceID                int      `gorm:"column:provider_namespace_id"`
		Name                       string   `gorm:"column:provider_name"`
		Description                *string  `gorm:"column:provider_description"`
		Tier                       string   `gorm:"column:provider_tier"`
		DefaultProviderSourceAuth *bool    `gorm:"column:provider_default_provider_source_auth"`
		ProviderCategoryID         *int     `gorm:"column:provider_provider_category_id"`
		RepositoryID               *int     `gorm:"column:provider_repository_id"`
		LatestVersionID            *int     `gorm:"column:provider_latest_version_id"`
		RelevanceScore             *int     `gorm:"column:relevance_score"`
		NamespaceName              string   `gorm:"column:namespace_namespace"`
		NamespaceDisplayName       *string  `gorm:"column:namespace_display_name"`
		NamespaceType              string   `gorm:"column:namespace_type"`
		VersionID                  *int     `gorm:"column:version_id"`
		Version                    *string  `gorm:"column:version_version"`
		GitTag                     *string  `gorm:"column:version_git_tag"`
		PublishedAt                *string  `gorm:"column:version_published_at"`
		RepositoryOwner            *string  `gorm:"column:repository_owner"`
		RepositoryDescription      *string  `gorm:"column:repository_description"`
		RepositoryCloneURL         *string  `gorm:"column:repository_clone_url"`
		RepositoryLogoURL          *string  `gorm:"column:repository_logo_url"`
		DownloadCount              int64    `gorm:"column:download_count"`
	}

	var results []SearchResult
	if err := r.db.WithContext(ctx).Raw(sql, finalArgs...).Scan(&results).Error; err != nil {
		return nil, fmt.Errorf("failed to search providers: %w", err)
	}

	// Convert to domain entities
	providers := make([]*provider.Provider, len(results))
	namespaceNames := make(map[int]string, len(results))
	versionData := make(map[int]repository.VersionData, len(results))
	for i, result := range results {
		// Store namespace name for this provider
		namespaceNames[result.ProviderID] = result.NamespaceName

		// Store version data for this provider (if available)
		if result.VersionID != nil {
			versionStr := ""
			if result.Version != nil {
				versionStr = *result.Version
			}
			versionData[result.ProviderID] = repository.VersionData{
				VersionID:             *result.VersionID,
				Version:               versionStr,
				GitTag:                result.GitTag,
				PublishedAt:           result.PublishedAt,
				RepositoryOwner:       result.RepositoryOwner,
				RepositoryDescription: result.RepositoryDescription,
				RepositoryCloneURL:    result.RepositoryCloneURL,
				RepositoryLogoURL:     result.RepositoryLogoURL,
				Downloads:             result.DownloadCount,
			}
		}

		// Handle nullable DefaultProviderSourceAuth
		defaultProviderSourceAuth := false
		if result.DefaultProviderSourceAuth != nil {
			defaultProviderSourceAuth = *result.DefaultProviderSourceAuth
		}

		// Create provider (using Reconstruct to match existing pattern)
		prov := provider.ReconstructProvider(
			result.ProviderID,
			result.NamespaceID,
			result.Name,
			result.Description,
			result.Tier,
			result.ProviderCategoryID,
			result.RepositoryID,
			result.LatestVersionID,
			defaultProviderSourceAuth,
		)
		providers[i] = prov

		// Set relevance score if available
		if len(queryTerms) > 0 && result.RelevanceScore != nil {
			// TODO: Add SetRelevanceScore method to domain model if needed
		}
	}

	return &repository.ProviderSearchResult{
		Providers:      providers,
		TotalCount:     int(totalCount.Total),
		NamespaceNames: namespaceNames,
		VersionData:    versionData,
	}, nil
}

// GetSearchFilters gets available filters and counts for a search query
func (r *ProviderRepository) GetSearchFilters(ctx context.Context, query string, trustedNamespacesList []string) (*repository.ProviderSearchFilters, error) {
	// Build base search query (without scoring) for filtering
	queryTerms := strings.Fields(query)

	var sql string
	var args []interface{}

	sql = `
		SELECT
			provider.id as provider_id,
			namespace.id as namespace_id,
			namespace.namespace as namespace_namespace,
			provider_category.slug as provider_category_slug
		FROM provider
		JOIN namespace ON namespace.id = provider.namespace_id
		LEFT JOIN provider_category ON provider_category.id = provider.provider_category_id
		WHERE provider.latest_version_id IS NOT NULL AND provider.latest_version_id > 0`

	if len(queryTerms) > 0 {
		// Build OR conditions for each query term
		termConditions := []string{}
		for _, term := range queryTerms {
			termConditions = append(termConditions, "(provider.name LIKE ? OR namespace.namespace LIKE ? OR provider.description LIKE ?)")
			termLower := strings.ToLower(term)
			args = append(args, "%"+termLower+"%", "%"+termLower+"%", "%"+termLower+"%")
		}
		// Combine all term conditions with OR
		sql += " AND (" + strings.Join(termConditions, " OR ") + ")"
	}

	// Execute query to get all matching providers
	type FilterResult struct {
		ProviderID           int     `gorm:"column:provider_id"`
		NamespaceID          int     `gorm:"column:namespace_id"`
		NamespaceName        string  `gorm:"column:namespace_namespace"`
		ProviderCategorySlug *string `gorm:"column:provider_category_slug"`
	}

	var filterResults []FilterResult
	if err := r.db.WithContext(ctx).Raw(sql, args...).Scan(&filterResults).Error; err != nil {
		return nil, fmt.Errorf("failed to get search filters: %w", err)
	}

	// Convert trusted namespaces list to map for O(1) lookup
	trustedNamespaces := make(map[string]bool)
	for _, ns := range trustedNamespacesList {
		trustedNamespaces[ns] = true
	}

	// Count filters
	trustedCount := 0
	contributedCount := 0
	categoryCounts := make(map[string]int)
	namespaceCounts := make(map[string]int)

	for _, result := range filterResults {
		// Count trusted vs contributed
		if trustedNamespaces[result.NamespaceName] {
			trustedCount++
		} else {
			contributedCount++
		}

		// Count categories
		if result.ProviderCategorySlug != nil {
			categoryCounts[*result.ProviderCategorySlug]++
		}

		// Count namespaces
		namespaceCounts[result.NamespaceName]++
	}

	return &repository.ProviderSearchFilters{
		TrustedNamespaces:  trustedCount,
		Contributed:        contributedCount,
		ProviderCategories: categoryCounts,
		Namespaces:         namespaceCounts,
	}, nil
}

// FindByNamespaceAndName retrieves a provider by namespace and name
func (r *ProviderRepository) FindByNamespaceAndName(ctx context.Context, namespace, providerName string) (*provider.Provider, error) {
	var model sqldb.ProviderDB

	// Build query - use correct table name "provider"
	db := r.db.WithContext(ctx).
		Preload("Namespace").
		Preload("ProviderCategory").
		Preload("Repository").
		Preload("LatestVersion")
	if namespace != "" {
		// Join with namespace to filter by namespace name
		// Note: namespace table uses 'namespace' as column name, not 'name'
		db = db.Joins("JOIN namespace ON namespace.id = provider.namespace_id").
			Where("namespace.namespace = ? AND provider.name = ?", namespace, providerName)
	} else {
		db = db.Where("provider.name = ?", providerName)
	}

	if err := db.First(&model).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find provider: %w", err)
	}

	// Convert to domain entity
	return toDomainProvider(&model), nil
}

// FindByID retrieves a provider by its ID
func (r *ProviderRepository) FindByID(ctx context.Context, providerID int) (*provider.Provider, error) {
	var model sqldb.ProviderDB

	if err := r.db.WithContext(ctx).
		Preload("Namespace").
		Preload("ProviderCategory").
		Preload("Repository").
		Preload("LatestVersion").
		First(&model, providerID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find provider: %w", err)
	}

	// Convert to domain entity
	return toDomainProvider(&model), nil
}

// FindVersionsByProvider retrieves all versions for a provider
func (r *ProviderRepository) FindVersionsByProvider(ctx context.Context, providerID int) ([]*provider.ProviderVersion, error) {
	var models []sqldb.ProviderVersionDB

	if err := r.db.WithContext(ctx).
		Preload("GPGKey").
		Where("provider_id = ?", providerID).
		Order("id DESC").
		Find(&models).Error; err != nil {
		return nil, fmt.Errorf("failed to find provider versions: %w", err)
	}

	// Convert to domain entities
	versions := make([]*provider.ProviderVersion, len(models))
	for i, model := range models {
		version := toDomainProviderVersion(&model)
		versions[i] = version
	}

	return versions, nil
}

// FindVersionByProviderAndVersion retrieves a specific version
func (r *ProviderRepository) FindVersionByProviderAndVersion(ctx context.Context, providerID int, version string) (*provider.ProviderVersion, error) {
	var model sqldb.ProviderVersionDB

	if err := r.db.WithContext(ctx).
		Preload("GPGKey").
		Where("provider_id = ? AND version = ?", providerID, version).
		First(&model).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find provider version: %w", err)
	}

	// Convert to domain entity
	return toDomainProviderVersion(&model), nil
}

// FindVersionByID retrieves a version by its ID
func (r *ProviderRepository) FindVersionByID(ctx context.Context, versionID int) (*provider.ProviderVersion, error) {
	var model sqldb.ProviderVersionDB

	if err := r.db.WithContext(ctx).
		Preload("GPGKey").
		First(&model, versionID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find provider version: %w", err)
	}

	// Convert to domain entity
	return toDomainProviderVersion(&model), nil
}

// FindBinariesByVersion retrieves all binaries for a provider version
func (r *ProviderRepository) FindBinariesByVersion(ctx context.Context, versionID int) ([]*provider.ProviderBinary, error) {
	var models []sqldb.ProviderVersionBinaryDB

	if err := r.db.WithContext(ctx).
		Where("provider_version_id = ?", versionID).
		Find(&models).Error; err != nil {
		return nil, fmt.Errorf("failed to find provider binaries: %w", err)
	}

	// Convert to domain entities
	binaries := make([]*provider.ProviderBinary, len(models))
	for i, model := range models {
		binary := toDomainProviderBinary(&model)
		binaries[i] = binary
	}

	return binaries, nil
}

// FindBinaryByPlatform retrieves a specific binary for a platform
func (r *ProviderRepository) FindBinaryByPlatform(ctx context.Context, versionID int, os, arch string) (*provider.ProviderBinary, error) {
	var model sqldb.ProviderVersionBinaryDB

	if err := r.db.WithContext(ctx).
		Where("provider_version_id = ? AND operating_system = ? AND architecture = ?", versionID, os, arch).
		First(&model).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find provider binary: %w", err)
	}

	// Convert to domain entity
	return toDomainProviderBinary(&model), nil
}

// FindGPGKeysByProvider retrieves all GPG keys for a provider
// NOTE: GPG keys belong to namespaces, not providers. This method is kept for compatibility
// but should be deprecated. Use namespace-based GPG key lookup instead.
func (r *ProviderRepository) FindGPGKeysByProvider(ctx context.Context, providerID int) ([]*provider.GPGKey, error) {
	// First get the provider to find its namespace
	var providerModel sqldb.ProviderDB
	if err := r.db.WithContext(ctx).First(&providerModel, providerID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, shared.ErrNotFound
		}
		return nil, fmt.Errorf("failed to find provider: %w", err)
	}

	// Now get GPG keys for the namespace
	var models []sqldb.GPGKeyDB
	if err := r.db.WithContext(ctx).
		Where("namespace_id = ?", providerModel.NamespaceID).
		Find(&models).Error; err != nil {
		return nil, fmt.Errorf("failed to find GPG keys: %w", err)
	}

	// Convert to domain entities
	keys := make([]*provider.GPGKey, len(models))
	for i, model := range models {
		key := toDomainGPGKey(&model)
		keys[i] = key
	}

	return keys, nil
}

// FindGPGKeyByKeyID retrieves a GPG key by its key identifier
func (r *ProviderRepository) FindGPGKeyByKeyID(ctx context.Context, keyID string) (*provider.GPGKey, error) {
	var model sqldb.GPGKeyDB

	if err := r.db.WithContext(ctx).
		Where("key_id = ?", keyID).
		First(&model).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find GPG key: %w", err)
	}

	// Convert to domain entity
	return toDomainGPGKey(&model), nil
}

// Save persists a provider aggregate to the database
func (r *ProviderRepository) Save(ctx context.Context, prov *provider.Provider) error {
	model := toDBProvider(prov)

	// Handle create or update
	if prov.ID() == 0 {
		// Create new provider
		if err := r.db.WithContext(ctx).Create(model).Error; err != nil {
			return fmt.Errorf("failed to create provider: %w", err)
		}
		prov.SetID(model.ID)
	} else {
		// Update existing provider
		if err := r.db.WithContext(ctx).Save(model).Error; err != nil {
			return fmt.Errorf("failed to update provider: %w", err)
		}
	}

	return nil
}

// SaveVersion persists a provider version
func (r *ProviderRepository) SaveVersion(ctx context.Context, version *provider.ProviderVersion) error {
	model := toDBProviderVersion(version)

	// Handle create or update
	if version.ID() == 0 {
		// Create new version
		if err := r.db.WithContext(ctx).Create(model).Error; err != nil {
			return fmt.Errorf("failed to create provider version: %w", err)
		}
		version.SetID(model.ID)
	} else {
		// Update existing version
		if err := r.db.WithContext(ctx).Save(model).Error; err != nil {
			return fmt.Errorf("failed to update provider version: %w", err)
		}
	}

	return nil
}

// SaveBinary persists a provider binary
func (r *ProviderRepository) SaveBinary(ctx context.Context, binary *provider.ProviderBinary) error {
	model := toDBProviderBinary(binary)

	// Handle create or update
	if binary.ID() == 0 {
		// Create new binary
		if err := r.db.WithContext(ctx).Create(model).Error; err != nil {
			return fmt.Errorf("failed to create provider binary: %w", err)
		}
		binary.SetID(model.ID)
	} else {
		// Update existing binary
		if err := r.db.WithContext(ctx).Save(model).Error; err != nil {
			return fmt.Errorf("failed to update provider binary: %w", err)
		}
	}

	return nil
}

// SaveGPGKey persists a GPG key
func (r *ProviderRepository) SaveGPGKey(ctx context.Context, gpgKey *provider.GPGKey) error {
	model := toDBGPGKey(gpgKey)

	// Handle create or update
	if gpgKey.ID() == 0 {
		// Create new GPG key
		if err := r.db.WithContext(ctx).Create(model).Error; err != nil {
			return fmt.Errorf("failed to create GPG key: %w", err)
		}
		gpgKey.SetID(model.ID)
	} else {
		// Update existing GPG key
		if err := r.db.WithContext(ctx).Save(model).Error; err != nil {
			return fmt.Errorf("failed to update GPG key: %w", err)
		}
	}

	return nil
}

// DeleteVersion removes a provider version
func (r *ProviderRepository) DeleteVersion(ctx context.Context, versionID int) error {
	if err := r.db.WithContext(ctx).Delete(&sqldb.ProviderVersionDB{}, versionID).Error; err != nil {
		return fmt.Errorf("failed to delete provider version: %w", err)
	}
	return nil
}

// DeleteBinary removes a provider binary
func (r *ProviderRepository) DeleteBinary(ctx context.Context, binaryID int) error {
	if err := r.db.WithContext(ctx).Delete(&sqldb.ProviderVersionBinaryDB{}, binaryID).Error; err != nil {
		return fmt.Errorf("failed to delete provider binary: %w", err)
	}
	return nil
}

// DeleteGPGKey removes a GPG key
func (r *ProviderRepository) DeleteGPGKey(ctx context.Context, keyID int) error {
	if err := r.db.WithContext(ctx).Delete(&sqldb.GPGKeyDB{}, keyID).Error; err != nil {
		return fmt.Errorf("failed to delete GPG key: %w", err)
	}
	return nil
}

// SetLatestVersion updates the latest version for a provider
func (r *ProviderRepository) SetLatestVersion(ctx context.Context, providerID, versionID int) error {
	if err := r.db.WithContext(ctx).
		Model(&sqldb.ProviderDB{}).
		Where("id = ?", providerID).
		Update("latest_version_id", versionID).Error; err != nil {
		return fmt.Errorf("failed to set latest version: %w", err)
	}
	return nil
}

// GetProviderVersionCount returns the number of versions for a provider
func (r *ProviderRepository) GetProviderVersionCount(ctx context.Context, providerID int) (int, error) {
	var count int64
	if err := r.db.WithContext(ctx).
		Model(&sqldb.ProviderVersionDB{}).
		Where("provider_id = ?", providerID).
		Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to count provider versions: %w", err)
	}
	return int(count), nil
}

// GetBinaryDownloadCount returns the download count for a binary
func (r *ProviderRepository) GetBinaryDownloadCount(ctx context.Context, binaryID int) (int64, error) {
	// Placeholder implementation - would need download tracking table
	return 0, nil
}

// FindDocumentationByID retrieves documentation by its ID
func (r *ProviderRepository) FindDocumentationByID(ctx context.Context, id int) (*provider.ProviderVersionDocumentation, error) {
	var dbDoc sqldb.ProviderVersionDocumentationDB

	if err := r.db.WithContext(ctx).
		Where("id = ?", id).
		First(&dbDoc).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return toDomainProviderDocumentation(&dbDoc), nil
}

// FindDocumentationByVersion retrieves all documentation for a provider version
func (r *ProviderRepository) FindDocumentationByVersion(ctx context.Context, versionID int) ([]*provider.ProviderVersionDocumentation, error) {
	var dbDocs []sqldb.ProviderVersionDocumentationDB

	if err := r.db.WithContext(ctx).
		Where("provider_version_id = ?", versionID).
		Order("id ASC").
		Find(&dbDocs).Error; err != nil {
		return nil, err
	}

	docs := make([]*provider.ProviderVersionDocumentation, len(dbDocs))
	for i := range dbDocs {
		docs[i] = toDomainProviderDocumentation(&dbDocs[i])
	}

	return docs, nil
}

// FindDocumentationByTypeSlugAndLanguage retrieves documentation by type, slug, and language
func (r *ProviderRepository) FindDocumentationByTypeSlugAndLanguage(ctx context.Context, versionID int, docType, slug, language string) (*provider.ProviderVersionDocumentation, error) {
	var dbDoc sqldb.ProviderVersionDocumentationDB

	if err := r.db.WithContext(ctx).
		Where("provider_version_id = ? AND documentation_type = ? AND slug = ? AND language = ?",
			versionID, docType, slug, language).
		First(&dbDoc).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return toDomainProviderDocumentation(&dbDoc), nil
}

// SearchDocumentation searches for documentation by category, slug, and language
func (r *ProviderRepository) SearchDocumentation(ctx context.Context, versionID int, category, slug, language string) ([]*provider.ProviderVersionDocumentation, error) {
	var dbDocs []sqldb.ProviderVersionDocumentationDB

	query := r.db.WithContext(ctx).Where("provider_version_id = ?", versionID)

	if category != "" {
		query = query.Where("documentation_type = ?", category)
	}
	if slug != "" {
		query = query.Where("slug = ?", slug)
	}
	if language != "" {
		query = query.Where("language = ?", language)
	}

	if err := query.Order("id ASC").Find(&dbDocs).Error; err != nil {
		return nil, err
	}

	docs := make([]*provider.ProviderVersionDocumentation, len(dbDocs))
	for i := range dbDocs {
		docs[i] = toDomainProviderDocumentation(&dbDocs[i])
	}

	return docs, nil
}

// SaveDocumentation persists provider documentation
func (r *ProviderRepository) SaveDocumentation(ctx context.Context, doc *provider.ProviderVersionDocumentation) error {
	dbDoc := toDBProviderDocumentation(doc)

	if doc.ID() == 0 {
		// Insert new documentation
		if err := r.db.WithContext(ctx).Create(dbDoc).Error; err != nil {
			return err
		}
		doc.SetID(dbDoc.ID)
	} else {
		// Update existing documentation
		if err := r.db.WithContext(ctx).Save(dbDoc).Error; err != nil {
			return err
		}
	}

	return nil
}

// Mapper functions (domain → DB)
// Following the same pattern as the module package

func toDBProvider(p *provider.Provider) *sqldb.ProviderDB {
	return &sqldb.ProviderDB{
		ID:                        p.ID(),
		NamespaceID:               p.NamespaceID(),
		Name:                      p.Name(),
		Description:               p.Description(),
		Tier:                      sqldb.ProviderTier(p.Tier()),
		DefaultProviderSourceAuth: p.UseProviderSourceAuth(),
		ProviderCategoryID:        p.CategoryID(),
		RepositoryID:              p.RepositoryID(),
		LatestVersionID:           p.LatestVersionID(),
	}
}

func toDBProviderVersion(v *provider.ProviderVersion) *sqldb.ProviderVersionDB {
	// Serialize protocol versions to JSON bytes
	protocolVersionsJSON, _ := json.Marshal(v.ProtocolVersions())

	return &sqldb.ProviderVersionDB{
		ID:                v.ID(),
		ProviderID:        v.ProviderID(),
		Version:           v.Version(),
		GitTag:            v.GitTag(),
		Beta:              v.Beta(),
		PublishedAt:       v.PublishedAt(),
		GPGKeyID:          v.GPGKeyID(), // Non-nullable in correct schema
		ExtractionVersion: nil,
		ProtocolVersions:  protocolVersionsJSON, // []byte, not string
	}
}

func toDBProviderBinary(b *provider.ProviderBinary) *sqldb.ProviderVersionBinaryDB {
	return &sqldb.ProviderVersionBinaryDB{
		ID:                b.ID(),
		ProviderVersionID: b.VersionID(),
		Name:              b.FileName(),
		OperatingSystem:   sqldb.ProviderBinaryOperatingSystemType(b.OperatingSystem()),
		Architecture:      sqldb.ProviderBinaryArchitectureType(b.Architecture()),
		Checksum:          b.FileHash(), // Domain uses FileHash, DB uses Checksum
	}
}

func toDBGPGKey(k *provider.GPGKey) *sqldb.GPGKeyDB {
	// Handle nullable field conversions
	var keyID *string
	if keyIDVal := k.KeyID(); keyIDVal != "" {
		keyID = &keyIDVal
	}

	// Convert time.Time to *time.Time
	var createdAt, updatedAt *time.Time
	if createdVal := k.CreatedAt(); !createdVal.IsZero() {
		createdAt = &createdVal
	}
	if updatedVal := k.UpdatedAt(); !updatedVal.IsZero() {
		updatedAt = &updatedVal
	}

	return &sqldb.GPGKeyDB{
		ID:          k.ID(),
		NamespaceID: 0, // Will be set by caller (GPG keys belong to namespace, not provider)
		ASCIIArmor:  []byte(k.AsciiArmor()),
		KeyID:       keyID,
		Fingerprint: nil, // Not in domain model
		Source:      nil, // Not in domain model
		SourceURL:   nil, // Not in domain model
		CreatedAt:   createdAt,
		UpdatedAt:   updatedAt,
	}
}

func toDBProviderDocumentation(d *provider.ProviderVersionDocumentation) *sqldb.ProviderVersionDocumentationDB {
	return &sqldb.ProviderVersionDocumentationDB{
		ID:                d.ID(),
		ProviderVersionID: d.ProviderVersionID(),
		Name:              d.Name(),
		Slug:              d.Slug(),
		Title:             d.Title(),
		Description:       d.Description(),
		Language:          d.Language(),
		Subcategory:       d.Subcategory(),
		Filename:          d.Filename(),
		DocumentationType: sqldb.ProviderDocumentationType(d.DocumentationType()),
		Content:           d.Content(),
	}
}

// Mapper functions (DB → domain)

func toDomainProvider(db *sqldb.ProviderDB) *provider.Provider {
	return provider.ReconstructProvider(
		db.ID,
		db.NamespaceID,
		db.Name,
		db.Description,
		string(db.Tier),
		db.ProviderCategoryID, // Field name: ProviderCategoryID
		db.RepositoryID,
		db.LatestVersionID,
		db.DefaultProviderSourceAuth, // Field name: DefaultProviderSourceAuth
	)
}

func toDomainProviderVersion(db *sqldb.ProviderVersionDB) *provider.ProviderVersion {
	// Parse protocol versions from JSON bytes
	var protocolVersions []string
	if len(db.ProtocolVersions) > 0 {
		if err := json.Unmarshal(db.ProtocolVersions, &protocolVersions); err != nil {
			// Fallback to default protocol
			protocolVersions = []string{"5.0"}
		}
	} else {
		protocolVersions = []string{"5.0"}
	}

	return provider.ReconstructProviderVersion(
		db.ID,
		db.ProviderID,
		db.Version,
		db.GitTag,
		db.Beta,
		db.PublishedAt,
		db.GPGKeyID, // Non-nullable in correct schema
		protocolVersions,
	)
}

func toDomainProviderBinary(db *sqldb.ProviderVersionBinaryDB) *provider.ProviderBinary {
	return provider.ReconstructProviderBinary(
		db.ID,
		db.ProviderVersionID,
		string(db.OperatingSystem),
		string(db.Architecture),
		db.Name,     // DB field: Name, domain: FileName
		0,           // FileSize - not stored in DB
		db.Checksum, // DB field: Checksum, domain: FileHash
		"",          // DownloadURL - not stored in DB
	)
}

func toDomainGPGKey(db *sqldb.GPGKeyDB) *provider.GPGKey {
	// Handle nullable to non-nullable conversions
	var keyID string
	if db.KeyID != nil {
		keyID = *db.KeyID
	}

	var createdAt, updatedAt time.Time
	if db.CreatedAt != nil {
		createdAt = *db.CreatedAt
	}
	if db.UpdatedAt != nil {
		updatedAt = *db.UpdatedAt
	}

	return provider.ReconstructGPGKey(
		db.ID,
		"", // KeyText - not in DB
		string(db.ASCIIArmor),
		keyID,
		nil, // TrustSignature - not in DB
		createdAt,
		updatedAt,
	)
}

func toDomainProviderDocumentation(db *sqldb.ProviderVersionDocumentationDB) *provider.ProviderVersionDocumentation {
	return provider.ReconstructProviderVersionDocumentation(
		db.ID,
		db.ProviderVersionID,
		db.Name,
		db.Slug,
		db.Title,
		db.Description,
		db.Language,
		db.Subcategory,
		db.Filename,
		string(db.DocumentationType),
		db.Content,
	)
}
