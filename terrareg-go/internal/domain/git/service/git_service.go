package service

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// GitService handles Git repository operations for module indexing
// This supports the critical Git clone → temp dir → process workflow
type GitService interface {
	// CloneRepository clones a repository to a target directory
	// This supports both SSH and HTTPS repositories with authentication
	CloneRepository(ctx context.Context, repoURL, branch, targetDir string, options *CloneOptions) error

	// GetCommitSHA extracts the commit SHA from a repository directory
	// This is used for tracking and provenance
	GetCommitSHA(ctx context.Context, dir string) (string, error)

	// ValidateRepository checks if a repository URL is valid and accessible
	// This validates before attempting expensive clone operations
	ValidateRepository(ctx context.Context, repoURL string) error

	// IsGitRepository checks if a directory is a valid Git repository
	IsGitRepository(ctx context.Context, dir string) bool
}


// GitServiceImpl implements GitService
type GitServiceImpl struct{}

// NewGitService creates a new Git service
func NewGitService() *GitServiceImpl {
	return &GitServiceImpl{}
}

// CloneRepository clones a Git repository to the specified directory
// This implements the critical Git clone step in the module indexing workflow
func (g *GitServiceImpl) CloneRepository(ctx context.Context, repoURL, branch, targetDir string, options *CloneOptions) error {
	if options == nil {
		options = &CloneOptions{}
	}

	// Create target directory if it doesn't exist
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return fmt.Errorf("failed to create target directory: %w", err)
	}

	// Build git clone command
	args := []string{"clone", repoURL, targetDir}

	// Add branch option if specified
	if branch != "" {
		args = append(args, "--branch", branch)
	}

	
	// Create command context with timeout if specified
	cmdCtx := ctx
	if options.Timeout > 0 {
		var cancel context.CancelFunc
		cmdCtx, cancel = context.WithTimeout(ctx, options.Timeout)
		defer cancel()
	}

	// Create git command
	cmd := exec.CommandContext(cmdCtx, "git", args...)

	// Set up environment for authentication
	if options.Credentials != nil {
		env := os.Environ()

		// HTTPS authentication
		if options.Credentials.Username != "" && options.Credentials.Password != "" {
			env = append(env, fmt.Sprintf("GIT_USERNAME=%s", options.Credentials.Username))
			env = append(env, fmt.Sprintf("GIT_PASSWORD=%s", options.Credentials.Password))
		}

		
		cmd.Env = env
	}

	// Execute clone command
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git clone failed: %v\nOutput: %s", err, string(output))
	}

	// Verify the clone was successful
	if !g.IsGitRepository(ctx, targetDir) {
		return fmt.Errorf("git clone completed but target directory is not a valid repository")
	}

	return nil
}

// GetCommitSHA extracts the current commit SHA from a Git repository
// This is used for provenance and tracking in the module indexing workflow
func (g *GitServiceImpl) GetCommitSHA(ctx context.Context, dir string) (string, error) {
	// Check if directory is a Git repository
	if !g.IsGitRepository(ctx, dir) {
		return "", fmt.Errorf("directory is not a Git repository: %s", dir)
	}

	// Get current commit SHA
	cmd := exec.CommandContext(ctx, "git", "-C", dir, "rev-parse", "HEAD")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get commit SHA: %v", err)
	}

	commitSHA := strings.TrimSpace(string(output))
	if len(commitSHA) != 40 {
		return "", fmt.Errorf("invalid commit SHA format: %s", commitSHA)
	}

	return commitSHA, nil
}

// ValidateRepository checks if a repository URL is valid and accessible
// This prevents expensive clone operations on invalid repositories
func (g *GitServiceImpl) ValidateRepository(ctx context.Context, repoURL string) error {
	// Basic URL validation
	if repoURL == "" {
		return fmt.Errorf("repository URL cannot be empty")
	}

	// Check URL format
	if !strings.HasPrefix(repoURL, "http://") &&
		!strings.HasPrefix(repoURL, "https://") &&
		!strings.HasPrefix(repoURL, "git@") &&
		!strings.HasPrefix(repoURL, "git://") {
		return fmt.Errorf("invalid repository URL format: %s", repoURL)
	}

	// Try to validate repository accessibility (lightweight check)
	cmd := exec.CommandContext(ctx, "git", "ls-remote", "--exit-code", repoURL, "HEAD")
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("repository is not accessible: %v", err)
	}

	return nil
}

// IsGitRepository checks if a directory contains a Git repository
// This validates that the clone operation was successful
func (g *GitServiceImpl) IsGitRepository(ctx context.Context, dir string) bool {
	gitDir := filepath.Join(dir, ".git")
	info, err := os.Stat(gitDir)
	if err != nil {
		return false
	}

	return info.IsDir()
}

// ParseRepositoryURL extracts repository information from a URL
// This helps in extracting repository metadata for the module indexing workflow
func (g *GitServiceImpl) ParseRepositoryURL(repoURL string) (*RepositoryInfo, error) {
	if repoURL == "" {
		return nil, fmt.Errorf("repository URL cannot be empty")
	}

	// Remove .git suffix if present
	repoURL = strings.TrimSuffix(repoURL, ".git")

	// Parse different URL formats
	var owner, name string
	var isSSH bool

	if strings.HasPrefix(repoURL, "git@") {
		// SSH format: git@github.com:owner/repo.git
		parts := strings.Split(repoURL, ":")
		if len(parts) >= 2 {
			hostRepo := parts[1]
			hostRepoParts := strings.Split(hostRepo, "/")
			if len(hostRepoParts) >= 2 {
				owner = hostRepoParts[0]
				name = strings.TrimSuffix(hostRepoParts[1], ".git")
			}
		}
		isSSH = true
	} else {
		// HTTPS format: https://github.com/owner/repo
		parts := strings.Split(repoURL, "/")
		if len(parts) >= 5 {
			owner = parts[3]
			name = strings.TrimSuffix(parts[4], ".git")
		}
	}

	if owner == "" || name == "" {
		return nil, fmt.Errorf("could not parse repository owner and name from URL: %s", repoURL)
	}

	return &RepositoryInfo{
		URL:    repoURL,
		Owner:  owner,
		Name:   name,
		IsSSH:  isSSH,
	}, nil
}

// RepositoryInfo contains parsed repository information
type RepositoryInfo struct {
	URL   string
	Owner string
	Name  string
	IsSSH bool
}

// GetRepositoryDisplayName returns a human-readable repository name
func (r *RepositoryInfo) GetRepositoryDisplayName() string {
	if r.Owner != "" && r.Name != "" {
		return fmt.Sprintf("%s/%s", r.Owner, r.Name)
	}
	return r.URL
}