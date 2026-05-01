# Git Domain Architecture

## Overview

The Git domain manages Git integration for module publishing, webhook handling, and repository operations. It provides URL template handling, Git tag format conversion, and repository URL building.

---

## Core Functionality

The git domain provides the following capabilities:

- **URL Template Handling** - Build repository URLs for clone, browse, and archive operations
- **Git Tag Format Conversion** - Convert module versions to git tags and vice versa
- **Webhook Processing** - Handle Git provider webhooks for automatic publishing
- **Repository URL Building** - Construct consistent URLs for Git operations

---

## Domain Components

### Models

**Location**: `/internal/domain/git/model/`

#### URLTemplate Model

**File**: `url_template.go`

Templates for module repository URLs with placeholders:
- `{namespace}` - Namespace name
- `{module}` - Module name
- `{provider}` - Provider name
- `{git_tag}` - Git tag (for archive URLs)
- `{version}` - Module version (for archive URLs)
- `{tag}` - Git tag (for browse URLs)
- `{path}` - Repository path (for browse URLs)

```go
type URLTemplate struct {
    cloneURLTemplate   string
    browseURLTemplate  string
    archiveURLTemplate string
}
```

#### GitUrlTemplateValidator Model

**File**: `git_url_template_validator.go`

Templates for Git Provider URLs:
- `{namespace}` - Namespace name
- `{module}` - Module name
- `{provider}` - Provider name
- `{path}` - Repository path
- `{tag}` - Git tag
- `{tag_uri_encoded}` - URL-encoded git tag

#### GitTagFormat Model

**File**: `git_tag_format.go`

Converts module version to git tag and vice versa:

```go
type GitTagFormat struct {
    Raw     string
    IsValid bool
}
```

**Placeholders**:
- `{version}` - Full semantic version (e.g., "1.2.3")
- `{major}` - Major version component (e.g., "1")
- `{minor}` - Minor version component (e.g., "2")
- `{patch}` - Patch version component (e.g., "3")

**Examples**:

| Format | Version | Git Tag |
|--------|---------|---------|
| `v{version}` | 1.2.3 | v1.2.3 |
| `releases/v{major}.{minor}` | 1.2.3 | releases/v1.2 |
| `{major}.{patch}` | 1.2.3 | 1.3 (minor defaults to 0) |

### Repository Interface

**Location**: `/internal/domain/git/repository/git_provider_repository.go`

```go
type GitProviderRepository interface {
    Save(ctx context.Context, provider *model.GitProvider) error
    FindByID(ctx context.Context, id int) (*model.GitProvider, error)
    FindAll(ctx context.Context) ([]*model.GitProvider, error)
}
```

### Services

**Location**: `/internal/domain/git/service/`

#### GitURLBuilderService

Builds clone, browse, and archive URLs from templates:

```go
type GitURLBuilderService struct {
    urlTemplate *model.URLTemplate
}

func (s *GitURLBuilderService) BuildCloneURL(namespace, module, provider, tag string) string
func (s *GitURLBuilderService) BuildBrowseURL(namespace, module, provider, tag, path string) string
func (s *GitURLBuilderService) BuildArchiveURL(namespace, module, provider, version string) string
```

#### GitProviderFactory

Creates Git provider instances:

```go
type GitProviderFactory struct{}

func (f *GitProviderFactory) CreateGitProvider(config *model.GitProviderConfig) (GitProvider, error)
```

#### WebhookHandler

Handles Git provider webhooks:

```go
type WebhookHandler struct {
    gitService     *GitService
    moduleService  module.ModuleService
}

func (h *WebhookHandler) HandleWebhook(ctx context.Context, provider string, payload []byte) error
```

---

## Dependencies

### Domain Dependencies

| Domain | Purpose |
|--------|---------|
| **config** | For Git provider configuration |
| **storage** | For clone operations and temporary storage |
| **module** | For webhook-based module publishing |

### Infrastructure Dependencies

| Component | Purpose |
|-----------|---------|
| **Git** - Git binary for clone operations |
| **HTTP Client** - For Git provider API calls |

### Domains That Use Git

| Domain | Purpose |
|--------|---------|
| **module** - For Git-based module imports |
| **provider_source** - For provider source integration |

---

## Key Design Principles

1. **Template-Based** - URLs built from configurable templates
2. **Git Provider Agnostic** - Works with GitHub, GitLab, Bitbucket
3. **Version to Tag Mapping** - Flexible version-to-git-tag conversion
4. **Placeholder Validation** - Validates template placeholders
5. **Parse-Once** - Validation happens during parsing

---

## URL Template Types

### Git Provider URL Templates

**Model**: `GitUrlTemplateValidator`
**File**: `git_url_template_validator.go`

For Git Provider URLs (clone, browse, archive):
- Clone URL template
- Browse URL template
- Archive URL template

### Module Repository URL Templates

**Model**: `URLTemplate`
**File**: `url_template.go`

For module repository URLs:
- Clone URL
- Browse URL
- Archive URL

---

## Git Tag Format Operations

### 1. Version to Git Tag

Converts module version to git tag for cloning:

```go
format := gitTagFormat.Parse("v{version}")
tag := format.VersionToGitTag("1.2.3")
// Returns: "v1.2.3"
```

### 2. Git Tag to Version

Parses git tag to extract version for indexing:

```go
format := gitTagFormat.Parse("v{version}")
version, ok := format.GitTagToVersion("v1.2.3")
// Returns: "1.2.3", true
```

### 3. Validation

Validates format contains valid placeholders:

```go
format := gitTagFormat.Parse("v{version}")
if !format.IsValid {
    // Handle invalid format
}
```

---

## Template Validation Strategy

### URL Template Validation

1. **Parse** - Replace known placeholders with test values
2. **Validate** - Check for:
   - Unknown placeholders (remaining `{` `}` patterns)
   - URL validity (scheme, host, path)
   - Required placeholders (based on context)

### Git Tag Format Validation

1. **Parse** - Replace known placeholders with test values
2. **Validate** - Check for:
   - Unknown placeholders (remaining `{` `}` patterns)
   - At least one required placeholder present (`{version}`, `{major}`, `{minor}`, or `{patch}`)

---

## Usage Examples

### Building Clone URLs

```go
urlBuilder := git.NewGitURLBuilderService(urlTemplate)

cloneURL := urlBuilder.BuildCloneURL("aws", "vpc", "aws", "v1.0.0")
// Returns: "https://github.com/namespace/aws-vpc.git"
```

### Converting Version to Git Tag

```go
format := gitTagFormat.Parse("v{version}")
tag := format.VersionToGitTag("1.2.3")
// Returns: "v1.2.3"

// Custom format
format = gitTagFormat.Parse("releases/v{major}.{minor}")
tag = format.VersionToGitTag("1.2.3")
// Returns: "releases/v1.2"
```

### Parsing Git Tag to Version

```go
format := gitTagFormat.Parse("v{version}")
version, ok := format.GitTagToVersion("v1.2.3")
// Returns: "1.2.3", true
```

---

## Webhook Integration

The git domain handles webhooks from Git providers:

1. **Receive Webhook** - Provider (GitHub, GitLab, etc.) sends webhook
2. **Validate Signature** - Verify webhook signature
3. **Parse Payload** - Extract repository and tag information
4. **Trigger Import** - Import module version from repository

For detailed webhook documentation, see:
- [`/docs/WEBHOOK_INTEGRATION.md`](../../docs/WEBHOOK_INTEGRATION.md)
- [`/docs/WEBHOOK_QUICK_START.md`](../../docs/WEBHOOK_QUICK_START.md)

---

## References

- [`/internal/domain/git/model/`](./model/) - Git models and templates
- [`/internal/domain/git/service/`](./service/) - Git services
- [`/docs/WEBHOOK_INTEGRATION.md`](../../docs/WEBHOOK_INTEGRATION.md) - Webhook documentation
