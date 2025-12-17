package git

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/git/service"
)

// GitClientImpl implements the GitClient interface using the git command.
type GitClientImpl struct{}

// NewGitClientImpl creates a new GitClientImpl.
func NewGitClientImpl() *GitClientImpl {
	return &GitClientImpl{}
}

// Clone clones a git repository from a given URL into a destination path.
func (g *GitClientImpl) Clone(ctx context.Context, repoURL, destinationPath string) error {
	cmd := exec.CommandContext(ctx, "git", "clone", repoURL, destinationPath)
	return cmd.Run()
}

// CloneWithOptions clones a git repository with additional options including timeout and credentials.
func (g *GitClientImpl) CloneWithOptions(ctx context.Context, repoURL, destinationPath string, options service.CloneOptions) error {
	// Create a context with timeout if specified
	var gitCtx context.Context
	var cancel context.CancelFunc

	if options.Timeout > 0 {
		gitCtx, cancel = context.WithTimeout(ctx, options.Timeout)
		defer cancel()
	} else {
		gitCtx = ctx
	}

	// Build git clone command with optional depth 1 for faster cloning
	args := []string{"clone", "--depth", "1", repoURL, destinationPath}

	cmd := exec.CommandContext(gitCtx, "git", args...)

	// Set up credentials if provided
	if options.Credentials != nil && options.Credentials.Username != "" && options.Credentials.Password != "" {
		// Create a custom environment with git credentials
		env := os.Environ()
		env = append(env, fmt.Sprintf("GIT_USERNAME=%s", options.Credentials.Username))
		env = append(env, fmt.Sprintf("GIT_PASSWORD=%s", options.Credentials.Password))

		// For HTTPS URLs, modify the URL to include credentials
		if strings.HasPrefix(repoURL, "https://") {
			authURL := strings.Replace(repoURL, "https://", fmt.Sprintf("https://%s:%s@",
				options.Credentials.Username, options.Credentials.Password), 1)
			args[2] = authURL
			cmd = exec.CommandContext(gitCtx, "git", args...)
		}

		cmd.Env = env
	}

	return cmd.Run()
}

// Checkout switches to a specific tag or branch in a repository.
func (g *GitClientImpl) Checkout(ctx context.Context, repositoryPath, tag string) error {
	cmd := exec.CommandContext(ctx, "git", "-C", repositoryPath, "checkout", tag)

	// Capture stdout and stderr for debugging
	stdout, err := cmd.CombinedOutput()
	if err != nil {
		// Include both stdout and stderr in the error for better debugging
		if len(stdout) > 0 {
			return fmt.Errorf("git checkout failed: %w\nOutput:\n%s", err, string(stdout))
		}
		return fmt.Errorf("git checkout failed: %w", err)
	}

	return nil
}
