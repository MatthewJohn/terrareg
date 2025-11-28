package repository

import (
	"context"

	"github.com/terrareg/terrareg/internal/domain/module/model"
)

// ModuleProviderRepository defines the interface for module provider persistence
type ModuleProviderRepository interface {
	// Save persists a module provider (aggregate root)
	Save(ctx context.Context, mp *model.ModuleProvider) error

	// FindByID retrieves a module provider by ID
	FindByID(ctx context.Context, id int) (*model.ModuleProvider, error)

	// FindByNamespaceModuleProvider retrieves a module provider by namespace/module/provider
	FindByNamespaceModuleProvider(ctx context.Context, namespace, module, provider string) (*model.ModuleProvider, error)

	// FindByNamespace retrieves all module providers in a namespace
	FindByNamespace(ctx context.Context, namespace string) ([]*model.ModuleProvider, error)

	// Search searches for module providers
	Search(ctx context.Context, query ModuleSearchQuery) (*ModuleSearchResult, error)

	// Delete removes a module provider
	Delete(ctx context.Context, id int) error

	// Exists checks if a module provider exists
	Exists(ctx context.Context, namespace, module, provider string) (bool, error)
}

// ModuleSearchQuery represents search criteria
type ModuleSearchQuery struct {
	Query      string
	Namespace  *string
	Provider   *string
	Verified   *bool
	Limit      int
	Offset     int
	OrderBy    string
	OrderDir   string
}

// ModuleSearchResult represents search results
type ModuleSearchResult struct {
	Modules    []*model.ModuleProvider
	TotalCount int
}
