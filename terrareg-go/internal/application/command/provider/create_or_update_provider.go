package provider

import (
	"context"
	"fmt"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/provider"
	providerRepo "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/provider/repository"
	namespaceRepo "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/module/repository"
)

// CreateOrUpdateProviderCommand handles creating or updating a provider
type CreateOrUpdateProviderCommand struct {
	providerRepo  providerRepo.ProviderRepository
	namespaceRepo namespaceRepo.NamespaceRepository
}

// NewCreateOrUpdateProviderCommand creates a new create/update provider command
func NewCreateOrUpdateProviderCommand(
	providerRepo providerRepo.ProviderRepository,
	namespaceRepo namespaceRepo.NamespaceRepository,
) *CreateOrUpdateProviderCommand {
	return &CreateOrUpdateProviderCommand{
		providerRepo:  providerRepo,
		namespaceRepo: namespaceRepo,
	}
}

// CreateOrUpdateProviderRequest represents the request to create/update a provider
type CreateOrUpdateProviderRequest struct {
	Namespace        string  `json:"namespace"`
	Name             string  `json:"name"`
	Description      *string `json:"description,omitempty"`
	Tier             string  `json:"tier"`
	Source           *string `json:"source,omitempty"`
	Alias            *string `json:"alias,omitempty"`
	GitProviderID    *int    `json:"git_provider_id,omitempty"`
	RepoCloneURL     *string `json:"repo_clone_url,omitempty"`
	GitTagFormat     *string `json:"git_tag_format,omitempty"`
	GitPath          *string `json:"git_path,omitempty"`
}

// Execute creates or updates a provider
func (c *CreateOrUpdateProviderCommand) Execute(ctx context.Context, req CreateOrUpdateProviderRequest) (*provider.Provider, error) {
	// Validate tier
	if req.Tier == "" {
		req.Tier = "community" // Default tier
	}

	// Find namespace
	namespace, err := c.namespaceRepo.FindByName(ctx, req.Namespace)
	if err != nil {
		return nil, fmt.Errorf("namespace not found: %w", err)
	}

	// Check if provider already exists
	existingProvider, err := c.providerRepo.FindByNamespaceAndName(ctx, req.Namespace, req.Name)
	if err == nil && existingProvider != nil {
		// Update existing provider
		existingProvider.UpdateDetails(req.Description, req.Tier, req.Source, req.Alias)
		existingProvider.UpdateGitConfig(req.GitProviderID, req.RepoCloneURL, req.GitTagFormat, req.GitPath)

		if err := c.providerRepo.Save(ctx, existingProvider); err != nil {
			return nil, fmt.Errorf("failed to update provider: %w", err)
		}

		return existingProvider, nil
	}

	// Create new provider
	newProvider := provider.NewProvider(
		0, // ID - will be set by repository
		namespace.ID(),
		req.Name,
		req.Description,
		req.Tier,
		nil, // CategoryID
		nil, // RepositoryID
		nil, // LatestVersionID
		false, // UseProviderSourceAuth
	)

	if err := c.providerRepo.Save(ctx, newProvider); err != nil {
		return nil, fmt.Errorf("failed to save provider: %w", err)
	}

	return newProvider, nil
}