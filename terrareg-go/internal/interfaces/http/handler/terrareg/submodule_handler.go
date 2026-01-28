package terrareg

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"

	apperrors "github.com/matthewjohn/terrareg/terrareg-go/internal/application/errors"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/application/query/module"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/shared/types"
)

// SubmoduleHandler handles submodule-related requests
type SubmoduleHandler struct {
	getSubmoduleDetailsQuery    *module.GetSubmoduleDetailsQuery
	getSubmoduleReadmeHTMLQuery *module.GetSubmoduleReadmeHTMLQuery
}

// NewSubmoduleHandler creates a new submodule handler
func NewSubmoduleHandler(
	getSubmoduleDetailsQuery *module.GetSubmoduleDetailsQuery,
	getSubmoduleReadmeHTMLQuery *module.GetSubmoduleReadmeHTMLQuery,
) *SubmoduleHandler {
	return &SubmoduleHandler{
		getSubmoduleDetailsQuery:    getSubmoduleDetailsQuery,
		getSubmoduleReadmeHTMLQuery: getSubmoduleReadmeHTMLQuery,
	}
}

// HandleSubmoduleDetails handles GET /modules/{namespace}/{name}/{provider}/{version}/submodules/details/{submodule}
func (h *SubmoduleHandler) HandleSubmoduleDetails(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Extract path parameters and convert to typed values
	namespace := types.NamespaceName(chi.URLParam(r, "namespace"))
	moduleName := types.ModuleName(chi.URLParam(r, "name"))
	provider := types.ModuleProviderName(chi.URLParam(r, "provider"))
	version := types.ModuleVersion(chi.URLParam(r, "version"))
	submodulePath := chi.URLParam(r, "*")

	// Validate required parameters
	if namespace == "" || moduleName == "" || provider == "" || version == "" || submodulePath == "" {
		RespondError(w, http.StatusBadRequest, "Missing required path parameters")
		return
	}

	// Extract request domain for terraform source URL generation
	// Python reference: /app/terrareg/server/__init__.py - request_domain = request.host
	requestDomain := r.Host

	// Execute query to get submodule details
	submoduleDetails, err := h.getSubmoduleDetailsQuery.Execute(ctx, namespace, moduleName, provider, version, submodulePath, requestDomain)
	if err != nil {
		if apperrors.IsNotFound(err) || errors.Is(err, apperrors.ErrModuleVersionNotPublished) {
			RespondError(w, http.StatusNotFound, err.Error())
			return
		}
		RespondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Send response
	RespondJSON(w, http.StatusOK, submoduleDetails)
}

// HandleSubmoduleReadmeHTML handles GET /modules/{namespace}/{name}/{provider}/{version}/submodules/readme_html/{submodule}
func (h *SubmoduleHandler) HandleSubmoduleReadmeHTML(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Extract path parameters and convert to typed values
	namespace := types.NamespaceName(chi.URLParam(r, "namespace"))
	moduleName := types.ModuleName(chi.URLParam(r, "name"))
	provider := types.ModuleProviderName(chi.URLParam(r, "provider"))
	version := types.ModuleVersion(chi.URLParam(r, "version"))
	submodulePath := chi.URLParam(r, "*")

	// Validate required parameters
	if namespace == "" || moduleName == "" || provider == "" || version == "" || submodulePath == "" {
		RespondError(w, http.StatusBadRequest, "Missing required path parameters")
		return
	}

	// Execute query to get submodule README HTML
	readmeHTML, err := h.getSubmoduleReadmeHTMLQuery.Execute(ctx, namespace, moduleName, provider, version, submodulePath)
	if err != nil {
		if apperrors.IsNotFound(err) || errors.Is(err, apperrors.ErrModuleVersionNotPublished) || errors.Is(err, apperrors.ErrNoReadmeContent) {
			// Return HTML error message for missing README
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`<div class="alert alert-warning">No README found for this submodule</div>`))
			return
		}
		RespondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Set content type to HTML
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	// Write the HTML content
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(readmeHTML))
}
