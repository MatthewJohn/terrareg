package module

import (
	"context"
	"errors"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/module/repository"
)

// GetExampleReadmeHTMLQuery retrieves README HTML for an example
type GetExampleReadmeHTMLQuery struct {
	moduleProviderRepo repository.ModuleProviderRepository
	moduleVersionRepo  repository.ModuleVersionRepository
}

// NewGetExampleReadmeHTMLQuery creates a new query
func NewGetExampleReadmeHTMLQuery(
	moduleProviderRepo repository.ModuleProviderRepository,
	moduleVersionRepo repository.ModuleVersionRepository,
) *GetExampleReadmeHTMLQuery {
	return &GetExampleReadmeHTMLQuery{
		moduleProviderRepo: moduleProviderRepo,
		moduleVersionRepo:  moduleVersionRepo,
	}
}

// Execute retrieves example README as HTML
func (q *GetExampleReadmeHTMLQuery) Execute(ctx context.Context, namespace, moduleName, provider, version, path string) (string, error) {
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

	// TODO: Implement GetExampleReadme in ModuleVersion model
	// For now, return placeholder
	return "<p>README for example at " + path + "</p>", nil
}