package service

import (
	"context"
	"time"
)

// GitCredentials represents Git authentication credentials
type GitCredentials struct {
	Username string
	Password string
}

// CloneOptions represents options for cloning a repository
type CloneOptions struct {
	Timeout    time.Duration
	Credentials *GitCredentials
}

// GitClient defines the interface for interacting with Git.
type GitClient interface {
	// Clone clones a git repository from a given URL into a destination path.
	Clone(ctx context.Context, repoURL, destinationPath string) error
	// CloneWithOptions clones a git repository with additional options.
	CloneWithOptions(ctx context.Context, repoURL, destinationPath string, options CloneOptions) error
	// Checkout switches to a specific tag or branch in a repository.
	Checkout(ctx context.Context, repositoryPath, tag string) error
}
