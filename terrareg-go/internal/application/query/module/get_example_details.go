package module

import (
	"context"
	"errors"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/module/repository"
)

// GetExampleDetailsQuery retrieves details for a specific example
type GetExampleDetailsQuery struct {
	moduleProviderRepo  repository.ModuleProviderRepository
	moduleVersionRepo  repository.ModuleVersionRepository
}

// NewGetExampleDetailsQuery creates a new query
func NewGetExampleDetailsQuery(
	moduleProviderRepo repository.ModuleProviderRepository,
	moduleVersionRepo repository.ModuleVersionRepository,
) *GetExampleDetailsQuery {
	return &GetExampleDetailsQuery{
		moduleProviderRepo:  moduleProviderRepo,
		moduleVersionRepo:  moduleVersionRepo,
	}
}

// ExampleDetails represents example details
type ExampleDetails struct {
	Path        string        `json:"path"`
	Description string        `json:"description,omitempty"`
	Readme      string        `json:"readme,omitempty"`
	Files       []ModuleFile   `json:"files,omitempty"`
}

// ModuleFile represents a file in an example
type ModuleFile struct {
	Path     string `json:"path"`
	Content  string `json:"content,omitempty"`
	IsBinary bool   `json:"is_binary"`
}

// Execute retrieves example details
func (q *GetExampleDetailsQuery) Execute(ctx context.Context, namespace, moduleName, provider, version, path string) (*ExampleDetails, error) {
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

	// TODO: Implement GetExampleByPath in ModuleVersion model
	// The ModuleVersion model should have examples
	// For now, return basic details
	return &ExampleDetails{
		Path:        path,
		Description: "Example at " + path,
	}, nil
}