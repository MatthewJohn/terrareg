package module

import (
	"context"
	"fmt"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/audit/model"
	auditrepo "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/audit/repository"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/module/repository"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/shared/types"
)

// UpdateModuleProviderInheritanceDisabledCommand handles updating whether provider source inheritance is disabled
// Python reference: /app/terrareg/models.py lines 2771-2792
type UpdateModuleProviderInheritanceDisabledCommand struct {
	moduleProviderRepo repository.ModuleProviderRepository
	auditRepo          auditrepo.AuditHistoryRepository
}

// NewUpdateModuleProviderInheritanceDisabledCommand creates a new command
func NewUpdateModuleProviderInheritanceDisabledCommand(
	moduleProviderRepo repository.ModuleProviderRepository,
	auditRepo auditrepo.AuditHistoryRepository,
) *UpdateModuleProviderInheritanceDisabledCommand {
	return &UpdateModuleProviderInheritanceDisabledCommand{
		moduleProviderRepo: moduleProviderRepo,
		auditRepo:          auditRepo,
	}
}

// Execute updates whether provider source inheritance is disabled for a module provider
func (c *UpdateModuleProviderInheritanceDisabledCommand) Execute(
	ctx context.Context,
	namespace types.NamespaceName,
	moduleName types.ModuleName,
	providerName types.ModuleProviderName,
	disabled *bool,
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

	// Only create audit entry and update if value is different
	// Python reference: /app/terrareg/models.py lines 2781-2791
	if disabled == nil {
		return nil
	}

	oldValue := moduleProvider.ProviderSourceInheritanceDisabled()
	if oldValue == *disabled {
		// No change, don't create audit entry or update
		return nil
	}

	// Update inheritance disabled
	if err := moduleProvider.UpdateProviderSourceInheritanceDisabled(ctx, disabled); err != nil {
		return err
	}

	// Create audit entry
	oldStr := fmt.Sprintf("%t", oldValue)
	newStr := fmt.Sprintf("%t", *disabled)

	audit := model.NewAuditHistory(
		username,
		model.AuditActionModuleProviderUpdateProviderSourceInheritanceDisabled,
		"ModuleProvider",
		fmt.Sprintf("%d", moduleProvider.ID()),
		&oldStr,
		&newStr,
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
