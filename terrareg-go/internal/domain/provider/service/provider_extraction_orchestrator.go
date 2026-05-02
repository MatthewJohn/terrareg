package service

import (
	"context"
	"encoding/json"
	"fmt"


	configModel "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/config/model"
	gpgkeyModel "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/gpgkey/model"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/gpgkey/repository"
	providermodel "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/provider"
	providerModel "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/provider/model"
	providerSourceModel "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/provider_source/model"
	providerSourceService "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/provider_source/service"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/persistence/sqldb"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/logging"
)

// ProviderExtractionOrchestrator orchestrates the complete provider extraction workflow
// Python reference: provider_extractor.py::ProviderExtractor.process_version
type ProviderExtractionOrchestrator struct {
	gpgService              *ProviderExtractionGPGService
	sourceExtractionService *ProviderSourceExtractionService
	binaryProcessingService *ProviderBinaryProcessingService
	documentationService    *ProviderDocumentationService
	providerRepo            ProviderRepository
	gpgKeyRepo              repository.GPGKeyRepository
	providerSourceFactory   *providerSourceService.ProviderSourceFactory
	config                  *configModel.DomainConfig
	logger                  logging.Logger
}

// ProviderRepository defines the provider repository interface needed for extraction
type ProviderRepository interface {
	FindByID(ctx context.Context, id int) (*providermodel.Provider, error)
	FindVersionByID(ctx context.Context, versionID int) (*providermodel.ProviderVersion, error)
	SaveVersion(ctx context.Context, version *providermodel.ProviderVersion) error
}

// NewProviderExtractionOrchestrator creates a new provider extraction orchestrator
func NewProviderExtractionOrchestrator(
	gpgService *ProviderExtractionGPGService,
	sourceExtractionService *ProviderSourceExtractionService,
	binaryProcessingService *ProviderBinaryProcessingService,
	documentationService *ProviderDocumentationService,
	providerRepo ProviderRepository,
	gpgKeyRepo repository.GPGKeyRepository,
	providerSourceFactory *providerSourceService.ProviderSourceFactory,
	config *configModel.DomainConfig,
	logger logging.Logger,
) *ProviderExtractionOrchestrator {
	return &ProviderExtractionOrchestrator{
		gpgService:              gpgService,
		sourceExtractionService: sourceExtractionService,
		binaryProcessingService: binaryProcessingService,
		documentationService:    documentationService,
		providerRepo:            providerRepo,
		gpgKeyRepo:              gpgKeyRepo,
		providerSourceFactory:   providerSourceFactory,
		config:                  config,
		logger:                  logger,
	}
}

// ExtractProviderVersionRequest contains the parameters for extraction
type ExtractProviderVersionRequest struct {
	ProviderID   int
	Version      string
	GitTag       string
	ProviderName string
	Namespace    string
}

// ExtractProviderVersion orchestrates the complete provider extraction workflow
// Python reference: provider_extractor.py::ProviderExtractor.process_version
func (o *ProviderExtractionOrchestrator) ExtractProviderVersion(
	ctx context.Context,
	req ExtractProviderVersionRequest,
	repository *sqldb.RepositoryDB,
) error {
	o.logger.Info().
		Int("provider_id", req.ProviderID).
		Str("version", req.Version).
		Str("git_tag", req.GitTag).
		Msg("Starting provider version extraction")

	// Get provider entity
	providerEntity, err := o.providerRepo.FindByID(ctx, req.ProviderID)
	if err != nil {
		return fmt.Errorf("failed to find provider: %w", err)
	}
	if providerEntity == nil {
		return fmt.Errorf("provider not found")
	}

	// Get provider version entity
	providerVersion, err := o.providerRepo.FindVersionByID(ctx, req.ProviderID)
	if err != nil {
		return fmt.Errorf("failed to find provider version: %w", err)
	}
	if providerVersion == nil {
		return fmt.Errorf("provider version not found")
	}

	// Get provider source
	var apiName string
	if repository.Name != nil {
		apiName = *repository.Name
	} else {
		return fmt.Errorf("repository name is nil")
	}
	providerSource, err := o.providerSourceFactory.GetProviderSourceByApiName(ctx, apiName)
	if err != nil {
		return fmt.Errorf("failed to get provider source: %w", err)
	}
	if providerSource == nil {
		return fmt.Errorf("provider source not found: %s", apiName)
	}

	// Get access token (if needed)
	accessToken := ""
	// Note: Some provider sources may require access tokens for downloading artifacts

	// Get namespace GPG keys
	namespaceGPGKeys, err := o.gpgKeyRepo.FindByNamespace(ctx, req.Namespace)
	if err != nil {
		o.logger.Warn().Err(err).Str("namespace", req.Namespace).Msg("Failed to get namespace GPG keys, continuing without GPG verification")
		namespaceGPGKeys = []*gpgkeyModel.GPGKey{}
	}

	// Convert domain GPG keys to provider GPG keys
	gpgKeys := make([]*providermodel.GPGKey, len(namespaceGPGKeys))
	for i, gpgKey := range namespaceGPGKeys {
		gpgKeys[i] = providermodel.ReconstructGPGKey(
			gpgKey.ID(),
			gpgKey.ASCIIArmor(), // keyText - use ASCII armor as key text
			gpgKey.ASCIIArmor(),
			gpgKey.KeyID(),
			gpgKey.TrustSignature(),
			gpgKey.CreatedAt(), // No dereference - returns time.Time
			gpgKey.UpdatedAt(), // No dereference - returns time.Time
		)
	}

	// Create release metadata from repository and version info
	var archiveURL string
	if repository.CloneURL != nil && *repository.CloneURL != "" {
		// Use CloneURL for constructing archive URL
		archiveURL = fmt.Sprintf("%s/archive/%s.tar.gz", *repository.CloneURL, req.GitTag)
	} else {
		// Fallback to a default pattern if CloneURL is not available
		// This should be improved to properly construct the URL based on provider source
		archiveURL = fmt.Sprintf("archive/%s.tar.gz", req.GitTag)
	}
	releaseMetadata := &providerSourceModel.RepositoryReleaseMetadata{
		Name:             fmt.Sprintf("Release %s", req.Version),
		Tag:              req.GitTag,
		ArchiveURL:       archiveURL,
		CommitHash:       "", // Not needed for extraction
		ProviderID:       0,  // Not needed for extraction
		ReleaseArtifacts: []*providerSourceModel.ReleaseArtifactMetadata{},
	}

	// Get repository release metadata from provider source
	// This would normally fetch the actual release artifacts from GitHub/GitLab
	// For now, we'll use the repository to download artifacts

	// Step 1: Extract and verify manifest file
	if err := o.extractManifestFile(ctx, providerSource, repository, releaseMetadata, providerVersion, accessToken); err != nil {
		return fmt.Errorf("manifest extraction failed: %w", err)
	}

	// Step 2: Process binaries with checksum verification
	if err := o.extractBinaries(ctx, providerEntity, providerVersion, repository, releaseMetadata, providerSource, accessToken); err != nil {
		return fmt.Errorf("binary extraction failed: %w", err)
	}

	// Step 3: Extract documentation
	if err := o.extractDocumentation(ctx, providerEntity, repository, releaseMetadata, providerSource, accessToken); err != nil {
		// Documentation extraction failure is not fatal - log and continue
		o.logger.Warn().Err(err).Msg("Documentation extraction failed, continuing")
	}

	// Step 4: Verify and attach GPG key if available
	if len(gpgKeys) > 0 {
		matchedGPGKey, err := o.gpgService.ObtainGPGKey(
			ctx,
			providerEntity,
			gpgKeys,
			releaseMetadata,
			func(ctx context.Context, artifactName string) ([]byte, error) {
				return o.gpgService.DownloadArtifact(
					ctx,
					providerEntity,
					releaseMetadata,
					repository,
					artifactName,
					providerSource.GetReleaseArtifact,
					accessToken,
				)
			},
		)
		if err != nil {
			o.logger.Warn().Err(err).Msg("GPG key verification failed")
		} else if matchedGPGKey != nil {
			// Attach GPG key to provider version
			providerVersion.SetGPGKeyID(matchedGPGKey.ID())
			if err := o.providerRepo.SaveVersion(ctx, providerVersion); err != nil {
				return fmt.Errorf("failed to save provider version: %w", err)
			}
			o.logger.Info().
				Int("gpg_key_id", matchedGPGKey.ID()).
				Str("key_id", matchedGPGKey.KeyID()).
				Msg("GPG key attached to provider version")
		}
	}

	o.logger.Info().
		Int("provider_id", req.ProviderID).
		Str("version", req.Version).
		Msg("Provider version extraction completed successfully")

	return nil
}

// extractManifestFile extracts and validates the provider manifest file
// Python reference: provider_extractor.py::ProviderExtractor.extract_manifest_file
func (o *ProviderExtractionOrchestrator) extractManifestFile(
	ctx context.Context,
	providerSource providerSourceService.ProviderSourceInstance,
	repository *sqldb.RepositoryDB,
	releaseMetadata *providerSourceModel.RepositoryReleaseMetadata,
	providerVersion *providermodel.ProviderVersion,
	accessToken string,
) error {
	o.logger.Debug().Msg("Extracting manifest file")

	// Try to download manifest file
	manifestFileName := o.generateManifestFileName(providerVersion, repository)
	manifestContent, err := providerSource.GetReleaseArtifact(
		ctx,
		repository,
		providerSourceModel.NewReleaseArtifactMetadata(manifestFileName, 0),
		accessToken,
	)
	if err != nil {
		o.logger.Debug().Err(err).Msg("Manifest file not found, using default protocol versions")
		// Set default protocol versions if manifest not found
		defaultProtocolVersions := []string{"5.0"}
		providerVersion.SetProtocolVersions(defaultProtocolVersions)
		if err := o.providerRepo.SaveVersion(ctx, providerVersion); err != nil {
			return fmt.Errorf("failed to save provider version: %w", err)
		}
		return nil
	}

	// Parse manifest JSON
	var manifest providerModel.ManifestFile
	if err := json.Unmarshal(manifestContent, &manifest); err != nil {
		return fmt.Errorf("%w: %s", providermodel.ErrInvalidManifestFile, err)
	}

	// Validate manifest version
	if manifest.Version != 1 {
		return fmt.Errorf("%w: only version 1 is supported, got %d", providermodel.ErrInvalidManifestVersion, manifest.Version)
	}

	// Validate protocol versions
	if len(manifest.Metadata.ProtocolVersions) == 0 {
		return fmt.Errorf("%w: manifest must contain protocol versions", providermodel.ErrInvalidProtocolVersions)
	}

	// Update provider version with protocol versions
	providerVersion.SetProtocolVersions(manifest.Metadata.ProtocolVersions)
	if err := o.providerRepo.SaveVersion(ctx, providerVersion); err != nil {
		return fmt.Errorf("failed to save provider version: %w", err)
	}

	o.logger.Debug().
		Strs("protocol_versions", manifest.Metadata.ProtocolVersions).
		Msg("Manifest file processed successfully")

	return nil
}

// extractBinaries downloads, verifies, and stores provider binaries
// Python reference: provider_extractor.py::ProviderExtractor.extract_binaries
func (o *ProviderExtractionOrchestrator) extractBinaries(
	ctx context.Context,
	providerEntity *providermodel.Provider,
	providerVersion *providermodel.ProviderVersion,
	repository *sqldb.RepositoryDB,
	releaseMetadata *providerSourceModel.RepositoryReleaseMetadata,
	providerSource providerSourceService.ProviderSourceInstance,
	accessToken string,
) error {
	o.logger.Debug().Msg("Extracting binaries")

	// Get checksums for verification
	checksums, err := o.binaryProcessingService.ValidateChecksumFile(
		ctx,
		providerEntity,
		releaseMetadata,
		providerVersion.Version(),
		func(ctx context.Context, artifactName string) ([]byte, error) {
			return o.gpgService.DownloadArtifact(
				ctx,
				providerEntity,
				releaseMetadata,
				repository,
				artifactName,
				providerSource.GetReleaseArtifact,
				accessToken,
			)
		},
	)
	if err != nil {
		o.logger.Warn().Err(err).Msg("Failed to validate checksum file, continuing without checksums")
	}

	// Process binaries
	binaries, err := o.binaryProcessingService.ProcessBinaries(
		ctx,
		providerEntity,
		releaseMetadata,
		checksums,
		func(ctx context.Context, artifactName string) ([]byte, error) {
			return o.gpgService.DownloadArtifact(
				ctx,
				providerEntity,
				releaseMetadata,
				repository,
				artifactName,
				providerSource.GetReleaseArtifact,
				accessToken,
			)
		},
	)
	if err != nil {
		return fmt.Errorf("binary processing failed: %w", err)
	}

	// Store binaries in database and filesystem
	for _, binary := range binaries {
		o.logger.Debug().
			Str("os", binary.OS).
			Str("arch", binary.Architecture).
			Str("filename", binary.Filename).
			Msg("Processing binary")

		// TODO: Store binary in database and filesystem
		// This would require:
		// 1. Creating ProviderVersionBinaryDB record
		// 2. Storing binary content in filesystem
		// 3. Linking to provider version
		_ = binary
	}

	o.logger.Debug().Int("binary_count", len(binaries)).Msg("Binaries extracted successfully")
	return nil
}

// extractDocumentation extracts provider documentation
// Python reference: provider_extractor.py::ProviderExtractor.extract_documentation
func (o *ProviderExtractionOrchestrator) extractDocumentation(
	ctx context.Context,
	providerEntity *providermodel.Provider,
	repository *sqldb.RepositoryDB,
	releaseMetadata *providerSourceModel.RepositoryReleaseMetadata,
	providerSource providerSourceService.ProviderSourceInstance,
	accessToken string,
) error {
	o.logger.Debug().Msg("Extracting documentation")

	// Obtain source code
	sourceResult, err := o.sourceExtractionService.ObtainSourceCode(
		ctx,
		providerEntity,
		releaseMetadata,
		repository,
		providerSource.GetReleaseArchive,
		accessToken,
	)
	if err != nil {
		return fmt.Errorf("failed to obtain source code: %w", err)
	}
	defer sourceResult.Cleanup()

	// Extract documentation from source
	docs, err := o.documentationService.ExtractDocumentation(
		ctx,
		sourceResult.SourceDir,
		providerEntity,
	)
	if err != nil {
		return fmt.Errorf("documentation extraction failed: %w", err)
	}

	// Store documentation in database
	for _, doc := range docs {
		o.logger.Debug().
			Str("type", string(doc.Type)).
			Str("name", doc.Name).
			Msg("Processing documentation")

		// TODO: Store documentation in database
		// This would require:
		// 1. Creating ProviderVersionDocumentationDB record
		// 2. Storing content as blob
		// 3. Linking to provider version
		_ = doc
	}

	o.logger.Debug().Int("doc_count", len(docs)).Msg("Documentation extracted successfully")
	return nil
}

// generateManifestFileName generates the manifest file name
// Python reference: provider_extractor.py (manifest file naming pattern)
func (o *ProviderExtractionOrchestrator) generateManifestFileName(
	providerVersion *providermodel.ProviderVersion,
	repository *sqldb.RepositoryDB,
) string {
	version := providerVersion.Version()
	var repoName string
	if repository.Name != nil {
		repoName = *repository.Name
	} else {
		repoName = "unknown"
	}
	return fmt.Sprintf("terraform-provider-%s_%s_manifest.json", repoName, version)
}
