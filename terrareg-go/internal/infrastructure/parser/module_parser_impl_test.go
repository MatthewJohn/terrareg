package parser

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	configModel "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/config/model"
	sharedService "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/shared/service"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/logging"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockStorageService is a mock implementation of StorageService for testing
type mockStorageService struct {
	files     map[string][]byte
	dirs      map[string]bool
	statError map[string]error
	readError map[string]error
}

func newMockStorageService() *mockStorageService {
	return &mockStorageService{
		files:     make(map[string][]byte),
		dirs:      make(map[string]bool),
		statError: make(map[string]error),
		readError: make(map[string]error),
	}
}

func (m *mockStorageService) CopyDir(src, dest string) error {
	return nil
}

func (m *mockStorageService) MkdirTemp(dir, pattern string) (string, error) {
	return "", nil
}

func (m *mockStorageService) RemoveAll(path string) error {
	delete(m.files, path)
	delete(m.dirs, path)
	return nil
}

func (m *mockStorageService) Stat(name string) (os.FileInfo, error) {
	if err, ok := m.statError[name]; ok {
		return nil, err
	}
	if content, ok := m.files[name]; ok {
		return &mockFileInfo{isDir: false, size: len(content)}, nil
	}
	if ok := m.dirs[name]; ok {
		return &mockFileInfo{isDir: true, size: 0}, nil
	}
	// Fall back to real file system for directories created during tests
	if info, err := os.Stat(name); err == nil {
		return info, nil
	}
	return nil, os.ErrNotExist
}

func (m *mockStorageService) MkdirAll(path string, perm os.FileMode) error {
	m.dirs[path] = true
	return nil
}

func (m *mockStorageService) ReadFile(filename string) ([]byte, error) {
	if err, ok := m.readError[filename]; ok {
		return nil, err
	}
	if content, ok := m.files[filename]; ok {
		return content, nil
	}
	return nil, os.ErrNotExist
}

func (m *mockStorageService) ReadDir(dirname string) ([]os.DirEntry, error) {
	var entries []os.DirEntry
	// Add entries from mock data
	for path := range m.files {
		if filepath.Dir(path) == dirname {
			entries = append(entries, &mockDirEntry{name: filepath.Base(path), isDir: false})
		}
	}
	for path := range m.dirs {
		if filepath.Dir(path) == dirname && path != dirname {
			entries = append(entries, &mockDirEntry{name: filepath.Base(path), isDir: true})
		}
	}
	// Fall back to real file system for directories created during tests
	if realEntries, err := os.ReadDir(dirname); err == nil {
		// Add real entries that aren't already in our mock list
		existing := make(map[string]bool)
		for _, e := range entries {
			existing[e.Name()] = true
		}
		for _, e := range realEntries {
			if !existing[e.Name()] {
				entries = append(entries, e)
			}
		}
	}
	return entries, nil
}

func (m *mockStorageService) ExtractArchive(src, dest string) error {
	return nil
}

func (m *mockStorageService) addFile(path, content string) {
	m.files[path] = []byte(content)
}

func (m *mockStorageService) addDir(path string) {
	m.dirs[path] = true
}

type mockFileInfo struct {
	isDir bool
	size  int
}

func (m *mockFileInfo) Name() string       { return "" }
func (m *mockFileInfo) Size() int64        { return int64(m.size) }
func (m *mockFileInfo) Mode() os.FileMode  { return 0 }
func (m *mockFileInfo) ModTime() time.Time { return time.Now() }
func (m *mockFileInfo) Sys() interface{}   { return nil }
func (m *mockFileInfo) IsDir() bool        { return m.isDir }

type mockDirEntry struct {
	name  string
	isDir bool
}

func (m *mockDirEntry) Name() string               { return m.name }
func (m *mockDirEntry) IsDir() bool                { return m.isDir }
func (m *mockDirEntry) Type() os.FileMode          { return 0 }
func (m *mockDirEntry) Info() (os.FileInfo, error) { return nil, nil }

// mockSystemCommandService is a mock implementation for testing
type mockSystemCommandService struct {
	mockOutputs map[string]*sharedService.CommandResult
	mockErrors  map[string]error
}

func newMockSystemCommandService() *mockSystemCommandService {
	return &mockSystemCommandService{
		mockOutputs: make(map[string]*sharedService.CommandResult),
		mockErrors:  make(map[string]error),
	}
}

func (m *mockSystemCommandService) Execute(ctx context.Context, cmd *sharedService.Command) (*sharedService.CommandResult, error) {
	key := cmd.Name + " " + cmd.Args[0]
	if err, ok := m.mockErrors[key]; ok {
		return nil, err
	}
	if result, ok := m.mockOutputs[key]; ok {
		return result, nil
	}
	return &sharedService.CommandResult{Stdout: "", Stderr: "", ExitCode: 0}, nil
}

func (m *mockSystemCommandService) ExecuteWithInput(ctx context.Context, cmd *sharedService.Command, input string) (*sharedService.CommandResult, error) {
	return m.Execute(ctx, cmd)
}

func (m *mockSystemCommandService) setMockOutput(command string, args []string, result *sharedService.CommandResult) {
	key := command + " " + args[0]
	m.mockOutputs[key] = result
}

// mockLogger is a mock implementation for testing
type mockLogger struct{}

func (m *mockLogger) Debug() logging.Event                            { return &mockEvent{} }
func (m *mockLogger) Info() logging.Event                             { return &mockEvent{} }
func (m *mockLogger) Warn() logging.Event                             { return &mockEvent{} }
func (m *mockLogger) Error() logging.Event                            { return &mockEvent{} }
func (m *mockLogger) With() logging.Logger                            { return m }
func (m *mockLogger) WithContext(ctx context.Context) context.Context { return ctx }

type mockEvent struct{}

func (m *mockEvent) Str(key, val string) logging.Event             { return m }
func (m *mockEvent) Strs(key string, vals []string) logging.Event  { return m }
func (m *mockEvent) Int(key string, i int) logging.Event           { return m }
func (m *mockEvent) Int64(key string, i int64) logging.Event       { return m }
func (m *mockEvent) Bool(key string, b bool) logging.Event         { return m }
func (m *mockEvent) Dur(key string, d time.Duration) logging.Event { return m }
func (m *mockEvent) Time(key string, t time.Time) logging.Event    { return m }
func (m *mockEvent) Err(err error) logging.Event                   { return m }
func (m *mockEvent) Msg(msg string)                                {}

func TestNewModuleParserImpl(t *testing.T) {
	storage := newMockStorageService()
	config := &configModel.DomainConfig{}
	cmdService := newMockSystemCommandService()
	logger := &mockLogger{}

	parser := NewModuleParserImpl(storage, config, cmdService, logger)

	assert.NotNil(t, parser)
	assert.Equal(t, storage, parser.storageService)
	assert.Equal(t, config, parser.config)
	assert.Equal(t, cmdService, parser.commandService)
	assert.Equal(t, logger, parser.logger)
}

func TestParseModule(t *testing.T) {
	t.Run("successful parse with terraform-docs", func(t *testing.T) {
		storage := newMockStorageService()
		storage.addDir("/test/module")
		storage.addFile("/test/module/README.md", "# Test Module\n\nThis is a test module with a description that is long enough to pass validation.")

		config := &configModel.DomainConfig{}
		cmdService := newMockSystemCommandService()
		logger := &mockLogger{}

		// Mock terraform-docs output
		tfDocsJSON := `{
			"header": "Test Module Header",
			"inputs": [
				{"name": "var1", "type": "string", "description": "First variable", "required": true}
			],
			"outputs": [
				{"name": "out1", "description": "First output"}
			],
			"providers": [
				{"name": "aws", "version": ">= 4.0"}
			],
			"resources": [
				{"type": "aws_instance", "name": "test"}
			],
			"modules": [],
			"requirements": []
		}`
		cmdService.setMockOutput("terraform-docs", []string{"json", "/test/module"}, &sharedService.CommandResult{
			Stdout:   tfDocsJSON,
			Stderr:   "",
			ExitCode: 0,
		})

		parser := NewModuleParserImpl(storage, config, cmdService, logger)
		result, err := parser.ParseModule("/test/module")

		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.NotEmpty(t, result.ReadmeContent)
		assert.NotEmpty(t, result.Description)
		assert.Len(t, result.Variables, 1)
		assert.Len(t, result.Outputs, 1)
		assert.Len(t, result.ProviderVersions, 1)
		assert.Len(t, result.Resources, 1)
	})

	t.Run("parse with terraform-docs failure", func(t *testing.T) {
		storage := newMockStorageService()
		storage.addDir("/test/module")
		storage.addFile("/test/module/README.md", "# Test\n\nDescription that is long enough.")

		config := &configModel.DomainConfig{}
		cmdService := newMockSystemCommandService()
		logger := &mockLogger{}

		// Mock terraform-docs failure
		cmdService.setMockOutput("terraform-docs", []string{"json", "/test/module"}, &sharedService.CommandResult{
			Stdout:   "",
			Stderr:   "terraform-docs not found",
			ExitCode: 1,
		})

		parser := NewModuleParserImpl(storage, config, cmdService, logger)
		result, err := parser.ParseModule("/test/module")

		// Should not fail, just log warning
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.NotEmpty(t, result.ReadmeContent)
		assert.Empty(t, result.Variables)
	})

	t.Run("parse with missing README", func(t *testing.T) {
		storage := newMockStorageService()
		storage.addDir("/test/module")

		config := &configModel.DomainConfig{}
		cmdService := newMockSystemCommandService()
		logger := &mockLogger{}

		cmdService.setMockOutput("terraform-docs", []string{"json", "/test/module"}, &sharedService.CommandResult{
			Stdout:   `{"inputs": [], "outputs": [], "providers": [], "resources": [], "modules": [], "requirements": []}`,
			Stderr:   "",
			ExitCode: 0,
		})

		parser := NewModuleParserImpl(storage, config, cmdService, logger)
		result, err := parser.ParseModule("/test/module")

		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Empty(t, result.ReadmeContent)
		assert.Empty(t, result.Description)
	})

	t.Run("parse removes terraform-docs config files", func(t *testing.T) {
		storage := newMockStorageService()
		storage.addDir("/test/module")
		storage.addFile("/test/module/.terraform-docs.yml", "{}")

		config := &configModel.DomainConfig{}
		cmdService := newMockSystemCommandService()
		logger := &mockLogger{}

		cmdService.setMockOutput("terraform-docs", []string{"json", "/test/module"}, &sharedService.CommandResult{
			Stdout:   `{"inputs": [], "outputs": [], "providers": [], "resources": [], "modules": [], "requirements": []}`,
			Stderr:   "",
			ExitCode: 0,
		})

		parser := NewModuleParserImpl(storage, config, cmdService, logger)
		_, err := parser.ParseModule("/test/module")

		require.NoError(t, err)
		// Verify config file was removed
		_, err = storage.Stat("/test/module/.terraform-docs.yml")
		assert.Error(t, err) // File should not exist
	})

	t.Run("parse with terrareg.json metadata", func(t *testing.T) {
		storage := newMockStorageService()
		storage.addDir("/test/module")
		storage.addFile("/test/module/README.md", "# Test Module\n\nThis is a test module with a description that is long enough to pass validation.")

		// Add metadata file
		metadata := `{
			"owner": "test-owner-from-metadata",
			"description": "Description from terrareg.json"
		}`
		storage.addFile("/test/module/terrareg.json", metadata)

		config := &configModel.DomainConfig{}
		cmdService := newMockSystemCommandService()
		logger := &mockLogger{}

		cmdService.setMockOutput("terraform-docs", []string{"json", "/test/module"}, &sharedService.CommandResult{
			Stdout:   `{"inputs": [], "outputs": [], "providers": [], "resources": [], "modules": [], "requirements": []}`,
			Stderr:   "",
			ExitCode: 0,
		})

		parser := NewModuleParserImpl(storage, config, cmdService, logger)
		result, err := parser.ParseModule("/test/module")

		require.NoError(t, err)
		assert.NotNil(t, result)
		// Metadata description should take precedence over README extraction
		assert.Equal(t, "Description from terrareg.json", result.Description)
		// Metadata owner should be populated
		assert.Equal(t, "test-owner-from-metadata", result.Owner)
		// Metadata file should be removed
		_, err = storage.Stat("/test/module/terrareg.json")
		assert.Error(t, err) // File should not exist
	})
}

func TestExtractDescriptionFromReadme(t *testing.T) {
	config := &configModel.DomainConfig{}
	cmdService := newMockSystemCommandService()
	logger := &mockLogger{}

	parser := NewModuleParserImpl(newMockStorageService(), config, cmdService, logger)

	tests := []struct {
		name     string
		readme   string
		expected string
	}{
		{
			name:     "empty readme",
			readme:   "",
			expected: "",
		},
		{
			name:     "readme with only headers",
			readme:   "# Header\n\n## Subheader\n",
			expected: "",
		},
		{
			name:     "readme with URL",
			readme:   "Check out https://example.com for more info",
			expected: "",
		},
		{
			name:     "readme with email",
			readme:   "Contact test@example.com for support",
			expected: "",
		},
		{
			name:     "readme with too few characters",
			readme:   "Short text",
			expected: "",
		},
		{
			name:     "readme with too few words",
			readme:   "This is short",
			expected: "",
		},
		{
			name:     "valid description with single sentence",
			readme:   "This is a much longer description that contains sufficient characters and words for successful validation",
			expected: "This is a much longer description that contains sufficient characters and words for successful validation",
		},
		{
			name:     "valid description with multiple sentences truncated",
			readme:   "This is the first sentence that has enough content. This is the second sentence that would make it too long if combined together.",
			expected: "This is the first sentence that has enough content",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parser.extractDescriptionFromReadme(tt.readme)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestDetectSubmodules_DefaultDirectory tests that submodules are detected from the default "modules" directory
func TestDetectSubmodules_DefaultDirectory(t *testing.T) {
	tempDir := t.TempDir()
	modulesDir := filepath.Join(tempDir, "modules")
	err := os.MkdirAll(modulesDir, 0755)
	require.NoError(t, err)

	submodule1 := filepath.Join(modulesDir, "database")
	err = os.MkdirAll(submodule1, 0755)
	require.NoError(t, err)
	err = os.WriteFile(filepath.Join(submodule1, "main.tf"), []byte("resource \"aws_db_instance\" \"example\" {}"), 0644)
	require.NoError(t, err)

	submodule2 := filepath.Join(modulesDir, "network")
	err = os.MkdirAll(submodule2, 0755)
	require.NoError(t, err)
	err = os.WriteFile(filepath.Join(submodule2, "main.tf"), []byte("resource \"aws_vpc\" \"example\" {}"), 0644)
	require.NoError(t, err)

	otherDir := filepath.Join(tempDir, "other_stuff")
	err = os.MkdirAll(otherDir, 0755)
	require.NoError(t, err)
	err = os.WriteFile(filepath.Join(otherDir, "main.tf"), []byte("resource \"aws_s3_bucket\" \"example\" {}"), 0644)
	require.NoError(t, err)

	config := &configModel.DomainConfig{
		ModulesDirectory:  "modules",
		ExamplesDirectory: "examples",
	}
	storage := newMockStorageService()
	cmdService := newMockSystemCommandService()
	logger := &mockLogger{}
	parser := NewModuleParserImpl(storage, config, cmdService, logger)

	submodules, err := parser.DetectSubmodules(tempDir)
	require.NoError(t, err)
	assert.Len(t, submodules, 2)
	assert.Contains(t, submodules, filepath.Join("modules", "database"))
	assert.Contains(t, submodules, filepath.Join("modules", "network"))
	assert.NotContains(t, submodules, "other_stuff")
}

// TestDetectSubmodules_CustomDirectory tests that submodules are detected from a custom MODULES_DIRECTORY
func TestDetectSubmodules_CustomDirectory(t *testing.T) {
	tempDir := t.TempDir()
	customModulesDir := filepath.Join(tempDir, "subcomponents")
	err := os.MkdirAll(customModulesDir, 0755)
	require.NoError(t, err)

	submodule1 := filepath.Join(customModulesDir, "auth")
	err = os.MkdirAll(submodule1, 0755)
	require.NoError(t, err)
	err = os.WriteFile(filepath.Join(submodule1, "main.tf"), []byte("resource \"aws_iam_role\" \"example\" {}"), 0644)
	require.NoError(t, err)

	defaultModulesDir := filepath.Join(tempDir, "modules")
	err = os.MkdirAll(defaultModulesDir, 0755)
	require.NoError(t, err)
	err = os.WriteFile(filepath.Join(defaultModulesDir, "main.tf"), []byte("resource \"aws_s3_bucket\" \"default\" {}"), 0644)
	require.NoError(t, err)

	config := &configModel.DomainConfig{
		ModulesDirectory:  "subcomponents",
		ExamplesDirectory: "examples",
	}
	storage := newMockStorageService()
	cmdService := newMockSystemCommandService()
	logger := &mockLogger{}
	parser := NewModuleParserImpl(storage, config, cmdService, logger)

	submodules, err := parser.DetectSubmodules(tempDir)
	require.NoError(t, err)
	assert.Len(t, submodules, 1)
	assert.Contains(t, submodules, filepath.Join("subcomponents", "auth"))
	assert.NotContains(t, submodules, filepath.Join("modules", "main"))
}

// TestDetectSubmodules_NonExistentDirectory tests that non-existent MODULES_DIRECTORY returns empty list
func TestDetectSubmodules_NonExistentDirectory(t *testing.T) {
	tempDir := t.TempDir()
	config := &configModel.DomainConfig{
		ModulesDirectory:  "non_existent_modules",
		ExamplesDirectory: "examples",
	}
	storage := newMockStorageService()
	cmdService := newMockSystemCommandService()
	logger := &mockLogger{}
	parser := NewModuleParserImpl(storage, config, cmdService, logger)

	submodules, err := parser.DetectSubmodules(tempDir)
	require.NoError(t, err)
	assert.Empty(t, submodules)
}

func TestParseTerraregMetadata(t *testing.T) {
	config := &configModel.DomainConfig{}
	cmdService := newMockSystemCommandService()
	logger := &mockLogger{}

	t.Run("valid terrareg.json", func(t *testing.T) {
		storage := newMockStorageService()
		storage.addDir("/test/module")

		metadata := `{
			"owner": "test-owner",
			"description": "Test description",
			"repo_clone_url": "https://github.com/test/repo.git",
			"repo_browse_url": "https://github.com/test/repo"
		}`
		storage.addFile("/test/module/terrareg.json", metadata)

		parser := NewModuleParserImpl(storage, config, cmdService, logger)
		result, err := parser.parseTerraregMetadata("/test/module")

		require.NoError(t, err)
		assert.NotNil(t, result)
		if result.Owner != nil {
			assert.Equal(t, "test-owner", *result.Owner)
		}
		if result.Description != nil {
			assert.Equal(t, "Test description", *result.Description)
		}
		if result.RepoCloneURL != nil {
			assert.Equal(t, "https://github.com/test/repo.git", *result.RepoCloneURL)
		}
	})

	t.Run("valid .terrareg.json", func(t *testing.T) {
		storage := newMockStorageService()
		storage.addDir("/test/module")

		metadata := `{"owner": "test-owner"}`
		storage.addFile("/test/module/.terrareg.json", metadata)

		parser := NewModuleParserImpl(storage, config, cmdService, logger)
		result, err := parser.parseTerraregMetadata("/test/module")

		require.NoError(t, err)
		assert.NotNil(t, result)
		if result.Owner != nil {
			assert.Equal(t, "test-owner", *result.Owner)
		}
	})

	t.Run("no metadata file", func(t *testing.T) {
		storage := newMockStorageService()
		storage.addDir("/test/module")

		parser := NewModuleParserImpl(storage, config, cmdService, logger)
		result, err := parser.parseTerraregMetadata("/test/module")

		require.NoError(t, err)
		assert.Nil(t, result) // No error, no metadata
	})

	t.Run("invalid JSON", func(t *testing.T) {
		storage := newMockStorageService()
		storage.addDir("/test/module")
		storage.addFile("/test/module/terrareg.json", "invalid json")

		parser := NewModuleParserImpl(storage, config, cmdService, logger)
		result, err := parser.parseTerraregMetadata("/test/module")

		require.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "failed to parse metadata JSON")
	})

	t.Run("terrareg.json takes priority over .terrareg.json", func(t *testing.T) {
		storage := newMockStorageService()
		storage.addDir("/test/module")
		storage.addFile("/test/module/terrareg.json", `{"owner": "first"}`)
		storage.addFile("/test/module/.terrareg.json", `{"owner": "second"}`)

		parser := NewModuleParserImpl(storage, config, cmdService, logger)
		result, err := parser.parseTerraregMetadata("/test/module")

		require.NoError(t, err)
		assert.NotNil(t, result)
		if result.Owner != nil {
			assert.Equal(t, "first", *result.Owner)
		}
	})
}

func TestParseSubmodule(t *testing.T) {
	storage := newMockStorageService()
	storage.addDir("/test/module/submodule")
	storage.addFile("/test/module/submodule/README.md", "# Submodule\n\nDescription here.")

	config := &configModel.DomainConfig{}
	cmdService := newMockSystemCommandService()
	logger := &mockLogger{}

	cmdService.setMockOutput("terraform-docs", []string{"json", "/test/module/submodule"}, &sharedService.CommandResult{
		Stdout:   `{"inputs": [], "outputs": [], "providers": [], "resources": [], "modules": [], "requirements": []}`,
		Stderr:   "",
		ExitCode: 0,
	})

	parser := NewModuleParserImpl(storage, config, cmdService, logger)
	result, err := parser.ParseSubmodule("/test/module/submodule")

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.NotEmpty(t, result.ReadmeContent)
}

func TestParseExample(t *testing.T) {
	storage := newMockStorageService()
	storage.addDir("/test/module/examples/test")
	storage.addFile("/test/module/examples/test/README.md", "# Example\n\nDescription here.")

	config := &configModel.DomainConfig{}
	cmdService := newMockSystemCommandService()
	logger := &mockLogger{}

	cmdService.setMockOutput("terraform-docs", []string{"json", "/test/module/examples/test"}, &sharedService.CommandResult{
		Stdout:   `{"inputs": [], "outputs": [], "providers": [], "resources": [], "modules": [], "requirements": []}`,
		Stderr:   "",
		ExitCode: 0,
	})

	parser := NewModuleParserImpl(storage, config, cmdService, logger)

	infracostJSON := []byte(`{"total_monthly_cost": 10.5}`)
	result, err := parser.ParseExample("/test/module/examples/test", infracostJSON)

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.NotNil(t, result.ParseResult)
	assert.Equal(t, infracostJSON, result.InfracostJSON)
}

func TestExtractExampleFiles(t *testing.T) {
	config := &configModel.DomainConfig{}
	cmdService := newMockSystemCommandService()
	logger := &mockLogger{}

	t.Run("extracts all terraform-related files", func(t *testing.T) {
		storage := newMockStorageService()
		storage.addDir("/test/module/examples/test")
		storage.addFile("/test/module/examples/test/main.tf", "resource {}")
		storage.addFile("/test/module/examples/test/variables.tfvars", "var = {}")
		storage.addFile("/test/module/examples/test/setup.sh", "#!/bin/bash")
		storage.addFile("/test/module/examples/test/metadata.json", "{}")
		storage.addFile("/test/module/examples/test/README.md", "# Example")
		storage.addFile("/test/module/examples/test/ignore.txt", "ignore me")

		parser := NewModuleParserImpl(storage, config, cmdService, logger)
		files, err := parser.ExtractExampleFiles("/test/module/examples/test")

		require.NoError(t, err)
		assert.Len(t, files, 4) // main.tf, variables.tfvars, setup.sh, metadata.json

		// Check that correct files were extracted
		filenames := make([]string, len(files))
		for i, f := range files {
			filenames[i] = f.Path()
		}
		assert.Contains(t, filenames, "main.tf")
		assert.Contains(t, filenames, "variables.tfvars")
		assert.Contains(t, filenames, "setup.sh")
		assert.Contains(t, filenames, "metadata.json")
		assert.NotContains(t, filenames, "README.md")
		assert.NotContains(t, filenames, "ignore.txt")
	})

	t.Run("handles empty directory", func(t *testing.T) {
		storage := newMockStorageService()
		storage.addDir("/test/module/examples/test")

		parser := NewModuleParserImpl(storage, config, cmdService, logger)
		files, err := parser.ExtractExampleFiles("/test/module/examples/test")

		require.NoError(t, err)
		assert.Empty(t, files)
	})

	t.Run("handles subdirectories", func(t *testing.T) {
		storage := newMockStorageService()
		storage.addDir("/test/module/examples/test")
		storage.addDir("/test/module/examples/test/subdir")
		storage.addFile("/test/module/examples/test/main.tf", "resource {}")

		parser := NewModuleParserImpl(storage, config, cmdService, logger)
		files, err := parser.ExtractExampleFiles("/test/module/examples/test")

		require.NoError(t, err)
		assert.Len(t, files, 1) // Only main.tf, not subdirectory
	})
}
