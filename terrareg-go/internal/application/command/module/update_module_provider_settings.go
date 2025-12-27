package module

import (
	"context"
	"fmt"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/module/repository"
)

// UpdateModuleProviderSettingsCommand handles updating module provider settings
type UpdateModuleProviderSettingsCommand struct {
	moduleProviderRepo repository.ModuleProviderRepository
}

// NewUpdateModuleProviderSettingsCommand creates a new command
func NewUpdateModuleProviderSettingsCommand(moduleProviderRepo repository.ModuleProviderRepository) *UpdateModuleProviderSettingsCommand {
	return &UpdateModuleProviderSettingsCommand{
		moduleProviderRepo: moduleProviderRepo,
	}
}

// UpdateModuleProviderSettingsRequest represents the request to update settings
type UpdateModuleProviderSettingsRequest struct {
	Namespace string
	Module    string
	Provider  string

	// Git configuration
	GitProviderID         *int
	RepoBaseURLTemplate   *string
	RepoCloneURLTemplate  *string
	RepoBrowseURLTemplate *string
	GitTagFormat          *string
	GitPath               *string
	ArchiveGitPath        *bool

	// Module settings
	Verified *bool
}

// Execute updates the module provider settings
func (c *UpdateModuleProviderSettingsCommand) Execute(ctx context.Context, req UpdateModuleProviderSettingsRequest) error {
	// Find the module provider
	moduleProvider, err := c.moduleProviderRepo.FindByNamespaceModuleProvider(ctx, req.Namespace, req.Module, req.Provider)
	if err != nil {
		return fmt.Errorf("module provider not found: %w", err)
	}

	// Update Git configuration if provided
	if req.GitProviderID != nil || req.RepoBaseURLTemplate != nil || req.RepoCloneURLTemplate != nil ||
		req.RepoBrowseURLTemplate != nil || req.GitTagFormat != nil || req.GitPath != nil || req.ArchiveGitPath != nil {

		archiveGitPath := false
		if req.ArchiveGitPath != nil {
			archiveGitPath = *req.ArchiveGitPath
		}

		moduleProvider.SetGitConfiguration(
			req.GitProviderID,
			req.RepoBaseURLTemplate,
			req.RepoCloneURLTemplate,
			req.RepoBrowseURLTemplate,
			req.GitTagFormat,
			req.GitPath,
			archiveGitPath,
		)
	}

	// Update verified status if provided
	if req.Verified != nil {
		if *req.Verified {
			moduleProvider.Verify()
		} else {
			moduleProvider.Unverify()
		}
	}

	// Persist changes
	if err := c.moduleProviderRepo.Save(ctx, moduleProvider); err != nil {
		return fmt.Errorf("failed to save module provider: %w", err)
	}

	return nil
}
