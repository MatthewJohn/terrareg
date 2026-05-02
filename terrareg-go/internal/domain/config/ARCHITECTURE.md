# Config Domain Architecture

## Overview

The Config domain provides centralized configuration management for the Terrareg application. It implements a three-tier configuration system separating business logic configuration from technical infrastructure settings, with a read-only view for presentation layer consumption.

---

## Core Functionality

The config domain provides the following capabilities:

- **Environment Variable Loading** - Loads all configuration from environment variables
- **Configuration Validation** - Validates required settings and detects conflicts
- **Three-Tier Separation** - Business, infrastructure, and UI configuration
- **Type Conversion** - Parses strings to appropriate types (int, bool, duration, etc.)
- **Default Values** - Provides sensible defaults for all configuration options

---

## Domain Components

### Models

**Location**: `/internal/domain/config/model/`

#### Three-Tier Configuration System

```go
// 1. DomainConfig - Business logic, UI settings, feature flags
type DomainConfig struct {
    // Feature flags
    AllowModuleHosting              ModuleHostingMode
    UploadAPIKeysEnabled            bool
    PublishAPIKeysEnabled           bool
    SecretKeySet                    bool

    // Namespace settings
    TrustedNamespaces               []string
    VerifiedModuleNamespaces        []string

    // UI configuration
    TrustedNamespaceLabel           string
    ContributedNamespaceLabel       string
    VerifiedModuleLabel             string

    // Analytics configuration
    AnalyticsTokenPhrase            string
    AnalyticsTokenDescription       string
    ExampleAnalyticsToken           string
    DisableAnalytics                bool

    // Module processing configuration
    AutoPublishModuleVersions       bool
    ModuleVersionReindexMode        ModuleVersionReindexMode
    RequiredModuleMetadataAttributes []string

    // Provider sources
    ProviderSources                 map[string]ProviderSourceConfig

    // Authentication status (computed by auth services)
    OpenIDConnectEnabled            bool
    SAMLEnabled                     bool
    AdminLoginEnabled               bool
}

// 2. InfrastructureConfig - Technical: DB, storage, secrets
// Location: /internal/infrastructure/config/model/
type InfrastructureConfig struct {
    // Server settings
    ListenPort                      int
    PublicURL                       string
    DomainName                      string
    Debug                           bool

    // Database settings
    DatabaseURL                     string

    // Storage settings
    DataDirectory                   string
    UploadDirectory                 string

    // Authentication settings
    SAML2IDPMetadataURL             string
    OpenIDConnectClientID           string
    AdminAuthenticationToken        string
    SecretKey                       string

    // Session settings
    SessionExpiry                   time.Duration
    SessionCookieName               string

    // External service settings
    InfracostAPIKey                 string
    SentryDSN                       string
}

// 3. UIConfig - Read-only view for presentation
type UIConfig struct {
    // Derived from DomainConfig and InfrastructureConfig
    // Contains only what the UI needs to display
}
```

#### Configuration Enums

| Enum | Values |
|------|--------|
| `ModuleHostingMode` | `Allow`, `Disallow`, `Enforce` |
| `ModuleVersionReindexMode` | `Legacy`, `New` |
| `Product` | `Terraform`, `TerraformEnterprise` |
| `ServerType` | `Builtin`, `Gunicorn` |
| `DefaultUiInputOutputView` | `Table`, `Expanded` |

### Service

**Location**: `/internal/domain/config/service/configuration_service.go`

The `ConfigurationService` is the single source of truth for all configuration:

```go
type ConfigurationService struct {
    envLoader *config.EnvironmentLoader
    validator *config.ConfigValidator
    opts      ConfigurationServiceOptions
}

// LoadConfiguration loads all configuration from environment
func (s *ConfigurationService) LoadConfiguration() (
    *model.DomainConfig,
    *config.InfrastructureConfig,
    error,
)
```

**Key Methods**:
- `LoadConfiguration()` - Main entry point for loading all config
- `buildDomainConfig()` - Constructs business configuration
- `buildInfrastructureConfig()` - Constructs technical configuration
- Various `parse*()` helper methods for type conversion

---

## Dependencies

### Domain Dependencies

| Domain | Purpose |
|--------|---------|
| **auth** | To determine if auth methods are configured (OIDC, SAML) |
| **url** | For URL building and public URL detection |

### Infrastructure Dependencies

| Component | Purpose |
|-----------|---------|
| **Environment Variables** - All configuration comes from environment |
| **Version Reader** - For application version info |

### Domains That Depend on Config

All domains depend on configuration for their behavior:
- **module** - Module processing settings
- **auth** - Authentication configuration
- **storage** - Storage paths and configuration
- **git** - Git provider configuration
- **provider_source** - Provider source JSON configuration

---

## Key Design Principles

1. **Single Source of Truth** - ConfigurationService is the only place configuration is loaded
2. **Environment-First** - All configuration comes from environment variables
3. **Three-Tier Separation** - Business vs. Infrastructure vs. UI concerns
4. **Validation Upfront** - All configuration validated at startup
5. **Sensible Defaults** - Every setting has a documented default
6. **Immutable After Load** - Configuration loaded once at startup

---

## Configuration Loading Process

### Phase 1: Environment Variable Collection

```go
// Load all environment variables once
rawConfig := s.envLoader.LoadAllEnvironmentVariables()
```

### Phase 2: Validation

```go
// Validate configuration
if err := s.validator.Validate(rawConfig); err != nil {
    return nil, nil, fmt.Errorf("configuration validation failed: %w", err)
}
```

### Phase 3: Build Infrastructure Config

```go
infrastructureConfig := s.buildInfrastructureConfig(rawConfig)
```

### Phase 4: Build Domain Config

```go
domainConfig := s.buildDomainConfig(rawConfig, infrastructureConfig)
```

---

## Configuration Hierarchy

```
Environment Variables
         ↓
   ConfigurationService
         ↓
    ┌────┴────┐
    ↓         ↓
DomainConfig  InfrastructureConfig
    ↓         ↓
    └────┬────┘
         ↓
     UIConfig (read-only)
```

---

## Type Conversion Helpers

The ConfigurationService provides helper methods for converting environment variable strings to appropriate types:

| Helper | Purpose | Default |
|--------|---------|---------|
| `parseBool()` | Parse "true"/"false" to bool | `false` |
| `parseInt()` | Parse string to int | `0` |
| `parseFloat()` | Parse string to float64 | `0.0` |
| `parseDuration()` | Parse minutes to time.Duration | `0` |
| `parseStringSlice()` | Parse comma-separated string to []string | `[]` |
| `parseModuleHostingMode()` | Parse module hosting mode enum | `Allow` |
| `parseModuleVersionReindexMode()` | Parse reindex mode enum | `Legacy` |
| `parseProduct()` | Parse product enum | `Terraform` |
| `parseServerType()` | Parse server type enum | `Builtin` |

---

## Key Configuration Categories

### Feature Flags

Control major feature behavior:

```bash
ALLOW_MODULE_HOSTING=true|false|enforce
ALLOW_PROVIDER_HOSTING=true
UPLOAD_API_KEYS=key1,key2
PUBLISH_API_KEYS=key1,key2
ENABLE_ACCESS_CONTROLS=false
ENABLE_SECURITY_SCANNING=true
```

### Namespace Settings

```bash
TRUSTED_NAMESPACES=ns1,ns2,ns3
VERIFIED_MODULE_NAMESPACES=verified1,verified2
TRUSTED_NAMESPACE_LABEL="Trusted"
CONTRIBUTED_NAMESPACE_LABEL="Contributed"
```

### Module Processing

```bash
AUTO_PUBLISH_MODULE_VERSIONS=true
MODULE_VERSION_REINDEX_MODE=legacy|new
AUTO_CREATE_NAMESPACE=true
AUTO_CREATE_MODULE_PROVIDER=true
REQUIRED_MODULE_METADATA_ATTRIBUTES=attr1,attr2
```

### Analytics

```bash
ANALYTICS_TOKEN_PHRASE="analytics token"
EXAMPLE_ANALYTICS_TOKEN="my-tf-application"
DISABLE_ANALYTICS=false
ANALYTICS_AUTH_KEYS=key1,key2
ALLOW_UNIDENTIFIED_DOWNLOADS=false
```

### Authentication

```bash
ADMIN_AUTHENTICATION_TOKEN=admin-token
SECRET_KEY=minimum-32-characters-long
SESSION_EXPIRY_MINS=1440
SESSION_COOKIE_NAME=terrareg_session
```

### Storage

```bash
DATA_DIRECTORY=./data
UPLOAD_DIRECTORY=./data/upload
```

### Provider Sources

```bash
PROVIDER_SOURCES='[{"name": "GitHub", "type": "github", ...}]'
PROVIDER_CATEGORIES='[{"id": 1, "name": "Example", "slug": "example", "user-selectable": true}]'
```

---

## Configuration Validation

The `ConfigValidator` performs validation checks:

1. **Required Fields** - Ensures critical settings are present
2. **Type Validation** - Validates enums and numeric ranges
3. **Conflict Detection** - Detects incompatible settings
4. **Path Validation** - Validates file paths exist when required

---

## Authentication Status Computation

The config domain delegates authentication method status determination to the auth services:

```go
// In buildDomainConfig - auth services are the single source of truth
OpenIDConnectEnabled: authservice.IsOIDCConfigured(infrastructureConfig),
SAMLEnabled:          authservice.IsSAMLConfigured(infrastructureConfig),
AdminLoginEnabled:    rawConfig["ADMIN_AUTHENTICATION_TOKEN"] != "",
```

This ensures authentication status is determined by the same logic that performs authentication.

---

## Usage Examples

### Loading Configuration

```go
configService := configService.NewConfigurationService(
    configService.ConfigurationServiceOptions{
        AllowHotReload: false,
    },
    versionReader,
)

domainConfig, infraConfig, err := configService.LoadConfiguration()
if err != nil {
    log.Fatal().Err(err).Msg("Failed to load configuration")
}
```

### Accessing Configuration in Services

```go
type ModuleService struct {
    config *model.DomainConfig  // Only business logic config
    repo   ModuleRepository
}

type SessionService struct {
    config *config.InfrastructureConfig  // Only technical config
    repo   SessionRepository
}
```

### Using Configuration Queries

In handlers, use queries to get UI configuration:

```go
type ConfigHandler struct {
    getConfigQuery *configQuery.GetConfigQuery
}

func (h *ConfigHandler) HandleGetConfig(w http.ResponseWriter, r *http.Request) {
    uiConfig, err := h.getConfigQuery.Execute(r.Context())
    // Returns UIConfig - read-only view for presentation
}
```

---

## Configuration Migration Pattern

When adding new configuration fields:

1. **Add to struct** with env/envDefault tags:
```go
type InfrastructureConfig struct {
    NewSetting string `env:"NEW_SETTING" envDefault:"default-value"`
}
```

2. **Add to buildInfrastructureConfig()**:
```go
NewSetting: s.getEnvStringWithDefault(rawConfig, "NEW_SETTING", "default-value"),
```

3. **Add validation** if required:
```go
if infraConfig.NewSetting == "" {
    return errors.New("NEW_SETTING cannot be empty")
}
```

---

## References

For complete configuration reference, see:
- [`/internal/infrastructure/config/model/config.go`](../../infrastructure/config/model/config.go) - InfrastructureConfig model
- [`/internal/domain/config/model/config.go`](./model/config.go) - DomainConfig model
- [`CONFIGURATION_MIGRATION_COMPLETE.md`](../../docs/CONFIGURATION_MIGRATION_COMPLETE.md) - Migration documentation
