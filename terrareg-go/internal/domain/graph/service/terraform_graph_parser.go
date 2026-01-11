package service

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/graph/model"
)

// TerraformGraphParser handles parsing of Terraform graph output
type TerraformGraphParser struct {
	tempDir string
}

// NewTerraformGraphParser creates a new Terraform graph parser
func NewTerraformGraphParser() *TerraformGraphParser {
	tempDir := filepath.Join(os.TempDir(), "terrareg-graph-cache")
	return &TerraformGraphParser{
		tempDir: tempDir,
	}
}

// ParseModuleDependencyGraph parses the dependency graph for a specific module version
func (p *TerraformGraphParser) ParseModuleDependencyGraph(
	ctx context.Context,
	moduleDir string,
	namespace, moduleName, provider, version string,
	includeBeta, includeOptional bool,
) (*model.ModuleDependencyGraph, error) {
	// Run terraform graph command
	graphOutput, err := p.runTerraformGraph(ctx, moduleDir)
	if err != nil {
		return nil, fmt.Errorf("failed to run terraform graph: %w", err)
	}

	// Parse the output to extract dependencies and modules
	dependencies, modules, err := p.parseGraphOutput(graphOutput)
	if err != nil {
		return nil, fmt.Errorf("failed to parse graph output: %w", err)
	}

	// Filter results based on options
	dependencies = p.filterDependencies(dependencies, includeOptional)
	modules = p.filterModules(modules, includeBeta)

	// Create the response
	moduleNode := model.ModuleNode{
		ID:        fmt.Sprintf("%s/%s/%s/%s", namespace, moduleName, provider, version),
		Namespace: namespace,
		Name:      moduleName,
		Provider:  provider,
		Version:   version,
	}

	return &model.ModuleDependencyGraph{
		Module:       moduleNode,
		Dependencies: dependencies,
		Modules:      modules,
		Metadata: model.ModuleGraphMetadata{
			IncludeBeta:       includeBeta,
			IncludeOptional:   includeOptional,
			TotalDependencies: len(dependencies),
			TotalModules:      len(modules),
			GeneratedAt:       time.Now().Format(time.RFC3339),
		},
	}, nil
}

// ParseGlobalGraph parses the global module graph
func (p *TerraformGraphParser) ParseGlobalGraph(
	ctx context.Context,
	namespace string,
	includeBeta bool,
) (*model.DependencyGraph, error) {
	// For now, return empty graph
	// In a real implementation, this would aggregate data from all modules
	nodes := []model.GraphNode{}
	edges := []model.GraphEdge{}

	return &model.DependencyGraph{
		Nodes: nodes,
		Edges: edges,
		Metadata: model.GraphMetadata{
			IncludeBeta: includeBeta,
			Namespace:   namespace,
			TotalNodes:  0,
			TotalEdges:  0,
			GeneratedAt: time.Now().Format(time.RFC3339),
		},
	}, nil
}

// runTerraformGraph executes terraform graph command
func (p *TerraformGraphParser) runTerraformGraph(ctx context.Context, moduleDir string) (string, error) {
	cmd := exec.CommandContext(ctx, "terraform", "graph")
	cmd.Dir = moduleDir

	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("terraform graph command failed: %w", err)
	}

	return string(output), nil
}

// parseGraphOutput parses terraform graph DOT format output
func (p *TerraformGraphParser) parseGraphOutput(output string) ([]model.DependencyNode, []model.DependencyNode, error) {
	dependencies := []model.DependencyNode{}
	modules := []model.DependencyNode{}

	// Regular expressions for matching different node types
	providerRegex := regexp.MustCompile(`"provider\[(.+?)\]"`)
	moduleRegex := regexp.MustCompile(`"module\.(.+?)"`)
	resourceRegex := regexp.MustCompile(`"((.+?)\.(.+?))"`)

	lines := strings.Split(output, "\n")
	for _, line := range lines {
		// Skip empty lines and graph formatting
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "digraph") || strings.Contains(line, "{") || strings.Contains(line, "}") {
			continue
		}

		// Extract provider dependencies
		if matches := providerRegex.FindStringSubmatch(line); len(matches) > 1 {
			dep := model.DependencyNode{
				ID:      "provider/" + matches[1],
				Label:   matches[1],
				Type:    "provider",
				Version: "",
			}
			dependencies = append(dependencies, dep)
			continue
		}

		// Extract module dependencies
		if matches := moduleRegex.FindStringSubmatch(line); len(matches) > 1 {
			dep := model.DependencyNode{
				ID:      "module/" + matches[1],
				Label:   matches[1],
				Type:    "module",
				Version: "",
			}
			modules = append(modules, dep)
			continue
		}

		// Extract resources (we can infer providers from resource types)
		if matches := resourceRegex.FindStringSubmatch(line); len(matches) > 1 {
			resourceType := matches[2]
			if p.isProviderResource(resourceType) {
				provider := p.extractProviderFromResource(resourceType)
				dep := model.DependencyNode{
					ID:      "provider/" + provider,
					Label:   provider,
					Type:    "provider",
					Version: "",
				}
				dependencies = append(dependencies, dep)
			}
		}
	}

	return dependencies, modules, nil
}

// isProviderResource checks if a resource type indicates a provider dependency
func (p *TerraformGraphParser) isProviderResource(resourceType string) bool {
	// Common patterns that indicate built-in providers
	providerResources := []string{
		"aws_", "azurerm_", "google_", "gcp_", "vsphere_",
		"docker_", "kubernetes_", "helm_", "null_",
		"random_", "tls_", "local_", "template_",
		"http_", "external_", "time_",
	}

	for _, prefix := range providerResources {
		if strings.HasPrefix(resourceType, prefix) {
			return true
		}
	}

	return false
}

// extractProviderFromResource extracts provider name from resource type
func (p *TerraformGraphParser) extractProviderFromResource(resourceType string) string {
	// Handle common provider prefixes
	if strings.HasPrefix(resourceType, "azurerm_") {
		return "azurerm"
	}
	if strings.HasPrefix(resourceType, "google_") {
		return "google"
	}
	if strings.HasPrefix(resourceType, "gcp_") {
		return "google"
	}
	if strings.HasPrefix(resourceType, "vsphere_") {
		return "vsphere"
	}

	// Remove suffix for other providers
	parts := strings.Split(resourceType, "_")
	if len(parts) > 1 {
		return parts[0]
	}

	return resourceType
}

// filterDependencies filters dependencies based on options
func (p *TerraformGraphParser) filterDependencies(dependencies []model.DependencyNode, includeOptional bool) []model.DependencyNode {
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
func (p *TerraformGraphParser) filterModules(modules []model.DependencyNode, includeBeta bool) []model.DependencyNode {
	if includeBeta {
		return modules
	}

	// Filter out beta modules (simple heuristic - modules with "beta" in name)
	filtered := []model.DependencyNode{}
	for _, mod := range modules {
		if !strings.Contains(strings.ToLower(mod.Label), "beta") {
			filtered = append(filtered, mod)
		}
	}

	return filtered
}
