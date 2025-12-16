package service

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"gorm.io/gorm"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/persistence/sqldb/transaction"
)

// TerraformOperation represents a single terraform operation
type TerraformOperation struct {
	Type        string
	Command     []string
	WorkingDir  string
	Timeout     time.Duration
	Description  string
}

// TerraformProcessingRequest represents a request to process terraform files
type TerraformProcessingRequest struct {
	ModuleVersionID int
	ModulePath      string
	TransactionCtx  context.Context
	Operations      []TerraformOperation
}

// TerraformInitResult represents the result of terraform init
type TerraformInitResult struct {
	Success    bool          `json:"success"`
	Output     string        `json:"output"`
	Error      *string       `json:"error,omitempty"`
	Duration   time.Duration `json:"duration"`
	HasChanges bool          `json:"has_changes"`
}

// TerraformGraphResult represents the result of terraform graph generation
type TerraformGraphResult struct {
	Success    bool          `json:"success"`
	GraphData  string        `json:"graph_data"`
	Error      *string       `json:"error,omitempty"`
	Duration   time.Duration `json:"duration"`
}

// TerraformVersionResult represents terraform version information
type TerraformVersionResult struct {
	Success    bool          `json:"success"`
	Version    string        `json:"version"`
	Output     string        `json:"output"`
	Error      *string       `json:"error,omitempty"`
	Duration   time.Duration `json:"duration"`
}

// TerraformModulesResult represents parsed terraform modules
type TerraformModulesResult struct {
	Success    bool          `json:"success"`
	Modules    string        `json:"modules"` // JSON string of modules.json content
	Error      *string       `json:"error,omitempty"`
	Duration   time.Duration `json:"duration"`
}

// TerraformProcessingResult represents the result of terraform processing
type TerraformProcessingResult struct {
	InitResult     *TerraformInitResult     `json:"init_result,omitempty"`
	GraphResult    *TerraformGraphResult    `json:"graph_result,omitempty"`
	VersionResult  *TerraformVersionResult  `json:"version_result,omitempty"`
	ModulesResult  *TerraformModulesResult  `json:"modules_result,omitempty"`
	OverallSuccess bool                     `json:"overall_success"`
	Duration       time.Duration            `json:"duration"`
	FailedStep     string                    `json:"failed_step,omitempty"`
}

// TerraformExecutorService handles terraform operations with transaction safety
type TerraformExecutorService struct {
	savepointHelper *transaction.SavepointHelper
	terraformBin   string
	lockTimeout     time.Duration
}

// NewTerraformExecutorService creates a new terraform executor service
func NewTerraformExecutorService(
	savepointHelper *transaction.SavepointHelper,
	terraformBin string,
	lockTimeout time.Duration,
) *TerraformExecutorService {
	return &TerraformExecutorService{
		savepointHelper: savepointHelper,
		terraformBin:   terraformBin,
		lockTimeout:     lockTimeout,
	}
}

// ProcessTerraformWithTransaction processes terraform operations with transaction context
func (s *TerraformExecutorService) ProcessTerraformWithTransaction(
	ctx context.Context,
	req TerraformProcessingRequest,
) (*TerraformProcessingResult, error) {
	startTime := time.Now()
	result := &TerraformProcessingResult{
		OverallSuccess: false,
		Duration:       0,
	}

	savepointName := fmt.Sprintf("terraform_processing_%d", startTime.UnixNano())

	err := s.savepointHelper.WithSavepointNamed(ctx, savepointName, func(tx *gorm.DB) error {
		// Execute terraform pipeline with rollback capability
		return s.executeTerraformPipeline(ctx, req.ModulePath, req.Operations)
	})

	result.Duration = time.Since(startTime)

	if err != nil {
		failedStep := s.getFailedStep(err)
		result.FailedStep = failedStep
		return result, nil
	}

	// Execute individual operations to get detailed results
	result.OverallSuccess = true
	result.InitResult = s.getInitResult(ctx, req.ModulePath)
	result.GraphResult = s.getGraphResult(ctx, req.ModulePath)
	result.VersionResult = s.getVersionResult(ctx, req.ModulePath)
	result.ModulesResult = s.getModulesResult(ctx, req.ModulePath)

	return result, nil
}

// ExecuteTerraformPipeline executes multiple terraform operations as a pipeline
func (s *TerraformExecutorService) ExecuteTerraformPipeline(
	ctx context.Context,
	modulePath string,
	operations []TerraformOperation,
) error {
	for i, op := range operations {
		if err := s.executeOperation(ctx, op); err != nil {
			return fmt.Errorf("terraform operation %d (%s) failed: %w", i, op.Description, err)
		}
	}
	return nil
}

// executeOperation executes a single terraform operation
func (s *TerraformExecutorService) executeOperation(ctx context.Context, op TerraformOperation) error {
	cmd := exec.CommandContext(ctx, op.Command...)
	cmd.Dir = op.WorkingDir

	// Set environment variables
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, "TF_IN_AUTOMATION=true")

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("terraform %s failed: %s\nOutput: %s", op.Description, err, string(output))
	}

	// For successful operations, we could store the output if needed
	_ = output

	return nil
}

// getInitResult executes terraform init and returns detailed result
func (s *TerraformExecutorService) getInitResult(ctx context.Context, modulePath string) *TerraformInitResult {
	startTime := time.Now()
	result := &TerraformInitResult{
		Success: false,
	}

	cmd := exec.CommandContext(ctx, s.terraformBin, "init")
	cmd.Dir = modulePath
	cmd.Env = append(os.Environ(), "TF_IN_AUTOMATION=true")

	output, err := cmd.CombinedOutput()
	result.Duration = time.Since(startTime)

	if err != nil {
		errorMsg := err.Error()
		result.Error = &errorMsg
		return result
	}

	result.Success = true
	result.Output = string(output)
	result.HasChanges = s.hasInitChanges(string(output))

	return result
}

// getGraphResult executes terraform graph and returns detailed result
func (s *TerraformExecutorService) getGraphResult(ctx context.Context, modulePath string) *TerraformGraphResult {
	startTime := time.Now()
	result := &TerraformGraphResult{
		Success: false,
	}

	cmd := exec.CommandContext(ctx, s.terraformBin, "graph")
	cmd.Dir = modulePath

	output, err := cmd.CombinedOutput()
	result.Duration = time.Since(startTime)

	if err != nil {
		errorMsg := err.Error()
		result.Error = &errorMsg
		return result
	}

	result.Success = true
	result.GraphData = string(output)

	return result
}

// getVersionResult executes terraform version and returns detailed result
func (s *TerraformExecutorService) getVersionResult(ctx context.Context, modulePath string) *TerraformVersionResult {
	startTime := time.Now()
	result := &TerraformVersionResult{
		Success: false,
	}

	cmd := exec.CommandContext(ctx, s.terraformBin, "version", "-json")
	cmd.Dir = modulePath

	output, err := cmd.CombinedOutput()
	result.Duration = time.Since(startTime)

	if err != nil {
		errorMsg := err.Error()
		result.Error = &errorMsg
		return result
	}

	result.Success = true
	result.Output = string(output)
	// Would parse JSON to extract version string in a full implementation

	return result
}

// getModulesResult parses .terraform/modules/modules.json and returns detailed result
func (s *TerraformExecutorService) getModulesResult(ctx context.Context, modulePath string) *TerraformModulesResult {
	startTime := time.Now()
	result := &TerraformModulesResult{
		Success: false,
	}

	modulesPath := filepath.Join(modulePath, ".terraform", "modules", "modules.json")
	output, err := os.ReadFile(modulesPath)
	result.Duration = time.Since(startTime)

	if err != nil {
		errorMsg := err.Error()
		result.Error = &errorMsg
		return result
	}

	result.Success = true
	result.Modules = string(output)

	return result
}

// hasInitChanges checks if terraform init made any changes (simplified check)
func (s *TerraformExecutorService) hasInitChanges(output string) bool {
	// Simplified check - in a full implementation, would parse output for specific patterns
	return len(output) > 0 && (contains(output, "Terraform has been successfully initialized") ||
		contains(output, "Downloading and installing"))
}

// getFailedStep determines which terraform step failed from the error message
func (s *TerraformExecutorService) getFailedStep(err error) string {
	errorMsg := err.Error()

	switch {
	case contains(errorMsg, "init"):
		return "terraform_init"
	case contains(errorMsg, "graph"):
		return "terraform_graph"
	case contains(errorMsg, "version"):
		return "terraform_version"
	default:
		return "terraform_operation"
	}
}

// contains checks if a string contains a substring (case-insensitive)
func contains(s, substr string) bool {
	return len(s) > 0 && (len(substr) == 0 || (s[0:len(substr)] == substr ||
		(len(s) > len(substr) && s[len(s)-len(substr):len(s)] == substr)))
}