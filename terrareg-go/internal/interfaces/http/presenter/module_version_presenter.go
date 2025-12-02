package presenter

import (
	"fmt"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/module/model"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/interfaces/http/dto"
)

// ModuleVersionPresenter converts module version domain models to DTOs
type ModuleVersionPresenter struct{}

// NewModuleVersionPresenter creates a new module version presenter
func NewModuleVersionPresenter() *ModuleVersionPresenter {
	return &ModuleVersionPresenter{}
}

// ToDTO converts a module version domain model to a DTO
func (p *ModuleVersionPresenter) ToDTO(mv *model.ModuleVersion, namespace, moduleName, provider string) dto.ModuleVersionResponse {
	// Build version ID in format: namespace/name/provider/version
	id := fmt.Sprintf("%s/%s/%s/%s", namespace, moduleName, provider, mv.Version().String())

	response := dto.ModuleVersionResponse{
		ID:       id,
		Version:  mv.Version().String(),
		Published: mv.IsPublished(),
		Beta:     mv.IsBeta(),
		Internal: mv.IsInternal(),
	}

	// Add optional fields
	if owner := mv.Owner(); owner != nil {
		response.Owner = owner
	}

	if desc := mv.Description(); desc != nil {
		response.Description = desc
	}

	if publishedAt := mv.PublishedAt(); publishedAt != nil {
		publishedAtStr := publishedAt.Format("2006-01-02T15:04:05Z")
		response.PublishedAt = &publishedAtStr
	}

	return response
}
