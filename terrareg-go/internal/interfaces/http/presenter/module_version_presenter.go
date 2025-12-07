package presenter

import (
	"fmt"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/module/model"
	moduledto "github.com/matthewjohn/terrareg/terrareg-go/internal/interfaces/http/dto/module"
)

// ModuleVersionPresenter converts module version domain models to DTOs
type ModuleVersionPresenter struct{}

// NewModuleVersionPresenter creates a new module version presenter
func NewModuleVersionPresenter() *ModuleVersionPresenter {
	return &ModuleVersionPresenter{}
}

// ToDTO converts a module version domain model to a DTO
func (p *ModuleVersionPresenter) ToDTO(mv *model.ModuleVersion, namespace, moduleName, provider string) moduledto.ModuleVersionResponse {
	// Build version ID in format: namespace/name/provider/version
	id := fmt.Sprintf("%s/%s/%s/%s", namespace, moduleName, provider, mv.Version().String())

	response := moduledto.ModuleVersionResponse{
		VersionBase: moduledto.VersionBase{
			ProviderBase: moduledto.ProviderBase{
				ID:        id,
				Namespace: namespace,
				Name:      moduleName,
				Provider:  provider,
				Verified:  false, // TODO: Get from module provider
				Trusted:   false, // TODO: Get from namespace service
			},
			Version:  mv.Version().String(),
			Internal: mv.IsInternal(),
		},
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
