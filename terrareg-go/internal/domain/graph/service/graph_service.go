package service

import (
	"context"
	"strings"
	"time"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/graph/model"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/graph/repository"
)

// GraphService handles dependency graph operations
type GraphService struct {
	graphRepo repository.DependencyGraphRepository
}

// NewGraphService creates a new graph service
func NewGraphService(graphRepo repository.DependencyGraphRepository) *GraphService {
	return &GraphService{
		graphRepo: graphRepo,
	}
}

// ParseModuleDependencyGraph retrieves the dependency graph for a specific module version from the database
func (s *GraphService) ParseModuleDependencyGraph(
	ctx context.Context,
	namespace, moduleName, provider, version string,
	includeBeta, includeOptional bool,
) (*model.ModuleDependencyGraph, error) {
	// Get dependencies from database
	dependencies, err := s.graphRepo.GetModuleVersionDependencies(ctx, namespace, moduleName, provider, version)
	if err != nil {
		return nil, err
	}

	// Get module dependencies from database
	modules, err := s.graphRepo.GetModuleVersionModules(ctx, namespace, moduleName, provider, version)
	if err != nil {
		return nil, err
	}

	// Convert database dependencies to model
	modelDependencies := make([]model.DependencyNode, len(dependencies))
	for i, dep := range dependencies {
		modelDep := model.DependencyNode{
			ID:       dep.Source,
			Label:    dep.Source,
			Type:     dep.Type,
			Version:  dep.Version,
			Optional: dep.Optional,
		}
		modelDependencies[i] = modelDep
	}

	// Convert database modules to model
	modelModules := make([]model.DependencyNode, len(modules))
	for i, mod := range modules {
		modelMod := model.DependencyNode{
			ID:       mod.Source,
			Label:    mod.Source,
			Type:     mod.Type,
			Version:  mod.Version,
			Optional: mod.Optional,
		}
		modelModules[i] = modelMod
	}

	// Filter results based on options
	modelDependencies = s.filterDependencies(modelDependencies, includeOptional)
	modelModules = s.filterModules(modelModules, includeBeta)

	// Create the response
	moduleNode := model.ModuleNode{
		ID:        namespace + "/" + moduleName + "/" + provider + "/" + version,
		Namespace: namespace,
		Name:      moduleName,
		Provider:  provider,
		Version:   version,
	}

	return &model.ModuleDependencyGraph{
		Module:       moduleNode,
		Dependencies: modelDependencies,
		Modules:      modelModules,
		Metadata: model.ModuleGraphMetadata{
			IncludeBeta:       includeBeta,
			IncludeOptional:   includeOptional,
			TotalDependencies: len(modelDependencies),
			TotalModules:      len(modelModules),
			GeneratedAt:       "2025-01-13T10:00:00Z",
		},
	}, nil
}

// ParseGlobalGraph retrieves the global dependency graph
func (s *GraphService) ParseGlobalGraph(
	ctx context.Context,
	namespace string,
	includeBeta bool,
) (*model.DependencyGraph, error) {
	// TODO: Implement real global graph parsing
	// For now, return an empty graph
	return &model.DependencyGraph{
		Nodes: []model.GraphNode{},
		Edges: []model.GraphEdge{},
		Metadata: model.GraphMetadata{
			IncludeBeta:     includeBeta,
			IncludeOptional: false,
			Namespace:       namespace,
			TotalNodes:      0,
			TotalEdges:      0,
			GeneratedAt:     time.Now().Format(time.RFC3339),
		},
	}, nil
}

// filterDependencies filters dependencies based on options
func (s *GraphService) filterDependencies(dependencies []model.DependencyNode, includeOptional bool) []model.DependencyNode {
	if includeOptional {
		return dependencies
	}

	// Filter out optional dependencies
	filtered := []model.DependencyNode{}
	for _, dep := range dependencies {
		if !dep.Optional {
			filtered = append(filtered, dep)
		}
	}

	return filtered
}

// filterModules filters modules based on beta inclusion
func (s *GraphService) filterModules(modules []model.DependencyNode, includeBeta bool) []model.DependencyNode {
	if includeBeta {
		return modules
	}

	// Filter out beta modules (simple heuristic - modules with "beta" in name)
	filtered := []model.DependencyNode{}
	for _, mod := range modules {
		if !containsIgnoreCase(mod.Label, "beta") {
			filtered = append(filtered, mod)
		}
	}

	return filtered
}

// containsIgnoreCase checks if a string contains another string, case-insensitive
func containsIgnoreCase(s, substr string) bool {
	s = strings.ToLower(s)
	substr = strings.ToLower(substr)
	return strings.Contains(s, substr)
}
