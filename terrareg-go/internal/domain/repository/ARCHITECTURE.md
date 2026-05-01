# Repository Domain Architecture

## Overview

The Repository domain manages Git repository links to providers. This domain represents the Git repositories that are associated with Terraform providers for automatic publishing and webhook integration.

**Note**: This is NOT the repository pattern (data access) - it's the domain of Git repositories linked to providers.

---

## Core Functionality

The repository domain provides the following capabilities:

- **Repository Linking** - Link Git repositories to providers
- **Repository Metadata** - Store repository URL and clone information
- **Webhook Integration** - Enable webhook-based publishing from repositories
- **Provider Source Auth** - Use provider source authentication for repository access

---

## Domain Components

### Models

**Location**: `/internal/domain/repository/model/repository.go`

#### Repository Model

```go
type Repository struct {
    id              int
    namespaceID     int
    name            string
    url             string
    cloneURL        string
    providerSourceID *int
    createdAt       time.Time
    updatedAt       time.Time
}
```

**Key Fields**:
- `namespaceID` - Links repository to a namespace
- `name` - Repository name
- `url` - Repository web URL
- `cloneURL` - URL for cloning the repository
- `providerSourceID` - Optional link to a provider source for authentication

### Repository Interface

**Location**: `/internal/domain/repository/repository/repository.go`

```go
type RepositoryRepository interface {
    Save(ctx context.Context, repo *Repository) error
    FindByID(ctx context.Context, id int) (*Repository, error)
    FindByNamespaceAndName(ctx context.Context, namespace, name string) (*Repository, error)
    FindByProviderID(ctx context.Context, providerID int) (*Repository, error)
    Delete(ctx context.Context, id int) error
}
```

---

## Dependencies

### Domain Dependencies

| Domain | Purpose |
|--------|---------|
| **module** | For namespace repository |
| **provider_source** | For provider source authentication |
| **provider** | For provider association |

### Infrastructure Dependencies

| Component | Purpose |
|-----------|---------|
| **Database** | Persistent storage for repository metadata |

---

## Key Design Principles

1. **Namespace Scoped** - Repositories belong to namespaces
2. **Provider Association** - Repositories linked to providers for automatic publishing
3. **Provider Source Integration** - Can use provider source authentication
4. **Git Native** - Designed for Git repository operations

---

## Repository Lifecycle

### Creation Flow

```
1. Validate namespace exists
2. Validate repository URL
3. Generate clone URL from web URL
4. Create Repository entity
5. Link to provider source (optional)
6. Save to repository
```

### Provider Association

```
1. Create or update provider
2. Link repository to provider
3. Enable webhook-based publishing
4. Configure automatic publishing settings
```

---

## Clone URL Generation

Clone URLs are generated from repository web URLs:

| Provider | Web URL | Clone URL |
|----------|---------|-----------|
| GitHub | `https://github.com/org/repo` | `https://github.com/org/repo.git` |
| GitLab | `https://gitlab.com/org/repo` | `https://gitlab.com/org/repo.git` |
| Bitbucket | `https://bitbucket.org/org/repo` | `https://bitbucket.org/org/repo.git` |

---

## Provider Source Authentication

Repositories can use provider source authentication:

```go
repo.SetProviderSourceID(providerSourceID)
```

This enables:
- Private repository access
- Webhook authentication
- Automatic publishing without credentials

---

## Usage Examples

### Creating a Repository

```go
repo := &repository.Repository{
    NamespaceID:      1,
    Name:            "my-terraform-provider",
    URL:             "https://github.com/myorg/my-terraform-provider",
    CloneURL:        "https://github.com/myorg/my-terraform-provider.git",
    ProviderSourceID: &providerSourceID,
}

err := repoRepo.Save(ctx, repo)
```

### Linking to a Provider

```go
// When creating a provider
req := providerService.CreateProviderRequest{
    // ...
    RepositoryID: &repoID,
}

provider, err := providerService.CreateProvider(ctx, req)
```

### Finding by Provider

```go
repo, err := repoRepo.FindByProviderID(ctx, providerID)
if err != nil {
    // Handle not found
}

cloneURL := repo.CloneURL()
```

---

## Webhook Integration

Repositories with linked provider sources support webhook-based publishing:

1. **Webhook Received** - Provider source receives webhook
2. **Repository Identified** - Match webhook to repository
3. **Provider Found** - Find linked provider
4. **Version Published** - Publish new version from release

---

## Repository vs Repository Pattern

This domain is about **Git repositories**, not the **repository pattern** for data access:

| Repository Domain | Repository Pattern |
|-------------------|-------------------|
| Git repository links to providers | Data access abstraction |
| `Repository` entity | `*Repository` interface |
| `/internal/domain/repository/` | `/internal/domain/*/repository/` |

---

## References

- [`/internal/domain/repository/model/`](./model/) - Repository models
- [`/internal/domain/provider_source/`](../provider_source/) - Provider source authentication
- [`/internal/domain/provider/`](../provider/) - Provider domain
- [`/docs/WEBHOOK_INTEGRATION.md`](../../docs/WEBHOOK_INTEGRATION.md) - Webhook documentation
