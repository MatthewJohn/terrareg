package service

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/auth"
)

// AuthenticationService orchestrates session and cookie operations for authentication flows
type AuthenticationService struct {
	sessionService *SessionService
	cookieService  *CookieService
}

// NewAuthenticationService creates a new authentication service
func NewAuthenticationService(sessionService *SessionService, cookieService *CookieService) *AuthenticationService {
	return &AuthenticationService{
		sessionService: sessionService,
		cookieService:  cookieService,
	}
}

// AuthenticationContext contains authentication information for a request
type AuthenticationContext struct {
	SessionData     *SessionData
	Session         *auth.Session
	IsAuthenticated bool
	AuthMethod      string
	IsAdmin         bool
	Username        string
	Permissions     map[string]string
}

// CreateAuthenticatedSession creates a new session and sets the authentication cookie
func (as *AuthenticationService) CreateAuthenticatedSession(
	ctx context.Context,
	w http.ResponseWriter,
	authMethod string,
	providerData map[string]interface{},
	ttl *time.Duration,
) error {
	// Convert provider data to JSON bytes
	providerDataBytes, err := json.Marshal(providerData)
	if err != nil {
		return fmt.Errorf("failed to marshal provider data: %w", err)
	}

	// Create session in database
	session, err := as.sessionService.CreateSession(ctx, authMethod, providerDataBytes, ttl)
	if err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}

	// Create session data for cookie
	sessionData := &SessionData{
		SessionID:   session.ID,
		AuthMethod:  authMethod,
		IsAdmin:     false, // Default to false, can be updated based on auth method
		Permissions: make(map[string]string),
	}

	// Update session data based on auth method
	switch authMethod {
	case "ADMIN_API_KEY":
		sessionData.Username = "admin"
		sessionData.IsAdmin = true
		// Extract permissions from provider data if present
		if permissions, ok := providerData["permissions"].(map[string]string); ok {
			sessionData.Permissions = permissions
		}
	case "OIDC", "SAML", "OAUTH":
		// Extract username from provider data
		if username, ok := providerData["username"].(string); ok {
			sessionData.Username = username
		}
		if isAdmin, ok := providerData["is_admin"].(bool); ok {
			sessionData.IsAdmin = isAdmin
		}
		if permissions, ok := providerData["permissions"].(map[string]string); ok {
			sessionData.Permissions = permissions
		}
	}

	// Encrypt session data and set cookie
	encryptedSession, err := as.cookieService.EncryptSession(sessionData)
	if err != nil {
		return fmt.Errorf("failed to encrypt session data: %w", err)
	}

	// Set session cookie
	as.cookieService.SetCookie(w, as.cookieService.GetSessionCookieName(), encryptedSession, &CookieOptions{
		Path:     "/",
		MaxAge:   int(ttl.Seconds()), // Convert to seconds
		Secure:   true,               // Always use secure cookies for authenticated sessions
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})

	return nil
}

// ValidateRequest validates authentication from HTTP request
func (as *AuthenticationService) ValidateRequest(ctx context.Context, r *http.Request) (*AuthenticationContext, error) {
	// Extract session cookie
	cookie, err := r.Cookie(as.cookieService.GetSessionCookieName())
	if err != nil {
		// No cookie found - return unauthenticated context
		return &AuthenticationContext{
			IsAuthenticated: false,
		}, nil
	}

	// Validate and decrypt session cookie
	sessionData, err := as.cookieService.ValidateSessionCookie(cookie.Value)
	if err != nil {
		// Invalid or expired cookie - return unauthenticated context
		return &AuthenticationContext{
			IsAuthenticated: false,
		}, nil
	}

	// Validate session in database
	session, err := as.sessionService.ValidateSession(ctx, sessionData.SessionID)
	if err != nil {
		// Session not found or expired - return unauthenticated context
		return &AuthenticationContext{
			IsAuthenticated: false,
		}, nil
	}

	// Build authentication context
	authCtx := &AuthenticationContext{
		SessionData:     sessionData,
		Session:         session,
		IsAuthenticated: true,
		AuthMethod:      sessionData.AuthMethod,
		IsAdmin:         sessionData.IsAdmin,
		Username:        sessionData.Username,
		Permissions:     sessionData.Permissions,
	}

	return authCtx, nil
}

// InvalidateSession removes a session (both database record and cookie)
func (as *AuthenticationService) InvalidateSession(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	// Extract and validate session cookie
	cookie, err := r.Cookie(as.cookieService.GetSessionCookieName())
	if err == nil {
		sessionData, err := as.cookieService.ValidateSessionCookie(cookie.Value)
		if err == nil {
			// Delete session from database
			if sessionData.SessionID != "" {
				_ = as.sessionService.DeleteSession(ctx, sessionData.SessionID)
			}
		}
	}

	// Clear the cookie
	as.cookieService.ClearCookie(w, as.cookieService.GetSessionCookieName())

	return nil
}

// RefreshSession extends a session's expiry time
func (as *AuthenticationService) RefreshSession(
	ctx context.Context,
	w http.ResponseWriter,
	r *http.Request,
	ttl time.Duration,
) error {
	// Extract and validate session cookie
	cookie, err := r.Cookie(as.cookieService.GetSessionCookieName())
	if err != nil {
		return fmt.Errorf("no session cookie found")
	}

	sessionData, err := as.cookieService.ValidateSessionCookie(cookie.Value)
	if err != nil {
		return fmt.Errorf("invalid session cookie: %w", err)
	}

	// Refresh session in database
	session, err := as.sessionService.RefreshSession(ctx, sessionData.SessionID, ttl)
	if err != nil {
		return fmt.Errorf("failed to refresh session: %w", err)
	}

	// Update session data expiry to match the refreshed session
	sessionData.Expiry = &session.Expiry

	// Re-encrypt and update cookie
	encryptedSession, err := as.cookieService.EncryptSession(sessionData)
	if err != nil {
		return fmt.Errorf("failed to encrypt updated session data: %w", err)
	}

	// Set updated cookie
	as.cookieService.SetCookie(w, as.cookieService.GetSessionCookieName(), encryptedSession, &CookieOptions{
		Path:     "/",
		MaxAge:   int(ttl.Seconds()),
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})

	return nil
}

// CreateAdminSession creates a complete admin session (convenience method)
func (as *AuthenticationService) CreateAdminSession(ctx context.Context, w http.ResponseWriter, sessionID string) error {
	log.Info().
		Str("session_id", sessionID).
		Msg("CreateAdminSession: starting")

	// Get session from database to ensure it's valid
	session, err := as.sessionService.GetSession(sessionID)
	if err != nil {
		log.Error().
			Err(err).
			Str("session_id", sessionID).
			Msg("CreateAdminSession: failed to get session from database")
		return fmt.Errorf("failed to get session: %w", err)
	}

	log.Info().
		Str("session_id", sessionID).
		Time("expiry", session.Expiry).
		Msg("CreateAdminSession: retrieved session successfully")

	// Create admin session data
	sessionData := &SessionData{
		SessionID:   sessionID,
		Username:    "admin",
		AuthMethod:  "ADMIN_API_KEY",
		IsAdmin:     true,
		Permissions: make(map[string]string),
		Expiry:      &session.Expiry,
	}

	// Use the session variable to avoid unused variable warning
	_ = session

	// Encrypt and set cookie
	log.Info().
		Str("session_id", sessionID).
		Msg("CreateAdminSession: encrypting session data")

	encryptedSession, err := as.cookieService.EncryptSession(sessionData)
	if err != nil {
		log.Error().
			Err(err).
			Str("session_id", sessionID).
			Msg("CreateAdminSession: failed to encrypt session")
		return fmt.Errorf("failed to encrypt admin session: %w", err)
	}

	// Calculate TTL for cookie
	ttl := time.Until(session.Expiry)
	log.Info().
		Str("session_id", sessionID).
		Dur("ttl", ttl).
		Msg("CreateAdminSession: setting cookie")

	if ttl <= 0 {
		ttl = 24 * time.Hour // Default to 24 hours
		log.Warn().
			Str("session_id", sessionID).
			Dur("default_ttl", ttl).
			Msg("CreateAdminSession: session expiry in past, using default TTL")
	}

	as.cookieService.SetCookie(w, as.cookieService.GetSessionCookieName(), encryptedSession, &CookieOptions{
		Path:     "/",
		MaxAge:   int(ttl.Seconds()),
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})

	log.Info().
		Str("session_id", sessionID).
		Msg("CreateAdminSession: completed successfully")

	return nil
}
