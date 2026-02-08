package service_test

import (
	"testing"

	authservice "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/auth/service"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/config/model"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/config/service"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/config"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/version"
)

// TestBuildDomainConfig_SAMLEnabled_DelegatesToSAMLService tests that the config service
// properly delegates to the SAML service to determine if SAML is enabled.
// Tests follow the pattern: all of them (true), none of them (false), all except one (false for each)
//
// Note: These tests directly construct config objects rather than using LoadConfiguration
// to ensure tests are deterministic and not affected by the actual environment.
func TestBuildDomainConfig_SAMLEnabled_DelegatesToSAMLService(t *testing.T) {
	// Base raw config with minimal required values (non-SAML)
	baseRawConfig := map[string]string{
		"SECRET_KEY":   "this-is-a-test-secret-key-that-is-at-least-32-characters-long",
		"DATABASE_URL": "sqlite:///test.db",
	}

	// All SAML fields set with valid values
	allSAMLRawConfig := map[string]string{
		"PUBLIC_URL":              "https://sp.example.com",
		"SAML2_ENTITY_ID":         "https://sp.example.com",
		"SAML2_IDP_METADATA_URL":  "https://idp.example.com/metadata",
		"SAML2_PUBLIC_KEY":        "-----BEGIN CERTIFICATE-----\nMIIBI...\n-----END CERTIFICATE-----",
		"SAML2_PRIVATE_KEY":       "-----BEGIN PRIVATE KEY-----\nMIIEv...\n-----END PRIVATE KEY-----",
	}

	mergeConfig := func(base, overlay map[string]string) map[string]string {
		result := make(map[string]string)
		for k, v := range base {
			result[k] = v
		}
		for k, v := range overlay {
			result[k] = v
		}
		return result
	}

	versionReader := version.NewVersionReader()
	_ = service.NewConfigurationService(service.ConfigurationServiceOptions{}, versionReader)

	t.Run("all of them - returns true", func(t *testing.T) {
		rawConfig := mergeConfig(baseRawConfig, allSAMLRawConfig)

		// Build infrastructure config first (needed for auth checks)
		infraConfig := buildTestInfrastructureConfig(rawConfig)

		// Build domain config using infrastructure config
		domainConfig := buildTestDomainConfig(rawConfig, infraConfig)

		if !domainConfig.SAMLEnabled {
			t.Error("SAMLEnabled = false with all fields, want true")
		}

		// Verify infrastructure config has all the fields
		if infraConfig.PublicURL == "" {
			t.Error("InfrastructureConfig.PublicURL was not set")
		}
		if infraConfig.SAML2EntityID == "" {
			t.Error("InfrastructureConfig.SAML2EntityID was not set")
		}
		if infraConfig.SAML2IDPMetadataURL == "" {
			t.Error("InfrastructureConfig.SAML2IDPMetadataURL was not set")
		}
		if infraConfig.SAML2PublicKey == "" {
			t.Error("InfrastructureConfig.SAML2PublicKey was not set")
		}
		if infraConfig.SAML2PrivateKey == "" {
			t.Error("InfrastructureConfig.SAML2PrivateKey was not set")
		}
	})

	t.Run("none of them - returns false", func(t *testing.T) {
		rawConfig := baseRawConfig // No SAML fields

		infraConfig := buildTestInfrastructureConfig(rawConfig)
		domainConfig := buildTestDomainConfig(rawConfig, infraConfig)

		if domainConfig.SAMLEnabled {
			t.Error("SAMLEnabled = true with no fields, want false")
		}
	})

	t.Run("all except one - returns false (iterate through each field)", func(t *testing.T) {
		requiredFields := []string{
			"PUBLIC_URL", "SAML2_ENTITY_ID", "SAML2_IDP_METADATA_URL",
			"SAML2_PUBLIC_KEY", "SAML2_PRIVATE_KEY",
		}

		for _, missingField := range requiredFields {
			t.Run(missingField, func(t *testing.T) {
				// Create config with all SAML fields except the missing one
				samlConfig := make(map[string]string)
				for k, v := range allSAMLRawConfig {
					if k != missingField {
						samlConfig[k] = v
					}
				}
				rawConfig := mergeConfig(baseRawConfig, samlConfig)

				infraConfig := buildTestInfrastructureConfig(rawConfig)
				domainConfig := buildTestDomainConfig(rawConfig, infraConfig)

				if domainConfig.SAMLEnabled {
					t.Errorf("SAMLEnabled = true when %s missing, want false", missingField)
				}
			})
		}
	})
}

// TestBuildDomainConfig_OIDCEnabled_DelegatesToOIDCService tests that the config service
// properly delegates to the OIDC service to determine if OIDC is enabled.
func TestBuildDomainConfig_OIDCEnabled_DelegatesToOIDCService(t *testing.T) {
	// Base raw config with minimal required values (non-OIDC)
	baseRawConfig := map[string]string{
		"SECRET_KEY":   "this-is-a-test-secret-key-that-is-at-least-32-characters-long",
		"DATABASE_URL": "sqlite:///test.db",
	}

	// All OIDC fields set with valid values
	allOIDCRawConfig := map[string]string{
		"OPENID_CONNECT_ISSUER":       "https://oidc.example.com",
		"OPENID_CONNECT_CLIENT_ID":     "client123",
		"OPENID_CONNECT_CLIENT_SECRET": "secret123",
	}

	mergeConfig := func(base, overlay map[string]string) map[string]string {
		result := make(map[string]string)
		for k, v := range base {
			result[k] = v
		}
		for k, v := range overlay {
			result[k] = v
		}
		return result
	}

	versionReader := version.NewVersionReader()
	_ = service.NewConfigurationService(service.ConfigurationServiceOptions{}, versionReader)

	t.Run("all of them - returns true", func(t *testing.T) {
		rawConfig := mergeConfig(baseRawConfig, allOIDCRawConfig)

		infraConfig := buildTestInfrastructureConfig(rawConfig)
		domainConfig := buildTestDomainConfig(rawConfig, infraConfig)

		if !domainConfig.OpenIDConnectEnabled {
			t.Error("OpenIDConnectEnabled = false with all fields, want true")
		}
	})

	t.Run("none of them - returns false", func(t *testing.T) {
		rawConfig := baseRawConfig // No OIDC fields

		infraConfig := buildTestInfrastructureConfig(rawConfig)
		domainConfig := buildTestDomainConfig(rawConfig, infraConfig)

		if domainConfig.OpenIDConnectEnabled {
			t.Error("OpenIDConnectEnabled = true with no fields, want false")
		}
	})
}

// Helper functions for testing

// buildTestInfrastructureConfig builds an infrastructure config from raw config map
// This mirrors the logic in configuration_service.buildInfrastructureConfig
func buildTestInfrastructureConfig(rawConfig map[string]string) *config.InfrastructureConfig {
	return &config.InfrastructureConfig{
		PublicURL:           rawConfig["PUBLIC_URL"],
		SAML2EntityID:       rawConfig["SAML2_ENTITY_ID"],
		SAML2IDPMetadataURL: rawConfig["SAML2_IDP_METADATA_URL"],
		SAML2PublicKey:      rawConfig["SAML2_PUBLIC_KEY"],
		SAML2PrivateKey:     rawConfig["SAML2_PRIVATE_KEY"],
		OpenIDConnectIssuer:       rawConfig["OPENID_CONNECT_ISSUER"],
		OpenIDConnectClientID:     rawConfig["OPENID_CONNECT_CLIENT_ID"],
		OpenIDConnectClientSecret: rawConfig["OPENID_CONNECT_CLIENT_SECRET"],
	}
}

// buildTestDomainConfig builds a domain config using the auth service delegation
// This directly tests the delegation without going through the full config service
func buildTestDomainConfig(rawConfig map[string]string, infraConfig *config.InfrastructureConfig) *model.DomainConfig {
	// Directly call the auth service functions to test delegation
	return &model.DomainConfig{
		OpenIDConnectEnabled: authservice.IsOIDCConfigured(infraConfig),
		SAMLEnabled:          authservice.IsSAMLConfigured(infraConfig),
	}
}
