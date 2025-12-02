package module

import (
	"context"

	moduleService "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/module/service"
)

// ImportModuleVersionCommand handles importing module versions from Git
type ImportModuleVersionCommand struct {
	moduleImporterService *moduleService.ModuleImporterService
}

// NewImportModuleVersionCommand creates a new command
func NewImportModuleVersionCommand(
	moduleImporterService *moduleService.ModuleImporterService,
) *ImportModuleVersionCommand {
	return &ImportModuleVersionCommand{
		moduleImporterService: moduleImporterService,
	}
}

// ImportModuleVersionRequest represents the import request
type ImportModuleVersionRequest struct {
	Namespace string
	Module    string
	Provider  string
	Version   *string // Optional - derived from git tag if not provided
	GitTag    *string // Optional - conflicts with Version
}

// Execute imports a module version from Git
func (c *ImportModuleVersionCommand) Execute(ctx context.Context, req ImportModuleVersionRequest) error {
	importReq := moduleService.ImportModuleVersionRequest{
		Namespace: req.Namespace,
		Module:    req.Module,
		Provider:  req.Provider,
		Version:   req.Version,
		GitTag:    req.GitTag,
	}
	return c.moduleImporterService.ImportModuleVersion(ctx, importReq)
}
