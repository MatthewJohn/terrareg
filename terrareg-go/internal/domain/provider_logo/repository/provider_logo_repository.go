package repository

import "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/provider_logo/model"

// ProviderLogoRepository defines the interface for provider logo data access
type ProviderLogoRepository interface {
	GetProviderLogo(providerName string) (*model.ProviderLogoInfo, bool)
	GetAllProviderLogos() map[string]model.ProviderLogoInfo
}
