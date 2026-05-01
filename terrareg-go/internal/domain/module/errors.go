package module

import (
	"errors"
)

// Domain-specific errors for the module context
var (
	ErrNamespaceNotFound       = errors.New("namespace not found")
	ErrModuleProviderNotFound  = errors.New("module provider not found")
	ErrModuleVersionNotFound   = errors.New("module version not found")
	ErrNamespaceAlreadyExists  = errors.New("namespace already exists")
	ErrModuleAlreadyExists     = errors.New("module provider already exists")
	ErrVersionAlreadyExists    = errors.New("version already exists")
	ErrVersionNotPublished     = errors.New("version not published")
	ErrInvalidModuleArchive    = errors.New("invalid module archive")
	ErrModuleExtractionFailed  = errors.New("module extraction failed")
	ErrNoVersionsAvailable     = errors.New("no versions available")
	ErrInvalidGitConfiguration = errors.New("invalid git configuration")
)
