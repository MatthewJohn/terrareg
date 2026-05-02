package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/security/csrf"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/logging"
)

// CSRFMiddleware provides CSRF protection for HTTP requests
type CSRFMiddleware struct {
	validator *csrf.SecureTokenValidator
	logger    logging.Logger
}

// NewCSRFMiddleware creates a new CSRF middleware
func NewCSRFMiddleware(logger logging.Logger) *CSRFMiddleware {
	return &CSRFMiddleware{
		validator: csrf.NewSecureTokenValidator(),
		logger:    logger,
	}
}

// RequireCSRF wraps a handler to require CSRF token validation
func (m *CSRFMiddleware) RequireCSRF(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Only validate CSRF tokens for state-changing methods
		if !m.requiresCSRFProtection(r.Method) {
			next.ServeHTTP(w, r)
			return
		}

		// Get expected CSRF token from session
		expectedToken := m.getCSRFTokenFromSession(r)

		// Get provided CSRF token from request
		providedToken := m.getCSRFTokenFromRequest(r)

		// Validate the token (required=true for session-based auth)
		err := m.validator.ValidateToken(expectedToken, providedToken, true)
		if err != nil {
			m.logger.Warn().Err(err).Str("method", r.Method).Str("path", r.URL.Path).Msg("CSRF validation failed")
			http.Error(w, "Invalid CSRF token", http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// requiresCSRFProtection determines if a request method requires CSRF protection
func (m *CSRFMiddleware) requiresCSRFProtection(method string) bool {
	// CSRF protection is required for state-changing methods
	switch method {
	case "POST", "PUT", "PATCH", "DELETE":
		return true
	default:
		return false
	}
}

// getCSRFTokenFromSession extracts CSRF token from session
func (m *CSRFMiddleware) getCSRFTokenFromSession(r *http.Request) csrf.CSRFToken {
	// TODO: Implement proper session extraction
	// For now, check for session cookie and retrieve token from session store
	cookie, err := r.Cookie("session_id")
	if err != nil {
		return ""
	}

	// TODO: Implement session lookup - for now return empty
	_ = cookie
	return ""
}

// getCSRFTokenFromRequest extracts CSRF token from HTTP request
// For backwards compatibility with Python implementation, this checks:
// 1. JSON body csrf_token field
// 2. Form data csrf_token field
// 3. X-CSRF-Token header
func (m *CSRFMiddleware) getCSRFTokenFromRequest(r *http.Request) csrf.CSRFToken {
	// Check in request body for JSON requests
	if r.Header.Get("Content-Type") == "application/json" {
		// Read body to extract CSRF token
		bodyBytes, err := io.ReadAll(r.Body)
		if err == nil && len(bodyBytes) > 0 {
			// Parse JSON to extract csrf_token
			var jsonData map[string]interface{}
			if err := json.Unmarshal(bodyBytes, &jsonData); err == nil {
				if csrfToken, ok := jsonData["csrf_token"].(string); ok {
					// Replace body so it can be read again by the handler
					r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
					return csrf.CSRFToken(csrfToken)
				}
			}
			// Replace body even if parsing failed
			r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		}
		return ""
	}

	// Check in form data for regular form submissions
	if err := r.ParseForm(); err == nil {
		if token := r.FormValue("csrf_token"); token != "" {
			return csrf.CSRFToken(token)
		}
	}

	// Check in headers
	if token := r.Header.Get("X-CSRF-Token"); token != "" {
		return csrf.CSRFToken(token)
	}

	return ""
}
