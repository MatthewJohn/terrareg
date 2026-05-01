package module

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/storage/service"
	storageInfrastructure "github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/storage"
)

// TestTemporaryDirectoryScopedStorageIntegration tests the complete flow of
// creating a scoped storage service for a temporary directory and verifying
// that it properly restricts file access to within that directory.
func TestTemporaryDirectoryScopedStorageIntegration(t *testing.T) {
	// Create a base temporary directory for testing
	baseTempDir := t.TempDir()

	// Create a path builder
	pathConfig := service.GetDefaultPathConfig(baseTempDir)
	pathBuilder := service.NewPathBuilderService(pathConfig)

	// Create a storage factory
	factory := storageInfrastructure.NewStorageFactory(pathBuilder)

	// Test Case 1: Create a temporary directory for processing
	tempDir, err := os.MkdirTemp(baseTempDir, "test-module-processing-*")
	require.NoError(t, err, "Should create temp directory")
	defer os.RemoveAll(tempDir) // Cleanup

	// Test Case 2: Create a scoped storage service for this temp directory
	scopedStorage, err := factory.CreateTemporaryStorageService(tempDir)
	require.NoError(t, err, "Should create scoped storage service")
	assert.NotNil(t, scopedStorage, "Scoped storage service should not be nil")

	// Verify it's a LocalStorageService
	localStorage, ok := scopedStorage.(*storageInfrastructure.LocalStorageService)
	require.True(t, ok, "CreateTemporaryStorageService should return *LocalStorageService")

	// Test Case 3: Write a file to the scoped directory and read it back
	ctx := context.Background()
	testFilePath := filepath.Join(tempDir, "test.tf")
	testContent := `resource "aws_s3_bucket" "example" {
	bucket = "my-test-bucket"
}`

	err = localStorage.WriteFile(ctx, testFilePath, testContent, false)
	require.NoError(t, err, "Should write file to scoped directory")

	// Verify the file exists
	exists, err := localStorage.FileExists(ctx, testFilePath)
	require.NoError(t, err, "Should check file existence")
	assert.True(t, exists, "File should exist in scoped directory")

	// Read the file back
	readContent, err := localStorage.ReadFile(ctx, testFilePath, true)
	require.NoError(t, err, "Should read file from scoped directory")
	assert.Equal(t, []byte(testContent), readContent, "Read content should match written content")

	// Test Case 4: Verify security - files outside the scoped directory are protected
	outsideFile := filepath.Join(baseTempDir, "outside.txt")
	outsideContent := []byte("This should not be readable")

	err = os.WriteFile(outsideFile, outsideContent, 0644)
	require.NoError(t, err, "Should create file outside scoped directory")

	// Attempting to read the file outside the scoped directory should fail
	// or return an error (not the outside content)
	readOutsideContent, err := localStorage.ReadFile(ctx, outsideFile, true)
	assert.Error(t, err, "Reading file outside scoped directory should return an error")
	assert.Nil(t, readOutsideContent, "Content should be nil when error occurs")

	// Test Case 5: Verify the scoped storage can be wrapped with ModuleStorageAdapter
	// for use in the module layer
	adapterStorage := storageInfrastructure.NewModuleStorageAdapter(scopedStorage, pathBuilder)
	assert.NotNil(t, adapterStorage, "ModuleStorageAdapter should be created")

	// Test Case 6: Verify cleanup - the temp directory can be removed
	err = localStorage.DeleteDirectory(ctx, tempDir)
	require.NoError(t, err, "Should be able to delete the scoped directory")

	// Verify directory is gone
	exists, err = localStorage.DirectoryExists(ctx, tempDir)
	require.NoError(t, err, "Should check directory existence")
	assert.False(t, exists, "Directory should not exist after deletion")
}
