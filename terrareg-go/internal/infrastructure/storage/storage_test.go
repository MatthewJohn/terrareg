package storage

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/storage/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBasicStorageWorkflow(t *testing.T) {
	// Simple test to verify basic storage functionality works
	tempDir := t.TempDir()
	pathConfig := service.GetDefaultPathConfig(tempDir)
	pathBuilder := service.NewPathBuilderService(pathConfig)

	// Test path builder
	namespacePath := pathBuilder.BuildNamespacePath("test")
	assert.Equal(t, filepath.Join(tempDir, "modules", "test"), namespacePath)

	// Test LocalStorageService creation
	storage, err := NewLocalStorageService(tempDir, pathBuilder)
	require.NoError(t, err)
	assert.NotNil(t, storage)

	// Test basic file operations
	ctx := context.Background()
	testFile := filepath.Join(tempDir, "test.txt")

	// Write file
	err = storage.WriteFile(ctx, testFile, "test content", false)
	assert.NoError(t, err)

	// Check file exists
	exists, err := storage.FileExists(ctx, testFile)
	assert.NoError(t, err)
	assert.True(t, exists)

	// Read file
	content, err := storage.ReadFile(ctx, testFile, true)
	assert.NoError(t, err)
	assert.Equal(t, []byte("test content"), content)

	// Delete file
	err = storage.DeleteFile(ctx, testFile)
	assert.NoError(t, err)

	// Verify file is deleted
	exists, err = storage.FileExists(ctx, testFile)
	assert.NoError(t, err)
	assert.False(t, exists)
}

func TestCreateTemporaryStorageService(t *testing.T) {
	// Test that CreateTemporaryStorageService creates a LocalStorageService
	// scoped to the specified temporary directory
	tempBaseDir := t.TempDir()
	pathConfig := service.GetDefaultPathConfig(tempBaseDir)
	pathBuilder := service.NewPathBuilderService(pathConfig)

	factory := NewStorageFactory(pathBuilder)

	// Create a temporary directory to use as base
	scopedTempDir, err := os.MkdirTemp(tempBaseDir, "test-scope-*")
	require.NoError(t, err)
	defer os.RemoveAll(scopedTempDir)

	// Create a storage service scoped to this directory
	scopedStorage, err := factory.CreateTemporaryStorageService(scopedTempDir)
	require.NoError(t, err)
	assert.NotNil(t, scopedStorage)

	// Verify it's a LocalStorageService
	localStorage, ok := scopedStorage.(*LocalStorageService)
	require.True(t, ok, "CreateTemporaryStorageService should return *LocalStorageService")

	// Verify the basePath is set to the scoped directory
	// We can verify this by checking that files written to the scoped directory
	// can be read back, but files outside cannot
	ctx := context.Background()

	// Write a test file inside the scoped directory
	testFile := filepath.Join(scopedTempDir, "test.txt")
	err = localStorage.WriteFile(ctx, testFile, "test content", false)
	require.NoError(t, err)

	// Verify we can read the file back
	content, err := localStorage.ReadFile(ctx, testFile, true)
	require.NoError(t, err)
	assert.Equal(t, []byte("test content"), content)

	// Verify that the file exists in the scoped directory
	exists, err := localStorage.FileExists(ctx, testFile)
	require.NoError(t, err)
	assert.True(t, exists)

	// Create a file outside the scoped directory to verify security
	outsideFile := filepath.Join(tempBaseDir, "outside.txt")
	err = os.WriteFile(outsideFile, []byte("outside content"), 0644)
	require.NoError(t, err)

	// Verify that reading from outside the scoped directory is properly handled
	// (The generatePath method should return the basePath for paths outside)
	// This is a security feature to prevent path traversal
	content, err = localStorage.ReadFile(ctx, outsideFile, true)
	// Expected behavior: ReadFile should fail because:
	// 1. generatePath returns basePath (the scoped directory) for paths outside
	// 2. Attempting to read a directory as a file returns an error
	assert.Error(t, err, "Reading file outside scoped directory should return an error")
	assert.Nil(t, content, "Content should be nil when error occurs")
}
