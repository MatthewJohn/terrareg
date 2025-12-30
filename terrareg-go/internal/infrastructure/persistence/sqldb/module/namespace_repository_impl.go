package module

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/module/model"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/module/repository"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/shared"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/persistence/sqldb"
)

// NamespaceRepositoryImpl implements NamespaceRepository using GORM
type NamespaceRepositoryImpl struct {
	db *gorm.DB
}

// NewNamespaceRepository creates a new namespace repository
func NewNamespaceRepository(db *gorm.DB) repository.NamespaceRepository {
	return &NamespaceRepositoryImpl{db: db}
}

// Save persists a namespace
func (r *NamespaceRepositoryImpl) Save(ctx context.Context, namespace *model.Namespace) error {
	dbModel := toDBNamespace(namespace)

	var err error
	if namespace.ID() == 0 {
		// Create
		err = r.db.WithContext(ctx).Create(&dbModel).Error
		if err != nil {
			return fmt.Errorf("failed to create namespace: %w", err)
		}

		// Update domain model with generated ID
		*namespace = *fromDBNamespace(&dbModel)
	} else {
		// Update
		err = r.db.WithContext(ctx).Save(&dbModel).Error
		if err != nil {
			return fmt.Errorf("failed to update namespace: %w", err)
		}
	}

	return nil
}

// FindByID retrieves a namespace by ID
func (r *NamespaceRepositoryImpl) FindByID(ctx context.Context, id int) (*model.Namespace, error) {
	var dbModel sqldb.NamespaceDB

	err := r.db.WithContext(ctx).First(&dbModel, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, shared.ErrNotFound
		}
		return nil, fmt.Errorf("failed to find namespace: %w", err)
	}

	return fromDBNamespace(&dbModel), nil
}

// FindByName retrieves a namespace by name
func (r *NamespaceRepositoryImpl) FindByName(ctx context.Context, name string) (*model.Namespace, error) {
	var dbModel sqldb.NamespaceDB

	err := r.db.WithContext(ctx).Where("namespace = ?", name).First(&dbModel).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, shared.ErrNotFound
		}
		return nil, fmt.Errorf("failed to find namespace: %w", err)
	}

	return fromDBNamespace(&dbModel), nil
}

// List retrieves all namespaces
func (r *NamespaceRepositoryImpl) List(ctx context.Context) ([]*model.Namespace, error) {
	var dbModels []sqldb.NamespaceDB

	err := r.db.WithContext(ctx).Order("namespace ASC").Find(&dbModels).Error
	if err != nil {
		return nil, fmt.Errorf("failed to list namespaces: %w", err)
	}

	namespaces := make([]*model.Namespace, len(dbModels))
	for i, dbModel := range dbModels {
		namespaces[i] = fromDBNamespace(&dbModel)
	}

	return namespaces, nil
}

// Delete removes a namespace
func (r *NamespaceRepositoryImpl) Delete(ctx context.Context, id int) error {
	result := r.db.WithContext(ctx).Delete(&sqldb.NamespaceDB{}, id)
	if result.Error != nil {
		return fmt.Errorf("failed to delete namespace: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return shared.ErrNotFound
	}

	return nil
}

// Exists checks if a namespace exists
func (r *NamespaceRepositoryImpl) Exists(ctx context.Context, name string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&sqldb.NamespaceDB{}).Where("namespace = ?", name).Count(&count).Error
	if err != nil {
		return false, fmt.Errorf("failed to check namespace existence: %w", err)
	}

	return count > 0, nil
}
