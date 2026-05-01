package model

import (
	"context"
	"path/filepath"
	"strings"
)

// ModuleVersionFile represents a file within a module version
type ModuleVersionFile struct {
	id            int
	moduleVersion *ModuleVersion
	path          string
	content       string
	fileName      string
	contentType   string
}

// NewModuleVersionFile creates a new module version file
func NewModuleVersionFile(id int, moduleVersion *ModuleVersion, path, content string) *ModuleVersionFile {
	fileName := filepath.Base(path)
	contentType := "text/plain"

	// Determine content type based on file extension
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".md":
		contentType = "text/markdown"
	case ".json":
		contentType = "application/json"
	case ".yml", ".yaml":
		contentType = "application/x-yaml"
	case ".tf":
		contentType = "text/plain"
	}

	return &ModuleVersionFile{
		id:            id,
		moduleVersion: moduleVersion,
		path:          path,
		content:       content,
		fileName:      fileName,
		contentType:   contentType,
	}
}

// ID returns the file ID
func (mvf *ModuleVersionFile) ID() int {
	return mvf.id
}

// ModuleVersion returns the associated module version
func (mvf *ModuleVersionFile) ModuleVersion() *ModuleVersion {
	return mvf.moduleVersion
}

// Path returns the file path within the module
func (mvf *ModuleVersionFile) Path() string {
	return mvf.path
}

// Content returns the file content
func (mvf *ModuleVersionFile) Content() string {
	return mvf.content
}

// FileName returns the file name
func (mvf *ModuleVersionFile) FileName() string {
	return mvf.fileName
}

// ContentType returns the content type
func (mvf *ModuleVersionFile) ContentType() string {
	return mvf.contentType
}

// IsMarkdown returns true if the file is a markdown file
func (mvf *ModuleVersionFile) IsMarkdown() bool {
	return strings.HasSuffix(strings.ToLower(mvf.path), ".md")
}

// ValidatePath validates that the file path is safe
func (mvf *ModuleVersionFile) ValidatePath() error {
	// Check for path traversal attempts
	if strings.Contains(mvf.path, "..") {
		return ErrInvalidFilePath
	}

	// Ensure path is relative and doesn't start with /
	if strings.HasPrefix(mvf.path, "/") {
		return ErrInvalidFilePath
	}

	// Check for empty path
	if mvf.path == "" {
		return ErrInvalidFilePath
	}

	return nil
}

// ModuleVersionFileRepository defines the interface for module version file persistence
type ModuleVersionFileRepository interface {
	FindByPath(ctx context.Context, moduleVersionID int, path string) (*ModuleVersionFile, error)
	FindByModuleVersionID(ctx context.Context, moduleVersionID int) ([]*ModuleVersionFile, error)
	Save(ctx context.Context, file *ModuleVersionFile) error
	Delete(ctx context.Context, id int) error
}

// FileStorageService defines the interface for file storage operations
type FileStorageService interface {
	GetFileContent(ctx context.Context, namespace, moduleName, provider, version, path string) (string, error)
	ValidateFilePath(path string) error
	SanitizeContent(content string, contentType string) (string, error)
}

// FileProcessingService defines the interface for file content processing
type FileProcessingService interface {
	ProcessMarkdownContent(content string) (string, error)
	FormatCodeContent(content string, language string) (string, error)
	SanitizeHTML(content string) (string, error)
}

// Error definitions
var (
	ErrInvalidFilePath = NewDomainError("invalid file path")
	ErrFileNotFound    = NewDomainError("file not found")
	ErrUnauthorized    = NewDomainError("unauthorized access")
)

// DomainError represents a domain-specific error
type DomainError struct {
	message string
}

// NewDomainError creates a new domain error
func NewDomainError(message string) *DomainError {
	return &DomainError{message: message}
}

// Error implements the error interface
func (e *DomainError) Error() string {
	return e.message
}
