package http

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	identityService "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/identity/service"
	httputils "github.com/matthewjohn/terrareg/terrareg-go/internal/interfaces/http/utils"
)

// AuthHandler handles authentication and authorization HTTP requests
type AuthHandler struct {
	userService identityService.UserService
}

// NewAuthHandler creates a new authentication handler
func NewAuthHandler(userService identityService.UserService) *AuthHandler {
	return &AuthHandler{
		userService: userService,
	}
}

// RegisterRoutes registers authentication routes
func (h *AuthHandler) RegisterRoutes(r chi.Router) {
	r.Route("/auth", func(r chi.Router) {
		r.Post("/login", h.HandleLogin)
		r.Post("/logout", h.HandleLogout)
		r.Post("/refresh", h.HandleRefreshToken)
		r.Get("/me", h.HandleGetCurrentUser)
		r.Post("/check", h.HandleCheckAuth)
	})
}

// LoginRequest represents a login request
type LoginRequest struct {
	AuthProvider string `json:"auth_provider"`
	Token        string `json:"token,omitempty"`
}

// LoginResponse represents a successful login response
type LoginResponse struct {
	Success      bool      `json:"success"`
	Token        string    `json:"token,omitempty"`
	ExpiresAt    string    `json:"expires_at,omitempty"`
	User         *UserInfo `json:"user,omitempty"`
	ErrorMessage string    `json:"error,omitempty"`
}

// UserInfo represents user information for API responses
type UserInfo struct {
	ID           string `json:"id"`
	Username     string `json:"username"`
	DisplayName  string `json:"display_name"`
	Email        string `json:"email,omitempty"`
	AuthProvider string `json:"auth_provider"`
	CreatedAt    string `json:"created_at"`
}

// HandleLogin handles user login
func (h *AuthHandler) HandleLogin(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := httputils.DecodeJSON(w, r, &req); err != nil {
		return
	}

	// For now, we'll implement a simple auth check
	// TODO: Implement proper OAuth/OIDC authentication flow
	if req.AuthProvider == "" || req.Token == "" {
		httputils.SendErrorResponse(w, http.StatusBadRequest, "auth_provider and token are required")
		return
	}

	// This is a placeholder - you'll need to implement proper authentication
	// For now, return an error to indicate this needs to be implemented
	httputils.SendErrorResponse(w, http.StatusNotImplemented, "authentication method not yet implemented")
}

// HandleLogout handles user logout
func (h *AuthHandler) HandleLogout(w http.ResponseWriter, r *http.Request) {
	token := httputils.ExtractBearerToken(r)
	if token == "" {
		httputils.SendErrorResponse(w, http.StatusBadRequest, "missing authorization header")
		return
	}

	// TODO: Implement logout method in UserService
	httputils.SendErrorResponse(w, http.StatusNotImplemented, "logout method not yet implemented")
}

// HandleRefreshToken handles token refresh
func (h *AuthHandler) HandleRefreshToken(w http.ResponseWriter, r *http.Request) {
	token := httputils.ExtractBearerToken(r)
	if token == "" {
		httputils.SendErrorResponse(w, http.StatusBadRequest, "missing authorization header")
		return
	}

	newToken, err := h.userService.RefreshToken(r.Context(), token)
	if err != nil {
		httputils.SendErrorResponse(w, http.StatusUnauthorized, "token refresh failed")
		return
	}

	httputils.SendJSONResponse(w, http.StatusOK, map[string]interface{}{
		"success":    true,
		"token":      newToken.Token,
		"expires_at": newToken.ExpiresAt().Format(time.RFC3339),
	})
}

// HandleGetCurrentUser handles getting current user info
func (h *AuthHandler) HandleGetCurrentUser(w http.ResponseWriter, r *http.Request) {
	token := httputils.ExtractBearerToken(r)
	if token == "" {
		httputils.SendErrorResponse(w, http.StatusUnauthorized, "missing authorization header")
		return
	}

	user, err := h.userService.GetUserByToken(r.Context(), token)
	if err != nil {
		httputils.SendErrorResponse(w, http.StatusUnauthorized, "invalid token")
		return
	}

	userInfo := UserInfo{
		ID:           user.ID(),
		Username:     user.Username(),
		DisplayName:  user.DisplayName(),
		Email:        user.Email(),
		AuthProvider: user.AuthMethod().String(),
		CreatedAt:    user.CreatedAt().Format(time.RFC3339),
	}

	httputils.SendJSONResponse(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"user":    userInfo,
	})
}

// HandleCheckAuth handles authentication check
func (h *AuthHandler) HandleCheckAuth(w http.ResponseWriter, r *http.Request) {
	token := httputils.ExtractBearerToken(r)
	if token == "" {
		httputils.SendJSONResponse(w, http.StatusOK, map[string]interface{}{
			"authenticated": false,
		})
		return
	}

	_, err := h.userService.ValidateToken(r.Context(), token)
	if err != nil {
		httputils.SendJSONResponse(w, http.StatusOK, map[string]interface{}{
			"authenticated": false,
		})
		return
	}

	httputils.SendJSONResponse(w, http.StatusOK, map[string]interface{}{
		"authenticated": true,
	})
}
