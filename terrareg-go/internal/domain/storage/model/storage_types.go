package model

import (
	"time"
)

// StorageType represents the type of storage backend
type StorageType int

const (
	StorageTypeLocal StorageType = iota
	StorageTypeS3
)

// String returns the string representation of storage type
func (st StorageType) String() string {
	switch st {
	case StorageTypeLocal:
		return "local"
	case StorageTypeS3:
		return "s3"
	default:
		return "unknown"
	}
}

// StorageConfig represents storage configuration
type StorageConfig struct {
	Type           StorageType `json:"type"`
	DataDirectory  string      `json:"data_directory"`
	UploadDirectory string     `json:"upload_directory"`
	S3Config       *S3Config   `json:"s3_config,omitempty"`
}

// S3Config represents S3-specific configuration
type S3Config struct {
	Bucket    string `json:"bucket"`
	Region    string `json:"region"`
	KeyPrefix string `json:"key_prefix"`
	AccessKey string `json:"access_key,omitempty"`
	SecretKey string `json:"secret_key,omitempty"`
}

// FileInfo represents information about a stored file
type FileInfo struct {
	Path         string    `json:"path"`
	Size         int64     `json:"size"`
	ContentType  string    `json:"content_type,omitempty"`
	LastModified time.Time `json:"last_modified"`
	ETag         string    `json:"etag,omitempty"`
	StorageType  StorageType `json:"storage_type"`
}

// DirectoryInfo represents information about a directory
type DirectoryInfo struct {
	Path         string    `json:"path"`
	FileCount    int       `json:"file_count"`
	TotalSize    int64     `json:"total_size"`
	LastModified time.Time `json:"last_modified"`
	StorageType  StorageType `json:"storage_type"`
}

// StoragePath represents a structured storage path
type StoragePath struct {
	Namespace string `json:"namespace"`
	Module    string `json:"module"`
	Provider  string `json:"provider"`
	Version   string `json:"version"`
}

// ArchivePath represents archive-specific path information
type ArchivePath struct {
	StoragePath `json:"storage_path"`
	ArchiveName string `json:"archive_name"` // e.g., "source.tar.gz", "source.zip"
}

// UploadResult represents the result of a file upload
type UploadResult struct {
	Success    bool      `json:"success"`
	Path       string    `json:"path"`
	Size       int64     `json:"size"`
	ETag       string    `json:"etag,omitempty"`
	UploadTime time.Time `json:"upload_time"`
	Error      *string   `json:"error,omitempty"`
}

// BatchUploadResult represents the result of a batch upload
type BatchUploadResult struct {
	TotalFiles     int           `json:"total_files"`
	SuccessfulFiles []UploadResult `json:"successful_files"`
	FailedFiles    []UploadResult `json:"failed_files"`
	OverallSuccess bool          `json:"overall_success"`
	PartialSuccess bool          `json:"partial_success"`
}

// StorageStats represents storage statistics
type StorageStats struct {
	TotalFiles     int64 `json:"total_files"`
	TotalSize      int64 `json:"total_size"`
	UploadCount    int64 `json:"upload_count"`
	DownloadCount  int64 `json:"download_count"`
	LastUploadTime time.Time `json:"last_upload_time"`
}

// StorageError represents storage-specific errors
type StorageError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Path    string `json:"path,omitempty"`
}

func (e *StorageError) Error() string {
	if e.Path != "" {
		return e.Message + " (path: " + e.Path + ")"
	}
	return e.Message
}

// Common error types
var (
	ErrFileNotFound     = &StorageError{Code: "FILE_NOT_FOUND", Message: "File not found"}
	ErrDirectoryNotFound = &StorageError{Code: "DIRECTORY_NOT_FOUND", Message: "Directory not found"}
	ErrPermissionDenied = &StorageError{Code: "PERMISSION_DENIED", Message: "Permission denied"}
	ErrStorageFull      = &StorageError{Code: "STORAGE_FULL", Message: "Storage full"}
	ErrInvalidPath      = &StorageError{Code: "INVALID_PATH", Message: "Invalid path"}
	ErrPathTraversal    = &StorageError{Code: "PATH_TRAVERSAL", Message: "Path traversal detected"}
	ErrUploadFailed     = &StorageError{Code: "UPLOAD_FAILED", Message: "Upload failed"}
	ErrDeleteFailed     = &StorageError{Code: "DELETE_FAILED", Message: "Delete failed"}
	ErrConfigInvalid    = &StorageError{Code: "CONFIG_INVALID", Message: "Invalid configuration"}
)