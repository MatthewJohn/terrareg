package v2

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"

	providerQuery "github.com/matthewjohn/terrareg/terrareg-go/internal/application/query/provider"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/interfaces/http/handler/terrareg"
)

// TerraformV2ProviderHandler groups all /v2/providers handlers
type TerraformV2ProviderHandler struct {
	getProviderQuery         *providerQuery.GetProviderQuery
	getProviderVersionsQuery *providerQuery.GetProviderVersionsQuery
	getProviderVersionQuery  *providerQuery.GetProviderVersionQuery
	listProvidersQuery       *providerQuery.ListProvidersQuery
}

// NewTerraformV2ProviderHandler creates a new TerraformV2ProviderHandler
func NewTerraformV2ProviderHandler(
	getProviderQuery *providerQuery.GetProviderQuery,
	getProviderVersionsQuery *providerQuery.GetProviderVersionsQuery,
	getProviderVersionQuery *providerQuery.GetProviderVersionQuery,
	listProvidersQuery *providerQuery.ListProvidersQuery,
) *TerraformV2ProviderHandler {
	return &TerraformV2ProviderHandler{
		getProviderQuery:         getProviderQuery,
		getProviderVersionsQuery: getProviderVersionsQuery,
		getProviderVersionQuery:  getProviderVersionQuery,
		listProvidersQuery:       listProvidersQuery,
	}
}

// HandleProviderDetails handles GET /v2/providers/{namespace}/{provider}
func (h *TerraformV2ProviderHandler) HandleProviderDetails(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse path parameters
	namespace := chi.URLParam(r, "namespace")
	providerName := chi.URLParam(r, "provider")

	// Execute query
	provider, err := h.getProviderQuery.Execute(ctx, namespace, providerName)
	if err != nil {
		terrareg.RespondError(w, http.StatusNotFound, fmt.Sprintf("Provider %s/%s not found: %s", namespace, providerName, err.Error()))
		return
	}

	// Build Terraform Registry v2 response
	response := h.buildV2ProviderResponse(provider)
	terrareg.RespondJSON(w, http.StatusOK, response)
}

// HandleProviderVersions handles GET /v2/providers/{namespace}/{provider}/versions
func (h *TerraformV2ProviderHandler) HandleProviderVersions(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse path parameters
	namespace := chi.URLParam(r, "namespace")
	providerName := chi.URLParam(r, "provider")

	// Get provider first
	provider, err := h.getProviderQuery.Execute(ctx, namespace, providerName)
	if err != nil {
		terrareg.RespondError(w, http.StatusNotFound, fmt.Sprintf("Provider %s/%s not found: %s", namespace, providerName, err.Error()))
		return
	}

	// Get versions
	versions, err := h.getProviderVersionsQuery.Execute(ctx, provider.ID())
	if err != nil {
		terrareg.RespondError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to get provider versions: %s", err.Error()))
		return
	}

	// Build response
	response := h.buildV2VersionsResponse(namespace, providerName, versions)
	terrareg.RespondJSON(w, http.StatusOK, response)
}

// HandleProviderVersion handles GET /v2/providers/{namespace}/{provider}/{version}
func (h *TerraformV2ProviderHandler) HandleProviderVersion(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse path parameters
	namespace := chi.URLParam(r, "namespace")
	providerName := chi.URLParam(r, "provider")
	version := chi.URLParam(r, "version")

	// Get provider first
	provider, err := h.getProviderQuery.Execute(ctx, namespace, providerName)
	if err != nil {
		terrareg.RespondError(w, http.StatusNotFound, fmt.Sprintf("Provider %s/%s not found: %s", namespace, providerName, err.Error()))
		return
	}

	// Get specific version
	providerVersion, err := h.getProviderVersionQuery.Execute(ctx, provider.ID(), version)
	if err != nil {
		terrareg.RespondError(w, http.StatusNotFound, fmt.Sprintf("Provider version %s/%s/%s not found: %s", namespace, providerName, version, err.Error()))
		return
	}

	// Build response
	response := h.buildV2VersionResponse(namespace, providerName, providerVersion)
	terrareg.RespondJSON(w, http.StatusOK, response)
}

// HandleProviderDownload handles GET /v2/providers/{namespace}/{provider}/{version}/download/{os}/{arch}
func (h *TerraformV2ProviderHandler) HandleProviderDownload(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse path parameters
	namespace := chi.URLParam(r, "namespace")
	providerName := chi.URLParam(r, "provider")
	version := chi.URLParam(r, "version")
	os := chi.URLParam(r, "os")
	arch := chi.URLParam(r, "arch")

	// Get provider first
	provider, err := h.getProviderQuery.Execute(ctx, namespace, providerName)
	if err != nil {
		terrareg.RespondError(w, http.StatusNotFound, fmt.Sprintf("Provider %s/%s not found: %s", namespace, providerName, err.Error()))
		return
	}

	// Get specific version
	providerVersion, err := h.getProviderVersionQuery.Execute(ctx, provider.ID(), version)
	if err != nil {
		terrareg.RespondError(w, http.StatusNotFound, fmt.Sprintf("Provider version %s/%s/%s not found: %s", namespace, providerName, version, err.Error()))
		return
	}

	// Find binary for this OS/Arch combination
	var binaryURL string
	for _, binary := range providerVersion.Binaries() {
		if binary.OS() == os && binary.Architecture() == arch {
			binaryURL = fmt.Sprintf("/providers/%s/%s/%s/download/%s/%s/%s",
				namespace, providerName, version, os, arch, binary.Filename())
			break
		}
	}

	if binaryURL == "" {
		terrareg.RespondError(w, http.StatusNotFound, fmt.Sprintf("Binary for %s/%s not found", os, arch))
		return
	}

	// Terraform Registry API expects a 204 No Content response with X-Terraform-Get header
	w.Header().Set("X-Terraform-Get", binaryURL)
	w.Header().Set("X-Terraform-Protocol-Version", "6.0")
	w.WriteHeader(http.StatusNoContent)
}

// HandleProviderDownloadsSummary handles GET /v2/providers/{provider_id}/downloads/summary
func (h *TerraformV2ProviderHandler) HandleProviderDownloadsSummary(w http.ResponseWriter, r *http.Request) {
	// For now, return empty summary - can be enhanced with analytics later
	response := map[string]interface{}{
		"id":         chi.URLParam(r, "provider_id"),
		"downloads": map[string]interface{}{
			"total":   0,
			"version": map[string]int{},
			"platform": map[string]interface{}{
				"linux":   0,
				"darwin":  0,
				"windows": 0,
			},
		},
	}
	terrareg.RespondJSON(w, http.StatusOK, response)
}

// buildV2ProviderResponse builds a Terraform Registry v2 provider response
func (h *TerraformV2ProviderHandler) buildV2ProviderResponse(provider interface{}) map[string]interface{} {
	// Convert provider to map (in a real implementation, this would use proper domain model methods)
	providerMap := map[string]interface{}{
		"id":           "placeholder", // This should come from provider.ID()
		"namespace":    "placeholder", // This should come from provider.Namespace().Name()
		"name":         "placeholder", // This should come from provider.Name()
		"versions":     []interface{}{},
		"logo_url":     "", // Optional logo URL
		"source":       "", // Optional source URL
		"tier":         "community", // Optional tier
		"description":  "", // Optional description
		"published_at": "", // Optional published date
	}
	return providerMap
}

// buildV2VersionsResponse builds a Terraform Registry v2 versions response
func (h *TerraformV2ProviderHandler) buildV2VersionsResponse(namespace, providerName string, versions interface{}) map[string]interface{} {
	versionsMap := map[string]interface{}{
		"id":           fmt.Sprintf("%s/%s", namespace, providerName),
		"versions":     []interface{}{},
		"permissions":  map[string]bool{
			"can_delete":      false,
			"can_create":      false,
			"can_sign":        false,
			"can_partner":     false,
		},
	}
	return versionsMap
}

// buildV2VersionResponse builds a Terraform Registry v2 version response
func (h *TerraformV2ProviderHandler) buildV2VersionResponse(namespace, providerName string, version interface{}) map[string]interface{} {
	versionMap := map[string]interface{}{
		"id":           fmt.Sprintf("%s/%s", namespace, providerName),
		"versions": []map[string]interface{}{
			{
				"version":     "placeholder", // This should come from version.Version()
				"protocols":   []string{"5.0", "6.0"}, // This should come from version.ProtocolVersions()
				"platforms":   []interface{}{}, // This should be populated from version.Binaries()
				"published_at": "", // Optional published date
				"beta":        false, // This should come from version.Beta()
			},
		},
		"permissions":  map[string]bool{
			"can_delete":      false,
			"can_create":      false,
			"can_sign":        false,
			"can_partner":     false,
		},
	}
	return versionMap
}