package service

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/provider"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/provider/repository"
)

// ProviderPublisherService handles publishing providers and their binaries
type ProviderPublisherService struct {
	providerRepo     repository.ProviderRepository
	extractorService *ProviderExtractorService
	storageService   ProviderStorageService
	gpgService       *ProviderGPGService
}

// NewProviderPublisherService creates a new provider publisher service
func NewProviderPublisherService(
	providerRepo repository.ProviderRepository,
	extractorService *ProviderExtractorService,
	storageService ProviderStorageService,
	gpgService *ProviderGPGService,
) *ProviderPublisherService {
	return &ProviderPublisherService{
		providerRepo:     providerRepo,
		extractorService: extractorService,
		storageService:   storageService,
		gpgService:       gpgService,
	}
}

// PublishRequest represents a request to publish a provider version
type PublishRequest struct {
	ProviderID      int
	Version         string
	SourcePath      string // Path to source code or archive
	ProtocolVersion []string
	IsBeta          bool
	GitTag          *string
	GPGKeyID        *string
}

// PublishResponse represents the response from publishing
type PublishResponse struct {
	VersionID   int
	Binaries    []*PublishedBinary
	Checksums   map[string]string
	DownloadURL string
}

// PublishedBinary represents information about a published binary
type PublishedBinary struct {
	BinaryID    int
	OS          string
	Arch        string
	Filename    string
	Size        int64
	Checksum    string
	DownloadURL string
}

// ProviderStorageService defines the interface for storing provider binaries
type ProviderStorageService interface {
	StoreBinary(ctx context.Context, key string, data io.Reader) (string, error)
	GetBinaryURL(ctx context.Context, key string) (string, error)
	DeleteBinary(ctx context.Context, key string) error
}

// Publish publishes a new provider version
func (s *ProviderPublisherService) Publish(ctx context.Context, req PublishRequest) (*PublishResponse, error) {
	// Find provider
	providerEntity, err := s.providerRepo.FindByID(ctx, req.ProviderID)
	if err != nil {
		return nil, fmt.Errorf("failed to find provider: %w", err)
	}
	if providerEntity == nil {
		return nil, fmt.Errorf("provider not found")
	}

	// Validate source
	if err := s.extractorService.ValidateProviderStructure(req.SourcePath); err != nil {
		return nil, fmt.Errorf("invalid provider source: %w", err)
	}

	// Extract metadata
	metadata, err := s.extractorService.ExtractFromArchive(ctx, req.SourcePath)
	if err != nil {
		return nil, fmt.Errorf("failed to extract metadata: %w", err)
	}

	// Use provided version or extracted version
	version := req.Version
	if version == "" {
		version = metadata.Version
	}
	if version == "" {
		return nil, fmt.Errorf("version is required")
	}

	// Check if version already exists
	existingVersion, err := s.providerRepo.FindVersionByProviderAndVersion(ctx, req.ProviderID, version)
	if err == nil && existingVersion != nil {
		return nil, fmt.Errorf("version %s already exists", version)
	}

	// Create version using aggregate
	protocolVersions := req.ProtocolVersion
	if len(protocolVersions) == 0 && len(metadata.Protocols) > 0 {
		protocolVersions = metadata.Protocols
	}
	if len(protocolVersions) == 0 {
		protocolVersions = []string{"5.0"} // Default protocol
	}

	newVersion, err := providerEntity.PublishVersion(version, protocolVersions, req.IsBeta)
	if err != nil {
		return nil, fmt.Errorf("failed to create version: %w", err)
	}

	// Set git tag if provided
	if req.GitTag != nil {
		newVersion.SetGitTag(req.GitTag)
	} else if metadata.GPGKeyID != "" {
		newVersion.SetGitTag(&metadata.GPGKeyID)
	}

	// Set GPG key if provided
	if req.GPGKeyID != nil {
		gpgKey := providerEntity.FindGPGKeyByKeyID(*req.GPGKeyID)
		if gpgKey != nil {
			newVersion.SetGPGKeyID(gpgKey.ID())
		}
	}

	// Save version
	if err := s.providerRepo.SaveVersion(ctx, newVersion); err != nil {
		return nil, fmt.Errorf("failed to save version: %w", err)
	}

	// Extract and publish binaries
	binaries, err := s.publishBinaries(ctx, newVersion, req.SourcePath)
	if err != nil {
		return nil, fmt.Errorf("failed to publish binaries: %w", err)
	}

	// Set as latest version if not beta
	if !req.IsBeta {
		if err := s.providerRepo.SetLatestVersion(ctx, req.ProviderID, newVersion.ID()); err != nil {
			// Non-critical error, log but don't fail
			fmt.Printf("Warning: failed to set latest version: %v\n", err)
		}
	}

	// Generate checksums
	checksums := s.generateChecksums(binaries)

	return &PublishResponse{
		VersionID:   newVersion.ID(),
		Binaries:    binaries,
		Checksums:   checksums,
		DownloadURL: fmt.Sprintf("/providers/%d/versions/%s/download", req.ProviderID, version),
	}, nil
}

// publishBinaries extracts and publishes provider binaries
func (s *ProviderPublisherService) publishBinaries(ctx context.Context, version *provider.ProviderVersion, sourcePath string) ([]*PublishedBinary, error) {
	// Extract binaries from source
	binaryInfos, err := s.extractorService.ExtractBinaries(ctx, sourcePath)
	if err != nil {
		return nil, fmt.Errorf("failed to extract binaries: %w", err)
	}

	publishedBinaries := make([]*PublishedBinary, 0, len(binaryInfos))

	for _, binaryInfo := range binaryInfos {
		// Upload binary to storage
		file, err := os.Open(binaryInfo.Path)
		if err != nil {
			return nil, fmt.Errorf("failed to open binary %s: %w", binaryInfo.Filename, err)
		}
		defer file.Close()

		// Generate storage key
		storageKey := fmt.Sprintf("providers/%d/versions/%d/%s", version.ProviderID(), version.ID(), binaryInfo.Filename)

		// Store binary
		downloadURL, err := s.storageService.StoreBinary(ctx, storageKey, file)
		if err != nil {
			return nil, fmt.Errorf("failed to store binary %s: %w", binaryInfo.Filename, err)
		}

		// Create provider binary entity
		fileInfo, err := file.Stat()
		if err != nil {
			return nil, fmt.Errorf("failed to get file info: %w", err)
		}

		providerBinary := provider.NewProviderBinary(
			version.ID(),
			binaryInfo.OS,
			binaryInfo.Arch,
			binaryInfo.Filename,
			binaryInfo.Checksum,
			downloadURL,
			fileInfo.Size(),
		)

		// Save binary to repository
		if err := s.providerRepo.SaveBinary(ctx, providerBinary); err != nil {
			return nil, fmt.Errorf("failed to save binary %s: %w", binaryInfo.Filename, err)
		}

		// Add binary to version
		if err := version.AddBinary(providerBinary); err != nil {
			return nil, fmt.Errorf("failed to add binary to version: %w", err)
		}

		publishedBinary := &PublishedBinary{
			BinaryID:    providerBinary.ID(),
			OS:          binaryInfo.OS,
			Arch:        binaryInfo.Arch,
			Filename:    binaryInfo.Filename,
			Size:        fileInfo.Size(),
			Checksum:    binaryInfo.Checksum,
			DownloadURL: downloadURL,
		}

		publishedBinaries = append(publishedBinaries, publishedBinary)
	}

	return publishedBinaries, nil
}

// PublishBinary publishes a single binary for an existing version
func (s *ProviderPublisherService) PublishBinary(ctx context.Context, versionID int, binaryPath, operatingSystem, arch string) (*PublishedBinary, error) {
	// Get version
	version, err := s.providerRepo.FindVersionByID(ctx, versionID)
	if err != nil {
		return nil, fmt.Errorf("failed to find version: %w", err)
	}
	if version == nil {
		return nil, fmt.Errorf("version not found")
	}

	// Validate binary file
	if _, err := os.Stat(binaryPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("binary file not found: %s", binaryPath)
	}

	// Open binary file
	file, err := os.Open(binaryPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open binary: %w", err)
	}
	defer file.Close()

	// Get file info
	fileInfo, err := file.Stat()
	if err != nil {
		return nil, fmt.Errorf("failed to get file info: %w", err)
	}

	// Generate storage key
	filename := filepath.Base(binaryPath)
	storageKey := fmt.Sprintf("providers/%d/versions/%d/%s", version.ProviderID(), versionID, filename)

	// Store binary
	downloadURL, err := s.storageService.StoreBinary(ctx, storageKey, file)
	if err != nil {
		return nil, fmt.Errorf("failed to store binary: %w", err)
	}

	// Generate checksum
	checksum, err := s.generateChecksum(binaryPath)
	if err != nil {
		return nil, fmt.Errorf("failed to generate checksum: %w", err)
	}

	// Create provider binary entity
	providerBinary := provider.NewProviderBinary(
		versionID,
		operatingSystem,
		arch,
		filename,
		checksum,
		downloadURL,
		fileInfo.Size(),
	)

	// Save binary to repository
	if err := s.providerRepo.SaveBinary(ctx, providerBinary); err != nil {
		return nil, fmt.Errorf("failed to save binary: %w", err)
	}

	// Add binary to version
	if err := version.AddBinary(providerBinary); err != nil {
		return nil, fmt.Errorf("failed to add binary to version: %w", err)
	}

	return &PublishedBinary{
		BinaryID:    providerBinary.ID(),
		OS:          operatingSystem,
		Arch:        arch,
		Filename:    filename,
		Size:        fileInfo.Size(),
		Checksum:    checksum,
		DownloadURL: downloadURL,
	}, nil
}

// generateChecksums generates checksums for all published binaries
func (s *ProviderPublisherService) generateChecksums(binaries []*PublishedBinary) map[string]string {
	checksums := make(map[string]string)
	for _, binary := range binaries {
		checksums[binary.Filename] = binary.Checksum
	}
	return checksums
}

// generateChecksum generates a SHA256 checksum for a file
func (s *ProviderPublisherService) generateChecksum(filePath string) (string, error) {
	// Placeholder implementation
	// In a real implementation, this would generate a SHA256 checksum
	return fmt.Sprintf("sha256-checksum-%s-%d", filepath.Base(filePath), time.Now().Unix()), nil
}