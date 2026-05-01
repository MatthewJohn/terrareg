package service

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/shared/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestTerraformExecutorService_NewTerraformExecutorService tests service creation
func TestTerraformExecutorService_NewTerraformExecutorService(t *testing.T) {
	t.Parallel()

	txManager := &mockTransactionManager{}
	commandService := &mockSystemCommandService{}
	tfswitchConfig := &TfswitchConfig{
		DefaultTerraformVersion: "1.5.0",
		TerraformProduct:        "terraform",
		ArchiveMirror:           "https://releases.hashicorp.com/terraform",
		BinaryPath:              "/usr/bin/terraform",
	}

	service := NewTerraformExecutorService(txManager, commandService, "terraform", 30*time.Second, tfswitchConfig)

	assert.NotNil(t, service)
	assert.Equal(t, txManager, service.txManager)
	assert.Equal(t, commandService, service.commandService)
	assert.Equal(t, "terraform", service.terraformBin)
	assert.Equal(t, 30*time.Second, service.lockTimeout)
	assert.Equal(t, tfswitchConfig, service.tfswitchConfig)
}

// TestTerraformExecutorService_NewTerraformExecutorServiceNilConfig tests that nil config returns nil
func TestTerraformExecutorService_NewTerraformExecutorServiceNilConfig(t *testing.T) {
	t.Parallel()

	txManager := &mockTransactionManager{}
	commandService := &mockSystemCommandService{}

	service := NewTerraformExecutorService(txManager, commandService, "terraform", 30*time.Second, nil)

	assert.Nil(t, service)
}

// TestTerraformExecutorService_ProcessTerraformWithTransaction_Success tests successful processing
func TestTerraformExecutorService_ProcessTerraformWithTransaction_Success(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	tempDir := t.TempDir()

	// Create a minimal terraform file
	tfContent := `variable "test" {}`
	err := os.WriteFile(filepath.Join(tempDir, "main.tf"), []byte(tfContent), 0644)
	require.NoError(t, err)

	// Create .terraform/modules directory for modules.json
	modulesDir := filepath.Join(tempDir, ".terraform", "modules")
	err = os.MkdirAll(modulesDir, 0755)
	require.NoError(t, err)

	modulesJSON := `{"modules": []}`
	err = os.WriteFile(filepath.Join(modulesDir, "modules.json"), []byte(modulesJSON), 0644)
	require.NoError(t, err)

	// Setup mocks
	txManager := &mockTransactionManager{
		withTransactionFunc: func(ctx context.Context, fn func(context.Context) error) error {
			return fn(ctx)
		},
	}

	commandService := &mockSystemCommandService{}
	tfswitchConfig := &TfswitchConfig{
		DefaultTerraformVersion: "1.5.0",
		TerraformProduct:        "terraform",
		BinaryPath:              "/usr/bin/terraform",
	}

	service := NewTerraformExecutorService(txManager, commandService, "terraform", 30*time.Second, tfswitchConfig)
	require.NotNil(t, service)

	req := TerraformProcessingRequest{
		ModuleVersionID: 1,
		ModulePath:      tempDir,
		TransactionCtx:  ctx,
	}

	result, err := service.ProcessTerraformWithTransaction(ctx, req)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.True(t, result.OverallSuccess)
	assert.NotNil(t, result.InitResult)
	assert.True(t, result.InitResult.Success)
	assert.NotNil(t, result.GraphResult)
	assert.True(t, result.GraphResult.Success)
	assert.NotNil(t, result.VersionResult)
	assert.True(t, result.VersionResult.Success)
	assert.NotNil(t, result.ModulesResult)
	assert.True(t, result.ModulesResult.Success)
}

// TestTerraformExecutorService_ProcessTerraformWithTransaction_InitFailure tests handling of init failure
func TestTerraformExecutorService_ProcessTerraformWithTransaction_InitFailure(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	tempDir := t.TempDir()

	// Create a terraform file with S3 backend (to test override)
	tfContent := `
terraform {
  backend "s3" {
    bucket = "test"
  }
}
variable "test" {}
`
	err := os.WriteFile(filepath.Join(tempDir, "main.tf"), []byte(tfContent), 0644)
	require.NoError(t, err)

	// Setup transaction manager to return error
	txManager := &mockTransactionManager{
		withTransactionFunc: func(ctx context.Context, fn func(context.Context) error) error {
			// Call the function, which will fail during terraform init
			return fn(ctx)
		},
	}

	// Create mock command service that returns error for terraform init
	errorCommandService := &mockErrorSystemCommandService{
		executeFunc: func(ctx context.Context, cmd *service.Command) (*service.CommandResult, error) {
			if len(cmd.Args) > 0 && cmd.Args[0] == "init" {
				return &service.CommandResult{
					Stdout:   "Error: Backend configuration changed",
					Stderr:   "Error initializing backend",
					ExitCode: 1,
				}, assert.AnError
			}
			return &service.CommandResult{
				Stdout:   `{"terraform_version": "1.5.0"}`,
				Stderr:   "",
				ExitCode: 0,
			}, nil
		},
	}

	tfswitchConfig := &TfswitchConfig{
		DefaultTerraformVersion: "1.5.0",
		TerraformProduct:        "terraform",
		BinaryPath:              "/usr/bin/terraform",
	}

	service := NewTerraformExecutorService(txManager, errorCommandService, "terraform", 30*time.Second, tfswitchConfig)
	require.NotNil(t, service)

	req := TerraformProcessingRequest{
		ModuleVersionID: 1,
		ModulePath:      tempDir,
		TransactionCtx:  ctx,
	}

	result, err := service.ProcessTerraformWithTransaction(ctx, req)

	// Should not return error, but result should indicate failure
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.False(t, result.OverallSuccess)
	assert.NotNil(t, result.InitResult)
	assert.False(t, result.InitResult.Success)
	assert.NotEmpty(t, result.FailedStep)
}

// TestTerraformExecutorService_OverrideTerraformBackend_Basic tests basic backend override
func TestTerraformExecutorService_OverrideTerraformBackend_Basic(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()

	// Create terraform file with S3 backend
	tfContent := `
terraform {
  backend "s3" {
    bucket = "my-bucket"
    key    = "path/to/my/key"
    region = "us-east-1"
  }
}
`
	err := os.WriteFile(filepath.Join(tempDir, "main.tf"), []byte(tfContent), 0644)
	require.NoError(t, err)

	txManager := &mockTransactionManager{}
	commandService := &mockSystemCommandService{}
	tfswitchConfig := &TfswitchConfig{
		DefaultTerraformVersion: "1.5.0",
		TerraformProduct:        "terraform",
	}

	service := NewTerraformExecutorService(txManager, commandService, "terraform", 30*time.Second, tfswitchConfig)

	overrideFile, err := service.OverrideTerraformBackend(tempDir)

	require.NoError(t, err)
	assert.NotNil(t, overrideFile)
	assert.Equal(t, "main_override.tf", *overrideFile)

	// Verify override file was created
	overridePath := filepath.Join(tempDir, *overrideFile)
	content, err := os.ReadFile(overridePath)
	require.NoError(t, err)
	assert.Contains(t, string(content), `backend "local"`)
	assert.Contains(t, string(content), `path = "./.local-state"`)
}

// TestTerraformExecutorService_OverrideTerraformBackend_NoBackend tests handling when no backend is present
func TestTerraformExecutorService_OverrideTerraformBackend_NoBackend(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()

	// Create terraform file without backend
	tfContent := `
resource "aws_instance" "example" {
  ami = "ami-12345678"
}
`
	err := os.WriteFile(filepath.Join(tempDir, "main.tf"), []byte(tfContent), 0644)
	require.NoError(t, err)

	txManager := &mockTransactionManager{}
	commandService := &mockSystemCommandService{}
	tfswitchConfig := &TfswitchConfig{
		DefaultTerraformVersion: "1.5.0",
		TerraformProduct:        "terraform",
	}

	service := NewTerraformExecutorService(txManager, commandService, "terraform", 30*time.Second, tfswitchConfig)

	overrideFile, err := service.OverrideTerraformBackend(tempDir)

	require.NoError(t, err)
	assert.Nil(t, overrideFile)
}

// TestTerraformExecutorService_OverrideTerraformBackend_MultipleFiles tests handling of multiple terraform files
func TestTerraformExecutorService_OverrideTerraformBackend_MultipleFiles(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()

	// Create multiple terraform files
	tfContent1 := `
terraform {
  backend "s3" {
    bucket = "bucket1"
  }
}
`
	tfContent2 := `
resource "aws_instance" "example" {}
`
	err := os.WriteFile(filepath.Join(tempDir, "backend.tf"), []byte(tfContent1), 0644)
	require.NoError(t, err)
	err = os.WriteFile(filepath.Join(tempDir, "main.tf"), []byte(tfContent2), 0644)
	require.NoError(t, err)

	txManager := &mockTransactionManager{}
	commandService := &mockSystemCommandService{}
	tfswitchConfig := &TfswitchConfig{
		DefaultTerraformVersion: "1.5.0",
		TerraformProduct:        "terraform",
	}

	service := NewTerraformExecutorService(txManager, commandService, "terraform", 30*time.Second, tfswitchConfig)

	overrideFile, err := service.OverrideTerraformBackend(tempDir)

	require.NoError(t, err)
	assert.NotNil(t, overrideFile)
	assert.Equal(t, "backend_override.tf", *overrideFile)
}

// TestTerraformExecutorService_OverrideTerraformBackend_ComplexBackend tests handling of complex backend configuration
func TestTerraformExecutorService_OverrideTerraformBackend_ComplexBackend(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()

	// Create terraform file with complex backend
	tfContent := `
# Some comments
provider "aws" {
  region = "us-east-1"
}

terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = ">= 2.7.0"
    }
  }
  backend "s3" {
    bucket         = "does-not-exist"
    key            = "path/to/my/key"
    region         = "thisdoesnotexistforterrareg"
    profile        = "thisdoesnotexistforterrareg"
    encrypt        = true
    kms_key_id     = "arn:aws:kms:us-east-1:123456789012:key/12345678-1234-1234-1234-123456789012"
    dynamodb_table = "terraform-lock"
  }
}
`
	err := os.WriteFile(filepath.Join(tempDir, "state.tf"), []byte(tfContent), 0644)
	require.NoError(t, err)

	txManager := &mockTransactionManager{}
	commandService := &mockSystemCommandService{}
	tfswitchConfig := &TfswitchConfig{
		DefaultTerraformVersion: "1.5.0",
		TerraformProduct:        "terraform",
	}

	service := NewTerraformExecutorService(txManager, commandService, "terraform", 30*time.Second, tfswitchConfig)

	overrideFile, err := service.OverrideTerraformBackend(tempDir)

	require.NoError(t, err)
	assert.NotNil(t, overrideFile)

	// Verify override file was created
	overridePath := filepath.Join(tempDir, *overrideFile)
	_, err = os.Stat(overridePath)
	assert.NoError(t, err)
}

// TestTerraformExecutorService_TerraformBinaryPath_Default tests default binary path
func TestTerraformExecutorService_TerraformBinaryPath_Default(t *testing.T) {
	t.Parallel()

	txManager := &mockTransactionManager{}
	commandService := &mockSystemCommandService{}
	tfswitchConfig := &TfswitchConfig{
		DefaultTerraformVersion: "1.5.0",
		TerraformProduct:        "terraform",
	}

	service := NewTerraformExecutorService(txManager, commandService, "terraform", 30*time.Second, tfswitchConfig)

	assert.Equal(t, "terraform", service.TerraformBinaryPath())
}

// TestTerraformExecutorService_TerraformBinaryPath_CustomBinaryPath tests custom binary path from tfswitch config
func TestTerraformExecutorService_TerraformBinaryPath_CustomBinaryPath(t *testing.T) {
	t.Parallel()

	txManager := &mockTransactionManager{}
	commandService := &mockSystemCommandService{}
	tfswitchConfig := &TfswitchConfig{
		DefaultTerraformVersion: "1.5.0",
		TerraformProduct:        "terraform",
		BinaryPath:              "/custom/path/to/terraform",
	}

	service := NewTerraformExecutorService(txManager, commandService, "terraform", 30*time.Second, tfswitchConfig)

	assert.Equal(t, "/custom/path/to/terraform", service.TerraformBinaryPath())
}

// TestTerraformExecutorService_TerraformBinaryPath_ProductSwitch tests binary path based on product
func TestTerraformExecutorService_TerraformBinaryPath_ProductSwitch(t *testing.T) {
	t.Parallel()

	t.Run("terraform product returns terraform", func(t *testing.T) {
		txManager := &mockTransactionManager{}
		commandService := &mockSystemCommandService{}
		tfswitchConfig := &TfswitchConfig{
			DefaultTerraformVersion: "1.5.0",
			TerraformProduct:        "terraform",
		}

		service := NewTerraformExecutorService(txManager, commandService, "terraform", 30*time.Second, tfswitchConfig)
		assert.Equal(t, "terraform", service.TerraformBinaryPath())
	})

	t.Run("opentofu product returns tofu", func(t *testing.T) {
		txManager := &mockTransactionManager{}
		commandService := &mockSystemCommandService{}
		tfswitchConfig := &TfswitchConfig{
			DefaultTerraformVersion: "1.5.0",
			TerraformProduct:        "opentofu",
			BinaryPath:              "tofu",
		}

		service := NewTerraformExecutorService(txManager, commandService, "tofu", 30*time.Second, tfswitchConfig)
		assert.Equal(t, "tofu", service.TerraformBinaryPath())
	})
}

// TestTerraformExecutorService_RunTerraformWithCallback_Success tests successful callback execution
func TestTerraformExecutorService_RunTerraformWithCallback_Success(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	tempDir := t.TempDir()

	// Create .terraform directory to avoid terraform init errors
	terraformDir := filepath.Join(tempDir, ".terraform")
	err := os.MkdirAll(terraformDir, 0755)
	require.NoError(t, err)

	txManager := &mockTransactionManager{}
	commandService := &mockSystemCommandService{}
	tfswitchConfig := &TfswitchConfig{
		DefaultTerraformVersion: "1.5.0",
		TerraformProduct:        "terraform",
		BinaryPath:              "/usr/bin/terraform",
	}

	service := NewTerraformExecutorService(txManager, commandService, "terraform", 30*time.Second, tfswitchConfig)

	callbackExecuted := false
	err = service.RunTerraformWithCallback(ctx, tempDir, func(ctx context.Context) error {
		callbackExecuted = true
		return nil
	})

	assert.NoError(t, err)
	assert.True(t, callbackExecuted)
}

// TestTerraformExecutorService_RunTerraformWithCallback_CallbackError tests callback error handling
func TestTerraformExecutorService_RunTerraformWithCallback_CallbackError(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	tempDir := t.TempDir()

	// Create .terraform directory
	terraformDir := filepath.Join(tempDir, ".terraform")
	err := os.MkdirAll(terraformDir, 0755)
	require.NoError(t, err)

	txManager := &mockTransactionManager{}
	commandService := &mockSystemCommandService{}
	tfswitchConfig := &TfswitchConfig{
		DefaultTerraformVersion: "1.5.0",
		TerraformProduct:        "terraform",
		BinaryPath:              "/usr/bin/terraform",
	}

	service := NewTerraformExecutorService(txManager, commandService, "terraform", 30*time.Second, tfswitchConfig)

	callbackErr := assert.AnError
	err = service.RunTerraformWithCallback(ctx, tempDir, func(ctx context.Context) error {
		return callbackErr
	})

	assert.Error(t, err)
	assert.Equal(t, callbackErr, err)
}

// Mock implementations

type mockTransactionManager struct {
	withTransactionFunc    func(ctx context.Context, fn func(context.Context) error) error
	withNamedTransactionFn func(ctx context.Context, name string, fn func(context.Context) error) error
	isActiveFunc          func(ctx context.Context) bool
}

func (m *mockTransactionManager) WithTransaction(ctx context.Context, fn func(context.Context) error) error {
	if m.withTransactionFunc != nil {
		return m.withTransactionFunc(ctx, fn)
	}
	return fn(ctx)
}

func (m *mockTransactionManager) WithNamedTransaction(ctx context.Context, name string, fn func(context.Context) error) error {
	if m.withNamedTransactionFn != nil {
		return m.withNamedTransactionFn(ctx, name, fn)
	}
	return fn(ctx)
}

func (m *mockTransactionManager) IsTransactionActive(ctx context.Context) bool {
	if m.isActiveFunc != nil {
		return m.isActiveFunc(ctx)
	}
	return false
}

type mockErrorSystemCommandService struct {
	executeFunc func(ctx context.Context, cmd *service.Command) (*service.CommandResult, error)
}

func (m *mockErrorSystemCommandService) Execute(ctx context.Context, cmd *service.Command) (*service.CommandResult, error) {
	if m.executeFunc != nil {
		return m.executeFunc(ctx, cmd)
	}
	return &service.CommandResult{
		Stdout:   "",
		Stderr:   "command failed",
		ExitCode: 1,
	}, assert.AnError
}

func (m *mockErrorSystemCommandService) ExecuteWithInput(ctx context.Context, cmd *service.Command, input string) (*service.CommandResult, error) {
	if m.executeFunc != nil {
		return m.executeFunc(ctx, cmd)
	}
	return &service.CommandResult{
		Stdout:   "",
		Stderr:   "command failed",
		ExitCode: 1,
	}, assert.AnError
}
