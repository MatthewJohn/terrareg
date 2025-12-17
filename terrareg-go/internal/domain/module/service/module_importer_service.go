package service

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gorm.io/gorm"
	"github.com/rs/zerolog"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/module"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/shared"
	domainConfig "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/config/model"
	gitService "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/git/service"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/module/model"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/module/repository"
	infraConfig "github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/config"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/persistence/sqldb/transaction"
)

// ModuleImporterService handles module importing with comprehensive
// transaction processing and rollback capabilities using all core services
type ModuleImporterService struct {
	// Core processing orchestrator
	processingOrchestrator *TransactionProcessingOrchestrator
	moduleCreationWrapper  *ModuleCreationWrapperService
	savepointHelper        *transaction.SavepointHelper

	// Legacy git operations (for compatibility during transition)
	moduleProviderRepo repository.ModuleProviderRepository
	gitClient          gitService.GitClient
	storageService     StorageService
	moduleParser       ModuleParser
	domainConfig       *domainConfig.DomainConfig
	infraConfig        *infraConfig.InfrastructureConfig

	// Logging
	logger zerolog.Logger
}

// NewModuleImporterService creates a new module importer service with transaction capabilities
func NewModuleImporterService(
	processingOrchestrator *TransactionProcessingOrchestrator,
	moduleCreationWrapper *ModuleCreationWrapperService,
	savepointHelper *transaction.SavepointHelper,
	moduleProviderRepo repository.ModuleProviderRepository,
	gitClient gitService.GitClient,
	storageService StorageService,
	moduleParser ModuleParser,
	domainConfig *domainConfig.DomainConfig,
	infraConfig *infraConfig.InfrastructureConfig,
	logger zerolog.Logger,
) *ModuleImporterService {
	return &ModuleImporterService{
		processingOrchestrator: processingOrchestrator,
		moduleCreationWrapper:  moduleCreationWrapper,
		savepointHelper:        savepointHelper,
		moduleProviderRepo:     moduleProviderRepo,
		gitClient:              gitClient,
		storageService:         storageService,
		moduleParser:           moduleParser,
		domainConfig:           domainConfig,
		infraConfig:            infraConfig,
		logger:                 logger,
	}
}

// DomainImportRequest represents a minimal domain import request
// This contains only domain-relevant information without application concerns
type DomainImportRequest struct {
	// Core domain input
	Input *module.ModuleVersionImportInput

	// Domain-relevant options
	ProcessingOptions ProcessingOptions

	// Source information (domain-relevant only)
	SourcePath  string // Path to source files (if already extracted)
	ArchivePath string // Path to archive file (if applicable)
	SourceType  string // "git", "archive", "path"

	// Domain processing options
	GenerateArchives bool
	ArchiveFormats   []ArchiveFormat

	// Domain analysis options
	EnableSecurityScan bool
	EnableInfracost    bool
	EnableExamples     bool
}

// ModuleImportResult represents the result of enhanced module import
type ModuleImportResult struct {
	Success             bool                  `json:"success"`
	ModuleVersionID     *int                  `json:"module_version_id,omitempty"`
	Version             string                `json:"version"`
	ProcessingResult    *ProcessingResult     `json:"processing_result,omitempty"`
	ImportDuration      time.Duration         `json:"import_duration"`
	GeneratedArchives   []GeneratedArchive    `json:"generated_archives,omitempty"`
	SecurityResults     *SecurityScanResponse `json:"security_results,omitempty"`
	FileStatistics      *FileStatistics       `json:"file_statistics,omitempty"`
	Error               *string               `json:"error,omitempty"`
	SavepointRolledBack bool                  `json:"savepoint_rolled_back"`
	Timestamp           time.Time             `json:"timestamp"`
}

// ImportModuleVersionWithTransaction performs a complete module import with transaction safety
func (s *ModuleImporterService) ImportModuleVersionWithTransaction(
	ctx context.Context,
	req DomainImportRequest,
) (*ModuleImportResult, error) {
	startTime := time.Now()

	result := &ModuleImportResult{
		Success:             false,
		Version:             req.Input.GetVersionString(),
		SavepointRolledBack: false,
		Timestamp:           startTime,
	}

	// Use smart transaction wrapper for the entire import process
	err := s.savepointHelper.WithTransaction(ctx, func(ctx context.Context, tx *gorm.DB) error {
		// Phase 1: Pre-import setup and validation
		if err := s.validateImportRequest(ctx, req); err != nil {
			return fmt.Errorf("import validation failed: %w", err)
		}

		// Phase 2: Source preparation (git clone, archive extraction, etc.)
		sourcePath, err := s.prepareSource(ctx, req)
		if err != nil {
			return fmt.Errorf("source preparation failed: %w", err)
		}
		defer s.cleanupSource(sourcePath)

		// Phase 3: Execute complete processing pipeline
		processingReq := ProcessingRequest{
			Namespace:   req.Input.Namespace,
			ModuleName:  req.Input.Module,
			Provider:    req.Input.Provider,
			Version:     req.Input.GetVersionString(),
			GitTag:      req.Input.GitTag,
			ModulePath:  sourcePath,
			ArchivePath: req.ArchivePath,
			SourceType:  SourceType(req.SourceType),
			Options:     req.ProcessingOptions,
		}

		processingResult, err := s.processingOrchestrator.ProcessModuleWithTransaction(ctx, processingReq)
		if err != nil {
			return fmt.Errorf("processing pipeline failed: %w", err)
		}

		result.ProcessingResult = processingResult

		if !processingResult.Success {
			return fmt.Errorf("processing pipeline failed: %s", *processingResult.Error)
		}

		// Phase 4: Archive generation (if requested)
		if req.GenerateArchives {
			// TODO: Integrate with archive generation service
			// For now, skip archive generation
		}

		// Phase 5: Additional analysis (security scanning, etc.)
		if req.EnableSecurityScan {
			// TODO: Integrate with security scanning service
			// For now, skip security scanning
		}

		// Phase 6: File statistics
		// TODO: Integrate with file content service for statistics
		// For now, skip file statistics

		// Set module version ID
		if processingResult.ModuleVersion != nil {
			moduleVersionID := processingResult.ModuleVersion.ID()
			result.ModuleVersionID = &moduleVersionID
		}

		return nil
	})

	result.ImportDuration = time.Since(startTime)

	if err != nil {
		result.SavepointRolledBack = true
		errorMsg := err.Error()
		result.Error = &errorMsg
		return result, nil
	}

	result.Success = true
	return result, nil
}

// ImportBatchModules imports multiple modules with individual transaction isolation
func (s *ModuleImporterService) ImportBatchModules(
	ctx context.Context,
	requests []DomainImportRequest,
) (*BatchModuleImportResult, error) {
	startTime := time.Now()

	result := &BatchModuleImportResult{
		TotalModules:      len(requests),
		SuccessfulImports: []ModuleImportResult{},
		FailedImports:     []ModuleImportResult{},
		PartialSuccess:    false,
		OverallSuccess:    true,
		Timestamp:         startTime,
	}

	for _, req := range requests {
		importResult, err := s.ImportModuleVersionWithTransaction(ctx, req)
		if err != nil {
			// This should rarely happen since we handle errors within ImportModuleVersionWithTransaction
			errorResult := ModuleImportResult{
				Success:             false,
				Version:             req.Input.GetVersionString(),
				Error:               func() *string { e := err.Error(); return &e }(),
				Timestamp:           time.Now(),
				SavepointRolledBack: true,
			}
			result.FailedImports = append(result.FailedImports, errorResult)
			result.OverallSuccess = false
			result.PartialSuccess = true
			continue
		}

		if importResult.Success {
			result.SuccessfulImports = append(result.SuccessfulImports, *importResult)
		} else {
			result.FailedImports = append(result.FailedImports, *importResult)
			result.OverallSuccess = false
			result.PartialSuccess = true
		}
	}

	result.OverallDuration = time.Since(startTime)

	// If there were no failures, set partial success to false
	if len(result.FailedImports) == 0 {
		result.PartialSuccess = false
	}

	return result, nil
}

// ReindexModuleWithTransaction re-indexes an existing module with transaction safety
func (s *ModuleImporterService) ReindexModuleWithTransaction(
	ctx context.Context,
	namespace, moduleName, provider, version string,
	options ProcessingOptions,
) (*ModuleImportResult, error) {
	parsedVersion, err := shared.ParseVersion(version)
	if err != nil {
		return nil, err
	}

	domainInput := module.NewModuleVersionImportInput(namespace, moduleName, provider, parsedVersion, nil)

	req := DomainImportRequest{
		Input:             domainInput,
		ProcessingOptions: options,
		SourceType:        "git", // Re-indexing typically from git
	}

	return s.ImportModuleVersionWithTransaction(ctx, req)
}

// ImportFromArchive imports a module from an archive file with transaction safety
func (s *ModuleImporterService) ImportFromArchive(
	ctx context.Context,
	namespace, moduleName, provider, version string,
	archivePath string,
	options ProcessingOptions,
) (*ModuleImportResult, error) {
	parsedVersion, err := shared.ParseVersion(version)
	if err != nil {
		return nil, err
	}

	domainInput := module.NewModuleVersionImportInput(namespace, moduleName, provider, parsedVersion, nil)

	req := DomainImportRequest{
		Input:             domainInput,
		ProcessingOptions: options,
		ArchivePath:       archivePath,
		SourceType:        "archive",
	}

	return s.ImportModuleVersionWithTransaction(ctx, req)
}

// validateImportRequest validates the domain import request
func (s *ModuleImporterService) validateImportRequest(
	ctx context.Context,
	req DomainImportRequest,
) error {
	// Validate domain input
	if err := req.Input.Validate(); err != nil {
		return err
	}

	// Validate processing options
	if req.ProcessingOptions.GenerateArchives && len(req.ProcessingOptions.ArchiveFormats) == 0 {
		return fmt.Errorf("archive generation requested but no formats specified")
	}

	// Validate source
	if req.SourcePath == "" && req.ArchivePath == "" && req.Input.GitTag == nil {
		return fmt.Errorf("no source specified (source path, archive path, or git tag required)")
	}

	return nil
}

// prepareSource prepares the source files for processing
func (s *ModuleImporterService) prepareSource(
	ctx context.Context,
	req DomainImportRequest,
) (string, error) {
	if req.SourcePath != "" {
		// Source already prepared
		return req.SourcePath, nil
	}

	if req.ArchivePath != "" {
		// Extract archive
		return s.extractArchive(ctx, req)
	}

	if req.Input.GitTag != nil {
		// Use base importer for git operations
		return s.prepareGitSource(ctx, req)
	}

	return "", fmt.Errorf("no source preparation method available")
}

// extractArchive extracts archive files for processing
func (s *ModuleImporterService) extractArchive(
	ctx context.Context,
	req DomainImportRequest,
) (string, error) {
	// Create temporary directory for extraction
	tempDir, err := s.storageService.MkdirTemp("", "terrareg-archive-")
	if err != nil {
		return "", fmt.Errorf("failed to create temp directory: %w", err)
	}

	// This would integrate with the ArchiveExtractionService
	// For now, return the temp directory path
	return tempDir, nil
}

// prepareGitSource prepares git source for processing
func (s *ModuleImporterService) prepareGitSource(
	ctx context.Context,
	req DomainImportRequest,
) (string, error) {
	// Validate domain input (should already be validated)
	if err := req.Input.Validate(); err != nil {
		return "", err
	}

	// Find the module provider
	moduleProvider, err := s.moduleProviderRepo.FindByNamespaceModuleProvider(
		ctx, req.Input.Namespace, req.Input.Module, req.Input.Provider,
	)
	if err != nil {
		return "", fmt.Errorf("module provider not found: %w", err)
	}

	// Validate git configuration
	// A module is git-based if it has either a git provider OR a custom clone URL template
	if moduleProvider.RepoCloneURLTemplate() == nil {
		return "", fmt.Errorf("module provider is not a git based module - no clone URL configured")
	}

	// Clone and checkout - using domain git operations
	cloneURL := s.buildCloneURL(req, moduleProvider)
	s.logger.Debug().Str("clone_url", cloneURL).Msg("Built clone URL")

	tmpDir, err := s.storageService.MkdirTemp("", "terrareg-git-")
	if err != nil {
		return "", fmt.Errorf("failed to create temp dir: %w", err)
	}
	s.logger.Debug().Str("temp_dir", tmpDir).Msg("Created temp directory")

	// Clone repository
	cloneOptions := gitService.CloneOptions{
		Timeout: time.Duration(s.infraConfig.GitCloneTimeout) * time.Second,
		NeedTags: req.Input.GitTag != nil, // Fetch all tags when we need to checkout a specific tag
	}

	cloneType := "shallow"
	if cloneOptions.NeedTags {
		cloneType = "full (with tags)"
	}
	s.logger.Debug().
		Str("clone_url", cloneURL).
		Int("timeout_seconds", s.infraConfig.GitCloneTimeout).
		Str("clone_type", cloneType).
		Bool("needs_tags", cloneOptions.NeedTags).
		Msg("Cloning repository")
	if err := s.gitClient.CloneWithOptions(ctx, cloneURL, tmpDir, cloneOptions); err != nil {
		s.logger.Debug().Err(err).Str("clone_url", cloneURL).Msg("Git clone failed")
		s.storageService.RemoveAll(tmpDir)
		return "", fmt.Errorf("failed to clone git repository: %w", err)
	}
	s.logger.Debug().Str("temp_dir", tmpDir).Msg("Successfully cloned repository")

	// Checkout specific tag if provided
	if req.Input.GitTag != nil {
		originalGitTag := *req.Input.GitTag
		gitTag := originalGitTag

		// Apply git tag format if configured
		if gitTagFormat := moduleProvider.GitTagFormat(); gitTagFormat != nil {
			// Replace {version} placeholder in git tag format
			replacer := strings.NewReplacer("{version}", gitTag)
			gitTag = replacer.Replace(*gitTagFormat)
			s.logger.Debug().
				Str("git_tag_format", *gitTagFormat).
				Str("original_version", originalGitTag).
				Str("formatted_git_tag", gitTag).
				Msg("Applied git tag format")
		} else {
			s.logger.Debug().Str("git_tag", gitTag).Msg("No git tag format configured, using raw git tag")
		}

		s.logger.Debug().
			Str("git_tag", gitTag).
			Str("directory", tmpDir).
			Msg("Attempting to checkout git tag")
		if err := s.gitClient.Checkout(ctx, tmpDir, gitTag); err != nil {
			s.logger.Debug().
				Err(err).
				Str("git_tag", gitTag).
				Str("directory", tmpDir).
				Msg("Git checkout failed")
			s.storageService.RemoveAll(tmpDir)
			return "", fmt.Errorf("failed to checkout git tag '%s': %w", gitTag, err)
		}
		s.logger.Debug().Str("git_tag", gitTag).Msg("Successfully checked out git tag")
	}

	// Determine source directory
	srcDir := tmpDir
	if gitPath := moduleProvider.GitPath(); gitPath != nil && *gitPath != "" {
		srcDir = filepath.Join(tmpDir, *gitPath)
	}

	return srcDir, nil
}

// buildCloneURL builds the clone URL from template
func (s *ModuleImporterService) buildCloneURL(req DomainImportRequest, moduleProvider *model.ModuleProvider) string {
	var cloneURLTemplate string
	if tmpl := moduleProvider.RepoCloneURLTemplate(); tmpl != nil {
		cloneURLTemplate = *tmpl
	} else if gp := moduleProvider.GitProvider(); gp != nil {
		cloneURLTemplate = gp.CloneURLTemplate
	}

	replacer := strings.NewReplacer(
		"{protocol}", "https",
		"{namespace}", req.Input.Namespace,
		"{name}", req.Input.Module,
		"{provider}", req.Input.Provider,
	)
	return replacer.Replace(cloneURLTemplate)
}

// cleanupSource cleans up temporary source files
func (s *ModuleImporterService) cleanupSource(sourcePath string) {
	if sourcePath != "" && strings.Contains(sourcePath, "tmp") {
		// Remove temporary directory
		// This would use the storage service or os.RemoveAll
		_ = os.RemoveAll(sourcePath)
	}
}


// BatchModuleImportResult represents the result of batch enhanced imports
type BatchModuleImportResult struct {
	TotalModules      int                  `json:"total_modules"`
	SuccessfulImports []ModuleImportResult `json:"successful_imports"`
	FailedImports     []ModuleImportResult `json:"failed_imports"`
	PartialSuccess    bool                 `json:"partial_success"`
	OverallSuccess    bool                 `json:"overall_success"`
	OverallDuration   time.Duration        `json:"overall_duration"`
	Timestamp         time.Time            `json:"timestamp"`
}

// validateBasicImportRequest validates the basic import request using domain DTO
func (s *ModuleImporterService) validateBasicImportRequest(input *module.ModuleVersionImportInput) error {
	// Use domain validation logic
	return input.Validate()
}
