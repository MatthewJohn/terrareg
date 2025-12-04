package config

import (
	"context"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/config/model"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/config/repository"
)

// GetConfigQuery handles retrieval of configuration for the UI
type GetConfigQuery struct {
	configRepo repository.ConfigRepository
}

// NewGetConfigQuery creates a new GetConfigQuery
func NewGetConfigQuery(configRepo repository.ConfigRepository) *GetConfigQuery {
	return &GetConfigQuery{
		configRepo: configRepo,
	}
}

// GetConfigResponse contains the configuration response
type GetConfigResponse struct {
	Config *model.Config
}

// Execute retrieves the configuration
func (q *GetConfigQuery) Execute(ctx context.Context) (*GetConfigResponse, error) {
	config, err := q.configRepo.GetConfig(ctx)
	if err != nil {
		return nil, err
	}

	return &GetConfigResponse{
		Config: config,
	}, nil
}