package presenter

import (
	"fmt"

	"github.com/terrareg/terrareg/internal/domain/module/model"
	"github.com/terrareg/terrareg/internal/interfaces/http/dto"
)

// ModulePresenter converts module domain models to DTOs
type ModulePresenter struct{}

// NewModulePresenter creates a new module presenter
func NewModulePresenter() *ModulePresenter {
	return &ModulePresenter{}
}

// ToDTO converts a module provider domain model to a DTO
func (p *ModulePresenter) ToDTO(mp *model.ModuleProvider) dto.ModuleProviderResponse {
	// Build module ID in Terraform format: namespace/name/provider
	id := fmt.Sprintf("%s/%s/%s", mp.Namespace().Name(), mp.Module(), mp.Provider())

	response := dto.ModuleProviderResponse{
		ID:        id,
		Namespace: mp.Namespace().Name(),
		Name:      mp.Module(),
		Provider:  mp.Provider(),
		Verified:  mp.IsVerified(),
		Downloads: 0, // TODO: Get from analytics
	}

	// Add latest version if available
	latestVersion := mp.GetLatestVersion()
	if latestVersion != nil {
		versionStr := latestVersion.Version().String()
		response.Version = &versionStr

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
func (p *ModulePresenter) ToListDTO(modules []*model.ModuleProvider) dto.ModuleListResponse {
	dtos := make([]dto.ModuleProviderResponse, len(modules))
	for i, mp := range modules {
		dtos[i] = p.ToDTO(mp)
	}

	return dto.ModuleListResponse{
		Modules: dtos,
	}
}

// ToSearchDTO converts search results to a search DTO
func (p *ModulePresenter) ToSearchDTO(modules []*model.ModuleProvider, totalCount, limit, offset int) dto.ModuleSearchResponse {
	dtos := make([]dto.ModuleProviderResponse, len(modules))
	for i, mp := range modules {
		dtos[i] = p.ToDTO(mp)
	}

	return dto.ModuleSearchResponse{
		Modules: dtos,
		Meta: dto.PaginationMeta{
			Limit:      limit,
			Offset:     offset,
			TotalCount: totalCount,
		},
	}
}
