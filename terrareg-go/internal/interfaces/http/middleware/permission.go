package middleware

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"
	identityModel "terrareg/internal/domain/identity/model"
	identityService "terrareg/internal/domain/identity/service"
)

// PermissionMiddleware handles permission-based access control
type PermissionMiddleware struct {
	userService *identityService.UserService
}

// NewPermissionMiddleware creates a new permission middleware
func NewPermissionMiddleware(userService *identityService.UserService) *PermissionMiddleware {
	return &PermissionMiddleware{
		userService: userService,
	}
}

// RequirePermission creates middleware that requires specific permission
func (m *PermissionMiddleware) RequirePermission(resourceType identityModel.ResourceType, action identityModel.Action) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user, err := m.getUserFromContext(r.Context())
			if err != nil {
				sendErrorResponse(w, http.StatusUnauthorized, "authentication required")
				return
			}

			// Extract resource ID from URL path
			resourceID := m.extractResourceID(r, chi.URLParam(r, "*"))

			// Check permission
			hasPermission, err := m.userService.CheckPermission(r.Context(), user.ID(), resourceType, resourceID, action)
			if err != nil {
				sendErrorResponse(w, http.StatusInternalServerError, "permission check failed")
				return
			}

			if !hasPermission {
				sendErrorResponse(w, http.StatusForbidden, "insufficient permissions")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// RequireUser creates middleware that requires authentication
func (m *PermissionMiddleware) RequireUser(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, err := m.getUserFromContext(r.Context())
		if err != nil {
			sendErrorResponse(w, http.StatusUnauthorized, "authentication required")
			return
		}

		if !user.Active() {
			sendErrorResponse(w, http.StatusUnauthorized, "user account is inactive")
			return
		}

		next.ServeHTTP(w, r)
	})
}

// OptionalAuth creates middleware that optionally authenticates user
func (m *PermissionMiddleware) OptionalAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Try to get user from context, but don't require it
		user, _ := m.getUserFromContext(r.Context())

		// Add user to context if found
		if user != nil {
			ctx := context.WithValue(r.Context(), "user", user)
			r = r.WithContext(ctx)
		}

		next.ServeHTTP(w, r)
	})
}

// RequireAdmin creates middleware that requires admin privileges
func (m *PermissionMiddleware) RequireAdmin(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user, err := m.getUserFromContext(r.Context())
			if err != nil {
				sendErrorResponse(w, http.StatusUnauthorized, "authentication required")
				return
			}

			// Check if user has admin privileges
			hasAdminPermission, err := m.userService.CheckPermission(r.Context(), user.ID(), "system", "*", identityModel.ActionAdmin)
			if err != nil {
				sendErrorResponse(w, http.StatusInternalServerError, "permission check failed")
				return
			}

			if !hasAdminPermission {
				sendErrorResponse(w, http.StatusForbidden, "admin privileges required")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// getUserFromContext extracts user from request context
func (m *PermissionMiddleware) getUserFromContext(ctx context.Context) (*identityModel.User, error) {
	user, ok := ctx.Value("user").(*identityModel.User)
	if !ok {
		return nil, ErrUserNotInContext
	}
	return user, nil
}

// extractResourceID extracts resource ID from request
func (m *PermissionMiddleware) extractResourceID(r *http.Request, pathParams map[string]string) string {
	// Try to extract ID from common patterns
	if id := chi.URLParam(r, "id"); id != "" {
		return id
	}
	if id := chi.URLParam(r, "namespace"); id != "" {
		return id
	}
	if id := chi.URLParam(r, "module"); id != "" {
		return id
	}
	if id := chi.URLParam(r, "provider"); id != "" {
		return id
	}

	// Extract from query parameter for some APIs
	if id := r.URL.Query().Get("id"); id != "" {
		return id
	}

	return "*"
}


// Permission types for middleware
const (
	ResourceTypeNamespace = "namespace"
	ResourceTypeModule    = "module"
	ResourceTypeProvider  = "provider"
)

// GetUserFromContext helper function to get user from context
func GetUserFromContext(ctx context.Context) (*identityModel.User, bool) {
	user, ok := ctx.Value("user").(*identityModel.User)
	return user, ok
}

// GetUserIDFromContext helper function to get user ID from context
func GetUserIDFromContext(ctx context.Context) (string, bool) {
	user, ok := ctx.Value("user").(*identityModel.User)
	if !ok {
		return "", false
	}
	return user.ID(), true
}