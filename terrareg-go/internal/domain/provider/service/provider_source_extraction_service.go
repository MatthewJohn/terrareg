package service

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"archive/tar"
	"compress/gzip"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/provider"
	providermodel "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/provider_source/model"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/persistence/sqldb"
)

// ProviderSourceExtractionService handles extracting source code from provider archives
// Python reference: provider_extractor.py::ProviderExtractor._obtain_source_code
type ProviderSourceExtractionService struct {
	// Dependencies would be injected here
}

// NewProviderSourceExtractionService creates a new source extraction service
func NewProviderSourceExtractionService() *ProviderSourceExtractionService {
	return &ProviderSourceExtractionService{}
}

// ExtractedSource represents an extracted source directory with cleanup function
type ExtractedSource struct {
	SourceDir string
	Cleanup   func() error
}

// ObtainSourceCode obtains and extracts source code from a provider release
// Python reference: provider_extractor.py::ProviderExtractor._obtain_source_code
func (s *ProviderSourceExtractionService) ObtainSourceCode(
	ctx context.Context,
	providerEntity *provider.Provider,
	releaseMetadata *providermodel.RepositoryReleaseMetadata,
	repository *sqldb.RepositoryDB,
	getReleaseArchiveFunc func(ctx context.Context, repo *sqldb.RepositoryDB, releaseMetadata *providermodel.RepositoryReleaseMetadata, accessToken string) ([]byte, string, error),
	accessToken string,
) (*ExtractedSource, error) {
	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "provider-extraction-*")
	if err != nil {
		return nil, fmt.Errorf("%w: failed to create temp directory: %s", provider.ErrUnableToObtainSource, err)
	}

	// Create cleanup function
	cleanup := func() error {
		return os.RemoveAll(tempDir)
	}

	// Create child directory for the provider name
	sourceDir := filepath.Join(tempDir, providerEntity.Name())
	if err := os.Mkdir(sourceDir, 0755); err != nil {
		cleanup()
		return nil, fmt.Errorf("%w: failed to create source directory: %s", provider.ErrUnableToObtainSource, err)
	}

	// Obtain archive of release
	archiveData, extractSubdirectory, err := getReleaseArchiveFunc(ctx, repository, releaseMetadata, accessToken)
	if err != nil {
		cleanup()
		return nil, fmt.Errorf("%w: failed to get release archive: %s", provider.ErrUnableToObtainSource, err)
	}

	if len(archiveData) == 0 {
		cleanup()
		return nil, fmt.Errorf("%w: unable to obtain release source for provider release", provider.ErrUnableToObtainSource)
	}

	// Extract archive
	if err := s.extractArchive(archiveData, sourceDir); err != nil {
		cleanup()
		return nil, fmt.Errorf("%w: %s", provider.ErrInvalidTarArchive, err)
	}

	// If the repository provider uses a sub-directory for the source, obtain this
	if extractSubdirectory != "" {
		sourceDir = filepath.Join(sourceDir, extractSubdirectory)
	}

	// Check if source directory is named after the repository
	// (apparently this is important for tfplugindocs)
	// and if not, rename it
	repoName := repository.Name
	if filepath.Base(sourceDir) != *repoName {
		newSourceDir := filepath.Join(filepath.Dir(sourceDir), *repoName)
		// Check if new source dir already exists
		if _, err := os.Stat(newSourceDir); err == nil {
			// Directory exists, remove it first
			os.RemoveAll(newSourceDir)
		}
		if err := os.Rename(sourceDir, newSourceDir); err != nil {
			cleanup()
			return nil, fmt.Errorf("%w: failed to rename source directory: %s", provider.ErrUnableToObtainSource, err)
		}
		sourceDir = newSourceDir
	}

	// Setup git repository inside directory (required for tfplugindocs)
	if err := s.setupGitRepository(sourceDir, repository); err != nil {
		cleanup()
		return nil, fmt.Errorf("%w: failed to setup git repository: %s", provider.ErrUnableToObtainSource, err)
	}

	return &ExtractedSource{
		SourceDir: sourceDir,
		Cleanup:   cleanup,
	}, nil
}

// extractArchive extracts a tar.gz archive to a directory
func (s *ProviderSourceExtractionService) extractArchive(archiveData []byte, destDir string) error {
	// Create a reader from the archive data
	archiveReader := bytes.NewReader(archiveData)

	// Create gzip reader
	gzipReader, err := gzip.NewReader(archiveReader)
	if err != nil {
		return fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer gzipReader.Close()

	// Create tar reader
	tarReader := tar.NewReader(gzipReader)

	// Extract each entry
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break // End of archive
		}
		if err != nil {
			return fmt.Errorf("failed to read tar entry: %w", err)
		}

		// SECURITY: Check for path traversal attempts
		// Python reference: provider_extractor.py::ProviderExtractor._obtain_source_code
		if filepath.IsAbs(header.Name) || strings.Contains(header.Name, "..") {
			return fmt.Errorf("illegal tar archive entry: %s", header.Name)
		}

		// Construct the destination path
		destPath := filepath.Join(destDir, header.Name)

		// Create directory or file
		if header.Typeflag == tar.TypeDir {
			// Create directory
			if err := os.MkdirAll(destPath, os.FileMode(header.Mode)); err != nil {
				return fmt.Errorf("failed to create directory %s: %w", destPath, err)
			}
		} else if header.Typeflag == tar.TypeReg || header.Typeflag == tar.TypeRegA {
			// Create parent directory if needed
			destDirPath := filepath.Dir(destPath)
			if err := os.MkdirAll(destDirPath, 0755); err != nil {
				return fmt.Errorf("failed to create parent directory %s: %w", destDirPath, err)
			}

			// Create file
			outFile, err := os.OpenFile(destPath, os.O_CREATE|os.O_WRONLY, os.FileMode(header.Mode))
			if err != nil {
				return fmt.Errorf("failed to create file %s: %w", destPath, err)
			}

			// Write file content
			_, err = io.Copy(outFile, tarReader)
			outFile.Close()

			if err != nil {
				return fmt.Errorf("failed to write file %s: %w", destPath, err)
			}
		}
		// Note: Symlinks and other file types are not handled for security reasons
	}

	return nil
}

// setupGitRepository sets up a git repository in the source directory
// This is required for tfplugindocs to work properly
// Python reference: provider_extractor.py::ProviderExtractor._obtain_source_code
func (s *ProviderSourceExtractionService) setupGitRepository(sourceDir string, repository *sqldb.RepositoryDB) error {
	// Create a clean environment for git
	// Remove any GIT_* environment variables to avoid conflicts
	env := s.cleanGitEnv(os.Environ())

	// Run git commands in the source directory
	commands := [][]string{
		{"git", "init"},
		{"git", "config", "user.email", "terrareg@localhost"},
		{"git", "config", "user.name", "Terrareg"},
		{"git", "add", "."},
		{"git", "commit", "-m", "Initial commit"},
	}

	for _, cmd := range commands {
		if err := s.runGitCommand(sourceDir, env, cmd); err != nil {
			return fmt.Errorf("failed to run %s: %w", strings.Join(cmd, " "), err)
		}
	}

	// Add remote if clone URL is available
	if repository.CloneURL != nil && *repository.CloneURL != "" {
		cloneURL := *repository.CloneURL
		// Remove .git suffix if present
		if strings.HasSuffix(cloneURL, ".git") {
			re := regexp.MustCompile(`\.git$`)
			cloneURL = re.ReplaceAllString(cloneURL, "")
		}

		cmd := []string{"git", "remote", "add", "origin", cloneURL}
		if err := s.runGitCommand(sourceDir, env, cmd); err != nil {
			// Non-fatal error, log but don't fail
			fmt.Printf("Warning: failed to add git remote: %v\n", err)
		}
	}

	return nil
}

// cleanGitEnv removes GIT_* environment variables to avoid conflicts
// Python reference: provider_extractor.py::ProviderExtractor._obtain_source_code
func (s *ProviderSourceExtractionService) cleanGitEnv(env []string) []string {
	cleanEnv := make([]string, 0, len(env))
	for _, e := range env {
		if !strings.HasPrefix(e, "GIT_") {
			cleanEnv = append(cleanEnv, e)
		}
	}
	return cleanEnv
}

// runGitCommand runs a git command in the specified directory
func (s *ProviderSourceExtractionService) runGitCommand(dir string, env []string, cmd []string) error {
	command := exec.Command(cmd[0], cmd[1:]...)
	command.Dir = dir
	command.Env = env

	output, err := command.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git command failed: %s\noutput: %s", err, string(output))
	}

	return nil
}

// ValidateProviderStructure validates that the source contains a valid provider
// Python reference: provider_extractor.py (implicit validation during extraction)
func (s *ProviderSourceExtractionService) ValidateProviderStructure(sourcePath string) error {
	// Check for main.go
	mainGoPath := filepath.Join(sourcePath, "main.go")
	if _, err := os.Stat(mainGoPath); os.IsNotExist(err) {
		return fmt.Errorf("main.go not found - not a valid Go provider")
	}

	// Check for go.mod
	goModPath := filepath.Join(sourcePath, "go.mod")
	if _, err := os.Stat(goModPath); os.IsNotExist(err) {
		return fmt.Errorf("go.mod not found - not a valid Go module")
	}

	return nil
}

// FindProviderMainFile looks for the main.go file and returns its path
func (s *ProviderSourceExtractionService) FindProviderMainFile(sourcePath string) (string, error) {
	mainGoPath := filepath.Join(sourcePath, "main.go")
	if _, err := os.Stat(mainGoPath); err != nil {
		return "", fmt.Errorf("main.go not found: %w", err)
	}
	return mainGoPath, nil
}
