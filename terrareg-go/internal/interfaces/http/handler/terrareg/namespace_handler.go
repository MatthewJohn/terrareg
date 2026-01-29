package terrareg

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/application/command/namespace"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/application/query"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/application/query/module"
	namespaceQuery "github.com/matthewjohn/terrareg/terrareg-go/internal/application/query/namespace"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/persistence/sqldb"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/interfaces/http/dto"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/interfaces/http/presenter"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/shared/types"
)

// NamespaceHandler handles namespace-related requests
type NamespaceHandler struct {
	// listNamespacesQuery lists all namespaces (required)
	listNamespacesQuery *module.ListNamespacesQuery
	// createNamespaceCmd creates new namespaces (required)
	createNamespaceCmd *namespace.CreateNamespaceCommand
	// updateNamespaceCmd updates existing namespaces (required)
	updateNamespaceCmd *namespace.UpdateNamespaceCommand
	// deleteNamespaceCmd deletes namespaces (required)
	deleteNamespaceCmd *namespace.DeleteNamespaceCommand
	// namespaceDetailsQuery gets namespace details (required)
	namespaceDetailsQuery *namespaceQuery.NamespaceDetailsQuery
	// presenter formats namespace responses (required)
	presenter *presenter.NamespacePresenter
}

// NewNamespaceHandler creates a new namespace handler
// Returns an error if any required dependency is nil
func NewNamespaceHandler(
	listNamespacesQuery *module.ListNamespacesQuery,
	createNamespaceCmd *namespace.CreateNamespaceCommand,
	updateNamespaceCmd *namespace.UpdateNamespaceCommand,
	deleteNamespaceCmd *namespace.DeleteNamespaceCommand,
	namespaceDetailsQuery *namespaceQuery.NamespaceDetailsQuery,
) (*NamespaceHandler, error) {
	if listNamespacesQuery == nil {
		return nil, fmt.Errorf("listNamespacesQuery cannot be nil")
	}
	if createNamespaceCmd == nil {
		return nil, fmt.Errorf("createNamespaceCmd cannot be nil")
	}
	if updateNamespaceCmd == nil {
		return nil, fmt.Errorf("updateNamespaceCmd cannot be nil")
	}
	if deleteNamespaceCmd == nil {
		return nil, fmt.Errorf("deleteNamespaceCmd cannot be nil")
	}
	if namespaceDetailsQuery == nil {
		return nil, fmt.Errorf("namespaceDetailsQuery cannot be nil")
	}

	return &NamespaceHandler{
		listNamespacesQuery:   listNamespacesQuery,
		createNamespaceCmd:    createNamespaceCmd,
		updateNamespaceCmd:    updateNamespaceCmd,
		deleteNamespaceCmd:    deleteNamespaceCmd,
		namespaceDetailsQuery: namespaceDetailsQuery,
		presenter:             presenter.NewNamespacePresenter(),
	}, nil
}

// HandleNamespaceList handles GET /v1/terrareg/namespaces
func (h *NamespaceHandler) HandleNamespaceList(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse resource type from query parameter - defaults to "module"
	// Python reference: terrareg/server/api/terrareg_namespaces.py:37
	resourceTypeStr := r.URL.Query().Get("type")
	if resourceTypeStr == "" {
		resourceTypeStr = "module" // Python default
	}

	var resourceType sqldb.RegistryResourceType
	switch resourceTypeStr {
	case "module":
		resourceType = sqldb.RegistryResourceTypeModule
	case "provider":
		resourceType = sqldb.RegistryResourceTypeProvider
	default:
		// For invalid type, Python returns 400 with "Invalid type argument"
		// Python reference: terrareg/server/api/terrareg_namespaces.py:66
		RespondJSON(w, http.StatusBadRequest, map[string][]string{
			"errors": {"Invalid type argument"},
		})
		return
	}

	// Parse pagination parameters
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	var listOpts *query.ListOptions
	if limitStr != "" || offsetStr != "" {
		listOpts = &query.ListOptions{}
		if offsetStr != "" {
			if offset, err := strconv.Atoi(offsetStr); err == nil {
				listOpts.Offset = offset
			}
		}
		if limitStr != "" {
			if limit, err := strconv.Atoi(limitStr); err == nil {
				listOpts.Limit = limit
			}
		}
	}

	// Execute query with pagination options
	namespaces, totalCount, err := h.listNamespacesQuery.Execute(ctx, listOpts)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Check if pagination params were provided
	if limitStr == "" && offsetStr == "" {
		// No pagination - return plain array to match Python behavior (legacy format)
		response := h.presenter.ToListArrayWithResourceType(namespaces, resourceType)
		RespondJSON(w, http.StatusOK, response)
	} else {
		// With pagination - return wrapped object with meta
		// Python reference: terrareg/server/api/terrareg_namespaces.py:82-86
		dtos := h.presenter.ToListArrayWithResourceType(namespaces, resourceType)

		// Build meta object (Python reference: ResultData.meta)
		meta := map[string]interface{}{
			"current_offset": listOpts.Offset,
			"limit":          listOpts.Limit,
			"total_count":    totalCount,
		}

		response := map[string]interface{}{
			"meta":       meta,
			"namespaces": dtos,
		}
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
		Name:        types.NamespaceName(req.Name),
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

	// Send response - Python returns 200, not 201
	RespondJSON(w, http.StatusOK, response)
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

	// Parse namespace from URL and convert to typed value
	namespaceName := types.NamespaceName(chi.URLParam(r, "namespace"))
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

	// Parse namespace from URL and convert to typed value
	namespaceName := types.NamespaceName(chi.URLParam(r, "namespace"))
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
