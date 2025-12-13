package namespace

import (
	"context"
	"fmt"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/module/repository"
)

// DeleteNamespaceCommand handles deleting a namespace
type DeleteNamespaceCommand struct {
	namespaceRepo repository.NamespaceRepository
}

// NewDeleteNamespaceCommand creates a new delete namespace command
func NewDeleteNamespaceCommand(namespaceRepo repository.NamespaceRepository) *DeleteNamespaceCommand {
	return &DeleteNamespaceCommand{
		namespaceRepo: namespaceRepo,
	}
}

// Execute executes the delete command
func (c *DeleteNamespaceCommand) Execute(ctx context.Context, namespaceName string) error {
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

	return nil
}