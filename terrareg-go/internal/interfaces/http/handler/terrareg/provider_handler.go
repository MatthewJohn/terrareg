package terrareg

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	providerCommand "github.com/matthewjohn/terrareg/terrareg-go/internal/application/command/provider"
	providerQuery "github.com/matthewjohn/terrareg/terrareg-go/internal/application/query/provider"
	providerRepo "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/provider/repository"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/interfaces/http/dto"
)

// ProviderHandler handles provider-related requests
type ProviderHandler struct {
	listProvidersQuery        *providerQuery.ListProvidersQuery
	searchProvidersQuery      *providerQuery.SearchProvidersQuery
	getProviderQuery          *providerQuery.GetProviderQuery
	getProviderVersionsQuery  *providerQuery.GetProviderVersionsQuery
	getProviderVersionQuery   *providerQuery.GetProviderVersionQuery
	createOrUpdateProviderCmd *providerCommand.CreateOrUpdateProviderCommand
	publishProviderVersionCmd *providerCommand.PublishProviderVersionCommand
	manageGPGKeyCmd           *providerCommand.ManageGPGKeyCommand
	getProviderDownloadQuery  *providerCommand.GetProviderDownloadQuery
}

// NewProviderHandler creates a new provider handler
func NewProviderHandler(
	listProvidersQuery *providerQuery.ListProvidersQuery,
	searchProvidersQuery *providerQuery.SearchProvidersQuery,
	getProviderQuery *providerQuery.GetProviderQuery,
	getProviderVersionsQuery *providerQuery.GetProviderVersionsQuery,
	getProviderVersionQuery *providerQuery.GetProviderVersionQuery,
	createOrUpdateProviderCmd *providerCommand.CreateOrUpdateProviderCommand,
	publishProviderVersionCmd *providerCommand.PublishProviderVersionCommand,
	manageGPGKeyCmd *providerCommand.ManageGPGKeyCommand,
	getProviderDownloadQuery *providerCommand.GetProviderDownloadQuery,
) *ProviderHandler {
	return &ProviderHandler{
		listProvidersQuery:        listProvidersQuery,
		searchProvidersQuery:      searchProvidersQuery,
		getProviderQuery:          getProviderQuery,
		getProviderVersionsQuery:  getProviderVersionsQuery,
		getProviderVersionQuery:   getProviderVersionQuery,
		createOrUpdateProviderCmd: createOrUpdateProviderCmd,
		publishProviderVersionCmd: publishProviderVersionCmd,
		manageGPGKeyCmd:           manageGPGKeyCmd,
		getProviderDownloadQuery:  getProviderDownloadQuery,
	}
}

// HandleProviderList handles GET /v1/providers
func (h *ProviderHandler) HandleProviderList(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse pagination parameters
	offset, limit := parsePaginationParams(r)

	// Execute query
	providers, total, err := h.listProvidersQuery.Execute(ctx, offset, limit)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Build response (no namespace names or version data available for simple list)
	response := dto.NewProviderListResponse(providers, map[int]string{}, map[int]providerRepo.VersionData{}, total, offset, limit)
	RespondJSON(w, http.StatusOK, response)
}

// HandleProviderSearch handles GET /v1/providers/search
func (h *ProviderHandler) HandleProviderSearch(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse query parameter
	query := r.URL.Query().Get("q")
	if query == "" {
		RespondJSON(w, http.StatusBadRequest, dto.NewError("query parameter 'q' is required"))
		return
	}

	// Parse pagination parameters
	offset, limit := parseProviderPaginationParams(r)

	// Parse filter parameters
	var namespaces []string
	if ns := r.URL.Query()["namespace"]; len(ns) > 0 {
		namespaces = ns
	}

	var categories []string
	if cat := r.URL.Query()["category"]; len(cat) > 0 {
		categories = cat
	}

	var trustedNamespaces *bool
	if tn := r.URL.Query().Get("trusted_namespaces"); tn != "" {
		if val, err := strconv.ParseBool(tn); err == nil {
			trustedNamespaces = &val
		}
	}

	var contributed *bool
	if cb := r.URL.Query().Get("contributed"); cb != "" {
		if val, err := strconv.ParseBool(cb); err == nil {
			contributed = &val
		}
	}

	// Build search query
	searchQuery := providerRepo.ProviderSearchQuery{
		Query:             query,
		Namespaces:        namespaces,
		Categories:        categories,
		TrustedNamespaces: trustedNamespaces,
		Contributed:       contributed,
		Limit:             limit,
		Offset:            offset,
	}

	// Execute query
	result, err := h.searchProvidersQuery.Execute(ctx, searchQuery)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Build response
	response := dto.NewProviderListResponse(result.Providers, result.NamespaceNames, result.VersionData, result.TotalCount, offset, limit)
	RespondJSON(w, http.StatusOK, response)
}

// HandleProviderDetails handles GET /v1/providers/{namespace}/{provider}
func (h *ProviderHandler) HandleProviderDetails(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse path parameters
	namespace := chi.URLParam(r, "namespace")
	providerName := chi.URLParam(r, "provider")

	// Execute query
	provider, err := h.getProviderQuery.Execute(ctx, namespace, providerName)
	if err != nil {
		RespondError(w, http.StatusNotFound, err.Error())
		return
	}

	// Build response
	response := dto.NewProviderDetailResponse(provider)
	RespondJSON(w, http.StatusOK, response)
}

// HandleProviderVersions handles GET /v1/providers/{namespace}/{provider}/versions
func (h *ProviderHandler) HandleProviderVersions(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse path parameters
	namespace := chi.URLParam(r, "namespace")
	providerName := chi.URLParam(r, "provider")

	// Get provider first
	provider, err := h.getProviderQuery.Execute(ctx, namespace, providerName)
	if err != nil {
		RespondError(w, http.StatusNotFound, err.Error())
		return
	}

	// Get versions
	versions, err := h.getProviderVersionsQuery.Execute(ctx, provider.ID())
	if err != nil {
		RespondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Build response
	response := dto.NewProviderVersionsResponse(namespace, providerName, versions)
	RespondJSON(w, http.StatusOK, response)
}

// HandleNamespaceProviders handles GET /v1/providers/{namespace}
func (h *ProviderHandler) HandleNamespaceProviders(w http.ResponseWriter, r *http.Request) {
	// For now, return empty list - can be enhanced later
	RespondJSON(w, http.StatusOK, map[string]interface{}{
		"providers": []interface{}{},
	})
}

// HandleCreateOrUpdateProvider handles POST /v1/providers and PUT /v1/providers/{namespace}/{provider}
func (h *ProviderHandler) HandleCreateOrUpdateProvider(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse request body
	var req providerCommand.CreateOrUpdateProviderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondError(w, http.StatusBadRequest, "Invalid JSON body")
		return
	}

	// Override namespace and provider from URL if available
	if namespace := chi.URLParam(r, "namespace"); namespace != "" {
		req.Namespace = namespace
	}
	if providerName := chi.URLParam(r, "provider"); providerName != "" {
		req.Name = providerName
	}

	// Validate required fields
	if req.Namespace == "" || req.Name == "" {
		RespondError(w, http.StatusBadRequest, "namespace and name are required")
		return
	}

	// Execute command
	provider, err := h.createOrUpdateProviderCmd.Execute(ctx, req)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Build response
	response := dto.NewProviderDetailResponse(provider)
	RespondJSON(w, http.StatusOK, response)
}

// HandlePublishProviderVersion handles POST /v1/providers/{namespace}/{provider}/versions
func (h *ProviderHandler) HandlePublishProviderVersion(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse URL parameters
	namespace := chi.URLParam(r, "namespace")
	providerName := chi.URLParam(r, "provider")

	// Parse request body
	var req providerCommand.PublishProviderVersionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondError(w, http.StatusBadRequest, "Invalid JSON body")
		return
	}

	// Set namespace and provider from URL
	req.Namespace = namespace
	req.ProviderName = providerName

	// Validate required fields
	if req.Namespace == "" || req.ProviderName == "" || req.Version == "" {
		RespondError(w, http.StatusBadRequest, "namespace, provider, and version are required")
		return
	}

	// Execute command
	providerVersion, err := h.publishProviderVersionCmd.Execute(ctx, req)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Build response
	response := map[string]interface{}{
		"version": providerVersion.Version(),
		"beta":    providerVersion.Beta(),
		"status":  "published",
	}
	RespondJSON(w, http.StatusCreated, response)
}

// HandleGetProviderVersion handles GET /v1/providers/{namespace}/{provider}/versions/{version}
func (h *ProviderHandler) HandleGetProviderVersion(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse URL parameters
	namespace := chi.URLParam(r, "namespace")
	providerName := chi.URLParam(r, "provider")
	version := chi.URLParam(r, "version")

	// Validate required fields
	if namespace == "" || providerName == "" || version == "" {
		RespondError(w, http.StatusBadRequest, "namespace, provider, and version are required")
		return
	}

	// Get provider first
	provider, err := h.getProviderQuery.Execute(ctx, namespace, providerName)
	if err != nil {
		RespondError(w, http.StatusNotFound, err.Error())
		return
	}

	// Get specific version
	providerVersion, err := h.getProviderVersionQuery.Execute(ctx, provider.ID(), version)
	if err != nil {
		RespondError(w, http.StatusNotFound, err.Error())
		return
	}

	// Build response
	response := map[string]interface{}{
		"id":        fmt.Sprintf("%s/%s", namespace, providerName),
		"version":   providerVersion.Version(),
		"beta":      providerVersion.Beta(),
		"protocols": providerVersion.ProtocolVersions(),
		"binaries":  []interface{}{}, // Will be populated from provider binaries
	}
	RespondJSON(w, http.StatusOK, response)
}

// HandleAddGPGKey handles POST /v1/providers/{namespace}/{provider}/gpg-keys
func (h *ProviderHandler) HandleAddGPGKey(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse URL parameters
	namespace := chi.URLParam(r, "namespace")
	providerName := chi.URLParam(r, "provider")

	// Parse request body
	var req providerCommand.AddGPGKeyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondError(w, http.StatusBadRequest, "Invalid JSON body")
		return
	}

	// Set namespace and provider from URL
	req.Namespace = namespace
	req.ProviderName = providerName

	// Validate required fields
	if req.Namespace == "" || req.ProviderName == "" || req.KeyID == "" || req.AsciiArmor == "" {
		RespondError(w, http.StatusBadRequest, "namespace, provider, key_id, and ascii_armor are required")
		return
	}

	// Execute command
	provider, err := h.manageGPGKeyCmd.ExecuteAdd(ctx, req)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Build response
	response := dto.NewProviderDetailResponse(provider)
	RespondJSON(w, http.StatusOK, response)
}

// HandleRemoveGPGKey handles DELETE /v1/providers/{namespace}/{provider}/gpg-keys/{key_id}
func (h *ProviderHandler) HandleRemoveGPGKey(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse URL parameters
	namespace := chi.URLParam(r, "namespace")
	providerName := chi.URLParam(r, "provider")
	keyID := chi.URLParam(r, "key_id")

	// Validate required fields
	if namespace == "" || providerName == "" || keyID == "" {
		RespondError(w, http.StatusBadRequest, "namespace, provider, and key_id are required")
		return
	}

	// Create request
	req := providerCommand.RemoveGPGKeyRequest{
		Namespace:    namespace,
		ProviderName: providerName,
		KeyID:        keyID,
	}

	// Execute command
	provider, err := h.manageGPGKeyCmd.ExecuteRemove(ctx, req)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Build response
	response := dto.NewProviderDetailResponse(provider)
	RespondJSON(w, http.StatusOK, response)
}

// parsePaginationParams extracts offset and limit from query parameters
func parsePaginationParams(r *http.Request) (int, int) {
	offset := 0
	limit := 20 // Default limit

	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		if val, err := strconv.Atoi(offsetStr); err == nil && val >= 0 {
			offset = val
		}
	}

	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if val, err := strconv.Atoi(limitStr); err == nil && val > 0 && val <= 100 {
			limit = val
		}
	}

	return offset, limit
}

// parseProviderPaginationParams extracts offset and limit from query parameters with max limit of 50
func parseProviderPaginationParams(r *http.Request) (int, int) {
	offset := 0
	limit := 20 // Default limit

	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		if val, err := strconv.Atoi(offsetStr); err == nil && val >= 0 {
			offset = val
		}
	}

	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if val, err := strconv.Atoi(limitStr); err == nil && val > 0 {
			// Enforce max limit of 50 for provider search (matching Python implementation)
			if val > 50 {
				val = 50
			}
			limit = val
		}
	}

	return offset, limit
}

// HandleProviderDownload handles GET /v1/providers/{namespace}/{provider}/{version}/download/{os}/{arch}
func (h *ProviderHandler) HandleProviderDownload(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Extract path parameters
	namespace := chi.URLParam(r, "namespace")
	provider := chi.URLParam(r, "provider")
	version := chi.URLParam(r, "version")
	os := chi.URLParam(r, "os")
	arch := chi.URLParam(r, "arch")

	// Extract headers for analytics
	userAgent := r.Header.Get("User-Agent")
	terraformVersion := r.Header.Get("X-Terraform-Version")

	// Create request
	req := &providerCommand.GetProviderDownloadRequest{
		Namespace:        namespace,
		Provider:         provider,
		Version:          version,
		OS:               os,
		Arch:             arch,
		UserAgent:        userAgent,
		TerraformVersion: terraformVersion,
	}

	// Execute query
	resp, err := h.getProviderDownloadQuery.Execute(ctx, req)
	if err != nil {
		if err.Error() == "namespace not found" ||
			err.Error() == "provider not found" ||
			err.Error() == "provider version not found" {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(`{"error": "Provider version not found"}`))
			return
		}
		if err.Error() == "unsupported OS" {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(fmt.Sprintf(`{"error": "Unsupported OS: %s"}`, os)))
			return
		}
		if err.Error() == "unsupported architecture" {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(fmt.Sprintf(`{"error": "Unsupported architecture: %s"}`, arch)))
			return
		}
		if err.Error() == "binary not found" {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(fmt.Sprintf(`{"error": "Binary not found for %s/%s"}`, os, arch)))
			return
		}
		RespondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Set content type and respond with JSON
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		RespondError(w, http.StatusInternalServerError, err.Error())
		return
	}
}
