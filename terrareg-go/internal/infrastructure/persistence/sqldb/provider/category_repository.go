package provider

import (
	"context"
	"fmt"
	"time"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/provider/model"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/persistence/sqldb"
	"gorm.io/gorm"
)

// ProviderCategoryRepositoryImpl implements ProviderCategoryRepository using GORM
type ProviderCategoryRepositoryImpl struct {
	db *gorm.DB
}

// NewProviderCategoryRepository creates a new ProviderCategoryRepository
func NewProviderCategoryRepository(database *sqldb.Database) *ProviderCategoryRepositoryImpl {
	return &ProviderCategoryRepositoryImpl{
		db: database.DB,
	}
}

// FindByID finds a category by ID
func (r *ProviderCategoryRepositoryImpl) FindByID(ctx context.Context, id int) (*model.ProviderCategory, error) {
	var dbCategory sqldb.ProviderCategoryDB
	err := r.db.WithContext(ctx).First(&dbCategory, id).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("provider category with ID %d not found", id)
		}
		return nil, fmt.Errorf("failed to find provider category: %w", err)
	}

	return r.dbToDomain(&dbCategory), nil
}

// FindBySlug finds a category by slug
func (r *ProviderCategoryRepositoryImpl) FindBySlug(ctx context.Context, slug string) (*model.ProviderCategory, error) {
	var dbCategory sqldb.ProviderCategoryDB
	err := r.db.WithContext(ctx).Where("slug = ?", slug).First(&dbCategory).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("provider category with slug '%s' not found", slug)
		}
		return nil, fmt.Errorf("failed to find provider category: %w", err)
	}

	return r.dbToDomain(&dbCategory), nil
}

// FindAll finds all categories
func (r *ProviderCategoryRepositoryImpl) FindAll(ctx context.Context) ([]*model.ProviderCategory, error) {
	var dbCategories []sqldb.ProviderCategoryDB
	err := r.db.WithContext(ctx).Order("id ASC").Find(&dbCategories).Error
	if err != nil {
		return nil, fmt.Errorf("failed to find all provider categories: %w", err)
	}

	categories := make([]*model.ProviderCategory, len(dbCategories))
	for i, dbCat := range dbCategories {
		categories[i] = r.dbToDomain(&dbCat)
	}

	return categories, nil
}

// FindUserSelectable finds only user-selectable categories
func (r *ProviderCategoryRepositoryImpl) FindUserSelectable(ctx context.Context) ([]*model.ProviderCategory, error) {
	var dbCategories []sqldb.ProviderCategoryDB
	err := r.db.WithContext(ctx).Where("user_selectable = ?", true).Order("id ASC").Find(&dbCategories).Error
	if err != nil {
		return nil, fmt.Errorf("failed to find user-selectable provider categories: %w", err)
	}

	categories := make([]*model.ProviderCategory, len(dbCategories))
	for i, dbCat := range dbCategories {
		categories[i] = r.dbToDomain(&dbCat)
	}

	return categories, nil
}

// Save saves a category (create or update)
func (r *ProviderCategoryRepositoryImpl) Save(ctx context.Context, category *model.ProviderCategory) error {
	dbCategory := r.domainToDB(category)

	if category.ID() == 0 {
		// Create new category
		err := r.db.WithContext(ctx).Create(dbCategory).Error
		if err != nil {
			return fmt.Errorf("failed to create provider category: %w", err)
		}
		// Note: Since the model doesn't have SetID, the caller would need to
		// fetch the newly created category to get the ID
	} else {
		// Update existing category
		err := r.db.WithContext(ctx).Save(dbCategory).Error
		if err != nil {
			return fmt.Errorf("failed to update provider category: %w", err)
		}
	}

	return nil
}

// Delete deletes a category
func (r *ProviderCategoryRepositoryImpl) Delete(ctx context.Context, id int) error {
	err := r.db.WithContext(ctx).Delete(&sqldb.ProviderCategoryDB{}, id).Error
	if err != nil {
		return fmt.Errorf("failed to delete provider category: %w", err)
	}
	return nil
}

// Exists checks if a category exists
func (r *ProviderCategoryRepositoryImpl) Exists(ctx context.Context, id int) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&sqldb.ProviderCategoryDB{}).Where("id = ?", id).Count(&count).Error
	if err != nil {
		return false, fmt.Errorf("failed to check if provider category exists: %w", err)
	}
	return count > 0, nil
}

// ExistsBySlug checks if a category with the given slug exists
func (r *ProviderCategoryRepositoryImpl) ExistsBySlug(ctx context.Context, slug string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&sqldb.ProviderCategoryDB{}).Where("slug = ?", slug).Count(&count).Error
	if err != nil {
		return false, fmt.Errorf("failed to check if provider category slug exists: %w", err)
	}
	return count > 0, nil
}

// Count returns the total number of categories
func (r *ProviderCategoryRepositoryImpl) Count(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&sqldb.ProviderCategoryDB{}).Count(&count).Error
	if err != nil {
		return 0, fmt.Errorf("failed to count provider categories: %w", err)
	}
	return count, nil
}

// dbToDomain converts a DB model to a domain model
func (r *ProviderCategoryRepositoryImpl) dbToDomain(dbCat *sqldb.ProviderCategoryDB) *model.ProviderCategory {
	// Use zero time for created/updated if not needed
	return model.ReconstructProviderCategory(
		dbCat.ID,
		dbCat.Name,
		dbCat.Slug,
		dbCat.UserSelectable,
		time.Time{}, // createdAt - not stored in DB
		time.Time{}, // updatedAt - not stored in DB
	)
}

// domainToDB converts a domain model to a DB model
func (r *ProviderCategoryRepositoryImpl) domainToDB(category *model.ProviderCategory) *sqldb.ProviderCategoryDB {
	dbCategory := &sqldb.ProviderCategoryDB{
		ID:             category.ID(),
		Slug:           category.Slug(),
		UserSelectable: category.UserSelectable(),
	}

	// Name is optional pointer field
	if category.Name() != nil {
		dbCategory.Name = category.Name()
	}

	return dbCategory
}
