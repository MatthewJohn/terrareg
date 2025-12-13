package module

import (
	"context"
	"fmt"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/shared"
)

// CreateModuleProviderRedirectRequest represents a request to create a module provider redirect
type CreateModuleProviderRedirectRequest struct {
	FromNamespace      string
	FromModule         string
	FromProvider       string
	ToModuleProviderID int
}

// CreateModuleProviderRedirectCommand creates a module provider redirect
type CreateModuleProviderRedirectCommand struct {
	redirectRepo ModuleProviderRedirectRepository
}

// ModuleProviderRedirectRepository defines the interface for module provider redirect persistence
type ModuleProviderRedirectRepository interface {
	Create(ctx context.Context, req CreateModuleProviderRedirectRequest) error
	GetAll(ctx context.Context) ([]*ModuleProviderRedirect, error)
	GetByFrom(ctx context.Context, namespace, module, provider string) (*ModuleProviderRedirect, error)
	Delete(ctx context.Context, namespace, module, provider string) error
}

// ModuleProviderRedirect represents a module provider redirect
type ModuleProviderRedirect struct {
	ID               int
	FromNamespace    string
	FromModule       string
	FromProvider     string
	ToModuleProviderID int
}

// NewCreateModuleProviderRedirectCommand creates a new CreateModuleProviderRedirectCommand
func NewCreateModuleProviderRedirectCommand(redirectRepo ModuleProviderRedirectRepository) *CreateModuleProviderRedirectCommand {
	return &CreateModuleProviderRedirectCommand{
		redirectRepo: redirectRepo,
	}
}

// Execute creates a module provider redirect
func (c *CreateModuleProviderRedirectCommand) Execute(ctx context.Context, req CreateModuleProviderRedirectRequest) error {
	// Validate request
	if req.FromNamespace == "" {
		return fmt.Errorf("%w: from namespace cannot be empty", shared.ErrInvalidName)
	}
	if req.FromModule == "" {
		return fmt.Errorf("%w: from module cannot be empty", shared.ErrInvalidName)
	}
	if req.FromProvider == "" {
		return fmt.Errorf("%w: from provider cannot be empty", shared.ErrInvalidProvider)
	}
	if req.ToModuleProviderID <= 0 {
		return fmt.Errorf("%w: to module provider ID must be positive", shared.ErrInvalidInput)
	}

	// Check if redirect already exists
	existing, err := c.redirectRepo.GetByFrom(ctx, req.FromNamespace, req.FromModule, req.FromProvider)
	if err != nil && err != shared.ErrNotFound {
		return fmt.Errorf("failed to check existing redirect: %w", err)
	}
	if existing != nil {
		return fmt.Errorf("%w: redirect already exists", shared.ErrAlreadyExists)
	}

	// Create redirect
	return c.redirectRepo.Create(ctx, req)
}