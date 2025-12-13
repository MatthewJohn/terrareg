package module

import (
	"context"
	"errors"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/module/repository"
)

// GetSubmoduleDetailsQuery retrieves details for a specific submodule
type GetSubmoduleDetailsQuery struct {
	moduleProviderRepo repository.ModuleProviderRepository
	moduleVersionRepo repository.ModuleVersionRepository
}

// NewGetSubmoduleDetailsQuery creates a new query
func NewGetSubmoduleDetailsQuery(
	moduleProviderRepo repository.ModuleProviderRepository,
	moduleVersionRepo repository.ModuleVersionRepository,
) *GetSubmoduleDetailsQuery {
	return &GetSubmoduleDetailsQuery{
		moduleProviderRepo:  moduleProviderRepo,
		moduleVersionRepo:  moduleVersionRepo,
	}
}

// SubmoduleDetails represents submodule details
type SubmoduleDetails struct {
	Path        string            `json:"path"`
	Description string            `json:"description,omitempty"`
	Readme      string            `json:"readme,omitempty"`
	Files       []SubmoduleFile `json:"files,omitempty"`
}

// SubmoduleFile represents a file in a submodule
type SubmoduleFile struct {
	Path     string `json:"path"`
	Content  string `json:"content,omitempty"`
	IsBinary bool   `json:"is_binary"`
}

// Execute retrieves submodule details
func (q *GetSubmoduleDetailsQuery) Execute(ctx context.Context, namespace, moduleName, provider, version, path string) (*SubmoduleDetails, error) {
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

	// TODO: Implement GetSubmoduleByPath in ModuleVersion model
	// The ModuleVersion model should have submodules and examples
	// For now, return basic details
	return &SubmoduleDetails{
		Path:        path,
		Description: "Submodule at " + path,
	}, nil
}