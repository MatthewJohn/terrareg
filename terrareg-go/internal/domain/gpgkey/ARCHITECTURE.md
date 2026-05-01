# GPG Key Domain Architecture

## Overview

The GPG Key domain manages GPG keys for provider signing and verification. It provides functionality for creating, storing, and managing GPG keys that are used to sign Terraform provider binaries.

---

## Core Functionality

The gpgkey domain provides the following capabilities:

- **GPG Key Creation** - Import and validate ASCII-armored GPG keys
- **Key Validation** - Parse and validate GPG key structure
- **Key Information Extraction** - Extract key ID and fingerprint from ASCII armor
- **Duplicate Detection** - Prevent duplicate fingerprints across namespaces
- **Usage Checking** - Verify if a key is in use before deletion
- **Signature Verification** - Verify signatures using stored GPG keys

---

## Domain Components

### Models

**Location**: `/internal/domain/gpgkey/model/gpg_key.go`

#### GPGKey Model

```go
type GPGKey struct {
    id              int
    namespaceID     int
    namespace       *Namespace
    asciiArmor      string
    keyID           string
    fingerprint     string
    trustSignature  *string
    source          *string
    sourceURL       *string
    createdAt       time.Time
}
```

**Key Fields**:
- `namespaceID` - Links key to a namespace
- `asciiArmor` - The full ASCII-armored GPG key
- `keyID` - Extracted from the key (e.g., "ABCD1234")
- `fingerprint` - Full key fingerprint for deduplication
- `trustSignature` - Optional trust signature
- `source` - Optional source description
- `sourceURL` - Optional source URL

#### Namespace Value Object

```go
type Namespace struct {
    id   int
    name string
}
```

### Repository Interface

**Location**: `/internal/domain/gpgkey/repository/gpg_key_repository.go`

```go
type GPGKeyRepository interface {
    Save(ctx context.Context, key *GPGKey) error
    FindByNamespaceAndKeyID(ctx context.Context, namespace, keyID string) (*GPGKey, error)
    FindByKeyID(ctx context.Context, keyID string) (*GPGKey, error)
    FindByNamespace(ctx context.Context, namespace string) ([]*GPGKey, error)
    FindMultipleByNamespaces(ctx context.Context, namespaces []string) ([]*GPGKey, error)
    ExistsByFingerprint(ctx context.Context, fingerprint string) (bool, error)
    IsInUse(ctx context.Context, keyID string) (bool, error)
    DeleteByNamespaceAndKeyID(ctx context.Context, namespace, keyID string) error
}
```

### Service

**Location**: `/internal/domain/gpgkey/service/gpg_key_service.go`

```go
type GPGKeyService struct {
    gpgKeyRepo    gpgkeyRepo.GPGKeyRepository
    namespaceRepo moduleRepo.NamespaceRepository
}
```

**Key Methods**:
- `CreateGPGKey(ctx, req)` - Create a new GPG key
- `GetNamespaceGPGKeys(ctx, namespace)` - Get all keys for a namespace
- `GetMultipleNamespaceGPGKeys(ctx, namespaces)` - Get keys for multiple namespaces
- `GetGPGKey(ctx, namespace, keyID)` - Get a specific key
- `DeleteGPGKey(ctx, namespace, keyID)` - Delete a key (checks usage first)
- `VerifySignature(ctx, keyID, signature, data)` - Verify a signature

---

## Dependencies

### Domain Dependencies

| Domain | Purpose |
|--------|---------|
| **module** | For namespace repository (to validate namespace exists) |

### Infrastructure Dependencies

| Component | Purpose |
|-----------|---------|
| **infrastructure/gpg** | GPG parsing and verification utilities |
| **Database** | Persistent storage for GPG keys |

### Domains That Depend on GPG Key

| Domain | Purpose |
|--------|---------|
| **provider** | Uses GPG keys to sign provider binaries |

---

## Key Design Principles

1. **Namespace Isolation** - GPG keys are scoped to namespaces
2. **Fingerprint Uniqueness** - No duplicate fingerprints across all namespaces
3. **Usage Safety** - Cannot delete keys in use by provider versions
4. **Immutable Key Data** - Key ID and fingerprint derived from ASCII armor, not modifiable
5. **Validation First** - ASCII armor validated before storage

---

## GPG Key Lifecycle

### Creation Flow

```
1. Validate namespace exists
2. Validate GPG key structure (ASCII armor)
3. Extract key ID and fingerprint
4. Check for duplicate fingerprint
5. Create GPGKey entity
6. Set optional fields (trust signature, source, source URL)
7. Save to repository
```

### Deletion Flow

```
1. Find GPG key by namespace and key ID
2. Check if key is in use by any provider versions
3. If in use, reject deletion
4. If not in use, delete from repository
```

---

## Error Handling

The domain defines specific errors for GPG key operations:

| Error | When Returned |
|-------|---------------|
| `ErrInvalidASCIIArmor` | ASCII armor structure is invalid |
| `ErrDuplicateFingerprint` | Fingerprint already exists |
| `ErrGPGKeyNotFound` | Key not found for namespace/key ID |
| `ErrGPGKeyInUse` | Key is in use by provider versions |

---

## Usage Examples

### Creating a GPG Key

```go
req := gpgkeyService.CreateGPGKeyRequest{
    Namespace:   "my-namespace",
    ASCIILArmor: "-----BEGIN PGP PUBLIC KEY BLOCK-----\n...",
    TrustSignature: &trustSignature,
    Source:       &source,
}

gpgKey, err := gpgKeyService.CreateGPGKey(ctx, req)
```

### Getting Namespace Keys

```go
keys, err := gpgKeyService.GetNamespaceGPGKeys(ctx, "my-namespace")
for _, key := range keys {
    fmt.Printf("Key ID: %s, Fingerprint: %s\n", key.KeyID(), key.Fingerprint())
}
```

### Deleting a GPG Key

```go
err := gpgKeyService.DeleteGPGKey(ctx, "my-namespace", "ABCD1234")
if err == gpgkeyModel.ErrGPGKeyInUse {
    // Handle in-use error
}
```

### Verifying a Signature

```go
valid, err := gpgKeyService.VerifySignature(
    ctx,
    "ABCD1234",
    signatureString,
    dataString,
)
if valid {
    // Signature is valid
}
```

---

## Integration with Provider Domain

Provider versions use GPG keys for signing:

```go
// When creating a provider version
providerVersion.SetGPGKeyID(gpgKey.KeyID())

// Signature verification occurs during download
signature, _ := gpg.GetSignatureFromProvider(provider)
valid, _ := gpgKeyService.VerifySignature(
    ctx,
    providerVersion.GPGKeyID(),
    signature,
    providerBinary,
)
```

---

## GPG Key Storage

### ASCII Armor Format

GPG keys are stored in ASCII-armored format:

```
-----BEGIN PGP PUBLIC KEY BLOCK-----
mQINBGA... (base64 encoded key data)
=ABC123
-----END PGP PUBLIC KEY BLOCK-----
```

### Key Extraction

The `infrastructure/gpg` package extracts:
- **Key ID** - Short identifier (e.g., "ABCD1234")
- **Fingerprint** - Full SHA-1 fingerprint (40 hex characters)

---

## Configuration

No specific configuration required. GPG key operations use the infrastructure GPG utilities which may have configuration for the GPG binary location.

---

## References

- [`/internal/infrastructure/gpg/`](../../infrastructure/gpg/) - GPG parsing and verification utilities
- [`/internal/domain/provider/`](../provider/) - Consumer of GPG keys for signing
