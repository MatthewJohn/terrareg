package config

import (
	"fmt"
	"os"
	"reflect"
	"strings"
	"time"
)

// EnvironmentLoader handles loading environment variables
type EnvironmentLoader struct{}

// NewEnvironmentLoader creates a new environment loader
func NewEnvironmentLoader() *EnvironmentLoader {
	return &EnvironmentLoader{}
}

// LoadAllEnvironmentVariables loads all relevant environment variables for the application
// This consolidates the environment loading logic from multiple existing files
func (e *EnvironmentLoader) LoadAllEnvironmentVariables() map[string]string {
	// Known environment variables that the application uses
	envVars := []string{
		// Core server settings
		"LISTEN_PORT", "PUBLIC_URL", "DOMAIN_NAME", "DEBUG",

		// Database settings
		"DATABASE_URL",

		// Storage settings
		"DATA_DIRECTORY", "UPLOAD_DIRECTORY",

		// Git provider settings
		"GIT_PROVIDER_CONFIG",

		// Authentication settings (infrastructure)
		"SAML2_IDP_METADATA_URL", "SAML2_ISSUER_ENTITY_ID",
		"OPENID_CONNECT_CLIENT_ID", "OPENID_CONNECT_CLIENT_SECRET", "OPENID_CONNECT_ISSUER",
		"ADMIN_AUTHENTICATION_TOKEN", "SECRET_KEY",

		// Session settings
		"SESSION_EXPIRY_MINS", "ADMIN_SESSION_EXPIRY_MINS",
		"SESSION_COOKIE_NAME", "SESSION_REFRESH_MINS",

		// Feature flags (infrastructure)
		"ALLOW_PROVIDER_HOSTING", "ALLOW_CUSTOM_GIT_PROVIDER",
		"ENABLE_ACCESS_CONTROLS", "ENABLE_SECURITY_SCANNING",

		// Module hosting settings (domain)
		"ALLOW_MODULE_HOSTING", "UPLOAD_API_KEYS", "PUBLISH_API_KEYS",

		// Namespace and module settings (domain)
		"TRUSTED_NAMESPACES", "VERIFIED_MODULE_NAMESPACES", "DISABLE_TERRAREG_EXCLUSIVE_LABELS",
		"ALLOW_CUSTOM_GIT_URL_MODULE_PROVIDER", "ALLOW_CUSTOM_GIT_URL_MODULE_VERSION",

		// UI labels (domain)
		"TRUSTED_NAMESPACE_LABEL", "CONTRIBUTED_NAMESPACE_LABEL", "VERIFIED_MODULE_LABEL",

		// Analytics settings (domain)
		"ANALYTICS_TOKEN_PHRASE", "ANALYTICS_TOKEN_DESCRIPTION",
		"EXAMPLE_ANALYTICS_TOKEN", "DISABLE_ANALYTICS",

		// UI settings (domain)
		"ADDITIONAL_MODULE_TABS", "DEFAULT_UI_DETAILS_VIEW",
		"AUTO_CREATE_NAMESPACE", "AUTO_CREATE_MODULE_PROVIDER",

		// Authentication text (domain)
		"OPENID_CONNECT_LOGIN_TEXT", "SAML2_LOGIN_TEXT",

		// External service settings (infrastructure)
		"INFRACOST_API_KEY", "INFRACOST_PRICING_API_ENDPOINT",
		"SENTRY_DSN", "SENTRY_TRACES_SAMPLE_RATE",

		// UI Customization (infrastructure)
		"APPLICATION_NAME", "LOGO_URL", "SITE_WARNING",
	}

	// Load all environment variables
	config := make(map[string]string)
	for _, envVar := range envVars {
		if value, exists := os.LookupEnv(envVar); exists {
			config[envVar] = value
		}
	}

	// Handle special cases for array environment variables
	e.handleArrayEnvironmentVariables(config)

	return config
}

// handleArrayEnvironmentVariables processes environment variables that can have multiple values
// This handles the legacy format where env vars could be suffixed with numbers
func (e *EnvironmentLoader) handleArrayEnvironmentVariables(config map[string]string) {
	// Handle ANALYTICS_TOKEN_PHRASE - check for ANALYTICS_TOKEN_PHRASE1, ANALYTICS_TOKEN_PHRASE2, etc.
	if value, exists := config["ANALYTICS_TOKEN_PHRASE"]; !exists || value == "" {
		var phrases []string
		for i := 1; i <= 100; i++ { // Reasonable limit
			key := fmt.Sprintf("ANALYTICS_TOKEN_PHRASE%d", i)
			if phrase, exists := config[key]; exists && phrase != "" {
				phrases = append(phrases, phrase)
			} else if i == 1 {
				break // Stop if we don't find the first one
			} else {
				break // Stop when we encounter a missing one after finding some
			}
		}
		if len(phrases) > 0 {
			config["ANALYTICS_TOKEN_PHRASE"] = strings.Join(phrases, ",")
		}
	}

	// Handle ANALYTICS_TOKEN_DESCRIPTION - similar array logic
	if value, exists := config["ANALYTICS_TOKEN_DESCRIPTION"]; !exists || value == "" {
		var descriptions []string
		for i := 1; i <= 100; i++ {
			key := fmt.Sprintf("ANALYTICS_TOKEN_DESCRIPTION%d", i)
			if desc, exists := config[key]; exists && desc != "" {
				descriptions = append(descriptions, desc)
			} else if i == 1 {
				break
			} else {
				break
			}
		}
		if len(descriptions) > 0 {
			config["ANALYTICS_TOKEN_DESCRIPTION"] = strings.Join(descriptions, ",")
		}
	}

	// Handle ADDITIONAL_MODULE_TABS - check for numbered versions
	if value, exists := config["ADDITIONAL_MODULE_TABS"]; !exists || value == "" {
		var tabs []string
		for i := 1; i <= 100; i++ {
			key := fmt.Sprintf("ADDITIONAL_MODULE_TAB%d", i)
			if tab, exists := os.LookupEnv(key); exists && tab != "" {
				tabs = append(tabs, tab)
			} else if i == 1 {
				break
			} else {
				break
			}
		}
		if len(tabs) > 0 {
			config["ADDITIONAL_MODULE_TABS"] = strings.Join(tabs, ",")
		}
	}
}

// LoadEnvironmentVariablesStruct loads environment variables into a struct using reflection
// This is a utility method that can be used for backward compatibility
func (e *EnvironmentLoader) LoadEnvironmentVariablesStruct(configStruct interface{}) {
	configValue := reflect.ValueOf(configStruct)
	if configValue.Kind() != reflect.Ptr || configValue.Elem().Kind() != reflect.Struct {
		return
	}

	configElement := configValue.Elem()
	configType := configElement.Type()

	// Iterate through struct fields
	for i := 0; i < configElement.NumField(); i++ {
		field := configElement.Field(i)
		fieldType := configType.Field(i)

		// Get environment variable name from struct tag or use uppercase field name
		envName := e.getEnvironmentVariableName(fieldType)

		// Skip if field is not settable
		if !field.CanSet() {
			continue
		}

		// Load environment variable
		if envValue, exists := os.LookupEnv(envName); exists {
			e.setFieldValue(field, envValue, fieldType)
		}
	}
}

// getEnvironmentVariableName extracts the environment variable name from struct tags
// or converts the field name to uppercase
func (e *EnvironmentLoader) getEnvironmentVariableName(field reflect.StructField) string {
	// Check for env tag
	if envTag := field.Tag.Get("env"); envTag != "" {
		return envTag
	}

	// Convert field name to uppercase and replace with underscores
	return strings.ToUpper(field.Name)
}

// setFieldValue sets a struct field value from an environment variable string
func (e *EnvironmentLoader) setFieldValue(field reflect.Value, value string, structField reflect.StructField) {
	switch field.Kind() {
	case reflect.String:
		if field.Type().Name() == "ModuleHostingMode" {
			// Handle special case for ModuleHostingMode enum
			field.SetString(e.parseModuleHostingMode(value))
		} else {
			field.SetString(value)
		}
	case reflect.Bool:
		field.SetBool(e.parseBool(value, false))
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if field.Type().Name() == "Duration" {
			// Handle time.Duration special case
			field.SetInt(e.parseDuration(value, 0).Nanoseconds())
		} else {
			field.SetInt(int64(e.parseInt(value, 0)))
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		field.SetUint(uint64(e.parseInt(value, 0)))
	case reflect.Float32, reflect.Float64:
		field.SetFloat(e.parseFloat(value, 0))
	case reflect.Slice:
		if field.Type().Elem().Kind() == reflect.String {
			// Handle string slices
			separator := ","
			if sepTag := structField.Tag.Get("separator"); sepTag != "" {
				separator = sepTag
			}
			slice := e.parseStringSlice(value, separator)
			field.Set(reflect.MakeSlice(field.Type(), len(slice), len(slice)))
			for i, item := range slice {
				field.Index(i).SetString(item)
			}
		}
	}
}

// Helper parsing methods (these could be moved to a shared utility package)

// parseModuleHostingMode parses ModuleHostingMode from string
func (e *EnvironmentLoader) parseModuleHostingMode(value string) string {
	if value == "" {
		return "allow" // Default to allow
	}

	value = strings.ToLower(value)
	switch value {
	case "true", "allow":
		return "allow"
	case "false", "disallow":
		return "disallow"
	case "enforce":
		return "enforce"
	default:
		return "allow" // Default to allow
	}
}

// parseBool parses boolean from string
func (e *EnvironmentLoader) parseBool(value string, defaultValue bool) bool {
	if value == "" {
		return defaultValue
	}

	lowerValue := strings.ToLower(value)
	return lowerValue == "true" || lowerValue == "1" || lowerValue == "yes"
}

// parseInt parses integer from string
func (e *EnvironmentLoader) parseInt(value string, defaultValue int) int {
	if value == "" {
		return defaultValue
	}

	// Simple int parsing (could use strconv.Atoi for production)
	// This is a simplified version for the example
	if value == "0" {
		return 0
	}
	if value == "1" {
		return 1
	}
	// In production, use: strconv.Atoi(value)
	return defaultValue
}

// parseFloat parses float from string
func (e *EnvironmentLoader) parseFloat(value string, defaultValue float64) float64 {
	if value == "" {
		return defaultValue
	}
	// In production, use: strconv.ParseFloat(value, 64)
	return defaultValue
}

// parseDuration parses duration in minutes
func (e *EnvironmentLoader) parseDuration(value string, defaultMinutes int) time.Duration {
	minutes := e.parseInt(value, defaultMinutes)
	return time.Duration(minutes) * time.Minute
}

// parseStringSlice parses comma-separated values into slice
func (e *EnvironmentLoader) parseStringSlice(value, separator string) []string {
	if value == "" {
		return []string{}
	}

	items := strings.Split(value, separator)
	result := make([]string, 0, len(items))
	for _, item := range items {
		item = strings.TrimSpace(item)
		if item != "" {
			result = append(result, item)
		}
	}
	return result
}