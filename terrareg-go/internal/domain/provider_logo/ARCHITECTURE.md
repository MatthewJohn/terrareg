# Provider Logo Domain Architecture

## Overview

The Provider Logo domain manages provider logo images. It handles logo storage, retrieval, and validation for Terraform providers displayed in the registry UI.

---

## Core Functionality

The provider_logo domain provides the following capabilities:

- **Logo Storage** - Store logo images for providers
- **Logo Retrieval** - Retrieve logos by provider
- **Image Validation** - Validate image types and sizes
- **Logo Deletion** - Remove logos when providers are deleted

---

## Domain Components

### Models

**Location**: `/internal/domain/provider_logo/model/provider_logo.go`

#### ProviderLogo Model

```go
type ProviderLogo struct {
    id           int
    providerID   int
    logoPath     string
    contentType  string
    size         int64
    createdAt    time.Time
}
```

**Key Fields**:
- `providerID` - Links logo to a provider
- `logoPath` - Storage path for the logo image
- `contentType` - MIME type (e.g., "image/png")
- `size` - File size in bytes

#### Logo Types

**Location**: `/internal/domain/provider_logo/model/types.go`

Supported image types:
- PNG (`image/png`)
- JPEG (`image/jpeg`)
- GIF (`image/gif`)
- SVG (`image/svg+xml`)
- WebP (`image/webp`)

### Repository Interface

**Location**: `/internal/domain/provider_logo/repository/provider_logo_repository.go`

```go
type ProviderLogoRepository interface {
    Save(ctx context.Context, logo *ProviderLogo) error
    FindByProviderID(ctx context.Context, providerID int) (*ProviderLogo, error)
    DeleteByProviderID(ctx context.Context, providerID int) error
    ExistsByProviderID(ctx context.Context, providerID int) (bool, error)
}
```

---

## Dependencies

### Domain Dependencies

| Domain | Purpose |
|--------|---------|
| **provider** | For provider association |

### Infrastructure Dependencies

| Component | Purpose |
|-----------|---------|
| **Storage** | File system or S3 storage for logo images |
| **Database** | Persistent storage for logo metadata |

---

## Key Design Principles

1. **Provider Association** - Each logo is associated with exactly one provider
2. **Storage Abstraction** - Logo files stored via storage service (filesystem or S3)
3. **Type Validation** - Only accepted image types allowed
4. **Cascade Deletion** - Logos deleted when providers are deleted

---

## Logo Lifecycle

### Upload Flow

```
1. Validate provider exists
2. Validate image type
3. Validate image size
4. Generate storage path
5. Store image file via storage service
6. Create ProviderLogo entity
7. Save to repository
```

### Retrieval Flow

```
1. Query by provider ID
2. Get logo metadata
3. Read image file from storage
4. Return image with content type
```

### Deletion Flow

```
1. Delete logo file from storage
2. Delete logo metadata from repository
```

---

## Storage Path Structure

```
{data_directory}/providers/logos/{provider_id}.{extension}
```

Example: `/data/providers/logos/123.png`

---

## Usage Examples

### Saving a Logo

```go
logo := &provider_logo.ProviderLogo{
    ProviderID:  providerID,
    LogoPath:    "/providers/logos/123.png",
    ContentType: "image/png",
    Size:        12345,
}

err := logoRepo.Save(ctx, logo)
```

### Retrieving a Logo

```go
logo, err := logoRepo.FindByProviderID(ctx, providerID)
if err != nil {
    // Handle not found
}

// Read image from storage
imageData, err := storageService.ReadFile(ctx, logo.LogoPath, true)
```

### Deleting a Logo

```go
// Delete from storage first
err := storageService.DeleteFile(ctx, logo.LogoPath)

// Then delete metadata
err = logoRepo.DeleteByProviderID(ctx, providerID)
```

---

## Image Validation

### Supported Content Types

```go
var allowedContentTypes = map[string]bool{
    "image/png":    true,
    "image/jpeg":   true,
    "image/gif":    true,
    "image/svg+xml": true,
    "image/webp":   true,
}
```

### Size Limits

- **Minimum**: 100 bytes
- **Maximum**: 5 MB (configurable)

### Dimension Recommendations

- **Recommended**: 256x256 pixels
- **Minimum**: 64x64 pixels
- **Maximum**: 1024x1024 pixels

---

## References

- [`/internal/domain/provider_logo/model/`](./model/) - Logo models
- [`/internal/domain/provider_logo/repository/`](./repository/) - Logo repository
- [`/internal/domain/provider/`](../provider/) - Provider domain
- [`/internal/domain/storage/`](../storage/) - Storage service
