package middleware

import (
	"context"
	"net/http"

	"github.com/rs/zerolog"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/auth/service"
)

// SessionMiddleware handles session management for HTTP requests
type SessionMiddleware struct {
	authService *service.AuthenticationService
	logger      zerolog.Logger
}

// NewSessionMiddleware creates a new session middleware
func NewSessionMiddleware(authService *service.AuthenticationService, logger zerolog.Logger) *SessionMiddleware {
	return &SessionMiddleware{
		authService: authService,
		logger:      logger,
	}
}

// Session extracts and validates session from cookie, adding it to the request context
func (m *SessionMiddleware) Session(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		// Use authentication service to validate request
		authCtx, err := m.authService.ValidateRequest(ctx, r)
		if err != nil {
			// Log error but continue without authentication
			m.logger.Warn().Err(err).Msg("Failed to validate authentication")
			next.ServeHTTP(w, r)
			return
		}

		// Add authentication context to request context
		ctx = withAuthenticationContext(ctx, authCtx)

		// Update the request with the new context
		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)
	})
}

// Authentication context key type to avoid context key collisions
type authContextKey string

const authenticationContextKey authContextKey = "authenticationContext"

// withAuthenticationContext adds authentication context to the context
func withAuthenticationContext(ctx context.Context, authCtx *service.AuthenticationContext) context.Context {
	return context.WithValue(ctx, authenticationContextKey, authCtx)
}

// GetAuthenticationContext retrieves authentication context from the context
func GetAuthenticationContext(ctx context.Context) *service.AuthenticationContext {
	if authCtx, ok := ctx.Value(authenticationContextKey).(*service.AuthenticationContext); ok {
		return authCtx
	}
	return nil
}

// Legacy session context functions for backward compatibility
type sessionKey string

const sessionDataKey sessionKey = "sessionData"

// withSessionData adds session data to the context
func withSessionData(ctx context.Context, sessionData *service.SessionData) context.Context {
	return context.WithValue(ctx, sessionDataKey, sessionData)
}

// GetSessionData retrieves session data from the context
func GetSessionData(ctx context.Context) *service.SessionData {
	if authCtx := GetAuthenticationContext(ctx); authCtx != nil {
		return authCtx.SessionData
	}

	if sessionData, ok := ctx.Value(sessionDataKey).(*service.SessionData); ok {
		return sessionData
	}
	return nil
}

// GetCSRFToken retrieves CSRF token from session data in context
func GetCSRFToken(ctx context.Context) string {
	// CSRF token is not part of the new SessionData structure
	// This function returns empty string for compatibility
	return ""
}

// GetUserID retrieves user ID from session data in context
func GetUserID(ctx context.Context) string {
	// User ID is not part of the new SessionData structure
	// Return SessionID as a fallback
	if sessionData := GetSessionData(ctx); sessionData != nil {
		return sessionData.SessionID
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
