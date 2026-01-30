package service

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/auth"
)

// Ensure SessionManagementService implements auth.SessionManager interface
var _ auth.SessionManager = (*SessionManagementService)(nil)

// SessionManagementService handles all session and cookie operations as single units of work
// Returns nil if cookieService is nil (SECRET_KEY not configured - login methods disabled)
type SessionManagementService struct {
	sessionService *SessionService
	cookieService  *CookieService
}

// NewSessionManagementService creates a new session management service
// Returns nil if cookieService is nil (no login methods available)
func NewSessionManagementService(
	sessionService *SessionService,
	cookieService *CookieService,
) *SessionManagementService {
	if cookieService == nil {
		// SECRET_KEY not configured - login methods not available
		return nil
	}
	return &SessionManagementService{
		sessionService: sessionService,
		cookieService:  cookieService,
	}
}

// CreateSessionAndCookie creates a session in the database and sets the encrypted cookie
// This is a single unit of work - both operations succeed or both fail
func (s *SessionManagementService) CreateSessionAndCookie(
	ctx context.Context,
	w http.ResponseWriter,
	authMethod auth.AuthMethodType,
	username string,
	isAdmin bool,
	userGroups []string,
	permissions map[string]string,
	providerData map[string]interface{},
	ttl *time.Duration,
) error {
	// Convert provider data to bytes
	var providerDataBytes []byte
	if providerData != nil {
		var err error
		providerDataBytes, err = json.Marshal(providerData)
		if err != nil {
			return fmt.Errorf("failed to marshal provider data: %w", err)
		}
	}

	// Create session using session service
	session, err := s.sessionService.CreateSession(ctx, string(authMethod), providerDataBytes, ttl)
	if err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}

	// Create session data for cookie
	sessionData := &auth.SessionData{
		SessionID:   session.ID,
	AuthMethod:  string(authMethod),
		Username:    username,
		IsAdmin:     isAdmin,
		SiteAdmin:   false,
		UserGroups:  userGroups,
		Permissions: permissions,
		Expiry:      &session.Expiry,
	}

	// Encrypt session data and set cookie
	encryptedSession, err := s.cookieService.EncryptSession(sessionData)
	if err != nil {
		return fmt.Errorf("failed to encrypt session data: %w", err)
	}

	// Calculate maxAge for cookie
	maxAge := 24 * 60 * 60 // Default 24 hours
	if ttl != nil {
		maxAge = int(ttl.Seconds())
	}

	// Set session cookie
	s.cookieService.SetCookie(w, s.cookieService.GetSessionCookieName(), encryptedSession, &CookieOptions{
		Path:     "/",
		MaxAge:   maxAge,
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})

	return nil
}

// ValidateSessionCookie validates a session cookie by decrypting and checking the database
// Returns the session if valid, nil if invalid/expired
func (s *SessionManagementService) ValidateSessionCookie(
	ctx context.Context,
	cookieValue string,
) (*auth.Session, error) {
	// Decrypt the cookie value
	sessionData, err := s.cookieService.DecryptSession(cookieValue)
	if err != nil {
		return nil, fmt.Errorf("invalid session cookie: %w", err)
	}

	// Validate session in database
	session, err := s.sessionService.ValidateSession(ctx, sessionData.SessionID)
	if err != nil {
		return nil, fmt.Errorf("session validation failed: %w", err)
	}

	return session, nil
}

// ClearSessionAndCookie removes a session from the database and clears the cookie
// This is a single unit of work - both operations are performed
func (s *SessionManagementService) ClearSessionAndCookie(
	ctx context.Context,
	w http.ResponseWriter,
	r *http.Request,
) error {
	// Extract and validate session cookie
	cookie, err := r.Cookie(s.cookieService.GetSessionCookieName())
	if err == nil {
		sessionData, err := s.cookieService.ValidateSessionCookie(cookie.Value)
		if err == nil {
			// Delete session from database
			if sessionData.SessionID != "" {
				_ = s.sessionService.DeleteSession(ctx, sessionData.SessionID)
			}
		}
	}

	// Clear the cookie
	s.cookieService.ClearCookie(w, s.cookieService.GetSessionCookieName())

	return nil
}

// RefreshSessionAndCookie refreshes a session expiry and updates the cookie
// This is a single unit of work - both operations succeed or both fail
func (s *SessionManagementService) RefreshSessionAndCookie(
	ctx context.Context,
	w http.ResponseWriter,
	r *http.Request,
	ttl time.Duration,
) error {
	// Extract and validate session cookie
	cookie, err := r.Cookie(s.cookieService.GetSessionCookieName())
	if err != nil {
		return fmt.Errorf("no session cookie found")
	}

	sessionData, err := s.cookieService.ValidateSessionCookie(cookie.Value)
	if err != nil {
		return fmt.Errorf("invalid session cookie: %w", err)
	}

	// Refresh session in database
	session, err := s.sessionService.RefreshSession(ctx, sessionData.SessionID, ttl)
	if err != nil {
		return fmt.Errorf("failed to refresh session: %w", err)
	}

	// Update session data expiry to match the refreshed session
	sessionData.Expiry = &session.Expiry

	// Re-encrypt and update cookie
	encryptedSession, err := s.cookieService.EncryptSession(sessionData)
	if err != nil {
		return fmt.Errorf("failed to encrypt updated session data: %w", err)
	}

	// Set updated cookie
	s.cookieService.SetCookie(w, s.cookieService.GetSessionCookieName(), encryptedSession, &CookieOptions{
		Path:     "/",
		MaxAge:   int(ttl.Seconds()),
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})

	return nil
}

// IsAvailable returns true if session management is available (SECRET_KEY configured)
func (s *SessionManagementService) IsAvailable() bool {
	return s.cookieService != nil
}

// GetSessionFromCookie extracts and validates session from HTTP request cookie
// Convenience method that combines cookie extraction + validation
func (s *SessionManagementService) GetSessionFromCookie(
	ctx context.Context,
	r *http.Request,
) (*auth.Session, *auth.SessionData, error) {
	// Extract session cookie from request
	cookie, err := r.Cookie(s.cookieService.GetSessionCookieName())
	if err != nil {
		return nil, nil, fmt.Errorf("no session cookie found")
	}

	// Validate session cookie
	session, err := s.ValidateSessionCookie(ctx, cookie.Value)
	if err != nil {
		return nil, nil, err
	}

	// Decrypt session data to get full details
	sessionData, err := s.cookieService.DecryptSession(cookie.Value)
	if err != nil {
		return session, nil, fmt.Errorf("failed to decrypt session data: %w", err)
	}

	return session, sessionData, nil
}

// SetCookieForExistingSession sets a session cookie for an already-existing session
// This is useful when a session was created elsewhere and we just need to set the cookie
func (s *SessionManagementService) SetCookieForExistingSession(
	ctx context.Context,
	w http.ResponseWriter,
	sessionID string,
) error {
	// Get the session from the database
	session, err := s.sessionService.ValidateSession(ctx, sessionID)
	if err != nil {
		return fmt.Errorf("failed to get session: %w", err)
	}

	// Create session data for cookie
	// Note: We don't have full user info here, so we use minimal data
	// The cookie contains the session ID which is the most important part
	sessionData := &auth.SessionData{
		SessionID:  session.ID,
		AuthMethod: "", // Will be filled from provider data if available
		Username:    "",
		IsAdmin:     false,
		SiteAdmin:   false,
		UserGroups:  []string{},
		Permissions: nil,
		Expiry:      &session.Expiry,
	}

	// Encrypt session data and set cookie
	encryptedSession, err := s.cookieService.EncryptSession(sessionData)
	if err != nil {
		return fmt.Errorf("failed to encrypt session data: %w", err)
	}

	// Calculate maxAge for cookie
	ttl := time.Until(session.Expiry)
	if ttl <= 0 {
		ttl = 24 * time.Hour
	}

	// Set session cookie
	s.cookieService.SetCookie(w, s.cookieService.GetSessionCookieName(), encryptedSession, &CookieOptions{
		Path:     "/",
		MaxAge:   int(ttl.Seconds()),
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})

	return nil
}
