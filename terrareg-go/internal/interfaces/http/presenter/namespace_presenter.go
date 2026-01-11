package presenter

import (
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/module/model"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/interfaces/http/dto"
)

// NamespacePresenter converts domain models to DTOs
type NamespacePresenter struct{}

// NewNamespacePresenter creates a new namespace presenter
func NewNamespacePresenter() *NamespacePresenter {
	return &NamespacePresenter{}
}

// ToDTO converts a namespace domain model to a DTO
func (p *NamespacePresenter) ToDTO(namespace *model.Namespace) dto.NamespaceResponse {
	return dto.NamespaceResponse{
		Name:        namespace.Name(),
		DisplayName: namespace.DisplayName(),
		Type:        string(namespace.Type()),
	}
}

// ToListDTO converts a list of namespaces to a list DTO (for paginated responses)
func (p *NamespacePresenter) ToListDTO(namespaces []*model.Namespace) dto.NamespaceListResponse {
	dtos := make([]dto.NamespaceResponse, len(namespaces))
	for i, ns := range namespaces {
		dtos[i] = p.ToDTO(ns)
	}

	return dto.NamespaceListResponse{
		Namespaces: dtos,
	}
}

// ToListArray converts a list of namespaces to a plain array (for legacy response format)
func (p *NamespacePresenter) ToListArray(namespaces []*model.Namespace) []dto.NamespaceResponse {
	dtos := make([]dto.NamespaceResponse, len(namespaces))
	for i, ns := range namespaces {
		dtos[i] = p.ToDTO(ns)
	}
	return dtos
}
