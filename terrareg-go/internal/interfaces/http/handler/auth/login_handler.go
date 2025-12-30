package auth

import (
	"encoding/json"
	"net/http"

	"github.com/rs/zerolog"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/application/command/auth"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/auth/service"
)

// LoginHandler handles authentication requests
type LoginHandler struct {
	createSessionCmd     *auth.CreateSessionCommand
	cookieSessionService *service.CookieSessionService
	logger               zerolog.Logger
}

// NewLoginHandler creates a new LoginHandler
func NewLoginHandler(
	createSessionCmd *auth.CreateSessionCommand,
	cookieSessionService *service.CookieSessionService,
	logger zerolog.Logger,
) *LoginHandler {
	return &LoginHandler{
		createSessionCmd:     createSessionCmd,
		cookieSessionService: cookieSessionService,
		logger:               logger,
	}
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

	// Create session
	sessionReq := &auth.CreateSessionRequest{
		Username:   req.Username,
		AuthMethod: "session_password",
		IsAdmin:    true, // TODO: Implement proper admin check
	}

	sessionResp, err := h.createSessionCmd.Execute(r.Context(), sessionReq)
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to create session")
		h.respondError(w, "Authentication failed", http.StatusUnauthorized)
		return
	}

	// Set session cookie
	sessionData := &service.SessionData{
		SessionID:  sessionResp.SessionID,
		UserID:     sessionResp.Username, // Use username as UserID for now
		Username:   sessionResp.Username,
		AuthMethod: sessionResp.AuthMethod,
		IsAdmin:    sessionResp.IsAdmin,
		SiteAdmin:  sessionResp.SiteAdmin,
		UserGroups: sessionResp.UserGroups,
		Expiry:     &sessionResp.Expiry,
	}
	if err := h.cookieSessionService.SetSessionCookie(w, sessionData); err != nil {
		h.logger.Error().Err(err).Msg("Failed to set session cookie")
		h.respondError(w, "Failed to create session", http.StatusInternalServerError)
		return
	}

	h.respondSuccess(w)
}

// HandleLogout handles logout requests
func (h *LoginHandler) HandleLogout(w http.ResponseWriter, r *http.Request) {
	// Clear session cookie
	h.cookieSessionService.ClearSessionCookie(w)
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
