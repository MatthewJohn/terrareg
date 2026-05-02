package git

import (
	"context"
	"fmt"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/git/model"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/git/repository"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/persistence/sqldb"
)

// gitProviderRepositoryImpl implements the GitProviderRepository interface
type gitProviderRepositoryImpl struct {
	db *sqldb.Database
}

// NewGitProviderRepository creates a new GitProvider repository
func NewGitProviderRepository(db *sqldb.Database) repository.GitProviderRepository {
	return &gitProviderRepositoryImpl{
		db: db,
	}
}

// FindByName retrieves a git provider by its name
// Python reference: models.py::GitProvider.get_by_name()
func (r *gitProviderRepositoryImpl) FindByName(ctx context.Context, name string) (*model.GitProvider, error) {
	var dbModel sqldb.GitProviderDB
	err := r.db.DB.WithContext(ctx).Where("name = ?", name).First(&dbModel).Error
	if err != nil {
		if err.Error() == "record not found" {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find git provider by name: %w", err)
	}
	return FromDBGitProvider(&dbModel), nil
}

// FindAll retrieves all git providers
// Python reference: models.py::GitProvider.get_all()
func (r *gitProviderRepositoryImpl) FindAll(ctx context.Context) ([]*model.GitProvider, error) {
	var dbModels []sqldb.GitProviderDB
	err := r.db.DB.WithContext(ctx).Order("name ASC").Find(&dbModels).Error
	if err != nil {
		return nil, fmt.Errorf("failed to find all git providers: %w", err)
	}

	result := make([]*model.GitProvider, len(dbModels))
	for i := range dbModels {
		result[i] = FromDBGitProvider(&dbModels[i])
	}
	return result, nil
}

// Upsert creates or updates a git provider
// Python reference: models.py::GitProvider.initialise_from_config() (upsert logic)
func (r *gitProviderRepositoryImpl) Upsert(ctx context.Context, provider *model.GitProvider) error {
	// Check if provider exists
	existing, err := r.FindByName(ctx, provider.Name)
	if err != nil {
		return fmt.Errorf("failed to check existing git provider: %w", err)
	}

	dbModel := ToDBGitProvider(provider)

	if existing != nil {
		// Update existing
		dbModel.ID = existing.ID
		err = r.db.DB.WithContext(ctx).Save(dbModel).Error
		if err != nil {
			return fmt.Errorf("failed to update git provider: %w", err)
		}
	} else {
		// Create new
		err = r.db.DB.WithContext(ctx).Create(dbModel).Error
		if err != nil {
			return fmt.Errorf("failed to create git provider: %w", err)
		}
	}

	return nil
}

// Delete removes a git provider by name
func (r *gitProviderRepositoryImpl) Delete(ctx context.Context, name string) error {
	err := r.db.DB.WithContext(ctx).Where("name = ?", name).Delete(&sqldb.GitProviderDB{}).Error
	if err != nil {
		return fmt.Errorf("failed to delete git provider: %w", err)
	}
	return nil
}

// Exists checks if a git provider with the given name exists
func (r *gitProviderRepositoryImpl) Exists(ctx context.Context, name string) (bool, error) {
	var count int64
	err := r.db.DB.WithContext(ctx).Model(&sqldb.GitProviderDB{}).Where("name = ?", name).Count(&count).Error
	if err != nil {
		return false, fmt.Errorf("failed to check git provider existence: %w", err)
	}
	return count > 0, nil
}
