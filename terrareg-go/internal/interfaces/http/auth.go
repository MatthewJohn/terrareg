package http

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"terrareg/internal/domain/identity"
)

// AuthHandler handles authentication and authorization HTTP requests
type AuthHandler struct {
	userService identity.UserService
}

// NewAuthHandler creates a new authentication handler
func NewAuthHandler(userService identity.UserService) *AuthHandler {
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
	Token        string     `json:"token,omitempty"`
	ExpiresAt    string     `json:"expires_at,omitempty"`
	User         *UserInfo  `json:"user,omitempty"`
	ErrorMessage string     `json:"error,omitempty"`
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
	if err := decodeJSON(w, r, &req); err != nil {
		return
	}

	// Authenticate user
	user, token, err := h.userService.Authenticate(r.Context(), req.AuthProvider, req.Token)
	if err != nil {
		sendErrorResponse(w, http.StatusUnauthorized, "authentication failed")
		return
	}

	response := LoginResponse{
		Success:   true,
		Token:     token.Token,
		ExpiresAt: token.ExpiresAt.Format(time.RFC3339),
		User: &UserInfo{
			ID:           user.ID,
			Username:     user.Username,
			DisplayName:  user.DisplayName,
			Email:        user.Email,
			AuthProvider: user.AuthProvider,
			CreatedAt:    user.CreatedAt.Format(time.RFC3339),
		},
	}

	sendJSONResponse(w, http.StatusOK, response)
}

// HandleLogout handles user logout
func (h *AuthHandler) HandleLogout(w http.ResponseWriter, r *http.Request) {
	token := extractBearerToken(r)
	if token == "" {
		sendErrorResponse(w, http.StatusBadRequest, "missing authorization header")
		return
	}

	if err := h.userService.Logout(r.Context(), token); err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, "logout failed")
		return
	}

	sendJSONResponse(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "logged out successfully",
	})
}

// HandleRefreshToken handles token refresh
func (h *AuthHandler) HandleRefreshToken(w http.ResponseWriter, r *http.Request) {
	token := extractBearerToken(r)
	if token == "" {
		sendErrorResponse(w, http.StatusBadRequest, "missing authorization header")
		return
	}

	newToken, err := h.userService.RefreshToken(r.Context(), token)
	if err != nil {
		sendErrorResponse(w, http.StatusUnauthorized, "token refresh failed")
		return
	}

	sendJSONResponse(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"token":   newToken.Token,
		"expires_at": newToken.ExpiresAt.Format(time.RFC3339),
	})
}

// HandleGetCurrentUser handles getting current user info
func (h *AuthHandler) HandleGetCurrentUser(w http.ResponseWriter, r *http.Request) {
	token := extractBearerToken(r)
	if token == "" {
		sendErrorResponse(w, http.StatusUnauthorized, "missing authorization header")
		return
	}

	user, err := h.userService.GetUserByToken(r.Context(), token)
	if err != nil {
		sendErrorResponse(w, http.StatusUnauthorized, "invalid token")
		return
	}

	userInfo := UserInfo{
		ID:           user.ID,
		Username:     user.Username,
		DisplayName:  user.DisplayName,
		Email:        user.Email,
		AuthProvider: user.AuthProvider,
		CreatedAt:    user.CreatedAt.Format(time.RFC3339),
	}

	sendJSONResponse(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"user":    userInfo,
	})
}

// HandleCheckAuth handles authentication check
func (h *AuthHandler) HandleCheckAuth(w http.ResponseWriter, r *http.Request) {
	token := extractBearerToken(r)
	if token == "" {
		sendJSONResponse(w, http.StatusOK, map[string]interface{}{
			"authenticated": false,
		})
		return
	}

	valid, err := h.userService.ValidateToken(r.Context(), token)
	if err != nil || !valid {
		sendJSONResponse(w, http.StatusOK, map[string]interface{}{
			"authenticated": false,
		})
		return
	}

	sendJSONResponse(w, http.StatusOK, map[string]interface{}{
		"authenticated": true,
	})
}