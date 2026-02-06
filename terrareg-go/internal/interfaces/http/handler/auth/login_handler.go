package auth

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/rs/zerolog"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/auth"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/auth/service"
)

// LoginHandler handles authentication requests
type LoginHandler struct {
	// sessionManagementService handles session and cookie operations (required)
	sessionManagementService *service.SessionManagementService
	// logger for logging (required)
	logger zerolog.Logger
}

// NewLoginHandler creates a new LoginHandler
// Returns an error if any required dependency is nil
func NewLoginHandler(
	sessionManagementService *service.SessionManagementService,
	logger zerolog.Logger,
) (*LoginHandler, error) {
	if sessionManagementService == nil {
		return nil, fmt.Errorf("sessionManagementService cannot be nil")
	}

	return &LoginHandler{
		sessionManagementService: sessionManagementService,
		logger:                   logger,
	}, nil
}

// LoginRequest represents a login request
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// LoginResponse represents a login response
type LoginResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
}

// HandleLogin handles admin login requests
func (h *LoginHandler) HandleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// TODO: Implement actual password validation
	// For now, we'll create a session for any non-empty username/password
	if req.Username == "" || req.Password == "" {
		h.respondError(w, "Username and password are required", http.StatusBadRequest)
		return
	}

	// Create session and set cookie in one operation
	err := h.sessionManagementService.CreateSessionAndCookie(
		r.Context(),
		w,
		auth.AuthMethodSession,
		req.Username,
		true, // TODO: Implement proper admin check
		nil, // userGroups
		nil, // permissions
		nil, // providerData
		nil, // ttl - will use default
	)
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to create session and cookie")
		h.respondError(w, "Authentication failed", http.StatusUnauthorized)
		return
	}

	h.respondSuccess(w)
}

// HandleLogout handles logout requests
func (h *LoginHandler) HandleLogout(w http.ResponseWriter, r *http.Request) {
	// Clear session and cookie
	if err := h.sessionManagementService.ClearSessionAndCookie(r.Context(), w, r); err != nil {
		h.logger.Error().Err(err).Msg("Failed to clear session and cookie")
		h.respondError(w, "Failed to clear session", http.StatusInternalServerError)
		return
	}
	h.respondSuccess(w)
}

// respondSuccess sends a successful response
func (h *LoginHandler) respondSuccess(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(LoginResponse{
		Success: true,
	})
}

// respondError sends an error response
func (h *LoginHandler) respondError(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(LoginResponse{
		Success: false,
		Message: message,
	})
}
