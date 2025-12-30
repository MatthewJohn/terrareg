package storage

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/storage/model"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/storage/service"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStorageWorkflowService_PrepareModuleVersionPath(t *testing.T) {
	// Setup
	tempDir := t.TempDir()
	servicePathConfig := &service.StoragePathConfig{
		BasePath:    tempDir,
		ModulesPath: filepath.Join(tempDir, "modules"),
	}
	modelPathConfig := &model.StoragePathConfig{
		BasePath:    tempDir,
		ModulesPath: filepath.Join(tempDir, "modules"),
	}

	pathBuilder := service.NewPathBuilderService(servicePathConfig)
	domainStorage, err := storage.NewLocalStorageService(tempDir, pathBuilder)
	require.NoError(t, err)

	tempDirManager, err := storage.NewTemporaryDirectoryManager()
	require.NoError(t, err)

	workflow := service.NewStorageWorkflowServiceImpl(
		domainStorage,
		pathBuilder,
		tempDirManager,
		modelPathConfig,
	)

	ctx := context.Background()

	// Test path preparation
	path, err := workflow.PrepareModuleVersionPath(
		ctx,
		"test-namespace",
		"test-module",
		"test-provider",
		"1.0.0",
	)

	assert.NoError(t, err)
	expectedPath := filepath.Join(tempDir, "modules", "test-namespace", "test-module", "test-provider", "1.0.0")
	assert.Equal(t, expectedPath, path)

	// Verify directory was created
	info, err := os.Stat(path)
	assert.NoError(t, err)
	assert.True(t, info.IsDir())
}

func TestStorageWorkflowService_GetArchivePath(t *testing.T) {
	// Setup
	tempDir := t.TempDir()
	servicePathConfig := &service.StoragePathConfig{
		BasePath:    tempDir,
		ModulesPath: filepath.Join(tempDir, "modules"),
	}
	modelPathConfig := &model.StoragePathConfig{
		BasePath:    tempDir,
		ModulesPath: filepath.Join(tempDir, "modules"),
	}

	pathBuilder := service.NewPathBuilderService(servicePathConfig)
	domainStorage, err := storage.NewLocalStorageService(tempDir, pathBuilder)
	require.NoError(t, err)

	tempDirManager, err := storage.NewTemporaryDirectoryManager()
	require.NoError(t, err)

	workflow := service.NewStorageWorkflowServiceImpl(
		domainStorage,
		pathBuilder,
		tempDirManager,
		modelPathConfig,
	)

	// Test archive path generation for tar.gz
	path := workflow.GetArchivePath(
		"test-namespace",
		"test-module",
		"test-provider",
		"1.0.0",
		"source.tar.gz",
	)

	expectedPath := filepath.Join(
		tempDir,
		"modules",
		"test-namespace",
		"test-module",
		"test-provider",
		"1.0.0",
		"source.tar.gz",
	)
	assert.Equal(t, expectedPath, path)

	// Test archive path generation for zip
	path = workflow.GetArchivePath(
		"test-namespace",
		"test-module",
		"test-provider",
		"1.0.0",
		"source.zip",
	)

	expectedPath = filepath.Join(
		tempDir,
		"modules",
		"test-namespace",
		"test-module",
		"test-provider",
		"1.0.0",
		"source.zip",
	)
	assert.Equal(t, expectedPath, path)
}

func TestStorageWorkflowService_CreateProcessingDirectory(t *testing.T) {
	// Setup
	tempDir := t.TempDir()
	servicePathConfig := &service.StoragePathConfig{
		BasePath:    tempDir,
		ModulesPath: filepath.Join(tempDir, "modules"),
	}
	modelPathConfig := &model.StoragePathConfig{
		BasePath:    tempDir,
		ModulesPath: filepath.Join(tempDir, "modules"),
	}

	pathBuilder := service.NewPathBuilderService(servicePathConfig)
	domainStorage, err := storage.NewLocalStorageService(tempDir, pathBuilder)
	require.NoError(t, err)

	tempDirManager, err := storage.NewTemporaryDirectoryManager()
	require.NoError(t, err)

	workflow := service.NewStorageWorkflowServiceImpl(
		domainStorage,
		pathBuilder,
		tempDirManager,
		modelPathConfig,
	)

	ctx := context.Background()

	// Test processing directory creation
	tempPath, cleanup, err := workflow.CreateProcessingDirectory(ctx, "test_processing_")
	assert.NoError(t, err)
	assert.NotNil(t, cleanup)
	assert.NotEmpty(t, tempPath)

	// Verify directory exists
	info, err := os.Stat(tempPath)
	assert.NoError(t, err)
	assert.True(t, info.IsDir())

	// Verify directory has correct prefix
	assert.Contains(t, filepath.Base(tempPath), "test_processing_")

	// Test cleanup function
	cleanup()

	// Verify directory is deleted after cleanup
	_, err = os.Stat(tempPath)
	assert.True(t, os.IsNotExist(err))
}

func TestStorageWorkflowService_CreateExtractionDirectory(t *testing.T) {
	// Setup
	tempDir := t.TempDir()
	servicePathConfig := &service.StoragePathConfig{
		BasePath:    tempDir,
		ModulesPath: filepath.Join(tempDir, "modules"),
	}
	modelPathConfig := &model.StoragePathConfig{
		BasePath:    tempDir,
		ModulesPath: filepath.Join(tempDir, "modules"),
	}

	pathBuilder := service.NewPathBuilderService(servicePathConfig)
	domainStorage, err := storage.NewLocalStorageService(tempDir, pathBuilder)
	require.NoError(t, err)

	tempDirManager, err := storage.NewTemporaryDirectoryManager()
	require.NoError(t, err)

	workflow := service.NewStorageWorkflowServiceImpl(
		domainStorage,
		pathBuilder,
		tempDirManager,
		modelPathConfig,
	)

	ctx := context.Background()
	moduleVersionID := 12345

	// Test extraction directory creation
	tempPath, cleanup, err := workflow.CreateExtractionDirectory(ctx, moduleVersionID)
	assert.NoError(t, err)
	assert.NotNil(t, cleanup)
	assert.NotEmpty(t, tempPath)

	// Verify directory exists
	info, err := os.Stat(tempPath)
	assert.NoError(t, err)
	assert.True(t, info.IsDir())

	// Verify directory name contains module version ID
	assert.Contains(t, filepath.Base(tempPath), "extract_12345_")

	// Test cleanup function
	cleanup()

	// Verify directory is deleted after cleanup
	_, err = os.Stat(tempPath)
	assert.True(t, os.IsNotExist(err))
}

func TestStorageWorkflowService_StoreModuleArchives(t *testing.T) {
	// Setup
	tempDir := t.TempDir()
	servicePathConfig := &service.StoragePathConfig{
		BasePath:    tempDir,
		ModulesPath: filepath.Join(tempDir, "modules"),
	}
	modelPathConfig := &model.StoragePathConfig{
		BasePath:    tempDir,
		ModulesPath: filepath.Join(tempDir, "modules"),
	}

	pathBuilder := service.NewPathBuilderService(servicePathConfig)
	domainStorage, err := storage.NewLocalStorageService(tempDir, pathBuilder)
	require.NoError(t, err)

	tempDirManager, err := storage.NewTemporaryDirectoryManager()
	require.NoError(t, err)

	workflow := service.NewStorageWorkflowServiceImpl(
		domainStorage,
		pathBuilder,
		tempDirManager,
		modelPathConfig,
	)

	ctx := context.Background()

	// Create test archives in temp directory
	sourceDir := t.TempDir()
	archivesDir := filepath.Join(sourceDir, "archives")
	err = os.MkdirAll(archivesDir, 0755)
	require.NoError(t, err)

	// Create test tar.gz file
	tarGzFile := filepath.Join(archivesDir, "source.tar.gz")
	tarGzContent := []byte("mock tar.gz content")
	err = os.WriteFile(tarGzFile, tarGzContent, 0644)
	require.NoError(t, err)

	// Create test zip file
	zipFile := filepath.Join(archivesDir, "source.zip")
	zipContent := []byte("mock zip content")
	err = os.WriteFile(zipFile, zipContent, 0644)
	require.NoError(t, err)

	// Test archive storage
	err = workflow.StoreModuleArchives(
		ctx,
		archivesDir,
		"test-namespace",
		"test-module",
		"test-provider",
		"1.0.0",
	)

	assert.NoError(t, err)

	// Verify tar.gz archive was stored
	targetTarGzPath := filepath.Join(
		tempDir,
		"modules",
		"test-namespace",
		"test-module",
		"test-provider",
		"1.0.0",
		"source.tar.gz",
	)

	storedTarGzContent, err := os.ReadFile(targetTarGzPath)
	assert.NoError(t, err)
	assert.Equal(t, tarGzContent, storedTarGzContent)

	// Verify zip archive was stored
	targetZipPath := filepath.Join(
		tempDir,
		"modules",
		"test-namespace",
		"test-module",
		"test-provider",
		"1.0.0",
		"source.zip",
	)

	storedZipContent, err := os.ReadFile(targetZipPath)
	assert.NoError(t, err)
	assert.Equal(t, zipContent, storedZipContent)
}

func TestStorageWorkflowService_CleanupDirectory(t *testing.T) {
	// Setup
	tempDir := t.TempDir()
	servicePathConfig := &service.StoragePathConfig{
		BasePath:    tempDir,
		ModulesPath: filepath.Join(tempDir, "modules"),
	}
	modelPathConfig := &model.StoragePathConfig{
		BasePath:    tempDir,
		ModulesPath: filepath.Join(tempDir, "modules"),
	}

	pathBuilder := service.NewPathBuilderService(servicePathConfig)
	domainStorage, err := storage.NewLocalStorageService(tempDir, pathBuilder)
	require.NoError(t, err)

	tempDirManager, err := storage.NewTemporaryDirectoryManager()
	require.NoError(t, err)

	workflow := service.NewStorageWorkflowServiceImpl(
		domainStorage,
		pathBuilder,
		tempDirManager,
		modelPathConfig,
	)

	ctx := context.Background()

	// Create test directory to cleanup
	testDir := filepath.Join(tempDir, "to_cleanup")
	err = os.MkdirAll(testDir, 0755)
	require.NoError(t, err)

	// Add some content
	subDir := filepath.Join(testDir, "subdir")
	err = os.MkdirAll(subDir, 0755)
	require.NoError(t, err)

	testFile := filepath.Join(testDir, "test.txt")
	err = os.WriteFile(testFile, []byte("content"), 0644)
	require.NoError(t, err)

	// Verify directory exists before cleanup
	_, err = os.Stat(testDir)
	assert.NoError(t, err)

	// Test cleanup
	err = workflow.CleanupDirectory(ctx, testDir)
	assert.NoError(t, err)

	// Verify directory is deleted
	_, err = os.Stat(testDir)
	assert.True(t, os.IsNotExist(err))

	// Test cleanup of non-existent directory (should not error)
	err = workflow.CleanupDirectory(ctx, filepath.Join(tempDir, "nonexistent"))
	assert.NoError(t, err)
}

func TestStorageWorkflowService_IntegrationWorkflow(t *testing.T) {
	// Integration test for complete workflow:
	// 1. Create processing directory
	// 2. Store archives
	// 3. Cleanup

	tempDir := t.TempDir()
	servicePathConfig := &service.StoragePathConfig{
		BasePath:    tempDir,
		ModulesPath: filepath.Join(tempDir, "modules"),
	}
	modelPathConfig := &model.StoragePathConfig{
		BasePath:    tempDir,
		ModulesPath: filepath.Join(tempDir, "modules"),
	}

	pathBuilder := service.NewPathBuilderService(servicePathConfig)
	domainStorage, err := storage.NewLocalStorageService(tempDir, pathBuilder)
	require.NoError(t, err)

	tempDirManager, err := storage.NewTemporaryDirectoryManager()
	require.NoError(t, err)

	workflow := service.NewStorageWorkflowServiceImpl(
		domainStorage,
		pathBuilder,
		tempDirManager,
		modelPathConfig,
	)

	ctx := context.Background()

	// 1. Create processing directory for a module version
	processingDir, cleanupProcessing, err := workflow.CreateProcessingDirectory(ctx, "integration_test_")
	require.NoError(t, err)
	defer cleanupProcessing()

	// 2. Prepare module version path
	versionPath, err := workflow.PrepareModuleVersionPath(
		ctx,
		"integration-namespace",
		"integration-module",
		"integration-provider",
		"1.0.0",
	)
	require.NoError(t, err)

	// Verify version path was created
	info, err := os.Stat(versionPath)
	assert.NoError(t, err)
	assert.True(t, info.IsDir())

	// 3. Create mock archives in processing directory
	archivesDir := filepath.Join(processingDir, "archives")
	err = os.MkdirAll(archivesDir, 0755)
	require.NoError(t, err)

	// Create mock tar.gz
	tarGzFile := filepath.Join(archivesDir, "source.tar.gz")
	err = os.WriteFile(tarGzFile, []byte("integration test tar.gz"), 0644)
	require.NoError(t, err)

	// Create mock zip
	zipFile := filepath.Join(archivesDir, "source.zip")
	err = os.WriteFile(zipFile, []byte("integration test zip"), 0644)
	require.NoError(t, err)

	// 4. Store archives
	err = workflow.StoreModuleArchives(
		ctx,
		archivesDir,
		"integration-namespace",
		"integration-module",
		"integration-provider",
		"1.0.0",
	)
	assert.NoError(t, err)

	// 5. Verify archives are stored at correct paths
	tarGzPath := workflow.GetArchivePath(
		"integration-namespace",
		"integration-module",
		"integration-provider",
		"1.0.0",
		"source.tar.gz",
	)

	zipPath := workflow.GetArchivePath(
		"integration-namespace",
		"integration-module",
		"integration-provider",
		"1.0.0",
		"source.zip",
	)

	// Verify tar.gz exists and has correct content
	tarGzContent, err := os.ReadFile(tarGzPath)
	assert.NoError(t, err)
	assert.Equal(t, []byte("integration test tar.gz"), tarGzContent)

	// Verify zip exists and has correct content
	zipContent, err := os.ReadFile(zipPath)
	assert.NoError(t, err)
	assert.Equal(t, []byte("integration test zip"), zipContent)

	// 6. Test archive path generation matches expected pattern
	expectedTarGzPath := filepath.Join(
		versionPath,
		"source.tar.gz",
	)
	assert.Equal(t, expectedTarGzPath, tarGzPath)

	expectedZipPath := filepath.Join(
		versionPath,
		"source.zip",
	)
	assert.Equal(t, expectedZipPath, zipPath)
}

func TestStorageWorkflowService_ConcurrentOperations(t *testing.T) {
	// Test concurrent processing directory creation and cleanup
	tempDir := t.TempDir()
	servicePathConfig := &service.StoragePathConfig{
		BasePath:    tempDir,
		ModulesPath: filepath.Join(tempDir, "modules"),
	}
	modelPathConfig := &model.StoragePathConfig{
		BasePath:    tempDir,
		ModulesPath: filepath.Join(tempDir, "modules"),
	}

	pathBuilder := service.NewPathBuilderService(servicePathConfig)
	domainStorage, err := storage.NewLocalStorageService(tempDir, pathBuilder)
	require.NoError(t, err)

	tempDirManager, err := storage.NewTemporaryDirectoryManager()
	require.NoError(t, err)

	workflow := service.NewStorageWorkflowServiceImpl(
		domainStorage,
		pathBuilder,
		tempDirManager,
		modelPathConfig,
	)

	ctx := context.Background()

	// Create multiple processing directories concurrently
	const numDirectories = 10
	done := make(chan bool, numDirectories)
	directories := make([]string, numDirectories)
	cleanups := make([]func(), numDirectories)

	for i := 0; i < numDirectories; i++ {
		go func(id int) {
			defer func() { done <- true }()

			tempPath, cleanup, err := workflow.CreateProcessingDirectory(ctx, fmt.Sprintf("concurrent_%d_", id))
			assert.NoError(t, err)
			assert.NotNil(t, cleanup)

			directories[id] = tempPath
			cleanups[id] = cleanup

			// Verify directory exists
			info, err := os.Stat(tempPath)
			assert.NoError(t, err)
			assert.True(t, info.IsDir())
		}(i)
	}

	// Wait for all operations to complete
	for i := 0; i < numDirectories; i++ {
		select {
		case <-done:
			// OK
		case <-time.After(5 * time.Second):
			t.Fatal("Timeout waiting for concurrent operations")
		}
	}

	// Verify all directories are unique
	dirSet := make(map[string]bool)
	for _, dir := range directories {
		assert.False(t, dirSet[dir], "Directory path should be unique: %s", dir)
		dirSet[dir] = true
	}

	// Cleanup all directories concurrently
	for i := 0; i < numDirectories; i++ {
		go func(id int) {
			defer func() { done <- true }()
			cleanups[id]()
		}(i)
	}

	// Wait for all cleanup operations
	for i := 0; i < numDirectories; i++ {
		select {
		case <-done:
			// OK
		case <-time.After(5 * time.Second):
			t.Fatal("Timeout waiting for cleanup operations")
		}
	}

	// Verify all directories are deleted
	for _, dir := range directories {
		_, err := os.Stat(dir)
		assert.True(t, os.IsNotExist(err), "Directory should be deleted: %s", dir)
	}
}