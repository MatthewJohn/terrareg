package module

import (
	"context"
	"fmt"

	moduleService "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/module/service"
)

// ImportModuleVersionCommand handles importing module versions from Git
type ImportModuleVersionCommand struct {
	moduleImporterService *moduleService.ModuleImporterService
}

// NewImportModuleVersionCommand creates a new command
func NewImportModuleVersionCommand(
	moduleImporterService *moduleService.ModuleImporterService,
) *ImportModuleVersionCommand {
	return &ImportModuleVersionCommand{
		moduleImporterService: moduleImporterService,
	}
}

// ImportModuleVersionRequest represents the import request
type ImportModuleVersionRequest struct {
	Namespace string
	Module    string
	Provider  string
	Version   *string // Optional - derived from git tag if not provided
	GitTag    *string // Optional - conflicts with Version
}

// Execute imports a module version from Git with full processing pipeline
func (c *ImportModuleVersionCommand) Execute(ctx context.Context, req ImportModuleVersionRequest) error {
	// Create comprehensive import request with all processing options
	importReq := moduleService.ModuleImportRequest{
		ImportModuleVersionRequest: moduleService.ImportModuleVersionRequest{
			Namespace: req.Namespace,
			Module:    req.Module,
			Provider:  req.Provider,
			Version:   req.Version,
			GitTag:    req.GitTag,
		},
		ProcessingOptions: moduleService.ProcessingOptions{
			SkipArchiveExtraction:   false,
			SkipTerraformProcessing: false,
			SkipMetadataProcessing:  false,
			SkipSecurityScanning:    false,
			SkipFileContentStorage:  false,
			SkipArchiveGeneration:   false,
			SecurityScanEnabled:     true,
			FileProcessingEnabled:   true,
			GenerateArchives:        true,
			PublishModule:          true,
			ArchiveFormats:        []moduleService.ArchiveFormat{
				moduleService.ArchiveFormatZIP,
				moduleService.ArchiveFormatTarGz,
			},
		},
		SourceType:         "git",
		UseTransaction:     true,
		EnableRollback:     true,
		EnableSecurityScan: true,
		GenerateArchives:   true,
	}

	// Use the transaction-aware import method
	result, err := c.moduleImporterService.ImportModuleVersionWithTransaction(ctx, importReq)
	if err != nil {
		return err
	}

	if !result.Success {
		if result.Error != nil {
			return fmt.Errorf("module import failed: %s", *result.Error)
		}
		return fmt.Errorf("module import failed")
	}

	return nil
}
