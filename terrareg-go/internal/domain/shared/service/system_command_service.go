package service

import (
	"context"
	"os/exec"
)

// SystemCommandService defines the interface for executing system commands.
// This abstraction allows for mocking in tests while using real exec.Command in production.
type SystemCommandService interface {
	// Execute runs a command and returns its output
	Execute(ctx context.Context, cmd *Command) (*CommandResult, error)
	// ExecuteWithInput runs a command with input and returns its output
	ExecuteWithInput(ctx context.Context, cmd *Command, input string) (*CommandResult, error)
}

// Command represents a system command to execute
type Command struct {
	Name string
	Args []string
	Dir  string
	Env  []string
}

// CommandResult represents the result of executing a command
type CommandResult struct {
	Stdout   string
	Stderr   string
	ExitCode int
}

// RealSystemCommandService provides the actual implementation using os/exec
type RealSystemCommandService struct{}

// NewRealSystemCommandService creates a new real system command service
func NewRealSystemCommandService() *RealSystemCommandService {
	return &RealSystemCommandService{}
}

// Execute executes a command and returns its result
func (s *RealSystemCommandService) Execute(ctx context.Context, cmd *Command) (*CommandResult, error) {
	execCmd := exec.CommandContext(ctx, cmd.Name, cmd.Args...)
	if cmd.Dir != "" {
		execCmd.Dir = cmd.Dir
	}
	if len(cmd.Env) > 0 {
		execCmd.Env = cmd.Env
	}

	stdout, err := execCmd.Output()
	if err != nil {
		// Try to get combined output for better error messages
		if combinedErr, ok := err.(*exec.ExitError); ok {
			return &CommandResult{
				Stdout:   string(stdout),
				Stderr:   string(combinedErr.Stderr),
				ExitCode: combinedErr.ExitCode(),
			}, err
		}
		return &CommandResult{
			Stdout:   string(stdout),
			Stderr:   err.Error(),
			ExitCode: -1,
		}, err
	}

	return &CommandResult{
		Stdout:   string(stdout),
		Stderr:   "",
		ExitCode: 0,
	}, nil
}

// ExecuteWithInput executes a command with input and returns its result
func (s *RealSystemCommandService) ExecuteWithInput(ctx context.Context, cmd *Command, input string) (*CommandResult, error) {
	execCmd := exec.CommandContext(ctx, cmd.Name, cmd.Args...)
	if cmd.Dir != "" {
		execCmd.Dir = cmd.Dir
	}
	if len(cmd.Env) > 0 {
		execCmd.Env = cmd.Env
	}

	stdin, err := execCmd.StdinPipe()
	if err != nil {
		return nil, err
	}

	go func() {
		defer stdin.Close()
		stdin.Write([]byte(input))
	}()

	output, err := execCmd.CombinedOutput()
	if err != nil {
		exitCode := 1
		if exitError, ok := err.(*exec.ExitError); ok {
			exitCode = exitError.ExitCode()
		}
		return &CommandResult{
			Stdout:   string(output),
			Stderr:   err.Error(),
			ExitCode: exitCode,
		}, err
	}

	return &CommandResult{
		Stdout:   string(output),
		Stderr:   "",
		ExitCode: 0,
	}, nil
}

// CombinedOutput executes a command and returns combined stdout/stderr
func (s *RealSystemCommandService) CombinedOutput(ctx context.Context, cmd *Command) (*CommandResult, error) {
	execCmd := exec.CommandContext(ctx, cmd.Name, cmd.Args...)
	if cmd.Dir != "" {
		execCmd.Dir = cmd.Dir
	}
	if len(cmd.Env) > 0 {
		execCmd.Env = cmd.Env
	}

	output, err := execCmd.CombinedOutput()
	if err != nil {
		exitCode := 1
		if exitError, ok := err.(*exec.ExitError); ok {
			exitCode = exitError.ExitCode()
		}
		return &CommandResult{
			Stdout:   string(output),
			Stderr:   "",
			ExitCode: exitCode,
		}, err
	}

	return &CommandResult{
		Stdout:   string(output),
		Stderr:   "",
		ExitCode: 0,
	}, nil
}
