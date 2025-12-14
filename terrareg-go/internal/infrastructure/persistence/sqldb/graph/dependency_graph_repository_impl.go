package graph

import (
	"context"
	"regexp"
	"strings"
	"time"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/graph/repository"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/persistence/sqldb"
	"gorm.io/gorm"
)

// dependencyGraphRepositoryImpl implements the dependency graph repository using SQL database
type dependencyGraphRepositoryImpl struct {
	db *gorm.DB
}

// TerraformGraph represents a parsed terraform graph
type TerraformGraph struct {
	Nodes []TerraformGraphNode
	Edges []TerraformGraphEdge
}

// TerraformGraphNode represents a node in the terraform graph
type TerraformGraphNode struct {
	ID       string
	Label    string
	Type     string // "module", "resource", "data", "provider", "variable", "output"
	Module   string
	Name     string
	Expanded bool
}

// TerraformGraphEdge represents an edge in the terraform graph
type TerraformGraphEdge struct {
	From string
	To   string
}

// NewDependencyGraphRepository creates a new dependency graph repository
func NewDependencyGraphRepository(db *gorm.DB) repository.DependencyGraphRepository {
	return &dependencyGraphRepositoryImpl{
		db: db,
	}
}

// GetModuleVersionDependencies retrieves provider dependencies from terraform graph
// This parses the terraform_graph field (DOT format) to extract provider dependencies
func (r *dependencyGraphRepositoryImpl) GetModuleVersionDependencies(
	ctx context.Context,
	namespace, moduleName, provider, version string,
) ([]repository.ModuleDependency, error) {
	graph, err := r.getTerraformGraph(ctx, namespace, moduleName, provider, version)
	if err != nil {
		return nil, err
	}

	var dependencies []repository.ModuleDependency

	// Extract provider dependencies from graph nodes
	for _, node := range graph.Nodes {
		if node.Type == "provider" {
			// Parse provider name from format: provider["registry.terraform.io/hashicorp/aws"]
			re := regexp.MustCompile(`provider\["registry\.terraform\.io/([^"]+)"\]`)
			matches := re.FindStringSubmatch(node.Label)

			if len(matches) > 1 {
				providerSource := matches[1] // e.g., "hashicorp/aws"

				dep := repository.ModuleDependency{
					ID:              0,
					ModuleVersionID: 0,
					Type:            "provider",
					Source:          providerSource,
					Version:         "", // Version constraints not available in graph
					Optional:        false,
					CreatedAt:       time.Now(),
				}

				dependencies = append(dependencies, dep)
			}
		}
	}

	return dependencies, nil
}

// GetModuleVersionModules retrieves module dependencies from terraform graph
// This parses the terraform_graph field (DOT format) to extract module dependencies
func (r *dependencyGraphRepositoryImpl) GetModuleVersionModules(
	ctx context.Context,
	namespace, moduleName, provider, version string,
) ([]repository.ModuleDependency, error) {
	graph, err := r.getTerraformGraph(ctx, namespace, moduleName, provider, version)
	if err != nil {
		return nil, err
	}

	var dependencies []repository.ModuleDependency

	// Extract module dependencies from graph nodes
	for _, node := range graph.Nodes {
		if node.Type == "module" {
			// Parse module name from format: module.submodule-call.resource_name (expand)
			parts := strings.Split(node.Label, ".")
			if len(parts) >= 2 && parts[0] == "module" {
				moduleName := parts[1]

				dep := repository.ModuleDependency{
					ID:              0,
					ModuleVersionID: 0,
					Type:            "module",
					Source:          moduleName,
					Version:         "", // Version not available in graph
					Optional:        false,
					CreatedAt:       time.Now(),
				}

				dependencies = append(dependencies, dep)
			}
		}
	}

	return dependencies, nil
}

// getTerraformGraph retrieves and parses terraform graph data from database
func (r *dependencyGraphRepositoryImpl) getTerraformGraph(
	ctx context.Context,
	namespace, moduleName, provider, version string,
) (*TerraformGraph, error) {
	var moduleDetails sqldb.ModuleDetailsDB

	// Query to get terraform_graph from module version
	query := r.db.WithContext(ctx).
		Table("module_details md").
		Select("md.*").
		Joins("INNER JOIN module_version mv ON md.id = mv.module_details_id").
		Joins("INNER JOIN module_provider mp ON mv.module_provider_id = mp.id").
		Joins("INNER JOIN namespace n ON mp.namespace_id = n.id").
		Where("n.name = ? AND mp.module = ? AND mp.provider = ? AND mv.version = ?",
			namespace, moduleName, provider, version)

	if err := query.First(&moduleDetails).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return &TerraformGraph{Nodes: []TerraformGraphNode{}, Edges: []TerraformGraphEdge{}}, nil
		}
		return nil, err
	}

	// Parse terraform_graph DOT format
	if len(moduleDetails.TerraformGraph) == 0 {
		return &TerraformGraph{Nodes: []TerraformGraphNode{}, Edges: []TerraformGraphEdge{}}, nil
	}

	return r.parseDotFormat(string(moduleDetails.TerraformGraph))
}

// parseDotFormat parses terraform graph DOT format into structured data
func (r *dependencyGraphRepositoryImpl) parseDotFormat(dotContent string) (*TerraformGraph, error) {
	graph := &TerraformGraph{
		Nodes: []TerraformGraphNode{},
		Edges: []TerraformGraphEdge{},
	}

	lines := strings.Split(dotContent, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Skip empty lines and graph braces
		if line == "" || line == "digraph" || line == "{" || line == "}" || line == "compound = \"true\"" || line == "newrank = \"true\"" {
			continue
		}

		// Parse node definitions
		if strings.Contains(line, "[label =") {
			node := r.parseNode(line)
			if node != nil {
				graph.Nodes = append(graph.Nodes, *node)
			}
		}

		// Parse edge definitions
		if strings.Contains(line, "->") {
			edge := r.parseEdge(line)
			if edge != nil {
				graph.Edges = append(graph.Edges, *edge)
			}
		}
	}

	return graph, nil
}

// parseNode parses a single node definition from DOT format
func (r *dependencyGraphRepositoryImpl) parseNode(line string) *TerraformGraphNode {
	// Remove trailing semicolon and whitespace
	line = strings.TrimSuffix(line, ";")
	line = strings.TrimSpace(line)

	// Extract node ID and label using regex
	// Pattern: "[root] provider[\"registry.terraform.io/hashicorp/aws\"]" [label = "provider[\"registry.terraform.io/hashicorp/aws\"]", shape = "diamond"]
	re := regexp.MustCompile(`"(\[[^\]]+\]\s+[^"]+)"\s+\[label\s*=\s*"([^"]+)"`)
	matches := re.FindStringSubmatch(line)

	if len(matches) < 3 {
		return nil
	}

	nodeID := matches[1]
	label := matches[2]

	node := &TerraformGraphNode{
		ID:       nodeID,
		Label:    label,
		Type:     r.determineNodeType(label),
		Expanded: strings.Contains(label, "(expand)"),
	}

	// Extract module and name
	if node.Type == "module" {
		r.parseModuleInfo(node)
	} else if node.Type == "resource" || node.Type == "data" {
		r.parseResourceInfo(node)
	}

	return node
}

// parseEdge parses a single edge definition from DOT format
func (r *dependencyGraphRepositoryImpl) parseEdge(line string) *TerraformGraphEdge {
	// Remove trailing semicolon and whitespace
	line = strings.TrimSuffix(line, ";")
	line = strings.TrimSpace(line)

	// Pattern: "[root] aws_s3_bucket.test_bucket (expand)" -> "[root] provider[\"registry.terraform.io/hashicorp/aws\"]"
	re := regexp.MustCompile(`"([^"]+)"\s*->\s*"([^"]+)"`)
	matches := re.FindStringSubmatch(line)

	if len(matches) < 3 {
		return nil
	}

	return &TerraformGraphEdge{
		From: matches[1],
		To:   matches[2],
	}
}

// determineNodeType determines the type of node based on its label
func (r *dependencyGraphRepositoryImpl) determineNodeType(label string) string {
	if strings.HasPrefix(label, "provider[") {
		return "provider"
	} else if strings.HasPrefix(label, "module.") {
		return "module"
	} else if strings.HasPrefix(label, "data.") {
		return "data"
	} else if strings.HasPrefix(label, "var.") {
		return "variable"
	} else if strings.HasPrefix(label, "output.") {
		return "output"
	} else {
		return "resource"
	}
}

// parseModuleInfo extracts module information from node label
func (r *dependencyGraphRepositoryImpl) parseModuleInfo(node *TerraformGraphNode) {
	// Format: module.submodule-call.resource_name (expand)
	parts := strings.Split(node.Label, ".")
	if len(parts) >= 2 {
		node.Module = parts[1] // submodule-call
		if len(parts) >= 3 {
			node.Name = strings.TrimSuffix(parts[2], " (expand)")
		}
	}
}

// parseResourceInfo extracts resource information from node label
func (r *dependencyGraphRepositoryImpl) parseResourceInfo(node *TerraformGraphNode) {
	// Format: aws_s3_bucket.test_bucket (expand)
	parts := strings.Split(strings.TrimSuffix(node.Label, " (expand)"), ".")
	if len(parts) >= 2 {
		node.Module = parts[0] // aws_s3_bucket
		node.Name = parts[1]  // test_bucket
	}
}