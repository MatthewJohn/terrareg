# Module Domain Architecture

## Overview

The Module domain is the core of the Terrareg application, handling module registry functionality including publishing, versioning, processing, and management. It implements the aggregate pattern with `ModuleProvider` as the aggregate root owning `ModuleVersion` entities.

---

## Core Functionality

The module domain provides the following capabilities:

- **Module CRUD** - Create, read, update, and delete modules
- **Version Management** - Manage module versions with publishing workflow
- **Module Processing** - Extract and process module archives (Terraform files, examples, submodules)
- **Git Integration** - Import modules from Git repositories
- **Example File Management** - Generate and store example Terraform files
- **Submodule Handling** - Detect and process module submodules
- **Search and Discovery** - Search modules by namespace, name, provider

---

## Domain Components

### Aggregate Structure

```
ModuleProvider (Aggregate Root)
    ├── ModuleVersion[] (Entities)
    └── ModuleDetails (Value Object - shared across versions)

Namespace (Separate Aggregate)
    └── No ownership relationship with ModuleProvider
```

### Models

**Location**: `/internal/domain/module/model/`

#### ModuleProvider (Aggregate Root)

```go
type ModuleProvider struct {
    id               int
    namespaceID      int
    name             string
    description      *string
    gitURLTemplate   *string
    gitTagFormat     *string
    versions         []*ModuleVersion
}
```

**Key Methods**:
- `AddVersion(version)` - Add a version to the module
- `GetVersion(version)` - Get a specific version
- `SetDescription(desc)` - Update description
- `SetGitURLTemplate(template)` - Update Git URL template
- `SetGitTagFormat(format)` - Update git tag format

#### ModuleVersion (Entity)

```go
type ModuleVersion struct {
    id                    int
    moduleProviderID      int
    version               string
    moduleDetailsID        *int
    variableTemplate      []byte
    terraformExampleIds   []int
    publishedAt           *time.Time
    isBeta                bool
    owner                 string
}
```

**Lifecycle**: Managed by `ModuleProvider` aggregate

#### ModuleDetails (Value Object)

```go
type ModuleDetails struct {
    id                     int
    owner                  string
    moduleProviderID        *int
    description            *string
    variableTemplate       []byte
    readme                 string
    inputs                 []byte
    outputs                []byte
    dependencies           []byte
    resources              []byte
    optionalResources      []byte
    providerDependencies   []byte
    moduleProviderID       *int  // FK to module provider
}
```

**Key**: Shared across versions (each version can have its own details)

#### Namespace (Separate Aggregate)

```go
type Namespace struct {
    id           int
    name         string
    displayName  *string
    namespaceType NamespaceType
}
```

**Note**: Namespace is a separate aggregate, not owned by ModuleProvider

#### Other Models

| Model | Purpose |
|-------|---------|
| `ModuleVersionFile` | Files within a module version |
| `Types` | Module-related type definitions |
| `Errors` | Module-specific errors |

### Repository Interfaces

**Location**: `/internal/domain/module/repository/`

```go
type ModuleProviderRepository interface {
    Save(ctx context.Context, mp *ModuleProvider) (*ModuleProvider, error)
    FindByID(ctx context.Context, id int) (*ModuleProvider, error)
    FindByNamespaceAndName(ctx context.Context, namespace, name string) (*ModuleProvider, error)
    FindAll(ctx context.Context, offset, limit int) ([]*ModuleProvider, int, error)
    Delete(ctx context.Context, id int) error
}

type ModuleVersionRepository interface {
    Save(ctx context.Context, mv *ModuleVersion) (*ModuleVersion, error)
    FindByID(ctx context.Context, id int) (*ModuleVersion, error)
    FindByModuleProviderAndVersion(ctx context.Context, mpID int, version string) (*ModuleVersion, error)
    FindVersionsByModuleProvider(ctx context.Context, mpID int) ([]*ModuleVersion, error)
    Delete(ctx context.Context, id int) error
    UpdateModuleDetailsID(ctx context.Context, versionID, detailsID int) error
}

type ModuleDetailsRepository interface {
    Save(ctx context.Context, md *ModuleDetails) (int, error)
    FindByID(ctx context.Context, id int) (*ModuleDetails, error)
}

type NamespaceRepository interface {
    Save(ctx context.Context, ns *Namespace) (*Namespace, error)
    FindByID(ctx context.Context, id int) (*Namespace, error)
    FindByName(ctx context.Context, name NamespaceName) (*Namespace, error)
    FindAll(ctx context.Context) ([]*Namespace, error)
}

type ExampleFileRepository interface {
    Save(ctx context.Context, example *ExampleFile) error
    FindByModuleVersionID(ctx context.Context, mvID int) ([]*ExampleFile, error)
    DeleteByModuleVersionID(ctx context.Context, mvID int) error
}

type SubmoduleRepository interface {
    Save(ctx context.Context, submodule *Submodule) error
    FindByModuleVersionID(ctx context.Context, mvID int) ([]*Submodule, error)
    DeleteByModuleVersionID(ctx context.Context, mvID int) error
}
```

### Services

**Location**: `/internal/domain/module/service/`

#### ModuleProcessorService

Main service for processing module archives:

```go
type ModuleProcessorService struct {
    moduleParser        parser.ModuleParser
    exampleGenerator    ExampleGenerator
    terraformExecutor   TerraformExecutorService
}
```

Processes uploaded module archives:
1. Extract archive
2. Parse Terraform files
3. Generate examples
4. Detect submodules
5. Run security scanning (optional)

#### ModuleImporterService

Handles Git-based module imports:

```go
type ModuleImporterService struct {
    gitService       git.GitService
    moduleProcessor  *ModuleProcessorService
}
```

Imports modules from Git repositories:
1. Clone repository at git tag
2. Extract module content
3. Process like uploaded module

#### ModuleCreationWrapperService

Wraps module creation with additional processing:

```go
type ModuleCreationWrapperService struct {
    processor        *ModuleProcessorService
    detailsRepo      ModuleDetailsRepository
}
```

#### TransactionProcessingOrchestrator

Coordinates multi-step processing with transactions:

```go
type TransactionProcessingOrchestrator struct {
    savepointHelper *savepoint.SavepointHelper
}
```

---

## Dependencies

### Domain Dependencies

| Domain | Purpose |
|--------|---------|
| **git** | For Git repository operations and URL templates |
| **config** | For module processing configuration |
| **storage** | For module archive storage |
| **audit** | For audit logging of module operations |
| **analytics** | For download tracking |
| **gpgkey** | For module signing (optional) |

### Infrastructure Dependencies

| Component | Purpose |
|-----------|---------|
| **Database** | Persistent storage for modules |
| **Terraform** - For `terraform init` and graph generation |
| **Git** - For cloning repositories |

### Domains That Depend on Module

| Domain | Purpose |
|--------|---------|
| **graph** | Uses module data for dependency graphs |
| **analytics** | Tracks module downloads |

---

## Key Design Principles

1. **Aggregate Pattern** - ModuleProvider is aggregate root, ModuleVersion are entities
2. **Parent-Child Relationships** - Established via `setModuleProvider()` and `AddVersion()` methods
3. **Delete-Then-Create** - Python compatibility: delete and recreate during re-indexing
4. **Context Propagation** - All database operations use context for transactions
5. **Relationship Restoration** - Repository implementations restore parent-child relationships

---

## Module Processing Pipeline

The module processing pipeline handles both ZIP uploads and Git imports:

```
┌─────────────────────────────────────────────────────────────────────┐
│                     Module Processing Pipeline                      │
├─────────────────────────────────────────────────────────────────────┤
│                                                                       │
│  ┌─────────────────────────────────────────────────────────────────┐ │
│  │                   Phase 1: Input                                │ │
│  │  • ZIP Upload → /modules/{ns}/{name}/{provider}/{version}/upload│ │
│  │  • Git Import → Clone repository at git tag                     │ │
│  │  • Webhook → Trigger import from repository                     │ │
│  └─────────────────────────────────────────────────────────────────┘ │
│                              ↓                                        │
│  ┌─────────────────────────────────────────────────────────────────┐ │
│  │                   Phase 2: Extraction                           │ │
│  │  • Extract ZIP archive                                          │ │
│  │  • Validate module structure                                    │ │
│  │  • Locate main.tf and Terraform files                          │ │
│  └─────────────────────────────────────────────────────────────────┘ │
│                              ↓                                        │
│  ┌─────────────────────────────────────────────────────────────────┐ │
│  │                   Phase 3: Parsing                              │ │
│  │  • Parse Terraform configuration                               │ │
│  │  • Extract variables, outputs, resources                        │ │
│  │  • Build dependency graph                                       │ │
│  └─────────────────────────────────────────────────────────────────┘ │
│                              ↓                                        │
│  ┌─────────────────────────────────────────────────────────────────┐ │
│  │                   Phase 4: Example Generation                   │ │
│  │  • Run `terraform init`                                         │ │
│  │  • Generate Terraform example code                             │ │
│  │  • Store example files                                         │ │
│  └─────────────────────────────────────────────────────────────────┘ │
│                              ↓                                        │
│  ┌─────────────────────────────────────────────────────────────────┐ │
│  │                   Phase 5: Submodule Detection                 │ │
│  │  • Scan for module blocks                                      │ │
│  │  • Detect submodules                                           │ │
│  │  • Store submodule references                                  │ │
│  └─────────────────────────────────────────────────────────────────┘ │
│                              ↓                                        │
│  ┌─────────────────────────────────────────────────────────────────┐ │
│  │                   Phase 6: Storage & Database                   │ │
│  │  • Store module archive                                         │ │
│  │  • Save ModuleProvider, ModuleVersion, ModuleDetails            │ │
│  │  • Index for search                                            │ │
│  └─────────────────────────────────────────────────────────────────┘ │
│                                                                       │
└─────────────────────────────────────────────────────────────────────┘
```

---

## Processing Flows

### 1. ZIP Upload Flow

```
Client → POST /modules/{ns}/{name}/{provider}/{version}/upload
         ↓
       Extract Archive
         ↓
       Process Module Files
         ↓
       Generate Examples
         ↓
       Store in Database
         ↓
       Return 200 OK
```

**Characteristics**:
- Fast (no Git operations)
- Direct archive processing
- User-provided version number

### 2. Git Import Flow

```
Client → POST /modules/{ns}/{name}/{provider}/import
         ↓
       Clone Repository at Git Tag
         ↓
       Extract and Process
         ↓
       Generate Examples
         ↓
       Store in Database
         ↓
       Return 200 OK
```

**Characteristics**:
- Slower (Git clone operation)
- Version derived from git tag
- Uses git_tag_format for conversion

### 3. Webhook Flow

```
Git Provider → POST /v1/terrareg/modules/{ns}/{name}/{provider}/hooks/*
                    ↓
               Validate Signature
                    ↓
               Parse Payload (tag, repo)
                    ↓
               Trigger Git Import
                    ↓
               Process Module
                    ↓
               Return 200 OK
```

**Characteristics**:
- Automatic publishing from releases
- Signature validation required
- Uses provider source authentication

---

## Module Storage Paths

```
{data_directory}/
├── modules/
│   └── {namespace}/
│       └── {module}/
│           └── {provider}/
│               └── {version}/
│                   ├── module.zip           # Main module archive
│                   ├── module.json          # Module metadata
│                   └── examples/            # Generated examples
│                       ├── main.tf
│                       ├── variables.tf
│                       └── outputs.tf
└── upload/                          # Temporary upload storage
    └── {upload_id}/
```

---

## Git Tag Format

Git tag format controls version-to-tag conversion:

| Format | Version | Git Tag |
|--------|---------|---------|
| `v{version}` | 1.2.3 | v1.2.3 |
| `releases/v{major}.{minor}` | 1.2.3 | releases/v1.2 |
| `{major}.{patch}` | 1.2.3 | 1.3 (minor defaults to 0) |

See [`/internal/domain/git/ARCHITECTURE.md`](../git/ARCHITECTURE.md) for details.

---

## Usage Examples

### Creating a Module Provider

```go
namespace, _ := namespaceRepo.FindByName(ctx, "aws")

moduleProvider := model.NewModuleProvider(
    namespace.ID(),
    "vpc",
    &description,
    &gitURLTemplate,
    &gitTagFormat,
)

saved, err := moduleProviderRepo.Save(ctx, moduleProvider)
```

### Adding a Version

```go
version, err := model.NewModuleVersion(
    "1.0.0",
    nil,  // module details ID
    false, // not beta
)

moduleProvider.AddVersion(version)
saved, err := moduleProviderRepo.Save(ctx, moduleProvider)
```

### Processing a Module Upload

```go
// In the upload command
err := moduleProcessorService.ProcessUploadedModule(
    ctx,
    namespace,
    moduleName,
    provider,
    version,
    uploadedFilePath,
    userID,
)
```

### Importing from Git

```go
err := moduleImporterService.ImportModuleFromGit(
    ctx,
    namespace,
    moduleName,
    provider,
    gitURL,
    gitTag,
    userID,
)
```

---

## Relationship Management

### Parent-Child Relationships

Critical: Module versions must maintain their parent relationship:

```go
// In repository implementation
func (r *ModuleVersionRepositoryImpl) FindByID(ctx context.Context, id int) (*model.ModuleVersion, error) {
    var dbVersion sqldb.ModuleVersionDB
    err := r.db.Preload("ModuleProvider").First(&dbVersion, id).Error

    // IMPORTANT: Restore the module provider relationship
    if dbVersion.ModuleProviderID > 0 {
        moduleProvider := fromDBModuleProvider(&dbVersion.ModuleProvider)
        moduleProvider.SetVersions([]*model.ModuleVersion{moduleVersion})
    }

    return moduleVersion, nil
}
```

---

## Configuration

Module processing behavior controlled by configuration:

```bash
# Module processing settings
AUTO_PUBLISH_MODULE_VERSIONS=true
MODULE_VERSION_REINDEX_MODE=legacy
REQUIRED_MODULE_METADATA_ATTRIBUTES=attr1,attr2
AUTOGENERATE_MODULE_PROVIDER_DESCRIPTION=true
AUTOGENERATE_USAGE_BUILDER_VARIABLES=true

# Example file generation
EXAMPLE_FILE_EXTENSIONS=tf,tfvars,sh,json
TERRAFORM_EXAMPLE_VERSION_TEMPLATE={major}.{minor}.{patch}
```

---

## References

### Domain Documentation

- [`/internal/domain/module/model/`](./model/) - Module models
- [`/internal/domain/module/repository/`](./repository/) - Repository interfaces
- [`/internal/domain/module/service/`](./service/) - Module services

### Processing Documentation

- [`/docs/MODULE_PROCESSING_FLOWS.md`](../../docs/MODULE_PROCESSING_FLOWS.md) - Detailed processing flows
- [`/docs/module-import-process.md`](../../docs/module-import-process.md) - Import process details
- [`/docs/WEBHOOK_INTEGRATION.md`](../../docs/WEBHOOK_INTEGRATION.md) - Webhook-based publishing

### Related Domains

- [`/internal/domain/git/`](../git/) - Git integration and URL templates
- [`/internal/domain/storage/`](../storage/) - Module storage
- [`/internal/domain/analytics/`](../analytics/) - Download tracking
