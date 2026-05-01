package repository

import (
	"context"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/module/model"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/shared/types"
)

// ModuleProviderRepository defines the interface for module provider persistence
type ModuleProviderRepository interface {
	// Save persists a module provider (aggregate root)
	Save(ctx context.Context, mp *model.ModuleProvider) error

	// FindByID retrieves a module provider by ID
	FindByID(ctx context.Context, id int) (*model.ModuleProvider, error)

	// FindByNamespaceModuleProvider retrieves a module provider by namespace/module/provider
	FindByNamespaceModuleProvider(ctx context.Context, namespace types.NamespaceName, module types.ModuleName, provider types.ModuleProviderName) (*model.ModuleProvider, error)

	// FindByNamespace retrieves all module providers in a namespace
	FindByNamespace(ctx context.Context, namespace types.NamespaceName) ([]*model.ModuleProvider, error)

	// Search searches for module providers
	Search(ctx context.Context, query ModuleSearchQuery) (*ModuleSearchResult, error)

	// Delete removes a module provider
	Delete(ctx context.Context, id int) error

	// Exists checks if a module provider exists
	Exists(ctx context.Context, namespace types.NamespaceName, module types.ModuleName, provider types.ModuleProviderName) (bool, error)
}

type OrderDir string

const (
	OrderDirDesc OrderDir = "DESC"
	OrderDirAsc  OrderDir = "ASC"
)

// ModuleSearchQuery represents search criteria
type ModuleSearchQuery struct {
	Query                  string
	Namespaces             []types.NamespaceName // Change from *string to []string for multiple values
	Module                 *types.ModuleName
	Providers              []types.ModuleProviderName // Change from *string to []string for multiple values
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

// ModuleProviderRedirectRepository defines operations for module provider redirects
// Matches Python ModuleProviderRedirect model
type ModuleProviderRedirectRepository interface {
	GetByOriginalDetails(ctx context.Context, namespace types.NamespaceName, module types.ModuleName, provider types.ModuleProviderName, caseInsensitive bool) (*model.ModuleProvider, error)
	GetByModuleProvider(ctx context.Context, moduleProviderID int) ([]*ModuleProviderRedirect, error)
}

// ModuleProviderRedirect represents a redirect from old module provider details to new ones
// Matches Python ModuleProviderRedirect model
type ModuleProviderRedirect struct {
	ID               int
	ModuleProviderID int
	NamespaceID      int
	Module           types.ModuleName
	Provider         types.ModuleProviderName
}
