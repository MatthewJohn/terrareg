package service

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	configmodel "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/config/model"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/git/service"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/module/model"
	moduleRepository "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/module/repository"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/shared/types"
	infraConfig "github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/config"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockModuleProviderRepo is a mock implementation of ModuleProviderRepository
type mockModuleProviderRepo struct {
	moduleProvider *model.ModuleProvider
	findErr        error
}

func (m *mockModuleProviderRepo) FindByNamespaceModuleProvider(
	ctx context.Context, namespace types.NamespaceName, moduleName types.ModuleName, provider types.ModuleProviderName,
) (*model.ModuleProvider, error) {
	if m.findErr != nil {
		return nil, m.findErr
	}
	return m.moduleProvider, nil
}

// Implement other required methods (no-ops for tests)
func (m *mockModuleProviderRepo) FindByID(ctx context.Context, id int) (*model.ModuleProvider, error) {
	return m.moduleProvider, nil
}
func (m *mockModuleProviderRepo) FindByNamespace(ctx context.Context, namespace types.NamespaceName) ([]*model.ModuleProvider, error) {
	return nil, nil
}
func (m *mockModuleProviderRepo) Save(ctx context.Context, mp *model.ModuleProvider) error {
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

// mockGitClient is a mock implementation of GitClient
type mockGitClient struct {
	cloneErr      error
	checkoutErr   error
	commitSHA     string
	clonePath     string
	checkedOutTag string
}

func (m *mockGitClient) Clone(ctx context.Context, url, dest string) error {
	return m.CloneWithOptions(ctx, url, dest, service.CloneOptions{})
}

func (m *mockGitClient) CloneWithOptions(ctx context.Context, url, dest string, opts service.CloneOptions) error {
	if m.cloneErr != nil {
		return m.cloneErr
	}
	m.clonePath = dest
	// Create the directory to simulate successful clone
	return os.MkdirAll(dest, 0755)
}

func (m *mockGitClient) Checkout(ctx context.Context, repoDir, tag string) error {
	if m.checkoutErr != nil {
		return m.checkoutErr
	}
	m.checkedOutTag = tag
	return nil
}

func (m *mockGitClient) GetCommitSHA(ctx context.Context, repoDir string) (string, error) {
	if m.commitSHA == "" {
		return "abc123def456", nil
	}
	return m.commitSHA, nil
}

// mockStorageService is a mock implementation of StorageService
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
	tempDir := filepath.Join(os.TempDir(), fmt.Sprintf("test-terrareg-%d", time.Now().UnixNano()))
	m.tempDirs = append(m.tempDirs, tempDir)
	err := os.MkdirAll(tempDir, 0755)
	return tempDir, err
}

func (m *mockStorageService) RemoveAll(path string) error {
	m.removedPaths = append(m.removedPaths, path)
	return os.RemoveAll(path)
}

// mockArchiveProcessor is a mock implementation of ArchiveProcessor
type mockArchiveProcessor struct {
	detectType        ArchiveType
	extractErr        error
	failOnNonExistent bool
}

func (m *mockArchiveProcessor) ExtractArchive(ctx context.Context, archivePath, targetDir string, archiveType ArchiveType) error {
	if m.extractErr != nil {
		return m.extractErr
	}
	// Create a test file to simulate successful extraction
	testFile := filepath.Join(targetDir, "main.tf")
	return os.WriteFile(testFile, []byte("resource \"aws_s3_bucket\" \"example\" {}"), 0644)
}

func (m *mockArchiveProcessor) DetectArchiveType(archivePath string) (ArchiveType, error) {
	// Check if file exists (for testing non-existent files)
	if _, err := os.Stat(archivePath); os.IsNotExist(err) && m.failOnNonExistent {
		return 0, fmt.Errorf("archive file does not exist: %s", archivePath)
	}
	if m.detectType != 0 {
		return m.detectType, nil
	}
	// Default to ZIP
	return ArchiveTypeZIP, nil
}

func (m *mockArchiveProcessor) ValidateArchive(archivePath string) error {
	return nil
}

// Helper function to create a test module provider with git configuration
func createTestModuleProviderWithGit(t *testing.T) *model.ModuleProvider {
	namespace, err := model.NewNamespace("test-namespace", nil, model.NamespaceTypeNone)
	require.NoError(t, err)

	moduleProvider, err := model.NewModuleProvider(namespace, "test-module", "aws")
	require.NoError(t, err)

	// Use reflection to set private fields for testing
	cloneURLTemplate := "https://github.com/{namespace}/{module}.git"
	tagFormat := "v{version}"
	gitPath := "terraform"

	// Reconstruct with git config
	moduleProvider = model.ReconstructModuleProvider(
		1,
		namespace,
		"test-module",
		"aws",
		false,
		nil,
		nil,
		&cloneURLTemplate,
		nil,
		&tagFormat,
		&gitPath,
		false,
		time.Now(),
		time.Now(),
	)

	return moduleProvider
}

// Helper function to create a test ZIP archive
func createTestZIPArchive(t *testing.T) *bytes.Reader {
	buf := new(bytes.Buffer)
	zipWriter := zip.NewWriter(buf)

	// Add some test files
	files := map[string]string{
		"main.tf":             `resource "aws_s3_bucket" "example" {}`,
		"variables.tf":        `variable "bucket_name" { type = string }`,
		"outputs.tf":          `output "bucket_id" { value = aws_s3_bucket.example.id }`,
		"README.md":           "# Test Module",
		"modules/sub/main.tf": `resource "aws_instance" "example" {}`,
	}

	for path, content := range files {
		writer, err := zipWriter.Create(path)
		require.NoError(t, err)
		_, err = io.WriteString(writer, content)
		require.NoError(t, err)
	}

	require.NoError(t, zipWriter.Close())
	return bytes.NewReader(buf.Bytes())
}

// Helper function to create a test tar.gz archive
func createTestTarGzArchive(t *testing.T) *bytes.Reader {
	buf := new(bytes.Buffer)
	gzipWriter := gzip.NewWriter(buf)
	tarWriter := tar.NewWriter(gzipWriter)

	// Add some test files
	files := map[string]string{
		"main.tf":      `resource "aws_s3_bucket" "example" {}`,
		"variables.tf": `variable "bucket_name" { type = string }`,
	}

	for path, content := range files {
		header := &tar.Header{
			Name:    path,
			Size:    int64(len(content)),
			Mode:    0644,
			ModTime: time.Now(),
		}
		require.NoError(t, tarWriter.WriteHeader(header))
		_, err := tarWriter.Write([]byte(content))
		require.NoError(t, err)
	}

	require.NoError(t, tarWriter.Close())
	require.NoError(t, gzipWriter.Close())
	return bytes.NewReader(buf.Bytes())
}

// Helper to create a test service
func createTestSourcePreparationService(t *testing.T, moduleProvider *model.ModuleProvider) *SourcePreparationService {
	domainConfig := &configmodel.DomainConfig{
		ModuleVersionUseGitCommit: true,
	}

	infraConfig := &infraConfig.InfrastructureConfig{
		GitCloneTimeout: 300,
	}

	logger := zerolog.New(os.Stderr)

	moduleProviderRepo := &mockModuleProviderRepo{
		moduleProvider: moduleProvider,
	}

	gitClient := &mockGitClient{
		commitSHA: "abc123def456",
	}

	storageService := &mockStorageService{}

	archiveProcessor := &mockArchiveProcessor{
		detectType: 0,
	}

	return NewSourcePreparationService(
		moduleProviderRepo,
		gitClient,
		storageService,
		archiveProcessor,
		domainConfig,
		infraConfig,
		logger,
	)
}

// TestPrepareFromUpload_ValidZIP tests successful ZIP upload preparation
func TestPrepareFromUpload_ValidZIP(t *testing.T) {
	ctx := context.Background()
	service := createTestSourcePreparationService(t, nil)

	zipContent := createTestZIPArchive(t)
	zipBytes, _ := io.ReadAll(zipContent)

	req := PrepareFromUploadRequest{
		Namespace:  "test-namespace",
		Module:     "test-module",
		Provider:   "aws",
		Version:    "1.0.0",
		Source:     bytes.NewReader(zipBytes),
		SourceSize: int64(len(zipBytes)),
	}

	result, err := service.PrepareFromUpload(ctx, req)

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.NotEmpty(t, result.SourcePath)
	assert.Nil(t, result.CommitSHA, "No commit SHA for uploads")
	assert.EqualValues(t, SourceTypeUpload, result.SourceType)
	assert.NotNil(t, result.Cleanup, "Cleanup function must be provided")

	// Verify source directory exists and has files
	files, err := os.ReadDir(result.SourcePath)
	require.NoError(t, err)
	assert.NotEmpty(t, files, "Extracted files should exist")

	// Cleanup
	result.Cleanup()
	_, err = os.Stat(result.SourcePath)
	assert.True(t, os.IsNotExist(err), "Source should be cleaned up")
}

// TestPrepareFromUpload_TarGzFormat tests tar.gz upload preparation
func TestPrepareFromUpload_TarGzFormat(t *testing.T) {
	ctx := context.Background()
	service := createTestSourcePreparationService(t, nil)

	// Set archive processor to detect tar.gz
	service.archiveProcessor = &mockArchiveProcessor{
		detectType: ArchiveTypeTarGZ,
		extractErr: nil,
	}

	tarGzContent := createTestTarGzArchive(t)
	tarGzBytes, _ := io.ReadAll(tarGzContent)

	req := PrepareFromUploadRequest{
		Namespace:  "test-namespace",
		Module:     "test-module",
		Provider:   "aws",
		Version:    "1.0.0",
		Source:     bytes.NewReader(tarGzBytes),
		SourceSize: int64(len(tarGzBytes)),
	}

	result, err := service.PrepareFromUpload(ctx, req)

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.NotEmpty(t, result.SourcePath)
	assert.EqualValues(t, SourceTypeUpload, result.SourceType)

	// Cleanup
	result.Cleanup()
}

// TestPrepareFromUpload_InvalidArchive tests upload with invalid archive
func TestPrepareFromUpload_InvalidArchive(t *testing.T) {
	ctx := context.Background()
	service := createTestSourcePreparationService(t, nil)

	// Set archive processor to fail detection
	service.archiveProcessor = &mockArchiveProcessor{
		detectType: 0,
		extractErr: fmt.Errorf("invalid archive format"),
	}

	req := PrepareFromUploadRequest{
		Namespace:  "test-namespace",
		Module:     "test-module",
		Provider:   "aws",
		Version:    "1.0.0",
		Source:     strings.NewReader("not a valid archive"),
		SourceSize: 20,
	}

	result, err := service.PrepareFromUpload(ctx, req)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to extract archive")
}

// TestPrepareFromGit_ValidRepository tests git clone preparation
func TestPrepareFromGit_ValidRepository(t *testing.T) {
	ctx := context.Background()
	moduleProvider := createTestModuleProviderWithGit(t)
	service := createTestSourcePreparationService(t, moduleProvider)

	gitTag := "v1.0.0"
	req := PrepareFromGitRequest{
		Namespace: "test-namespace",
		Module:    "test-module",
		Provider:  "aws",
		Version:   "1.0.0",
		GitTag:    &gitTag,
	}

	result, err := service.PrepareFromGit(ctx, req)

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.NotEmpty(t, result.SourcePath)
	assert.NotNil(t, result.CommitSHA, "Should have commit SHA")
	assert.Equal(t, "abc123def456", *result.CommitSHA)
	assert.EqualValues(t, SourceTypeGit, result.SourceType)
	assert.NotNil(t, result.Cleanup)

	// Verify git client was called correctly
	mockGit := service.gitClient.(*mockGitClient)
	assert.NotEmpty(t, mockGit.clonePath, "Clone should have been called")
	// The tag format "v{version}" with input "v1.0.0" becomes "vv1.0.0"
	assert.Equal(t, "vv1.0.0", mockGit.checkedOutTag, "Tag should have been formatted and checked out")

	// Cleanup
	result.Cleanup()
}

// TestPrepareFromGit_ModuleProviderNotFound tests git prep with missing module provider
func TestPrepareFromGit_ModuleProviderNotFound(t *testing.T) {
	ctx := context.Background()
	service := createTestSourcePreparationService(t, nil)

	// Set repo to return nil
	service.moduleProviderRepo = &mockModuleProviderRepo{
		moduleProvider: nil,
		findErr:        nil,
	}

	gitTag := "v1.0.0"
	req := PrepareFromGitRequest{
		Namespace: "test-namespace",
		Module:    "test-module",
		Provider:  "aws",
		Version:   "1.0.0",
		GitTag:    &gitTag,
	}

	result, err := service.PrepareFromGit(ctx, req)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "module provider not found")
}

// TestPrepareFromGit_NotGitBased tests git prep with non-git module provider
func TestPrepareFromGit_NotGitBased(t *testing.T) {
	ctx := context.Background()

	namespace, err := model.NewNamespace("test-namespace", nil, model.NamespaceTypeNone)
	require.NoError(t, err)

	moduleProvider, err := model.NewModuleProvider(namespace, "test-module", "aws")
	require.NoError(t, err)

	service := createTestSourcePreparationService(t, moduleProvider)

	gitTag := "v1.0.0"
	req := PrepareFromGitRequest{
		Namespace: "test-namespace",
		Module:    "test-module",
		Provider:  "aws",
		Version:   "1.0.0",
		GitTag:    &gitTag,
	}

	result, err := service.PrepareFromGit(ctx, req)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "not git-based")
}

// TestPrepareFromGit_CloneFailure tests git clone failure handling
func TestPrepareFromGit_CloneFailure(t *testing.T) {
	ctx := context.Background()
	moduleProvider := createTestModuleProviderWithGit(t)
	service := createTestSourcePreparationService(t, moduleProvider)

	// Set git client to fail clone
	service.gitClient = &mockGitClient{
		cloneErr: fmt.Errorf("git clone failed: repository not found"),
	}

	gitTag := "v1.0.0"
	req := PrepareFromGitRequest{
		Namespace: "test-namespace",
		Module:    "test-module",
		Provider:  "aws",
		Version:   "1.0.0",
		GitTag:    &gitTag,
	}

	result, err := service.PrepareFromGit(ctx, req)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to clone repository")
}

// TestPrepareFromArchive_ValidArchive tests archive extraction preparation
func TestPrepareFromArchive_ValidArchive(t *testing.T) {
	ctx := context.Background()
	service := createTestSourcePreparationService(t, nil)

	// Create a temporary archive file
	tempDir := t.TempDir()
	archivePath := filepath.Join(tempDir, "module.zip")

	// Write test ZIP content
	zipContent := createTestZIPArchive(t)
	zipBytes, _ := io.ReadAll(zipContent)
	require.NoError(t, os.WriteFile(archivePath, zipBytes, 0644))

	req := PrepareFromArchiveRequest{
		Namespace:   "test-namespace",
		Module:      "test-module",
		Provider:    "aws",
		Version:     "1.0.0",
		ArchivePath: archivePath,
	}

	result, err := service.PrepareFromArchive(ctx, req)

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.NotEmpty(t, result.SourcePath)
	assert.Nil(t, result.CommitSHA, "No commit SHA for archives")
	assert.EqualValues(t, SourceTypeArchive, result.SourceType)
	assert.NotNil(t, result.Cleanup)

	// Verify source directory has files
	files, err := os.ReadDir(result.SourcePath)
	require.NoError(t, err)
	assert.NotEmpty(t, files)

	// Cleanup
	result.Cleanup()
	_, err = os.Stat(result.SourcePath)
	assert.True(t, os.IsNotExist(err))
}

// TestPrepareFromArchive_NonExistentArchive tests archive prep with non-existent file
func TestPrepareFromArchive_NonExistentArchive(t *testing.T) {
	ctx := context.Background()
	service := createTestSourcePreparationService(t, nil)

	// Set archive processor to fail on non-existent files
	service.archiveProcessor = &mockArchiveProcessor{
		detectType:        0,
		extractErr:        nil,
		failOnNonExistent: true,
	}

	req := PrepareFromArchiveRequest{
		Namespace:   "test-namespace",
		Module:      "test-module",
		Provider:    "aws",
		Version:     "1.0.0",
		ArchivePath: "/nonexistent/path/to/archive.zip",
	}

	result, err := service.PrepareFromArchive(ctx, req)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to detect archive type")
}

// TestPrepareFromArchive_CleanupOnError tests cleanup is performed on extraction error
func TestPrepareFromArchive_CleanupOnError(t *testing.T) {
	ctx := context.Background()
	service := createTestSourcePreparationService(t, nil)

	// Set archive processor to fail extraction
	service.archiveProcessor = &mockArchiveProcessor{
		detectType: 0,
		extractErr: fmt.Errorf("extraction failed"),
	}

	// Create a temporary archive file
	tempDir := t.TempDir()
	archivePath := filepath.Join(tempDir, "module.zip")
	zipContent := createTestZIPArchive(t)
	zipBytes, _ := io.ReadAll(zipContent)
	require.NoError(t, os.WriteFile(archivePath, zipBytes, 0644))

	req := PrepareFromArchiveRequest{
		Namespace:   "test-namespace",
		Module:      "test-module",
		Provider:    "aws",
		Version:     "1.0.0",
		ArchivePath: archivePath,
	}

	result, err := service.PrepareFromArchive(ctx, req)

	assert.Error(t, err)
	assert.Nil(t, result)

	// Verify temporary directory was cleaned up
	mockStorage := service.storageService.(*mockStorageService)
	assert.NotEmpty(t, mockStorage.removedPaths, "Temporary directory should be cleaned up on error")
}

// TestBuildCloneURL tests clone URL building from template
func TestBuildCloneURL(t *testing.T) {
	namespace, err := model.NewNamespace("my-org", nil, model.NamespaceTypeNone)
	require.NoError(t, err)

	moduleProvider := model.ReconstructModuleProvider(
		1,
		namespace,
		"my-module",
		"aws",
		false,
		nil,
		nil,
		nil,
		nil, // repoBrowseURLTemplate
		nil, // gitTagFormat
		nil, // gitPath
		false,
		time.Now(),
		time.Now(),
	)

	service := createTestSourcePreparationService(t, moduleProvider)

	cloneURLTemplate := "https://github.com/{namespace}/{module}.git"
	moduleProvider = model.ReconstructModuleProvider(
		1,
		namespace,
		"my-module",
		"aws",
		false,
		nil,
		nil,               // repoBaseURLTemplate
		&cloneURLTemplate, // repoCloneURLTemplate
		nil,               // repoBrowseURLTemplate
		nil,               // gitTagFormat
		nil,               // gitPath
		false,
		time.Now(),
		time.Now(),
	)

	req := PrepareFromGitRequest{
		Namespace: "my-org",
		Module:    "my-module",
		Provider:  "aws",
	}

	result := service.buildCloneURL(req, moduleProvider)

	assert.Equal(t, "https://github.com/my-org/my-module.git", result)
}

// TestFormatGitTag tests git tag formatting
func TestFormatGitTag(t *testing.T) {
	namespace, err := model.NewNamespace("my-org", nil, model.NamespaceTypeNone)
	require.NoError(t, err)

	tests := []struct {
		name        string
		tagFormat   string
		inputTag    string
		expectedTag string
	}{
		{
			name:        "Default format (no template)",
			tagFormat:   "",
			inputTag:    "1.0.0",
			expectedTag: "1.0.0",
		},
		{
			name:        "Version prefix format",
			tagFormat:   "v{version}",
			inputTag:    "1.0.0",
			expectedTag: "v1.0.0",
		},
		{
			name:        "Module and version format",
			tagFormat:   "{module}/v{version}",
			inputTag:    "1.0.0",
			expectedTag: "test-module/v1.0.0",
		},
		{
			name:        "Provider module version format",
			tagFormat:   "{provider}/{module}/v{version}",
			inputTag:    "1.0.0",
			expectedTag: "aws/test-module/v1.0.0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			moduleProvider := model.ReconstructModuleProvider(
				1,
				namespace,
				"test-module",
				"aws",
				false,
				nil,
				nil,
				nil,
				nil,
				&tt.tagFormat,
				nil,
				false,
				time.Now(),
				time.Now(),
			)

			service := createTestSourcePreparationService(t, moduleProvider)
			result := service.formatGitTag(tt.inputTag, moduleProvider)

			assert.Equal(t, tt.expectedTag, result)
		})
	}
}
