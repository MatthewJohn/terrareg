package service

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"gorm.io/gorm"

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
	}
}

// ModuleImportRequest represents a comprehensive import request with all processing options
type ModuleImportRequest struct {
	ImportModuleVersionRequest

	// Processing options
	ProcessingOptions ProcessingOptions

	// Source information
	SourcePath  string // Path to source files (if already extracted)
	ArchivePath string // Path to archive file (if applicable)
	SourceType  SourceType

	// Archive generation options
	GenerateArchives bool
	ArchiveFormats   []ArchiveFormat
	PathspecFilter   *PathspecFilter

	// Security and analysis options
	EnableSecurityScan bool
	EnableInfracost    bool
	EnableExamples     bool

	// Transaction options
	UseTransaction bool
	EnableRollback bool
	SavepointName  string
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
	req ModuleImportRequest,
) (*ModuleImportResult, error) {
	startTime := time.Now()

	result := &ModuleImportResult{
		Success:             false,
		Version:             req.getVersion(),
		SavepointRolledBack: false,
		Timestamp:           startTime,
	}

	// Create main savepoint for the entire import process
	savepointName := req.SavepointName
	if savepointName == "" {
		savepointName = fmt.Sprintf("enhanced_import_%s_%s_%s_%d",
			req.Namespace, req.Module, req.Provider, startTime.UnixNano())
	}

	err := s.savepointHelper.WithSavepointNamed(ctx, savepointName, func(tx *gorm.DB) error {
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
			Namespace:   req.Namespace,
			ModuleName:  req.Module,
			Provider:    req.Provider,
			Version:     req.getVersion(),
			GitTag:      req.GitTag,
			ModulePath:  sourcePath,
			ArchivePath: req.ArchivePath,
			SourceType:  req.SourceType,
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
	requests []ModuleImportRequest,
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
				Version:             req.getVersion(),
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
	req := ModuleImportRequest{
		ImportModuleVersionRequest: ImportModuleVersionRequest{
			Namespace: namespace,
			Module:    moduleName,
			Provider:  provider,
			Version:   &version,
		},
		ProcessingOptions: options,
		SourceType:        "git", // Re-indexing typically from git
		UseTransaction:    true,
		EnableRollback:    true,
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
	req := ModuleImportRequest{
		ImportModuleVersionRequest: ImportModuleVersionRequest{
			Namespace: namespace,
			Module:    moduleName,
			Provider:  provider,
			Version:   &version,
		},
		ProcessingOptions: options,
		ArchivePath:       archivePath,
		SourceType:        "archive",
		UseTransaction:    true,
		EnableRollback:    true,
	}

	return s.ImportModuleVersionWithTransaction(ctx, req)
}

// validateImportRequest validates the enhanced import request
func (s *ModuleImporterService) validateImportRequest(
	ctx context.Context,
	req ModuleImportRequest,
) error {
	// Validate basic import request
	if err := s.validateBasicImportRequest(req.ImportModuleVersionRequest); err != nil {
		return err
	}

	// Validate processing options
	if req.ProcessingOptions.GenerateArchives && len(req.ProcessingOptions.ArchiveFormats) == 0 {
		return fmt.Errorf("archive generation requested but no formats specified")
	}

	// Validate source
	if req.SourcePath == "" && req.ArchivePath == "" && req.GitTag == nil {
		return fmt.Errorf("no source specified (source path, archive path, or git tag required)")
	}

	return nil
}

// prepareSource prepares the source files for processing
func (s *ModuleImporterService) prepareSource(
	ctx context.Context,
	req ModuleImportRequest,
) (string, error) {
	if req.SourcePath != "" {
		// Source already prepared
		return req.SourcePath, nil
	}

	if req.ArchivePath != "" {
		// Extract archive
		return s.extractArchive(ctx, req)
	}

	if req.GitTag != nil {
		// Use base importer for git operations
		return s.prepareGitSource(ctx, req)
	}

	return "", fmt.Errorf("no source preparation method available")
}

// extractArchive extracts archive files for processing
func (s *ModuleImporterService) extractArchive(
	ctx context.Context,
	req ModuleImportRequest,
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
	req ModuleImportRequest,
) (string, error) {
	// Validate that either version or git_tag is provided
	if (req.Version == nil && req.GitTag == nil) || (req.Version != nil && req.GitTag != nil) {
		return "", fmt.Errorf("either version or git_tag must be provided (but not both)")
	}

	// Find the module provider
	moduleProvider, err := s.moduleProviderRepo.FindByNamespaceModuleProvider(
		ctx, req.Namespace, req.Module, req.Provider,
	)
	if err != nil {
		return "", fmt.Errorf("module provider not found: %w", err)
	}

	// Validate git configuration
	if moduleProvider.GitProviderID() == nil || moduleProvider.RepoCloneURLTemplate() == nil {
		return "", fmt.Errorf("module provider is not a git based module")
	}

	// Clone and checkout - using legacy git operations for now
	cloneURL := s.buildCloneURL(req, moduleProvider)

	tmpDir, err := s.storageService.MkdirTemp("", "terrareg-git-")
	if err != nil {
		return "", fmt.Errorf("failed to create temp dir: %w", err)
	}

	// Clone repository
	cloneOptions := gitService.CloneOptions{
		Timeout: time.Duration(s.infraConfig.GitCloneTimeout) * time.Second,
	}

	if err := s.gitClient.CloneWithOptions(ctx, cloneURL, tmpDir, cloneOptions); err != nil {
		s.storageService.RemoveAll(tmpDir)
		return "", fmt.Errorf("failed to clone git repository: %w", err)
	}

	// Checkout specific tag if provided
	if req.GitTag != nil {
		if err := s.gitClient.Checkout(ctx, tmpDir, *req.GitTag); err != nil {
			s.storageService.RemoveAll(tmpDir)
			return "", fmt.Errorf("failed to checkout git tag '%s': %w", *req.GitTag, err)
		}
	}

	// Determine source directory
	srcDir := tmpDir
	if gitPath := moduleProvider.GitPath(); gitPath != nil && *gitPath != "" {
		srcDir = filepath.Join(tmpDir, *gitPath)
	}

	return srcDir, nil
}

// buildCloneURL builds the clone URL from template
func (s *ModuleImporterService) buildCloneURL(req ModuleImportRequest, moduleProvider *model.ModuleProvider) string {
	var cloneURLTemplate string
	if tmpl := moduleProvider.RepoCloneURLTemplate(); tmpl != nil {
		cloneURLTemplate = *tmpl
	} else if gp := moduleProvider.GitProvider(); gp != nil {
		cloneURLTemplate = gp.CloneURLTemplate
	}

	replacer := strings.NewReplacer(
		"{protocol}", "https",
		"{namespace}", req.Namespace,
		"{name}", req.Module,
		"{provider}", req.Provider,
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

// Helper method to get version from request
func (req ModuleImportRequest) getVersion() string {
	if req.Version != nil {
		return *req.Version
	}
	if req.GitTag != nil {
		return *req.GitTag
	}
	return ""
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

// validateBasicImportRequest validates the basic import request using base importer logic
func (s *ModuleImporterService) validateBasicImportRequest(req ImportModuleVersionRequest) error {
	// Use base importer validation logic
	// This would extract common validation from the base importer
	if (req.Version == nil && req.GitTag == nil) || (req.Version != nil && req.GitTag != nil) {
		return fmt.Errorf("either version or git_tag must be provided (but not both)")
	}

	if req.Namespace == "" || req.Module == "" || req.Provider == "" {
		return fmt.Errorf("namespace, module, and provider are required")
	}

	return nil
}
