package repository

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/repository/model"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/repository/repository"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/persistence/sqldb"
)

// RepositoryRepositoryImpl implements RepositoryRepository using GORM
type RepositoryRepositoryImpl struct {
	db *gorm.DB
}

// NewRepositoryRepository creates a new repository repository
func NewRepositoryRepository(db *gorm.DB) repository.RepositoryRepository {
	return &RepositoryRepositoryImpl{db: db}
}

// FindByID retrieves a repository by its database primary key
// Returns nil if not found (no error)
// Python reference: repository_model.py::Repository.get_by_pk()
func (r *RepositoryRepositoryImpl) FindByID(ctx context.Context, id int) (*model.Repository, error) {
	var dbModel sqldb.RepositoryDB
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&dbModel).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find repository by id: %w", err)
	}
	return r.dbModelToDomain(&dbModel)
}

// FindByProviderSourceAndProviderID retrieves a repository by provider source name and provider ID
// Returns nil if not found (no error)
// Python reference: repository_model.py::Repository.get_by_provider_source_and_provider_id()
func (r *RepositoryRepositoryImpl) FindByProviderSourceAndProviderID(ctx context.Context, providerSourceName string, providerID string) (*model.Repository, error) {
	var dbModel sqldb.RepositoryDB
	err := r.db.WithContext(ctx).
		Where("provider_source_name = ? AND provider_id = ?", providerSourceName, providerID).
		First(&dbModel).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find repository by provider source and provider id: %w", err)
	}
	return r.dbModelToDomain(&dbModel)
}

// FindByOwnerList retrieves repositories matching any of the given owners
// Returns empty slice if none found
// Python reference: repository_model.py::Repository.get_repositories_by_owner_list()
func (r *RepositoryRepositoryImpl) FindByOwnerList(ctx context.Context, owners []string) ([]*model.Repository, error) {
	if len(owners) == 0 {
		return []*model.Repository{}, nil
	}

	var dbModels []sqldb.RepositoryDB
	err := r.db.WithContext(ctx).
		Where("owner IN ?", owners).
		Find(&dbModels).Error
	if err != nil {
		return nil, fmt.Errorf("failed to find repositories by owner list: %w", err)
	}

	result := make([]*model.Repository, len(dbModels))
	for i, dbModel := range dbModels {
		domainModel, err := r.dbModelToDomain(&dbModel)
		if err != nil {
			return nil, fmt.Errorf("failed to convert db model to domain: %w", err)
		}
		result[i] = domainModel
	}
	return result, nil
}

// Create creates a new repository
// Returns the created repository with ID set, or error if creation fails
// Python reference: repository_model.py::Repository.create()
func (r *RepositoryRepositoryImpl) Create(ctx context.Context, repository *model.Repository) (*model.Repository, error) {
	dbModel := r.domainToDBModel(repository)

	err := r.db.WithContext(ctx).Create(dbModel).Error
	if err != nil {
		return nil, fmt.Errorf("failed to create repository: %w", err)
	}

	// Update the domain entity with the generated ID
	repository.SetID(int(dbModel.ID))
	return repository, nil
}

// Update updates an existing repository
// Python reference: repository_model.py::Repository.update_attributes()
func (r *RepositoryRepositoryImpl) Update(ctx context.Context, repository *model.Repository) error {
	dbModel := r.domainToDBModel(repository)

	err := r.db.WithContext(ctx).
		Model(&sqldb.RepositoryDB{}).
		Where("id = ?", repository.ID()).
		Updates(dbModel).Error
	if err != nil {
		return fmt.Errorf("failed to update repository: %w", err)
	}
	return nil
}

// Delete removes a repository by its ID
// Returns nil if repository doesn't exist (idempotent)
func (r *RepositoryRepositoryImpl) Delete(ctx context.Context, id int) error {
	err := r.db.WithContext(ctx).Where("id = ?", id).Delete(&sqldb.RepositoryDB{}).Error
	if err != nil {
		return fmt.Errorf("failed to delete repository: %w", err)
	}
	return nil
}

// Exists checks if a repository exists by provider source name and provider ID
// Python reference: used in Repository.create() before insertion
func (r *RepositoryRepositoryImpl) Exists(ctx context.Context, providerSourceName string, providerID string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&sqldb.RepositoryDB{}).
		Where("provider_source_name = ? AND provider_id = ?", providerSourceName, providerID).
		Count(&count).Error
	if err != nil {
		return false, fmt.Errorf("failed to check if repository exists: %w", err)
	}
	return count > 0, nil
}

// dbModelToDomain converts a DB model to a domain model
func (r *RepositoryRepositoryImpl) dbModelToDomain(dbModel *sqldb.RepositoryDB) (*model.Repository, error) {
	var description *string
	if len(dbModel.Description) > 0 {
		desc := string(dbModel.Description)
		description = &desc
	}

	return model.ReconstructRepository(
		int(dbModel.ID),
		dbModel.ProviderID,
		stringValue(dbModel.Owner),
		stringValue(dbModel.Name),
		description,
		dbModel.CloneURL,
		dbModel.LogoURL,
		dbModel.ProviderSourceName,
	), nil
}

// domainToDBModel converts a domain model to a DB model
func (r *RepositoryRepositoryImpl) domainToDBModel(repo *model.Repository) *sqldb.RepositoryDB {
	return repo.ToDBModel()
}

// stringValue safely converts a string pointer to a string
func stringValue(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
