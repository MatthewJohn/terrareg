package provider

import (
	"context"
	"fmt"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/provider"
	providerRepo "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/provider/repository"
	namespaceRepo "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/module/repository"
)

// PublishProviderVersionCommand handles publishing a provider version
type PublishProviderVersionCommand struct {
	providerRepo  providerRepo.ProviderRepository
	namespaceRepo namespaceRepo.NamespaceRepository
}

// NewPublishProviderVersionCommand creates a new publish provider version command
func NewPublishProviderVersionCommand(
	providerRepo providerRepo.ProviderRepository,
	namespaceRepo namespaceRepo.NamespaceRepository,
) *PublishProviderVersionCommand {
	return &PublishProviderVersionCommand{
		providerRepo:  providerRepo,
		namespaceRepo: namespaceRepo,
	}
}

// PublishProviderVersionRequest represents the request to publish a provider version
type PublishProviderVersionRequest struct {
	Namespace    string `json:"namespace"`
	ProviderName string `json:"provider"`
	Version      string `json:"version"`
	Protocol     string `json:"protocol,omitempty"`
	IsBeta       bool   `json:"is_beta,omitempty"`
}

// Execute publishes a new provider version
func (c *PublishProviderVersionCommand) Execute(ctx context.Context, req PublishProviderVersionRequest) (*provider.ProviderVersion, error) {
	// Find provider
	providerEntity, err := c.providerRepo.FindByNamespaceAndName(ctx, req.Namespace, req.ProviderName)
	if err != nil {
		return nil, fmt.Errorf("provider not found: %w", err)
	}

	// Create new provider version
	providerVersion := providerEntity.PublishVersion(req.Version, req.Protocol, req.IsBeta)

	// Save provider with new version
	if err := c.providerRepo.Save(ctx, providerEntity); err != nil {
		return nil, fmt.Errorf("failed to save provider: %w", err)
	}

	return providerVersion, nil
}