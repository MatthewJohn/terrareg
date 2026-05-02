# Provider Source Domain Architecture

## Overview

The Provider Source domain manages external provider source integrations (GitHub, GitLab, Bitbucket) for OAuth authentication and API operations. It implements a factory pattern with class registration to support multiple provider types dynamically.

---

## Core Functionality

The provider_source domain provides the following capabilities:

- **Provider Source Management** - Configure and manage external Git providers
- **OAuth Integration** - Handle OAuth flows for GitHub, GitLab, etc.
- **API Access** - Access provider APIs for repositories and organizations
- **Webhook Handling** - Process webhooks from external providers
- **Factory Pattern** - Register and instantiate provider sources dynamically
- **Authentication** - Use provider source authentication for repository access

---

## Domain Components

### Models

**Location**: `/internal/domain/provider_source/model/`

#### ProviderSource Model

```go
type ProviderSource struct {
    id           int
    name         string
    apiName      string
    type         ProviderSourceType
    config       *Config
}
```

**Key Fields**:
- `name` - Display name (e.g., "GitHub")
- `apiName` - API identifier (e.g., "github")
- `type` - Provider type (github, gitlab, bitbucket)
- `config` - Provider-specific configuration

#### Config Model

```go
type Config struct {
    BaseURL          string
    ApiURL           string
    ClientID         string
    ClientSecret     string
    PrivateKeyPath   string
    AppID            string
    LoginButtonText  string
}
```

#### Organization Model

```go
type Organization struct {
    ID       int
    Name     string
    AvatarURL string
}
```

#### Repository Model

```go
type Repository struct {
    ID              int
    Name            string
    FullName        string
    URL             string
    DefaultBranch   string
    IsPrivate       bool
}
```

#### ReleaseMetadata Model

```go
type ReleaseMetadata struct {
    Tag         string
    Name        string
    CreatedAt   time.Time
    PublishedAt time.Time
}
```

### ProviderSourceInstance Interface

```go
type ProviderSourceInstance interface {
    // Basic properties
    Name() string
    ApiName(ctx context.Context) (string, error)
    Type() model.ProviderSourceType

    // OAuth methods
    GetLoginRedirectURL(ctx context.Context) (string, error)
    GetUserAccessToken(ctx context.Context, code string) (string, error)
    GetUsername(ctx context.Context, accessToken string) (string, error)
    GetUserOrganizations(ctx context.Context, accessToken string) []string

    // API methods
    GetUserOrganizationsList(ctx context.Context, sessionID string) ([]*model.Organization, error)
    GetUserRepositories(ctx context.Context, sessionID string) ([]*model.Repository, error)
    RefreshNamespaceRepositories(ctx context.Context, namespace string) error
    PublishProviderFromRepository(ctx context.Context, repoID, categoryID int, namespace string) (*PublishProviderResult, error)
}
```

### ProviderSourceClass Interface

```go
type ProviderSourceClass interface {
    Type() model.ProviderSourceType
    GenerateDBConfigFromSourceConfig(config map[string]interface{}) (*model.Config, error)
    CreateInstance(name string, repo *repository.ProviderSourceRepository, db interface{}) (ProviderSourceInstance, error)
}
```

### Repository Interface

**Location**: `/internal/domain/provider_source/repository/provider_source_repository.go`

```go
type ProviderSourceRepository interface {
    Upsert(ctx context.Context, source *model.ProviderSource) error
    FindByName(ctx context.Context, name string) (*model.ProviderSource, error)
    FindByApiName(ctx context.Context, apiName string) (*model.ProviderSource, error)
    FindAll(ctx context.Context) ([]*model.ProviderSource, error)
}
```

### Factory

**Location**: `/internal/domain/provider_source/service/provider_source_factory.go`

```go
type ProviderSourceFactory struct {
    repo            repository.ProviderSourceRepository
    db              interface{}
    classMapping    map[model.ProviderSourceType]ProviderSourceClass
}
```

**Key Methods**:
- `RegisterProviderSourceClass(class)` - Register a provider class
- `GetProviderClasses()` - Get all registered classes
- `GetProviderSourceByName(ctx, name)` - Get provider by name
- `GetProviderSourceByApiName(ctx, apiName)` - Get provider by API name
- `GetAllProviderSources(ctx)` - Get all providers
- `InitialiseFromConfig(ctx, configJSON)` - Initialize from JSON config

---

## Dependencies

### Domain Dependencies

| Domain | Purpose |
|--------|---------|
| **git** | For Git provider configuration and operations |
| **config** | For provider source JSON configuration |
| **auth** | For OAuth integration |

### Infrastructure Dependencies

| Component | Purpose |
|-----------|---------|
| **Database** | Persistent storage for provider sources |
| **HTTP Client** - For API calls to providers |

### Domains That Use Provider Source

| Domain | Purpose |
|--------|---------|
| **module** | For GitHub/OAuth-based module publishing |
| **provider** | For automatic provider publishing from repositories |
| **repository** | For provider source authentication |

---

## Key Design Principles

1. **Factory Pattern** - Register and instantiate provider sources dynamically
2. **Interface-Based** - All providers implement ProviderSourceInstance
3. **Database Injection** - Factory stores DB reference and passes to instances
4. **Lazy Instantiation** - Provider instances created on-demand via `getProviderSourceImplementation()`
5. **Class Registration** - New provider types registered at startup

---

## Provider Source Types

| Type | Class | Implementation |
|------|-------|----------------|
| `github` | `GithubProviderSourceClass` | `GithubProviderSource` |
| `gitlab` | `GitlabProviderSourceClass` | `GitlabProviderSource` (TODO) |
| `bitbucket` | `BitbucketProviderSourceClass` | `BitbucketProviderSource` (TODO) |

---

## Factory Pattern

### Registration (during container init)

```go
factory := provider_source.NewProviderSourceFactory(repo, db)

githubClass := provider_source.NewGithubProviderSourceClass()
factory.RegisterProviderSourceClass(githubClass)
```

### Usage (in handlers/queries)

```go
providerSource, err := factory.GetProviderSourceByName(ctx, "GitHub")
if err != nil {
    return err
}

redirectURL, err := providerSource.GetLoginRedirectURL(ctx)
```

---

## Configuration

### JSON Configuration

```json
[
    {
        "name": "GitHub",
        "type": "github",
        "base_url": "https://github.com",
        "api_url": "https://api.github.com",
        "client_id": "xxx",
        "client_secret": "yyy",
        "private_key_path": "/path/to/key.pem",
        "app_id": "12345",
        "login_button_text": "Sign in with GitHub"
    }
]
```

### Environment Variable

```bash
PROVIDER_SOURCES='[{"name": "GitHub", "type": "github", ...}]'
```

---

## GitHub Provider Implementation

### GithubProviderSource

**Location**: `/internal/domain/provider_source/github_provider_source.go`

Implements `ProviderSourceInstance` interface:

```go
type GithubProviderSource struct {
    *BaseProviderSource
    config *model.Config
}

func (g *GithubProviderSource) GetLoginRedirectURL(ctx context.Context) (string, error)
func (g *GithubProviderSource) GetUserAccessToken(ctx context.Context, code string) (string, error)
func (g *GithubProviderSource) GetUserOrganizationsList(ctx context.Context, sessionID string) ([]*model.Organization, error)
func (g *GithubProviderSource) PublishProviderFromRepository(ctx context.Context, repoID, categoryID int, namespace string) (*PublishProviderResult, error)
```

### GithubProviderSourceClass

**Location**: `/internal/domain/provider_source/github_provider_source_class.go`

Implements `ProviderSourceClass` interface:

```go
type GithubProviderSourceClass struct{}

func (c *GithubProviderSourceClass) Type() model.ProviderSourceType
func (c *GithubProviderSourceClass) GenerateDBConfigFromSourceConfig(config map[string]interface{}) (*model.Config, error)
func (c *GithubProviderSourceClass) CreateInstance(name string, repo repository.ProviderSourceRepository, db interface{}) (ProviderSourceInstance, error)
```

---

## BaseProviderSource

**Location**: `/internal/domain/provider_source/base_provider_source.go`

Provides common functionality for all provider sources:

```go
type BaseProviderSource struct {
    name   string
    config *model.Config
    repo   repository.ProviderSourceRepository
    db     interface{}
}
```

Common methods shared across providers:
- `Name()` - Returns provider name
- `ApiName()` - Returns API name from config
- `Type()` - Returns provider type

---

## OAuth Flow

### GitHub OAuth Example

```
1. User clicks "Sign in with GitHub"
2. GetLoginRedirectURL() returns GitHub OAuth URL
3. User authorizes on GitHub
4. GitHub redirects back with code
5. GetUserAccessToken() exchanges code for access token
6. GetUsername() gets user's GitHub username
7. GetUserOrganizationsList() gets user's organizations
8. Create session with access token
```

---

## Repository Publishing

### Publishing from Repository

```go
result, err := providerSource.PublishProviderFromRepository(
    ctx,
    repoID,      // Repository ID
    categoryID,  // Provider category
    namespace,   // Target namespace
)
```

**Process**:
1. Fetch repository from provider API
2. Get latest release
3. Download provider binaries
4. Extract and validate
5. Create/update provider in Terrareg
6. Publish new version

---

## Usage Examples

### Getting a Provider Source

```go
factory := container.ProviderSourceFactory

providerSource, err := factory.GetProviderSourceByName(ctx, "GitHub")
if err != nil {
    return err
}

// Use provider source
redirectURL, err := providerSource.GetLoginRedirectURL(ctx)
```

### OAuth Login

```go
// Generate login URL
loginURL, err := githubProvider.GetLoginRedirectURL(ctx)

// After redirect, exchange code for token
accessToken, err := githubProvider.GetUserAccessToken(ctx, code)

// Get user info
username, err := githubProvider.GetUsername(ctx, accessToken)

// Get organizations
orgs, err := githubProvider.GetUserOrganizationsList(ctx, sessionID)
```

### Getting User Repositories

```go
repos, err := providerSource.GetUserRepositories(ctx, sessionID)
for _, repo := range repos {
    fmt.Printf("%s - %s\n", repo.FullName, repo.URL)
}
```

### Refreshing Namespace Repositories

```go
err := providerSource.RefreshNamespaceRepositories(ctx, "my-namespace")
```

---

## Integration with Container

```go
// In container initialization
factory := provider_source.NewProviderSourceFactory(repo, db)

// Register provider classes
githubClass := provider_source.NewGithubProviderSourceClass()
factory.RegisterProviderSourceClass(githubClass)

// Initialize from config
configJSON := infraConfig.ProviderSourcesJSON
err := factory.InitialiseFromConfig(ctx, configJSON)

// Set database for provider instances
factory.SetDatabase(db.DB)
```

---

## References

- [`/internal/domain/provider_source/model/`](./model/) - Provider source models
- [`/internal/domain/provider_source/service/provider_source_factory.go`](./service/provider_source_factory.go) - Factory implementation
- [`/internal/domain/provider_source/github_provider_source.go`](./github_provider_source.go) - GitHub implementation
- [`/internal/domain/provider_source/github_provider_source_class.go`](./github_provider_source_class.go) - GitHub class
- [`/CLAUDE.md`](../../../../CLAUDE.md) - Provider Source Architecture section
