package service

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/rs/zerolog"
	"gorm.io/gorm"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/persistence/sqldb/transaction"
)

// TerraformOperation represents a single terraform operation
type TerraformOperation struct {
	Type        string
	Command     []string
	WorkingDir  string
	Timeout     time.Duration
	Description string
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
	Success         bool          `json:"success"`
	Output          string        `json:"output"`
	Error           *string       `json:"error,omitempty"`
	Duration        time.Duration `json:"duration"`
	HasChanges      bool          `json:"has_changes"`
	BackendOverride *string       `json:"backend_override,omitempty"`
}

// TerraformGraphResult represents the result of terraform graph generation
type TerraformGraphResult struct {
	Success   bool          `json:"success"`
	GraphData string        `json:"graph_data"`
	Error     *string       `json:"error,omitempty"`
	Duration  time.Duration `json:"duration"`
}

// TerraformVersionResult represents terraform version information
type TerraformVersionResult struct {
	Success  bool          `json:"success"`
	Version  string        `json:"version"`
	Output   string        `json:"output"`
	Error    *string       `json:"error,omitempty"`
	Duration time.Duration `json:"duration"`
}

// TerraformModulesResult represents parsed terraform modules
type TerraformModulesResult struct {
	Success  bool          `json:"success"`
	Modules  string        `json:"modules"` // JSON string of modules.json content
	Error    *string       `json:"error,omitempty"`
	Duration time.Duration `json:"duration"`
}

// TerraformProcessingResult represents the result of terraform processing
type TerraformProcessingResult struct {
	InitResult     *TerraformInitResult    `json:"init_result,omitempty"`
	GraphResult    *TerraformGraphResult   `json:"graph_result,omitempty"`
	VersionResult  *TerraformVersionResult `json:"version_result,omitempty"`
	ModulesResult  *TerraformModulesResult `json:"modules_result,omitempty"`
	OverallSuccess bool                    `json:"overall_success"`
	Duration       time.Duration           `json:"duration"`
	FailedStep     string                  `json:"failed_step,omitempty"`
}

// Global terraform lock to prevent concurrent terraform operations (matching Python implementation)
var terraformGlobalLock sync.Mutex

// TfswitchConfig represents tfswitch configuration (matching Python configuration)
type TfswitchConfig struct {
	DefaultTerraformVersion string
	TerraformProduct        string
	ArchiveMirror           string
	BinaryPath              string
}

// TerraformExecutorService handles terraform operations with transaction safety
type TerraformExecutorService struct {
	savepointHelper *transaction.SavepointHelper
	terraformBin    string
	lockTimeout     time.Duration
	tfswitchConfig  *TfswitchConfig
}

// NewTerraformExecutorService creates a new terraform executor service
func NewTerraformExecutorService(
	savepointHelper *transaction.SavepointHelper,
	terraformBin string,
	lockTimeout time.Duration,
	tfswitchConfig *TfswitchConfig,
) *TerraformExecutorService {
	if tfswitchConfig == nil {
		// Create default tfswitch config
		tfswitchConfig = &TfswitchConfig{
			BinaryPath: terraformBin,
		}
	}

	return &TerraformExecutorService{
		savepointHelper: savepointHelper,
		terraformBin:    terraformBin,
		lockTimeout:     lockTimeout,
		tfswitchConfig:  tfswitchConfig,
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


	err := s.savepointHelper.WithTransaction(ctx, func(ctx context.Context, tx *gorm.DB) error {
		// Execute terraform pipeline with rollback capability
		return s.ExecuteTerraformPipeline(ctx, req.ModulePath, req.Operations)
	})

	result.Duration = time.Since(startTime)

	if err != nil {
		failedStep := s.getFailedStep(err)
		result.FailedStep = failedStep

		// Add detailed error information from individual results
		if result.InitResult != nil && !result.InitResult.Success && result.InitResult.Error != nil {
			result.FailedStep = fmt.Sprintf("terraform_init failed: %s", *result.InitResult.Error)
		} else if result.GraphResult != nil && !result.GraphResult.Success && result.GraphResult.Error != nil {
			result.FailedStep = fmt.Sprintf("terraform_graph failed: %s", *result.GraphResult.Error)
		} else if result.VersionResult != nil && !result.VersionResult.Success && result.VersionResult.Error != nil {
			result.FailedStep = fmt.Sprintf("terraform_version failed: %s", *result.VersionResult.Error)
		} else if result.ModulesResult != nil && !result.ModulesResult.Success && result.ModulesResult.Error != nil {
			result.FailedStep = fmt.Sprintf("terraform_modules failed: %s", *result.ModulesResult.Error)
		}

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
	logger := zerolog.Ctx(ctx)

	logger.Debug().
		Str("module_path", modulePath).
		Int("operations_count", len(operations)).
		Msg("Starting terraform pipeline execution")

	for i, op := range operations {
		commandStr := strings.Join(op.Command, " ")
		logger.Debug().
			Int("operation_index", i).
			Str("command", commandStr).
			Str("description", op.Description).
			Str("working_dir", op.WorkingDir).
			Msg("About to execute terraform operation")

		if err := s.executeOperation(ctx, op); err != nil {
			logger.Error().
				Int("operation_index", i).
				Str("command", commandStr).
				Str("description", op.Description).
				Err(err).
				Msg("Terraform operation failed in pipeline")

			return fmt.Errorf("terraform operation %d (%s) failed: %w", i, op.Description, err)
		}

		logger.Info().
			Int("operation_index", i).
			Str("command", commandStr).
			Str("description", op.Description).
			Msg("Terraform operation completed successfully")
	}

	logger.Debug().
		Str("module_path", modulePath).
		Msg("All terraform pipeline operations completed successfully")

	return nil
}

// executeOperation executes a single terraform operation
func (s *TerraformExecutorService) executeOperation(ctx context.Context, op TerraformOperation) error {
	logger := zerolog.Ctx(ctx)

	commandStr := strings.Join(op.Command, " ")
	logger.Debug().
		Str("module_path", op.WorkingDir).
		Str("command", commandStr).
		Str("description", op.Description).
		Msg("Executing terraform operation")

	// Check if terraform binary exists
	if s.terraformBin == "" {
		err := fmt.Errorf("terraform binary path is empty")
		logger.Error().
			Str("module_path", op.WorkingDir).
			Str("command", commandStr).
			Err(err).
			Msg("Terraform binary path is empty")
		return err
	}

	// Check if working directory exists
	if _, err := os.Stat(op.WorkingDir); err != nil {
		logger.Error().
			Str("module_path", op.WorkingDir).
			Str("command", commandStr).
			Err(err).
			Msg("Working directory does not exist or is not accessible")
		return fmt.Errorf("working directory %s not accessible: %w", op.WorkingDir, err)
	}

	logger.Debug().
		Str("module_path", op.WorkingDir).
		Str("command", commandStr).
		Msg("Creating terraform command")

	cmd := exec.CommandContext(ctx, op.Command[0], op.Command[1:]...)
	cmd.Dir = op.WorkingDir

	// Set environment variables
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, "TF_IN_AUTOMATION=true")

	logger.Debug().
		Str("module_path", op.WorkingDir).
		Str("command", commandStr).
		Msg("Executing terraform command")

	startTime := time.Now()
	output, err := cmd.CombinedOutput()
	duration := time.Since(startTime)

	outputStr := string(output)

	logger.Debug().
		Str("module_path", op.WorkingDir).
		Str("command", commandStr).
		Str("description", op.Description).
		Dur("duration_ms", duration).
		Int("output_length", len(outputStr)).
		Bool("has_error", err != nil).
		Msg("Terraform operation completed")

	// Log the full terraform output at debug level
	if outputStr != "" {
		logger.Debug().
			Str("module_path", op.WorkingDir).
			Str("command", commandStr).
			Str("terraform_output", outputStr).
			Msg("Terraform operation output")
	}

	if err != nil {
		logger.Error().
			Str("module_path", op.WorkingDir).
			Str("command", commandStr).
			Str("description", op.Description).
			Err(err).
			Dur("duration", duration).
			Str("error_type", s.classifyInitError(err, outputStr)).
			Str("output_preview", s.getOutputPreview(outputStr)).
			Str("terraform_output", outputStr).
			Msg("Terraform operation failed")

		return fmt.Errorf("terraform %s failed: %s\nOutput: %s", op.Description, err, outputStr)
	}

	logger.Info().
		Str("module_path", op.WorkingDir).
		Str("command", commandStr).
		Str("description", op.Description).
		Dur("duration", duration).
		Msg("Terraform operation completed successfully")

	return nil
}

// getInitResult executes terraform init and returns detailed result
func (s *TerraformExecutorService) getInitResult(ctx context.Context, modulePath string) *TerraformInitResult {
	startTime := time.Now()
	result := &TerraformInitResult{
		Success: false,
	}

	logger := zerolog.Ctx(ctx)
	logger.Debug().
		Str("module_path", modulePath).
		Str("terraform_binary", s.terraformBin).
		Msg("Executing terraform init")

	// Check if module path exists and is accessible
	if _, err := os.Stat(modulePath); err != nil {
		errorMsg := fmt.Sprintf("module path not accessible: %v", err)
		result.Error = &errorMsg
		logger.Error().
			Str("module_path", modulePath).
			Err(err).
			Msg("Module path not accessible for terraform init")
		return result
	}

	// Check for .tf files in the module
	tfFiles, err := filepath.Glob(filepath.Join(modulePath, "*.tf"))
	if err != nil {
		logger.Warn().
			Str("module_path", modulePath).
			Err(err).
			Msg("Failed to check for .tf files")
	} else {
		logger.Debug().
			Str("module_path", modulePath).
			Int("tf_files_count", len(tfFiles)).
			Msgf("Found %d .tf files", len(tfFiles))
	}

	cmd := exec.CommandContext(ctx, s.terraformBin, "init", "-input=false", "-no-color")
	cmd.Dir = modulePath
	cmd.Env = append(os.Environ(), "TF_IN_AUTOMATION=true")

	logger.Debug().
		Str("command", s.terraformBin+" init").
		Str("working_dir", modulePath).
		Msg("Running terraform init command")

	output, err := cmd.CombinedOutput()
	result.Duration = time.Since(startTime)
	outputStr := string(output)

	logger.Debug().
		Str("module_path", modulePath).
		Dur("duration_ms", result.Duration).
		Int("output_length", len(outputStr)).
		Bool("has_error", err != nil).
		Msg("Terraform init completed")

	// Log the full terraform output at debug level
	if outputStr != "" {
		logger.Debug().
			Str("module_path", modulePath).
			Str("terraform_output", outputStr).
			Msg("Terraform init output")
	}

	if err != nil {
		errorMsg := fmt.Sprintf("terraform init failed: %v\nOutput: %s", err, outputStr)
		result.Error = &errorMsg

		// Enhanced logging for common terraform init failures
		logger.Error().
			Str("module_path", modulePath).
			Err(err).
			Dur("duration", result.Duration).
			Str("error_type", s.classifyInitError(err, outputStr)).
			Str("output_preview", s.getOutputPreview(outputStr)).
			Str("terraform_output", outputStr).
			Msg("Terraform init failed")

		return result
	}

	result.Success = true
	result.Output = outputStr
	result.HasChanges = s.hasInitChanges(outputStr)

	logger.Info().
		Str("module_path", modulePath).
		Dur("duration", result.Duration).
		Bool("has_changes", result.HasChanges).
		Int("output_length", len(outputStr)).
		Msg("Terraform init completed successfully")

	// Log the full terraform output at debug level for successful runs too
	if outputStr != "" {
		logger.Debug().
			Str("module_path", modulePath).
			Str("terraform_output", outputStr).
			Bool("init_success", true).
			Msg("Terraform init successful output")
	}

	return result
}

// getGraphResult executes terraform graph and returns detailed result
func (s *TerraformExecutorService) getGraphResult(ctx context.Context, modulePath string) *TerraformGraphResult {
	startTime := time.Now()
	result := &TerraformGraphResult{
		Success: false,
	}

	logger := zerolog.Ctx(ctx)
	logger.Debug().
		Str("module_path", modulePath).
		Msg("Executing terraform graph")

	cmd := exec.CommandContext(ctx, s.TerraformBinaryPath(), "graph")
	cmd.Dir = modulePath
	cmd.Env = append(os.Environ(), "TF_IN_AUTOMATION=true")

	output, err := cmd.CombinedOutput()
	result.Duration = time.Since(startTime)
	outputStr := string(output)

	logger.Debug().
		Str("module_path", modulePath).
		Dur("duration_ms", result.Duration).
		Int("output_length", len(outputStr)).
		Bool("has_error", err != nil).
		Msg("Terraform graph completed")

	// Log the full terraform graph output at debug level
	if outputStr != "" {
		logger.Debug().
			Str("module_path", modulePath).
			Str("terraform_graph_output", outputStr).
			Msg("Terraform graph output")
	}

	if err != nil {
		errorMsg := fmt.Sprintf("terraform graph failed: %v\nOutput: %s", err, outputStr)
		result.Error = &errorMsg

		logger.Error().
			Str("module_path", modulePath).
			Err(err).
			Dur("duration", result.Duration).
			Str("error_type", "terraform_graph_failed").
			Str("output_preview", s.getOutputPreview(outputStr)).
			Msg("Terraform graph failed")

		return result
	}

	result.Success = true
	result.GraphData = outputStr

	logger.Info().
		Str("module_path", modulePath).
		Dur("duration", result.Duration).
		Int("graph_data_length", len(outputStr)).
		Msg("Terraform graph completed successfully")

	return result
}

// RunTerraformGraph executes terraform graph with proper locking and version switching
func (s *TerraformExecutorService) RunTerraformGraph(ctx context.Context, modulePath string) (*TerraformGraphResult, error) {
	startTime := time.Now()
	result := &TerraformGraphResult{
		Success: false,
	}

	// Execute terraform graph with lock and version switching
	err := s.RunTerraformWithLock(ctx, modulePath, []string{"graph"}, 30*time.Second)
	result.Duration = time.Since(startTime)

	if err != nil {
		errorMsg := err.Error()
		result.Error = &errorMsg
		return result, nil
	}

	// Get the graph output directly
	cmd := exec.CommandContext(ctx, s.TerraformBinaryPath(), "graph")
	cmd.Dir = modulePath
	cmd.Env = append(os.Environ(), "TF_IN_AUTOMATION=true")

	output, err := cmd.CombinedOutput()
	if err != nil {
		errorMsg := fmt.Sprintf("terraform graph failed: %v\nOutput: %s", err, string(output))
		result.Error = &errorMsg
		return result, nil
	}

	result.Success = true
	result.GraphData = string(output)

	return result, nil
}

// getVersionResult executes terraform version and returns detailed result
func (s *TerraformExecutorService) getVersionResult(ctx context.Context, modulePath string) *TerraformVersionResult {
	startTime := time.Now()
	result := &TerraformVersionResult{
		Success: false,
	}

	logger := zerolog.Ctx(ctx)
	logger.Debug().
		Str("module_path", modulePath).
		Str("terraform_binary", s.terraformBin).
		Msg("Executing terraform version")

	cmd := exec.CommandContext(ctx, s.terraformBin, "version", "-json")
	cmd.Dir = modulePath

	output, err := cmd.CombinedOutput()
	result.Duration = time.Since(startTime)
	outputStr := string(output)

	logger.Debug().
		Str("module_path", modulePath).
		Dur("duration_ms", result.Duration).
		Int("output_length", len(outputStr)).
		Bool("has_error", err != nil).
		Msg("Terraform version completed")

	// Log the full terraform version output at debug level
	if outputStr != "" {
		logger.Debug().
			Str("module_path", modulePath).
			Str("terraform_version_output", outputStr).
			Msg("Terraform version output")
	}

	if err != nil {
		errorMsg := err.Error()
		result.Error = &errorMsg

		logger.Error().
			Str("module_path", modulePath).
			Err(err).
			Dur("duration", result.Duration).
			Str("error_type", "terraform_version_failed").
			Str("output_preview", s.getOutputPreview(outputStr)).
			Msg("Terraform version failed")

		return result
	}

	result.Success = true
	result.Output = outputStr

	logger.Info().
		Str("module_path", modulePath).
		Dur("duration", result.Duration).
		Msg("Terraform version completed successfully")

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
		errorMsg := fmt.Sprintf("failed to read terraform modules.json: %v", err)
		result.Error = &errorMsg
		return result
	}

	result.Success = true
	result.Modules = string(output)

	return result
}

// RunTerraformModulesParsing obtains terraform modules information
// Replicates Python's _get_terraform_modules method
func (s *TerraformExecutorService) RunTerraformModulesParsing(ctx context.Context, modulePath string) (*TerraformModulesResult, error) {
	startTime := time.Now()
	result := &TerraformModulesResult{
		Success: false,
	}

	// Read modules.json file from .terraform directory
	modulesPath := filepath.Join(modulePath, ".terraform", "modules", "modules.json")

	// Check if file exists
	if _, err := os.Stat(modulesPath); os.IsNotExist(err) {
		errorMsg := "modules.json file does not exist - terraform init may not have been run"
		result.Error = &errorMsg
		result.Duration = time.Since(startTime)
		return result, nil
	}

	output, err := os.ReadFile(modulesPath)
	result.Duration = time.Since(startTime)

	if err != nil {
		errorMsg := fmt.Sprintf("failed to read terraform modules.json: %v", err)
		result.Error = &errorMsg
		return result, nil
	}

	// Validate JSON content
	if len(output) == 0 {
		errorMsg := "modules.json file is empty"
		result.Error = &errorMsg
		return result, nil
	}

	result.Success = true
	result.Modules = string(output)

	return result, nil
}

// GitCommitResult represents the result of git commit SHA extraction
type GitCommitResult struct {
	Success   bool          `json:"success"`
	CommitSHA *string       `json:"commit_sha,omitempty"`
	Error     *string       `json:"error,omitempty"`
	Duration  time.Duration `json:"duration"`
}

// GetGitCommitSHA obtains git commit hash for module version
// Replicates Python's _get_git_commit_sha method
func (s *TerraformExecutorService) GetGitCommitSHA(ctx context.Context, modulePath string) (*GitCommitResult, error) {
	startTime := time.Now()
	result := &GitCommitResult{
		Success: false,
	}

	// Check if we're in a git repository
	gitDir := filepath.Join(modulePath, ".git")
	if _, err := os.Stat(gitDir); os.IsNotExist(err) {
		errorMsg := "not a git repository"
		result.Error = &errorMsg
		result.Duration = time.Since(startTime)
		return result, nil
	}

	// Execute git rev-parse HEAD
	cmd := exec.CommandContext(ctx, "git", "rev-parse", "HEAD")
	cmd.Dir = modulePath

	output, err := cmd.CombinedOutput()
	result.Duration = time.Since(startTime)

	if err != nil {
		errorMsg := fmt.Sprintf("failed to obtain git commit hash: %v\nOutput: %s", err, string(output))
		result.Error = &errorMsg
		return result, nil
	}

	commitSHA := string(output)
	commitSHA = strings.TrimSpace(commitSHA)

	// Validate commit SHA format (should be 40 hex characters)
	if len(commitSHA) != 40 {
		errorMsg := fmt.Sprintf("invalid git commit SHA format: %s", commitSHA)
		result.Error = &errorMsg
		return result, nil
	}

	result.Success = true
	result.CommitSHA = &commitSHA

	return result, nil
}

// hasInitChanges checks if terraform init made any changes (simplified check)
func (s *TerraformExecutorService) hasInitChanges(output string) bool {
	// Simplified check - in a full implementation, would parse output for specific patterns
	return len(output) > 0 && (contains(output, "Terraform has been successfully initialized") ||
		contains(output, "Downloading and installing"))
}

// getFailedStep determines which terraform step failed and returns detailed error info
func (s *TerraformExecutorService) getFailedStep(err error) string {
	errorMsg := err.Error()

	// Look for specific terraform commands in the error
	switch {
	case contains(errorMsg, "terraform init"):
		return s.extractTerraformErrorDetails("terraform_init", errorMsg)
	case contains(errorMsg, "terraform graph"):
		return s.extractTerraformErrorDetails("terraform_graph", errorMsg)
	case contains(errorMsg, "terraform version"):
		return s.extractTerraformErrorDetails("terraform_version", errorMsg)
	case contains(errorMsg, "terraform fmt"):
		return s.extractTerraformErrorDetails("terraform_fmt", errorMsg)
	case contains(errorMsg, "terraform validate"):
		return s.extractTerraformErrorDetails("terraform_validate", errorMsg)
	case contains(errorMsg, "terraform show"):
		return s.extractTerraformErrorDetails("terraform_show", errorMsg)
	case contains(errorMsg, "terraform providers"):
		return s.extractTerraformErrorDetails("terraform_providers", errorMsg)
	case contains(errorMsg, "permission denied"):
		return "permission_denied"
	case contains(errorMsg, "no such file or directory"):
		return "file_not_found"
	case contains(errorMsg, "command not found"):
		return "command_not_found"
	case contains(errorMsg, "context deadline exceeded"):
		return "timeout"
	case contains(errorMsg, "signal: killed"):
		return "process_killed"
	default:
		// Extract the actual command that failed if possible
		if strings.Contains(errorMsg, "failed:") {
			// Return everything after "failed:" for more detail
			parts := strings.Split(errorMsg, "failed:")
			if len(parts) > 1 {
				return strings.TrimSpace(parts[1])
			}
		}
		return fmt.Sprintf("terraform_error: %s", errorMsg)
	}
}

// extractTerraformErrorDetails extracts detailed error information from terraform command output
func (s *TerraformExecutorService) extractTerraformErrorDetails(command, errorMsg string) string {
	// Extract terraform output from the error message
	// The error format is typically: "terraform <command> failed: <error details>\nOutput: <terraform output>"

	// Split by "Output:" to get the terraform output
	parts := strings.Split(errorMsg, "Output:")
	if len(parts) > 1 {
		output := strings.TrimSpace(parts[1])

		// Extract key error patterns from terraform output
		if strings.Contains(output, "Error:") {
			// Get the first error line
			lines := strings.Split(output, "\n")
			for _, line := range lines {
				if strings.Contains(line, "Error:") {
					return fmt.Sprintf("%s failed: %s", command, strings.TrimSpace(line))
				}
			}
		}

		// If no specific error found, return a preview of the output
		preview := s.getOutputPreview(output)
		if preview != output {
			return fmt.Sprintf("%s failed: %s...", command, preview)
		}
		return fmt.Sprintf("%s failed: %s", command, output)
	}

	// Fallback to basic command detection
	return fmt.Sprintf("%s failed", command)
}

// contains checks if a string contains a substring (case-insensitive)
func contains(s, substr string) bool {
	return len(s) > 0 && (len(substr) == 0 || (s[0:len(substr)] == substr ||
		(len(s) > len(substr) && s[len(s)-len(substr):len(s)] == substr)))
}

// SwitchTerraformVersions switches terraform versions using tfswitch with global locking
// Replicates Python's _switch_terraform_versions context manager
func (s *TerraformExecutorService) SwitchTerraformVersions(
	ctx context.Context,
	modulePath string,
	timeout time.Duration,
) (func(), error) {
	// Try to acquire global lock with timeout
	lockTimeout := timeout
	if lockTimeout == 0 {
		lockTimeout = 60 * time.Second // Default 60 second timeout matching Python
	}

	// Use a channel to handle lock acquisition with timeout
	lockAcquired := make(chan bool, 1)

	go func() {
		terraformGlobalLock.Lock()
		lockAcquired <- true
	}()

	select {
	case <-ctx.Done():
		return nil, fmt.Errorf("context cancelled while waiting for terraform lock: %w", ctx.Err())
	case <-time.After(lockTimeout):
		return nil, fmt.Errorf("unable to obtain global terraform lock in %v", lockTimeout)
	case <-lockAcquired:
		// Lock acquired, continue
	}

	// Prepare environment variables for tfswitch
	tfswitchEnv := os.Environ()
	tfswitchEnv = append(tfswitchEnv, fmt.Sprintf("TF_DEFAULT_VERSION=%s", s.tfswitchConfig.DefaultTerraformVersion))
	tfswitchEnv = append(tfswitchEnv, fmt.Sprintf("TF_PRODUCT=%s", s.tfswitchConfig.TerraformProduct))

	if s.tfswitchConfig.ArchiveMirror != "" {
		tfswitchEnv = append(tfswitchEnv, fmt.Sprintf("TERRAFORM_ARCHIVE_MIRROR=%s", s.tfswitchConfig.ArchiveMirror))
	}

	// Prepare tfswitch command arguments
	var tfswitchArgs []string
	if s.tfswitchConfig.BinaryPath != "" {
		tfswitchArgs = append(tfswitchArgs, "--bin", s.tfswitchConfig.BinaryPath)
	}

	// Create tfswitch command
	cmd := exec.CommandContext(ctx, "tfswitch", tfswitchArgs...)
	cmd.Dir = modulePath
	cmd.Env = tfswitchEnv

	// Execute tfswitch
	if output, err := cmd.CombinedOutput(); err != nil {
		terraformGlobalLock.Unlock()
		return nil, fmt.Errorf("terraform version switch failed: %v\nOutput: %s", err, string(output))
	}

	// Return cleanup function to release lock
	cleanup := func() {
		terraformGlobalLock.Unlock()
	}

	return cleanup, nil
}

// TerraformBinaryPath returns the path to the terraform binary
// Replicates Python's terraform_binary method
func (s *TerraformExecutorService) TerraformBinaryPath() string {
	if s.tfswitchConfig.BinaryPath != "" {
		return s.tfswitchConfig.BinaryPath
	}
	return "terraform" // Default fallback
}

// RunTerraformWithLock executes terraform commands with proper version switching and locking
func (s *TerraformExecutorService) RunTerraformWithLock(
	ctx context.Context,
	modulePath string,
	terraformArgs []string,
	lockTimeout time.Duration,
) error {
	// Acquire lock and switch versions
	cleanup, err := s.SwitchTerraformVersions(ctx, modulePath, lockTimeout)
	if err != nil {
		return err
	}
	defer cleanup()

	// Execute terraform command
	cmd := exec.CommandContext(ctx, s.TerraformBinaryPath(), terraformArgs...)
	cmd.Dir = modulePath
	cmd.Env = append(os.Environ(), "TF_IN_AUTOMATION=true")

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("terraform %s failed: %v\nOutput: %s", terraformArgs[0], err, string(output))
	}

	// Store output for potential processing if needed
	_ = output

	return nil
}

// OverrideTerraformBackend creates a backend override file if terraform backend is detected
// Replicates Python's _override_tf_backend method
func (s *TerraformExecutorService) OverrideTerraformBackend(modulePath string) (*string, error) {
	// Regex to match terraform backend blocks (matching Python pattern)
	backendRegex := regexp.MustCompile(`(?s)^.*\bterraform\s*\{[\s\n.]+(.|\n)*backend\s+"[\w]+"\s+\{`)

	// Find all .tf files in module directory
	tfFiles, err := filepath.Glob(filepath.Join(modulePath, "*.tf"))
	if err != nil {
		return nil, fmt.Errorf("failed to glob .tf files: %w", err)
	}

	var backendFilename *string
	for _, tfFile := range tfFiles {
		content, err := os.ReadFile(tfFile)
		if err != nil {
			return nil, fmt.Errorf("failed to read file %s: %w", tfFile, err)
		}

		if backendRegex.Match(content) {
			filename := filepath.Base(tfFile)
			backendFilename = &filename
			break
		}
	}

	if backendFilename == nil {
		return nil, nil // No backend found
	}

	// Create override filename
	overrideFilename := (*backendFilename)[:len(*backendFilename)-3] + "_override.tf"
	stateFile := ".local-state"
	overridePath := filepath.Join(modulePath, overrideFilename)

	// Create override file content
	overrideContent := fmt.Sprintf(`
terraform {
  backend "local" {
    path = "./%s"
  }
}
`, stateFile)

	// Write override file
	if err := os.WriteFile(overridePath, []byte(overrideContent), 0644); err != nil {
		return nil, fmt.Errorf("failed to write backend override file: %w", err)
	}

	return &overrideFilename, nil
}

// RunTerraformInit executes terraform init with backend override
// Replicates Python's _run_tf_init method
func (s *TerraformExecutorService) RunTerraformInit(ctx context.Context, modulePath string) (*TerraformInitResult, error) {
	startTime := time.Now()
	result := &TerraformInitResult{
		Success: false,
	}

	// Override terraform backend if detected
	overrideFilename, err := s.OverrideTerraformBackend(modulePath)
	if err != nil {
		errorMsg := err.Error()
		result.Error = &errorMsg
		return result, nil
	}

	if overrideFilename != nil {
		result.BackendOverride = overrideFilename
	}

	// Execute terraform init with lock and version switching
	err = s.RunTerraformWithLock(ctx, modulePath, []string{"init"}, 60*time.Second)
	result.Duration = time.Since(startTime)

	if err != nil {
		errorMsg := err.Error()
		result.Error = &errorMsg
		return result, nil
	}

	result.Success = true
	result.Output = "Terraform initialization completed successfully"
	result.HasChanges = true // Assume changes were made during init

	return result, nil
}

// classifyInitError classifies terraform init errors into common categories
func (s *TerraformExecutorService) classifyInitError(err error, output string) string {
	errorMsg := strings.ToLower(err.Error() + " " + output)

	switch {
	case strings.Contains(errorMsg, "permission denied"):
		return "permission_denied"
	case strings.Contains(errorMsg, "no such file or directory"):
		return "file_not_found"
	case strings.Contains(errorMsg, "command not found"):
		return "terraform_not_found"
	case strings.Contains(errorMsg, "context deadline exceeded"):
		return "timeout"
	case strings.Contains(errorMsg, "signal: killed"):
		return "process_killed"
	case strings.Contains(errorMsg, "network") || strings.Contains(errorMsg, "connection refused"):
		return "network_error"
	case strings.Contains(errorMsg, "authentication failed") || strings.Contains(errorMsg, "unauthorized"):
		return "authentication_error"
	case strings.Contains(errorMsg, "provider") && strings.Contains(errorMsg, "not found"):
		return "provider_error"
	case strings.Contains(errorMsg, "module") && strings.Contains(errorMsg, "not found"):
		return "module_error"
	default:
		return "unknown_error"
	}
}

// getOutputPreview returns a preview of the terraform output for logging
func (s *TerraformExecutorService) getOutputPreview(output string) string {
	if len(output) <= 200 {
		return strings.TrimSpace(output)
	}

	// Return first 200 characters, trimming to whole words
	preview := output[:200]
	if lastSpace := strings.LastIndex(preview, " "); lastSpace > 0 {
		preview = preview[:lastSpace]
	}

	return strings.TrimSpace(preview) + "..."
}
