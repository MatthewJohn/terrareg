package service

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"gorm.io/gorm"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/module/model"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/module/repository"
	storageService "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/storage/service"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/persistence/sqldb/transaction"
)

// FileContentTransactionService handles file content operations with transaction safety
// and rollback capabilities for partial failures during file processing
// Note: File content is stored in database, not filesystem storage
type FileContentTransactionService struct {
	moduleVersionFileRepo model.ModuleVersionFileRepository
	moduleVersionRepo     repository.ModuleVersionRepository
	fileProcessingService model.FileProcessingService
	pathBuilder           storageService.PathBuilder
	savepointHelper       *transaction.SavepointHelper
}

// NewFileContentTransactionService creates a new file content transaction service
func NewFileContentTransactionService(
	moduleVersionFileRepo model.ModuleVersionFileRepository,
	moduleVersionRepo repository.ModuleVersionRepository,
	fileProcessingService model.FileProcessingService,
	pathBuilder storageService.PathBuilder,
	savepointHelper *transaction.SavepointHelper,
) *FileContentTransactionService {
	return &FileContentTransactionService{
		moduleVersionFileRepo: moduleVersionFileRepo,
		moduleVersionRepo:     moduleVersionRepo,
		fileProcessingService: fileProcessingService,
		pathBuilder:           pathBuilder,
		savepointHelper:       savepointHelper,
	}
}

// FileStorageRequest represents a request to store files with transaction safety
type FileStorageRequest struct {
	ModuleVersionID int
	Files           []FileContentItem
	TransactionCtx  context.Context
	SavepointName   string
	ProcessContent  bool // Whether to process content (markdown, etc.)
	ValidatePaths   bool // Whether to validate file paths
}

// FileContentItem represents a file to be stored
type FileContentItem struct {
	Path     string
	Content  string
	Encoding string // "utf-8", "base64", etc.
	Type     string // "source", "example", "docs", etc.
}

// FileStorageResult represents the result of file storage operation
type FileStorageResult struct {
	Success             bool             `json:"success"`
	StoredFiles         []StoredFileInfo `json:"stored_files,omitempty"`
	FailedFiles         []FailedFileInfo `json:"failed_files,omitempty"`
	Error               *string          `json:"error,omitempty"`
	StorageDuration     time.Duration    `json:"storage_duration"`
	Timestamp           time.Time        `json:"timestamp"`
	SavepointRolledBack bool             `json:"savepoint_rolled_back"`
	TotalFiles          int              `json:"total_files"`
	ProcessedFiles      int              `json:"processed_files"`
}

// StoredFileInfo represents information about a successfully stored file
type StoredFileInfo struct {
	Path         string `json:"path"`
	FileID       int    `json:"file_id"`
	Size         int    `json:"size"`
	ContentType  string `json:"content_type"`
	Processed    bool   `json:"processed"`
	ProcessingMs int64  `json:"processing_ms"`
}

// FailedFileInfo represents information about a failed file storage
type FailedFileInfo struct {
	Path   string `json:"path"`
	Error  string `json:"error"`
	Reason string `json:"reason"` // "validation", "storage", "processing"
}

// ExampleFile represents an example file to be processed
type ExampleFile struct {
	Path          string `json:"path"`
	Content       string `json:"content"`
	IsMainExample bool   `json:"is_main_example"`
	Directory     string `json:"directory"`
}

// ExampleProcessingResult represents the result of example file processing
type ExampleProcessingResult struct {
	Success            bool               `json:"success"`
	ProcessedExamples  []ProcessedExample `json:"processed_examples,omitempty"`
	FailedExamples     []FailedExample    `json:"failed_examples,omitempty"`
	Error              *string            `json:"error,omitempty"`
	ProcessingDuration time.Duration      `json:"processing_duration"`
	Timestamp          time.Time          `json:"timestamp"`
}

// ProcessedExample represents a successfully processed example
type ProcessedExample struct {
	Path         string                `json:"path"`
	FileID       int                   `json:"file_id"`
	IsMain       bool                  `json:"is_main"`
	Directory    string                `json:"directory"`
	HasTerraform bool                  `json:"has_terraform"`
	SecurityScan *SecurityScanResponse `json:"security_scan,omitempty"`
}

// FailedExample represents a failed example processing
type FailedExample struct {
	Path  string `json:"path"`
	Error string `json:"error"`
}

// StoreFilesWithTransaction stores multiple files with transaction safety and rollback capability
func (s *FileContentTransactionService) StoreFilesWithTransaction(
	ctx context.Context,
	req FileStorageRequest,
) (*FileStorageResult, error) {
	startTime := time.Now()
	result := &FileStorageResult{
		Success:             false,
		StoredFiles:         []StoredFileInfo{},
		FailedFiles:         []FailedFileInfo{},
		SavepointRolledBack: false,
		Timestamp:           startTime,
		TotalFiles:          len(req.Files),
	}

	// Use provided savepoint name or create new one
	savepointName := req.SavepointName
	if savepointName == "" {
		savepointName = fmt.Sprintf("file_storage_%d_%d", req.ModuleVersionID, startTime.UnixNano())
	}

	err := s.savepointHelper.WithSmartSavepointOrTransaction(ctx, savepointName, func(tx *gorm.DB) error {
		for _, fileItem := range req.Files {
			fileStart := time.Now()

			// Validate file path if requested
			if req.ValidatePaths {
				if err := s.validateFilePath(fileItem.Path); err != nil {
					result.FailedFiles = append(result.FailedFiles, FailedFileInfo{
						Path:   fileItem.Path,
						Error:  err.Error(),
						Reason: "validation",
					})
					continue
				}
			}

			// Process content if requested
			processedContent := fileItem.Content
			contentProcessed := false
			if req.ProcessContent {
				processed, err := s.processFileContent(fileItem.Path, fileItem.Content)
				if err != nil {
					result.FailedFiles = append(result.FailedFiles, FailedFileInfo{
						Path:   fileItem.Path,
						Error:  err.Error(),
						Reason: "processing",
					})
					continue
				}
				processedContent = processed
				contentProcessed = true
			}

			// Create module version file entity
			moduleVersion := &model.ModuleVersion{} // Would normally find this
			moduleVersionFile := model.NewModuleVersionFile(0, moduleVersion, fileItem.Path, processedContent)

			// Store the file
			if err := s.moduleVersionFileRepo.Save(ctx, moduleVersionFile); err != nil {
				result.FailedFiles = append(result.FailedFiles, FailedFileInfo{
					Path:   fileItem.Path,
					Error:  err.Error(),
					Reason: "storage",
				})
				continue
			}

			// Add to successful files
			result.StoredFiles = append(result.StoredFiles, StoredFileInfo{
				Path:         fileItem.Path,
				FileID:       moduleVersionFile.ID(),
				Size:         len(processedContent),
				ContentType:  moduleVersionFile.ContentType(),
				Processed:    contentProcessed,
				ProcessingMs: time.Since(fileStart).Milliseconds(),
			})

			result.ProcessedFiles++
		}

		// Set success if no files failed (or if partial success is acceptable)
		result.Success = len(result.FailedFiles) == 0

		if !result.Success {
			return fmt.Errorf("failed to store %d out of %d files", len(result.FailedFiles), len(req.Files))
		}

		return nil
	})

	result.StorageDuration = time.Since(startTime)

	if err != nil {
		result.SavepointRolledBack = true
		errorMsg := err.Error()
		result.Error = &errorMsg
		return result, nil
	}

	return result, nil
}

// ProcessExampleFiles processes example files with individual savepoints for isolation
func (s *FileContentTransactionService) ProcessExampleFiles(
	ctx context.Context,
	moduleVersionID int,
	examples []ExampleFile,
) (*ExampleProcessingResult, error) {
	startTime := time.Now()
	result := &ExampleProcessingResult{
		Success:           false,
		ProcessedExamples: []ProcessedExample{},
		FailedExamples:    []FailedExample{},
		Timestamp:         startTime,
	}

	for _, example := range examples {
		// Each example gets its own savepoint for isolation
		savepointName := fmt.Sprintf("example_processing_%s_%d", strings.ReplaceAll(example.Path, "/", "_"), time.Now().UnixNano())

		err := s.savepointHelper.WithSmartSavepointOrTransaction(ctx, savepointName, func(tx *gorm.DB) error {
			// Validate example file
			if err := s.validateExampleFile(example); err != nil {
				return fmt.Errorf("example validation failed: %w", err)
			}

			// Store example file
			fileItem := FileContentItem{
				Path:    example.Path,
				Content: example.Content,
				Type:    "example",
			}

			storageReq := FileStorageRequest{
				ModuleVersionID: moduleVersionID,
				Files:           []FileContentItem{fileItem},
				TransactionCtx:  ctx,
				SavepointName:   savepointName,
				ProcessContent:  true,
				ValidatePaths:   true,
			}

			storageResult, err := s.StoreFilesWithTransaction(ctx, storageReq)
			if err != nil {
				return fmt.Errorf("failed to store example file: %w", err)
			}

			if !storageResult.Success {
				return fmt.Errorf("example file storage failed")
			}

			// Get the stored file info
			if len(storageResult.StoredFiles) > 0 {
				storedFile := storageResult.StoredFiles[0]

				// Check if example contains terraform files
				hasTerraform := s.hasTerraformFiles(example.Path, example.Content)

				processedExample := ProcessedExample{
					Path:         example.Path,
					FileID:       storedFile.FileID,
					IsMain:       example.IsMainExample,
					Directory:    example.Directory,
					HasTerraform: hasTerraform,
				}

				result.ProcessedExamples = append(result.ProcessedExamples, processedExample)
			}

			return nil
		})

		if err != nil {
			result.FailedExamples = append(result.FailedExamples, FailedExample{
				Path:  example.Path,
				Error: err.Error(),
			})
		}
	}

	result.ProcessingDuration = time.Since(startTime)
	result.Success = len(result.FailedExamples) == 0

	return result, nil
}

// StoreModuleArchiveContents stores all files from a module archive with transaction safety
func (s *FileContentTransactionService) StoreModuleArchiveContents(
	ctx context.Context,
	moduleVersionID int,
	archiveFiles map[string]string,
) (*FileStorageResult, error) {
	// Convert archive files to file content items
	var files []FileContentItem
	for path, content := range archiveFiles {
		files = append(files, FileContentItem{
			Path:    path,
			Content: content,
			Type:    "source",
		})
	}

	storageReq := FileStorageRequest{
		ModuleVersionID: moduleVersionID,
		Files:           files,
		TransactionCtx:  ctx,
		ProcessContent:  true,
		ValidatePaths:   true,
	}

	return s.StoreFilesWithTransaction(ctx, storageReq)
}

// UpdateFileContent updates a specific file's content within a transaction
func (s *FileContentTransactionService) UpdateFileContent(
	ctx context.Context,
	moduleVersionID int,
	filePath string,
	newContent string,
) error {
	savepointName := fmt.Sprintf("update_file_%s_%d", strings.ReplaceAll(filePath, "/", "_"), time.Now().UnixNano())

	return s.savepointHelper.WithSmartSavepointOrTransaction(ctx, savepointName, func(tx *gorm.DB) error {
		// Find existing file
		existingFile, err := s.moduleVersionFileRepo.FindByPath(ctx, moduleVersionID, filePath)
		if err != nil {
			return fmt.Errorf("failed to find existing file: %w", err)
		}

		if existingFile == nil {
			return fmt.Errorf("file not found: %s", filePath)
		}

		// Process the new content
		processedContent, err := s.processFileContent(filePath, newContent)
		if err != nil {
			return fmt.Errorf("failed to process content: %w", err)
		}

		// Update the file content
		// Note: This assumes ModuleVersionFile has a method to update content
		// In the current architecture, this might require creating a new instance
		updatedFile := model.NewModuleVersionFile(existingFile.ID(), existingFile.ModuleVersion(), filePath, processedContent)

		// Save the updated file
		if err := s.moduleVersionFileRepo.Save(ctx, updatedFile); err != nil {
			return fmt.Errorf("failed to save updated file: %w", err)
		}

		return nil
	})
}

// DeleteFilesWithTransaction deletes multiple files with transaction safety
func (s *FileContentTransactionService) DeleteFilesWithTransaction(
	ctx context.Context,
	moduleVersionID int,
	filePaths []string,
) error {
	savepointName := fmt.Sprintf("delete_files_%d_%d", moduleVersionID, time.Now().UnixNano())

	return s.savepointHelper.WithSmartSavepointOrTransaction(ctx, savepointName, func(tx *gorm.DB) error {
		for _, filePath := range filePaths {
			// Find the file
			file, err := s.moduleVersionFileRepo.FindByPath(ctx, moduleVersionID, filePath)
			if err != nil {
				return fmt.Errorf("failed to find file %s: %w", filePath, err)
			}

			if file == nil {
				continue // Skip non-existent files
			}

			// Delete the file
			if err := s.moduleVersionFileRepo.Delete(ctx, file.ID()); err != nil {
				return fmt.Errorf("failed to delete file %s: %w", filePath, err)
			}
		}

		return nil
	})
}

// validateFilePath validates that a file path is safe and allowed
func (s *FileContentTransactionService) validateFilePath(path string) error {
	return s.pathBuilder.ValidatePath(path)
}

// processFileContent processes file content based on its type
func (s *FileContentTransactionService) processFileContent(path, content string) (string, error) {
	// Use file processing service if available
	if s.fileProcessingService != nil {
		// Process based on content type
		contentType := s.getContentType(path)

		// Process based on content type
		switch contentType {
		case "text/markdown":
			return s.fileProcessingService.ProcessMarkdownContent(content)
		default:
			// For other content types, return content as-is
			return content, nil
		}
	}

	// Basic processing if no processing service available
	return content, nil
}

// validateExampleFile validates an example file
func (s *FileContentTransactionService) validateExampleFile(example ExampleFile) error {
	if example.Path == "" {
		return fmt.Errorf("example path cannot be empty")
	}

	if example.Content == "" {
		return fmt.Errorf("example content cannot be empty")
	}

	// Validate path
	if err := s.validateFilePath(example.Path); err != nil {
		return fmt.Errorf("invalid example path: %w", err)
	}

	return nil
}

// hasTerraformFiles checks if content contains terraform files
func (s *FileContentTransactionService) hasTerraformFiles(path, content string) bool {
	// Check file extension
	ext := strings.ToLower(filepath.Ext(path))
	if ext == ".tf" || ext == ".tfvars" {
		return true
	}

	// Check content for terraform patterns
	terraformPatterns := []string{
		"resource \"",
		"variable \"",
		"output \"",
		"provider \"",
		"module \"",
		"terraform {",
	}

	for _, pattern := range terraformPatterns {
		if strings.Contains(content, pattern) {
			return true
		}
	}

	return false
}

// getContentType determines the content type based on file extension
func (s *FileContentTransactionService) getContentType(path string) string {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".md":
		return "text/markdown"
	case ".json":
		return "application/json"
	case ".yml", ".yaml":
		return "application/x-yaml"
	case ".tf", ".tfvars":
		return "text/plain"
	default:
		return "text/plain"
	}
}

// GetFileStatistics returns statistics about stored files for a module version
func (s *FileContentTransactionService) GetFileStatistics(
	ctx context.Context,
	moduleVersionID int,
) (*FileStatistics, error) {
	files, err := s.moduleVersionFileRepo.FindByModuleVersionID(ctx, moduleVersionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get module files: %w", err)
	}

	stats := &FileStatistics{
		TotalFiles:     len(files),
		TotalSize:      0,
		FileTypes:      make(map[string]int),
		ProcessedTypes: make(map[string]int),
	}

	for _, file := range files {
		stats.TotalSize += len(file.Content())

		ext := strings.ToLower(filepath.Ext(file.Path()))
		stats.FileTypes[ext]++

		if file.ContentType() != "text/plain" {
			stats.ProcessedTypes[ext]++
		}
	}

	return stats, nil
}

// FileStatistics represents statistics about stored files
type FileStatistics struct {
	TotalFiles     int            `json:"total_files"`
	TotalSize      int            `json:"total_size"`
	FileTypes      map[string]int `json:"file_types"`
	ProcessedTypes map[string]int `json:"processed_types"`
}
