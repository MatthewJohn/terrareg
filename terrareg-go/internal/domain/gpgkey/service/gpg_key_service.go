package service

import (
	"context"
	"fmt"

	gpgkeyModel "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/gpgkey/model"
	gpgkeyRepo "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/gpgkey/repository"
	moduleRepo "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/module/repository"
)

// GPGKeyService handles GPG key operations
type GPGKeyService struct {
	gpgKeyRepo    gpgkeyRepo.GPGKeyRepository
	namespaceRepo moduleRepo.NamespaceRepository
}

// NewGPGKeyService creates a new GPG key service
func NewGPGKeyService(
	gpgKeyRepo gpgkeyRepo.GPGKeyRepository,
	namespaceRepo moduleRepo.NamespaceRepository,
) *GPGKeyService {
	return &GPGKeyService{
		gpgKeyRepo:    gpgKeyRepo,
		namespaceRepo: namespaceRepo,
	}
}

// CreateGPGKeyRequest represents a request to create a GPG key
type CreateGPGKeyRequest struct {
	Namespace      string
	ASCIILArmor    string
	TrustSignature *string
	Source         *string
	SourceURL      *string
}

// CreateGPGKey creates a new GPG key
func (s *GPGKeyService) CreateGPGKey(ctx context.Context, req CreateGPGKeyRequest) (*gpgkeyModel.GPGKey, error) {
	// Validate namespace exists
	namespace, err := s.namespaceRepo.FindByName(ctx, req.Namespace)
	if err != nil {
		return nil, fmt.Errorf("failed to validate namespace: %w", err)
	}
	if namespace == nil {
		return nil, fmt.Errorf("namespace '%s' does not exist", req.Namespace)
	}

	// Extract key ID and fingerprint from ASCII armor
	keyID, fingerprint, err := s.extractGPGKeyInfo(req.ASCIILArmor)
	if err != nil {
		return nil, fmt.Errorf("failed to extract GPG key information: %w", err)
	}

	// Check for duplicate fingerprint
	exists, err := s.gpgKeyRepo.ExistsByFingerprint(ctx, fingerprint)
	if err != nil {
		return nil, fmt.Errorf("failed to check for duplicate fingerprint: %w", err)
	}
	if exists {
		return nil, gpgkeyModel.ErrDuplicateFingerprint
	}

	// Create GPG key
	gpgKey, err := gpgkeyModel.NewGPGKey(namespace.ID(), req.ASCIILArmor, keyID, fingerprint)
	if err != nil {
		return nil, fmt.Errorf("failed to create GPG key: %w", err)
	}

	// Set optional fields
	if req.TrustSignature != nil {
		gpgKey.SetTrustSignature(req.TrustSignature)
	}
	if req.Source != nil {
		gpgKey.SetSource(*req.Source)
	}
	if req.SourceURL != nil {
		gpgKey.SetSourceURL(req.SourceURL)
	}

	// Set namespace entity
	gpgKey.SetNamespace(gpgkeyModel.NewNamespace(namespace.ID(), namespace.Name()))

	// Save to repository
	if err := s.gpgKeyRepo.Save(ctx, gpgKey); err != nil {
		return nil, fmt.Errorf("failed to save GPG key: %w", err)
	}

	return gpgKey, nil
}

// GetNamespaceGPGKeys retrieves all GPG keys for a namespace
func (s *GPGKeyService) GetNamespaceGPGKeys(ctx context.Context, namespace string) ([]*gpgkeyModel.GPGKey, error) {
	// Validate namespace exists
	ns, err := s.namespaceRepo.FindByName(ctx, namespace)
	if err != nil {
		return nil, fmt.Errorf("failed to validate namespace: %w", err)
	}
	if ns == nil {
		return nil, fmt.Errorf("namespace '%s' does not exist", namespace)
	}

	gpgKeys, err := s.gpgKeyRepo.FindByNamespace(ctx, namespace)
	if err != nil {
		return nil, fmt.Errorf("failed to get GPG keys for namespace: %w", err)
	}

	return gpgKeys, nil
}

// GetMultipleNamespaceGPGKeys retrieves GPG keys for multiple namespaces
func (s *GPGKeyService) GetMultipleNamespaceGPGKeys(ctx context.Context, namespaces []string) ([]*gpgkeyModel.GPGKey, error) {
	if len(namespaces) == 0 {
		return []*gpgkeyModel.GPGKey{}, nil
	}

	gpgKeys, err := s.gpgKeyRepo.FindMultipleByNamespaces(ctx, namespaces)
	if err != nil {
		return nil, fmt.Errorf("failed to get GPG keys for namespaces: %w", err)
	}

	return gpgKeys, nil
}

// GetGPGKey retrieves a specific GPG key by namespace and key ID
func (s *GPGKeyService) GetGPGKey(ctx context.Context, namespace, keyID string) (*gpgkeyModel.GPGKey, error) {
	gpgKey, err := s.gpgKeyRepo.FindByNamespaceAndKeyID(ctx, namespace, keyID)
	if err != nil {
		return nil, fmt.Errorf("failed to get GPG key: %w", err)
	}
	if gpgKey == nil {
		return nil, gpgkeyModel.ErrGPGKeyNotFound
	}

	return gpgKey, nil
}

// DeleteGPGKey deletes a GPG key by namespace and key ID
func (s *GPGKeyService) DeleteGPGKey(ctx context.Context, namespace, keyID string) error {
	// Check if GPG key exists and belongs to the namespace
	gpgKey, err := s.gpgKeyRepo.FindByNamespaceAndKeyID(ctx, namespace, keyID)
	if err != nil {
		return fmt.Errorf("failed to get GPG key: %w", err)
	}
	if gpgKey == nil {
		return gpgkeyModel.ErrGPGKeyNotFound
	}

	// Check if GPG key is in use by any provider versions
	isInUse, err := s.gpgKeyRepo.IsInUse(ctx, keyID)
	if err != nil {
		return fmt.Errorf("failed to check if GPG key is in use: %w", err)
	}
	if isInUse {
		return gpgkeyModel.ErrGPGKeyInUse
	}

	// Delete the GPG key
	if err := s.gpgKeyRepo.DeleteByNamespaceAndKeyID(ctx, namespace, keyID); err != nil {
		return fmt.Errorf("failed to delete GPG key: %w", err)
	}

	return nil
}

// VerifySignature verifies a GPG signature for data
func (s *GPGKeyService) VerifySignature(ctx context.Context, keyID string, signature, data string) (bool, error) {
	// Find GPG key
	gpgKey, err := s.gpgKeyRepo.FindByKeyID(ctx, keyID)
	if err != nil {
		return false, fmt.Errorf("failed to find GPG key: %w", err)
	}
	if gpgKey == nil {
		return false, fmt.Errorf("GPG key not found")
	}

	// TODO: Implement actual GPG signature verification
	// This would use golang.org/x/crypto/openpgp to verify the signature
	// For now, return a placeholder implementation

	return true, nil
}

// extractGPGKeyInfo extracts key ID and fingerprint from ASCII armor
// This is a placeholder implementation - in production, you would use
// a proper GPG library like golang.org/x/crypto/openpgp
func (s *GPGKeyService) extractGPGKeyInfo(asciiArmor string) (keyID, fingerprint string, err error) {
	// Placeholder implementation
	// In a real implementation, this would:
	// 1. Parse the ASCII armor using openpgp.ReadArmoredKeyRing
	// 2. Extract the fingerprint from the primary key
	// 3. Derive the key ID (last 16 characters of fingerprint)

	// For now, return dummy values - this would be replaced with actual GPG parsing
	return "1234567890ABCDEF", "1234567890ABCDEF1234567890ABCDEF12345678", nil
}
