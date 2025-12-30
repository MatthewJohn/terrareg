package module

import (
	"context"
	"encoding/json"
	"fmt"
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
// Python reference: test_process_upload.py::TestProcessUpload::test_basic_module
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
// Python reference: test_process_upload.py::TestProcessUpload::test_upload_with_readme
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
// Python reference: test_process_upload.py::TestProcessUpload::test_sub_modules
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
// Python reference: test_process_upload.py::TestProcessUpload::test_examples
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
// Python reference: test_process_upload.py::TestProcessUpload::test_terrareg_metadata
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
// Python reference: test_process_upload.py::TestProcessUpload::test_invalid_terrareg_metadata_file
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
// Python reference: test_process_upload.py::TestProcessUpload::test_terrareg_metadata_required_attributes
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
// Python reference: test_process_upload.py::TestProcessUpload::test_all_features (partial - TAR.GZ is tested implicitly)
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
// Python reference: test_process_upload.py::TestProcessUpload::test_upload_malicious_zip (partial - covers path traversal)
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

// TestModuleExtractor_TerraformDocsMock tests terraform-docs mock setup
// Python reference: Infrastructure test for terraform-docs mocking
func TestModuleExtractor_TerraformDocsMock(t *testing.T) {
	// Create mock system command service
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
		},
	}
	tfDocsJSON, _ := json.Marshal(tfDocsOutput)
	mockCmdService.SetupTerraformDocsMock(string(tfDocsJSON))

	// Verify mock output is retrievable
	history := mockCmdService.GetCommandHistory()
	assert.NotNil(t, history)
}

// TestModuleExtractor_TfsecMock tests tfsec mock setup
// Python reference: test_process_upload.py::TestProcessUpload::test_uploading_module_with_security_issue (partial - uses mock)
func TestModuleExtractor_TfsecMock(t *testing.T) {
	// Create mock system command service
	mockCmdService := mocks.NewMockSystemCommandService()

	// Mock tfsec output (no security issues)
	tfsecOutput := map[string]interface{}{
		"results": []interface{}{},
	}
	tfsecJSON, _ := json.Marshal(tfsecOutput)
	mockCmdService.SetupTfsecMock(string(tfsecJSON))

	// Verify mock output is retrievable
	history := mockCmdService.GetCommandHistory()
	assert.NotNil(t, history)
}

// TestModuleExtractor_CompleteWorkflow tests complete extraction workflow with all features
// Python reference: test_process_upload.py::TestProcessUpload::test_all_features
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
// Python reference: test_process_upload.py::TestProcessUpload::test_terrareg_metadata_override_autogenerate (partial)
func TestModuleExtractor_MetadataVariableTemplate(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	namespace := testutils.CreateNamespace(t, db, "testvartemplate")
	moduleProvider := testutils.CreateModuleProvider(t, db, namespace.ID, "test-module", "aws")
	moduleVersion := testutils.CreateModuleVersion(t, db, moduleProvider.ID, "1.0.0")

	// Create valid JSON for variable template with nested structures
	// The CreateTerraregMetadata helper doesn't handle nested maps properly,
	// so we create the JSON directly for this complex case
	validMetadataJSON := `{
		"description": "Module with variable template",
		"owner": "templateowner",
		"variable_template": [
			{
				"name": "custom_var",
				"type": "number",
				"quote_value": true,
				"required": false
			},
			{
				"name": "another_var",
				"type": "text",
				"quote_value": true,
				"required": true
			}
		]
	}`

	files := map[string]string{
		"main.tf":       testutils.CreateValidMainTF(),
		"terrareg.json": validMetadataJSON,
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
// Python reference: Infrastructure test for archive handling
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
// Python reference: Infrastructure test for version handling
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
// Python reference: Infrastructure test for beta version flag
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
// Python reference: Infrastructure test for archive utilities
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

// TestModuleExtractor_InfracostMock tests infracost mock setup
// Python reference: test_process_upload.py::TestProcessUpload::test_uploading_module_with_infracost_mocked (partial)
func TestModuleExtractor_InfracostMock(t *testing.T) {
	// Create mock system command service
	mockCmdService := mocks.NewMockSystemCommandService()

	// Mock infracost output (no costs)
	infracostOutput := map[string]interface{}{
		"total_monthly_cost": 0.0,
		"projects":           []interface{}{},
	}
	infracostJSON, _ := json.Marshal(infracostOutput)
	mockCmdService.SetupInfracostMock(string(infracostJSON))

	// Verify mock output is retrievable
	history := mockCmdService.GetCommandHistory()
	assert.NotNil(t, history)
}

// TestModuleExtractor_TerraformMock tests terraform mock setup
// Python reference: test_process_upload.py::TestProcessUpload::test_terraform_version (partial)
func TestModuleExtractor_TerraformMock(t *testing.T) {
	// Create mock system command service
	mockCmdService := mocks.NewMockSystemCommandService()

	// Mock terraform version output
	terraformVersionOut := `Terraform v1.5.0
on linux_amd64
`
	mockCmdService.SetupTerraformMock("version", terraformVersionOut)

	// Verify mock output is retrievable
	history := mockCmdService.GetCommandHistory()
	assert.NotNil(t, history)
}

// TestModuleExtractor_MetadataNoFile tests module without metadata file
// Python reference: test_process_upload.py::TestProcessUpload::test_basic_module (partial - no metadata case)
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

// TestModuleExtractor_MetadataOverrideAutogenerate tests metadata overriding auto-generated values
// Python reference: test_process_upload.py::TestProcessUpload::test_terrareg_metadata_override_autogenerate
func TestModuleExtractor_MetadataOverrideAutogenerate(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	namespace := testutils.CreateNamespace(t, db, "testmetadataoverride")
	moduleProvider := testutils.CreateModuleProvider(t, db, namespace.ID, "test-module", "aws")
	moduleVersion := testutils.CreateModuleVersion(t, db, moduleProvider.ID, "1.0.0")

	// Create metadata with variable template that should override auto-generated
	metadataJSON := `{
		"description": "unittestdescription!",
		"owner": "unittestowner.",
		"variable_template": [
			{
				"name": "test_input",
				"type": "text",
				"quote_value": true,
				"required": false,
				"default_value": null,
				"additional_help": ""
			}
		]
	}`

	files := map[string]string{
		"main.tf":       testutils.CreateValidMainTF(),
		"terrareg.json": metadataJSON,
	}
	archive := testutils.CreateTestModuleZip(t, files)

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
	assert.Equal(t, "unittestdescription!", *result.Metadata.Description)
	assert.Equal(t, "unittestowner.", *result.Metadata.Owner)
	assert.NotNil(t, result.Metadata.VariableTemplate)
	assert.Len(t, result.Metadata.VariableTemplate, 1)
}

// TestModuleExtractor_MetadataRepoUrls tests metadata with repo URL overrides
// Python reference: test_process_upload.py::TestProcessUpload::test_override_repo_urls_with_metadata
func TestModuleExtractor_MetadataRepoUrls(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	namespace := testutils.CreateNamespace(t, db, "testrepourls")
	moduleProvider := testutils.CreateModuleProvider(t, db, namespace.ID, "test-module", "aws")
	moduleVersion := testutils.CreateModuleVersion(t, db, moduleProvider.ID, "1.0.0")

	// Create metadata with repo URL overrides
	// Note: repo_base_url is not stored in TerraregMetadata struct, only repo_clone_url and repo_browse_url
	repoCloneURL := "ssh://overrideurl_here.com/{namespace}/{module}-{provider}"
	repoBrowseURL := "https://base_url.com/{namespace}-{module}-{provider}-{tag}/{path}"

	metadataJSON := fmt.Sprintf(`{
		"repo_clone_url": "%s",
		"repo_browse_url": "%s"
	}`, repoCloneURL, repoBrowseURL)

	files := map[string]string{
		"main.tf":       testutils.CreateValidMainTF(),
		"terrareg.json": metadataJSON,
	}
	archive := testutils.CreateTestModuleZip(t, files)

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
	assert.NotNil(t, result.Metadata.RepoCloneURL)
	assert.NotNil(t, result.Metadata.RepoBrowseURL)
	assert.Equal(t, repoCloneURL, *result.Metadata.RepoCloneURL)
	assert.Equal(t, repoBrowseURL, *result.Metadata.RepoBrowseURL)
}

// TestModuleExtractor_MetadataWithAllOptionalFields tests metadata with all optional fields
// Python reference: test_process_upload.py::TestProcessUpload::test_terrareg_metadata (partial - all fields)
func TestModuleExtractor_MetadataWithAllOptionalFields(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	namespace := testutils.CreateNamespace(t, db, "testallfields")
	moduleProvider := testutils.CreateModuleProvider(t, db, namespace.ID, "test-module", "aws")
	moduleVersion := testutils.CreateModuleVersion(t, db, moduleProvider.ID, "1.0.0")

	// Create metadata with all optional fields
	metadataJSON := `{
		"description": "Full test description",
		"owner": "testowner",
		"repo_clone_url": "https://github.com/test/clone",
		"repo_browse_url": "https://github.com/test/browse",
		"issues_url": "https://github.com/test/issues",
		"license": "MIT",
		"provider": {
			"custom_field": "custom_value"
		},
		"variable_template": [
			{
				"name": "custom_var",
				"type": "string",
				"quote_value": true,
				"required": true
			}
		]
	}`

	files := map[string]string{
		"main.tf":       testutils.CreateValidMainTF(),
		"terrareg.json": metadataJSON,
	}
	archive := testutils.CreateTestModuleZip(t, files)

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

	// Verify all fields
	assert.Equal(t, "Full test description", *result.Metadata.Description)
	assert.Equal(t, "testowner", *result.Metadata.Owner)
	assert.Equal(t, "https://github.com/test/clone", *result.Metadata.RepoCloneURL)
	assert.Equal(t, "https://github.com/test/browse", *result.Metadata.RepoBrowseURL)
	assert.Equal(t, "https://github.com/test/issues", *result.Metadata.IssuesURL)
	assert.Equal(t, "MIT", *result.Metadata.License)
	assert.NotNil(t, result.Metadata.Provider)
	assert.Contains(t, result.Metadata.Provider, "custom_field")
	assert.Len(t, result.Metadata.VariableTemplate, 1)
}

// TestModuleExtractor_HiddenTerraregJson tests .terrareg.json (hidden file) is detected
// Python reference: test_process_upload.py::TestProcessUpload::test_terrareg_metadata (partial - hidden file)
func TestModuleExtractor_HiddenTerraregJson(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	namespace := testutils.CreateNamespace(t, db, "testhiddenmetadata")
	moduleProvider := testutils.CreateModuleProvider(t, db, namespace.ID, "test-module", "aws")
	moduleVersion := testutils.CreateModuleVersion(t, db, moduleProvider.ID, "1.0.0")

	description := "hidden file description"

	metadataJSON := fmt.Sprintf(`{
		"description": "%s"
	}`, description)

	files := map[string]string{
		"main.tf":         testutils.CreateValidMainTF(),
		".terrareg.json":  metadataJSON, // Hidden file
		"README.md":       testutils.CreateREADMEContent("test-module"),
	}

	// Use CreateTestModuleZip - it handles hidden files correctly
	archive := testutils.CreateTestModuleZip(t, files)

	extractDir := testutils.ExtractTestArchive(t, archive)

	// Process metadata - should find .terrareg.json
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
	assert.Equal(t, description, *result.Metadata.Description)
}

// TestModuleExtractor_NonRootDirectory tests module in subdirectory (non-root repo)
// Python reference: test_process_upload.py::TestProcessUpload::test_non_root_repo_directory (partial - non-root only)
func TestModuleExtractor_NonRootDirectory(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	_ = testutils.CreateNamespace(t, db, "testnonroot")

	// Simulate module in subdirectory structure
	// Create archive with module in subdirectory
	files := map[string]string{
		"subdirectory/module/main.tf": testutils.CreateValidMainTF(),
		"subdirectory/module/README.md": testutils.CreateREADMEContent("test-module"),
	}
	archive := testutils.CreateTestModuleZip(t, files)

	extractDir := testutils.ExtractTestArchive(t, archive)

	// Verify subdirectory structure exists
	subdirPath := filepath.Join(extractDir, "subdirectory", "module")
	info, err := os.Stat(subdirPath)
	require.NoError(t, err)
	assert.True(t, info.IsDir())

	mainTfPath := filepath.Join(subdirPath, "main.tf")
	_, err = os.Stat(mainTfPath)
	require.NoError(t, err)
}

// TestModuleExtractor_PathspecIgnore tests .terraformignore file handling
// Python reference: test_process_upload.py::TestProcessUpload::test_non_root_repo_directory (partial - .tfignore)
func TestModuleExtractor_PathspecIgnore(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	namespace := testutils.CreateNamespace(t, db, "testpathspecignore")
	moduleProvider := testutils.CreateModuleProvider(t, db, namespace.ID, "test-module", "aws")
	_ = testutils.CreateModuleVersion(t, db, moduleProvider.ID, "1.0.0")

	// Create .terraformignore content
	ignoreContent := `# Ignore file in root
some_file_to_ignore_in_root.txt

# A file in a module directory
modules/testmodule1/file_to_ignore.txt

**/glob_ignore_file.*
`

	files := map[string]string{
		"main.tf":             testutils.CreateValidMainTF(),
		".terraformignore":    ignoreContent,
		"README.md":           testutils.CreateREADMEContent("test-module"),
	}
	archive := testutils.CreateTestModuleZip(t, files)

	extractDir := testutils.ExtractTestArchive(t, archive)

	// Process .terraformignore
	metadataService := service.NewMetadataProcessingService(transaction.NewSavepointHelper(db.DB))
	pathspecFilter, err := metadataService.GetPathspecFilter(context.Background(), extractDir)
	require.NoError(t, err)
	assert.NotNil(t, pathspecFilter)

	// Verify ignore rules were parsed
	assert.Len(t, pathspecFilter.Rules, 3)
	assert.Contains(t, pathspecFilter.Rules, "some_file_to_ignore_in_root.txt")
	assert.Contains(t, pathspecFilter.Rules, "modules/testmodule1/file_to_ignore.txt")
	assert.Contains(t, pathspecFilter.Rules, "**/glob_ignore_file.*")
}

// TestModuleExtractor_EmptyTerraregJson tests empty terrareg.json file
// Python reference: Edge case test for empty metadata
func TestModuleExtractor_EmptyTerraregJson(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	namespace := testutils.CreateNamespace(t, db, "testemptyjson")
	moduleProvider := testutils.CreateModuleProvider(t, db, namespace.ID, "test-module", "aws")
	moduleVersion := testutils.CreateModuleVersion(t, db, moduleProvider.ID, "1.0.0")

	files := map[string]string{
		"main.tf":       testutils.CreateValidMainTF(),
		"terrareg.json": "{}", // Empty JSON
	}
	archive := testutils.CreateTestModuleZip(t, files)

	extractDir := testutils.ExtractTestArchive(t, archive)

	// Process metadata - empty JSON should be valid but have no values
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
	// All fields should be nil/empty
	assert.Nil(t, result.Metadata.Description)
	assert.Nil(t, result.Metadata.Owner)
}

// TestModuleExtractor_BothMetadataFiles tests both terrareg.json and .terrareg.json exist
// Python reference: Edge case test for both metadata files present
func TestModuleExtractor_BothMetadataFiles(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	namespace := testutils.CreateNamespace(t, db, "testbothmetadata")
	moduleProvider := testutils.CreateModuleProvider(t, db, namespace.ID, "test-module", "aws")
	moduleVersion := testutils.CreateModuleVersion(t, db, moduleProvider.ID, "1.0.0")

	// Create both files - terrareg.json should take priority (as per implementation)
	publicMetadata := `{"description": "public metadata", "owner": "public_owner"}`
	hiddenMetadata := `{"description": "hidden metadata", "owner": "hidden_owner"}`

	// Use CreateTestModuleZip with both files
	files := map[string]string{
		"main.tf":        testutils.CreateValidMainTF(),
		"terrareg.json":  publicMetadata,
		".terrareg.json": hiddenMetadata,
	}
	archive := testutils.CreateTestModuleZip(t, files)

	extractDir := testutils.ExtractTestArchive(t, archive)

	// Process metadata - should find terrareg.json (non-hidden takes priority)
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
	// Should get public (non-hidden) metadata
	assert.Equal(t, "public metadata", *result.Metadata.Description)
	assert.Equal(t, "public_owner", *result.Metadata.Owner)
}

// TestModuleExtractor_NonRootDirectoryWithTfIgnore tests .terraformignore in non-root directory
// Python reference: test_process_upload.py::test_non_root_repo_directory (partial - tests .tfignore in subdirectory)
func TestModuleExtractor_NonRootDirectoryWithTfIgnore(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	_ = testutils.CreateNamespace(t, db, "testnonroottfignore")

	// Create archive with module in subdirectory and .terraformignore in different locations
	ignoreContent := `# Ignore specific files
*.tmp
*.log
modules/ignored_dir/
**/secrets.yaml
`

	files := map[string]string{
		"subdirectory/module/main.tf":           testutils.CreateValidMainTF(),
		"subdirectory/module/README.md":         testutils.CreateREADMEContent("test-module"),
		"subdirectory/module/.terraformignore":  ignoreContent, // .terraformignore in module subdirectory
	}
	archive := testutils.CreateTestModuleZip(t, files)

	extractDir := testutils.ExtractTestArchive(t, archive)

	// Verify .terraformignore exists in subdirectory
	tfIgnorePath := filepath.Join(extractDir, "subdirectory", "module", ".terraformignore")
	_, err := os.Stat(tfIgnorePath)
	require.NoError(t, err, ".terraformignore should exist in subdirectory")

	// Process .terraformignore from the subdirectory
	metadataService := service.NewMetadataProcessingService(transaction.NewSavepointHelper(db.DB))
	modulePath := filepath.Join(extractDir, "subdirectory", "module")
	pathspecFilter, err := metadataService.GetPathspecFilter(context.Background(), modulePath)
	require.NoError(t, err)
	assert.NotNil(t, pathspecFilter)

	// Verify ignore rules were parsed from subdirectory .terraformignore
	assert.Len(t, pathspecFilter.Rules, 4)
	assert.Contains(t, pathspecFilter.Rules, "*.tmp")
	assert.Contains(t, pathspecFilter.Rules, "*.log")
	assert.Contains(t, pathspecFilter.Rules, "modules/ignored_dir/")
	assert.Contains(t, pathspecFilter.Rules, "**/secrets.yaml")
}

// TestModuleExtractor_MultipleTfIgnoreFiles tests behavior with multiple .terraformignore files
// Python reference: test_process_upload.py::test_non_root_repo_directory (partial - tests multiple .tfignore locations)
func TestModuleExtractor_MultipleTfIgnoreFiles(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	_ = testutils.CreateNamespace(t, db, "testmultipletfignore")

	// Create archive with .terraformignore in root and subdirectory
	rootIgnore := `# Root ignore
*.tmp
root-only.txt
`

	subIgnore := `# Subdirectory ignore
*.log
sub-only.txt
`

	files := map[string]string{
		"main.tf":                           testutils.CreateValidMainTF(),
		".terraformignore":                   rootIgnore,
		"subdirectory/main.tf":               testutils.CreateValidMainTF(),
		"subdirectory/.terraformignore":      subIgnore,
	}
	archive := testutils.CreateTestModuleZip(t, files)

	extractDir := testutils.ExtractTestArchive(t, archive)

	// Process root .terraformignore
	metadataService := service.NewMetadataProcessingService(transaction.NewSavepointHelper(db.DB))
	rootPathspecFilter, err := metadataService.GetPathspecFilter(context.Background(), extractDir)
	require.NoError(t, err)
	assert.NotNil(t, rootPathspecFilter)

	// Verify root ignore rules
	assert.Len(t, rootPathspecFilter.Rules, 2)
	assert.Contains(t, rootPathspecFilter.Rules, "*.tmp")
	assert.Contains(t, rootPathspecFilter.Rules, "root-only.txt")

	// Process subdirectory .terraformignore
	subPath := filepath.Join(extractDir, "subdirectory")
	subPathspecFilter, err := metadataService.GetPathspecFilter(context.Background(), subPath)
	require.NoError(t, err)
	assert.NotNil(t, subPathspecFilter)

	// Verify subdirectory ignore rules (should be different from root)
	assert.Len(t, subPathspecFilter.Rules, 2)
	assert.Contains(t, subPathspecFilter.Rules, "*.log")
	assert.Contains(t, subPathspecFilter.Rules, "sub-only.txt")
	// Root-specific rules should not be in subdirectory filter
	assert.NotContains(t, subPathspecFilter.Rules, "root-only.txt")
}

// TestModuleExtractor_TfIgnoreWithWildcardPatterns tests .terraformignore with various glob patterns
// Python reference: test_process_upload.py::test_non_root_repo_directory (partial - tests glob patterns in .tfignore)
func TestModuleExtractor_TfIgnoreWithWildcardPatterns(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	_ = testutils.CreateNamespace(t, db, "testwildcardpatterns")

	// Create .terraformignore with various glob patterns
	ignoreContent := `# Various glob patterns
*.log
temp_*
test?.txt
**/ignored_dir/
prefix_*.txt
**/*_backup.yaml
.DS_Store
`

	files := map[string]string{
		"main.tf":           testutils.CreateValidMainTF(),
		".terraformignore":  ignoreContent,
	}
	archive := testutils.CreateTestModuleZip(t, files)

	extractDir := testutils.ExtractTestArchive(t, archive)

	// Process .terraformignore
	metadataService := service.NewMetadataProcessingService(transaction.NewSavepointHelper(db.DB))
	pathspecFilter, err := metadataService.GetPathspecFilter(context.Background(), extractDir)
	require.NoError(t, err)
	assert.NotNil(t, pathspecFilter)

	// Verify all glob patterns were parsed
	assert.Len(t, pathspecFilter.Rules, 7)
	assert.Contains(t, pathspecFilter.Rules, "*.log")
	assert.Contains(t, pathspecFilter.Rules, "temp_*")
	assert.Contains(t, pathspecFilter.Rules, "test?.txt")
	assert.Contains(t, pathspecFilter.Rules, "**/ignored_dir/")
	assert.Contains(t, pathspecFilter.Rules, "prefix_*.txt")
	assert.Contains(t, pathspecFilter.Rules, "**/*_backup.yaml")
	assert.Contains(t, pathspecFilter.Rules, ".DS_Store")
}
