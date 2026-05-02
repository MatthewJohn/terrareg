package config

import (
	"context"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/config/repository"
)

// GetVersionQuery handles retrieval of application version
type GetVersionQuery struct {
	configRepo repository.ConfigRepository
}

// NewGetVersionQuery creates a new GetVersionQuery
func NewGetVersionQuery(configRepo repository.ConfigRepository) *GetVersionQuery {
	return &GetVersionQuery{
		configRepo: configRepo,
	}
}

// GetVersionResponse contains the version response
type GetVersionResponse struct {
	Version string
}

// Execute retrieves the application version
func (q *GetVersionQuery) Execute(ctx context.Context) (*GetVersionResponse, error) {
	version, err := q.configRepo.GetVersion(ctx)
	if err != nil {
		return nil, err
	}

	return &GetVersionResponse{
		Version: version,
	}, nil
}
