package module

import (
	"context"
	"errors"
	"fmt"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/audit/service"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/module/model"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/module/repository"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/shared"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/interfaces/http/middleware"
)

// PublishModuleVersionCommand handles publishing a new module version
type PublishModuleVersionCommand struct {
	moduleProviderRepo repository.ModuleProviderRepository
	auditService       service.ModuleAuditServiceInterface
}

// NewPublishModuleVersionCommand creates a new publish module version command
func NewPublishModuleVersionCommand(
	moduleProviderRepo repository.ModuleProviderRepository,
	auditService service.ModuleAuditServiceInterface,
) *PublishModuleVersionCommand {
	return &PublishModuleVersionCommand{
		moduleProviderRepo: moduleProviderRepo,
		auditService:       auditService,
	}
}

// PublishModuleVersionRequest represents the request to publish a module version
type PublishModuleVersionRequest struct {
	Namespace   string
	Module      string
	Provider    string
	Version     string
	Beta        bool
	Description *string
	Owner       *string
}

// Execute executes the command
func (c *PublishModuleVersionCommand) Execute(ctx context.Context, req PublishModuleVersionRequest) (*model.ModuleVersion, error) {
	// Find the module provider
	moduleProvider, err := c.moduleProviderRepo.FindByNamespaceModuleProvider(ctx, req.Namespace, req.Module, req.Provider)
	if err != nil {
		if errors.Is(err, shared.ErrNotFound) {
			return nil, fmt.Errorf("module provider %s/%s/%s not found", req.Namespace, req.Module, req.Provider)
		}
		return nil, fmt.Errorf("failed to find module provider: %w", err)
	}

	// Check if version already exists (e.g., from upload)
	existingVersion, err := moduleProvider.GetVersion(req.Version)
	var version *model.ModuleVersion
	if err == nil && existingVersion != nil {
		// Version already exists (likely from upload)
		version = existingVersion
		// Check if already published
		if version.IsPublished() {
			// Already published, return success (idempotent)
			return version, nil
		}
		// Set optional metadata if not already set
		if req.Owner != nil || req.Description != nil {
			version.SetMetadata(req.Owner, req.Description)
		}
		// Mark as published
		if err := version.Publish(); err != nil {
			return nil, fmt.Errorf("failed to mark version as published: %w", err)
		}
	} else {
		// Version doesn't exist, create it
		// Create module details (can be expanded later with README, etc.)
		details := model.NewModuleDetails(nil) // Empty README for now

		// Publish the version using the aggregate root
		version, err = moduleProvider.PublishVersion(req.Version, details, req.Beta)
		if err != nil {
			return nil, fmt.Errorf("failed to publish version: %w", err)
		}

		// Set optional metadata
		version.SetMetadata(req.Owner, req.Description)

		// Mark as published
		if err := version.Publish(); err != nil {
			return nil, fmt.Errorf("failed to mark version as published: %w", err)
		}
	}

	// Persist the entire aggregate
	if err := c.moduleProviderRepo.Save(ctx, moduleProvider); err != nil {
		return nil, fmt.Errorf("failed to save module provider: %w", err)
	}

	// Log audit event (async, don't block the response)
	go func() {
		// Get username from context
		username := "system"
		if authCtx := middleware.GetAuthContext(ctx); authCtx.IsAuthenticated {
			username = authCtx.Username
		}

		// Log the version index and publish
		c.auditService.LogModuleVersionIndex(
			context.Background(),
			username,
			req.Namespace,
			req.Module,
			req.Provider,
			req.Version,
		)

		c.auditService.LogModuleVersionPublish(
			context.Background(),
			username,
			req.Namespace,
			req.Module,
			req.Provider,
			req.Version,
		)
	}()

	return version, nil
}
