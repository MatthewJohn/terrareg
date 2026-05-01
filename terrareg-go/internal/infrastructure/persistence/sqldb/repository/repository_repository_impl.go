package repository

import (
	"context"
	"database/sql"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/repository/model"
	repositoryRepo "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/repository/repository"
	"gorm.io/gorm"
)

// RepositoryRepositoryImpl implements RepositoryRepository using GORM
type RepositoryRepositoryImpl struct {
	db *gorm.DB
}

// NewRepositoryRepository creates a new RepositoryRepository
func NewRepositoryRepository(db *gorm.DB) repositoryRepo.RepositoryRepository {
	return &RepositoryRepositoryImpl{db: db}
}

// FindByProviderSourceAndProviderID retrieves a repository by provider source name and provider ID
// Python reference: repository_model.py::get_by_provider_source_and_provider_id
func (r *RepositoryRepositoryImpl) FindByProviderSourceAndProviderID(ctx context.Context, providerSourceName string, providerID string) (*model.Repository, error) {
	var dbRepository struct {
		ID                 int            `gorm:"column:id"`
		ProviderID         string         `gorm:"column:provider_id"`
		Name               string         `gorm:"column:name"`
		Owner              string         `gorm:"column:owner"`
		Description        sql.NullString `gorm:"column:description"`
		CloneURL           string         `gorm:"column:clone_url"`
		LogoURL            sql.NullString `gorm:"column:logo_url"`
		ProviderSourceName string         `gorm:"column:provider_source_name"`
	}

	err := r.db.WithContext(ctx).Table("repository").
		Where("provider_source_name = ? AND provider_id = ?", providerSourceName, providerID).
		First(&dbRepository).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}

	result := &model.Repository{
		ID:                 dbRepository.ID,
		ProviderID:         dbRepository.ProviderID,
		Name:               dbRepository.Name,
		Owner:              dbRepository.Owner,
		CloneURL:           dbRepository.CloneURL,
		ProviderSourceName: dbRepository.ProviderSourceName,
	}

	if dbRepository.Description.Valid {
		result.Description = &dbRepository.Description.String
	}

	if dbRepository.LogoURL.Valid {
		result.LogoURL = &dbRepository.LogoURL.String
	}

	return result, nil
}
