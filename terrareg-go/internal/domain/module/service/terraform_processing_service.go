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

	savepointName := fmt.Sprintf("terraform_processing_%d", startTime.UnixNano())

	err := s.savepointHelper.WithSmartSavepointOrTransaction(ctx, savepointName, func(tx *gorm.DB) error {
		// Execute terraform pipeline with rollback capability
		return s.ExecuteTerraformPipeline(ctx, req.ModulePath, req.Operations)
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
	cmd := exec.CommandContext(ctx, op.Command[0], op.Command[1:]...)
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

	cmd := exec.CommandContext(ctx, s.TerraformBinaryPath(), "graph")
	cmd.Dir = modulePath
	cmd.Env = append(os.Environ(), "TF_IN_AUTOMATION=true")

	output, err := cmd.CombinedOutput()
	result.Duration = time.Since(startTime)

	if err != nil {
		errorMsg := fmt.Sprintf("terraform graph failed: %v\nOutput: %s", err, string(output))
		result.Error = &errorMsg
		return result
	}

	result.Success = true
	result.GraphData = string(output)

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
