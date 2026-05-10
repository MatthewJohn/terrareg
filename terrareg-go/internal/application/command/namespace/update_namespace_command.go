package namespace

import (
	"context"
	"fmt"

	auditservice "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/audit/service"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/module/repository"
	providerSourceService "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/provider_source/service"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/shared/types"
)

// UpdateNamespaceCommand handles updating namespace details
type UpdateNamespaceCommand struct {
	namespaceRepo         repository.NamespaceRepository
	namespaceAuditService *auditservice.NamespaceAuditService
	providerSourceFactory  *providerSourceService.ProviderSourceFactory
}

// UpdateNamespaceRequest represents a request to update a namespace
type UpdateNamespaceRequest struct {
	Name                    *string `json:"name,omitempty"`
	DisplayName             *string `json:"display_name,omitempty"`
	DefaultProviderSource   *string `json:"default_provider_source,omitempty"`
}

// UpdateNamespaceResponse represents the response after updating a namespace
type UpdateNamespaceResponse struct {
	Name                   string  `json:"name"`
	DisplayName            *string `json:"display_name,omitempty"`
	DefaultProviderSource  *string `json:"default_provider_source"`
	ViewURL                string  `json:"view_href,omitempty"`
}

// NewUpdateNamespaceCommand creates a new update namespace command
func NewUpdateNamespaceCommand(
	namespaceRepo repository.NamespaceRepository,
	namespaceAuditService *auditservice.NamespaceAuditService,
	providerSourceFactory *providerSourceService.ProviderSourceFactory,
) *UpdateNamespaceCommand {
	return &UpdateNamespaceCommand{
		namespaceRepo:         namespaceRepo,
		namespaceAuditService: namespaceAuditService,
		providerSourceFactory:  providerSourceFactory,
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

	// Inject provider source factory if available (required for validation)
	if c.providerSourceFactory != nil {
		namespace.SetProviderSourceFactory(c.providerSourceFactory)
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

		// Log audit event for display name change (synchronous)
		// Python reference: /app/terrareg/models.py:1118 - AuditAction.NAMESPACE_MODIFY_DISPLAY_NAME
		if c.namespaceAuditService != nil {
			if err := c.namespaceAuditService.LogNamespaceModifyDisplayName(ctx, namespaceName, oldDisplayName, req.DisplayName); err != nil {
				return nil, fmt.Errorf("failed to log display name audit: %w", err)
			}
		}
	}

	// Update default provider source if provided
	// Python reference: /app/terrareg/models.py lines 1174-1213
	if req.DefaultProviderSource != nil {
		oldValue := namespace.DefaultProviderSourceName()

		// Update the provider source
		if err := namespace.UpdateDefaultProviderSource(ctx, req.DefaultProviderSource); err != nil {
			return nil, err
		}

		// Log audit event for default provider source change (synchronous)
		if c.namespaceAuditService != nil {
			// Convert empty string to nil for audit (unset operation)
			auditNewValue := req.DefaultProviderSource
			if auditNewValue != nil && *auditNewValue == "" {
				auditNewValue = nil
			}
			if err := c.namespaceAuditService.LogNamespaceModifyDefaultProviderSource(ctx, namespaceName, oldValue, auditNewValue); err != nil {
				return nil, fmt.Errorf("failed to log provider source audit: %w", err)
			}
		}
	}

	// Save the updated namespace
	if err := c.namespaceRepo.Save(ctx, namespace); err != nil {
		return nil, fmt.Errorf("failed to update namespace: %w", err)
	}

	// Return response
	response := &UpdateNamespaceResponse{
		Name:                   string(namespace.Name()),
		DisplayName:            namespace.DisplayName(),
		DefaultProviderSource:   namespace.DefaultProviderSourceName(),
		ViewURL:                fmt.Sprintf("/modules/%s", string(namespace.Name())),
	}

	return response, nil
}
