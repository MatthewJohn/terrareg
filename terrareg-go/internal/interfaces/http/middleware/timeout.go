package middleware

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
)

// WithStandardTimeout returns chi's timeout middleware with the specified duration
func WithStandardTimeout(timeoutSeconds int) func(http.Handler) http.Handler {
	return middleware.Timeout(time.Duration(timeoutSeconds) * time.Second)
}

// WithModuleIndexingTimeout returns chi's timeout middleware configured for long-running operations
func WithModuleIndexingTimeout(timeoutSeconds int) func(http.Handler) http.Handler {
	return middleware.Timeout(time.Duration(timeoutSeconds) * time.Second)
}
