package model

import (
	"fmt"
)

// ProviderLogo represents a provider's logo information
// This is a separate entity from the Provider domain as it deals with static logo assets
type ProviderLogo struct {
	providerName string
	source       string
	alt          string
	tos          string
	link         string
	exists       bool
}

// NewProviderLogo creates a new ProviderLogo instance with the provided data
func NewProviderLogo(providerName string, info *ProviderLogoInfo, exists bool) *ProviderLogo {
	if !exists || info == nil {
		return &ProviderLogo{
			providerName: providerName,
			exists:       false,
		}
	}

	return &ProviderLogo{
		providerName: providerName,
		source:       info.Source,
		alt:          info.Alt,
		tos:          info.Tos,
		link:         info.Link,
		exists:       true,
	}
}

// NewProviderLogoFromInfo creates a new ProviderLogo instance from ProviderLogoInfo
func NewProviderLogoFromInfo(providerName string, info ProviderLogoInfo) *ProviderLogo {
	return &ProviderLogo{
		providerName: providerName,
		source:       info.Source,
		alt:          info.Alt,
		tos:          info.Tos,
		link:         info.Link,
		exists:       true,
	}
}

// ProviderName returns the name of the provider
func (pl *ProviderLogo) ProviderName() string {
	return pl.providerName
}

// Source returns the path to the logo image
func (pl *ProviderLogo) Source() string {
	return pl.source
}

// Alt returns the alt text for the logo image
func (pl *ProviderLogo) Alt() string {
	return pl.alt
}

// Tos returns the terms of service text for the provider
func (pl *ProviderLogo) Tos() string {
	return pl.tos
}

// Link returns the URL for the provider
func (pl *ProviderLogo) Link() string {
	return pl.link
}

// Exists returns whether the provider has a logo defined
func (pl *ProviderLogo) Exists() bool {
	return pl.exists
}

// String returns a string representation of the provider logo
func (pl *ProviderLogo) String() string {
	if !pl.Exists() {
		return fmt.Sprintf("ProviderLogo{provider: %s, exists: false}", pl.providerName)
	}
	return fmt.Sprintf("ProviderLogo{provider: %s, source: %s, alt: %s}",
		pl.providerName, pl.source, pl.alt)
}
