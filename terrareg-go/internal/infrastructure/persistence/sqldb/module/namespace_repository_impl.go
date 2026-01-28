package module

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/application/query"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/shared/types"
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
func (r *NamespaceRepositoryImpl) FindByName(ctx context.Context, name types.NamespaceName) (*model.Namespace, error) {
	var dbModel sqldb.NamespaceDB

	err := r.db.WithContext(ctx).Where("namespace = ?", string(name)).First(&dbModel).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, shared.ErrNotFound
		}
		return nil, fmt.Errorf("failed to find namespace: %w", err)
	}

	return fromDBNamespace(&dbModel), nil
}

// List retrieves namespaces with optional pagination
// If opts is nil or opts.Limit is 0, returns all namespaces
// Returns: namespaces, total count (for pagination meta), error
func (r *NamespaceRepositoryImpl) List(ctx context.Context, opts *query.ListOptions) ([]*model.Namespace, int, error) {
	var dbModels []sqldb.NamespaceDB
	query := r.db.WithContext(ctx).Order("namespace ASC")

	// Get total count first (needed for pagination meta)
	var totalCount int64
	if err := query.Model(&sqldb.NamespaceDB{}).Count(&totalCount).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count namespaces: %w", err)
	}

	// Apply pagination if provided
	if opts != nil && opts.Limit > 0 {
		query = query.Offset(opts.Offset).Limit(opts.Limit)
	}

	err := query.Find(&dbModels).Error
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list namespaces: %w", err)
	}

	namespaces := make([]*model.Namespace, len(dbModels))
	for i, dbModel := range dbModels {
		namespaces[i] = fromDBNamespace(&dbModel)
	}

	return namespaces, int(totalCount), nil
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
func (r *NamespaceRepositoryImpl) Exists(ctx context.Context, name types.NamespaceName) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&sqldb.NamespaceDB{}).Where("namespace = ?", string(name)).Count(&count).Error
	if err != nil {
		return false, fmt.Errorf("failed to check namespace existence: %w", err)
	}

	return count > 0, nil
}
