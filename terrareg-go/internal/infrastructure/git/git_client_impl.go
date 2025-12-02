package git

import (
	"context"
	"os/exec"
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

// Checkout switches to a specific tag or branch in a repository.
func (g *GitClientImpl) Checkout(ctx context.Context, repositoryPath, tag string) error {
	cmd := exec.CommandContext(ctx, "git", "-C", repositoryPath, "checkout", tag)
	return cmd.Run()
}
