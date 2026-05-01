# Auth Domain Architecture

## Overview

The Auth domain provides a comprehensive multi-authentication system supporting GitHub, SAML, OIDC, API keys, and Terraform IDP. It follows Domain-Driven Design principles with clean separation between authentication factories (AuthMethods) and authentication state (AuthContexts).

---

## Core Functionality

The auth domain provides the following capabilities:

- **Multiple Authentication Methods** - Support for SAML, OIDC, GitHub OAuth, API keys, and Terraform IDP
- **Session Management** - Server-side session storage with encrypted client-side cookies
- **Permission Management** - Fine-grained access control with namespace-based permissions
- **User Group Support** - Group-based authorization integration
- **CSRF Protection** - Token generation and validation for state-changing operations

---

## Domain Components

### Models

**Location**: `/internal/domain/auth/model/`

#### Session Model

```go
type Session struct {
    ID           string      `json:"id" gorm:"primaryKey"`
    AuthMethod   string      `json:"auth_method"`
    ProviderData []byte      `json:"provider_data"`
    Expiry       time.Time   `json:"expiry"`
}
```

#### Authentication Token Model

```go
type AuthenticationToken struct {
    ID        string    `json:"id"`
    TokenType string    `json:"token_type"`
    Username  string    `json:"username"`
    IsAdmin   bool      `json:"is_admin"`
    Permissions map[string]string `json:"permissions,omitempty"`
}
```

#### Other Models

| Model | Purpose |
|-------|---------|
| `AuthRequest` | Authentication request data |
| `OAuthToken` | OAuth token storage |
| `TerraformSession` | Terraform IDP session data |
| `UserGroup` | User group for authorization |

### Repository Interfaces

**Location**: `/internal/domain/auth/repository/`

```go
type SessionRepository interface {
    Create(ctx context.Context, session *model.Session) error
    FindByID(ctx context.Context, id string) (*model.Session, error)
    Delete(ctx context.Context, id string) error
    Refresh(ctx context.Context, id string, expiry time.Time) error
}

type UserGroupRepository interface {
    FindByNamespace(ctx context.Context, namespace string) ([]string, error)
    FindAll(ctx context.Context) ([]*model.UserGroup, error)
}
```

### Services

**Location**: `/internal/domain/auth/service/`

#### AuthFactory (Orchestrator)

The central service that orchestrates authentication by trying multiple auth methods in priority order:

```go
type AuthFactory struct {
    authMethods   []auth.AuthMethod
    sessionRepo   repository.SessionRepository
    userGroupRepo repository.UserGroupRepository
    config        *infraConfig.InfrastructureConfig
}

// AuthenticateRequest tries each auth method and returns the first successful authentication
func (af *AuthFactory) AuthenticateRequest(
    ctx context.Context,
    headers, formData, queryParams map[string]string,
) (*model.AuthenticationResponse, error)
```

**Priority Order:**
1. AdminApiKey
2. AdminSession
3. UploadApiKey
4. PublishApiKey
5. SAML
6. OpenID Connect
7. GitHub OAuth (TODO)
8. Terraform OIDC
9. Terraform Analytics
10. Terraform Internal Extraction
11. NotAuthenticated (fallback)

#### SessionManagementService

Coordinates session and cookie operations:

```go
type SessionManagementService struct {
    sessionService *SessionService
    cookieService  *CookieService
}

// Methods for CRUD on both sessions and cookies
CreateSession(ctx, authMethod, username, isAdmin, permissions) (cookie, error)
ValidateSession(ctx, cookie) (*SessionData, error)
DeleteSession(ctx, cookie) error
RefreshSession(ctx, cookie) (string, error)
```

#### SessionService

Pure database operations for sessions:

```go
type SessionService struct {
    repo repository.SessionRepository
    config *config.InfrastructureConfig
}

// Create, Validate, Delete, Refresh - database only
CreateSession(ctx, authMethod, providerData, expiry) (*Session, error)
ValidateSession(ctx, sessionID) (*Session, error)
```

#### CookieService

Cookie encryption/decryption (AES-256-GCM):

```go
type CookieService struct {
    secretKey []byte
    config    *config.InfrastructureConfig
}

// Encrypt/Decrypt session data for HTTP cookies
EncryptSession(data *SessionData) (string, error)
DecryptSession(encryptedCookie) (*SessionData, error)
BuildCookie(sessionData, name, expiry) (*http.Cookie, error)
```

#### Auth Method Services

| Service | Purpose |
|---------|---------|
| `OidcService` | OpenID Connect authentication flow |
| `SamlService` | SAML authentication flow |
| `TerraformIdpService` | Terraform Cloud/Enterprise IDP authentication |

### AuthMethod/AuthContext Interface

The core abstraction separating authentication factories from authentication state.

#### AuthMethod Interfaces (Factory)

```go
// Base AuthMethod interface - all auth methods implement this
type AuthMethod interface {
    GetProviderType() AuthMethodType
    IsEnabled() bool
}

// Header-based authentication (Admin API Key, Upload API Key, Publish API Key)
type HeaderAuthMethod interface {
    AuthMethod
    Authenticate(ctx context.Context, headers, formData, queryParams map[string]string) (AuthContext, error)
}

// Session-based authentication (SAML, OpenID Connect, Admin Session)
type SessionAuthMethod interface {
    AuthMethod
    Authenticate(ctx context.Context, sessionData map[string]interface{}) (AuthContext, error)
}

// Token-based authentication (Terraform OIDC)
type TokenAuthMethod interface {
    AuthMethod
    Authenticate(ctx context.Context, token string) (AuthContext, error)
}

// Bearer token authentication
type BearerTokenAuthMethod interface {
    AuthMethod
    Authenticate(ctx context.Context, token string, additionalClaims map[string]interface{}) (AuthContext, error)
}
```

#### AuthContext Interface (State)

```go
type AuthContext interface {
    // Authentication state
    IsBuiltInAdmin() bool
    IsAdmin() bool
    IsAuthenticated() bool
    RequiresCSRF() bool
    CheckAuthState() bool

    // Permission checking
    CanPublishModuleVersion(namespace string) bool
    CanUploadModuleVersion(namespace string) bool
    CheckNamespaceAccess(permissionType, namespace string) bool
    GetAllNamespacePermissions() map[string]string

    // User information
    GetUsername() string
    GetUserGroupNames() []string

    // API access
    CanAccessReadAPI() bool
    CanAccessTerraformAPI() bool
    GetTerraformAuthToken() string

    // Provider information
    GetProviderType() AuthMethodType
    GetProviderData() map[string]interface{}
}
```

---

## Dependencies

### Domain Dependencies

| Domain | Purpose |
|--------|---------|
| **config** | For authentication configuration (InfrastructureConfig) |
| **shared** | For common error definitions |
| **url** | For URL building in authentication redirects |

### Infrastructure Dependencies

| Component | Purpose |
|-----------|---------|
| **Database** | Session and user group storage |
| **crewjam/saml** | SAML 2.0 implementation |
| **coreos/go-oidc/v3** | OpenID Connect implementation |

### Domains That Depend on Auth

All domains that require authentication:
- **module** - For module upload/publish authorization
- **provider** - For provider publishing authorization
- **namespace** - For namespace management

---

## Key Design Principles

1. **Four-Tier Architecture** - AuthFactory → SessionManagementService → SessionService/CookieService → Database
2. **Immutable Factories** - AuthMethods are stateless factories that create AuthContexts
3. **Request-Scoped Contexts** - Each AuthContext is created for a single request
4. **Interface Segregation** - Specialized interfaces for different authentication patterns
5. **No Mutable State** - Authentication state is captured in the AuthContext, not modified in the AuthMethod
6. **Separation of Concerns** - SessionService (DB) vs CookieService (crypto) vs SessionManagementService (coordination)

---

## Cookie Encryption

Client-side session data is encrypted using AES-256-GCM:

- **Algorithm**: AES-256-GCM (authenticated encryption)
- **Key**: 32-byte derived from `SECRET_KEY` config
- **Nonce**: 12-byte random per encryption
- **Format**: Base64(nonce + ciphertext + auth tag)

```go
// Encrypted cookie format
encryptedCookie = base64(nonce || ciphertext || auth_tag)
```

---

## Authentication Methods

### Admin API Key

**Interface**: `HeaderAuthMethod`

**Configuration**: `ADMIN_AUTHENTICATION_TOKEN`

**Usage**:
```bash
curl -H "X-Terrareg-ApiKey: your-admin-token" \
     http://localhost:3000/v1/terrareg/modules
```

**Features**:
- Full admin access to all namespaces
- No CSRF required (API-based)
- Simple header-based validation
- Case-insensitive header matching

### Upload API Key

**Interface**: `HeaderAuthMethod`

**Configuration**: `UPLOAD_API_KEYS` (comma-separated list)

**Features**:
- Can upload module versions to any namespace
- Cannot publish module versions
- Header: `X-Terrareg-Upload-Key`

### Publish API Key

**Interface**: `HeaderAuthMethod`

**Configuration**: `PUBLISH_API_KEYS` (comma-separated list)

**Features**:
- Can publish module versions to any namespace
- Cannot upload module versions
- Header: `X-Terrareg-Publish-Key`

### SAML

**Interface**: `SessionAuthMethod`

**Configuration**:
- `SAML2_IDP_METADATA_URL` - Identity Provider metadata URL
- `SAML2_ISSUER_ENTITY_ID` - SP entity ID
- `SAML2_PUBLIC_KEY` - SP certificate
- `SAML2_PRIVATE_KEY` - SP private key

**Features**:
- Production-ready SAML 2.0 using `crewjam/saml` library
- Proper signature verification and assertion validation
- Metadata-based configuration
- ACS endpoint for receiving SAML responses

### OpenID Connect

**Interface**: `SessionAuthMethod`

**Configuration**:
- `OPENID_CONNECT_CLIENT_ID` - OAuth client ID
- `OPENID_CONNECT_CLIENT_SECRET` - OAuth client secret
- `OPENID_CONNECT_ISSUER` - OpenID Connect issuer URL

**Features**:
- Production-ready OIDC using `coreos/go-oidc/v3` library
- ID token validation with signature verification
- Claims extraction (email, groups, etc.)
- User group mapping

### Terraform IDP

**Interface**: `BearerTokenAuthMethod`

**Configuration**:
- `TERRAFORM_OIDC_IDP_SIGNING_KEY_PATH` - Signing key path
- `TERRAFORM_OIDC_IDP_SUBJECT_ID_HASH_SALT` - Subject ID salt
- `TERRAFORM_OIDC_IDP_SESSION_EXPIRY` - Session expiry

**Features**:
- Terraform Cloud/Enterprise IDP authentication
- JWT token validation
- Subject ID hashing for privacy

---

## Authentication Flow

### Request Authentication Sequence

```
Client → AuthMiddleware → AuthFactory → AuthMethods (in priority order)
                                                    ↓
                                         Create AuthContext
                                                    ↓
                                    AuthMiddleware injects into context
                                                    ↓
                                      Handler uses AuthContext
```

### Header-Based Authentication (Admin API Key)

```
1. Extract X-Terrareg-ApiKey header
2. Validate against configured token
3. Check if admin token
4. Create AdminApiKeyAuthContext with admin permissions
5. Return AuthContext with IsAdmin=true
```

### Session-Based Authentication (SAML/OIDC)

```
1. Extract session cookie
2. Decrypt cookie using CookieService
3. Validate session using SessionService
4. Load provider data from database
5. Create appropriate AuthContext (SamlAuthContext/OidcAuthContext)
6. Return AuthContext with user permissions
```

---

## Security Considerations

### SAML Security

- **Signature Verification**: All SAML responses verified against IdP certificate
- **Assertion Validation**: Validates conditions, subject confirmation, attributes
- **Metadata Validation**: IdP metadata fetched and validated
- **Secure Storage**: SAML configuration stored securely with certificate/key files

### OIDC Security

- **ID Token Validation**: JWT signature verification using IdP's JWKS
- **Claims Validation**: Validates iss, aud, exp, nbf claims
- **Nonce Verification**: Prevents replay attacks
- **Group Mapping**: Maps user groups from IdP to Terrareg permissions

### Session Security

- **AES-256-GCM Encryption**: All session data encrypted client-side
- **Secure Cookies**: HttpOnly, Secure, SameSite settings
- **Session Expiry**: Configurable timeout with refresh capability
- **Secret Key**: Minimum 32 characters (256 bits) required

---

## Configuration

### Authentication Configuration

```bash
# Admin API Key Authentication
ADMIN_AUTHENTICATION_TOKEN="your-admin-token-here"

# Upload API Key Authentication (comma-separated)
UPLOAD_API_KEYS="upload-token-1,upload-token-2"

# Publish API Key Authentication (comma-separated)
PUBLISH_API_KEYS="publish-token-1,publish-token-2"

# SAML Configuration
SAML2_IDP_METADATA_URL="https://idp.example.com/metadata"
SAML2_ISSUER_ENTITY_ID="https://terrareg.example.com"
SAML2_PUBLIC_KEY="/path/to/cert.pem"
SAML2_PRIVATE_KEY="/path/to/key.pem"

# OpenID Connect Configuration
OPENID_CONNECT_CLIENT_ID="your-client-id"
OPENID_CONNECT_CLIENT_SECRET="your-client-secret"
OPENID_CONNECT_ISSUER="https://oidc.example.com"

# Terraform OIDC Configuration
TERRAFORM_OIDC_IDP_SIGNING_KEY_PATH="signing_key.pem"
TERRAFORM_OIDC_IDP_SUBJECT_ID_HASH_SALT="random-salt"
TERRAFORM_OIDC_IDP_SESSION_EXPIRY=3600

# Session Configuration (required for login methods)
SECRET_KEY="your-secret-key-minimum-32-characters-long"
SESSION_EXPIRY_MINS=1440
```

---

## Middleware Integration

### AuthenticationMiddleware

Validates requests and injects AuthContext into request context:

```go
// In handler
func (h *MyHandler) HandleRequest(w http.ResponseWriter, r *http.Request) {
    authCtx := middleware.GetAuthenticationContext(r)

    if !authCtx.IsAuthenticated() {
        RespondError(w, http.StatusUnauthorized, "Not authenticated")
        return
    }

    username := authCtx.GetUsername()
    isAdmin := authCtx.IsAdmin()
}
```

### AuthorizationMiddleware

Enforces access control based on AuthContext permissions:

```go
// Require upload permission for specific namespace
r.With(
    s.authMiddleware.RequireUploadPermission("{namespace}"),
).Post("/modules/{namespace}/{name}/{provider}/{version}/upload")

// Require admin access
r.With(
    s.authMiddleware.RequireAdmin,
).Get("/v1/terrareg/admin")
```

---

## References

For complete authentication architecture documentation including testing examples and migration guide, see:
- [`/docs/AUTHENTICATION_ARCHITECTURE.md`](../../docs/AUTHENTICATION_ARCHITECTURE.md)
