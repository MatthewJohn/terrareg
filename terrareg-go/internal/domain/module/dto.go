package module

import (
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/shared"
)

// ModuleVersionImportInput represents a minimal, pure domain DTO for module import
// This contains only the essential domain information needed for import
type ModuleVersionImportInput struct {
	Namespace string
	Module    string
	Provider  string
	Version   *shared.Version
	GitTag    *string
}

// NewModuleVersionImportInput creates a domain import input from application request
func NewModuleVersionImportInput(namespace, moduleName, provider string, version *shared.Version, gitTag *string) *ModuleVersionImportInput {
	return &ModuleVersionImportInput{
		Namespace: namespace,
		Module:    moduleName,
		Provider:  provider,
		Version:   version,
		GitTag:    gitTag,
	}
}

// Validate performs basic domain validation
func (i *ModuleVersionImportInput) Validate() error {
	// Exactly one of version or git_tag must be provided
	if (i.Version == nil && i.GitTag == nil) || (i.Version != nil && i.GitTag != nil) {
		return shared.ErrInvalidInput
	}

	// Validate required fields
	if i.Namespace == "" || i.Module == "" || i.Provider == "" {
		return shared.ErrInvalidInput
	}

	return nil
}

// GetVersionString returns the version string (either from Version or GitTag)
func (i *ModuleVersionImportInput) GetVersionString() string {
	if i.Version != nil {
		return i.Version.String()
	}
	if i.GitTag != nil {
		return *i.GitTag
	}
	return ""
}

// IsVersionProvided returns true if a specific version is provided
func (i *ModuleVersionImportInput) IsVersionProvided() bool {
	return i.Version != nil
}

// IsGitTagProvided returns true if a git tag is provided
func (i *ModuleVersionImportInput) IsGitTagProvided() bool {
	return i.GitTag != nil
}