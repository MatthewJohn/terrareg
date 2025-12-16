package service

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGitService_CloneRepository(t *testing.T) {
	// Setup - create a test repository first
	tempDir := t.TempDir()
	repoDir := filepath.Join(tempDir, "test-repo")
	targetDir := filepath.Join(tempDir, "cloned-repo")

	// Initialize a test git repository
	err := os.MkdirAll(repoDir, 0755)
	require.NoError(t, err)

	// Initialize git repo
	ctx := context.Background()
	gitService := NewGitService()

	// Create initial commit in test repository
	initCommands := [][]string{
		{"git", "init"},
		{"git", "config", "user.email", "test@example.com"},
		{"git", "config", "user.name", "Test User"},
		{"git", "commit", "--allow-empty", "-m", "Initial commit"},
	}

	for _, cmd := range initCommands {
		err := runCommand(repoDir, cmd)
		require.NoError(t, err)
	}

	// Create a tag
	err = runCommand(repoDir, []string{"git", "tag", "v1.0.0"})
	require.NoError(t, err)

	// Test successful clone
	cloneOptions := &gitService.CloneOptions{
		Timeout: 30 * time.Second,
	}

	err = gitService.CloneRepository(ctx, repoDir, "v1.0.0", targetDir, cloneOptions)
	assert.NoError(t, err)

	// Verify cloned repository exists
	info, err := os.Stat(targetDir)
	assert.NoError(t, err)
	assert.True(t, info.IsDir())

	// Verify .git directory exists
	gitDir := filepath.Join(targetDir, ".git")
	info, err = os.Stat(gitDir)
	assert.NoError(t, err)
	assert.True(t, info.IsDir())

	// Verify it's a valid git repository
	assert.True(t, gitService.IsGitRepository(ctx, targetDir))
}

func TestGitService_CloneRepository_WithTimeout(t *testing.T) {
	tempDir := t.TempDir()
	repoDir := filepath.Join(tempDir, "test-repo")
	targetDir := filepath.Join(tempDir, "cloned-repo-timeout")

	// Initialize a test repository
	err := os.MkdirAll(repoDir, 0755)
	require.NoError(t, err)

	ctx := context.Background()
	gitService := NewGitService()

	// Initialize git repo
	err = runCommand(repoDir, []string{"git", "init"})
	require.NoError(t, err)

	err = runCommand(repoDir, []string{"git", "config", "user.email", "test@example.com"})
	require.NoError(t, err)

	err = runCommand(repoDir, []string{"git", "config", "user.name", "Test User"})
	require.NoError(t, err)

	err = runCommand(repoDir, []string{"git", "commit", "--allow-empty", "-m", "Initial commit"})
	require.NoError(t, err)

	// Test clone with short timeout
	cloneOptions := &gitService.CloneOptions{
		Timeout: 100 * time.Millisecond, // Very short timeout
	}

	err = gitService.CloneRepository(ctx, repoDir, "", targetDir, cloneOptions)
	// This might succeed quickly or fail with timeout, both are acceptable
	// The important thing is that it doesn't hang
	assert.True(t, err == nil || err.Error() == "context deadline exceeded")
}

func TestGitService_CloneRepository_WithAuthentication(t *testing.T) {
	// This test would require actual authentication setup
	// For now, we'll test the error case with invalid credentials
	tempDir := t.TempDir()
	targetDir := filepath.Join(tempDir, "cloned-repo-auth")

	ctx := context.Background()
	gitService := NewGitService()

	// Test with invalid repository URL (should fail gracefully)
	cloneOptions := &gitService.CloneOptions{
		Timeout: 5 * time.Second,
		Credentials: &gitService.GitCredentials{
			Username: "invalid",
			Password: "invalid",
		},
	}

	err := gitService.CloneRepository(ctx, "https://github.com/nonexistent/repo.git", "", targetDir, cloneOptions)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "git clone failed")
}

func TestGitService_GetCommitSHA(t *testing.T) {
	tempDir := t.TempDir()
	repoDir := filepath.Join(tempDir, "test-repo-sha")

	ctx := context.Background()
	gitService := NewGitService()

	// Initialize repository
	err := os.MkdirAll(repoDir, 0755)
	require.NoError(t, err)

	err = runCommand(repoDir, []string{"git", "init"})
	require.NoError(t, err)

	err = runCommand(repoDir, []string{"git", "config", "user.email", "test@example.com"})
	require.NoError(t, err)

	err = runCommand(repoDir, []string{"git", "config", "user.name", "Test User"})
	require.NoError(t, err)

	// Create a test file and commit
	testFile := filepath.Join(repoDir, "test.txt")
	err = os.WriteFile(testFile, []byte("test content"), 0644)
	require.NoError(t, err)

	err = runCommand(repoDir, []string{"git", "add", "test.txt"})
	require.NoError(t, err)

	err = runCommand(repoDir, []string{"git", "commit", "-m", "Add test file"})
	require.NoError(t, err)

	// Test get commit SHA
	commitSHA, err := gitService.GetCommitSHA(ctx, repoDir)
	assert.NoError(t, err)
	assert.NotEmpty(t, commitSHA)
	assert.Len(t, commitSHA, 40) // Git SHA is 40 characters
	assert.Regexp(t, `^[a-f0-9]{40}$`, commitSHA)
}

func TestGitService_GetCommitSHA_NonGitRepo(t *testing.T) {
	tempDir := t.TempDir()
	nonGitDir := filepath.Join(tempDir, "not-a-git-repo")

	err := os.MkdirAll(nonGitDir, 0755)
	require.NoError(t, err)

	ctx := context.Background()
	gitService := NewGitService()

	// Test get commit SHA from non-git repository
	commitSHA, err := gitService.GetCommitSHA(ctx, nonGitDir)
	// This should return empty string and not error
	assert.Empty(t, commitSHA)
	assert.NoError(t, err)
}

func TestGitService_IsGitRepository(t *testing.T) {
	tempDir := t.TempDir()
	repoDir := filepath.Join(tempDir, "test-repo-valid")
	nonGitDir := filepath.Join(tempDir, "not-a-git-repo")

	ctx := context.Background()
	gitService := NewGitService()

	// Create non-git directory
	err := os.MkdirAll(nonGitDir, 0755)
	require.NoError(t, err)

	// Test non-git directory
	assert.False(t, gitService.IsGitRepository(ctx, nonGitDir))

	// Initialize git repository
	err = os.MkdirAll(repoDir, 0755)
	require.NoError(t, err)

	err = runCommand(repoDir, []string{"git", "init"})
	require.NoError(t, err)

	// Test git directory
	assert.True(t, gitService.IsGitRepository(ctx, repoDir))
}

func TestGitService_ValidateRepository(t *testing.T) {
	ctx := context.Background()
	gitService := NewGitService()

	// Test with invalid URL (should fail)
	err := gitService.ValidateRepository(ctx, "not-a-url")
	assert.Error(t, err)

	// Test with non-existent repository
	err = gitService.ValidateRepository(ctx, "https://github.com/nonexistent/repo.git")
	// This may fail due to network issues, but should not hang
	// The important thing is that it returns an error
	assert.Error(t, err)

	// Note: Testing with a real repository would require network access
	// and might be flaky, so we keep tests to local/error cases
}

func TestGitService_CloneRepository_BranchOption(t *testing.T) {
	tempDir := t.TempDir()
	repoDir := filepath.Join(tempDir, "test-repo-branches")
	targetDir := filepath.Join(tempDir, "cloned-repo-branch")

	ctx := context.Background()
	gitService := NewGitService()

	// Initialize repository with multiple branches
	err := os.MkdirAll(repoDir, 0755)
	require.NoError(t, err)

	err = runCommand(repoDir, []string{"git", "init"})
	require.NoError(t, err)

	err = runCommand(repoDir, []string{"git", "config", "user.email", "test@example.com"})
	require.NoError(t, err)

	err = runCommand(repoDir, []string{"git", "config", "user.name", "Test User"})
	require.NoError(t, err)

	// Initial commit
	err = runCommand(repoDir, []string{"git", "commit", "--allow-empty", "-m", "Initial commit"})
	require.NoError(t, err)

	// Create and switch to new branch
	err = runCommand(repoDir, []string{"git", "checkout", "-b", "feature-branch"})
	require.NoError(t, err)

	// Add commit on feature branch
	testFile := filepath.Join(repoDir, "feature.txt")
	err = os.WriteFile(testFile, []byte("feature content"), 0644)
	require.NoError(t, err)

	err = runCommand(repoDir, []string{"git", "add", "feature.txt"})
	require.NoError(t, err)

	err = runCommand(repoDir, []string{"git", "commit", "-m", "Add feature"})
	require.NoError(t, err)

	// Clone specific branch
	cloneOptions := &gitService.CloneOptions{
		Timeout: 30 * time.Second,
	}

	err = gitService.CloneRepository(ctx, repoDir, "feature-branch", targetDir, cloneOptions)
	assert.NoError(t, err)

	// Verify correct branch was cloned
	featureFile := filepath.Join(targetDir, "feature.txt")
	_, err = os.Stat(featureFile)
	assert.NoError(t, err)
}

func TestGitService_ErrorHandling(t *testing.T) {
	ctx := context.Background()
	gitService := NewGitService()

	// Test clone to invalid target directory
	tempDir := t.TempDir()
	repoDir := filepath.Join(tempDir, "test-repo")

	// Initialize a simple repo
	err := os.MkdirAll(repoDir, 0755)
	require.NoError(t, err)

	err = runCommand(repoDir, []string{"git", "init"})
	require.NoError(t, err)

	err = runCommand(repoDir, []string{"git", "config", "user.email", "test@example.com"})
	require.NoError(t, err)

	err = runCommand(repoDir, []string{"git", "config", "user.name", "Test User"})
	require.NoError(t, err)

	err = runCommand(repoDir, []string{"git", "commit", "--allow-empty", "-m", "Initial commit"})
	require.NoError(t, err)

	// Try to clone to non-existent parent directory
	invalidTarget := filepath.Join(tempDir, "nonexistent", "target")
	cloneOptions := &gitService.CloneOptions{
		Timeout: 5 * time.Second,
	}

	err = gitService.CloneRepository(ctx, repoDir, "", invalidTarget, cloneOptions)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "git clone failed")

	// Test operations on non-existent directory
	commitSHA, err := gitService.GetCommitSHA(ctx, "/nonexistent/directory")
	assert.Empty(t, commitSHA)
	assert.NoError(t, err) // Should not error, just return empty

	isRepo := gitService.IsGitRepository(ctx, "/nonexistent/directory")
	assert.False(t, isRepo)
}

// Helper function to run commands
func runCommand(dir string, cmd []string) error {
	command := exec.CommandContext(context.Background(), cmd[0], cmd[1:]...)
	command.Dir = dir
	return command.Run()
}