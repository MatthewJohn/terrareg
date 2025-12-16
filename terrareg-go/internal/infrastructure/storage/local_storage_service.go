package storage

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/storage/model"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/storage/service"
)

// LocalStorageService implements StorageService for local filesystem
// This replicates the Python LocalFileStorage class
type LocalStorageService struct {
	basePath string
	pathBuilder service.PathBuilder
}

// NewLocalStorageService creates a new local storage service
func NewLocalStorageService(basePath string, pathBuilder service.PathBuilder) (*LocalStorageService, error) {
	// Ensure base path exists
	if err := os.MkdirAll(basePath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create base directory: %w", err)
	}

	return &LocalStorageService{
		basePath:    basePath,
		pathBuilder: pathBuilder,
	}, nil
}

// generatePath generates the full path by prepending base directory
// This replicates Python's _generate_path method
func (s *LocalStorageService) generatePath(path string) string {
	if strings.HasPrefix(path, s.basePath) {
		return path
	}
	return s.pathBuilder.SafeJoinPaths(s.basePath, path)
}

// UploadFile uploads a file from source path to destination
// This replicates Python's upload_file method
func (s *LocalStorageService) UploadFile(ctx context.Context, sourcePath string, destDirectory string, destFilename string) (*model.UploadResult, error) {
	startTime := time.Now()

	// Generate full destination path
	fullDestPath := s.generatePath(s.pathBuilder.SafeJoinPaths(destDirectory, destFilename))

	// Create destination directory if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(fullDestPath), 0755); err != nil {
		errorMsg := fmt.Sprintf("failed to create destination directory: %v", err)
		return &model.UploadResult{
			Success:    false,
			Error:      &errorMsg,
			UploadTime: startTime,
		}, nil
	}

	// Copy file
	sourceFile, err := os.Open(sourcePath)
	if err != nil {
		errorMsg := fmt.Sprintf("failed to open source file: %v", err)
		return &model.UploadResult{
			Success:    false,
			Error:      &errorMsg,
			UploadTime: startTime,
		}, nil
	}
	defer sourceFile.Close()

	destFile, err := os.Create(fullDestPath)
	if err != nil {
		errorMsg := fmt.Sprintf("failed to create destination file: %v", err)
		return &model.UploadResult{
			Success:    false,
			Error:      &errorMsg,
			UploadTime: startTime,
		}, nil
	}
	defer destFile.Close()

	size, err := io.Copy(destFile, sourceFile)
	if err != nil {
		errorMsg := fmt.Sprintf("failed to copy file: %v", err)
		os.Remove(fullDestPath) // Clean up on failure
		return &model.UploadResult{
			Success:    false,
			Error:      &errorMsg,
			UploadTime: startTime,
		}, nil
	}

	// Get file info for etag
	info, err := os.Stat(fullDestPath)
	if err != nil {
		errorMsg := fmt.Sprintf("failed to get file info: %v", err)
		return &model.UploadResult{
			Success:    false,
			Error:      &errorMsg,
			UploadTime: startTime,
		}, nil
	}

	return &model.UploadResult{
		Success:    true,
		Path:       fullDestPath,
		Size:       size,
		ETag:       fmt.Sprintf("%d-%d", info.ModTime().Unix(), info.Size()),
		UploadTime: startTime,
	}, nil
}

// UploadFileContent uploads file content directly
func (s *LocalStorageService) UploadFileContent(ctx context.Context, content []byte, destDirectory string, destFilename string, contentType string) (*model.UploadResult, error) {
	startTime := time.Now()

	fullDestPath := s.generatePath(s.pathBuilder.SafeJoinPaths(destDirectory, destFilename))

	// Create destination directory if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(fullDestPath), 0755); err != nil {
		errorMsg := fmt.Sprintf("failed to create destination directory: %v", err)
		return &model.UploadResult{
			Success:    false,
			Error:      &errorMsg,
			UploadTime: startTime,
		}, nil
	}

	// Write content to file
	if err := os.WriteFile(fullDestPath, content, 0644); err != nil {
		errorMsg := fmt.Sprintf("failed to write file: %v", err)
		return &model.UploadResult{
			Success:    false,
			Error:      &errorMsg,
			UploadTime: startTime,
		}, nil
	}

	info, err := os.Stat(fullDestPath)
	if err != nil {
		errorMsg := fmt.Sprintf("failed to get file info: %v", err)
		return &model.UploadResult{
			Success:    false,
			Error:      &errorMsg,
			UploadTime: startTime,
		}, nil
	}

	return &model.UploadResult{
		Success:    true,
		Path:       fullDestPath,
		Size:       int64(len(content)),
		ETag:       fmt.Sprintf("%d-%d", info.ModTime().Unix(), info.Size()),
		UploadTime: startTime,
	}, nil
}

// DownloadFile downloads a file and returns a reader
func (s *LocalStorageService) DownloadFile(ctx context.Context, path string) (io.ReadCloser, *model.FileInfo, error) {
	fullPath := s.generatePath(path)

	info, err := os.Stat(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil, model.ErrFileNotFound
		}
		return nil, nil, fmt.Errorf("failed to get file info: %w", err)
	}

	file, err := os.Open(fullPath)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to open file: %w", err)
	}

	fileInfo := &model.FileInfo{
		Path:         path,
		Size:         info.Size(),
		LastModified: info.ModTime(),
		ContentType:  s.getContentType(fullPath),
		StorageType:  model.StorageTypeLocal,
	}

	return file, fileInfo, nil
}

// ReadFile reads file content
func (s *LocalStorageService) ReadFile(ctx context.Context, path string, bytesMode bool) ([]byte, *model.FileInfo, error) {
	fullPath := s.generatePath(path)

	info, err := os.Stat(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil, model.ErrFileNotFound
		}
		return nil, nil, fmt.Errorf("failed to get file info: %w", err)
	}

	content, err := os.ReadFile(fullPath)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read file: %w", err)
	}

	fileInfo := &model.FileInfo{
		Path:         path,
		Size:         info.Size(),
		LastModified: info.ModTime(),
		ContentType:  s.getContentType(fullPath),
		StorageType:  model.StorageTypeLocal,
	}

	return content, fileInfo, nil
}

// WriteFile writes content to a file
func (s *LocalStorageService) WriteFile(ctx context.Context, path string, content any, binary bool) error {
	fullPath := s.generatePath(path)

	// Create directory if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	var data []byte
	var err error

	switch v := content.(type) {
	case []byte:
		data = v
	case string:
		data = []byte(v)
	case io.Reader:
		data, err = io.ReadAll(v)
		if err != nil {
			return fmt.Errorf("failed to read content: %w", err)
		}
	default:
		return fmt.Errorf("unsupported content type: %T", content)
	}

	if err := os.WriteFile(fullPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// FileExists checks if a file exists
func (s *LocalStorageService) FileExists(ctx context.Context, path string) (bool, error) {
	fullPath := s.generatePath(path)

	info, err := os.Stat(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}

	return !info.IsDir(), nil
}

// DirectoryExists checks if a directory exists
func (s *LocalStorageService) DirectoryExists(ctx context.Context, path string) (bool, error) {
	fullPath := s.generatePath(path)

	info, err := os.Stat(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}

	return info.IsDir(), nil
}

// MakeDirectory creates a directory
func (s *LocalStorageService) MakeDirectory(ctx context.Context, directory string) error {
	fullPath := s.generatePath(directory)
	return os.MkdirAll(fullPath, 0755)
}

// ListDirectory lists files in a directory
func (s *LocalStorageService) ListDirectory(ctx context.Context, directory string) ([]*model.FileInfo, error) {
	fullPath := s.generatePath(directory)

	entries, err := os.ReadDir(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, model.ErrDirectoryNotFound
		}
		return nil, fmt.Errorf("failed to read directory: %w", err)
	}

	var files []*model.FileInfo
	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			continue // Skip entries we can't get info for
		}

		filePath := s.pathBuilder.SafeJoinPaths(directory, entry.Name())
		files = append(files, &model.FileInfo{
			Path:         filePath,
			Size:         info.Size(),
			LastModified: info.ModTime(),
			ContentType:  s.getContentType(filePath),
			StorageType:  model.StorageTypeLocal,
		})
	}

	return files, nil
}

// DeleteFile deletes a file
func (s *LocalStorageService) DeleteFile(ctx context.Context, path string) error {
	fullPath := s.generatePath(path)

	if err := os.Remove(fullPath); err != nil {
		if os.IsNotExist(err) {
			return model.ErrFileNotFound
		}
		return fmt.Errorf("failed to delete file: %w", err)
	}

	return nil
}

// DeleteDirectory deletes a directory and all its contents
func (s *LocalStorageService) DeleteDirectory(ctx context.Context, path string) error {
	fullPath := s.generatePath(path)

	if err := os.RemoveAll(fullPath); err != nil {
		if os.IsNotExist(err) {
			return model.ErrDirectoryNotFound
		}
		return fmt.Errorf("failed to delete directory: %w", err)
	}

	return nil
}

// BatchUpload uploads multiple files
func (s *LocalStorageService) BatchUpload(ctx context.Context, files []service.BatchUploadRequest) (*model.BatchUploadResult, error) {
	result := &model.BatchUploadResult{
		TotalFiles: len(files),
		SuccessfulFiles: []model.UploadResult{},
		FailedFiles:    []model.UploadResult{},
		OverallSuccess: true,
		PartialSuccess: false,
	}

	for _, req := range files {
		var uploadResult *model.UploadResult
		var err error

		if req.IsContent {
			uploadResult, err = s.UploadFileContent(ctx, req.Content, req.DestDirectory, req.DestFilename, req.ContentType)
		} else {
			uploadResult, err = s.UploadFile(ctx, req.SourcePath, req.DestDirectory, req.DestFilename)
		}

		if err != nil || !uploadResult.Success {
			result.FailedFiles = append(result.FailedFiles, *uploadResult)
			result.OverallSuccess = false
			result.PartialSuccess = true
		} else {
			result.SuccessfulFiles = append(result.SuccessfulFiles, *uploadResult)
		}
	}

	return result, nil
}

// GeneratePath generates a path from components
func (s *LocalStorageService) GeneratePath(pathComponents ...string) string {
	return s.pathBuilder.SafeJoinPaths(pathComponents...)
}

// ValidatePath validates a path
func (s *LocalStorageService) ValidatePath(path string) error {
	return s.pathBuilder.ValidatePath(path)
}

// GetStorageType returns the storage type
func (s *LocalStorageService) GetStorageType() model.StorageType {
	return model.StorageTypeLocal
}

// GetStorageStats returns storage statistics
func (s *LocalStorageService) GetStorageStats(ctx context.Context) (*model.StorageStats, error) {
	// Implementation would scan the base path and calculate stats
	// For now, return empty stats
	return &model.StorageStats{
		TotalFiles:     0,
		TotalSize:      0,
		UploadCount:    0,
		DownloadCount:  0,
		LastUploadTime: time.Time{},
	}, nil
}

// getContentType determines content type based on file extension
func (s *LocalStorageService) getContentType(path string) string {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".tar", ".gz", ".zip":
		return "application/gzip"
	case ".json":
		return "application/json"
	case ".md":
		return "text/markdown"
	case ".txt":
		return "text/plain"
	case ".tf":
		return "text/plain"
	case ".yaml", ".yml":
		return "text/yaml"
	default:
		return "application/octet-stream"
	}
}