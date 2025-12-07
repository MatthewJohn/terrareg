package namespace

import (
	"context"
	"fmt"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/module/repository"
)

// UpdateNamespaceCommand handles updating namespace details
type UpdateNamespaceCommand struct {
	namespaceRepo repository.NamespaceRepository
}

// UpdateNamespaceRequest represents a request to update a namespace
type UpdateNamespaceRequest struct {
	Name        *string `json:"name,omitempty"`
	DisplayName *string `json:"display_name,omitempty"`
}

// UpdateNamespaceResponse represents the response after updating a namespace
type UpdateNamespaceResponse struct {
	Name        string  `json:"name"`
	DisplayName *string `json:"display_name,omitempty"`
	ViewURL     string  `json:"view_href,omitempty"`
}

// NewUpdateNamespaceCommand creates a new update namespace command
func NewUpdateNamespaceCommand(namespaceRepo repository.NamespaceRepository) *UpdateNamespaceCommand {
	return &UpdateNamespaceCommand{
		namespaceRepo: namespaceRepo,
	}
}

// Execute executes the update namespace command
func (c *UpdateNamespaceCommand) Execute(ctx context.Context, namespaceName string, req UpdateNamespaceRequest) (*UpdateNamespaceResponse, error) {
	// Get existing namespace
	namespace, err := c.namespaceRepo.FindByName(ctx, namespaceName)
	if err != nil {
		return nil, fmt.Errorf("failed to find namespace: %w", err)
	}

	if namespace == nil {
		return nil, fmt.Errorf("namespace '%s' not found", namespaceName)
	}

	// TODO: Implement name change with redirect logic
	// For now, we only support display name updates
	if req.Name != nil {
		return nil, fmt.Errorf("namespace name changes are not yet supported")
	}

	// Update display name if provided and different
	if req.DisplayName != nil {
		// Note: In Python, display name validation allows spaces, hyphens, underscores
		// For now, we'll just set it as-is
		namespace.SetDisplayName(req.DisplayName)
	}

	// Save the updated namespace
	if err := c.namespaceRepo.Save(ctx, namespace); err != nil {
		return nil, fmt.Errorf("failed to update namespace: %w", err)
	}

	// Return response
	response := &UpdateNamespaceResponse{
		Name:        namespace.Name(),
		DisplayName: namespace.DisplayName(),
		// TODO: Generate view URL when URL service is available
		// ViewURL: namespace.GetViewURL(),
	}

	return response, nil
}
