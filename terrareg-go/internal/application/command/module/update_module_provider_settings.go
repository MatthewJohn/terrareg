package module

import (
	"context"
	"fmt"

	auditService "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/audit/service"
	gitModel "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/git/model"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/module/repository"
	providerSourceService "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/provider_source/service"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/shared/types"
)

// UpdateModuleProviderSettingsCommand handles updating module provider settings
type UpdateModuleProviderSettingsCommand struct {
	moduleProviderRepo      repository.ModuleProviderRepository
	providerSourceFactory   *providerSourceService.ProviderSourceFactory
	moduleAuditService      *auditService.ModuleAuditService
}

// NewUpdateModuleProviderSettingsCommand creates a new command
func NewUpdateModuleProviderSettingsCommand(
	moduleProviderRepo repository.ModuleProviderRepository,
	providerSourceFactory *providerSourceService.ProviderSourceFactory,
	moduleAuditService *auditService.ModuleAuditService,
) *UpdateModuleProviderSettingsCommand {
	return &UpdateModuleProviderSettingsCommand{
		moduleProviderRepo:    moduleProviderRepo,
		providerSourceFactory: providerSourceFactory,
		moduleAuditService:    moduleAuditService,
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

	// Provider source settings
	ProviderSource                    *string
	ProviderSourceInheritanceDisabled   *bool
}

// Execute updates the module provider settings
func (c *UpdateModuleProviderSettingsCommand) Execute(ctx context.Context, req UpdateModuleProviderSettingsRequest) error {
	// Find the module provider
	moduleProvider, err := c.moduleProviderRepo.FindByNamespaceModuleProvider(ctx, types.NamespaceName(req.Namespace), types.ModuleName(req.Module), types.ModuleProviderName(req.Provider))
	if err != nil {
		return fmt.Errorf("module provider not found: %w", err)
	}

	// Inject provider source factory if available (required for validation)
	if c.providerSourceFactory != nil {
		moduleProvider.SetProviderSourceFactory(c.providerSourceFactory)
	}

	// Validate git_tag_format if provided
	// Uses same validation as create - GitTagFormat.Validate()
	if req.GitTagFormat != nil && *req.GitTagFormat != "" {
		gitTagFormat := gitModel.NewGitTagFormat(*req.GitTagFormat)
		if err := gitTagFormat.Validate(); err != nil {
			return err
		}
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

	// Update provider source settings if provided
	if req.ProviderSource != nil {
		oldValue := moduleProvider.ProviderSourceName()

		// Empty string means "clear/unset" the provider source
		// Non-empty string means set to that value
		if err := moduleProvider.UpdateProviderSource(ctx, req.ProviderSource); err != nil {
			return err
		}

		// Log audit event for provider source change
		if c.moduleAuditService != nil {
			// Convert empty string to nil for audit (unset operation)
			newValue := req.ProviderSource
			if newValue != nil && *newValue == "" {
				newValue = nil
			}

			c.moduleAuditService.LogModuleProviderUpdateProviderSource(ctx, types.NamespaceName("anonymous"), moduleProvider.ID(), oldValue, newValue)
		}
	}

	if req.ProviderSourceInheritanceDisabled != nil {
		oldValue := moduleProvider.ProviderSourceInheritanceDisabled()

		if err := moduleProvider.UpdateProviderSourceInheritanceDisabled(ctx, req.ProviderSourceInheritanceDisabled); err != nil {
			return err
		}

		// Log audit event for inheritance disabled change (only if value actually changed)
		if c.moduleAuditService != nil && oldValue != *req.ProviderSourceInheritanceDisabled {
			c.moduleAuditService.LogModuleProviderUpdateProviderSourceInheritanceDisabled(ctx, types.NamespaceName("anonymous"), moduleProvider.ID(), oldValue, *req.ProviderSourceInheritanceDisabled)
		}
	}

	// Persist changes
	if err := c.moduleProviderRepo.Save(ctx, moduleProvider); err != nil {
		return fmt.Errorf("failed to save module provider: %w", err)
	}

	return nil
}
