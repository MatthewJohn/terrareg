package config

import (
	"fmt"
	"time"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/config/model"
)

// InfrastructureConfig contains all infrastructure-related configuration
// This includes database connections, server settings, external services, etc.
type InfrastructureConfig struct {
	// Server settings
	ListenPort int    `env:"LISTEN_PORT"`
	PublicURL  string `env:"PUBLIC_URL"`
	DomainName string `env:"DOMAIN_NAME"`
	Debug      bool   `env:"DEBUG"`

	// Database settings
	DatabaseURL string `env:"DATABASE_URL"`

	// Storage settings
	DataDirectory   string `env:"DATA_DIRECTORY"`
	UploadDirectory string `env:"UPLOAD_DIRECTORY"`

	// Git provider settings
	GitProviderConfig string `env:"GIT_PROVIDER_CONFIG"`

	// Authentication settings (infrastructure concerns)
	// SAML
	SAML2IDPMetadataURL string `env:"SAML2_IDP_METADATA_URL"`
	SAML2IssuerEntityID string `env:"SAML2_ISSUER_ENTITY_ID"`

	// OpenID Connect
	OpenIDConnectClientID     string `env:"OPENID_CONNECT_CLIENT_ID"`
	OpenIDConnectClientSecret string `env:"OPENID_CONNECT_CLIENT_SECRET"`
	OpenIDConnectIssuer       string `env:"OPENID_CONNECT_ISSUER"`

	// Admin authentication
	AdminAuthenticationToken string   `env:"ADMIN_AUTHENTICATION_TOKEN"`
	UploadApiKeys            []string `env:"UPLOAD_API_KEYS"`
	PublishApiKeys           []string `env:"PUBLISH_API_KEYS"`
	SecretKey                string   `env:"SECRET_KEY"`

	// Analytics authentication (from Python ANALYTICS_AUTH_KEYS)
	AnalyticsAuthKeys []string `env:"ANALYTICS_AUTH_KEYS" envDefault:""`

	// Internal extraction authentication (from Python INTERNAL_EXTRACTION_ANALYTICS_TOKEN)
	InternalExtractionAnalyticsToken string `env:"INTERNAL_EXTRACTION_ANALYTICS_TOKEN"`

	// Feature flags (infrastructure concerns)
	AllowProviderHosting   bool `env:"ALLOW_PROVIDER_HOSTING"`
	AllowCustomGitProvider bool `env:"ALLOW_CUSTOM_GIT_PROVIDER"`
	EnableAccessControls   bool `env:"ENABLE_ACCESS_CONTROLS"`
	EnableSecurityScanning bool `env:"ENABLE_SECURITY_SCANNING"`

	// UI Customization (infrastructure assets)
	ApplicationName string `env:"APPLICATION_NAME"`
	LogoURL         string `env:"LOGO_URL"`
	SiteWarning     string `env:"SITE_WARNING"`

	// Session settings
	SessionExpiry          time.Duration `env:"SESSION_EXPIRY_MINS"`
	AdminSessionExpiryMins int           `env:"ADMIN_SESSION_EXPIRY_MINS"`
	SessionCookieName      string        `env:"SESSION_COOKIE_NAME"`
	SessionRefreshAge      time.Duration `env:"SESSION_REFRESH_MINS"`

	// External service settings
	InfracostAPIKey             string  `env:"INFRACOST_API_KEY"`
	InfracostPricingAPIEndpoint string  `env:"INFRACOST_PRICING_API_ENDPOINT"`
	SentryDSN                   string  `env:"SENTRY_DSN"`
	SentryTracesSampleRate      float64 `env:"SENTRY_TRACES_SAMPLE_RATE"`

	// Terraform OIDC settings
	TerraformOidcIdpSigningKeyPath    string `env:"TERRAFORM_OIDC_IDP_SIGNING_KEY_PATH"`
	TerraformOidcIdpSubjectIdHashSalt string `env:"TERRAFORM_OIDC_IDP_SUBJECT_ID_HASH_SALT"`
	TerraformOidcIdpSessionExpiry     int    `env:"TERRAFORM_OIDC_IDP_SESSION_EXPIRY"`

	// SSL/TLS Configuration
	SSLCertPrivateKey string `env:"SSL_CERT_PRIVATE_KEY"`
	SSLCertPublicKey  string `env:"SSL_CERT_PUBLIC_KEY"`

	// Complete SAML Configuration
	SAML2EntityID       string `env:"SAML2_ENTITY_ID"`
	SAML2PublicKey      string `env:"SAML2_PUBLIC_KEY"`
	SAML2PrivateKey     string `env:"SAML2_PRIVATE_KEY"`
	SAML2GroupAttribute string `env:"SAML2_GROUP_ATTRIBUTE"`
	SAML2Debug          bool   `env:"SAML2_DEBUG"`

	// Enhanced OpenID Connect Configuration
	OpenIDConnectScopes []string `env:"OPENID_CONNECT_SCOPES"`
	OpenIDConnectDebug  bool     `env:"OPENID_CONNECT_DEBUG"`

	// Access Control Configuration
	AllowUnauthenticatedAccess bool `env:"ALLOW_UNAUTHENTICATED_ACCESS"`

	// Git Provider Configuration
	GitCloneTimeout                int    `env:"GIT_CLONE_TIMEOUT"`
	UpstreamGitCredentialsUsername string `env:"UPSTREAM_GIT_CREDENTIALS_USERNAME"`
	UpstreamGitCredentialsPassword string `env:"UPSTREAM_GIT_CREDENTIALS_PASSWORD"`

	// Terraform Configuration
	TerraformBinaryPath    string `env:"TERRAFORM_BINARY_PATH"`
	TerraformDefaultVersion string `env:"TERRAFORM_DEFAULT_VERSION" envDefault:"1.5.7"`
	TerraformProduct       string `env:"TERRAFORM_PRODUCT" envDefault:"terraform"`
	TerraformArchiveMirror string `env:"TERRAFORM_ARCHIVE_MIRROR"`

	// Server Configuration
	ServerType       model.ServerType `env:"SERVER"`
	Threaded         bool             `env:"THREADED"`
	AllowedProviders []string         `env:"ALLOWED_PROVIDERS"`

	// HTTP Timeouts (in seconds)
	StandardRequestTimeoutSeconds int `env:"STANDARD_REQUEST_TIMEOUT_SECONDS" envDefault:"60"`   // 1 minute default
	ModuleIndexingTimeoutSeconds  int `env:"MODULE_INDEXING_TIMEOUT_SECONDS" envDefault:"1800"` // 30 minutes default

	// Terraform Processing Timeouts (in seconds)
	TerraformLockTimeoutSeconds   int `env:"TERRAFORM_LOCK_TIMEOUT_SECONDS" envDefault:"1800"`   // 30 minutes default

	// Terraform Presigned URL Configuration
	TerraformPresignedUrlSecret        string `env:"TERRAFORM_PRESIGNED_URL_SECRET"`
	TerraformPresignedUrlExpirySeconds int    `env:"TERRAFORM_PRESIGNED_URL_EXPIRY_SECONDS"`

	// Additional infrastructure settings
	// Note: Add any other infrastructure-specific settings here
	// such as Redis connections, message queue settings, etc.
}

// ProviderSourceConfig holds configuration for external provider sources
type ProviderSourceConfig struct {
	Type         string `env:"TYPE"`
	APIName      string `env:"API_NAME"`
	ClientID     string `env:"CLIENT_ID"`
	ClientSecret string `env:"CLIENT_SECRET"`
	LoginURL     string `env:"LOGIN_URL"`
	CallbackURL  string `env:"CALLBACK_URL"`
}

// IsDebug returns whether debug mode is enabled
func (c *InfrastructureConfig) IsDebug() bool {
	return c.Debug
}

// GetDatabaseConnectionURL returns the database connection URL
func (c *InfrastructureConfig) GetDatabaseConnectionURL() string {
	return c.DatabaseURL
}

// GetListenAddress returns the server listen address
func (c *InfrastructureConfig) GetListenAddress() string {
	return fmt.Sprintf(":%d", c.ListenPort)
}

// ValidateSAMLConfig validates SAML configuration
func (c *InfrastructureConfig) ValidateSAMLConfig() error {
	if c.SAML2IDPMetadataURL != "" {
		if c.SAML2EntityID == "" {
			return fmt.Errorf("SAML2_ENTITY_ID is required when SAML2_IDP_METADATA_URL is set")
		}

		if c.SAML2PublicKey == "" {
			return fmt.Errorf("SAML2_PUBLIC_KEY is required for SAML authentication")
		}
		if c.SAML2PrivateKey == "" {
			return fmt.Errorf("SAML2_PRIVATE_KEY is required for SAML authentication")
		}
	}

	return nil
}

// ValidateOIDCConfig validates OpenID Connect configuration
func (c *InfrastructureConfig) ValidateOIDCConfig() error {
	if c.OpenIDConnectIssuer != "" {
		if c.OpenIDConnectClientID == "" {
			return fmt.Errorf("OPENID_CONNECT_CLIENT_ID is required when OPENID_CONNECT_ISSUER is set")
		}
		if c.OpenIDConnectClientSecret == "" {
			return fmt.Errorf("OPENID_CONNECT_CLIENT_SECRET is required when OPENID_CONNECT_ISSUER is set")
		}
	}

	return nil
}

// ValidateAuthConfig validates all authentication configuration
func (c *InfrastructureConfig) ValidateAuthConfig() error {
	// Validate SAML configuration
	if err := c.ValidateSAMLConfig(); err != nil {
		return fmt.Errorf("SAML configuration error: %w", err)
	}

	// Validate OIDC configuration
	if err := c.ValidateOIDCConfig(); err != nil {
		return fmt.Errorf("OIDC configuration error: %w", err)
	}

	// Validate session configuration
	if c.SessionExpiry == 0 {
		return fmt.Errorf("SESSION_EXPIRY_MINS must be greater than 0")
	}

	if c.SessionCookieName == "" {
		return fmt.Errorf("SESSION_COOKIE_NAME cannot be empty")
	}

	// Validate secret key for session encryption
	if c.SecretKey == "" {
		return fmt.Errorf("SECRET_KEY is required for session security")
	}

	return nil
}

// ValidateTerraformOIDCConfig validates Terraform OIDC configuration
func (c *InfrastructureConfig) ValidateTerraformOIDCConfig() error {
	if c.TerraformOidcIdpSessionExpiry == 0 {
		return fmt.Errorf("TERRAFORM_OIDC_IDP_SESSION_EXPIRY must be greater than 0")
	}

	if c.TerraformOidcIdpSubjectIdHashSalt == "" {
		return fmt.Errorf("TERRAFORM_OIDC_IDP_SUBJECT_ID_HASH_SALT is required for Terraform OIDC security")
	}

	return nil
}

// ValidateServerConfig validates server configuration
func (c *InfrastructureConfig) ValidateServerConfig() error {
	if c.ListenPort == 0 {
		return fmt.Errorf("LISTEN_PORT must be greater than 0")
	}

	if c.ListenPort < 1 || c.ListenPort > 65535 {
		return fmt.Errorf("LISTEN_PORT must be between 1 and 65535")
	}

	if c.PublicURL == "" {
		return fmt.Errorf("PUBLIC_URL is required")
	}

	// Validate timeout configurations
	if c.StandardRequestTimeoutSeconds <= 0 {
		return fmt.Errorf("STANDARD_REQUEST_TIMEOUT_SECONDS must be greater than 0")
	}

	if c.ModuleIndexingTimeoutSeconds <= 0 {
		return fmt.Errorf("MODULE_INDEXING_TIMEOUT_SECONDS must be greater than 0")
	}

	if c.TerraformLockTimeoutSeconds <= 0 {
		return fmt.Errorf("TERRAFORM_LOCK_TIMEOUT_SECONDS must be greater than 0")
	}

	return nil
}

// ValidateDatabaseConfig validates database configuration
func (c *InfrastructureConfig) ValidateDatabaseConfig() error {
	if c.DatabaseURL == "" {
		return fmt.Errorf("DATABASE_URL is required")
	}

	// Additional database URL validation could be added here
	return nil
}

// ValidateAll performs comprehensive validation of all configuration
func (c *InfrastructureConfig) ValidateAll() error {
	var errors []string

	// Validate server configuration
	if err := c.ValidateServerConfig(); err != nil {
		errors = append(errors, err.Error())
	}

	// Validate database configuration
	if err := c.ValidateDatabaseConfig(); err != nil {
		errors = append(errors, err.Error())
	}

	// Validate authentication configuration
	if err := c.ValidateAuthConfig(); err != nil {
		errors = append(errors, err.Error())
	}

	// Validate Terraform OIDC configuration
	if err := c.ValidateTerraformOIDCConfig(); err != nil {
		errors = append(errors, err.Error())
	}

	if len(errors) > 0 {
		return fmt.Errorf("configuration validation failed: %v", errors)
	}

	return nil
}
