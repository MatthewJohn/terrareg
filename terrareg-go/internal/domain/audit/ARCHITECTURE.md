# Audit Domain Architecture

## Overview

The Audit domain provides comprehensive audit trail functionality for all system changes and user actions in Terrareg. It maintains a complete history of modifications across all entities including modules, providers, namespaces, authentication events, and more.

---

## Core Functionality

The audit domain provides the following capabilities:

- **Event Logging** - Records all changes to system entities
- **Search and Filtering** - Queries audit history with pagination and text search
- **Entity-Specific Auditing** - Specialized services for different entity types
- **Metadata Tracking** - Records user context, timestamps, and change details

---

## Domain Components

### Models

**Location**: `/internal/domain/audit/model/`

#### AuditHistory Model

The core domain model representing an audit entry:

```go
type AuditHistory struct {
    ID          int       `json:"id"`
    EntityType  string    `json:"entity_type"`   // e.g., "module", "provider", "namespace"
    EntityID    int       `json:"entity_id"`
    ActionType  string    `json:"action_type"`    // e.g., "create", "update", "delete"
    Username    string    `json:"username"`
    OldValue    string    `json:"old_value"`
    NewValue    string    `json:"new_value"`
    ParentID    *int      `json:"parent_id,omitempty"`
    Timestamp   time.Time `json:"timestamp"`
}
```

#### AuditHistorySearchQuery

Query model for searching audit history:

```go
type AuditHistorySearchQuery struct {
    EntityType  string
    EntityID    int
    ActionType  string
    Username    string
    SearchValue string
    Limit       int
    Offset      int
}
```

### Repository

**Location**: `/internal/domain/audit/repository/`

The `AuditHistoryRepository` interface defines data access operations:

```go
type AuditHistoryRepository interface {
    Create(ctx context.Context, audit *model.AuditHistory) error
    Search(ctx context.Context, query model.AuditHistorySearchQuery) (*model.AuditHistorySearchResult, error)
    GetTotalCount(ctx context.Context) (int, error)
    GetFilteredCount(ctx context.Context, searchValue string) (int, error)
}
```

### Services

**Location**: `/internal/domain/audit/service/`

#### Core AuditService

The main service for audit operations:

```go
type AuditService struct {
    auditRepo repository.AuditHistoryRepository
}
```

**Methods**:
- `LogEvent(ctx, audit)` - Records an audit event
- `SearchHistory(ctx, query)` - Retrieves paginated audit history
- `GetTotalCount(ctx)` - Returns total number of audit entries
- `GetFilteredCount(ctx, searchValue)` - Returns count matching search criteria

#### Specialized Audit Services

Each major entity type has a specialized audit service that provides convenience methods:

| Service | Purpose |
|---------|---------|
| `ModuleAuditService` | Audit logging for module providers and versions |
| `NamespaceAuditService` | Audit logging for namespace operations |
| `ProviderAuditService` | Audit logging for Terraform providers |
| `RepositoryAuditService` | Audit logging for Git repository operations |
| `GpgKeyAuditService` | Audit logging for GPG key management |
| `AuthenticationAuditService` | Audit logging for authentication events |
| `UserGroupAuditService` | Audit logging for user group changes |

---

## Dependencies

### Domain Dependencies

The audit domain has **no dependencies** on other domains - it is a pure cross-cutting concern that records events from all other domains.

### Domains That Depend on Audit

| Domain | Audit Integration |
|--------|-------------------|
| **module** | ModuleAuditService logs all module changes |
| **provider** | ProviderAuditService logs provider operations |
| **namespace** | NamespaceAuditService logs namespace CRUD |
| **gpgkey** | GpgKeyAuditService logs key management |
| **auth** | AuthenticationAuditService logs auth events |
| **repository** | RepositoryAuditService logs Git repository operations |
| **user_groups** | UserGroupAuditService logs group changes |

### Infrastructure Dependencies

| Component | Purpose |
|-----------|---------|
| **Database** | Persistent storage for audit records |
| **Middleware** | HTTP middleware for request-level audit logging |

---

## Key Design Principles

1. **Immutability** - Audit records are never modified once created
2. **Completeness** - All significant state changes are logged
3. **Performance** - Async logging to avoid blocking main operations
4. **Traceability** - Links to parent entities for full change history
5. **Searchability** - Full-text search and filtering capabilities

---

## Audit Event Types

### Standard Action Types

| Action | Description |
|--------|-------------|
| `create` | Entity creation |
| `update` | Entity modification |
| `delete` | Entity deletion |
| `publish` | Module/provider version publishing |
| `import` | Module import from Git |
| `upload` | File upload operations |
| `login` | User authentication |
| `logout` | User logout |
| `authorize` | Authorization decision |

### Entity Types

| Type | Description |
|------|-------------|
| `module_provider` | Module provider entities |
| `module_version` | Module version entities |
| `namespace` | Namespace entities |
| `provider` | Terraform providers |
| `provider_version` | Provider version entities |
| `gpg_key` | GPG key entities |
| `authentication_token` | Auth session tokens |
| `user_group` | User group entities |
| `repository` | Git repository links |

---

## Usage Examples

### Logging an Audit Event

```go
audit := &model.AuditHistory{
    EntityType: "module_provider",
    EntityID:   moduleProvider.ID(),
    ActionType: "update",
    Username:   authCtx.Username,
    OldValue:   oldDescription,
    NewValue:   newDescription,
    Timestamp:  time.Now(),
}
err := auditService.LogEvent(ctx, audit)
```

### Searching Audit History

```go
query := model.AuditHistorySearchQuery{
    EntityType:  "module_provider",
    EntityID:    moduleProvider.ID(),
    SearchValue: "description",
    Limit:       50,
    Offset:      0,
}
result, err := auditService.SearchHistory(ctx, query)
```

### Using Specialized Audit Service

```go
// Module-specific audit logging
err := moduleAuditService.LogModuleUpdate(ctx, moduleProvider, oldDetails, newDetails, username)
```

---

## Integration Patterns

### Service Layer Integration

Services in other domains inject audit services and log events:

```go
type ModuleService struct {
    // ... other dependencies
    auditService *audit.ModuleAuditService
}

func (s *ModuleService) UpdateDescription(ctx context.Context, id int, desc string) error {
    // Perform update
    // ...

    // Log audit event
    s.auditService.LogModuleUpdate(ctx, module, oldDesc, desc, username)
}
```

### HTTP Middleware

Request-level audit logging can be performed via middleware:

```go
func AuditMiddleware(auditService *audit.AuditService) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            // Process request
            next.ServeHTTP(w, r)

            // Log audit event for mutating operations
            if r.Method != "GET" && r.Method != "HEAD" {
                auditService.LogHTTPRequest(ctx, r)
            }
        })
    }
}
```
