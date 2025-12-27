package service

import (
	"context"
	"io"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/storage/model"
)

// StorageService defines the interface matching Python BaseFileStorage exactly
// Only 8 core methods that are actually used in Python implementation
type StorageService interface {
	// Core 8 methods from Python - NO UNUSED METHODS

	// Archive operations (used by module_extractor.py)
	// Python equivalent: upload_file(source_path, dest_directory, dest_filename)
	UploadFile(ctx context.Context, sourcePath string, destDirectory string, destFilename string) error

	// File serving operations (used by download endpoints)
	// Python equivalent: read_file(path, bytes_mode=False)
	ReadFile(ctx context.Context, path string, bytesMode bool) ([]byte, error)

	// Binary content operations (used by provider binaries)
	// Python equivalent: write_file(path, content, binary=False)
	WriteFile(ctx context.Context, path string, content any, binary bool) error

	// Directory operations (used by extraction and module creation)
	// Python equivalent: make_directory(directory)
	MakeDirectory(ctx context.Context, directory string) error

	// Existence checks (used by cleanup and validation)
	// Python equivalent: file_exists(path)
	FileExists(ctx context.Context, path string) (bool, error)

	// Python equivalent: directory_exists(path)
	DirectoryExists(ctx context.Context, path string) (bool, error)

	// Cleanup operations (used by module deletion)
	// Python equivalent: delete_file(path)
	DeleteFile(ctx context.Context, path string) error

	// Python equivalent: delete_directory(path)
	DeleteDirectory(ctx context.Context, path string) error

	// Streaming operations for performance optimization
	UploadStream(ctx context.Context, reader io.Reader, destPath string) error
	StreamToHTTPResponse(ctx context.Context, path string, writer io.Writer) error
	GetFileSize(ctx context.Context, path string) (int64, error)
}


// StorageFactory defines the interface for creating storage services
type StorageFactory interface {
	CreateStorageService(config *model.StorageConfig) (StorageService, error)
	GetDefaultStorageService() (StorageService, error)
	DetectStorageType(dataDirectory string) model.StorageType
}

// PathBuilder defines the interface for constructing storage paths
// This replicates the Python path construction logic
type PathBuilder interface {
	BuildNamespacePath(namespace string) string
	BuildModulePath(namespace string, module string) string
	BuildProviderPath(namespace string, module string, provider string) string
	BuildVersionPath(namespace string, module string, provider string, version string) string
	BuildArchivePath(namespace string, module string, provider string, version string, archiveName string) string
	BuildUploadPath(filename string) string
	SafeJoinPaths(basePath string, subPaths ...string) string
	ValidatePath(path string) error
}

// TemporaryDirectoryManager manages temporary directories for module processing
// This replicates the Python temporary directory handling
type TemporaryDirectoryManager interface {
	CreateTemporaryDirectory(ctx context.Context, prefix string) (string, error)
	CreateExtractionDirectory(ctx context.Context, moduleVersionID int) (string, func(), error)
	CreateUploadDirectory(ctx context.Context) (string, func(), error)
	CleanupTemporaryDirectory(ctx context.Context, path string) error
}

// StoragePathConfig represents path configuration matching Python implementation
type StoragePathConfig struct {
	BasePath        string `json:"base_path"`
	ModulesPath     string `json:"modules_path"`
	ProvidersPath   string `json:"providers_path"`
	UploadPath      string `json:"upload_path"`
	TempPath        string `json:"temp_path"`
}