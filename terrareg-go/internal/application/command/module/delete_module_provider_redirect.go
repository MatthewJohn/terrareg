package module

import (
	"context"
	"fmt"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/shared"
)

// DeleteModuleProviderRedirectRequest represents a request to delete a module provider redirect
type DeleteModuleProviderRedirectRequest struct {
	Namespace string
	Module    string
	Provider  string
}

// DeleteModuleProviderRedirectCommand deletes a module provider redirect
type DeleteModuleProviderRedirectCommand struct {
	redirectRepo ModuleProviderRedirectRepository
}

// NewDeleteModuleProviderRedirectCommand creates a new DeleteModuleProviderRedirectCommand
func NewDeleteModuleProviderRedirectCommand(redirectRepo ModuleProviderRedirectRepository) *DeleteModuleProviderRedirectCommand {
	return &DeleteModuleProviderRedirectCommand{
		redirectRepo: redirectRepo,
	}
}

// Execute deletes a module provider redirect
func (c *DeleteModuleProviderRedirectCommand) Execute(ctx context.Context, req DeleteModuleProviderRedirectRequest) error {
	// Validate request
	if req.Namespace == "" {
		return fmt.Errorf("%w: namespace cannot be empty", shared.ErrInvalidName)
	}
	if req.Module == "" {
		return fmt.Errorf("%w: module cannot be empty", shared.ErrInvalidName)
	}
	if req.Provider == "" {
		return fmt.Errorf("%w: provider cannot be empty", shared.ErrInvalidProvider)
	}

	// Check if redirect exists
	existing, err := c.redirectRepo.GetByFrom(ctx, req.Namespace, req.Module, req.Provider)
	if err != nil {
		if err == shared.ErrNotFound {
			return fmt.Errorf("%w: redirect not found", shared.ErrNotFound)
		}
		return fmt.Errorf("failed to check existing redirect: %w", err)
	}
	if existing == nil {
		return fmt.Errorf("%w: redirect not found", shared.ErrNotFound)
	}

	// Delete redirect
	return c.redirectRepo.Delete(ctx, req.Namespace, req.Module, req.Provider)
}