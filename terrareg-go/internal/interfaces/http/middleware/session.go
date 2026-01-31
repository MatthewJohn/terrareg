package middleware

import (
	"context"
	"net/http"

	"github.com/rs/zerolog"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/auth"
	authservice "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/auth/service"
)

// SessionMiddleware handles session management for HTTP requests
// This is a global middleware that adds authentication context to all requests
// It does NOT require authentication - just adds auth info if available
type SessionMiddleware struct {
	authFactory *authservice.AuthFactory
	logger      zerolog.Logger
}

// NewSessionMiddleware creates a new session middleware
func NewSessionMiddleware(authFactory *authservice.AuthFactory, logger zerolog.Logger) *SessionMiddleware {
	return &SessionMiddleware{
		authFactory: authFactory,
		logger:      logger,
	}
}

// Session extracts and validates session from cookie, adding it to the request context
// This middleware does not require authentication - it just makes auth info available
func (m *SessionMiddleware) Session(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		// Use auth factory to validate request - now returns domain AuthContext directly
		authCtx, err := m.authFactory.AuthenticateRequest(ctx, getHeadersMap(r), getFormDataMap(r), getQueryParamsMap(r))
		if err != nil {
			// Log error but continue without authentication
			m.logger.Debug().Err(err).Msg("Failed to validate authentication, continuing without auth")
			// Set not authenticated context and continue
			authCtx = authservice.NewNotAuthenticatedAuthContext()
		}

		// Add domain auth context to request context
		ctx = context.WithValue(ctx, authContextKey, authCtx)

		// Update the request with the new context
		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)
	})
}

// Helper functions to extract request data
func getHeadersMap(r *http.Request) map[string]string {
	headers := make(map[string]string)
	for name, values := range r.Header {
		if len(values) > 0 {
			headers[name] = values[0]
		}
	}
	return headers
}

func getFormDataMap(r *http.Request) map[string]string {
	if err := r.ParseForm(); err != nil {
		return make(map[string]string)
	}
	form := make(map[string]string)
	for key, values := range r.Form {
		if len(values) > 0 {
			form[key] = values[0]
		}
	}
	return form
}

func getQueryParamsMap(r *http.Request) map[string]string {
	params := make(map[string]string)
	for key, values := range r.URL.Query() {
		if len(values) > 0 {
			params[key] = values[0]
		}
	}
	return params
}

// Legacy compatibility functions

// GetCSRFToken retrieves CSRF token (placeholder for compatibility)
// CSRF tokens are not part of the current authentication architecture
func GetCSRFToken(ctx context.Context) string {
	return ""
}

// GetSessionData is a compatibility function that returns auth context
// This maps the old GetSessionData API to the new GetAuthContext API
// Returns domain auth.AuthContext interface - never nil
func GetSessionData(ctx context.Context) auth.AuthContext {
	return GetAuthContext(ctx)
}

// GetAuthenticationContext is an alias for GetAuthContext for compatibility
// Returns domain auth.AuthContext interface - never nil
func GetAuthenticationContext(ctx context.Context) auth.AuthContext {
	return GetAuthContext(ctx)
}

// WithAuthenticationContext sets auth context in the request context (for testing compatibility)
func WithAuthenticationContext(ctx context.Context, authCtx auth.AuthContext) context.Context {
	return context.WithValue(ctx, authContextKey, authCtx)
}
