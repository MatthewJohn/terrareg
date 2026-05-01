# URL Domain Architecture

## Overview

The URL domain provides centralized URL generation and processing utilities for the Terrareg application. It handles public URL detection, protocol handling, and URL construction for various purposes including Terraform source URLs.

---

## Core Functionality

The url domain provides the following capabilities:

- **Public URL Detection** - Extract protocol, domain, and port from configuration
- **HTTPS Detection** - Determine if the application is running in HTTPS mode
- **Base URL Construction** - Build consistent base URLs
- **URL Building** - Construct URLs with proper path joining
- **Terraform Source URLs** - Generate Terraform-specific source URLs for modules

---

## Domain Components

### Models

**Location**: `/internal/domain/url/service/url_service.go`

#### Protocol Type

```go
type Protocol string

const (
    ProtocolHttp  Protocol = "http"
    ProtocolHttps Protocol = "https"
)
```

#### PublicURLDetails

```go
type PublicURLDetails struct {
    Protocol Protocol
    Domain   string
    Port     int
}
```

Represents the parsed components of the public URL:
- `Protocol` - HTTP or HTTPS
- `Domain` - Domain name or IP address
- `Port` - Port number (standard ports: 80 for HTTP, 443 for HTTPS)

### Service

**Location**: `/internal/domain/url/service/url_service.go`

```go
type URLService struct {
    config *infraConfig.InfrastructureConfig
}
```

**Key Methods**:
- `GetPublicURLDetails(fallbackDomain)` - Extract protocol, domain, port from config
- `IsHTTPS(fallbackDomain)` - Determine if running in HTTPS mode
- `GetBaseURL(fallbackDomain)` - Get base URL without trailing slash
- `BuildURL(path, fallbackDomain)` - Build full URL with path
- `GetHostWithPort(fallbackDomain)` - Get host with port if non-standard
- `BuildTerraformSourceURL(providerID, version, modulePath, requestDomain)` - Build Terraform source URL

---

## Dependencies

### Domain Dependencies

| Domain | Purpose |
|--------|---------|
| **config** | For InfrastructureConfig (PUBLIC_URL, DOMAIN_NAME) |

### Infrastructure Dependencies

| Component | Purpose |
|-----------|---------|
| **net/url** | URL parsing and construction |

### Domains That Use URL

| Domain | Purpose |
|--------|---------|
| **auth** | For redirect URLs and HTTPS detection |
| **module** | For module download URLs |
| **git** | For Git clone URLs |

---

## Key Design Principles

1. **Single Source of Truth** - All URL building goes through URLService
2. **Configuration First** - URLs derived from PUBLIC_URL configuration
3. **Fallback Support** - Supports fallback domain from request
4. **Standard Port Handling** - Omits standard ports (80, 443) from URLs
5. **Terraform Compatibility** - Generates Terraform-compatible source URLs

---

## URL Detection Logic

### Priority Order

```
1. PUBLIC_URL environment variable (if valid)
2. DOMAIN_NAME environment variable (if set)
3. Fallback domain from request (if provided)
4. Defaults to localhost
```

### Public URL Parsing

```go
// From PUBLIC_URL
if s.config.PublicURL != "" {
    parsedURL, err := url.Parse(s.config.PublicURL)
    if err == nil && parsedURL.Hostname() != "" {
        protocol = parsedURL.Scheme
        port = getPortFromURL(parsedURL)
        domain = parsedURL.Hostname()
    }
}
```

---

## Port Handling

### Standard Ports

| Protocol | Standard Port | Included in URL |
|----------|---------------|-----------------|
| HTTP | 80 | No |
| HTTPS | 443 | No |

### Non-Standard Ports

Always included in URL:
- `http://example.com:8080` - Port 8088 included
- `https://example.com:8443` - Port 8443 included

---

## Terraform Source URL Format

### HTTP (Non-HTTPS)

```
http://{domain}:{port}/modules/{provider_id}/{version}//{module_path}
```

Example: `http://localhost:5000/modules/hashicorp/aws/5.0.0//modules/vpc`

### HTTPS

```
{domain}/{provider_id}//{module_path}
```

Example: `registry.example.com/hashicorp/aws//modules/vpc`

**Note**: HTTPS URLs omit protocol and port (if standard) for Terraform compatibility.

### With Analytics Token

If analytics is enabled:
```
{analytics_token}__{provider_id}//{module_path}
```

Example: `my-tf-app__hashicorp/aws//modules/vpc`

---

## Usage Examples

### Getting Public URL Details

```go
details := urlService.GetPublicURLDetails(nil)

fmt.Printf("Protocol: %s\n", details.Protocol)   // "https"
fmt.Printf("Domain: %s\n", details.Domain)      // "registry.example.com"
fmt.Printf("Port: %d\n", details.Port)          // 443
```

### Checking HTTPS Mode

```go
isHTTPS := urlService.IsHTTPS(nil)

if isHTTPS {
    // Use HTTPS-specific logic
}
```

### Building Base URL

```go
baseURL := urlService.GetBaseURL(nil)
// Returns: "https://registry.example.com"
// Or with non-standard port: "http://localhost:8080"
```

### Building Full URL

```go
fullURL := urlService.BuildURL("/v1/modules", nil)
// Returns: "https://registry.example.com/v1/modules"
```

### Building Terraform Source URL

```go
sourceURL := urlService.BuildTerraformSourceURL(
    "hashicorp/aws",  // provider ID
    "5.0.0",         // version
    "modules/vpc",   // module path
    "",              // request domain (empty = use config)
)
// Returns: "registry.example.com/hashicorp/aws/5.0.0//modules/vpc"
```

---

## Fallback Domain Pattern

When handling requests, use the request domain as fallback:

```go
func (h *Handler) HandleRequest(w http.ResponseWriter, r *http.Request) {
    requestDomain := r.Host

    details := urlService.GetPublicURLDetails(&requestDomain)
    // Uses request domain if PUBLIC_URL not configured
}
```

This allows the application to work correctly:
- Behind reverse proxies with different domains
- In development with localhost
- In production with configured domain

---

## Host With Port

```go
hostWithPort := urlService.GetHostWithPort(nil)
// Returns: "registry.example.com" (standard port)
// Or: "localhost:8080" (non-standard port)
```

Useful for:
- HTTP Host headers
- CORS configuration
- Cookie domain settings

---

## URL Construction Rules

### Path Joining

```go
// Handles leading slashes correctly
urlService.BuildURL("modules", nil)       // "/modules"
urlService.BuildURL("/modules", nil)      // "/modules"
urlService.BuildURL("modules/", nil)      // "/modules/"
```

### Module Path Normalization

```go
// Removes leading slashes from module path
urlService.BuildTerraformSourceURL(..., "modules/vpc", ...)
// Correctly handles: "/modules/vpc" → "modules/vpc"
//                        "modules/vpc" → "modules/vpc"
```

---

## Configuration

### Environment Variables

| Variable | Description | Example |
|----------|-------------|---------|
| `PUBLIC_URL` | Full public URL | `https://registry.example.com` |
| `DOMAIN_NAME` | Domain name only | `registry.example.com` |
| `LISTEN_PORT` | Server port | `5000` |

### Configuration Priority

```
PUBLIC_URL > DOMAIN_NAME > Request Domain > localhost
```

---

## References

- [`/internal/domain/url/service/url_service.go`](./service/url_service.go) - URL service implementation
- [`/internal/infrastructure/config/model/config.go`](../../infrastructure/config/model/config.go) - Configuration model
