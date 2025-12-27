package module

import (
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/module"
	domainService "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/module/service"
)

// ModuleImportRequest represents a comprehensive module import request with all processing options
// This is an application-layer construct that coordinates domain services
type ModuleImportRequest struct {
	module.ImportModuleVersionRequest

	// Application-layer processing options
	ProcessingOptions domainService.ProcessingOptions

	// Source information
	SourcePath  string // Path to source files (if already extracted)
	ArchivePath string // Path to archive file (if applicable)
	SourceType  string // "git", "archive", "path"

	// Archive generation options
	GenerateArchives bool
	ArchiveFormats   []domainService.ArchiveFormat

	// Security and analysis options
	EnableSecurityScan bool
	EnableInfracost    bool
	EnableExamples     bool

	// Transaction options (application concern)
	UseTransaction bool
	EnableRollback bool
	SavepointName  string
}

// ModuleImportResult represents the result of module import
// This is an application-layer response type
type ModuleImportResult struct {
	Success             bool                              `json:"success"`
	ModuleVersionID     *int                              `json:"module_version_id,omitempty"`
	Version             string                            `json:"version"`
	ProcessingResult    *domainService.ProcessingResult   `json:"processing_result,omitempty"`
	GeneratedArchives   []domainService.GeneratedArchive   `json:"generated_archives,omitempty"`
	SecurityResults     *domainService.SecurityScanResponse `json:"security_results,omitempty"`
	Error               *string                           `json:"error,omitempty"`
	SavepointRolledBack bool                              `json:"savepoint_rolled_back"`
	Timestamp           string                            `json:"timestamp"` // Application-layer formatting
}

// BatchModuleImportRequest represents a batch import request
type BatchModuleImportRequest struct {
	Requests []ModuleImportRequest
	Options  BatchImportOptions
}

// BatchImportOptions represents options for batch imports
type BatchImportOptions struct {
	ContinueOnError    bool
	ParallelProcessing bool
	MaxConcurrency     int
}

// BatchModuleImportResult represents the result of batch imports
type BatchModuleImportResult struct {
	TotalModules      int                        `json:"total_modules"`
	SuccessfulImports []ModuleImportResult       `json:"successful_imports"`
	FailedImports     []ModuleImportResult       `json:"failed_imports"`
	PartialSuccess    bool                       `json:"partial_success"`
	OverallSuccess    bool                       `json:"overall_success"`
	FailureSummary    string                     `json:"failure_summary,omitempty"`
	TotalDuration     string                     `json:"total_duration"` // Application-layer formatting
	Timestamp         string                     `json:"timestamp"`     // Application-layer formatting
}