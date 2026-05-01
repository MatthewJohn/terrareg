# Analytics Domain Architecture

## Overview

The Analytics domain tracks module downloads and usage statistics across the Terrareg registry. It provides a simple but essential service for monitoring module popularity and usage patterns.

---

## Core Functionality

The analytics domain provides the following capabilities:

- **Download Tracking** - Records and increments download counts for module versions
- **Statistics Retrieval** - Queries download counts for specific modules or versions
- **Usage Analytics** - Supports analytics token-based tracking for external applications

---

## Domain Components

### Models

**Location**: `/internal/domain/analytics/model/`

The analytics domain is lightweight and primarily uses repository interfaces defined in the application layer. The domain model is minimal as analytics data is primarily counters.

### Repository

**Location**: `/internal/domain/analytics/repository/`

The `AnalyticsRepository` interface defines data access operations:

```go
type AnalyticsRepository interface {
    GetDownloadsByVersionID(ctx context.Context, moduleVersionID int) (int, error)
    IncrementDownloadCount(ctx context.Context, moduleVersionID int) error
}
```

**Note**: The interface is actually defined in the application layer (`internal/application/command/analytics`) to maintain clean architecture boundaries.

### Service

**Location**: `/internal/domain/analytics/service/analytics_service.go`

The `AnalyticsService` provides domain logic for analytics operations:

```go
type AnalyticsService struct {
    analyticsRepository analyticsCmd.AnalyticsRepository
}
```

**Methods**:
- `GetDownloadsByVersionID(ctx, moduleVersionID)` - Retrieves download count for a specific module version

---

## Dependencies

### Domain Dependencies

| Domain | Purpose |
|--------|---------|
| **module** | Provides module version IDs for analytics tracking |

### Infrastructure Dependencies

| Component | Purpose |
|-----------|---------|
| **Database** | Persistent storage for download counters |
| **Application Layer** | Repository interfaces defined in command layer |

---

## Key Design Principles

1. **Simplicity** - Analytics is a read-heavy domain with simple counter operations
2. **Performance** - Download counting must be fast and not block module downloads
3. **Loose Coupling** - Uses repository interfaces from application layer for clean architecture
4. **Configuration-Driven** - Analytics behavior controlled by `DomainConfig` settings

---

## Analytics Configuration

Analytics behavior is controlled via configuration in the `config` domain:

```go
type DomainConfig struct {
    // Analytics settings
    AnalyticsTokenPhrase      string  // Token phrase for analytics tracking
    AnalyticsTokenDescription string  // Description for analytics token
    ExampleAnalyticsToken     string  // Default analytics token
    DisableAnalytics          bool    // Global analytics disable flag
}
```

---

## Usage Examples

### Recording a Download

When a module version is downloaded:

```go
// In the download handler
err := analyticsRepository.IncrementDownloadCount(ctx, moduleVersionID)
```

### Retrieving Download Statistics

```go
// Get download count for display
count, err := analyticsService.GetDownloadsByVersionID(ctx, moduleVersionID)
```

---

## Integration Points

### Module Downloads

Analytics are automatically updated when:
- Module version downloads occur via the Terraform protocol
- Module archive downloads occur via HTTP endpoints
- Provider binary downloads occur

### Analytics Tokens

The system supports analytics token-based tracking to differentiate usage by external applications. See the `config` domain for analytics token configuration.
