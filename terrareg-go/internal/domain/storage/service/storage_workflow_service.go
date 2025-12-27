package service

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/storage/model"
)

// StorageWorkflowService integrates storage components for module processing workflows
// This orchestrates the critical module indexing workflow: Git clone → temp dir → process → upload → cleanup
type StorageWorkflowService interface {
	// Module archive workflow matching Python
	PrepareModuleVersionPath(ctx context.Context, namespace, module, provider, version string) (string, error)
	StoreModuleArchives(ctx context.Context, tempDir string, namespace, module, provider, version string) error
	GetArchivePath(namespace, module, provider, version, archiveName string) string

	// Temporary directory management for module processing
	CreateProcessingDirectory(ctx context.Context, prefix string) (string, func(), error)
	CreateExtractionDirectory(ctx context.Context, moduleVersionID int) (string, func(), error)
	CleanupDirectory(ctx context.Context, path string) error
}

// StorageWorkflowServiceImpl implements StorageWorkflowService
type StorageWorkflowServiceImpl struct {
	storageService StorageService
	pathBuilder    PathBuilder
	tempDirManager TemporaryDirectoryManager
	config         *model.StoragePathConfig
}

// NewStorageWorkflowServiceImpl creates a new storage workflow service
func NewStorageWorkflowServiceImpl(
	storageService StorageService,
	pathBuilder PathBuilder,
	tempDirManager TemporaryDirectoryManager,
	config *model.StoragePathConfig,
) *StorageWorkflowServiceImpl {
	return &StorageWorkflowServiceImpl{
		storageService: storageService,
		pathBuilder:    pathBuilder,
		tempDirManager: tempDirManager,
		config:         config,
	}
}

// PrepareModuleVersionPath creates the directory structure for a module version
// This replicates Python's directory structure: /modules/{ns}/{module}/{provider}/{version}/
func (s *StorageWorkflowServiceImpl) PrepareModuleVersionPath(ctx context.Context, namespace, module, provider, version string) (string, error) {
	versionPath := s.pathBuilder.BuildVersionPath(namespace, module, provider, version)

	// Create the directory structure
	if err := s.storageService.MakeDirectory(ctx, versionPath); err != nil {
		return "", fmt.Errorf("failed to create module version directory: %w", err)
	}

	return versionPath, nil
}

// StoreModuleArchives stores module archives (tar.gz and zip) from temp directory to storage
// This handles the critical archive upload step in the module indexing workflow
func (s *StorageWorkflowServiceImpl) StoreModuleArchives(ctx context.Context, tempDir string, namespace, module, provider, version string) error {
	// Archive names must match Python exactly
	archives := []string{
		"source.tar.gz",
		"source.zip",
	}

	destinationPath := s.pathBuilder.BuildVersionPath(namespace, module, provider, version)

	// Upload each archive
	for _, archiveName := range archives {
		sourcePath := filepath.Join(tempDir, archiveName)

		// Check if archive exists
		if _, err := os.Stat(sourcePath); os.IsNotExist(err) {
			// Try to find archive in archives subdirectory
			sourcePath = filepath.Join(tempDir, "archives", archiveName)
			if _, err := os.Stat(sourcePath); os.IsNotExist(err) {
				continue // Skip if archive doesn't exist
			}
		}

		// Upload archive to storage
		if err := s.storageService.UploadFile(ctx, sourcePath, destinationPath, archiveName); err != nil {
			return fmt.Errorf("failed to upload archive %s: %w", archiveName, err)
		}
	}

	return nil
}

// GetArchivePath returns the full path for an archive file
// This replicates Python's archive path construction
func (s *StorageWorkflowServiceImpl) GetArchivePath(namespace, module, provider, version, archiveName string) string {
	versionPath := s.pathBuilder.BuildVersionPath(namespace, module, provider, version)
	return s.pathBuilder.SafeJoinPaths(versionPath, archiveName)
}

// CreateProcessingDirectory creates a temporary directory for module processing
// This is used for the Git clone → processing → cleanup workflow
func (s *StorageWorkflowServiceImpl) CreateProcessingDirectory(ctx context.Context, prefix string) (string, func(), error) {
	tempDir, err := s.tempDirManager.CreateTemporaryDirectory(ctx, prefix)
	if err != nil {
		return "", nil, err
	}

	// Return cleanup function
	cleanup := func() {
		s.tempDirManager.CleanupTemporaryDirectory(ctx, tempDir)
	}

	return tempDir, cleanup, nil
}

// CreateExtractionDirectory creates a temporary directory for archive extraction
// This replicates Python's temporary directory extraction pattern
func (s *StorageWorkflowServiceImpl) CreateExtractionDirectory(ctx context.Context, moduleVersionID int) (string, func(), error) {
	return s.tempDirManager.CreateExtractionDirectory(ctx, moduleVersionID)
}

// CleanupDirectory removes a temporary directory
// This ensures proper cleanup in the module indexing workflow
func (s *StorageWorkflowServiceImpl) CleanupDirectory(ctx context.Context, path string) error {
	return s.tempDirManager.CleanupTemporaryDirectory(ctx, path)
}

// ArchivePaths represents the standard archive paths generated by Python
var StandardArchivePaths = []string{
	"source.tar.gz",
	"source.zip",
}

// IsStandardArchive checks if an archive name is one of the standard Python archives
func (s *StorageWorkflowServiceImpl) IsStandardArchive(archiveName string) bool {
	for _, standardName := range StandardArchivePaths {
		if archiveName == standardName {
			return true
		}
	}
	return false
}

// GetModuleVersionArchivePaths returns all archive paths for a module version
// This replicates Python's archive path generation exactly
func (s *StorageWorkflowServiceImpl) GetModuleVersionArchivePaths(namespace, module, provider, version string) []string {
	versionPath := s.pathBuilder.BuildVersionPath(namespace, module, provider, version)

	paths := make([]string, len(StandardArchivePaths))
	for i, archiveName := range StandardArchivePaths {
		paths[i] = s.pathBuilder.SafeJoinPaths(versionPath, archiveName)
	}

	return paths
}

// ValidateModuleArchiveStructure ensures the directory structure matches Python expectations
// This validates the critical path structure for module indexing
func (s *StorageWorkflowServiceImpl) ValidateModuleArchiveStructure(namespace, module, provider, version string) error {
	versionPath := s.pathBuilder.BuildVersionPath(namespace, module, provider, version)

	// Check if version directory exists
	exists, err := s.storageService.DirectoryExists(context.Background(), versionPath)
	if err != nil {
		return fmt.Errorf("failed to check version directory: %w", err)
	}
	if !exists {
		return fmt.Errorf("module version directory does not exist: %s", versionPath)
	}

	// Check for required archives
	for _, archiveName := range StandardArchivePaths {
		archivePath := s.pathBuilder.SafeJoinPaths(versionPath, archiveName)
		exists, err := s.storageService.FileExists(context.Background(), archivePath)
		if err != nil {
			return fmt.Errorf("failed to check archive %s: %w", archiveName, err)
		}
		if !exists {
			return fmt.Errorf("required archive missing: %s", archivePath)
		}
	}

	return nil
}