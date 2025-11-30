package module

import (
	"context"
	"fmt"

	"github.com/terrareg/terrareg/internal/domain/module/repository"
)

// ImportModuleVersionCommand handles importing module versions from Git
type ImportModuleVersionCommand struct {
	moduleProviderRepo repository.ModuleProviderRepository
}

// NewImportModuleVersionCommand creates a new command
func NewImportModuleVersionCommand(
	moduleProviderRepo repository.ModuleProviderRepository,
) *ImportModuleVersionCommand {
	return &ImportModuleVersionCommand{
		moduleProviderRepo: moduleProviderRepo,
	}
}

// ImportModuleVersionRequest represents the import request
type ImportModuleVersionRequest struct {
	Namespace string
	Module    string
	Provider  string
	Version   *string // Optional - derived from git tag if not provided
	GitTag    *string // Optional - conflicts with Version
}

// Execute imports a module version from Git
func (c *ImportModuleVersionCommand) Execute(ctx context.Context, req ImportModuleVersionRequest) error {
	// Validate that either version or git_tag is provided (not both, not neither)
	if (req.Version == nil && req.GitTag == nil) || (req.Version != nil && req.GitTag != nil) {
		return fmt.Errorf("either version or git_tag must be provided (but not both)")
	}

	// Find the module provider
	moduleProvider, err := c.moduleProviderRepo.FindByNamespaceModuleProvider(
		ctx, req.Namespace, req.Module, req.Provider,
	)
	if err != nil {
		return fmt.Errorf("module provider not found: %w", err)
	}

	// TODO: Validate that the module provider has Git configuration
	// TODO: If git_tag is provided, derive version from it
	// TODO: Clone the Git repository
	// TODO: Extract module files
	// TODO: Run terraform-docs to extract metadata
	// TODO: Parse README
	// TODO: Create/update module version
	// TODO: Publish the version

	// For now, return a not implemented error
	_ = moduleProvider
	return fmt.Errorf("Git import is not yet fully implemented")
}
