package module

import (
	"context"
	"io"
	"strings"
	"testing"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/module/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockSourcePreparationService is a mock for SourcePreparationService
type mockSourcePreparationService struct {
	preparedSource *service.PreparedSource
	prepareErr     error
	sourceType     service.SourceType
}

func (m *mockSourcePreparationService) PrepareFromUpload(
	ctx context.Context,
	req service.PrepareFromUploadRequest,
) (*service.PreparedSource, error) {
	if m.prepareErr != nil {
		return nil, m.prepareErr
	}
	if m.preparedSource != nil {
		m.preparedSource.SourceType = service.SourceTypeUpload
	}
	return m.preparedSource, m.prepareErr
}

func (m *mockSourcePreparationService) PrepareFromGit(
	ctx context.Context,
	req service.PrepareFromGitRequest,
) (*service.PreparedSource, error) {
	if m.prepareErr != nil {
		return nil, m.prepareErr
	}
	if m.preparedSource != nil {
		m.preparedSource.SourceType = service.SourceTypeGit
	}
	return m.preparedSource, m.prepareErr
}

func (m *mockSourcePreparationService) PrepareFromArchive(
	ctx context.Context,
	req service.PrepareFromArchiveRequest,
) (*service.PreparedSource, error) {
	if m.prepareErr != nil {
		return nil, m.prepareErr
	}
	if m.preparedSource != nil {
		m.preparedSource.SourceType = service.SourceTypeArchive
	}
	return m.preparedSource, m.prepareErr
}

// mockProcessingOrchestrator is a mock for TransactionProcessingOrchestrator
type mockProcessingOrchestrator struct {
	result *service.ProcessingResult
	err    error
}

func (m *mockProcessingOrchestrator) ProcessModuleWithTransaction(
	ctx context.Context,
	req service.ProcessingRequest,
) (*service.ProcessingResult, error) {
	return m.result, m.err
}

// Helper to create test ZIP content
func createTestZIPContent(t *testing.T) io.Reader {
	// Simple ZIP content for testing
	return strings.NewReader("PK\x03\x04") // ZIP file header
}

// Helper to create default processing options
func defaultProcessingOptions() service.ProcessingOptions {
	return service.ProcessingOptions{
		SkipArchiveExtraction:   false,
		SkipTerraformProcessing: false,
		SkipMetadataProcessing:  false,
		SkipSecurityScanning:    false,
		SkipFileContentStorage:  false,
		SkipArchiveGeneration:   false,
		SecurityScanEnabled:     true,
		FileProcessingEnabled:   true,
		GenerateArchives:        true,
		ArchiveFormats: []service.ArchiveFormat{
			service.ArchiveFormatZIP,
			service.ArchiveFormatTarGz,
		},
	}
}

// TestProcessModuleCommand_UploadSource tests processing from upload source
func TestProcessModuleCommand_UploadSource(t *testing.T) {
	ctx := context.Background()

	// Create mock services
	cleanupCalled := false
	preparedSource := &service.PreparedSource{
		SourcePath: "/tmp/test-module",
		CommitSHA:  nil,
		Cleanup: func() {
			cleanupCalled = true
		},
	}

	sourcePrepService := &mockSourcePreparationService{
		preparedSource: preparedSource,
		prepareErr:     nil,
	}

	processingResult := &service.ProcessingResult{
		Success: true,
	}

	orchestrator := &mockProcessingOrchestrator{
		result: processingResult,
		err:    nil,
	}

	// Create command with mocks
	cmd := NewProcessModuleCommand(sourcePrepService, orchestrator)

	// Create request
	zipContent := createTestZIPContent(t)
	req := ProcessModuleRequest{
		Namespace:    "test",
		Module:       "test-module",
		Provider:     "aws",
		Version:      "1.0.0",
		UploadSource: zipContent,
		UploadSize:   1024,
		Options:      defaultProcessingOptions(),
	}

	// Execute
	err := cmd.Execute(ctx, req)

	// Verify
	require.NoError(t, err)
	assert.True(t, cleanupCalled, "Cleanup should be called after processing")
}

// TestProcessModuleCommand_GitSource tests processing from git source
func TestProcessModuleCommand_GitSource(t *testing.T) {
	ctx := context.Background()

	// Create mock services
	cleanupCalled := false
	commitSHA := "abc123"
	preparedSource := &service.PreparedSource{
		SourcePath: "/tmp/test-module",
		CommitSHA:  &commitSHA,
		Cleanup: func() {
			cleanupCalled = true
		},
	}

	sourcePrepService := &mockSourcePreparationService{
		preparedSource: preparedSource,
		prepareErr:     nil,
	}

	processingResult := &service.ProcessingResult{
		Success: true,
	}

	orchestrator := &mockProcessingOrchestrator{
		result: processingResult,
		err:    nil,
	}

	// Create command with mocks
	cmd := NewProcessModuleCommand(sourcePrepService, orchestrator)

	// Create request
	gitTag := "v1.0.0"
	req := ProcessModuleRequest{
		Namespace: "test",
		Module:    "test-module",
		Provider:  "aws",
		Version:   "1.0.0",
		GitTag:    &gitTag,
		Options:   defaultProcessingOptions(),
	}

	// Execute
	err := cmd.Execute(ctx, req)

	// Verify
	require.NoError(t, err)
	assert.True(t, cleanupCalled, "Cleanup should be called after processing")
}

// TestProcessModuleCommand_ArchiveSource tests processing from archive source
func TestProcessModuleCommand_ArchiveSource(t *testing.T) {
	ctx := context.Background()

	// Create mock services
	cleanupCalled := false
	preparedSource := &service.PreparedSource{
		SourcePath: "/tmp/test-module",
		CommitSHA:  nil,
		Cleanup: func() {
			cleanupCalled = true
		},
	}

	sourcePrepService := &mockSourcePreparationService{
		preparedSource: preparedSource,
		prepareErr:     nil,
	}

	processingResult := &service.ProcessingResult{
		Success: true,
	}

	orchestrator := &mockProcessingOrchestrator{
		result: processingResult,
		err:    nil,
	}

	// Create command with mocks
	cmd := NewProcessModuleCommand(sourcePrepService, orchestrator)

	// Create request
	req := ProcessModuleRequest{
		Namespace:   "test",
		Module:      "test-module",
		Provider:    "aws",
		Version:     "1.0.0",
		ArchivePath: "/tmp/module.zip",
		Options:     defaultProcessingOptions(),
	}

	// Execute
	err := cmd.Execute(ctx, req)

	// Verify
	require.NoError(t, err)
	assert.True(t, cleanupCalled, "Cleanup should be called after processing")
}

// TestProcessModuleCommand_NoSourceError tests error when no source is specified
func TestProcessModuleCommand_NoSourceError(t *testing.T) {
	ctx := context.Background()

	// Create mock services
	sourcePrepService := &mockSourcePreparationService{}
	orchestrator := &mockProcessingOrchestrator{}

	// Create command with mocks
	cmd := NewProcessModuleCommand(sourcePrepService, orchestrator)

	// Create request with no source
	req := ProcessModuleRequest{
		Namespace: "test",
		Module:    "test-module",
		Provider:  "aws",
		Version:   "1.0.0",
		Options:   defaultProcessingOptions(),
	}

	// Execute
	err := cmd.Execute(ctx, req)

	// Verify
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no source specified")
}

// TestProcessModuleCommand_SourcePreparationError tests error handling during source preparation
func TestProcessModuleCommand_SourcePreparationError(t *testing.T) {
	ctx := context.Background()

	// Create mock services that return error
	sourcePrepService := &mockSourcePreparationService{
		preparedSource: nil,
		prepareErr:     assert.AnError,
	}
	orchestrator := &mockProcessingOrchestrator{}

	// Create command with mocks
	cmd := NewProcessModuleCommand(sourcePrepService, orchestrator)

	// Create request
	zipContent := createTestZIPContent(t)
	req := ProcessModuleRequest{
		Namespace:    "test",
		Module:       "test-module",
		Provider:     "aws",
		Version:      "1.0.0",
		UploadSource: zipContent,
		UploadSize:   1024,
		Options:      defaultProcessingOptions(),
	}

	// Execute
	err := cmd.Execute(ctx, req)

	// Verify
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "source preparation failed")
}

// TestProcessModuleCommand_ProcessingError tests error handling during processing
func TestProcessModuleCommand_ProcessingError(t *testing.T) {
	ctx := context.Background()

	// Create mock services where processing fails
	cleanupCalled := false
	preparedSource := &service.PreparedSource{
		SourcePath: "/tmp/test-module",
		CommitSHA:  nil,
		Cleanup: func() {
			cleanupCalled = true
		},
	}

	sourcePrepService := &mockSourcePreparationService{
		preparedSource: preparedSource,
		prepareErr:     nil,
	}

	orchestrator := &mockProcessingOrchestrator{
		result: &service.ProcessingResult{
			Success: false,
			Error:   func() *string { s := "processing failed"; return &s }(),
		},
		err: assert.AnError,
	}

	// Create command with mocks
	cmd := NewProcessModuleCommand(sourcePrepService, orchestrator)

	// Create request
	zipContent := createTestZIPContent(t)
	req := ProcessModuleRequest{
		Namespace:    "test",
		Module:       "test-module",
		Provider:     "aws",
		Version:      "1.0.0",
		UploadSource: zipContent,
		UploadSize:   1024,
		Options:      defaultProcessingOptions(),
	}

	// Execute
	err := cmd.Execute(ctx, req)

	// Verify
	assert.Error(t, err)
	assert.True(t, cleanupCalled, "Cleanup should be called even on error")
}

// TestProcessModuleCommand_CleanupOnSuccess tests cleanup is called on successful processing
func TestProcessModuleCommand_CleanupOnSuccess(t *testing.T) {
	ctx := context.Background()

	// Create mock services
	cleanupCalled := false
	preparedSource := &service.PreparedSource{
		SourcePath: "/tmp/test-module",
		CommitSHA:  nil,
		Cleanup: func() {
			cleanupCalled = true
		},
	}

	sourcePrepService := &mockSourcePreparationService{
		preparedSource: preparedSource,
		prepareErr:     nil,
	}

	processingResult := &service.ProcessingResult{
		Success: true,
	}

	orchestrator := &mockProcessingOrchestrator{
		result: processingResult,
		err:    nil,
	}

	// Create command with mocks
	cmd := NewProcessModuleCommand(sourcePrepService, orchestrator)

	// Create request
	zipContent := createTestZIPContent(t)
	req := ProcessModuleRequest{
		Namespace:    "test",
		Module:       "test-module",
		Provider:     "aws",
		Version:      "1.0.0",
		UploadSource: zipContent,
		UploadSize:   1024,
		Options:      defaultProcessingOptions(),
	}

	// Execute
	err := cmd.Execute(ctx, req)

	// Verify
	require.NoError(t, err)
	assert.True(t, cleanupCalled, "Cleanup must be called after successful processing")
}
