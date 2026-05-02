package presenter

import (
	"fmt"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/module/model"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/persistence/sqldb"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/interfaces/http/dto"
)

// NamespacePresenter converts domain models to DTOs
type NamespacePresenter struct{}

// NewNamespacePresenter creates a new namespace presenter
func NewNamespacePresenter() *NamespacePresenter {
	return &NamespacePresenter{}
}

// ToDTO converts a namespace domain model to a DTO
// Defaults to MODULE resource type for view_href
func (p *NamespacePresenter) ToDTO(namespace *model.Namespace) dto.NamespaceResponse {
	return p.ToDTOWithResourceType(namespace, sqldb.RegistryResourceTypeModule)
}

// ToDTOWithResourceType converts a namespace domain model to a DTO with the specified resource type
// Python reference: models.py:1170 - get_view_url(resource_type)
func (p *NamespacePresenter) ToDTOWithResourceType(namespace *model.Namespace, resourceType sqldb.RegistryResourceType) dto.NamespaceResponse {
	// Generate view_href based on resource type
	// Python reference: models.py:1170
	var urlPart string
	switch resourceType {
	case sqldb.RegistryResourceTypeModule:
		urlPart = "modules"
	case sqldb.RegistryResourceTypeProvider:
		urlPart = "providers"
	default:
		urlPart = "modules" // Default to modules
	}
	viewHref := fmt.Sprintf("/%s/%s", urlPart, string(namespace.Name()))

	return dto.NamespaceResponse{
		Name:        string(namespace.Name()),
		DisplayName: namespace.DisplayName(),
		Type:        string(namespace.Type()),
		ViewHref:    viewHref,
	}
}

// ToListDTO converts a list of namespaces to a list DTO (for paginated responses)
func (p *NamespacePresenter) ToListDTO(namespaces []*model.Namespace) dto.NamespaceListResponse {
	return p.ToListDTOWithResourceType(namespaces, sqldb.RegistryResourceTypeModule)
}

// ToListDTOWithResourceType converts a list of namespaces to a list DTO with the specified resource type
func (p *NamespacePresenter) ToListDTOWithResourceType(namespaces []*model.Namespace, resourceType sqldb.RegistryResourceType) dto.NamespaceListResponse {
	dtos := make([]dto.NamespaceResponse, len(namespaces))
	for i, ns := range namespaces {
		dtos[i] = p.ToDTOWithResourceType(ns, resourceType)
	}

	return dto.NamespaceListResponse{
		Namespaces: dtos,
	}
}

// ToListArray converts a list of namespaces to a plain array (for legacy response format)
func (p *NamespacePresenter) ToListArray(namespaces []*model.Namespace) []dto.NamespaceResponse {
	return p.ToListArrayWithResourceType(namespaces, sqldb.RegistryResourceTypeModule)
}

// ToListArrayWithResourceType converts a list of namespaces to a plain array with the specified resource type
func (p *NamespacePresenter) ToListArrayWithResourceType(namespaces []*model.Namespace, resourceType sqldb.RegistryResourceType) []dto.NamespaceResponse {
	dtos := make([]dto.NamespaceResponse, len(namespaces))
	for i, ns := range namespaces {
		dtos[i] = p.ToDTOWithResourceType(ns, resourceType)
	}
	return dtos
}
