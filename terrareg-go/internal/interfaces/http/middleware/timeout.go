package middleware

import (
	"context"
	"net/http"
	"time"
)

// TimeoutConfig holds timeout configuration for different endpoint types
type TimeoutConfig struct {
	// Module indexing timeouts (in seconds)
	ModuleIndexingReadTimeout  int
	ModuleIndexingWriteTimeout int
}

// WithModuleIndexingTimeout creates a middleware that applies extended timeout
// for long-running module operations
func WithModuleIndexingTimeout(config TimeoutConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Use the longer of the two timeouts for the context
			timeoutSeconds := max(config.ModuleIndexingWriteTimeout, config.ModuleIndexingReadTimeout)

			// Create a custom context with extended timeout
			ctx, cancel := context.WithTimeout(r.Context(), time.Duration(timeoutSeconds)*time.Second)
			defer cancel()

			// Replace the request context with the timeout context
			r = r.WithContext(ctx)

			// Call the next handler
			next.ServeHTTP(w, r)
		})
	}
}

// max returns the maximum of two integers
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}