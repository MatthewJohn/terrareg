package provider_source

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/provider_source/model"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/provider_source/repository"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/persistence/sqldb"
)

// ProviderSourceRepositoryImpl implements ProviderSourceRepository using GORM
type ProviderSourceRepositoryImpl struct {
	db *gorm.DB
}

// NewProviderSourceRepository creates a new provider source repository
func NewProviderSourceRepository(db *gorm.DB) repository.ProviderSourceRepository {
	return &ProviderSourceRepositoryImpl{db: db}
}

// FindByName retrieves a provider source by its display name
// Returns nil if not found (no error)
func (r *ProviderSourceRepositoryImpl) FindByName(ctx context.Context, name string) (*model.ProviderSource, error) {
	var dbModel sqldb.ProviderSourceDB
	err := r.db.WithContext(ctx).Where("name = ?", name).First(&dbModel).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find provider source by name: %w", err)
	}
	return r.dbModelToDomain(&dbModel)
}

// FindByApiName retrieves a provider source by its API-friendly name
// Returns nil if not found (no error)
func (r *ProviderSourceRepositoryImpl) FindByApiName(ctx context.Context, apiName string) (*model.ProviderSource, error) {
	var dbModel sqldb.ProviderSourceDB
	err := r.db.WithContext(ctx).Where("api_name = ?", apiName).First(&dbModel).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find provider source by api_name: %w", err)
	}
	return r.dbModelToDomain(&dbModel)
}

// FindAll retrieves all provider sources from the database
// Returns empty slice if none exist
func (r *ProviderSourceRepositoryImpl) FindAll(ctx context.Context) ([]*model.ProviderSource, error) {
	var dbModels []sqldb.ProviderSourceDB
	err := r.db.WithContext(ctx).Find(&dbModels).Error
	if err != nil {
		return nil, fmt.Errorf("failed to find all provider sources: %w", err)
	}

	result := make([]*model.ProviderSource, len(dbModels))
	for i, dbModel := range dbModels {
		domainModel, err := r.dbModelToDomain(&dbModel)
		if err != nil {
			return nil, fmt.Errorf("failed to convert db model to domain: %w", err)
		}
		result[i] = domainModel
	}
	return result, nil
}

// Upsert creates a new provider source or updates an existing one
// Python reference: factory.py::initialise_from_config() - uses insert or update
func (r *ProviderSourceRepositoryImpl) Upsert(ctx context.Context, source *model.ProviderSource) error {
	dbModel := r.domainToDBModel(source)

	// Use GORM's Save which performs upsert based on primary key
	err := r.db.WithContext(ctx).Save(dbModel).Error
	if err != nil {
		return fmt.Errorf("failed to upsert provider source: %w", err)
	}
	return nil
}

// Delete removes a provider source by name
// Returns nil if source doesn't exist (idempotent)
func (r *ProviderSourceRepositoryImpl) Delete(ctx context.Context, name string) error {
	err := r.db.WithContext(ctx).Where("name = ?", name).Delete(&sqldb.ProviderSourceDB{}).Error
	if err != nil {
		return fmt.Errorf("failed to delete provider source: %w", err)
	}
	return nil
}

// Exists checks if a provider source with the given name exists
func (r *ProviderSourceRepositoryImpl) Exists(ctx context.Context, name string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&sqldb.ProviderSourceDB{}).Where("name = ?", name).Count(&count).Error
	if err != nil {
		return false, fmt.Errorf("failed to check if provider source exists: %w", err)
	}
	return count > 0, nil
}

// ExistsByApiName checks if a provider source with the given API name exists
func (r *ProviderSourceRepositoryImpl) ExistsByApiName(ctx context.Context, apiName string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&sqldb.ProviderSourceDB{}).Where("api_name = ?", apiName).Count(&count).Error
	if err != nil {
		return false, fmt.Errorf("failed to check if provider source exists by api_name: %w", err)
	}
	return count > 0, nil
}

// dbModelToDomain converts a DB model to a domain model
func (r *ProviderSourceRepositoryImpl) dbModelToDomain(dbModel *sqldb.ProviderSourceDB) (*model.ProviderSource, error) {
	// Decode config blob
	config := &model.ProviderSourceConfig{}
	if len(dbModel.Config) > 0 {
		if err := sqldb.DecodeBlob(dbModel.Config, config); err != nil {
			return nil, fmt.Errorf("failed to decode config: %w", err)
		}
	}

	apiName := ""
	if dbModel.APIName != nil {
		apiName = *dbModel.APIName
	}

	return model.NewProviderSource(
		dbModel.Name,
		apiName,
		model.ProviderSourceType(dbModel.ProviderSourceType),
		config,
	), nil
}

// domainToDBModel converts a domain model to a DB model
func (r *ProviderSourceRepositoryImpl) domainToDBModel(source *model.ProviderSource) *sqldb.ProviderSourceDB {
	return source.ToDBModel()
}
