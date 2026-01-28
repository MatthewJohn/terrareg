package presenter

import (
	"context"
	"fmt"

	analyticsCmd "github.com/matthewjohn/terrareg/terrareg-go/internal/application/command/analytics"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/module/model"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/interfaces/http/dto"
	moduledto "github.com/matthewjohn/terrareg/terrareg-go/internal/interfaces/http/dto/module"
)

// ModulePresenter converts module domain models to DTOs
type ModulePresenter struct {
	analyticsRepo analyticsCmd.AnalyticsRepository
}

// NewModulePresenter creates a new module presenter
func NewModulePresenter(analyticsRepo analyticsCmd.AnalyticsRepository) *ModulePresenter {
	return &ModulePresenter{
		analyticsRepo: analyticsRepo,
	}
}

// ToDTO converts a module provider domain model to a DTO
func (p *ModulePresenter) ToDTO(ctx context.Context, mp *model.ModuleProvider) moduledto.ModuleProviderResponse {
	// Build module ID in Terraform format: namespace/name/provider
	id := fmt.Sprintf("%s/%s/%s", mp.Namespace().Name(), mp.Module(), mp.Provider())

	// Get download statistics from analytics
	downloads := 0
	stats, err := p.analyticsRepo.GetDownloadStats(ctx,
		mp.Namespace().Name(), mp.Module(), mp.Provider())
	if err == nil {
		downloads = stats.TotalDownloads
	}

	response := moduledto.ModuleProviderResponse{
		ProviderBase: moduledto.ProviderBase{
			ID:        id,
			Namespace: string(mp.Namespace().Name()),
			Name:      string(mp.Module()),
			Provider:  string(mp.Provider()),
			Verified:  mp.IsVerified(),
			Trusted:   false, // TODO: Get from namespace service
		},
		Downloads: downloads,
	}

	// Add latest version if available
	latestVersion := mp.GetLatestVersion()
	if latestVersion != nil {
		// Note: Version is not in the base ProviderBase, need to handle differently

		// Get details from latest version
		details := latestVersion.Details()
		if details != nil {
			if owner := latestVersion.Owner(); owner != nil {
				response.Owner = owner
			}
			if desc := latestVersion.Description(); desc != nil {
				response.Description = desc
			}
		}

		// Published at
		if publishedAt := latestVersion.PublishedAt(); publishedAt != nil {
			publishedAtStr := publishedAt.Format("2006-01-02T15:04:05Z")
			response.PublishedAt = &publishedAtStr
		}
	}

	return response
}

// ToListDTO converts a list of module providers to a list DTO
func (p *ModulePresenter) ToListDTO(ctx context.Context, modules []*model.ModuleProvider) moduledto.ModuleListResponse {
	dtos := make([]moduledto.ModuleProviderResponse, len(modules))
	for i, mp := range modules {
		dtos[i] = p.ToDTO(ctx, mp)
	}

	// Include default pagination meta (matching Python behavior)
	meta := &dto.PaginationMeta{
		Limit:         len(modules), // Return actual count as limit
		CurrentOffset: 0,
	}

	return moduledto.ModuleListResponse{
		Modules: dtos,
		Meta:    meta,
	}
}

// ToListDTOWithMeta converts a list of module providers to a list DTO with full pagination metadata
// Python reference: /app/test/unit/terrareg/server/test_api_module_list.py
// Supports offset, limit, next_offset, prev_offset matching Python behavior
func (p *ModulePresenter) ToListDTOWithMeta(ctx context.Context, modules []*model.ModuleProvider, offset, limit, totalCount int) map[string]interface{} {
	dtos := make([]moduledto.ModuleProviderResponse, len(modules))
	for i, mp := range modules {
		dtos[i] = p.ToDTO(ctx, mp)
	}

	// Build pagination meta matching Python ResultData.meta
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

	// Build response with meta (matching Python format)
	// Python: {'meta': {...}, 'modules': [...]}
	return map[string]interface{}{
		"meta":    meta,
		"modules": dtos,
	}
}

// ToSearchDTO converts search results to a search DTO
func (p *ModulePresenter) ToSearchDTO(ctx context.Context, modules []*model.ModuleProvider, totalCount, limit, offset int) moduledto.ModuleSearchResponse {
	dtos := make([]moduledto.ModuleProviderResponse, len(modules))
	for i, mp := range modules {
		dtos[i] = p.ToDTO(ctx, mp)
	}

	// Build pagination meta matching Python ResultData.meta
	meta := dto.PaginationMeta{
		Limit:         limit,
		CurrentOffset: offset,
	}

	// Add prev_offset if current offset > 0 (matching Python)
	if offset > 0 {
		prevOffset := offset - limit
		if prevOffset < 0 {
			prevOffset = 0
		}
		meta.PrevOffset = &prevOffset
	}

	// Add next_offset if there are more results (matching Python)
	if totalCount > offset+limit {
		nextOffset := offset + limit
		meta.NextOffset = &nextOffset
	}

	return moduledto.ModuleSearchResponse{
		Modules: dtos,
		Meta:    meta,
		Count:   &totalCount, // Always include count for terrareg API
	}
}
