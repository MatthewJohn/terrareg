package provider_source

import (
	"context"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/provider_source/model"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/provider_source/repository"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/provider_source/service"
)

// BaseProviderSource is the base implementation for provider sources
// Python reference: provider_source/base.py::BaseProviderSource
type BaseProviderSource struct {
	name         string
	repo         repository.ProviderSourceRepository
	sourceClass  service.ProviderSourceClass
}

// NewBaseProviderSource creates a new base provider source
func NewBaseProviderSource(
	name string,
	repo repository.ProviderSourceRepository,
	sourceClass service.ProviderSourceClass,
) *BaseProviderSource {
	return &BaseProviderSource{
		name:        name,
		repo:        repo,
		sourceClass: sourceClass,
	}
}

// Name returns the provider source name
// Python reference: base.py::name property
func (b *BaseProviderSource) Name() string {
	return b.name
}

// ApiName returns the API-friendly name
// Python reference: base.py::api_name property
func (b *BaseProviderSource) ApiName(ctx context.Context) (string, error) {
	source, err := b.repo.FindByName(ctx, b.name)
	if err != nil {
		return "", err
	}
	if source == nil {
		return "", nil
	}
	return source.ApiName(), nil
}

// Config returns the provider source configuration
// Python reference: base.py::_config property
func (b *BaseProviderSource) Config(ctx context.Context) (*model.ProviderSourceConfig, error) {
	source, err := b.repo.FindByName(ctx, b.name)
	if err != nil {
		return nil, err
	}
	if source == nil {
		return nil, nil
	}
	return source.Config(), nil
}
