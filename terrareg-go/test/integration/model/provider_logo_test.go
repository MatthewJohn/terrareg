package model

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/provider_logo/model"
	providerlogoinfrastructure "github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/provider_logo"
)

// TestProviderLogo_Exists tests the exists property of provider logos
func TestProviderLogo_Exists(t *testing.T) {
	testCases := []struct {
		name         string
		providerName string
		expectExists bool
	}{
		{"aws provider exists", "aws", true},
		{"gcp provider exists", "gcp", true},
		{"null provider exists", "null", true},
		{"datadog provider exists", "datadog", true},
		{"consul provider exists", "consul", true},
		{"nomad provider exists", "nomad", true},
		{"vagrant provider exists", "vagrant", true},
		{"vault provider exists", "vault", true},
		{"doesnotexist provider does not exist", "doesnotexist", false},
		{"empty provider name", "", false},
		{"random provider name", "randomprovider", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			repo := providerlogoinfrastructure.NewProviderLogoRepository()
			info, exists := repo.GetProviderLogo(tc.providerName)

			logo := model.NewProviderLogo(tc.providerName, info, exists)

			assert.Equal(t, tc.expectExists, logo.Exists())
		})
	}
}

// TestProviderLogo_Tos tests the terms of service text for provider logos
func TestProviderLogo_Tos(t *testing.T) {
	testCases := []struct {
		name        string
		providerName string
		expectedTos string
	}{
		{
			"aws provider tos",
			"aws",
			"Amazon Web Services, AWS, the Powered by AWS logo are trademarks of Amazon.com, Inc. or its affiliates.",
		},
		{
			"gcp provider tos",
			"gcp",
			"Google Cloud and the Google Cloud logo are trademarks of Google LLC.",
		},
		{
			"null provider tos",
			"null",
			" ",
		},
		{
			"datadog provider tos",
			"datadog",
			"All 'Datadog' modules are designed to work with Datadog. Modules are in no way affiliated with nor endorsed by Datadog Inc.",
		},
		{
			"consul provider tos",
			"consul",
			"All 'Consul' modules are designed to work with HashiCorp Consul. Terrareg and modules hosted within it are in no way affiliated with, nor endorsed by, HashiCorp. HashiCorp, HashiCorp Consul and the HashiCorp Consul logo are trademarks of HashiCorp.",
		},
		{
			"nomad provider tos",
			"nomad",
			"All 'Nomad' modules are designed to work with HashiCorp Nomad. Terrareg and modules hosted within it are in no way affiliated with, nor endorsed by, HashiCorp. HashiCorp, HashiCorp Nomad and the HashiCorp Nomad logo are trademarks of HashiCorp.",
		},
		{
			"vagrant provider tos",
			"vagrant",
			"All 'Vagrant' modules are designed to work with HashiCorp Vagrant. Terrareg and modules hosted within it are in no way affiliated with, nor endorsed by, HashiCorp. HashiCorp, HashiCorp Vagrant and the HashiCorp Vagrant logo are trademarks of HashiCorp.",
		},
		{
			"vault provider tos",
			"vault",
			"All 'Vault' modules are designed to work with HashiCorp Vault. Terrareg and modules hosted within it are in no way affiliated with, nor endorsed by, HashiCorp. HashiCorp, HashiCorp Vault and the HashiCorp Vault logo are trademarks of HashiCorp.",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			repo := providerlogoinfrastructure.NewProviderLogoRepository()
			info, exists := repo.GetProviderLogo(tc.providerName)

			logo := model.NewProviderLogo(tc.providerName, info, exists)

			assert.True(t, logo.Exists(), "Provider should exist: %s", tc.providerName)
			assert.Equal(t, tc.expectedTos, logo.Tos())
		})
	}
}

// TestProviderLogo_Alt tests the alt text for provider logos
func TestProviderLogo_Alt(t *testing.T) {
	testCases := []struct {
		name         string
		providerName string
		expectedAlt  string
	}{
		{"aws provider alt", "aws", "Powered by AWS Cloud Computing"},
		{"gcp provider alt", "gcp", "Google Cloud"},
		{"null provider alt", "null", "Null Provider"},
		{"datadog provider alt", "datadog", "Works with Datadog"},
		{"consul provider alt", "consul", "Hashicorp Consul"},
		{"nomad provider alt", "nomad", "Hashicorp Nomad"},
		{"vagrant provider alt", "vagrant", "Hashicorp Vagrant"},
		{"vault provider alt", "vault", "Hashicorp Vault"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			repo := providerlogoinfrastructure.NewProviderLogoRepository()
			info, exists := repo.GetProviderLogo(tc.providerName)

			logo := model.NewProviderLogo(tc.providerName, info, exists)

			assert.True(t, logo.Exists(), "Provider should exist: %s", tc.providerName)
			assert.Equal(t, tc.expectedAlt, logo.Alt())
		})
	}
}

// TestProviderLogo_Link tests the link URLs for provider logos
func TestProviderLogo_Link(t *testing.T) {
	testCases := []struct {
		name         string
		providerName string
		expectedLink string
	}{
		{"aws provider link", "aws", "https://aws.amazon.com/"},
		{"gcp provider link", "gcp", "https://cloud.google.com/"},
		{"null provider link", "null", "#"},
		{"datadog provider link", "datadog", "https://www.datadoghq.com/"},
		{"consul provider link", "consul", "#"},
		{"nomad provider link", "nomad", "#"},
		{"vagrant provider link", "vagrant", "#"},
		{"vault provider link", "vault", "#"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			repo := providerlogoinfrastructure.NewProviderLogoRepository()
			info, exists := repo.GetProviderLogo(tc.providerName)

			logo := model.NewProviderLogo(tc.providerName, info, exists)

			assert.True(t, logo.Exists(), "Provider should exist: %s", tc.providerName)
			assert.Equal(t, tc.expectedLink, logo.Link())
		})
	}
}

// TestProviderLogo_Source tests the image source paths for provider logos
func TestProviderLogo_Source(t *testing.T) {
	testCases := []struct {
		name         string
		providerName string
		expectedSource string
	}{
		{
			"aws provider source",
			"aws",
			"/static/images/PB_AWS_logo_RGB_stacked.547f032d90171cdea4dd90c258f47373c5573db5.png",
		},
		{
			"gcp provider source",
			"gcp",
			"/static/images/gcp.png",
		},
		{
			"null provider source",
			"null",
			"/static/images/null.png",
		},
		{
			"datadog provider source",
			"datadog",
			"/static/images/dd_logo_v_rgb.png",
		},
		{
			"consul provider source",
			"consul",
			"/static/images/consul.png",
		},
		{
			"nomad provider source",
			"nomad",
			"/static/images/nomad.png",
		},
		{
			"vagrant provider source",
			"vagrant",
			"/static/images/vagrant.png",
		},
		{
			"vault provider source",
			"vault",
			"/static/images/vault.png",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			repo := providerlogoinfrastructure.NewProviderLogoRepository()
			info, exists := repo.GetProviderLogo(tc.providerName)

			logo := model.NewProviderLogo(tc.providerName, info, exists)

			assert.True(t, logo.Exists(), "Provider should exist: %s", tc.providerName)
			assert.Equal(t, tc.expectedSource, logo.Source())
		})
	}
}

// TestProviderLogo_NonExistentProvider tests properties for non-existent providers
func TestProviderLogo_NonExistentProvider(t *testing.T) {
	providerName := "doesnotexist"

	repo := providerlogoinfrastructure.NewProviderLogoRepository()
	info, exists := repo.GetProviderLogo(providerName)

	logo := model.NewProviderLogo(providerName, info, exists)

	assert.False(t, logo.Exists(), "Non-existent provider should not exist")
	assert.Equal(t, "", logo.Tos(), "Non-existent provider should have empty tos")
	assert.Equal(t, "", logo.Alt(), "Non-existent provider should have empty alt")
	assert.Equal(t, "", logo.Link(), "Non-existent provider should have empty link")
	assert.Equal(t, "", logo.Source(), "Non-existent provider should have empty source")
	assert.Equal(t, providerName, logo.ProviderName(), "ProviderName should be preserved")
}

// TestProviderLogo_GetAllProviderLogos tests getting all provider logos
func TestProviderLogo_GetAllProviderLogos(t *testing.T) {
	repo := providerlogoinfrastructure.NewProviderLogoRepository()
	allLogos := repo.GetAllProviderLogos()

	// Verify expected providers are present
	expectedProviders := []string{
		"aws", "gcp", "null", "datadog", "consul", "nomad", "vagrant", "vault",
	}

	assert.Len(t, allLogos, len(expectedProviders), "Should have exactly 8 providers")

	for _, provider := range expectedProviders {
		info, exists := allLogos[provider]
		assert.True(t, exists, "Provider %s should exist in all logos", provider)
		assert.NotEmpty(t, info.Source, "Provider %s should have a source", provider)
		assert.NotEmpty(t, info.Alt, "Provider %s should have alt text", provider)
		assert.NotEmpty(t, info.Tos, "Provider %s should have tos text", provider)
		assert.NotEmpty(t, info.Link, "Provider %s should have a link", provider)
	}
}

// TestProviderLogo_NewProviderLogoFromInfo tests creating logo from info
func TestProviderLogo_NewProviderLogoFromInfo(t *testing.T) {
	info := model.ProviderLogoInfo{
		Source: "/static/images/test.png",
		Alt:    "Test Provider",
		Tos:    "Test terms of service",
		Link:   "https://example.com/",
	}

	logo := model.NewProviderLogoFromInfo("testprovider", info)

	assert.True(t, logo.Exists(), "Logo created from info should exist")
	assert.Equal(t, "testprovider", logo.ProviderName())
	assert.Equal(t, "/static/images/test.png", logo.Source())
	assert.Equal(t, "Test Provider", logo.Alt())
	assert.Equal(t, "Test terms of service", logo.Tos())
	assert.Equal(t, "https://example.com/", logo.Link())
}

// TestProviderLogo_NilInfo tests creating logo with nil info
func TestProviderLogo_NilInfo(t *testing.T) {
	logo := model.NewProviderLogo("testprovider", nil, false)

	assert.False(t, logo.Exists(), "Logo with nil info should not exist")
	assert.Equal(t, "testprovider", logo.ProviderName())
	assert.Equal(t, "", logo.Source())
	assert.Equal(t, "", logo.Alt())
	assert.Equal(t, "", logo.Tos())
	assert.Equal(t, "", logo.Link())
}

// TestProviderLogo_String tests the String method
func TestProviderLogo_String(t *testing.T) {
	t.Run("existing provider", func(t *testing.T) {
		repo := providerlogoinfrastructure.NewProviderLogoRepository()
		info, exists := repo.GetProviderLogo("aws")

		logo := model.NewProviderLogo("aws", info, exists)

		expected := "ProviderLogo{provider: aws, source: /static/images/PB_AWS_logo_RGB_stacked.547f032d90171cdea4dd90c258f47373c5573db5.png, alt: Powered by AWS Cloud Computing}"
		assert.Equal(t, expected, logo.String())
	})

	t.Run("non-existing provider", func(t *testing.T) {
		logo := model.NewProviderLogo("doesnotexist", nil, false)

		expected := "ProviderLogo{provider: doesnotexist, exists: false}"
		assert.Equal(t, expected, logo.String())
	})
}
