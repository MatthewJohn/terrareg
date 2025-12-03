package provider

import (
	"context"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/provider/model"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/provider/repository"
)

// ListProviderCategoriesQuery is a query to list provider categories
type ListProviderCategoriesQuery struct {
	categoryRepo repository.ProviderCategoryRepository
}

// NewListProviderCategoriesQuery creates a new ListProviderCategoriesQuery
func NewListProviderCategoriesQuery(categoryRepo repository.ProviderCategoryRepository) *ListProviderCategoriesQuery {
	return &ListProviderCategoriesQuery{
		categoryRepo: categoryRepo,
	}
}

// Execute executes the query to list provider categories
func (q *ListProviderCategoriesQuery) Execute(ctx context.Context) ([]*model.ProviderCategory, error) {
	return q.categoryRepo.FindAll(ctx)
}

// ListUserSelectableProviderCategoriesQuery is a query to list only user-selectable categories
type ListUserSelectableProviderCategoriesQuery struct {
	categoryRepo repository.ProviderCategoryRepository
}

// NewListUserSelectableProviderCategoriesQuery creates a new ListUserSelectableProviderCategoriesQuery
func NewListUserSelectableProviderCategoriesQuery(categoryRepo repository.ProviderCategoryRepository) *ListUserSelectableProviderCategoriesQuery {
	return &ListUserSelectableProviderCategoriesQuery{
		categoryRepo: categoryRepo,
	}
}

// Execute executes the query to list user-selectable provider categories
func (q *ListUserSelectableProviderCategoriesQuery) Execute(ctx context.Context) ([]*model.ProviderCategory, error) {
	return q.categoryRepo.FindUserSelectable(ctx)
}

// GetProviderCategoryQuery is a query to get a specific provider category
type GetProviderCategoryQuery struct {
	categoryRepo repository.ProviderCategoryRepository
}

// NewGetProviderCategoryQuery creates a new GetProviderCategoryQuery
func NewGetProviderCategoryQuery(categoryRepo repository.ProviderCategoryRepository) *GetProviderCategoryQuery {
	return &GetProviderCategoryQuery{
		categoryRepo: categoryRepo,
	}
}

// Execute executes the query to get a provider category by ID
func (q *GetProviderCategoryQuery) Execute(ctx context.Context, id int) (*model.ProviderCategory, error) {
	return q.categoryRepo.FindByID(ctx, id)
}

// ExecuteBySlug executes the query to get a provider category by slug
func (q *GetProviderCategoryQuery) ExecuteBySlug(ctx context.Context, slug string) (*model.ProviderCategory, error) {
	return q.categoryRepo.FindBySlug(ctx, slug)
}