package service

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	auditservice "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/audit/service"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/auth"
	"github.com/rs/zerolog/log"
)

// AuthenticationService orchestrates session and cookie operations for authentication flows
type AuthenticationService struct {
	// sessionService manages session persistence (required)
	sessionService *SessionService
	// cookieService handles cookie encryption and validation (required)
	cookieService *CookieService
	// authAuditService handles audit logging for login events (required)
	authAuditService auditservice.AuthenticationAuditServiceInterface
}

// NewAuthenticationService creates a new authentication service
// Returns nil if cookieService is nil (SECRET_KEY not configured - cookie-based auth unavailable)
func NewAuthenticationService(sessionService *SessionService, cookieService *CookieService, authAuditService auditservice.AuthenticationAuditServiceInterface) (*AuthenticationService, error) {
	if sessionService == nil {
		return nil, fmt.Errorf("sessionService cannot be nil")
	}
	if cookieService == nil {
		// SECRET_KEY not configured - cookie-based auth is not available
		return nil, nil
	}
	if authAuditService == nil {
		return nil, fmt.Errorf("authAuditService cannot be nil")
	}

	return &AuthenticationService{
		sessionService:   sessionService,
		cookieService:    cookieService,
		authAuditService: authAuditService,
	}, nil
}

// SessionData contains session information for HTTP handlers
type SessionData struct {
	SessionID    string            `json:"session_id"`
	UserID       string            `json:"user_id,omitempty"`
	Username     string            `json:"username,omitempty"`
	AuthMethod   string            `json:"auth_method"`
	IsAdmin      bool              `json:"is_admin"`
	SiteAdmin    bool              `json:"site_admin"`
	UserGroups   []string          `json:"user_groups,omitempty"`
	LastAccessed *time.Time        `json:"last_accessed,omitempty"`
	Expiry       *time.Time        `json:"expiry,omitempty"`
	Permissions  map[string]string `json:"permissions,omitempty"`
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

// CreateSessionFromAuthContext creates a new session and sets the authentication cookie
// using an AuthContext for type-safe access to authentication properties
func (as *AuthenticationService) CreateSessionFromAuthContext(
	ctx context.Context,
	w http.ResponseWriter,
	authCtx auth.AuthContext,
	ttl *time.Duration,
) error {
	// Convert AuthContext to providerData for persistence
	providerData := authCtx.GetProviderData()

	// Marshal to JSON for database storage
	providerDataBytes, err := json.Marshal(providerData)
	if err != nil {
		return fmt.Errorf("failed to marshal provider data: %w", err)
	}

	// Create session in database
	session, err := as.sessionService.CreateSession(ctx, string(authCtx.GetProviderType()), providerDataBytes, ttl)
	if err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}

	// Create session data for cookie - use AuthContext methods instead of type assertions
	sessionData := &SessionData{
		SessionID:   session.ID,
		AuthMethod:  string(authCtx.GetProviderType()),
		IsAdmin:     authCtx.IsAdmin(),
		SiteAdmin:   false, // TODO: Add IsSiteAdmin() to AuthContext interface if needed
		UserGroups:  authCtx.GetUserGroupNames(),
		Expiry:      &session.Expiry,
		Permissions: authCtx.GetAllNamespacePermissions(),
	}

	if username := authCtx.GetUsername(); username != "" {
		sessionData.Username = username
	}

	// Encrypt session data and set cookie
	encryptedSession, err := as.cookieService.EncryptSession(sessionData)
	if err != nil {
		return fmt.Errorf("failed to encrypt session data: %w", err)
	}

	// Calculate maxAge for cookie
	maxAge := 24 * 60 * 60 // Default 24 hours
	if ttl != nil {
		maxAge = int(ttl.Seconds())
	}

	// Set session cookie
	as.cookieService.SetCookie(w, as.cookieService.GetSessionCookieName(), encryptedSession, &CookieOptions{
		Path:     "/",
		MaxAge:   maxAge,
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})

	// Log audit event for successful login
	// Python reference: /app/terrareg/server/api/github/github_login_callback.py:65
	as.authAuditService.LogUserLogin(ctx, authCtx.GetUsername(), sessionData.AuthMethod)

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
	if err != nil || sessionData == nil {
		// Invalid or expired cookie - return unauthenticated context
		return &AuthenticationContext{
			IsAuthenticated: false,
		}, nil
	}

	// Validate session in database
	session, err := as.sessionService.ValidateSession(ctx, sessionData.SessionID)
	if err != nil || session == nil {
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

// CreateSession creates a session cookie from an existing session ID
func (as *AuthenticationService) CreateSession(ctx context.Context, w http.ResponseWriter, sessionID string) error {
	log.Info().
		Str("session_id", sessionID).
		Msg("CreateSession: starting")

	// Get session from database to ensure it's valid
	session, err := as.sessionService.GetSession(ctx, sessionID)
	if err != nil {
		log.Error().
			Err(err).
			Str("session_id", sessionID).
			Msg("CreateSession: failed to get session from database")
		return fmt.Errorf("failed to get session: %w", err)
	}

	log.Info().
		Str("session_id", sessionID).
		Time("expiry", session.Expiry).
		Msg("CreateSession: retrieved session successfully")

	// Create basic session data
	sessionData := &SessionData{
		SessionID:   sessionID,
		AuthMethod:  "",    // We'll extract this from provider data
		IsAdmin:     false, // Default to false, can be updated based on session
		SiteAdmin:   false,
		UserGroups:  []string{},
		Permissions: make(map[string]string),
		Expiry:      &session.Expiry,
	}

	// Parse provider data from session to extract auth method and user details if available
	if len(session.ProviderSourceAuth) > 0 {
		var providerData map[string]interface{}
		if err := json.Unmarshal(session.ProviderSourceAuth, &providerData); err == nil {
			// Extract auth method from provider data if present
			if authMethod, ok := providerData["auth_method"].(string); ok {
				sessionData.AuthMethod = authMethod
			}
			// Extract username from provider data if present
			if username, ok := providerData["username"].(string); ok {
				sessionData.Username = username
			}
			// Extract admin status from provider data if present
			if isAdmin, ok := providerData["is_admin"].(bool); ok {
				sessionData.IsAdmin = isAdmin
			}
			// Extract permissions from provider data if present
			if permissions, ok := providerData["permissions"].(map[string]string); ok {
				sessionData.Permissions = permissions
			}
		}
	}

	// Encrypt and set cookie
	log.Info().
		Str("session_id", sessionID).
		Msg("CreateSession: encrypting session data")

	encryptedSession, err := as.cookieService.EncryptSession(sessionData)
	if err != nil {
		log.Error().
			Err(err).
			Str("session_id", sessionID).
			Msg("CreateSession: failed to encrypt session")
		return fmt.Errorf("failed to encrypt session: %w", err)
	}

	// Calculate TTL for cookie
	ttl := time.Until(session.Expiry)
	log.Info().
		Str("session_id", sessionID).
		Dur("ttl", ttl).
		Msg("CreateSession: setting cookie")

	if ttl <= 0 {
		ttl = 24 * time.Hour // Default to 24 hours
		log.Warn().
			Str("session_id", sessionID).
			Dur("default_ttl", ttl).
			Msg("CreateSession: session expiry in past, using default TTL")
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
		Msg("CreateSession: completed successfully")

	// Log audit event for successful login
	// Python reference: /app/terrareg/server/api/open_id_callback.py:86
	as.authAuditService.LogUserLogin(ctx, sessionData.Username, sessionData.AuthMethod)

	return nil
}

// ClearSession removes the session cookie and invalidates the session
func (as *AuthenticationService) ClearSession(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	log.Info().
		Msg("ClearSession: starting")

	// Extract and validate session cookie
	cookie, err := r.Cookie(as.cookieService.GetSessionCookieName())
	if err == nil {
		sessionData, err := as.cookieService.ValidateSessionCookie(cookie.Value)
		if err == nil {
			// Delete session from database
			if sessionData.SessionID != "" {
				log.Info().
					Str("session_id", sessionData.SessionID).
					Msg("ClearSession: invalidating session in database")

				_ = as.sessionService.DeleteSession(ctx, sessionData.SessionID)
			}
		}
	}

	// Clear the cookie
	log.Info().
		Msg("ClearSession: clearing session cookie")

	as.cookieService.ClearCookie(w, as.cookieService.GetSessionCookieName())

	log.Info().
		Msg("ClearSession: completed successfully")

	return nil
}

// CreateAdminSession creates a complete admin session (convenience method)
func (as *AuthenticationService) CreateAdminSession(ctx context.Context, w http.ResponseWriter, sessionID string) error {
	log.Info().
		Str("session_id", sessionID).
		Msg("CreateAdminSession: starting")

	// Get session from database to ensure it's valid
	session, err := as.sessionService.GetSession(ctx, sessionID)
	if err != nil {
		log.Error().
			Err(err).
			Str("session_id", sessionID).
			Msg("CreateAdminSession: failed to get session from database")
		return fmt.Errorf("failed to get session: %w", err)
	}

	if session == nil {
		log.Error().
			Str("session_id", sessionID).
			Msg("CreateAdminSession: session not found in database")
		return fmt.Errorf("session not found: %s", sessionID)
	}

	log.Info().
		Str("session_id", sessionID).
		Time("expiry", session.Expiry).
		Msg("CreateAdminSession: retrieved session successfully")

	// Create admin session data
	sessionData := &SessionData{
		SessionID:   sessionID,
		Username:    "Built-in admin",
		AuthMethod:  "ADMIN_API_KEY",
		IsAdmin:     true,
		SiteAdmin:   false,
		UserGroups:  []string{},
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

	// Log audit event for successful admin login
	// Python reference: /app/terrareg/server/api/terrareg_admin_authenticate.py:28
	as.authAuditService.LogUserLogin(ctx, sessionData.Username, sessionData.AuthMethod)

	return nil
}
