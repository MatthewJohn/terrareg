package storage

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/storage/service"
)

// TemporaryDirectoryManagerImpl manages temporary directories
// This replicates Python's temporary directory handling for module processing
type TemporaryDirectoryManagerImpl struct {
	baseTempDir string
	mu          sync.RWMutex
	directories map[string]time.Time // Track created directories for cleanup
}

// NewTemporaryDirectoryManager creates a new temporary directory manager
func NewTemporaryDirectoryManager() (*TemporaryDirectoryManagerImpl, error) {
	baseTempDir := filepath.Join(os.TempDir(), "terrareg")

	if err := os.MkdirAll(baseTempDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create base temp directory: %w", err)
	}

	return &TemporaryDirectoryManagerImpl{
		baseTempDir: baseTempDir,
		directories: make(map[string]time.Time),
	}, nil
}

// CreateTemporaryDirectory creates a temporary directory with given prefix
func (m *TemporaryDirectoryManagerImpl) CreateTemporaryDirectory(ctx context.Context, prefix string) (string, error) {
	tempDir, err := os.MkdirTemp(m.baseTempDir, prefix)
	if err != nil {
		return "", fmt.Errorf("failed to create temporary directory: %w", err)
	}

	m.mu.Lock()
	m.directories[tempDir] = time.Now()
	m.mu.Unlock()

	return tempDir, nil
}

// CreateExtractionDirectory creates a temporary directory for module extraction
func (m *TemporaryDirectoryManagerImpl) CreateExtractionDirectory(ctx context.Context, moduleVersionID int) (string, func(), error) {
	tempDir, err := m.CreateTemporaryDirectory(ctx, fmt.Sprintf("extract_%d_", moduleVersionID))
	if err != nil {
		return "", nil, err
	}

	// Return cleanup function
	cleanup := func() {
		m.CleanupTemporaryDirectory(ctx, tempDir)
	}

	return tempDir, cleanup, nil
}

// CreateUploadDirectory creates a temporary directory for uploads
func (m *TemporaryDirectoryManagerImpl) CreateUploadDirectory(ctx context.Context) (string, func(), error) {
	tempDir, err := m.CreateTemporaryDirectory(ctx, "upload_")
	if err != nil {
		return "", nil, err
	}

	// Return cleanup function
	cleanup := func() {
		m.CleanupTemporaryDirectory(ctx, tempDir)
	}

	return tempDir, cleanup, nil
}

// CleanupTemporaryDirectory removes a temporary directory
func (m *TemporaryDirectoryManagerImpl) CleanupTemporaryDirectory(ctx context.Context, path string) error {
	m.mu.Lock()
	delete(m.directories, path)
	m.mu.Unlock()

	if err := os.RemoveAll(path); err != nil {
		return fmt.Errorf("failed to cleanup temporary directory: %w", err)
	}

	return nil
}

// GetTemporaryDirectoryStats returns statistics about temporary directories
func (m *TemporaryDirectoryManagerImpl) GetTemporaryDirectoryStats(ctx context.Context) (*service.TemporaryDirectoryStats, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	stats := &service.TemporaryDirectoryStats{
		TotalDirectories: len(m.directories),
		Directories:      []service.DirectoryInfo{},
	}

	for path, createdAt := range m.directories {
		info, err := os.Stat(path)
		if err != nil {
			continue // Skip directories we can't stat
		}

		var size int64
		var files []string

		// Walk directory to get size and files
		filepath.Walk(path, func(filePath string, file os.FileInfo, err error) error {
			if err != nil {
				return nil // Skip errors
			}
			if !file.IsDir() {
				size += file.Size()
				relPath, _ := filepath.Rel(path, filePath)
				files = append(files, relPath)
			}
			return nil
		})

		dirInfo := service.DirectoryInfo{
			Path:      path,
			Size:      size,
			CreatedAt: createdAt,
			Age:       time.Since(createdAt),
			Files:     files,
		}

		stats.Directories = append(stats.Directories, dirInfo)
		stats.TotalSize += size

		if stats.OldestDirectory.IsZero() || createdAt.Before(stats.OldestDirectory) {
			stats.OldestDirectory = createdAt
		}
	}

	return stats, nil
}

// CleanupOldDirectories removes temporary directories older than maxAge
func (m *TemporaryDirectoryManagerImpl) CleanupOldDirectories(ctx context.Context, maxAge time.Duration) error {
	m.mu.RLock()
	toCleanup := []string{}
	now := time.Now()

	for path, createdAt := range m.directories {
		if now.Sub(createdAt) > maxAge {
			toCleanup = append(toCleanup, path)
		}
	}
	m.mu.RUnlock()

	var errors []string
	for _, path := range toCleanup {
		if err := m.CleanupTemporaryDirectory(ctx, path); err != nil {
			errors = append(errors, fmt.Sprintf("failed to cleanup %s: %v", path, err))
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("cleanup errors: %v", errors)
	}

	return nil
}