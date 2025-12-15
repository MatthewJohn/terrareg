package module

import (
	"archive/zip"
	"bytes"
	"context"
	"fmt"

	moduleRepo "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/module/repository"
	moduleService "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/module/service"
)

// GenerateModuleSourceCommand handles generating source archives for module versions
type GenerateModuleSourceCommand struct {
	moduleProviderRepo moduleRepo.ModuleProviderRepository
	moduleFileService  *moduleService.ModuleFileService
}

// NewGenerateModuleSourceCommand creates a new generate module source command
func NewGenerateModuleSourceCommand(
	moduleProviderRepo moduleRepo.ModuleProviderRepository,
	moduleFileService *moduleService.ModuleFileService,
) *GenerateModuleSourceCommand {
	return &GenerateModuleSourceCommand{
		moduleProviderRepo: moduleProviderRepo,
		moduleFileService:  moduleFileService,
	}
}

// GenerateModuleSourceRequest represents a request to generate module source
type GenerateModuleSourceRequest struct {
	Namespace string
	Module    string
	Provider  string
	Version   string
}

// GenerateModuleSourceResponse represents the generated module source archive
type GenerateModuleSourceResponse struct {
	Filename    string
	ContentType string
	Content     []byte
	ContentSize int64
}

// Execute generates a source archive for the module version
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
		ContentSize: int64(buf.Len()),
	}, nil
}
