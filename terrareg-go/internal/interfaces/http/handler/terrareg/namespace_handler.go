package terrareg

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/application/command/namespace"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/application/query/module"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/interfaces/http/dto"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/interfaces/http/presenter"
)

// NamespaceHandler handles namespace-related requests
type NamespaceHandler struct {
	listNamespacesQuery *module.ListNamespacesQuery
	createNamespaceCmd  *namespace.CreateNamespaceCommand
	presenter           *presenter.NamespacePresenter
}

// NewNamespaceHandler creates a new namespace handler
func NewNamespaceHandler(
	listNamespacesQuery *module.ListNamespacesQuery,
	createNamespaceCmd *namespace.CreateNamespaceCommand,
) *NamespaceHandler {
	return &NamespaceHandler{
		listNamespacesQuery: listNamespacesQuery,
		createNamespaceCmd:  createNamespaceCmd,
		presenter:           presenter.NewNamespacePresenter(),
	}
}

// HandleNamespaceList handles GET /v1/terrareg/namespaces
func (h *NamespaceHandler) HandleNamespaceList(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Execute query
	namespaces, err := h.listNamespacesQuery.Execute(ctx)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Convert to DTO
	response := h.presenter.ToListDTO(namespaces)

	// Send response
	RespondJSON(w, http.StatusOK, response)
}

// HandleNamespaceCreate handles POST /v1/terrareg/namespaces
func (h *NamespaceHandler) HandleNamespaceCreate(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse request body
	var req dto.NamespaceCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondError(w, http.StatusBadRequest, fmt.Errorf("invalid request body: %w", err).Error())
		return
	}

	// Execute command
	cmdReq := namespace.CreateNamespaceRequest{
		Name:        req.Name,
		DisplayName: req.DisplayName,
		Type:        req.Type,
	}

	ns, err := h.createNamespaceCmd.Execute(ctx, cmdReq)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Convert to DTO
	response := h.presenter.ToDTO(ns)

	// Send response
	RespondJSON(w, http.StatusCreated, response)
}
