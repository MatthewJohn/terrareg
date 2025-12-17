package service

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gorm.io/gorm"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/persistence/sqldb/transaction"
)

// ArchiveGenerationTransactionService handles archive generation with transaction safety
// and rollback capabilities for partial failures during archive creation
type ArchiveGenerationTransactionService struct {
	savepointHelper *transaction.SavepointHelper
}

// NewArchiveGenerationTransactionService creates a new archive generation transaction service
func NewArchiveGenerationTransactionService(
	savepointHelper *transaction.SavepointHelper,
) *ArchiveGenerationTransactionService {
	return &ArchiveGenerationTransactionService{
		savepointHelper: savepointHelper,
	}
}

// ArchiveFormat represents supported archive formats
type ArchiveFormat int

const (
	ArchiveFormatZIP ArchiveFormat = iota
	ArchiveFormatTarGz
)

// String returns the string representation of archive format
func (af ArchiveFormat) String() string {
	switch af {
	case ArchiveFormatZIP:
		return "zip"
	case ArchiveFormatTarGz:
		return "tar.gz"
	default:
		return "unknown"
	}
}

// ArchiveGenerationRequest represents a request to generate archives
type ArchiveGenerationRequest struct {
	ModuleVersionID int
	SourcePath      string // Path to source files
	ArchivePath     string // Output directory for archives
	Formats         []ArchiveFormat
	PathspecFilter  *PathspecFilter // For filtering files (from .terraformignore)
	TransactionCtx  context.Context
	SavepointName   string
}

// ArchiveGenerationResult represents the result of archive generation
type ArchiveGenerationResult struct {
	Success             bool               `json:"success"`
	GeneratedArchives   []GeneratedArchive `json:"generated_archives,omitempty"`
	FailedFormats       []ArchiveFormat    `json:"failed_formats,omitempty"`
	Error               *string            `json:"error,omitempty"`
	GenerationDuration  time.Duration      `json:"generation_duration"`
	SourceFilesCount    int                `json:"source_files_count"`
	TotalArchiveSize    int64              `json:"total_archive_size"`
	Timestamp           time.Time          `json:"timestamp"`
	SavepointRolledBack bool               `json:"savepoint_rolled_back"`
}

// GeneratedArchive represents information about a generated archive
type GeneratedArchive struct {
	Format      ArchiveFormat `json:"format"`
	Path        string        `json:"path"`
	Size        int64         `json:"size"`
	FileCount   int           `json:"file_count"`
	Compression string        `json:"compression"`
	CreatedAt   time.Time     `json:"created_at"`
}

// ArchiveFile represents a file to be included in an archive
type ArchiveFile struct {
	Path         string    `json:"path"`
	Content      []byte    `json:"content"`
	Size         int64     `json:"size"`
	LastModified time.Time `json:"last_modified"`
}

// MultiFormatResult represents the result of generating multiple archive formats
type MultiFormatResult struct {
	OverallStatus   string                    `json:"overall_status"`
	Formats         map[string]*ArchiveResult `json:"formats"`
	HasFailures     bool                      `json:"has_failures"`
	FailureSummary  string                    `json:"failure_summary,omitempty"`
	TotalFormats    int                       `json:"total_formats"`
	SuccessCount    int                       `json:"success_count"`
	FailureCount    int                       `json:"failure_count"`
	OverallDuration time.Duration             `json:"overall_duration"`
}

// ArchiveResult represents the result of generating a single archive format
type ArchiveResult struct {
	Format    ArchiveFormat `json:"format"`
	Success   bool          `json:"success"`
	Path      string        `json:"path,omitempty"`
	Size      int64         `json:"size,omitempty"`
	FileCount int           `json:"file_count,omitempty"`
	Error     *string       `json:"error,omitempty"`
	Duration  time.Duration `json:"duration"`
}

// GenerateArchivesWithTransaction generates multiple archive formats with transaction safety
func (s *ArchiveGenerationTransactionService) GenerateArchivesWithTransaction(
	ctx context.Context,
	req ArchiveGenerationRequest,
) (*ArchiveGenerationResult, error) {
	startTime := time.Now()

	result := &ArchiveGenerationResult{
		Success:             false,
		GeneratedArchives:   []GeneratedArchive{},
		FailedFormats:       []ArchiveFormat{},
		SavepointRolledBack: false,
		Timestamp:           startTime,
	}

	// Use provided savepoint name or create new one
	savepointName := req.SavepointName
	if savepointName == "" {
		savepointName = fmt.Sprintf("archive_generation_%d_%d", req.ModuleVersionID, startTime.UnixNano())
	}

	err := s.savepointHelper.WithSavepointNamed(ctx, savepointName, func(tx *gorm.DB) error {
		// Scan source files
		sourceFiles, err := s.scanSourceFiles(req.SourcePath, req.PathspecFilter)
		if err != nil {
			return fmt.Errorf("failed to scan source files: %w", err)
		}

		result.SourceFilesCount = len(sourceFiles)

		// Generate each requested format
		for _, format := range req.Formats {
			archiveResult, err := s.generateArchiveFormat(ctx, req.SourcePath, req.ArchivePath, format, sourceFiles)
			if err != nil {
				result.FailedFormats = append(result.FailedFormats, format)
				continue
			}

			if archiveResult.Success {
				generatedArchive := GeneratedArchive{
					Format:      format,
					Path:        archiveResult.Path,
					Size:        archiveResult.Size,
					FileCount:   archiveResult.FileCount,
					Compression: s.getCompressionType(format),
					CreatedAt:   time.Now(),
				}
				result.GeneratedArchives = append(result.GeneratedArchives, generatedArchive)
				result.TotalArchiveSize += archiveResult.Size
			} else {
				result.FailedFormats = append(result.FailedFormats, format)
			}
		}

		// Check if all formats were generated successfully
		if len(result.FailedFormats) > 0 {
			return fmt.Errorf("failed to generate %d archive formats", len(result.FailedFormats))
		}

		result.Success = true
		return nil
	})

	result.GenerationDuration = time.Since(startTime)

	if err != nil {
		result.SavepointRolledBack = true
		errorMsg := err.Error()
		result.Error = &errorMsg
		return result, nil
	}

	return result, nil
}

// GenerateMultipleFormats generates multiple archive formats with individual savepoints
func (s *ArchiveGenerationTransactionService) GenerateMultipleFormats(
	ctx context.Context,
	moduleVersionID int,
	sourcePath string,
	archivePath string,
	formats []ArchiveFormat,
	pathspecFilter *PathspecFilter,
) (*MultiFormatResult, error) {
	startTime := time.Now()

	result := &MultiFormatResult{
		OverallStatus: "Success",
		Formats:       make(map[string]*ArchiveResult),
		HasFailures:   false,
		TotalFormats:  len(formats),
		SuccessCount:  0,
		FailureCount:  0,
	}

	// Scan source files once
	sourceFiles, err := s.scanSourceFiles(sourcePath, pathspecFilter)
	if err != nil {
		return nil, fmt.Errorf("failed to scan source files: %w", err)
	}

	// Generate each format with its own savepoint
	for _, format := range formats {
		savepointName := fmt.Sprintf("archive_format_%s_%d", format.String(), time.Now().UnixNano())

		err := s.savepointHelper.WithSavepointNamed(ctx, savepointName, func(tx *gorm.DB) error {
			archiveResult, err := s.generateArchiveFormat(ctx, sourcePath, archivePath, format, sourceFiles)
			if err != nil {
				result.Formats[format.String()] = &ArchiveResult{
					Format:   format,
					Success:  false,
					Error:    func() *string { e := err.Error(); return &e }(),
					Duration: 0,
				}
				result.FailureCount++
				result.HasFailures = true
				return nil // Continue with other formats
			}

			result.Formats[format.String()] = archiveResult
			if archiveResult.Success {
				result.SuccessCount++
			} else {
				result.FailureCount++
				result.HasFailures = true
			}

			return nil
		})

		if err != nil {
			// Savepoint creation failed
			result.Formats[format.String()] = &ArchiveResult{
				Format:   format,
				Success:  false,
				Error:    func() *string { e := err.Error(); return &e }(),
				Duration: 0,
			}
			result.FailureCount++
			result.HasFailures = true
		}
	}

	// Set overall status
	if result.HasFailures {
		result.OverallStatus = "Error"
		result.FailureSummary = fmt.Sprintf("%d of %d formats failed to generate", result.FailureCount, result.TotalFormats)
	}

	result.OverallDuration = time.Since(startTime)

	return result, nil
}

// generateArchiveFormat generates a single archive format
func (s *ArchiveGenerationTransactionService) generateArchiveFormat(
	ctx context.Context,
	sourcePath string,
	archivePath string,
	format ArchiveFormat,
	sourceFiles []string,
) (*ArchiveResult, error) {
	startTime := time.Now()

	result := &ArchiveResult{
		Format:  format,
		Success: false,
	}

	// Create archive filename
	archiveFileName := fmt.Sprintf("module.%s", format.String())
	archiveFilePath := filepath.Join(archivePath, archiveFileName)

	// Generate archive based on format
	switch format {
	case ArchiveFormatZIP:
		fileCount, size, err := s.generateZIPArchive(sourcePath, archiveFilePath, sourceFiles)
		if err != nil {
			result.Error = func() *string { e := err.Error(); return &e }()
			result.Duration = time.Since(startTime)
			return result, nil
		}
		result.Path = archiveFilePath
		result.Size = size
		result.FileCount = fileCount

	case ArchiveFormatTarGz:
		fileCount, size, err := s.generateTarGzArchive(sourcePath, archiveFilePath, sourceFiles)
		if err != nil {
			result.Error = func() *string { e := err.Error(); return &e }()
			result.Duration = time.Since(startTime)
			return result, nil
		}
		result.Path = archiveFilePath
		result.Size = size
		result.FileCount = fileCount

	default:
		result.Error = func() *string { e := fmt.Sprintf("unsupported archive format: %s", format.String()); return &e }()
		result.Duration = time.Since(startTime)
		return result, nil
	}

	result.Success = true
	result.Duration = time.Since(startTime)
	return result, nil
}

// generateZIPArchive generates a ZIP archive
func (s *ArchiveGenerationTransactionService) generateZIPArchive(
	sourcePath string,
	archivePath string,
	sourceFiles []string,
) (int, int64, error) {
	// Create ZIP file
	zipFile, err := os.Create(archivePath)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to create ZIP file: %w", err)
	}
	defer zipFile.Close()

	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	fileCount := 0

	// Add files to archive
	for _, filePath := range sourceFiles {
		fullPath := filepath.Join(sourcePath, filePath)

		// Read file content
		content, err := os.ReadFile(fullPath)
		if err != nil {
			return 0, 0, fmt.Errorf("failed to read file %s: %w", filePath, err)
		}

		// Create file in archive
		header := &zip.FileHeader{
			Name:   filePath,
			Method: zip.Deflate,
		}

		writer, err := zipWriter.CreateHeader(header)
		if err != nil {
			return 0, 0, fmt.Errorf("failed to create file in archive %s: %w", filePath, err)
		}

		// Write file content
		_, err = writer.Write(content)
		if err != nil {
			return 0, 0, fmt.Errorf("failed to write file to archive %s: %w", filePath, err)
		}

		fileCount++
	}

	// Get archive size
	stat, err := os.Stat(archivePath)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to get archive size: %w", err)
	}

	return fileCount, stat.Size(), nil
}

// generateTarGzArchive generates a tar.gz archive
func (s *ArchiveGenerationTransactionService) generateTarGzArchive(
	sourcePath string,
	archivePath string,
	sourceFiles []string,
) (int, int64, error) {
	// Create tar.gz file
	tarFile, err := os.Create(archivePath)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to create tar.gz file: %w", err)
	}
	defer tarFile.Close()

	gzipWriter := gzip.NewWriter(tarFile)
	defer gzipWriter.Close()

	tarWriter := tar.NewWriter(gzipWriter)
	defer tarWriter.Close()

	fileCount := 0

	// Add files to archive
	for _, filePath := range sourceFiles {
		fullPath := filepath.Join(sourcePath, filePath)

		// Read file info
		fileInfo, err := os.Stat(fullPath)
		if err != nil {
			return 0, 0, fmt.Errorf("failed to get file info for %s: %w", filePath, err)
		}

		// Read file content
		content, err := os.ReadFile(fullPath)
		if err != nil {
			return 0, 0, fmt.Errorf("failed to read file %s: %w", filePath, err)
		}

		// Create tar header
		header := &tar.Header{
			Name:     filePath,
			Size:     int64(len(content)),
			Mode:     int64(fileInfo.Mode()),
			ModTime:  fileInfo.ModTime(),
			Typeflag: tar.TypeReg,
		}

		// Write header
		if err := tarWriter.WriteHeader(header); err != nil {
			return 0, 0, fmt.Errorf("failed to write tar header for %s: %w", filePath, err)
		}

		// Write file content
		if _, err := tarWriter.Write(content); err != nil {
			return 0, 0, fmt.Errorf("failed to write file to tar archive %s: %w", filePath, err)
		}

		fileCount++
	}

	// Get archive size
	stat, err := os.Stat(archivePath)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to get archive size: %w", err)
	}

	return fileCount, stat.Size(), nil
}

// scanSourceFiles scans source directory for files to include in archives
func (s *ArchiveGenerationTransactionService) scanSourceFiles(
	sourcePath string,
	pathspecFilter *PathspecFilter,
) ([]string, error) {
	var files []string

	err := filepath.Walk(sourcePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Get relative path
		relPath, err := filepath.Rel(sourcePath, path)
		if err != nil {
			return fmt.Errorf("failed to get relative path for %s: %w", path, err)
		}

		// Normalize path separators
		relPath = filepath.ToSlash(relPath)

		// Apply pathspec filter if provided
		if pathspecFilter != nil {
			if s.shouldIgnoreFile(relPath, pathspecFilter) {
				return nil
			}
		}

		files = append(files, relPath)
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to scan source files: %w", err)
	}

	return files, nil
}

// shouldIgnoreFile checks if a file should be ignored based on pathspec filter
func (s *ArchiveGenerationTransactionService) shouldIgnoreFile(filePath string, filter *PathspecFilter) bool {
	if filter == nil || len(filter.Rules) == 0 {
		return false
	}

	for _, rule := range filter.Rules {
		if s.matchesRule(filePath, rule) {
			return true
		}
	}

	return false
}

// matchesRule checks if a file path matches a pathspec rule
func (s *ArchiveGenerationTransactionService) matchesRule(filePath, rule string) bool {
	// Simple glob matching - in a full implementation, this would use
	// a proper pathspec library like github.com/mozillazg/go-pathspec
	if strings.Contains(rule, "*") {
		// Basic glob support
		if matched, _ := filepath.Match(rule, filePath); matched {
			return true
		}
	}

	// Directory-based rules
	if strings.HasSuffix(rule, "/") && strings.HasPrefix(filePath, rule) {
		return true
	}

	// Exact match
	if filePath == rule {
		return true
	}

	// Prefix match
	if strings.HasPrefix(filePath, rule) {
		return true
	}

	return false
}

// getCompressionType returns the compression type for an archive format
func (s *ArchiveGenerationTransactionService) getCompressionType(format ArchiveFormat) string {
	switch format {
	case ArchiveFormatZIP:
		return "deflate"
	case ArchiveFormatTarGz:
		return "gzip"
	default:
		return "none"
	}
}

// ValidateArchiveRequest validates an archive generation request
func (s *ArchiveGenerationTransactionService) ValidateArchiveRequest(
	req ArchiveGenerationRequest,
) error {
	if req.ModuleVersionID <= 0 {
		return fmt.Errorf("invalid module version ID")
	}

	if req.SourcePath == "" {
		return fmt.Errorf("source path is required")
	}

	if _, err := os.Stat(req.SourcePath); os.IsNotExist(err) {
		return fmt.Errorf("source path does not exist: %s", req.SourcePath)
	}

	if req.ArchivePath == "" {
		return fmt.Errorf("archive path is required")
	}

	if len(req.Formats) == 0 {
		return fmt.Errorf("at least one format must be specified")
	}

	for _, format := range req.Formats {
		if !s.isValidFormat(format) {
			return fmt.Errorf("invalid archive format: %s", format.String())
		}
	}

	return nil
}

// isValidFormat checks if an archive format is valid
func (s *ArchiveGenerationTransactionService) isValidFormat(format ArchiveFormat) bool {
	switch format {
	case ArchiveFormatZIP, ArchiveFormatTarGz:
		return true
	default:
		return false
	}
}

// CleanupGeneratedArchives removes generated archives on transaction rollback
func (s *ArchiveGenerationTransactionService) CleanupGeneratedArchives(
	ctx context.Context,
	archives []GeneratedArchive,
) error {
	var errors []string

	for _, archive := range archives {
		if err := os.Remove(archive.Path); err != nil {
			errors = append(errors, fmt.Sprintf("failed to remove %s: %v", archive.Path, err))
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("cleanup failed: %s", strings.Join(errors, "; "))
	}

	return nil
}

// GetArchiveInfo returns information about a generated archive
func (s *ArchiveGenerationTransactionService) GetArchiveInfo(
	archivePath string,
) (*GeneratedArchive, error) {
	stat, err := os.Stat(archivePath)
	if os.IsNotExist(err) {
		return nil, fmt.Errorf("archive does not exist: %s", archivePath)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get archive info: %w", err)
	}

	// Determine format from extension
	ext := strings.ToLower(filepath.Ext(archivePath))
	var format ArchiveFormat
	switch ext {
	case ".zip":
		format = ArchiveFormatZIP
	case ".gz":
		if strings.HasSuffix(strings.ToLower(archivePath), ".tar.gz") {
			format = ArchiveFormatTarGz
		} else {
			return nil, fmt.Errorf("unrecognized gzip archive format: %s", archivePath)
		}
	default:
		return nil, fmt.Errorf("unrecognized archive format: %s", ext)
	}

	return &GeneratedArchive{
		Format:      format,
		Path:        archivePath,
		Size:        stat.Size(),
		Compression: s.getCompressionType(format),
		CreatedAt:   stat.ModTime(),
	}, nil
}
