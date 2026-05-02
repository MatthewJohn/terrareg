package provider_source

import (
	"context"
	"fmt"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/module/repository"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/provider_source/service"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/shared"
	types "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/shared/types"
)

// RefreshNamespaceCommand refreshes repositories for a namespace from a provider source
type RefreshNamespaceCommand struct {
	providerSourceFactory *service.ProviderSourceFactory
	namespaceRepo         repository.NamespaceRepository
}

// NewRefreshNamespaceCommand creates a new RefreshNamespaceCommand
func NewRefreshNamespaceCommand(
	providerSourceFactory *service.ProviderSourceFactory,
	namespaceRepo repository.NamespaceRepository,
) (*RefreshNamespaceCommand, error) {
	if providerSourceFactory == nil {
		return nil, fmt.Errorf("providerSourceFactory cannot be nil")
	}
	if namespaceRepo == nil {
		return nil, fmt.Errorf("namespaceRepo cannot be nil")
	}

	return &RefreshNamespaceCommand{
		providerSourceFactory: providerSourceFactory,
		namespaceRepo:         namespaceRepo,
	}, nil
}

// RefreshNamespaceRequest represents a request to refresh namespace repositories
type RefreshNamespaceRequest struct {
	// ProviderSource is the name of the provider source (e.g., "github")
	ProviderSource string
	// Namespace is the namespace to refresh
	Namespace string
}

// Execute refreshes repositories for a namespace from the provider source
func (c *RefreshNamespaceCommand) Execute(ctx context.Context, req RefreshNamespaceRequest) error {
	// Validate inputs
	if req.ProviderSource == "" {
		return fmt.Errorf("provider_source is required")
	}
	if req.Namespace == "" {
		return fmt.Errorf("namespace is required")
	}

	// Verify namespace exists
	namespace, err := c.namespaceRepo.FindByName(ctx, types.NamespaceName(req.Namespace))
	if err != nil {
		return fmt.Errorf("failed to find namespace: %w", err)
	}
	if namespace == nil {
		return fmt.Errorf("namespace not found: %s", req.Namespace)
	}

	// Get provider source instance
	providerSource, err := c.providerSourceFactory.GetProviderSourceByName(ctx, req.ProviderSource)
	if err != nil {
		return fmt.Errorf("failed to get provider source: %w", err)
	}
	if providerSource == nil {
		return shared.ErrNotFound
	}

	// Refresh namespace repositories
	if err := providerSource.RefreshNamespaceRepositories(ctx, req.Namespace); err != nil {
		return fmt.Errorf("failed to refresh namespace repositories: %w", err)
	}

	return nil
}
