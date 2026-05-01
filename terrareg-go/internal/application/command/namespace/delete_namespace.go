package namespace

import (
	"context"
	"fmt"

	auditservice "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/audit/service"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/module/repository"
	providerrepo "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/provider/repository"
	types "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/shared/types"
)

// DeleteNamespaceCommand handles deleting a namespace
type DeleteNamespaceCommand struct {
	namespaceRepo         repository.NamespaceRepository
	moduleProviderRepo    repository.ModuleProviderRepository
	providerRepo          providerrepo.ProviderRepository
	namespaceAuditService *auditservice.NamespaceAuditService
}

// NewDeleteNamespaceCommand creates a new delete namespace command
func NewDeleteNamespaceCommand(
	namespaceRepo repository.NamespaceRepository,
	moduleProviderRepo repository.ModuleProviderRepository,
	providerRepo providerrepo.ProviderRepository,
	namespaceAuditService *auditservice.NamespaceAuditService,
) *DeleteNamespaceCommand {
	return &DeleteNamespaceCommand{
		namespaceRepo:         namespaceRepo,
		moduleProviderRepo:    moduleProviderRepo,
		providerRepo:          providerRepo,
		namespaceAuditService: namespaceAuditService,
	}
}

// Execute executes the delete command
func (c *DeleteNamespaceCommand) Execute(ctx context.Context, namespaceName types.NamespaceName) error {
	// Check if namespace exists
	existing, err := c.namespaceRepo.FindByName(ctx, namespaceName)
	if err != nil {
		return fmt.Errorf("failed to find namespace: %w", err)
	}
	if existing == nil {
		return fmt.Errorf("namespace %s not found", namespaceName)
	}

	// Check for any module providers in the namespace
	// Python reference: /app/terrareg/models.py:1229 - if self.get_all_modules():
	moduleProviders, err := c.moduleProviderRepo.FindByNamespace(ctx, namespaceName)
	if err != nil {
		return fmt.Errorf("failed to check for module providers: %w", err)
	}
	if len(moduleProviders) > 0 {
		return fmt.Errorf("namespace cannot be deleted as it contains modules")
	}

	// Check for any terraform providers in the namespace
	// Python reference: /app/terrareg/models.py:1232 - if self.get_all_providers():
	// Note: We need to check if there are any providers with this namespace ID
	// Since ProviderRepository doesn't have FindByNamespace, we use Search with namespace filter
	providerSearchResult, err := c.providerRepo.Search(ctx, providerrepo.ProviderSearchQuery{
		Namespaces: []string{string(namespaceName)},
		Limit:      1,
	})
	if err != nil {
		return fmt.Errorf("failed to check for providers: %w", err)
	}
	if providerSearchResult.TotalCount > 0 {
		return fmt.Errorf("namespace cannot be deleted as it contains providers")
	}

	// Delete the namespace by ID
	if err := c.namespaceRepo.Delete(ctx, existing.ID()); err != nil {
		return fmt.Errorf("failed to delete namespace: %w", err)
	}

	// Log audit event (synchronous)
	// Python reference: /app/terrareg/models.py:460 - AuditAction.NAMESPACE_DELETE
	if c.namespaceAuditService != nil {
		c.namespaceAuditService.LogNamespaceDelete(ctx, namespaceName)
	}

	return nil
}
