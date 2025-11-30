package provider

import (
	"context"
	"encoding/json"
	"fmt"

	"gorm.io/gorm"

	"github.com/terrareg/terrareg/internal/domain/provider"
	"github.com/terrareg/terrareg/internal/infrastructure/persistence/sqldb"
)

// ProviderRepositoryImpl implements the provider repository using GORM
type ProviderRepositoryImpl struct {
	db *gorm.DB
}

// NewProviderRepository creates a new provider repository
func NewProviderRepository(db *gorm.DB) *ProviderRepositoryImpl {
	return &ProviderRepositoryImpl{db: db}
}

// FindAll retrieves all providers with pagination
func (r *ProviderRepositoryImpl) FindAll(ctx context.Context, offset, limit int) ([]*provider.Provider, int, error) {
	var providers []sqldb.ProviderDB
	var total int64

	// Get total count
	if err := r.db.WithContext(ctx).Model(&sqldb.ProviderDB{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated results
	if err := r.db.WithContext(ctx).
		Preload("Namespace").
		Offset(offset).
		Limit(limit).
		Find(&providers).Error; err != nil {
		return nil, 0, err
	}

	return r.toDomainList(providers), int(total), nil
}

// Search searches for providers by query with pagination
func (r *ProviderRepositoryImpl) Search(ctx context.Context, query string, offset, limit int) ([]*provider.Provider, int, error) {
	var providers []sqldb.ProviderDB
	var total int64

	searchPattern := "%" + query + "%"

	// Get total count
	if err := r.db.WithContext(ctx).
		Model(&sqldb.ProviderDB{}).
		Joins("JOIN namespace ON namespace.id = provider.namespace_id").
		Where("provider.name LIKE ? OR namespace.namespace LIKE ? OR provider.description LIKE ?",
			searchPattern, searchPattern, searchPattern).
		Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated results
	if err := r.db.WithContext(ctx).
		Preload("Namespace").
		Joins("JOIN namespace ON namespace.id = provider.namespace_id").
		Where("provider.name LIKE ? OR namespace.namespace LIKE ? OR provider.description LIKE ?",
			searchPattern, searchPattern, searchPattern).
		Offset(offset).
		Limit(limit).
		Find(&providers).Error; err != nil {
		return nil, 0, err
	}

	return r.toDomainList(providers), int(total), nil
}

// FindByNamespaceAndName retrieves a provider by namespace and name
func (r *ProviderRepositoryImpl) FindByNamespaceAndName(ctx context.Context, namespace, providerName string) (*provider.Provider, error) {
	var providerDB sqldb.ProviderDB

	err := r.db.WithContext(ctx).
		Preload("Namespace").
		Joins("JOIN namespace ON namespace.id = provider.namespace_id").
		Where("namespace.namespace = ? AND provider.name = ?", namespace, providerName).
		First(&providerDB).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("provider not found: %s/%s", namespace, providerName)
		}
		return nil, err
	}

	return r.toDomain(&providerDB), nil
}

// FindVersionsByProvider retrieves all versions for a provider
func (r *ProviderRepositoryImpl) FindVersionsByProvider(ctx context.Context, providerID int) ([]*provider.ProviderVersion, error) {
	var versions []sqldb.ProviderVersionDB

	if err := r.db.WithContext(ctx).
		Where("provider_id = ?", providerID).
		Order("published_at DESC").
		Find(&versions).Error; err != nil {
		return nil, err
	}

	return r.toVersionDomainList(versions), nil
}

// FindVersionByProviderAndVersion retrieves a specific version
func (r *ProviderRepositoryImpl) FindVersionByProviderAndVersion(ctx context.Context, providerID int, version string) (*provider.ProviderVersion, error) {
	var versionDB sqldb.ProviderVersionDB

	err := r.db.WithContext(ctx).
		Where("provider_id = ? AND version = ?", providerID, version).
		First(&versionDB).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("provider version not found: %d/%s", providerID, version)
		}
		return nil, err
	}

	return r.toVersionDomain(&versionDB), nil
}

// toDomain converts database model to domain model
func (r *ProviderRepositoryImpl) toDomain(db *sqldb.ProviderDB) *provider.Provider {
	return provider.NewProvider(
		db.ID,
		db.NamespaceID,
		db.Name,
		db.Description,
		string(db.Tier),
		db.ProviderCategoryID,
		db.RepositoryID,
		db.LatestVersionID,
		db.DefaultProviderSourceAuth,
	)
}

// toDomainList converts a list of database models to domain models
func (r *ProviderRepositoryImpl) toDomainList(dbs []sqldb.ProviderDB) []*provider.Provider {
	providers := make([]*provider.Provider, len(dbs))
	for i, db := range dbs {
		providers[i] = r.toDomain(&db)
	}
	return providers
}

// toVersionDomain converts database version model to domain model
func (r *ProviderRepositoryImpl) toVersionDomain(db *sqldb.ProviderVersionDB) *provider.ProviderVersion {
	// Parse protocol versions from JSON
	var protocolVersions []string
	if db.ProtocolVersions != nil {
		_ = json.Unmarshal(db.ProtocolVersions, &protocolVersions)
	}

	return provider.NewProviderVersion(
		db.ID,
		db.ProviderID,
		db.Version,
		db.GitTag,
		db.Beta,
		db.PublishedAt,
		db.GPGKeyID,
		protocolVersions,
	)
}

// toVersionDomainList converts a list of version database models to domain models
func (r *ProviderRepositoryImpl) toVersionDomainList(dbs []sqldb.ProviderVersionDB) []*provider.ProviderVersion {
	versions := make([]*provider.ProviderVersion, len(dbs))
	for i, db := range dbs {
		versions[i] = r.toVersionDomain(&db)
	}
	return versions
}
