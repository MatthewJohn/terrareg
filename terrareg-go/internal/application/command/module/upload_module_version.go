package module

import (
	"archive/zip"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/terrareg/terrareg/internal/config"
	"github.com/terrareg/terrareg/internal/domain/module/repository"
)

// UploadModuleVersionCommand handles uploading module source code
type UploadModuleVersionCommand struct {
	moduleProviderRepo repository.ModuleProviderRepository
	config             *config.Config
}

// NewUploadModuleVersionCommand creates a new command
func NewUploadModuleVersionCommand(
	moduleProviderRepo repository.ModuleProviderRepository,
	config *config.Config,
) *UploadModuleVersionCommand {
	return &UploadModuleVersionCommand{
		moduleProviderRepo: moduleProviderRepo,
		config:             config,
	}
}

// UploadModuleVersionRequest represents the upload request
type UploadModuleVersionRequest struct {
	Namespace string
	Module    string
	Provider  string
	Version   string
	Source    io.Reader
	SourceSize int64
}

// Execute uploads and extracts a module version
func (c *UploadModuleVersionCommand) Execute(ctx context.Context, req UploadModuleVersionRequest) error {
	// Find the module provider
	moduleProvider, err := c.moduleProviderRepo.FindByNamespaceModuleProvider(
		ctx, req.Namespace, req.Module, req.Provider,
	)
	if err != nil {
		return fmt.Errorf("module provider not found: %w", err)
	}

	// Find the module version within the aggregate
	version, err := moduleProvider.GetVersion(req.Version)
	if err != nil {
		return fmt.Errorf("module version not found: %w", err)
	}

	// Create upload directory if it doesn't exist
	uploadDir := filepath.Join(c.config.UploadDirectory, fmt.Sprintf("%d", version.ID()))
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		return fmt.Errorf("failed to create upload directory: %w", err)
	}

	// Save the uploaded file temporarily
	tempFile := filepath.Join(uploadDir, "source.zip")
	f, err := os.Create(tempFile)
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	defer f.Close()

	// Copy the uploaded content
	if _, err := io.Copy(f, req.Source); err != nil {
		return fmt.Errorf("failed to save upload: %w", err)
	}
	f.Close()

	// Extract the ZIP file
	extractDir := filepath.Join(c.config.DataDirectory, "modules", req.Namespace, req.Module, req.Provider, req.Version)
	if err := extractZip(tempFile, extractDir); err != nil {
		return fmt.Errorf("failed to extract module: %w", err)
	}

	// Publish the module version
	// In the future, this would parse terraform files, README, etc.
	if err := version.Publish(); err != nil {
		return fmt.Errorf("failed to publish version: %w", err)
	}

	// Save the aggregate (persistence of version is handled by the aggregate)
	if err := c.moduleProviderRepo.Save(ctx, moduleProvider); err != nil {
		return fmt.Errorf("failed to save module provider: %w", err)
	}

	return nil
}

// extractZip extracts a ZIP file to a destination directory
func extractZip(src, dest string) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer r.Close()

	// Create destination directory
	if err := os.MkdirAll(dest, 0755); err != nil {
		return err
	}

	// Extract each file
	for _, f := range r.File {
		// Prevent path traversal attacks
		if strings.Contains(f.Name, "..") {
			continue
		}

		filePath := filepath.Join(dest, f.Name)

		if f.FileInfo().IsDir() {
			os.MkdirAll(filePath, 0755)
			continue
		}

		// Create parent directories
		if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
			return err
		}

		// Create the file
		outFile, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return err
		}

		rc, err := f.Open()
		if err != nil {
			outFile.Close()
			return err
		}

		_, err = io.Copy(outFile, rc)
		outFile.Close()
		rc.Close()

		if err != nil {
			return err
		}
	}

	return nil
}
