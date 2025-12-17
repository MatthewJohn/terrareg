package service

import (
	"context"
	"fmt"
	"time"

	"github.com/rs/zerolog"
	"gorm.io/gorm"

	configmodel "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/config/model"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/module/model"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/module/repository"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/persistence/sqldb/transaction"
)

// TransactionProcessingOrchestrator coordinates the complete module processing pipeline
// with comprehensive transaction management and rollback capabilities
type TransactionProcessingOrchestrator struct {
	// Core processing services with transaction safety
	archiveService     *ArchiveExtractionService
	terraformService   *TerraformExecutorService
	metadataService    *MetadataProcessingService
	securityService    *SecurityScanningService
	fileContentService *FileContentTransactionService
	archiveGenService  *ArchiveGenerationTransactionService

	// Foundation services
	moduleCreationWrapper *ModuleCreationWrapperService
	savepointHelper       *transaction.SavepointHelper

	// Configuration
	domainConfig *configmodel.DomainConfig
	logger       zerolog.Logger

	// Repositories
	moduleVersionRepo  repository.ModuleVersionRepository
	moduleProviderRepo repository.ModuleProviderRepository
}

// NewTransactionProcessingOrchestrator creates a new transaction processing orchestrator
func NewTransactionProcessingOrchestrator(
	archiveService *ArchiveExtractionService,
	terraformService *TerraformExecutorService,
	metadataService *MetadataProcessingService,
	securityService *SecurityScanningService,
	fileContentService *FileContentTransactionService,
	archiveGenService *ArchiveGenerationTransactionService,
	moduleCreationWrapper *ModuleCreationWrapperService,
	savepointHelper *transaction.SavepointHelper,
	domainConfig *configmodel.DomainConfig,
	logger zerolog.Logger,
	moduleVersionRepo repository.ModuleVersionRepository,
	moduleProviderRepo repository.ModuleProviderRepository,
) *TransactionProcessingOrchestrator {
	return &TransactionProcessingOrchestrator{
		archiveService:        archiveService,
		terraformService:      terraformService,
		metadataService:       metadataService,
		securityService:       securityService,
		fileContentService:    fileContentService,
		archiveGenService:     archiveGenService,
		moduleCreationWrapper: moduleCreationWrapper,
		savepointHelper:       savepointHelper,
		domainConfig:          domainConfig,
		logger:                logger,
		moduleVersionRepo:     moduleVersionRepo,
		moduleProviderRepo:    moduleProviderRepo,
	}
}

// ProcessingRequest represents a complete module processing request
type ProcessingRequest struct {
	Namespace   string
	ModuleName  string
	Provider    string
	Version     string
	GitTag      *string
	ModulePath  string // Path to extracted module files
	ArchivePath string // Path to module archive (if applicable)
	SourceType  SourceType
	Options     ProcessingOptions
}

// ProcessingOptions represents options for module processing
type ProcessingOptions struct {
	SkipArchiveExtraction   bool
	SkipTerraformProcessing bool
	SkipMetadataProcessing  bool
	SkipSecurityScanning    bool
	SkipFileContentStorage  bool
	SkipArchiveGeneration   bool
	RequiredMetadataFields  []string
	SecurityScanEnabled     bool
	FileProcessingEnabled   bool
	GenerateArchives        bool
	PublishModule           bool
	ArchiveFormats          []ArchiveFormat
}

// ProcessingResult represents the result of complete module processing
type ProcessingResult struct {
	Success             bool                    `json:"success"`
	ModuleVersion       *model.ModuleVersion    `json:"module_version,omitempty"`
	PhaseResults        map[string]*PhaseResult `json:"phase_results,omitempty"`
	OverallDuration     time.Duration           `json:"overall_duration"`
	Error               *string                 `json:"error,omitempty"`
	SavepointRolledBack bool                    `json:"savepoint_rolled_back"`
	ProcessedFiles      int                     `json:"processed_files"`
	SecurityIssues      int                     `json:"security_issues"`
	GeneratedArchives   []string                `json:"generated_archives,omitempty"`
	Timestamp           time.Time               `json:"timestamp"`
}

// PhaseResult represents the result of a processing phase
type PhaseResult struct {
	Success  bool          `json:"success"`
	Duration time.Duration `json:"duration"`
	Error    *string       `json:"error,omitempty"`
	Data     interface{}   `json:"data,omitempty"`
}

// ProcessModuleWithTransaction executes the complete module processing pipeline
// with comprehensive transaction management and rollback capabilities
func (o *TransactionProcessingOrchestrator) ProcessModuleWithTransaction(
	ctx context.Context,
	req ProcessingRequest,
) (*ProcessingResult, error) {
	startTime := time.Now()

	result := &ProcessingResult{
		Success:             false,
		PhaseResults:        make(map[string]*PhaseResult),
		SavepointRolledBack: false,
		Timestamp:           startTime,
	}

	// Log the start of processing
	o.logger.Info().
		Str("namespace", req.Namespace).
		Str("module", req.ModuleName).
		Str("provider", req.Provider).
		Str("version", req.Version).
		Str("source_type", string(req.SourceType)).
		Str("module_path", req.ModulePath).
		Msg("Starting module processing")

	// Use the smart transaction wrapper that properly handles nested transactions
	err := o.savepointHelper.WithTransaction(ctx, func(ctx context.Context, tx *gorm.DB) error {
		// Use module creation wrapper for atomic module creation and publishing
		prepareReq := PrepareModuleRequest{
			Namespace:    req.Namespace,
			ModuleName:   req.ModuleName,
			Provider:     req.Provider,
			Version:      req.Version,
			GitTag:       req.GitTag,
			SourceGitTag: req.GitTag,
		}

		// Execute within the transaction context
		err := o.moduleCreationWrapper.WithModuleCreationWrapper(ctx, prepareReq,
			func(ctx context.Context, moduleVersion *model.ModuleVersion) error {
				// Execute all processing phases
				if err := o.executeProcessingPhases(ctx, req, moduleVersion, result); err != nil {
					return fmt.Errorf("processing phases failed for module %s/%s/%s version %s: %w",
						req.Namespace, req.ModuleName, req.Provider, req.Version, err)
				}

				// Set the module version in result
				result.ModuleVersion = moduleVersion
				return nil
			})

		if err != nil {
			return fmt.Errorf("module creation failed for %s/%s/%s version %s: %w",
				req.Namespace, req.ModuleName, req.Provider, req.Version, err)
		}

		// Mark overall success
		result.Success = true
		result.ModuleVersion = result.ModuleVersion

		return nil
	})

	result.OverallDuration = time.Since(startTime)

	if err != nil {
		result.SavepointRolledBack = true
		errorMsg := err.Error()
		result.Error = &errorMsg
		o.logger.Error().
			Str("namespace", req.Namespace).
			Str("module", req.ModuleName).
			Str("provider", req.Provider).
			Str("version", req.Version).
			Dur("duration", result.OverallDuration).
			Err(err).
			Msg("Module processing failed")
		return result, nil
	}

	result.Success = true
	o.logger.Info().
		Str("namespace", req.Namespace).
		Str("module", req.ModuleName).
		Str("provider", req.Provider).
		Str("version", req.Version).
		Dur("duration", result.OverallDuration).
		Msg("Module processing completed successfully")
	return result, nil
}

// executeProcessingPhases executes all processing phases in sequence
func (o *TransactionProcessingOrchestrator) executeProcessingPhases(
	ctx context.Context,
	req ProcessingRequest,
	moduleVersion *model.ModuleVersion,
	result *ProcessingResult,
) error {
	// Phase 1: Archive Extraction (if needed)
	if !req.Options.SkipArchiveExtraction && req.ArchivePath != "" {
		phaseResult := o.executeArchiveExtractionPhase(ctx, req, moduleVersion)
		result.PhaseResults["archive_extraction"] = phaseResult
		if !phaseResult.Success {
			return fmt.Errorf("archive extraction failed: %s", *phaseResult.Error)
		}
	}

	// Phase 2: Terraform Processing
	if !req.Options.SkipTerraformProcessing && req.ModulePath != "" {
		phaseResult := o.executeTerraformProcessingPhase(ctx, req, moduleVersion)
		result.PhaseResults["terraform_processing"] = phaseResult
		if !phaseResult.Success {
			return fmt.Errorf("terraform processing failed: %s", *phaseResult.Error)
		}
	}

	// Phase 3: Metadata Processing
	if !req.Options.SkipMetadataProcessing && req.ModulePath != "" {
		phaseResult := o.executeMetadataProcessingPhase(ctx, req, moduleVersion)
		result.PhaseResults["metadata_processing"] = phaseResult
		if !phaseResult.Success {
			return fmt.Errorf("metadata processing failed: %s", *phaseResult.Error)
		}
	}

	// Phase 4: File Content Storage
	if !req.Options.SkipFileContentStorage && req.ModulePath != "" {
		phaseResult := o.executeFileContentPhase(ctx, req, moduleVersion)
		result.PhaseResults["file_content"] = phaseResult
		if !phaseResult.Success {
			return fmt.Errorf("file content processing failed: %s", *phaseResult.Error)
		}
	}

	// Phase 5: Security Scanning
	if !req.Options.SkipSecurityScanning && req.Options.SecurityScanEnabled {
		phaseResult := o.executeSecurityScanningPhase(ctx, req, moduleVersion)
		result.PhaseResults["security_scanning"] = phaseResult
		if !phaseResult.Success {
			return fmt.Errorf("security scanning failed: %s", *phaseResult.Error)
		}
	}

	// Phase 6: Archive Generation
	if !req.Options.SkipArchiveGeneration && req.Options.GenerateArchives {
		phaseResult := o.executeArchiveGenerationPhase(ctx, req, moduleVersion)
		result.PhaseResults["archive_generation"] = phaseResult
		if !phaseResult.Success {
			return fmt.Errorf("archive generation failed: %s", *phaseResult.Error)
		}
	}

	return nil
}

// executeArchiveExtractionPhase executes archive extraction with its own savepoint
func (o *TransactionProcessingOrchestrator) executeArchiveExtractionPhase(
	ctx context.Context,
	req ProcessingRequest,
	moduleVersion *model.ModuleVersion,
) *PhaseResult {
	startTime := time.Now()
	phaseResult := &PhaseResult{Success: false}

	extractionReq := ArchiveExtractionRequest{
		ModuleVersionID: moduleVersion.ID(),
		ArchivePath:     req.ArchivePath,
		SourceType:      req.SourceType,
		TransactionCtx:  ctx,
	}

	extractionResult, err := o.archiveService.ExtractAndProcessWithTransaction(ctx, extractionReq)
	phaseResult.Duration = time.Since(startTime)

	if err != nil {
		errorMsg := err.Error()
		phaseResult.Error = &errorMsg
		return phaseResult
	}

	if !extractionResult.Success {
		errorMsg := fmt.Sprintf("Archive extraction failed: %s", extractionResult.Error)
		phaseResult.Error = &errorMsg
		return phaseResult
	}

	phaseResult.Success = true
	phaseResult.Data = extractionResult
	return phaseResult
}

// executeTerraformProcessingPhase executes terraform processing with its own savepoint
func (o *TransactionProcessingOrchestrator) executeTerraformProcessingPhase(
	ctx context.Context,
	req ProcessingRequest,
	moduleVersion *model.ModuleVersion,
) *PhaseResult {
	startTime := time.Now()
	phaseResult := &PhaseResult{Success: false}

	o.logger.Debug().
		Int("module_version_id", moduleVersion.ID()).
		Str("module_path", req.ModulePath).
		Msg("Starting terraform processing phase")

	// Add validation and context for terraform processing
	if req.ModulePath == "" {
		errorMsg := "module path is empty for terraform processing"
		o.logger.Error().
			Int("module_version_id", moduleVersion.ID()).
			Str("namespace", req.Namespace).
			Str("module", req.ModuleName).
			Str("provider", req.Provider).
			Str("version", req.Version).
			Msg("Module path is empty for terraform processing")
		phaseResult.Error = &errorMsg
		phaseResult.Duration = time.Since(startTime)
		return phaseResult
	}

	// Add terraform context logging
	o.logger.Debug().
		Int("module_version_id", moduleVersion.ID()).
		Str("namespace", req.Namespace).
		Str("module", req.ModuleName).
		Str("provider", req.Provider).
		Str("version", req.Version).
		Str("module_path", req.ModulePath).
		Msg("Executing terraform operations on module")

	// Log terraform service details
	o.logger.Debug().
		Str("terraform_service_type", "TerraformExecutorService").
		Msg("Using terraform executor service")

	// Create terraform processing operations
	operations := []TerraformOperation{
		{Type: "init", Command: []string{"init", "-input=false", "-no-color"}, WorkingDir: req.ModulePath, Description: "Initialize Terraform"},
		{Type: "version", Command: []string{"version", "-no-color"}, WorkingDir: req.ModulePath, Description: "Get Terraform version"},
		{Type: "graph", Command: []string{"graph", "-no-color"}, WorkingDir: req.ModulePath, Description: "Generate dependency graph"},
	}

	terraformReq := TerraformProcessingRequest{
		ModuleVersionID: moduleVersion.ID(),
		ModulePath:      req.ModulePath,
		TransactionCtx:  ctx,
		Operations:      operations,
	}

	terraformResult, err := o.terraformService.ProcessTerraformWithTransaction(ctx, terraformReq)
	phaseResult.Duration = time.Since(startTime)

	if err != nil {
		errorMsg := err.Error()
		phaseResult.Error = &errorMsg
		return phaseResult
	}

	if !terraformResult.OverallSuccess {
		// Provide detailed error with command output and context
		var detailedError string
		var commandOutput string

		// Extract command output from the most relevant failed step
		if terraformResult.InitResult != nil && terraformResult.InitResult.Error != nil {
			commandOutput = *terraformResult.InitResult.Error
		} else if terraformResult.GraphResult != nil && terraformResult.GraphResult.Error != nil {
			commandOutput = *terraformResult.GraphResult.Error
		} else if terraformResult.VersionResult != nil && terraformResult.VersionResult.Error != nil {
			commandOutput = *terraformResult.VersionResult.Error
		}

		if commandOutput != "" {
			detailedError = fmt.Sprintf("Terraform processing failed for module version %d (%s/%s/%s): %s\nCommand output: %s",
				moduleVersion.ID(), req.Namespace, req.ModuleName, req.Provider,
				terraformResult.FailedStep, commandOutput)
		} else {
			detailedError = fmt.Sprintf("Terraform processing failed for module version %d (%s/%s/%s): %s",
				moduleVersion.ID(), req.Namespace, req.ModuleName, req.Provider,
				terraformResult.FailedStep)
		}

		// Log the terraform processing failure with details
		o.logger.Error().
			Int("module_version_id", moduleVersion.ID()).
			Str("failed_step", terraformResult.FailedStep).
			Str("namespace", req.Namespace).
			Str("module", req.ModuleName).
			Str("provider", req.Provider).
			Dur("phase_duration", phaseResult.Duration).
			Err(fmt.Errorf(detailedError)).
			Msg("Terraform processing phase failed")

		phaseResult.Error = &detailedError
		return phaseResult
	}

	phaseResult.Success = true
	phaseResult.Data = terraformResult

	o.logger.Debug().
		Int("module_version_id", moduleVersion.ID()).
		Str("module_path", req.ModulePath).
		Dur("phase_duration", phaseResult.Duration).
		Msg("Terraform processing completed successfully")

	return phaseResult
}

// executeMetadataProcessingPhase executes metadata processing with its own savepoint
func (o *TransactionProcessingOrchestrator) executeMetadataProcessingPhase(
	ctx context.Context,
	req ProcessingRequest,
	moduleVersion *model.ModuleVersion,
) *PhaseResult {
	startTime := time.Now()
	phaseResult := &PhaseResult{Success: false}

	metadataReq := MetadataProcessingRequest{
		ModuleVersionID:    moduleVersion.ID(),
		MetadataPath:       req.ModulePath,
		ModulePath:         req.ModulePath,
		TransactionCtx:     ctx,
		RequiredAttributes: req.Options.RequiredMetadataFields,
	}

	metadataResult, err := o.metadataService.ProcessMetadataWithTransaction(ctx, metadataReq)
	phaseResult.Duration = time.Since(startTime)

	if err != nil {
		errorMsg := err.Error()
		phaseResult.Error = &errorMsg
		return phaseResult
	}

	if !metadataResult.Success {
		errorMsg := fmt.Sprintf("Metadata processing failed: %s", *metadataResult.Error)
		phaseResult.Error = &errorMsg
		return phaseResult
	}

	phaseResult.Success = true
	phaseResult.Data = metadataResult
	return phaseResult
}

// executeFileContentPhase executes file content processing with its own savepoint
func (o *TransactionProcessingOrchestrator) executeFileContentPhase(
	ctx context.Context,
	req ProcessingRequest,
	moduleVersion *model.ModuleVersion,
) *PhaseResult {
	startTime := time.Now()
	phaseResult := &PhaseResult{Success: false}

	// This would need to get the archive files from the extraction result
	// For now, assume we have a way to get the file map
	archiveFiles := make(map[string]string) // Would be populated from extraction

	storageReq := FileStorageRequest{
		ModuleVersionID: moduleVersion.ID(),
		Files:           o.convertArchiveFilesToFileItems(archiveFiles),
		TransactionCtx:  ctx,
		ProcessContent:  req.Options.FileProcessingEnabled,
		ValidatePaths:   true,
	}

	storageResult, err := o.fileContentService.StoreFilesWithTransaction(ctx, storageReq)
	phaseResult.Duration = time.Since(startTime)

	if err != nil {
		errorMsg := err.Error()
		phaseResult.Error = &errorMsg
		return phaseResult
	}

	if !storageResult.Success {
		errorMsg := fmt.Sprintf("File content storage failed: %s", *storageResult.Error)
		phaseResult.Error = &errorMsg
		return phaseResult
	}

	phaseResult.Success = true
	phaseResult.Data = storageResult
	return phaseResult
}

// executeSecurityScanningPhase executes security scanning with its own savepoint
func (o *TransactionProcessingOrchestrator) executeSecurityScanningPhase(
	ctx context.Context,
	req ProcessingRequest,
	moduleVersion *model.ModuleVersion,
) *PhaseResult {
	startTime := time.Now()
	phaseResult := &PhaseResult{Success: false}

	scanReq := SecurityScanTransactionRequest{
		ModuleVersionID: moduleVersion.ID(),
		ModulePath:      req.ModulePath,
		Namespace:       req.Namespace,
		Module:          req.ModuleName,
		Provider:        req.Provider,
		Version:         req.Version,
		TransactionCtx:  ctx,
	}

	// Use the context-aware transaction method
	scanResult, err := o.securityService.ScanWithTransaction(ctx, scanReq)
	phaseResult.Duration = time.Since(startTime)

	if err != nil {
		errorMsg := err.Error()
		phaseResult.Error = &errorMsg
		return phaseResult
	}

	if !scanResult.Success {
		errorMsg := fmt.Sprintf("Security scanning failed: %s", *scanResult.Error)
		phaseResult.Error = &errorMsg
		return phaseResult
	}

	phaseResult.Success = true
	phaseResult.Data = scanResult
	return phaseResult
}

// executeArchiveGenerationPhase executes archive generation with its own savepoint
func (o *TransactionProcessingOrchestrator) executeArchiveGenerationPhase(
	ctx context.Context,
	req ProcessingRequest,
	moduleVersion *model.ModuleVersion,
) *PhaseResult {
	startTime := time.Now()
	phaseResult := &PhaseResult{Success: false}

	// Get git clone URL from module provider (may be empty if not externally hosted)
	var gitCloneURL string
	if moduleVersion.ModuleProvider() != nil && moduleVersion.ModuleProvider().GetGitCloneURL() != nil {
		gitCloneURL = *moduleVersion.ModuleProvider().GetGitCloneURL()
	}

	// Integrate with the archive generation service
	genReq := ArchiveGenerationRequest{
		ModuleVersionID:               moduleVersion.ID(),
		SourcePath:                    req.ModulePath,
		Formats:                       []ArchiveFormat{ArchiveFormatZIP, ArchiveFormatTarGz},
		TransactionCtx:                ctx,
		GitCloneURL:                   gitCloneURL,
		DeleteExternallyHostedArtifacts: o.domainConfig.DeleteExternallyHostedArtifacts,
	}

	genResult, err := o.archiveGenService.GenerateArchivesWithTransaction(ctx, genReq)
	phaseResult.Duration = time.Since(startTime)

	if err != nil {
		errorMsg := fmt.Sprintf("Archive generation failed: %v", err)
		phaseResult.Success = false
		phaseResult.Error = &errorMsg
		return phaseResult
	}

	if !genResult.Success {
		// Check if archives were skipped (this is actually success for externally hosted modules)
		if genResult.SkippedReason != "" {
			phaseResult.Success = true
			phaseResult.Data = struct {
				SkippedReason string `json:"skipped_reason"`
			}{
				SkippedReason: genResult.SkippedReason,
			}
			return phaseResult
		}

		errorMsg := fmt.Sprintf("Archive generation failed: %s", *genResult.Error)
		phaseResult.Success = false
		phaseResult.Error = &errorMsg
		return phaseResult
	}

	phaseResult.Success = true
	generatedFiles := make([]string, len(genResult.GeneratedArchives))
	for i, archive := range genResult.GeneratedArchives {
		generatedFiles[i] = archive.Path
	}

	phaseResult.Data = struct {
		GeneratedArchives []string `json:"generated_archives"`
		SkippedReason     string   `json:"skipped_reason,omitempty"`
		TotalArchiveSize  int64    `json:"total_archive_size"`
		SourceFilesCount  int      `json:"source_files_count"`
	}{
		GeneratedArchives: generatedFiles,
		SkippedReason:     genResult.SkippedReason,
		TotalArchiveSize:  genResult.TotalArchiveSize,
		SourceFilesCount:  genResult.SourceFilesCount,
	}

	return phaseResult
}

// ProcessBatchModules processes multiple modules with individual transaction isolation
func (o *TransactionProcessingOrchestrator) ProcessBatchModules(
	ctx context.Context,
	requests []ProcessingRequest,
) (*BatchProcessingResult, error) {
	startTime := time.Now()

	result := &BatchProcessingResult{
		TotalModules:      len(requests),
		SuccessfulModules: []ProcessingResult{},
		FailedModules:     []ProcessingResult{},
		PartialSuccess:    false,
		OverallSuccess:    true,
		Timestamp:         startTime,
	}

	for _, req := range requests {
		moduleResult, err := o.ProcessModuleWithTransaction(ctx, req)
		if err != nil {
			// This should rarely happen since we handle errors within ProcessModuleWithTransaction
			errorResult := ProcessingResult{
				Success:             false,
				Error:               func() *string { e := err.Error(); return &e }(),
				Timestamp:           time.Now(),
				SavepointRolledBack: true,
			}
			result.FailedModules = append(result.FailedModules, errorResult)
			result.OverallSuccess = false
			result.PartialSuccess = true
			continue
		}

		if moduleResult.Success {
			result.SuccessfulModules = append(result.SuccessfulModules, *moduleResult)
		} else {
			result.FailedModules = append(result.FailedModules, *moduleResult)
			result.OverallSuccess = false
			result.PartialSuccess = true
		}
	}

	result.OverallDuration = time.Since(startTime)

	// If there were no failures, set partial success to false
	if len(result.FailedModules) == 0 {
		result.PartialSuccess = false
	}

	return result, nil
}

// BatchProcessingResult represents the result of batch module processing
type BatchProcessingResult struct {
	TotalModules      int                `json:"total_modules"`
	SuccessfulModules []ProcessingResult `json:"successful_modules"`
	FailedModules     []ProcessingResult `json:"failed_modules"`
	PartialSuccess    bool               `json:"partial_success"`
	OverallSuccess    bool               `json:"overall_success"`
	OverallDuration   time.Duration      `json:"overall_duration"`
	Timestamp         time.Time          `json:"timestamp"`
}

// convertArchiveFilesToFileItems converts archive file map to file content items
func (o *TransactionProcessingOrchestrator) convertArchiveFilesToFileItems(
	archiveFiles map[string]string,
) []FileContentItem {
	var files []FileContentItem
	for path, content := range archiveFiles {
		files = append(files, FileContentItem{
			Path:    path,
			Content: content,
			Type:    "source",
		})
	}
	return files
}

// GetProcessingStatus returns the current status of processing for a module version
func (o *TransactionProcessingOrchestrator) GetProcessingStatus(
	ctx context.Context,
	moduleVersionID int,
) (*ProcessingStatus, error) {
	moduleVersion, err := o.moduleVersionRepo.FindByID(ctx, moduleVersionID)
	if err != nil {
		return nil, fmt.Errorf("failed to find module version: %w", err)
	}

	if moduleVersion == nil {
		return nil, fmt.Errorf("module version not found")
	}

	details := moduleVersion.Details()
	status := &ProcessingStatus{
		ModuleVersionID: moduleVersionID,
		Version:         moduleVersion.Version().String(),
		Published:       moduleVersion.IsPublished(),
		Timestamp:       time.Now(),
	}

	// Check various processing stages based on module details
	if details != nil {
		status.HasReadme = details.HasReadme()
		status.HasTerraformDocs = details.HasTerraformDocs()
		status.HasTerraformGraph = details.HasTerraformGraph()
		status.HasTfsecResults = details.HasTfsec()
		status.HasInfracostResults = details.HasInfracost()
		status.HasTerraformModules = details.HasTerraformModules()
		status.TerraformVersion = details.TerraformVersion()
	}

	return status, nil
}

// ProcessingStatus represents the processing status of a module version
type ProcessingStatus struct {
	ModuleVersionID     int       `json:"module_version_id"`
	Version             string    `json:"version"`
	Published           bool      `json:"published"`
	HasReadme           bool      `json:"has_readme"`
	HasTerraformDocs    bool      `json:"has_terraform_docs"`
	HasTerraformGraph   bool      `json:"has_terraform_graph"`
	HasTfsecResults     bool      `json:"has_tfsec_results"`
	HasInfracostResults bool      `json:"has_infracost_results"`
	HasTerraformModules bool      `json:"has_terraform_modules"`
	TerraformVersion    string    `json:"terraform_version"`
	Timestamp           time.Time `json:"timestamp"`
}
