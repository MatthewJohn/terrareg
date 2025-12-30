package storage

import (
	"context"
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