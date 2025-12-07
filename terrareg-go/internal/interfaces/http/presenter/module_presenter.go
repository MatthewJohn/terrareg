package presenter

import (
	"fmt"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/module/model"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/interfaces/http/dto"
	moduledto "github.com/matthewjohn/terrareg/terrareg-go/internal/interfaces/http/dto/module"
)

// ModulePresenter converts module domain models to DTOs
type ModulePresenter struct{}

// NewModulePresenter creates a new module presenter
func NewModulePresenter() *ModulePresenter {
	return &ModulePresenter{}
}

// ToDTO converts a module provider domain model to a DTO
func (p *ModulePresenter) ToDTO(mp *model.ModuleProvider) moduledto.ModuleProviderResponse {
	// Build module ID in Terraform format: namespace/name/provider
	id := fmt.Sprintf("%s/%s/%s", mp.Namespace().Name(), mp.Module(), mp.Provider())

	response := moduledto.ModuleProviderResponse{
		ProviderBase: moduledto.ProviderBase{
			ID:        id,
			Namespace: mp.Namespace().Name(),
			Name:      mp.Module(),
			Provider:  mp.Provider(),
			Verified:  mp.IsVerified(),
			Trusted:   false, // TODO: Get from namespace service
		},
		Downloads: 0, // TODO: Get from analytics
		Score:     mp.RelevanceScore(), // Include relevance score if available
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
func (p *ModulePresenter) ToListDTO(modules []*model.ModuleProvider) moduledto.ModuleListResponse {
	dtos := make([]moduledto.ModuleProviderResponse, len(modules))
	for i, mp := range modules {
		dtos[i] = p.ToDTO(mp)
	}

	return moduledto.ModuleListResponse{
		Modules: dtos,
	}
}

// ToSearchDTO converts search results to a search DTO
func (p *ModulePresenter) ToSearchDTO(modules []*model.ModuleProvider, totalCount, limit, offset int) moduledto.ModuleSearchResponse {
	dtos := make([]moduledto.ModuleProviderResponse, len(modules))
	for i, mp := range modules {
		dtos[i] = p.ToDTO(mp)
	}

	return moduledto.ModuleSearchResponse{
		Modules: dtos,
		Meta: dto.PaginationMeta{
			Limit:      limit,
			Offset:     offset,
			TotalCount: totalCount,
		},
	}
}
