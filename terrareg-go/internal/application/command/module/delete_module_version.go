package module

import (
	"context"
	"errors"

	moduleRepo "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/module/repository"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/shared"
)

// DeleteModuleVersionCommand handles deleting a specific module version
type DeleteModuleVersionCommand struct {
	moduleProviderRepo    moduleRepo.ModuleProviderRepository
	moduleVersionRepo     moduleRepo.ModuleVersionRepository
}

// NewDeleteModuleVersionCommand creates a new delete module version command
func NewDeleteModuleVersionCommand(
	moduleProviderRepo moduleRepo.ModuleProviderRepository,
	moduleVersionRepo moduleRepo.ModuleVersionRepository,
) *DeleteModuleVersionCommand {
	return &DeleteModuleVersionCommand{
		moduleProviderRepo: moduleProviderRepo,
		moduleVersionRepo:  moduleVersionRepo,
	}
}

// DeleteModuleVersionRequest represents a request to delete a module version
type DeleteModuleVersionRequest struct {
	Namespace string
	Module    string
	Provider  string
	Version   string
}

// Execute executes the command to delete a module version
func (c *DeleteModuleVersionCommand) Execute(ctx context.Context, req DeleteModuleVersionRequest) error {
	// Get module provider
	moduleProvider, err := c.moduleProviderRepo.FindByNamespaceModuleProvider(
		ctx,
		req.Namespace,
		req.Module,
		req.Provider,
	)
	if err != nil {
		if errors.Is(err, shared.ErrNotFound) {
			return errors.New("module provider not found")
		}
		return err
	}

	// Get the specific version to delete
	moduleVersion, err := moduleProvider.GetVersion(req.Version)
	if err != nil || moduleVersion == nil {
		return errors.New("module version not found")
	}

	// Delete the module version
	err = c.moduleVersionRepo.Delete(ctx, moduleVersion.ID())
	if err != nil {
		return err
	}

	// TODO: Update provider's latest_version_id if this was the latest
	// TODO: Delete associated files from storage
	// TODO: Create audit event
	// TODO: Delete analytics data

	return nil
}