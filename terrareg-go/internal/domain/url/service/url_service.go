package service

import (
	"fmt"
	"net/url"
	"strconv"

	infraConfig "github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/config"
)

// URL Protocol
type Protocol string

const (
	ProtocolHttp  Protocol = "http"
	ProtocolHttps Protocol = "https"
)

// PublicURLDetails represents the details of the public URL
// Matching Python's get_public_url_details return tuple
type PublicURLDetails struct {
	Protocol Protocol
	Domain   string
	Port     int
}

// URLService provides centralized URL handling functionality
// Following DDD principles by separating URL logic from business layers
type URLService struct {
	config *infraConfig.InfrastructureConfig
}

// NewURLService creates a new URL service
func NewURLService(config *infraConfig.InfrastructureConfig) (*URLService, error) {
	if config == nil {
		return nil, fmt.Errorf("NewURLService cannot be passed nil config")
	}
	return &URLService{
		config: config,
	}, nil
}

// GetPublicURLDetails returns protocol, domain, and port used to access Terrareg
// This matches Python's get_public_url_details function exactly
func (s *URLService) GetPublicURLDetails(fallbackDomain *string) *PublicURLDetails {
	// Set default values matching Python implementation
	domain := s.config.DomainName
	if domain == "" && fallbackDomain != nil {
		domain = *fallbackDomain
	}

	port := 443
	protocol := ProtocolHttps

	if s.config.PublicURL != "" {
		parsedURL, err := url.Parse(s.config.PublicURL)
		if err == nil && parsedURL.Hostname() != "" {
			// Only use values from parsed URL if it has a hostname,
			// otherwise it is invalid (matching Python logic)
			if parsedURL.Scheme == "http" {
				protocol = ProtocolHttp
			}

			if parsedURL.Port() != "" {
				port, _ = strconv.Atoi(parsedURL.Port())
			} else {
				// Default port based on protocol
				if protocol == ProtocolHttp {
					port = 80
				} else {
					port = 443
				}
			}

			domain = parsedURL.Hostname()
		}
	}

	return &PublicURLDetails{
		Protocol: protocol,
		Domain:   domain,
		Port:     port,
	}
}

// IsHTTPS determines if the application is running in HTTPS mode
// This centralizes HTTPS detection logic used across auth components
func (s *URLService) IsHTTPS(fallbackDomain *string) bool {
	details := s.GetPublicURLDetails(fallbackDomain)
	return details.Protocol == "https"
}

// GetBaseURL returns the base URL from configuration
func (s *URLService) GetBaseURL(fallbackDomain *string) string {
	details := s.GetPublicURLDetails(fallbackDomain)

	// Handle standard ports - don't include them in the URL
	if (details.Protocol == "https" && details.Port == 443) ||
		(details.Protocol == "http" && details.Port == 80) {
		return string(details.Protocol) + "://" + details.Domain
	}

	return string(details.Protocol) + "://" + details.Domain + ":" + strconv.Itoa(details.Port)
}

// BuildURL constructs a URL with the given path using the base configuration
func (s *URLService) BuildURL(path string, fallbackDomain *string) string {
	baseURL := s.GetBaseURL(fallbackDomain)

	// Ensure proper path joining
	if len(path) > 0 && path[0] != '/' {
		path = "/" + path
	}

	return baseURL + path
}

// GetHostWithPort returns the host including port if non-standard
func (s *URLService) GetHostWithPort(fallbackDomain *string) string {
	details := s.GetPublicURLDetails(fallbackDomain)

	// Handle standard ports - don't include them in the host
	if (details.Protocol == "https" && details.Port == 443) ||
		(details.Protocol == "http" && details.Port == 80) {
		return details.Domain
	}

	return details.Domain + ":" + strconv.Itoa(details.Port)
}

// BuildTerraformSourceURL builds the terraform source URL for a module
// Python reference: /app/terrareg/models.py TerraformSpecsObject.get_terraform_url_and_version_strings()
// For HTTP: http://{domain}:{port}/modules/{provider_id}/{version}//{module_path}
// For HTTPS: {domain}:{port}/{provider_id}//{module_path}
// Domain is ALWAYS included
func (s *URLService) BuildTerraformSourceURL(providerID, version, modulePath, requestDomain string) string {
	// Get public URL details (protocol, domain, port)
	// Python reference: get_public_url_details(fallback_domain=request_domain)
	var domainPtr *string
	if requestDomain != "" {
		domainPtr = &requestDomain
	}
	details := s.GetPublicURLDetails(domainPtr)

	// Check if using HTTPS - use same domainPtr to get consistent protocol
	// Python: isHttps = protocol.lower() == "https"
	isHttps := s.IsHTTPS(domainPtr)

	// Determine if port should be included (non-standard ports only)
	// Python: isDefaultPort = not port or (str(port) == "443" and isHttps) or (str(port) == "80" and not isHttps)
	isDefaultPort := details.Port == 0 ||
		(details.Port == 443 && isHttps) ||
		(details.Port == 80 && !isHttps)

	// Build source URL
	// Python: source_url = '' if isHttps else 'http://'
	var sourceURL string
	if !isHttps {
		sourceURL = "http://"
	}

	// Python: source_url += domain (domain is ALWAYS added)
	sourceURL += details.Domain

	// Python: source_url += '' if isDefaultPort else f':{port}'
	if !isDefaultPort && details.Port != 0 {
		sourceURL += fmt.Sprintf(":%d", details.Port)
	}

	// Python: source_url += '' if isHttps else '/modules'
	if !isHttps {
		sourceURL += "/modules"
	}

	// Python: source_url += '/'
	sourceURL += "/"

	// Python: Add analytics token if configured (TODO: implement this)
	// source_url += (EXAMPLE_ANALYTICS_TOKEN + '__') if EXAMPLE_ANALYTICS_TOKEN and not DISABLE_ANALYTICS else ''

	// Python: source_url += module_provider_id
	sourceURL += providerID

	// Python: source_url += '' if isHttps else f'/{version}'
	if !isHttps && version != "" {
		sourceURL += "/" + version
	}

	// Python: source_url += f'//{module_path}' if module_path else ''
	// Remove any leading slashes from modulePath (Python: module_path = re.sub(r'^\/+', '', module_path))
	cleanPath := modulePath
	for len(cleanPath) > 0 && cleanPath[0] == '/' {
		cleanPath = cleanPath[1:]
	}
	if cleanPath != "" {
		sourceURL += "//" + cleanPath
	}

	return sourceURL
}
