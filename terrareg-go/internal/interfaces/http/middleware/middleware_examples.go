package middleware

import (
	"encoding/json"
	"net/http"
)

// Example usage of the authentication middleware

// ExampleProtectedRoute shows how to protect a route with authentication
func ExampleProtectedRoute(middleware *AuthMiddleware) http.Handler {
	// This requires any valid authentication
	return middleware.RequireAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get user information from context
		username, _ := GetUserFromContext(r.Context())
		isAdmin := GetIsAdminFromContext(r.Context())
		permissions, _ := GetPermissionsFromContext(r.Context())

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		response := map[string]interface{}{
			"message":     "Authenticated access granted",
			"username":    username,
			"is_admin":    isAdmin,
			"permissions": permissions,
		}

		_ = json.NewEncoder(w).Encode(response)
	}))
}

// ExampleAdminRoute shows how to protect a route requiring admin access
func ExampleAdminRoute(middleware *AuthMiddleware) http.Handler {
	// This requires admin authentication
	return middleware.RequireAdmin(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		username, _ := GetUserFromContext(r.Context())

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		response := map[string]interface{}{
			"message":  "Admin access granted",
			"username": username,
		}
		_ = json.NewEncoder(w).Encode(response)
	}))
}

// ExampleNamespaceProtectedRoute shows how to protect a route requiring specific namespace permissions
func ExampleNamespaceProtectedRoute(middleware *AuthMiddleware) http.Handler {
	// This requires FULL permission on the "example" namespace
	return middleware.RequireNamespacePermission("FULL", "example")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		username, _ := GetUserFromContext(r.Context())

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		response := map[string]interface{}{
			"message":  "Namespace access granted for 'example' namespace",
			"username": username,
		}
		_ = json.NewEncoder(w).Encode(response)
	}))
}

// ExampleOptionalAuthRoute shows how to have optional authentication
func ExampleOptionalAuthRoute(middleware *AuthMiddleware) http.Handler {
	// This works with or without authentication
	return middleware.OptionalAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		username, authenticated := GetUserFromContext(r.Context())

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		response := map[string]interface{}{
			"message": "Optional access",
		}

		if authenticated {
			response["username"] = username
			response["authenticated"] = true
		} else {
			response["authenticated"] = false
		}

		_ = json.NewEncoder(w).Encode(response)
	}))
}

// ExampleRouteHandlerWithPermissionCheck shows how to manually check permissions within a handler
func ExampleRouteHandlerWithPermissionCheck(middleware *AuthMiddleware) http.Handler {
	return middleware.RequireAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		// Manually check if user can publish to the "production" namespace
		canPublish := middleware.CheckNamespacePermission(ctx, "FULL", "production")

		if !canPublish {
			w.WriteHeader(http.StatusForbidden)
			_ = json.NewEncoder(w).Encode(map[string]string{
				"error": "Insufficient permissions for production namespace",
			})
			return
		}

		// User has permission, proceed with the operation
		username, _ := GetUserFromContext(ctx)

		response := map[string]interface{}{
			"message":  "Operation completed successfully",
			"username": username,
		}

		_ = json.NewEncoder(w).Encode(response)
	}))
}

// ExampleRouteSetup shows how to set up routes with different protection levels
func ExampleRouteSetup(middleware *AuthMiddleware) {
	mux := http.NewServeMux()

	// Public route (no authentication required)
	mux.HandleFunc("/public", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Public access - no authentication required"))
	})

	// Protected route (any valid authentication)
	mux.Handle("/protected", ExampleProtectedRoute(middleware))

	// Admin-only route
	mux.Handle("/admin", ExampleAdminRoute(middleware))

	// Namespace-protected route
	mux.Handle("/modules/example/manage", ExampleNamespaceProtectedRoute(middleware))

	// Optional authentication route
	mux.Handle("/optional", ExampleOptionalAuthRoute(middleware))

	// Route with manual permission checking
	mux.Handle("/check-permissions", ExampleRouteHandlerWithPermissionCheck(middleware))
}

// ExampleRouteProtectionDocumentation provides examples for developers

/*
HTTP Authentication Middleware Usage Examples:

1. Basic Authentication Protection:
   middleware.RequireAuth(handler)
   - Requires any valid authentication method
   - Rejects unauthenticated requests with 401 Unauthorized

2. Admin-Only Access:
   middleware.RequireAdmin(handler)
   - Requires authenticated user with admin privileges
   - Rejects non-admin users with 403 Forbidden

3. Namespace-Specific Permissions:
   middleware.RequireNamespacePermission("FULL", "my-namespace")(handler)
   - Requires specific permission level for a namespace
   - Supports permission hierarchy: READ < MODIFY < FULL
   - Admins bypass namespace restrictions

4. Optional Authentication:
   middleware.OptionalAuth(handler)
   - Attempts authentication but doesn't require it
   - Handler can check authentication status using context helpers

5. Manual Permission Checking:
   middleware.CheckNamespacePermission(ctx, "MODIFY", "namespace")
   - Used within handlers for dynamic permission checks
   - Returns boolean result

Context Helper Functions:
- GetUserFromContext(ctx) (string, bool) - Get authenticated username
- GetIsAdminFromContext(ctx) bool - Check if user is admin
- GetPermissionsFromContext(ctx) (map[string]string, bool) - Get all namespace permissions
- GetAuthMethodFromContext(ctx) (auth.AuthMethodType, bool) - Get authentication method type
- GetSessionIDFromContext(ctx) string - Get session ID if available

Authentication Methods Supported:
1. Admin API Keys (Bearer tokens)
2. Admin Session Authentication
3. SAML SSO
4. OpenID Connect
5. Terraform OIDC (for CLI access)
6. Upload API Keys
7. Publish API Keys
8. Terraform Analytics Auth Keys
9. Terraform Internal Extraction
10. NotAuthenticated (fallback)

Example Request Headers:
- Admin API Key: Authorization: Bearer admin-api-key-12345
- Upload API Key: Authorization: Bearer upload-api-key-67890
- Terraform OIDC: Authorization: Bearer terraform-access-token

Example Response Context:
The middleware sets the following context values:
- auth_method: The authentication method type used
- user: The authenticated username
- is_admin: Boolean indicating admin status
- permissions: Map of namespace -> permission_type
- session_id: Session ID if available
*/
