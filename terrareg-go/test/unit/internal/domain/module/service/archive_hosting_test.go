package service_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	configmodel "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/config/model"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/module/model"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/module/repository"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/module/service"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/persistence/sqldb/transaction"
)

// MockArchiveGenerationService is a mock for ArchiveGenerationTransactionService
type MockArchiveGenerationService struct {
	mock.Mock
}

func (m *MockArchiveGenerationService) GenerateArchivesWithTransaction(
	ctx context.Context,
	req service.ArchiveGenerationRequest,
) (*service.ArchiveGenerationResult, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(*service.ArchiveGenerationResult), args.Error(1)
}

func TestArchiveHostingLogic_DeleteExternallyHostedEnabled(t *testing.T) {
	// Create mock services
	mockArchiveService := &MockArchiveGenerationService{}

	// Create domain config with DELETE_EXTERNALLY_HOSTED_ARTIFACTS enabled
	domainConfig := &configmodel.DomainConfig{
		DeleteExternallyHostedArtifacts: true,
	}

	// Create orchestrator with domain config
	orchestrator := &service.TransactionProcessingOrchestrator{
		ArchiveGenService: mockArchiveService,
		DomainConfig:      domainConfig,
	}

	// Create a mock module version with git clone URL
	gitCloneURL := "https://github.com/example/module.git"
	moduleProvider := model.ReconstructModuleProvider(
		1, nil, "test-module", "aws", false,
		nil, &gitCloneURL, nil, nil, nil, false,
		nil, nil,
	)

	moduleVersion := model.ReconstructModuleVersion(
		1, "1.0.0", nil, false, false, false, nil,
		nil, nil, false, nil, nil, nil, nil, nil,
		nil, nil, nil,
	)

	// Set the provider relationship using reflection or helper method
	moduleProvider.SetVersions([]*model.ModuleVersion{moduleVersion})

	// Create processing request
	req := service.ProcessingRequest{
		Namespace:  "test-namespace",
		ModuleName: "test-module",
		Provider:   "aws",
		Version:    "1.0.0",
		ModulePath: "/tmp/module",
		Options: service.ProcessingOptions{
			GenerateArchives: true,
		},
	}

	// Mock expectation: ArchiveGenerationRequest should have GitCloneURL set and DeleteExternallyHostedArtifacts true
	expectedGenReq := service.ArchiveGenerationRequest{
		ModuleVersionID:                 1,
		SourcePath:                      "/tmp/module",
		Formats:                         []service.ArchiveFormat{service.ArchiveFormatZIP, service.ArchiveFormatTarGz},
		GitCloneURL:                     gitCloneURL,
		DeleteExternallyHostedArtifacts: true,
	}

	// Mock response: Skipped due to external hosting
	mockResult := &service.ArchiveGenerationResult{
		Success:             true,
		GeneratedArchives:   []service.GeneratedArchive{},
		SkippedReason:       "Module is externally hosted and DELETE_EXTERNALLY_HOSTED_ARTIFACTS is enabled",
		GenerationDuration:  0,
		SourceFilesCount:    0,
		TotalArchiveSize:    0,
		SavepointRolledBack: false,
	}

	mockArchiveService.On("GenerateArchivesWithTransaction", mock.Anything, expectedGenReq).Return(mockResult, nil)

	// Execute the archive generation phase
	ctx := context.Background()
	result := orchestrator.ExecuteArchiveGenerationPhase(ctx, req, moduleVersion)

	// Assertions
	assert.True(t, result.Success)
	assert.NotNil(t, result.Data)

	// Check that the skipped reason is properly set
	dataMap := result.Data.(map[string]interface{})
	assert.Equal(t, "Module is externally hosted and DELETE_EXTERNALLY_HOSTED_ARTIFACTS is enabled", dataMap["skipped_reason"])

	mockArchiveService.AssertExpectations(t)
}

func TestArchiveHostingLogic_DeleteExternallyHostedDisabled(t *testing.T) {
	// Create mock services
	mockArchiveService := &MockArchiveGenerationService{}

	// Create domain config with DELETE_EXTERNALLY_HOSTED_ARTIFACTS disabled
	domainConfig := &configmodel.DomainConfig{
		DeleteExternallyHostedArtifacts: false,
	}

	// Create orchestrator with domain config
	orchestrator := &service.TransactionProcessingOrchestrator{
		ArchiveGenService: mockArchiveService,
		DomainConfig:      domainConfig,
	}

	// Create a mock module version with git clone URL
	gitCloneURL := "https://github.com/example/module.git"
	moduleProvider := model.ReconstructModuleProvider(
		1, nil, "test-module", "aws", false,
		nil, &gitCloneURL, nil, nil, nil, false,
		nil, nil,
	)

	moduleVersion := model.ReconstructModuleVersion(
		1, "1.0.0", nil, false, false, false, nil,
		nil, nil, false, nil, nil, nil, nil, nil,
		nil, nil, nil,
	)

	// Set the provider relationship
	moduleProvider.SetVersions([]*model.ModuleVersion{moduleVersion})

	// Create processing request
	req := service.ProcessingRequest{
		Namespace:  "test-namespace",
		ModuleName: "test-module",
		Provider:   "aws",
		Version:    "1.0.0",
		ModulePath: "/tmp/module",
		Options: service.ProcessingOptions{
			GenerateArchives: true,
		},
	}

	// Mock expectation: ArchiveGenerationRequest should have GitCloneURL set but DeleteExternallyHostedArtifacts false
	expectedGenReq := service.ArchiveGenerationRequest{
		ModuleVersionID:                 1,
		SourcePath:                      "/tmp/module",
		Formats:                         []service.ArchiveFormat{service.ArchiveFormatZIP, service.ArchiveFormatTarGz},
		GitCloneURL:                     gitCloneURL,
		DeleteExternallyHostedArtifacts: false,
	}

	// Mock response: Archives generated successfully
	mockResult := &service.ArchiveGenerationResult{
		Success: true,
		GeneratedArchives: []service.GeneratedArchive{
			{Format: service.ArchiveFormatZIP, Path: "/tmp/module.zip", Size: 1024},
			{Format: service.ArchiveFormatTarGz, Path: "/tmp/module.tar.gz", Size: 2048},
		},
		GenerationDuration:  1000000, // 1ms
		SourceFilesCount:    10,
		TotalArchiveSize:    3072,
		SavepointRolledBack: false,
	}

	mockArchiveService.On("GenerateArchivesWithTransaction", mock.Anything, expectedGenReq).Return(mockResult, nil)

	// Execute the archive generation phase
	ctx := context.Background()
	result := orchestrator.ExecuteArchiveGenerationPhase(ctx, req, moduleVersion)

	// Assertions
	assert.True(t, result.Success)
	assert.NotNil(t, result.Data)

	// Check that archives were generated
	dataMap := result.Data.(map[string]interface{})
	assert.Equal(t, []string{"/tmp/module.zip", "/tmp/module.tar.gz"}, dataMap["generated_archives"])
	assert.Equal(t, int64(3072), dataMap["total_archive_size"])
	assert.Equal(t, 10, dataMap["source_files_count"])
	assert.Empty(t, dataMap["skipped_reason"])

	mockArchiveService.AssertExpectations(t)
}

func TestArchiveHostingLogic_NoGitCloneURL(t *testing.T) {
	// Create mock services
	mockArchiveService := &MockArchiveGenerationService{}

	// Create domain config with DELETE_EXTERNALLY_HOSTED_ARTIFACTS enabled
	domainConfig := &configmodel.DomainConfig{
		DeleteExternallyHostedArtifacts: true,
	}

	// Create orchestrator with domain config
	orchestrator := &service.TransactionProcessingOrchestrator{
		ArchiveGenService: mockArchiveService,
		DomainConfig:      domainConfig,
	}

	// Create a mock module version WITHOUT git clone URL (nil)
	moduleProvider := model.ReconstructModuleProvider(
		1, nil, "test-module", "aws", false,
		nil, nil, nil, nil, nil, false,
		nil, nil,
	)

	moduleVersion := model.ReconstructModuleVersion(
		1, "1.0.0", nil, false, false, false, nil,
		nil, nil, false, nil, nil, nil, nil, nil,
		nil, nil, nil,
	)

	// Set the provider relationship
	moduleProvider.SetVersions([]*model.ModuleVersion{moduleVersion})

	// Create processing request
	req := service.ProcessingRequest{
		Namespace:  "test-namespace",
		ModuleName: "test-module",
		Provider:   "aws",
		Version:    "1.0.0",
		ModulePath: "/tmp/module",
		Options: service.ProcessingOptions{
			GenerateArchives: true,
		},
	}

	// Mock expectation: ArchiveGenerationRequest should have empty GitCloneURL and DeleteExternallyHostedArtifacts true
	expectedGenReq := service.ArchiveGenerationRequest{
		ModuleVersionID:                 1,
		SourcePath:                      "/tmp/module",
		Formats:                         []service.ArchiveFormat{service.ArchiveFormatZIP, service.ArchiveFormatTarGz},
		GitCloneURL:                     "", // Empty because no git clone URL configured
		DeleteExternallyHostedArtifacts: true,
	}

	// Mock response: Archives generated (because module is not externally hosted)
	mockResult := &service.ArchiveGenerationResult{
		Success: true,
		GeneratedArchives: []service.GeneratedArchive{
			{Format: service.ArchiveFormatZIP, Path: "/tmp/module.zip", Size: 1024},
		},
		GenerationDuration:  1000000,
		SourceFilesCount:    5,
		TotalArchiveSize:    1024,
		SavepointRolledBack: false,
	}

	mockArchiveService.On("GenerateArchivesWithTransaction", mock.Anything, expectedGenReq).Return(mockResult, nil)

	// Execute the archive generation phase
	ctx := context.Background()
	result := orchestrator.ExecuteArchiveGenerationPhase(ctx, req, moduleVersion)

	// Assertions
	assert.True(t, result.Success)
	assert.NotNil(t, result.Data)

	// Check that archives were generated (not skipped, because no git clone URL)
	dataMap := result.Data.(map[string]interface{})
	assert.Equal(t, []string{"/tmp/module.zip"}, dataMap["generated_archives"])
	assert.Empty(t, dataMap["skipped_reason"])

	mockArchiveService.AssertExpectations(t)
}
