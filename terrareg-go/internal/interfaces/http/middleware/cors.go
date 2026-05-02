package middleware

import (
	"net/http"
)

// CORSOptions holds CORS configuration
type CORSOptions struct {
	AllowedOrigins   []string
	AllowedMethods   []string
	AllowedHeaders   []string
	AllowCredentials bool
	MaxAge           int
}

// DefaultCORSOptions returns default CORS options
func DefaultCORSOptions() CORSOptions {
	return CORSOptions{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		AllowCredentials: false,
		MaxAge:           300,
	}
}

// CORS creates a CORS middleware
func CORS(options CORSOptions) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")

			// Check if origin is allowed
			allowed := false
			for _, allowedOrigin := range options.AllowedOrigins {
				if allowedOrigin == "*" || allowedOrigin == origin {
					allowed = true
					break
				}
			}

			if allowed {
				if len(options.AllowedOrigins) == 1 && options.AllowedOrigins[0] == "*" {
					w.Header().Set("Access-Control-Allow-Origin", "*")
				} else {
					w.Header().Set("Access-Control-Allow-Origin", origin)
					w.Header().Add("Vary", "Origin")
				}

				w.Header().Set("Access-Control-Allow-Methods", joinStrings(options.AllowedMethods))
				w.Header().Set("Access-Control-Allow-Headers", joinStrings(options.AllowedHeaders))

				if options.AllowCredentials {
					w.Header().Set("Access-Control-Allow-Credentials", "true")
				}

				if options.MaxAge > 0 {
					w.Header().Set("Access-Control-Max-Age", string(rune(options.MaxAge)))
				}
			}

			// Handle preflight requests
			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func joinStrings(strs []string) string {
	result := ""
	for i, str := range strs {
		if i > 0 {
			result += ", "
		}
		result += str
	}
	return result
}
