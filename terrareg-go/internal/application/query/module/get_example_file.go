package module

import (
	"context"
	"errors"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/module/repository"
)

// GetExampleFileQuery retrieves a specific file from an example
type GetExampleFileQuery struct {
	moduleProviderRepo repository.ModuleProviderRepository
	moduleVersionRepo  repository.ModuleVersionRepository
}

// NewGetExampleFileQuery creates a new query
func NewGetExampleFileQuery(
	moduleProviderRepo repository.ModuleProviderRepository,
	moduleVersionRepo repository.ModuleVersionRepository,
) *GetExampleFileQuery {
	return &GetExampleFileQuery{
		moduleProviderRepo: moduleProviderRepo,
		moduleVersionRepo:  moduleVersionRepo,
	}
}

// Execute retrieves example file content
func (q *GetExampleFileQuery) Execute(ctx context.Context, namespace, moduleName, provider, version, path string) ([]byte, error) {
	// Get module provider first
	moduleProvider, err := q.moduleProviderRepo.FindByNamespaceModuleProvider(ctx, namespace, moduleName, provider)
	if err != nil {
		return nil, err
	}

	if moduleProvider == nil {
		return nil, errors.New("module provider not found")
	}

	// Get module version from the provider
	moduleVersion, err := q.moduleVersionRepo.FindByModuleProviderAndVersion(ctx, moduleProvider.ID(), version)
	if err != nil {
		return nil, err
	}

	if moduleVersion == nil {
		return nil, errors.New("module version not found")
	}

	// TODO: Implement GetExampleFileContent in ModuleVersion model
	// For now, return empty content
	return []byte{}, nil
}