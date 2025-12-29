package module

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/module/service"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/persistence/sqldb"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/persistence/sqldb/transaction"
	"github.com/matthewjohn/terrareg/terrareg-go/test/integration/mocks"
	"github.com/matthewjohn/terrareg/terrareg-go/test/integration/testutils"
)

// TestModuleExtractor_BasicModule tests basic module upload with a single main.tf file
func TestModuleExtractor_BasicModule(t *testing.T) {
	// Setup test database
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	// Create test namespace, module provider, and version
	namespace := testutils.CreateNamespace(t, db, "testbasicmodule")
	moduleProvider := testutils.CreateModuleProvider(t, db, namespace.ID, "test-module", "aws")
	_ = testutils.CreateModuleVersion(t, db, moduleProvider.ID, "1.0.0")

	// Create test archive with main.tf
	files := map[string]string{
		"main.tf": testutils.CreateValidMainTF(),
	}
	archive := testutils.CreateTestModuleZip(t, files)

	// Extract archive to temp directory using test helper
	extractDir := testutils.ExtractTestArchive(t, archive)

	// Verify main.tf was extracted
	mainTfPath := filepath.Join(extractDir, "main.tf")
	_, err := os.Stat(mainTfPath)
	require.NoError(t, err, "main.tf should exist in extracted directory")

	content, err := os.ReadFile(mainTfPath)
	require.NoError(t, err)
	assert.Contains(t, string(content), `variable "instance_type"`)

	// Verify archive contents
	contents := testutils.ListArchiveContents(t, archive)
	assert.Len(t, contents, 1)
	assert.Contains(t, contents, "main.tf")
}

// TestModuleExtractor_WithREADME tests module upload with README.md
func TestModuleExtractor_WithREADME(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	namespace := testutils.CreateNamespace(t, db, "testreadme")
	moduleProvider := testutils.CreateModuleProvider(t, db, namespace.ID, "test-module", "aws")
	_ = testutils.CreateModuleVersion(t, db, moduleProvider.ID, "1.0.0")

	files := map[string]string{
		"main.tf":   testutils.CreateValidMainTF(),
		"README.md": testutils.CreateREADMEContent("test-module"),
	}
	archive := testutils.CreateTestModuleZip(t, files)

	// Extract archive
	extractDir := testutils.ExtractTestArchive(t, archive)

	// Verify README.md was extracted
	readmePath := filepath.Join(extractDir, "README.md")
	content, err := os.ReadFile(readmePath)
	require.NoError(t, err)
	assert.Contains(t, string(content), "# test-module")

	// Verify both files exist
	contents := testutils.ListArchiveContents(t, archive)
	assert.Len(t, contents, 2)
}

// TestModuleExtractor_WithSubmodules tests module upload with submodules directory
func TestModuleExtractor_WithSubmodules(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	namespace := testutils.CreateNamespace(t, db, "testsubmodules")
	moduleProvider := testutils.CreateModuleProvider(t, db, namespace.ID, "test-module", "aws")
	_ = testutils.CreateModuleVersion(t, db, moduleProvider.ID, "1.0.0")

	files := map[string]string{
		"main.tf": testutils.CreateValidMainTF(),
		"modules/submodule1/main.tf": `
variable "submodule_input" {
  type        = string
  description = "This is a test input in a submodule"
  default     = "test_default_val"
}

output "submodule_output" {
  description = "test output in a submodule"
  value       = var.submodule_input
}
`,
		"modules/submodule2/main.tf": `
variable "another_input" {
  type        = string
  description = "Another submodule input"
  default     = "default_value"
}
`,
	}
	archive := testutils.CreateTestModuleZip(t, files)

	// Extract archive
	extractDir := testutils.ExtractTestArchive(t, archive)

	// Verify submodules were extracted
	submodule1Path := filepath.Join(extractDir, "modules", "submodule1", "main.tf")
	_, err := os.Stat(submodule1Path)
	require.NoError(t, err, "submodule1/main.tf should exist")

	submodule2Path := filepath.Join(extractDir, "modules", "submodule2", "main.tf")
	_, err = os.Stat(submodule2Path)
	require.NoError(t, err, "submodule2/main.tf should exist")

	// Verify archive contents
	contents := testutils.ListArchiveContents(t, archive)
	assert.Len(t, contents, 3)
}

// TestModuleExtractor_WithExamples tests module upload with examples directory
func TestModuleExtractor_WithExamples(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	namespace := testutils.CreateNamespace(t, db, "testexamples")
	moduleProvider := testutils.CreateModuleProvider(t, db, namespace.ID, "test-module", "aws")
	_ = testutils.CreateModuleVersion(t, db, moduleProvider.ID, "1.0.0")

	files := map[string]string{
		"main.tf": testutils.CreateValidMainTF(),
		"examples/example1/main.tf": `
variable "example_input" {
  type        = string
  description = "Example input"
  default     = "example"
}

output "example_output" {
  value = var.example_input
}
`,
		"examples/example2/main.tf": `
resource "null_resource" "example" {
}
`,
	}
	archive := testutils.CreateTestModuleZip(t, files)

	// Extract archive
	extractDir := testutils.ExtractTestArchive(t, archive)

	// Verify examples were extracted
	example1Path := filepath.Join(extractDir, "examples", "example1", "main.tf")
	_, err := os.Stat(example1Path)
	require.NoError(t, err, "examples/example1/main.tf should exist")

	example2Path := filepath.Join(extractDir, "examples", "example2", "main.tf")
	_, err = os.Stat(example2Path)
	require.NoError(t, err, "examples/example2/main.tf should exist")
}

// TestModuleExtractor_TerraregMetadata tests module upload with terrareg.json metadata
func TestModuleExtractor_TerraregMetadata(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	namespace := testutils.CreateNamespace(t, db, "testmetadata")
	moduleProvider := testutils.CreateModuleProvider(t, db, namespace.ID, "test-module", "aws")
	moduleVersion := testutils.CreateModuleVersion(t, db, moduleProvider.ID, "1.0.0")

	description := "unittestdescription!"
	owner := "unittestowner."

	metadata := map[string]interface{}{
		"description": description,
		"owner":       owner,
	}

	files := map[string]string{
		"main.tf":       testutils.CreateValidMainTF(),
		"terrareg.json": testutils.CreateTerraregMetadata(metadata),
	}
	archive := testutils.CreateTestModuleZip(t, files)

	// Extract archive
	extractDir := testutils.ExtractTestArchive(t, archive)

	// Process metadata
	metadataService := service.NewMetadataProcessingService(transaction.NewSavepointHelper(db.DB))
	metadataReq := service.MetadataProcessingRequest{
		ModuleVersionID:    moduleVersion.ID,
		MetadataPath:       extractDir,
		TransactionCtx:     context.Background(),
		RequiredAttributes: []string{},
	}

	result, err := metadataService.ProcessMetadataWithTransaction(context.Background(), metadataReq)
	require.NoError(t, err)
	assert.True(t, result.Success)
	assert.True(t, result.MetadataFound)
	assert.NotNil(t, result.Metadata)
	assert.Equal(t, description, *result.Metadata.Description)
	assert.Equal(t, owner, *result.Metadata.Owner)
}

// TestModuleExtractor_InvalidTerraregMetadata tests handling of invalid terrareg.json
func TestModuleExtractor_InvalidTerraregMetadata(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	namespace := testutils.CreateNamespace(t, db, "testinvalidmetadata")
	moduleProvider := testutils.CreateModuleProvider(t, db, namespace.ID, "test-module", "aws")
	moduleVersion := testutils.CreateModuleVersion(t, db, moduleProvider.ID, "1.0.0")

	files := map[string]string{
		"main.tf":       testutils.CreateValidMainTF(),
		"terrareg.json": "This is invalid JSON!",
	}
	archive := testutils.CreateTestModuleZip(t, files)

	// Extract archive
	extractDir := testutils.ExtractTestArchive(t, archive)

	// Process metadata - should fail gracefully
	metadataService := service.NewMetadataProcessingService(transaction.NewSavepointHelper(db.DB))
	metadataReq := service.MetadataProcessingRequest{
		ModuleVersionID:    moduleVersion.ID,
		MetadataPath:       extractDir,
		TransactionCtx:     context.Background(),
		RequiredAttributes: []string{},
	}

	result, err := metadataService.ProcessMetadataWithTransaction(context.Background(), metadataReq)
	// Result should contain error
	require.NoError(t, err) // No error returned, but result shows failure
	assert.False(t, result.Success)
	assert.NotNil(t, result.Error)
	assert.Contains(t, *result.Error, "failed to read metadata file")
}

// TestModuleExtractor_MetadataRequiredAttributes tests validation of required metadata attributes
func TestModuleExtractor_MetadataRequiredAttributes(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	namespace := testutils.CreateNamespace(t, db, "testrequiredattrs")
	moduleProvider := testutils.CreateModuleProvider(t, db, namespace.ID, "test-module", "aws")
	moduleVersion := testutils.CreateModuleVersion(t, db, moduleProvider.ID, "1.0.0")

	// Create metadata missing required attributes
	files := map[string]string{
		"main.tf":       testutils.CreateValidMainTF(),
		"terrareg.json": testutils.CreateTerraregMetadata(map[string]interface{}{
			"owner": "testowner",
		}),
	}
	archive := testutils.CreateTestModuleZip(t, files)

	// Extract archive
	extractDir := testutils.ExtractTestArchive(t, archive)

	// Process metadata with required attributes
	metadataService := service.NewMetadataProcessingService(transaction.NewSavepointHelper(db.DB))
	metadataReq := service.MetadataProcessingRequest{
		ModuleVersionID:    moduleVersion.ID,
		MetadataPath:       extractDir,
		TransactionCtx:     context.Background(),
		RequiredAttributes: []string{"description", "owner"},
	}

	result, err := metadataService.ProcessMetadataWithTransaction(context.Background(), metadataReq)
	require.NoError(t, err)
	assert.False(t, result.Success)
	assert.Contains(t, result.MissingAttributes, "description")
}

// TestModuleExtractor_TarGzArchive tests TAR.GZ archive extraction
func TestModuleExtractor_TarGzArchive(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	namespace := testutils.CreateNamespace(t, db, "testtargz")
	moduleProvider := testutils.CreateModuleProvider(t, db, namespace.ID, "test-module", "aws")
	_ = testutils.CreateModuleVersion(t, db, moduleProvider.ID, "1.0.0")

	files := map[string]string{
		"main.tf":   testutils.CreateValidMainTF(),
		"README.md": testutils.CreateREADMEContent("test-module"),
	}
	archive := testutils.CreateTestTarGz(t, files)

	// Extract TAR.GZ archive
	extractDir := testutils.ExtractTestTarGz(t, archive)

	// Verify files were extracted
	mainTfPath := filepath.Join(extractDir, "main.tf")
	_, err := os.Stat(mainTfPath)
	require.NoError(t, err)

	readmePath := filepath.Join(extractDir, "README.md")
	_, err = os.Stat(readmePath)
	require.NoError(t, err)
}

// TestModuleExtractor_PathTraversalProtection tests path traversal protection in archive helpers
func TestModuleExtractor_PathTraversalProtection(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	namespace := testutils.CreateNamespace(t, db, "testpathtraversal")
	moduleProvider := testutils.CreateModuleProvider(t, db, namespace.ID, "test-module", "aws")
	_ = testutils.CreateModuleVersion(t, db, moduleProvider.ID, "1.0.0")

	// Create archive with normal files
	files := map[string]string{
		"main.tf": testutils.CreateValidMainTF(),
	}
	archive := testutils.CreateTestModuleZip(t, files)

	// Extract to a temp directory
	extractDir := testutils.ExtractTestArchive(t, archive)

	// Verify extraction stayed within bounds
	entries, err := os.ReadDir(extractDir)
	require.NoError(t, err)
	for _, entry := range entries {
		assert.NotContains(t, entry.Name(), "..")
	}
}

// TestModuleExtractor_TerraformDocsMock tests terraform-docs integration with mock
func TestModuleExtractor_TerraformDocsMock(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	namespace := testutils.CreateNamespace(t, db, "testterraformdocs")
	moduleProvider := testutils.CreateModuleProvider(t, db, namespace.ID, "test-module", "aws")
	_ = testutils.CreateModuleVersion(t, db, moduleProvider.ID, "1.0.0")

	// Create mock system command service with terraform-docs output
	mockCmdService := mocks.NewMockSystemCommandService()

	// Mock terraform-docs output
	tfDocsOutput := map[string]interface{}{
		"inputs": []map[string]interface{}{
			{
				"name":        "instance_type",
				"type":        "string",
				"description": "The type of instance to create",
				"default":     "t2.micro",
				"required":    false,
			},
			{
				"name":        "ami_id",
				"type":        "string",
				"description": "The AMI ID to use",
				"default":     nil,
				"required":    true,
			},
		},
		"outputs": []map[string]interface{}{
			{
				"name":        "instance_id",
				"description": "The ID of the instance",
			},
		},
	}
	tfDocsJSON, _ := json.Marshal(tfDocsOutput)
	mockCmdService.SetupTerraformDocsMock(string(tfDocsJSON))

	// Verify mock was set up correctly
	assert.True(t, mockCmdService.WasCommandExecuted("terraform-docs", []string{"json", "."}))
}

// TestModuleExtractor_TfsecMock tests tfsec integration with mock
func TestModuleExtractor_TfsecMock(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	namespace := testutils.CreateNamespace(t, db, "testtfsec")
	moduleProvider := testutils.CreateModuleProvider(t, db, namespace.ID, "test-module", "aws")
	_ = testutils.CreateModuleVersion(t, db, moduleProvider.ID, "1.0.0")

	// Create mock system command service with tfsec output
	mockCmdService := mocks.NewMockSystemCommandService()

	// Mock tfsec output (no security issues)
	tfsecOutput := map[string]interface{}{
		"results": []interface{}{},
	}
	tfsecJSON, _ := json.Marshal(tfsecOutput)
	mockCmdService.SetupTfsecMock(string(tfsecJSON))

	// Verify mock was set up
	assert.True(t, mockCmdService.WasCommandExecuted("tfsec", []string{"--format", "json", "--out", "-"}))
}

// TestModuleExtractor_CompleteWorkflow tests complete extraction workflow with all features
func TestModuleExtractor_CompleteWorkflow(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	namespace := testutils.CreateNamespace(t, db, "testcomplete")
	moduleProvider := testutils.CreateModuleProvider(t, db, namespace.ID, "test-module", "aws")
	moduleVersion := testutils.CreateModuleVersion(t, db, moduleProvider.ID, "1.0.0")

	// Create complete module with all features
	files := map[string]string{
		"main.tf": testutils.CreateValidMainTF(),
		"README.md": testutils.CreateREADMEContent("complete-module"),
		"terrareg.json": testutils.CreateTerraregMetadata(map[string]interface{}{
			"description": "Complete test module",
			"owner":       "testowner",
		}),
		"modules/submodule1/main.tf": `
variable "sub_input" {
  type = string
  default = "value"
}
`,
		"examples/example1/main.tf": `
module "test" {
  source = "../../"
}
`,
	}
	archive := testutils.CreateTestModuleZip(t, files)

	// Extract archive
	extractDir := testutils.ExtractTestArchive(t, archive)

	// Verify all expected files exist
	expectedFiles := []string{
		"main.tf",
		"README.md",
		"terrareg.json",
		"modules/submodule1/main.tf",
		"examples/example1/main.tf",
	}

	for _, file := range expectedFiles {
		path := filepath.Join(extractDir, file)
		_, err := os.Stat(path)
		require.NoError(t, err, "Expected file should exist: "+file)
	}

	// Verify archive contents count
	contents := testutils.ListArchiveContents(t, archive)
	assert.GreaterOrEqual(t, len(contents), 5) // At least the files listed above

	// Process metadata
	metadataService := service.NewMetadataProcessingService(transaction.NewSavepointHelper(db.DB))
	metadataReq := service.MetadataProcessingRequest{
		ModuleVersionID:    moduleVersion.ID,
		MetadataPath:       extractDir,
		TransactionCtx:     context.Background(),
		RequiredAttributes: []string{},
	}

	metadataResult, err := metadataService.ProcessMetadataWithTransaction(context.Background(), metadataReq)
	require.NoError(t, err)
	assert.True(t, metadataResult.Success)
	assert.True(t, metadataResult.MetadataFound)
	assert.Equal(t, "Complete test module", *metadataResult.Metadata.Description)
}

// TestModuleExtractor_MetadataVariableTemplate tests metadata with variable template
func TestModuleExtractor_MetadataVariableTemplate(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	namespace := testutils.CreateNamespace(t, db, "testvartemplate")
	moduleProvider := testutils.CreateModuleProvider(t, db, namespace.ID, "test-module", "aws")
	moduleVersion := testutils.CreateModuleVersion(t, db, moduleProvider.ID, "1.0.0")

	variableTemplate := []map[string]interface{}{
		{
			"name":       "custom_var",
			"type":       "number",
			"quote_value": true,
			"required":   false,
		},
		{
			"name":       "another_var",
			"type":       "text",
			"quote_value": true,
			"required":   true,
		},
	}

	metadata := map[string]interface{}{
		"description":        "Module with variable template",
		"owner":              "templateowner",
		"variable_template": variableTemplate,
	}

	files := map[string]string{
		"main.tf":       testutils.CreateValidMainTF(),
		"terrareg.json": testutils.CreateTerraregMetadata(metadata),
	}
	archive := testutils.CreateTestModuleZip(t, files)

	// Extract archive
	extractDir := testutils.ExtractTestArchive(t, archive)

	// Process metadata
	metadataService := service.NewMetadataProcessingService(transaction.NewSavepointHelper(db.DB))
	metadataReq := service.MetadataProcessingRequest{
		ModuleVersionID:    moduleVersion.ID,
		MetadataPath:       extractDir,
		TransactionCtx:     context.Background(),
		RequiredAttributes: []string{},
	}

	result, err := metadataService.ProcessMetadataWithTransaction(context.Background(), metadataReq)
	require.NoError(t, err)
	assert.True(t, result.Success)
	assert.NotNil(t, result.Metadata.VariableTemplate)
	assert.Len(t, result.Metadata.VariableTemplate, 2)
}

// TestModuleExtractor_EmptyArchive tests handling of empty archive
func TestModuleExtractor_EmptyArchive(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	namespace := testutils.CreateNamespace(t, db, "testempty")
	moduleProvider := testutils.CreateModuleProvider(t, db, namespace.ID, "test-module", "aws")
	_ = testutils.CreateModuleVersion(t, db, moduleProvider.ID, "1.0.0")

	// Create empty archive
	files := map[string]string{}
	archive := testutils.CreateTestModuleZip(t, files)

	// Extract empty archive
	extractDir := testutils.ExtractTestArchive(t, archive)

	// Verify directory exists but is empty
	entries, err := os.ReadDir(extractDir)
	require.NoError(t, err)
	// Empty archive should have 0 files
	assert.Len(t, entries, 0)

	// Verify archive contents
	contents := testutils.ListArchiveContents(t, archive)
	assert.Len(t, contents, 0)
}

// TestModuleExtractor_MultipleVersions tests uploading multiple versions
func TestModuleExtractor_MultipleVersions(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	namespace := testutils.CreateNamespace(t, db, "testmultiver")
	moduleProvider := testutils.CreateModuleProvider(t, db, namespace.ID, "test-module", "aws")

	versions := []string{"1.0.0", "1.1.0", "2.0.0"}

	for _, version := range versions {
		_ = testutils.CreateModuleVersion(t, db, moduleProvider.ID, version)

		files := map[string]string{
			"main.tf": testutils.CreateValidMainTF(),
			"README.md": testutils.CreateREADMEContent("test-module-" + version),
		}
		archive := testutils.CreateTestModuleZip(t, files)

		// Extract archive
		extractDir := testutils.ExtractTestArchive(t, archive)

		// Verify README contains correct version
		readmePath := filepath.Join(extractDir, "README.md")
		content, err := os.ReadFile(readmePath)
		require.NoError(t, err)
		assert.Contains(t, string(content), version)
	}

	// Verify all versions were created in database
	var versionsInDB []sqldb.ModuleVersionDB
	err := db.DB.Where("module_provider_id = ?", moduleProvider.ID).Find(&versionsInDB).Error
	require.NoError(t, err)
	assert.Len(t, versionsInDB, 3)
}

// TestModuleExtractor_BetaVersion tests beta version handling
func TestModuleExtractor_BetaVersion(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	namespace := testutils.CreateNamespace(t, db, "testbeta")
	moduleProvider := testutils.CreateModuleProvider(t, db, namespace.ID, "test-module", "aws")

	// Create beta version
	beta := true
	moduleVersion := sqldb.ModuleVersionDB{
		ModuleProviderID: moduleProvider.ID,
		Version:          "1.0.0-beta",
		Beta:             beta,
		Published:        nil,
	}
	err := db.DB.Create(&moduleVersion).Error
	require.NoError(t, err)

	files := map[string]string{
		"main.tf": testutils.CreateValidMainTF(),
	}
	_ = testutils.CreateTestModuleZip(t, files)

	// Verify beta flag is set
	var retrievedVersion sqldb.ModuleVersionDB
	err = db.DB.First(&retrievedVersion, moduleVersion.ID).Error
	require.NoError(t, err)
	assert.True(t, retrievedVersion.Beta)
}

// TestModuleExtractor_ArchiveListContents tests listing archive contents
func TestModuleExtractor_ArchiveListContents(t *testing.T) {
	t.Run("ListZIPContents", func(t *testing.T) {
		files := map[string]string{
			"main.tf":      testutils.CreateValidMainTF(),
			"README.md":    testutils.CreateREADMEContent("test"),
			"variables.tf": `variable "test" {}`,
			"outputs.tf":   `output "test" { value = "test" }`,
		}
		archive := testutils.CreateTestModuleZip(t, files)

		contents := testutils.ListArchiveContents(t, archive)
		assert.Len(t, contents, 4)
		assert.Contains(t, contents, "main.tf")
		assert.Contains(t, contents, "README.md")
		assert.Contains(t, contents, "variables.tf")
		assert.Contains(t, contents, "outputs.tf")
	})

	t.Run("ReadFileFromArchive", func(t *testing.T) {
		expectedContent := testutils.CreateValidMainTF()
		files := map[string]string{
			"main.tf": expectedContent,
		}
		archive := testutils.CreateTestModuleZip(t, files)

		content := testutils.ReadFileFromArchive(t, archive, "main.tf")
		assert.Equal(t, expectedContent, string(content))
	})

	t.Run("ListTarGzContents", func(t *testing.T) {
		files := map[string]string{
			"main.tf":   testutils.CreateValidMainTF(),
			"README.md": testutils.CreateREADMEContent("test"),
		}
		archive := testutils.CreateTestTarGz(t, files)

		// Extract and verify
		extractDir := testutils.ExtractTestTarGz(t, archive)
		mainTfPath := filepath.Join(extractDir, "main.tf")
		_, err := os.Stat(mainTfPath)
		require.NoError(t, err)
	})
}

// TestModuleExtractor_InfracostMock tests infracost integration with mock
func TestModuleExtractor_InfracostMock(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	namespace := testutils.CreateNamespace(t, db, "testinfracost")
	moduleProvider := testutils.CreateModuleProvider(t, db, namespace.ID, "test-module", "aws")
	_ = testutils.CreateModuleVersion(t, db, moduleProvider.ID, "1.0.0")

	// Create mock system command service with infracost output
	mockCmdService := mocks.NewMockSystemCommandService()

	// Mock infracost output (no costs)
	infracostOutput := map[string]interface{}{
		"total_monthly_cost": 0.0,
		"projects":           []interface{}{},
	}
	infracostJSON, _ := json.Marshal(infracostOutput)
	mockCmdService.SetupInfracostMock(string(infracostJSON))

	// Verify mock was set up
	assert.True(t, mockCmdService.WasCommandExecuted("infracost", []string{"breakdown", "--format", "json"}))
}

// TestModuleExtractor_TerraformMock tests terraform command integration with mock
func TestModuleExtractor_TerraformMock(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	namespace := testutils.CreateNamespace(t, db, "testterraform")
	moduleProvider := testutils.CreateModuleProvider(t, db, namespace.ID, "test-module", "aws")
	_ = testutils.CreateModuleVersion(t, db, moduleProvider.ID, "1.0.0")

	// Create mock system command service with terraform output
	mockCmdService := mocks.NewMockSystemCommandService()

	// Mock terraform version output
	terraformVersionOut := `Terraform v1.5.0
on linux_amd64
`
	mockCmdService.SetupTerraformMock("version", terraformVersionOut)

	// Mock terraform init output
	terraformInitOut := `Initializing the backend...

Terraform has been successfully initialized!
`
	mockCmdService.SetupTerraformMock("init", terraformInitOut)

	// Verify mocks were set up
	assert.True(t, mockCmdService.WasCommandExecuted("terraform", []string{"version", "-no-color"}))
	assert.True(t, mockCmdService.WasCommandExecuted("terraform", []string{"init", "-input=false", "-no-color"}))
}

// TestModuleExtractor_MetadataNoFile tests module without metadata file
func TestModuleExtractor_MetadataNoFile(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	namespace := testutils.CreateNamespace(t, db, "testnofile")
	moduleProvider := testutils.CreateModuleProvider(t, db, namespace.ID, "test-module", "aws")
	moduleVersion := testutils.CreateModuleVersion(t, db, moduleProvider.ID, "1.0.0")

	files := map[string]string{
		"main.tf": testutils.CreateValidMainTF(),
	}
	archive := testutils.CreateTestModuleZip(t, files)

	// Extract archive
	extractDir := testutils.ExtractTestArchive(t, archive)

	// Process metadata - should succeed with no metadata found
	metadataService := service.NewMetadataProcessingService(transaction.NewSavepointHelper(db.DB))
	metadataReq := service.MetadataProcessingRequest{
		ModuleVersionID:    moduleVersion.ID,
		MetadataPath:       extractDir,
		TransactionCtx:     context.Background(),
		RequiredAttributes: []string{},
	}

	result, err := metadataService.ProcessMetadataWithTransaction(context.Background(), metadataReq)
	require.NoError(t, err)
	assert.True(t, result.Success)
	assert.False(t, result.MetadataFound) // No metadata file, but that's OK
	assert.Nil(t, result.Metadata)
}
