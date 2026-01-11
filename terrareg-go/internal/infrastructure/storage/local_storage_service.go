package storage

import (
	"archive/zip"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/storage/service"
)

// LocalStorageService implements StorageService for local filesystem
// This replicates the Python LocalFileStorage class
type LocalStorageService struct {
	basePath    string
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
func (s *LocalStorageService) UploadFile(ctx context.Context, sourcePath string, destDirectory string, destFilename string) error {
	// Generate full destination path
	fullDestPath := s.generatePath(s.pathBuilder.SafeJoinPaths(destDirectory, destFilename))

	// Create destination directory if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(fullDestPath), 0755); err != nil {
		return fmt.Errorf("failed to create destination directory: %w", err)
	}

	// Copy file
	sourceFile, err := os.Open(sourcePath)
	if err != nil {
		return fmt.Errorf("failed to open source file: %w", err)
	}
	defer sourceFile.Close()

	destFile, err := os.Create(fullDestPath)
	if err != nil {
		return fmt.Errorf("failed to create destination file: %w", err)
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	if err != nil {
		os.Remove(fullDestPath) // Clean up on failure
		return fmt.Errorf("failed to copy file: %w", err)
	}

	return nil
}

// ReadFile reads a file and returns its contents
// This replicates Python's read_file method
func (s *LocalStorageService) ReadFile(ctx context.Context, path string, bytesMode bool) ([]byte, error) {
	fullPath := s.generatePath(path)

	// Check if file exists
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("file not found: %s", fullPath)
	}

	content, err := os.ReadFile(fullPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	if !bytesMode {
		// For text mode, just return the bytes as-is
		// Python's text mode handles encoding differently
		return content, nil
	}

	return content, nil
}

// WriteFile writes content to a file
// This replicates Python's write_file method
func (s *LocalStorageService) WriteFile(ctx context.Context, path string, content any, binary bool) error {
	fullPath := s.generatePath(path)

	// Create directory if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	var file *os.File
	var err error

	if binary {
		file, err = os.OpenFile(fullPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	} else {
		file, err = os.OpenFile(fullPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	}

	if err != nil {
		return fmt.Errorf("failed to open file for writing: %w", err)
	}
	defer file.Close()

	switch v := content.(type) {
	case []byte:
		_, err = file.Write(v)
	case string:
		_, err = file.Write([]byte(v))
	default:
		return fmt.Errorf("unsupported content type: %T", content)
	}

	if err != nil {
		return fmt.Errorf("failed to write content: %w", err)
	}

	return nil
}

// MakeDirectory creates a directory
// This replicates Python's make_directory method
func (s *LocalStorageService) MakeDirectory(ctx context.Context, directory string) error {
	fullPath := s.generatePath(directory)
	return os.MkdirAll(fullPath, 0755)
}

// FileExists checks if a file exists
// This replicates Python's file_exists method
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
// This replicates Python's directory_exists method
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

// DeleteFile deletes a file
// This replicates Python's delete_file method
func (s *LocalStorageService) DeleteFile(ctx context.Context, path string) error {
	fullPath := s.generatePath(path)

	// Check if file exists
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		return nil // File doesn't exist, nothing to delete
	}

	return os.Remove(fullPath)
}

// DeleteDirectory deletes a directory and all its contents
// This replicates Python's delete_directory method
func (s *LocalStorageService) DeleteDirectory(ctx context.Context, path string) error {
	fullPath := s.generatePath(path)

	// Check if directory exists
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		return nil // Directory doesn't exist, nothing to delete
	}

	return os.RemoveAll(fullPath)
}

// CopyDir recursively copies a directory from source to destination
// Merged from duplicate LocalStorage implementation
func (s *LocalStorageService) CopyDir(src, dest string) error {
	info, err := os.Stat(src)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(dest, info.Mode()); err != nil {
		return err
	}

	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		destPath := filepath.Join(dest, entry.Name())

		if entry.IsDir() {
			if err := s.CopyDir(srcPath, destPath); err != nil {
				return err
			}
		} else {
			// Prevent symlink traversal
			if entry.Type()&os.ModeSymlink != 0 {
				continue
			}
			if err := copyFileHelper(srcPath, destPath); err != nil {
				return err
			}
		}
	}
	return nil
}

// ExtractArchive extracts a ZIP archive from src to dest
// Merged from duplicate LocalStorage implementation
func (s *LocalStorageService) ExtractArchive(src, dest string) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer r.Close()

	// Create destination directory
	if err := os.MkdirAll(dest, 0755); err != nil {
		return err
	}

	// Extract each file
	for _, f := range r.File {
		// Prevent path traversal attacks
		if strings.Contains(f.Name, "..") {
			continue
		}

		filePath := filepath.Join(dest, f.Name)

		if f.FileInfo().IsDir() {
			os.MkdirAll(filePath, 0755)
			continue
		}

		// Create parent directories
		if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
			return err
		}

		// Create the file
		outFile, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return err
		}

		rc, err := f.Open()
		if err != nil {
			outFile.Close()
			return err
		}

		_, err = io.Copy(outFile, rc)
		outFile.Close()
		rc.Close()

		if err != nil {
			return err
		}
	}

	return nil
}

// copyFileHelper copies a file from src to dest
// Helper function from duplicate LocalStorage implementation
func copyFileHelper(src, dest string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.OpenFile(dest, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer out.Close()

	if _, err := io.Copy(out, in); err != nil {
		return err
	}
	return out.Close()
}

// UploadStream uploads data from a reader directly to storage
// This provides streaming upload without loading entire file into memory
func (s *LocalStorageService) UploadStream(ctx context.Context, reader io.Reader, destPath string) error {
	fullPath := s.generatePath(destPath)

	// Create destination directory if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
		return fmt.Errorf("failed to create destination directory: %w", err)
	}

	// Create destination file
	file, err := os.Create(fullPath)
	if err != nil {
		return fmt.Errorf("failed to create destination file: %w", err)
	}
	defer file.Close()

	// Copy data from reader to file
	_, err = io.Copy(file, reader)
	if err != nil {
		os.Remove(fullPath) // Clean up on failure
		return fmt.Errorf("failed to write data: %w", err)
	}

	return nil
}

// StreamToHTTPResponse streams a file directly to HTTP response writer
// This provides streaming download without loading entire file into memory
func (s *LocalStorageService) StreamToHTTPResponse(ctx context.Context, path string, writer io.Writer) error {
	fullPath := s.generatePath(path)

	// Open file for reading
	file, err := os.Open(fullPath)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// Copy file directly to response writer
	_, err = io.Copy(writer, file)
	if err != nil {
		return fmt.Errorf("failed to stream file: %w", err)
	}

	return nil
}

// GetFileSize returns the size of a file in bytes
func (s *LocalStorageService) GetFileSize(ctx context.Context, path string) (int64, error) {
	fullPath := s.generatePath(path)

	fileInfo, err := os.Stat(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return 0, fmt.Errorf("file not found: %s", path)
		}
		return 0, fmt.Errorf("failed to stat file: %w", err)
	}

	return fileInfo.Size(), nil
}
