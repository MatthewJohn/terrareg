package graph

import (
	"context"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/graph/model"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/graph/service"
)

// GetModuleDependencyGraphRequest represents a request to get a module's dependency graph
type GetModuleDependencyGraphRequest struct {
	Namespace       string
	ModuleName      string
	Provider        string
	Version         string
	IncludeBeta     bool
	IncludeOptional bool
}

// GetModuleDependencyGraphQuery retrieves a module's dependency graph
type GetModuleDependencyGraphQuery struct {
	graphService *service.GraphService
}

// NewGetModuleDependencyGraphQuery creates a new GetModuleDependencyGraphQuery
func NewGetModuleDependencyGraphQuery(
	graphService *service.GraphService,
) *GetModuleDependencyGraphQuery {
	return &GetModuleDependencyGraphQuery{
		graphService: graphService,
	}
}

// Execute retrieves a module's dependency graph
func (q *GetModuleDependencyGraphQuery) Execute(
	ctx context.Context,
	req GetModuleDependencyGraphRequest,
) (*model.ModuleDependencyGraph, error) {
	// Use the graph service to get the dependency graph
	graph, err := q.graphService.ParseModuleDependencyGraph(
		ctx,
		req.Namespace,
		req.ModuleName,
		req.Provider,
		req.Version,
		req.IncludeBeta,
		req.IncludeOptional,
	)
	if err != nil {
		return nil, err
	}

	return graph, nil
}
