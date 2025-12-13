package module

import (
	"context"
	"errors"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/module/repository"
)

// GetSubmoduleReadmeHTMLQuery retrieves README HTML for a submodule
type GetSubmoduleReadmeHTMLQuery struct {
	moduleProviderRepo repository.ModuleProviderRepository
	moduleVersionRepo  repository.ModuleVersionRepository
}

// NewGetSubmoduleReadmeHTMLQuery creates a new query
func NewGetSubmoduleReadmeHTMLQuery(
	moduleProviderRepo repository.ModuleProviderRepository,
	moduleVersionRepo repository.ModuleVersionRepository,
) *GetSubmoduleReadmeHTMLQuery {
	return &GetSubmoduleReadmeHTMLQuery{
		moduleProviderRepo: moduleProviderRepo,
		moduleVersionRepo:  moduleVersionRepo,
	}
}

// Execute retrieves submodule README as HTML
func (q *GetSubmoduleReadmeHTMLQuery) Execute(ctx context.Context, namespace, moduleName, provider, version, path string) (string, error) {
	// Get module provider first
	moduleProvider, err := q.moduleProviderRepo.FindByNamespaceModuleProvider(ctx, namespace, moduleName, provider)
	if err != nil {
		return "", err
	}

	if moduleProvider == nil {
		return "", errors.New("module provider not found")
	}

	// Get module version from the provider
	moduleVersion, err := q.moduleVersionRepo.FindByModuleProviderAndVersion(ctx, moduleProvider.ID(), version)
	if err != nil {
		return "", err
	}

	if moduleVersion == nil {
		return "", errors.New("module version not found")
	}

	// TODO: Implement GetSubmoduleReadme in ModuleVersion model
	// For now, return placeholder
	return "<p>README for submodule at " + path + "</p>", nil
}