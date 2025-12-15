package service

import (
	"context"
	"fmt"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/provider"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/provider/repository"
)

// ProviderGPGService handles GPG key operations for providers
type ProviderGPGService struct {
	providerRepo repository.ProviderRepository
}

// NewProviderGPGService creates a new provider GPG service
func NewProviderGPGService(providerRepo repository.ProviderRepository) *ProviderGPGService {
	return &ProviderGPGService{
		providerRepo: providerRepo,
	}
}

// AddGPGKeyRequest represents a request to add a GPG key
type AddGPGKeyRequest struct {
	ProviderID     int
	KeyText        string
	AsciiArmor     string
	KeyID          string
	TrustSignature *string
}

// AddGPGKey adds a new GPG key to a provider
func (s *ProviderGPGService) AddGPGKey(ctx context.Context, req AddGPGKeyRequest) (*provider.GPGKey, error) {
	// Find provider
	providerEntity, err := s.providerRepo.FindByID(ctx, req.ProviderID)
	if err != nil {
		return nil, fmt.Errorf("failed to find provider: %w", err)
	}
	if providerEntity == nil {
		return nil, fmt.Errorf("provider not found")
	}

	// Create GPG key
	gpgKey, err := provider.NewGPGKey(req.KeyText, req.AsciiArmor, req.KeyID, "")
	if err != nil {
		return nil, fmt.Errorf("failed to create GPG key: %w", err)
	}

	// Set trust signature if provided
	if req.TrustSignature != nil {
		gpgKey.SetTrustSignature(req.TrustSignature)
	}

	// Add to provider
	if err := providerEntity.AddGPGKey(gpgKey); err != nil {
		return nil, fmt.Errorf("failed to add GPG key to provider: %w", err)
	}

	// Save to repository
	if err := s.providerRepo.SaveGPGKey(ctx, gpgKey); err != nil {
		return nil, fmt.Errorf("failed to save GPG key: %w", err)
	}

	return gpgKey, nil
}

// ImportGPGKeyRequest represents a request to import a GPG key from text
type ImportGPGKeyRequest struct {
	ProviderID int
	KeyText    string
}

// ImportGPGKey imports a GPG key from text and extracts metadata
func (s *ProviderGPGService) ImportGPGKey(ctx context.Context, req ImportGPGKeyRequest) (*provider.GPGKey, error) {
	// Find provider
	providerEntity, err := s.providerRepo.FindByID(ctx, req.ProviderID)
	if err != nil {
		return nil, fmt.Errorf("failed to find provider: %w", err)
	}
	if providerEntity == nil {
		return nil, fmt.Errorf("provider not found")
	}

	// Extract key ID from text if not provided
	keyID, err := provider.ExtractKeyIDFromText(req.KeyText)
	if err != nil {
		return nil, fmt.Errorf("failed to extract key ID from text: %w", err)
	}

	// Create GPG key with extracted information
	gpgKey, err := provider.NewGPGKey(req.KeyText, "", keyID, "")
	if err != nil {
		return nil, fmt.Errorf("failed to create GPG key: %w", err)
	}

	// Add to provider
	if err := providerEntity.AddGPGKey(gpgKey); err != nil {
		return nil, fmt.Errorf("failed to add GPG key to provider: %w", err)
	}

	// Save to repository
	if err := s.providerRepo.SaveGPGKey(ctx, gpgKey); err != nil {
		return nil, fmt.Errorf("failed to save GPG key: %w", err)
	}

	return gpgKey, nil
}

// RemoveGPGKey removes a GPG key from a provider
func (s *ProviderGPGService) RemoveGPGKey(ctx context.Context, providerID, keyID int) error {
	// Find provider
	providerEntity, err := s.providerRepo.FindByID(ctx, providerID)
	if err != nil {
		return fmt.Errorf("failed to find provider: %w", err)
	}
	if providerEntity == nil {
		return fmt.Errorf("provider not found")
	}

	// Check if key is being used by any version
	versions, err := s.providerRepo.FindVersionsByProvider(ctx, providerID)
	if err != nil {
		return fmt.Errorf("failed to check versions: %w", err)
	}

	for _, version := range versions {
		if version.GPGKeyID() == keyID {
			return fmt.Errorf("cannot remove GPG key that is in use by version %s", version.Version())
		}
	}

	// Remove from provider
	if err := providerEntity.RemoveGPGKey(keyID); err != nil {
		return fmt.Errorf("failed to remove GPG key from provider: %w", err)
	}

	// Delete from repository
	if err := s.providerRepo.DeleteGPGKey(ctx, keyID); err != nil {
		return fmt.Errorf("failed to delete GPG key: %w", err)
	}

	return nil
}

// GetGPGKeys retrieves all GPG keys for a provider
func (s *ProviderGPGService) GetGPGKeys(ctx context.Context, providerID int) ([]*provider.GPGKey, error) {
	gpgKeys, err := s.providerRepo.FindGPGKeysByProvider(ctx, providerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get GPG keys: %w", err)
	}
	return gpgKeys, nil
}

// GetGPGKey retrieves a specific GPG key
func (s *ProviderGPGService) GetGPGKey(ctx context.Context, keyID string) (*provider.GPGKey, error) {
	gpgKey, err := s.providerRepo.FindGPGKeyByKeyID(ctx, keyID)
	if err != nil {
		return nil, fmt.Errorf("failed to get GPG key: %w", err)
	}
	return gpgKey, nil
}

// SignVersion signs a provider version with a GPG key
func (s *ProviderGPGService) SignVersion(ctx context.Context, providerID, versionID, keyID int) error {
	// Find provider
	providerEntity, err := s.providerRepo.FindByID(ctx, providerID)
	if err != nil {
		return fmt.Errorf("failed to find provider: %w", err)
	}
	if providerEntity == nil {
		return fmt.Errorf("provider not found")
	}

	// Find version
	version, err := s.providerRepo.FindVersionByID(ctx, versionID)
	if err != nil {
		return fmt.Errorf("failed to find version: %w", err)
	}
	if version == nil {
		return fmt.Errorf("version not found")
	}

	// Check if version belongs to provider
	if version.ProviderID() != providerID {
		return fmt.Errorf("version does not belong to provider")
	}

	// Find GPG key
	gpgKey := providerEntity.FindGPGKeyByID(keyID)
	if gpgKey == nil {
		return fmt.Errorf("GPG key not found")
	}

	// Set GPG key for version
	version.SetGPGKeyID(keyID)

	// Save version
	if err := s.providerRepo.SaveVersion(ctx, version); err != nil {
		return fmt.Errorf("failed to save version: %w", err)
	}

	return nil
}

// VerifySignature verifies a GPG signature for a binary
func (s *ProviderGPGService) VerifySignature(ctx context.Context, binaryPath, signaturePath, keyID string) (bool, error) {
	// Find GPG key
	gpgKey, err := s.providerRepo.FindGPGKeyByKeyID(ctx, keyID)
	if err != nil {
		return false, fmt.Errorf("failed to find GPG key: %w", err)
	}
	if gpgKey == nil {
		return false, fmt.Errorf("GPG key not found")
	}

	// Placeholder implementation
	// In a real implementation, this would:
	// 1. Use golang.org/x/crypto/openpgp to verify the signature
	// 2. Check that the signature was made with the provided key
	// 3. Return true if verification succeeds

	fmt.Printf("Verifying signature for %s with key %s\n", binaryPath, keyID)
	return true, nil
}

// GenerateSignature generates a GPG signature for a binary
func (s *ProviderGPGService) GenerateSignature(ctx context.Context, binaryPath, keyID string) (string, error) {
	// Find GPG key
	gpgKey, err := s.providerRepo.FindGPGKeyByKeyID(ctx, keyID)
	if err != nil {
		return "", fmt.Errorf("failed to find GPG key: %w", err)
	}
	if gpgKey == nil {
		return "", fmt.Errorf("GPG key not found")
	}

	// Placeholder implementation
	// In a real implementation, this would:
	// 1. Use golang.org/x/crypto/openpgp to generate a signature
	// 2. Sign the binary with the provided key
	// 3. Return the signature as ASCII armored text

	signature := fmt.Sprintf("-----BEGIN PGP SIGNATURE-----\nGenerated signature for %s with key %s\n-----END PGP SIGNATURE-----", binaryPath, keyID)
	return signature, nil
}

// TrustGPGKey adds or updates a trust signature for a GPG key
func (s *ProviderGPGService) TrustGPGKey(ctx context.Context, providerID, keyID int, trustSignature string) error {
	// Find provider
	providerEntity, err := s.providerRepo.FindByID(ctx, providerID)
	if err != nil {
		return fmt.Errorf("failed to find provider: %w", err)
	}
	if providerEntity == nil {
		return fmt.Errorf("provider not found")
	}

	// Find GPG key
	gpgKey := providerEntity.FindGPGKeyByID(keyID)
	if gpgKey == nil {
		return fmt.Errorf("GPG key not found")
	}

	// Update trust signature
	gpgKey.SetTrustSignature(&trustSignature)

	// Save GPG key
	if err := s.providerRepo.SaveGPGKey(ctx, gpgKey); err != nil {
		return fmt.Errorf("failed to save GPG key: %w", err)
	}

	return nil
}
