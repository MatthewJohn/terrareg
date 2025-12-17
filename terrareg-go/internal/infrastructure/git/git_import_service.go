package git

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/module"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/module/model"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/module/repository"
	moduleService "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/module/service"
	infraConfig "github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/config"
)

// GitImportService handles importing modules from Git repositories
type GitImportService struct {
	moduleProviderRepo repository.ModuleProviderRepository
	storageService     moduleService.StorageService
	infraConfig        *infraConfig.InfrastructureConfig
}

// NewGitImportService creates a new Git import service
func NewGitImportService(
	moduleProviderRepo repository.ModuleProviderRepository,
	storageService moduleService.StorageService,
	infraConfig *infraConfig.InfrastructureConfig,
) *GitImportService {
	return &GitImportService{
		moduleProviderRepo: moduleProviderRepo,
		storageService:     storageService,
		infraConfig:        infraConfig,
	}
}

// ImportModuleVersionResult represents the result of a Git import
type ImportModuleVersionResult struct {
	Version    string
	GitSHA     string
	CommitDate string
	Message    string
}

// Execute imports a module version from Git
func (s *GitImportService) Execute(ctx context.Context, req module.ImportModuleVersionRequest) (*ImportModuleVersionResult, error) {
	// Find the module provider
	moduleProvider, err := s.moduleProviderRepo.FindByNamespaceModuleProvider(
		ctx, req.Namespace, req.Module, req.Provider,
	)
	if err != nil {
		return nil, fmt.Errorf("module provider not found: %w", err)
	}

	// Find or create the module version
	var moduleVersion *model.ModuleVersion

	if req.Version == nil {
		if req.GitTag == nil {
			return nil, fmt.Errorf("either version or git_tag must be provided")
		}
		// Derive version from git tag
		version, err := s.deriveVersionFromGitTag(*req.GitTag, moduleProvider)
		if err != nil {
			return nil, fmt.Errorf("failed to derive version from git tag: %w", err)
		}
		if version == "" {
			return nil, fmt.Errorf("could not derive version from git tag: %s", *req.GitTag)
		}
		req.Version = &version
	}

	moduleVersion, err = moduleProvider.GetVersion(*req.Version)
	if err != nil {
		// Create new version if it doesn't exist
		moduleVersion, err = moduleProvider.PublishVersion(*req.Version, nil, false)
		if err != nil {
			return nil, fmt.Errorf("failed to create module version: %w", err)
		}
	}

	// Ensure that module provider has a repository URL configured
	gitCloneURL := moduleProvider.RepoCloneURLTemplate()
	if gitCloneURL == nil || *gitCloneURL == "" {
		return nil, fmt.Errorf("module provider is not configured with a repository")
	}

	// Create temporary directory for cloning
	tmpDir, err := os.MkdirTemp("", "terrareg-git-import-")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp dir for git clone: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	// Determine git reference for cloning
	var gitRef string
	if req.Version != nil {
		// Check if module provider can index by version
		gitTagFormat := moduleProvider.GitTagFormat()
		if gitTagFormat == nil || *gitTagFormat == "" {
			return nil, fmt.Errorf("module provider is not configured with a git tag format containing a {version} placeholder")
		}
		// Generate git tag from format
		gitRef = strings.ReplaceAll(*gitTagFormat, "{version}", *req.Version)
	} else if req.GitTag != nil {
		gitRef = *req.GitTag
	}

	// Clone repository
	if err := s.cloneRepository(*gitCloneURL, tmpDir, gitRef); err != nil {
		return nil, fmt.Errorf("failed to clone repository: %w", err)
	}

	// Get git commit SHA and date
	gitSHA, err := s.getGitCommitSHA(tmpDir)
	if err != nil {
		return nil, fmt.Errorf("failed to get git commit SHA: %w", err)
	}

	commitDate, err := s.getGitCommitDate(tmpDir)
	if err != nil {
		return nil, fmt.Errorf("failed to get git commit date: %w", err)
	}

	// Extract module files from cloned repository
	if err := s.extractModuleFromClone(tmpDir, moduleVersion); err != nil {
		return nil, fmt.Errorf("failed to extract module from clone: %w", err)
	}

	// Update module version with git metadata
	moduleVersion.SetGitInfo(&gitSHA, moduleVersion.GitPath(), moduleVersion.ArchiveGitPath())

	// Save module version and provider
	if err := s.moduleProviderRepo.Save(ctx, moduleProvider); err != nil {
		return nil, fmt.Errorf("failed to save module version: %w", err)
	}

	return &ImportModuleVersionResult{
		Version:    moduleVersion.Version().String(),
		GitSHA:     gitSHA,
		CommitDate: commitDate.Format(time.RFC3339),
		Message:    "Module imported successfully",
	}, nil
}

// deriveVersionFromGitTag derives a version string from a git tag
func (s *GitImportService) deriveVersionFromGitTag(gitTag string, moduleProvider *model.ModuleProvider) (string, error) {
	// Get git tag format from module provider
	gitTagFormat := moduleProvider.GitTagFormat()
	if gitTagFormat == nil || *gitTagFormat == "" {
		// If no format is specified, try to match the git tag directly as version
		return gitTag, nil
	}

	// Convert tag format to regex pattern
	tagRegex, err := s.getTagRegex(*gitTagFormat)
	if err != nil {
		return "", fmt.Errorf("failed to convert tag format to regex: %w", err)
	}

	// Match git tag against regex to extract version
	return s.getVersionFromRegex(tagRegex, gitTag), nil
}

// getTagRegex converts tag format string to regex pattern
func (s *GitImportService) getTagRegex(format string) (*regexp.Regexp, error) {
	// Replace placeholders with named regex groups
	// This mirrors the Python implementation's approach

	// Escape regex special characters in the format first
	escaped := regexp.QuoteMeta(format)

	// Replace the escaped placeholders back with regex groups
	replacements := map[string]string{
		`\{version\}`: `(?P<version>[^}]+)`,
		`\{major\}`:   `(?P<major>\d+)`,
		`\{minor\}`:   `(?P<minor>\d+)`,
		`\{patch\}`:   `(?P<patch>\d+)`,
		`\{build\}`:   `(?P<build>[^}]+)`,
	}

	for placeholder, regex := range replacements {
		escaped = strings.ReplaceAll(escaped, placeholder, regex)
	}

	// Add anchors to match the entire string
	pattern := "^" + escaped + "$"

	return regexp.Compile(pattern)
}

// getVersionFromRegex extracts version from git tag using regex
func (s *GitImportService) getVersionFromRegex(regex *regexp.Regexp, gitTag string) string {
	if !regex.MatchString(gitTag) {
		return ""
	}

	matches := regex.FindStringSubmatch(gitTag)
	if matches == nil {
		return ""
	}

	// Create map of group names to values
	groups := make(map[string]string)
	for i, name := range regex.SubexpNames() {
		if i > 0 && i < len(matches) {
			groups[name] = matches[i]
		}
	}

	// If regex contains a 'version' group, return that directly
	if version, ok := groups["version"]; ok && version != "" {
		return version
	}

	// Otherwise, construct version from major.minor.patch
	major := groups["major"]
	if major == "" {
		major = "0"
	}

	minor := groups["minor"]
	if minor == "" {
		minor = "0"
	}

	patch := groups["patch"]
	if patch == "" {
		patch = "0"
	}

	version := fmt.Sprintf("%s.%s.%s", major, minor, patch)

	// Add build metadata if present
	if build, ok := groups["build"]; ok && build != "" {
		version += "-" + build
	}

	return version
}

// cloneRepository clones a Git repository
func (s *GitImportService) cloneRepository(gitURL, targetDir string, gitRef string) error {
	// Set up environment for Git clone
	env := os.Environ()
	env = append(env, "GIT_SSH_COMMAND=ssh -o StrictHostKeyChecking=accept-new")

	args := []string{
		"git", "clone", "--single-branch",
	}

	// Add branch/tag reference if provided
	if gitRef != "" {
		args = append(args, "--branch", gitRef)
	}

	args = append(args, gitURL, targetDir)

	cmd := exec.Command("git", args...)
	cmd.Env = env

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to execute git clone: %w", err)
	}

	if cmd.ProcessState != nil {
		return fmt.Errorf("git clone failed: %s", output)
	}

	return nil
}

// getGitCommitSHA extracts the Git commit SHA from a repository
func (s *GitImportService) getGitCommitSHA(repoDir string) (string, error) {
	cmd := exec.Command("git", "rev-parse", "HEAD")
	cmd.Dir = repoDir

	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to get git commit SHA: %w", err)
	}

	return strings.TrimSpace(string(output)), nil
}

// getGitCommitDate extracts Git commit date from a repository
func (s *GitImportService) getGitCommitDate(repoDir string) (time.Time, error) {
	cmd := exec.Command("git", "show", "-s", "--format=%cI", "HEAD")
	cmd.Dir = repoDir

	output, err := cmd.CombinedOutput()
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to get git commit date: %w", err)
	}

	dateStr := strings.TrimSpace(string(output))
	return time.Parse(time.RFC3339, dateStr)
}

// extractModuleFromClone extracts module files from a cloned repository
func (s *GitImportService) extractModuleFromClone(repoDir string, moduleVersion *model.ModuleVersion) error {
	// Determine module source directory in the cloned repository
	gitPath := ""
	if moduleVersion.ModuleProvider().GitPath() != nil {
		gitPath = *moduleVersion.ModuleProvider().GitPath()
	}

	sourceDir := filepath.Join(repoDir, gitPath)

	// Ensure source directory exists
	if _, err := os.Stat(sourceDir); os.IsNotExist(err) {
		return fmt.Errorf("module source directory does not exist: %s", sourceDir)
	}

	// Create destination directory in data storage
	versionStr := moduleVersion.Version().String()
	destDir := filepath.Join(s.infraConfig.DataDirectory,
		moduleVersion.ModuleProvider().Namespace().Name(),
		moduleVersion.ModuleProvider().Module(),
		moduleVersion.ModuleProvider().Provider(),
		versionStr,
	)

	// Create extraction directory
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return fmt.Errorf("failed to create extraction directory: %w", err)
	}

	// Copy all files from source to destination
	return filepath.Walk(sourceDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip .git directory and hidden files
		if info.IsDir() && info.Name() == ".git" {
			return filepath.SkipDir
		}

		// Skip hidden files and directories (except for .terraformignore which is valid)
		if strings.HasPrefix(info.Name(), ".") && info.Name() != ".terraformignore" {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// Calculate relative path
		relPath, err := filepath.Rel(sourceDir, path)
		if err != nil {
			return fmt.Errorf("failed to calculate relative path: %w", err)
		}

		// Create destination path
		destPath := filepath.Join(destDir, relPath)

		// Create directory if it's a directory
		if info.IsDir() {
			return os.MkdirAll(destPath, info.Mode())
		}

		// Copy file if it's a file
		return copyFile(path, destPath, info.Mode())
	})
}

// copyFile copies a file from src to dest with specified permissions
func copyFile(src, dest string, mode os.FileMode) error {
	// Read source file
	data, err := os.ReadFile(src)
	if err != nil {
		return fmt.Errorf("failed to read source file: %w", err)
	}

	// Write destination file
	if err := os.WriteFile(dest, data, mode); err != nil {
		return fmt.Errorf("failed to write destination file: %w", err)
	}

	return nil
}
