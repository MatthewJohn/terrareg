package v1

import (
	"fmt"
	"net/http"
	"time"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/application/query/module"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/module/model"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/interfaces/http/dto"
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
func (h *ModuleListHandler) HandleListModules(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	moduleProviders, err := h.listModulesQuery.Execute(ctx)
	if err != nil {
		terrareg.RespondInternalServerError(w, err, "Failed to list modules")
		return
	}

	// Convert domain models to DTOs for the API response
	moduleDTOs := make([]dto.ModuleProviderResponse, len(moduleProviders))
	for i, mp := range moduleProviders {
		moduleDTOs[i] = convertModuleProviderToListResponse(mp)
	}

	// For the /v1/modules endpoint, the response wraps the modules in a "modules" field
	response := map[string][]dto.ModuleProviderResponse{
		"modules": moduleDTOs,
	}

	terrareg.RespondJSON(w, http.StatusOK, response)
}

// convertModuleProviderToListResponse converts a domain ModuleProvider model to a DTO for listing.
// This function is specifically for the /v1/modules list endpoint.
func convertModuleProviderToListResponse(mp *model.ModuleProvider) dto.ModuleProviderResponse {
	var latestVersionStr *string
	if latest := mp.GetLatestVersion(); latest != nil {
		str := latest.Version().String()
		latestVersionStr = &str
	}

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
	// Owner is typically the namespace, but the DTO allows for an explicit owner field
	// For now, use namespace name as owner
	nsName := mp.Namespace().Name()
	if nsName != "" {
		owner = &nsName
	}

	// Dummy values for fields not yet fully implemented or easily accessible from ModuleProvider
	src := "github.com/some/repo" // Placeholder for source
	source := &src
	downloads := 0 // Placeholder

	return dto.ModuleProviderResponse{
		ID:          fmt.Sprintf("%d", mp.ID()),
		Namespace:   mp.Namespace().Name(),
		Name:        mp.Module(),
		Provider:    mp.Provider(),
		Verified:    mp.IsVerified(),
		Description: description,
		Owner:       owner,
		Source:      source,
		PublishedAt: publishedAtStr,
		Downloads:   downloads,
		Version:     latestVersionStr,
	}
}
