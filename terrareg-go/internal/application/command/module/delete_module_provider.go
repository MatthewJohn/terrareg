package module

import (
	"context"
	"fmt"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/audit/service"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/module/repository"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/interfaces/http/middleware"
)

// DeleteModuleProviderCommand handles deleting a module provider
type DeleteModuleProviderCommand struct {
	moduleProviderRepo repository.ModuleProviderRepository
	auditService       service.ModuleAuditServiceInterface
}

// NewDeleteModuleProviderCommand creates a new delete module provider command
func NewDeleteModuleProviderCommand(
	moduleProviderRepo repository.ModuleProviderRepository,
	auditService service.ModuleAuditServiceInterface,
) *DeleteModuleProviderCommand {
	return &DeleteModuleProviderCommand{
		moduleProviderRepo: moduleProviderRepo,
		auditService:       auditService,
	}
}

// DeleteModuleProviderRequest represents a request to delete a module provider
type DeleteModuleProviderRequest struct {
	Namespace string
	Module    string
	Provider  string
}

// Execute deletes the module provider
func (c *DeleteModuleProviderCommand) Execute(ctx context.Context, req DeleteModuleProviderRequest) error {
	// Find the module provider to ensure it exists
	moduleProvider, err := c.moduleProviderRepo.FindByNamespaceModuleProvider(ctx, req.Namespace, req.Module, req.Provider)
	if err != nil {
		return fmt.Errorf("module provider not found: %w", err)
	}

	// Delete the module provider
	if err := c.moduleProviderRepo.Delete(ctx, moduleProvider.ID()); err != nil {
		return fmt.Errorf("failed to delete module provider: %w", err)
	}

	// Log audit event (async, don't block the response)
	go func() {
		username := "system"
		// Try to get username from auth context if available
		if authCtx := middleware.GetAuthContext(ctx); authCtx.IsAuthenticated {
			username = authCtx.Username
		}

		c.auditService.LogModuleProviderDelete(
			context.Background(),
			username,
			req.Namespace,
			req.Module,
			req.Provider,
		)
	}()

	return nil
}
