package storage

import (
	"context"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/storage/service"
)

// ModuleStorageAdapter adapts the domain StorageService to the module service StorageService interface
// This bridges the gap between the 8-method domain interface and the filesystem-specific module interface
type ModuleStorageAdapter struct {
	domainStorage service.StorageService
	pathBuilder   service.PathBuilder
}

// NewModuleStorageAdapter creates a new adapter for module storage operations
func NewModuleStorageAdapter(domainStorage service.StorageService, pathBuilder service.PathBuilder) *ModuleStorageAdapter {
	return &ModuleStorageAdapter{
		domainStorage: domainStorage,
		pathBuilder:   pathBuilder,
	}
}

// CopyDir recursively copies a directory from source to destination
func (a *ModuleStorageAdapter) CopyDir(src, dest string) error {
	// For now, use filesystem operations directly
	// In a full implementation, this could use the domain storage service
	return filepath.Walk(src, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}

		destPath := filepath.Join(dest, relPath)

		if info.IsDir() {
			return os.MkdirAll(destPath, info.Mode())
		}

		// Copy file
		ctx := context.Background()
		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		return a.domainStorage.WriteFile(ctx, destPath, content, true)
	})
}

// MkdirTemp creates a new temporary directory
func (a *ModuleStorageAdapter) MkdirTemp(dir, pattern string) (string, error) {
	return os.MkdirTemp(dir, pattern)
}

// RemoveAll removes a path and any children it contains
func (a *ModuleStorageAdapter) RemoveAll(path string) error {
	// Use domain storage service for file operations
	ctx := context.Background()

	// Try to delete as directory first
	err := a.domainStorage.DeleteDirectory(ctx, path)
	if err != nil {
		// If that fails, try to delete as file
		return a.domainStorage.DeleteFile(ctx, path)
	}

	return nil
}

// Stat returns a FileInfo describing the named file
func (a *ModuleStorageAdapter) Stat(name string) (fs.FileInfo, error) {
	return os.Stat(name)
}

// MkdirAll creates a directory path
func (a *ModuleStorageAdapter) MkdirAll(path string, perm fs.FileMode) error {
	ctx := context.Background()
	return a.domainStorage.MakeDirectory(ctx, path)
}

// ReadFile reads the file named by filename and returns the contents
func (a *ModuleStorageAdapter) ReadFile(filename string) ([]byte, error) {
	ctx := context.Background()
	return a.domainStorage.ReadFile(ctx, filename, true)
}

// ReadDir reads the directory named by dirname and returns a list of directory entries
func (a *ModuleStorageAdapter) ReadDir(dirname string) ([]fs.DirEntry, error) {
	return os.ReadDir(dirname)
}

// ExtractArchive extracts a ZIP archive from src to dest
func (a *ModuleStorageAdapter) ExtractArchive(src, dest string) error {
	// Use the LocalStorageService's ExtractArchive method if available
	// This would need type assertion in a real implementation
	// For now, use the domain storage service to handle this
	if localStorageService, ok := a.domainStorage.(*LocalStorageService); ok {
		return localStorageService.ExtractArchive(src, dest)
	}

	// Fallback implementation would go here
	return os.ErrNotExist
}
