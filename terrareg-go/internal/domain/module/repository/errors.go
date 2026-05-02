package repository

import "errors"

// Common repository errors
var (
	ErrNamespaceNotFound      = errors.New("namespace not found")
	ErrModuleProviderNotFound = errors.New("module provider not found")
	ErrModuleVersionNotFound  = errors.New("module version not found")
	ErrNamespaceAlreadyExists = errors.New("namespace already exists")
)
