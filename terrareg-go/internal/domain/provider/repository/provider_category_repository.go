package repository

import (
	"context"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/provider/model"
)

// ProviderCategoryRepository defines the interface for provider category persistence
type ProviderCategoryRepository interface {
	// FindByID finds a category by ID
	FindByID(ctx context.Context, id int) (*model.ProviderCategory, error)

	// FindBySlug finds a category by slug
	FindBySlug(ctx context.Context, slug string) (*model.ProviderCategory, error)

	// FindAll finds all categories
	FindAll(ctx context.Context) ([]*model.ProviderCategory, error)

	// FindUserSelectable finds only user-selectable categories
	FindUserSelectable(ctx context.Context) ([]*model.ProviderCategory, error)

	// Save saves a category (create or update)
	Save(ctx context.Context, category *model.ProviderCategory) error

	// Delete deletes a category
	Delete(ctx context.Context, id int) error

	// Exists checks if a category exists
	Exists(ctx context.Context, id int) (bool, error)

	// ExistsBySlug checks if a category with the given slug exists
	ExistsBySlug(ctx context.Context, slug string) (bool, error)

	// Count returns the total number of categories
	Count(ctx context.Context) (int64, error)
}