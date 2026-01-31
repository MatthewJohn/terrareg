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
	sessionManagementService *service.SessionManagementService
	logger                   zerolog.Logger
}

// NewSessionHandler creates a new SessionHandler
func NewSessionHandler(
	sessionManagementService *service.SessionManagementService,
	logger zerolog.Logger,
) *SessionHandler {
	return &SessionHandler{
		sessionManagementService: sessionManagementService,
		logger:                   logger,
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
	if sessionData.IsAuthenticated {
		response = GetSessionResponse{
			Authenticated: true,
			UserID:        sessionData.UserID,
			Username:      sessionData.Username,
			AuthMethod:    string(sessionData.AuthMethod),
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
	if !sessionData.IsAuthenticated {
		http.Error(w, "No active session", http.StatusBadRequest)
		return
	}

	// Clear session and cookie in one operation
	if err := h.sessionManagementService.ClearSessionAndCookie(r.Context(), w, r); err != nil {
		h.logger.Error().Err(err).Msg("Failed to clear session and cookie")
		http.Error(w, "Failed to delete session", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]bool{
		"success": true,
	})
}

// HandleRefreshSession handles requests to refresh the current session
func (h *SessionHandler) HandleRefreshSession(w http.ResponseWriter, r *http.Request) {
	sessionData := middleware.GetSessionData(r.Context())
	if !sessionData.IsAuthenticated {
		http.Error(w, "No active session", http.StatusBadRequest)
		return
	}

	// Default TTL of 24 hours if session has no expiry
	ttl := 24 * time.Hour
	if sessionData.Expiry != nil {
		ttl = time.Until(*sessionData.Expiry)
		if ttl <= 0 {
			ttl = 24 * time.Hour
		}
	}

	// Refresh session and update cookie in one operation
	if err := h.sessionManagementService.RefreshSessionAndCookie(r.Context(), w, r, ttl); err != nil {
		h.logger.Error().Err(err).Msg("Failed to refresh session and cookie")
		http.Error(w, "Failed to refresh session", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]bool{
		"success": true,
	})
}
