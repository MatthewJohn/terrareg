package auth

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/rs/zerolog"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/auth/service"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/interfaces/http/middleware"
)

// SessionHandler handles session management requests
type SessionHandler struct {
	cookieSessionService *service.CookieSessionService
	logger               zerolog.Logger
}

// NewSessionHandler creates a new SessionHandler
func NewSessionHandler(
	cookieSessionService *service.CookieSessionService,
	logger zerolog.Logger,
) *SessionHandler {
	return &SessionHandler{
		cookieSessionService: cookieSessionService,
		logger:               logger,
	}
}

// GetSessionResponse represents a session response
type GetSessionResponse struct {
	Authenticated bool     `json:"authenticated"`
	UserID        string   `json:"user_id,omitempty"`
	Username      string   `json:"username,omitempty"`
	AuthMethod    string   `json:"auth_method,omitempty"`
	IsAdmin       bool     `json:"is_admin,omitempty"`
	UserGroups    []string `json:"user_groups,omitempty"`
}

// HandleGetSession handles requests to get current session info
func (h *SessionHandler) HandleGetSession(w http.ResponseWriter, r *http.Request) {
	sessionData := middleware.GetSessionData(r.Context())

	var response GetSessionResponse
	if sessionData != nil {
		response = GetSessionResponse{
			Authenticated: true,
			UserID:        sessionData.UserID,
			Username:      sessionData.Username,
			AuthMethod:    sessionData.AuthMethod,
			IsAdmin:       sessionData.IsAdmin,
			UserGroups:    sessionData.UserGroups,
		}
	} else {
		response = GetSessionResponse{
			Authenticated: false,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// HandleDeleteSession handles requests to delete the current session (logout)
func (h *SessionHandler) HandleDeleteSession(w http.ResponseWriter, r *http.Request) {
	sessionData := middleware.GetSessionData(r.Context())
	if sessionData == nil {
		http.Error(w, "No active session", http.StatusBadRequest)
		return
	}

	// Delete session from database
	if err := h.cookieSessionService.DeleteSession(r.Context(), sessionData.SessionID); err != nil {
		h.logger.Error().Err(err).Msg("Failed to delete session from database")
		http.Error(w, "Failed to delete session", http.StatusInternalServerError)
		return
	}

	// Clear session cookie
	h.cookieSessionService.ClearSessionCookie(w)

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]bool{
		"success": true,
	})
}

// HandleRefreshSession handles requests to refresh the current session
func (h *SessionHandler) HandleRefreshSession(w http.ResponseWriter, r *http.Request) {
	sessionData := middleware.GetSessionData(r.Context())
	if sessionData == nil {
		http.Error(w, "No active session", http.StatusBadRequest)
		return
	}

	// Update last accessed time
	now := time.Now()
	sessionData.LastAccessed = &now

	// Re-encrypt and set updated session cookie
	if err := h.cookieSessionService.SetSessionCookie(w, sessionData); err != nil {
		h.logger.Error().Err(err).Msg("Failed to refresh session cookie")
		http.Error(w, "Failed to refresh session", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]bool{
		"success": true,
	})
}
