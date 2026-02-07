package service

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/shared/service"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockCommandService is a mock for SystemCommandService
type MockCommandService struct {
	executeFunc func(ctx context.Context, cmd *service.Command) (*service.CommandResult, error)
}

func (m *MockCommandService) Execute(ctx context.Context, cmd *service.Command) (*service.CommandResult, error) {
	if m.executeFunc != nil {
		return m.executeFunc(ctx, cmd)
	}
	return &service.CommandResult{
		Stdout: "",
		Stderr: "",
	}, nil
}

func (m *MockCommandService) ExecuteWithInput(ctx context.Context, cmd *service.Command, input string) (*service.CommandResult, error) {
	return nil, nil
}

func TestNewInfracostService(t *testing.T) {
	config := &InfracostConfig{
		InfracostAPIKey: "test-api-key",
	}
	logger := zerolog.Nop()
	commandService := &MockCommandService{}

	service := NewInfracostService(config, logger, commandService)

	assert.NotNil(t, service)
	assert.NotNil(t, service.config)
	assert.NotNil(t, service.logger)
	assert.NotNil(t, service.commandService)
}

func TestInfracostService_IsAvailable_WithAPIKey(t *testing.T) {
	config := &InfracostConfig{
		InfracostAPIKey: "test-api-key",
	}
	service := NewInfracostService(config, zerolog.Nop(), &MockCommandService{})

	assert.True(t, service.IsAvailable())
}

func TestInfracostService_IsAvailable_WithoutAPIKey(t *testing.T) {
	config := &InfracostConfig{
		InfracostAPIKey: "",
	}
	service := NewInfracostService(config, zerolog.Nop(), &MockCommandService{})

	assert.False(t, service.IsAvailable())
}

func TestInfracostService_IsAvailable_NilConfig(t *testing.T) {
	service := NewInfracostService(nil, zerolog.Nop(), &MockCommandService{})

	assert.False(t, service.IsAvailable())
}

func TestInfracostService_AnalyzeExample_NotAvailable(t *testing.T) {
	config := &InfracostConfig{
		InfracostAPIKey: "",
	}
	service := NewInfracostService(config, zerolog.Nop(), &MockCommandService{})

	result, err := service.AnalyzeExample(context.Background(), "/some/path")

	assert.NoError(t, err)
	assert.Nil(t, result) // Should return nil, nil when not available
}

func TestInfracostService_AnalyzeExample_EmptyPath(t *testing.T) {
	config := &InfracostConfig{
		InfracostAPIKey: "test-api-key",
	}
	service := NewInfracostService(config, zerolog.Nop(), &MockCommandService{})

	result, err := service.AnalyzeExample(context.Background(), "")

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "example path cannot be empty")
}

func TestInfracostService_AnalyzeExample_NonExistentPath(t *testing.T) {
	config := &InfracostConfig{
		InfracostAPIKey: "test-api-key",
	}
	service := NewInfracostService(config, zerolog.Nop(), &MockCommandService{})

	result, err := service.AnalyzeExample(context.Background(), "/nonexistent/path/to/example")

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "example path does not exist")
}

func TestInfracostService_AnalyzeExample_ExecutableNotFound(t *testing.T) {
	config := &InfracostConfig{
		InfracostAPIKey: "test-api-key",
	}
	mockCommand := &MockCommandService{
		executeFunc: func(ctx context.Context, cmd *service.Command) (*service.CommandResult, error) {
			return &service.CommandResult{
				Stdout: "",
				Stderr: "exec: \"infracost\": executable file not found",
			}, fmt.Errorf("executable file not found")
		},
	}
	service := NewInfracostService(config, zerolog.Nop(), mockCommand)

	// Create a temp directory for the test
	tempDir := t.TempDir()
	result, err := service.AnalyzeExample(context.Background(), tempDir)

	assert.NoError(t, err)
	assert.Nil(t, result) // Should return nil, nil when executable not found
}

func TestInfracostService_AnalyzeExample_CommandFailure(t *testing.T) {
	config := &InfracostConfig{
		InfracostAPIKey: "test-api-key",
	}
	mockCommand := &MockCommandService{
		executeFunc: func(ctx context.Context, cmd *service.Command) (*service.CommandResult, error) {
			return &service.CommandResult{
				Stdout: "",
				Stderr: "infracost: some error occurred",
			}, fmt.Errorf("some error occurred")
		},
	}
	service := NewInfracostService(config, zerolog.Nop(), mockCommand)

	tempDir := t.TempDir()
	result, err := service.AnalyzeExample(context.Background(), tempDir)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "infracost failed")
}

func TestInfracostService_AnalyzeExample_Success(t *testing.T) {
	// Create a temporary directory to simulate an example
	tempDir := t.TempDir()
	examplePath := filepath.Join(tempDir, "example")
	err := os.Mkdir(examplePath, 0755)
	require.NoError(t, err)

	// Create a simple Terraform file
	tfFile := filepath.Join(examplePath, "main.tf")
	err = os.WriteFile(tfFile, []byte(`
resource "aws_s3_bucket" "example" {
  bucket = "test-bucket"
}
`), 0644)
	require.NoError(t, err)

	// Create a temporary file to simulate infracost output
	tmpOutputFile, err := os.CreateTemp(tempDir, "infracost-output-*.json")
	require.NoError(t, err)
	tmpOutputPath := tmpOutputFile.Name()

	// Write mock infracost JSON output
	infracostOutput := map[string]interface{}{
		"total_monthly_cost": 12.50,
		"total_hourly_cost":  0.0173611,
		"projects": []interface{}{},
	}
	outputJSON, _ := json.Marshal(infracostOutput)
	err = os.WriteFile(tmpOutputPath, outputJSON, 0644)
	require.NoError(t, err)
	tmpOutputFile.Close()

	config := &InfracostConfig{
		InfracostAPIKey: "test-api-key",
	}
	mockCommand := &MockCommandService{
		executeFunc: func(ctx context.Context, cmd *service.Command) (*service.CommandResult, error) {
			// Verify command arguments
			assert.Equal(t, "infracost", cmd.Name)
			assert.Contains(t, cmd.Args, "--path", examplePath)
			assert.Contains(t, cmd.Args, "--format", "json")
			assert.Contains(t, cmd.Args, "--out-file")

			// Copy the temp file to the output path specified in the command
			for i, arg := range cmd.Args {
				if arg == "--out-file" && i+1 < len(cmd.Args) {
					outputPath := cmd.Args[i+1]
					input, _ := os.ReadFile(tmpOutputPath)
					os.WriteFile(outputPath, input, 0644)
				}
			}

			return &service.CommandResult{
				Stdout: "infracost analysis complete",
				Stderr: "",
			}, nil
		},
	}
	service := NewInfracostService(config, zerolog.Nop(), mockCommand)

	result, err := service.AnalyzeExample(context.Background(), examplePath)

	assert.NoError(t, err)
	assert.NotNil(t, result)

	// Verify the JSON can be parsed
	var parsed map[string]interface{}
	err = json.Unmarshal(result, &parsed)
	assert.NoError(t, err)
	assert.Equal(t, 12.50, parsed["total_monthly_cost"])
}

func TestInfracostService_AnalyzeExample_InvalidJSONOutput(t *testing.T) {
	tempDir := t.TempDir()
	examplePath := filepath.Join(tempDir, "example")
	err := os.Mkdir(examplePath, 0755)
	require.NoError(t, err)

	// Create a temp file with invalid JSON
	tmpOutputFile, err := os.CreateTemp(tempDir, "infracost-output-*.json")
	require.NoError(t, err)
	tmpOutputPath := tmpOutputFile.Name()

	err = os.WriteFile(tmpOutputPath, []byte("invalid json{{{"), 0644)
	require.NoError(t, err)
	tmpOutputFile.Close()

	config := &InfracostConfig{
		InfracostAPIKey: "test-api-key",
	}
	mockCommand := &MockCommandService{
		executeFunc: func(ctx context.Context, cmd *service.Command) (*service.CommandResult, error) {
			// Copy the temp file to the output path specified in the command
			for i, arg := range cmd.Args {
				if arg == "--out-file" && i+1 < len(cmd.Args) {
					outputPath := cmd.Args[i+1]
					input, _ := os.ReadFile(tmpOutputPath)
					os.WriteFile(outputPath, input, 0644)
				}
			}

			return &service.CommandResult{
				Stdout: "infracost analysis complete",
				Stderr: "",
			}, nil
		},
	}
	service := NewInfracostService(config, zerolog.Nop(), mockCommand)

	result, err := service.AnalyzeExample(context.Background(), examplePath)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "not valid JSON")
}

func TestInfracostService_buildEnvironment_NoConfig(t *testing.T) {
	config := &InfracostConfig{
		InfracostAPIKey:                     "test-key",
		InternalExtractionAnalyticsToken: "",
		PublicURL:                          "",
	}
	service := NewInfracostService(config, zerolog.Nop(), &MockCommandService{})

	env := service.buildEnvironment()
	assert.NotNil(t, env)
	// Should have at least the current environment
	assert.NotEmpty(t, env)
}

func TestInfracostService_buildEnvironment_WithTerraformCloud(t *testing.T) {
	config := &InfracostConfig{
		InfracostAPIKey:                     "test-key",
		InternalExtractionAnalyticsToken: "internal-token-123",
		PublicURL:                          "https://terrareg.example.com",
	}
	service := NewInfracostService(config, zerolog.Nop(), &MockCommandService{})

	env := service.buildEnvironment()
	assert.NotNil(t, env)

	// Convert to map for easier checking
	envMap := make(map[string]string)
	for _, e := range env {
		parts := strings.SplitN(e, "=", 2)
		if len(parts) == 2 {
			envMap[parts[0]] = parts[1]
		}
	}

	assert.Equal(t, "internal-token-123", envMap["INFRACOST_TERRAFORM_CLOUD_TOKEN"])
	assert.Equal(t, "terrareg.example.com", envMap["INFRACOST_TERRAFORM_CLOUD_HOST"])
}

func TestInfracostService_buildEnvironment_InvalidPublicURL(t *testing.T) {
	config := &InfracostConfig{
		InfracostAPIKey:                     "test-key",
		InternalExtractionAnalyticsToken: "internal-token-123",
		PublicURL:                          "://invalid-url", // Invalid URL
	}
	service := NewInfracostService(config, zerolog.Nop(), &MockCommandService{})

	env := service.buildEnvironment()
	assert.NotNil(t, env)

	// Convert to map for checking
	envMap := make(map[string]string)
	for _, e := range env {
		parts := strings.SplitN(e, "=", 2)
		if len(parts) == 2 {
			envMap[parts[0]] = parts[1]
		}
	}

	// Terraform Cloud env vars should NOT be set due to invalid URL
	_, hasToken := envMap["INFRACOST_TERRAFORM_CLOUD_TOKEN"]
	_, hasHost := envMap["INFRACOST_TERRAFORM_CLOUD_HOST"]
	assert.False(t, hasToken)
	assert.False(t, hasHost)
}
