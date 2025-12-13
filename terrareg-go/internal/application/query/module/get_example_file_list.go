package module

import (
	"context"
	"errors"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/module/repository"
)

// GetExampleFileListQuery retrieves the list of files in an example
type GetExampleFileListQuery struct {
	moduleProviderRepo repository.ModuleProviderRepository
	moduleVersionRepo  repository.ModuleVersionRepository
}

// NewGetExampleFileListQuery creates a new query
func NewGetExampleFileListQuery(
	moduleProviderRepo repository.ModuleProviderRepository,
	moduleVersionRepo repository.ModuleVersionRepository,
) *GetExampleFileListQuery {
	return &GetExampleFileListQuery{
		moduleProviderRepo: moduleProviderRepo,
		moduleVersionRepo:  moduleVersionRepo,
	}
}

// ExampleFile represents a file in an example
type ExampleFile struct {
	Path string `json:"path"`
	Type string `json:"type"` // file or directory
}

// Execute retrieves example file list
func (q *GetExampleFileListQuery) Execute(ctx context.Context, namespace, moduleName, provider, version, path string) ([]ExampleFile, error) {
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

	// TODO: Implement GetExampleFiles in ModuleVersion model
	// For now, return empty list
	return []ExampleFile{}, nil
}