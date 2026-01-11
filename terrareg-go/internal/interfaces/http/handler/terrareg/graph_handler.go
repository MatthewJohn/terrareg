package terrareg

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/application/query/graph"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/graph/model"
	graphservice "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/graph/service"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/interfaces/http/dto"
)

// GraphHandler handles graph-related requests
type GraphHandler struct {
	getModuleDependencyGraphQuery *graph.GetModuleDependencyGraphQuery
	graphService                  *graphservice.GraphService
}

// NewGraphHandler creates a new graph handler
func NewGraphHandler(
	getModuleDependencyGraphQuery *graph.GetModuleDependencyGraphQuery,
	graphService *graphservice.GraphService,
) *GraphHandler {
	return &GraphHandler{
		getModuleDependencyGraphQuery: getModuleDependencyGraphQuery,
		graphService:                  graphService,
	}
}

// HandleModuleDependencyGraph handles GET /v1/terrareg/modules/{namespace}/{name}/{provider}/{version}/graph
// Returns dependency graph for a specific module version
func (h *GraphHandler) HandleModuleDependencyGraph(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse path parameters
	namespace := chi.URLParam(r, "namespace")
	moduleName := chi.URLParam(r, "name")
	provider := chi.URLParam(r, "provider")
	version := chi.URLParam(r, "version")

	if namespace == "" || moduleName == "" || provider == "" || version == "" {
		RespondJSON(w, http.StatusBadRequest, dto.NewError("Missing required path parameters"))
		return
	}

	// Parse query parameters
	includeBeta := r.URL.Query().Get("include-beta") == "true"
	includeOptional := r.URL.Query().Get("include-optional") == "true"

	// Create request
	req := graph.GetModuleDependencyGraphRequest{
		Namespace:       namespace,
		ModuleName:      moduleName,
		Provider:        provider,
		Version:         version,
		IncludeBeta:     includeBeta,
		IncludeOptional: includeOptional,
	}

	// Execute query to get module dependency graph
	graph, err := h.getModuleDependencyGraphQuery.Execute(ctx, req)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	RespondJSON(w, http.StatusOK, graph)
}

// HandleGraphExport handles GET /v1/terrareg/graph/export
// Exports graph data in various formats (dot, json, svg)
func (h *GraphHandler) HandleGraphExport(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse query parameters
	format := r.URL.Query().Get("format")
	if format == "" {
		format = "json" // Default format
	}

	includeBeta := r.URL.Query().Get("include-beta") == "true"
	namespace := r.URL.Query().Get("namespace")

	// Get global graph data
	graph, err := h.graphService.ParseGlobalGraph(ctx, namespace, includeBeta)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	switch format {
	case "dot":
		// Return Graphviz DOT format
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)

		// Generate DOT format from graph data
		dotOutput := h.convertToDOT(graph)
		w.Write([]byte(dotOutput))

	case "json":
		// Return JSON format
		RespondJSON(w, http.StatusOK, graph)

	default:
		RespondJSON(w, http.StatusBadRequest, dto.NewError("Unsupported format. Supported formats: dot, json"))
	}
}

// HandleGraphStatistics handles GET /v1/terrareg/graph/statistics
// Returns statistics about the module graph
func (h *GraphHandler) HandleGraphStatistics(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	namespace := r.URL.Query().Get("namespace")

	// TODO: Implement actual statistics calculation
	response := map[string]interface{}{
		"overview": map[string]interface{}{
			"total_module_providers": 150,
			"total_modules":          125,
			"total_namespaces":       15,
			"total_dependencies":     450,
		},
		"by_namespace": []map[string]interface{}{
			{
				"namespace":        "hashicorp",
				"module_providers": 45,
				"modules":          40,
				"dependencies":     120,
				"avg_dependencies": 3.0,
			},
			{
				"namespace":        "terraform-aws-modules",
				"module_providers": 30,
				"modules":          25,
				"dependencies":     85,
				"avg_dependencies": 3.4,
			},
		},
		"top_dependencies": []map[string]interface{}{
			{
				"provider":   "hashicorp/aws",
				"count":      85,
				"percentage": 68.0,
			},
			{
				"provider":   "hashicorp/null",
				"count":      35,
				"percentage": 28.0,
			},
		},
		"metadata": map[string]interface{}{
			"namespace":    namespace,
			"generated_at": "2025-12-12T10:00:00Z",
		},
	}

	RespondJSON(w, http.StatusOK, response)
}

// convertToDOT converts graph data to DOT format
func (h *GraphHandler) convertToDOT(graph *model.DependencyGraph) string {
	dot := "digraph module_dependencies {\n"
	dot += "    rankdir=TB;\n"
	dot += "    node [shape=box];\n\n"

	// Add nodes
	for _, node := range graph.Nodes {
		dot += fmt.Sprintf("    \"%s\" [label=\"%s\", type=\"%s\", group=\"%s\"];\n",
			node.ID, node.Label, node.Type, node.Group)
	}

	dot += "\n"

	// Add edges
	for _, edge := range graph.Edges {
		dot += fmt.Sprintf("    \"%s\" -> \"%s\" [label=\"%s\"];\n",
			edge.Source, edge.Target, edge.Label)
	}

	dot += "}\n"
	return dot
}
