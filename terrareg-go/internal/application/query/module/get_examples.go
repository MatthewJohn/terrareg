package module

import (
	"context"
	"fmt"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/module/repository"
	)

// GetExamplesQuery retrieves examples for a module version
type GetExamplesQuery struct {
	moduleProviderRepo repository.ModuleProviderRepository
}

// NewGetExamplesQuery creates a new query
func NewGetExamplesQuery(
	moduleProviderRepo repository.ModuleProviderRepository,
) *GetExamplesQuery {
	return &GetExamplesQuery{
		moduleProviderRepo: moduleProviderRepo,
	}
}

// ExampleInfo represents information about an example
type ExampleInfo struct {
	Path string `json:"path"`
	Href string `json:"href"`
}

// Execute retrieves examples for a module version
func (q *GetExamplesQuery) Execute(ctx context.Context, namespace, module, provider, version string) ([]ExampleInfo, error) {
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

	// Get examples from database using domain model
	exampleSpecs := moduleVersion.GetExamples()

	// Convert to ExampleInfo
	result := make([]ExampleInfo, len(exampleSpecs))
	for i, exampleSpec := range exampleSpecs {
		// Generate href URL similar to Python terrareg
		href := fmt.Sprintf("/modules/%s/%s/%s/%s/example/%s", namespace, module, provider, version, exampleSpec.Path)
		result[i] = ExampleInfo{
			Path: exampleSpec.Path,
			Href: href,
		}
	}

	return result, nil
}
