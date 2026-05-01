package module

import (
	"context"
	"errors"
	"fmt"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/module/model"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/module/repository"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/shared"
	types "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/shared/types"
)

// GetModuleVersionQuery handles retrieving a specific module version
type GetModuleVersionQuery struct {
	// moduleProviderRepo handles module provider persistence (required)
	moduleProviderRepo repository.ModuleProviderRepository
}

// NewGetModuleVersionQuery creates a new get module version query
// Returns an error if moduleProviderRepo is nil
func NewGetModuleVersionQuery(moduleProviderRepo repository.ModuleProviderRepository) (*GetModuleVersionQuery, error) {
	if moduleProviderRepo == nil {
		return nil, fmt.Errorf("moduleProviderRepo cannot be nil")
	}
	return &GetModuleVersionQuery{
		moduleProviderRepo: moduleProviderRepo,
	}, nil
}

// Execute executes the query
func (q *GetModuleVersionQuery) Execute(ctx context.Context, namespace types.NamespaceName, module types.ModuleName, provider types.ModuleProviderName, version types.ModuleVersion) (*model.ModuleVersion, error) {
	// First get the module provider
	moduleProvider, err := q.moduleProviderRepo.FindByNamespaceModuleProvider(ctx, namespace, module, provider)
	if err != nil {
		if errors.Is(err, shared.ErrNotFound) {
			return nil, fmt.Errorf("module provider %s/%s/%s not found", namespace, module, provider)
		}
		return nil, fmt.Errorf("failed to get module provider: %w", err)
	}

	// Get the specific version
	moduleVersion, err := moduleProvider.GetVersion(version)
	if err != nil || moduleVersion == nil {
		return nil, fmt.Errorf("version %s not found for %s/%s/%s", version, namespace, module, provider)
	}

	return moduleVersion, nil
}
