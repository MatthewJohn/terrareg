package provider_source

import (
	"context"
	"fmt"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/module/repository"
	providerRepo "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/provider/repository"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/provider_source/service"
	types "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/shared/types"
)

// PublishProviderCommand publishes a provider from a repository
// Python reference: github_repository_publish_provider.py
type PublishProviderCommand struct {
	providerSourceFactory *service.ProviderSourceFactory
	providerRepo          providerRepo.ProviderRepository
	namespaceRepo         repository.NamespaceRepository
	categoryRepo          providerRepo.ProviderCategoryRepository
}

// NewPublishProviderCommand creates a new PublishProviderCommand
func NewPublishProviderCommand(
	providerSourceFactory *service.ProviderSourceFactory,
	providerRepo providerRepo.ProviderRepository,
	namespaceRepo repository.NamespaceRepository,
	categoryRepo providerRepo.ProviderCategoryRepository,
) (*PublishProviderCommand, error) {
	if providerSourceFactory == nil {
		return nil, fmt.Errorf("providerSourceFactory cannot be nil")
	}
	if providerRepo == nil {
		return nil, fmt.Errorf("providerRepo cannot be nil")
	}
	if namespaceRepo == nil {
		return nil, fmt.Errorf("namespaceRepo cannot be nil")
	}
	if categoryRepo == nil {
		return nil, fmt.Errorf("categoryRepo cannot be nil")
	}

	return &PublishProviderCommand{
		providerSourceFactory: providerSourceFactory,
		providerRepo:          providerRepo,
		namespaceRepo:         namespaceRepo,
		categoryRepo:          categoryRepo,
	}, nil
}

// PublishProviderRequest represents a request to publish a provider from a repository
type PublishProviderRequest struct {
	// ProviderSource is the name of the provider source (e.g., "github")
	ProviderSource string
	// RepoID is the ID of the repository to publish from
	RepoID int
	// CategoryID is the ID of the provider category
	CategoryID int
	// Namespace is the namespace to publish the provider under
	Namespace string
}

// Execute publishes a provider from a repository
// Python reference: github_repository_publish_provider.py::publish_provider_from_repository
func (c *PublishProviderCommand) Execute(ctx context.Context, req PublishProviderRequest) (*service.PublishProviderResult, error) {
	// Validate inputs
	if req.ProviderSource == "" {
		return nil, fmt.Errorf("provider_source is required")
	}
	if req.RepoID <= 0 {
		return nil, fmt.Errorf("repo_id is required")
	}
	if req.CategoryID <= 0 {
		return nil, fmt.Errorf("category_id is required")
	}
	if req.Namespace == "" {
		return nil, fmt.Errorf("namespace is required")
	}

	// Verify namespace exists
	namespace, err := c.namespaceRepo.FindByName(ctx, types.NamespaceName(req.Namespace))
	if err != nil {
		return nil, fmt.Errorf("failed to find namespace: %w", err)
	}
	if namespace == nil {
		return nil, fmt.Errorf("namespace not found: %s", req.Namespace)
	}

	// Verify category exists
	category, err := c.categoryRepo.FindByID(ctx, req.CategoryID)
	if err != nil {
		return nil, fmt.Errorf("failed to find category: %w", err)
	}
	if category == nil {
		return nil, fmt.Errorf("category not found: %d", req.CategoryID)
	}

	// Get provider source instance
	providerSource, err := c.providerSourceFactory.GetProviderSourceByName(ctx, req.ProviderSource)
	if err != nil {
		return nil, fmt.Errorf("failed to get provider source: %w", err)
	}
	if providerSource == nil {
		return nil, fmt.Errorf("provider source not found: %s", req.ProviderSource)
	}

	// Publish provider from repository
	result, err := providerSource.PublishProviderFromRepository(ctx, req.RepoID, req.CategoryID, req.Namespace)
	if err != nil {
		return nil, fmt.Errorf("failed to publish provider from repository: %w", err)
	}

	return result, nil
}
