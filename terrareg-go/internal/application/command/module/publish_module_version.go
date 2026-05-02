package module

import (
	"context"
	"errors"
	"fmt"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/audit/service"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/module/repository"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/shared"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/shared/types"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/interfaces/http/middleware"
)

// PublishModuleVersionCommand handles publishing an existing module version
// Python reference: terrareg/server/api/terrareg_module_version_publish.py
type PublishModuleVersionCommand struct {
	// moduleProviderRepo handles module provider persistence (required)
	moduleProviderRepo repository.ModuleProviderRepository
	// auditService logs audit events (required)
	auditService service.ModuleAuditServiceInterface
}

// NewPublishModuleVersionCommand creates a new publish module version command
// Returns an error if any required dependency is nil
func NewPublishModuleVersionCommand(
	moduleProviderRepo repository.ModuleProviderRepository,
	auditService service.ModuleAuditServiceInterface,
) (*PublishModuleVersionCommand, error) {
	if moduleProviderRepo == nil {
		return nil, fmt.Errorf("moduleProviderRepo cannot be nil")
	}
	if auditService == nil {
		return nil, fmt.Errorf("auditService cannot be nil")
	}

	return &PublishModuleVersionCommand{
		moduleProviderRepo: moduleProviderRepo,
		auditService:       auditService,
	}, nil
}

// Execute executes the publish command
// Python reference: ApiTerraregModuleVersionPublish._post()
// This matches Python behavior:
// 1. Gets the existing module version from database
// 2. Calls publish() which marks it as published and updates latest_version_id
// 3. Returns {'status': 'Success'}
func (c *PublishModuleVersionCommand) Execute(ctx context.Context, namespace, module, provider, version string) error {
	// Find the module provider
	moduleProvider, err := c.moduleProviderRepo.FindByNamespaceModuleProvider(
		ctx,
		types.NamespaceName(namespace),
		types.ModuleName(module),
		types.ModuleProviderName(provider),
	)
	if err != nil {
		if errors.Is(err, shared.ErrNotFound) {
			return fmt.Errorf("module provider %s/%s/%s not found", namespace, module, provider)
		}
		return fmt.Errorf("failed to find module provider: %w", err)
	}

	// Get the existing version - it must exist from upload/index
	// Python: module_version, error = self.get_module_version_by_name(...)
	existingVersion, err := moduleProvider.GetVersion(types.ModuleVersion(version))
	if err != nil {
		return fmt.Errorf("version %s not found", version)
	}

	// Check if already published (idempotent)
	// Python: No explicit check, but publish() would be a no-op for already published versions
	if existingVersion.IsPublished() {
		// Already published, return success (idempotent)
		return nil
	}

	// Mark as published
	// Python: module_version.publish()
	// This sets published=True and updates module_provider.latest_version_id if this is the latest non-beta version
	if err := existingVersion.Publish(); err != nil {
		return fmt.Errorf("failed to mark version as published: %w", err)
	}

	// Persist the entire aggregate (saves latest_version_id to database)
	// Python: self.update_attributes(published=True)
	//         self._module_provider.update_attributes(latest_version_id=self.pk)
	if err := c.moduleProviderRepo.Save(ctx, moduleProvider); err != nil {
		return fmt.Errorf("failed to save module provider: %w", err)
	}

	// Log audit event
	// Python: AuditEvent.create_audit_event(action=AuditAction.MODULE_VERSION_PUBLISH, ...)
	username := "system"
	if authCtx := middleware.GetAuthContext(ctx); authCtx.IsAuthenticated() {
		username = authCtx.GetUsername()
	}

	_ = c.auditService.LogModuleVersionPublish(
		ctx,
		types.NamespaceName(username),
		types.NamespaceName(namespace),
		types.ModuleName(module),
		types.ModuleProviderName(provider),
		types.ModuleVersion(version),
	)

	return nil
}
