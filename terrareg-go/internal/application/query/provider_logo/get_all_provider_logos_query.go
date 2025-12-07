package provider_logo

import (
	"context"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/provider_logo/model"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/provider_logo/repository"
)

// GetAllProviderLogosQuery retrieves all provider logos
type GetAllProviderLogosQuery struct {
	repo repository.ProviderLogoRepository
}

// NewGetAllProviderLogosQuery creates a new GetAllProviderLogosQuery
func NewGetAllProviderLogosQuery(repo repository.ProviderLogoRepository) *GetAllProviderLogosQuery {
	return &GetAllProviderLogosQuery{
		repo: repo,
	}
}

// Execute returns all available provider logos
func (q *GetAllProviderLogosQuery) Execute(ctx context.Context) map[string]*model.ProviderLogo {
	logosData := q.repo.GetAllProviderLogos()
	result := make(map[string]*model.ProviderLogo)
	for providerName, info := range logosData {
		result[providerName] = model.NewProviderLogoFromInfo(providerName, info)
	}
	return result
}
