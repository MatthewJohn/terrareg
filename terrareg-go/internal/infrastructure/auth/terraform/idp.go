package terraform

import (
	"crypto/rsa"
	"time"
)

// TerraformIDPConfig holds the Terraform IDP configuration
type TerraformIDPConfig struct {
	IssuerURL            string
	ClientID             string
	RedirectURIs         []string
	TokenExpiration      time.Duration
	AllowUnsafeRedirects bool
}

// TerraformIDP implements Terrareg as an OIDC Identity Provider for Terraform Cloud/Enterprise
type TerraformIDP struct {
	config TerraformIDPConfig
}

// NewTerraformIDP creates a new Terraform IDP
func NewTerraformIDP(config TerraformIDPConfig, privateKey *rsa.PrivateKey) (*TerraformIDP, error) {
	// TODO: Implement full JWT library integration
	// For now, return a basic implementation to get build working
	return &TerraformIDP{
		config: config,
	}, nil
}

// TODO: Implement remaining OIDC IDP methods when JWT library integration is complete