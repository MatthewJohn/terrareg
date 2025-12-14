package repository

import (
	"context"
	"time"
)

// DependencyGraphRepository defines the interface for dependency graph operations
type DependencyGraphRepository interface {
	// GetModuleVersionDependencies retrieves stored dependencies for a module version
	GetModuleVersionDependencies(ctx context.Context, namespace, moduleName, provider, version string) ([]ModuleDependency, error)

	// GetModuleVersionModules retrieves stored module dependencies for a module version
	GetModuleVersionModules(ctx context.Context, namespace, moduleName, provider, version string) ([]ModuleDependency, error)
}

// ModuleDependency represents a dependency in the database
type ModuleDependency struct {
	ID         int
	ModuleVersionID int
	Type       string // "provider" or "module"
	Source     string // provider name or module path
	Version    string // version constraint
	Optional   bool
	CreatedAt  time.Time
}