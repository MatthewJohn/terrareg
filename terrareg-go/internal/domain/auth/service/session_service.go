package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	terraregAppConfig "github.com/matthewjohn/terrareg/terrareg-go/internal/config"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/auth"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/auth/repository"
)

// SessionService handles session management independent of authentication methods
// Matches Python's session management approach
type SessionService struct {
	sessionRepo repository.SessionRepository
	config      *SessionConfig
}

// SessionConfig contains session configuration
type SessionConfig struct {
	DefaultTTL      time.Duration `json:"default_ttl"`
	MaxTTL          time.Duration `json:"max_ttl"`
	CleanupInterval time.Duration `json:"cleanup_interval"`
	CookieName      string        `json:"cookie_name"`
	SecureCookies   bool          `json:"secure_cookies"`
	HttpOnlyCookies bool          `json:"http_only_cookies"`
}

// NewSessionService creates a new session service
func NewSessionService(appConfig *terraregAppConfig.Config, sessionRepo repository.SessionRepository, config *SessionConfig) *SessionService {
	if config == nil {
		config = DefaultSessionConfig(appConfig)
	}

	ss := &SessionService{
		sessionRepo: sessionRepo,
		config:      config,
	}

	return ss
}

// DefaultSessionConfig returns default session configuration
func DefaultSessionConfig(config *terraregAppConfig.Config) *SessionConfig {
	return &SessionConfig{
		DefaultTTL:      1 * time.Hour,
		MaxTTL:          24 * time.Hour,
		CleanupInterval: 1 * time.Hour,
		CookieName:      config.SessionCookieName,
		SecureCookies:   true,
		HttpOnlyCookies: true,
	}
}

// CreateSession creates a new session with provider-specific auth data
func (ss *SessionService) CreateSession(ctx context.Context, authMethod auth.AuthMethod, providerData map[string]interface{}, ttl *time.Duration) (*auth.Session, error) {
	sessionID := fmt.Sprintf("%d", time.Now().UnixNano()) // Use timestamp-based ID like Python
	expiry := time.Now().Add(ss.getEffectiveTTL(ttl))

	// Create provider source auth data matching Python's approach
	providerSourceAuth := ProviderSourceAuthData{
		AuthMethod: authMethod.GetProviderType(),
		Username:   authMethod.GetUsername(),
		IsAdmin:    authMethod.IsAdmin(),
		Data:       providerData,
		CreatedAt:  time.Now(),
	}

	// Serialize to JSON
	authDataJSON, err := json.Marshal(providerSourceAuth)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize provider auth data: %w", err)
	}

	now := time.Now()
	session := &auth.Session{
		ID:                 sessionID,
		Expiry:             expiry,
		ProviderSourceAuth: authDataJSON,
		CreatedAt:          &now,
		LastAccessedAt:     nil,
	}

	if err := ss.sessionRepo.Create(ctx, session); err != nil {
		return nil, fmt.Errorf("failed to save session: %w", err)
	}

	return session, nil
}

// ValidateSession checks if a session is valid and not expired
func (ss *SessionService) ValidateSession(ctx context.Context, sessionID string) (*auth.Session, error) {
	session, err := ss.sessionRepo.FindByID(ctx, sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to find session: %w", err)
	}

	if session == nil {
		return nil, ErrSessionNotFound
	}

	if session.IsExpired() {
		// Clean up expired session
		_ = ss.sessionRepo.Delete(ctx, sessionID)
		return nil, ErrSessionExpired
	}

	// Update last accessed time
	now := time.Now()
	session.LastAccessedAt = &now
	if err := ss.sessionRepo.UpdateProviderSourceAuth(ctx, sessionID, session.ProviderSourceAuth); err != nil {
		// Log error but don't fail validation
		fmt.Printf("Warning: failed to update session last accessed time: %v\n", err)
	}

	return session, nil
}

// GetAuthProviderData extracts provider-specific auth data from a session
func (ss *SessionService) GetAuthProviderData(session *auth.Session) (*ProviderSourceAuthData, error) {
	if session.ProviderSourceAuth == nil {
		return nil, fmt.Errorf("session has no provider auth data")
	}

	var authData ProviderSourceAuthData
	if err := json.Unmarshal(session.ProviderSourceAuth, &authData); err != nil {
		return nil, fmt.Errorf("failed to deserialize provider auth data: %w", err)
	}

	return &authData, nil
}

// RefreshSession extends a session's expiry time
func (ss *SessionService) RefreshSession(ctx context.Context, sessionID string, ttl *time.Duration) (*auth.Session, error) {
	session, err := ss.ValidateSession(ctx, sessionID)
	if err != nil {
		return nil, err
	}

	// Update expiry time
	newExpiry := time.Now().Add(ss.getEffectiveTTL(ttl))
	session.Expiry = newExpiry

	// Update provider source auth with new expiry
	if err := ss.sessionRepo.UpdateProviderSourceAuth(ctx, sessionID, session.ProviderSourceAuth); err != nil {
		return nil, fmt.Errorf("failed to update session: %w", err)
	}

	return session, nil
}

// InvalidateSession removes a session
func (ss *SessionService) InvalidateSession(ctx context.Context, sessionID string) error {
	return ss.sessionRepo.Delete(ctx, sessionID)
}

// CleanupExpiredSessions removes all expired sessions
func (ss *SessionService) CleanupExpiredSessions(ctx context.Context) error {
	return ss.sessionRepo.CleanupExpired(ctx)
}

// GetSessionFromRequest extracts session ID from HTTP request headers/cookies
func (ss *SessionService) GetSessionFromRequest(headers map[string]string) (*string, error) {
	// Check for session cookie first
	if cookie, exists := headers["Cookie"]; exists {
		sessionID := ss.extractSessionFromCookie(cookie)
		if sessionID != nil {
			return sessionID, nil
		}
	}

	// Check for session header
	if sessionID, exists := headers["X-Session-ID"]; exists && sessionID != "" {
		return &sessionID, nil
	}

	// Check for Authorization header (for API-based sessions)
	if authHeader, exists := headers["Authorization"]; exists {
		if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
			token := authHeader[7:]
			// This could be a session token or API key
			return &token, nil
		}
	}

	return nil, ErrNoSessionFound
}

// extractSessionFromCookie extracts session ID from cookie string
func (ss *SessionService) extractSessionFromCookie(cookie string) *string {
	// Parse cookie string to find session cookie
	// Format: "name1=value1; name2=value2; ..."
	cookies := ss.parseCookies(cookie)

	for name, value := range cookies {
		if name == ss.config.CookieName {
			return &value
		}
	}

	return nil
}

// parseCookies parses cookie string into map
func (ss *SessionService) parseCookies(cookieString string) map[string]string {
	cookies := make(map[string]string)

	// Simple cookie parsing - in production, use proper cookie parsing
	pairs := ss.splitCookieString(cookieString)

	for _, pair := range pairs {
		parts := ss.splitNameValuePair(pair)
		if len(parts) == 2 {
			cookies[parts[0]] = parts[1]
		}
	}

	return cookies
}

// splitCookieString splits cookie string by semicolons
func (ss *SessionService) splitCookieString(cookieString string) []string {
	var pairs []string
	current := ""

	for _, char := range cookieString {
		if char == ';' {
			if current != "" {
				pairs = append(pairs, ss.trim(current))
				current = ""
			}
		} else {
			current += string(char)
		}
	}

	if current != "" {
		pairs = append(pairs, ss.trim(current))
	}

	return pairs
}

// splitNameValuePair splits name=value pair
func (ss *SessionService) splitNameValuePair(pair string) []string {
	for i, char := range pair {
		if char == '=' {
			return []string{ss.trim(pair[:i]), ss.trim(pair[i+1:])}
		}
	}
	return []string{ss.trim(pair), ""}
}

// trim removes leading/trailing whitespace
func (ss *SessionService) trim(s string) string {
	// Simple trim implementation
	start := 0
	end := len(s)

	for start < end && (s[start] == ' ' || s[start] == '\t') {
		start++
	}

	for end > start && (s[end-1] == ' ' || s[end-1] == '\t') {
		end--
	}

	return s[start:end]
}

// getEffectiveTTL returns the effective TTL for a session
func (ss *SessionService) getEffectiveTTL(ttl *time.Duration) time.Duration {
	if ttl != nil && *ttl > 0 {
		if *ttl > ss.config.MaxTTL {
			return ss.config.MaxTTL
		}
		return *ttl
	}
	return ss.config.DefaultTTL
}

// ProviderSourceAuthData represents provider-specific authentication data
// Matches Python's provider_source_auth field structure
type ProviderSourceAuthData struct {
	AuthMethod auth.AuthMethodType    `json:"auth_method"`
	Username   string                 `json:"username"`
	IsAdmin    bool                   `json:"is_admin"`
	Data       map[string]interface{} `json:"data"`
	CreatedAt  time.Time              `json:"created_at"`
	UpdatedAt  *time.Time             `json:"updated_at,omitempty"`
}

// Update updates the provider auth data
func (psad *ProviderSourceAuthData) Update(newData map[string]interface{}) {
	now := time.Now()
	psad.Data = newData
	psad.UpdatedAt = &now
}

// GetProviderData retrieves provider-specific data
func (psad *ProviderSourceAuthData) GetProviderData(key string) (interface{}, bool) {
	value, exists := psad.Data[key]
	return value, exists
}

// SetProviderData sets provider-specific data
func (psad *ProviderSourceAuthData) SetProviderData(key string, value interface{}) {
	if psad.Data == nil {
		psad.Data = make(map[string]interface{})
	}
	psad.Data[key] = value
	now := time.Now()
	psad.UpdatedAt = &now
}

// Session errors
var (
	ErrSessionNotFound = fmt.Errorf("session not found")
	ErrSessionExpired  = fmt.Errorf("session expired")
	ErrNoSessionFound  = fmt.Errorf("no session found")
	ErrInvalidSession  = fmt.Errorf("invalid session")
)
