package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/config"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/auth"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/auth/model"
	authservice "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/auth/service"
)

// contextKey is a custom type for context keys
type contextKey string

const (
	authMethodContextKey  contextKey = "auth_method"
	authMethodInstanceKey contextKey = "auth_method_instance"
	userContextKey        contextKey = "user"
	isAdminContextKey     contextKey = "is_admin"
	sessionIDContextKey   contextKey = "session_id"
	namespaceContextKey   contextKey = "namespace"
	permissionsContextKey contextKey = "permissions"
)

// AuthMiddleware provides authentication middleware
type AuthMiddleware struct {
	config      *config.Config
	authFactory *authservice.AuthFactory
}

// NewAuthMiddleware creates a new auth middleware
func NewAuthMiddleware(cfg *config.Config, authFactory *authservice.AuthFactory) *AuthMiddleware {
	return &AuthMiddleware{
		config:      cfg,
		authFactory: authFactory,
	}
}

// RequireAuth is a middleware that requires authentication
func (m *AuthMiddleware) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		// Extract headers and form data for authentication
		headers := make(map[string]string)
		formData := make(map[string]string)
		queryParams := make(map[string]string)

		// Copy all headers
		for name, values := range r.Header {
			if len(values) > 0 {
				headers[name] = values[0]
			}
		}

		// Copy query parameters
		for name, values := range r.URL.Query() {
			if len(values) > 0 {
				queryParams[name] = values[0]
			}
		}

		// Use AuthFactory to authenticate the request
		authResponse, err := m.authFactory.AuthenticateRequest(ctx, headers, formData, queryParams)
		if err != nil || authResponse == nil || !authResponse.Success {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Set authentication context
		ctx = context.WithValue(ctx, authMethodContextKey, authResponse.AuthMethod)
		ctx = context.WithValue(ctx, authMethodInstanceKey, m.authFactory.GetCurrentAuthMethod())
		ctx = context.WithValue(ctx, userContextKey, authResponse.Username)
		ctx = context.WithValue(ctx, isAdminContextKey, authResponse.IsAdmin)
		ctx = context.WithValue(ctx, permissionsContextKey, authResponse.Permissions)

		// Set session ID if available
		if authResponse.SessionID != nil {
			ctx = context.WithValue(ctx, sessionIDContextKey, *authResponse.SessionID)
		}

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// OptionalAuth is a middleware that optionally authenticates
func (m *AuthMiddleware) OptionalAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		// Extract headers and form data for authentication
		headers := make(map[string]string)
		formData := make(map[string]string)
		queryParams := make(map[string]string)

		// Copy all headers
		for name, values := range r.Header {
			if len(values) > 0 {
				headers[name] = values[0]
			}
		}

		// Copy query parameters
		for name, values := range r.URL.Query() {
			if len(values) > 0 {
				queryParams[name] = values[0]
			}
		}

		// Use AuthFactory to authenticate the request (optional)
		authResponse, err := m.authFactory.AuthenticateRequest(ctx, headers, formData, queryParams)
		if err == nil && authResponse != nil && authResponse.Success {
			// Set authentication context if successful
			ctx = context.WithValue(ctx, authMethodContextKey, authResponse.AuthMethod)
			ctx = context.WithValue(ctx, userContextKey, authResponse.Username)
			ctx = context.WithValue(ctx, isAdminContextKey, authResponse.IsAdmin)
			ctx = context.WithValue(ctx, permissionsContextKey, authResponse.Permissions)

			// Set session ID if available
			if authResponse.SessionID != nil {
				ctx = context.WithValue(ctx, sessionIDContextKey, *authResponse.SessionID)
			}
		}

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetAuthMethodFromContext retrieves the auth method from the request context
func GetAuthMethodFromContext(ctx context.Context) (auth.AuthMethodType, bool) {
	authMethod, ok := ctx.Value(authMethodContextKey).(auth.AuthMethodType)
	return authMethod, ok
}

// GetAuthMethodInstanceFromContext retrieves the auth method instance from the request context
func GetAuthMethodInstanceFromContext(ctx context.Context) (auth.AuthMethod, bool) {
	authMethod, ok := ctx.Value(authMethodInstanceKey).(auth.AuthMethod)
	return authMethod, ok
}

// GetUserFromContext retrieves the user from the request context
func GetUserFromContext(ctx context.Context) (string, bool) {
	user, ok := ctx.Value(userContextKey).(string)
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

// GetPermissionsFromContext retrieves user permissions from context
func GetPermissionsFromContext(ctx context.Context) (map[string]string, bool) {
	permissions, ok := ctx.Value(permissionsContextKey).(map[string]string)
	return permissions, ok
}

// CheckNamespacePermission checks if the current user has permission for a namespace
func (m *AuthMiddleware) CheckNamespacePermission(ctx context.Context, permissionType, namespace string) bool {
	permissions, ok := GetPermissionsFromContext(ctx)
	if !ok {
		return false
	}

	// Check if user is admin
	if GetIsAdminFromContext(ctx) {
		return true
	}

	// Check specific namespace permission
	storedPermission, exists := permissions[namespace]
	if !exists {
		return false
	}

	// Check permission hierarchy
	switch permissionType {
	case "READ":
		return storedPermission == "READ" || storedPermission == "MODIFY" || storedPermission == "FULL"
	case "MODIFY":
		return storedPermission == "MODIFY" || storedPermission == "FULL"
	case "FULL":
		return storedPermission == "FULL"
	default:
		return false
	}
}

// RequireAdmin is a middleware that requires admin authentication
func (m *AuthMiddleware) RequireAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		// Extract headers and form data for authentication
		headers := make(map[string]string)
		formData := make(map[string]string)
		queryParams := make(map[string]string)

		// Copy all headers
		for name, values := range r.Header {
			if len(values) > 0 {
				headers[name] = values[0]
			}
		}

		// Copy query parameters
		for name, values := range r.URL.Query() {
			if len(values) > 0 {
				queryParams[name] = values[0]
			}
		}

		// Use AuthFactory to authenticate the request
		authResponse, err := m.authFactory.AuthenticateRequest(ctx, headers, formData, queryParams)
		if err != nil || authResponse == nil || !authResponse.Success || !authResponse.IsAdmin {
			http.Error(w, "Admin access required", http.StatusForbidden)
			return
		}

		// Set authentication context
		ctx = context.WithValue(ctx, authMethodContextKey, authResponse.AuthMethod)
		ctx = context.WithValue(ctx, authMethodInstanceKey, m.authFactory.GetCurrentAuthMethod())
		ctx = context.WithValue(ctx, userContextKey, authResponse.Username)
		ctx = context.WithValue(ctx, isAdminContextKey, authResponse.IsAdmin)
		ctx = context.WithValue(ctx, permissionsContextKey, authResponse.Permissions)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RequireNamespacePermission creates middleware that requires specific namespace permission
func (m *AuthMiddleware) RequireNamespacePermission(permissionType, namespace string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			// Extract headers and form data for authentication
			headers := make(map[string]string)
			formData := make(map[string]string)
			queryParams := make(map[string]string)

			// Copy all headers
			for name, values := range r.Header {
				if len(values) > 0 {
					headers[name] = values[0]
				}
			}

			// Copy query parameters
			for name, values := range r.URL.Query() {
				if len(values) > 0 {
					queryParams[name] = values[0]
				}
			}

			// Use AuthFactory to authenticate the request
			authResponse, err := m.authFactory.AuthenticateRequest(ctx, headers, formData, queryParams)
			if err != nil || authResponse == nil || !authResponse.Success {
				http.Error(w, "Authentication required", http.StatusUnauthorized)
				return
			}

			// Check namespace permission
			if !m.checkNamespacePermissionInResponse(authResponse, permissionType, namespace) {
				http.Error(w, "Insufficient permissions", http.StatusForbidden)
				return
			}

			// Set authentication context
			ctx = context.WithValue(ctx, authMethodContextKey, authResponse.AuthMethod)
			ctx = context.WithValue(ctx, userContextKey, authResponse.Username)
			ctx = context.WithValue(ctx, isAdminContextKey, authResponse.IsAdmin)
			ctx = context.WithValue(ctx, permissionsContextKey, authResponse.Permissions)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// checkNamespacePermissionInResponse checks permission in an authentication response
func (m *AuthMiddleware) checkNamespacePermissionInResponse(authResponse *model.AuthenticationResponse, permissionType, namespace string) bool {
	// Check if user is admin
	if authResponse.IsAdmin {
		return true
	}

	// Check specific namespace permission
	storedPermission, exists := authResponse.Permissions[namespace]
	if !exists {
		return false
	}

	// Check permission hierarchy
	switch permissionType {
	case "READ":
		return storedPermission == "READ" || storedPermission == "MODIFY" || storedPermission == "FULL"
	case "MODIFY":
		return storedPermission == "MODIFY" || storedPermission == "FULL"
	case "FULL":
		return storedPermission == "FULL"
	default:
		return false
	}
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
