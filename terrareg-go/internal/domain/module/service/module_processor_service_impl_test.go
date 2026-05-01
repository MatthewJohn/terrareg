package service

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	configModel "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/config/model"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/module/model"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/shared/service"
	types "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/shared/types"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/logging"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/persistence/sqldb"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Mock implementations for testing

type mockModuleParser struct {
	parseModuleFunc          func(modulePath string) (*ParseResult, error)
	detectSubmodulesFunc     func(modulePath string) ([]string, error)
	detectExamplesFunc       func(modulePath string) ([]string, error)
	parseSubmoduleFunc       func(submodulePath string) (*ParseResult, error)
	parseExampleFunc         func(examplePath string, infracostJSON []byte) (*ExampleParseResult, error)
	extractExampleFilesFunc  func(examplePath string) ([]*model.ExampleFile, error)
}

func (m *mockModuleParser) ParseModule(modulePath string) (*ParseResult, error) {
	if m.parseModuleFunc != nil {
		return m.parseModuleFunc(modulePath)
	}
	return &ParseResult{}, nil
}

func (m *mockModuleParser) DetectSubmodules(modulePath string) ([]string, error) {
	if m.detectSubmodulesFunc != nil {
		return m.detectSubmodulesFunc(modulePath)
	}
	return []string{}, nil
}

func (m *mockModuleParser) DetectExamples(modulePath string) ([]string, error) {
	if m.detectExamplesFunc != nil {
		return m.detectExamplesFunc(modulePath)
	}
	return []string{}, nil
}

func (m *mockModuleParser) ParseSubmodule(submodulePath string) (*ParseResult, error) {
	if m.parseSubmoduleFunc != nil {
		return m.parseSubmoduleFunc(submodulePath)
	}
	return &ParseResult{}, nil
}

func (m *mockModuleParser) ParseExample(examplePath string, infracostJSON []byte) (*ExampleParseResult, error) {
	if m.parseExampleFunc != nil {
		return m.parseExampleFunc(examplePath, infracostJSON)
	}
	return &ExampleParseResult{ParseResult: &ParseResult{}}, nil
}

func (m *mockModuleParser) ExtractExampleFiles(examplePath string) ([]*model.ExampleFile, error) {
	if m.extractExampleFilesFunc != nil {
		return m.extractExampleFilesFunc(examplePath)
	}
	return []*model.ExampleFile{}, nil
}

type mockModuleDetailsRepo struct {
	saveAndReturnIDFunc func(ctx context.Context, details *model.ModuleDetails) (int, error)
}

func (m *mockModuleDetailsRepo) Save(ctx context.Context, details *model.ModuleDetails) (*model.ModuleDetails, error) {
	return details, nil
}

func (m *mockModuleDetailsRepo) SaveAndReturnID(ctx context.Context, details *model.ModuleDetails) (int, error) {
	if m.saveAndReturnIDFunc != nil {
		return m.saveAndReturnIDFunc(ctx, details)
	}
	return 1, nil
}

func (m *mockModuleDetailsRepo) FindByID(ctx context.Context, id int) (*model.ModuleDetails, error) {
	return nil, nil
}

func (m *mockModuleDetailsRepo) FindByModuleVersionID(ctx context.Context, moduleVersionID int) (*model.ModuleDetails, error) {
	return nil, nil
}

func (m *mockModuleDetailsRepo) Update(ctx context.Context, id int, details *model.ModuleDetails) (*model.ModuleDetails, error) {
	return details, nil
}

func (m *mockModuleDetailsRepo) Delete(ctx context.Context, id int) error {
	return nil
}

type mockModuleVersionRepo struct {
	updateModuleDetailsIDFunc func(ctx context.Context, moduleVersionID int, moduleDetailsID int) error
}

func (m *mockModuleVersionRepo) UpdateModuleDetailsID(ctx context.Context, moduleVersionID int, moduleDetailsID int) error {
	if m.updateModuleDetailsIDFunc != nil {
		return m.updateModuleDetailsIDFunc(ctx, moduleVersionID, moduleDetailsID)
	}
	return nil
}

func (m *mockModuleVersionRepo) FindByID(ctx context.Context, id int) (*model.ModuleVersion, error) {
	return nil, nil
}

func (m *mockModuleVersionRepo) FindByModuleProvider(ctx context.Context, moduleProviderID int, includeBeta, includeUnpublished bool) ([]*model.ModuleVersion, error) {
	return []*model.ModuleVersion{}, nil
}

func (m *mockModuleVersionRepo) Save(ctx context.Context, moduleVersion *model.ModuleVersion) (*model.ModuleVersion, error) {
	return moduleVersion, nil
}

func (m *mockModuleVersionRepo) FindByModuleProviderAndVersion(ctx context.Context, moduleProviderID int, version types.ModuleVersion) (*model.ModuleVersion, error) {
	return nil, nil
}

func (m *mockModuleVersionRepo) Delete(ctx context.Context, id int) error {
	return nil
}

func (m *mockModuleVersionRepo) Exists(ctx context.Context, moduleProviderID int, version types.ModuleVersion) (bool, error) {
	return false, nil
}

type mockSubmoduleRepo struct {
	saveFunc                  func(ctx context.Context, moduleVersionID int, submodule *sqldb.SubmoduleDB) (*sqldb.SubmoduleDB, error)
	saveWithDetailsFunc       func(ctx context.Context, moduleVersionID int, submodule *sqldb.SubmoduleDB, detailsID int) (*sqldb.SubmoduleDB, error)
	updateModuleDetailsIDFunc func(ctx context.Context, submoduleID int, detailsID int) error
}

func (m *mockSubmoduleRepo) Save(ctx context.Context, moduleVersionID int, submodule *sqldb.SubmoduleDB) (*sqldb.SubmoduleDB, error) {
	if m.saveFunc != nil {
		return m.saveFunc(ctx, moduleVersionID, submodule)
	}
	return &sqldb.SubmoduleDB{ID: 1}, nil
}

func (m *mockSubmoduleRepo) SaveWithDetails(ctx context.Context, moduleVersionID int, submodule *sqldb.SubmoduleDB, detailsID int) (*sqldb.SubmoduleDB, error) {
	if m.saveWithDetailsFunc != nil {
		return m.saveWithDetailsFunc(ctx, moduleVersionID, submodule, detailsID)
	}
	submodule.ID = 1
	return submodule, nil
}

func (m *mockSubmoduleRepo) UpdateModuleDetailsID(ctx context.Context, submoduleID int, detailsID int) error {
	if m.updateModuleDetailsIDFunc != nil {
		return m.updateModuleDetailsIDFunc(ctx, submoduleID, detailsID)
	}
	return nil
}

func (m *mockSubmoduleRepo) FindByParentModuleVersion(ctx context.Context, moduleVersionID int) ([]sqldb.SubmoduleDB, error) {
	return []sqldb.SubmoduleDB{}, nil
}

func (m *mockSubmoduleRepo) FindByPath(ctx context.Context, moduleVersionID int, path string) (*sqldb.SubmoduleDB, error) {
	return nil, nil
}

func (m *mockSubmoduleRepo) DeleteByParentModuleVersion(ctx context.Context, moduleVersionID int) error {
	return nil
}

type mockExampleFileRepo struct {
	saveBatchFunc func(ctx context.Context, files []*sqldb.ExampleFileDB) ([]*sqldb.ExampleFileDB, error)
}

func (m *mockExampleFileRepo) Save(ctx context.Context, file *sqldb.ExampleFileDB) (*sqldb.ExampleFileDB, error) {
	return file, nil
}

func (m *mockExampleFileRepo) SaveBatch(ctx context.Context, files []*sqldb.ExampleFileDB) ([]*sqldb.ExampleFileDB, error) {
	if m.saveBatchFunc != nil {
		return m.saveBatchFunc(ctx, files)
	}
	return files, nil
}

func (m *mockExampleFileRepo) FindBySubmoduleID(ctx context.Context, submoduleID int) ([]sqldb.ExampleFileDB, error) {
	return []sqldb.ExampleFileDB{}, nil
}

func (m *mockExampleFileRepo) DeleteBySubmoduleID(ctx context.Context, submoduleID int) error {
	return nil
}

func (m *mockExampleFileRepo) DeleteByModuleVersion(ctx context.Context, moduleVersionID int) error {
	return nil
}

type mockInfracostService struct {
	isAvailableFunc    func() bool
	analyzeExampleFunc func(ctx context.Context, examplePath string) ([]byte, error)
}

func (m *mockInfracostService) IsAvailable() bool {
	if m.isAvailableFunc != nil {
		return m.isAvailableFunc()
	}
	return false
}

func (m *mockInfracostService) AnalyzeExample(ctx context.Context, examplePath string) ([]byte, error) {
	if m.analyzeExampleFunc != nil {
		return m.analyzeExampleFunc(ctx, examplePath)
	}
	return nil, nil
}

type mockSystemCommandService struct{}

func (m *mockSystemCommandService) Execute(ctx context.Context, cmd *service.Command) (*service.CommandResult, error) {
	return &service.CommandResult{
		Stdout:   `{"terraform_version": "1.5.0"}`,
		Stderr:   "",
		ExitCode: 0,
	}, nil
}

func (m *mockSystemCommandService) ExecuteWithInput(ctx context.Context, cmd *service.Command, input string) (*service.CommandResult, error) {
	return &service.CommandResult{
		Stdout:   "",
		Stderr:   "",
		ExitCode: 0,
	}, nil
}

// Test NewModuleProcessorServiceImpl
func TestNewModuleProcessorServiceImpl(t *testing.T) {
	parser := &mockModuleParser{}
	detailsRepo := &mockModuleDetailsRepo{}
	versionRepo := &mockModuleVersionRepo{}
	submoduleRepo := &mockSubmoduleRepo{}
	exampleFileRepo := &mockExampleFileRepo{}
	infracostService := &mockInfracostService{}
	var securityService *SecurityScanningService = nil
	var commandService service.SystemCommandService = &mockSystemCommandService{}
	config := &configModel.DomainConfig{}
	logger := logging.NewZeroLogger(zerolog.New(zerolog.Nop()).With().Timestamp().Logger())

	service := NewModuleProcessorServiceImpl(
		parser,
		detailsRepo,
		versionRepo,
		submoduleRepo,
		exampleFileRepo,
		infracostService,
		securityService,
		commandService,
		config,
		logger,
	)

	assert.NotNil(t, service)
}

// Test ProcessModule_Success
func TestProcessModule_Success(t *testing.T) {
	ctx := context.Background()
	moduleDir := "/test/module"

	parseResult := &ParseResult{
		Description:      "Test module",
		ReadmeContent:    "# Test Module\nThis is a test",
		RawTerraformDocs: []byte(`{"header":"Test"}`),
		Variables: []Variable{
			{Name: "var1", Type: "string", Description: "Test variable", Required: true},
		},
		Outputs: []Output{
			{Name: "output1", Description: "Test output"},
		},
		TerraformRequirements: []TerraformRequirement{
			{Name: "terraform", Version: ">= 1.0"},
		},
	}

	parser := &mockModuleParser{
		parseModuleFunc: func(modulePath string) (*ParseResult, error) {
			return parseResult, nil
		},
		detectSubmodulesFunc: func(modulePath string) ([]string, error) {
			return []string{"submodule1"}, nil
		},
		detectExamplesFunc: func(modulePath string) ([]string, error) {
			return []string{"example1"}, nil
		},
		extractExampleFilesFunc: func(examplePath string) ([]*model.ExampleFile, error) {
			return []*model.ExampleFile{
				model.NewExampleFile("main.tf", []byte("resource \"aws_s3_bucket\" \"test\" {}")),
			}, nil
		},
	}

	detailsRepo := &mockModuleDetailsRepo{
		saveAndReturnIDFunc: func(ctx context.Context, details *model.ModuleDetails) (int, error) {
			return 123, nil
		},
	}

	versionRepo := &mockModuleVersionRepo{
		updateModuleDetailsIDFunc: func(ctx context.Context, moduleVersionID int, moduleDetailsID int) error {
			return nil
		},
	}

	submoduleRepo := &mockSubmoduleRepo{
		saveWithDetailsFunc: func(ctx context.Context, moduleVersionID int, submodule *sqldb.SubmoduleDB, detailsID int) (*sqldb.SubmoduleDB, error) {
			return submodule, nil
		},
		saveFunc: func(ctx context.Context, moduleVersionID int, submodule *sqldb.SubmoduleDB) (*sqldb.SubmoduleDB, error) {
			return &sqldb.SubmoduleDB{ID: 1}, nil
		},
	}

	exampleFileRepo := &mockExampleFileRepo{}

	infracostService := &mockInfracostService{
		isAvailableFunc: func() bool {
			return false
		},
	}

	// Don't use security scanning service in tests to avoid nil pointer issues
	var securityService *SecurityScanningService = nil
	var commandService service.SystemCommandService = &mockSystemCommandService{}

	config := &configModel.DomainConfig{
		ExamplesDirectory: "examples",
	}

	logger := logging.NewZeroLogger(zerolog.New(zerolog.Nop()).With().Timestamp().Logger())

	service := NewModuleProcessorServiceImpl(
		parser,
		detailsRepo,
		versionRepo,
		submoduleRepo,
		exampleFileRepo,
		infracostService,
		securityService,
		commandService,
		config,
		logger,
	)

	metadata := &ModuleProcessingMetadata{
		ModuleVersionID: 456,
		GitTag:          "v1.0.0",
		GitURL:          "https://github.com/test/module",
		GitPath:         "/",
		CommitSHA:       "abc123",
	}

	result, err := service.ProcessModule(ctx, moduleDir, metadata)

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "Test module", result.ModuleMetadata.Description)
	assert.Len(t, result.Submodules, 1)
	assert.Len(t, result.Examples, 1)
	assert.Equal(t, "# Test Module\nThis is a test", result.ReadmeContent)
}

// Test ProcessModule_ParseModuleFailure
func TestProcessModule_ParseModuleFailure(t *testing.T) {
	ctx := context.Background()
	moduleDir := "/test/module"

	parser := &mockModuleParser{
		parseModuleFunc: func(modulePath string) (*ParseResult, error) {
			return nil, errors.New("parse failed")
		},
		detectSubmodulesFunc: func(modulePath string) ([]string, error) {
			return []string{}, nil
		},
		detectExamplesFunc: func(modulePath string) ([]string, error) {
			return []string{}, nil
		},
	}

	detailsRepo := &mockModuleDetailsRepo{
		saveAndReturnIDFunc: func(ctx context.Context, details *model.ModuleDetails) (int, error) {
			return 123, nil
		},
	}

	versionRepo := &mockModuleVersionRepo{
		updateModuleDetailsIDFunc: func(ctx context.Context, moduleVersionID int, moduleDetailsID int) error {
			return nil
		},
	}

	submoduleRepo := &mockSubmoduleRepo{}
	exampleFileRepo := &mockExampleFileRepo{}
	infracostService := &mockInfracostService{}
	var securityService *SecurityScanningService = nil
	var commandService service.SystemCommandService = &mockSystemCommandService{}
	config := &configModel.DomainConfig{}
	logger := logging.NewZeroLogger(zerolog.New(zerolog.Nop()).With().Timestamp().Logger())

	service := NewModuleProcessorServiceImpl(
		parser,
		detailsRepo,
		versionRepo,
		submoduleRepo,
		exampleFileRepo,
		infracostService,
		securityService,
		commandService,
		config,
		logger,
	)

	metadata := &ModuleProcessingMetadata{
		ModuleVersionID: 456,
		GitTag:          "v1.0.0",
		GitURL:          "https://github.com/test/module",
		GitPath:         "/",
		CommitSHA:       "abc123",
	}

	result, err := service.ProcessModule(ctx, moduleDir, metadata)

	// Should continue with partial processing
	require.NoError(t, err)
	assert.NotNil(t, result)
}

// Test ValidateModuleStructure_Success
func TestValidateModuleStructure_Success(t *testing.T) {
	ctx := context.Background()

	// Use a real temporary directory with .tf files
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "main.tf")
	err := os.WriteFile(testFile, []byte("resource \"test\" \"example\" {}"), 0644)
	require.NoError(t, err)

	service := &ModuleProcessorServiceImpl{}

	err = service.ValidateModuleStructure(ctx, tempDir)
	assert.NoError(t, err)
}

// Test ValidateModuleStructure_NoTerraformFiles
func TestValidateModuleStructure_NoTerraformFiles(t *testing.T) {
	ctx := context.Background()

	// Use a real temporary directory without .tf files
	tempDir := t.TempDir()

	service := &ModuleProcessorServiceImpl{}

	err := service.ValidateModuleStructure(ctx, tempDir)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no Terraform files")
}

// Test ValidateModuleStructure_DirectoryNotExists
func TestValidateModuleStructure_DirectoryNotExists(t *testing.T) {
	ctx := context.Background()

	service := &ModuleProcessorServiceImpl{}

	err := service.ValidateModuleStructure(ctx, "/nonexistent/directory")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "does not exist")
}

// Test ExtractMetadata_Success
func TestExtractMetadata_Success(t *testing.T) {
	ctx := context.Background()
	moduleDir := "/test/module"

	parseResult := &ParseResult{
		Description: "Test module",
		Variables: []Variable{
			{Name: "var1", Type: "string", Description: "Test variable", Required: true},
		},
		Outputs: []Output{
			{Name: "output1", Description: "Test output"},
		},
		ProviderVersions: []ProviderVersion{
			{Name: "aws", Version: ">= 4.0"},
		},
		Resources: []Resource{
			{Type: "aws_s3_bucket", Name: "test"},
		},
		Dependencies: []model.Dependency{
			{Source: "terraform-aws-modules/vpc/aws", Version: "5.0.0"},
		},
	}

	parser := &mockModuleParser{
		parseModuleFunc: func(modulePath string) (*ParseResult, error) {
			return parseResult, nil
		},
	}

	config := &configModel.DomainConfig{}
	logger := logging.NewZeroLogger(zerolog.New(zerolog.Nop()).With().Timestamp().Logger())

	service := NewModuleProcessorServiceImpl(
		parser,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		config,
		logger,
	)

	metadata, err := service.ExtractMetadata(ctx, moduleDir)

	require.NoError(t, err)
	assert.NotNil(t, metadata)
	assert.Equal(t, "Test module", metadata.Description)
	assert.Len(t, metadata.Variables, 1)
	assert.Len(t, metadata.Outputs, 1)
	assert.Len(t, metadata.Providers, 1)
	assert.Len(t, metadata.Resources, 1)
	assert.Len(t, metadata.Dependencies, 1)
}

// Test ExtractMetadata_ParseFailure
func TestExtractMetadata_ParseFailure(t *testing.T) {
	ctx := context.Background()
	moduleDir := "/test/module"

	parser := &mockModuleParser{
		parseModuleFunc: func(modulePath string) (*ParseResult, error) {
			return nil, errors.New("parse failed")
		},
	}

	config := &configModel.DomainConfig{}
	logger := logging.NewZeroLogger(zerolog.New(zerolog.Nop()).With().Timestamp().Logger())

	service := NewModuleProcessorServiceImpl(
		parser,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		config,
		logger,
	)

	metadata, err := service.ExtractMetadata(ctx, moduleDir)

	assert.Error(t, err)
	assert.Nil(t, metadata)
	assert.Contains(t, err.Error(), "failed to parse module")
}

// Test SubmoduleInfo_Structure (after removing Source/Version fields)
func TestSubmoduleInfo_Structure(t *testing.T) {
	submodule := SubmoduleInfo{
		Path: "modules/submodule1",
	}

	assert.Equal(t, "modules/submodule1", submodule.Path)
}

// Test ExampleInfo_Structure
func TestExampleInfo_Structure(t *testing.T) {
	example := ExampleInfo{
		Name:        "example1",
		Description: "Test example",
		Files:       []string{"main.tf", "variables.tf"},
	}

	assert.Equal(t, "example1", example.Name)
	assert.Equal(t, "Test example", example.Description)
	assert.Len(t, example.Files, 2)
	assert.Equal(t, "main.tf", example.Files[0])
	assert.Equal(t, "variables.tf", example.Files[1])
}

// Test ModuleProcessingMetadata_Structure
func TestModuleProcessingMetadata_Structure(t *testing.T) {
	metadata := ModuleProcessingMetadata{
		ModuleVersionID: 123,
		GitTag:          "v1.0.0",
		GitURL:          "https://github.com/test/module",
		GitPath:         "/",
		CommitSHA:       "abc123",
	}

	assert.Equal(t, 123, metadata.ModuleVersionID)
	assert.Equal(t, "v1.0.0", metadata.GitTag)
	assert.Equal(t, "https://github.com/test/module", metadata.GitURL)
	assert.Equal(t, "/", metadata.GitPath)
	assert.Equal(t, "abc123", metadata.CommitSHA)
}

// Test ModuleProcessingResult_Structure
func TestModuleProcessingResult_Structure(t *testing.T) {
	moduleMetadata := &ModuleMetadata{
		Name:        "test-module",
		Description: "Test module",
		Version:     "v1.0.0",
	}

	result := ModuleProcessingResult{
		ModuleMetadata:   moduleMetadata,
		Submodules:       []SubmoduleInfo{{Path: "sub1"}},
		Examples:         []ExampleInfo{{Name: "ex1"}},
		ReadmeContent:    "# Test",
		VariableTemplate: "{}",
		ProcessedFiles:   []string{"main.tf"},
	}

	assert.Equal(t, moduleMetadata, result.ModuleMetadata)
	assert.Len(t, result.Submodules, 1)
	assert.Len(t, result.Examples, 1)
	assert.Equal(t, "# Test", result.ReadmeContent)
	assert.Equal(t, "{}", result.VariableTemplate)
	assert.Len(t, result.ProcessedFiles, 1)
}
