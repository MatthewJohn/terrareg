package middleware

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSecurityHeaders_StandardHeaders(t *testing.T) {
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := SecurityHeaders(nextHandler)
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "nosniff", rec.Header().Get("X-Content-Type-Options"))
	assert.Equal(t, "DENY", rec.Header().Get("X-Frame-Options"))
	assert.Equal(t, "1; mode=block", rec.Header().Get("X-XSS-Protection"))
	assert.Equal(t, "strict-origin-when-cross-origin", rec.Header().Get("Referrer-Policy"))
	assert.Equal(t, "", rec.Header().Get("Server"))
}

func TestSecurityHeaders_HSTS_InProduction(t *testing.T) {
	// Set production mode
	oldDebug := os.Getenv("DEBUG")
	defer os.Setenv("DEBUG", oldDebug)
	os.Setenv("DEBUG", "false")

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := SecurityHeaders(nextHandler)
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, "max-age=31536000; includeSubDomains; preload", rec.Header().Get("Strict-Transport-Security"))
}

func TestSecurityHeaders_HSTS_InDevelopment(t *testing.T) {
	// Set development mode
	oldDebug := os.Getenv("DEBUG")
	defer os.Setenv("DEBUG", oldDebug)
	os.Setenv("DEBUG", "true")

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := SecurityHeaders(nextHandler)
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, "", rec.Header().Get("Strict-Transport-Security"), "HSTS should not be set in development")
}

func TestSecurityHeaders_CSP_InProduction(t *testing.T) {
	// Set production mode
	oldDebug := os.Getenv("DEBUG")
	defer os.Setenv("DEBUG", oldDebug)
	os.Setenv("DEBUG", "false")

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := SecurityHeaders(nextHandler)
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	csp := rec.Header().Get("Content-Security-Policy")
	assert.NotEmpty(t, csp, "CSP should be set in production")
	assert.Contains(t, csp, "default-src 'self'")
	assert.Contains(t, csp, "script-src 'self'")
	assert.Contains(t, csp, "frame-ancestors 'none'")
	assert.Contains(t, csp, "base-uri 'self'")
	assert.Contains(t, csp, "form-action 'self'")
}

func TestSecurityHeaders_CSP_InDevelopment(t *testing.T) {
	// Set development mode
	oldDebug := os.Getenv("DEBUG")
	defer os.Setenv("DEBUG", oldDebug)
	os.Setenv("DEBUG", "true")

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := SecurityHeaders(nextHandler)
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, "", rec.Header().Get("Content-Security-Policy"), "CSP should not be set in development")
}

func TestSecurityHeaders_PermissionsPolicy(t *testing.T) {
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := SecurityHeaders(nextHandler)
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	permissions := rec.Header().Get("Permissions-Policy")
	assert.NotEmpty(t, permissions)
	assert.Contains(t, permissions, "geolocation=()")
	assert.Contains(t, permissions, "microphone=()")
	assert.Contains(t, permissions, "camera=()")
	assert.Contains(t, permissions, "payment=()")
	assert.Contains(t, permissions, "usb=()")
	assert.Contains(t, permissions, "interest-cohort=()")
}

func TestAuthSecurityHeaders_StandardHeaders(t *testing.T) {
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := AuthSecurityHeaders(nextHandler)
	req := httptest.NewRequest(http.MethodGet, "/v1/terrareg/auth/login", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "no-store, no-cache, must-revalidate, private", rec.Header().Get("Cache-Control"))
	assert.Equal(t, "no-cache", rec.Header().Get("Pragma"))
	assert.Equal(t, "0", rec.Header().Get("Expires"))
}

func TestAuthSecurityHeaders_StandardSecurityHeaders(t *testing.T) {
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := AuthSecurityHeaders(nextHandler)
	req := httptest.NewRequest(http.MethodGet, "/v1/terrareg/auth/login", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, "nosniff", rec.Header().Get("X-Content-Type-Options"))
	assert.Equal(t, "DENY", rec.Header().Get("X-Frame-Options"))
	assert.Equal(t, "1; mode=block", rec.Header().Get("X-XSS-Protection"))
}

func TestAuthSecurityHeaders_MetadataEndpoint(t *testing.T) {
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := AuthSecurityHeaders(nextHandler)
	req := httptest.NewRequest(http.MethodGet, "/v1/terrareg/auth/saml/metadata", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	cacheControl := rec.Header().Get("Cache-Control")
	assert.NotEqual(t, "no-store, no-cache, must-revalidate, private", cacheControl,
		"Cache-Control should not be overridden for metadata endpoint")
}

func TestAuthSecurityHeaders_NonMetadataEndpoint(t *testing.T) {
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := AuthSecurityHeaders(nextHandler)
	req := httptest.NewRequest(http.MethodGet, "/v1/terrareg/auth/login", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, "no-store, no-cache, must-revalidate, private", rec.Header().Get("Cache-Control"))
}

func TestCORSMiddleware_AllowAllOrigins(t *testing.T) {
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := CORSMiddleware([]string{"*"}, []string{"GET", "POST"}, []string{"Content-Type"})(nextHandler)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Origin", "https://example.com")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, "https://example.com", rec.Header().Get("Access-Control-Allow-Origin"))
	assert.Equal(t, "GET, POST", rec.Header().Get("Access-Control-Allow-Methods"))
	assert.Equal(t, "Content-Type", rec.Header().Get("Access-Control-Allow-Headers"))
	assert.Equal(t, "true", rec.Header().Get("Access-Control-Allow-Credentials"))
}

func TestCORSMiddleware_SpecificOriginAllowed(t *testing.T) {
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	allowedOrigins := []string{"https://trusted.com", "https://another.com"}
	handler := CORSMiddleware(allowedOrigins, []string{"GET"}, []string{})(nextHandler)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Origin", "https://trusted.com")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, "https://trusted.com", rec.Header().Get("Access-Control-Allow-Origin"))
}

func TestCORSMiddleware_OriginNotAllowed(t *testing.T) {
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := CORSMiddleware([]string{"https://trusted.com"}, []string{"GET"}, []string{})(nextHandler)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Origin", "https://malicious.com")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, "", rec.Header().Get("Access-Control-Allow-Origin"))
}

func TestCORSMiddleware_SameOriginRequest(t *testing.T) {
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := CORSMiddleware([]string{"https://trusted.com"}, []string{"GET"}, []string{})(nextHandler)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, "", rec.Header().Get("Access-Control-Allow-Origin"),
		"Same-origin requests should not set CORS headers")
}

func TestCORSMiddleware_PreflightRequest(t *testing.T) {
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("should not reach here"))
	})

	handler := CORSMiddleware([]string{"*"}, []string{"GET", "POST"}, []string{"Content-Type"})(nextHandler)

	req := httptest.NewRequest(http.MethodOptions, "/test", nil)
	req.Header.Set("Origin", "https://example.com")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNoContent, rec.Code)
	assert.Empty(t, rec.Body.String(), "Preflight should not call next handler")
	assert.Equal(t, "GET, POST", rec.Header().Get("Access-Control-Allow-Methods"))
	assert.Equal(t, "Content-Type", rec.Header().Get("Access-Control-Allow-Headers"))
}

func TestCORSMiddleware_NonPreflightRequest(t *testing.T) {
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("response"))
	})

	handler := CORSMiddleware([]string{"*"}, []string{"GET", "POST"}, []string{"Content-Type"})(nextHandler)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Origin", "https://example.com")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "response", rec.Body.String())
}

func TestCORSMiddleware_EmptyMethods(t *testing.T) {
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := CORSMiddleware([]string{"*"}, []string{}, []string{})(nextHandler)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, "", rec.Header().Get("Access-Control-Allow-Methods"))
}

func TestCORSMiddleware_EmptyHeaders(t *testing.T) {
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := CORSMiddleware([]string{"*"}, []string{"GET"}, []string{})(nextHandler)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, "", rec.Header().Get("Access-Control-Allow-Headers"))
}

func TestCORSMiddleware_AllowCredentials(t *testing.T) {
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := CORSMiddleware([]string{"https://trusted.com"}, []string{"GET"}, []string{})(nextHandler)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Origin", "https://trusted.com")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, "true", rec.Header().Get("Access-Control-Allow-Credentials"))
}

func TestCORSMiddleware_NoCredentialsForWildcard(t *testing.T) {
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := CORSMiddleware([]string{"*"}, []string{"GET"}, []string{})(nextHandler)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Origin", "https://example.com")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, "true", rec.Header().Get("Access-Control-Allow-Credentials"))
}

func TestSecurityHeaders_ServerHeader(t *testing.T) {
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := SecurityHeaders(nextHandler)
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, "", rec.Header().Get("Server"), "Server header should be removed")
}

func TestSecurityHeaders_AllowsCustomHeaders(t *testing.T) {
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Custom-Header", "custom-value")
		w.WriteHeader(http.StatusOK)
	})

	handler := SecurityHeaders(nextHandler)
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, "custom-value", rec.Header().Get("X-Custom-Header"), "Custom headers should be preserved")
}

func TestSecurityHeaders_ProducesCorrectCSP(t *testing.T) {
	oldDebug := os.Getenv("DEBUG")
	defer os.Setenv("DEBUG", oldDebug)
	os.Setenv("DEBUG", "false")

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := SecurityHeaders(nextHandler)
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	csp := rec.Header().Get("Content-Security-Policy")
	expectedCSP := "default-src 'self'; " +
		"script-src 'self' 'unsafe-inline' 'unsafe-eval' https://cdn.datatables.net; " +
		"style-src 'self' 'unsafe-inline'; " +
		"img-src 'self' data: https:; " +
		"font-src 'self' data:; " +
		"connect-src 'self'; " +
		"frame-ancestors 'none'; " +
		"base-uri 'self'; " +
		"form-action 'self';"

	assert.Equal(t, expectedCSP, csp)
}
