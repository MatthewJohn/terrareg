package terrareg

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/application/query/module"
)

// ExampleHandler handles example-related requests
type ExampleHandler struct {
	getExampleDetailsQuery   *module.GetExampleDetailsQuery
	getExampleReadmeHTMLQuery *module.GetExampleReadmeHTMLQuery
	getExampleFileListQuery  *module.GetExampleFileListQuery
	getExampleFileQuery      *module.GetExampleFileQuery
}

// NewExampleHandler creates a new example handler
func NewExampleHandler(
	getExampleDetailsQuery *module.GetExampleDetailsQuery,
	getExampleReadmeHTMLQuery *module.GetExampleReadmeHTMLQuery,
	getExampleFileListQuery *module.GetExampleFileListQuery,
	getExampleFileQuery *module.GetExampleFileQuery,
) *ExampleHandler {
	return &ExampleHandler{
		getExampleDetailsQuery:   getExampleDetailsQuery,
		getExampleReadmeHTMLQuery: getExampleReadmeHTMLQuery,
		getExampleFileListQuery:  getExampleFileListQuery,
		getExampleFileQuery:      getExampleFileQuery,
	}
}

// HandleExampleDetails handles GET /modules/{namespace}/{name}/{provider}/{version}/examples/details/{example}
func (h *ExampleHandler) HandleExampleDetails(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Extract path parameters
	namespace := chi.URLParam(r, "namespace")
	moduleName := chi.URLParam(r, "name")
	provider := chi.URLParam(r, "provider")
	version := chi.URLParam(r, "version")
	examplePath := chi.URLParam(r, "example")

	// Validate required parameters
	if namespace == "" || moduleName == "" || provider == "" || version == "" || examplePath == "" {
		RespondError(w, http.StatusBadRequest, "Missing required path parameters")
		return
	}

	// Execute query to get example details
	exampleDetails, err := h.getExampleDetailsQuery.Execute(ctx, namespace, moduleName, provider, version, examplePath)
	if err != nil {
		if err.Error() == "module provider not found" ||
			err.Error() == "module version not found" ||
			err.Error() == "module version is not published" ||
			err.Error() == "example not found" {
			RespondError(w, http.StatusNotFound, err.Error())
			return
		}
		RespondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Send response
	RespondJSON(w, http.StatusOK, exampleDetails)
}

// HandleExampleReadmeHTML handles GET /modules/{namespace}/{name}/{provider}/{version}/examples/readme_html/{example}
func (h *ExampleHandler) HandleExampleReadmeHTML(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Extract path parameters
	namespace := chi.URLParam(r, "namespace")
	moduleName := chi.URLParam(r, "name")
	provider := chi.URLParam(r, "provider")
	version := chi.URLParam(r, "version")
	examplePath := chi.URLParam(r, "example")

	// Validate required parameters
	if namespace == "" || moduleName == "" || provider == "" || version == "" || examplePath == "" {
		RespondError(w, http.StatusBadRequest, "Missing required path parameters")
		return
	}

	// Execute query to get example README HTML
	readmeHTML, err := h.getExampleReadmeHTMLQuery.Execute(ctx, namespace, moduleName, provider, version, examplePath)
	if err != nil {
		if err.Error() == "module provider not found" ||
			err.Error() == "module version not found" ||
			err.Error() == "module version is not published" ||
			err.Error() == "example not found" ||
			err.Error() == "no README content found" {
			// Return HTML error message for missing README
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`<div class="alert alert-warning">No README found for this example</div>`))
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

// HandleExampleFileList handles GET /modules/{namespace}/{name}/{provider}/{version}/examples/filelist/{example}
func (h *ExampleHandler) HandleExampleFileList(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Extract path parameters
	namespace := chi.URLParam(r, "namespace")
	moduleName := chi.URLParam(r, "name")
	provider := chi.URLParam(r, "provider")
	version := chi.URLParam(r, "version")
	examplePath := chi.URLParam(r, "example")

	// Validate required parameters
	if namespace == "" || moduleName == "" || provider == "" || version == "" || examplePath == "" {
		RespondError(w, http.StatusBadRequest, "Missing required path parameters")
		return
	}

	// Execute query to get example file list
	fileList, err := h.getExampleFileListQuery.Execute(ctx, namespace, moduleName, provider, version, examplePath)
	if err != nil {
		if err.Error() == "module provider not found" ||
			err.Error() == "module version not found" ||
			err.Error() == "module version is not published" ||
			err.Error() == "example not found" {
			RespondError(w, http.StatusNotFound, err.Error())
			return
		}
		RespondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Send response
	RespondJSON(w, http.StatusOK, fileList)
}

// HandleExampleFile handles GET /modules/{namespace}/{name}/{provider}/{version}/examples/file/{file}
func (h *ExampleHandler) HandleExampleFile(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Extract path parameters
	namespace := chi.URLParam(r, "namespace")
	moduleName := chi.URLParam(r, "name")
	provider := chi.URLParam(r, "provider")
	version := chi.URLParam(r, "version")

	// The example path and file path are encoded in the URL path
	// We need to parse the remaining path to separate example and file
	pathParam := chi.URLParam(r, "file")
	if pathParam == "" {
		RespondError(w, http.StatusBadRequest, "Missing required path parameters")
		return
	}

	// Split the path to get example path and file path
	// Format: /namespace/module/provider/version/examples/file/{example}/{file-path}
	// chi gives us the combined path after "file/"
	parts := splitExampleFilePath(pathParam)
	if len(parts) < 2 {
		RespondError(w, http.StatusBadRequest, "Invalid file path format")
		return
	}

	examplePath := parts[0]
	filePath := joinPathParts(parts[1:]...)

	// Validate required parameters
	if namespace == "" || moduleName == "" || provider == "" || version == "" || examplePath == "" || filePath == "" {
		RespondError(w, http.StatusBadRequest, "Missing required path parameters")
		return
	}

	// Combine example path and file path for the query
	fullPath := examplePath + "/" + filePath

	// Execute query to get example file
	fileContent, err := h.getExampleFileQuery.Execute(ctx, namespace, moduleName, provider, version, fullPath)
	if err != nil {
		if err.Error() == "module provider not found" ||
			err.Error() == "module version not found" ||
			err.Error() == "module version is not published" ||
			err.Error() == "example not found" ||
			err.Error() == "file not found" {
			RespondError(w, http.StatusNotFound, err.Error())
			return
		}
		RespondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Determine content type based on file extension
	contentType := "text/plain"
	if len(filePath) > 0 {
		switch {
		case filePath[len(filePath)-3:] == ".tf":
			contentType = "text/plain"
		case filePath[len(filePath)-5:] == ".json":
			contentType = "application/json"
		case filePath[len(filePath)-4:] == ".md":
			contentType = "text/markdown"
		case filePath[len(filePath)-5:] == ".yaml" || filePath[len(filePath)-4:] == ".yml":
			contentType = "application/x-yaml"
		}
	}

	// Set content type header
	w.Header().Set("Content-Type", contentType)

	// Write the file content
	w.WriteHeader(http.StatusOK)
	w.Write(fileContent)
}

// splitExampleFilePath splits the combined path parameter into example path and file path parts
func splitExampleFilePath(path string) []string {
	// Find the first slash to separate example from file path
	for i, char := range path {
		if char == '/' {
			return []string{path[:i], path[i+1:]}
		}
	}
	// No slash found, treat the entire path as the example name
	return []string{path}
}

// joinPathParts joins path parts with slashes
func joinPathParts(parts ...string) string {
	if len(parts) == 0 {
		return ""
	}
	result := parts[0]
	for _, part := range parts[1:] {
		result += "/" + part
	}
	return result
}