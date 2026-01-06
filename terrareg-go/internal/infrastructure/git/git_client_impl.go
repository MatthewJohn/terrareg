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
	// Note: Use --depth 1 only when we don't need to checkout specific tags
	args := []string{"clone", "--depth", "1", repoURL, destinationPath}
	if options.NeedTags {
		// For tag-based operations, we need full history and all tags
		// Remove --depth 1 to fetch all tags and history
		args = []string{"clone", repoURL, destinationPath}
	}

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
	// First, validate that the tag exists to provide better error messages
	if !g.tagExists(ctx, repositoryPath, tag) {
		// List available tags to help with debugging
		availableTags := g.listTags(ctx, repositoryPath)
		return fmt.Errorf("git tag '%s' not found in repository. Available tags: %s", tag, availableTags)
	}

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

// tagExists checks if a tag exists in the repository
func (g *GitClientImpl) tagExists(ctx context.Context, repositoryPath, tag string) bool {
	cmd := exec.CommandContext(ctx, "git", "-C", repositoryPath, "tag", "-l", tag)
	output, err := cmd.Output()
	if err != nil {
		return false
	}
	return strings.TrimSpace(string(output)) == tag
}

// listTags returns a comma-separated list of available tags (limited to first 10 for readability)
func (g *GitClientImpl) listTags(ctx context.Context, repositoryPath string) string {
	cmd := exec.CommandContext(ctx, "git", "-C", repositoryPath, "tag", "--sort=-version:refname")
	output, err := cmd.Output()
	if err != nil {
		return "unable to list tags"
	}

	tags := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(tags) == 0 {
		return "no tags found"
	}

	// Limit to first 10 tags for readability
	if len(tags) > 10 {
		tags = tags[:10]
		tags = append(tags, "...")
	}

	return strings.Join(tags, ", ")
}

// GetCommitSHA returns the current git commit SHA for the repository at the given path.
func (g *GitClientImpl) GetCommitSHA(ctx context.Context, repositoryPath string) (string, error) {
	cmd := exec.CommandContext(ctx, "git", "-C", repositoryPath, "rev-parse", "HEAD")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get commit SHA: %w", err)
	}
	return strings.TrimSpace(string(output)), nil
}
