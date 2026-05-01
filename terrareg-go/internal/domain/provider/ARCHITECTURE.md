# Provider Domain Architecture

## Overview

The Provider domain manages Terraform provider registry functionality including binary management, extraction, and publishing. It handles the complete lifecycle of Terraform providers from creation to version management.

---

## Core Functionality

The provider domain provides the following capabilities:

- **Provider CRUD** - Create, read, update, and delete providers
- **Version Management** - Manage provider versions with protocol versions
- **Binary Extraction** - Extract provider binaries from uploaded files
- **Publishing** - Publish provider versions with GPG signing
- **Category Management** - Organize providers into categories
- **Repository Integration** - Link providers to Git repositories

---

## Domain Components

### Models

**Location**: `/internal/domain/provider/model/`

#### Provider Aggregate Root

```go
type Provider struct {
    id                    int
    namespace            *model.Namespace  // Namespace entity reference
    namespaceID           int               // Denormalized for persistence
    name                  string
    description           *string
    tier                  string
    categoryID            *int
    repositoryID          *int
    useProviderSourceAuth bool
    versions              []*ProviderVersion
    gpgKeys               []*GPGKey
    latestVersionID       *int
}
```

**Key Methods**:
- `Namespace() *model.Namespace` - Returns namespace entity
- `Name() string` - Provider name
- `VersionID(namespace, version string)` - Returns formatted ID (temporary, until Namespace is always populated)

#### ProviderVersion Entity

```go
type ProviderVersion struct {
    id               int
    provider         *Provider  // Parent provider reference
    providerID       int
    version          string
    gitTag           *string
    beta             bool
    publishedAt      *time.Time
    gpgKeyID         int
    protocolVersions []string
    binaries         []*ProviderBinary
}
```

**Key Methods**:
- `FormattedID() string` - Returns `"namespace/provider/version"` (Python: ProviderVersion.id property)
- `Provider() *Provider` - Returns parent provider
- `Version() string` - Version string

#### ProviderCategory Model

```go
type ProviderCategory struct {
    id              int
    name            string
    slug            string
    userSelectable  bool
}
```

#### ProviderExtraction Model

```go
type ProviderExtraction struct {
    archivePath     string
    extractPath     string
    platform        string
    architecture    string
    version         string
    executablePath  string
}
```

### Repository Interfaces

**Location**: `/internal/domain/provider/repository/`

```go
type ProviderRepository interface {
    Save(ctx context.Context, provider *Provider) error
    FindByID(ctx context.Context, id int) (*Provider, error)
    FindByNamespaceAndName(ctx context.Context, namespace, name string) (*Provider, error)
    FindAll(ctx context.Context, offset, limit int) ([]*Provider, map[int]string, map[int]VersionData, int, error)
    Search(ctx context.Context, query ProviderSearchQuery) (*ProviderSearchResult, error)

    // Version operations
    SaveVersion(ctx context.Context, version *ProviderVersion) error
    FindVersionByProviderAndVersion(ctx context.Context, providerID int, version string) (*ProviderVersion, error)
    FindVersionsByProvider(ctx context.Context, providerID int) ([]*ProviderVersion, error)
    DeleteVersion(ctx context.Context, versionID int) error
    SetLatestVersion(ctx context.Context, providerID, versionID int) error
}

type ProviderCategoryRepository interface {
    Save(ctx context.Context, category *ProviderCategory) error
    FindByID(ctx context.Context, id int) (*ProviderCategory, error)
    FindAll(ctx context.Context) ([]*ProviderCategory, error)
    FindBySlug(ctx context.Context, slug string) (*ProviderCategory, error)
}
```

#### Repository "Not Found" Pattern

**Important**: The repository layer returns `(nil, nil)` for "no results", not an error:

```go
// ✅ CORRECT: Repository returns (nil, nil) for "not found"
if errors.Is(err, gorm.ErrRecordNotFound) {
    return nil, nil  // "no results", not an error
}
```

**Query layer** converts to descriptive error:

```go
// Query layer handles (nil, nil) → descriptive error
p, err := repo.FindByNamespaceAndName(ctx, namespace, name)
if err != nil {
    return nil, fmt.Errorf("failed to get provider: %w", err)
}
if p == nil {
    return nil, fmt.Errorf("provider %s/%s not found", namespace, name)
}
```

### Services

**Location**: `/internal/domain/provider/service/`

#### ProviderService

```go
type ProviderService struct {
    providerRepo repository.ProviderRepository
}
```

**Key Methods**:
- `CreateProvider(ctx, req)` - Create a new provider
- `UpdateProvider(ctx, providerID, req)` - Update provider details
- `DeleteProvider(ctx, providerID)` - Delete a provider
- `GetProvider(ctx, providerID)` - Get a provider by ID
- `GetProviderByName(ctx, namespace, name)` - Get a provider by namespace and name
- `ListProviders(ctx, offset, limit)` - List all providers with pagination
- `SearchProviders(ctx, query, offset, limit)` - Search providers

#### ProviderExtractorService

Handles binary extraction from uploaded archives:

```go
type ProviderExtractorService struct {
    storageService storage.StorageService
    gpgKeyService  gpgkey.GPGKeyServiceInterface
}
```

**Key Methods**:
- `ExtractProviderBinaries(ctx, archivePath, extractPath)` - Extract binaries
- `ValidateBinary(ctx, binaryPath)` - Validate provider binary
- `GetBinaryPlatform(binaryPath)` - Determine platform and architecture

#### ProviderPublisherService

Handles provider version publishing:

```go
type ProviderPublisherService struct {
    providerRepo    repository.ProviderRepository
    storageService  storage.StorageService
}
```

**Key Methods**:
- `PublishVersion(ctx, providerID, req)` - Publish a new version
- `SignBinary(ctx, binaryPath, gpgKeyID)` - Sign binary with GPG key
- `GenerateChecksums(ctx, binaryPath)` - Generate SHA256 checksums

---

## Dependencies

### Domain Dependencies

| Domain | Purpose |
|--------|---------|
| **module** | For namespace repository |
| **gpgkey** | For GPG key signing |
| **storage** | For binary storage |
| **audit** | For audit logging |

### Infrastructure Dependencies

| Component | Purpose |
|-----------|---------|
| **Database** | Persistent storage for providers |
| **File System** | Temporary storage during extraction |
| **GPG** | Binary signing |

---

## Key Design Principles

1. **Aggregate Pattern** - Provider is aggregate root, versions are entities
2. **Tier System** - Providers have tiers (community, partner, official)
3. **Protocol Versions** - Each version specifies supported protocol versions
4. **GPG Signing** - Provider versions can be signed with GPG keys
5. **Platform Support** - Multi-platform binary support (linux, windows, darwin)

---

## Provider Lifecycle

### Creation Flow

```
1. Validate namespace exists
2. Validate provider name
3. Check for duplicate (namespace + name)
4. Create Provider entity
5. Set optional fields (description, tier, category, repository)
6. Save to repository
```

### Version Publishing Flow

```
1. Upload provider archive
2. Extract binaries from archive
3. Validate each binary
4. Generate checksums
5. Sign with GPG key (optional)
6. Store binaries in storage
7. Create ProviderVersion entity
8. Set as latest version
9. Save to repository
```

---

## Provider Tiers

| Tier | Description |
|------|-------------|
| `community` | Community-maintained providers |
| `partner` | Partner-maintained providers |
| `official` | Officially maintained providers |
| `example` | Example providers |

---

## Platform and Architecture

**Platforms**: `linux`, `windows`, `darwin`, `freebsd`

**Architectures**: `amd64`, `arm`, `arm64`, `386`

**Binary Naming Convention**:
```
{provider}_{version}_{platform}_{architecture}.zip
```

Example: `terraform-provider-aws_5.0.0_linux_amd64.zip`

---

## Usage Examples

### Creating a Provider

```go
req := providerService.CreateProviderRequest{
    NamespaceID:           1,
    Name:                  "aws",
    Description:           &desc,
    Tier:                  "official",
    CategoryID:            &categoryID,
    RepositoryID:          &repoID,
    UseProviderSourceAuth:  true,
}

provider, err := providerService.CreateProvider(ctx, req)
```

### Publishing a Version

```go
req := providerService.PublishVersionRequest{
    Version:          "5.0.0",
    GitTag:           &gitTag,
    ProtocolVersions: []string{"5.0"},
    IsBeta:           false,
}

version, err := providerService.PublishVersion(ctx, providerID, req)
```

### Listing Providers

```go
providers, namespaceNames, versionData, total, err := providerService.ListProviders(ctx, 0, 50)

for _, provider := range providers {
    nsName := namespaceNames[provider.NamespaceID()]
    fmt.Printf("%s/%s - %s\n", nsName, provider.Name(), provider.Tier())
}
```

### Searching Providers

```go
providers, total, err := providerService.SearchProviders(ctx, "aws", 0, 20)
```

---

## GPG Integration

Provider versions can be signed with GPG keys:

```go
// Set GPG key on version
version.SetGPGKeyID("ABCD1234")

// Sign binary during publishing
signature, err := gpgKeyService.Sign(ctx, binaryPath, "ABCD1234")

// Store signature with binary
storageService.WriteFile(ctx, signaturePath, signature, true)
```

---

## Provider Categories

Categories organize providers:

```go
type ProviderCategory struct {
    id             int
    name           string
    slug           string
    userSelectable bool
}
```

**Example Categories**:
- Infrastructure (AWS, Azure, GCP)
- Networking (Cloudflare, Akamai)
- Monitoring (Datadog, New Relic)
- Security (Vault, Boundary)

---

## Repository Integration

Providers can be linked to Git repositories for automatic publishing:

```go
provider.SetRepositoryID(repoID)
provider.SetUseProviderSourceAuth(true)
```

This enables:
- Automatic publishing from Git releases
- Webhook-based updates
- Provider source authentication

---

## References

- [`/internal/domain/provider/model/`](./model/) - Provider models
- [`/internal/domain/provider/service/`](./service/) - Provider services
- [`/internal/domain/gpgkey/`](../gpgkey/) - GPG key management
- [`/internal/domain/storage/`](../storage/) - Storage for provider binaries
