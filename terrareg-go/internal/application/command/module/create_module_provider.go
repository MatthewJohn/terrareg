package module

import (
	"context"
	"errors"
	"fmt"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/module/model"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/module/repository"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/shared"
)

// CreateModuleProviderCommand handles creating a new module provider
type CreateModuleProviderCommand struct {
	namespaceRepo      repository.NamespaceRepository
	moduleProviderRepo repository.ModuleProviderRepository
}

// NewCreateModuleProviderCommand creates a new create module provider command
func NewCreateModuleProviderCommand(
	namespaceRepo repository.NamespaceRepository,
	moduleProviderRepo repository.ModuleProviderRepository,
) *CreateModuleProviderCommand {
	return &CreateModuleProviderCommand{
		namespaceRepo:      namespaceRepo,
		moduleProviderRepo: moduleProviderRepo,
	}
}

// CreateModuleProviderRequest represents the request to create a module provider
type CreateModuleProviderRequest struct {
	Namespace string
	Module    string
	Provider  string
}

// Execute executes the command
func (c *CreateModuleProviderCommand) Execute(ctx context.Context, req CreateModuleProviderRequest) (*model.ModuleProvider, error) {
	// Find the namespace
	namespace, err := c.namespaceRepo.FindByName(ctx, req.Namespace)
	if err != nil {
		if errors.Is(err, shared.ErrNotFound) {
			return nil, fmt.Errorf("namespace %s not found", req.Namespace)
		}
		return nil, fmt.Errorf("failed to find namespace: %w", err)
	}

	// Check if module provider already exists
	existing, err := c.moduleProviderRepo.FindByNamespaceModuleProvider(ctx, req.Namespace, req.Module, req.Provider)
	if err != nil && !errors.Is(err, shared.ErrNotFound) {
		return nil, fmt.Errorf("failed to check module provider existence: %w", err)
	}
	if existing != nil {
		return nil, fmt.Errorf("module provider %s/%s/%s already exists", req.Namespace, req.Module, req.Provider)
	}

	// Create module provider domain model
	moduleProvider, err := model.NewModuleProvider(namespace, req.Module, req.Provider)
	if err != nil {
		return nil, fmt.Errorf("failed to create module provider: %w", err)
	}

	// Persist to repository
	if err := c.moduleProviderRepo.Save(ctx, moduleProvider); err != nil {
		return nil, fmt.Errorf("failed to save module provider: %w", err)
	}

	return moduleProvider, nil
}
