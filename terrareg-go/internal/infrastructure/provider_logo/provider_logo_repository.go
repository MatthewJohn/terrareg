package provider_logo

import (
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/provider_logo/model"
)

// ProviderLogoRepository handles provider logo data
// This is an infrastructure concern that provides static data
type ProviderLogoRepository struct {
	logos map[string]model.ProviderLogoInfo
}

// NewProviderLogoRepository creates a new provider logo repository
func NewProviderLogoRepository() *ProviderLogoRepository {
	return &ProviderLogoRepository{
		logos: providerLogoData,
	}
}

// GetProviderLogo returns the logo information for a provider
func (r *ProviderLogoRepository) GetProviderLogo(providerName string) (*model.ProviderLogoInfo, bool) {
	info, exists := r.logos[providerName]
	return &info, exists
}

// GetAllProviderLogos returns all provider logos
func (r *ProviderLogoRepository) GetAllProviderLogos() map[string]model.ProviderLogoInfo {
	// Return a copy to prevent modification of the internal data
	result := make(map[string]model.ProviderLogoInfo)
	for k, v := range r.logos {
		result[k] = v
	}
	return result
}

// providerLogoData contains the static information about provider logos
// This mirrors the Python implementation's INFO dictionary
var providerLogoData = map[string]model.ProviderLogoInfo{
	"aws": {
		Source: "/static/images/PB_AWS_logo_RGB_stacked.547f032d90171cdea4dd90c258f47373c5573db5.png",
		Alt:    "Powered by AWS Cloud Computing",
		Tos:    "Amazon Web Services, AWS, the Powered by AWS logo are trademarks of Amazon.com, Inc. or its affiliates.",
		Link:   "https://aws.amazon.com/",
	},
	"gcp": {
		Source: "/static/images/gcp.png",
		Alt:    "Google Cloud",
		Tos:    "Google Cloud and the Google Cloud logo are trademarks of Google LLC.",
		Link:   "https://cloud.google.com/",
	},
	"null": {
		Source: "/static/images/null.png",
		Alt:    "Null Provider",
		Tos:    " ",
		Link:   "#",
	},
	"datadog": {
		Source: "/static/images/dd_logo_v_rgb.png",
		Alt:    "Works with Datadog",
		Tos:    "All 'Datadog' modules are designed to work with Datadog. Modules are in no way affiliated with nor endorsed by Datadog Inc.",
		Link:   "https://www.datadoghq.com/",
	},
	"consul": {
		Source: "/static/images/consul.png",
		Alt:    "Hashicorp Consul",
		Tos:    "All 'Consul' modules are designed to work with HashiCorp Consul. Terrareg and modules hosted within it are in no way affiliated with, nor endorsed by, HashiCorp. HashiCorp, HashiCorp Consul and the HashiCorp Consul logo are trademarks of HashiCorp.",
		Link:   "#",
	},
	"nomad": {
		Source: "/static/images/nomad.png",
		Alt:    "Hashicorp Nomad",
		Tos:    "All 'Nomad' modules are designed to work with HashiCorp Nomad. Terrareg and modules hosted within it are in no way affiliated with, nor endorsed by, HashiCorp. HashiCorp, HashiCorp Nomad and the HashiCorp Nomad logo are trademarks of HashiCorp.",
		Link:   "#",
	},
	"vagrant": {
		Source: "/static/images/vagrant.png",
		Alt:    "Hashicorp Vagrant",
		Tos:    "All 'Vagrant' modules are designed to work with HashiCorp Vagrant. Terrareg and modules hosted within it are in no way affiliated with, nor endorsed by, HashiCorp. HashiCorp, HashiCorp Vagrant and the HashiCorp Vagrant logo are trademarks of HashiCorp.",
		Link:   "#",
	},
	"vault": {
		Source: "/static/images/vault.png",
		Alt:    "Hashicorp Vault",
		Tos:    "All 'Vault' modules are designed to work with HashiCorp Vault. Terrareg and modules hosted within it are in no way affiliated with, nor endorsed by, HashiCorp. HashiCorp, HashiCorp Vault and the HashiCorp Vault logo are trademarks of HashiCorp.",
		Link:   "#",
	},
}
