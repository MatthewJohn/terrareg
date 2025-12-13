package terrareg

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/application/query/module"
)

// SubmoduleHandler handles submodule-related requests
type SubmoduleHandler struct {
	getSubmoduleDetailsQuery   *module.GetSubmoduleDetailsQuery
	getSubmoduleReadmeHTMLQuery *module.GetSubmoduleReadmeHTMLQuery
}

// NewSubmoduleHandler creates a new submodule handler
func NewSubmoduleHandler(
	getSubmoduleDetailsQuery *module.GetSubmoduleDetailsQuery,
	getSubmoduleReadmeHTMLQuery *module.GetSubmoduleReadmeHTMLQuery,
) *SubmoduleHandler {
	return &SubmoduleHandler{
		getSubmoduleDetailsQuery:   getSubmoduleDetailsQuery,
		getSubmoduleReadmeHTMLQuery: getSubmoduleReadmeHTMLQuery,
	}
}

// HandleSubmoduleDetails handles GET /modules/{namespace}/{name}/{provider}/{version}/submodules/details/{submodule}
func (h *SubmoduleHandler) HandleSubmoduleDetails(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Extract path parameters
	namespace := chi.URLParam(r, "namespace")
	moduleName := chi.URLParam(r, "name")
	provider := chi.URLParam(r, "provider")
	version := chi.URLParam(r, "version")
	submodulePath := chi.URLParam(r, "submodule")

	// Validate required parameters
	if namespace == "" || moduleName == "" || provider == "" || version == "" || submodulePath == "" {
		RespondError(w, http.StatusBadRequest, "Missing required path parameters")
		return
	}

	// Execute query to get submodule details
	submoduleDetails, err := h.getSubmoduleDetailsQuery.Execute(ctx, namespace, moduleName, provider, version, submodulePath)
	if err != nil {
		if err.Error() == "module provider not found" ||
			err.Error() == "module version not found" ||
			err.Error() == "module version is not published" ||
			err.Error() == "submodule not found" {
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

	// Extract path parameters
	namespace := chi.URLParam(r, "namespace")
	moduleName := chi.URLParam(r, "name")
	provider := chi.URLParam(r, "provider")
	version := chi.URLParam(r, "version")
	submodulePath := chi.URLParam(r, "submodule")

	// Validate required parameters
	if namespace == "" || moduleName == "" || provider == "" || version == "" || submodulePath == "" {
		RespondError(w, http.StatusBadRequest, "Missing required path parameters")
		return
	}

	// Execute query to get submodule README HTML
	readmeHTML, err := h.getSubmoduleReadmeHTMLQuery.Execute(ctx, namespace, moduleName, provider, version, submodulePath)
	if err != nil {
		if err.Error() == "module provider not found" ||
			err.Error() == "module version not found" ||
			err.Error() == "module version is not published" ||
			err.Error() == "submodule not found" ||
			err.Error() == "no README content found" {
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