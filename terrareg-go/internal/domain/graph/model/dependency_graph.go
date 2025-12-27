package model

// NodeType represents the type of node in the dependency graph
type NodeType string

const (
	NodeTypeModule   NodeType = "module"
	NodeTypeResource NodeType = "resource"
	NodeTypeData     NodeType = "data"
	NodeTypeProvider NodeType = "provider"
)

// GraphNode represents a node in the dependency graph
type GraphNode struct {
	ID        string            `json:"id"`
	Label     string            `json:"label"`
	Type      NodeType          `json:"type"`
	Group     string            `json:"group,omitempty"`
	Namespace string            `json:"namespace,omitempty"`
	Version   string            `json:"version,omitempty"`
	Optional  bool              `json:"optional,omitempty"`
	Style     map[string]string `json:"style,omitempty"`
}

// GraphEdge represents an edge in the dependency graph
type GraphEdge struct {
	ID      string            `json:"id"`
	Source  string            `json:"source"`
	Target  string            `json:"target"`
	Label   string            `json:"label,omitempty"`
	Classes []string          `json:"classes,omitempty"`
	Style   map[string]string `json:"style,omitempty"`
}

// DependencyGraph represents the complete dependency graph
type DependencyGraph struct {
	Nodes    []GraphNode   `json:"nodes"`
	Edges    []GraphEdge   `json:"edges"`
	Metadata GraphMetadata `json:"metadata"`
}

// GraphMetadata contains metadata about the graph
type GraphMetadata struct {
	IncludeBeta     bool   `json:"include_beta"`
	IncludeOptional bool   `json:"include_optional"`
	Namespace       string `json:"namespace,omitempty"`
	TotalNodes      int    `json:"total_nodes"`
	TotalEdges      int    `json:"total_edges"`
	GeneratedAt     string `json:"generated_at"`
}

// ModuleDependencyGraph represents a dependency graph for a specific module
type ModuleDependencyGraph struct {
	Module       ModuleNode          `json:"module"`
	Dependencies []DependencyNode    `json:"dependencies"`
	Modules      []DependencyNode    `json:"modules"`
	Metadata     ModuleGraphMetadata `json:"metadata"`
}

// ModuleNode represents the module at the root of the dependency graph
type ModuleNode struct {
	ID        string `json:"id"`
	Namespace string `json:"namespace"`
	Name      string `json:"name"`
	Provider  string `json:"provider"`
	Version   string `json:"version"`
}

// DependencyNode represents a dependency (either a module or provider)
type DependencyNode struct {
	ID       string `json:"id"`
	Label    string `json:"label"`
	Type     string `json:"type"`
	Version  string `json:"version,omitempty"`
	Optional bool   `json:"optional,omitempty"`
}

// ModuleGraphMetadata contains metadata for a module dependency graph
type ModuleGraphMetadata struct {
	IncludeBeta       bool   `json:"include_beta"`
	IncludeOptional   bool   `json:"include_optional"`
	TotalDependencies int    `json:"total_dependencies"`
	TotalModules      int    `json:"total_modules"`
	GeneratedAt       string `json:"generated_at"`
}

// NewGraphNode creates a new graph node
func NewGraphNode(id, label string, nodeType NodeType) *GraphNode {
	return &GraphNode{
		ID:    id,
		Label: label,
		Type:  nodeType,
		Style: make(map[string]string),
	}
}

// NewGraphEdge creates a new graph edge
func NewGraphEdge(source, target string) *GraphEdge {
	return &GraphEdge{
		ID:     source + "." + target,
		Source: source,
		Target: target,
		Style:  make(map[string]string),
	}
}

// SetNodeStyle sets the style for a graph node
func (n *GraphNode) SetNodeStyle(key, value string) {
	if n.Style == nil {
		n.Style = make(map[string]string)
	}
	n.Style[key] = value
}

// SetEdgeStyle sets the style for a graph edge
func (e *GraphEdge) SetEdgeStyle(key, value string) {
	if e.Style == nil {
		e.Style = make(map[string]string)
	}
	e.Style[key] = value
}

// AddEdgeClass adds a class to an edge
func (e *GraphEdge) AddEdgeClass(class string) {
	if e.Classes == nil {
		e.Classes = []string{}
	}
	e.Classes = append(e.Classes, class)
}
