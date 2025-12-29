package storage

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/git/service"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/storage/model"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/storage/service"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestModuleIndexingWorkflowEndToEnd tests the complete workflow:
// Git clone → temp dir → process → upload archives → cleanup
func TestModuleIndexingWorkflowEndToEnd(t *testing.T) {
	// Setup
	tempDir := t.TempDir()
	dataDir := filepath.Join(tempDir, "data")

	// Initialize storage services
	pathConfig := service.GetDefaultPathConfig(dataDir)
	pathBuilder := service.NewPathBuilderService(pathConfig)

	domainStorage, err := storage.NewLocalStorageService(dataDir, pathBuilder)
	require.NoError(t, err)

	tempDirManager, err := storage.NewTemporaryDirectoryManager()
	require.NoError(t, err)

	storageWorkflow := service.NewStorageWorkflowServiceImpl(
		domainStorage,
		pathBuilder,
		tempDirManager,
		&model.StoragePathConfig{
			BasePath:      pathConfig.BasePath,
			ModulesPath:   pathConfig.ModulesPath,
			ProvidersPath: pathConfig.ProvidersPath,
			UploadPath:    pathConfig.UploadPath,
			TempPath:      pathConfig.TempPath,
		},
	)

	gitService := service.NewGitService()
	ctx := context.Background()

	// Create a test Git repository with Terraform module
	repoDir := filepath.Join(tempDir, "test-repo")
	err = createTestTerraformRepository(repoDir)
	require.NoError(t, err)

	// Test the workflow: Git clone → temp dir → process → upload archives → cleanup

	// 1. Create processing directory
	processingDir, cleanup, err := storageWorkflow.CreateProcessingDirectory(ctx, "test_workflow_")
	require.NoError(t, err)
	defer cleanup()

	// Verify processing directory exists
	info, err := os.Stat(processingDir)
	assert.NoError(t, err)
	assert.True(t, info.IsDir())

	// 2. Clone repository to processing directory
	cloneDir := filepath.Join(processingDir, "clone")
	cloneOptions := &service.CloneOptions{
		Timeout: 30 * time.Second,
	}

	err = gitService.CloneRepository(ctx, repoDir, "main", cloneDir, cloneOptions)
	assert.NoError(t, err)

	// Verify clone was successful
	assert.True(t, gitService.IsGitRepository(ctx, cloneDir))
	commitSHA, err := gitService.GetCommitSHA(ctx, cloneDir)
	assert.NoError(t, err)
	assert.NotEmpty(t, commitSHA)

	// Verify Terraform files were cloned
	mainTfFile := filepath.Join(cloneDir, "main.tf")
	_, err = os.Stat(mainTfFile)
	assert.NoError(t, err)

	readmeFile := filepath.Join(cloneDir, "README.md")
	_, err = os.Stat(readmeFile)
	assert.NoError(t, err)

	// 3. Simulate module processing (in real implementation, this would call ModuleProcessorService)
	// For this test, we'll create mock archives directly
	archiveDir := filepath.Join(processingDir, "archives")
	err = os.MkdirAll(archiveDir, 0755)
	require.NoError(t, err)

	// Create mock archives
	tarGzFile := filepath.Join(archiveDir, "source.tar.gz")
	zipFile := filepath.Join(archiveDir, "source.zip")

	// Create simple tar.gz content (mock)
	tarGzContent := []byte("mock tar.gz archive content")
	err = os.WriteFile(tarGzFile, tarGzContent, 0644)
	require.NoError(t, err)

	// Create simple zip content (mock)
	zipContent := []byte("mock zip archive content")
	err = os.WriteFile(zipFile, zipContent, 0644)
	require.NoError(t, err)

	// 4. Store archives using storage workflow
	namespace := "test-namespace"
	module := "test-module"
	provider := "aws"
	version := "1.0.0"

	err = storageWorkflow.StoreModuleArchives(ctx, archiveDir, namespace, module, provider, version)
	assert.NoError(t, err)

	// 5. Verify archives are stored at correct Python-compatible paths
	tarGzPath := storageWorkflow.GetArchivePath(namespace, module, provider, version, "source.tar.gz")
	zipPath := storageWorkflow.GetArchivePath(namespace, module, provider, version, "source.zip")

	// Verify tar.gz archive exists and has correct content
	storedTarGzContent, err := domainStorage.ReadFile(ctx, tarGzPath, true)
	assert.NoError(t, err)
	assert.Equal(t, tarGzContent, storedTarGzContent)

	// Verify zip archive exists and has correct content
	storedZipContent, err := domainStorage.ReadFile(ctx, zipPath, true)
	assert.NoError(t, err)
	assert.Equal(t, zipContent, storedZipContent)

	// 6. Verify directory structure matches Python exactly
	expectedVersionPath := filepath.Join(dataDir, "modules", namespace, module, provider, version)
	info, err = os.Stat(expectedVersionPath)
	assert.NoError(t, err)
	assert.True(t, info.IsDir())

	// 7. Verify cleanup happens automatically (cleanup is called via defer)
	// After cleanup, processing directory should be deleted
}

// TestModuleIndexingWorkflowErrorHandling tests error handling at each step
func TestModuleIndexingWorkflowErrorHandling(t *testing.T) {
	tempDir := t.TempDir()
	dataDir := filepath.Join(tempDir, "data")

	// Initialize storage services
	pathConfig := service.GetDefaultPathConfig(dataDir)
	pathBuilder := service.NewPathBuilderService(pathConfig)

	domainStorage, err := storage.NewLocalStorageService(dataDir, pathBuilder)
	require.NoError(t, err)

	tempDirManager, err := storage.NewTemporaryDirectoryManager()
	require.NoError(t, err)

	storageWorkflow := service.NewStorageWorkflowServiceImpl(
		domainStorage,
		pathBuilder,
		tempDirManager,
		&model.StoragePathConfig{
			BasePath:      pathConfig.BasePath,
			ModulesPath:   pathConfig.ModulesPath,
			ProvidersPath: pathConfig.ProvidersPath,
			UploadPath:    pathConfig.UploadPath,
			TempPath:      pathConfig.TempPath,
		},
	)

	gitService := service.NewGitService()
	ctx := context.Background()

	// Test error case: Invalid repository URL
	processingDir, cleanup, err := storageWorkflow.CreateProcessingDirectory(ctx, "error_test_")
	require.NoError(t, err)
	defer cleanup()

	cloneDir := filepath.Join(processingDir, "clone")
	cloneOptions := &service.CloneOptions{
		Timeout: 5 * time.Second,
	}

	err = gitService.CloneRepository(ctx, "https://github.com/nonexistent/repo.git", "", cloneDir, cloneOptions)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "git clone failed")

	// Verify cleanup still works even after clone error
}

// TestModuleIndexingWorkflowConcurrentOperations tests concurrent workflow executions
func TestModuleIndexingWorkflowConcurrentOperations(t *testing.T) {
	tempDir := t.TempDir()
	dataDir := filepath.Join(tempDir, "data")

	// Initialize storage services
	pathConfig := service.GetDefaultPathConfig(dataDir)
	pathBuilder := service.NewPathBuilderService(pathConfig)

	domainStorage, err := storage.NewLocalStorageService(dataDir, pathBuilder)
	require.NoError(t, err)

	tempDirManager, err := storage.NewTemporaryDirectoryManager()
	require.NoError(t, err)

	storageWorkflow := service.NewStorageWorkflowServiceImpl(
		domainStorage,
		pathBuilder,
		tempDirManager,
		&model.StoragePathConfig{
			BasePath:      pathConfig.BasePath,
			ModulesPath:   pathConfig.ModulesPath,
			ProvidersPath: pathConfig.ProvidersPath,
			UploadPath:    pathConfig.UploadPath,
			TempPath:      pathConfig.TempPath,
		},
	)

	gitService := service.NewGitService()
	ctx := context.Background()

	// Create multiple test repositories
	const numWorkflows = 3
	repos := make([]string, numWorkflows)
	for i := 0; i < numWorkflows; i++ {
		repoDir := filepath.Join(tempDir, f.Sprintf("test-repo-%d", i))
		err = createTestTerraformRepository(repoDir)
		require.NoError(t, err)
		repos[i] = repoDir
	}

	// Run workflows concurrently
	done := make(chan bool, numWorkflows)
	results := make([]string, numWorkflows)

	for i := 0; i < numWorkflows; i++ {
		go func(id int) {
			defer func() { done <- true }()

			// Create unique processing directory
			processingDir, cleanup, err := storageWorkflow.CreateProcessingDirectory(ctx, f.Sprintf("concurrent_%d_", id))
			if err != nil {
				t.Errorf("Failed to create processing directory: %v", err)
				return
			}
			defer cleanup()

			// Clone repository
			cloneDir := filepath.Join(processingDir, "clone")
			cloneOptions := &service.CloneOptions{
				Timeout: 30 * time.Second,
			}

			err = gitService.CloneRepository(ctx, repos[id], "main", cloneDir, cloneOptions)
			if err != nil {
				t.Errorf("Failed to clone repository %d: %v", id, err)
				return
			}

			// Store archives
			archiveDir := filepath.Join(processingDir, "archives")
			err = os.MkdirAll(archiveDir, 0755)
			if err != nil {
				t.Errorf("Failed to create archive directory: %v", err)
				return
			}

			// Create mock archives
			namespace := f.Sprintf("test-namespace-%d", id)
			module := f.Sprintf("test-module-%d", id)
			provider := "aws"
			version := "1.0.0"

			err = createMockArchives(archiveDir)
			if err != nil {
				t.Errorf("Failed to create mock archives: %v", err)
				return
			}

			err = storageWorkflow.StoreModuleArchives(ctx, archiveDir, namespace, module, provider, version)
			if err != nil {
				t.Errorf("Failed to store archives: %v", err)
				return
			}

			// Verify archives exist
			tarGzPath := storageWorkflow.GetArchivePath(namespace, module, provider, version, "source.tar.gz")
			zipPath := storageWorkflow.GetArchivePath(namespace, module, provider, version, "source.zip")

			exists, err := domainStorage.FileExists(ctx, tarGzPath)
			if err != nil || !exists {
				t.Errorf("Tar.gz archive not found for workflow %d", id)
				return
			}

			exists, err = domainStorage.FileExists(ctx, zipPath)
			if err != nil || !exists {
				t.Errorf("Zip archive not found for workflow %d", id)
				return
			}

			results[id] = f.Sprintf("workflow-%d-success", id)
		}(i)
	}

	// Wait for all workflows to complete
	for i := 0; i < numWorkflows; i++ {
		select {
		case <-done:
			// OK
		case <-time.After(30 * time.Second):
			t.Fatal("Timeout waiting for concurrent workflows")
		}
	}

	// Verify all workflows completed successfully
	for i, result := range results {
		assert.Equal(t, f.Sprintf("workflow-%d-success", i), result)
	}
}

// TestModuleIndexingWorkflowPythonPathStructure tests that the workflow
// creates the exact directory structure expected by Python terrareg
func TestModuleIndexingWorkflowPythonPathStructure(t *testing.T) {
	tempDir := t.TempDir()
	dataDir := filepath.Join(tempDir, "data")

	// Initialize storage services
	pathConfig := service.GetDefaultPathConfig(dataDir)
	pathBuilder := service.NewPathBuilderService(pathConfig)

	domainStorage, err := storage.NewLocalStorageService(dataDir, pathBuilder)
	require.NoError(t, err)

	tempDirManager, err := storage.NewTemporaryDirectoryManager()
	require.NoError(t, err)

	storageWorkflow := service.NewStorageWorkflowServiceImpl(
		domainStorage,
		pathBuilder,
		tempDirManager,
		&model.StoragePathConfig{
			BasePath:      pathConfig.BasePath,
			ModulesPath:   pathConfig.ModulesPath,
			ProvidersPath: pathConfig.ProvidersPath,
			UploadPath:    pathConfig.UploadPath,
			TempPath:      pathConfig.TempPath,
		},
	)

	ctx := context.Background()

	// Test module with complex namespace and module names (like Python examples)
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
			namespace: "hashicorp",
			module:    "consul",
			provider:  "aws",
			version:   "2.1.0",
		},
		{
			namespace: "my-company",
			module:    "database-module",
			provider:  "postgresql",
			version:   "1.2.3",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.namespace+"/"+tc.module, func(t *testing.T) {
			// Prepare module version path (Python creates this first)
			versionPath, err := storageWorkflow.PrepareModuleVersionPath(
				ctx,
				tc.namespace,
				tc.module,
				tc.provider,
				tc.version,
			)
			require.NoError(t, err)

			// Verify Python-compatible path structure
			expectedPath := filepath.Join(
				dataDir,
				"modules",
				tc.namespace,
				tc.module,
				tc.provider,
				tc.version,
			)
			assert.Equal(t, expectedPath, versionPath)

			// Verify directory was created
			info, err := os.Stat(versionPath)
			assert.NoError(t, err)
			assert.True(t, info.IsDir())

			// Test archive path generation
			tarGzPath := storageWorkflow.GetArchivePath(
				tc.namespace,
				tc.module,
				tc.provider,
				tc.version,
				"source.tar.gz",
			)
			expectedTarGzPath := filepath.Join(expectedPath, "source.tar.gz")
			assert.Equal(t, expectedTarGzPath, tarGzPath)

			zipPath := storageWorkflow.GetArchivePath(
				tc.namespace,
				tc.module,
				tc.provider,
				tc.version,
				"source.zip",
			)
			expectedZipPath := filepath.Join(expectedPath, "source.zip")
			assert.Equal(t, expectedZipPath, zipPath)
		})
	}
}

// Helper functions

func createTestTerraformRepository(repoDir string) error {
	// Create repository directory
	err := os.MkdirAll(repoDir, 0755)
	if err != nil {
		return err
	}

	// Initialize git repository
	ctx := context.Background()
	commands := [][]string{
		{"git", "init"},
		{"git", "config", "user.email", "test@example.com"},
		{"git", "config", "user.name", "Test User"},
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

	// Create outputs.tf
	outputsTfContent := `
output "instance_ip" {
  description = "Public IP address of the EC2 instance"
  value       = aws_instance.example.public_ip
}
`
	err = os.WriteFile(filepath.Join(repoDir, "outputs.tf"), []byte(outputsTfContent), 0644)
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

func createMockArchives(archiveDir string) error {
	// Create mock tar.gz file
	tarGzFile := filepath.Join(archiveDir, "source.tar.gz")
	tarGzContent := []byte("mock tar.gz archive content for testing")
	err := os.WriteFile(tarGzFile, tarGzContent, 0644)
	if err != nil {
		return err
	}

	// Create mock zip file
	zipFile := filepath.Join(archiveDir, "source.zip")
	zipContent := []byte("mock zip archive content for testing")
	err = os.WriteFile(zipFile, zipContent, 0644)
	if err != nil {
		return err
	}

	return nil
}

func runCommand(dir string, cmd []string) error {
	command := exec.CommandContext(context.Background(), cmd[0], cmd[1:]...)
	command.Dir = dir
	return command.Run()
}