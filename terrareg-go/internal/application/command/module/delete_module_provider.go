package module

import (
	"context"
	"fmt"

	"github.com/terrareg/terrareg/internal/domain/module/repository"
)

// DeleteModuleProviderCommand handles deleting a module provider
type DeleteModuleProviderCommand struct {
	moduleProviderRepo repository.ModuleProviderRepository
}

// NewDeleteModuleProviderCommand creates a new delete module provider command
func NewDeleteModuleProviderCommand(moduleProviderRepo repository.ModuleProviderRepository) *DeleteModuleProviderCommand {
	return &DeleteModuleProviderCommand{
		moduleProviderRepo: moduleProviderRepo,
	}
}

// DeleteModuleProviderRequest represents a request to delete a module provider
type DeleteModuleProviderRequest struct {
	Namespace string
	Module    string
	Provider  string
}

// Execute deletes the module provider
func (c *DeleteModuleProviderCommand) Execute(ctx context.Context, req DeleteModuleProviderRequest) error {
	// Find the module provider to ensure it exists
	moduleProvider, err := c.moduleProviderRepo.FindByNamespaceModuleProvider(ctx, req.Namespace, req.Module, req.Provider)
	if err != nil {
		return fmt.Errorf("module provider not found: %w", err)
	}

	// Delete the module provider
	if err := c.moduleProviderRepo.Delete(ctx, moduleProvider.ID()); err != nil {
		return fmt.Errorf("failed to delete module provider: %w", err)
	}

	return nil
}
