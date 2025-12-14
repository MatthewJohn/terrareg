package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/auth"
	domainAuthModel "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/auth/model"
	authservice "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/auth/service"
	domainConfig "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/config/model"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/interfaces/http/middleware/model"
)

// authMiddlewareKeyType is a custom type for auth middleware context keys
type authMiddlewareKeyType string

const (
	middlewareAuthContextKey authMiddlewareKeyType = "auth_middleware_context"
)

// AuthMiddleware provides authentication middleware
type AuthMiddleware struct {
	domainConfig *domainConfig.DomainConfig
	authFactory  *authservice.AuthFactory
}

// NewAuthMiddleware creates a new auth middleware
func NewAuthMiddleware(cfg *domainConfig.DomainConfig, authFactory *authservice.AuthFactory) *AuthMiddleware {
	return &AuthMiddleware{
		domainConfig: cfg,
		authFactory:  authFactory,
	}
}

// createAuthContext creates a structured AuthContext from authentication response
func (m *AuthMiddleware) createAuthContext(authResponse *domainAuthModel.AuthenticationResponse) *model.AuthContext {
	if authResponse == nil || !authResponse.Success {
		return &model.AuthContext{
			IsAuthenticated: false,
		}
	}

	authCtx := &model.AuthContext{
		AuthMethod:      authResponse.AuthMethod,
		Username:        authResponse.Username,
		IsAdmin:         authResponse.IsAdmin,
		Permissions:     authResponse.Permissions,
		IsAuthenticated: true,
	}

	// Set session ID if available
	if authResponse.SessionID != nil {
		authCtx.SessionID = *authResponse.SessionID
	}

	return authCtx
}

// extractRequestData extracts headers, form data, and query params from request
func (m *AuthMiddleware) extractRequestData(r *http.Request) (map[string]string, map[string]string, map[string]string) {
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

	return headers, formData, queryParams
}

// authenticateRequest performs common authentication logic
func (m *AuthMiddleware) authenticateRequest(ctx context.Context, headers, formData, queryParams map[string]string) (*domainAuthModel.AuthenticationResponse, error) {
	return m.authFactory.AuthenticateRequest(ctx, headers, formData, queryParams)
}

// RequireAuth is a middleware that requires authentication
func (m *AuthMiddleware) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		// Extract request data
		headers, formData, queryParams := m.extractRequestData(r)

		// Authenticate the request
		authResponse, err := m.authenticateRequest(ctx, headers, formData, queryParams)
		if err != nil || authResponse == nil || !authResponse.Success {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Create and set auth context
		authCtx := m.createAuthContext(authResponse)
		ctx = context.WithValue(ctx, middlewareAuthContextKey, authCtx)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// OptionalAuth is a middleware that optionally authenticates
func (m *AuthMiddleware) OptionalAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		// Extract request data
		headers, formData, queryParams := m.extractRequestData(r)

		// Try to authenticate the request (optional)
		if authResponse, err := m.authenticateRequest(ctx, headers, formData, queryParams); err == nil {
			authCtx := m.createAuthContext(authResponse)
			ctx = context.WithValue(ctx, middlewareAuthContextKey, authCtx)
		} else {
			// Set empty auth context for consistency
			authCtx := &model.AuthContext{IsAuthenticated: false}
			ctx = context.WithValue(ctx, middlewareAuthContextKey, authCtx)
		}

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetAuthContext retrieves the complete auth context from the request context
func GetAuthContext(ctx context.Context) *model.AuthContext {
	if authCtx, ok := ctx.Value(middlewareAuthContextKey).(*model.AuthContext); ok {
		return authCtx
	}
	return &model.AuthContext{IsAuthenticated: false}
}

// GetAuthMethodFromContext retrieves the auth method from the request context
func GetAuthMethodFromContext(ctx context.Context) (auth.AuthMethodType, bool) {
	authCtx := GetAuthContext(ctx)
	return authCtx.AuthMethod, authCtx.IsAuthenticated
}

// GetUserFromContext retrieves the user from the request context
func GetUserFromContext(ctx context.Context) (string, bool) {
	authCtx := GetAuthContext(ctx)
	return authCtx.Username, authCtx.IsAuthenticated
}

// GetIsAdminFromContext retrieves admin status from context
func GetIsAdminFromContext(ctx context.Context) bool {
	authCtx := GetAuthContext(ctx)
	return authCtx.IsAuthenticated && authCtx.IsAdmin
}

// GetSessionIDFromContext retrieves session ID from context
func GetSessionIDFromContext(ctx context.Context) string {
	authCtx := GetAuthContext(ctx)
	if authCtx.IsAuthenticated {
		return authCtx.SessionID
	}
	return ""
}

// GetPermissionsFromContext retrieves user permissions from context
func GetPermissionsFromContext(ctx context.Context) (map[string]string, bool) {
	authCtx := GetAuthContext(ctx)
	if authCtx.IsAuthenticated {
		return authCtx.Permissions, true
	}
	return nil, false
}

// CheckNamespacePermission checks if the current user has permission for a namespace
func (m *AuthMiddleware) CheckNamespacePermission(ctx context.Context, permissionType, namespace string) bool {
	authCtx := GetAuthContext(ctx)
	return authCtx.HasPermission(permissionType, namespace)
}

// RequireAdmin is a middleware that requires admin authentication
func (m *AuthMiddleware) RequireAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		// Extract request data
		headers, formData, queryParams := m.extractRequestData(r)

		// Authenticate the request
		authResponse, err := m.authenticateRequest(ctx, headers, formData, queryParams)
		if err != nil || authResponse == nil || !authResponse.Success || !authResponse.IsAdmin {
			http.Error(w, "Admin access required", http.StatusForbidden)
			return
		}

		// Create and set auth context
		authCtx := m.createAuthContext(authResponse)
		ctx = context.WithValue(ctx, middlewareAuthContextKey, authCtx)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RequireNamespacePermission creates middleware that requires specific namespace permission
func (m *AuthMiddleware) RequireNamespacePermission(permissionType, namespaceParam string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			// Extract request data
			headers, formData, queryParams := m.extractRequestData(r)

			// Authenticate the request
			authResponse, err := m.authenticateRequest(ctx, headers, formData, queryParams)
			if err != nil || authResponse == nil || !authResponse.Success {
				http.Error(w, "Authentication required", http.StatusUnauthorized)
				return
			}

			// Extract namespace name from URL parameter if needed
			namespaceName := namespaceParam
			if strings.HasPrefix(namespaceParam, "{") && strings.HasSuffix(namespaceParam, "}") {
				// Extract from URL path (e.g., "{namespace}" -> actual namespace value)
				paramName := namespaceParam[1 : len(namespaceParam)-1]
				namespaceName = chi.URLParam(r, paramName)
			}

			if namespaceName == "" {
				http.Error(w, "Namespace parameter required", http.StatusBadRequest)
				return
			}

			// Create auth context and check namespace permission
			authCtx := m.createAuthContext(authResponse)
			if !authCtx.HasPermission(permissionType, namespaceName) {
				http.Error(w, "Insufficient permissions", http.StatusForbidden)
				return
			}

			// Set auth context
			ctx = context.WithValue(ctx, middlewareAuthContextKey, authCtx)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// SetAuthContextInContext sets the auth context in the request context
func SetAuthContextInContext(ctx context.Context, authCtx *model.AuthContext) context.Context {
	return context.WithValue(ctx, middlewareAuthContextKey, authCtx)
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
