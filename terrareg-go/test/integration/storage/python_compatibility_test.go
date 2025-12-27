package storage

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/storage/model"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/storage/service"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestPythonStorageStructureCompatibility verifies that the Go storage paths
// match the Python terrareg storage structure exactly
func TestPythonStorageStructureCompatibility(t *testing.T) {
	// Setup
	tempDir := t.TempDir()
	dataDir := filepath.Join(tempDir, "data")

	pathConfig := service.GetDefaultPathConfig(dataDir)

	// Verify Python-compatible path structure
	assert.Equal(t, dataDir+"/", pathConfig.BasePath)
	assert.Equal(t, dataDir+"/modules", pathConfig.ModulesPath)
	assert.Equal(t, dataDir+"/providers", pathConfig.ProvidersPath)
	assert.Equal(t, dataDir+"/upload", pathConfig.UploadPath)
	assert.Contains(t, pathConfig.TempPath, "terrareg")

	pathBuilder := service.NewPathBuilderService(pathConfig)

	// Test namespace path
	namespacePath := pathBuilder.BuildNamespacePath("test-namespace")
	expectedNamespacePath := filepath.Join(dataDir, "modules", "test-namespace")
	assert.Equal(t, expectedNamespacePath, namespacePath)

	// Test module path
	modulePath := pathBuilder.BuildModulePath("test-namespace", "test-module")
	expectedModulePath := filepath.Join(dataDir, "modules", "test-namespace", "test-module")
	assert.Equal(t, expectedModulePath, modulePath)

	// Test provider path
	providerPath := pathBuilder.BuildProviderPath("test-namespace", "test-module", "aws")
	expectedProviderPath := filepath.Join(dataDir, "modules", "test-namespace", "test-module", "aws")
	assert.Equal(t, expectedProviderPath, providerPath)

	// Test version path
	versionPath := pathBuilder.BuildVersionPath("test-namespace", "test-module", "aws", "1.0.0")
	expectedVersionPath := filepath.Join(dataDir, "modules", "test-namespace", "test-module", "aws", "1.0.0")
	assert.Equal(t, expectedVersionPath, versionPath)

	// Test archive paths (Python expects source.tar.gz and source.zip)
	archiveTarGzPath := pathBuilder.GetArchivePath("test-namespace", "test-module", "aws", "1.0.0", "source.tar.gz")
	expectedTarGzPath := filepath.Join(dataDir, "modules", "test-namespace", "test-module", "aws", "1.0.0", "source.tar.gz")
	assert.Equal(t, expectedTarGzPath, archiveTarGzPath)

	archiveZipPath := pathBuilder.GetArchivePath("test-namespace", "test-module", "aws", "1.0.0", "source.zip")
	expectedZipPath := filepath.Join(dataDir, "modules", "test-namespace", "test-module", "aws", "1.0.0", "source.zip")
	assert.Equal(t, expectedZipPath, archiveZipPath)
}

// TestPythonArchiveNamingCompatibility verifies that archive generation
// produces files with the exact names expected by Python
func TestPythonArchiveNamingCompatibility(t *testing.T) {
	// Setup
	tempDir := t.TempDir()
	dataDir := filepath.Join(tempDir, "data")

	pathConfig := service.GetDefaultPathConfig(dataDir)
	pathBuilder := service.NewPathBuilderService(pathConfig)

	// Test all possible archive names that Python expects
	testCases := []struct {
		namespace   string
		module      string
		provider    string
		version     string
		archiveName string
		expectedPath string
	}{
		{
			namespace:   "hashicorp",
			module:      "consul",
			provider:    "aws",
			version:     "2.1.0",
			archiveName: "source.tar.gz",
			expectedPath: filepath.Join(dataDir, "modules", "hashicorp", "consul", "aws", "2.1.0", "source.tar.gz"),
		},
		{
			namespace:   "terraform-aws-modules",
			module:      "vpc",
			provider:    "aws",
			version:     "3.0.0",
			archiveName: "source.zip",
			expectedPath: filepath.Join(dataDir, "modules", "terraform-aws-modules", "vpc", "aws", "3.0.0", "source.zip"),
		},
		{
			namespace:   "my-company",
			module:      "database",
			provider:    "postgresql",
			version:     "1.2.3",
			archiveName: "source.tar.gz",
			expectedPath: filepath.Join(dataDir, "modules", "my-company", "database", "postgresql", "1.2.3", "source.tar.gz"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.namespace+"/"+tc.module, func(t *testing.T) {
			archivePath := pathBuilder.GetArchivePath(tc.namespace, tc.module, tc.provider, tc.version, tc.archiveName)
			assert.Equal(t, tc.expectedPath, archivePath)

			// Verify path builder can identify archive paths correctly
			assert.True(t, pathBuilder.IsArchivePath(archivePath))
			assert.Equal(t, tc.archiveName, pathBuilder.GetArchiveName(archivePath))
		})
	}
}

// TestPythonSafeJoinPathsCompatibility verifies that path joining
// behaves exactly like Python's safe_join_paths function
func TestPythonSafeJoinPathsCompatibility(t *testing.T) {
	tempDir := t.TempDir()
	pathConfig := service.GetDefaultPathConfig(tempDir)
	pathBuilder := service.NewPathBuilderService(pathConfig)

	testCases := []struct {
		name         string
		basePath     string
		subPaths     []string
		expectedPath string
	}{
		{
			name:         "Simple path join",
			basePath:     "/data",
			subPaths:     []string{"modules", "test"},
			expectedPath: "/data/modules/test",
		},
		{
			name:         "Single subpath",
			basePath:     "/data/modules",
			subPaths:     []string{"namespace"},
			expectedPath: "/data/modules/namespace",
		},
		{
			name:         "Empty subpath ignored",
			basePath:     "/data",
			subPaths:     []string{"", "modules", "test"},
			expectedPath: "/data/modules/test",
		},
		{
			name:         "Leading slash in subpath trimmed",
			basePath:     "/data",
			subPaths:     []string{"/modules", "test"},
			expectedPath: "/data/modules/test",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := pathBuilder.SafeJoinPaths(tc.basePath, tc.subPaths...)
			assert.Equal(t, tc.expectedPath, result)
		})
	}
}

// TestPythonPathTraversalProtection verifies that path traversal
// attempts are blocked exactly like Python's implementation
func TestPythonPathTraversalProtection(t *testing.T) {
	tempDir := t.TempDir()
	pathConfig := service.GetDefaultPathConfig(tempDir)
	pathBuilder := service.NewPathBuilderService(pathConfig)

	testCases := []struct {
		name        string
		basePath    string
		subPaths    []string
		expectError bool
	}{
		{
			name:        "Relative path traversal blocked",
			basePath:    "/data/modules",
			subPaths:    []string{"../../etc/passwd"},
			expectError: true,
		},
		{
			name:        "Absolute path traversal blocked",
			basePath:    "/data/modules",
			subPaths:    []string{"/etc/passwd"},
			expectError: true,
		},
		{
			name:        "Multiple traversal attempts blocked",
			basePath:    "/data/modules",
			subPaths:    []string{"test", "../../../etc/passwd"},
			expectError: true,
		},
		{
			name:        "Safe path allowed",
			basePath:    "/data/modules",
			subPaths:    []string{"test", "subdir"},
			expectError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := pathBuilder.SafeJoinPaths(tc.basePath, tc.subPaths...)

			if tc.expectError {
				// Should contain path traversal attempts
				assert.Contains(t, result, "..")
			} else {
				// Should be a valid path without traversal
				assert.NotContains(t, result, "..")
				assert.True(t, filepath.IsAbs(result) || filepath.IsLocal(result))
			}

			// Test path validation
			err := pathBuilder.ValidatePath(result)
			if tc.expectError {
				assert.Error(t, err)
				assert.Equal(t, model.ErrPathTraversal, err)
			} else {
				// Note: The validation might still error for absolute paths not in base
				// but should not error for path traversal
				assert.NotEqual(t, model.ErrPathTraversal, err)
			}
		})
	}
}

// TestPythonStorageWorkflowCompatibility tests the complete workflow
// to ensure it matches Python behavior
func TestPythonStorageWorkflowCompatibility(t *testing.T) {
	// Setup
	tempDir := t.TempDir()
	dataDir := filepath.Join(tempDir, "data")

	pathConfig := service.GetDefaultPathConfig(dataDir)
	pathBuilder := service.NewPathBuilderService(pathConfig)

	domainStorage, err := storage.NewLocalStorageService(dataDir, pathBuilder)
	require.NoError(t, err)

	tempDirManager, err := storage.NewTemporaryDirectoryManager()
	require.NoError(t, err)

	workflow := service.NewStorageWorkflowServiceImpl(
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

	// Simulate Python workflow: Create module version structure
	namespace := "hashicorp"
	module := "consul"
	provider := "aws"
	version := "2.1.0"

	// Python creates the version directory first
	versionPath, err := workflow.PrepareModuleVersionPath(ctx, namespace, module, provider, version)
	require.NoError(t, err)

	// Verify Python-compatible path structure
	expectedPath := filepath.Join(dataDir, "modules", namespace, module, provider, version)
	assert.Equal(t, expectedPath, versionPath)

	// Python creates archives in the version directory
	tarGzPath := workflow.GetArchivePath(namespace, module, provider, version, "source.tar.gz")
	zipPath := workflow.GetArchivePath(namespace, module, provider, version, "source.zip")

	// Verify Python expected archive names and paths
	assert.Equal(t, expectedPath+"/source.tar.gz", tarGzPath)
	assert.Equal(t, expectedPath+"/source.zip", zipPath)

	// Python checks that archives exist before serving
	// Create mock archives to simulate this
	tarGzContent := []byte("mock tar.gz content")
	zipContent := []byte("mock zip content")

	err = domainStorage.WriteFile(ctx, tarGzPath, tarGzContent, true)
	require.NoError(t, err)

	err = domainStorage.WriteFile(ctx, zipPath, zipContent, true)
	require.NoError(t, err)

	// Verify Python can read archives
	readTarGz, err := domainStorage.ReadFile(ctx, tarGzPath, true)
	assert.NoError(t, err)
	assert.Equal(t, tarGzContent, readTarGz)

	readZip, err := domainStorage.ReadFile(ctx, zipPath, true)
	assert.NoError(t, err)
	assert.Equal(t, zipContent, readZip)

	// Python checks file existence before operations
	exists, err := domainStorage.FileExists(ctx, tarGzPath)
	assert.NoError(t, err)
	assert.True(t, exists)

	exists, err = domainStorage.FileExists(ctx, zipPath)
	assert.NoError(t, err)
	assert.True(t, exists)

	// Python also checks directory existence
	exists, err = domainStorage.DirectoryExists(ctx, versionPath)
	assert.NoError(t, err)
	assert.True(t, exists)
}

// TestPythonDirectoryOperationsCompatibility verifies that directory
// operations behave like Python's implementation
func TestPythonDirectoryOperationsCompatibility(t *testing.T) {
	tempDir := t.TempDir()
	pathConfig := service.GetDefaultPathConfig(tempDir)
	pathBuilder := service.NewPathBuilderService(pathConfig)

	domainStorage, err := storage.NewLocalStorageService(tempDir, pathBuilder)
	require.NoError(t, err)

	ctx := context.Background()

	// Python creates directories automatically
	testDir := filepath.Join(tempDir, "modules", "test", "subdir")
	err = domainStorage.MakeDirectory(ctx, testDir)
	assert.NoError(t, err)

	// Python checks directory existence
	exists, err := domainStorage.DirectoryExists(ctx, testDir)
	assert.NoError(t, err)
	assert.True(t, exists)

	// Python distinguishes between file and directory existence
	exists, err = domainStorage.FileExists(ctx, testDir)
	assert.NoError(t, err)
	assert.False(t, exists) // Directories return false for FileExists

	// Python cleanup operations
	err = domainStorage.DeleteDirectory(ctx, testDir)
	assert.NoError(t, err)

	exists, err = domainStorage.DirectoryExists(ctx, testDir)
	assert.NoError(t, err)
	assert.False(t, exists)
}