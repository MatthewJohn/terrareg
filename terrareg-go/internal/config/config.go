package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// Config holds all application configuration
type Config struct {
	// Server settings
	ListenPort int
	PublicURL  string
	DomainName string
	Debug      bool

	// Database settings
	DatabaseURL string

	// Storage settings
	DataDirectory   string
	UploadDirectory string

	// Git provider settings
	GitProviderConfig string

	// Authentication settings
	// SAML
	SAML2IDPMetadataURL string
	SAML2IssuerEntityID string
	SAML2Enabled        bool

	// OpenID Connect
	OpenIDConnectEnabled      bool
	OpenIDConnectClientID     string
	OpenIDConnectClientSecret string
	OpenIDConnectIssuer       string

	// Admin authentication
	AdminAuthenticationToken string
	SecretKey                string

	// Feature flags
	AllowModuleHosting     bool
	AllowProviderHosting   bool
	AllowCustomGitProvider bool
	EnableAccessControls   bool
	EnableSecurityScanning bool

	// UI Customization
	ApplicationName string
	LogoURL         string
	SiteWarning     string

	// External services
	InfracostAPIKey             string
	InfracostPricingAPIEndpoint string
	SentryDSN                   string
	SentryTracesSampleRate      float64

	// Session settings
	SessionExpiry          time.Duration
	AdminSessionExpiryMins int
	SessionCookieName      string
	SessionRefreshAge      time.Duration

	// Namespace settings
	VerifiedModuleNamespaces []string
	TrustedNamespaces        []string

	// Provider source settings
	ProviderSources map[string]ProviderSourceConfig
}

// ProviderSourceConfig holds configuration for external provider sources
type ProviderSourceConfig struct {
	Type         string
	APIName      string
	ClientID     string
	ClientSecret string
	LoginURL     string
	CallbackURL  string
}

// New creates a new Config from environment variables
func New() (*Config, error) {
	cfg := &Config{
		ListenPort:        getEnvInt("LISTEN_PORT", 5000),
		PublicURL:         getEnv("PUBLIC_URL", "http://localhost:5000"),
		DomainName:        getEnv("DOMAIN_NAME", ""),
		Debug:             getEnvBool("DEBUG", false),
		DatabaseURL:       getEnv("DATABASE_URL", "sqlite:///modules.db"),
		DataDirectory:     getEnv("DATA_DIRECTORY", "./data"),
		UploadDirectory:   getEnv("UPLOAD_DIRECTORY", "./data/upload"),
		GitProviderConfig: getEnv("GIT_PROVIDER_CONFIG", ""),

		// SAML
		SAML2IDPMetadataURL: getEnv("SAML2_IDP_METADATA_URL", ""),
		SAML2IssuerEntityID: getEnv("SAML2_ISSUER_ENTITY_ID", ""),

		// OpenID Connect
		OpenIDConnectClientID:     getEnv("OPENID_CONNECT_CLIENT_ID", ""),
		OpenIDConnectClientSecret: getEnv("OPENID_CONNECT_CLIENT_SECRET", ""),
		OpenIDConnectIssuer:       getEnv("OPENID_CONNECT_ISSUER", ""),

		// Admin auth
		AdminAuthenticationToken: getEnv("ADMIN_AUTHENTICATION_TOKEN", ""),
		SecretKey:                getEnv("SECRET_KEY", ""),

		// Feature flags
		AllowModuleHosting:     getEnvBool("ALLOW_MODULE_HOSTING", true),
		AllowProviderHosting:   getEnvBool("ALLOW_PROVIDER_HOSTING", true),
		AllowCustomGitProvider: getEnvBool("ALLOW_CUSTOM_GIT_PROVIDER", true),
		EnableAccessControls:   getEnvBool("ENABLE_ACCESS_CONTROLS", false),
		EnableSecurityScanning: getEnvBool("ENABLE_SECURITY_SCANNING", true),

		// UI Customization
		ApplicationName: getEnv("APPLICATION_NAME", "Terrareg"),
		LogoURL:         getEnv("LOGO_URL", "/static/images/logo.png"),
		SiteWarning:     getEnv("SITE_WARNING", ""),

		// External services
		InfracostAPIKey:             getEnv("INFRACOST_API_KEY", ""),
		InfracostPricingAPIEndpoint: getEnv("INFRACOST_PRICING_API_ENDPOINT", ""),
		SentryDSN:                   getEnv("SENTRY_DSN", ""),
		SentryTracesSampleRate:      getEnvFloat("SENTRY_TRACES_SAMPLE_RATE", 0.1),

		// Session
		SessionExpiry:          time.Duration(getEnvInt("SESSION_EXPIRY_MINS", 60)) * time.Minute,
		AdminSessionExpiryMins: getEnvInt("ADMIN_SESSION_EXPIRY_MINS", 60),
		SessionCookieName:      getEnv("SESSION_COOKIE_NAME", "terrareg_session"),
		SessionRefreshAge:      time.Duration(getEnvInt("SESSION_REFRESH_MINS", 25)) * time.Minute,

		// Namespaces
		VerifiedModuleNamespaces: getEnvSlice("VERIFIED_MODULE_NAMESPACES", []string{}),
		TrustedNamespaces:        getEnvSlice("TRUSTED_NAMESPACES", []string{}),

		ProviderSources: make(map[string]ProviderSourceConfig),
	}

	// Derived settings
	cfg.SAML2Enabled = cfg.SAML2IDPMetadataURL != ""
	cfg.OpenIDConnectEnabled = cfg.OpenIDConnectClientID != "" && cfg.OpenIDConnectIssuer != ""

	// Create directories if they don't exist
	if err := os.MkdirAll(cfg.DataDirectory, 0755); err != nil {
		return nil, fmt.Errorf("failed to create data directory: %w", err)
	}
	if err := os.MkdirAll(cfg.UploadDirectory, 0755); err != nil {
		return nil, fmt.Errorf("failed to create upload directory: %w", err)
	}

	return cfg, nil
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.DatabaseURL == "" {
		return fmt.Errorf("DATABASE_URL is required")
	}

	if c.PublicURL == "" {
		return fmt.Errorf("PUBLIC_URL is required")
	}

	if c.SecretKey == "" {
		return fmt.Errorf("SECRET_KEY is required. Generate with: python -c 'import secrets; print(secrets.token_hex())'")
	}

	// Ensure SECRET_KEY is at least 32 characters for AES-256
	if len(c.SecretKey) < 32 {
		return fmt.Errorf("SECRET_KEY must be at least 32 characters long for AES-256 encryption (current length: %d). Generate with: python -c 'import secrets; print(secrets.token_hex())'", len(c.SecretKey))
	}

	return nil
}

// Helper functions

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		value = strings.ToLower(value)
		return value == "true" || value == "1" || value == "yes"
	}
	return defaultValue
}

func getEnvFloat(key string, defaultValue float64) float64 {
	if value := os.Getenv(key); value != "" {
		if floatValue, err := strconv.ParseFloat(value, 64); err == nil {
			return floatValue
		}
	}
	return defaultValue
}

func getEnvSlice(key string, defaultValue []string) []string {
	if value := os.Getenv(key); value != "" {
		// Split by comma and trim whitespace
		values := strings.Split(value, ",")
		result := make([]string, 0, len(values))
		for _, v := range values {
			if trimmed := strings.TrimSpace(v); trimmed != "" {
				result = append(result, trimmed)
			}
		}
		return result
	}
	return defaultValue
}
