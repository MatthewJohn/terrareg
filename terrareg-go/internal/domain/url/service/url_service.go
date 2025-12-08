package service

import (
	"net/url"
	"strconv"

	infraConfig "github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/config"
)

// PublicURLDetails represents the details of the public URL
// Matching Python's get_public_url_details return tuple
type PublicURLDetails struct {
	Protocol string
	Domain   string
	Port     int
}

// URLService provides centralized URL handling functionality
// Following DDD principles by separating URL logic from business layers
type URLService struct {
	config *infraConfig.InfrastructureConfig
}

// NewURLService creates a new URL service
func NewURLService(config *infraConfig.InfrastructureConfig) *URLService {
	return &URLService{
		config: config,
	}
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
	protocol := "https"

	if s.config.PublicURL != "" {
		parsedURL, err := url.Parse(s.config.PublicURL)
		if err == nil && parsedURL.Hostname() != "" {
			// Only use values from parsed URL if it has a hostname,
			// otherwise it is invalid (matching Python logic)
			protocol = parsedURL.Scheme
			if protocol == "" {
				protocol = "https"
			}

			if parsedURL.Port() != "" {
				port, _ = strconv.Atoi(parsedURL.Port())
			} else {
				// Default port based on protocol
				if protocol == "http" {
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
		return details.Protocol + "://" + details.Domain
	}

	return details.Protocol + "://" + details.Domain + ":" + strconv.Itoa(details.Port)
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
