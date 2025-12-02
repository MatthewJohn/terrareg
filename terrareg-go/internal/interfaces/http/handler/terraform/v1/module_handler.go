package v1

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"

	"github.com/terrareg/terrareg/internal/application/query/module"
	"github.com/terrareg/terrareg/internal/domain/module/model"
	"github.com/terrareg/terrareg/internal/interfaces/http/dto"
	"github.com/terrareg/terrareg/internal/interfaces/http/handler/terrareg" // For respondJSON and respondError
)

// TerraformV1ModuleHandler groups all /v1/modules handlers
type TerraformV1ModuleHandler struct {
	moduleListHandler       *ModuleListHandler
	searchModulesQuery      *module.SearchModulesQuery
	getModuleProviderQuery  *module.GetModuleProviderQuery
	listModuleVersionsQuery *module.ListModuleVersionsQuery
	getModuleDownloadQuery  *module.GetModuleDownloadQuery
	getModuleVersionQuery   *module.GetModuleVersionQuery
	// Other /v1/modules handlers will be added here
}

// NewTerraformV1ModuleHandler creates a new TerraformV1ModuleHandler
func NewTerraformV1ModuleHandler(
	moduleListQuery *module.ListModulesQuery,
	searchModulesQuery *module.SearchModulesQuery,
	getModuleProviderQuery *module.GetModuleProviderQuery,
	listModuleVersionsQuery *module.ListModuleVersionsQuery,
	getModuleDownloadQuery *module.GetModuleDownloadQuery,
	getModuleVersionQuery *module.GetModuleVersionQuery,
) *TerraformV1ModuleHandler {
	return &TerraformV1ModuleHandler{
		moduleListHandler:       NewModuleListHandler(moduleListQuery),
		searchModulesQuery:      searchModulesQuery,
		getModuleProviderQuery:  getModuleProviderQuery,
		listModuleVersionsQuery: listModuleVersionsQuery,
		getModuleDownloadQuery:  getModuleDownloadQuery,
		getModuleVersionQuery:   getModuleVersionQuery,
	}
}

// HandleModuleList dispatches to ModuleListHandler
func (h *TerraformV1ModuleHandler) HandleModuleList(w http.ResponseWriter, r *http.Request) {
	h.moduleListHandler.HandleListModules(w, r)
}

// HandleModuleSearch handles the HTTP request to search for modules.
func (h *TerraformV1ModuleHandler) HandleModuleSearch(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Extract query parameters
	query := r.URL.Query().Get("q")
	namespace := r.URL.Query().Get("namespace")
	provider := r.URL.Query().Get("provider")
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	var limit int = 20 // Default limit
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil {
			limit = l
		}
	}

	var offset int = 0 // Default offset
	if offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil {
			offset = o
		}
	}

	// Build search parameters
	params := module.SearchParams{
		Query:  query,
		Limit:  limit,
		Offset: offset,
	}
	if namespace != "" {
		params.Namespace = &namespace
	}
	if provider != "" {
		params.Provider = &provider
	}

	searchResult, err := h.searchModulesQuery.Execute(ctx, params)
	if err != nil {
		terrareg.RespondInternalServerError(w, err, "Failed to search modules")
		return
	}

	// Convert domain models to DTOs for the API response
	moduleDTOs := make([]dto.ModuleProviderResponse, len(searchResult.Modules))
	for i, mp := range searchResult.Modules {
		moduleDTOs[i] = toModuleProviderResponse(mp) // Reuse the conversion function
	}

	// For search, the response wraps the modules in a "modules" field and includes meta
	response := dto.ModuleSearchResponse{
		Modules: moduleDTOs,
		Meta: dto.PaginationMeta{
			Limit:      params.Limit,
			Offset:     params.Offset,
			TotalCount: searchResult.TotalCount,
		},
	}

	terrareg.RespondJSON(w, http.StatusOK, response)
}

// HandleModuleProviderDetails handles the HTTP request to get details for a specific module provider.
func (h *TerraformV1ModuleHandler) HandleModuleProviderDetails(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	namespace := chi.URLParam(r, "namespace")
	name := chi.URLParam(r, "name")
	provider := chi.URLParam(r, "provider")

	moduleProvider, err := h.getModuleProviderQuery.Execute(ctx, namespace, name, provider)
	if err != nil {
		if errors.Is(err, shared.ErrNotFound) {
			terrareg.RespondError(w, http.StatusNotFound, fmt.Sprintf("Module provider %s/%s/%s not found", namespace, name, provider))
			return
		}
		terrareg.RespondInternalServerError(w, err, "Failed to get module provider details")
		return
	}

	response := toModuleProviderResponse(moduleProvider) // Reuse the conversion function

	terrareg.RespondJSON(w, http.StatusOK, response)
}

// HandleModuleVersions handles the HTTP request to list all versions for a specific module provider.
func (h *TerraformV1ModuleHandler) HandleModuleVersions(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	namespace := chi.URLParam(r, "namespace")
	name := chi.URLParam(r, "name")
	provider := chi.URLParam(r, "provider")

	moduleVersions, err := h.listModuleVersionsQuery.Execute(ctx, namespace, name, provider)
	if err != nil {
		if errors.Is(err, shared.ErrNotFound) {
			terrareg.RespondError(w, http.StatusNotFound, fmt.Sprintf("Module provider %s/%s/%s not found", namespace, name, provider))
			return
		}
		terrareg.RespondInternalServerError(w, err, "Failed to list module versions")
		return
	}

	// Convert domain models to DTOs for the API response
	// The Terraform Registry API expects a specific format for versions.
	// Example: {"versions": [{"version": "1.0.0"}, {"version": "0.1.0"}]}
	type VersionDTO struct {
		Version string `json:"version"`
	}
	versionsDTOs := make([]VersionDTO, 0, len(moduleVersions))
	for _, mv := range moduleVersions {
		versionsDTOs = append(versionsDTOs, VersionDTO{Version: mv.Version().String()})
	}

	response := map[string][]VersionDTO{"versions": versionsDTOs}
	terrareg.RespondJSON(w, http.StatusOK, response)
}

// HandleModuleDownload handles the HTTP request to download a specific module version.
func (h *TerraformV1ModuleHandler) HandleModuleDownload(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	namespace := chi.URLParam(r, "namespace")
	name := chi.URLParam(r, "name")
	provider := chi.URLParam(r, "provider")
	version := chi.URLParam(r, "version") // Optional

	downloadInfo, err := h.getModuleDownloadQuery.Execute(ctx, namespace, name, provider, version)
	if err != nil {
		if errors.Is(err, shared.ErrNotFound) {
			terrareg.RespondError(w, http.StatusNotFound, err.Error())
			return
		}
		terrareg.RespondInternalServerError(w, err, "Failed to get module download info")
		return
	}

	// Construct the redirect URL for the module archive
	// Assuming a structure like /modules/namespace/name/provider/version/archive.zip
	// This should align with the actual file storage path
	archiveURL := fmt.Sprintf("/modules/%s/%s/%s/%s/archive.zip",
		downloadInfo.ModuleProvider.Namespace().Name(),
		downloadInfo.ModuleProvider.Module(),
		downloadInfo.ModuleProvider.Provider(),
		downloadInfo.Version.Version().String())

	// Terraform Registry API expects a 204 No Content response with a X-Terraform-Get header
	w.Header().Set("X-Terraform-Get", archiveURL)
	w.WriteHeader(http.StatusNoContent)
}

// HandleModuleVersionDetails handles the HTTP request to get details for a specific module version.
func (h *TerraformV1ModuleHandler) HandleModuleVersionDetails(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	namespace := chi.URLParam(r, "namespace")
	name := chi.URLParam(r, "name")
	provider := chi.URLParam(r, "provider")
	version := chi.URLParam(r, "version")

	moduleVersion, err := h.getModuleVersionQuery.Execute(ctx, namespace, name, provider, version)
	if err != nil {
		if errors.Is(err, shared.ErrNotFound) {
			terrareg.RespondError(w, http.StatusNotFound, err.Error())
			return
		}
		terrareg.RespondInternalServerError(w, err, "Failed to get module version details")
		return
	}

	// Convert domain model to DTO
	response := toModuleVersionResponse(moduleVersion)

	terrareg.RespondJSON(w, http.StatusOK, response)
}

// toModuleVersionResponse converts a domain ModuleVersion model to a DTO.
func toModuleVersionResponse(mv *model.ModuleVersion) dto.ModuleVersionResponse {
	var description *string
	if mv.Details() != nil {
		desc := mv.Details().Description()
		if desc != "" {
			description = &desc
		}
	}

	var owner *string
	if mv.Details() != nil {
		own := mv.Details().Owner()
		if own != "" {
			owner = &own
		}
	}

	var publishedAt *string
	if mv.PublishedAt() != nil {
		str := mv.PublishedAt().Format(time.RFC3339)
		publishedAt = &str
	}

	return dto.ModuleVersionResponse{
		ID:          fmt.Sprintf("%d", mv.ID()),
		Version:     mv.Version().String(),
		Published:   mv.IsPublished(),
		Beta:        mv.IsBeta(),
		Internal:    mv.IsInternal(),
		Description: description,
		Owner:       owner,
		PublishedAt: publishedAt,
	}
}

// Other handler dispatch methods will be added here
