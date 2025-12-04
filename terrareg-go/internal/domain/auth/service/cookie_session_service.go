package service

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/application/security"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/config"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/auth"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/auth/repository"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/security/csrf"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/url/service"
)

// SessionData represents the data stored in the encrypted session cookie
type SessionData struct {
	SessionID          string            `json:"session_id"`
	UserID             string            `json:"user_id,omitempty"`
	Username           string            `json:"username,omitempty"`
	AuthMethod         string            `json:"auth_method"`
	IsAdmin            bool              `json:"is_admin,omitempty"`
	Permissions        map[string]string `json:"permissions,omitempty"`
	CSRFToken          string            `json:"csrf_token"`
	CreatedAt          time.Time         `json:"created_at"`
	LastAccessed       time.Time         `json:"last_accessed"`
	UserGroups         []string          `json:"user_groups,omitempty"`
	ProviderSourceAuth []byte            `json:"provider_source_auth,omitempty"`
	Theme              string            `json:"theme,omitempty"`
}

// CookieSessionService manages encrypted client-side cookie sessions
type CookieSessionService struct {
	sessionRepo repository.SessionRepository
	csrfService *security.CSRFService
	config      *config.Config
	cipher      *SessionCipher
	urlService  *service.URLService
}

// NewCookieSessionService creates a new cookie session service
func NewCookieSessionService(
	sessionRepo repository.SessionRepository,
	csrfService *security.CSRFService,
	config *config.Config,
	urlService *service.URLService,
) (*CookieSessionService, error) {
	if config.SecretKey == "" {
		return nil, fmt.Errorf("secret key is required for session encryption")
	}

	cipher, err := NewSessionCipher(config.SecretKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create session cipher: %w", err)
	}

	return &CookieSessionService{
		sessionRepo: sessionRepo,
		csrfService: csrfService,
		config:      config,
		cipher:      cipher,
		urlService:  urlService,
	}, nil
}

// CreateSession creates a new session and returns the encrypted cookie data
func (s *CookieSessionService) CreateSession(ctx context.Context, userID, username, authMethod string, isAdmin bool) (*SessionData, error) {
	// Generate secure session ID
	sessionID, err := s.generateSessionID()
	if err != nil {
		return nil, fmt.Errorf("failed to generate session ID: %w", err)
	}

	// Generate CSRF token
	csrfToken, err := csrf.NewCSRFToken()
	if err != nil {
		return nil, fmt.Errorf("failed to generate CSRF token: %w", err)
	}

	// Create database session record
	expiry := time.Now().Add(s.config.SessionExpiry)
	dbSession := &auth.Session{
		ID:     sessionID,
		Expiry: expiry,
	}

	if err := s.sessionRepo.Create(ctx, dbSession); err != nil {
		return nil, fmt.Errorf("failed to create session in database: %w", err)
	}

	// Create session data
	now := time.Now()
	sessionData := &SessionData{
		SessionID:    sessionID,
		UserID:       userID,
		Username:     username,
		AuthMethod:   authMethod,
		IsAdmin:      isAdmin,
		CSRFToken:    csrfToken.String(),
		CreatedAt:    now,
		LastAccessed: now,
	}

	return sessionData, nil
}

// ValidateSession validates an encrypted session cookie against the database
func (s *CookieSessionService) ValidateSession(ctx context.Context, cookieValue string) (*SessionData, error) {
	if cookieValue == "" {
		return nil, fmt.Errorf("empty session cookie")
	}

	// Decrypt session data
	decryptedData, err := s.cipher.Decrypt(cookieValue)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt session cookie: %w", err)
	}

	// Parse session data
	var sessionData SessionData
	if err := json.Unmarshal(decryptedData, &sessionData); err != nil {
		return nil, fmt.Errorf("failed to parse session data: %w", err)
	}

	// Validate session exists in database
	dbSession, err := s.sessionRepo.FindByID(ctx, sessionData.SessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to validate session in database: %w", err)
	}

	// Check if session has expired
	if dbSession.IsExpired() {
		// Clean up expired session
		s.sessionRepo.Delete(ctx, sessionData.SessionID)
		return nil, fmt.Errorf("session has expired")
	}

	// Update last accessed time
	sessionData.LastAccessed = time.Now()

	// Update provider source auth data if it exists in database
	if len(dbSession.ProviderSourceAuth) > 0 {
		sessionData.ProviderSourceAuth = dbSession.ProviderSourceAuth
	}

	return &sessionData, nil
}

// EncryptSession encrypts session data for storage in a cookie
func (s *CookieSessionService) EncryptSession(sessionData *SessionData) (string, error) {
	jsonData, err := json.Marshal(sessionData)
	if err != nil {
		return "", fmt.Errorf("failed to marshal session data: %w", err)
	}

	return s.cipher.Encrypt(jsonData)
}

// DeleteSession deletes a session from the database
func (s *CookieSessionService) DeleteSession(ctx context.Context, sessionID string) error {
	return s.sessionRepo.Delete(ctx, sessionID)
}

// CleanupExpiredSessions removes expired sessions from the database
func (s *CookieSessionService) CleanupExpiredSessions(ctx context.Context) error {
	return s.sessionRepo.CleanupExpired(ctx)
}

// generateSessionID generates a secure random session ID
func (s *CookieSessionService) generateSessionID() (string, error) {
	bytes := make([]byte, 32) // 256 bits
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}

	return base64.URLEncoding.EncodeToString(bytes), nil
}

// SetSessionCookie sets the encrypted session cookie in the HTTP response
func (s *CookieSessionService) SetSessionCookie(w http.ResponseWriter, sessionData *SessionData) error {
	encryptedData, err := s.EncryptSession(sessionData)
	if err != nil {
		return fmt.Errorf("failed to encrypt session data: %w", err)
	}

	s.setCookie(encryptedData, w)
	return nil
}

// ClearSessionCookie clears the session cookie
func (s *CookieSessionService) ClearSessionCookie(w http.ResponseWriter) {
	s.setCookie("", w)
}

func (s *CookieSessionService) setCookie(data string, w http.ResponseWriter) {
	cookie := &http.Cookie{
		Name:     s.getSessionCookieName(),
		Value:    data,
		Path:     "/",
		MaxAge:   -1,
		Secure:   s.getSessionSecure(),
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	}

	http.SetCookie(w, cookie)
}

// GetSessionCookieName returns the session cookie name
func (s *CookieSessionService) GetSessionCookieName() string {
	if s.config.SessionCookieName != "" {
		return s.config.SessionCookieName
	}
	return "terrareg_session"
}

// getSessionCookieName returns the session cookie name (private method)
func (s *CookieSessionService) getSessionCookieName() string {
	return s.GetSessionCookieName()
}

// getSessionSecure returns whether session cookies should be secure
func (s *CookieSessionService) getSessionSecure() bool {
	// Default to true unless explicitly configured otherwise
	// In a real implementation, this might be based on environment or config
	return true
}

// SetBasicSessionCookie sets a basic session_id cookie for compatibility
// This method centralizes session_id cookie management used across auth methods
// Uses centralized HTTPS detection via URL service following DDD principles
func (s *CookieSessionService) SetBasicSessionCookie(w http.ResponseWriter, sessionID string, expiry time.Time) {
	// Use URL service for HTTPS detection - this centralizes URL logic
	isHTTPS := s.urlService.IsHTTPS(nil)

	cookie := &http.Cookie{
		Name:     "session_id",
		Value:    sessionID,
		Path:     "/",
		HttpOnly: true,
		Secure:   isHTTPS,
		SameSite: http.SameSiteLaxMode,
		Expires:  expiry,
	}

	http.SetCookie(w, cookie)
}

// SetAdminAuthenticationCookie sets the admin authentication flag cookie
// Uses centralized HTTPS detection via URL service following DDD principles
func (s *CookieSessionService) SetAdminAuthenticationCookie(w http.ResponseWriter, isAuthenticated bool) {
	// Use URL service for HTTPS detection - this centralizes URL logic
	isHTTPS := s.urlService.IsHTTPS(nil)

	if !isAuthenticated {
		// Clear the cookie by setting expired date
		cookie := &http.Cookie{
			Name:     "is_admin_authenticated",
			Value:    "",
			Path:     "/",
			HttpOnly: true,
			Secure:   isHTTPS,
			SameSite: http.SameSiteLaxMode,
			Expires:  time.Now().Add(-24 * time.Hour), // Expire immediately
		}
		http.SetCookie(w, cookie)
		return
	}

	cookie := &http.Cookie{
		Name:     "is_admin_authenticated",
		Value:    "true",
		Path:     "/",
		HttpOnly: true,
		Secure:   isHTTPS,
		SameSite: http.SameSiteLaxMode,
		Expires:  time.Now().Add(24 * time.Hour), // Admin sessions typically last longer
	}

	http.SetCookie(w, cookie)
}

// ClearSessionCookieByName clears any session cookie by name
// Uses centralized HTTPS detection via URL service following DDD principles
func (s *CookieSessionService) ClearSessionCookieByName(w http.ResponseWriter, cookieName string) {
	// Use URL service for HTTPS detection - this centralizes URL logic
	isHTTPS := s.urlService.IsHTTPS(nil)

	cookie := &http.Cookie{
		Name:     cookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   isHTTPS,
		SameSite: http.SameSiteLaxMode,
		Expires:  time.Now().Add(-24 * time.Hour), // Expire immediately
	}

	http.SetCookie(w, cookie)
}
