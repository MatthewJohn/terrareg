package middleware

import (
	"net/http"
	"os"
	"strings"
)

// SecurityHeaders adds security-related HTTP headers to responses
func SecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Prevent MIME type sniffing
		w.Header().Set("X-Content-Type-Options", "nosniff")

		// Prevent clickjacking
		w.Header().Set("X-Frame-Options", "DENY")

		// Enable XSS protection (legacy browsers)
		w.Header().Set("X-XSS-Protection", "1; mode=block")

		// Force HTTPS (remove in development if not using HTTPS)
		if !isDevelopmentMode() {
			w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains; preload")
		}

		// Control referrer information
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")

		// Content Security Policy
		// Only set CSP in production (HTTPS) - Python terrareg doesn't set CSP at all
		// This allows external CDN scripts (e.g., cdn.datatables.net) in development
		if !isDevelopmentMode() {
			csp := "default-src 'self'; " +
				"script-src 'self' 'unsafe-inline' 'unsafe-eval' https://cdn.datatables.net; " +
				"style-src 'self' 'unsafe-inline'; " +
				"img-src 'self' data: https:; " +
				"font-src 'self' data:; " +
				"connect-src 'self'; " +
				"frame-ancestors 'none'; " +
				"base-uri 'self'; " +
				"form-action 'self';"

			w.Header().Set("Content-Security-Policy", csp)
		}

		// Permissions Policy (formerly Feature Policy)
		permissionsPolicy := "geolocation=(), " +
			"microphone=(), " +
			"camera=(), " +
			"payment=(), " +
			"usb=(), " +
			"interest-cohort=()"

		w.Header().Set("Permissions-Policy", permissionsPolicy)

		// Remove server information
		w.Header().Set("Server", "")

		next.ServeHTTP(w, r)
	})
}

// AuthSecurityHeaders adds additional security headers for authentication endpoints
func AuthSecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Apply standard security headers
		SecurityHeaders(next).ServeHTTP(w, r)

		// Add auth-specific headers
		// Prevent caching of authentication responses, except for metadata endpoint
		if r.URL.Path != "/v1/terrareg/auth/saml/metadata" {
			w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate, private")
			w.Header().Set("Pragma", "no-cache")
			w.Header().Set("Expires", "0")
		}
	})
}

// CORSMiddleware provides CORS configuration
func CORSMiddleware(allowedOrigins []string, allowedMethods []string, allowedHeaders []string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")

			// Check if origin is allowed
			allowed := false
			if origin == "" {
				allowed = true // Allow same-origin requests
			} else {
				for _, allowedOrigin := range allowedOrigins {
					if allowedOrigin == "*" || allowedOrigin == origin {
						allowed = true
						break
					}
				}
			}

			if allowed {
				w.Header().Set("Access-Control-Allow-Origin", origin)
			}

			// Set other CORS headers
			if len(allowedMethods) > 0 {
				w.Header().Set("Access-Control-Allow-Methods", strings.Join(allowedMethods, ", "))
			}

			if len(allowedHeaders) > 0 {
				w.Header().Set("Access-Control-Allow-Headers", strings.Join(allowedHeaders, ", "))
			}

			// Allow credentials for specific origins
			if allowed && origin != "" {
				w.Header().Set("Access-Control-Allow-Credentials", "true")
			}

			// Handle preflight requests
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// isDevelopmentMode checks if we're running in development mode
func isDevelopmentMode() bool {
	// Check DEBUG environment variable (matching Python terrareg behavior)
	// Python reference: /app/terrareg/config.py - Config.debug property
	return os.Getenv("DEBUG") == "true"
}
