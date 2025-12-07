package repository

import (
	"context"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/module/model"
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
	Query                  string
	Namespaces             []string // Change from *string to []string for multiple values
	Module                 *string
	Providers              []string // Change from *string to []string for multiple values
	Verified               *bool
	TrustedNamespaces      *bool   // New: Filter for trusted namespaces only
	Contributed            *bool   // New: Filter for contributed modules only
	TargetTerraformVersion *string // New: Check compatibility with specific Terraform version
	Limit                  int
	Offset                 int
	OrderBy                string
	OrderDir               string
}

// ModuleSearchResult represents search results
type ModuleSearchResult struct {
	Modules    []*model.ModuleProvider
	TotalCount int
}
