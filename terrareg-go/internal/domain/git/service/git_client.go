package service

import "context"

// GitClient defines the interface for interacting with Git.
type GitClient interface {
	// Clone clones a git repository from a given URL into a destination path.
	Clone(ctx context.Context, repoURL, destinationPath string) error
	// Checkout switches to a specific tag or branch in a repository.
	Checkout(ctx context.Context, repositoryPath, tag string) error
}
