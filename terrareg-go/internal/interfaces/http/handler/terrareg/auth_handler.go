package terrareg

import (
	"net/http"

	authCmd "github.com/matthewjohn/terrareg/terrareg-go/internal/application/command/auth"
	authQuery "github.com/matthewjohn/terrareg/terrareg-go/internal/application/query/auth"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/config"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/auth/service"
)

// AuthHandler handles authentication-related requests
type AuthHandler struct {
	adminLoginCmd         *authCmd.AdminLoginCommand
	checkSessionQuery     *authQuery.CheckSessionQuery
	cookieSessionService  *service.CookieSessionService
	config                *config.Config
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(
	adminLoginCmd *authCmd.AdminLoginCommand,
	checkSessionQuery *authQuery.CheckSessionQuery,
	cookieSessionService *service.CookieSessionService,
	config *config.Config,
) *AuthHandler {
	return &AuthHandler{
		adminLoginCmd:         adminLoginCmd,
		checkSessionQuery:     checkSessionQuery,
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

	// Set session cookie using centralized service (DDD-compliant)
	// This centralizes cookie management used across all auth methods
	// HTTPS detection is handled centrally by the cookie service
	h.cookieSessionService.SetBasicSessionCookie(w, response.SessionID, response.Expiry)

	// Set admin authentication flag using centralized service
	h.cookieSessionService.SetAdminAuthenticationCookie(w, response.Authenticated)

	// Respond with Python-compatible format
	RespondJSON(w, http.StatusOK, map[string]interface{}{
		"authenticated": response.Authenticated,
	})
}

// HandleIsAuthenticated handles GET /v1/terrareg/auth/admin/is_authenticated
func (h *AuthHandler) HandleIsAuthenticated(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Check for session cookie
	sessionCookie, err := r.Cookie("session_id")
	if err != nil {
		RespondJSON(w, http.StatusOK, map[string]interface{}{
			"authenticated": false,
		})
		return
	}

	// Check session validity
	session, err := h.checkSessionQuery.Execute(ctx, sessionCookie.Value)
	if err != nil || session == nil || session.IsExpired() {
		RespondJSON(w, http.StatusOK, map[string]interface{}{
			"authenticated": false,
		})
		return
	}

	// Check admin authenticated cookie
	adminCookie, err := r.Cookie("is_admin_authenticated")
	isAdmin := err == nil && adminCookie.Value == "true"

	RespondJSON(w, http.StatusOK, map[string]interface{}{
		"authenticated": isAdmin,
	})
}

