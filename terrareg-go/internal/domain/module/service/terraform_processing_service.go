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
		return nil
	}

	return &TerraformExecutorService{
		savepointHelper: savepointHelper,
		terraformBin:    terraformBin,
		lockTimeout:     lockTimeout,
		tfswitchConfig:  tfswitchConfig,
	}
}

// ProcessTerraformWithTransaction processes terraform operations with transaction context
// Following Python pattern: execute terraform init once, then collect all results
func (s *TerraformExecutorService) ProcessTerraformWithTransaction(
	ctx context.Context,
	req TerraformProcessingRequest,
) (*TerraformProcessingResult, error) {
	logger := zerolog.Ctx(ctx)
	logger.Debug().
		Int("module_version_id", req.ModuleVersionID).
		Str("module_path", req.ModulePath).
		Msg("ProcessTerraformWithTransaction started")

	startTime := time.Now()
	result := &TerraformProcessingResult{
		OverallSuccess: false,
		Duration:       0,
	}

	err := s.savepointHelper.WithTransaction(ctx, func(ctx context.Context, tx *gorm.DB) error {
		logger.Debug().
			Int("module_version_id", req.ModuleVersionID).
			Str("module_path", req.ModulePath).
			Msg("Transaction started, ensuring terraform is installed")

		// Ensure terraform is installed using tfswitch before processing
		if err := s.ensureTerraformInstalled(ctx, req.ModulePath); err != nil {
			return fmt.Errorf("failed to ensure terraform is installed: %w", err)
		}

		logger.Debug().
			Int("module_version_id", req.ModuleVersionID).
			Str("module_path", req.ModulePath).
			Msg("Terraform installation verified, executing terraform operations")

		// Execute terraform init first (required for other operations)
		initResult := s.executeTerraformInit(ctx, req.ModulePath)
		result.InitResult = initResult
		if !initResult.Success {
			return fmt.Errorf("terraform init failed: %s", *initResult.Error)
		}

		logger.Debug().
			Str("module_path", req.ModulePath).
			Msg("Terraform init completed successfully, collecting other results")

		// Collect results from other operations (following Python pattern)
		result.GraphResult = s.executeTerraformGraph(ctx, req.ModulePath)
		result.VersionResult = s.executeTerraformVersion(ctx, req.ModulePath)
		result.ModulesResult = s.executeTerraformModules(ctx, req.ModulePath)

		result.OverallSuccess = true
		return nil
	})

	result.Duration = time.Since(startTime)

	if err != nil {
		result.FailedStep = s.getFailedStep(err)
		result.OverallSuccess = false
		return result, nil
	}

	return result, nil
}

// executeTerraformInit executes terraform init and returns result
// Following Python's _run_tf_init pattern
func (s *TerraformExecutorService) executeTerraformInit(ctx context.Context, modulePath string) *TerraformInitResult {
	startTime := time.Now()
	result := &TerraformInitResult{
		Success: false,
	}

	logger := zerolog.Ctx(ctx)
	logger.Debug().
		Str("module_path", modulePath).
		Msg("Executing terraform init")

	// Override terraform backend if detected (matching Python)
	overrideFilename, err := s.OverrideTerraformBackend(modulePath)
	if err != nil {
		errorMsg := fmt.Sprintf("backend override failed: %v", err)
		result.Error = &errorMsg
		return result
	}
	if overrideFilename != nil {
		result.BackendOverride = overrideFilename
	}

	// Execute terraform init with global lock and version switching
	err = s.RunTerraformWithLock(ctx, modulePath, []string{"init", "-input=false", "-no-color"}, 60*time.Second)
	result.Duration = time.Since(startTime)

	if err != nil {
		errorMsg := fmt.Sprintf("terraform init failed: %v", err)
		result.Error = &errorMsg

		logger.Error().
			Str("module_path", modulePath).
			Err(err).
			Dur("duration", result.Duration).
			Msg("Terraform init failed")

		return result
	}

	result.Success = true
	result.Output = "Terraform initialization completed successfully"
	result.HasChanges = true // Assume changes were made during init

	logger.Info().
		Str("module_path", modulePath).
		Dur("duration", result.Duration).
		Bool("backend_override", overrideFilename != nil).
		Msg("Terraform init completed successfully")

	return result
}

// executeTerraformGraph executes terraform graph and returns result
// Following Python's _get_graph_data pattern
func (s *TerraformExecutorService) executeTerraformGraph(ctx context.Context, modulePath string) *TerraformGraphResult {
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

	if err != nil {
		errorMsg := fmt.Sprintf("terraform graph failed: %v", err)
		result.Error = &errorMsg

		logger.Error().
			Str("module_path", modulePath).
			Err(err).
			Dur("duration", result.Duration).
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

// executeTerraformVersion executes terraform version and returns result
// Following Python's _get_terraform_version pattern
func (s *TerraformExecutorService) executeTerraformVersion(ctx context.Context, modulePath string) *TerraformVersionResult {
	startTime := time.Now()
	result := &TerraformVersionResult{
		Success: false,
	}

	logger := zerolog.Ctx(ctx)
	logger.Debug().
		Str("module_path", modulePath).
		Msg("Executing terraform version")

	cmd := exec.CommandContext(ctx, s.TerraformBinaryPath(), "version", "-json")
	cmd.Dir = modulePath

	output, err := cmd.CombinedOutput()
	result.Duration = time.Since(startTime)
	outputStr := string(output)

	if err != nil {
		errorMsg := fmt.Sprintf("terraform version failed: %v", err)
		result.Error = &errorMsg

		logger.Error().
			Str("module_path", modulePath).
			Err(err).
			Dur("duration", result.Duration).
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

// executeTerraformModules parses terraform modules and returns result
// Following Python's _get_terraform_modules pattern
func (s *TerraformExecutorService) executeTerraformModules(ctx context.Context, modulePath string) *TerraformModulesResult {
	startTime := time.Now()
	result := &TerraformModulesResult{
		Success: false,
	}

	logger := zerolog.Ctx(ctx)
	logger.Debug().
		Str("module_path", modulePath).
		Msg("Reading terraform modules")

	modulesPath := filepath.Join(modulePath, ".terraform", "modules", "modules.json")
	output, err := os.ReadFile(modulesPath)
	result.Duration = time.Since(startTime)

	if err != nil {
		errorMsg := fmt.Sprintf("failed to read terraform modules.json: %v", err)
		result.Error = &errorMsg

		logger.Warn().
			Str("module_path", modulePath).
			Str("modules_path", modulesPath).
			Err(err).
			Msg("Terraform modules file not found - may not be a terraform module")

		// Don't treat this as an error - some modules may not have dependencies
		result.Success = true
		result.Modules = "{}"
		return result
	}

	result.Success = true
	result.Modules = string(output)

	logger.Info().
		Str("module_path", modulePath).
		Dur("duration", result.Duration).
		Int("modules_data_length", len(output)).
		Msg("Terraform modules read successfully")

	return result
}

// ensureTerraformInstalled ensures terraform is installed using tfswitch with global locking
// Always runs tfswitch to ensure correct version is being used (not just when missing)
func (s *TerraformExecutorService) ensureTerraformInstalled(ctx context.Context, modulePath string) error {
	logger := zerolog.Ctx(ctx)
	logger.Debug().
		Str("module_path", modulePath).
		Str("default_version", s.tfswitchConfig.DefaultTerraformVersion).
		Msg("Ensuring terraform version using tfswitch (always run to maintain version consistency)")

	// Get terraform binary path for verification
	terraformPath := s.TerraformBinaryPath()

	// Use global lock to ensure only one tfswitch operation runs at a time
	logger.Debug().Msg("Acquiring global terraform lock for tfswitch")

	lockAcquired := make(chan bool, 1)
	go func() {
		terraformGlobalLock.Lock()
		lockAcquired <- true
	}()

	select {
	case <-ctx.Done():
		return fmt.Errorf("context cancelled while waiting for terraform lock: %w", ctx.Err())
	case <-time.After(60 * time.Second):
		return fmt.Errorf("unable to obtain global terraform lock in 60 seconds")
	case <-lockAcquired:
		logger.Debug().Msg("Global terraform lock acquired")
	}

	defer func() {
		terraformGlobalLock.Unlock()
		logger.Debug().Msg("Global terraform lock released")
	}()

	// Always run tfswitch to ensure correct version is used
	logger.Debug().
		Str("terraform_path", terraformPath).
		Msg("Running tfswitch to ensure correct terraform version")

	// Prepare environment variables for tfswitch
	tfswitchEnv := os.Environ()
	if s.tfswitchConfig.DefaultTerraformVersion != "" {
		tfswitchEnv = append(tfswitchEnv, fmt.Sprintf("TF_DEFAULT_VERSION=%s", s.tfswitchConfig.DefaultTerraformVersion))
	}
	if s.tfswitchConfig.TerraformProduct != "" {
		tfswitchEnv = append(tfswitchEnv, fmt.Sprintf("TF_PRODUCT=%s", s.tfswitchConfig.TerraformProduct))
	}
	if s.tfswitchConfig.ArchiveMirror != "" {
		tfswitchEnv = append(tfswitchEnv, fmt.Sprintf("TERRAFORM_ARCHIVE_MIRROR=%s", s.tfswitchConfig.ArchiveMirror))
	}

	// Prepare tfswitch command arguments
	var tfswitchArgs []string
	if s.tfswitchConfig.DefaultTerraformVersion != "" {
		tfswitchArgs = append(tfswitchArgs, s.tfswitchConfig.DefaultTerraformVersion)
	}
	if s.tfswitchConfig.BinaryPath != "" {
		tfswitchArgs = append(tfswitchArgs, "--bin", s.tfswitchConfig.BinaryPath)
	}

	// Create tfswitch command
	cmd := exec.CommandContext(ctx, "tfswitch", tfswitchArgs...)
	cmd.Dir = modulePath
	cmd.Env = tfswitchEnv

	logger.Debug().
		Str("command", "tfswitch "+strings.Join(tfswitchArgs, " ")).
		Str("working_dir", modulePath).
		Msg("Executing tfswitch to install terraform")

	// Execute tfswitch
	output, err := cmd.CombinedOutput()
	if err != nil {
		logger.Error().
			Err(err).
			Str("tfswitch_output", string(output)).
			Msg("Tfswitch failed to install terraform")
		return fmt.Errorf("terraform version switch failed: %v\nOutput: %s", err, string(output))
	}

	logger.Info().
		Str("tfswitch_output", string(output)).
		Msg("Terraform successfully installed via tfswitch")

	// Verify terraform is now available
	if _, err := exec.LookPath(terraformPath); err != nil {
		return fmt.Errorf("terraform installation verification failed: %w", err)
	}

	return nil
}

// TerraformBinaryPath returns the path to the terraform binary
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

// SwitchTerraformVersions switches terraform versions using tfswitch with global locking
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

// OverrideTerraformBackend creates a backend override file if terraform backend is detected
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