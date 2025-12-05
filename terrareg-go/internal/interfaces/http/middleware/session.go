package middleware

import (
	"context"
	"net/http"

	"github.com/rs/zerolog"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/auth/service"
)

// SessionMiddleware handles session management for HTTP requests
type SessionMiddleware struct {
	sessionService *service.CookieSessionService
	logger         zerolog.Logger
}

// NewSessionMiddleware creates a new session middleware
func NewSessionMiddleware(sessionService *service.CookieSessionService, logger zerolog.Logger) *SessionMiddleware {
	return &SessionMiddleware{
		sessionService: sessionService,
		logger:         logger,
	}
}

// Session extracts and validates session from cookie, adding it to the request context
func (m *SessionMiddleware) Session(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		// Extract session cookie
		cookie, err := r.Cookie(m.sessionService.GetSessionCookieName())
		if err != nil {
			// No session cookie - continue without session
			m.logger.Warn().Err(err).Msg("No session cookie")
			next.ServeHTTP(w, r)
			return
		}

		// Validate session
		sessionData, err := m.sessionService.ValidateSession(ctx, cookie.Value)
		if err != nil {
			// Invalid session - clear cookie and continue
			m.logger.Warn().Err(err).Msg("Invalid session cookie, clearing")
			m.sessionService.ClearSessionCookie(w)
			next.ServeHTTP(w, r)
			return
		}

		// Add session data to context
		ctx = withSessionData(ctx, sessionData)

		// Update the request with the new context
		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)
	})
}

// RequireSession requires a valid session, otherwise returns 401
func (m *SessionMiddleware) RequireSession(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sessionData := GetSessionData(r.Context())
		if sessionData == nil {
			http.Error(w, "Authentication required", http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// RequireAdminSession requires a valid admin session, otherwise returns 403
func (m *SessionMiddleware) RequireAdminSession(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sessionData := GetSessionData(r.Context())
		if sessionData == nil || !sessionData.IsAdmin {
			http.Error(w, "Admin access required", http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// Session context key type to avoid context key collisions
type sessionKey string

const sessionDataKey sessionKey = "sessionData"

// withSessionData adds session data to the context
func withSessionData(ctx context.Context, sessionData *service.SessionData) context.Context {
	return context.WithValue(ctx, sessionDataKey, sessionData)
}

// GetSessionData retrieves session data from the context
func GetSessionData(ctx context.Context) *service.SessionData {
	if sessionData, ok := ctx.Value(sessionDataKey).(*service.SessionData); ok {
		return sessionData
	}
	return nil
}

// GetCSRFToken retrieves CSRF token from session data in context
func GetCSRFToken(ctx context.Context) string {
	if sessionData := GetSessionData(ctx); sessionData != nil {
		return sessionData.CSRFToken
	}
	return ""
}

// GetUserID retrieves user ID from session data in context
func GetUserID(ctx context.Context) string {
	if sessionData := GetSessionData(ctx); sessionData != nil {
		return sessionData.UserID
	}
	return ""
}

// GetUsername retrieves username from session data in context
func GetUsername(ctx context.Context) string {
	if sessionData := GetSessionData(ctx); sessionData != nil {
		return sessionData.Username
	}
	return ""
}

// IsAdmin checks if the current session is admin
func IsAdmin(ctx context.Context) bool {
	if sessionData := GetSessionData(ctx); sessionData != nil {
		return sessionData.IsAdmin
	}
	return false
}
