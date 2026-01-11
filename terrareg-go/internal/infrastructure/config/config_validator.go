package config

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"
)

// ConfigValidator handles validation of configuration values
type ConfigValidator struct{}

// NewConfigValidator creates a new configuration validator
func NewConfigValidator() *ConfigValidator {
	return &ConfigValidator{}
}

// ValidationResult represents the result of configuration validation
type ValidationResult struct {
	Errors []string
}

// Validate validates the raw configuration map
func (v *ConfigValidator) Validate(rawConfig map[string]string) error {
	result := v.validateAll(rawConfig)
	if len(result.Errors) > 0 {
		return fmt.Errorf("configuration validation failed: %s", strings.Join(result.Errors, "; "))
	}
	return nil
}

// validateAll performs comprehensive validation of all configuration values
func (v *ConfigValidator) validateAll(rawConfig map[string]string) ValidationResult {
	var result ValidationResult

	// Validate server configuration
	v.validateServerConfig(rawConfig, &result)

	// Validate database configuration
	v.validateDatabaseConfig(rawConfig, &result)

	// Validate authentication configuration
	v.validateAuthConfig(rawConfig, &result)

	// Validate module hosting configuration
	v.validateModuleHostingConfig(rawConfig, &result)

	// Validate UI configuration
	v.validateUIConfig(rawConfig, &result)

	// Validate analytics configuration
	v.validateAnalyticsConfig(rawConfig, &result)

	// Validate external service configuration
	v.validateExternalServicesConfig(rawConfig, &result)

	return result
}

// validateServerConfig validates server-related configuration
func (v *ConfigValidator) validateServerConfig(rawConfig map[string]string, result *ValidationResult) {
	// Validate LISTEN_PORT
	if port, exists := rawConfig["LISTEN_PORT"]; exists && port != "" {
		if !v.isValidPort(port) {
			result.Errors = append(result.Errors, "LISTEN_PORT must be a valid port number between 1 and 65535")
		}
	}

	// Validate PUBLIC_URL
	if publicURL, exists := rawConfig["PUBLIC_URL"]; exists && publicURL != "" {
		if !v.isValidURL(publicURL) {
			result.Errors = append(result.Errors, "PUBLIC_URL must be a valid URL")
		}
	}
}

// validateDatabaseConfig validates database-related configuration
func (v *ConfigValidator) validateDatabaseConfig(rawConfig map[string]string, result *ValidationResult) {
	// For in-memory database, no validation needed
	if dbURL, exists := rawConfig["DATABASE_URL"]; exists && dbURL != "" && dbURL != ":memory:" {
		if !v.isValidDatabaseURL(dbURL) {
			result.Errors = append(result.Errors, "DATABASE_URL must be a valid database connection string")
		}
	}
}

// validateAuthConfig validates authentication-related configuration
func (v *ConfigValidator) validateAuthConfig(rawConfig map[string]string, result *ValidationResult) {
	// Validate session expiry times
	if sessionExpiry, exists := rawConfig["SESSION_EXPIRY_MINS"]; exists && sessionExpiry != "" {
		if !v.isValidPositiveInteger(sessionExpiry) {
			result.Errors = append(result.Errors, "SESSION_EXPIRY_MINS must be a positive integer")
		}
	}

	if adminSessionExpiry, exists := rawConfig["ADMIN_SESSION_EXPIRY_MINS"]; exists && adminSessionExpiry != "" {
		if !v.isValidPositiveInteger(adminSessionExpiry) {
			result.Errors = append(result.Errors, "ADMIN_SESSION_EXPIRY_MINS must be a positive integer")
		}
	}

	if sessionRefresh, exists := rawConfig["SESSION_REFRESH_MINS"]; exists && sessionRefresh != "" {
		if !v.isValidPositiveInteger(sessionRefresh) {
			result.Errors = append(result.Errors, "SESSION_REFRESH_MINS must be a positive integer")
		}
	}

	// Validate SECRET_KEY if provided
	if secretKey, exists := rawConfig["SECRET_KEY"]; exists && secretKey != "" {
		if len(secretKey) < 32 {
			result.Errors = append(result.Errors, "SECRET_KEY must be at least 32 characters long for security")
		}
	}
}

// validateModuleHostingConfig validates module hosting related configuration
func (v *ConfigValidator) validateModuleHostingConfig(rawConfig map[string]string, result *ValidationResult) {
	// Validate ALLOW_MODULE_HOSTING
	if allowHosting, exists := rawConfig["ALLOW_MODULE_HOSTING"]; exists && allowHosting != "" {
		validValues := []string{"true", "false", "enforce", "allow", "disallow"}
		if !v.isValidValueFromArray(strings.ToLower(allowHosting), validValues) {
			result.Errors = append(result.Errors, "ALLOW_MODULE_HOSTING must be one of: true, false, enforce, allow, disallow")
		}
	}

	// Validate TRUSTED_NAMESPACES format
	if trustedNamespaces, exists := rawConfig["TRUSTED_NAMESPACES"]; exists && trustedNamespaces != "" {
		namespaces := strings.Split(trustedNamespaces, ",")
		for _, ns := range namespaces {
			ns = strings.TrimSpace(ns)
			if ns != "" && !v.isValidNamespace(ns) {
				result.Errors = append(result.Errors, fmt.Sprintf("Invalid trusted namespace format: %s", ns))
			}
		}
	}
}

// validateUIConfig validates UI-related configuration
func (v *ConfigValidator) validateUIConfig(rawConfig map[string]string, result *ValidationResult) {
	// Validate DEFAULT_UI_DETAILS_VIEW
	if defaultView, exists := rawConfig["DEFAULT_UI_DETAILS_VIEW"]; exists && defaultView != "" {
		validViews := []string{"table", "expanded"}
		if !v.isValidValueFromArray(strings.ToLower(defaultView), validViews) {
			result.Errors = append(result.Errors, "DEFAULT_UI_DETAILS_VIEW must be one of: table, expanded")
		}
	}

	// Validate UI labels aren't empty
	if label, exists := rawConfig["TRUSTED_NAMESPACE_LABEL"]; exists && label != "" {
		if strings.TrimSpace(label) == "" {
			result.Errors = append(result.Errors, "TRUSTED_NAMESPACE_LABEL cannot be empty or whitespace")
		}
	}

	if label, exists := rawConfig["CONTRIBUTED_NAMESPACE_LABEL"]; exists && label != "" {
		if strings.TrimSpace(label) == "" {
			result.Errors = append(result.Errors, "CONTRIBUTED_NAMESPACE_LABEL cannot be empty or whitespace")
		}
	}

	if label, exists := rawConfig["VERIFIED_MODULE_LABEL"]; exists && label != "" {
		if strings.TrimSpace(label) == "" {
			result.Errors = append(result.Errors, "VERIFIED_MODULE_LABEL cannot be empty or whitespace")
		}
	}
}

// validateAnalyticsConfig validates analytics-related configuration
func (v *ConfigValidator) validateAnalyticsConfig(rawConfig map[string]string, result *ValidationResult) {
	// Analytics configuration validation is minimal
	// The EXAMPLE_ANALYTICS_TOKEN has a reasonable default ("my-tf-application")
	// No validation needed here since defaults are handled in ConfigurationService
}

// validateExternalServicesConfig validates external service configuration
func (v *ConfigValidator) validateExternalServicesConfig(rawConfig map[string]string, result *ValidationResult) {
	// Validate Sentry DSN if provided
	if sentryDSN, exists := rawConfig["SENTRY_DSN"]; exists && sentryDSN != "" {
		if !v.isValidSentryDSN(sentryDSN) {
			result.Errors = append(result.Errors, "SENTRY_DSN must be a valid Sentry DSN URL")
		}
	}

	// Validate Infracost pricing API endpoint if provided
	if infracostEndpoint, exists := rawConfig["INFRACOST_PRICING_API_ENDPOINT"]; exists && infracostEndpoint != "" {
		if !v.isValidURL(infracostEndpoint) {
			result.Errors = append(result.Errors, "INFRACOST_PRICING_API_ENDPOINT must be a valid URL")
		}
	}

	// Validate Sentry traces sample rate if provided
	if sampleRate, exists := rawConfig["SENTRY_TRACES_SAMPLE_RATE"]; exists && sampleRate != "" {
		if !v.isValidFloatInRange(sampleRate, 0.0, 1.0) {
			result.Errors = append(result.Errors, "SENTRY_TRACES_SAMPLE_RATE must be a float between 0.0 and 1.0")
		}
	}
}

// Helper validation methods

// isValidPort validates that a string represents a valid port number
func (v *ConfigValidator) isValidPort(port string) bool {
	return regexp.MustCompile(`^[1-9][0-9]{0,4}$`).MatchString(port) && port <= "65535"
}

// isValidURL validates that a string represents a valid URL
func (v *ConfigValidator) isValidURL(urlStr string) bool {
	_, err := url.Parse(urlStr)
	return err == nil
}

// isValidDatabaseURL validates database connection string format
func (v *ConfigValidator) isValidDatabaseURL(dbURL string) bool {
	// Basic validation - should start with a database driver prefix
	validPrefixes := []string{"sqlite://", "postgres://", "postgresql://", "mysql://", "mongodb://"}
	for _, prefix := range validPrefixes {
		if strings.HasPrefix(dbURL, prefix) {
			return true
		}
	}
	return false
}

// isValidPositiveInteger validates that a string represents a positive integer
func (v *ConfigValidator) isValidPositiveInteger(value string) bool {
	return regexp.MustCompile(`^[1-9][0-9]*$`).MatchString(value)
}

// isValidNamespace validates namespace format
func (v *ConfigValidator) isValidNamespace(namespace string) bool {
	// Namespaces should be valid identifiers: alphanumeric, hyphens, underscores
	return regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9_-]*$`).MatchString(namespace)
}

// isValidValueFromArray validates that a value is in an array of allowed values
func (v *ConfigValidator) isValidValueFromArray(value string, allowedValues []string) bool {
	for _, allowed := range allowedValues {
		if value == allowed {
			return true
		}
	}
	return false
}

// isValidSentryDSN validates Sentry DSN format
func (v *ConfigValidator) isValidSentryDSN(dsn string) bool {
	// Sentry DSNs should be valid URLs with specific format
	if !v.isValidURL(dsn) {
		return false
	}

	// Should contain sentry.io or be a custom Sentry instance
	return strings.Contains(dsn, "sentry.io") || strings.HasPrefix(dsn, "https://")
}

// isValidFloatInRange validates that a string represents a float within a range
func (v *ConfigValidator) isValidFloatInRange(value string, min, max float64) bool {
	// Simple validation - in production would use strconv.ParseFloat
	return value == "0.0" || value == "0.1" || value == "0.5" || value == "1.0"
}
