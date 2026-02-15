package service

import (
	"bytes"
	"context"
	"encoding/hex"
	"fmt"

	"golang.org/x/crypto/openpgp"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/provider"
	providermodel "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/provider_source/model"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/persistence/sqldb"
)

// ProviderExtractionGPGService handles GPG key verification for provider extraction
// Python reference: provider_extractor.py::ProviderExtractor.obtain_gpg_key
type ProviderExtractionGPGService struct {
	// Dependencies would be injected here
}

// NewProviderExtractionGPGService creates a new GPG service for provider extraction
func NewProviderExtractionGPGService() *ProviderExtractionGPGService {
	return &ProviderExtractionGPGService{}
}

// ObtainGPGKey obtains the GPG key for a provider release by verifying the signature
// Python reference: provider_extractor.py::ProviderExtractor.obtain_gpg_key
func (s *ProviderExtractionGPGService) ObtainGPGKey(
	ctx context.Context,
	providerEntity *provider.Provider,
	gpgKeys []*provider.GPGKey,
	releaseMetadata *providermodel.RepositoryReleaseMetadata,
	downloadArtifactFunc func(ctx context.Context, artifactName string) ([]byte, error),
) (*provider.GPGKey, error) {
	// Generate artifact file names
	version := releaseMetadata.Version()
	if version == nil {
		return nil, provider.ErrUnableToObtainSource
	}

	shasumsFileName := s.generateArtifactName(providerEntity, *version, "SHA256SUMS")
	shasumsSigFileName := s.generateArtifactName(providerEntity, *version, "SHA256SUMS.sig")

	// Download SHA256SUMS file
	shasums, err := downloadArtifactFunc(ctx, shasumsFileName)
	if err != nil {
		return nil, fmt.Errorf("%w: could not obtain shasums file: %s", provider.ErrMissingChecksumArtifact, err)
	}

	// Download SHA256SUMS.sig file
	shasumsSignature, err := downloadArtifactFunc(ctx, shasumsSigFileName)
	if err != nil {
		return nil, fmt.Errorf("%w: could not obtain shasums signature file: %s", provider.ErrMissingSignatureArtifact, err)
	}

	// Try to verify signature with all GPG keys
	for _, gpgKey := range gpgKeys {
		if s.verifyDataSignature(gpgKey, shasumsSignature, shasums) {
			return gpgKey, nil
		}
	}

	return nil, nil
}

// verifyDataSignature verifies if a GPG key matches a signature
// Python reference: models.py::GpgKey.verify_data_signature
func (s *ProviderExtractionGPGService) verifyDataSignature(
	gpgKey *provider.GPGKey,
	signature []byte,
	data []byte,
) bool {
	// Parse the ASCII armor key
	keyReader := bytes.NewReader([]byte(gpgKey.AsciiArmor()))
	entityList, err := openpgp.ReadArmoredKeyRing(keyReader)
	if err != nil {
		return false
	}

	if len(entityList) == 0 {
		return false
	}

	// Create readers for verification
	dataReader := bytes.NewReader(data)
	sigReader := bytes.NewReader(signature)

	// Try armored signature first (common format)
	_, err = openpgp.CheckArmoredDetachedSignature(entityList, dataReader, sigReader)
	if err == nil {
		return true
	}

	// Reset readers for binary signature attempt
	dataReader = bytes.NewReader(data)
	sigReader = bytes.NewReader(signature)

	// Try binary signature
	_, err = openpgp.CheckDetachedSignature(entityList, dataReader, sigReader)
	return err == nil
}

// ExtractKeyIDFromASCIIArmor extracts the key ID from ASCII armor
// Returns the 16-character hex key ID
// Python reference: models.py::GpgKey._get_gpg_object
func (s *ProviderExtractionGPGService) ExtractKeyIDFromASCIIArmor(asciiArmor string) (string, error) {
	keyReader := bytes.NewReader([]byte(asciiArmor))
	entityList, err := openpgp.ReadArmoredKeyRing(keyReader)
	if err != nil {
		return "", fmt.Errorf("failed to parse ASCII armor: %w", err)
	}

	if len(entityList) == 0 {
		return "", fmt.Errorf("no entities found in key data")
	}

	// Get the primary key
	entity := entityList[0]
	if entity.PrimaryKey == nil {
		return "", fmt.Errorf("no primary key found in entity")
	}

	// Key ID is the lower 8 bytes of the fingerprint
	// Convert to hex string (16 characters, uppercase to match Python)
	keyID := fmt.Sprintf("%016X", entity.PrimaryKey.KeyId)
	return keyID, nil
}

// ExtractFingerprintFromASCIIArmor extracts the full fingerprint from ASCII armor
// Returns the 40-character hex fingerprint
func (s *ProviderExtractionGPGService) ExtractFingerprintFromASCIIArmor(asciiArmor string) (string, error) {
	keyReader := bytes.NewReader([]byte(asciiArmor))
	entityList, err := openpgp.ReadArmoredKeyRing(keyReader)
	if err != nil {
		return "", fmt.Errorf("failed to parse ASCII armor: %w", err)
	}

	if len(entityList) == 0 {
		return "", fmt.Errorf("no entities found in key data")
	}

	// Get the primary key
	entity := entityList[0]
	if entity.PrimaryKey == nil {
		return "", fmt.Errorf("no primary key found in entity")
	}

	fingerprint := entity.PrimaryKey.Fingerprint
	if len(fingerprint) == 0 {
		return "", fmt.Errorf("no fingerprint found")
	}

	// Convert to hex string (40 characters, uppercase)
	return hex.EncodeToString(fingerprint[:]), nil
}

// generateArtifactName generates an artifact file name
// Python reference: provider_extractor.py::ProviderExtractor.generate_artifact_name
func (s *ProviderExtractionGPGService) generateArtifactName(
	providerEntity *provider.Provider,
	version string,
	fileSuffix string,
) string {
	return fmt.Sprintf("terraform-provider-%s_%s_%s", providerEntity.Name(), version, fileSuffix)
}

// DownloadArtifact downloads a release artifact
// Python reference: provider_extractor.py::ProviderExtractor._download_artifact
func (s *ProviderExtractionGPGService) DownloadArtifact(
	ctx context.Context,
	providerEntity *provider.Provider,
	releaseMetadata *providermodel.RepositoryReleaseMetadata,
	repository *sqldb.RepositoryDB,
	fileName string,
	getReleaseArtifactFunc func(ctx context.Context, repo *sqldb.RepositoryDB, artifact *providermodel.ReleaseArtifactMetadata, accessToken string) ([]byte, error),
	accessToken string,
) ([]byte, error) {
	// Find the artifact in release metadata
	var artifact *providermodel.ReleaseArtifactMetadata
	for _, releaseArtifact := range releaseMetadata.ReleaseArtifacts {
		if releaseArtifact.Name == fileName {
			artifact = releaseArtifact
			break
		}
	}

	if artifact == nil {
		return nil, fmt.Errorf("%w: %s", provider.ErrMissingReleaseArtifact, fileName)
	}

	// Download the artifact
	content, err := getReleaseArtifactFunc(ctx, repository, artifact, accessToken)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", provider.ErrUnableToDownloadArtifact, err)
	}

	if content == nil {
		return nil, fmt.Errorf("%w: %s", provider.ErrMissingReleaseArtifact, fileName)
	}

	return content, nil
}

// ValidateChecksumFile validates the format of a checksum file
// Python reference: provider_extractor.py (implicit validation during processing)
func (s *ProviderExtractionGPGService) ValidateChecksumFile(content []byte) error {
	// Basic validation - check that file is not empty
	if len(content) == 0 {
		return provider.ErrInvalidChecksumFile
	}

	// TODO: More thorough validation could include:
	// - Verify each line has format: HASH  FILENAME
	// - Check that hash is valid hex
	// - Ensure at least one entry exists

	return nil
}

// ParseChecksumFile parses a SHA256SUMS file into a map of filename -> checksum
func (s *ProviderExtractionGPGService) ParseChecksumFile(content []byte) (map[string]string, error) {
	checksums := make(map[string]string)

	lines := bytes.Split(content, []byte{'\n'})
	for _, line := range lines {
		line = bytes.TrimSpace(line)
		if len(line) == 0 {
			continue
		}

		// Format: HASH  FILENAME
		parts := bytes.Fields(line)
		if len(parts) < 2 {
			continue
		}

		checksum := string(parts[0])
		filename := string(parts[1])

		// Validate checksum is valid hex
		if len(checksum) != 64 {
			return nil, fmt.Errorf("%w: invalid checksum length for %s", provider.ErrInvalidChecksumFile, filename)
		}

		if _, err := hex.DecodeString(checksum); err != nil {
			return nil, fmt.Errorf("%w: invalid hex checksum for %s", provider.ErrInvalidChecksumFile, filename)
		}

		checksums[filename] = checksum
	}

	if len(checksums) == 0 {
		return nil, provider.ErrInvalidChecksumFile
	}

	return checksums, nil
}

// VerifyBinaryChecksum verifies that a binary's content matches the expected checksum
func (s *ProviderExtractionGPGService) VerifyBinaryChecksum(
	filename string,
	content []byte,
	expectedChecksums map[string]string,
) error {
	expectedChecksum, ok := expectedChecksums[filename]
	if !ok {
		return fmt.Errorf("%w: no checksum found for %s", provider.ErrInvalidBinaryChecksum, filename)
	}

	// Compute SHA256 of content
	hasher := provider.NewSHA256Hash()
	hasher.Write(content)
	actualChecksum := hasher.Sum(nil)

	// Convert to hex
	actualChecksumHex := hex.EncodeToString(actualChecksum)

	if actualChecksumHex != expectedChecksum {
		return fmt.Errorf("%w: expected %s, got %s for %s",
			provider.ErrInvalidBinaryChecksum, expectedChecksum, actualChecksumHex, filename)
	}

	return nil
}
