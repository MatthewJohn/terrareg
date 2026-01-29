package namespace

import (
	"context"
	"fmt"

	auditservice "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/audit/service"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/module/repository"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/shared/types"
)

// UpdateNamespaceCommand handles updating namespace details
type UpdateNamespaceCommand struct {
	namespaceRepo         repository.NamespaceRepository
	namespaceAuditService *auditservice.NamespaceAuditService
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
func NewUpdateNamespaceCommand(namespaceRepo repository.NamespaceRepository, namespaceAuditService *auditservice.NamespaceAuditService) *UpdateNamespaceCommand {
	return &UpdateNamespaceCommand{
		namespaceRepo:         namespaceRepo,
		namespaceAuditService: namespaceAuditService,
	}
}

// Execute executes the update namespace command
func (c *UpdateNamespaceCommand) Execute(ctx context.Context, namespaceName types.NamespaceName, req UpdateNamespaceRequest) (*UpdateNamespaceResponse, error) {
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
		oldDisplayName := namespace.DisplayName()
		namespace.SetDisplayName(req.DisplayName)

		// Log audit event for display name change (async, non-blocking)
		// Python reference: /app/terrareg/models.py:1118 - AuditAction.NAMESPACE_MODIFY_DISPLAY_NAME
		go c.namespaceAuditService.LogNamespaceModifyDisplayName(ctx, namespaceName, oldDisplayName, req.DisplayName)
	}

	// Save the updated namespace
	if err := c.namespaceRepo.Save(ctx, namespace); err != nil {
		return nil, fmt.Errorf("failed to update namespace: %w", err)
	}

	// Return response
	response := &UpdateNamespaceResponse{
		Name:        string(namespace.Name()),
		DisplayName: namespace.DisplayName(),
		// TODO: Generate view URL when URL service is available
		// ViewURL: namespace.GetViewURL(),
	}

	return response, nil
}
