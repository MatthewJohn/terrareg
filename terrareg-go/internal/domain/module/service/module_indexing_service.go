package service

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	gitService "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/git/service"
	storageService "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/storage/service"
	"github.com/rs/zerolog"
)

// ModuleIndexingService orchestrates the complete module indexing workflow
// This implements the critical Git clone → temp dir → process → upload archives → cleanup workflow
type ModuleIndexingService interface {
	// IndexModuleVersion performs complete module indexing from Git repository
	IndexModuleVersion(ctx context.Context, req *IndexModuleVersionRequest) error
}

// IndexModuleVersionRequest contains all parameters for module indexing
type IndexModuleVersionRequest struct {
	// Module identification
	Namespace string `json:"namespace"`
	Module    string `json:"module"`
	Provider  string `json:"provider"`
	Version   string `json:"version"`

	// Git repository information
	GitURL  string `json:"git_url"`
	GitTag  string `json:"git_tag"`
	GitPath string `json:"git_path"` // Optional: Path within repository to module

	// Processing options
	ArchiveTarget string `json:"archive_target"` // "all" or "git_path_only"
}

// ModuleIndexingServiceImpl implements ModuleIndexingService
type ModuleIndexingServiceImpl struct {
	gitService       gitService.GitService
	storageWorkflow  storageService.StorageWorkflowService
	moduleProcessor  ModuleProcessorService
	archiveGenerator ArchiveGenerationService
	logger           zerolog.Logger
}

// NewModuleIndexingService creates a new module indexing service
func NewModuleIndexingServiceImpl(
	gitService gitService.GitService,
	storageWorkflow storageService.StorageWorkflowService,
	moduleProcessor ModuleProcessorService,
	archiveGenerator ArchiveGenerationService,
	logger zerolog.Logger,
) *ModuleIndexingServiceImpl {
	return &ModuleIndexingServiceImpl{
		gitService:       gitService,
		storageWorkflow:  storageWorkflow,
		moduleProcessor:  moduleProcessor,
		archiveGenerator: archiveGenerator,
		logger:           logger,
	}
}

// IndexModuleVersion performs the complete module indexing workflow
// This implements the critical workflow: Git clone → temp dir → process → upload archives → cleanup
func (s *ModuleIndexingServiceImpl) IndexModuleVersion(ctx context.Context, req *IndexModuleVersionRequest) error {
	s.logger.Info().
		Str("namespace", req.Namespace).
		Str("module", req.Module).
		Str("provider", req.Provider).
		Str("version", req.Version).
		Str("git_url", req.GitURL).
		Str("git_tag", req.GitTag).
		Msg("Starting module indexing workflow")

	// 1. Create temporary directory for processing
	tempDir, cleanup, err := s.storageWorkflow.CreateProcessingDirectory(ctx, fmt.Sprintf("module_%s_%s_", req.Module, req.Version))
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to create processing directory")
		return fmt.Errorf("failed to create processing directory: %w", err)
	}
	defer cleanup()

	s.logger.Debug().Str("temp_dir", tempDir).Msg("Created processing directory")

	// 2. Clone repository to temporary directory
	cloneDir := filepath.Join(tempDir, "clone")
	cloneOptions := &gitService.CloneOptions{
		Timeout: 5 * time.Minute,
	}

	if err := s.gitService.CloneRepository(ctx, req.GitURL, req.GitTag, cloneDir, cloneOptions); err != nil {
		s.logger.Error().Err(err).Msg("Failed to clone repository")
		return fmt.Errorf("failed to clone repository: %w", err)
	}

	s.logger.Debug().Str("clone_dir", cloneDir).Msg("Successfully cloned repository")

	// 3. Get repository provenance information
	commitSHA, err := s.gitService.GetCommitSHA(ctx, cloneDir)
	if err != nil {
		s.logger.Warn().Err(err).Msg("Failed to get commit SHA, continuing anyway")
		commitSHA = ""
	} else {
		s.logger.Info().Str("commit_sha", commitSHA).Msg("Repository commit SHA")
	}

	// 4. Determine source directory for processing
	sourceDir := cloneDir
	if req.GitPath != "" {
		sourceDir = filepath.Join(cloneDir, req.GitPath)

		// Validate path exists
		if _, err := os.Stat(sourceDir); os.IsNotExist(err) {
			return fmt.Errorf("git path does not exist in repository: %s", req.GitPath)
		}
	}

	// 5. Process module (terraform docs, tfsec, metadata extraction)
	s.logger.Debug().Str("source_dir", sourceDir).Msg("Starting module processing")

	processingMetadata := &ModuleProcessingMetadata{
		ModuleVersionID: 0, // Set from database in real implementation
		GitTag:          req.GitTag,
		GitURL:          req.GitURL,
		GitPath:         req.GitPath,
		CommitSHA:       commitSHA,
	}

	result, err := s.moduleProcessor.ProcessModule(ctx, sourceDir, processingMetadata)
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to process module")
		return fmt.Errorf("failed to process module: %w", err)
	}

	s.logger.Info().
		Int("readme_length", len(result.ReadmeContent)).
		Int("submodules", len(result.Submodules)).
		Int("examples", len(result.Examples)).
		Msg("Module processing completed")

	// 6. Generate archives (tar.gz and zip) matching Python
	archiveDir := filepath.Join(tempDir, "archives")
	if err := os.MkdirAll(archiveDir, 0755); err != nil {
		s.logger.Error().Err(err).Msg("Failed to create archive directory")
		return fmt.Errorf("failed to create archive directory: %w", err)
	}

	archivePaths, err := s.archiveGenerator.GenerateArchives(ctx, sourceDir, archiveDir)
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to generate archives")
		return fmt.Errorf("failed to generate archives: %w", err)
	}

	s.logger.Info().
		Strs("archives", archivePaths).
		Msg("Successfully generated archives")

	// 7. Upload archives to storage (local or S3)
	if err := s.storageWorkflow.StoreModuleArchives(ctx, archiveDir, req.Namespace, req.Module, req.Provider, req.Version); err != nil {
		s.logger.Error().Err(err).Msg("Failed to upload archives to storage")
		return fmt.Errorf("failed to upload archives: %w", err)
	}

	s.logger.Info().
		Str("namespace", req.Namespace).
		Str("module", req.Module).
		Str("provider", req.Provider).
		Str("version", req.Version).
		Msg("Successfully completed module indexing workflow")

	return nil
}

// ModuleProcessorService interface for processing modules
// This handles the actual module analysis and extraction
type ModuleProcessorService interface {
	ProcessModule(ctx context.Context, moduleDir string, metadata *ModuleProcessingMetadata) (*ModuleProcessingResult, error)
	ValidateModuleStructure(ctx context.Context, moduleDir string) error
	ExtractMetadata(ctx context.Context, moduleDir string) (*ModuleMetadata, error)
}

// ModuleProcessingMetadata contains metadata for module processing
type ModuleProcessingMetadata struct {
	ModuleVersionID int
	GitTag          string
	GitURL          string
	GitPath         string
	CommitSHA       string
}

// ModuleProcessingResult contains results from module processing
type ModuleProcessingResult struct {
	ModuleMetadata   *ModuleMetadata `json:"module_metadata"`
	Submodules       []SubmoduleInfo `json:"submodules"`
	Examples         []ExampleInfo   `json:"examples"`
	ReadmeContent    string          `json:"readme_content"`
	VariableTemplate string          `json:"variable_template"`
	ProcessedFiles   []string        `json:"processed_files"`
}

// ModuleMetadata contains extracted module metadata
type ModuleMetadata struct {
	Name         string           `json:"name"`
	Description  string           `json:"description"`
	Version      string           `json:"version"`
	Providers    []ProviderInfo   `json:"providers"`
	Variables    []VariableInfo   `json:"variables"`
	Outputs      []OutputInfo     `json:"outputs"`
	Resources    []ResourceInfo   `json:"resources"`
	Dependencies []DependencyInfo `json:"dependencies"`
}

// ProviderInfo represents a Terraform provider
type ProviderInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
	Source  string `json:"source"`
}

// VariableInfo represents a Terraform variable
type VariableInfo struct {
	Name        string      `json:"name"`
	Type        string      `json:"type"`
	Description string      `json:"description"`
	Default     interface{} `json:"default"`
	Required    bool        `json:"required"`
}

// OutputInfo represents a Terraform output
type OutputInfo struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Value       interface{} `json:"value"`
	Sensitive   bool        `json:"sensitive"`
}

// ResourceInfo represents a Terraform resource
type ResourceInfo struct {
	Type string `json:"type"`
	Name string `json:"name"`
}

// DependencyInfo represents a module dependency
type DependencyInfo struct {
	Source  string `json:"source"`
	Version string `json:"version"`
}

// SubmoduleInfo represents a Terraform submodule
type SubmoduleInfo struct {
	Path    string `json:"path"`
	Source  string `json:"source"`
	Version string `json:"version"`
}

// ExampleInfo represents a Terraform example
type ExampleInfo struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Files       []string `json:"files"`
}

// ArchiveGenerationService interface for generating archives
type ArchiveGenerationService interface {
	GenerateArchives(ctx context.Context, sourceDir string, outputDir string) ([]string, error)
	GenerateTarGz(ctx context.Context, sourceDir, outputPath string) error
	GenerateZip(ctx context.Context, sourceDir, outputPath string) error
}
