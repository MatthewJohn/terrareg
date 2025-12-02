package module

import (
	"context"

	"github.com/terrareg/terrareg/internal/domain/module/service"
)

// ImportModuleVersionCommand handles importing module versions from Git
type ImportModuleVersionCommand struct {
	moduleImporter *service.ModuleImporterService
}

// NewImportModuleVersionCommand creates a new command
func NewImportModuleVersionCommand(
	moduleImporter *service.ModuleImporterService,
) *ImportModuleVersionCommand {
	return &ImportModuleVersionCommand{
		moduleImporter: moduleImporter,
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
	importReq := service.ImportModuleVersionRequest{
		Namespace: req.Namespace,
		Module:    req.Module,
		Provider:  req.Provider,
		Version:   req.Version,
		GitTag:    req.GitTag,
	}
	return c.moduleImporter.ImportModuleVersion(ctx, importReq)
}
