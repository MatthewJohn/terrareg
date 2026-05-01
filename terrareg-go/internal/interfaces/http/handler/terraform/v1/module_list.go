package v1

import (
	"fmt"
	"net/http"
	"time"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/application/query/module"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/module/model"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/shared/types"
	moduledto "github.com/matthewjohn/terrareg/terrareg-go/internal/interfaces/http/dto/module"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/interfaces/http/handler/terrareg" // For respondJSON and respondError
)

// ModuleListHandler handles the HTTP request to list all modules
type ModuleListHandler struct {
	listModulesQuery *module.ListModulesQuery
}

// NewModuleListHandler creates a new ModuleListHandler
func NewModuleListHandler(listModulesQuery *module.ListModulesQuery) *ModuleListHandler {
	return &ModuleListHandler{
		listModulesQuery: listModulesQuery,
	}
}

// HandleListModules executes the query and returns the list of modules
// Python reference: /app/test/unit/terrareg/server/test_api_module_list.py
func (h *ModuleListHandler) HandleListModules(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse query parameters matching Python's test_api_module_list.py
	queryParams := r.URL.Query()

	// Parse providers (support multiple values, like Python)
	var providers []types.ModuleProviderName
	if p := queryParams["provider"]; len(p) > 0 {
		for _, prov := range p {
			if prov != "" {
				providers = append(providers, types.ModuleProviderName(prov))
			}
		}
	}

	// Parse verified parameter
	verified := parseBoolPtr(queryParams.Get("verified"))

	// Parse pagination parameters
	offset := parseInt(queryParams.Get("offset"), 0)
	limit := parseInt(queryParams.Get("limit"), 10)

	// Build input for query
	input := module.ListModulesInput{
		Offset:       offset,
		Limit:        limit,
		Providers:    providers,
		Verified:     verified,
		IncludeCount: true, // Always include count to determine if more results exist
	}

	// Execute query
	moduleProviders, totalCount, err := h.listModulesQuery.Execute(ctx, input)
	if err != nil {
		terrareg.RespondInternalServerError(w, err, "Failed to list modules")
		return
	}

	// Convert domain models to DTOs for the API response
	moduleDTOs := make([]moduledto.ModuleProviderResponse, len(moduleProviders))
	for i, mp := range moduleProviders {
		moduleDTOs[i] = convertModuleProviderToListResponse(mp)
	}

	// Build response with pagination meta (matching Python format)
	// Python: {'meta': {'current_offset': 0, 'limit': 10}, 'modules': [...]}
	meta := map[string]interface{}{
		"current_offset": offset,
		"limit":          limit,
	}

	// Add prev_offset if current offset > 0 (matching Python)
	if offset > 0 {
		prevOffset := offset - limit
		if prevOffset < 0 {
			prevOffset = 0
		}
		meta["prev_offset"] = prevOffset
	}

	// Add next_offset if there are more results (matching Python)
	if totalCount > offset+limit {
		nextOffset := offset + limit
		meta["next_offset"] = nextOffset
	}

	response := map[string]interface{}{
		"meta":    meta,
		"modules": moduleDTOs,
	}

	terrareg.RespondJSON(w, http.StatusOK, response)
}

// convertModuleProviderToListResponse converts a domain ModuleProvider model to a DTO for listing.
// This function is specifically for the /v1/modules list endpoint.
func convertModuleProviderToListResponse(mp *model.ModuleProvider) moduledto.ModuleProviderResponse {
	var publishedAtStr *string
	if latest := mp.GetLatestVersion(); latest != nil && latest.PublishedAt() != nil {
		str := latest.PublishedAt().Format(time.RFC3339) // Format time to string
		publishedAtStr = &str
	}

	var description *string
	if latest := mp.GetLatestVersion(); latest != nil && latest.Details() != nil {
		desc := latest.Description()
		if desc != nil && *desc != "" {
			description = desc
		}
	}

	var owner *string
	if latest := mp.GetLatestVersion(); latest != nil {
		owner = latest.Owner()
	}
	// Fallback to namespace name if no owner is set on the version
	if owner == nil || *owner == "" {
		nsName := string(mp.Namespace().Name())
		if nsName != "" {
			owner = &nsName
		}
	}

	// Dummy values for fields not yet fully implemented or easily accessible from ModuleProvider
	src := "github.com/some/repo" // Placeholder for source
	source := &src
	downloads := 0 // Placeholder

	return moduledto.ModuleProviderResponse{
		ProviderBase: moduledto.ProviderBase{
			ID:        fmt.Sprintf("%d", mp.ID()),
			Namespace: string(mp.Namespace().Name()),
			Name:      string(mp.Module()),
			Provider:  string(mp.Provider()),
			Verified:  mp.IsVerified(),
			Trusted:   false, // TODO: Get from namespace service
		},
		Description: description,
		Owner:       owner,
		Source:      source,
		PublishedAt: publishedAtStr,
		Downloads:   downloads,
	}
}
