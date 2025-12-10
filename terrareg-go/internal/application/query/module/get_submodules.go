package module

import (
	"context"
	"fmt"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/module/repository"
)

// GetSubmodulesQuery retrieves submodules for a module version
type GetSubmodulesQuery struct {
	moduleProviderRepo repository.ModuleProviderRepository
}

// NewGetSubmodulesQuery creates a new query
func NewGetSubmodulesQuery(
	moduleProviderRepo repository.ModuleProviderRepository,
) *GetSubmodulesQuery {
	return &GetSubmodulesQuery{
		moduleProviderRepo: moduleProviderRepo,
	}
}

// SubmoduleInfo represents information about a submodule
type SubmoduleInfo struct {
	Path string `json:"path"`
	Href string `json:"href"`
}

// Execute retrieves submodules for a module version
func (q *GetSubmodulesQuery) Execute(ctx context.Context, namespace, module, provider, version string) ([]SubmoduleInfo, error) {
	// Find the module provider
	moduleProvider, err := q.moduleProviderRepo.FindByNamespaceModuleProvider(
		ctx, namespace, module, provider,
	)
	if err != nil {
		return nil, fmt.Errorf("module provider not found: %w", err)
	}

	// Find the version
	moduleVersion, err := moduleProvider.GetVersion(version)
	if err != nil {
		return nil, fmt.Errorf("module version not found: %w", err)
	}

	// Check if version is published
	if !moduleVersion.IsPublished() {
		return nil, fmt.Errorf("module version is not published")
	}

	// Get submodules from database using domain model
	submoduleSpecs := moduleVersion.GetSubmodules()

	// Convert to SubmoduleInfo
	result := make([]SubmoduleInfo, len(submoduleSpecs))
	for i, submoduleSpec := range submoduleSpecs {
		// Generate href URL similar to Python terrareg
		href := fmt.Sprintf("/modules/%s/%s/%s/%s/submodule/%s", namespace, module, provider, version, submoduleSpec.Path)
		result[i] = SubmoduleInfo{
			Path: submoduleSpec.Path,
			Href: href,
		}
	}

	return result, nil
}
