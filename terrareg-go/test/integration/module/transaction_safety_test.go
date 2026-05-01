package module

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/module/service"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/container"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/logging"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/persistence/sqldb"
	"github.com/matthewjohn/terrareg/terrareg-go/test/integration/testutils"
)

// TestTransactionSafety_ModuleProcessingWithSubmodules tests that all database operations
// within the module processing flow use the transaction context correctly.
// This test would have caught the "database is locked" bug where SubmoduleLoader was using
// the base DB instead of the transaction-aware DB.
func TestTransactionSafety_ModuleProcessingWithSubmodules(t *testing.T) {
	// Use file-based SQLite (not in-memory) to expose locking issues
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")
	db, err := sqldb.NewDatabase("sqlite://"+dbPath, true)
	require.NoError(t, err)
	defer func() {
		testutils.CleanupTestDatabase(t, db)
		os.Remove(dbPath)
	}()

	// Run auto-migration
	err = db.DB.AutoMigrate(
		&sqldb.NamespaceDB{},
		&sqldb.ModuleProviderDB{},
		&sqldb.ModuleVersionDB{},
		&sqldb.ModuleDetailsDB{},
		&sqldb.SubmoduleDB{},
		&sqldb.ExampleFileDB{},
	)
	require.NoError(t, err)

	// Create a test git repository with submodules
	repoDir := t.TempDir()
	setupTestGitRepoWithSubmodules(t, repoDir)

	// Create container with all dependencies
	domainConfig := testutils.CreateTestDomainConfig(t)
	infraConfig := testutils.CreateTestInfraConfig(t)
	logger := logging.NewTestLogger(t)

	cont, err := container.NewContainer(domainConfig, infraConfig, nil, logger, db)
	require.NoError(t, err)

	// Create namespace, module provider, and version
	namespace := testutils.CreateNamespace(t, db, "test-namespace", nil)
	moduleProvider := testutils.CreateModuleProvider(t, db, namespace.ID, "test-module", "aws")
	moduleVersion := testutils.CreateModuleVersion(t, db, moduleProvider.ID, "1.0.0")

	// Mock security scanning to avoid external dependencies
	// (Security scanning is handled by the mock system command service in tests)

	ctx := context.Background()

	// Process the module - this should NOT cause "database is locked" errors
	// The bug was that LoadSubmodulesAndExamples was called from within a transaction
	// but used the base DB instead of the transaction-aware DB
	metadata := &service.ModuleProcessingMetadata{
		ModuleVersionID: moduleVersion.ID,
		GitTag:          "v1.0.0",
	}
	result, err := cont.ModuleProcessorService.ProcessModule(ctx, repoDir, metadata)

	// This should succeed without database lock errors
	assert.NoError(t, err, "Module processing should complete without database locks")
	assert.NotNil(t, result, "ProcessModule result should not be nil")

	// Verify the module version was properly processed
	updatedVersion, err := cont.ModuleVersionRepo.FindByID(ctx, moduleVersion.ID)
	require.NoError(t, err)
	assert.NotNil(t, updatedVersion)

	// Verify submodules were loaded within the transaction
	// This is the critical check - if SubmoduleLoader is not using transaction context,
	// this would have failed with "database is locked"
	var submodulesDB []sqldb.SubmoduleDB
	err = db.DB.Where("parent_module_version = ?", moduleVersion.ID).Find(&submodulesDB).Error
	require.NoError(t, err)
	assert.Greater(t, len(submodulesDB), 0, "Submodules should be indexed")
}

// TestTransactionSafety_WithSecurityScanning tests that security scanning
// operations use the transaction context correctly
func TestTransactionSafety_WithSecurityScanning(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")
	db, err := sqldb.NewDatabase("sqlite://"+dbPath, true)
	require.NoError(t, err)
	defer func() {
		testutils.CleanupTestDatabase(t, db)
		os.Remove(dbPath)
	}()

	// Run auto-migration
	err = db.DB.AutoMigrate(
		&sqldb.NamespaceDB{},
		&sqldb.ModuleProviderDB{},
		&sqldb.ModuleVersionDB{},
		&sqldb.ModuleDetailsDB{},
		&sqldb.SubmoduleDB{},
	)
	require.NoError(t, err)

	// Create a test git repository
	repoDir := t.TempDir()
	setupTestGitRepo(t, repoDir)

	// Create container
	domainConfig := testutils.CreateTestDomainConfig(t)
	infraConfig := testutils.CreateTestInfraConfig(t)
	logger := logging.NewTestLogger(t)

	cont, err := container.NewContainer(domainConfig, infraConfig, nil, logger, db)
	require.NoError(t, err)

	// Create test data
	namespace := testutils.CreateNamespace(t, db, "test-namespace", nil)
	moduleProvider := testutils.CreateModuleProvider(t, db, namespace.ID, "test-module", "aws")
	moduleVersion := testutils.CreateModuleVersion(t, db, moduleProvider.ID, "1.0.0")

	// Mock security scanning
	// (Security scanning is handled by the mock system command service in tests)

	ctx := context.Background()

	// Process module with security scanning
	// The bug was that security scanning calls FindByID -> mapToDomainModel -> LoadSubmodulesAndExamples
	// and if LoadSubmodulesAndExamples doesn't use transaction context, it will lock
	metadata := &service.ModuleProcessingMetadata{
		ModuleVersionID: moduleVersion.ID,
		GitTag:          "v1.0.0",
	}
	_, err = cont.ModuleProcessorService.ProcessModule(ctx, repoDir, metadata)

	assert.NoError(t, err, "Module processing with security scan should complete without locks")

	// Verify module details were created
	var detailsDB sqldb.ModuleDetailsDB
	err = db.DB.Where("id = (SELECT module_details_id FROM module_version WHERE id = ?)", moduleVersion.ID).First(&detailsDB).Error
	require.NoError(t, err)
	assert.NotNil(t, detailsDB)
}

// setupTestGitRepo creates a minimal git repository for testing
func setupTestGitRepo(t *testing.T, dir string) {
	t.Helper()

	// Initialize git repo
	runCommand(t, dir, "git", "init", "-b", "main")
	runCommand(t, dir, "git", "config", "user.email", "test@example.com")
	runCommand(t, dir, "git", "config", "user.name", "Test User")
	runCommand(t, dir, "git", "config", "commit.gpgsign", "false")

	// Create main.tf
	mainTF := `
variable "instance_type" {
  description = "EC2 instance type"
  type        = string
  default     = "t2.micro"
}

resource "aws_instance" "example" {
  ami           = "ami-12345678"
  instance_type = var.instance_type
}

output "instance_id" {
  description = "ID of the EC2 instance"
  value       = aws_instance.example.id
}
`
	err := os.WriteFile(filepath.Join(dir, "main.tf"), []byte(mainTF), 0644)
	require.NoError(t, err)

	// Create README.md
	readme := `# Test Module

This is a test module for transaction safety testing.
It demonstrates proper module processing with database transactions.
`
	err = os.WriteFile(filepath.Join(dir, "README.md"), []byte(readme), 0644)
	require.NoError(t, err)

	// Commit files
	runCommand(t, dir, "git", "add", ".")
	runCommand(t, dir, "git", "commit", "-m", "Initial commit")
	runCommand(t, dir, "git", "tag", "v1.0.0")
}

// setupTestGitRepoWithSubmodules creates a git repository with submodules
func setupTestGitRepoWithSubmodules(t *testing.T, dir string) {
	t.Helper()

	// First create base repo
	setupTestGitRepo(t, dir)

	// Create modules directory
	modulesDir := filepath.Join(dir, "modules")
	err := os.MkdirAll(modulesDir, 0755)
	require.NoError(t, err)

	// Create a submodule
	submoduleDir := filepath.Join(modulesDir, "submodule1")
	err = os.MkdirAll(submoduleDir, 0755)
	require.NoError(t, err)

	submoduleTF := `
variable "subnet_id" {
  description = "VPC subnet ID"
  type        = string
}

resource "aws_subnet" "example" {
  vpc_id     = var.vpc_id
  cidr_block = "10.0.1.0/24"
}
`
	err = os.WriteFile(filepath.Join(submoduleDir, "main.tf"), []byte(submoduleTF), 0644)
	require.NoError(t, err)

	// Create examples directory
	examplesDir := filepath.Join(dir, "examples")
	err = os.MkdirAll(examplesDir, 0755)
	require.NoError(t, err)

	// Create an example
	exampleDir := filepath.Join(examplesDir, "example1")
	err = os.MkdirAll(exampleDir, 0755)
	require.NoError(t, err)

	exampleTF := `
# Example usage of the module

module "test_module" {
  source = "../../."

  instance_type = "t2.micro"
}
`
	err = os.WriteFile(filepath.Join(exampleDir, "main.tf"), []byte(exampleTF), 0644)
	require.NoError(t, err)

	// Commit the changes
	runCommand(t, dir, "git", "add", ".")
	runCommand(t, dir, "git", "commit", "-m", "Add submodules and examples")
}

// runCommand executes a command in the given directory
func runCommand(t *testing.T, dir string, name string, args ...string) {
	t.Helper()

	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Command %s %v failed: %v\nOutput: %s", name, args, err, string(output))
	}
}
