package mocks

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/shared/service"
)

// MockSystemCommandService provides a mock implementation for testing
// It implements the service.SystemCommandService interface
type MockSystemCommandService struct {
	mu             sync.RWMutex
	mockOutputs    map[string]*service.CommandResult
	mockErrors     map[string]error
	commandHistory []*CommandCall
	shouldFail     bool
	failureMessage string
}

// CommandCall records a command invocation for testing assertions
type CommandCall struct {
	Command *service.Command
	Input   string
}

// NewMockSystemCommandService creates a new mock system command service
func NewMockSystemCommandService() *MockSystemCommandService {
	return &MockSystemCommandService{
		mockOutputs:    make(map[string]*service.CommandResult),
		mockErrors:     make(map[string]error),
		commandHistory: make([]*CommandCall, 0),
	}
}

// Execute executes a command and returns the mocked result
func (m *MockSystemCommandService) Execute(ctx context.Context, cmd *service.Command) (*service.CommandResult, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Record the call
	m.commandHistory = append(m.commandHistory, &CommandCall{
		Command: cmd,
		Input:   "",
	})

	// Check if we should fail
	if m.shouldFail {
		return nil, fmt.Errorf("command execution failed: %s", m.failureMessage)
	}

	// Generate command key
	key := m.commandKey(cmd.Name, cmd.Args)

	// Check for mock error
	if err, ok := m.mockErrors[key]; ok {
		return nil, err
	}

	// Return mock output if exists
	if result, ok := m.mockOutputs[key]; ok {
		return result, nil
	}

	// Default empty result for commands without explicit mocks
	return &service.CommandResult{
		Stdout:   "",
		Stderr:   "",
		ExitCode: 0,
	}, nil
}

// ExecuteWithInput executes a command with input and returns the mocked result
func (m *MockSystemCommandService) ExecuteWithInput(ctx context.Context, cmd *service.Command, input string) (*service.CommandResult, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Record the call
	m.commandHistory = append(m.commandHistory, &CommandCall{
		Command: cmd,
		Input:   input,
	})

	// Check if we should fail
	if m.shouldFail {
		return nil, fmt.Errorf("command execution failed: %s", m.failureMessage)
	}

	// Generate command key
	key := m.commandKey(cmd.Name, cmd.Args)

	// Check for mock error
	if err, ok := m.mockErrors[key]; ok {
		return nil, err
	}

	// Return mock output if exists
	if result, ok := m.mockOutputs[key]; ok {
		return result, nil
	}

	// Default empty result for commands without explicit mocks
	return &service.CommandResult{
		Stdout:   "",
		Stderr:   "",
		ExitCode: 0,
	}, nil
}

// SetMockOutput sets a mock output for a specific command
func (m *MockSystemCommandService) SetMockOutput(command string, args []string, result *service.CommandResult) {
	m.mu.Lock()
	defer m.mu.Unlock()

	key := m.commandKey(command, args)
	m.mockOutputs[key] = result
}

// SetMockError sets a mock error for a specific command
func (m *MockSystemCommandService) SetMockError(command string, args []string, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	key := m.commandKey(command, args)
	m.mockErrors[key] = err
}

// SetMockOutputByPattern sets mock output for commands matching a pattern
func (m *MockSystemCommandService) SetMockOutputByPattern(pattern string, result *service.CommandResult) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.mockOutputs[pattern] = result
}

// SetGlobalFailure configures all commands to fail
func (m *MockSystemCommandService) SetGlobalFailure(fail bool, message string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.shouldFail = fail
	m.failureMessage = message
}

// GetCommandHistory returns the history of executed commands
func (m *MockSystemCommandService) GetCommandHistory() []*CommandCall {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Return a copy to prevent external modification
	history := make([]*CommandCall, len(m.commandHistory))
	copy(history, m.commandHistory)
	return history
}

// ClearHistory clears the command history
func (m *MockSystemCommandService) ClearHistory() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.commandHistory = make([]*CommandCall, 0)
}

// ClearMocks clears all mock outputs and errors
func (m *MockSystemCommandService) ClearMocks() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.mockOutputs = make(map[string]*service.CommandResult)
	m.mockErrors = make(map[string]error)
}

// WasCommandExecuted checks if a specific command was executed
func (m *MockSystemCommandService) WasCommandExecuted(command string, args []string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	targetKey := m.commandKey(command, args)
	for _, call := range m.commandHistory {
		key := m.commandKey(call.Command.Name, call.Command.Args)
		if key == targetKey {
			return true
		}
	}
	return false
}

// GetExecutionCount returns how many times a command was executed
func (m *MockSystemCommandService) GetExecutionCount(command string, args []string) int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	targetKey := m.commandKey(command, args)
	count := 0
	for _, call := range m.commandHistory {
		key := m.commandKey(call.Command.Name, call.Command.Args)
		if key == targetKey {
			count++
		}
	}
	return count
}

// commandKey generates a unique key for a command
func (m *MockSystemCommandService) commandKey(command string, args []string) string {
	return command + " " + strings.Join(args, " ")
}

// Helper methods to set up common mocks

// SetupTerraformDocsMock sets up mock output for terraform-docs
func (m *MockSystemCommandService) SetupTerraformDocsMock(output string) {
	m.SetMockOutput("terraform-docs", []string{"json", "."}, &service.CommandResult{
		Stdout:   output,
		Stderr:   "",
		ExitCode: 0,
	})
}

// SetupTerraformDocsError sets up mock error for terraform-docs
func (m *MockSystemCommandService) SetupTerraformDocsError(message string) {
	m.SetMockError("terraform-docs", []string{"json", "."}, fmt.Errorf("terraform-docs error: %s", message))
}

// SetupTfsecMock sets up mock output for tfsec
func (m *MockSystemCommandService) SetupTfsecMock(output string) {
	m.SetMockOutput("tfsec", []string{"--format", "json", "--out", "-"}, &service.CommandResult{
		Stdout:   output,
		Stderr:   "",
		ExitCode: 0,
	})
}

// SetupTfsecError sets up mock error for tfsec
func (m *MockSystemCommandService) SetupTfsecError(message string) {
	m.SetMockError("tfsec", []string{"--format", "json", "--out", "-"}, fmt.Errorf("tfsec error: %s", message))
}

// SetupInfracostMock sets up mock output for infracost
func (m *MockSystemCommandService) SetupInfracostMock(output string) {
	m.SetMockOutput("infracost", []string{"breakdown", "--format", "json"}, &service.CommandResult{
		Stdout:   output,
		Stderr:   "",
		ExitCode: 0,
	})
}

// SetupInfracostError sets up mock error for infracost
func (m *MockSystemCommandService) SetupInfracostError(message string) {
	m.SetMockError("infracost", []string{"breakdown", "--format", "json"}, fmt.Errorf("infracost error: %s", message))
}

// SetupTerraformMock sets up mock output for terraform commands
func (m *MockSystemCommandService) SetupTerraformMock(subcommand string, output string) {
	m.SetMockOutput("terraform", []string{subcommand, "-json"}, &service.CommandResult{
		Stdout:   output,
		Stderr:   "",
		ExitCode: 0,
	})
}

// SetupTerraformError sets up mock error for terraform commands
func (m *MockSystemCommandService) SetupTerraformError(subcommand string, message string) {
	m.SetMockError("terraform", []string{subcommand, "-json"}, fmt.Errorf("terraform error: %s", message))
}

// SetupGitMock sets up mock output for git commands
func (m *MockSystemCommandService) SetupGitMock(subcommand string, args []string, output string) {
	gitArgs := append([]string{subcommand}, args...)
	m.SetMockOutput("git", gitArgs, &service.CommandResult{
		Stdout:   output,
		Stderr:   "",
		ExitCode: 0,
	})
}

// Ensure MockSystemCommandService implements service.SystemCommandService
var _ service.SystemCommandService = (*MockSystemCommandService)(nil)
