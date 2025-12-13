package terrareg

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/application/command/namespace"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/application/query/module"
	namespaceQuery "github.com/matthewjohn/terrareg/terrareg-go/internal/application/query/namespace"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/interfaces/http/dto"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/interfaces/http/presenter"
)

// NamespaceHandler handles namespace-related requests
type NamespaceHandler struct {
	listNamespacesQuery   *module.ListNamespacesQuery
	createNamespaceCmd    *namespace.CreateNamespaceCommand
	updateNamespaceCmd    *namespace.UpdateNamespaceCommand
	deleteNamespaceCmd    *namespace.DeleteNamespaceCommand
	namespaceDetailsQuery *namespaceQuery.NamespaceDetailsQuery
	presenter             *presenter.NamespacePresenter
}

// NewNamespaceHandler creates a new namespace handler
func NewNamespaceHandler(
	listNamespacesQuery *module.ListNamespacesQuery,
	createNamespaceCmd *namespace.CreateNamespaceCommand,
	updateNamespaceCmd *namespace.UpdateNamespaceCommand,
	deleteNamespaceCmd *namespace.DeleteNamespaceCommand,
	namespaceDetailsQuery *namespaceQuery.NamespaceDetailsQuery,
) *NamespaceHandler {
	return &NamespaceHandler{
		listNamespacesQuery:   listNamespacesQuery,
		createNamespaceCmd:    createNamespaceCmd,
		updateNamespaceCmd:    updateNamespaceCmd,
		deleteNamespaceCmd:    deleteNamespaceCmd,
		namespaceDetailsQuery: namespaceDetailsQuery,
		presenter:             presenter.NewNamespacePresenter(),
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

	// Check if limit/offset pagination params are provided
	limit := r.URL.Query().Get("limit")
	offset := r.URL.Query().Get("offset")

	if limit == "" && offset == "" {
		// No pagination - return plain array to match Python behavior (legacy format)
		response := h.presenter.ToListArray(namespaces)
		RespondJSON(w, http.StatusOK, response)
	} else {
		// With pagination - return wrapped object
		response := h.presenter.ToListDTO(namespaces)
		RespondJSON(w, http.StatusOK, response)
	}
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

// HandleNamespaceDetails handles GET /v1/terrareg/namespaces/{namespace}
func (h *NamespaceHandler) HandleNamespaceDetails(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse namespace from URL
	namespaceName := chi.URLParam(r, "namespace")
	if namespaceName == "" {
		RespondError(w, http.StatusBadRequest, "namespace is required")
		return
	}

	// Execute query
	details, err := h.namespaceDetailsQuery.Execute(ctx, namespaceName)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Return 404 if namespace not found (matching Python behavior)
	if details == nil {
		RespondJSON(w, http.StatusNotFound, map[string]interface{}{})
		return
	}

	// Send response
	RespondJSON(w, http.StatusOK, details)
}

// HandleNamespaceUpdate handles POST /v1/terrareg/namespaces/{namespace}
func (h *NamespaceHandler) HandleNamespaceUpdate(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse namespace from URL
	namespaceName := chi.URLParam(r, "namespace")
	if namespaceName == "" {
		RespondError(w, http.StatusBadRequest, "namespace is required")
		return
	}

	// Parse request body
	var req dto.NamespaceUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondError(w, http.StatusBadRequest, fmt.Errorf("invalid request body: %w", err).Error())
		return
	}

	// TODO: Add CSRF token validation when authentication is implemented
	// For now, we'll skip CSRF validation

	// Execute update command
	cmdReq := namespace.UpdateNamespaceRequest{
		Name:        req.Name,
		DisplayName: req.DisplayName,
	}

	response, err := h.updateNamespaceCmd.Execute(ctx, namespaceName, cmdReq)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Send response
	RespondJSON(w, http.StatusOK, response)
}

// HandleNamespaceDelete handles DELETE /v1/terrareg/namespaces/{namespace}
func (h *NamespaceHandler) HandleNamespaceDelete(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse namespace from URL
	namespaceName := chi.URLParam(r, "namespace")
	if namespaceName == "" {
		RespondError(w, http.StatusBadRequest, "namespace is required")
		return
	}

	// Execute delete command
	if err := h.deleteNamespaceCmd.Execute(ctx, namespaceName); err != nil {
		RespondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Send success response (empty JSON object like Python)
	RespondJSON(w, http.StatusOK, map[string]interface{}{})
}
