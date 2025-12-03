package middleware

import (
	"context"
	"encoding/json"
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

// decodeJSON parses JSON from request body
func decodeJSON(w http.ResponseWriter, r *http.Request, v interface{}) error {
	decoder := json.NewDecoder(r.Body)
	return decoder.Decode(v)
}

// sendJSONResponse sends a JSON response
func sendJSONResponse(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}

// sendErrorResponse sends a JSON error response
func sendErrorResponse(w http.ResponseWriter, statusCode int, message string) {
	sendJSONResponse(w, statusCode, map[string]interface{}{
		"error": message,
	})
}

// extractBearerToken extracts Bearer token from Authorization header
func extractBearerToken(r *http.Request) string {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return ""
	}

	// Extract Bearer token
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || parts[0] != "Bearer" {
		return ""
	}

	return parts[1]
}
