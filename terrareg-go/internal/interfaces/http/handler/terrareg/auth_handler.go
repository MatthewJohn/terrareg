package terrareg

import (
	"net/http"

	authCmd "github.com/matthewjohn/terrareg/terrareg-go/internal/application/command/auth"
	authQuery "github.com/matthewjohn/terrareg/terrareg-go/internal/application/query/auth"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/config"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/auth/service"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/interfaces/http/dto"
)

// AuthHandler handles authentication-related requests
type AuthHandler struct {
	adminLoginCmd         *authCmd.AdminLoginCommand
	checkSessionQuery     *authQuery.CheckSessionQuery
	isAuthenticatedQuery  *authQuery.IsAuthenticatedQuery
	cookieSessionService  *service.CookieSessionService
	config                *config.Config
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(
	adminLoginCmd *authCmd.AdminLoginCommand,
	checkSessionQuery *authQuery.CheckSessionQuery,
	isAuthenticatedQuery *authQuery.IsAuthenticatedQuery,
	cookieSessionService *service.CookieSessionService,
	config *config.Config,
) *AuthHandler {
	return &AuthHandler{
		adminLoginCmd:         adminLoginCmd,
		checkSessionQuery:     checkSessionQuery,
		isAuthenticatedQuery:  isAuthenticatedQuery,
		cookieSessionService:  cookieSessionService,
		config:                config,
	}
}

// HandleAdminLogin handles POST /v1/terrareg/auth/admin/login
// Now follows DDD principles and matches Python API contract
func (h *AuthHandler) HandleAdminLogin(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Only allow POST requests
	if r.Method != http.MethodPost {
		RespondJSON(w, http.StatusMethodNotAllowed, map[string]interface{}{
			"message": "Method not allowed",
		})
		return
	}

	// Execute admin login command using the request (extracts X-Terrareg-ApiKey header)
	response, err := h.adminLoginCmd.ExecuteWithRequest(ctx, r)
	if err != nil {
		// Handle different error scenarios appropriately
		if !h.adminLoginCmd.IsConfigured() {
			RespondJSON(w, http.StatusForbidden, map[string]interface{}{
				"message": "Admin authentication is not enabled",
			})
		} else {
			RespondJSON(w, http.StatusUnauthorized, map[string]interface{}{
				"message": "Invalid API key",
			})
		}
		return
	}

	// If authentication failed, return 401
	if !response.Authenticated {
		RespondJSON(w, http.StatusUnauthorized, map[string]interface{}{
			"message": "Authentication failed",
		})
		return
	}

	// Create admin session using the session service (DDD-compliant approach)
	// The session service handles all cookie operations internally
	if err := h.cookieSessionService.CreateAdminSession(w, response.SessionID); err != nil {
		// If session creation fails, return error
		RespondJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"message": "Failed to create session",
		})
		return
	}

	// Respond with Python-compatible format
	RespondJSON(w, http.StatusOK, map[string]interface{}{
		"authenticated": response.Authenticated,
	})
}

// HandleIsAuthenticated handles GET /v1/terrareg/auth/admin/is_authenticated
func (h *AuthHandler) HandleIsAuthenticated(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get authentication status using the query
	response, err := h.isAuthenticatedQuery.Execute(ctx)
	if err != nil {
		// If there's an error, return unauthenticated status
		response = &dto.IsAuthenticatedResponse{
			Authenticated:        false,
			ReadAccess:          false,
			SiteAdmin:           false,
			NamespacePermissions: make(map[string]string),
		}
	}

	RespondJSON(w, http.StatusOK, response)
}

