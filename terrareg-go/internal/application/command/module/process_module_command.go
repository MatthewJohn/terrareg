package module

import (
	"context"
	"fmt"
	"io"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/module/service"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/shared"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/shared/types"
)

// SourcePreparationServiceInterface defines the interface for source preparation operations
type SourcePreparationServiceInterface interface {
	PrepareFromUpload(ctx context.Context, req service.PrepareFromUploadRequest) (*service.PreparedSource, error)
	PrepareFromGit(ctx context.Context, req service.PrepareFromGitRequest) (*service.PreparedSource, error)
	PrepareFromArchive(ctx context.Context, req service.PrepareFromArchiveRequest) (*service.PreparedSource, error)
}

// ProcessingOrchestratorInterface defines the interface for processing orchestration
type ProcessingOrchestratorInterface interface {
	ProcessModuleWithTransaction(ctx context.Context, req service.ProcessingRequest) (*service.ProcessingResult, error)
}

// ProcessModuleCommand handles all module processing with unified pipeline
type ProcessModuleCommand struct {
	sourcePrepService      SourcePreparationServiceInterface
	processingOrchestrator ProcessingOrchestratorInterface
}

// ProcessModuleRequest represents a unified module processing request
type ProcessModuleRequest struct {
	// Common fields
	Namespace string `json:"namespace"`
	Module    string `json:"module"`
	Provider  string `json:"provider"`
	Version   string `json:"version"`

	// Source type (one of these)
	UploadSource io.Reader // For file uploads
	UploadSize   int64     // For file uploads
	GitTag       *string   // For git-based imports
	ArchivePath  string    // For archive files

	// Processing options (all optional, with sensible defaults)
	Options service.ProcessingOptions `json:"options"`
}

// NewProcessModuleCommand creates a new ProcessModuleCommand
func NewProcessModuleCommand(
	sourcePrepService SourcePreparationServiceInterface,
	processingOrchestrator ProcessingOrchestratorInterface,
) *ProcessModuleCommand {
	return &ProcessModuleCommand{
		sourcePrepService:      sourcePrepService,
		processingOrchestrator: processingOrchestrator,
	}
}

// Execute processes a module using the unified pipeline
func (c *ProcessModuleCommand) Execute(ctx context.Context, req ProcessModuleRequest) error {
	// 1. Parse version
	version, err := shared.ParseVersion(req.Version)
	if err != nil {
		return fmt.Errorf("invalid version: %w", err)
	}
	_ = version // Use validated version

	// 2. Determine source type and prepare source
	var preparedSource *service.PreparedSource

	if req.UploadSource != nil {
		// File upload
		preparedSource, err = c.sourcePrepService.PrepareFromUpload(ctx,
			service.PrepareFromUploadRequest{
				Namespace:  req.Namespace,
				Module:     req.Module,
				Provider:   req.Provider,
				Version:    req.Version,
				Source:     req.UploadSource,
				SourceSize: req.UploadSize,
			})
		if err != nil {
			return fmt.Errorf("source preparation failed: %w", err)
		}
	} else if req.GitTag != nil {
		// Git-based import
		preparedSource, err = c.sourcePrepService.PrepareFromGit(ctx,
			service.PrepareFromGitRequest{
				Namespace: types.NamespaceName(req.Namespace),
				Module:    types.ModuleName(req.Module),
				Provider:  types.ModuleProviderName(req.Provider),
				Version:   req.Version,
				GitTag:    req.GitTag,
			})
		if err != nil {
			return fmt.Errorf("source preparation failed: %w", err)
		}
	} else if req.ArchivePath != "" {
		// Archive file
		preparedSource, err = c.sourcePrepService.PrepareFromArchive(ctx,
			service.PrepareFromArchiveRequest{
				Namespace:   types.NamespaceName(req.Namespace),
				Module:      types.ModuleName(req.Module),
				Provider:    types.ModuleProviderName(req.Provider),
				Version:     req.Version,
				ArchivePath: req.ArchivePath,
			})
		if err != nil {
			return fmt.Errorf("source preparation failed: %w", err)
		}
	} else {
		return fmt.Errorf("no source specified")
	}

	// Ensure cleanup is always called
	defer preparedSource.Cleanup()

	// 3. Execute processing pipeline
	processingReq := service.ProcessingRequest{
		Namespace:  types.NamespaceName(req.Namespace),
		ModuleName: types.ModuleName(req.Module),
		Provider:   types.ModuleProviderName(req.Provider),
		Version:    req.Version,
		GitTag:     req.GitTag,
		CommitSHA:  preparedSource.CommitSHA,
		ModulePath: preparedSource.SourcePath,
		SourceType: preparedSource.SourceType,
		Options:    req.Options,
	}

	result, err := c.processingOrchestrator.ProcessModuleWithTransaction(ctx, processingReq)
	if err != nil {
		return fmt.Errorf("processing failed: %w", err)
	}

	if !result.Success {
		if result.Error != nil {
			return fmt.Errorf("processing failed: %s", *result.Error)
		}
		return fmt.Errorf("processing failed")
	}

	return nil
}
