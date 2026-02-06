# Terrareg Go - AI Development Guide

> **Comprehensive guide for AI development on the Terrareg Go codebase**

## Table of Contents

1. [Application Overview](#application-overview)
2. [DDD Architecture](#ddd-architecture)
3. [Code Flow & Request Processing](#code-flow--request-processing)
4. [Database Handling](#database-handling)
5. [Authentication System](#authentication-system)
6. [Validation Patterns](#validation-patterns)
7. [Configuration Management](#configuration-management)
8. [HTTP Timeout System](#http-timeout-system)
9. [Testing Patterns](#testing-patterns)
10. [Nil Checking Conventions](#nil-checking-conventions)
11. [Critical Constraints](#critical-constraints)
12. [Common Pitfalls](#common-pitfalls)
13. [Key Files to Understand](#key-files-to-understand)

---

## Application Overview

### What is Terrareg?

**Terrareg** is a self-hosted Terraform Module and Provider Registry. It allows teams to:

- **Host and distribute Terraform modules** - Private registry for custom modules
- **Host and distribute Terraform providers** - Custom provider binaries
- **Git integration** - Auto-publish modules from GitHub/GitLab/Bitbucket releases
- **Webhook automation** - Automated module publishing from Git releases
- **Multi-authentication** - SAML, OIDC, GitHub OAuth, API keys
- **Analytics tracking** - Download counts and usage statistics
- **Audit logging** - Complete change tracking

### Python vs Go Context

This is a **complete rewrite** of the Python Terrareg application with **100% API compatibility**:

| Python (Original) | Go (Current) |
|-------------------|--------------|
| Flask + SQLAlchemy | Chi + GORM |
| Direct DB access | Repository pattern |
| Duck typing | Strong interfaces |
| Celery background jobs | Goroutines |
| Jinja2 templates | API-first (JSON) |
| Session-based auth | Token + cookie hybrid |

### Key Directories

```
terrareg-go/
├── cmd/                      # Entry points (server, migrate)
├── internal/
│   ├── domain/               # Core business logic (bounded contexts)
│   │   ├── module/           # Module domain entities, services, repo interfaces
│   │   ├── provider/         # Provider domain (Terraform providers)
│   │   ├── auth/             # Authentication domain
│   │   ├── config/           # Configuration domain models
│   │   └── shared/           # Shared domain utilities
│   ├── application/          # Use case orchestration (CQRS)
│   │   ├── command/          # Write operations (commands)
│   │   └── query/            # Read operations (queries)
│   ├── infrastructure/       # Technical implementations
│   │   ├── persistence/      # Database repositories, GORM models
│   │   ├── auth/             # Auth method implementations
│   │   ├── storage/          # File system and S3 storage
│   │   └── config/           # Configuration loading
│   └── interfaces/           # HTTP handlers, middleware
│       └── http/
│           ├── handler/      # API handlers (terraform, terrareg)
│           ├── dto/          # Data transfer objects
│           └── middleware/   # Auth, logging, CORS
├── test/
│   ├── integration/          # End-to-end tests
│   └── testutils/            # Test helpers and mocks
└── static/                   # Front-end assets
```

---

## DDD Architecture

### Layer Separation (CRITICAL)

```
┌─────────────────────────────────────────────────────────────┐
│                    Interfaces Layer                         │
│  (HTTP handlers, middleware - NO business logic)             │
└─────────────────────────────────────────────────────────────┘
                            │
                            ▼
┌─────────────────────────────────────────────────────────────┐
│                   Application Layer                         │
│  (Commands/Queries - orchestrate domain use cases)          │
└─────────────────────────────────────────────────────────────┘
                            │
                            ▼
┌─────────────────────────────────────────────────────────────┐
│                     Domain Layer                            │
│  (Pure business logic - NO external dependencies)           │
│  - Entities, Value Objects, Aggregates                      │
│  - Domain Services                                          │
│  - Repository Interfaces (only interfaces, not impls)       │
└─────────────────────────────────────────────────────────────┘
                            │
                            ▼
┌─────────────────────────────────────────────────────────────┐
│                  Infrastructure Layer                        │
│  (Technical implementations - DB, storage, external APIs)    │
│  - Repository Implementations                                │
│  - External service clients                                 │
│  - Storage implementations                                   │
└─────────────────────────────────────────────────────────────┘
```

### Key DDD Concepts

**Bounded Contexts** (in `/internal/domain/`):
- `module/` - Module registry, versioning, publishing
- `provider/` - Provider registry, binaries
- `auth/` - Authentication, authorization, sessions
- `git/` - Git provider integration (anti-corruption layer)
- `analytics/` - Download tracking
- `audit/` - Audit trail

**Aggregates**:
- `ModuleProvider` (aggregate root) → owns `ModuleVersion` entities
- `Namespace` (aggregate root) → tenant/isolation boundary

**Repository Pattern**:
```go
// Domain: Interface definition only
type ModuleVersionRepository interface {
    FindByID(ctx context.Context, id int) (*model.ModuleVersion, error)
    Save(ctx context.Context, mv *model.ModuleVersion) (*model.ModuleVersion, error)
}

// Infrastructure: Implementation
type ModuleVersionRepositoryImpl struct {
    *baserepo.BaseRepository
}
```

### Core Domain Model

#### Module Aggregate (Most Important)
- **ModuleProvider**: Aggregate root - owns ModuleVersion entities
- **ModuleVersion**: Entity with lifecycle managed by ModuleProvider
- **Namespace**: Separate aggregate for namespace management
- **ModuleDetails**: Value object for module metadata

#### Critical Relationships
- `ModuleProvider` has many `ModuleVersion` entities
- `ModuleVersion` belongs to exactly one `ModuleProvider`
- `ModuleProvider` belongs to exactly one `Namespace`
- **Parent-child relationships are established via `setModuleProvider()` and `AddVersion()` methods**

---

## Code Flow & Request Processing

### Typical HTTP Request Flow

```
1. HTTP Request arrives
   │
2. Chi Router matches route
   │
3. Middleware executes (Session → Auth → Logging)
   │
4. Handler method called
   │
5. Handler invokes Command/Query
   │
6. Command/Query orchestrates Domain Services
   │
7. Domain Services use Repositories
   │
8. Repositories access Database (via context)
   │
9. Response propagates back
   │
10. Handler sends JSON response
```

### Example: Module Publishing Flow

```go
// 1. HTTP Handler
func (h *ModuleHandler) HandlePublishModuleVersion(w http.ResponseWriter, r *http.Request) {
    // Parse request
    var req dto.ModuleVersionPublishRequest
    json.NewDecoder(r.Body).Decode(&req)

    // 2. Invoke Command
    err := h.publishModuleVersionCommand.Execute(r.Context(), &req)

    // 3. Respond
    RespondJSON(w, http.StatusCreated, response)
}

// 2. Command (Application Layer)
func (c *PublishModuleVersionCommand) Execute(ctx context.Context, req *Request) error {
    return c.transactionHelper.WithTransaction(ctx, func(txCtx context.Context, tx *gorm.DB) error {
        // 3. Call Domain Service
        return c.moduleService.PublishModuleVersion(txCtx, req)
    })
}

// 3. Domain Service (Domain Layer)
func (s *ModuleService) PublishModuleVersion(ctx context.Context, req *Request) error {
    // Business logic
    module := model.NewModuleProvider(...)

    // 4. Use Repository
    return s.repo.Save(ctx, module)
}

// 4. Repository Implementation (Infrastructure Layer)
func (r *ModuleProviderRepositoryImpl) Save(ctx context.Context, mp *model.ModuleProvider) error {
    db := r.GetDBFromContext(ctx)  // Context-aware transaction
    return db.Create(&dbModel).Error
}
```

---

## Database Handling

### Context-Based Transaction Management (CRITICAL)

**All database operations MUST use context:**

```go
// ✅ CORRECT
func (r *MyRepo) Save(ctx context.Context, entity *Entity) error {
    db := r.GetDBFromContext(ctx)  // Gets existing transaction or creates new
    return db.Create(entity).Error
}

// ❌ WRONG - No context
func (r *MyRepo) Save(entity *Entity) error {
    return r.db.Create(entity).Error
}
```

### BaseRepository Pattern

Every repository inherits from `BaseRepository`:

```go
type MyRepositoryImpl struct {
    *baserepo.BaseRepository  // Embedded for transaction support
}

func NewMyRepository(db *gorm.DB, helper *savepoint.SavepointHelper) *MyRepositoryImpl {
    return &MyRepositoryImpl{
        BaseRepository: baserepo.NewBaseRepository(db, helper),
    }
}
```

### Savepoint Transactions

For nested operations, use savepoint transactions:

```go
err := s.savepointHelper.WithTransaction(ctx, func(ctx context.Context, tx *gorm.DB) error {
    // All operations here use same transaction
    version, err := s.versionRepo.Save(ctx, version)
    if err != nil {
        return err  // Automatic rollback
    }

    detailsID, err := s.detailsRepo.SaveAndReturnID(ctx, details)
    if err != nil {
        return err  // Automatic rollback
    }

    return s.versionRepo.UpdateModuleDetailsID(ctx, version.ID(), detailsID)
})
```

### GORM Models

- **Prefix**: All DB models end with `DB` (e.g., `ModuleVersionDB`)
- **Location**: `/internal/infrastructure/persistence/sqldb/models.go`
- **Mapping**: Domain ↔ Database models via mapper functions

```go
type ModuleVersionDB struct {
    ID               int        `gorm:"primaryKey;autoIncrement"`
    ModuleProviderID int        `gorm:"not null"`
    Version          string     `gorm:"type:varchar(128)"`
    VariableTemplate []byte     `gorm:"type:mediumblob"` // JSON blob
    ModuleProvider   ModuleProviderDB `gorm:"foreignKey:ModuleProviderID"`
}

func (ModuleVersionDB) TableName() string {
    return "module_version"  // Match Python table name
}
```

### Database Compatibility

- **Development**: SQLite (`sqlite:///terrareg.db`)
- **Production**: PostgreSQL or MySQL
- **Schema**: IDENTICAL to Python version for migration compatibility
- **Migrations**: `/internal/infrastructure/persistence/sqldb/migrations/`

### Critical Mappers

- `fromDBModuleVersion()`: Converts DB model to domain entity (⚠️ DOES NOT restore relationships)
- `toDBModuleVersion()`: Converts domain entity to DB model
- `fromDBModuleProvider()`: Converts DB model to domain entity
- `fromDBNamespace()`: Converts DB model to domain entity

### Critical Issues Fixed

#### Module Provider Relationship Loss

**Problem**: Module versions loaded from database lose their parent ModuleProvider relationship, causing `module_provider_id=0` corruption.

**Root Cause**: `fromDBModuleVersion()` creates domain entities but doesn't establish parent-child relationships.

**Solution**: Enhanced `mapToDomainModel()` in `ModuleVersionRepositoryImpl`:
```go
// IMPORTANT: Restore the module provider relationship if module_provider_id exists
if dbVersion.ModuleProviderID > 0 {
    var moduleProviderDB sqldb.ModuleProviderDB
    err := r.db.Preload("Namespace").First(&moduleProviderDB, dbVersion.ModuleProviderID).Error
    if err == nil {
        namespace := fromDBNamespace(&moduleProviderDB.Namespace)
        moduleProvider := fromDBModuleProvider(&moduleProviderDB, namespace)
        if moduleProvider != nil {
            moduleProvider.SetVersions([]*model.ModuleVersion{moduleVersion})
        }
    }
}
```

---

## Authentication System

### Architecture

Four-tier authentication system following DDD principles:

```
┌─────────────────────────────────────────────────────┐
│              AuthFactory (Orchestrator)              │
│  - Tries auth methods in priority order              │
│  - Returns AuthContext interface                     │
└─────────────────────────────────────────────────────┘
          │
    ┌─────┴─────┐
    ▼           ▼
┌────────────────┐  ┌──────────────────────────────┐
│SessionManagement│  │   Infrastructure AuthMethods  │
│Service (Coord)  │  │ - AdminApiKey, SAML, OIDC... │
│                 │  └──────────────────────────────┘
│  Coordinates    │              │
│  session +      │              ▼
│  cookie ops     │        ┌────────────┐
└────────────────┘        │   Domain   │
       │      │           │  AuthCtx   │
   ┌───┴──────┴───┐       │  Interface │
   ▼              ▼       └────────────┘
┌────────┐  ┌────────┐
│ Session│  │ Cookie │
│Service │  │Service │
│(DB Ops)│  │(Crypto)│
└────────┘  └────────┘
```

### Key Services

1. **SessionService** - Pure database operations (create, validate, delete, refresh sessions)
2. **CookieService** - Cookie encryption/decryption (AES-256-GCM) and HTTP cookie management
3. **SessionManagementService** - Coordinates session + cookie operations (CRUD on both)
4. **AuthFactory** - Orchestrates authentication by trying auth methods in priority order

### Session Management

**Server-side storage** (database):
```go
type Session struct {
    ID           string      `json:"id" gorm:"primaryKey"`
    AuthMethod   string      `json:"auth_method"`
    ProviderData []byte      `json:"provider_data"`
    Expiry       time.Time   `json:"expiry"`
}
```

**Client-side storage** (encrypted cookie):
```go
type SessionData struct {
    SessionID   string            `json:"session_id"`
    Username    string            `json:"username"`
    AuthMethod  string            `json:"auth_method"`
    IsAdmin     bool              `json:"is_admin"`
    Permissions map[string]string `json:"permissions,omitempty"`
}
```

### Cookie Encryption

- **Algorithm**: AES-256-GCM (authenticated encryption)
- **Key**: 32-byte derived from `SECRET_KEY` config
- **Nonce**: 12-byte random per encryption
- **Format**: Base64(nonce + ciphertext + auth tag)

### Auth Methods

**Supported**:
- `SAML` - SAML2 identity provider
- `OIDC` - OpenID Connect
- `GITHUB` - GitHub OAuth
- `API_KEY` - Admin API key authentication
- `TERRAFORM_IDP` - Terraform Cloud/Enterprise IDP

### Authentication Context

Access auth info in handlers:

```go
func (h *MyHandler) HandleRequest(w http.ResponseWriter, r *http.Request) {
    authCtx := middleware.GetAuthenticationContext(r)

    if !authCtx.IsAuthenticated {
        RespondError(w, http.StatusUnauthorized, "Not authenticated")
        return
    }

    username := authCtx.Username
    isAdmin := authCtx.IsAdmin
}
```

### Critical Security Note

Upload endpoints require specific permission checks, not just generic authentication:
```go
// Correct approach - specific permission
r.With(s.authMiddleware.RequireUploadPermission("{namespace}")).Post("/modules/{namespace}/{name}/{provider}/{version}/upload")

// Incorrect approach - generic auth
r.With(s.authMiddleware.RequireAuth).Post("/modules/{namespace}/{name}/{provider}/{version}/upload")
```

---

## Validation Patterns

### Multi-Layer Validation

```
Handler Layer        → Basic parameter checks (empty, required)
Application Layer    → Request DTO validation (format, type)
Domain Layer         → Business rule validation (invariants)
Database Layer       → Constraint enforcement (unique, foreign key)
```

### Handler Level Validation

```go
func (h *ModuleHandler) HandleGetModule(w http.ResponseWriter, r *http.Request) {
    namespace := chi.URLParam(r, "namespace")

    // Basic validation
    if namespace == "" {
        RespondError(w, http.StatusBadRequest, "Missing namespace")
        return
    }

    // Continue processing...
}
```

### Service Level Validation

```go
func (c *OidcLoginCommand) ValidateRequest(req *OidcLoginRequest) error {
    if req == nil {
        return fmt.Errorf("request cannot be nil")
    }
    if req.RedirectURL == "" {
        return fmt.Errorf("redirect URL cannot be empty")
    }

    // URL format validation
    if _, err := url.Parse(req.RedirectURL); err != nil {
        return fmt.Errorf("invalid redirect URL: %w", err)
    }

    return nil
}
```

### Domain Model Validation

```go
func NewNamespace(name string, displayName *string, nsType NamespaceType) (*Namespace, error) {
    if err := ValidateNamespaceName(name); err != nil {
        return nil, err
    }
    // ... create and return namespace
}

func ValidateNamespaceName(name string) error {
    if name == "" {
        return fmt.Errorf("namespace name cannot be empty")
    }
    if len(name) < 2 {
        return fmt.Errorf("namespace name must be at least 2 characters")
    }
    if !namespaceRegex.MatchString(name) {
        return fmt.Errorf("namespace name contains invalid characters")
    }
    return nil
}
```

### Error Response Format

```go
// Simple error
RespondError(w, http.StatusBadRequest, "Invalid input")

// JSON error
RespondJSON(w, http.StatusBadRequest, dto.Error{Message: "Invalid input"})

// Domain error
type DomainError struct {
    Code    string
    Message string
}
```

---

## Error Handling Patterns

### Use Error Types, Not String Matching

**CRITICAL**: Always use proper error types for error checking, never string matching.

```go
// ❌ WRONG - String matching
if err != nil && strings.Contains(err.Error(), "not found") {
    return http.StatusNotFound
}

// ✅ CORRECT - Error type checking
if errors.Is(err, shared.ErrNotFound) {
    return http.StatusNotFound
}
```

### Standard Error Types

Common errors are defined in `/internal/domain/shared/errors.go`:

```go
var (
    ErrNotFound            = errors.New("not found")
    ErrAlreadyExists       = errors.New("already exists")
    ErrInvalidInput        = errors.New("invalid input")
    ErrUnauthorized        = errors.New("unauthorized")
    ErrForbidden           = errors.New("forbidden")
    // ... more standard errors
)
```

### Repository Error Pattern

Repositories should return `shared.ErrNotFound` when a record doesn't exist:

```go
// In repository implementation
func (r *MyRepositoryImpl) FindByID(ctx context.Context, id int) (*Model, error) {
    var dbModel DBModel
    err := r.db.WithContext(ctx).First(&dbModel, id).Error
    if err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return nil, shared.ErrNotFound  // ✅ Return standard error type
        }
        return nil, fmt.Errorf("failed to find: %w", err)
    }
    return r.dbModelToDomain(&dbModel), nil
}
```

### Query/Command Error Handling

Application layer queries and commands should use `errors.Is()` to check for specific errors:

```go
// In query or command
func (q *MyQuery) Execute(ctx context.Context, req Request) (*Result, error) {
    entity, err := q.repo.FindByName(ctx, req.Name)
    if err != nil {
        // Check for not found error
        if errors.Is(err, shared.ErrNotFound) {
            return nil, shared.ErrNotFound  // ✅ Return standard error type
        }
        return nil, fmt.Errorf("failed to get entity: %w", err)
    }
    // ... process entity
}
```

### Handler Error Response

HTTP handlers should check error types and return appropriate HTTP status codes:

```go
// In HTTP handler
func (h *MyHandler) HandleRequest(w http.ResponseWriter, r *http.Request) {
    result, err := h.myQuery.Execute(r.Context(), req)
    if err != nil {
        // Check for specific error types
        if errors.Is(err, shared.ErrNotFound) {
            RespondError(w, http.StatusNotFound, "Resource not found")
            return
        }
        if errors.Is(err, shared.ErrInvalidInput) {
            RespondError(w, http.StatusBadRequest, err.Error())
            return
        }
        // Generic error handling
        RespondError(w, http.StatusInternalServerError, err.Error())
        return
    }
    RespondJSON(w, http.StatusOK, result)
}
```

### DomainError for Custom Errors

For domain-specific errors, use the `DomainError` type:

```go
// Domain error with code and message
type DomainError struct {
    Code    string
    Message string
    Err     error
}

func (e *DomainError) Error() string {
    if e.Err != nil {
        return e.Message + ": " + e.Err.Error()
    }
    return e.Message
}

func (e *DomainError) Unwrap() error {
    return e.Err
}

// Usage
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

### Error Wrapping Guidelines

**When to wrap errors:**
- ✅ Wrap with context when passing up layers: `fmt.Errorf("failed to get user: %w", err)`
- ✅ Wrap with `shared.ErrNotFound` for not found cases: `return nil, shared.ErrNotFound`
- ❌ Don't wrap if you're just adding context to a standard error type

**Error wrapping patterns:**
```go
// ✅ GOOD - Direct return of standard error type
if entity == nil {
    return nil, shared.ErrNotFound
}

// ✅ GOOD - Wrap with context using %w
if err != nil {
    return nil, fmt.Errorf("failed to query database: %w", err)
}

// ❌ BAD - Wrapping standard error (loses error type)
if entity == nil {
    return nil, fmt.Errorf("entity not found: %w", shared.ErrNotFound)
}
```

### Summary

| Pattern | Correct Usage |
|---------|--------------|
| **Define errors** | Use `var ErrNotFound = errors.New("not found")` in shared package |
| **Return errors** | Return `shared.ErrNotFound` directly, not wrapped |
| **Check errors** | Use `errors.Is(err, shared.ErrNotFound)` not string matching |
| **HTTP status** | Check error type in handler, return appropriate status code |
| **Custom errors** | Use `DomainError` for domain-specific error codes |

---

## Provider Source Architecture

### Overview

The Provider Source system manages external Git provider integrations (GitHub, GitLab, etc.) for OAuth authentication and repository operations. It uses a factory pattern with class registration to support multiple provider types.

### Python Reference

- **Base**: `terrareg/server/provider_source/base.py::BaseProviderSource`
- **GitHub**: `terrareg/server/provider_source/github.py::GithubProviderSource`
- **Factory**: `terrareg/server/provider_source/factory.py::ProviderSourceFactory`

### Architecture Diagram

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                              Provider Source System                         │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│  ┌─────────────────────────────────────────────────────────────────────┐    │
│  │                    Domain Layer (provider_source)                    │    │
│  │                                                                      │    │
│  │  ┌──────────────────────┐  ┌──────────────────────────────────────┐ │    │
│  │  │ GithubProviderSource │  │      BaseProviderSource              │ │    │
│  │  │                      │  │  (embeds ProviderSourceClass)        │ │    │
│  │  │ - GetLoginRedirectURL│  │                                      │ │    │
│  │  │ - GetUserAccessToken │  │ - Name() string                     │ │    │
│  │  │ - GetUsername        │  │ - ApiName(ctx) (string, error)      │ │    │
│  │  │ - GetUserOrganizations│ │ - Config(ctx) (*Config, error)       │ │    │
│  │  │ - RefreshNamespace   │  │                                      │ │    │
│  │  │ - PublishProvider    │  │                                      │ │    │
│  │  └──────────┬───────────┘  └──────────────────┬───────────────────┘ │    │
│  │             │                                  │                       │    │
│  │             │                                  │                       │    │
│  │  ┌──────────▼──────────────────────────────────▼───────────────────┐ │    │
│  │  │           GithubProviderSourceClass                             │ │    │
│  │  │           (implements ProviderSourceClass)                      │ │    │
│  │  │                                                                  │ │    │
│  │  │  - Type() ProviderSourceType                                    │ │    │
│  │  │  - GenerateDBConfigFromSourceConfig(config) (*Config, error)    │ │    │
│  │  │  - CreateInstance(name, repo, db) (ProviderSourceInstance, error)│ │    │
│  │  └──────────────────────────────────────────────────────────────────┘ │    │
│  └─────────────────────────────────────────────────────────────────────┘    │
│                                    │                                        │
│                                    │ implements                              │
│                                    ▼                                        │
│  ┌─────────────────────────────────────────────────────────────────────┐    │
│  │              Service Layer (provider_source/service)                 │    │
│  │                                                                      │    │
│  │  ┌──────────────────────────────────────────────────────────────┐   │    │
│  │  │           ProviderSourceClass (interface)                    │   │    │
│  │  │                                                              │   │    │
│  │  │  - Type() ProviderSourceType                                │   │    │
│  │  │  - GenerateDBConfigFromSourceConfig(config) (*Config, error) │   │    │
│  │  │  - CreateInstance(name, repo, db) (ProviderSourceInstance, error)│   │    │
│  │  └──────────────────────────────────────────────────────────────┘   │    │
│  │                                                                      │    │
│  │  ┌──────────────────────────────────────────────────────────────┐   │    │
│  │  │            ProviderSourceFactory                             │   │    │
│  │  │                                                              │   │    │
│  │  │  - repo: ProviderSourceRepository                           │   │    │
│  │  │  - db: interface{} (database reference)                      │   │    │
│  │  │  - classMapping: map[Type]ProviderSourceClass                │   │    │
│  │  │                                                              │   │    │
│  │  │  Methods:                                                    │   │    │
│  │  │  - RegisterProviderSourceClass(class)                        │   │    │
│  │  │  - GetProviderClasses() map[Type]Class                       │   │    │
│  │  │  - GetProviderSourceByName(ctx, name) (Instance, error)      │   │    │
│  │  │  - GetProviderSourceByApiName(ctx, apiName) (Instance, error) │   │    │
│  │  │  - GetAllProviderSources(ctx) ([]Instance, error)             │   │    │
│  │  │  - InitialiseFromConfig(ctx, configJSON) error                │   │    │
│  │  └──────────────────────────────────────────────────────────────┘   │    │
│  └─────────────────────────────────────────────────────────────────────┘    │
│                                    │                                        │
│                                    │ uses                                   │
│                                    ▼                                        │
│  ┌─────────────────────────────────────────────────────────────────────┐    │
│  │              Model Layer (provider_source/model)                     │    │
│  │                                                                      │    │
│  │  ┌──────────────────────┐  ┌──────────────────────────────────────┐ │    │
│  │  │   ProviderSource     │  │      ProviderSourceConfig            │ │    │
│  │  │                      │  │                                      │ │    │
│  │  │  - name string       │  │  - BaseURL string                    │ │    │
│  │  │  - apiName string    │  │  - ApiURL string                     │ │    │
│  │  │  - type Type         │  │  - ClientID string                   │ │    │
│  │  │  - config *Config    │  │  - ClientSecret string               │ │    │
│  │  └──────────────────────┘  │  - PrivateKeyPath string             │ │    │
│  │                             │  - AppID string                      │ │    │
│  │                             │  - LoginButtonText string            │ │    │
│  │                             │  - DefaultAccessToken string         │ │    │
│  │                             │  - DefaultInstallationID string      │ │    │
│  │                             │  - AutoGenerateNamespaces bool       │ │    │
│  │                             └──────────────────────────────────────┘ │    │
│  └─────────────────────────────────────────────────────────────────────┘    │
│                                    │                                        │
│                                    │ persists                               │
│                                    ▼                                        │
│  ┌─────────────────────────────────────────────────────────────────────┐    │
│  │         Repository Layer (provider_source/repository)                │    │
│  │                                                                      │    │
│  │  ┌──────────────────────────────────────────────────────────────┐   │    │
│  │  │      ProviderSourceRepository (interface)                    │   │    │
│  │  │                                                              │   │    │
│  │  │  - Upsert(ctx, source) error                                 │   │    │
│  │  │  - FindByName(ctx, name) (*ProviderSource, error)            │   │    │
│  │  │  - FindByApiName(ctx, apiName) (*ProviderSource, error)       │   │    │
│  │  │  - FindAll(ctx) ([]*ProviderSource, error)                   │   │    │
│  │  └──────────────────────────────────────────────────────────────┘   │    │
│  └─────────────────────────────────────────────────────────────────────┘    │
│                                    │                                        │
│                                    │ implements                             │
│                                    ▼                                        │
│  ┌─────────────────────────────────────────────────────────────────────┐    │
│  │      Infrastructure Layer (persistence/sqldb/provider_source)         │    │
│  │                                                                      │    │
│  │  ┌──────────────────────────────────────────────────────────────┐   │    │
│  │  │     ProviderSourceRepositoryImpl                              │   │    │
│  │  │     (implements ProviderSourceRepository)                     │   │    │
│  │  │                                                              │   │    │
│  │  │  - db *gorm.DB                                               │   │    │
│  │  │  - Converts between domain/model and DB models               │   │    │
│  │  └──────────────────────────────────────────────────────────────┘   │    │
│  └─────────────────────────────────────────────────────────────────────┘    │
│                                                                              │
└─────────────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                      Container Initialization                                │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│  1. Create repository:                                                       │
│     repo := provider_source.NewProviderSourceRepository(db.DB)               │
│                                                                              │
│  2. Create factory:                                                          │
│     factory := service.NewProviderSourceFactory(repo)                        │
│                                                                              │
│  3. Set database on factory (for provider instances):                        │
│     factory.SetDatabase(db)                                                  │
│                                                                              │
│  4. Register provider source classes:                                        │
│     githubClass := provider_source.NewGithubProviderSourceClass()            │
│     factory.RegisterProviderSourceClass(githubClass)                         │
│                                                                              │
└─────────────────────────────────────────────────────────────────────────────┘
```

### Key Files

| Layer | File | Purpose |
|-------|------|---------|
| **Domain** | `internal/domain/provider_source/github_provider_source.go` | GitHub provider implementation |
| **Domain** | `internal/domain/provider_source/base_provider_source.go` | Base provider source functionality |
| **Domain** | `internal/domain/provider_source/github_provider_source_class.go` | GitHub provider class (config validation, instance creation) |
| **Model** | `internal/domain/provider_source/model/provider_source.go` | Domain models |
| **Service** | `internal/domain/provider_source/service/provider_source_factory.go` | Factory pattern for provider sources |
| **Repository** | `internal/domain/provider_source/repository/provider_source_repository.go` | Repository interface |
| **Infra** | `internal/infrastructure/persistence/sqldb/provider_source/` | Repository implementation |

### ProviderSourceInstance Interface

The `ProviderSourceInstance` interface defines all operations that can be performed on a provider source:

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
    PublishProviderFromRepository(ctx context.Context, repoID int, categoryID int, namespace string) (*PublishProviderResult, error)
}
```

### Factory Pattern

The factory pattern allows for:
1. **Class Registration**: Provider classes are registered at startup
2. **Type-based Discovery**: Get provider class by type
3. **Lazy Instantiation**: Provider instances created on-demand
4. **Database Injection**: Database passed to instances for operations

```go
// Registration (during container init)
githubClass := provider_source.NewGithubProviderSourceClass()
factory.RegisterProviderSourceClass(githubClass)

// Usage (in handlers/queries)
providerSource, err := factory.GetProviderSourceByName(ctx, "GitHub")
if err != nil {
    return err
}
redirectURL, err := providerSource.GetLoginRedirectURL(ctx)
```

### Configuration Management

Provider sources are configured via JSON and initialized into the database:

```go
configJSON := `[
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
]`

err := factory.InitialiseFromConfig(ctx, configJSON)
```

### Critical Design Decisions

1. **Class Moved to provider_source Package**: `GithubProviderSourceClass` is in the `provider_source` package (not `service`) to avoid circular imports since it needs to create `GithubProviderSource` instances.

2. **Database Stored in Factory**: The factory stores a database reference and passes it to provider instances via `CreateInstance()`. This allows providers to perform database operations without tight coupling.

3. **Interface-based Design**: All provider operations go through the `ProviderSourceInstance` interface, allowing different provider types (GitHub, GitLab, etc.) to be used interchangeably.

4. **Lazy Instance Creation**: The factory stores provider source data (from DB) and creates actual implementation instances (`GithubProviderSource`) only when methods are called, via `getProviderSourceImplementation()`.

---

## Configuration Management

### Three-Tier Configuration

```
DomainConfig           → Business logic, UI settings, feature flags
InfrastructureConfig   → Technical: DB, storage, secrets
UIConfig              → Read-only view for presentation
```

### DomainConfig (Business Logic)

```go
type DomainConfig struct {
    // Namespace settings
    TrustedNamespaces          []string `env:"TRUSTED_NAMESPACES" envDefault:""`
    VerifiedModuleNamespaces  []string `env:"VERIFIED_MODULE_NAMESPACES" envDefault:""`

    // Feature flags
    AllowModuleHosting         bool    `env:"ALLOW_MODULE_HOSTING" envDefault:"true"`
    UploadAPIKeysEnabled       bool    `env:"UPLOAD_API_KEYS_ENABLED" envDefault:"true"`

    // UI labels
    TrustedNamespaceLabel      string  `env:"TRUSTED_NAMESPACE_LABEL" envDefault:"Trusted"`

    // Analytics
    AnalyticsTokenPhrase       string  `env:"ANALYTICS_TOKEN_PHRASE" envDefault:"my-tf-application"`
    DisableAnalytics           bool    `env:"DISABLE_ANALYTICS" envDefault:"false"`
}
```

### InfrastructureConfig (Technical)

```go
type InfrastructureConfig struct {
    // Server
    ListenPort     int    `env:"LISTEN_PORT" envDefault:"5000"`
    PublicURL      string `env:"PUBLIC_URL" envDefault:"http://localhost:5000"`

    // Database
    DatabaseURL    string `env:"DATABASE_URL" envDefault:"sqlite:///terrareg.db"`

    // Storage
    DataDirectory  string `env:"DATA_DIRECTORY" envDefault:"./data"`

    // Secrets
    SecretKey      string `env:"SECRET_KEY" envDefault:""`

    // Session
    SessionExpiry  string `env:"SESSION_EXPIRY" envDefault:"24h"`
}
```

### Usage Pattern

```go
// Domain Service → Inject DomainConfig
type NamespaceService struct {
    config *model.DomainConfig  // ✅ Only domain config
    repo   NamespaceRepository
}

// Infrastructure Service → Inject InfrastructureConfig
type SessionService struct {
    config *config.InfrastructureConfig  // ✅ Only infra config
    repo   SessionRepository
}

// Application Handler → Use Query
type ConfigHandler struct {
    getConfigQuery *configQuery.GetConfigQuery  // ✅ Returns UIConfig
}
```

### Configuration System Architecture & Intricacies

#### Current Implementation Challenges

**1. Manual Configuration Building**

**Problem**: The configuration service manually constructs `InfrastructureConfig` instead of using struct tag libraries.

**Current Approach**:
```go
// In buildInfrastructureConfig()
return &config.InfrastructureConfig{
    ListenPort: s.parseInt(rawConfig["LISTEN_PORT"], 5000),
    DatabaseURL: s.getEnvStringWithDefault(rawConfig, "DATABASE_URL", "sqlite:///modules.db"),
    TerraformLockTimeoutSeconds: s.parseInt(rawConfig["TERRAFORM_LOCK_TIMEOUT_SECONDS"], 1800),
    // ... many more fields
}
```

**Why This Is Problematic**:
- Error-prone: Easy to miss new fields when adding struct properties
- Repetitive: Every field requires manual parsing logic
- Maintenance burden: Changes require updates in multiple places
- Inconsistent defaults: Different developers might use different default values

**2. Struct Tag Mismatch**

**The Issue**: Struct has `env:` and `envDefault:` tags but no library processes them:
```go
type InfrastructureConfig struct {
    ListenPort                     int    `env:"LISTEN_PORT" envDefault:"5000"`
    TerraformLockTimeoutSeconds   int    `env:"TERRAFORM_LOCK_TIMEOUT_SECONDS" envDefault:"1800"`
}
```

**Current Reality**: These tags are ignored - configuration is manually parsed from raw environment variables.

**3. Architectural Debt**

**What Should Happen**: Use a configuration library (Viper, envconfig, etc.) that automatically handles:
- Environment variable loading
- Default value application
- Type conversion
- Validation

**Current State**: Manual parsing with helper methods like `s.parseInt()`, `s.parseBool()`, etc.

### Recommended Future Improvements

#### 1. Adopt a Configuration Library
```go
// Example with envconfig
type InfrastructureConfig struct {
    ListenPort                     int    `env:"LISTEN_PORT" envDefault:"5000"`
    TerraformLockTimeoutSeconds   int    `env:"TERRAFORM_LOCK_TIMEOUT_SECONDS" envDefault:"1800"`
}

// Automatic loading
var config InfrastructureConfig
if err := envconfig.Process("TERRAREG_", &config); err != nil {
    log.Fatal(err)
}
```

#### 2. Configuration Validation
```go
func (c *InfrastructureConfig) Validate() error {
    if c.TerraformLockTimeoutSeconds < 60 {
        return errors.New("TERRAFORM_LOCK_TIMEOUT_SECONDS must be at least 60 seconds")
    }
    return nil
}
```

### Current Configuration Loading Process

#### 1. Environment Variable Collection
```go
// In ConfigurationService
rawConfig := s.envLoader.LoadAllEnvironmentVariables()
```

#### 2. Manual Struct Construction
```go
// In buildInfrastructureConfig
return &config.InfrastructureConfig{
    TerraformLockTimeoutSeconds: s.parseInt(rawConfig["TERRAFORM_LOCK_TIMEOUT_SECONDS"], 1800),
    // ... every single field manually
}
```

#### 3. Manual Type Conversion
```go
func (s *ConfigurationService) parseInt(key string, defaultValue int) int {
    if value, exists := rawConfig[key]; exists {
        if parsed, err := strconv.Atoi(value); err == nil {
            return parsed
        }
    }
    return defaultValue
}
```

### Key Takeaway

The current manual configuration approach works but is fragile. When adding new configuration fields:
1. **Add field to struct** (with env/envDefault tags for future use)
2. **Add field to manual construction** in `buildInfrastructureConfig()`
3. **Add appropriate parsing logic** if needed
4. **Test with both set and unset environment variables**

This dual requirement makes the system error-prone and explains why configuration fields are sometimes missed.

---

## HTTP Timeout System

### Overview

The Go Terrareg implementation uses a sophisticated timeout system to handle both standard HTTP requests and long-running operations like module imports, uploads, and webhook processing.

### Configuration

All timeout values are configurable via environment variables:

```bash
# Standard HTTP requests (default: 60 seconds)
STANDARD_REQUEST_TIMEOUT_SECONDS=60

# Extended timeout for long-running operations (default: 30 minutes)
MODULE_INDEXING_TIMEOUT_SECONDS=1800

# Terraform processing timeout (default: 30 minutes)
TERRAFORM_LOCK_TIMEOUT_SECONDS=1800
```

### Implementation Architecture

#### 1. Chi Middleware for HTTP Timeouts

- **Location**: `/internal/interfaces/http/server.go`
- **Pattern**: Route-specific timeout middleware applied before authentication

```go
r.With(
    middleware.Timeout(time.Duration(s.infraConfig.ModuleIndexingTimeoutSeconds) * time.Second),
    s.authMiddleware.RequireAuth,
).Post("/modules/{namespace}/{name}/{provider}/{version}/import", s.handleModuleVersionCreate)
```

#### 2. Terraform Processing Timeouts

- **Location**: `/internal/infrastructure/container/container.go`
- **Configuration**: Passed to `TerraformExecutorService` during container initialization
- **Purpose**: Controls global terraform lock acquisition for security scanning, graph generation, etc.

#### 3. Routes With Extended Timeout (30 minutes)

- `POST /modules/{namespace}/{name}/{provider}/{version}/upload` - File upload and processing
- `POST /modules/{namespace}/{name}/{provider}/{version}/import` - Git cloning and extraction
- `POST /modules/{namespace}/{name}/{provider}/import` - Module provider import
- `POST /v1/terrareg/modules/{namespace}/{name}/{provider}/hooks/*` - Webhook processing

#### 4. Routes With Standard Timeout (60 seconds)

- `POST /modules/{namespace}/{name}/{provider}/{version}/publish` - Quick database flag update
- All other routes (GET, DELETE, etc.)

### Critical Implementation Details

#### 1. Configuration Loading

The configuration service manually builds `InfrastructureConfig` from environment variables. **Important**: When adding new configuration fields, they must be added to the `buildInfrastructureConfig()` method in `/internal/domain/config/service/configuration_service.go`.

#### 2. Middleware Ordering

Timeout middleware must be applied **before** authentication middleware:

```go
// ✅ Correct: Timeout context created first, then auth runs within timeout
r.With(
    middleware.Timeout(30 * time.Minute),    // Timeout first
    s.authMiddleware.RequireAuth,            // Then auth within timeout
).Post("/import", handler)

// ❌ Wrong: Auth runs first, then timeout is applied (can corrupt context)
r.With(
    s.authMiddleware.RequireAuth,
    middleware.Timeout(30 * time.Minute),
).Post("/import", handler)
```

#### 3. Terraform Global Lock

The terraform processing service uses a global mutex to ensure only one terraform operation runs at a time. The lock timeout is configurable and prevents deadlocks.

### Common Timeout-Related Issues

#### 1. "context deadline exceeded" with 0s timeout

**Problem**: TerraformLockTimeoutSeconds configuration field not being loaded

**Solution**: Add the field to `buildInfrastructureConfig()` method with proper default
```go
TerraformLockTimeoutSeconds: s.parseInt(rawConfig["TERRAFORM_LOCK_TIMEOUT_SECONDS"], 1800),
```

#### 2. Immediate timeouts in under 1ms

**Problem**: Hard-coded timeouts in service layer (30s or 60s) instead of using configured values

**Solution**: Use configured timeout from InfrastructureConfig
```go
// In container.go
time.Duration(infraConfig.TerraformLockTimeoutSeconds) * time.Second
```

#### 3. Middleware timeout conflicts

**Problem**: Global HTTP server timeout conflicting with route-specific timeouts

**Solution**: Remove global timeout middleware, use only route-specific timeouts

### Testing Timeout Behavior

#### 1. Test Environment Variables
```go
os.Setenv("MODULE_INDEXING_TIMEOUT_SECONDS", "10") // 10 seconds for testing
os.Setenv("TERRAFORM_LOCK_TIMEOUT_SECONDS", "15")  // 15 seconds for testing
```

#### 2. Test Long-Running Operations

- Simulate slow Git operations
- Test with large repositories
- Verify timeout cancellation works properly

#### 3. Monitor Timeout Logs

- Check for "Unable to obtain global terraform lock" messages
- Verify timeout values are being loaded correctly
- Monitor duration of operations

---

## Testing Patterns

### Comprehensive Validation Standards

**CRITICAL**: Go tests must validate **"does it respond correctly?"** not just **"does it respond?"**

- ✅ Validate complete response structure
- ✅ Validate all nested objects and fields
- ✅ Validate exact field values match expectations
- ✅ Validate data types (boolean vs string, null handling)
- ✅ Validate error messages, not just status codes
- ❌ Only validate status code
- ❌ Only check that a field exists, not its value

**See**: [`TESTING_STANDARDS.md`](./TESTING_STANDARDS.md) for complete testing guidelines, examples, and checklists.

### Feature Parity Testing

**Reference**: [`TEST_PARITY_ANALYSIS.md`](./TEST_PARITY_ANALYSIS.md) - Detailed comparison of Python vs Go test coverage

When writing tests:
1. Find the corresponding Python test
2. Match the validation depth of the Python test
3. Ensure all assertions from Python are represented in Go
4. Add Python reference comments
5. Document any parity gaps

### Test Organization

```
Unit Tests (co-located)
├── internal/domain/auth/service/*_test.go         # Beside source files
├── internal/infrastructure/auth/*_test.go         # Beside source files
└── internal/application/command/*_test.go         # Beside source files

Integration Tests (separate directory)
├── test/integration/domain/                      # Domain integration
├── test/integration/handler/                     # HTTP endpoint tests
├── test/integration/auth_integration_test.go     # Auth flows
└── test/integration/complete_workflow_test.go    # E2E workflows

Test Utilities
├── test/integration/testutils/database.go         # DB setup helpers
├── test/integration/testutils/python_test_data.go # Python test data mirroring
└── test/testutils/mocks/                         # Repository mocks
```

### Test Data Standards

Use comprehensive test data helpers that mirror Python's `test_data.py`:

```go
// For modules - use fully populated test data
_, moduleProvider, _ := testutils.SetupFullyPopulatedModule(t, db)

// For providers - include repository information
provider, repository, version := testutils.CreateProviderVersionWithRepository(
    t, db, namespaceID, "test-provider", "1.0.0", "v1.0.0",
    &description, sqldb.ProviderTierCommunity, gpgKeyID, nil,
)
```

### Table-Driven Tests (Primary Pattern)

```go
func TestParseVersion(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        want    *Version
        wantErr bool
    }{
        {
            name:  "valid semantic version",
            input: "1.2.3",
            want:  &Version{Major: 1, Minor: 2, Patch: 3},
        },
        {
            name:    "invalid version",
            input:   "not.a.version",
            wantErr: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := ParseVersion(tt.input)
            if (err != nil) != tt.wantErr {
                t.Errorf("ParseVersion() error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            if !reflect.DeepEqual(got, tt.want) {
                t.Errorf("ParseVersion() = %v, want %v", got, tt.want)
            }
        })
    }
}
```

### Integration Test Setup

```go
func TestModulePublishing(t *testing.T) {
    // Setup test database
    db := testutils.SetupTestDatabase(t)
    defer cleanupTestDatabase(t, db)

    // Create container with all services
    container := testutils.CreateTestContainer(t, db)

    // Create test request
    req := httptest.NewRequest("POST", "/v1/modules/...", body)

    // Execute
    w := httptest.NewRecorder()
    handler.HandlePublishModuleVersion(w, req)

    // Assert
    assert.Equal(t, http.StatusCreated, w.Code)
}
```

### Container vs TestContainer - Critical Testing Patterns

**⚠️ CRITICAL**: Understanding the difference between `Container` and `TestContainer` is essential for writing correct integration tests, especially for authentication tests.

#### Types

| Type | Location | Purpose |
|------|----------|---------|
| `*container.Container` | `internal/infrastructure/container/` | DI container with all services/repositories |
| `*testutils.TestContainer` | `test/integration/testutils/database.go` | Test helper that wraps `Container` + adds `Router` |

#### Structure

```go
// internal/infrastructure/container/container.go
type Container struct {
    // Configuration
    DomainConfig  *domainConfig.DomainConfig
    InfraConfig   *config.InfrastructureConfig
    ConfigService *configService.ConfigurationService
    Logger        zerolog.Logger
    DB            *sqldb.Database

    // Repositories (20+ repositories)
    NamespaceRepo   ...
    ModuleProviderRepo ...
    SessionRepo     ...
    // ... etc

    // Services
    Server          *Server
    AuthFactory     *auth.AuthFactory
    SessionService  *auth.SessionManagementService
    // ... etc
}
```

```go
// test/integration/testutils/database.go
type TestContainer struct {
    *Container           // Embedded: has ALL dependencies
    Router    http.Handler // Added: HTTP router for testing
}
```

#### Usage Patterns

**❌ WRONG - Creating multiple containers in a test:**
```go
func TestAuthBroken(t *testing.T) {
    db := testutils.SetupTestDatabase(t)
    defer testutils.CleanupTestDatabase(t, db)

    // Create namespace using one container
    _ = testutils.CreateNamespace(t, db, "test-ns", nil)

    // Create auth helper with one container
    cont1 := testutils.CreateTestServer(t, db)
    authHelper := testutils.NewAuthHelper(t, db, cont1)
    cookie := authHelper.CreateSessionForUser("user", false, []string{}, nil)

    // ❌ WRONG: Create ANOTHER container for HTTP requests!
    // This has DIFFERENT services (different signing keys, sessions, etc.)
    cont2 := testutils.CreateTestContainer(t, db)
    router := cont2.Server.Router()

    // The cookie was encrypted with cont1's signing key
    // But router uses cont2's signing key
    // Result: Authentication fails!
}
```

**✅ CORRECT - Using the same container throughout:**
```go
func TestAuthCorrect(t *testing.T) {
    db := testutils.SetupTestDatabase(t)
    defer testutils.CleanupTestDatabase(t, db)

    // Create namespace using test helpers (they use the db directly)
    _ = testutils.CreateNamespace(t, db, "test-ns", nil)

    // ✅ Create ONE container and use it throughout
    cont := testutils.CreateTestServer(t, db)  // Returns *TestContainer
    authHelper := testutils.NewAuthHelper(t, db, cont)  // Needs *TestContainer

    // Create session using same container's CookieService
    cookie := authHelper.CreateSessionForUser("user", false, []string{}, nil)

    // ✅ Use same container's router for HTTP requests
    router := cont.Router

    // Now auth works - same signing keys, same session service
    req := httptest.NewRequest("GET", "/v1/modules/...", nil)
    req.Header.Set("Cookie", cookie)

    w := httptest.NewRecorder()
    router.ServeHTTP(w, req)
}
```

#### When to Use Each Helper

| Helper | Returns | Use When |
|--------|--------|----------|
| `CreateTestContainer(t, db)` | `*container.Container` | Need services but won't make HTTP requests |
| `CreateTestServer(t, db)` | `*testutils.TestContainer` | **DEFAULT** - Need HTTP routing + auth |
| `CreateTestContainerWithConfig(t, db, opts...)` | `*container.Container` | Need custom config (OIDC, etc.) |

**For tests with authentication, ALWAYS use `CreateTestServer()`** (or wrap `CreateTestContainerWithConfig()` in a `TestContainer` struct):

```go
// For tests with Terraform IDP (needs custom signing key)
func setupTestContainerWithSigningKey(t *testing.T, db *sqldb.Database, authMethodName string) *testutils.TestContainer {
    if strings.HasPrefix(authMethodName, "terraform_idp") {
        keyPath, _ := testutils.CreateTestTerraformOIDCSigningKey(t)
        cont := testutils.CreateTestContainerWithConfig(t, db, testutils.WithTerraformOIDCConfig(keyPath))
        return &testutils.TestContainer{
            Container: cont,
            Router:    cont.Server.Router(),
        }
    }
    return testutils.CreateTestServer(t, db)
}
```

#### Key Rules

1. **Never create multiple containers in one test** - different signing keys, sessions, etc.
2. **AuthHelper needs `*TestContainer`** - not `*Container`, not empty struct
3. **Use `TestContainer.Router`** - not `Container.Server.Router()`
4. **Share the container** - between auth helper setup and HTTP requests

### Test Utilities

**Database Setup**:
```go
db := testutils.SetupTestDatabase(t)
config := testutils.CreateTestDomainConfig(t)
container := testutils.CreateTestContainer(t, db)
```

**HTTP Test Helpers**:
```go
testutils.AssertHTTPError(t, w, 400, "Invalid input")
testutils.AssertJSONSuccess(t, w)
testutils.MakeAuthenticatedRequest(t, handler, "POST", url, body, authConfig)
```

**Mock Servers**:
```go
mockOIDC := testutils.NewMockOIDCServer(t)
mockSAML := testutils.NewMockSAMLServer(t)
defer mockOIDC.Close()
```

### Running Tests

```bash
# All tests
go test ./... -v

# Specific package
go test ./internal/domain/auth/... -v

# With coverage
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

### When to Write Tests

#### 1. New Domain Logic

**Trigger**: Adding new business rules or domain services

**Tests to Write**:
- Unit tests for domain entities
- Service tests for business logic
- Integration tests for complete use cases

#### 2. New Repository Methods

**Trigger**: Adding new database operations

**Tests to Write**:
- Unit tests with mock database
- Integration tests with test database
- Performance tests for complex queries

#### 3. New API Endpoints

**Trigger**: Adding new HTTP routes

**Tests to Write**:
- Unit tests for handlers
- Integration tests for API contracts
- Authentication/authorization tests

#### 4. Bug Fixes

**Trigger**: Fixing defects

**Tests to Write**:
- Regression tests that reproduce the bug
- Verify the fix works
- Ensure no side effects

#### 5. Configuration Changes

**Trigger**: Adding new configuration options

**Tests to Write**:
- Configuration parsing tests
- Default value verification tests
- Integration tests with different configurations

---

## Nil Checking Conventions

### Core Principle

**Constructor validation establishes invariants - methods trust them.**

All nil checking happens in constructors (`NewX()` functions). Once an object is created, its methods assume all required dependencies are non-nil and do NOT re-check.

### Service Struct Rules

#### 1. Every Service Struct MUST Have a Constructor

```go
// ❌ BAD: Direct struct creation
service := &MyService{repo: someRepo}

// ✅ GOOD: Use constructor
service, err := NewMyService(someRepo)
```

#### 2. Constructor Signature Returns Error

```go
func NewX(required1 *Type1, required2 *Type2) (*X, error) {
	if required1 == nil {
		return nil, fmt.Errorf("required1 cannot be nil")
	}
	if required2 == nil {
		return nil, fmt.Errorf("required2 cannot be nil")
	}
	return &X{required1: required1, required2: required2}, nil
}
```

#### 3. Struct Fields Must Document Nil Usage

```go
type MyService struct {
	// requiredDep is required for all operations (never nil after construction)
	requiredDep *Dependency

	// optionalDep may be nil for read-only operations (check before use)
	optionalDep *OptionalDependency

	// config holds shared configuration (never nil after construction)
	config *Config
}
```

#### 4. Methods DO NOT Re-Check Required Dependencies

```go
// ✅ GOOD: Trust the constructor invariant
func (s *MyService) DoWork() error {
	return s.requiredDep.Perform() // No nil check needed
}

// ❌ BAD: Redundant nil check
func (s *MyService) DoWork() error {
	if s.requiredDep == nil { // This check is redundant
		return fmt.Errorf("requiredDep is nil")
	}
	return s.requiredDep.Perform()
}
```

#### 5. Optional Fields Must Be Checked With Comments

```go
func (s *MyService) ReadOnlyOperation() error {
	// optionalDep is only needed for write operations
	if s.optionalDep != nil {
		return s.optionalDep.Write()
	}
	return nil // Read-only, optionalDep not needed
}
```

### Non-Service Functions Taking Pointers

Functions that accept pointer parameters must either:
1. **Nil-check at the start**, OR
2. **Document the valid nil use-case with a comment**

```go
// ✅ GOOD: Nil check with clear error
func ProcessModule(ctx context.Context, module *Module) error {
	if module == nil {
		return fmt.Errorf("module cannot be nil")
	}
	// ... process module
}

// ✅ GOOD: Documented nil usage
func UpdateMetadata(module *Module) error {
	// module may be nil - if nil, clears metadata
	if module == nil {
		return clearMetadata()
	}
	return setMetadata(module)
}

// ❌ BAD: No nil check, no documentation
func ProcessModule(ctx context.Context, module *Module) error {
	module.ID // Will panic if module is nil!
}
```

### Implementation Examples

#### Example 1: Simple Service

```go
type ProviderService struct {
	// providerRepo handles provider persistence (required)
	providerRepo repository.ProviderRepository
}

func NewProviderService(providerRepo repository.ProviderRepository) (*ProviderService, error) {
	if providerRepo == nil {
		return nil, fmt.Errorf("providerRepo cannot be nil")
	}
	return &ProviderService{providerRepo: providerRepo}, nil
}

// No nil check needed - constructor guarantees providerRepo is non-nil
func (s *ProviderService) GetByID(id int) (*Provider, error) {
	return s.providerRepo.FindByID(context.Background(), id)
}
```

#### Example 2: Service with Optional Dependency

```go
type ModuleProcessorService struct {
	// parser is required for all operations (never nil)
	parser ModuleParser

	// archiveGenerator is optional - nil means skip archive generation
	archiveGenerator ArchiveGenerator
}

func NewModuleProcessorService(
	parser ModuleParser,
	archiveGenerator ArchiveGenerator,
) (*ModuleProcessorService, error) {
	if parser == nil {
		return nil, fmt.Errorf("parser cannot be nil")
	}
	// archiveGenerator can be nil - documented in struct field comment
	return &ModuleProcessorService{
		parser: parser,
		archiveGenerator: archiveGenerator,
	}, nil
}

func (s *ModuleProcessorService) Process(module *Module) error {
	// parser is never nil - no check needed
	parsed, err := s.parser.Parse(module)
	if err != nil {
		return err
	}

	// archiveGenerator may be nil - check before use
	if s.archiveGenerator != nil {
		return s.archiveGenerator.Generate(parsed)
	}
	return nil
}
```

### Container Initialization

When initializing services in the container, handle constructor errors:

```go
// In container.go
// SessionManagementService returns nil if SECRET_KEY is empty (login methods disabled)
sessionManagementService := authservice.NewSessionManagementService(sessionService, cookieService)
// No error returned - nil is valid when SECRET_KEY not configured
c.SessionManagementService = sessionManagementService
```

### Testing Nil Checks

Every constructor MUST have tests verifying nil parameter rejection:

```go
func TestNewMyService_NilRequiredDep(t *testing.T) {
	service, err := NewMyService(nil, &validOtherDep{})
	if err == nil {
		t.Error("Expected error when requiredDep is nil")
	}
	if service != nil {
		t.Error("Expected nil service when requiredDep is nil")
	}
}

func TestNewMyService_ValidDependencies(t *testing.T) {
	service, err := NewMyService(&validDep{}, &validOtherDep{})
	if err != nil {
		t.Errorf("Expected no error with valid dependencies: %v", err)
	}
	if service == nil {
		t.Error("Expected service when dependencies are valid")
	}
}
```

### Summary

| What | Rule |
|------|------|
| **Service structs** | MUST have `NewX()` constructor |
| **Constructor** | MUST check all pointer params and return error on nil |
| **Struct fields** | MUST have comments documenting nil usage |
| **Required fields** | Methods assume non-nil, no re-checks needed |
| **Optional fields** | Document when nil is valid, check before use |
| **Other functions** | Either nil-check OR document valid nil use-case |

---

## Critical Constraints

### Python Parity (ABSOLUTE REQUIREMENT)

**This is a refactoring, NOT a feature change:**

- ✅ API must match Python exactly
- ✅ Database schema must be identical
- ✅ Business logic must produce same results
- ✅ Error handling must match Python behavior
- ✅ Configuration comes from environment variables

### Delete-Then-Create Pattern

Python deletes and recreates records during re-indexing. Go must follow this:

```go
func (s *Service) ReimportModuleVersion(ctx context.Context, req *Request) error {
    // 1. Delete existing version
    existing, _ := s.repo.FindByVersion(ctx, req.Version)
    if existing != nil {
        s.repo.Delete(ctx, existing.ID())
    }

    // 2. Create NEW version (ID=0 forces Create not Update)
    newVersion := model.NewModuleVersion(req.Version, nil, false)

    // 3. Save
    return s.repo.Save(ctx, newVersion)
}
```

### Context Propagation (MANDATORY)

**Every database operation requires context:**

```go
// ✅ CORRECT - Context propagated
func (s *Service) MethodA(ctx context.Context) error {
    return s.repo.MethodB(ctx)  // Context passed through
}

// ❌ WRONG - No context
func (s *Service) MethodA() error {
    return s.repo.MethodB()  // Missing context
}
```

### Configuration Defaults (Single Source)

**All defaults defined in domain config models:**

```go
// ✅ CORRECT - In domain config
ExampleFileExtensions []string `env:"EXAMPLE_FILE_EXTENSIONS" envDefault:"tf,tfvars,sh,json"`

// ❌ WRONG - Hardcoded in service
extensions := []string{"tf", "tfvars", "sh", "json"}
```

---

## Common Pitfalls

### 1. Direct Database Access

```go
// ❌ WRONG
func (s *Service) BadPattern() error {
    return s.db.Raw("SELECT * FROM module_version").Error
}

// ✅ CORRECT
func (s *Service) GoodPattern(ctx context.Context) ([]*ModuleVersion, error) {
    return s.repo.FindAll(ctx)
}
```

### 2. Missing Context

```go
// ❌ WRONG
func (h *Handler) HandleRequest(w http.ResponseWriter, r *http.Request) {
    h.service.DoSomething()  // No context
}

// ✅ CORRECT
func (h *Handler) HandleRequest(w http.ResponseWriter, r *http.Request) {
    h.service.DoSomething(r.Context())  // Context passed
}
```

### 3. Hardcoded Configuration

```go
// ❌ WRONG
extensions := []string{"tf", "tfvars", "sh"}

// ✅ CORRECT
extensions := s.config.ModuleProcessing.ExampleFileExtensions
```

### 4. Business Logic in Infrastructure

```go
// ❌ WRONG - HTTP in domain service
func (s *DomainService) DoSomething() error {
    resp, err := http.Get("https://example.com")
    return err
}

// ✅ CORRECT - Pure business logic
func (s *DomainService) DoSomething(data *Data) (*Result, error) {
    return s.processData(data), nil
}
```

### 5. Domain Model Relationship Loss

```go
// ❌ BAD: Loading entities without relationships
moduleVersion, err := repo.FindByID(id) // Loses ModuleProvider()

// ✅ GOOD: Enhanced repository implementation automatically restores relationships
moduleVersion, err := repo.FindByID(id) // Has ModuleProvider() intact
```

### 6. Transaction Mismanagement

```go
// ❌ BAD: Creating transactions in repositories
func (r *repo) Save(ctx context.Context, entity *Entity) error {
    tx := r.db.Begin() // Wrong: Repository shouldn't manage transactions
}

// ✅ GOOD: Use existing transaction context
func (r *repo) Save(ctx context.Context, entity *Entity) error {
    db := r.getDBFromContext(ctx) // Correct: Use existing transaction
}
```

### 7. Permission Bypass

```go
// ❌ BAD: Generic authentication on sensitive endpoints
r.With(auth.RequireAuth).Post("/modules/upload") // Too permissive

// ✅ GOOD: Specific permission checks
r.With(auth.RequireUploadPermission("{namespace}")).Post("/modules/{namespace}/upload")
```

---

## Key Files to Understand

### Critical Implementation Files

| File | Purpose |
|------|---------|
| `/internal/domain/config/model/config.go` | Configuration models |
| `/internal/infrastructure/persistence/sqldb/repository/base_repository.go` | Base repository pattern |
| `/internal/infrastructure/parser/module_parser_impl.go` | Module parsing logic |
| `/internal/domain/module/service/module_processor_service_impl.go` | Content extraction |
| `/internal/infrastructure/container/container.go` | Dependency injection |
| `/internal/domain/auth/service/authentication_service.go` | Auth orchestration |
| `/internal/interfaces/http/server.go` | HTTP routing and middleware |
| `/test/integration/testutils/` | Test utilities and helpers |

### Current Module Processing Flow

1. `ImportModuleVersionCommand` (Application)
2. `ModuleImporterService` (Domain)
3. `TransactionProcessingOrchestrator` (Domain)
4. `ModuleCreationWrapperService` (Domain)
5. `ModuleProcessorService` (Domain)
6. `ModuleParser` (Infrastructure) → Extract content
7. `ModuleDetailsRepository` + `ModuleVersionRepository` (Infrastructure) → Persist

### Development Guidelines

#### When Adding New Features

1. **Domain First**: Define domain models and interfaces in `/internal/domain/`
2. **Repository Pattern**: Implement interfaces in `/internal/infrastructure/persistence/`
3. **Configuration**: Add config to DomainConfig models, use throughout
4. **Transactions**: Always use `GetDBFromContext(ctx)` for database operations

#### When Modifying Existing Code

1. **Context Propagation**: Ensure any new methods accept and propagate context
2. **Repository Usage**: Use BaseRepository pattern for new repositories
3. **Configuration**: Use DomainConfig instead of hardcoded values
4. **DDD Compliance**: Maintain domain/infrastructure separation

### Additional Resources

#### Key Documentation Files

- `TESTING_STANDARDS.md` - Comprehensive testing guidelines, standards, and examples
- `TEST_PARITY_ANALYSIS.md` - Python vs Go test parity analysis and gap documentation
- `GOLANG_DEVELOPMENT_PATTERNS.md` - Development patterns and anti-patterns
- `AUTHENTICATION_ARCHITECTURE.md` - Authentication system details
- `CONFIG_ARCHITECTURE.md` - Configuration management patterns
- `AI_DEVELOPMENT_GUIDE.md` - Architecture overview

#### Python Reference

When implementing features, reference:
- `/terrareg/server/api/` - Python Flask routes
- `/terrareg/server/model_sqlalchemy.py` - Python database models
- `/terrareg/terrareg/` - Python business logic