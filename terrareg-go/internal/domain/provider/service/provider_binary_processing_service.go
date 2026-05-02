package service

import (
	"context"
	"encoding/hex"
	"fmt"
	"regexp"
	"strings"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/provider"
	providermodel "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/provider_source/model"
)

// ProviderBinaryProcessingService handles processing provider binaries
// Python reference: provider_extractor.py::ProviderExtractor.extract_binaries
type ProviderBinaryProcessingService struct {
	gpgService *ProviderExtractionGPGService
}

// NewProviderBinaryProcessingService creates a new binary processing service
func NewProviderBinaryProcessingService(gpgService *ProviderExtractionGPGService) *ProviderBinaryProcessingService {
	return &ProviderBinaryProcessingService{
		gpgService: gpgService,
	}
}

// BinaryResult represents a processed binary with metadata
type BinaryResult struct {
	OS           string
	Architecture string
	Filename     string
	Checksum     string
	Content      []byte
}

// ProcessBinaries processes binaries from release artifacts
// Python reference: provider_extractor.py::ProviderExtractor.extract_binaries
func (s *ProviderBinaryProcessingService) ProcessBinaries(
	ctx context.Context,
	providerEntity *provider.Provider,
	releaseMetadata *providermodel.RepositoryReleaseMetadata,
	checksums map[string]string,
	downloadArtifactFunc func(ctx context.Context, artifactName string) ([]byte, error),
) ([]*BinaryResult, error) {
	binaries := make([]*BinaryResult, 0)

	// Process each artifact in the release
	for _, artifact := range releaseMetadata.ReleaseArtifacts {
		// Skip manifest and checksum files
		if s.isMetadataFile(artifact.Name) {
			continue
		}

		// Check if this is a provider binary
		platform, err := s.extractPlatformFromFilename(artifact.Name)
		if err != nil {
			// Not a binary file, skip
			continue
		}

		// Download the binary
		content, err := downloadArtifactFunc(ctx, artifact.Name)
		if err != nil {
			return nil, fmt.Errorf("%w: failed to download %s", provider.ErrUnableToDownloadArtifact, artifact.Name)
		}

		// Verify checksum
		if len(checksums) > 0 {
			if err := s.gpgService.VerifyBinaryChecksum(artifact.Name, content, checksums); err != nil {
				return nil, err
			}
		}

		// Compute checksum for verification
		hasher := provider.NewSHA256Hash()
		hasher.Write(content)
		checksum := hex.EncodeToString(hasher.Sum(nil))

		binary := &BinaryResult{
			OS:           platform.OS,
			Architecture: platform.Architecture,
			Filename:     artifact.Name,
			Checksum:     checksum,
			Content:      content,
		}

		binaries = append(binaries, binary)
	}

	return binaries, nil
}

// isMetadataFile checks if a file is metadata (manifest, checksums, signature)
func (s *ProviderBinaryProcessingService) isMetadataFile(filename string) bool {
	metadataPatterns := []string{
		"_SHA256SUMS",
		"_SHA256SUMS.sig",
		"_SHA256SUMS.asc",
		"manifest.json",
	}

	for _, pattern := range metadataPatterns {
		if strings.Contains(filename, pattern) {
			return true
		}
	}

	return false
}

// BinaryPlatform represents the OS and architecture for a binary
// Python reference: provider_extractor.py::ProviderExtractor._get_os_and_arch
type BinaryPlatform struct {
	OS           string
	Architecture string
}

// extractPlatformFromFilename extracts platform info from filename
// Python reference: provider_extractor.py::ProviderExtractor._get_os_and_arch
func (s *ProviderBinaryProcessingService) extractPlatformFromFilename(filename string) (*BinaryPlatform, error) {
	// Convert to lowercase for matching
	lowerFilename := strings.ToLower(filename)

	// OS patterns (ordered by specificity)
	osPatterns := map[string]string{
		"windows": "windows",
		"linux":   "linux",
		"darwin":  "darwin",
		"freebsd": "freebsd",
		"openbsd": "openbsd",
		"solaris": "solaris",
	}

	// Architecture patterns (ordered by specificity - longer patterns first)
	// Important: longer patterns must be checked before shorter ones
	archPatterns := []struct {
		pattern string
		value   string
	}{
		{"amd64", "amd64"},
		{"x86_64", "amd64"},
		{"arm64", "arm64"},
		{"aarch64", "arm64"},
		{"armv7l", "arm"},
		{"ppc64le", "ppc64le"},
		{"ppc64", "ppc64"},
		{"mips64le", "mips64le"},
		{"mips64", "mips64"},
		{"mipsle", "mipsle"},
		{"mips", "mips"},
		{"riscv64", "riscv64"},
		{"s390x", "s390x"},
		{"386", "386"},
		{"i386", "386"},
		{"arm", "arm"},
	}

	var detectedOS, detectedArch string

	// Detect OS
	for osKey, osValue := range osPatterns {
		if strings.Contains(lowerFilename, osKey) {
			detectedOS = osValue
			break
		}
	}

	// Detect Architecture
	// Try to find OS separator first (e.g., "linux_" or "windows_")
	osSeparator := "_" + detectedOS + "_"
	osIndex := strings.Index(lowerFilename, osSeparator)
	if osIndex == -1 {
		osSeparator = "_" + detectedOS + "." // e.g., "linux.zip"
		osIndex = strings.Index(lowerFilename, osSeparator)
	}

	// Look for architecture after the OS separator
	if osIndex != -1 {
		afterOS := lowerFilename[osIndex+len(osSeparator):]
		// Check if any arch pattern matches at the start
		for _, archPattern := range archPatterns {
			if strings.HasPrefix(afterOS, archPattern.pattern) {
				detectedArch = archPattern.value
				break
			}
		}
	}

	// Fallback: try to find arch pattern anywhere (but be careful with partial matches)
	if detectedArch == "" {
		for _, archPattern := range archPatterns {
			// Match only if surrounded by separators, dots, or version numbers
			pattern := regexp.MustCompile(`(?:_|\.|^)` + regexp.QuoteMeta(archPattern.pattern) + `(?:_|\.|zip|$)`)
			if pattern.MatchString(lowerFilename) {
				detectedArch = archPattern.value
				break
			}
		}
	}

	// Must detect both OS and architecture
	if detectedOS == "" || detectedArch == "" {
		return nil, fmt.Errorf("%w: could not detect platform from filename %s (os=%s, arch=%s)",
			provider.ErrInvalidBinaryPlatform, filename, detectedOS, detectedArch)
	}

	return &BinaryPlatform{
		OS:           detectedOS,
		Architecture: detectedArch,
	}, nil
}

// ValidateChecksumFile validates and parses a checksum file
// Python reference: provider_extractor.py (checksum validation during binary extraction)
func (s *ProviderBinaryProcessingService) ValidateChecksumFile(
	ctx context.Context,
	providerEntity *provider.Provider,
	releaseMetadata *providermodel.RepositoryReleaseMetadata,
	version string,
	downloadArtifactFunc func(ctx context.Context, artifactName string) ([]byte, error),
) (map[string]string, error) {
	// Generate checksum file name
	checksumFileName := fmt.Sprintf("terraform-provider-%s_%s_SHA256SUMS", providerEntity.Name(), version)

	// Download checksum file
	content, err := downloadArtifactFunc(ctx, checksumFileName)
	if err != nil {
		// Checksums are optional - return empty map if not found
		return nil, nil
	}

	// Validate and parse checksum file
	checksums, err := s.gpgService.ParseChecksumFile(content)
	if err != nil {
		return nil, fmt.Errorf("failed to parse checksum file: %w", err)
	}

	return checksums, nil
}

// GetBinaryPlatformString returns the platform string for database storage
// Python reference: provider_extractor.py::ProviderExtractor._get_os_and_arch
func (s *ProviderBinaryProcessingService) GetBinaryPlatformString(os, arch string) string {
	return fmt.Sprintf("%s_%s", os, arch)
}

// ParseBinaryPlatformString parses a platform string back to OS and architecture
func (s *ProviderBinaryProcessingService) ParseBinaryPlatformString(platform string) (string, string, error) {
	parts := strings.Split(platform, "_")
	if len(parts) != 2 {
		return "", "", fmt.Errorf("%w: invalid platform format: %s", provider.ErrInvalidBinaryPlatform, platform)
	}
	return parts[0], parts[1], nil
}
