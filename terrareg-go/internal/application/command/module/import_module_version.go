package module

import (
	"context"
	"fmt"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/module"
	moduleService "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/module/service"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/shared"
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

// Execute imports a module version from Git with full processing pipeline
func (c *ImportModuleVersionCommand) Execute(ctx context.Context, req module.ImportModuleVersionRequest) error {
	// Parse version for domain input
	var version *shared.Version
	var err error
	if req.Version != nil {
		version, err = shared.ParseVersion(*req.Version)
		if err != nil {
			return fmt.Errorf("invalid version: %w", err)
		}
	}

	// Create domain input DTO
	domainInput := module.NewModuleVersionImportInput(
		req.Namespace,
		req.Module,
		req.Provider,
		version,
		req.GitTag,
	)

	// Create import request with processing options
	// NOTE: Publishing is handled by the domain layer based on reindex mode configuration
	importReq := ModuleImportRequest{
		ImportModuleVersionRequest: req,
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
			PublishModule:           false, // Publishing handled by domain logic based on reindex mode
			ArchiveFormats: []moduleService.ArchiveFormat{
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

	// Convert application request to domain request
	domainReq := moduleService.DomainImportRequest{
		Input:              domainInput,
		ProcessingOptions:  importReq.ProcessingOptions,
		SourceType:         importReq.SourceType,
		GenerateArchives:   importReq.GenerateArchives,
		EnableSecurityScan: importReq.EnableSecurityScan,
	}

	// Use the transaction-aware import method
	result, err := c.moduleImporterService.ImportModuleVersionWithTransaction(ctx, domainReq)
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
