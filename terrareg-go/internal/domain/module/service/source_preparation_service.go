package service

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	configmodel "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/config/model"
	gitService "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/git/service"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/module/model"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/module/repository"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/shared/types"
	infraConfig "github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/config"
	"github.com/rs/zerolog"
)

const (
	SourceTypeUpload SourceType = "upload" // Uploaded file
	SourceTypePath   SourceType = "path"   // Existing path
)

// PrepareFromUploadRequest represents a request to prepare from uploaded file
type PrepareFromUploadRequest struct {
	Namespace  string
	Module     string
	Provider   string
	Version    string
	Source     io.Reader
	SourceSize int64
}

// PrepareFromGitRequest represents a request to prepare from git repository
type PrepareFromGitRequest struct {
	Namespace types.NamespaceName
	Module    types.ModuleName
	Provider  types.ModuleProviderName
	Version   string
	GitTag    *string
}

// PrepareFromArchiveRequest represents a request to prepare from archive file
type PrepareFromArchiveRequest struct {
	Namespace   types.NamespaceName
	Module      types.ModuleName
	Provider    types.ModuleProviderName
	Version     string
	ArchivePath string
}

// PreparedSource represents the result of source preparation
type PreparedSource struct {
	SourcePath string  // Path to prepared source files
	CommitSHA  *string // Git commit SHA (nil for non-git sources)
	SourceType SourceType
	Cleanup    func() // Cleanup function to remove temporary resources
}

// SourcePreparationService prepares module sources for processing
type SourcePreparationService struct {
	moduleProviderRepo repository.ModuleProviderRepository
	gitClient          gitService.GitClient
	storageService     StorageService
	archiveProcessor   ArchiveProcessor
	domainConfig       *configmodel.DomainConfig
	infraConfig        *infraConfig.InfrastructureConfig
	logger             zerolog.Logger
}

// NewSourcePreparationService creates a new source preparation service
func NewSourcePreparationService(
	moduleProviderRepo repository.ModuleProviderRepository,
	gitClient gitService.GitClient,
	storageService StorageService,
	archiveProcessor ArchiveProcessor,
	domainConfig *configmodel.DomainConfig,
	infraConfig *infraConfig.InfrastructureConfig,
	logger zerolog.Logger,
) *SourcePreparationService {
	return &SourcePreparationService{
		moduleProviderRepo: moduleProviderRepo,
		gitClient:          gitClient,
		storageService:     storageService,
		archiveProcessor:   archiveProcessor,
		domainConfig:       domainConfig,
		infraConfig:        infraConfig,
		logger:             logger,
	}
}

// PrepareFromUpload prepares source from an uploaded file
func (s *SourcePreparationService) PrepareFromUpload(
	ctx context.Context,
	req PrepareFromUploadRequest,
) (*PreparedSource, error) {
	s.logger.Debug().
		Str("namespace", req.Namespace).
		Str("module", req.Module).
		Str("provider", req.Provider).
		Str("version", req.Version).
		Msg("Preparing source from upload")

	// 1. Save uploaded file to temp location
	tempFile, err := s.saveUploadedFile(req.Source)
	if err != nil {
		return nil, fmt.Errorf("failed to save uploaded file: %w", err)
	}
	defer os.Remove(tempFile.Name())

	// 2. Create temp directory for extraction
	extractDir, err := s.storageService.MkdirTemp("", "terrareg-upload-*")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp directory: %w", err)
	}

	// 3. Extract archive
	archiveType, err := s.archiveProcessor.DetectArchiveType(tempFile.Name())
	if err != nil {
		s.storageService.RemoveAll(extractDir)
		return nil, fmt.Errorf("failed to detect archive type: %w", err)
	}

	if err := s.archiveProcessor.ExtractArchive(ctx, tempFile.Name(), extractDir, archiveType); err != nil {
		s.storageService.RemoveAll(extractDir)
		return nil, fmt.Errorf("failed to extract archive: %w", err)
	}

	s.logger.Debug().
		Str("extract_dir", extractDir).
		Msg("Successfully extracted uploaded archive")

	// 4. Return prepared source with cleanup
	return &PreparedSource{
		SourcePath: extractDir,
		CommitSHA:  nil,
		SourceType: SourceTypeUpload,
		Cleanup: func() {
			s.logger.Debug().Str("path", extractDir).Msg("Cleaning up upload source directory")
			s.storageService.RemoveAll(extractDir)
		},
	}, nil
}

// PrepareFromGit prepares source from a git repository
func (s *SourcePreparationService) PrepareFromGit(
	ctx context.Context,
	req PrepareFromGitRequest,
) (*PreparedSource, error) {
	logEvent := s.logger.Debug().
		Str("namespace", string(req.Namespace)).
		Str("module", string(req.Module)).
		Str("provider", string(req.Provider)).
		Str("version", req.Version)
	if req.GitTag != nil {
		logEvent.Str("git_tag", *req.GitTag)
	}
	logEvent.Msg("Preparing source from git")

	// 1. Find module provider to get git configuration
	moduleProvider, err := s.moduleProviderRepo.FindByNamespaceModuleProvider(
		ctx, req.Namespace, req.Module, req.Provider,
	)
	if err != nil {
		return nil, fmt.Errorf("module provider not found: %w", err)
	}

	if moduleProvider == nil {
		return nil, fmt.Errorf("module provider not found: %s/%s/%s", req.Namespace, req.Module, req.Provider)
	}

	if moduleProvider.RepoCloneURLTemplate() == nil {
		return nil, fmt.Errorf("module provider is not git-based")
	}

	// 2. Build clone URL
	cloneURL := s.buildCloneURL(req, moduleProvider)

	// 3. Create temp directory for clone
	cloneDir, err := s.storageService.MkdirTemp("", "terrareg-git-*")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp directory: %w", err)
	}

	s.logger.Debug().
		Str("clone_url", cloneURL).
		Str("clone_dir", cloneDir).
		Msg("Cloning git repository")

	// 4. Clone repository
	cloneOptions := gitService.CloneOptions{
		Timeout:  time.Duration(s.infraConfig.GitCloneTimeout) * time.Second,
		NeedTags: req.GitTag != nil,
	}

	if err := s.gitClient.CloneWithOptions(ctx, cloneURL, cloneDir, cloneOptions); err != nil {
		s.storageService.RemoveAll(cloneDir)
		return nil, fmt.Errorf("failed to clone repository: %w", err)
	}

	s.logger.Debug().
		Str("clone_dir", cloneDir).
		Msg("Successfully cloned repository")

	// 5. Checkout tag if provided
	if req.GitTag != nil {
		gitTag := s.formatGitTag(*req.GitTag, moduleProvider)
		s.logger.Debug().
			Str("git_tag", gitTag).
			Str("directory", cloneDir).
			Msg("Checking out git tag")

		if err := s.gitClient.Checkout(ctx, cloneDir, gitTag); err != nil {
			s.storageService.RemoveAll(cloneDir)
			return nil, fmt.Errorf("failed to checkout tag '%s': %w", gitTag, err)
		}

		s.logger.Debug().
			Str("git_tag", gitTag).
			Msg("Successfully checked out git tag")
	}

	// 6. Extract commit SHA if configured
	var commitSHA *string
	if s.domainConfig.ModuleVersionUseGitCommit {
		if sha, err := s.gitClient.GetCommitSHA(ctx, cloneDir); err == nil {
			commitSHA = &sha
			s.logger.Debug().Str("commit_sha", sha).Msg("Extracted git commit SHA")
		} else {
			s.logger.Debug().Err(err).Msg("Failed to get commit SHA, continuing without it")
		}
	}

	// 7. Determine source directory (handle git_path)
	sourceDir := cloneDir
	if gitPath := moduleProvider.GitPath(); gitPath != nil && *gitPath != "" {
		sourceDir = filepath.Join(cloneDir, *gitPath)
		s.logger.Debug().
			Str("git_path", *gitPath).
			Str("source_dir", sourceDir).
			Msg("Using git_path for source directory")
	}

	return &PreparedSource{
		SourcePath: sourceDir,
		CommitSHA:  commitSHA,
		SourceType: SourceTypeGit,
		Cleanup: func() {
			s.logger.Debug().Str("path", cloneDir).Msg("Cleaning up git clone directory")
			s.storageService.RemoveAll(cloneDir)
		},
	}, nil
}

// PrepareFromArchive prepares source from an archive file
func (s *SourcePreparationService) PrepareFromArchive(
	ctx context.Context,
	req PrepareFromArchiveRequest,
) (*PreparedSource, error) {
	s.logger.Debug().
		Str("namespace", string(req.Namespace)).
		Str("module", string(req.Module)).
		Str("provider", string(req.Provider)).
		Str("version", req.Version).
		Str("archive_path", req.ArchivePath).
		Msg("Preparing source from archive")

	// 1. Detect archive type
	archiveType, err := s.archiveProcessor.DetectArchiveType(req.ArchivePath)
	if err != nil {
		return nil, fmt.Errorf("failed to detect archive type: %w", err)
	}

	// 2. Create temp directory for extraction
	extractDir, err := s.storageService.MkdirTemp("", "terrareg-archive-*")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp directory: %w", err)
	}

	// 3. Extract archive
	if err := s.archiveProcessor.ExtractArchive(ctx, req.ArchivePath, extractDir, archiveType); err != nil {
		s.storageService.RemoveAll(extractDir)
		return nil, fmt.Errorf("failed to extract archive: %w", err)
	}

	s.logger.Debug().
		Str("extract_dir", extractDir).
		Msg("Successfully extracted archive")

	return &PreparedSource{
		SourcePath: extractDir,
		CommitSHA:  nil,
		SourceType: SourceTypeArchive,
		Cleanup: func() {
			s.logger.Debug().Str("path", extractDir).Msg("Cleaning up archive source directory")
			s.storageService.RemoveAll(extractDir)
		},
	}, nil
}

// saveUploadedFile saves uploaded file to a temporary location
func (s *SourcePreparationService) saveUploadedFile(source io.Reader) (*os.File, error) {
	tempFile, err := os.CreateTemp("", "terrareg-upload-*.zip")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp file: %w", err)
	}

	_, err = io.Copy(tempFile, source)
	if err != nil {
		tempFile.Close()
		return nil, fmt.Errorf("failed to save upload: %w", err)
	}

	return tempFile, nil
}

// buildCloneURL builds the clone URL from template
func (s *SourcePreparationService) buildCloneURL(req PrepareFromGitRequest, moduleProvider *model.ModuleProvider) string {
	cloneURLTemplate := moduleProvider.RepoCloneURLTemplate()
	if cloneURLTemplate == nil {
		return ""
	}

	// Simple template replacement - in production, use a proper template library
	cloneURL := *cloneURLTemplate
	cloneURL = strings.ReplaceAll(cloneURL, "{namespace}", string(req.Namespace))
	cloneURL = strings.ReplaceAll(cloneURL, "{name}", string(req.Module))
	cloneURL = strings.ReplaceAll(cloneURL, "{module}", string(req.Module))
	cloneURL = strings.ReplaceAll(cloneURL, "{provider}", string(req.Provider))

	return cloneURL
}

// formatGitTag formats the git tag according to module provider's git tag format
func (s *SourcePreparationService) formatGitTag(gitTag string, moduleProvider *model.ModuleProvider) string {
	gitTagFormat := moduleProvider.GitTagFormat()
	if gitTagFormat == nil || *gitTagFormat == "" {
		return gitTag
	}

	// Simple template replacement
	formatted := *gitTagFormat
	formatted = strings.ReplaceAll(formatted, "{version}", gitTag)
	formatted = strings.ReplaceAll(formatted, "{module}", string(moduleProvider.Module()))
	formatted = strings.ReplaceAll(formatted, "{provider}", string(moduleProvider.Provider()))

	return formatted
}
