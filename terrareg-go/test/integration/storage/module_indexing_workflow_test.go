package storage

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/git/service"
	storageservice "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/storage/service"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestLocalStorageService_BasicOperations tests basic storage operations
func TestLocalStorageService_BasicOperations(t *testing.T) {
	tempDir := t.TempDir()
	dataDir := filepath.Join(tempDir, "data")

	ctx := context.Background()

	// Initialize storage services
	pathConfig := storageservice.GetDefaultPathConfig(dataDir)
	pathBuilder := storageservice.NewPathBuilderService(pathConfig)

	domainStorage, err := storage.NewLocalStorageService(dataDir, pathBuilder)
	require.NoError(t, err)

	t.Run("Upload and Read File", func(t *testing.T) {
		// Create a test file in temp directory
		testContent := []byte("test file content for terrareg")
		sourceFile := filepath.Join(tempDir, "test-source.txt")
		err = os.WriteFile(sourceFile, testContent, 0644)
		require.NoError(t, err)

		// Upload file to storage
		destDir := filepath.Join(dataDir, "test-dir")
		destFilename := "test.txt"
		err = domainStorage.UploadFile(ctx, sourceFile, destDir, destFilename)
		require.NoError(t, err)

		// Verify file was uploaded
		fullPath := pathBuilder.SafeJoinPaths(destDir, destFilename)
		_, err = os.Stat(fullPath)
		assert.NoError(t, err)

		// Read file back
		readContent, err := domainStorage.ReadFile(ctx, fullPath, true)
		require.NoError(t, err)
		assert.Equal(t, testContent, readContent)
	})

	t.Run("File Exists Check", func(t *testing.T) {
		// Create a test file
		testPath := filepath.Join(dataDir, "exists-test.txt")
		testContent := []byte("exists test")
		err = os.WriteFile(testPath, testContent, 0644)
		require.NoError(t, err)

		// Check file exists
		exists, err := domainStorage.FileExists(ctx, testPath)
		require.NoError(t, err)
		assert.True(t, exists)

		// Check non-existent file
		exists, err = domainStorage.FileExists(ctx, filepath.Join(dataDir, "nonexistent.txt"))
		require.NoError(t, err)
		assert.False(t, exists)
	})

	t.Run("Directory Operations", func(t *testing.T) {
		testDir := filepath.Join(dataDir, "test-directory")

		// Create directory
		err = domainStorage.MakeDirectory(ctx, testDir)
		require.NoError(t, err)

		// Check directory exists
		exists, err := domainStorage.DirectoryExists(ctx, testDir)
		require.NoError(t, err)
		assert.True(t, exists)

		// Check non-existent directory
		exists, err = domainStorage.DirectoryExists(ctx, filepath.Join(dataDir, "nonexistent-dir"))
		require.NoError(t, err)
		assert.False(t, exists)
	})

	t.Run("Write File", func(t *testing.T) {
		testPath := filepath.Join(dataDir, "write-test.txt")
		testContent := "written content"

		err = domainStorage.WriteFile(ctx, testPath, testContent, false)
		require.NoError(t, err)

		// Verify content
		readContent, err := domainStorage.ReadFile(ctx, testPath, false)
		require.NoError(t, err)
		assert.Equal(t, []byte(testContent), readContent)
	})

	t.Run("Delete File", func(t *testing.T) {
		testFile := filepath.Join(dataDir, "delete-test.txt")
		testContent := []byte("to be deleted")
		err = os.WriteFile(testFile, testContent, 0644)
		require.NoError(t, err)

		// Verify file exists
		exists, err := domainStorage.FileExists(ctx, testFile)
		require.NoError(t, err)
		assert.True(t, exists)

		// Delete file
		err = domainStorage.DeleteFile(ctx, testFile)
		require.NoError(t, err)

		// Verify file is gone
		exists, err = domainStorage.FileExists(ctx, testFile)
		require.NoError(t, err)
		assert.False(t, exists)
	})

	t.Run("Delete Directory", func(t *testing.T) {
		testDir := filepath.Join(dataDir, "delete-dir-test")
		testFile := filepath.Join(testDir, "test.txt")
		err = os.MkdirAll(testDir, 0755)
		require.NoError(t, err)
		err = os.WriteFile(testFile, []byte("test"), 0644)
		require.NoError(t, err)

		// Delete directory
		err = domainStorage.DeleteDirectory(ctx, testDir)
		require.NoError(t, err)

		// Verify directory is gone
		exists, err := domainStorage.DirectoryExists(ctx, testDir)
		require.NoError(t, err)
		assert.False(t, exists)
	})
}

// TestPathBuilderService tests path building for module storage
func TestPathBuilderService(t *testing.T) {
	tempDir := t.TempDir()
	dataDir := filepath.Join(tempDir, "data")

	pathConfig := storageservice.GetDefaultPathConfig(dataDir)
	pathBuilder := storageservice.NewPathBuilderService(pathConfig)

	t.Run("Build Namespace Path", func(t *testing.T) {
		namespace := "test-namespace"
		path := pathBuilder.BuildNamespacePath(namespace)

		expectedPath := filepath.Join(dataDir, "modules", namespace)
		assert.Equal(t, expectedPath, path)
	})

	t.Run("Build Module Path", func(t *testing.T) {
		namespace := "hashicorp"
		module := "consul"
		path := pathBuilder.BuildModulePath(namespace, module)

		expectedPath := filepath.Join(dataDir, "modules", namespace, module)
		assert.Equal(t, expectedPath, path)
	})

	t.Run("Build Provider Path", func(t *testing.T) {
		namespace := "terraform-aws-modules"
		module := "vpc"
		provider := "aws"
		path := pathBuilder.BuildProviderPath(namespace, module, provider)

		expectedPath := filepath.Join(dataDir, "modules", namespace, module, provider)
		assert.Equal(t, expectedPath, path)
	})

	t.Run("Build Version Path", func(t *testing.T) {
		namespace := "my-company"
		module := "database-module"
		provider := "postgresql"
		version := "1.2.3"
		path := pathBuilder.BuildVersionPath(namespace, module, provider, version)

		expectedPath := filepath.Join(dataDir, "modules", namespace, module, provider, version)
		assert.Equal(t, expectedPath, path)
	})

	t.Run("Build Archive Path", func(t *testing.T) {
		namespace := "test-ns"
		module := "test-module"
		provider := "aws"
		version := "1.0.0"
		archiveName := "source.tar.gz"

		path := pathBuilder.BuildArchivePath(namespace, module, provider, version, archiveName)

		expectedPath := filepath.Join(dataDir, "modules", namespace, module, provider, version, archiveName)
		assert.Equal(t, expectedPath, path)
	})

	t.Run("Safe Join Paths Prevents Traversal", func(t *testing.T) {
		basePath := filepath.Join(dataDir, "safe")
		// Attempt to use path traversal
		result := pathBuilder.SafeJoinPaths(basePath, "../unsafe")

		// The path should not contain ".." after joining
		assert.NotContains(t, result, "..")
		assert.Contains(t, result, "safe")
	})

	t.Run("Build Module Archive Paths", func(t *testing.T) {
		namespace := "test-ns"
		module := "test-module"
		provider := "aws"
		version := "1.0.0"

		paths := pathBuilder.BuildModuleArchivePaths(namespace, module, provider, version)

		require.Len(t, paths, 2)

		// Should have tar.gz and zip archives
		hasTarGz := false
		hasZip := false
		for _, path := range paths {
			if filepath.Base(path) == "source.tar.gz" {
				hasTarGz = true
			}
			if filepath.Base(path) == "source.zip" {
				hasZip = true
			}
		}
		assert.True(t, hasTarGz, "Should have tar.gz archive path")
		assert.True(t, hasZip, "Should have zip archive path")
	})
}

// TestGitService_BasicOperations tests basic git operations
func TestGitService_BasicOperations(t *testing.T) {
	gitService := service.NewGitService()
	ctx := context.Background()

	t.Run("Create And Validate Repository", func(t *testing.T) {
		tempDir := t.TempDir()
		repoDir := filepath.Join(tempDir, "test-repo")

		// Create a test repository
		err := createTestTerraformRepository(repoDir)
		require.NoError(t, err)

		// Check if it's a valid git repository
		isValid := gitService.IsGitRepository(ctx, repoDir)
		assert.True(t, isValid)

		// Get commit SHA
		commitSHA, err := gitService.GetCommitSHA(ctx, repoDir)
		require.NoError(t, err)
		assert.Len(t, commitSHA, 40)
	})

	t.Run("IsGitRepository Returns False For NonGitDir", func(t *testing.T) {
		tempDir := t.TempDir()

		isValid := gitService.IsGitRepository(ctx, tempDir)
		assert.False(t, isValid)
	})

	t.Run("ValidateRepository With InvalidURL", func(t *testing.T) {
		err := gitService.ValidateRepository(ctx, "not-a-valid-url")
		assert.Error(t, err)
	})

	t.Run("ParseRepositoryURL", func(t *testing.T) {
		testCases := []struct {
			name     string
			url      string
			expected *service.RepositoryInfo
		}{
			{
				name: "HTTPS URL",
				url:  "https://github.com/hashicorp/consul",
				expected: &service.RepositoryInfo{
					URL:   "https://github.com/hashicorp/consul",
					Owner: "hashicorp",
					Name:  "consul",
					IsSSH: false,
				},
			},
			{
				name: "HTTPS URL with .git suffix",
				url:  "https://github.com/hashicorp/consul.git",
				expected: &service.RepositoryInfo{
					URL:   "https://github.com/hashicorp/consul",
					Owner: "hashicorp",
					Name:  "consul",
					IsSSH: false,
				},
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				info, err := gitService.ParseRepositoryURL(tc.url)
				require.NoError(t, err)
				assert.Equal(t, tc.expected.Owner, info.Owner)
				assert.Equal(t, tc.expected.Name, info.Name)
				assert.Equal(t, tc.expected.IsSSH, info.IsSSH)
			})
		}
	})
}

// TestModuleStorageWorkflow tests the complete workflow for storing module archives
func TestModuleStorageWorkflow(t *testing.T) {
	tempDir := t.TempDir()
	dataDir := filepath.Join(tempDir, "data")

	ctx := context.Background()

	// Initialize storage services
	pathConfig := storageservice.GetDefaultPathConfig(dataDir)
	pathBuilder := storageservice.NewPathBuilderService(pathConfig)

	domainStorage, err := storage.NewLocalStorageService(dataDir, pathBuilder)
	require.NoError(t, err)

	t.Run("Store Module Archive", func(t *testing.T) {
		namespace := "test-namespace"
		module := "test-module"
		provider := "aws"
		version := "1.0.0"

		// Create mock archive content
		tarGzContent := []byte("mock tar.gz archive content")
		zipContent := []byte("mock zip archive content")

		// Build version path
		versionPath := pathBuilder.BuildVersionPath(namespace, module, provider, version)

		// Create archive paths
		tarGzPath := pathBuilder.BuildArchivePath(namespace, module, provider, version, "source.tar.gz")
		zipPath := pathBuilder.BuildArchivePath(namespace, module, provider, version, "source.zip")

		// Write archives
		err = domainStorage.WriteFile(ctx, tarGzPath, tarGzContent, true)
		require.NoError(t, err)

		err = domainStorage.WriteFile(ctx, zipPath, zipContent, true)
		require.NoError(t, err)

		// Verify archives exist
		exists, err := domainStorage.FileExists(ctx, tarGzPath)
		require.NoError(t, err)
		assert.True(t, exists)

		exists, err = domainStorage.FileExists(ctx, zipPath)
		require.NoError(t, err)
		assert.True(t, exists)

		// Verify version directory exists
		exists, err = domainStorage.DirectoryExists(ctx, versionPath)
		require.NoError(t, err)
		assert.True(t, exists)

		// Read back and verify content
		readTarGz, err := domainStorage.ReadFile(ctx, tarGzPath, true)
		require.NoError(t, err)
		assert.Equal(t, tarGzContent, readTarGz)

		readZip, err := domainStorage.ReadFile(ctx, zipPath, true)
		require.NoError(t, err)
		assert.Equal(t, zipContent, readZip)
	})

	t.Run("Multiple Module Versions", func(t *testing.T) {
		namespace := "multi-version"
		module := "test-module"
		provider := "aws"
		versions := []string{"1.0.0", "1.1.0", "2.0.0"}

		for _, version := range versions {
			// Create version directory
			versionPath := pathBuilder.BuildVersionPath(namespace, module, provider, version)
			err = domainStorage.MakeDirectory(ctx, versionPath)
			require.NoError(t, err)

			// Create archive
			archivePath := pathBuilder.BuildArchivePath(namespace, module, provider, version, "source.tar.gz")
			content := []byte(fmt.Sprintf("archive content for %s", version))
			err = domainStorage.WriteFile(ctx, archivePath, content, true)
			require.NoError(t, err)
		}

		// Verify all versions exist
		for _, version := range versions {
			archivePath := pathBuilder.BuildArchivePath(namespace, module, provider, version, "source.tar.gz")
			exists, err := domainStorage.FileExists(ctx, archivePath)
			require.NoError(t, err)
			assert.True(t, exists, fmt.Sprintf("Version %s archive should exist", version))
		}
	})

	t.Run("Complex Namespaces and Modules", func(t *testing.T) {
		testCases := []struct {
			namespace string
			module    string
			provider  string
			version   string
		}{
			{
				namespace: "terraform-aws-modules",
				module:    "vpc",
				provider:  "aws",
				version:   "3.0.0",
			},
			{
				namespace: "my-company-name",
				module:    "database-module-with-dashes",
				provider:  "postgresql",
				version:   "1.2.3",
			},
		}

		for _, tc := range testCases {
			t.Run(tc.namespace+"/"+tc.module, func(t *testing.T) {
				// Build paths
				versionPath := pathBuilder.BuildVersionPath(tc.namespace, tc.module, tc.provider, tc.version)
				archivePath := pathBuilder.BuildArchivePath(tc.namespace, tc.module, tc.provider, tc.version, "source.tar.gz")

				// Create version directory
				err = domainStorage.MakeDirectory(ctx, versionPath)
				require.NoError(t, err)

				// Create archive
				content := []byte("test archive")
				err = domainStorage.WriteFile(ctx, archivePath, content, true)
				require.NoError(t, err)

				// Verify
				exists, err := domainStorage.FileExists(ctx, archivePath)
				require.NoError(t, err)
				assert.True(t, exists)
			})
		}
	})
}

// Helper functions

func createTestTerraformRepository(repoDir string) error {
	// Create repository directory
	err := os.MkdirAll(repoDir, 0755)
	if err != nil {
		return err
	}

	// Initialize git repository
	commands := [][]string{
		{"git", "init"},
		{"git", "config", "user.email", "test@example.com"},
		{"git", "config", "user.name", "Test User"},
		{"git", "config", "commit.gpgsign", "false"},
	}

	for _, cmd := range commands {
		if err := runCommand(repoDir, cmd); err != nil {
			return err
		}
	}

	// Create main.tf
	mainTfContent := `
provider "aws" {
  region = "us-west-2"
}

resource "aws_instance" "example" {
  ami           = "ami-12345678"
  instance_type = "t2.micro"
}

output "instance_id" {
  value = aws_instance.example.id
}
`
	err = os.WriteFile(filepath.Join(repoDir, "main.tf"), []byte(mainTfContent), 0644)
	if err != nil {
		return err
	}

	// Create README.md
	readmeContent := "# Test Terraform Module\n\n" +
		"This is a test module for terrareg integration testing.\n\n" +
		"## Usage\n\n" +
		"```hcl\n" +
		"module \"test\" {\n" +
		"  source = \"./path/to/module\"\n\n" +
		"  providers = {\n" +
		"    aws = aws\n" +
		"  }\n" +
		"}\n" +
		"```\n"
	err = os.WriteFile(filepath.Join(repoDir, "README.md"), []byte(readmeContent), 0644)
	if err != nil {
		return err
	}

	// Add files and commit
	commands = [][]string{
		{"git", "add", "."},
		{"git", "commit", "-m", "Initial commit"},
		{"git", "tag", "v1.0.0"},
	}

	for _, cmd := range commands {
		if err := runCommand(repoDir, cmd); err != nil {
			return err
		}
	}

	return nil
}

func runCommand(dir string, cmd []string) error {
	command := exec.Command(cmd[0], cmd[1:]...)
	command.Dir = dir
	output, err := command.CombinedOutput()
	if err != nil {
		return fmt.Errorf("command %v failed: %v\nOutput: %s", cmd, err, string(output))
	}
	return nil
}
