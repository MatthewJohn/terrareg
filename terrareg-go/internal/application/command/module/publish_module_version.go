package module

import (
	"context"
	"errors"
	"fmt"

	"github.com/terrareg/terrareg/internal/domain/module/model"
	"github.com/terrareg/terrareg/internal/domain/module/repository"
	"github.com/terrareg/terrareg/internal/domain/shared"
)

// PublishModuleVersionCommand handles publishing a new module version
type PublishModuleVersionCommand struct {
	moduleProviderRepo repository.ModuleProviderRepository
}

// NewPublishModuleVersionCommand creates a new publish module version command
func NewPublishModuleVersionCommand(moduleProviderRepo repository.ModuleProviderRepository) *PublishModuleVersionCommand {
	return &PublishModuleVersionCommand{
		moduleProviderRepo: moduleProviderRepo,
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

	// Check if version already exists
	existingVersion, err := moduleProvider.GetVersion(req.Version)
	if err == nil && existingVersion != nil {
		return nil, fmt.Errorf("version %s already exists for %s/%s/%s", req.Version, req.Namespace, req.Module, req.Provider)
	}

	// Create module details (can be expanded later with README, etc.)
	details := model.NewModuleDetails(nil) // Empty README for now

	// Publish the version using the aggregate root
	version, err := moduleProvider.PublishVersion(req.Version, details, req.Beta)
	if err != nil {
		return nil, fmt.Errorf("failed to publish version: %w", err)
	}

	// Set optional metadata
	version.SetMetadata(req.Owner, req.Description)

	// Mark as published
	if err := version.Publish(); err != nil {
		return nil, fmt.Errorf("failed to mark version as published: %w", err)
	}

	// Persist the entire aggregate
	if err := c.moduleProviderRepo.Save(ctx, moduleProvider); err != nil {
		return nil, fmt.Errorf("failed to save module provider: %w", err)
	}

	return version, nil
}
