package provider

import (
	"context"
	"fmt"

	namespaceRepo "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/module/repository"
)

// GetNamespaceGPGKeysQuery handles retrieving GPG keys for a namespace
type GetNamespaceGPGKeysQuery struct {
	namespaceRepo namespaceRepo.NamespaceRepository
	// TODO: Add GPG key repository when it exists
}

// NewGetNamespaceGPGKeysQuery creates a new get namespace GPG keys query
func NewGetNamespaceGPGKeysQuery(namespaceRepo namespaceRepo.NamespaceRepository) *GetNamespaceGPGKeysQuery {
	return &GetNamespaceGPGKeysQuery{
		namespaceRepo: namespaceRepo,
	}
}

// Execute retrieves all GPG keys for a namespace
func (q *GetNamespaceGPGKeysQuery) Execute(ctx context.Context, namespace string) ([]GPGKeyResponse, error) {
	// Get namespace
	_, err := q.namespaceRepo.FindByName(ctx, namespace)
	if err != nil {
		return nil, fmt.Errorf("namespace not found: %w", err)
	}

	// TODO: Implement when GPG key repository is available
	// For now, return empty slice
	return []GPGKeyResponse{}, nil
}

// GPGKeyResponse represents a GPG key in the API response
type GPGKeyResponse struct {
	ID            string `json:"id"`
	Namespace     string `json:"namespace"`
	KeyID         string `json:"key_id"`
	ASCIIArmor    string `json:"ascii_armor"`
	TrustSignature string `json:"trust_signature,omitempty"`
	CreatedAt     string `json:"created_at"`
	UpdatedAt     string `json:"updated_at"`
}