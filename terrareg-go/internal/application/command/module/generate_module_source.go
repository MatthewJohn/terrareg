package module

import (
	"archive/zip"
	"bytes"
	"context"
	"fmt"
	"io"

	moduleRepo "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/module/repository"
	moduleService "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/module/service"
	storageService "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/storage/service"
)

// GenerateModuleSourceCommand handles generating source archives for module versions
type GenerateModuleSourceCommand struct {
	moduleProviderRepo moduleRepo.ModuleProviderRepository
	moduleFileService  *moduleService.ModuleFileService
	storageService    storageService.StorageService
}

// NewGenerateModuleSourceCommand creates a new generate module source command
func NewGenerateModuleSourceCommand(
	moduleProviderRepo moduleRepo.ModuleProviderRepository,
	moduleFileService *moduleService.ModuleFileService,
	storageService storageService.StorageService,
) *GenerateModuleSourceCommand {
	return &GenerateModuleSourceCommand{
		moduleProviderRepo: moduleProviderRepo,
		moduleFileService:  moduleFileService,
		storageService:    storageService,
	}
}

// GenerateModuleSourceRequest represents a request to generate module source
type GenerateModuleSourceRequest struct {
	Namespace string
	Module    string
	Provider  string
	Version   string
}

// GenerateModuleSourceResponse represents the generated module source archive information
type GenerateModuleSourceResponse struct {
	Filename    string
	ContentType string
	Content     []byte
	Size        int64
}

// GenerateModuleSourceStorageResponse represents the result when storing the generated archive
type GenerateModuleSourceStorageResponse struct {
	Filename    string
	ContentType string
	Size        int64
	StoragePath string
	Stored      bool
}

// Execute generates a source archive for the module version (in-memory)
func (c *GenerateModuleSourceCommand) Execute(ctx context.Context, req *GenerateModuleSourceRequest) (*GenerateModuleSourceResponse, error) {
	// Validate request
	if req.Namespace == "" || req.Module == "" || req.Provider == "" || req.Version == "" {
		return nil, fmt.Errorf("missing required parameters")
	}

	// Get all module files for the version
	files, err := c.moduleFileService.ListModuleFiles(ctx, req.Namespace, req.Module, req.Provider, req.Version)
	if err != nil {
		return nil, fmt.Errorf("failed to get module files: %w", err)
	}

	if len(files) == 0 {
		return nil, fmt.Errorf("no files found for module version")
	}

	// Create ZIP archive in memory
	var buf bytes.Buffer
	zipWriter := zip.NewWriter(&buf)

	// Add each file to the ZIP
	for _, file := range files {
		// Create file entry in ZIP
		writer, err := zipWriter.Create(file.Path())
		if err != nil {
			zipWriter.Close()
			return nil, fmt.Errorf("failed to create file entry: %w", err)
		}

		// Write file content
		_, err = writer.Write([]byte(file.Content()))
		if err != nil {
			zipWriter.Close()
			return nil, fmt.Errorf("failed to write file content: %w", err)
		}
	}

	// Close ZIP writer
	err = zipWriter.Close()
	if err != nil {
		return nil, fmt.Errorf("failed to close ZIP archive: %w", err)
	}

	// Generate filename
	filename := fmt.Sprintf("%s-%s-%s-%s.zip", req.Namespace, req.Module, req.Provider, req.Version)

	return &GenerateModuleSourceResponse{
		Filename:    filename,
		ContentType: "application/zip",
		Content:     buf.Bytes(),
		Size:        int64(buf.Len()),
	}, nil
}

// ExecuteAndStore generates a source archive and streams it directly to storage
func (c *GenerateModuleSourceCommand) ExecuteAndStore(ctx context.Context, req *GenerateModuleSourceRequest) (*GenerateModuleSourceStorageResponse, error) {
	// Validate request
	if req.Namespace == "" || req.Module == "" || req.Provider == "" || req.Version == "" {
		return nil, fmt.Errorf("missing required parameters")
	}

	// Get all module files for the version
	files, err := c.moduleFileService.ListModuleFiles(ctx, req.Namespace, req.Module, req.Provider, req.Version)
	if err != nil {
		return nil, fmt.Errorf("failed to get module files: %w", err)
	}

	if len(files) == 0 {
		return nil, fmt.Errorf("no files found for module version")
	}

	// Generate filename
	filename := fmt.Sprintf("%s-%s-%s-%s.zip", req.Namespace, req.Module, req.Provider, req.Version)

	// Build storage path
	storagePath := fmt.Sprintf("archives/%s/%s/%s/%s/%s",
		req.Namespace, req.Module, req.Provider, req.Version, filename)

	// Create a pipe for streaming
	pipeReader, pipeWriter := io.Pipe()

	// Start writing to storage in a goroutine
	errChan := make(chan error, 1)
	go func() {
		defer pipeWriter.Close()

		// Create ZIP writer that writes to pipe
		zipWriter := zip.NewWriter(pipeWriter)

		// Add each file to the ZIP
		for _, file := range files {
			// Create file entry in ZIP
			writer, err := zipWriter.Create(file.Path())
			if err != nil {
				zipWriter.Close()
				errChan <- fmt.Errorf("failed to create file entry: %w", err)
				return
			}

			// Write file content
			_, err = writer.Write([]byte(file.Content()))
			if err != nil {
				zipWriter.Close()
				errChan <- fmt.Errorf("failed to write file content: %w", err)
				return
			}
		}

		// Close ZIP writer
		err = zipWriter.Close()
		if err != nil {
			errChan <- fmt.Errorf("failed to close ZIP archive: %w", err)
			return
		}

		errChan <- nil
	}()

	// Upload the streaming data to storage
	err = c.storageService.UploadStream(ctx, pipeReader, storagePath)
	if err != nil {
		pipeWriter.Close()
		<-errChan
		return nil, fmt.Errorf("failed to upload archive to storage: %w", err)
	}

	// Wait for ZIP creation to complete
	if err := <-errChan; err != nil {
		return nil, err
	}

	// Get file size from storage
	size, err := c.storageService.GetFileSize(ctx, storagePath)
	if err != nil {
		// Log error but don't fail
		size = 0
	}

	return &GenerateModuleSourceStorageResponse{
		Filename:    filename,
		ContentType: "application/zip",
		Size:        size,
		StoragePath: storagePath,
		Stored:      true,
	}, nil
}

// StreamFromStorage streams a file directly from storage to an HTTP response writer
func (c *GenerateModuleSourceCommand) StreamFromStorage(ctx context.Context, storagePath string, writer io.Writer) error {
	// Use storage service's streaming capability
	err := c.storageService.StreamToHTTPResponse(ctx, storagePath, writer)
	if err != nil {
		return fmt.Errorf("failed to stream file from storage: %w", err)
	}

	return nil
}
