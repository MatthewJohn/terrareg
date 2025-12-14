package gpgkey

import (
	"context"
	"fmt"
	"time"

	gpgkeyModel "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/gpgkey/model"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/gpgkey/service"
)

// GetNamespaceGPGKeysQuery retrieves all GPG keys for a namespace
type GetNamespaceGPGKeysQuery struct {
	gpgKeyService *service.GPGKeyService
}

// NewGetNamespaceGPGKeysQuery creates a new query for retrieving namespace GPG keys
func NewGetNamespaceGPGKeysQuery(gpgKeyService *service.GPGKeyService) *GetNamespaceGPGKeysQuery {
	return &GetNamespaceGPGKeysQuery{
		gpgKeyService: gpgKeyService,
	}
}

// GetMultipleNamespaceGPGKeysQuery retrieves GPG keys for multiple namespaces
type GetMultipleNamespaceGPGKeysQuery struct {
	gpgKeyService *service.GPGKeyService
}

// NewGetMultipleNamespaceGPGKeysQuery creates a new query for retrieving GPG keys from multiple namespaces
func NewGetMultipleNamespaceGPGKeysQuery(gpgKeyService *service.GPGKeyService) *GetMultipleNamespaceGPGKeysQuery {
	return &GetMultipleNamespaceGPGKeysQuery{
		gpgKeyService: gpgKeyService,
	}
}

// GetGPGKeyQuery retrieves a specific GPG key
type GetGPGKeyQuery struct {
	gpgKeyService *service.GPGKeyService
}

// NewGetGPGKeyQuery creates a new query for retrieving a specific GPG key
func NewGetGPGKeyQuery(gpgKeyService *service.GPGKeyService) *GetGPGKeyQuery {
	return &GetGPGKeyQuery{
		gpgKeyService: gpgKeyService,
	}
}

// Execute executes the query to retrieve namespace GPG keys
func (q *GetNamespaceGPGKeysQuery) Execute(ctx context.Context, namespace string) ([]GPGKeyResponse, error) {
	gpgKeys, err := q.gpgKeyService.GetNamespaceGPGKeys(ctx, namespace)
	if err != nil {
		return nil, fmt.Errorf("failed to get namespace GPG keys: %w", err)
	}

	return gpgKeysToResponse(gpgKeys), nil
}

// Execute executes the query to retrieve GPG keys from multiple namespaces
func (q *GetMultipleNamespaceGPGKeysQuery) Execute(ctx context.Context, namespaces []string) ([]GPGKeyResponse, error) {
	gpgKeys, err := q.gpgKeyService.GetMultipleNamespaceGPGKeys(ctx, namespaces)
	if err != nil {
		return nil, fmt.Errorf("failed to get multiple namespace GPG keys: %w", err)
	}

	return gpgKeysToResponse(gpgKeys), nil
}

// Execute executes the query to retrieve a specific GPG key
func (q *GetGPGKeyQuery) Execute(ctx context.Context, namespace, keyID string) (*GPGKeyResponse, error) {
	gpgKey, err := q.gpgKeyService.GetGPGKey(ctx, namespace, keyID)
	if err != nil {
		if err == gpgkeyModel.ErrGPGKeyNotFound {
			return nil, nil // Not found is not an error
		}
		return nil, fmt.Errorf("failed to get GPG key: %w", err)
	}

	return gpgKeyToResponse(gpgKey), nil
}

// GPGKeyResponse represents the API response format for GPG keys
type GPGKeyResponse struct {
	Type       string `json:"type"`
	ID         string `json:"id"`
	Attributes struct {
		Namespace      string    `json:"namespace"`
		ASCIILArmor    string    `json:"ascii-armor"`
		CreatedAt      time.Time `json:"created-at"`
		KeyID          string    `json:"key-id"`
		Source         string    `json:"source"`
		SourceURL      *string   `json:"source-url"`
		TrustSignature string    `json:"trust-signature"`
		UpdatedAt      time.Time `json:"updated-at"`
	} `json:"attributes"`
}

// Helper functions to convert domain models to response DTOs

func gpgKeysToResponse(gpgKeys []*gpgkeyModel.GPGKey) []GPGKeyResponse {
	responses := make([]GPGKeyResponse, len(gpgKeys))
	for i, gpgKey := range gpgKeys {
		responses[i] = *gpgKeyToResponse(gpgKey)
	}
	return responses
}

func gpgKeyToResponse(gpgKey *gpgkeyModel.GPGKey) *GPGKeyResponse {
	response := &GPGKeyResponse{
		Type: "gpg-keys",
		ID:   gpgKey.KeyID(), // Use key_id as the ID for API responses
	}

	response.Attributes.Namespace = gpgKey.Namespace().Name()
	response.Attributes.ASCIILArmor = gpgKey.ASCIIArmor()
	response.Attributes.CreatedAt = gpgKey.CreatedAt()
	response.Attributes.KeyID = gpgKey.KeyID()
	response.Attributes.Source = gpgKey.Source()
	response.Attributes.SourceURL = gpgKey.SourceURL()
	response.Attributes.UpdatedAt = gpgKey.UpdatedAt()

	if trustSignature := gpgKey.TrustSignature(); trustSignature != nil {
		response.Attributes.TrustSignature = *trustSignature
	} else {
		response.Attributes.TrustSignature = ""
	}

	return response
}
