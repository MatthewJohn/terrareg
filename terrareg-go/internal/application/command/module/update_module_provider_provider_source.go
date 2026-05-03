package module

import (
	"context"
	"fmt"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/audit/model"
	auditrepo "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/audit/repository"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/module/repository"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/shared/types"
)

// UpdateModuleProviderProviderSourceCommand handles updating the provider source for a module provider
// Python reference: /app/terrareg/models.py lines 2730-2769
type UpdateModuleProviderProviderSourceCommand struct {
	moduleProviderRepo repository.ModuleProviderRepository
	auditRepo          auditrepo.AuditHistoryRepository
}

// NewUpdateModuleProviderProviderSourceCommand creates a new command
func NewUpdateModuleProviderProviderSourceCommand(
	moduleProviderRepo repository.ModuleProviderRepository,
	auditRepo auditrepo.AuditHistoryRepository,
) *UpdateModuleProviderProviderSourceCommand {
	return &UpdateModuleProviderProviderSourceCommand{
		moduleProviderRepo: moduleProviderRepo,
		auditRepo:          auditRepo,
	}
}

// Execute updates the provider source for a module provider
func (c *UpdateModuleProviderProviderSourceCommand) Execute(
	ctx context.Context,
	namespace types.NamespaceName,
	moduleName types.ModuleName,
	providerName types.ModuleProviderName,
	providerSource *string,
	username string,
) error {
	// Find the module provider
	moduleProvider, err := c.moduleProviderRepo.FindByNamespaceModuleProvider(
		ctx,
		namespace,
		moduleName,
		providerName,
	)
	if err != nil {
		return fmt.Errorf("module provider not found: %w", err)
	}

	// Get old value for audit
	oldValue := moduleProvider.ProviderSourceName()

	// Update provider source
	if err := moduleProvider.UpdateProviderSource(ctx, providerSource); err != nil {
		return err
	}

	// Create audit entry
	// Python reference: /app/terrareg/models.py lines 2761-2767
	var oldValueStr *string
	if oldValue != nil {
		oldValueStr = oldValue
	}
	var newValueStr *string
	if providerSource != nil && *providerSource != "" {
		newValueStr = providerSource
	}

	audit := model.NewAuditHistory(
		username,
		model.AuditActionModuleProviderUpdateProviderSource,
		"ModuleProvider",
		fmt.Sprintf("%d", moduleProvider.ID()),
		oldValueStr,
		newValueStr,
	)

	if err := c.auditRepo.Create(ctx, audit); err != nil {
		return fmt.Errorf("failed to create audit entry: %w", err)
	}

	// Persist changes
	if err := c.moduleProviderRepo.Save(ctx, moduleProvider); err != nil {
		return fmt.Errorf("failed to save module provider: %w", err)
	}

	return nil
}
