package service

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"archive/tar"
	"archive/zip"
	"compress/gzip"
)

// DefaultArchiveProcessor provides a default implementation of ArchiveProcessor
type DefaultArchiveProcessor struct{}

// NewDefaultArchiveProcessor creates a new default archive processor
func NewDefaultArchiveProcessor() *DefaultArchiveProcessor {
	return &DefaultArchiveProcessor{}
}

// DetectArchiveType detects the type of archive (ZIP or tar.gz)
func (p *DefaultArchiveProcessor) DetectArchiveType(archivePath string) (ArchiveType, error) {
	file, err := os.Open(archivePath)
	if err != nil {
		return 0, fmt.Errorf("failed to open archive: %w", err)
	}
	defer file.Close()

	// Read first few bytes to check magic numbers
	header := make([]byte, 4)
	if _, err := io.ReadFull(file, header); err != nil {
		return 0, fmt.Errorf("failed to read archive header: %w", err)
	}

	// Check for ZIP magic number (PK\x03\x04 or PK\x05\x06)
	if header[0] == 0x50 && header[1] == 0x4B && (header[2] == 0x03 || header[2] == 0x05) {
		return ArchiveTypeZIP, nil
	}

	// Check for gzip magic number
	if header[0] == 0x1F && header[1] == 0x8B {
		// It's gzip, check if it contains a tar archive
		return ArchiveTypeTarGZ, nil
	}

	// Check file extension as fallback
	ext := strings.ToLower(filepath.Ext(archivePath))
	switch ext {
	case ".zip":
		return ArchiveTypeZIP, nil
	case ".gz", ".tgz":
		if strings.HasSuffix(strings.ToLower(archivePath), ".tar.gz") {
			return ArchiveTypeTarGZ, nil
		}
		return ArchiveTypeTarGZ, nil
	}

	return 0, fmt.Errorf("unsupported archive format")
}

// ExtractArchive extracts an archive to the target directory
func (p *DefaultArchiveProcessor) ExtractArchive(ctx context.Context, archivePath string, targetDir string, archiveType ArchiveType) error {
	switch archiveType {
	case ArchiveTypeZIP:
		return p.extractZIP(archivePath, targetDir)
	case ArchiveTypeTarGZ:
		return p.extractTarGz(archivePath, targetDir)
	default:
		return fmt.Errorf("unsupported archive type: %v", archiveType)
	}
}

// ValidateArchive validates that the file is a valid archive
func (p *DefaultArchiveProcessor) ValidateArchive(archivePath string) error {
	_, err := p.DetectArchiveType(archivePath)
	return err
}

// extractZIP extracts a ZIP archive
func (p *DefaultArchiveProcessor) extractZIP(archivePath, targetDir string) error {
	// Open the ZIP file
	reader, err := zip.OpenReader(archivePath)
	if err != nil {
		return fmt.Errorf("failed to open ZIP file: %w", err)
	}
	defer reader.Close()

	// Extract each file
	for _, file := range reader.File {
		filePath := filepath.Join(targetDir, file.Name)

		// Create directory structure if needed
		if file.FileInfo().IsDir() {
			if err := os.MkdirAll(filePath, 0755); err != nil {
				return fmt.Errorf("failed to create directory: %w", err)
			}
			continue
		}

		// Create parent directory if needed
		if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
			return fmt.Errorf("failed to create parent directory: %w", err)
		}

		// Extract file
		fileReader, err := file.Open()
		if err != nil {
			return fmt.Errorf("failed to open file in archive: %w", err)
		}

		destFile, err := os.Create(filePath)
		if err != nil {
			fileReader.Close()
			return fmt.Errorf("failed to create destination file: %w", err)
		}

		if _, err := io.Copy(destFile, fileReader); err != nil {
			destFile.Close()
			fileReader.Close()
			return fmt.Errorf("failed to extract file: %w", err)
		}

		destFile.Close()
		fileReader.Close()
	}

	return nil
}

// extractTarGz extracts a tar.gz archive
func (p *DefaultArchiveProcessor) extractTarGz(archivePath, targetDir string) error {
	// Open the gzip file
	gzipFile, err := os.Open(archivePath)
	if err != nil {
		return fmt.Errorf("failed to open gzip file: %w", err)
	}
	defer gzipFile.Close()

	// Create gzip reader
	gzipReader, err := gzip.NewReader(gzipFile)
	if err != nil {
		return fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer gzipReader.Close()

	// Create tar reader
	tarReader := tar.NewReader(gzipReader)

	// Extract files
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("failed to read tar header: %w", err)
		}

		filePath := filepath.Join(targetDir, header.Name)

		// Create directory structure if needed
		if header.Typeflag == tar.TypeDir {
			if err := os.MkdirAll(filePath, 0755); err != nil {
				return fmt.Errorf("failed to create directory: %w", err)
			}
			continue
		}

		// Skip non-regular files (symlinks, etc.)
		if header.Typeflag != tar.TypeReg {
			continue
		}

		// Create parent directory if needed
		if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
			return fmt.Errorf("failed to create parent directory: %w", err)
		}

		// Create file
		destFile, err := os.Create(filePath)
		if err != nil {
			return fmt.Errorf("failed to create destination file: %w", err)
		}

		if _, err := io.Copy(destFile, tarReader); err != nil {
			destFile.Close()
			return fmt.Errorf("failed to extract file: %w", err)
		}

		destFile.Close()
	}

	return nil
}
