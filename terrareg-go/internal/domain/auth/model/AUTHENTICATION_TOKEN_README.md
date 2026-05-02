# AuthenticationToken Domain Model

## Overview

The `AuthenticationToken` domain model provides a secure way to manage API authentication tokens for the TerraReg module registry. It supports three types of tokens:

1. **Admin Tokens** - Full administrative access to the system
2. **Upload Tokens** - Can upload module versions to any namespace
3. **Publish Tokens** - Can publish module versions only to a specific namespace

## Implementation Details

### Core Components

#### AuthenticationToken Structure
- `id` - Unique identifier for the token
- `tokenType` - Type of token (Admin/Upload/Publish)
- `tokenValue` - Cryptographically secure random token value
- `namespace` - Associated namespace (only for publish tokens)
- `description` - Human-readable description of the token
- `createdAt` - Token creation timestamp
- `expiresAt` - Optional expiration timestamp
- `isActive` - Whether the token is currently active
- `createdBy` - User who created the token

#### AuthenticationTokenType Enum
- `AuthenticationTokenTypeAdmin` - Admin tokens with full system access
- `AuthenticationTokenTypeUpload` - Upload tokens for module uploads
- `AuthenticationTokenTypePublish` - Publish tokens for specific namespaces

#### AuthenticationTokenRepository Interface
Defines the contract for token persistence with methods for:
- Validating tokens
- Managing token lifecycle (create, revoke, update)
- Searching and filtering tokens
- Cleanup operations

#### AuthenticationTokenService
Provides business logic for:
- Token creation and validation
- Namespace access control
- Token management operations
- Cleanup of expired tokens

## Security Features

1. **Cryptographically Secure Tokens**: Uses `crypto/rand` to generate 256-bit tokens
2. **Base64 URL Encoding**: Safe for use in HTTP headers and URLs
3. **Namespace Isolation**: Publish tokens are restricted to specific namespaces
4. **Expiration Support**: Optional expiration dates for tokens
5. **Revocation Support**: Tokens can be deactivated without deletion
6. **No Plaintext Storage**: Token values are stored securely (implementation-dependent)

## Usage Examples

### Creating an Admin Token
```go
token, err := model.NewAuthenticationToken(
    model.AuthenticationTokenTypeAdmin,
    "Admin API token for automation",
    nil, // No namespace for admin tokens
    nil, // No expiration
    "admin-user",
)
```

### Creating a Publish Token
```go
namespace, _ := model.NewNamespace("myorg", nil, model.NamespaceTypeGithubOrg)
token, err := model.NewAuthenticationToken(
    model.AuthenticationTokenTypePublish,
    "Publish token for myorg namespace",
    namespace,
    time.Now().Add(30 * 24 * time.Hour), // 30 days
    "admin-user",
)
```

### Validating a Token
```go
token, err := service.ValidateAndAuthenticate(tokenValue)
if err != nil {
    // Handle invalid token
}

// Check namespace access
if !token.CanAccessNamespace(requestedNamespace) {
    // Access denied
}
```

## Integration with Python TerraReg

This implementation aligns with the Python TerraReg authentication system:

1. **Token Types**: Maps to Python's AdminApiKey, UploadApiKey, and PublishApiKey auth methods
2. **Header Validation**: Expects `X-Terrareg-ApiKey` header (matching Python implementation)
3. **Access Control**: Mirrors Python's namespace-based permission system
4. **Auth Method Mapping**: Direct mapping to Python's auth method types

## Testing

Comprehensive unit tests are provided covering:
- Token creation and validation
- Type conversions and string representations
- Namespace access control
- Expiration handling
- Revocation operations
- Edge cases and error conditions

Run tests with:
```bash
go test ./internal/domain/auth/model -v
```