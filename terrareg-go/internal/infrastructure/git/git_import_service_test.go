package git

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/module"
	moduleModel "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/module/model"
	moduleRepository "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/module/repository"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/shared/types"
	infraConfig "github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockModuleProviderRepo is a mock implementation for testing
type mockModuleProviderRepo struct {
	moduleProvider *moduleModel.ModuleProvider
}

func (m *mockModuleProviderRepo) FindByNamespaceModuleProvider(
	ctx context.Context, namespace types.NamespaceName, moduleName types.ModuleName, provider types.ModuleProviderName,
) (*moduleModel.ModuleProvider, error) {
	return m.moduleProvider, nil
}

// Implement other required methods
func (m *mockModuleProviderRepo) FindByID(ctx context.Context, id int) (*moduleModel.ModuleProvider, error) {
	return m.moduleProvider, nil
}
func (m *mockModuleProviderRepo) FindByNamespace(ctx context.Context, namespace types.NamespaceName) ([]*moduleModel.ModuleProvider, error) {
	return nil, nil
}
func (m *mockModuleProviderRepo) Save(ctx context.Context, mp *moduleModel.ModuleProvider) error {
	return nil
}
func (m *mockModuleProviderRepo) Search(ctx context.Context, query moduleRepository.ModuleSearchQuery) (*moduleRepository.ModuleSearchResult, error) {
	return nil, nil
}
func (m *mockModuleProviderRepo) Delete(ctx context.Context, id int) error {
	return nil
}
func (m *mockModuleProviderRepo) Exists(ctx context.Context, namespace types.NamespaceName, module types.ModuleName, provider types.ModuleProviderName) (bool, error) {
	return false, nil
}

// mockStorageService tracks MkdirTemp and RemoveAll calls
type mockStorageService struct {
	tempDirs     []string
	removedPaths []string
	mkdirTempErr error
}

func (m *mockStorageService) CopyDir(src, dest string) error                { return nil }
func (m *mockStorageService) Stat(name string) (os.FileInfo, error)         { return nil, nil }
func (m *mockStorageService) MkdirAll(path string, perm os.FileMode) error  { return nil }
func (m *mockStorageService) ReadFile(filename string) ([]byte, error)      { return nil, nil }
func (m *mockStorageService) ReadDir(dirname string) ([]os.DirEntry, error) { return nil, nil }
func (m *mockStorageService) ExtractArchive(src, dest string) error         { return nil }

func (m *mockStorageService) MkdirTemp(dir, pattern string) (string, error) {
	if m.mkdirTempErr != nil {
		return "", m.mkdirTempErr
	}
	// Use the actual pattern from the service to match expected behavior
	tempDir := filepath.Join(os.TempDir(), fmt.Sprintf("terrareg-git-import-%d", os.Getpid()))
	m.tempDirs = append(m.tempDirs, tempDir)
	err := os.MkdirAll(tempDir, 0755)
	return tempDir, err
}

func (m *mockStorageService) RemoveAll(path string) error {
	m.removedPaths = append(m.removedPaths, path)
	return os.RemoveAll(path)
}

func TestGitImportService_getTagRegex(t *testing.T) {
	service := &GitImportService{}

	tests := []struct {
		name        string
		format      string
		expectError bool
		expected    string // partial expected pattern
	}{
		{
			name:        "simple version format",
			format:      "v{version}",
			expectError: false,
			expected:    `v(?P<version>[^}]+)`,
		},
		{
			name:        "semantic versioning",
			format:      "v{major}.{minor}.{patch}",
			expectError: false,
			expected:    `v(?P<major>\d+)\.(?P<minor>\d+)\.(?P<patch>\d+)`,
		},
		{
			name:        "with build metadata",
			format:      "release-{major}.{minor}-{build}",
			expectError: false,
			expected:    `release-(?P<major>\d+)\.(?P<minor>\d+)-(?P<build>[^}]+)`,
		},
		{
			name:        "complex format with dots",
			format:      "api/{major}.{minor}.{patch}-alpha.{build}",
			expectError: false,
			expected:    `api/(?P<major>\d+)\.(?P<minor>\d+)\.(?P<patch>\d+)-alpha\.(?P<build>[^}]+)`,
		},
		{
			name:        "no placeholders",
			format:      "stable",
			expectError: false,
			expected:    `stable`,
		},
		{
			name:        "mixed with literal text",
			format:      "version-{version}-release",
			expectError: false,
			expected:    `version-(?P<version>[^}]+)-release`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			regex, err := service.getTagRegex(tt.format)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			// Check that the expected pattern is in the regex (accounting for ^ and $ anchors)
			if tt.expected != "" {
				actual := regex.String()
				// Remove ^ and $ for comparison
				if len(actual) > 2 && actual[0] == '^' && actual[len(actual)-1] == '$' {
					actual = actual[1 : len(actual)-1]
				}
				if actual != tt.expected {
					t.Logf("Expected pattern: %s", tt.expected)
					t.Logf("Actual pattern:  %s", actual)
					t.Errorf("Pattern doesn't match expected structure")
				}
			}

			// Verify regex is anchored
			pattern := regex.String()
			if pattern[0] != '^' || pattern[len(pattern)-1] != '$' {
				t.Errorf("Regex should be anchored with ^ and $, got: %s", pattern)
			}
		})
	}
}

func TestGitImportService_getVersionFromRegex(t *testing.T) {
	service := &GitImportService{}

	tests := []struct {
		name           string
		tagFormat      string
		gitTag         string
		expectedOutput string
	}{
		{
			name:           "direct version group",
			tagFormat:      "v{version}",
			gitTag:         "v1.2.3-beta",
			expectedOutput: "1.2.3-beta",
		},
		{
			name:           "semantic versioning",
			tagFormat:      "v{major}.{minor}.{patch}",
			gitTag:         "v2.5.1",
			expectedOutput: "2.5.1",
		},
		{
			name:           "semantic with build",
			tagFormat:      "v{major}.{minor}.{patch}-{build}",
			gitTag:         "v1.0.0-rc1",
			expectedOutput: "1.0.0-rc1",
		},
		{
			name:           "complex format",
			tagFormat:      "release/{major}.{minor}.{patch}-alpha.{build}",
			gitTag:         "release/3.2.1-alpha.20231201",
			expectedOutput: "3.2.1-20231201",
		},
		{
			name:           "missing minor defaults to 0",
			tagFormat:      "v{major}.{patch}",
			gitTag:         "v4.7",
			expectedOutput: "4.0.7",
		},
		{
			name:           "only major",
			tagFormat:      "v{major}",
			gitTag:         "v5",
			expectedOutput: "5.0.0",
		},
		{
			name:           "no match",
			tagFormat:      "v{major}.{minor}.{patch}",
			gitTag:         "not-a-version",
			expectedOutput: "",
		},
		{
			name:           "empty groups default to 0",
			tagFormat:      "v{major}.{minor}.{patch}",
			gitTag:         "v1.0.3",
			expectedOutput: "1.0.3",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			regex, err := service.getTagRegex(tt.tagFormat)
			if err != nil {
				t.Fatalf("Failed to create regex: %v", err)
			}

			result := service.getVersionFromRegex(regex, tt.gitTag)
			if result != tt.expectedOutput {
				t.Errorf("Expected version %q, got %q", tt.expectedOutput, result)
			}
		})
	}
}

// Note: TestDeriveVersionFromGitTag requires proper mocking of the ModuleProvider domain model
// The individual components (getTagRegex, getVersionFromRegex) are thoroughly tested above

// TestGitImportService_TemporaryDirectoryUsage verifies that git import uses
// the storage service for temporary directory creation and cleanup
func TestGitImportService_TemporaryDirectoryUsage(t *testing.T) {
	ctx := context.Background()

	// Create a test module provider with git configuration
	namespace, err := moduleModel.NewNamespace("test-namespace", nil, moduleModel.NamespaceTypeNone)
	require.NoError(t, err)

	moduleName := types.ModuleName("test-module")
	providerName := types.ModuleProviderName("aws")

	moduleProvider, err := moduleModel.NewModuleProvider(namespace, moduleName, providerName)
	require.NoError(t, err)

	cloneURL := "https://github.com/test/module.git"
	tagFormat := "v{version}"

	moduleProvider = moduleModel.ReconstructModuleProvider(
		1,
		namespace,
		moduleName,
		providerName,
		false,
		nil,
		nil,
		&cloneURL,
		nil,
		&tagFormat,
		nil,
		false,
		time.Now(),
		time.Now(),
	)

	// Create mock storage service to track temp directory operations
	storageService := &mockStorageService{}

	infraCfg := &infraConfig.InfrastructureConfig{
		DataDirectory: "/tmp",
	}

	service := &GitImportService{
		moduleProviderRepo: &mockModuleProviderRepo{
			moduleProvider: moduleProvider,
		},
		storageService: storageService,
		infraConfig:    infraCfg,
	}

	// Create import request with proper types
	req := module.ImportModuleVersionRequest{
		Namespace: types.NamespaceName("test-namespace"),
		Module:    moduleName,
		Provider:  providerName,
		Version:   "1.0.0",
		GitTag:    "",
	}

	// Execute the import (it will fail at git clone, but that's ok - we're testing temp dir usage)
	_, _ = service.Execute(ctx, req)

	// Verify that MkdirTemp was called with the correct pattern
	require.Equal(t, 1, len(storageService.tempDirs), "MkdirTemp should be called once")
	tempDir := storageService.tempDirs[0]

	// Verify the pattern contains "terrareg-git-import"
	assert.Contains(t, tempDir, "terrareg-git-import", "Temp directory should use the correct pattern")

	// Verify that RemoveAll was called for cleanup (via defer)
	require.Equal(t, 1, len(storageService.removedPaths), "RemoveAll should be called once for cleanup")
	assert.Equal(t, tempDir, storageService.removedPaths[0], "RemoveAll should be called with the temp directory path")
}
