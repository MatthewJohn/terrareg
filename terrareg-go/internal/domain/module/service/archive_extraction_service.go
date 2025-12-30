package service

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"gorm.io/gorm"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/persistence/sqldb/transaction"
)

// ArchiveType represents the type of archive
type ArchiveType int

const (
	ArchiveTypeZIP ArchiveType = iota
	ArchiveTypeTarGZ
)

type SourceType string

const (
	SourceTypeArchive = "archive"
	SourceTypeGit     = "git"
)

// ArchiveExtractionRequest represents a request to extract and process an archive
type ArchiveExtractionRequest struct {
	ModuleVersionID int
	ArchivePath     string
	SourceType      SourceType
	TargetDirectory string
	TransactionCtx  context.Context
}

// ArchiveExtractionResult represents the result of archive extraction and processing
type ArchiveExtractionResult struct {
	Success        bool          `json:"success"`
	FilesExtracted int           `json:"files_extracted"`
	Duration       time.Duration `json:"duration"`
	Error          *string       `json:"error,omitempty"`
}

// ArchiveRequest represents a single archive processing request
type ArchiveRequest struct {
	ID string
	ArchiveExtractionRequest
}

// BatchArchiveResult represents the result of processing multiple archives
type BatchArchiveResult struct {
	TotalArchives      int                       `json:"total_archives"`
	SuccessfulArchives []ArchiveExtractionResult `json:"successful_archives"`
	FailedArchives     []ArchiveExtractionResult `json:"failed_archives"`
	PartialSuccess     bool                      `json:"partial_success"`
	OverallSuccess     bool                      `json:"overall_success"`
	Duration           time.Duration             `json:"duration"`
}

// ArchiveProcessor interface for archive processing operations
type ArchiveProcessor interface {
	ExtractArchive(ctx context.Context, archivePath string, targetDir string, archiveType ArchiveType) error
	DetectArchiveType(archivePath string) (ArchiveType, error)
	ValidateArchive(archivePath string) error
}

// ArchiveExtractionService handles archive extraction with transaction safety
type ArchiveExtractionService struct {
	archiveProcessor ArchiveProcessor
	savepointHelper  *transaction.SavepointHelper
}

// NewArchiveExtractionService creates a new archive extraction service
func NewArchiveExtractionService(
	archiveProcessor ArchiveProcessor,
	savepointHelper *transaction.SavepointHelper,
) *ArchiveExtractionService {
	return &ArchiveExtractionService{
		archiveProcessor: archiveProcessor,
		savepointHelper:  savepointHelper,
	}
}

// ExtractAndProcessWithTransaction extracts an archive and processes it within a transaction
func (s *ArchiveExtractionService) ExtractAndProcessWithTransaction(
	ctx context.Context,
	req ArchiveExtractionRequest,
) (*ArchiveExtractionResult, error) {
	startTime := time.Now()
	result := &ArchiveExtractionResult{
		Success:  false,
		Duration: 0,
	}

	// Create savepoint for this archive operation

	err := s.savepointHelper.WithTransaction(ctx, func(ctx context.Context, tx *gorm.DB) error {
		// Detect archive type
		archiveType, err := s.archiveProcessor.DetectArchiveType(req.ArchivePath)
		if err != nil {
			return fmt.Errorf("failed to detect archive type: %w", err)
		}

		// Validate archive
		if err := s.archiveProcessor.ValidateArchive(req.ArchivePath); err != nil {
			return fmt.Errorf("archive validation failed: %w", err)
		}

		// Extract archive
		if err := s.archiveProcessor.ExtractArchive(ctx, req.ArchivePath, req.TargetDirectory, archiveType); err != nil {
			return fmt.Errorf("archive extraction failed: %w", err)
		}

		// Count extracted files
		fileCount, err := s.countExtractedFiles(req.TargetDirectory)
		if err != nil {
			return fmt.Errorf("failed to count extracted files: %w", err)
		}

		result.Success = true
		result.FilesExtracted = fileCount
		result.Duration = time.Since(startTime)

		return nil
	})

	result.Duration = time.Since(startTime)

	if err != nil {
		errorMsg := err.Error()
		result.Error = &errorMsg
		return result, nil
	}

	return result, nil
}

// ProcessBatchArchives processes multiple archives with individual rollback capability
func (s *ArchiveExtractionService) ProcessBatchArchives(
	ctx context.Context,
	archives []ArchiveRequest,
) (*BatchArchiveResult, error) {
	batchResult := &BatchArchiveResult{
		TotalArchives:      len(archives),
		SuccessfulArchives: []ArchiveExtractionResult{},
		FailedArchives:     []ArchiveExtractionResult{},
		PartialSuccess:     false,
		OverallSuccess:     true,
		Duration:           0,
	}

	startTime := time.Now()

	for _, archiveReq := range archives {
		archiveResult := s.extractArchiveWithSavepoint(ctx, archiveReq.ArchiveExtractionRequest)

		if archiveResult.Success {
			batchResult.SuccessfulArchives = append(batchResult.SuccessfulArchives, *archiveResult)
		} else {
			batchResult.FailedArchives = append(batchResult.FailedArchives, *archiveResult)
			batchResult.OverallSuccess = false
		}
	}

	batchResult.Duration = time.Since(startTime)
	batchResult.PartialSuccess = len(batchResult.SuccessfulArchives) > 0 && len(batchResult.FailedArchives) > 0

	return batchResult, nil
}

// extractArchiveWithSavepoint extracts a single archive with savepoint isolation
func (s *ArchiveExtractionService) extractArchiveWithSavepoint(
	ctx context.Context,
	req ArchiveExtractionRequest,
) *ArchiveExtractionResult {
	startTime := time.Now()
	result := &ArchiveExtractionResult{
		Success:  false,
		Duration: 0,
	}

	err := s.savepointHelper.WithTransaction(ctx, func(ctx context.Context, tx *gorm.DB) error {
		return s.extractArchive(ctx, req)
	})

	result.Duration = time.Since(startTime)

	if err != nil {
		errorMsg := err.Error()
		result.Error = &errorMsg
		return result
	}

	result.Success = true
	return result
}

// extractArchive performs the actual archive extraction
func (s *ArchiveExtractionService) extractArchive(
	ctx context.Context,
	req ArchiveExtractionRequest,
) error {
	// Detect archive type
	archiveType, err := s.archiveProcessor.DetectArchiveType(req.ArchivePath)
	if err != nil {
		return fmt.Errorf("failed to detect archive type: %w", err)
	}

	// Validate archive
	if err := s.archiveProcessor.ValidateArchive(req.ArchivePath); err != nil {
		return fmt.Errorf("archive validation failed: %w", err)
	}

	// Extract archive
	if err := s.archiveProcessor.ExtractArchive(ctx, req.ArchivePath, req.TargetDirectory, archiveType); err != nil {
		return fmt.Errorf("archive extraction failed: %w", err)
	}

	return nil
}

// countExtractedFiles counts the number of files in an extracted directory
func (s *ArchiveExtractionService) countExtractedFiles(directory string) (int, error) {
	var count int

	err := filepath.Walk(directory, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		count++
		return nil
	})

	return count, err
}
