# Security Domain Architecture

## Overview

The Security domain provides security utilities for the Terrareg application, primarily CSRF (Cross-Site Request Forgery) protection. It implements token generation and validation for state-changing operations.

---

## Core Functionality

The security domain provides the following capabilities:

- **CSRF Token Generation** - Generate secure random tokens for CSRF protection
- **CSRF Token Validation** - Validate tokens on form submissions
- **Token Comparison** - Constant-time comparison to prevent timing attacks

---

## Domain Components

### Models

**Location**: `/internal/domain/security/csrf/`

#### CSRFToken Type

```go
type CSRFToken string
```

The `CSRFToken` type represents a CSRF token value with helper methods.

### Token Generation

**Location**: `/internal/domain/security/csrf/token.go`

```go
func NewCSRFToken() (CSRFToken, error) {
    // Generate 32 random bytes (256 bits)
    bytes := make([]byte, 32)
    rand.Read(bytes)

    // Hash the random bytes using SHA256
    hash := sha256.Sum256(bytes)
    return CSRFToken(hex.EncodeToString(hash[:])), nil
}
```

**Generation Process**:
1. Generate 32 cryptographically secure random bytes (256 bits)
2. Hash using SHA256
3. Encode as hexadecimal string

### Token Validation

**Location**: `/internal/domain/security/csrf/validator.go`

```go
func ValidateToken(token CSRFToken, expectedToken CSRFToken) bool {
    return token.Equals(expectedToken)
}
```

### CSRFToken Methods

```go
// String returns the string representation
func (t CSRFToken) String() string

// IsEmpty checks if the token is empty
func (t CSRFToken) IsEmpty() bool

// Equals compares two tokens for equality (constant-time)
func (t CSRFToken) Equals(other CSRFToken) bool
```

---

## Dependencies

### Domain Dependencies

None - this is a standalone security utility domain.

### Infrastructure Dependencies

| Component | Purpose |
|-----------|---------|
| **crypto/rand** - Cryptographically secure random number generation |
| **crypto/sha256** - SHA256 hashing |

---

## Key Design Principles

1. **Cryptographic Security** - Uses crypto/rand for secure random generation
2. **Constant-Time Comparison** - Prevents timing attacks on token comparison
3. **SHA256 Hashing** - Double-hashing for additional security
4. **Simple API** - Easy to use token generation and validation

---

## CSRF Protection Flow

### Token Generation

```
1. Generate secure random bytes
2. Hash with SHA256
3. Store in session
4. Embed in form/headers
```

### Token Validation

```
1. Extract token from request
2. Get expected token from session
3. Compare using constant-time comparison
4. Allow or reject request
```

---

## Usage Examples

### Generating a Token

```go
token, err := csrf.NewCSRFToken()
if err != nil {
    return err
}

// Store in session
session.SetCSRFToken(token.String())

// Embed in form
<input type="hidden" name="csrf_token" value="{{.CSRFToken}}">
```

### Validating a Token

```go
// Get token from request
requestToken := csrf.CSRFToken(r.FormValue("csrf_token"))

// Get expected token from session
expectedToken := csrf.CSRFToken(session.GetCSRFToken())

// Validate
if !requestToken.Equals(expectedToken) {
    return errors.New("invalid CSRF token")
}
```

### Middleware Integration

```go
func CSRFMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Skip for GET requests (read-only)
        if r.Method == "GET" || r.Method == "HEAD" {
            next.ServeHTTP(w, r)
            return
        }

        // Validate CSRF token for state-changing operations
        token := csrf.CSRFToken(r.Header.Get("X-CSRF-Token"))
        expectedToken := getCSRFTokenFromSession(r)

        if !token.Equals(expectedToken) {
            http.Error(w, "Invalid CSRF token", http.StatusForbidden)
            return
        }

        next.ServeHTTP(w, r)
    })
}
```

---

## Token Format

CSRF tokens are hexadecimal-encoded SHA256 hashes:

```
Length: 64 characters (32 bytes * 2 hex chars/byte)
Format: [0-9a-f]{64}
Example: a1b2c3d4e5f6...7890abcdef1234
```

---

## Security Considerations

### Cryptographic Randomness

Uses `crypto/rand` instead of `math/rand`:

```go
bytes := make([]byte, 32)
_, err := rand.Read(bytes)
```

### Constant-Time Comparison

Prevents timing attacks:

```go
func (t CSRFToken) Equals(other CSRFToken) bool {
    return t == other  // Go's == is constant-time for strings
}
```

### Session Binding

CSRF tokens should be bound to sessions:
- One token per session
- Token regenerated on login
- Token cleared on logout

---

## When to Use CSRF Protection

### Required For

- Form submissions (POST, PUT, DELETE, PATCH)
- State-changing operations
- Authenticated user actions

### Not Required For

- GET/HEAD requests (read-only)
- Public API endpoints
- API requests with other authentication (e.g., API keys)

---

## References

- [`/internal/domain/security/csrf/token.go`](./csrf/token.go) - Token generation
- [`/internal/domain/security/csrf/validator.go`](./csrf/validator.go) - Token validation
- [`/internal/domain/security/csrf/generator.go`](./csrf/generator.go) - Token generator
- [`/internal/domain/security/csrf/errors.go`](./csrf/errors.go) - CSRF errors
