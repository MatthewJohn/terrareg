package service_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/url/service"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/config"
)

// TestBuildTerraformSourceURL_EdgeCases tests edge cases
func TestBuildUrlServiceWithoutConfig(t *testing.T) {
	urlService, err := service.NewURLService(nil)
	assert.Error(t, err)
	assert.Nil(t, urlService)
}

// TestBuildTerraformSourceURL_HTTP tests terraform source URL generation for HTTP
func TestBuildTerraformSourceURL_HTTP(t *testing.T) {
	tests := []struct {
		name           string
		publicURL      string
		providerID     string
		version        string
		modulePath     string
		requestDomain  string
		expectedResult string
	}{
		{
			name:           "HTTP with standard port",
			publicURL:      "http://localhost:80",
			providerID:     "test-namespace/test-module/aws",
			version:        "1.0.0",
			modulePath:     "modules/example-submodule",
			requestDomain:  "",
			expectedResult: "http://localhost/modules/test-namespace/test-module/aws/1.0.0//modules/example-submodule",
		},
		{
			name:           "HTTP with non-standard port",
			publicURL:      "http://localhost:5000",
			providerID:     "test-namespace/test-module/aws",
			version:        "1.0.0",
			modulePath:     "modules/example-submodule",
			requestDomain:  "",
			expectedResult: "http://localhost:5000/modules/test-namespace/test-module/aws/1.0.0//modules/example-submodule",
		},
		{
			name:           "HTTP with version in URL",
			publicURL:      "http://example.com:8080",
			providerID:     "ns/mod/prov",
			version:        "2.3.4",
			modulePath:     "submodules/foo",
			requestDomain:  "",
			expectedResult: "http://example.com:8080/modules/ns/mod/prov/2.3.4//submodules/foo",
		},
		{
			name:           "HTTP without module path (root module)",
			publicURL:      "http://localhost:5000",
			providerID:     "test-namespace/test-module/aws",
			version:        "1.0.0",
			modulePath:     "",
			requestDomain:  "",
			expectedResult: "http://localhost:5000/modules/test-namespace/test-module/aws/1.0.0",
		},
		{
			name:           "HTTP with leading slashes in module path (should be stripped)",
			publicURL:      "http://localhost:5000",
			providerID:     "test-namespace/test-module/aws",
			version:        "1.0.0",
			modulePath:     "///modules/example-submodule",
			requestDomain:  "",
			expectedResult: "http://localhost:5000/modules/test-namespace/test-module/aws/1.0.0//modules/example-submodule",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			infraConfig := &config.InfrastructureConfig{
				PublicURL: tc.publicURL,
			}
			urlService, err := service.NewURLService(infraConfig)
			assert.NoError(t, err)

			result := urlService.BuildTerraformSourceURL(tc.providerID, tc.version, tc.modulePath, tc.requestDomain)

			assert.Equal(t, tc.expectedResult, result)
		})
	}
}

// TestBuildTerraformSourceURL_HTTPS tests terraform source URL generation for HTTPS
func TestBuildTerraformSourceURL_HTTPS(t *testing.T) {
	tests := []struct {
		name           string
		publicURL      string
		providerID     string
		version        string
		modulePath     string
		requestDomain  string
		expectedResult string
	}{
		{
			name:           "HTTPS with standard port 443",
			publicURL:      "https://example.com",
			providerID:     "test-namespace/test-module/aws",
			version:        "1.0.0",
			modulePath:     "modules/example-submodule",
			requestDomain:  "",
			expectedResult: "example.com/test-namespace/test-module/aws//modules/example-submodule",
		},
		{
			name:           "HTTPS with explicit standard port 443",
			publicURL:      "https://example.com:443",
			providerID:     "test-namespace/test-module/aws",
			version:        "1.0.0",
			modulePath:     "modules/example-submodule",
			requestDomain:  "",
			expectedResult: "example.com/test-namespace/test-module/aws//modules/example-submodule",
		},
		{
			name:           "HTTPS with non-standard port",
			publicURL:      "https://local-dev.dock.studio:5000",
			providerID:     "test-namespace/test-module/aws",
			version:        "1.0.0",
			modulePath:     "modules/example-submodule",
			requestDomain:  "",
			expectedResult: "local-dev.dock.studio:5000/test-namespace/test-module/aws//modules/example-submodule",
		},
		{
			name:           "HTTPS without version in URL (version not included for HTTPS)",
			publicURL:      "https://example.com:8443",
			providerID:     "ns/mod/prov",
			version:        "2.3.4",
			modulePath:     "submodules/foo",
			requestDomain:  "",
			expectedResult: "example.com:8443/ns/mod/prov//submodules/foo",
		},
		{
			name:           "HTTPS without module path (root module)",
			publicURL:      "https://local-dev.dock.studio:5000",
			providerID:     "test-namespace/test-module/aws",
			version:        "1.0.0",
			modulePath:     "",
			requestDomain:  "",
			expectedResult: "local-dev.dock.studio:5000/test-namespace/test-module/aws",
		},
		{
			name:           "HTTPS with leading slashes in module path (should be stripped)",
			publicURL:      "https://example.com",
			providerID:     "test-namespace/test-module/aws",
			version:        "1.0.0",
			modulePath:     "///modules/example-submodule",
			requestDomain:  "",
			expectedResult: "example.com/test-namespace/test-module/aws//modules/example-submodule",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			infraConfig := &config.InfrastructureConfig{
				PublicURL: tc.publicURL,
			}
			urlService, err := service.NewURLService(infraConfig)
			assert.NoError(t, err)

			result := urlService.BuildTerraformSourceURL(tc.providerID, tc.version, tc.modulePath, tc.requestDomain)

			assert.Equal(t, tc.expectedResult, result)
		})
	}
}

// TestBuildTerraformSourceURL_WithRequestDomain tests that request domain is used as fallback
func TestBuildTerraformSourceURL_WithRequestDomain(t *testing.T) {
	tests := []struct {
		name           string
		publicURL      string
		providerID     string
		version        string
		modulePath     string
		requestDomain  string
		expectedResult string
	}{
		{
			name:          "Request domain fallback when PUBLIC_URL not set (defaults to HTTPS)",
			publicURL:     "",
			providerID:    "test-namespace/test-module/aws",
			version:       "1.0.0",
			modulePath:    "modules/example-submodule",
			requestDomain: "example.com:8080",
			// When PUBLIC_URL is not set, defaults to HTTPS protocol, so no /modules/ or version in URL
			expectedResult: "example.com:8080/test-namespace/test-module/aws//modules/example-submodule",
		},
		{
			name:           "PUBLIC_URL takes precedence over request domain",
			publicURL:      "http://public.example.com:5000",
			providerID:     "test-namespace/test-module/aws",
			version:        "1.0.0",
			modulePath:     "modules/example-submodule",
			requestDomain:  "request.example.com",
			expectedResult: "http://public.example.com:5000/modules/test-namespace/test-module/aws/1.0.0//modules/example-submodule",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			infraConfig := &config.InfrastructureConfig{
				PublicURL: tc.publicURL,
			}
			urlService, err := service.NewURLService(infraConfig)
			assert.NoError(t, err)

			result := urlService.BuildTerraformSourceURL(tc.providerID, tc.version, tc.modulePath, tc.requestDomain)

			assert.Equal(t, tc.expectedResult, result)
		})
	}
}

// TestBuildTerraformSourceURL_EdgeCases tests edge cases
func TestBuildTerraformSourceURL_EdgeCases(t *testing.T) {
	tests := []struct {
		name           string
		publicURL      string
		providerID     string
		version        string
		modulePath     string
		requestDomain  string
		expectedResult string
	}{
		{
			name:           "Empty version",
			publicURL:      "http://localhost:5000",
			providerID:     "test-namespace/test-module/aws",
			version:        "",
			modulePath:     "modules/example-submodule",
			requestDomain:  "",
			expectedResult: "http://localhost:5000/modules/test-namespace/test-module/aws//modules/example-submodule",
		},
		{
			name:           "Module path with special characters",
			publicURL:      "https://example.com",
			providerID:     "test-namespace/test-module/aws",
			version:        "1.0.0",
			modulePath:     "modules/sub-module/v2.0",
			requestDomain:  "",
			expectedResult: "example.com/test-namespace/test-module/aws//modules/sub-module/v2.0",
		},
		{
			name:           "Provider ID is always included",
			publicURL:      "https://local-dev.dock.studio:5000",
			providerID:     "moduledetails/fullypopulated/testprovider",
			version:        "1.5.0",
			modulePath:     "modules/example-submodule1",
			requestDomain:  "",
			expectedResult: "local-dev.dock.studio:5000/moduledetails/fullypopulated/testprovider//modules/example-submodule1",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			infraConfig := &config.InfrastructureConfig{
				PublicURL: tc.publicURL,
			}
			urlService, err := service.NewURLService(infraConfig)
			assert.NoError(t, err)

			result := urlService.BuildTerraformSourceURL(tc.providerID, tc.version, tc.modulePath, tc.requestDomain)

			assert.Equal(t, tc.expectedResult, result)
		})
	}
}
