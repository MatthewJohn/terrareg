# Graph Domain Architecture

## Overview

The Graph domain provides dependency graph generation and analysis for Terraform modules. It maintains and queries dependency information for module versions, supporting both module-to-module dependencies and module dependencies.

---

## Core Functionality

The graph domain provides the following capabilities:

- **Dependency Graph Retrieval** - Get complete dependency graphs for module versions
- **Module Graph Retrieval** - Get module-to-module dependencies
- **Graph Filtering** - Filter by beta versions and optional dependencies
- **Metadata Tracking** - Track graph generation metadata

---

## Domain Components

### Models

**Location**: `/internal/domain/graph/model/dependency_graph.go`

#### ModuleDependencyGraph Model

```go
type ModuleDependencyGraph struct {
    Module       ModuleNode
    Dependencies []DependencyNode
    Modules      []DependencyNode
    Metadata     ModuleGraphMetadata
}
```

#### DependencyNode Model

```go
type DependencyNode struct {
    ID       string
    Label    string
    Type     string
    Version  string
    Optional bool
}
```

#### ModuleNode Model

```go
type ModuleNode struct {
    ID        string
    Namespace string
    Name      string
    Provider  string
    Version   string
}
```

#### ModuleGraphMetadata Model

```go
type ModuleGraphMetadata struct {
    IncludeBeta       bool
    IncludeOptional   bool
    TotalDependencies int
    TotalModules      int
    GeneratedAt       string
}
```

### Repository Interface

**Location**: `/internal/domain/graph/repository/dependency_graph_repository.go`

```go
type DependencyGraphRepository interface {
    GetModuleVersionDependencies(
        ctx context.Context,
        namespace, moduleName, provider, version string,
    ) ([]*model.DependencyNode, error)

    GetModuleVersionModules(
        ctx context.Context,
        namespace, moduleName, provider, version string,
    ) ([]*model.DependencyNode, error)
}
```

### Service

**Location**: `/internal/domain/graph/service/graph_service.go`

```go
type GraphService struct {
    graphRepo repository.DependencyGraphRepository
}
```

**Key Methods**:
- `ParseModuleDependencyGraph()` - Get dependency graph for a module version
- `ParseGlobalGraph()` - Get global dependency graph (TODO)

---

## Dependencies

### Domain Dependencies

| Domain | Purpose |
|--------|---------|
| **module** | Provides module version data for graph generation |

### Infrastructure Dependencies

| Component | Purpose |
|-----------|---------|
| **Database** | Persistent storage for dependency data |

### Domains That Use Graph

| Domain | Purpose |
|--------|---------|
| **module** | Display dependency information in module details |

---

## Key Design Principles

1. **Read-Only** - Graph data is populated during module processing, not directly modified
2. **Filtering** - Supports filtering by beta and optional dependencies
3. **Metadata** - Tracks generation metadata for caching and validation
4. **Structured Data** - Returns structured graph data for visualization

---

## Graph Types

### Dependency Graph (Terraform Dependencies)

Represents Terraform resource dependencies within a module:

```
module "vpc" {
  source = "terraform-aws-modules/vpc/aws"
  ...
}
```

**Node Structure**:
- `ID`: Source identifier
- `Label`: Human-readable name
- `Type`: Dependency type (e.g., "module", "provider")
- `Version`: Version constraint
- `Optional`: Whether dependency is optional

### Module Graph (Module-to-Module Dependencies)

Represents dependencies between Terraform modules:

```
module "database" {
  source = "./modules/database"
  ...
}
```

---

## Usage Examples

### Getting Dependency Graph

```go
graph, err := graphService.ParseModuleDependencyGraph(
    ctx,
    "aws",      // namespace
    "vpc",      // module name
    "aws",      // provider
    "3.0.0",    // version
    false,      // include beta
    false,      // include optional
)

for _, dep := range graph.Dependencies {
    fmt.Printf("Dependency: %s (type: %s, version: %s)\n",
        dep.Label, dep.Type, dep.Version)
}
```

### Filtering Options

```go
// Include beta modules
graph, _ := graphService.ParseModuleDependencyGraph(
    ctx, ns, name, provider, version,
    true,  // includeBeta
    false, // includeOptional
)

// Include optional dependencies
graph, _ := graphService.ParseModuleDependencyGraph(
    ctx, ns, name, provider, version,
    false, // includeBeta
    true,  // includeOptional
)
```

### Accessing Metadata

```go
graph, _ := graphService.ParseModuleDependencyGraph(...)

fmt.Printf("Total Dependencies: %d\n", graph.Metadata.TotalDependencies)
fmt.Printf("Total Modules: %d\n", graph.Metadata.TotalModules)
fmt.Printf("Generated At: %s\n", graph.Metadata.GeneratedAt)
```

---

## Graph Generation

Dependency graphs are generated during module processing:

1. **Module Upload** - User uploads module ZIP
2. **Terraform Init** - Run `terraform init` to get dependencies
3. **Parse Output** - Extract dependency information
4. **Store** - Save dependencies to database
5. **Query** - GraphService queries dependencies for display

---

## Beta Detection

Beta modules are detected using a simple heuristic:

```go
func containsIgnoreCase(s, substr string) bool {
    s = strings.ToLower(s)
    substr = strings.ToLower(substr)
    return strings.Contains(s, substr)
}

// Usage
if !containsIgnoreCase(mod.Label, "beta") {
    filtered = append(filtered, mod)
}
```

---

## Global Graph

The `ParseGlobalGraph()` method is a placeholder for future implementation:

```go
func (s *GraphService) ParseGlobalGraph(
    ctx context.Context,
    namespace string,
    includeBeta bool,
) (*model.DependencyGraph, error) {
    // TODO: Implement real global graph parsing
    // Returns empty graph for now
    return &model.DependencyGraph{
        Nodes: []model.GraphNode{},
        Edges: []model.GraphEdge{},
        Metadata: model.GraphMetadata{...},
    }, nil
}
```

---

## Data Flow

```
Module Processing
       ↓
Terraform Init
       ↓
Parse Dependencies
       ↓
Store in Database
       ↓
GraphService Query
       ↓
Return ModuleDependencyGraph
```

---

## References

- [`/internal/domain/graph/model/dependency_graph.go`](./model/dependency_graph.go) - Graph models
- [`/internal/domain/graph/service/terraform_graph_parser.go`](./service/terraform_graph_parser.go) - Terraform-specific parsing
