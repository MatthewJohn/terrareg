package service

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/auth"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/auth/repository"
	infraConfig "github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/config"
)

// CookieSessionService combines session management with cookie operations
// This bridges the pure SessionService with HTTP cookie handling
type CookieSessionService struct {
	sessionService *SessionService
	cookieService  *CookieService
	sessionRepo    repository.SessionRepository
	config         *infraConfig.InfrastructureConfig
}

// NewCookieSessionService creates a new cookie session service
func NewCookieSessionService(
	sessionService *SessionService,
	cookieService *CookieService,
	sessionRepo repository.SessionRepository,
	config *infraConfig.InfrastructureConfig,
) *CookieSessionService {
	return &CookieSessionService{
		sessionService: sessionService,
		cookieService:  cookieService,
		sessionRepo:    sessionRepo,
		config:         config,
	}
}

// CreateSessionRequest represents a session creation request
type CreateSessionRequest struct {
	AuthMethod    string          `json:"auth_method"`
	Username      string          `json:"username,omitempty"`
	IsAdmin       bool            `json:"is_admin,omitempty"`
	SiteAdmin     bool            `json:"site_admin,omitempty"`
	UserGroups    []string        `json:"user_groups,omitempty"`
	ProviderData  json.RawMessage `json:"provider_data,omitempty"`
	TTL           *time.Duration  `json:"ttl,omitempty"`
}

// CreateSessionResponse represents a session creation response
type CreateSessionResponse struct {
	SessionID     string    `json:"session_id"`
	Expiry        time.Time `json:"expiry"`
	Authenticated bool      `json:"authenticated"`
	Username      string    `json:"username,omitempty"`
	IsAdmin       bool      `json:"is_admin,omitempty"`
	SiteAdmin     bool      `json:"site_admin,omitempty"`
	UserGroups    []string  `json:"user_groups,omitempty"`
}

// CreateSession creates a new session and returns the response
func (cs *CookieSessionService) CreateSession(ctx context.Context, req *CreateSessionRequest) (*CreateSessionResponse, error) {
	// Convert provider data to bytes
	var providerData []byte
	if req.ProviderData != nil {
		providerData = req.ProviderData
	}

	// Create session using session service
	session, err := cs.sessionService.CreateSession(ctx, req.AuthMethod, providerData, req.TTL)
	if err != nil {
		return nil, err
	}

	// Create response
	response := &CreateSessionResponse{
		SessionID:     session.ID,
		Expiry:        session.Expiry,
		Authenticated: true,
		Username:      req.Username,
		IsAdmin:       req.IsAdmin,
		SiteAdmin:     req.SiteAdmin,
		UserGroups:    req.UserGroups,
	}

	return response, nil
}

// ValidateSession validates a session ID and returns the session data
func (cs *CookieSessionService) ValidateSession(ctx context.Context, sessionID string) (*auth.Session, error) {
	return cs.sessionRepo.FindByID(ctx, sessionID)
}

// DeleteSession deletes a session
func (cs *CookieSessionService) DeleteSession(ctx context.Context, sessionID string) error {
	return cs.sessionRepo.Delete(ctx, sessionID)
}

// SetSessionCookie sets a session cookie in the HTTP response
func (cs *CookieSessionService) SetSessionCookie(w http.ResponseWriter, sessionData *SessionData) error {
	// Use the cookie service to encrypt and set the cookie
	return cs.cookieService.SetSessionCookie(w, sessionData)
}

// ClearSessionCookie clears the session cookie
func (cs *CookieSessionService) ClearSessionCookie(w http.ResponseWriter) error {
	// Use the cookie service to clear the cookie
	return cs.cookieService.ClearSessionCookie(w)
}