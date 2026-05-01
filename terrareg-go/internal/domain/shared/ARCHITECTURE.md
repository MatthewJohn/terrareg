# Shared Domain Architecture

## Overview

The Shared domain provides cross-cutting concerns and shared utilities used across all domains in the Terrareg application. It contains common error definitions, shared types, utility services, and helper functions.

---

## Core Functionality

The shared domain provides the following capabilities:

- **Common Error Definitions** - Standard errors used across all domains
- **Shared Types** - Reusable type definitions (identifiers, git types)
- **Utility Services** - File processing, system commands, markdown rendering
- **Version Utilities** - Version parsing and comparison
- **Helper Functions** - Common utility functions

---

## Domain Components

### Errors

**Location**: `/internal/domain/shared/errors.go`

#### Standard Error Definitions

```go
var (
    ErrNotFound            = errors.New("not found")
    ErrAlreadyExists       = errors.New("already exists")
    ErrInvalidInput        = errors.New("invalid input")
    ErrUnauthorized        = errors.New("unauthorized")
    ErrForbidden           = errors.New("forbidden")
    ErrInvalidVersion      = errors.New("invalid version")
    ErrInvalidName         = errors.New("invalid name")
    ErrInvalidNamespace    = errors.New("invalid namespace")
    ErrInvalidProvider     = errors.New("invalid provider")
    ErrInvalidCategorySlug = errors.New("invalid category slug")
    ErrDomainViolation     = errors.New("domain rule violation")
)
```

#### DomainError Type

```go
type DomainError struct {
    Code    string
    Message string
    Err     error
}

func (e *DomainError) Error() string
func (e *DomainError) Unwrap() error
```

### Types

**Location**: `/internal/domain/shared/types/`

#### Identifier Types

```go
type NamespaceName string
type ModuleName string
type ProviderName string
type Version string
```

#### Git Types

```go
type GitURL string
type GitTag string
type GitCommit string
```

### Services

**Location**: `/internal/domain/shared/service/`

#### FileProcessingAdapter

Adapts file processing operations:

```go
type FileProcessingAdapter struct {
    // Config for file processing
}

func (a *FileProcessingAdapter) ProcessFile(path string) ([]byte, error)
func (a *FileProcessingAdapter) ValidateFileType(path string) bool
```

#### SystemCommandService

Executes system commands safely:

```go
type SystemCommandService struct {
    timeout time.Duration
}

func (s *SystemCommandService) Execute(ctx context.Context, cmd string, args ...string) ([]byte, error)
func (s *SystemCommandService) ExecuteWithDir(ctx context.Context, dir, cmd string, args ...string) ([]byte, error)
```

#### MarkdownService

Renders markdown to HTML:

```go
type MarkdownService struct {
    // Markdown rendering config
}

func (s *MarkdownService) Render(markdown string) (string, error)
func (s *MarkdownService) RenderSanitized(markdown string) (string, error)
```

### Utilities

**Location**: `/internal/domain/shared/`

#### Version Utilities

```go
func ParseVersion(version string) (*Version, error)
func CompareVersions(v1, v2 string) int
```

#### Helper Functions

```go
func SanitizeFilename(name string) string
func GenerateSlug(text string) string
func TruncateString(s string, maxLen int) string
```

---

## Dependencies

### Domain Dependencies

None - this domain provides utilities to other domains.

### Infrastructure Dependencies

| Component | Purpose |
|-----------|---------|
| **os/exec** | System command execution |
| **crypto** | Hashing and encryption utilities |

---

## Domains That Use Shared

All domains use the shared domain for:
- **Common errors** - Consistent error handling
- **Types** - Reusable type definitions
- **Utilities** - Common operations

---

## Key Design Principles

1. **No Dependencies** - Shared domain has no dependencies on other domains
2. **Reusability** - All utilities designed for reuse across domains
3. **Consistency** - Standard errors and types used consistently
4. **Simplicity** - Simple, focused utilities

---

## Error Handling Patterns

### Using Standard Errors

```go
// In repository
func (r *MyRepository) FindByID(ctx context.Context, id int) (*Entity, error) {
    var dbModel DBModel
    err := r.db.WithContext(ctx).First(&dbModel, id).Error
    if err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return nil, shared.ErrNotFound
        }
        return nil, fmt.Errorf("failed to find: %w", err)
    }
    return r.dbModelToDomain(&dbModel), nil
}

// In handler
if errors.Is(err, shared.ErrNotFound) {
    RespondError(w, http.StatusNotFound, "Resource not found")
    return
}
```

### Using DomainError

```go
func ValidateVersion(version string) error {
    if !semver.IsValid(version) {
        return &shared.DomainError{
            Code:    "INVALID_VERSION",
            Message: fmt.Sprintf("version %q is invalid", version),
        }
    }
    return nil
}
```

---

## Type Usage

### NamespaceName

```go
func CreateNamespace(name NamespaceName) (*Namespace, error) {
    if err := ValidateNamespaceName(string(name)); err != nil {
        return nil, err
    }
    // ...
}
```

### Git Types

```go
func CloneRepository(url GitURL, tag GitTag) error {
    cmd := exec.Command("git", "clone", string(url), "-b", string(tag))
    // ...
}
```

---

## System Command Execution

### Basic Execution

```go
sysCmd := shared.NewSystemCommandService(30 * time.Second)

output, err := sysCmd.Execute(ctx, "terraform", "version")
if err != nil {
    return fmt.Errorf("terraform not found: %w", err)
}
```

### Execution with Directory

```go
output, err := sysCmd.ExecuteWithDir(ctx, moduleDir, "terraform", "init")
if err != nil {
    return fmt.Errorf("terraform init failed: %w", err)
}
```

---

## Markdown Rendering

```go
markdownService := shared.NewMarkdownService()

html, err := markdownService.Render("# Module Documentation\n...")
if err != nil {
    return err
}
```

---

## File Processing

```go
adapter := shared.NewFileProcessingAdapter()

content, err := adapter.ProcessFile(modulePath)
if err != nil {
    return fmt.Errorf("failed to process file: %w", err)
}
```

---

## Version Utilities

```go
v1, err := shared.ParseVersion("1.2.3")
if err != nil {
    return err
}

v2, _ := shared.ParseVersion("1.2.4")

cmp := shared.CompareVersions(v1.String(), v2.String())
if cmp < 0 {
    fmt.Println("v1 is less than v2")
}
```

---

## Utility Functions

### Filename Sanitization

```go
safeName := shared.SanitizeFilename("module/name@123")
// Returns: "module-name-123"
```

### Slug Generation

```go
slug := shared.GenerateSlug("My Awesome Module")
// Returns: "my-awesome-module"
```

### String Truncation

```go
short := shared.TruncateString("Very long description...", 50)
// Returns first 50 characters with "..." if truncated
```

---

## References

- [`/internal/domain/shared/errors.go`](./errors.go) - Common errors
- [`/internal/domain/shared/types/`](./types/) - Shared types
- [`/internal/domain/shared/service/`](./service/) - Utility services
