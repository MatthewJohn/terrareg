package service

import (
	"context"
	"io"
	"time"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/storage/model"
)

// StorageService defines the interface for file storage operations
// This follows the Python BaseFileStorage interface pattern
type StorageService interface {
	// File operations
	UploadFile(ctx context.Context, sourcePath string, destDirectory string, destFilename string) (*model.UploadResult, error)
	UploadFileContent(ctx context.Context, content []byte, destDirectory string, destFilename string, contentType string) (*model.UploadResult, error)
	DownloadFile(ctx context.Context, path string) (io.ReadCloser, *model.FileInfo, error)
	ReadFile(ctx context.Context, path string, bytesMode bool) ([]byte, *model.FileInfo, error)
	WriteFile(ctx context.Context, path string, content any, binary bool) error

	// File/directory existence checks
	FileExists(ctx context.Context, path string) (bool, error)
	DirectoryExists(ctx context.Context, path string) (bool, error)

	// Directory operations
	MakeDirectory(ctx context.Context, directory string) error
	ListDirectory(ctx context.Context, directory string) ([]*model.FileInfo, error)

	// Delete operations
	DeleteFile(ctx context.Context, path string) error
	DeleteDirectory(ctx context.Context, path string) error

	// Batch operations
	BatchUpload(ctx context.Context, files []BatchUploadRequest) (*model.BatchUploadResult, error)

	// Path operations
	GeneratePath(pathComponents ...string) string
	ValidatePath(path string) error

	// Storage info
	GetStorageType() model.StorageType
	GetStorageStats(ctx context.Context) (*model.StorageStats, error)
}

// BatchUploadRequest represents a single file upload request in a batch
type BatchUploadRequest struct {
	SourcePath   string `json:"source_path,omitempty"`
	Content      []byte `json:"content,omitempty"`
	DestDirectory string `json:"dest_directory"`
	DestFilename string `json:"dest_filename"`
	ContentType  string `json:"content_type,omitempty"`
	IsContent    bool   `json:"is_content"` // true if Content is used, false if SourcePath is used
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
	GetTemporaryDirectoryStats(ctx context.Context) (*TemporaryDirectoryStats, error)
}

// TemporaryDirectoryStats represents statistics about temporary directories
type TemporaryDirectoryStats struct {
	TotalDirectories int               `json:"total_directories"`
	TotalSize        int64             `json:"total_size"`
	OldestDirectory  time.Time         `json:"oldest_directory"`
	Directories      []DirectoryInfo   `json:"directories"`
}

// DirectoryInfo represents information about a temporary directory
type DirectoryInfo struct {
	Path      string    `json:"path"`
	Size      int64     `json:"size"`
	CreatedAt time.Time `json:"created_at"`
	Age       time.Duration `json:"age"`
	Files     []string  `json:"files"`
}

// StoragePathConfig represents path configuration matching Python implementation
type StoragePathConfig struct {
	BasePath        string `json:"base_path"`
	ModulesPath     string `json:"modules_path"`
	ProvidersPath   string `json:"providers_path"`
	UploadPath      string `json:"upload_path"`
	ArchivePrefix   string `json:"archive_prefix"`
	SourcePrefix    string `json:"source_prefix"`
}