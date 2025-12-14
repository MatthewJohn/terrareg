package gpgkey

import (
	"context"
	"fmt"

	gpgkey "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/gpgkey/model"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/gpgkey/service"
)

// ManageGPGKeyCommand handles GPG key management operations
type ManageGPGKeyCommand struct {
	gpgKeyService *service.GPGKeyService
}

// NewManageGPGKeyCommand creates a new command for managing GPG keys
func NewManageGPGKeyCommand(gpgKeyService *service.GPGKeyService) *ManageGPGKeyCommand {
	return &ManageGPGKeyCommand{
		gpgKeyService: gpgKeyService,
	}
}

// CreateGPGKeyRequest represents a request to create a GPG key
type CreateGPGKeyRequest struct {
	Namespace      string  `json:"namespace"`
	ASCIILArmor    string  `json:"ascii-armor"`
	TrustSignature *string `json:"trust-signature,omitempty"`
	Source         *string `json:"source,omitempty"`
	SourceURL      *string `json:"source-url,omitempty"`
}

// CreateGPGKeyResponse represents the response for creating a GPG key
type CreateGPGKeyResponse struct {
	Type       string `json:"type"`
	ID         string `json:"id"`
	Attributes struct {
		Namespace      string  `json:"namespace"`
		ASCIILArmor    string  `json:"ascii-armor"`
		CreatedAt      string  `json:"created-at"`
		KeyID          string  `json:"key-id"`
		Source         string  `json:"source"`
		SourceURL      *string `json:"source-url"`
		TrustSignature string  `json:"trust-signature"`
		UpdatedAt      string  `json:"updated-at"`
	} `json:"attributes"`
}

// DeleteGPGKeyRequest represents a request to delete a GPG key
type DeleteGPGKeyRequest struct {
	Namespace string `json:"namespace"`
	KeyID     string `json:"key_id"`
}

// CreateGPGKey creates a new GPG key
func (c *ManageGPGKeyCommand) CreateGPGKey(ctx context.Context, req CreateGPGKeyRequest) (*CreateGPGKeyResponse, error) {
	// Convert request to service request
	serviceReq := service.CreateGPGKeyRequest{
		Namespace:      req.Namespace,
		ASCIILArmor:    req.ASCIILArmor,
		TrustSignature: req.TrustSignature,
		Source:         req.Source,
		SourceURL:      req.SourceURL,
	}

	// Create GPG key using service
	gpgKey, err := c.gpgKeyService.CreateGPGKey(ctx, serviceReq)
	if err != nil {
		return nil, fmt.Errorf("failed to create GPG key: %w", err)
	}

	// Convert domain model to response
	response := &CreateGPGKeyResponse{
		Type: "gpg-keys",
		ID:   gpgKey.KeyID(), // Use key_id as the ID for API responses
	}

	response.Attributes.Namespace = gpgKey.Namespace().Name()
	response.Attributes.ASCIILArmor = gpgKey.ASCIIArmor()
	response.Attributes.CreatedAt = gpgKey.CreatedAt().Format("2006-01-02T15:04:05Z")
	response.Attributes.KeyID = gpgKey.KeyID()
	response.Attributes.Source = gpgKey.Source()
	response.Attributes.SourceURL = gpgKey.SourceURL()
	response.Attributes.UpdatedAt = gpgKey.UpdatedAt().Format("2006-01-02T15:04:05Z")

	if trustSignature := gpgKey.TrustSignature(); trustSignature != nil {
		response.Attributes.TrustSignature = *trustSignature
	} else {
		response.Attributes.TrustSignature = ""
	}

	return response, nil
}

// DeleteGPGKey deletes a GPG key
func (c *ManageGPGKeyCommand) DeleteGPGKey(ctx context.Context, req DeleteGPGKeyRequest) error {
	err := c.gpgKeyService.DeleteGPGKey(ctx, req.Namespace, req.KeyID)
	if err != nil {
		if err == gpgkey.ErrGPGKeyNotFound {
			return fmt.Errorf("GPG key not found")
		}
		if err == gpgkey.ErrGPGKeyInUse {
			return fmt.Errorf("cannot delete GPG key that is in use")
		}
		return fmt.Errorf("failed to delete GPG key: %w", err)
	}

	return nil
}
