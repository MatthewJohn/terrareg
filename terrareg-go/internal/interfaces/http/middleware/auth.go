package middleware

import (
	"context"
	"net/http"
	"strings"

	authQuery "github.com/matthewjohn/terrareg/terrareg-go/internal/application/query/auth"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/config"
)

// contextKey is a custom type for context keys
type contextKey string

const (
	userContextKey        contextKey = "user"
	isAdminContextKey     contextKey = "is_admin"
	sessionIDContextKey   contextKey = "session_id"
)

// AuthMiddleware provides authentication middleware
type AuthMiddleware struct {
	config            *config.Config
	checkSessionQuery *authQuery.CheckSessionQuery
}

// NewAuthMiddleware creates a new auth middleware
func NewAuthMiddleware(cfg *config.Config, checkSessionQuery *authQuery.CheckSessionQuery) *AuthMiddleware {
	return &AuthMiddleware{
		config:            cfg,
		checkSessionQuery: checkSessionQuery,
	}
}

// RequireAuth is a middleware that requires authentication
func (m *AuthMiddleware) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		authenticated := false

		// Check for API key in Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader != "" {
			// Bearer token authentication
			if strings.HasPrefix(authHeader, "Bearer ") {
				token := strings.TrimPrefix(authHeader, "Bearer ")
				// Check if token matches admin auth token
				if m.config.AdminAuthenticationToken != "" && token == m.config.AdminAuthenticationToken {
					authenticated = true
					ctx = context.WithValue(ctx, isAdminContextKey, true)
				}
			}
		}

		// Check for session cookie if not already authenticated
		if !authenticated {
			sessionCookie, err := r.Cookie("session_id")
			if err == nil {
				// Validate session
				session, err := m.checkSessionQuery.Execute(ctx, sessionCookie.Value)
				if err == nil && session != nil && !session.IsExpired() {
					// Check for admin authentication cookie
					adminCookie, _ := r.Cookie("is_admin_authenticated")
					if adminCookie != nil && adminCookie.Value == "true" {
						authenticated = true
						ctx = context.WithValue(ctx, isAdminContextKey, true)
						ctx = context.WithValue(ctx, sessionIDContextKey, session.ID())
					}
				}
			}
		}

		if !authenticated {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// OptionalAuth is a middleware that optionally authenticates
func (m *AuthMiddleware) OptionalAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		// Try to authenticate via session cookie
		sessionCookie, err := r.Cookie("session_id")
		if err == nil {
			session, err := m.checkSessionQuery.Execute(ctx, sessionCookie.Value)
			if err == nil && session != nil && !session.IsExpired() {
				adminCookie, _ := r.Cookie("is_admin_authenticated")
				if adminCookie != nil && adminCookie.Value == "true" {
					ctx = context.WithValue(ctx, isAdminContextKey, true)
					ctx = context.WithValue(ctx, sessionIDContextKey, session.ID())
				}
			}
		}

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetUserFromContext retrieves the user from the request context
func GetUserFromContext(ctx context.Context) (interface{}, bool) {
	user, ok := ctx.Value(userContextKey).(interface{})
	return user, ok
}

// GetIsAdminFromContext retrieves admin status from context
func GetIsAdminFromContext(ctx context.Context) bool {
	isAdmin, ok := ctx.Value(isAdminContextKey).(bool)
	return ok && isAdmin
}

// GetSessionIDFromContext retrieves session ID from context
func GetSessionIDFromContext(ctx context.Context) string {
	sessionID, ok := ctx.Value(sessionIDContextKey).(string)
	if !ok {
		return ""
	}
	return sessionID
}

// SetUserInContext sets the user in the request context
func SetUserInContext(ctx context.Context, user interface{}) context.Context {
	return context.WithValue(ctx, userContextKey, user)
}
