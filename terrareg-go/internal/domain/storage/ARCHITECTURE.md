# Storage Domain Architecture

## Overview

The Storage domain provides file system and storage management for modules, providers, and assets. It abstracts storage operations behind an interface, supporting both local filesystem and S3-compatible storage backends.

---

## Core Functionality

The storage domain provides the following capabilities:

- **File Operations** - Read, write, delete files
- **Directory Operations** - Create, check existence, delete directories
- **Streaming Operations** - Stream uploads and downloads for performance
- **Path Building** - Construct consistent storage paths
- **Temporary Directory Management** - Manage temp directories for processing
- **Storage Backend Abstraction** - Support multiple storage backends

---

## Domain Components

### Models

**Location**: `/internal/domain/storage/model/storage_types.go`

#### StorageType Enum

```go
type StorageType string

const (
    StorageTypeLocal  StorageType = "local"
    StorageTypeS3     StorageType = "s3"
    StorageTypeMinIO  StorageType = "minio"
)
```

#### StorageConfig

```go
type StorageConfig struct {
    Type         StorageType
    LocalPath    string  // For local storage
    S3Bucket     string  // For S3 storage
    S3Endpoint   string  // For S3-compatible storage
    S3Region     string  // For S3 storage
    S3AccessKey  string  // For S3 storage
    S3SecretKey  string  // For S3 storage
}
```

#### StoragePathConfig

```go
type StoragePathConfig struct {
    BasePath      string
    ModulesPath   string
    ProvidersPath string
    UploadPath    string
    TempPath      string
}
```

### Service Interface

**Location**: `/internal/domain/storage/service/storage_service.go`

#### StorageService Interface

The core interface matching Python's BaseFileStorage exactly:

```go
type StorageService interface {
    // Archive operations
    UploadFile(ctx context.Context, sourcePath, destDirectory, destFilename string) error

    // File serving operations
    ReadFile(ctx context.Context, path string, bytesMode bool) ([]byte, error)

    // Binary content operations
    WriteFile(ctx context.Context, path string, content any, binary bool) error

    // Directory operations
    MakeDirectory(ctx context.Context, directory string) error

    // Existence checks
    FileExists(ctx context.Context, path string) (bool, error)
    DirectoryExists(ctx context.Context, path string) (bool, error)

    // Cleanup operations
    DeleteFile(ctx context.Context, path string) error
    DeleteDirectory(ctx context.Context, path string) error

    // Streaming operations
    UploadStream(ctx context.Context, reader io.Reader, destPath string) error
    StreamToHTTPResponse(ctx context.Context, path string, writer io.Writer) error
    GetFileSize(ctx context.Context, path string) (int64, error)
}
```

### Factory Interface

**Location**: `/internal/domain/storage/service/storage_service.go`

```go
type StorageFactory interface {
    CreateStorageService(config *model.StorageConfig) (StorageService, error)
    GetDefaultStorageService() (StorageService, error)
    DetectStorageType(dataDirectory string) model.StorageType
}
```

### PathBuilder Interface

```go
type PathBuilder interface {
    BuildNamespacePath(namespace string) string
    BuildModulePath(namespace, module string) string
    BuildProviderPath(namespace, module, provider string) string
    BuildVersionPath(namespace, module, provider, version string) string
    BuildArchivePath(namespace, module, provider, version, archiveName string) string
    BuildUploadPath(filename string) string
    SafeJoinPaths(basePath string, subPaths ...string) string
    ValidatePath(path string) error
}
```

### TemporaryDirectoryManager Interface

```go
type TemporaryDirectoryManager interface {
    CreateTemporaryDirectory(ctx context.Context, prefix string) (string, error)
    CreateExtractionDirectory(ctx context.Context, moduleVersionID int) (string, func(), error)
    CreateUploadDirectory(ctx context.Context) (string, func(), error)
    CleanupTemporaryDirectory(ctx context.Context, path string) error
}
```

---

## Dependencies

### Domain Dependencies

| Domain | Purpose |
|--------|---------|
| **config** | For storage configuration (data directory, storage type) |

### Infrastructure Dependencies

| Component | Purpose |
|-----------|---------|
| **File System** - Local file system storage |
| **S3 API** - AWS S3 or S3-compatible storage |
| **MinIO** - Self-hosted S3-compatible storage |

### Domains That Use Storage

| Domain | Purpose |
|--------|---------|
| **module** | Store module archives and examples |
| **provider** | Store provider binaries |
| **git** | Store cloned repositories |
| **provider_logo** | Store provider logos |

---

## Key Design Principles

1. **Python Parity** - Interface matches Python BaseFileStorage exactly
2. **Backend Agnostic** - Same interface for local, S3, MinIO
3. **Streaming Support** - Efficient large file handling
4. **Path Consistency** - Standardized path construction
5. **Temporary Safety** - Cleanup callbacks for temp directories

---

## Storage Path Structure

### Module Storage

```
{data_directory}/
├── modules/
│   └── {namespace}/
│       └── {module}/
│           └── {provider}/
│               └── {version}/
│                   ├── module.zip
│                   ├── module.json
│                   └── examples/
└── upload/
    └── {upload_id}/
```

### Provider Storage

```
{data_directory}/
├── providers/
│   └── {namespace}/
│       └── {provider}/
│           └── {version}/
│               ├── terraform-provider-{name}_{version}_{os}_{arch}.zip
│               ├── terraform-provider-{name}_{version}_SHA256SUMS
│               └── terraform-provider-{name}_{version}_SHA256SUMS.sig
```

---

## Usage Examples

### Uploading a File

```go
err := storageService.UploadFile(
    ctx,
    "/tmp/module.zip",
    "/modules/aws/vpc/aws",
    "1.0.0/module.zip",
)
```

### Reading a File

```go
content, err := storageService.ReadFile(ctx, "/modules/aws/vpc/aws/1.0.0/module.zip", true)
```

### Streaming Upload

```go
file, err := os.Open("/tmp/large-file.zip")
defer file.Close()

err := storageService.UploadStream(ctx, file, "/modules/.../large-file.zip")
```

### Streaming to HTTP Response

```go
w := http.ResponseWriter
err := storageService.StreamToHTTPResponse(ctx, "/modules/.../module.zip", w)
```

### Creating Temporary Directory

```go
tempDir, cleanup, err := tempMgr.CreateUploadDirectory(ctx)
defer cleanup() // Automatic cleanup

// Use tempDir for processing
```

---

## Storage Backend Detection

The factory can detect storage type from data directory:

```go
storageType := factory.DetectStorageType("/data")
// Returns StorageTypeLocal for local filesystem
// Could detect S3/MinIO from config
```

---

## Temporary Directory Management

### Extraction Directory

For module extraction during processing:

```go
extractDir, cleanup, err := tempMgr.CreateExtractionDirectory(ctx, moduleVersionID)
defer cleanup()

// Extract and process module
unzipModule(extractDir)
processModuleFiles(extractDir)
```

### Upload Directory

For file uploads:

```go
uploadDir, cleanup, err := tempMgr.CreateUploadDirectory(ctx)
defer cleanup()

// Save uploaded file
saveUploadedFile(uploadDir, file)
```

---

## Path Building Examples

```go
pathBuilder := storage.NewPathBuilder(config)

// Namespace path
pathBuilder.BuildNamespacePath("aws")
// Returns: "/modules/aws"

// Module path
pathBuilder.BuildModulePath("aws", "vpc")
// Returns: "/modules/aws/vpc"

// Provider path
pathBuilder.BuildProviderPath("aws", "vpc", "aws")
// Returns: "/modules/aws/vpc/aws"

// Version path
pathBuilder.BuildVersionPath("aws", "vpc", "aws", "1.0.0")
// Returns: "/modules/aws/vpc/aws/1.0.0"

// Archive path
pathBuilder.BuildArchivePath("aws", "vpc", "aws", "1.0.0", "module.zip")
// Returns: "/modules/aws/vpc/aws/1.0.0/module.zip"
```

---

## Storage Backend Implementations

### Local Storage

```go
type LocalStorageService struct {
    basePath string
}
```

Uses standard Go `os` and `io` packages for file operations.

### S3 Storage

```go
type S3StorageService struct {
    client   *s3.Client
    bucket   string
    basePath string
}
```

Uses AWS SDK for Go v2 for S3 operations.

---

## Python Compatibility

The StorageService interface matches Python's BaseFileStorage exactly:

| Python Method | Go Method |
|---------------|-----------|
| `upload_file()` | `UploadFile()` |
| `read_file()` | `ReadFile()` |
| `write_file()` | `WriteFile()` |
| `make_directory()` | `MakeDirectory()` |
| `file_exists()` | `FileExists()` |
| `directory_exists()` | `DirectoryExists()` |
| `delete_file()` | `DeleteFile()` |
| `delete_directory()` | `DeleteDirectory()` |

---

## References

- [`/internal/domain/storage/model/`](./model/) - Storage models
- [`/internal/domain/storage/service/`](./service/) - Storage service interfaces
- [`/internal/infrastructure/storage/`](../../infrastructure/storage/) - Storage implementations
