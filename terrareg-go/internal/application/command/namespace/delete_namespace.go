package namespace

import (
	"context"
	"fmt"

	auditservice "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/audit/service"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/module/repository"
	types "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/shared/types"
)

// DeleteNamespaceCommand handles deleting a namespace
type DeleteNamespaceCommand struct {
	namespaceRepo         repository.NamespaceRepository
	namespaceAuditService *auditservice.NamespaceAuditService
}

// NewDeleteNamespaceCommand creates a new delete namespace command
func NewDeleteNamespaceCommand(namespaceRepo repository.NamespaceRepository, namespaceAuditService *auditservice.NamespaceAuditService) *DeleteNamespaceCommand {
	return &DeleteNamespaceCommand{
		namespaceRepo:         namespaceRepo,
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

	// Delete the namespace by ID
	if err := c.namespaceRepo.Delete(ctx, existing.ID()); err != nil {
		return fmt.Errorf("failed to delete namespace: %w", err)
	}

	// Log audit event (async, non-blocking)
	// Python reference: /app/terrareg/models.py:460 - AuditAction.NAMESPACE_DELETE
	go c.namespaceAuditService.LogNamespaceDelete(ctx, namespaceName)

	return nil
}
