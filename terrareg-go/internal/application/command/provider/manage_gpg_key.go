package provider

import (
	"context"
	"fmt"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/provider"
	providerRepo "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/provider/repository"
	namespaceRepo "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/module/repository"
)

// ManageGPGKeyCommand handles managing GPG keys for providers
type ManageGPGKeyCommand struct {
	providerRepo  providerRepo.ProviderRepository
	namespaceRepo namespaceRepo.NamespaceRepository
}

// NewManageGPGKeyCommand creates a new manage GPG key command
func NewManageGPGKeyCommand(
	providerRepo providerRepo.ProviderRepository,
	namespaceRepo namespaceRepo.NamespaceRepository,
) *ManageGPGKeyCommand {
	return &ManageGPGKeyCommand{
		providerRepo:  providerRepo,
		namespaceRepo: namespaceRepo,
	}
}

// AddGPGKeyRequest represents the request to add a GPG key to a provider
type AddGPGKeyRequest struct {
	Namespace     string `json:"namespace"`
	ProviderName  string `json:"provider"`
	KeyText       string `json:"key_text"`
	AsciiArmor    string `json:"ascii_armor"`
	KeyID         string `json:"key_id"`
	TrustSignature string `json:"trust_signature,omitempty"`
}

// RemoveGPGKeyRequest represents the request to remove a GPG key from a provider
type RemoveGPGKeyRequest struct {
	Namespace    string `json:"namespace"`
	ProviderName string `json:"provider"`
	KeyID        string `json:"key_id"`
}

// ExecuteAdd adds a GPG key to a provider
func (c *ManageGPGKeyCommand) ExecuteAdd(ctx context.Context, req AddGPGKeyRequest) (*provider.Provider, error) {
	// Find provider
	providerEntity, err := c.providerRepo.FindByNamespaceAndName(ctx, req.Namespace, req.ProviderName)
	if err != nil {
		return nil, fmt.Errorf("provider not found: %w", err)
	}

	// Create GPG key value object
	gpgKey, err := provider.NewGPGKey(
		req.KeyText,
		req.AsciiArmor,
		req.KeyID,
		req.TrustSignature,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create GPG key: %w", err)
	}

	// Add GPG key to provider
	providerEntity.AddGPGKey(gpgKey)

	// Save provider
	if err := c.providerRepo.Save(ctx, providerEntity); err != nil {
		return nil, fmt.Errorf("failed to save provider: %w", err)
	}

	return providerEntity, nil
}

// ExecuteRemove removes a GPG key from a provider
func (c *ManageGPGKeyCommand) ExecuteRemove(ctx context.Context, req RemoveGPGKeyRequest) (*provider.Provider, error) {
	// Find provider
	providerEntity, err := c.providerRepo.FindByNamespaceAndName(ctx, req.Namespace, req.ProviderName)
	if err != nil {
		return nil, fmt.Errorf("provider not found: %w", err)
	}

	// Remove GPG key from provider
	if err := providerEntity.RemoveGPGKey(req.KeyID); err != nil {
		return nil, fmt.Errorf("failed to remove GPG key: %w", err)
	}

	// Save provider
	if err := c.providerRepo.Save(ctx, providerEntity); err != nil {
		return nil, fmt.Errorf("failed to save provider: %w", err)
	}

	return providerEntity, nil
}