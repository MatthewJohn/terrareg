package middleware

import (
	"context"
	"net/http"
	"strings"
)

// contextKey is a custom type for context keys
type contextKey string

const (
	userContextKey contextKey = "user"
)

// AuthMiddleware provides authentication middleware
type AuthMiddleware struct {
	secretKey string
}

// NewAuthMiddleware creates a new auth middleware
func NewAuthMiddleware(secretKey string) *AuthMiddleware {
	return &AuthMiddleware{
		secretKey: secretKey,
	}
}

// RequireAuth is a middleware that requires authentication
func (m *AuthMiddleware) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check for API key in Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader != "" {
			// Bearer token or API key
			if strings.HasPrefix(authHeader, "Bearer ") {
				token := strings.TrimPrefix(authHeader, "Bearer ")
				// TODO: Validate token
				_ = token
			}
		}

		// Check for session cookie
		cookie, err := r.Cookie("session_id")
		if err == nil {
			// TODO: Validate session
			_ = cookie
		}

		// For now, allow all requests (will implement proper auth later)
		next.ServeHTTP(w, r)
	})
}

// OptionalAuth is a middleware that optionally authenticates
func (m *AuthMiddleware) OptionalAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Try to authenticate but don't fail if not authenticated
		// TODO: Implement optional authentication
		next.ServeHTTP(w, r)
	})
}

// GetUserFromContext retrieves the user from the request context
func GetUserFromContext(ctx context.Context) (interface{}, bool) {
	user, ok := ctx.Value(userContextKey).(interface{})
	return user, ok
}

// SetUserInContext sets the user in the request context
func SetUserInContext(ctx context.Context, user interface{}) context.Context {
	return context.WithValue(ctx, userContextKey, user)
}
